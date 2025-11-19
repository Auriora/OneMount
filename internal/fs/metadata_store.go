package fs

import (
	"context"
	"encoding/json"
	goerrors "errors"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// metadataEntryFromInode captures a snapshot of an inode for persistence.
func (f *Filesystem) metadataEntryFromInode(id string, inode *Inode, snapshot time.Time) *metadata.Entry {
	if inode == nil {
		return nil
	}

	inode.mu.RLock()
	defer inode.mu.RUnlock()

	isDir := inode.DriveItem.IsDir()
	hasChanges := inode.hasChanges
	isVirtual := inode.virtual

	entry := &metadata.Entry{
		ID:            id,
		Name:          inode.DriveItem.Name,
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		Children:      cloneStringSlice(inode.children),
		SubdirCount:   inode.subdir,
		Mode:          inode.mode,
		Xattrs:        cloneXattrs(inode.xattrs),
		Size:          inode.DriveItem.Size,
		Pin: metadata.PinState{
			Mode: metadata.PinModeUnset,
		},
		CreatedAt: snapshot,
		UpdatedAt: snapshot,
	}

	if inode.DriveItem.Parent != nil {
		entry.ParentID = inode.DriveItem.Parent.ID
	}
	if !isLocalID(id) {
		entry.RemoteID = id
	}
	if inode.DriveItem.ETag != "" {
		entry.ETag = inode.DriveItem.ETag
	}
	if inode.DriveItem.ModTime != nil {
		ts := inode.DriveItem.ModTime.UTC()
		entry.LastModified = &ts
	}
	if isDir {
		entry.ItemType = metadata.ItemKindDirectory
		entry.State = metadata.ItemStateHydrated
	}

	if hasChanges {
		entry.State = metadata.ItemStateDirtyLocal
	}

	if isVirtual {
		entry.Virtual = true
		entry.State = metadata.ItemStateHydrated
		entry.OverlayPolicy = metadata.OverlayPolicyLocalWins
	}

	if f != nil && f.content != nil && !isDir && !isVirtual && !hasChanges {
		if f.content.HasContent(id) {
			entry.State = metadata.ItemStateHydrated
		}
	}

	if f != nil && f.isChildPendingRemote(id) {
		entry.PendingRemote = true
	}

	return entry
}

func cloneStringSlice(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	out := make([]string, len(src))
	copy(out, src)
	return out
}

func cloneXattrs(src map[string][]byte) map[string][]byte {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string][]byte, len(src))
	for k, v := range src {
		if v == nil {
			out[k] = nil
			continue
		}
		buf := make([]byte, len(v))
		copy(buf, v)
		out[k] = buf
	}
	return out
}

func (f *Filesystem) metadataEntryFromSerializedInode(id string, raw []byte, snapshot time.Time) (*metadata.Entry, error) {
	inode, err := NewInodeJSON(raw)
	if err != nil {
		return nil, err
	}
	return f.metadataEntryFromInode(id, inode, snapshot), nil
}

func (f *Filesystem) inodeFromMetadataEntry(entry *metadata.Entry) *Inode {
	if entry == nil {
		return nil
	}

	inode := &Inode{
		mu:       &sync.RWMutex{},
		children: cloneStringSlice(entry.Children),
		subdir:   entry.SubdirCount,
		mode:     entry.Mode,
		xattrs:   cloneXattrs(entry.Xattrs),
		virtual:  entry.Virtual,
	}

	itemID := entry.ID
	inode.DriveItem.ID = itemID
	inode.DriveItem.Name = entry.Name
	inode.DriveItem.Size = entry.Size
	inode.DriveItem.ETag = entry.ETag

	if entry.ParentID != "" {
		inode.DriveItem.Parent = &graph.DriveItemParent{
			ID: entry.ParentID,
		}
	}
	if entry.LastModified != nil {
		ts := entry.LastModified.UTC()
		inode.DriveItem.ModTime = &ts
	}

	switch entry.ItemType {
	case metadata.ItemKindDirectory:
		inode.DriveItem.Folder = &graph.Folder{ChildCount: entry.SubdirCount}
		if inode.mode == 0 {
			inode.mode = fuse.S_IFDIR | 0755
		}
	default:
		inode.DriveItem.File = &graph.File{}
		if inode.mode == 0 {
			inode.mode = fuse.S_IFREG | 0644
		}
	}

	switch entry.State {
	case metadata.ItemStateDirtyLocal, metadata.ItemStateConflict:
		inode.hasChanges = true
	default:
		inode.hasChanges = false
	}

	return inode
}

// bootstrapMetadataStore initializes the v2 bucket and migrates legacy entries.
func (f *Filesystem) bootstrapMetadataStore() error {
	if f.db == nil {
		return errors.New("filesystem database is not initialized")
	}
	now := time.Now().UTC()
	return f.db.Update(func(tx *bolt.Tx) error {
		v2 := tx.Bucket(bucketMetadataV2)
		if v2 == nil {
			return errors.New("metadata_v2 bucket missing")
		}
		if stats := v2.Stats(); stats.KeyN > 0 {
			return nil
		}
		legacy := tx.Bucket(bucketMetadata)
		if legacy == nil {
			return nil
		}
		migrated := 0
		if err := legacy.ForEach(func(k, v []byte) error {
			if len(v) == 0 {
				return nil
			}
			entry, err := f.metadataEntryFromSerializedInode(string(k), v, now)
			if err != nil {
				logging.Warn().
					Err(err).
					Str("id", string(k)).
					Msg("Skipping legacy metadata during migration")
				return nil
			}
			if entry == nil {
				return nil
			}
			if err := entry.Validate(); err != nil {
				logging.Warn().
					Err(err).
					Str("id", entry.ID).
					Msg("Skipping invalid metadata entry during migration")
				return nil
			}
			payload, err := json.Marshal(entry)
			if err != nil {
				return errors.Wrap(err, "marshal metadata entry")
			}
			if err := v2.Put(k, payload); err != nil {
				return errors.Wrap(err, "persist metadata_v2 entry")
			}
			migrated++
			return nil
		}); err != nil {
			return err
		}
		if migrated > 0 {
			logging.Info().Int("entries", migrated).Msg("Migrated legacy metadata to metadata_v2 bucket")
		}
		return nil
	})
}

// loadMetadataEntry retrieves a metadata entry from the structured store,
// converting from legacy sources when required.
func (f *Filesystem) loadMetadataEntry(id string) (*metadata.Entry, error) {
	if id == "" {
		return nil, goerrors.New("metadata id is required")
	}

	ctx := context.Background()
	if f.metadataStore != nil {
		entry, err := f.metadataStore.Get(ctx, id)
		if err == nil {
			return entry, nil
		}
		if !goerrors.Is(err, metadata.ErrNotFound) {
			return nil, err
		}
	}

	if inode := f.GetID(id); inode != nil {
		if entry := f.metadataEntryFromInode(id, inode, time.Now().UTC()); entry != nil {
			return entry, nil
		}
	}

	if f.db == nil {
		return nil, metadata.ErrNotFound
	}

	var entry *metadata.Entry
	err := f.db.View(func(tx *bolt.Tx) error {
		legacy := tx.Bucket(bucketMetadata)
		if legacy == nil {
			return metadata.ErrNotFound
		}
		raw := legacy.Get([]byte(id))
		if len(raw) == 0 && id == "root" {
			raw = legacy.Get([]byte(f.root))
		}
		if len(raw) == 0 {
			return metadata.ErrNotFound
		}
		converted, convErr := f.metadataEntryFromSerializedInode(id, raw, time.Now().UTC())
		if convErr != nil {
			return convErr
		}
		entry = converted
		return nil
	})
	if err != nil {
		return nil, err
	}

	if entry != nil && f.metadataStore != nil {
		if err := f.metadataStore.Save(ctx, entry); err != nil {
			logging.Warn().Err(err).Str("id", entry.ID).Msg("Failed to persist legacy metadata into metadata_v2")
		}
	}
	return entry, nil
}

// GetMetadataEntry returns the structured metadata entry for the provided ID.
func (f *Filesystem) GetMetadataEntry(id string) (*metadata.Entry, error) {
	return f.loadMetadataEntry(id)
}

// SaveMetadataEntry validates and persists the provided entry.
func (f *Filesystem) SaveMetadataEntry(entry *metadata.Entry) error {
	if entry == nil {
		return goerrors.New("metadata entry is nil")
	}
	if f.metadataStore == nil {
		return goerrors.New("metadata store not initialized")
	}
	return f.metadataStore.Save(context.Background(), entry)
}

// UpdateMetadataEntry applies the provided mutation function atomically.
func (f *Filesystem) UpdateMetadataEntry(id string, fn func(*metadata.Entry) error) (*metadata.Entry, error) {
	if f.metadataStore == nil {
		return nil, goerrors.New("metadata store not initialized")
	}
	return f.metadataStore.Update(context.Background(), id, fn)
}
