package fs

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/metadata"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

// TestStateTransitionAtomicity verifies that state transitions are atomic
// and no intermediate inconsistent states are visible.
// Validates: Requirements 21.1-21.10
func TestUT_FS_State_TransitionAtomicity(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	// Create a test entry
	entry := &metadata.Entry{
		ID:            "test-file",
		Name:          "test.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, entry)

	// Transition from GHOST to HYDRATING
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("test-worker"))

	// Verify the transition was atomic - state should be HYDRATING
	updated, err := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrating, updated.State)
	require.NotNil(t, updated.Hydration.StartedAt)
	require.Equal(t, "test-worker", updated.Hydration.WorkerID)

	// Transition to HYDRATED
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("test-worker"))

	// Verify atomic transition to HYDRATED
	updated, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, updated.State)
	require.NotNil(t, updated.LastHydrated)
	require.NotNil(t, updated.Hydration.CompletedAt)
	require.Nil(t, updated.LastError)
}

// TestNoIntermediateInconsistentStates verifies that during state transitions,
// no intermediate inconsistent states are visible to concurrent readers.
// Validates: Requirements 21.1-21.10
func TestUT_FS_State_NoIntermediateInconsistentStates(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	entry := &metadata.Entry{
		ID:            "concurrent-file",
		Name:          "concurrent.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, entry)

	var wg sync.WaitGroup
	inconsistentStateFound := false
	var mu sync.Mutex

	// Start multiple readers that check state consistency
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
				if err != nil {
					continue
				}

				// Check for inconsistent states
				switch retrieved.State {
				case metadata.ItemStateHydrating:
					// If HYDRATING, StartedAt must be set
					if retrieved.Hydration.StartedAt == nil {
						mu.Lock()
						inconsistentStateFound = true
						mu.Unlock()
					}
				case metadata.ItemStateHydrated:
					// If HYDRATED, LastHydrated should be set
					if retrieved.LastHydrated == nil && !retrieved.Virtual {
						mu.Lock()
						inconsistentStateFound = true
						mu.Unlock()
					}
				case metadata.ItemStateError:
					// If ERROR, LastError must be set
					if retrieved.LastError == nil {
						mu.Lock()
						inconsistentStateFound = true
						mu.Unlock()
					}
				}
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Perform state transitions while readers are active
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)

		// GHOST -> HYDRATING
		fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
			metadata.WithHydrationEvent(),
			metadata.WithWorker("writer"))

		time.Sleep(10 * time.Millisecond)

		// HYDRATING -> HYDRATED
		fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
			metadata.WithHydrationEvent(),
			metadata.WithWorker("writer"))

		time.Sleep(10 * time.Millisecond)

		// HYDRATED -> DIRTY_LOCAL
		fs.transitionItemState(entry.ID, metadata.ItemStateDirtyLocal,
			metadata.WithUploadEvent(),
			metadata.WithWorker("writer"))

		time.Sleep(10 * time.Millisecond)

		// DIRTY_LOCAL -> HYDRATED
		fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
			metadata.WithUploadEvent(),
			metadata.WithWorker("writer"))
	}()

	wg.Wait()

	require.False(t, inconsistentStateFound, "Inconsistent state detected during concurrent access")
}

// TestStatePersistenceAcrossRestarts verifies that state transitions
// are persisted correctly and survive filesystem restarts.
// Validates: Requirements 21.1-21.10
func TestUT_FS_State_PersistenceAcrossRestarts(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "meta.db")

	// Create initial filesystem and perform transitions
	{
		db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
		require.NoError(t, err)
		require.NoError(t, db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(bucketMetadataV2)
			return err
		}))

		store, err := metadata.NewBoltStore(db, bucketMetadataV2)
		require.NoError(t, err)
		stateMgr, err := metadata.NewStateManager(store)
		require.NoError(t, err)

		fs := &Filesystem{
			db:            db,
			metadataStore: store,
			stateManager:  stateMgr,
			content:       NewLoopbackCacheWithSize(filepath.Join(tmp, "content"), 0),
		}

		now := time.Now().UTC()
		entry := &metadata.Entry{
			ID:            "persist-file",
			Name:          "persist.txt",
			ParentID:      "parent",
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateGhost,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		require.NoError(t, store.Save(context.Background(), entry))

		// Perform state transitions
		fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
			metadata.WithHydrationEvent(),
			metadata.WithWorker("persist-worker"))

		fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
			metadata.WithHydrationEvent(),
			metadata.WithWorker("persist-worker"),
			metadata.WithETag("test-etag"),
			metadata.WithSize(1024))

		// Verify state before closing
		retrieved, err := store.Get(context.Background(), entry.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
		require.Equal(t, "test-etag", retrieved.ETag)
		require.Equal(t, uint64(1024), retrieved.Size)

		// Close database
		require.NoError(t, db.Close())
	}

	// Reopen database and verify persistence
	{
		db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
		require.NoError(t, err)
		defer db.Close()

		store, err := metadata.NewBoltStore(db, bucketMetadataV2)
		require.NoError(t, err)

		// Verify state persisted correctly
		retrieved, err := store.Get(context.Background(), "persist-file")
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
		require.Equal(t, "test-etag", retrieved.ETag)
		require.Equal(t, uint64(1024), retrieved.Size)
		require.NotNil(t, retrieved.LastHydrated)
		require.NotNil(t, retrieved.Hydration.CompletedAt)
	}
}

// TestConcurrentStateTransitionSafety verifies that concurrent state transitions
// on different files are safe and don't interfere with each other.
// Validates: Requirements 21.1-21.10
func TestUT_FS_State_ConcurrentStateTransitionSafety(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	// Create multiple test entries
	numFiles := 20
	fileIDs := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		fileID := fmt.Sprintf("concurrent-file-%d", i)
		fileIDs[i] = fileID

		entry := &metadata.Entry{
			ID:            fileID,
			Name:          fmt.Sprintf("file-%d.txt", i),
			ParentID:      "parent",
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateGhost,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		seedEntry(t, fs, entry)
	}

	var wg sync.WaitGroup
	errors := make(chan error, numFiles)

	// Perform concurrent state transitions on different files
	for _, fileID := range fileIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// GHOST -> HYDRATING
			fs.transitionItemState(id, metadata.ItemStateHydrating,
				metadata.WithHydrationEvent(),
				metadata.WithWorker("worker-"+id))

			// Small delay to simulate work
			time.Sleep(time.Millisecond)

			// HYDRATING -> HYDRATED
			fs.transitionItemState(id, metadata.ItemStateHydrated,
				metadata.WithHydrationEvent(),
				metadata.WithWorker("worker-"+id))

			// Verify final state
			retrieved, err := fs.metadataStore.Get(context.Background(), id)
			if err != nil {
				errors <- err
				return
			}
			if retrieved.State != metadata.ItemStateHydrated {
				errors <- fmt.Errorf("file %s: expected HYDRATED, got %s", id, retrieved.State)
			}
		}(fileID)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		require.NoError(t, err)
	}

	// Verify all files reached HYDRATED state
	for _, fileID := range fileIDs {
		retrieved, err := fs.metadataStore.Get(context.Background(), fileID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
		require.NotNil(t, retrieved.LastHydrated)
	}
}

// TestConcurrentStateTransitionOnSameFile verifies that concurrent transitions
// on the same file are handled safely without corruption.
// Validates: Requirements 21.1-21.10
func TestUT_FS_State_ConcurrentStateTransitionOnSameFile(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	entry := &metadata.Entry{
		ID:            "same-file",
		Name:          "same.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, entry)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Multiple goroutines try to transition the same file
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Try to transition GHOST -> HYDRATING
			fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
				metadata.WithHydrationEvent(),
				metadata.WithWorker(fmt.Sprintf("worker-%d", workerID)))

			time.Sleep(time.Millisecond)
		}(i)
	}

	wg.Wait()

	// Verify the file is in a valid state (should be HYDRATING)
	retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrating, retrieved.State)
	require.NotNil(t, retrieved.Hydration.StartedAt)
	require.NotEmpty(t, retrieved.Hydration.WorkerID)
}

// TestStateTransitionWithError verifies that error transitions
// preserve state consistency and error information.
// Validates: Requirements 21.1-21.10
func TestUT_FS_State_TransitionWithError(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	entry := &metadata.Entry{
		ID:            "error-file",
		Name:          "error.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, entry)

	// Transition to HYDRATING
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("error-worker"))

	// Transition to ERROR with error information
	testErr := errors.New("download failed: network timeout")
	fs.transitionItemState(entry.ID, metadata.ItemStateError,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("error-worker"),
		metadata.WithTransitionError(testErr, true))

	// Verify error state is consistent
	retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateError, retrieved.State)
	require.NotNil(t, retrieved.LastError)
	require.Equal(t, "download failed: network timeout", retrieved.LastError.Message)
	require.True(t, retrieved.LastError.Temporary)
	require.NotNil(t, retrieved.Hydration.Error)
	require.Equal(t, retrieved.LastError.Message, retrieved.Hydration.Error.Message)
	require.NotNil(t, retrieved.Hydration.CompletedAt)
}

// TestVirtualFileStateImmutability verifies that virtual files
// remain in HYDRATED state and cannot transition to other states.
// Validates: Requirements 21.10
func TestUT_FS_State_VirtualFileStateImmutability(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	entry := &metadata.Entry{
		ID:            "virtual-file",
		Name:          ".xdg-volume-info",
		ParentID:      "root",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		Virtual:       true,
		OverlayPolicy: metadata.OverlayPolicyLocalWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, entry)

	// Attempt to transition virtual file to GHOST (should fail or be ignored)
	fs.transitionItemState(entry.ID, metadata.ItemStateGhost)

	// Verify virtual file remains HYDRATED
	retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
	require.True(t, retrieved.Virtual)

	// Attempt to transition to DIRTY_LOCAL (should fail or be ignored)
	fs.transitionItemState(entry.ID, metadata.ItemStateDirtyLocal)

	// Verify still HYDRATED
	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
}

// TestCompleteStateLifecycle verifies a complete state lifecycle
// from GHOST through various states and back.
// Validates: Requirements 21.1-21.10
func TestUT_FS_State_CompleteStateLifecycle(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	entry := &metadata.Entry{
		ID:            "lifecycle-file",
		Name:          "lifecycle.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntry(t, fs, entry)

	// GHOST -> HYDRATING
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("lifecycle-worker"))

	retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrating, retrieved.State)

	// HYDRATING -> HYDRATED
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("lifecycle-worker"),
		metadata.WithETag("etag-1"))

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
	require.Equal(t, "etag-1", retrieved.ETag)

	// HYDRATED -> DIRTY_LOCAL (user modifies file)
	fs.transitionItemState(entry.ID, metadata.ItemStateDirtyLocal,
		metadata.WithUploadEvent(),
		metadata.WithWorker("upload-worker"))

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateDirtyLocal, retrieved.State)

	// DIRTY_LOCAL -> HYDRATED (upload succeeds)
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
		metadata.WithUploadEvent(),
		metadata.WithWorker("upload-worker"),
		metadata.WithETag("etag-2"))

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
	require.Equal(t, "etag-2", retrieved.ETag)

	// HYDRATED -> GHOST (cache eviction)
	fs.transitionItemState(entry.ID, metadata.ItemStateGhost)

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateGhost, retrieved.State)
	require.Equal(t, "etag-2", retrieved.ETag) // ETag preserved

	// GHOST -> DELETED_LOCAL (user deletes file)
	fs.transitionItemState(entry.ID, metadata.ItemStateDeleted)

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateDeleted, retrieved.State)
}
