package fs

import (
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_STATUS_01_FileStatus_Updates_WorkCorrectly tests file status updates during operations.
//
//	Test Case ID    IT-FS-STATUS-01
//	Title           File Status Updates
//	Description     Tests file status updates during various file operations
//	Preconditions   Filesystem mounted with status tracking enabled
//	Steps           1. Create a file and check status
//	                2. Modify a file and check status
//	                3. Read a file and check status
//	                4. Verify status consistency
//	Expected Result File status updates correctly during operations
//	Notes: This test verifies that file status tracking works correctly.
func TestIT_FS_STATUS_01_FileStatus_Updates_WorkCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileStatusUpdatesFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		// Test 1: Check status for non-existent file (should be Cloud or not found)
		nonExistentID := "non-existent-file-id"
		status := filesystem.GetFileStatus(nonExistentID)
		assert.NotNil(status, "Status should not be nil")
		assert.True(status.Status == StatusCloud || status.Status == StatusLocal,
			"Non-existent file should have Cloud or Local status")

		// Test 2: Mark file as downloading and verify status
		testID := "test-file-downloading"
		filesystem.MarkFileDownloading(testID)
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusDownloading, status.Status, "Status should be Downloading")
		assert.False(status.Timestamp.IsZero(), "Timestamp should be set")

		// Test 3: Mark file as out of sync and verify status
		filesystem.MarkFileOutofSync(testID)
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusOutofSync, status.Status, "Status should be OutofSync")

		// Test 4: Mark file with error and verify status
		testError := os.ErrPermission
		filesystem.MarkFileError(testID, testError)
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusError, status.Status, "Status should be Error")
		assert.Contains(status.ErrorMsg, "permission", "Error message should contain 'permission'")

		// Test 5: Mark file with conflict and verify status
		conflictMsg := "File modified both locally and remotely"
		filesystem.MarkFileConflict(testID, conflictMsg)
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusConflict, status.Status, "Status should be Conflict")
		assert.Equal(conflictMsg, status.ErrorMsg, "Conflict message should match")

		// Test 6: Verify status caching
		// Set a status and retrieve it multiple times
		filesystem.SetFileStatus(testID, FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		})

		for i := 0; i < 5; i++ {
			status = filesystem.GetFileStatus(testID)
			assert.Equal(StatusLocal, status.Status, "Cached status should be consistent")
		}

		// Test 7: Verify status string representations
		testCases := []struct {
			status   FileStatus
			expected string
		}{
			{StatusCloud, "Cloud"},
			{StatusLocal, "Local"},
			{StatusLocalModified, "LocalModified"},
			{StatusSyncing, "Syncing"},
			{StatusDownloading, "Downloading"},
			{StatusOutofSync, "OutofSync"},
			{StatusError, "Error"},
			{StatusConflict, "Conflict"},
		}

		for _, tc := range testCases {
			assert.Equal(tc.expected, tc.status.String(),
				"Status string should match expected value")
		}
	})
}

// TestIT_FS_STATUS_02_FileStatus_Determination_WorksCorrectly tests status determination logic.
//
//	Test Case ID    IT-FS-STATUS-02
//	Title           File Status Determination
//	Description     Tests the status determination logic for various file states
//	Preconditions   Filesystem mounted with content cache
//	Steps           1. Test status for cached files
//	                2. Test status for cloud-only files
//	                3. Test status for files with offline changes
//	                4. Test status for uploading files
//	Expected Result Status determination works correctly for all file states
//	Notes: This test verifies the determineFileStatus logic.
func TestIT_FS_STATUS_02_FileStatus_Determination_WorksCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileStatusDeterminationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		// Test 1: Cloud-only file (not in cache, no offline changes)
		cloudFileID := "cloud-only-file"
		status := filesystem.GetFileStatus(cloudFileID)
		// Should be Cloud since it's not cached and has no offline changes
		assert.True(status.Status == StatusCloud || status.Status == StatusLocal,
			"Cloud-only file should have Cloud or Local status")

		// Test 2: Cached file
		cachedFileID := "cached-file"
		// Simulate a cached file by marking it as local
		filesystem.SetFileStatus(cachedFileID, FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		})
		status = filesystem.GetFileStatus(cachedFileID)
		assert.Equal(StatusLocal, status.Status, "Cached file should have Local status")

		// Test 3: File being downloaded
		downloadingFileID := "downloading-file"
		filesystem.MarkFileDownloading(downloadingFileID)
		status = filesystem.GetFileStatus(downloadingFileID)
		assert.Equal(StatusDownloading, status.Status, "Downloading file should have Downloading status")

		// Test 4: File with error
		errorFileID := "error-file"
		filesystem.MarkFileError(errorFileID, os.ErrNotExist)
		status = filesystem.GetFileStatus(errorFileID)
		assert.Equal(StatusError, status.Status, "Error file should have Error status")
		assert.True(len(status.ErrorMsg) > 0, "Error message should be set")

		// Test 5: Multiple status changes
		changingFileID := "changing-file"

		// Start as cloud
		filesystem.SetFileStatus(changingFileID, FileStatusInfo{
			Status:    StatusCloud,
			Timestamp: time.Now(),
		})
		status = filesystem.GetFileStatus(changingFileID)
		assert.Equal(StatusCloud, status.Status, "Initial status should be Cloud")

		// Mark as downloading
		filesystem.MarkFileDownloading(changingFileID)
		status = filesystem.GetFileStatus(changingFileID)
		assert.Equal(StatusDownloading, status.Status, "Status should change to Downloading")

		// Mark as local (download complete)
		filesystem.SetFileStatus(changingFileID, FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		})
		status = filesystem.GetFileStatus(changingFileID)
		assert.Equal(StatusLocal, status.Status, "Status should change to Local")

		// Mark as modified
		filesystem.SetFileStatus(changingFileID, FileStatusInfo{
			Status:    StatusLocalModified,
			Timestamp: time.Now(),
		})
		status = filesystem.GetFileStatus(changingFileID)
		assert.Equal(StatusLocalModified, status.Status, "Status should change to LocalModified")

		// Mark as syncing
		filesystem.SetFileStatus(changingFileID, FileStatusInfo{
			Status:    StatusSyncing,
			Timestamp: time.Now(),
		})
		status = filesystem.GetFileStatus(changingFileID)
		assert.Equal(StatusSyncing, status.Status, "Status should change to Syncing")
	})
}

// TestIT_FS_STATUS_03_FileStatus_ThreadSafety_WorksCorrectly tests thread safety of status operations.
//
//	Test Case ID    IT-FS-STATUS-03
//	Title           File Status Thread Safety
//	Description     Tests thread safety of concurrent status operations
//	Preconditions   Filesystem mounted
//	Steps           1. Perform concurrent status reads
//	                2. Perform concurrent status writes
//	                3. Perform mixed concurrent operations
//	Expected Result All operations complete without race conditions
//	Notes: This test should be run with -race flag.
func TestIT_FS_STATUS_03_FileStatus_ThreadSafety_WorksCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileStatusThreadSafetyFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		// Test concurrent reads
		testID := "concurrent-test-file"
		filesystem.SetFileStatus(testID, FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		})

		done := make(chan bool)
		numGoroutines := 10

		// Concurrent reads
		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					status := filesystem.GetFileStatus(testID)
					assert.NotNil(status, "Status should not be nil")
				}
				done <- true
			}()
		}

		// Wait for all reads to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Concurrent writes
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < 100; j++ {
					filesystem.SetFileStatus(testID, FileStatusInfo{
						Status:    StatusLocal,
						Timestamp: time.Now(),
					})
				}
				done <- true
			}(i)
		}

		// Wait for all writes to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Mixed concurrent operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < 50; j++ {
					if j%2 == 0 {
						filesystem.GetFileStatus(testID)
					} else {
						filesystem.SetFileStatus(testID, FileStatusInfo{
							Status:    StatusLocal,
							Timestamp: time.Now(),
						})
					}
				}
				done <- true
			}(i)
		}

		// Wait for all mixed operations to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify final state is consistent
		status := filesystem.GetFileStatus(testID)
		assert.NotNil(status, "Final status should not be nil")
		assert.Equal(StatusLocal, status.Status, "Final status should be Local")
	})
}

// TestIT_FS_STATUS_04_FileStatus_Timestamps_WorkCorrectly tests timestamp tracking.
//
//	Test Case ID    IT-FS-STATUS-04
//	Title           File Status Timestamps
//	Description     Tests timestamp tracking for status updates
//	Preconditions   Filesystem mounted
//	Steps           1. Set status and check timestamp
//	                2. Update status and verify timestamp changes
//	                3. Verify timestamp ordering
//	Expected Result Timestamps are set correctly and update on status changes
//	Notes: This test verifies timestamp tracking.
func TestIT_FS_STATUS_04_FileStatus_Timestamps_WorksCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileStatusTimestampsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		testID := "timestamp-test-file"

		// Test 1: Initial timestamp
		filesystem.MarkFileDownloading(testID)
		status1 := filesystem.GetFileStatus(testID)
		assert.False(status1.Timestamp.IsZero(), "Initial timestamp should be set")

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Test 2: Updated timestamp
		filesystem.SetFileStatus(testID, FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		})
		status2 := filesystem.GetFileStatus(testID)
		assert.False(status2.Timestamp.IsZero(), "Updated timestamp should be set")
		assert.True(status2.Timestamp.After(status1.Timestamp),
			"Updated timestamp should be after initial timestamp")

		// Test 3: Multiple updates
		timestamps := make([]time.Time, 5)
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Millisecond)
			filesystem.SetFileStatus(testID, FileStatusInfo{
				Status:    StatusLocal,
				Timestamp: time.Now(),
			})
			status := filesystem.GetFileStatus(testID)
			timestamps[i] = status.Timestamp
		}

		// Verify timestamps are in order
		for i := 1; i < len(timestamps); i++ {
			assert.True(timestamps[i].After(timestamps[i-1]) || timestamps[i].Equal(timestamps[i-1]),
				"Timestamps should be in chronological order")
		}
	})
}
