package fs

import (
	"context"
	"encoding/json"
	goerrors "errors"
	"sort"
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
	if f != nil && f.defaultOverlayPolicy != "" {
		entry.OverlayPolicy = f.defaultOverlayPolicy
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

	if f.metadataStore == nil {
		if val, ok := f.metadata.Load(id); ok {
			if inode, ok := val.(*Inode); ok {
				return f.metadataEntryFromInode(id, inode, time.Now().UTC()), nil
			}
		}
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
		return nil, err
	}

	return nil, metadata.ErrNotFound
}

// entryFromDriveItem builds a structured metadata entry from a DriveItem snapshot.
func (f *Filesystem) entryFromDriveItem(item *graph.DriveItem, snapshot time.Time) *metadata.Entry {
	if item == nil {
		return nil
	}
	if snapshot.IsZero() {
		snapshot = time.Now().UTC()
	}

	entry := &metadata.Entry{
		ID:            item.ID,
		Name:          item.Name,
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     snapshot,
		UpdatedAt:     snapshot,
		Pin: metadata.PinState{
			Mode: metadata.PinModeUnset,
		},
	}

	if item.Parent != nil {
		entry.ParentID = item.Parent.ID
	}
	if !isLocalID(item.ID) {
		entry.RemoteID = item.ID
	}
	if item.ETag != "" {
		entry.ETag = item.ETag
	}
	if item.ModTime != nil {
		ts := item.ModTime.UTC()
		entry.LastModified = &ts
	}
	if item.IsDir() {
		entry.ItemType = metadata.ItemKindDirectory
		entry.State = metadata.ItemStateHydrated
		entry.SubdirCount = uint32(item.Folder.ChildCount)
		if entry.Mode == 0 {
			entry.Mode = fuse.S_IFDIR | 0755
		}
	} else {
		entry.Size = item.Size
		if entry.Mode == 0 {
			entry.Mode = fuse.S_IFREG | 0644
		}
	}

	if f != nil && f.defaultOverlayPolicy != "" {
		entry.OverlayPolicy = f.defaultOverlayPolicy
	}

	return entry
}

// applyDriveItemToEntry mutates an existing metadata entry with fields from a DriveItem.
func (f *Filesystem) applyDriveItemToEntry(entry *metadata.Entry, item *graph.DriveItem, snapshot time.Time) {
	if entry == nil || item == nil {
		return
	}
	if snapshot.IsZero() {
		snapshot = time.Now().UTC()
	}

	entry.ID = item.ID
	entry.Name = item.Name
	if item.Parent != nil {
		entry.ParentID = item.Parent.ID
	}
	if !isLocalID(item.ID) {
		entry.RemoteID = item.ID
	}
	if item.ETag != "" {
		entry.ETag = item.ETag
	}
	if item.ModTime != nil {
		ts := item.ModTime.UTC()
		entry.LastModified = &ts
	}

	if item.IsDir() {
		entry.ItemType = metadata.ItemKindDirectory
		entry.SubdirCount = uint32(item.Folder.ChildCount)
		if entry.Mode == 0 {
			entry.Mode = fuse.S_IFDIR | 0755
		}
		if entry.State == "" {
			entry.State = metadata.ItemStateHydrated
		}
	} else {
		entry.ItemType = metadata.ItemKindFile
		entry.Size = item.Size
		if entry.Mode == 0 {
			entry.Mode = fuse.S_IFREG | 0644
		}
		if entry.State == "" {
			entry.State = metadata.ItemStateGhost
		}
	}

	if entry.OverlayPolicy == "" {
		entry.OverlayPolicy = metadata.OverlayPolicyRemoteWins
		if f != nil && f.defaultOverlayPolicy != "" {
			entry.OverlayPolicy = f.defaultOverlayPolicy
		}
	}

	if entry.Pin.Mode == "" {
		entry.Pin.Mode = metadata.PinModeUnset
	}

	entry.UpdatedAt = snapshot
}

// cloneMetadataEntry creates a shallow copy with cloned slices and maps for safe comparison.
func cloneMetadataEntry(entry *metadata.Entry) *metadata.Entry {
	if entry == nil {
		return nil
	}
	copied := *entry
	if entry.Children != nil {
		copied.Children = cloneStringSlice(entry.Children)
	}
	if entry.Xattrs != nil {
		copied.Xattrs = cloneXattrs(entry.Xattrs)
	}
	if entry.Pin.Since != nil {
		ts := *entry.Pin.Since
		copied.Pin.Since = &ts
	}
	if entry.LastModified != nil {
		ts := entry.LastModified.UTC()
		copied.LastModified = &ts
	}
	if entry.LastHydrated != nil {
		ts := entry.LastHydrated.UTC()
		copied.LastHydrated = &ts
	}
	if entry.LastUploaded != nil {
		ts := entry.LastUploaded.UTC()
		copied.LastUploaded = &ts
	}
	return &copied
}

// upsertDriveItemEntry writes a DriveItem snapshot into metadata_v2 and returns the updated entry with its prior value.
func (f *Filesystem) upsertDriveItemEntry(ctx context.Context, item *graph.DriveItem, snapshot time.Time) (*metadata.Entry, *metadata.Entry, error) {
	if f.metadataStore == nil {
		return nil, nil, goerrors.New("metadata store not initialized")
	}
	if item == nil {
		return nil, nil, goerrors.New("drive item is nil")
	}

	var previous *metadata.Entry
	updated, err := f.metadataStore.Update(ctx, item.ID, func(entry *metadata.Entry) error {
		if entry == nil {
			return metadata.ErrNotFound
		}
		previous = cloneMetadataEntry(entry)
		f.applyDriveItemToEntry(entry, item, snapshot)
		entry.PendingRemote = false
		return nil
	})
	if err == metadata.ErrNotFound {
		entry := f.entryFromDriveItem(item, snapshot)
		if entry == nil {
			return nil, nil, metadata.ErrNotFound
		}
		entry.PendingRemote = false
		if saveErr := f.metadataStore.Save(ctx, entry); saveErr != nil {
			return nil, nil, saveErr
		}
		return entry, nil, nil
	}
	return updated, previous, err
}

// replaceParentChildren overwrites the parent's child list based on the provided entries.
func (f *Filesystem) replaceParentChildren(ctx context.Context, parentID string, children []*metadata.Entry) error {
	if f.metadataStore == nil || parentID == "" {
		return nil
	}
	childIDs := make([]string, 0, len(children))
	var dirCount uint32
	for _, child := range children {
		if child == nil {
			continue
		}
		childIDs = append(childIDs, child.ID)
		if child.ItemType == metadata.ItemKindDirectory {
			dirCount++
		}
	}
	sort.Strings(childIDs)

	_, err := f.metadataStore.Update(ctx, parentID, func(entry *metadata.Entry) error {
		if entry == nil {
			return metadata.ErrNotFound
		}
		entry.Children = childIDs
		entry.SubdirCount = dirCount
		entry.PendingRemote = false
		return nil
	})
	if err != nil && err != metadata.ErrNotFound {
		return err
	}
	return nil
}

// moveChildBetweenParents updates parent child lists when an item moves.
func (f *Filesystem) moveChildBetweenParents(ctx context.Context, oldParentID, newParentID string, child *metadata.Entry) {
	if child == nil || f.metadataStore == nil {
		return
	}
	if oldParentID != "" && oldParentID != newParentID {
		_ = f.removeChildFromParent(ctx, oldParentID, child.ID, child.ItemType == metadata.ItemKindDirectory)
	}
	if newParentID != "" {
		_ = f.addChildToParent(ctx, newParentID, child)
	}
}

// addChildToParent appends a child ID to the parent's metadata, updating subdir counts.
func (f *Filesystem) addChildToParent(ctx context.Context, parentID string, child *metadata.Entry) error {
	if f.metadataStore == nil || parentID == "" || child == nil {
		return nil
	}
	_, err := f.metadataStore.Update(ctx, parentID, func(entry *metadata.Entry) error {
		if entry == nil {
			return metadata.ErrNotFound
		}
		present := false
		for _, existing := range entry.Children {
			if existing == child.ID {
				present = true
				break
			}
		}
		if !present {
			entry.Children = append(entry.Children, child.ID)
			if child.ItemType == metadata.ItemKindDirectory {
				entry.SubdirCount++
			}
			sort.Strings(entry.Children)
		}
		entry.PendingRemote = false
		return nil
	})
	if err != nil && err != metadata.ErrNotFound {
		return err
	}
	return nil
}

// removeChildFromParent removes a child ID from the parent's metadata.
func (f *Filesystem) removeChildFromParent(ctx context.Context, parentID, childID string, childWasDir bool) error {
	if f.metadataStore == nil || parentID == "" || childID == "" {
		return nil
	}
	_, err := f.metadataStore.Update(ctx, parentID, func(entry *metadata.Entry) error {
		if entry == nil {
			return metadata.ErrNotFound
		}
		out := entry.Children[:0]
		removed := false
		for _, existing := range entry.Children {
			if existing == childID {
				removed = true
				continue
			}
			out = append(out, existing)
		}
		entry.Children = out
		if removed && childWasDir && entry.SubdirCount > 0 {
			entry.SubdirCount--
		}
		entry.PendingRemote = false
		return nil
	})
	if err != nil && err != metadata.ErrNotFound {
		return err
	}
	return nil
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

func (f *Filesystem) persistMetadataEntry(id string, inode *Inode) {
	if f.metadataStore == nil || inode == nil {
		return
	}
	entry := f.metadataEntryFromInode(id, inode, time.Now().UTC())
	if entry == nil {
		return
	}
	if err := f.metadataStore.Save(context.Background(), entry); err != nil {
		logging.Debug().
			Err(err).
			Str("id", id).
			Msg("Failed to persist metadata entry")
	}
}

func (f *Filesystem) transitionItemState(id string, target metadata.ItemState, opts ...metadata.TransitionOption) {
	if f.stateManager == nil || id == "" {
		return
	}
	if _, err := f.stateManager.Transition(context.Background(), id, target, opts...); err != nil && !goerrors.Is(err, metadata.ErrNotFound) {
		logging.Debug().
			Err(err).
			Str("id", id).
			Str("state", string(target)).
			Msg("Metadata state transition failed")
	}
}

// transitionToState transitions via the state manager, forcing the transition when the current state matches.
func (f *Filesystem) transitionToState(id string, target metadata.ItemState, opts ...metadata.TransitionOption) {
	if id == "" {
		return
	}
	if f.metadataStore != nil {
		if entry, err := f.metadataStore.Get(context.Background(), id); err == nil && entry.State == target {
			opts = append(opts, metadata.ForceTransition())
		}
	}
	f.transitionItemState(id, target, opts...)
}

func (f *Filesystem) markEntryDeleted(id string) {
	if id == "" {
		return
	}
	if f.metadataStore != nil {
		entry, err := f.metadataStore.Get(context.Background(), id)
		if err == nil && entry.Virtual {
			return
		}
	}
	f.transitionItemState(id, metadata.ItemStateDeleted)
	if f.metadataStore == nil {
		return
	}
	_, err := f.metadataStore.Update(context.Background(), id, func(entry *metadata.Entry) error {
		if entry == nil {
			return metadata.ErrNotFound
		}
		entry.Children = nil
		entry.SubdirCount = 0
		entry.PendingRemote = false
		entry.LastHydrated = nil
		entry.LastUploaded = nil
		entry.Hydration = metadata.HydrationState{}
		entry.Upload = metadata.UploadState{}
		entry.LastError = nil
		return nil
	})
	if err != nil && err != metadata.ErrNotFound {
		logging.Debug().
			Err(err).
			Str("id", id).
			Msg("Failed to scrub metadata entry after delete")
	}
}

func (f *Filesystem) updateMetadataFromDelta(id string, delta *graph.DriveItem) {
	if id == "" || delta == nil || f.metadataStore == nil {
		return
	}
	_, err := f.metadataStore.Update(context.Background(), id, func(entry *metadata.Entry) error {
		if entry == nil {
			return metadata.ErrNotFound
		}
		if delta.Name != "" {
			entry.Name = delta.Name
		}
		if delta.Parent != nil && delta.Parent.ID != "" {
			entry.ParentID = delta.Parent.ID
		}
		if delta.ETag != "" {
			entry.ETag = delta.ETag
		}
		if !delta.IsDir() && delta.Size != 0 {
			entry.Size = delta.Size
		}
		if delta.ModTime != nil {
			ts := delta.ModTime.UTC()
			entry.LastModified = &ts
		}
		entry.PendingRemote = false
		return nil
	})
	if err != nil && err != metadata.ErrNotFound {
		logging.Debug().Err(err).Str("id", id).Msg("Failed to update metadata from delta")
	}
}

func (f *Filesystem) ensureInodeFromMetadataStore(id string) *Inode {
	if f.metadataStore == nil || id == "" {
		return nil
	}
	entry, err := f.metadataStore.Get(context.Background(), id)
	if err != nil {
		return nil
	}
	if entry.State == metadata.ItemStateDeleted {
		return nil
	}
	inode := f.inodeFromMetadataEntry(entry)
	if inode == nil {
		return nil
	}
	f.InsertID(id, inode)
	return inode
}

func (f *Filesystem) loadLegacyMetadataEntry(id string) (*metadata.Entry, error) {
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
		if err := f.metadataStore.Save(context.Background(), entry); err != nil {
			logging.Warn().Err(err).Str("id", entry.ID).Msg("Failed to persist legacy metadata into metadata_v2")
		}
	}
	return entry, nil
}
