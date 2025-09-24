package fs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
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
func TestDBusServer_GetFileStatus(t *testing.T) {
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
		filesystem := fsFixture.FS.(FilesystemInterface)

		// Set a unique D-Bus service name prefix for this test
		SetDBusServiceNamePrefix("test_get_file_status")

		// Create and start D-Bus server
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.StartForTesting()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Test GetFileStatus with various paths
		testPaths := []string{
			"/test/path/file.txt",
			"/another/path/document.pdf",
			"/root/file.doc",
			"",
			"/path/with/unicode/файл.txt",
		}

		for _, path := range testPaths {
			status, dbusErr := dbusServer.GetFileStatus(path)
			assert.Nil(dbusErr, "GetFileStatus should not return D-Bus error for path: %s", path)
			// Currently the implementation returns "Unknown" for all paths
			// since GetPath is not available in FilesystemInterface
			assert.Equal("Unknown", status, "GetFileStatus should return Unknown for path: %s", path)
		}
	})
}

// TestDBusServer_SendFileStatusUpdate tests the SendFileStatusUpdate signal emission.
func TestDBusServer_SendFileStatusUpdate(t *testing.T) {
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
func TestDBusServiceNameGeneration(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)

	// Test default service name generation
	originalServiceName := DBusServiceName
	defer func() {
		DBusServiceName = originalServiceName
	}()

	// Test with custom prefix
	SetDBusServiceNamePrefix("test_prefix")
	assert.Contains(DBusServiceName, "test_prefix", "Service name should contain custom prefix")
	assert.Contains(DBusServiceName, DBusServiceNameBase, "Service name should contain base name")

	// Test with empty prefix
	SetDBusServiceNamePrefix("")
	assert.Contains(DBusServiceName, "instance", "Service name should contain default prefix when empty")

	// Test uniqueness - generate multiple service names
	names := make(map[string]bool)
	for i := 0; i < 10; i++ {
		SetDBusServiceNamePrefix("unique_test")
		names[DBusServiceName] = true
	}

	// All names should be unique due to timestamp/PID components
	assert.Equal(10, len(names), "All generated service names should be unique")

	// Test service name format
	SetDBusServiceNamePrefix("format_test")
	assert.True(len(DBusServiceName) > len(DBusServiceNameBase), "Generated name should be longer than base")
	assert.True(strings.HasPrefix(DBusServiceName, DBusServiceNameBase), "Generated name should start with base")
}

// TestDBusServer_MultipleInstances tests running multiple D-Bus server instances.
func TestDBusServer_MultipleInstances(t *testing.T) {
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
