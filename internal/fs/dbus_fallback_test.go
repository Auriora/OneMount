package fs

import (
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_STATUS_08_DBusFallback_SystemContinuesOperating tests system operation without D-Bus.
//
//	Test Case ID    IT-FS-STATUS-08
//	Title           D-Bus Fallback - System Continues Operating
//	Description     Tests that the system continues operating when D-Bus is unavailable
//	Preconditions   Filesystem mounted without D-Bus server
//	Steps           1. Create filesystem without starting D-Bus server
//	                2. Perform file status operations
//	                3. Verify operations complete successfully
//	Expected Result System operates normally without D-Bus
//	Notes: This test verifies graceful degradation when D-Bus is unavailable.
func TestIT_FS_STATUS_08_DBusFallback_SystemContinuesOperating(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusFallbackFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem WITHOUT starting D-Bus server
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		// Note: We intentionally do NOT start the D-Bus server here
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

		// Note: D-Bus server may be started automatically by the filesystem
		// This test verifies that operations work regardless of D-Bus state
		if filesystem.dbusServer != nil {
			t.Logf("D-Bus server is running (automatic start)")
			// Stop it to test fallback
			filesystem.dbusServer.Stop()
			filesystem.dbusServer = nil
			t.Logf("D-Bus server stopped to test fallback")
		} else {
			t.Logf("D-Bus server is not running")
		}

		// Test 1: File status operations should work without D-Bus
		testID := "fallback-test-file"

		// Set status
		filesystem.SetFileStatus(testID, FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		})

		// Get status
		status := filesystem.GetFileStatus(testID)
		assert.Equal(StatusLocal, status.Status, "Status should be retrievable without D-Bus")

		// Test 2: Mark file operations should work
		filesystem.MarkFileDownloading(testID)
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusDownloading, status.Status, "MarkFileDownloading should work without D-Bus")

		filesystem.MarkFileOutofSync(testID)
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusOutofSync, status.Status, "MarkFileOutofSync should work without D-Bus")

		filesystem.MarkFileError(testID, os.ErrPermission)
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusError, status.Status, "MarkFileError should work without D-Bus")

		filesystem.MarkFileConflict(testID, "Test conflict")
		status = filesystem.GetFileStatus(testID)
		assert.Equal(StatusConflict, status.Status, "MarkFileConflict should work without D-Bus")

		// Test 3: Multiple file status operations
		for i := 0; i < 10; i++ {
			testFileID := "fallback-file-" + string(rune(i+'0'))
			filesystem.SetFileStatus(testFileID, FileStatusInfo{
				Status:    StatusLocal,
				Timestamp: time.Now(),
			})
			status := filesystem.GetFileStatus(testFileID)
			assert.Equal(StatusLocal, status.Status, "Status operations should work for multiple files")
		}

		t.Logf("✓ System operates normally without D-Bus")
	})
}

// TestIT_FS_STATUS_10_DBusFallback_NoDBusPanics tests that missing D-Bus doesn't cause panics.
//
//	Test Case ID    IT-FS-STATUS-10
//	Title           D-Bus Fallback - No Panics
//	Description     Tests that operations don't panic when D-Bus is unavailable
//	Preconditions   Filesystem mounted without D-Bus server
//	Steps           1. Create filesystem without D-Bus server
//	                2. Perform various operations that would use D-Bus
//	                3. Verify no panics occur
//	Expected Result No panics occur when D-Bus is unavailable
//	Notes: This test verifies robustness of D-Bus fallback.
func TestIT_FS_STATUS_10_DBusFallback_NoDBusPanics(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusNoPanicsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem WITHOUT starting D-Bus server
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

		// Note: D-Bus server may be started automatically
		// Stop it to test fallback behavior
		if filesystem.dbusServer != nil {
			filesystem.dbusServer.Stop()
			filesystem.dbusServer = nil
		}

		// Test that these operations don't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Operation panicked when D-Bus unavailable: %v", r)
			}
		}()

		// Create test ID
		testID := "test-no-panic"

		// These should all be safe even without D-Bus
		filesystem.SetFileStatus(testID, FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		})

		filesystem.GetFileStatus(testID)
		filesystem.MarkFileDownloading(testID)
		filesystem.MarkFileOutofSync(testID)
		filesystem.MarkFileError(testID, os.ErrNotExist)
		filesystem.MarkFileConflict(testID, "conflict")

		// If we get here without panicking, the test passes
		assert.True(true, "All operations completed without panicking")
		t.Logf("✓ No panics occur when D-Bus is unavailable")
	})
}
