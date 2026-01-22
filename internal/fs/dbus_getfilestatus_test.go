package fs

import (
	"fmt"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestDBusServer_GetFileStatus_ValidPaths tests GetFileStatus with valid file paths
// This test verifies that the GetFileStatus method correctly returns file status
// for valid paths in the filesystem.
//
// Test ID: TestDBusServer_GetFileStatus_ValidPaths
// Requirements: 8.2
// Expected Result: GetFileStatus returns correct status for valid paths
// Notes: This test verifies the fix for Issue #FS-001
func TestDBusServer_GetFileStatus_ValidPaths(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusGetFileStatusValidPathsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_valid_paths")

		// Create a directory structure
		// /Documents
		docsDir := NewInode("Documents", fuse.S_IFDIR|0755, root)
		filesystem.InsertNodeID(docsDir)
		filesystem.InsertChild(rootID, docsDir)

		// /Documents/Work
		workDir := NewInode("Work", fuse.S_IFDIR|0755, docsDir)
		filesystem.InsertNodeID(workDir)
		filesystem.InsertChild(docsDir.ID(), workDir)

		// /Documents/Work/report.pdf
		reportFile := NewInode("report.pdf", fuse.S_IFREG|0644, workDir)
		filesystem.InsertNodeID(reportFile)
		filesystem.InsertChild(workDir.ID(), reportFile)

		// Set different statuses for testing
		filesystem.SetFileStatus(docsDir.ID(), FileStatusInfo{Status: StatusLocal})
		filesystem.SetFileStatus(workDir.ID(), FileStatusInfo{Status: StatusLocal})
		filesystem.SetFileStatus(reportFile.ID(), FileStatusInfo{Status: StatusDownloading})

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test GetFileStatus for root
		status, dbusErr := dbusServer.GetFileStatus("/")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for root")
		assert.NotEqual("Unknown", status, "GetFileStatus should return actual status for root")

		// Test GetFileStatus for /Documents
		status, dbusErr = dbusServer.GetFileStatus("/Documents")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for /Documents")
		assert.Equal("Local", status, "GetFileStatus should return Local status for /Documents")

		// Test GetFileStatus for /Documents/Work
		status, dbusErr = dbusServer.GetFileStatus("/Documents/Work")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for /Documents/Work")
		assert.Equal("Local", status, "GetFileStatus should return Local status for /Documents/Work")

		// Test GetFileStatus for /Documents/Work/report.pdf
		status, dbusErr = dbusServer.GetFileStatus("/Documents/Work/report.pdf")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for /Documents/Work/report.pdf")
		assert.Equal("Downloading", status, "GetFileStatus should return Downloading status for /Documents/Work/report.pdf")
	})
}

// TestDBusServer_GetFileStatus_InvalidPaths tests GetFileStatus with invalid file paths
// This test verifies that the GetFileStatus method correctly handles invalid paths
// and returns "Unknown" status.
//
// Test ID: TestDBusServer_GetFileStatus_InvalidPaths
// Requirements: 8.2
// Expected Result: GetFileStatus returns Unknown for invalid paths
// Notes: This test verifies the fix for Issue #FS-001
func TestDBusServer_GetFileStatus_InvalidPaths(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusGetFileStatusInvalidPathsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_invalid_paths")

		// Create a simple directory structure
		docsDir := NewInode("Documents", fuse.S_IFDIR|0755, root)
		filesystem.InsertNodeID(docsDir)
		filesystem.InsertChild(rootID, docsDir)

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test GetFileStatus with non-existent paths
		testCases := []struct {
			path        string
			description string
		}{
			{"/NonExistent", "non-existent directory in root"},
			{"/Documents/NonExistent", "non-existent file in existing directory"},
			{"/Documents/SubDir/file.txt", "file in non-existent subdirectory"},
			{"/a/b/c/d/e/f/g", "deeply nested non-existent path"},
		}

		for _, tc := range testCases {
			status, dbusErr := dbusServer.GetFileStatus(tc.path)
			assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for %s", tc.description)
			assert.Equal("Unknown", status, "GetFileStatus should return Unknown for %s", tc.description)
		}
	})
}

// TestDBusServer_GetFileStatus_StatusChanges tests GetFileStatus with changing file statuses
// This test verifies that the GetFileStatus method correctly reflects status changes
// as files are modified, downloaded, or synced.
//
// Test ID: TestDBusServer_GetFileStatus_StatusChanges
// Requirements: 8.2
// Expected Result: GetFileStatus returns updated status after status changes
// Notes: This test verifies the fix for Issue #FS-001
func TestDBusServer_GetFileStatus_StatusChanges(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusGetFileStatusChangesFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_status_changes")

		// Create a test file
		testFile := NewInode("test.txt", fuse.S_IFREG|0644, root)
		filesystem.InsertNodeID(testFile)
		filesystem.InsertChild(rootID, testFile)

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test status progression: Cloud -> Downloading -> Local -> LocalModified -> Syncing -> Local
		statusProgression := []struct {
			status      FileStatus
			description string
		}{
			{StatusCloud, "Cloud"},
			{StatusDownloading, "Downloading"},
			{StatusLocal, "Local"},
			{StatusLocalModified, "LocalModified"},
			{StatusSyncing, "Syncing"},
			{StatusLocal, "Local"},
		}

		for _, sp := range statusProgression {
			// Set the status
			filesystem.SetFileStatus(testFile.ID(), FileStatusInfo{Status: sp.status})

			// Query the status via D-Bus
			status, dbusErr := dbusServer.GetFileStatus("/test.txt")
			assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error")
			assert.Equal(sp.description, status, "GetFileStatus should return %s status", sp.description)
		}

		// Test error status
		filesystem.MarkFileError(testFile.ID(), fmt.Errorf("test error"))
		status, dbusErr := dbusServer.GetFileStatus("/test.txt")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error")
		assert.Equal("Error", status, "GetFileStatus should return Error status")

		// Test conflict status
		filesystem.MarkFileConflict(testFile.ID(), "test conflict")
		status, dbusErr = dbusServer.GetFileStatus("/test.txt")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error")
		assert.Equal("Conflict", status, "GetFileStatus should return Conflict status")
	})
}

// TestDBusServer_GetFileStatus_SpecialCharacters tests GetFileStatus with special characters in paths
// This test verifies that the GetFileStatus method correctly handles paths with special characters.
//
// Test ID: TestDBusServer_GetFileStatus_SpecialCharacters
// Requirements: 8.2
// Expected Result: GetFileStatus handles special characters correctly
// Notes: This test verifies the fix for Issue #FS-001
func TestDBusServer_GetFileStatus_SpecialCharacters(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusGetFileStatusSpecialCharsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_special_chars")

		// Create files with special characters in names
		testFiles := []struct {
			name        string
			description string
		}{
			{"file with spaces.txt", "spaces"},
			{"file-with-dashes.txt", "dashes"},
			{"file_with_underscores.txt", "underscores"},
			{"file.multiple.dots.txt", "multiple dots"},
		}

		for _, tf := range testFiles {
			file := NewInode(tf.name, fuse.S_IFREG|0644, root)
			filesystem.InsertNodeID(file)
			filesystem.InsertChild(rootID, file)
			filesystem.SetFileStatus(file.ID(), FileStatusInfo{Status: StatusLocal})
		}

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test GetFileStatus for files with special characters
		for _, tf := range testFiles {
			path := "/" + tf.name
			status, dbusErr := dbusServer.GetFileStatus(path)
			assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for file with %s", tf.description)
			assert.Equal("Local", status, "GetFileStatus should return Local status for file with %s", tf.description)
		}
	})
}
