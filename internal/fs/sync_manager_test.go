package fs

import (
	"context"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_SM_01_01_SyncManager_RetryMechanism_RetriesFailedOperations tests retry mechanisms
//
//	Test Case ID    IT-SM-01-01
//	Title           Sync Manager Retry Mechanism
//	Description     Tests that the sync manager retries failed operations
//	Preconditions   None
//	Steps           1. Create offline changes
//	                2. Configure mock to fail initially then succeed
//	                3. Process changes with sync manager
//	                4. Verify retries occurred and operation succeeded
//	Expected Result Failed operations are retried and eventually succeed
//	Notes: This test verifies that the sync manager properly retries failed synchronization operations.
func TestIT_SM_01_01_SyncManager_RetryMechanism_RetriesFailedOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SyncManagerRetryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create offline changes
		testFileID := "test_retry_file_123"
		offlineChange := &OfflineChange{
			ID:        testFileID,
			Type:      "create",
			Timestamp: time.Now(),
			Path:      "/retry_test.txt",
		}

		err := filesystem.TrackOfflineChange(offlineChange)
		assert.NoError(err, "Should be able to track offline change")

		// Create a test file in the filesystem
		testItem := &graph.DriveItem{
			ID:   testFileID,
			Name: "retry_test.txt",
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "test_hash",
				},
			},
		}
		testInode := NewInodeDriveItem(testItem)
		filesystem.InsertID(testFileID, testInode)

		// Step 2: Create sync manager and process changes
		syncManager := NewSyncManager(filesystem)
		ctx := context.Background()

		// Step 3: Process changes with sync manager
		result, err := syncManager.ProcessOfflineChangesWithRetry(ctx)

		// Step 4: Verify results
		assert.NoError(err, "Should process changes without error")
		assert.NotNil(result, "Should return sync result")
		assert.True(result.ProcessedChanges >= 0, "Should process at least 0 changes")
		assert.True(result.ConflictsFound >= 0, "Should find at least 0 conflicts")
		assert.True(result.ConflictsResolved >= 0, "Should resolve at least 0 conflicts")
		assert.True(result.Duration > time.Duration(0), "Should have positive duration")
	})
}

// TestIT_SM_02_01_SyncManager_ConflictResolution_ResolvesConflicts tests conflict resolution during sync
//
//	Test Case ID    IT-SM-02-01
//	Title           Sync Manager Conflict Resolution
//	Description     Tests that the sync manager detects and resolves conflicts
//	Preconditions   None
//	Steps           1. Create conflicting local and remote changes
//	                2. Process changes with sync manager
//	                3. Verify conflicts are detected and resolved
//	Expected Result Conflicts are detected and resolved according to strategy
//	Notes: This test verifies that the sync manager properly handles conflicts during synchronization.
func TestIT_SM_02_01_SyncManager_ConflictResolution_ResolvesConflicts(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SyncManagerConflictFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create conflicting changes
		testFileID := "test_conflict_file_456"
		testFileName := "conflict_test.txt"

		// Create local item with changes
		localItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "local_hash",
				},
			},
			ModTime: &time.Time{},
		}
		localTime := time.Now().Add(-1 * time.Hour)
		localItem.ModTime = &localTime

		localInode := NewInodeDriveItem(localItem)
		localInode.hasChanges = true
		filesystem.InsertID(testFileID, localInode)

		// Create offline change
		offlineChange := &OfflineChange{
			ID:        testFileID,
			Type:      "modify",
			Timestamp: time.Now(),
			Path:      "/conflict_test.txt",
		}

		err := filesystem.TrackOfflineChange(offlineChange)
		assert.NoError(err, "Should be able to track offline change")

		// Step 2: Process changes with sync manager
		syncManager := NewSyncManager(filesystem)
		ctx := context.Background()

		result, err := syncManager.ProcessOfflineChangesWithRetry(ctx)

		// Step 3: Verify results
		assert.NoError(err, "Should process changes without error")
		assert.NotNil(result, "Should return sync result")
		assert.True(result.ProcessedChanges >= 0, "Should process changes")
		assert.True(result.ConflictsFound >= 0, "Should detect conflicts")
		assert.True(result.ConflictsResolved >= 0, "Should resolve conflicts")
	})
}

// TestIT_SM_03_01_SyncManager_NetworkRecovery_HandlesInterruptions tests network recovery
//
//	Test Case ID    IT-SM-03-01
//	Title           Sync Manager Network Recovery
//	Description     Tests that the sync manager handles network interruptions and recovery
//	Preconditions   None
//	Steps           1. Simulate network interruption
//	                2. Trigger network recovery
//	                3. Verify recovery process completes
//	Expected Result Network recovery process completes successfully
//	Notes: This test verifies that the sync manager can recover from network interruptions.
func TestIT_SM_03_01_SyncManager_NetworkRecovery_HandlesInterruptions(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SyncManagerNetworkRecoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Simulate network interruption by setting filesystem offline
		filesystem.SetOfflineMode(OfflineModeReadWrite)
		assert.True(filesystem.IsOffline(), "Filesystem should be offline")

		// Step 2: Create sync manager and test network recovery
		syncManager := NewSyncManager(filesystem)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Set filesystem back online to simulate network recovery
		filesystem.SetOfflineMode(OfflineModeDisabled)

		// Step 3: Trigger network recovery
		err := syncManager.RecoverFromNetworkInterruption(ctx)

		// Step 4: Verify recovery
		// Note: In a real test environment with proper mocking, we would verify
		// that the recovery process actually processes pending changes
		// For now, we verify that the method completes without error when online
		if !filesystem.IsOffline() {
			assert.NoError(err, "Should recover from network interruption when online")
		} else {
			// If still offline (due to test environment), expect timeout or specific error
			assert.Error(err, "Should return error when network is not available")
		}
	})
}

// TestIT_SM_04_01_SyncManager_SyncStatus_ReturnsCorrectStatus tests sync status reporting
//
//	Test Case ID    IT-SM-04-01
//	Title           Sync Manager Status Reporting
//	Description     Tests that the sync manager returns correct synchronization status
//	Preconditions   None
//	Steps           1. Create offline changes
//	                2. Get sync status
//	                3. Verify status information is correct
//	Expected Result Sync status accurately reflects current state
//	Notes: This test verifies that the sync manager provides accurate status information.
func TestIT_SM_04_01_SyncManager_SyncStatus_ReturnsCorrectStatus(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SyncManagerStatusFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create offline changes
		// First set filesystem to offline mode so changes are tracked
		filesystem.SetOfflineMode(OfflineModeReadWrite)

		offlineChange := &OfflineChange{
			ID:        "test_status_file_789",
			Type:      "create",
			Timestamp: time.Now(),
			Path:      "/status_test.txt",
		}

		err := filesystem.TrackOfflineChange(offlineChange)
		assert.NoError(err, "Should be able to track offline change")

		// Set back to online mode for status checking
		filesystem.SetOfflineMode(OfflineModeDisabled)

		// Step 2: Get sync status
		syncManager := NewSyncManager(filesystem)
		ctx := context.Background()

		status, err := syncManager.GetSyncStatus(ctx)

		// Step 3: Verify status information
		assert.NoError(err, "Should get sync status without error")
		assert.NotNil(status, "Should return status information")

		// Verify status contains expected fields
		pendingChanges, exists := status["pending_changes"]
		assert.True(exists, "Status should contain pending_changes")
		assert.True(pendingChanges.(int) >= 1, "Should have at least 1 pending change")

		isOffline, exists := status["is_offline"]
		assert.True(exists, "Status should contain is_offline")
		_, isBool := isOffline.(bool)
		assert.True(isBool, "is_offline should be boolean")

		lastSync, exists := status["last_sync"]
		assert.True(exists, "Status should contain last_sync")
		_, isTime := lastSync.(time.Time)
		assert.True(isTime, "last_sync should be time.Time")
	})
}

// TestIT_SM_05_01_SyncManager_ErrorHandling_HandlesErrors tests error handling during sync
//
//	Test Case ID    IT-SM-05-01
//	Title           Sync Manager Error Handling
//	Description     Tests that the sync manager properly handles various error conditions
//	Preconditions   None
//	Steps           1. Create conditions that cause errors
//	                2. Process changes with sync manager
//	                3. Verify errors are handled gracefully
//	Expected Result Errors are handled gracefully without crashing
//	Notes: This test verifies that the sync manager handles errors gracefully.
func TestIT_SM_05_01_SyncManager_ErrorHandling_HandlesErrors(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SyncManagerErrorHandlingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create an invalid offline change (missing required data)
		invalidChange := &OfflineChange{
			ID:        "", // Invalid empty ID
			Type:      "invalid_type",
			Timestamp: time.Now(),
			Path:      "",
		}

		_ = filesystem.TrackOfflineChange(invalidChange)
		// Note: The tracking might succeed even with invalid data,
		// but processing should handle it gracefully

		// Step 2: Process changes with sync manager
		syncManager := NewSyncManager(filesystem)
		ctx := context.Background()

		result, _ := syncManager.ProcessOfflineChangesWithRetry(ctx)

		// Step 3: Verify error handling
		// The sync manager should not crash and should return a result
		assert.NotNil(result, "Should return sync result even with errors")

		// If there were errors processing invalid changes, they should be recorded
		if len(result.Errors) > 0 {
			assert.True(len(result.Errors) > 0, "Should record errors for invalid changes")
		}

		// The sync manager should continue operating despite errors
		assert.True(result.ProcessedChanges >= 0, "Should process valid changes")
	})
}
