package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDBusServerStartStop tests starting and stopping the D-Bus server
func TestDBusServerStartStop(t *testing.T) {
	// Create a temporary filesystem for testing
	tempDir := filepath.Join(testDBLoc, "test_dbus_start_stop")
	os.RemoveAll(tempDir)
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Create a new filesystem
	fs, err := NewFilesystem(auth, tempDir, 30)
	require.NoError(t, err, "Failed to create filesystem")

	// The D-Bus server should be started automatically
	assert.NotNil(t, fs.dbusServer, "D-Bus server should be initialized")
	assert.True(t, fs.dbusServer.started, "D-Bus server should be started")

	// Stop the D-Bus server
	fs.dbusServer.Stop()
	assert.False(t, fs.dbusServer.started, "D-Bus server should be stopped")

	// Start the D-Bus server again
	err = fs.dbusServer.StartForTesting()
	assert.NoError(t, err, "Failed to start D-Bus server")
	assert.True(t, fs.dbusServer.started, "D-Bus server should be started")

	// Stop the D-Bus server again
	fs.dbusServer.Stop()
	assert.False(t, fs.dbusServer.started, "D-Bus server should be stopped")
}

// TestDBusGetFileStatus tests the GetFileStatus method
func TestDBusGetFileStatus(t *testing.T) {
	// Create a test file
	testFilePath := filepath.Join(TestDir, "dbus_test_file.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0644)
	require.NoError(t, err, "Failed to create test file")
	defer os.Remove(testFilePath)

	// Wait for the file to be recognized by the filesystem
	var inode *Inode
	require.Eventually(t, func() bool {
		inode, err = fs.GetPath(testFilePath, auth)
		return err == nil && inode != nil
	}, 10*time.Second, 100*time.Millisecond, "Failed to get inode for test file")

	// Call the GetFileStatus method directly
	status := fs.GetFileStatus(inode.ID())
	statusStr := status.Status.String()

	// The status should be "LocalModified" since we just created the file with os.WriteFile
	assert.Equal(t, "LocalModified", statusStr, "File status should be 'LocalModified'")
}

// TestDBusFileStatusSignal tests the FileStatusChanged signal
func TestDBusFileStatusSignal(t *testing.T) {
	// Create a test file
	testFilePath := filepath.Join(TestDir, "dbus_test_signal.txt")
	err := os.WriteFile(testFilePath, []byte("test content"), 0644)
	require.NoError(t, err, "Failed to create test file")
	defer os.Remove(testFilePath)

	// Wait for the file to be recognized by the filesystem
	var inode *Inode
	require.Eventually(t, func() bool {
		inode, err = fs.GetPath(testFilePath, auth)
		return err == nil && inode != nil
	}, 10*time.Second, 100*time.Millisecond, "Failed to get inode for test file")

	// Connect to the D-Bus service
	conn, err := dbus.SessionBus()
	require.NoError(t, err, "Failed to connect to D-Bus session bus")
	defer conn.Close()

	// Set up a signal handler
	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)

	// Add a match rule for the FileStatusChanged signal
	err = conn.AddMatchSignal(
		dbus.WithMatchInterface(DBusInterface),
		dbus.WithMatchMember("FileStatusChanged"),
		dbus.WithMatchObjectPath(DBusObjectPath),
	)
	require.NoError(t, err, "Failed to add match rule for signal")

	// Trigger a file status update
	fs.updateFileStatus(inode)

	// Wait for the signal
	var signal *dbus.Signal
	timeout := time.After(5 * time.Second)
	for {
		select {
		case signal = <-signalChan:
			// Check if this is the signal we're looking for
			if signal.Path == DBusObjectPath && signal.Name == DBusInterface+".FileStatusChanged" {
				// Got the signal we're looking for
				goto signalFound
			}
			// Not the signal we're looking for, continue waiting
			t.Logf("Received signal: %v, continuing to wait", signal)
		case <-timeout:
			t.Fatal("Timed out waiting for FileStatusChanged signal")
		}
	}

signalFound:
	// Verify the signal
	// Convert dbus.ObjectPath to string for comparison
	assert.Equal(t, string(DBusObjectPath), string(signal.Path), "Signal path should match")
	assert.Equal(t, DBusInterface+".FileStatusChanged", signal.Name, "Signal name should match")
	assert.Len(t, signal.Body, 2, "Signal should have 2 arguments")
	// The path in the signal is the OneDrive API path, not the local filesystem path
	// Just check that it contains the filename
	assert.Contains(t, signal.Body[0].(string), "dbus_test_signal.txt", "Signal path should contain the test file name")
	assert.Equal(t, "LocalModified", signal.Body[1].(string), "Signal status should be 'LocalModified'")
}

// TestDBusServerReconnect tests that the D-Bus server can reconnect after being stopped
func TestDBusServerReconnect(t *testing.T) {
	// Create a temporary filesystem for testing
	tempDir := filepath.Join(testDBLoc, "test_dbus_reconnect")
	os.RemoveAll(tempDir)
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// Create a new filesystem
	testFS, err := NewFilesystem(auth, tempDir, 30)
	require.NoError(t, err, "Failed to create filesystem")

	// The D-Bus server should be started automatically
	assert.NotNil(t, testFS.dbusServer, "D-Bus server should be initialized")
	assert.True(t, testFS.dbusServer.started, "D-Bus server should be started")

	// Stop the D-Bus server
	testFS.dbusServer.Stop()
	assert.False(t, testFS.dbusServer.started, "D-Bus server should be stopped")

	// Create a test file
	testFilePath := filepath.Join(TestDir, "dbus_test_reconnect.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	require.NoError(t, err, "Failed to create test file")
	defer os.Remove(testFilePath)

	// Wait for the file to be recognized by the global filesystem
	// We use the global fs variable here because the test file is created in the mounted filesystem
	var inode *Inode
	require.Eventually(t, func() bool {
		inode, err = fs.GetPath(testFilePath, auth)
		return err == nil && inode != nil
	}, 10*time.Second, 100*time.Millisecond, "Failed to get inode for test file")

	// Updating file status should not cause an error even though the D-Bus server is stopped
	// We use the test filesystem's updateFileStatus method to test that it doesn't cause an error
	testFS.updateFileStatus(inode)

	// Start the D-Bus server again
	err = testFS.dbusServer.StartForTesting()
	assert.NoError(t, err, "Failed to start D-Bus server")
	assert.True(t, testFS.dbusServer.started, "D-Bus server should be started")

	// Connect to the D-Bus service
	conn, err := dbus.SessionBus()
	require.NoError(t, err, "Failed to connect to D-Bus session bus")
	defer conn.Close()

	// Get the D-Bus object
	obj := conn.Object(DBusServiceName, DBusObjectPath)
	require.NotNil(t, obj, "Failed to get D-Bus object")

	// Call the GetFileStatus method on the test filesystem's D-Bus server
	var status string
	err = obj.Call(DBusInterface+".GetFileStatus", 0, testFilePath).Store(&status)
	require.NoError(t, err, "Failed to call GetFileStatus method")

	// The status should be "LocalModified" since we just created the file with os.WriteFile
	assert.Equal(t, "LocalModified", status, "File status should be 'LocalModified'")
}
