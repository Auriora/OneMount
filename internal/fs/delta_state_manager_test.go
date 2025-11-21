package fs

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func newTestFilesystemWithMetadata(t *testing.T) *Filesystem {
	t.Helper()
	tmp := t.TempDir()

	db, err := bolt.Open(filepath.Join(tmp, "meta.db"), 0600, &bolt.Options{Timeout: time.Second})
	require.NoError(t, err)
	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketMetadataV2)
		return err
	}))

	store, err := metadata.NewBoltStore(db, bucketMetadataV2)
	require.NoError(t, err)
	stateMgr, err := metadata.NewStateManager(store)
	require.NoError(t, err)

	return &Filesystem{
		db:            db,
		metadataStore: store,
		stateManager:  stateMgr,
		content:       NewLoopbackCacheWithSize(filepath.Join(tmp, "content"), 0),
		nodeIndex:     make(map[uint64]*Inode),
		statuses:      make(map[string]FileStatusInfo),
		uploads: &UploadManager{
			deletionQueue:              make(chan string, 1),
			sessions:                   make(map[string]*UploadSession),
			sessionPriorities:          make(map[string]UploadPriority),
			pendingHighPriorityUploads: make(map[string]bool),
			pendingLowPriorityUploads:  make(map[string]bool),
			shutdownContext:            context.Background(),
			shutdownCancel:             func() {},
		},
	}
}

func seedEntry(t *testing.T, fs *Filesystem, entry *metadata.Entry) {
	t.Helper()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	if entry.UpdatedAt.IsZero() {
		entry.UpdatedAt = entry.CreatedAt
	}
	if entry.OverlayPolicy == "" {
		entry.OverlayPolicy = metadata.OverlayPolicyRemoteWins
	}
	require.NoError(t, fs.metadataStore.Save(context.Background(), entry))
}

func TestApplyDeltaSetsGhostOnRemoteChange(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()
	parent := &metadata.Entry{
		ID:            "parent",
		Name:          "parent",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, parent)

	child := &metadata.Entry{
		ID:            "child",
		Name:          "file.txt",
		ParentID:      parent.ID,
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		ETag:          "old-etag",
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, child)
	require.NoError(t, fs.addChildToParent(context.Background(), parent.ID, child))
	require.NoError(t, fs.content.Insert(child.ID, []byte("payload")))
	require.True(t, fs.content.HasContent(child.ID))

	delta := &graph.DriveItem{
		ID:     child.ID,
		Name:   child.Name,
		Parent: &graph.DriveItemParent{ID: parent.ID},
		File:   &graph.File{},
		ETag:   "new-etag",
		Size:   1024,
	}

	require.NoError(t, fs.applyDelta(delta))

	updated, err := fs.metadataStore.Get(context.Background(), child.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateGhost, updated.State)
	require.False(t, fs.content.HasContent(child.ID))
}

func TestApplyDeltaHydratesWhenMetadataMatches(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()
	parent := &metadata.Entry{
		ID:            "parent",
		Name:          "parent",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, parent)

	child := &metadata.Entry{
		ID:            "child",
		Name:          "file.txt",
		ParentID:      parent.ID,
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		ETag:          "same-etag",
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, child)
	require.NoError(t, fs.addChildToParent(context.Background(), parent.ID, child))

	delta := &graph.DriveItem{
		ID:     child.ID,
		Name:   child.Name,
		Parent: &graph.DriveItemParent{ID: parent.ID},
		File:   &graph.File{},
		ETag:   "same-etag",
		Size:   2048,
	}

	require.NoError(t, fs.applyDelta(delta))

	updated, err := fs.metadataStore.Get(context.Background(), child.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, updated.State)
	require.Nil(t, updated.LastError)
}

func TestApplyDeltaMarksDeletedAndScrubsParent(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()
	parent := &metadata.Entry{
		ID:            "parent",
		Name:          "parent",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, parent)

	child := &metadata.Entry{
		ID:            "child",
		Name:          "file.txt",
		ParentID:      parent.ID,
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, child)
	require.NoError(t, fs.addChildToParent(context.Background(), parent.ID, child))

	delta := &graph.DriveItem{
		ID:     child.ID,
		Name:   child.Name,
		Parent: &graph.DriveItemParent{ID: parent.ID},
		Deleted: &graph.Deleted{
			State: "deleted",
		},
	}

	require.NoError(t, fs.applyDelta(delta))

	updated, err := fs.metadataStore.Get(context.Background(), child.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateDeleted, updated.State)

	parentUpdated, err := fs.metadataStore.Get(context.Background(), parent.ID)
	require.NoError(t, err)
	for _, cid := range parentUpdated.Children {
		require.NotEqual(t, child.ID, cid)
	}
}
