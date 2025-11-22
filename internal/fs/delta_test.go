package fs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/socketio"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/stretchr/testify/require"
)

func seedDeltaTestFile(t *testing.T, filesystem *Filesystem, rootID, fileID, name, quickHash string, content []byte) *graph.DriveItem {
	t.Helper()
	modTime := time.Now().Add(-2 * time.Hour)
	item := &graph.DriveItem{
		ID:   fileID,
		Name: name,
		Size: uint64(len(content)),
		File: &graph.File{
			Hashes: graph.Hashes{
				QuickXorHash: quickHash,
			},
		},
		Parent: &graph.DriveItemParent{
			ID: rootID,
		},
		ETag:    "etag-local-" + fileID,
		ModTime: &modTime,
	}

	inode := NewInodeDriveItem(item)
	filesystem.InsertChild(rootID, inode)

	require.NoError(t, filesystem.content.Insert(fileID, content))
	filesystem.persistMetadataEntry(fileID, inode)
	filesystem.transitionItemState(fileID, metadata.ItemStateHydrated)
	return item
}

func TestApplyDeltaPersistsMetadataOnMetadataOnlyChange(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "DeltaMetadataPersistenceFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, fx interface{}) {
			unit := fx.(*framework.UnitTestFixture)
			fsFixture := unit.SetupData.(*helpers.FSTestFixture)
			filesystem := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID

			localItem := seedDeltaTestFile(t, filesystem, rootID, "delta-meta-file", "delta-meta.txt", "hash-meta", []byte("local-content"))

			remoteMod := time.Now()
			delta := *localItem
			delta.ModTime = &remoteMod
			delta.ETag = localItem.ETag // metadata-only change; keep same etag to avoid invalidation
			delta.Parent = &graph.DriveItemParent{ID: rootID}

			require.NoError(t, filesystem.applyDelta(&delta))

			entry, err := filesystem.GetMetadataEntry(localItem.ID)
			require.NoError(t, err)
			require.Equal(t, localItem.ETag, entry.ETag)
			require.Equal(t, metadata.ItemStateHydrated, entry.State)
			require.NotNil(t, entry.LastModified)
			require.WithinDuration(t, remoteMod.UTC(), *entry.LastModified, time.Second)
		})
	})
}

func TestApplyDeltaRemoteInvalidationTransitionsMetadata(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "DeltaInvalidationPersistenceFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, fx interface{}) {
			unit := fx.(*framework.UnitTestFixture)
			fsFixture := unit.SetupData.(*helpers.FSTestFixture)
			filesystem := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID

			localItem := seedDeltaTestFile(t, filesystem, rootID, "delta-change-file", "delta-change.txt", "hash-initial", []byte("initial-content"))

			remoteMod := time.Now()
			delta := *localItem
			delta.ModTime = &remoteMod
			delta.ETag = "etag-remote-2"
			delta.Parent = &graph.DriveItemParent{ID: rootID}
			delta.File = &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "hash-remote-new",
				},
			}
			delta.Size = uint64(len("remote-new-content"))

			require.NoError(t, filesystem.applyDelta(&delta))

			entry, err := filesystem.GetMetadataEntry(localItem.ID)
			require.NoError(t, err)
			require.Equal(t, metadata.ItemStateGhost, entry.State)
			require.False(t, filesystem.content.HasContent(localItem.ID))
			require.Equal(t, uint64(len("remote-new-content")), entry.Size)
			require.Equal(t, "etag-remote-2", entry.ETag)
		})
	})
}

func TestApplyDeltaMoveUpdatesMetadataEntry(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "DeltaMoveMetadataFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, fx interface{}) {
			unit := fx.(*framework.UnitTestFixture)
			fsFixture := unit.SetupData.(*helpers.FSTestFixture)
			filesystem := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID

			root := filesystem.GetID(rootID)
			require.NotNil(t, root)

			newParent := &graph.DriveItem{
				ID:   "subdir-id",
				Name: "subdir",
				Folder: &graph.Folder{
					ChildCount: 0,
				},
				Parent: &graph.DriveItemParent{ID: rootID},
			}
			filesystem.InsertChild(rootID, NewInodeDriveItem(newParent))

			file := seedDeltaTestFile(t, filesystem, rootID, "delta-move-file", "delta-move.txt", "hash-move", []byte("move"))

			moveDelta := *file
			moveDelta.Parent = &graph.DriveItemParent{ID: newParent.ID}
			moveDelta.ModTime = timePtr(time.Now())

			require.NoError(t, filesystem.applyDelta(&moveDelta))

			entry, err := filesystem.GetMetadataEntry(file.ID)
			require.NoError(t, err)
			require.Equal(t, newParent.ID, entry.ParentID)
			require.Equal(t, moveDelta.Name, entry.Name)
		})
	})
}

func TestApplyDeltaPinnedFileQueuesHydration(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "DeltaPinnedAutoHydrationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, fx interface{}) {
			unit := fx.(*framework.UnitTestFixture)
			fsFixture := unit.SetupData.(*helpers.FSTestFixture)
			filesystem := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID

			localItem := seedDeltaTestFile(t, filesystem, rootID, "delta-pin-file", "delta-pin.txt", "hash-pin", []byte("pin-content"))
			_, err := filesystem.UpdateMetadataEntry(localItem.ID, func(entry *metadata.Entry) error {
				entry.Pin.Mode = metadata.PinModeAlways
				entry.ItemType = metadata.ItemKindFile
				return nil
			})
			require.NoError(t, err)

			var autoHydrateCount int32
			filesystem.SetTestHooks(&FilesystemTestHooks{
				AutoHydrateHook: func(_ *Filesystem, id string) bool {
					if id == localItem.ID {
						atomic.AddInt32(&autoHydrateCount, 1)
						return true
					}
					return true
				},
			})
			defer filesystem.ClearTestHooks()

			remoteMod := time.Now()
			delta := *localItem
			delta.ModTime = &remoteMod
			delta.ETag = "etag-remote-pin"
			delta.Parent = &graph.DriveItemParent{ID: rootID}
			delta.File = &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "hash-pin-remote",
				},
			}
			delta.Size = uint64(len("pin-remote-content"))

			require.NoError(t, filesystem.applyDelta(&delta))
			require.Equal(t, int32(1), atomic.LoadInt32(&autoHydrateCount), "delta invalidation should auto hydrate pinned items")
		})
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// TestIT_FS_03_01_Delta_SyncOperations_ChangesAreSynced tests delta operations for syncing changes.
//
//	Test Case ID    IT-FS-03-01
//	Title           Delta Sync Operations
//	Description     Tests delta operations for syncing changes
//	Preconditions   None
//	Steps           1. Set up test files/directories
//	                2. Perform operations on the server (create, delete, rename, move)
//	                3. Verify that changes are synced to the client
//	Expected Result Delta operations correctly sync changes
//	Notes: This test verifies that delta operations correctly sync changes from the server to the client.
func TestIT_FS_03_01_Delta_SyncOperations_ChangesAreSynced(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaSyncOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// TODO: Implement the test case
		// 1. Set up test files/directories
		// 2. Perform operations on the server (create, delete, rename, move)
		// 3. Verify that changes are synced to the client
		assert.True(true, "Placeholder assertion")
		t.Skip("Test not implemented yet")
	})
}

// TestIT_FS_04_01_Delta_RemoteContentChange_ClientIsUpdated tests syncing content changes from server to client.
//
//	Test Case ID    IT-FS-04-01
//	Title           Delta Remote Content Change
//	Description     Tests syncing content changes from server to client
//	Preconditions   None
//	Steps           1. Create a file on the client
//	                2. Change the content on the server
//	                3. Verify that the client content is updated
//	Expected Result Remote content changes are synced to the client
//	Notes: This test verifies that content changes on the server are synced to the client.
func TestIT_FS_04_01_Delta_RemoteContentChange_ClientIsUpdated(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaRemoteContentChangeFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// TODO: Implement the test case
		// 1. Create a file on the client
		// 2. Change the content on the server
		// 3. Verify that the client content is updated
		assert.True(true, "Placeholder assertion")
		t.Skip("Test not implemented yet")
	})
}

// TestIT_FS_05_01_Delta_ConflictingChanges_LocalChangesPreserved tests handling of conflicting changes.
//
//	Test Case ID    IT-FS-05-01
//	Title           Delta Conflicting Changes
//	Description     Tests handling of conflicting changes
//	Preconditions   None
//	Steps           1. Create a file with initial content
//	                2. Change the content both locally and remotely
//	                3. Verify that local changes are preserved
//	Expected Result Local changes are preserved when there are conflicts
//	Notes: This test verifies that local changes are preserved when there are conflicts with remote changes.
func TestIT_FS_05_01_Delta_ConflictingChanges_LocalChangesPreserved(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaConflictingChangesFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the filesystem from the fixture
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		// Step 1: Create a file with initial content
		testFileID := "test_conflict_file_123"
		testFileName := "conflict_test.txt"
		initialContent := "Initial content"

		// Create local item with changes
		localItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Size: uint64(len(initialContent)),
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "local_hash_123",
				},
			},
			ModTime: &time.Time{},
			ETag:    "local_etag",
		}
		localTime := time.Now().Add(-1 * time.Hour)
		localItem.ModTime = &localTime

		localInode := NewInodeDriveItem(localItem)
		localInode.hasChanges = true // Mark as having local changes
		filesystem.InsertID(testFileID, localInode)

		// Step 2: Simulate remote changes with different content
		remoteContent := "Different remote content"
		remoteItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Size: uint64(len(remoteContent)),
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "remote_hash_456", // Different hash
				},
			},
			ModTime: &time.Time{},
			ETag:    "remote_etag", // Different ETag
			Parent: &graph.DriveItemParent{
				ID: "parent123",
			},
		}
		remoteTime := time.Now() // More recent than local
		remoteItem.ModTime = &remoteTime

		// Step 3: Apply delta and verify local changes are preserved
		err := filesystem.applyDelta(remoteItem)
		assert.NoError(err, "Should apply delta without error")

		// Verify that local item still exists
		localItemAfter := filesystem.GetID(testFileID)
		assert.NotNil(localItemAfter, "Local item should still exist")

		// Verify that local changes are preserved (hasChanges should still be true)
		assert.True(localItemAfter.HasChanges(), "Local changes should be preserved")

		// Note: In a real conflict scenario, the system should detect the conflict
		// and handle it according to the configured strategy. For this test,
		// we're verifying that the delta application doesn't lose local changes.
	})
}

// TestIT_FS_06_01_Delta_CorruptedCache_ContentIsRestored tests handling of corrupted cache content.
//
//	Test Case ID    IT-FS-06-01
//	Title           Delta Corrupted Cache
//	Description     Tests handling of corrupted cache content
//	Preconditions   None
//	Steps           1. Create a file with correct content
//	                2. Corrupt the cache content
//	                3. Verify that the correct content is restored
//	Expected Result Corrupted cache content is detected and fixed
//	Notes: This test verifies that corrupted cache content is detected and fixed.
func TestIT_FS_06_01_Delta_CorruptedCache_ContentIsRestored(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaCorruptedCacheFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// TODO: Implement the test case
		// 1. Create a file with correct content
		// 2. Corrupt the cache content
		// 3. Verify that the correct content is restored
		assert.True(true, "Placeholder assertion")
		t.Skip("Test not implemented yet")
	})
}

// TestIT_FS_07_01_Delta_FolderDeletion_EmptyFoldersAreDeleted tests folder deletion during sync.
//
//	Test Case ID    IT-FS-07-01
//	Title           Delta Folder Deletion
//	Description     Tests folder deletion during sync
//	Preconditions   None
//	Steps           1. Create a nested directory structure
//	                2. Delete the folder on the server
//	                3. Verify that the folder is deleted on the client
//	Expected Result Folders are deleted when empty
//	Notes: This test verifies that empty folders are deleted during sync.
func TestIT_FS_07_01_Delta_FolderDeletion_EmptyFoldersAreDeleted(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaFolderDeletionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// TODO: Implement the test case
		// 1. Create a nested directory structure
		// 2. Delete the folder on the server
		// 3. Verify that the folder is deleted on the client
		assert.True(true, "Placeholder assertion")
		t.Skip("Test not implemented yet")
	})
}

// TestIT_FS_08_01_Delta_NonEmptyFolderDeletion_FolderIsPreserved tests handling of non-empty folder deletion.
//
//	Test Case ID    IT-FS-08-01
//	Title           Delta Non-Empty Folder Deletion
//	Description     Tests handling of non-empty folder deletion
//	Preconditions   None
//	Steps           1. Create a folder with files
//	                2. Attempt to delete the folder via delta sync
//	                3. Verify that the folder is not deleted until empty
//	Expected Result Non-empty folders are not deleted
//	Notes: This test verifies that non-empty folders are not deleted until they are empty.
func TestIT_FS_08_01_Delta_NonEmptyFolderDeletion_FolderIsPreserved(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaNonEmptyFolderDeletionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// TODO: Implement the test case
		// 1. Create a folder with files
		// 2. Attempt to delete the folder via delta sync
		// 3. Verify that the folder is not deleted until empty
		assert.True(true, "Placeholder assertion")
		t.Skip("Test not implemented yet")
	})
}

// TestIT_FS_09_01_Delta_UnchangedContent_ModTimeIsPreserved tests preservation of modification times.
//
//	Test Case ID    IT-FS-09-01
//	Title           Delta Unchanged Content
//	Description     Tests preservation of modification times
//	Preconditions   None
//	Steps           1. Create a file with initial content
//	                2. Wait for delta sync to run multiple times
//	                3. Verify that the modification time is not updated
//	Expected Result Modification times are preserved when content is unchanged
//	Notes: This test verifies that modification times are preserved when content is unchanged.
func TestIT_FS_09_01_Delta_UnchangedContent_ModTimeIsPreserved(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaUnchangedContentFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// TODO: Implement the test case
		// 1. Create a file with initial content
		// 2. Wait for delta sync to run multiple times
		// 3. Verify that the modification time is not updated
		assert.True(true, "Placeholder assertion")
		t.Skip("Test not implemented yet")
	})
}

// TestIT_FS_10_01_Delta_MissingHash_HandledCorrectly tests handling of deltas with missing hash.
//
//	Test Case ID    IT-FS-10-01
//	Title           Delta Missing Hash
//	Description     Tests handling of deltas with missing hash
//	Preconditions   None
//	Steps           1. Create a file in the filesystem
//	                2. Apply a delta with missing hash information
//	                3. Verify that the delta is applied without errors
//	Expected Result Deltas with missing hash information are handled correctly
//	Notes: This test verifies that deltas with missing hash information are handled correctly.
func TestIT_FS_10_01_Delta_MissingHash_HandledCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DeltaMissingHashFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// TODO: Implement the test case
		// 1. Create a file in the filesystem
		// 2. Apply a delta with missing hash information
		// 3. Verify that the delta is applied without errors
		assert.True(true, "Placeholder assertion")
		t.Skip("Test not implemented yet")
	})
}

func TestDesiredDeltaIntervalUsesActiveWindow(t *testing.T) {
	fs := &Filesystem{}
	fs.ConfigureDeltaTuning(DeltaTuning{
		ActiveInterval: time.Minute,
		ActiveWindow:   2 * time.Minute,
	})
	fs.deltaInterval = 5 * time.Minute
	fs.RecordForegroundActivity()

	if interval := fs.desiredDeltaInterval(); interval != time.Minute {
		t.Fatalf("expected active interval 1m, got %s", interval)
	}
}

func TestDesiredDeltaIntervalFallsBackAfterWindow(t *testing.T) {
	fs := &Filesystem{}
	fs.ConfigureDeltaTuning(DeltaTuning{
		ActiveInterval: 30 * time.Second,
		ActiveWindow:   30 * time.Second,
	})
	fs.deltaInterval = 2 * time.Minute
	fs.lastForegroundActivity.Store(time.Now().Add(-time.Minute).UnixNano())

	if interval := fs.desiredDeltaInterval(); interval != 2*time.Minute {
		t.Fatalf("expected base interval 2m, got %s", interval)
	}
}

type fakeNotifier struct {
	active bool
	health socketio.HealthState
}

func (f *fakeNotifier) Start(context.Context) error          { return nil }
func (f *fakeNotifier) Stop(context.Context) error           { return nil }
func (f *fakeNotifier) Notifications() <-chan struct{}       { return nil }
func (f *fakeNotifier) IsActive() bool                       { return f.active }
func (f *fakeNotifier) HealthSnapshot() socketio.HealthState { return f.health }

type fakeMetadataStore struct {
	items map[string]*metadata.Entry
}

func newFakeMetadataStore() *fakeMetadataStore {
	return &fakeMetadataStore{items: make(map[string]*metadata.Entry)}
}

func (s *fakeMetadataStore) Get(_ context.Context, id string) (*metadata.Entry, error) {
	entry, ok := s.items[id]
	if !ok {
		return nil, metadata.ErrNotFound
	}
	cp := *entry
	return &cp, nil
}

func (s *fakeMetadataStore) Save(_ context.Context, entry *metadata.Entry) error {
	if entry == nil {
		return metadata.ErrNotFound
	}
	cp := *entry
	s.items[entry.ID] = &cp
	return nil
}

func (s *fakeMetadataStore) Update(ctx context.Context, id string, fn func(*metadata.Entry) error) (*metadata.Entry, error) {
	current, _ := s.Get(ctx, id)
	if current == nil {
		current = &metadata.Entry{ID: id}
	}
	if err := fn(current); err != nil {
		return nil, err
	}
	if err := s.Save(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func TestDesiredDeltaIntervalUsesNotifierHealthHealthy(t *testing.T) {
	fs := &Filesystem{}
	fs.ConfigureRealtime(RealtimeOptions{Enabled: true, FallbackInterval: 45 * time.Minute})
	fs.subscriptionManager = &fakeNotifier{
		active: true,
		health: socketio.HealthState{Status: socketio.StatusHealthy},
	}

	expected := 45 * time.Minute
	if interval := fs.desiredDeltaInterval(); interval != expected {
		t.Fatalf("expected realtime healthy interval %s, got %s", expected, interval)
	}
}

func TestDesiredDeltaIntervalUsesNotifierHealthDegraded(t *testing.T) {
	fs := &Filesystem{}
	fs.ConfigureRealtime(RealtimeOptions{Enabled: true})
	fs.subscriptionManager = &fakeNotifier{
		active: true,
		health: socketio.HealthState{Status: socketio.StatusDegraded, ConsecutiveFailures: 3},
	}

	if interval := fs.desiredDeltaInterval(); interval != defaultPollingInterval {
		t.Fatalf("expected degraded interval %s, got %s", defaultPollingInterval, interval)
	}
}

func TestDesiredDeltaIntervalUsesNotifierHealthFailedRecovery(t *testing.T) {
	fs := &Filesystem{}
	fs.ConfigureRealtime(RealtimeOptions{Enabled: true})
	fs.subscriptionManager = &fakeNotifier{
		active: true,
		health: socketio.HealthState{Status: socketio.StatusFailed, ConsecutiveFailures: 5},
	}

	if interval := fs.desiredDeltaInterval(); interval != defaultRecoveryInterval {
		t.Fatalf("expected recovery interval %s, got %s", defaultRecoveryInterval, interval)
	}
}

func TestApplyDeltaTransitionsStateOnRemoteInvalidation(t *testing.T) {
	store := newFakeMetadataStore()
	manager, err := metadata.NewStateManager(store)
	require.NoError(t, err)
	fs := &Filesystem{
		metadataStore: store,
		stateManager:  manager,
		statuses:      make(map[string]FileStatusInfo),
	}

	ctx := context.Background()
	parent := &metadata.Entry{
		ID:        "parent-id",
		Name:      "Parent",
		ItemType:  metadata.ItemKindDirectory,
		State:     metadata.ItemStateHydrated,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_ = fs.metadataStore.Save(ctx, parent)

	item := &metadata.Entry{
		ID:       "file-id",
		ParentID: "parent-id",
		Name:     "file.txt",
		ItemType: metadata.ItemKindFile,
		State:    metadata.ItemStateDirtyLocal,
		ETag:     "old",
	}
	_ = fs.metadataStore.Save(ctx, item)

	delta := &graph.DriveItem{
		ID:   item.ID,
		Name: item.Name,
		ETag: "new",
		Parent: &graph.DriveItemParent{
			ID: item.ParentID,
		},
		File: &graph.File{},
	}

	err = fs.applyDelta(delta)
	require.NoError(t, err)

	entry, err := fs.metadataStore.Get(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateConflict, entry.State, "dirty local item should become conflict on remote change")
}
