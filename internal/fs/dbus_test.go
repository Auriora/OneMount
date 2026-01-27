package fs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestIT_FS_02_01_DBusServer_BasicFunctionality_WorksCorrectly tests D-Bus server functionality.
//
//	Test Case ID    IT-FS-02-01
//	Title           D-Bus Server Functionality
//	Description     Tests D-Bus server functionality
//	Preconditions   None
//	Steps           1. Set up a D-Bus server
//	                2. Perform operations (get file status, emit signals, reconnect)
//	                3. Verify the results of each operation
//	Expected Result D-Bus server functionality works correctly
//	Notes: This test verifies that the D-Bus server functionality works correctly.
func TestIT_FS_02_01_DBusServer_BasicFunctionality_WorksCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusServerFunctionalityFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		filesystem := fsFixture.FS.(FilesystemInterface)

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_basic_functionality")

		// Create a D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		assert.NotNil(dbusServer, "D-Bus server should be created")

		// Start the D-Bus server in test mode
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")

		// Verify server is started
		assert.True(dbusServer.started, "D-Bus server should be marked as started")
		assert.NotNil(dbusServer.conn, "D-Bus connection should be established")

		// Test GetFileStatus method
		status, dbusErr := dbusServer.GetFileStatus("/test/path/file.txt")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error")
		assert.Equal("Unknown", status, "GetFileStatus should return Unknown for non-existent files")

		// Test SendFileStatusUpdate signal emission
		// This should not panic or error
		dbusServer.SendFileStatusUpdate("/test/path/file.txt", "Local")

		// Stop the D-Bus server
		dbusServer.Stop()

		// Verify server is stopped
		assert.False(dbusServer.started, "D-Bus server should be marked as stopped")
		assert.Nil(dbusServer.conn, "D-Bus connection should be nil after stop")
	})
}

// TestIT_FS_11_01_DBusServer_StartStop_OperatesCorrectly tests D-Bus server start/stop operations.
//
//	Test Case ID    IT-FS-11-01
//	Title           D-Bus Server Start/Stop Operations
//	Description     Tests D-Bus server start/stop operations
//	Preconditions   None
//	Steps           1. Create a D-Bus server
//	                2. Perform start/stop operations
//	                3. Verify the server state after each operation
//	Expected Result D-Bus server start/stop operations work correctly
//	Notes: This test verifies that the D-Bus server start/stop operations work correctly.
func TestIT_FS_11_01_DBusServer_StartStop_OperatesCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusServerStartStopFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		filesystem := fsFixture.FS.(FilesystemInterface)

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_start_stop")

		// Create a D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		assert.NotNil(dbusServer, "D-Bus server should be created")

		// Initial state - server should not be started
		assert.False(dbusServer.started, "D-Bus server should not be started initially")
		assert.Nil(dbusServer.conn, "D-Bus connection should be nil initially")

		// Start the server in test mode
		err := dbusServer.StartForTesting()
		assert.NoError(err, "First start should succeed")
		assert.True(dbusServer.started, "D-Bus server should be started after StartForTesting")
		assert.NotNil(dbusServer.conn, "D-Bus connection should be established after start")

		// Try to start again - should be idempotent
		err = dbusServer.StartForTesting()
		assert.NoError(err, "Second start should be idempotent")
		assert.True(dbusServer.started, "D-Bus server should still be started")

		// Stop the server
		dbusServer.Stop()
		assert.False(dbusServer.started, "D-Bus server should be stopped after Stop")
		assert.Nil(dbusServer.conn, "D-Bus connection should be nil after stop")

		// Try to stop again - should be idempotent
		dbusServer.Stop()
		assert.False(dbusServer.started, "D-Bus server should still be stopped")

		// Restart the server
		err = dbusServer.StartForTesting()
		assert.NoError(err, "Restart should succeed")
		assert.True(dbusServer.started, "D-Bus server should be started after restart")
		assert.NotNil(dbusServer.conn, "D-Bus connection should be re-established")

		// Final cleanup
		dbusServer.Stop()
	})
}

// TestDBusServer_GetFileStatus tests the GetFileStatus D-Bus method.
func TestIT_FS_DBus_GetFileStatus(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusGetFileStatusFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_get_file_status")

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test GetFileStatus with non-existent paths - should return Unknown
		testPaths := []string{
			"/test/path/file.txt",
			"/another/path/document.pdf",
			"/nonexistent/file.doc",
		}

		for _, path := range testPaths {
			status, dbusErr := dbusServer.GetFileStatus(path)
			assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for path: %s", path)
			assert.Equal("Unknown", status, "GetFileStatus should return Unknown for non-existent path: %s", path)
		}

		// Test with empty path - should return Unknown or root status
		status, dbusErr := dbusServer.GetFileStatus("")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for empty path")
		// Empty path maps to root, which should have a status
		assert.NotEqual("", status, "GetFileStatus should return a status for empty path")

		// Test with root path
		status, dbusErr = dbusServer.GetFileStatus("/")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for root path")
		assert.NotEqual("", status, "GetFileStatus should return a status for root path")
	})
}

// TestDBusServer_GetFileStatus_WithRealFiles tests GetFileStatus with actual files in the filesystem.
func TestIT_FS_DBus_GetFileStatus_WithRealFiles(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusGetFileStatusRealFilesFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_get_file_status_real")

		// Create some test files in the filesystem
		// Create a directory
		testDir := NewInode("testdir", fuse.S_IFDIR|0755, root)
		filesystem.InsertNodeID(testDir)
		filesystem.InsertChild(rootID, testDir)

		// Create a file in the directory
		testFile := NewInode("testfile.txt", fuse.S_IFREG|0644, testDir)
		filesystem.InsertNodeID(testFile)
		filesystem.InsertChild(testDir.ID(), testFile)

		// Set file status to a known value
		filesystem.SetFileStatus(testFile.ID(), FileStatusInfo{
			Status: StatusLocal,
		})

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test GetFileStatus for the directory
		status, dbusErr := dbusServer.GetFileStatus("/testdir")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for directory")
		assert.NotEqual("Unknown", status, "GetFileStatus should return actual status for existing directory")

		// Test GetFileStatus for the file
		status, dbusErr = dbusServer.GetFileStatus("/testdir/testfile.txt")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for file")
		assert.Equal("Local", status, "GetFileStatus should return Local status for the test file")

		// Test with non-existent file in existing directory
		status, dbusErr = dbusServer.GetFileStatus("/testdir/nonexistent.txt")
		assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for non-existent file")
		assert.Equal("Unknown", status, "GetFileStatus should return Unknown for non-existent file")
	})
}

// TestDBusServer_SendFileStatusUpdate tests the SendFileStatusUpdate signal emission.
func TestIT_FS_DBus_SendFileStatusUpdate(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSendSignalFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		filesystem := fsFixture.FS.(FilesystemInterface)

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_send_signal")

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test signal emission with various status updates
		testCases := []struct {
			path   string
			status string
		}{
			{"/test/path/file1.txt", "Local"},
			{"/test/path/file2.txt", "Cloud"},
			{"/test/path/file3.txt", "Syncing"},
			{"/test/path/file4.txt", "Error"},
			{"/test/path/file5.txt", "Downloading"},
			{"", "Unknown"},
			{"/path/with/unicode/файл.txt", "LocalModified"},
		}

		for _, tc := range testCases {
			// This should not panic or return an error
			// Note: We can't easily test signal reception in unit tests without
			// setting up a full D-Bus client, but we can verify the method doesn't crash
			dbusServer.SendFileStatusUpdate(tc.path, tc.status)
		}

		// Test signal emission when server is not started
		dbusServer.Stop()
		// This should be safe and not panic
		dbusServer.SendFileStatusUpdate("/test/path/stopped.txt", "Unknown")
	})
}

// TestDBusServiceNameGeneration tests D-Bus service name generation and uniqueness.
func TestIT_FS_DBus_ServiceNameGeneration(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)

	// Test default service name generation
	originalServiceName := DBusServiceName
	defer func() {
		DBusServiceName = originalServiceName
	}()

	// Test with custom prefix
	SetDBusServiceNamePrefix("test_prefix")
	assert.Equal(DBusServiceNameBase+".test_prefix", DBusServiceName, "Service name should include custom prefix deterministically")

	// Test with empty prefix
	SetDBusServiceNamePrefix("")
	assert.Equal(DBusServiceNameBase+".instance", DBusServiceName, "Empty prefix should fall back to default")

	// Repeated calls with same prefix should keep identical name
	SetDBusServiceNamePrefix("stable")
	first := DBusServiceName
	SetDBusServiceNamePrefix("stable")
	assert.Equal(first, DBusServiceName, "Deterministic naming should produce identical values for same prefix")

	// Test service name format
	SetDBusServiceNamePrefix("format_test")
	assert.True(len(DBusServiceName) > len(DBusServiceNameBase), "Generated name should be longer than base")
	assert.True(strings.HasPrefix(DBusServiceName, DBusServiceNameBase), "Generated name should start with base")
}

func TestIT_FS_DBus_SetServiceNameForMount(t *testing.T) {
	assert := framework.NewAssert(t)
	original := DBusServiceName
	defer func() { DBusServiceName = original }()

	SetDBusServiceNameForMount("/home/bcherrington/OneMountTest")
	assert.Equal("org.onemount.FileStatus.mnt_home_bcherrington_OneMountTest", DBusServiceName)

	SetDBusServiceNameForMount("/tmp/onemount auth")
	assert.Equal("org.onemount.FileStatus.mnt_tmp_onemount_x20auth", DBusServiceName)
}

// TestDBusServer_MultipleInstances tests running multiple D-Bus server instances.
func TestIT_FS_DBus_MultipleInstances(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusMultipleInstancesFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		filesystem := fsFixture.FS.(FilesystemInterface)

		// Create multiple D-Bus servers with unique names
		servers := make([]*FileStatusDBusServer, 3)
		for i := 0; i < 3; i++ {
			SetDBusServiceNamePrefix(fmt.Sprintf("test_multi_%d", i))
			servers[i] = NewFileStatusDBusServer(filesystem)

			err := servers[i].StartForTesting()
			assert.NoError(err, "Server %d should start successfully", i)
			assert.True(servers[i].started, "Server %d should be marked as started", i)
		}

		// All servers should be running independently
		for i, server := range servers {
			assert.True(server.started, "Server %d should still be running", i)
			assert.NotNil(server.conn, "Server %d should have active connection", i)

			// Test that each server can handle requests
			status, dbusErr := server.GetFileStatus(fmt.Sprintf("/test/path/server_%d.txt", i))
			assert.Nil(dbusErr, "Server %d should handle GetFileStatus", i)
			assert.Equal("Unknown", status, "Server %d should return Unknown status", i)
		}

		// Stop all servers
		for i, server := range servers {
			server.Stop()
			assert.False(server.started, "Server %d should be stopped", i)
			assert.Nil(server.conn, "Server %d connection should be nil", i)
		}
	})
}

// TestFindInodeByPath_PathTraversal tests the path traversal logic without D-Bus.
func TestIT_FS_DBus_FindInodeByPath_PathTraversal(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "PathTraversalFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a directory structure:
		// /
		// ├── dir1/
		// │   ├── file1.txt
		// │   └── subdir/
		// │       └── file2.txt
		// └── dir2/
		//     └── file3.txt

		// Get the actual root inode
		root := filesystem.GetID(filesystem.root)
		assert.NotNil(root, "Root inode should exist")

		// Create dir1
		dir1 := NewInode("dir1", fuse.S_IFDIR|0755, root)
		filesystem.InsertID(dir1.ID(), dir1)
		filesystem.InsertNodeID(dir1)

		// Create file1.txt in dir1
		file1 := NewInode("file1.txt", fuse.S_IFREG|0644, dir1)
		filesystem.InsertID(file1.ID(), file1)
		filesystem.InsertNodeID(file1)

		// Create subdir in dir1
		subdir := NewInode("subdir", fuse.S_IFDIR|0755, dir1)
		filesystem.InsertID(subdir.ID(), subdir)
		filesystem.InsertNodeID(subdir)

		// Create file2.txt in subdir
		file2 := NewInode("file2.txt", fuse.S_IFREG|0644, subdir)
		filesystem.InsertID(file2.ID(), file2)
		filesystem.InsertNodeID(file2)

		// Create dir2
		dir2 := NewInode("dir2", fuse.S_IFDIR|0755, root)
		filesystem.InsertID(dir2.ID(), dir2)
		filesystem.InsertNodeID(dir2)

		// Create file3.txt in dir2
		file3 := NewInode("file3.txt", fuse.S_IFREG|0644, dir2)
		filesystem.InsertID(file3.ID(), file3)
		filesystem.InsertNodeID(file3)

		// Set up the directory structure
		root.SetChildren([]string{dir1.ID(), dir2.ID()})
		dir1.SetChildren([]string{file1.ID(), subdir.ID()})
		subdir.SetChildren([]string{file2.ID()})
		dir2.SetChildren([]string{file3.ID()})

		// Set file statuses
		filesystem.SetFileStatus(file1.ID(), FileStatusInfo{Status: StatusLocal})
		filesystem.SetFileStatus(file2.ID(), FileStatusInfo{Status: StatusCloud})
		filesystem.SetFileStatus(file3.ID(), FileStatusInfo{Status: StatusLocalModified})

		// Create D-Bus server (without starting it to avoid D-Bus dependency)
		dbusServer := NewFileStatusDBusServer(filesystem)

		// Test path traversal using GetIDByPath
		testCases := []struct {
			path          string
			expectedID    string
			shouldBeEmpty bool
			description   string
		}{
			{"/", filesystem.root, false, "Root path"},
			{"", filesystem.root, false, "Empty path (maps to root)"},
			{"/dir1", dir1.ID(), false, "First level directory"},
			{"/dir1/file1.txt", file1.ID(), false, "File in first level directory"},
			{"/dir1/subdir", subdir.ID(), false, "Second level directory"},
			{"/dir1/subdir/file2.txt", file2.ID(), false, "File in second level directory"},
			{"/dir2", dir2.ID(), false, "Another first level directory"},
			{"/dir2/file3.txt", file3.ID(), false, "File in another directory"},
			{"/nonexistent", "", true, "Non-existent path"},
			{"/dir1/nonexistent.txt", "", true, "Non-existent file in existing directory"},
			{"/dir1/subdir/nonexistent", "", true, "Non-existent path in deep directory"},
		}

		for _, tc := range testCases {
			id := filesystem.GetIDByPath(tc.path)
			if tc.shouldBeEmpty {
				assert.Equal("", id, "Expected empty ID for path: %s (%s)", tc.path, tc.description)
			} else {
				assert.NotEqual("", id, "Expected non-empty ID for path: %s (%s)", tc.path, tc.description)
				assert.Equal(tc.expectedID, id, "ID mismatch for path: %s (%s)", tc.path, tc.description)
			}
		}

		// Test GetFileStatus method (without D-Bus connection)
		// This will use the findInodeByPath logic we just tested
		testStatusCases := []struct {
			path           string
			expectedStatus string
			description    string
		}{
			{"/dir1/file1.txt", "Local", "File with Local status"},
			{"/dir1/subdir/file2.txt", "Cloud", "File with Cloud status"},
			{"/dir2/file3.txt", "LocalModified", "File with LocalModified status"},
			{"/nonexistent", "Unknown", "Non-existent file"},
		}

		for _, tc := range testStatusCases {
			status, dbusErr := dbusServer.GetFileStatus(tc.path)
			assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for path: %s", tc.path)
			assert.Equal(tc.expectedStatus, status, "Status mismatch for path: %s (%s)", tc.path, tc.description)
		}
	})
}
