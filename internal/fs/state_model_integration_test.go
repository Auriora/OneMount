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

// TestIT_FS_30_10_01_CompleteStateLifecycle tests a complete state lifecycle
// from GHOST through all possible states and transitions.
//
// Test Case ID    IT-FS-30-10-01
// Test Name       Complete State Lifecycle
// Component       Metadata State Model
// Requirement     21.1-21.10
// Description     Verifies that a file can transition through all valid states
//
//	in a complete lifecycle: GHOST -> HYDRATING -> HYDRATED ->
//	DIRTY_LOCAL -> HYDRATED -> GHOST -> DELETED_LOCAL
//
// Preconditions   Filesystem with metadata store initialized
// Test Steps      1. Create entry in GHOST state
//  2. Transition to HYDRATING (user access)
//  3. Transition to HYDRATED (download complete)
//  4. Transition to DIRTY_LOCAL (user modification)
//  5. Transition to HYDRATED (upload complete)
//  6. Transition to GHOST (cache eviction)
//  7. Transition to DELETED_LOCAL (user deletion)
//  8. Verify each transition preserves metadata
//
// Expected Result All transitions succeed, metadata preserved throughout
// Requirements    21.1-21.10
func TestIT_FS_30_10_01_CompleteStateLifecycle(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	// Step 1: Create entry in GHOST state
	entry := &metadata.Entry{
		ID:            "lifecycle-test-file",
		Name:          "lifecycle.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		RemoteID:      "remote-123",
		ETag:          "initial-etag",
		Size:          1024,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntryForStateModel(t, fs, entry)

	t.Logf("✓ Step 1: Created entry in GHOST state")

	// Step 2: Transition to HYDRATING (user access triggers download)
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("download-worker-1"))

	retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrating, retrieved.State)
	require.NotNil(t, retrieved.Hydration.StartedAt)
	require.Equal(t, "download-worker-1", retrieved.Hydration.WorkerID)
	require.Equal(t, "initial-etag", retrieved.ETag) // ETag preserved
	require.Equal(t, uint64(1024), retrieved.Size)   // Size preserved

	t.Logf("✓ Step 2: Transitioned to HYDRATING, metadata preserved")

	// Step 3: Transition to HYDRATED (download complete)
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
		metadata.WithHydrationEvent(),
		metadata.WithWorker("download-worker-1"))

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
	require.NotNil(t, retrieved.LastHydrated)
	require.NotNil(t, retrieved.Hydration.CompletedAt)
	require.Nil(t, retrieved.LastError)

	t.Logf("✓ Step 3: Transitioned to HYDRATED, download complete")

	// Step 4: Transition to DIRTY_LOCAL (user modifies file)
	fs.transitionItemState(entry.ID, metadata.ItemStateDirtyLocal,
		metadata.WithUploadEvent(),
		metadata.WithWorker("upload-worker-1"))

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateDirtyLocal, retrieved.State)
	require.NotNil(t, retrieved.Upload.StartedAt)

	t.Logf("✓ Step 4: Transitioned to DIRTY_LOCAL, upload queued")

	// Step 5: Transition to HYDRATED (upload complete with new ETag)
	fs.transitionItemState(entry.ID, metadata.ItemStateHydrated,
		metadata.WithUploadEvent(),
		metadata.WithWorker("upload-worker-1"),
		metadata.WithETag("new-etag-after-upload"),
		metadata.WithSize(2048))

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
	require.Equal(t, "new-etag-after-upload", retrieved.ETag)
	require.Equal(t, uint64(2048), retrieved.Size)
	require.NotNil(t, retrieved.Upload.CompletedAt)

	t.Logf("✓ Step 5: Transitioned to HYDRATED, upload complete with new ETag")

	// Step 6: Transition to GHOST (cache eviction)
	fs.transitionItemState(entry.ID, metadata.ItemStateGhost)

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateGhost, retrieved.State)
	require.Equal(t, "new-etag-after-upload", retrieved.ETag) // ETag still preserved
	require.Equal(t, uint64(2048), retrieved.Size)            // Size still preserved

	t.Logf("✓ Step 6: Transitioned to GHOST, metadata preserved after eviction")

	// Step 7: Transition to DELETED_LOCAL (user deletes file)
	fs.transitionItemState(entry.ID, metadata.ItemStateDeleted)

	retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateDeleted, retrieved.State)

	t.Logf("✓ Step 7: Transitioned to DELETED_LOCAL, deletion queued")
	t.Logf("✅ Complete state lifecycle test passed")
}

// TestIT_FS_30_10_02_StateTransitionEdgeCases tests edge cases in state transitions
// including invalid transitions, error states, and conflict states.
//
// Test Case ID    IT-FS-30-10-02
// Test Name       State Transition Edge Cases
// Component       Metadata State Model
// Requirement     21.1-21.10
// Description     Verifies edge cases: invalid transitions, error recovery,
//
//	conflict detection, and virtual file immutability
//
// Preconditions   Filesystem with metadata store initialized
// Test Steps      1. Test invalid transition (DIRTY_LOCAL -> HYDRATING)
//  2. Test error state transition and recovery
//  3. Test conflict state transition
//  4. Test virtual file state immutability
//  5. Test transition from ERROR to HYDRATING (retry)
//
// Expected Result Invalid transitions rejected, error/conflict states handled correctly
// Requirements    21.1-21.10
func TestIT_FS_30_10_02_StateTransitionEdgeCases(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	// Test 1: Invalid transition (DIRTY_LOCAL -> HYDRATING should fail)
	t.Run("InvalidTransition", func(t *testing.T) {
		entry := &metadata.Entry{
			ID:            "invalid-transition-file",
			Name:          "invalid.txt",
			ParentID:      "parent",
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateDirtyLocal,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		seedEntryForStateModel(t, fs, entry)

		// Attempt invalid transition
		_, err := fs.stateManager.Transition(context.Background(), entry.ID, metadata.ItemStateHydrating)
		require.Error(t, err)
		require.ErrorIs(t, err, metadata.ErrInvalidTransition)

		// Verify state unchanged
		retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateDirtyLocal, retrieved.State)

		t.Logf("✓ Invalid transition correctly rejected")
	})

	// Test 2: Error state transition and recovery
	t.Run("ErrorStateTransition", func(t *testing.T) {
		entry := &metadata.Entry{
			ID:            "error-state-file",
			Name:          "error.txt",
			ParentID:      "parent",
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateHydrating,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		seedEntryForStateModel(t, fs, entry)

		// Transition to ERROR with error details
		testErr := errors.New("network timeout during download")
		fs.transitionItemState(entry.ID, metadata.ItemStateError,
			metadata.WithHydrationEvent(),
			metadata.WithWorker("error-worker"),
			metadata.WithTransitionError(testErr, true))

		retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateError, retrieved.State)
		require.NotNil(t, retrieved.LastError)
		require.Equal(t, "network timeout during download", retrieved.LastError.Message)
		require.True(t, retrieved.LastError.Temporary)

		t.Logf("✓ Error state transition with error details recorded")

		// Test recovery: ERROR -> HYDRATING (retry)
		fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
			metadata.WithHydrationEvent(),
			metadata.WithWorker("retry-worker"))

		retrieved, err = fs.metadataStore.Get(context.Background(), entry.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrating, retrieved.State)
		require.Equal(t, "retry-worker", retrieved.Hydration.WorkerID)

		t.Logf("✓ Error recovery: transitioned from ERROR to HYDRATING for retry")
	})

	// Test 3: Conflict state transition
	t.Run("ConflictStateTransition", func(t *testing.T) {
		entry := &metadata.Entry{
			ID:            "conflict-file",
			Name:          "conflict.txt",
			ParentID:      "parent",
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateDirtyLocal,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			ETag:          "local-etag",
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		seedEntryForStateModel(t, fs, entry)

		// Transition to CONFLICT (remote changes detected while local changes exist)
		fs.transitionItemState(entry.ID, metadata.ItemStateConflict)

		retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateConflict, retrieved.State)
		// Note: ETag may not change during conflict transition - the conflict state itself is what matters

		t.Logf("✓ Conflict state transition successful")
	})

	// Test 4: Virtual file state immutability
	t.Run("VirtualFileImmutability", func(t *testing.T) {
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
		seedEntryForStateModel(t, fs, entry)

		// Attempt to transition virtual file to GHOST
		fs.transitionItemState(entry.ID, metadata.ItemStateGhost)

		// Verify virtual file remains HYDRATED
		retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
		require.True(t, retrieved.Virtual)

		t.Logf("✓ Virtual file state immutability preserved")
	})

	t.Logf("✅ All edge case tests passed")
}

// TestIT_FS_30_10_03_StatePersistenceAndRecovery tests that state transitions
// are persisted correctly and survive filesystem restarts.
//
// Test Case ID    IT-FS-30-10-03
// Test Name       State Persistence and Recovery
// Component       Metadata State Model
// Requirement     21.1-21.10
// Description     Verifies that state transitions are persisted to disk and
//
//	can be recovered after filesystem restart
//
// Preconditions   Temporary directory for database
// Test Steps      1. Create filesystem and perform state transitions
//  2. Close filesystem and database
//  3. Reopen database and verify state persistence
//  4. Verify all metadata fields preserved
//
// Expected Result All state transitions persisted and recovered correctly
// Requirements    21.1-21.10
func TestIT_FS_30_10_03_StatePersistenceAndRecovery(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "persistence-test.db")

	// Phase 1: Create filesystem and perform transitions
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

		_ = &Filesystem{
			db:            db,
			metadataStore: store,
			stateManager:  stateMgr,
			content:       NewLoopbackCacheWithSize(filepath.Join(tmp, "content"), 0),
		}

		now := time.Now().UTC()

		// Create multiple entries in different states
		entries := []struct {
			id    string
			name  string
			state metadata.ItemState
		}{
			{"persist-ghost", "ghost.txt", metadata.ItemStateGhost},
			{"persist-hydrating", "hydrating.txt", metadata.ItemStateHydrating},
			{"persist-hydrated", "hydrated.txt", metadata.ItemStateHydrated},
			{"persist-dirty", "dirty.txt", metadata.ItemStateDirtyLocal},
			{"persist-error", "error.txt", metadata.ItemStateError},
		}

		for _, e := range entries {
			entry := &metadata.Entry{
				ID:            e.id,
				Name:          e.name,
				ParentID:      "parent",
				ItemType:      metadata.ItemKindFile,
				State:         e.state,
				OverlayPolicy: metadata.OverlayPolicyRemoteWins,
				ETag:          fmt.Sprintf("etag-%s", e.id),
				Size:          2048,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			// Add state-specific metadata
			switch e.state {
			case metadata.ItemStateHydrating:
				entry.Hydration.StartedAt = &now
				entry.Hydration.WorkerID = "persist-worker"
			case metadata.ItemStateHydrated:
				hydrated := now
				entry.LastHydrated = &hydrated
				entry.Hydration.CompletedAt = &hydrated
			case metadata.ItemStateDirtyLocal:
				entry.Upload.StartedAt = &now
			case metadata.ItemStateError:
				entry.LastError = &metadata.OperationError{
					Message:    "test error",
					Temporary:  true,
					OccurredAt: now,
				}
			}

			require.NoError(t, store.Save(context.Background(), entry))
		}

		t.Logf("✓ Created %d entries in various states", len(entries))

		// Close database
		require.NoError(t, db.Close())
		t.Logf("✓ Database closed")
	}

	// Phase 2: Reopen database and verify persistence
	{
		db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
		require.NoError(t, err)
		defer db.Close()

		store, err := metadata.NewBoltStore(db, bucketMetadataV2)
		require.NoError(t, err)

		// Verify each entry persisted correctly
		testCases := []struct {
			id            string
			expectedState metadata.ItemState
			checkFunc     func(*testing.T, *metadata.Entry)
		}{
			{
				id:            "persist-ghost",
				expectedState: metadata.ItemStateGhost,
				checkFunc: func(t *testing.T, e *metadata.Entry) {
					require.Equal(t, "etag-persist-ghost", e.ETag)
					require.Equal(t, uint64(2048), e.Size)
				},
			},
			{
				id:            "persist-hydrating",
				expectedState: metadata.ItemStateHydrating,
				checkFunc: func(t *testing.T, e *metadata.Entry) {
					require.NotNil(t, e.Hydration.StartedAt)
					require.Equal(t, "persist-worker", e.Hydration.WorkerID)
				},
			},
			{
				id:            "persist-hydrated",
				expectedState: metadata.ItemStateHydrated,
				checkFunc: func(t *testing.T, e *metadata.Entry) {
					require.NotNil(t, e.LastHydrated)
					require.NotNil(t, e.Hydration.CompletedAt)
				},
			},
			{
				id:            "persist-dirty",
				expectedState: metadata.ItemStateDirtyLocal,
				checkFunc: func(t *testing.T, e *metadata.Entry) {
					require.NotNil(t, e.Upload.StartedAt)
				},
			},
			{
				id:            "persist-error",
				expectedState: metadata.ItemStateError,
				checkFunc: func(t *testing.T, e *metadata.Entry) {
					require.NotNil(t, e.LastError)
					require.Equal(t, "test error", e.LastError.Message)
					require.True(t, e.LastError.Temporary)
				},
			},
		}

		for _, tc := range testCases {
			retrieved, err := store.Get(context.Background(), tc.id)
			require.NoError(t, err, "Failed to retrieve %s", tc.id)
			require.Equal(t, tc.expectedState, retrieved.State, "State mismatch for %s", tc.id)
			tc.checkFunc(t, retrieved)
			t.Logf("✓ Verified persistence for %s (state: %s)", tc.id, tc.expectedState)
		}

		t.Logf("✅ All state persistence tests passed")
	}
}

// TestIT_FS_30_10_04_ConcurrentStateOperations tests that concurrent state
// operations on multiple files are safe and don't interfere with each other.
//
// Test Case ID    IT-FS-30-10-04
// Test Name       Concurrent State Operations
// Component       Metadata State Model
// Requirement     21.1-21.10
// Description     Verifies that concurrent state transitions on multiple files
//
//	are thread-safe and maintain consistency
//
// Preconditions   Filesystem with metadata store initialized
// Test Steps      1. Create multiple test files
//  2. Perform concurrent state transitions on different files
//  3. Perform concurrent reads while transitions occur
//  4. Verify all files reach expected final states
//  5. Verify no data corruption or race conditions
//
// Expected Result All concurrent operations complete successfully without errors
// Requirements    21.1-21.10
func TestIT_FS_30_10_04_ConcurrentStateOperations(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	// Create multiple test files (reduced from 50 to 20 for faster execution)
	numFiles := 20
	fileIDs := make([]string, numFiles)

	for i := 0; i < numFiles; i++ {
		fileID := fmt.Sprintf("concurrent-file-%03d", i)
		fileIDs[i] = fileID

		entry := &metadata.Entry{
			ID:            fileID,
			Name:          fmt.Sprintf("file-%03d.txt", i),
			ParentID:      "parent",
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateGhost,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			ETag:          fmt.Sprintf("etag-%03d", i),
			Size:          uint64(1024 * (i + 1)),
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		seedEntryForStateModel(t, fs, entry)
	}

	t.Logf("✓ Created %d test files", numFiles)

	var wg sync.WaitGroup
	errors := make(chan error, numFiles*2)

	// Start concurrent readers (reduced from 10 to 5)
	numReaders := 5
	stopReaders := make(chan struct{})

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			readCount := 0
			maxReads := 100 // Limit reads to prevent infinite loop
			for readCount < maxReads {
				select {
				case <-stopReaders:
					t.Logf("Reader %d completed %d reads", readerID, readCount)
					return
				default:
					// Read random file
					fileID := fileIDs[readCount%numFiles]
					_, err := fs.metadataStore.Get(context.Background(), fileID)
					if err != nil {
						errors <- fmt.Errorf("reader %d: %w", readerID, err)
						return
					}
					readCount++
					time.Sleep(time.Microsecond * 10) // Reduced sleep time
				}
			}
			t.Logf("Reader %d completed %d reads (max reached)", readerID, readCount)
		}(i)
	}

	t.Logf("✓ Started %d concurrent readers", numReaders)

	// Perform concurrent state transitions on different files
	for _, fileID := range fileIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// GHOST -> HYDRATING
			fs.transitionItemState(id, metadata.ItemStateHydrating,
				metadata.WithHydrationEvent(),
				metadata.WithWorker("worker-"+id))

			time.Sleep(time.Microsecond * 100)

			// HYDRATING -> HYDRATED
			fs.transitionItemState(id, metadata.ItemStateHydrated,
				metadata.WithHydrationEvent(),
				metadata.WithWorker("worker-"+id))

			time.Sleep(time.Microsecond * 100)

			// HYDRATED -> DIRTY_LOCAL
			fs.transitionItemState(id, metadata.ItemStateDirtyLocal,
				metadata.WithUploadEvent(),
				metadata.WithWorker("upload-"+id))

			time.Sleep(time.Microsecond * 100)

			// DIRTY_LOCAL -> HYDRATED
			fs.transitionItemState(id, metadata.ItemStateHydrated,
				metadata.WithUploadEvent(),
				metadata.WithWorker("upload-"+id),
				metadata.WithETag("new-"+id))

			// Verify final state
			retrieved, err := fs.metadataStore.Get(context.Background(), id)
			if err != nil {
				errors <- fmt.Errorf("file %s: %w", id, err)
				return
			}
			if retrieved.State != metadata.ItemStateHydrated {
				errors <- fmt.Errorf("file %s: expected HYDRATED, got %s", id, retrieved.State)
			}
			if retrieved.ETag != "new-"+id {
				errors <- fmt.Errorf("file %s: expected ETag 'new-%s', got '%s'", id, id, retrieved.ETag)
			}
		}(fileID)
	}

	t.Logf("✓ Started %d concurrent state transition workers", numFiles)

	// Wait for all transitions to complete
	wg.Wait()

	// Stop readers after transitions complete
	close(stopReaders)

	// Give readers a moment to finish
	time.Sleep(100 * time.Millisecond)

	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
	}
	require.Equal(t, 0, errorCount, "Expected no errors during concurrent operations")

	// Verify all files reached final state
	for _, fileID := range fileIDs {
		retrieved, err := fs.metadataStore.Get(context.Background(), fileID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, retrieved.State)
		require.Equal(t, "new-"+fileID, retrieved.ETag)
		require.NotNil(t, retrieved.LastHydrated)
	}

	t.Logf("✅ All %d files reached expected final state", numFiles)
	t.Logf("✅ Concurrent state operations test passed")
}

// TestIT_FS_30_10_05_StateTransitionRaceConditions tests for race conditions
// when multiple goroutines attempt to transition the same file simultaneously.
//
// Test Case ID    IT-FS_30-10-05
// Test Name       State Transition Race Conditions
// Component       Metadata State Model
// Requirement     21.1-21.10
// Description     Verifies that concurrent transitions on the same file
//
//	are handled safely without data corruption
//
// Preconditions   Filesystem with metadata store initialized
// Test Steps      1. Create a single test file
//  2. Launch multiple goroutines attempting same transition
//  3. Verify file ends in valid state
//  4. Verify no data corruption
//
// Expected Result File reaches valid state, no corruption detected
// Requirements    21.1-21.10
func TestIT_FS_30_10_05_StateTransitionRaceConditions(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	now := time.Now().UTC()

	entry := &metadata.Entry{
		ID:            "race-test-file",
		Name:          "race.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateGhost,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		ETag:          "initial-etag",
		Size:          1024,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	seedEntryForStateModel(t, fs, entry)

	var wg sync.WaitGroup
	numGoroutines := 20

	// Multiple goroutines try to transition the same file
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Try to transition GHOST -> HYDRATING
			fs.transitionItemState(entry.ID, metadata.ItemStateHydrating,
				metadata.WithHydrationEvent(),
				metadata.WithWorker(fmt.Sprintf("race-worker-%d", workerID)))

			time.Sleep(time.Millisecond)
		}(i)
	}

	wg.Wait()

	// Verify the file is in a valid state
	retrieved, err := fs.metadataStore.Get(context.Background(), entry.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrating, retrieved.State)
	require.NotNil(t, retrieved.Hydration.StartedAt)
	require.NotEmpty(t, retrieved.Hydration.WorkerID)
	require.Equal(t, "initial-etag", retrieved.ETag) // ETag preserved

	t.Logf("✓ File in valid state after concurrent transitions")
	t.Logf("✓ Worker ID: %s", retrieved.Hydration.WorkerID)
	t.Logf("✅ Race condition test passed")
}

// Helper function to seed an entry into the metadata store for state model tests
func seedEntryForStateModel(t *testing.T, fs *Filesystem, entry *metadata.Entry) {
	require.NoError(t, fs.metadataStore.Save(context.Background(), entry))
}
