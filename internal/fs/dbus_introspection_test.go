package fs

import (
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	dbus "github.com/godbus/dbus/v5"
)

// TestIT_FS_DBus_IntrospectionValidation verifies D-Bus interface structure is correct.
//
//	Test Case ID    IT-FS-DBUS-INTROSPECTION-01
//	Title           D-Bus Interface Introspection Validation
//	Description     Tests that D-Bus interface structure matches expected specification
//	Preconditions   D-Bus session bus is available
//	Steps           1. Create and start a D-Bus server
//	                2. Connect to D-Bus session bus
//	                3. Call org.freedesktop.DBus.Introspectable.Introspect
//	                4. Verify org.onemount.FileStatus interface is present
//	                5. Verify GetFileStatus method signature is correct
//	                6. Verify FileStatusChanged signal signature is correct
//	                7. Verify standard D-Bus interfaces are present
//	Expected Result Interface structure matches specification
//	Notes: This test automates the introspection functionality previously tested manually with D-Feet.
//	       It verifies that the D-Bus interface contract is correct and external clients can
//	       discover the interface structure programmatically.
func TestIT_FS_DBus_IntrospectionValidation(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusIntrospectionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_introspection")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Connect to D-Bus session bus
		conn, err := dbus.SessionBus()
		assert.NoError(err, "Should connect to D-Bus session bus")

		// Get D-Bus object
		serviceName := DBusServiceName
		obj := conn.Object(serviceName, DBusObjectPath)

		// Call Introspect method
		var introspectXML string
		err = obj.Call(
			"org.freedesktop.DBus.Introspectable.Introspect", 0,
		).Store(&introspectXML)
		assert.NoError(err, "Should be able to introspect interface")
		assert.True(len(introspectXML) > 0, "Introspection XML should not be empty")

		// Verify org.onemount.FileStatus interface is present
		assert.True(
			strings.Contains(introspectXML, `interface name="org.onemount.FileStatus"`),
			"Should expose org.onemount.FileStatus interface",
		)

		// Verify GetFileStatus method is present with correct signature
		assert.True(
			strings.Contains(introspectXML, `<method name="GetFileStatus">`),
			"Should expose GetFileStatus method",
		)
		assert.True(
			strings.Contains(introspectXML, `<arg name="path" type="s" direction="in">`),
			"GetFileStatus should have 'path' input parameter of type string",
		)
		assert.True(
			strings.Contains(introspectXML, `<arg name="status" type="s" direction="out">`),
			"GetFileStatus should return 'status' output parameter of type string",
		)

		// Verify FileStatusChanged signal is present with correct signature
		assert.True(
			strings.Contains(introspectXML, `<signal name="FileStatusChanged">`),
			"Should expose FileStatusChanged signal",
		)
		// Note: Signal arguments don't have direction attribute
		assert.True(
			strings.Contains(introspectXML, `<arg name="path" type="s">`),
			"FileStatusChanged signal should have 'path' parameter of type string",
		)
		assert.True(
			strings.Contains(introspectXML, `<arg name="status" type="s">`),
			"FileStatusChanged signal should have 'status' parameter of type string",
		)

		// Verify standard D-Bus interfaces are present
		assert.True(
			strings.Contains(introspectXML, `interface name="org.freedesktop.DBus.Introspectable"`),
			"Should support org.freedesktop.DBus.Introspectable interface",
		)
		// Note: The godbus library may or may not expose Peer/Properties interfaces
		// depending on the implementation. We just verify Introspectable is present.

		// Log the introspection XML for debugging (truncated for readability)
		if len(introspectXML) > 1000 {
			t.Logf("Introspection XML (first 1000 chars): %s...", introspectXML[:1000])
		} else {
			t.Logf("Introspection XML: %s", introspectXML)
		}
	})
}
