package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	dbus "github.com/godbus/dbus/v5"
)

// TestIT_FS_DBus_ServiceDiscovery verifies OneMount service is discoverable on D-Bus session bus.
//
//	Test Case ID    IT-FS-DBUS-DISCOVERY-01
//	Title           D-Bus Service Discovery
//	Description     Tests that OneMount service can be discovered on D-Bus session bus
//	Preconditions   D-Bus session bus is available
//	Steps           1. Create and start a D-Bus server
//	                2. Connect to D-Bus session bus
//	                3. List all services using org.freedesktop.DBus.ListNames
//	                4. Verify OneMount service is in the list
//	                5. Verify service is reachable using Peer.Ping
//	Expected Result OneMount service is discoverable and reachable
//	Notes: This test automates the service discovery functionality previously tested manually with D-Feet.
//	       It verifies that external clients (like Nemo extension) can discover the OneMount service.
func TestIT_FS_DBus_ServiceDiscovery(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusServiceDiscoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_service_discovery")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Connect to D-Bus session bus
		conn, err := dbus.SessionBus()
		assert.NoError(err, "Should connect to D-Bus session bus")

		// List all services using org.freedesktop.DBus.ListNames
		var names []string
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.ListNames", 0,
		).Store(&names)
		assert.NoError(err, "Should list D-Bus services")
		assert.True(len(names) > 0, "Should have at least one service on the bus")

		// Verify OneMount service is in the list
		serviceName := DBusServiceName
		found := false
		for _, name := range names {
			if name == serviceName {
				found = true
				break
			}
		}
		assert.True(found, "OneMount service '%s' should be discoverable on D-Bus (found %d services)", serviceName, len(names))

		// Verify service is reachable using Peer.Ping
		obj := conn.Object(serviceName, DBusObjectPath)
		err = obj.Call("org.freedesktop.DBus.Peer.Ping", 0).Err
		assert.NoError(err, "Should be able to ping OneMount service at %s", DBusObjectPath)

		// Additional verification: Verify we can actually call a method on the service
		var status string
		err = obj.Call(DBusInterface+".GetFileStatus", 0, "/test/path").Store(&status)
		assert.NoError(err, "Should be able to call GetFileStatus method on discovered service")
		assert.NotEqual("", status, "GetFileStatus should return a status value")
	})
}
