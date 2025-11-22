package fs

import (
	"context"
	"errors"
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

func TestHydrationErrorPersistsLastErrorSnapshot(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()
	entry := &metadata.Entry{
		ID:        "file",
		Name:      "file.txt",
		ParentID:  "parent",
		ItemType:  metadata.ItemKindFile,
		State:     metadata.ItemStateGhost,
		CreatedAt: now,
		UpdatedAt: now,
	}
	seedEntry(t, fs, entry)

	fs.transitionToState(entry.ID, metadata.ItemStateHydrating, metadata.WithHydrationEvent(), metadata.WithWorker("hydrator"))
	err := errors.New("network timeout")
	fs.transitionItemState(entry.ID, metadata.ItemStateError,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("hydrator"),
		metadata.WithTransitionError(err, true))

	updated, getErr := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, getErr)
	require.Equal(t, metadata.ItemStateError, updated.State)
	require.NotNil(t, updated.LastError)
	require.Equal(t, "network timeout", updated.LastError.Message)
	require.True(t, updated.LastError.Temporary)
	require.NotNil(t, updated.Hydration.Error)
	require.Equal(t, updated.LastError.Message, updated.Hydration.Error.Message)
	require.NotNil(t, updated.Hydration.CompletedAt)
}

func TestUploadErrorPersistsSnapshot(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()
	entry := &metadata.Entry{
		ID:        "upload-file",
		Name:      "upload.txt",
		ParentID:  "parent",
		ItemType:  metadata.ItemKindFile,
		State:     metadata.ItemStateGhost,
		CreatedAt: now,
		UpdatedAt: now,
	}
	seedEntry(t, fs, entry)

	fs.transitionToState(entry.ID, metadata.ItemStateDirtyLocal, metadata.WithUploadEvent(), metadata.WithWorker("upload-session"))
	uerr := errors.New("etag mismatch")
	fs.transitionItemState(entry.ID, metadata.ItemStateError,
		metadata.WithUploadEvent(),
		metadata.WithWorker("upload-session"),
		metadata.WithTransitionError(uerr, false))

	updated, getErr := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, getErr)
	require.Equal(t, metadata.ItemStateError, updated.State)
	require.NotNil(t, updated.Upload.LastError)
	require.Equal(t, "etag mismatch", updated.Upload.LastError.Message)
	require.NotNil(t, updated.Upload.CompletedAt)
}

func TestApplyDeltaPinnedItemRequeuesHydration(t *testing.T) {
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
		Children:      []string{"child"},
	}
	child := &metadata.Entry{
		ID:            "child",
		Name:          "file.txt",
		ParentID:      parent.ID,
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		Pin: metadata.PinState{
			Mode: metadata.PinModeAlways,
		},
		CreatedAt: now,
		UpdatedAt: now,
		ETag:      "old-etag",
	}
	seedEntry(t, fs, parent)
	seedEntry(t, fs, child)

	var requeued bool
	fs.SetTestHooks(&FilesystemTestHooks{
		AutoHydrateHook: func(_ *Filesystem, id string) bool {
			requiredID := "child"
			requeued = id == requiredID
			return true // prevent real download queueing
		},
	})

	delta := &graph.DriveItem{
		ID:     "child",
		Name:   "file.txt",
		Parent: &graph.DriveItemParent{ID: parent.ID},
		ETag:   "new-etag",
		File:   &graph.File{},
	}

	require.NoError(t, fs.applyDelta(delta))
	require.True(t, requeued, "expected pinned item to requeue hydration")
	updated, err := fs.metadataStore.Get(context.Background(), child.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateGhost, updated.State)
	require.Equal(t, metadata.PinModeAlways, updated.Pin.Mode)
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
