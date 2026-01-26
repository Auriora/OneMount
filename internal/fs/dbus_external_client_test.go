package fs

import (
	"fmt"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	dbus "github.com/godbus/dbus/v5"
)

// TestIT_FS_DBus_ExternalClientSimulation simulates an external D-Bus client (like Nemo extension).
//
//	Test Case ID    IT-FS-DBUS-EXTERNAL-CLIENT-01
//	Title           D-Bus External Client Simulation
//	Description     Tests that OneMount works correctly from external client perspective
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Discover OneMount service on D-Bus
//	                2. Connect to the service
//	                3. Subscribe to FileStatusChanged signals
//	                4. Call GetFileStatus method
//	                5. Perform file operation (trigger status change)
//	                6. Verify signal received by external client
//	Expected Result External client can discover, connect, subscribe, call methods, and receive signals
//	Requirements    8.2, 8.3, 10.2
//	Notes: This test simulates the behavior of an external D-Bus client like the Nemo file manager
//	       extension. It automates the external client simulation functionality previously tested
//	       manually with D-Feet. The test verifies that OneMount works correctly from the perspective
//	       of an external application that needs to:
//	       - Discover the OneMount service on D-Bus
//	       - Connect to the service
//	       - Subscribe to status change signals
//	       - Query file status via GetFileStatus method
//	       - Receive and process status change signals
//	       This is exactly what the Nemo extension does when displaying file status icons.
func TestIT_FS_DBus_ExternalClientSimulation(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusExternalClientFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_external_client")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		t.Log("=== Simulating External D-Bus Client (like Nemo extension) ===")

		// Step 1: Discover service (like Nemo would on startup)
		t.Log("Step 1: Discovering OneMount service on D-Bus...")
		conn, err := dbus.SessionBus()
		assert.NoError(err, "External client should connect to D-Bus session bus")

		// List all services to discover OneMount
		var names []string
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.ListNames", 0,
		).Store(&names)
		assert.NoError(err, "External client should be able to list D-Bus services")

		serviceName := DBusServiceName
		found := false
		for _, name := range names {
			if name == serviceName {
				found = true
				break
			}
		}
		assert.True(found, "External client should discover OneMount service '%s'", serviceName)
		t.Logf("✓ Service discovered: %s", serviceName)

		// Step 2: Connect to service (like Nemo would)
		t.Log("Step 2: Connecting to OneMount service...")
		obj := conn.Object(serviceName, DBusObjectPath)

		// Verify service is reachable
		err = obj.Call("org.freedesktop.DBus.Peer.Ping", 0).Err
		assert.NoError(err, "External client should be able to ping OneMount service")
		t.Logf("✓ Connected to service at %s", DBusObjectPath)

		// Step 3: Subscribe to signals (like Nemo would to update icons)
		t.Log("Step 3: Subscribing to FileStatusChanged signals...")
		signalChan := make(chan *dbus.Signal, 10)
		conn.Signal(signalChan)
		defer conn.RemoveSignal(signalChan)

		// Add match rule for FileStatusChanged signals
		matchRule := fmt.Sprintf(
			"type='signal',sender='%s',interface='%s',member='FileStatusChanged'",
			serviceName,
			DBusInterface,
		)
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.AddMatch", 0, matchRule,
		).Err
		assert.NoError(err, "External client should be able to subscribe to signals")
		t.Logf("✓ Subscribed to signals with match rule: %s", matchRule)

		// Step 4: Call GetFileStatus method (like Nemo would for each file)
		t.Log("Step 4: Calling GetFileStatus method...")
		testPath := "/test-file.txt"

		var status string
		err = obj.Call(
			DBusInterface+".GetFileStatus", 0,
			testPath,
		).Store(&status)
		assert.NoError(err, "External client should be able to call GetFileStatus method")
		assert.NotEqual("", status, "GetFileStatus should return a status value")
		t.Logf("✓ GetFileStatus('%s') returned: %s", testPath, status)

		// Step 5: Perform file operation (simulate file status change)
		t.Log("Step 5: Triggering file status change...")
		// Simulate a file operation that would trigger a status change
		// In real scenario, this would be a file access, modification, etc.
		dbusServer.SendFileStatusUpdate(testPath, "Downloading")
		t.Logf("✓ Triggered status change: %s → Downloading", testPath)

		// Step 6: Verify signal received (like Nemo would update icon)
		t.Log("Step 6: Waiting for FileStatusChanged signal...")
		select {
		case sig := <-signalChan:
			// Verify signal details
			// Note: sig.Name includes the full interface name (e.g., "org.onemount.FileStatus.FileStatusChanged")
			assert.True(
				sig.Name == "FileStatusChanged" || sig.Name == DBusInterface+".FileStatusChanged",
				"Signal name should be FileStatusChanged (got: %s)", sig.Name,
			)
			assert.Equal(2, len(sig.Body), "Signal should have 2 parameters (path, status)")

			// Extract path and status from signal
			signalPath, pathOk := sig.Body[0].(string)
			signalStatus, statusOk := sig.Body[1].(string)

			assert.True(pathOk, "Signal path should be a string")
			assert.True(statusOk, "Signal status should be a string")
			assert.Equal(testPath, signalPath, "Signal path should match test path")
			assert.Equal("Downloading", signalStatus, "Signal status should be 'Downloading'")

			t.Logf("✓ Signal received: FileStatusChanged('%s', '%s')", signalPath, signalStatus)
			t.Log("✓ External client successfully received and processed signal")

		case <-time.After(5 * time.Second):
			t.Fatal("External client did not receive signal within timeout")
		}

		// Additional verification: Call GetFileStatus again to verify updated status
		t.Log("Step 7: Verifying updated status via GetFileStatus...")
		var updatedStatus string
		err = obj.Call(
			DBusInterface+".GetFileStatus", 0,
			testPath,
		).Store(&updatedStatus)
		assert.NoError(err, "External client should be able to call GetFileStatus after status change")
		t.Logf("✓ GetFileStatus('%s') now returns: %s", testPath, updatedStatus)

		// Simulate another status change (download complete)
		t.Log("Step 8: Simulating download completion...")
		dbusServer.SendFileStatusUpdate(testPath, "Cached")

		// Verify second signal
		select {
		case sig := <-signalChan:
			signalPath, _ := sig.Body[0].(string)
			signalStatus, _ := sig.Body[1].(string)

			assert.Equal(testPath, signalPath, "Second signal path should match")
			assert.Equal("Cached", signalStatus, "Second signal status should be 'Cached'")

			t.Logf("✓ Second signal received: FileStatusChanged('%s', '%s')", signalPath, signalStatus)

		case <-time.After(5 * time.Second):
			t.Fatal("External client did not receive second signal within timeout")
		}

		t.Log("=== External Client Simulation Complete ===")
		t.Log("✓ All external client operations successful:")
		t.Log("  - Service discovery")
		t.Log("  - Service connection")
		t.Log("  - Signal subscription")
		t.Log("  - Method invocation (GetFileStatus)")
		t.Log("  - Signal reception and processing")
		t.Log("  - Multiple signal handling")
		t.Log("")
		t.Log("This test verifies that OneMount works correctly from the perspective")
		t.Log("of external D-Bus clients like the Nemo file manager extension.")
	})
}
