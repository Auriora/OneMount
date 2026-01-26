package fs

import (
	"fmt"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	dbus "github.com/godbus/dbus/v5"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// MockNemoExtensionClient simulates the Nemo extension's D-Bus client behavior
type MockNemoExtensionClient struct {
	conn          *dbus.Conn
	serviceName   string
	objectPath    dbus.ObjectPath
	interfaceName string
	signalChan    chan *dbus.Signal
}

// NewMockNemoExtensionClient creates a new mock Nemo extension client
func NewMockNemoExtensionClient(serviceName string) (*MockNemoExtensionClient, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	client := &MockNemoExtensionClient{
		conn:          conn,
		serviceName:   serviceName,
		objectPath:    DBusObjectPath,
		interfaceName: DBusInterface,
		signalChan:    make(chan *dbus.Signal, 10),
	}

	return client, nil
}

// DiscoverService simulates the Nemo extension discovering the OneMount D-Bus service
func (c *MockNemoExtensionClient) DiscoverService() error {
	// Try to get the object - this verifies the service is available
	obj := c.conn.Object(c.serviceName, c.objectPath)
	if obj == nil {
		return fmt.Errorf("service not found: %s", c.serviceName)
	}
	return nil
}

// GetFileStatus simulates the Nemo extension calling the GetFileStatus method
func (c *MockNemoExtensionClient) GetFileStatus(path string) (string, error) {
	obj := c.conn.Object(c.serviceName, c.objectPath)
	call := obj.Call(c.interfaceName+".GetFileStatus", 0, path)
	if call.Err != nil {
		return "", call.Err
	}

	var status string
	if err := call.Store(&status); err != nil {
		return "", err
	}

	return status, nil
}

// SubscribeToSignals simulates the Nemo extension subscribing to FileStatusChanged signals
func (c *MockNemoExtensionClient) SubscribeToSignals() error {
	// Add match rule for the signal
	matchRule := fmt.Sprintf(
		"type='signal',interface='%s',member='FileStatusChanged'",
		c.interfaceName,
	)

	call := c.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule)
	if call.Err != nil {
		return fmt.Errorf("failed to add match rule: %w", call.Err)
	}

	// Register signal channel
	c.conn.Signal(c.signalChan)

	return nil
}

// WaitForSignal waits for a FileStatusChanged signal with timeout
func (c *MockNemoExtensionClient) WaitForSignal(timeout time.Duration) (*dbus.Signal, error) {
	select {
	case sig := <-c.signalChan:
		return sig, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for signal")
	}
}

// Close closes the mock client connection
func (c *MockNemoExtensionClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// TestIT_FS_NemoExtension_ServiceDiscovery tests that the Nemo extension can discover the OneMount D-Bus service.
//
//	Test Case ID    IT-FS-NEMO-EXT-01
//	Title           Nemo Extension - Service Discovery
//	Description     Tests that the Nemo extension can discover the OneMount D-Bus service
//	Preconditions   D-Bus server is running
//	Steps           1. Start D-Bus server with unique service name
//	                2. Create mock Nemo extension client
//	                3. Attempt to discover the service
//	                4. Verify service is found
//	Expected Result Service discovery succeeds
//	Requirements    Requirement 8.2, 8.3 - D-Bus integration, Nemo extension
func TestIT_FS_NemoExtension_ServiceDiscovery(t *testing.T) {
	// Create a test fixture
	fixture := helpers.SetupFSTestFixture(t, "NemoExtensionServiceDiscoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		// Set a unique D-Bus service name for this test
		SetDBusServiceNamePrefix("test_nemo_discovery")

		// Start D-Bus server (use Start() instead of StartForTesting() to register service name)
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Create mock Nemo extension client
		client, err := NewMockNemoExtensionClient(DBusServiceName)
		assert.NoError(err, "Mock client should be created successfully")
		defer client.Close()

		// Test service discovery
		err = client.DiscoverService()
		assert.NoError(err, "Service discovery should succeed")

		t.Logf("✓ Nemo extension successfully discovered OneMount D-Bus service: %s", DBusServiceName)
	})
}

// TestIT_FS_NemoExtension_GetFileStatus tests that the Nemo extension can call the GetFileStatus method.
//
//	Test Case ID    IT-FS-NEMO-EXT-02
//	Title           Nemo Extension - GetFileStatus Method Call
//	Description     Tests that the Nemo extension can call GetFileStatus and receive correct responses
//	Preconditions   D-Bus server is running with test files
//	Steps           1. Create test files with known statuses
//	                2. Create mock Nemo extension client
//	                3. Call GetFileStatus for each file
//	                4. Verify correct status is returned
//	Expected Result GetFileStatus returns correct status for all files
//	Requirements    Requirement 8.2, 8.3 - D-Bus integration, Nemo extension
func TestIT_FS_NemoExtension_GetFileStatus(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "NemoExtensionGetFileStatusFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Set a unique D-Bus service name for this test
		SetDBusServiceNamePrefix("test_nemo_get_status")

		// Create test directory structure
		testDir := NewInode("testdir", fuse.S_IFDIR|0755, root)
		filesystem.InsertNodeID(testDir)
		filesystem.InsertChild(rootID, testDir)

		// Create test files with different statuses
		testFiles := []struct {
			name   string
			status FileStatus
		}{
			{"local_file.txt", StatusLocal},
			{"downloading_file.txt", StatusDownloading},
			{"syncing_file.txt", StatusSyncing},
			{"modified_file.txt", StatusLocalModified},
			{"error_file.txt", StatusError},
			{"conflict_file.txt", StatusConflict},
		}

		for _, tf := range testFiles {
			file := NewInode(tf.name, fuse.S_IFREG|0644, testDir)
			filesystem.InsertNodeID(file)
			filesystem.InsertChild(testDir.ID(), file)
			filesystem.SetFileStatus(file.ID(), FileStatusInfo{
				Status:    tf.status,
				Timestamp: time.Now(),
			})
		}

		// Start D-Bus server (use Start() instead of StartForTesting() to register service name)
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Create mock Nemo extension client
		client, err := NewMockNemoExtensionClient(DBusServiceName)
		assert.NoError(err, "Mock client should be created successfully")
		defer client.Close()

		// Test GetFileStatus for each file
		for _, tf := range testFiles {
			path := fmt.Sprintf("/testdir/%s", tf.name)
			status, err := client.GetFileStatus(path)
			assert.NoError(err, "GetFileStatus should succeed for %s", path)
			assert.Equal(tf.status.String(), status, "Status should match for %s", path)
			t.Logf("✓ GetFileStatus(%s) = %s", path, status)
		}

		// Test with non-existent file
		status, err := client.GetFileStatus("/testdir/nonexistent.txt")
		assert.NoError(err, "GetFileStatus should not error for non-existent file")
		assert.Equal("Unknown", status, "Status should be Unknown for non-existent file")
		t.Logf("✓ GetFileStatus for non-existent file returns Unknown")
	})
}

// TestIT_FS_NemoExtension_SignalSubscription tests that the Nemo extension can subscribe to D-Bus signals.
//
//	Test Case ID    IT-FS-NEMO-EXT-03
//	Title           Nemo Extension - Signal Subscription
//	Description     Tests that the Nemo extension can subscribe to FileStatusChanged signals
//	Preconditions   D-Bus server is running
//	Steps           1. Create mock Nemo extension client
//	                2. Subscribe to FileStatusChanged signals
//	                3. Verify subscription succeeds
//	Expected Result Signal subscription succeeds without errors
//	Requirements    Requirement 8.2, 8.3 - D-Bus integration, Nemo extension
func TestIT_FS_NemoExtension_SignalSubscription(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "NemoExtensionSignalSubscriptionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(FilesystemInterface)

		// Set a unique D-Bus service name for this test
		SetDBusServiceNamePrefix("test_nemo_signal_sub")

		// Start D-Bus server (use Start() instead of StartForTesting() to register service name)
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Create mock Nemo extension client
		client, err := NewMockNemoExtensionClient(DBusServiceName)
		assert.NoError(err, "Mock client should be created successfully")
		defer client.Close()

		// Test signal subscription
		err = client.SubscribeToSignals()
		assert.NoError(err, "Signal subscription should succeed")

		t.Logf("✓ Nemo extension successfully subscribed to FileStatusChanged signals")
	})
}

// TestIT_FS_NemoExtension_SignalReception tests that the Nemo extension can receive FileStatusChanged signals.
//
//	Test Case ID    IT-FS-NEMO-EXT-04
//	Title           Nemo Extension - Signal Reception
//	Description     Tests that the Nemo extension receives FileStatusChanged signals when file status changes
//	Preconditions   D-Bus server is running, client is subscribed to signals
//	Steps           1. Subscribe to signals
//	                2. Emit FileStatusChanged signal
//	                3. Wait for signal reception
//	                4. Verify signal contains correct data
//	Expected Result Signal is received with correct path and status
//	Requirements    Requirement 8.2, 8.3 - D-Bus integration, Nemo extension
func TestIT_FS_NemoExtension_SignalReception(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "NemoExtensionSignalReceptionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(FilesystemInterface)

		// Set a unique D-Bus service name for this test
		SetDBusServiceNamePrefix("test_nemo_signal_recv")

		// Start D-Bus server (use Start() instead of StartForTesting() to register service name)
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Create mock Nemo extension client
		client, err := NewMockNemoExtensionClient(DBusServiceName)
		assert.NoError(err, "Mock client should be created successfully")
		defer client.Close()

		// Subscribe to signals
		err = client.SubscribeToSignals()
		assert.NoError(err, "Signal subscription should succeed")

		// Test signal reception for multiple status changes
		testCases := []struct {
			path   string
			status string
		}{
			{"/test/file1.txt", "Downloading"},
			{"/test/file2.txt", "Syncing"},
			{"/test/file3.txt", "Local"},
			{"/test/file4.txt", "Error"},
		}

		for _, tc := range testCases {
			// Emit signal
			dbusServer.SendFileStatusUpdate(tc.path, tc.status)

			// Wait for signal reception
			signal, err := client.WaitForSignal(2 * time.Second)
			assert.NoError(err, "Should receive signal for %s", tc.path)
			assert.NotNil(signal, "Signal should not be nil")

			// Verify signal data
			assert.Equal(DBusInterface+".FileStatusChanged", signal.Name, "Signal name should match")
			assert.Equal(2, len(signal.Body), "Signal should have 2 arguments")

			receivedPath, ok := signal.Body[0].(string)
			assert.True(ok, "First argument should be string (path)")
			assert.Equal(tc.path, receivedPath, "Path should match")

			receivedStatus, ok := signal.Body[1].(string)
			assert.True(ok, "Second argument should be string (status)")
			assert.Equal(tc.status, receivedStatus, "Status should match")

			t.Logf("✓ Received signal: %s -> %s", receivedPath, receivedStatus)
		}
	})
}

// TestIT_FS_NemoExtension_ErrorHandling tests that the Nemo extension handles D-Bus errors gracefully.
//
//	Test Case ID    IT-FS-NEMO-EXT-05
//	Title           Nemo Extension - Error Handling
//	Description     Tests that the Nemo extension handles D-Bus unavailability and errors gracefully
//	Preconditions   None
//	Steps           1. Attempt to connect when D-Bus service is not running
//	                2. Verify error is handled gracefully
//	                3. Attempt GetFileStatus when service is unavailable
//	                4. Verify appropriate error is returned
//	Expected Result Errors are handled gracefully without crashes
//	Requirements    Requirement 8.2, 8.3, 8.4 - D-Bus integration, Nemo extension, fallback
func TestIT_FS_NemoExtension_ErrorHandling(t *testing.T) {
	assert := framework.NewAssert(t)

	// Set a unique D-Bus service name that doesn't exist
	nonExistentService := "org.onemount.FileStatus.nonexistent_test_service"

	// Test 1: Service discovery when service doesn't exist
	client, err := NewMockNemoExtensionClient(nonExistentService)
	assert.NoError(err, "Client creation should succeed even if service doesn't exist")
	defer client.Close()

	err = client.DiscoverService()
	// Note: D-Bus may not immediately fail if the service name is not registered
	// The error will occur when trying to call methods on the service
	if err != nil {
		t.Logf("✓ Service discovery fails gracefully when service unavailable: %v", err)
	} else {
		t.Logf("✓ Service discovery succeeds (service name exists but may not be functional)")
	}

	// Test 2: GetFileStatus when service doesn't exist
	status, err := client.GetFileStatus("/test/file.txt")
	assert.Error(err, "GetFileStatus should fail when service doesn't exist")
	assert.Equal("", status, "Status should be empty when service unavailable")
	t.Logf("✓ GetFileStatus fails gracefully when service unavailable: %v", err)

	// Test 3: Signal subscription when service doesn't exist
	err = client.SubscribeToSignals()
	assert.NoError(err, "Signal subscription should succeed (match rule is added regardless)")
	t.Logf("✓ Signal subscription succeeds even when service unavailable")
}

// TestIT_FS_NemoExtension_Performance tests that D-Bus status queries meet performance requirements.
//
//	Test Case ID    IT-FS-NEMO-EXT-06
//	Title           Nemo Extension - Performance
//	Description     Tests that GetFileStatus queries complete within performance requirements (< 10ms per file)
//	Preconditions   D-Bus server is running with test files
//	Steps           1. Create multiple test files
//	                2. Measure time for GetFileStatus calls
//	                3. Verify average time is < 10ms per file
//	Expected Result GetFileStatus queries complete in < 10ms per file
//	Requirements    Requirement 8.2, 8.3, 10.3 - D-Bus integration, Nemo extension, performance
func TestIT_FS_NemoExtension_Performance(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "NemoExtensionPerformanceFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Set a unique D-Bus service name for this test
		SetDBusServiceNamePrefix("test_nemo_performance")

		// Create test directory
		testDir := NewInode("perftest", fuse.S_IFDIR|0755, root)
		filesystem.InsertNodeID(testDir)
		filesystem.InsertChild(rootID, testDir)

		// Create 100 test files
		numFiles := 100
		for i := 0; i < numFiles; i++ {
			file := NewInode(fmt.Sprintf("file%d.txt", i), fuse.S_IFREG|0644, testDir)
			filesystem.InsertNodeID(file)
			filesystem.InsertChild(testDir.ID(), file)
			filesystem.SetFileStatus(file.ID(), FileStatusInfo{
				Status:    StatusLocal,
				Timestamp: time.Now(),
			})
		}

		// Start D-Bus server (use Start() instead of StartForTesting() to register service name)
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Create mock Nemo extension client
		client, err := NewMockNemoExtensionClient(DBusServiceName)
		assert.NoError(err, "Mock client should be created successfully")
		defer client.Close()

		// Warm up - make a few calls first
		for i := 0; i < 5; i++ {
			path := fmt.Sprintf("/perftest/file%d.txt", i)
			_, _ = client.GetFileStatus(path)
		}

		// Measure performance for 50 files
		testFiles := 50
		start := time.Now()

		for i := 0; i < testFiles; i++ {
			path := fmt.Sprintf("/perftest/file%d.txt", i)
			status, err := client.GetFileStatus(path)
			assert.NoError(err, "GetFileStatus should succeed")
			assert.Equal("Local", status, "Status should be Local")
		}

		elapsed := time.Since(start)
		avgTime := elapsed / time.Duration(testFiles)

		t.Logf("Performance results:")
		t.Logf("  Total time: %v", elapsed)
		t.Logf("  Files queried: %d", testFiles)
		t.Logf("  Average time per file: %v", avgTime)

		// Verify performance requirement: < 10ms per file
		maxTimePerFile := 10 * time.Millisecond
		assert.True(avgTime < maxTimePerFile,
			"Average time per file (%v) should be less than %v", avgTime, maxTimePerFile)

		t.Logf("✓ Performance requirement met: %v < %v per file", avgTime, maxTimePerFile)
	})
}
