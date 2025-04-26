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

// TestDBusServerOperations tests various operations on the D-Bus server
func TestDBusServerOperations(t *testing.T) {
	t.Parallel()

	// Create a temporary filesystem for testing
	tempDir := filepath.Join(testDBLoc, "test_dbus_operations_"+t.Name())
	err := os.RemoveAll(tempDir)
	require.NoError(t, err, "Failed to remove temp directory")

	err = os.MkdirAll(tempDir, 0755)
	require.NoError(t, err, "Failed to create temp directory")

	// Setup cleanup to remove the temp directory after test completes or fails
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to clean up temp directory %s: %v", tempDir, err)
		}
	})

	// Create a new filesystem
	fs, err := NewFilesystem(auth, tempDir, 30)
	require.NoError(t, err, "Failed to create filesystem")

	// Define test cases
	testCases := []struct {
		name           string
		operation      func(t *testing.T, fs *Filesystem) error
		expectedState  bool
		errorExpected  bool
	}{
		{
			name: "InitialState_ShouldBeStarted",
			operation: func(t *testing.T, fs *Filesystem) error {
				// No operation, just check initial state
				return nil
			},
			expectedState: true,
			errorExpected: false,
		},
		{
			name: "StopServer_ShouldBeStopped",
			operation: func(t *testing.T, fs *Filesystem) error {
				fs.dbusServer.Stop()
				return nil
			},
			expectedState: false,
			errorExpected: false,
		},
		{
			name: "StartServer_ShouldBeStarted",
			operation: func(t *testing.T, fs *Filesystem) error {
				return fs.dbusServer.StartForTesting()
			},
			expectedState: true,
			errorExpected: false,
		},
		{
			name: "StopAgain_ShouldBeStopped",
			operation: func(t *testing.T, fs *Filesystem) error {
				fs.dbusServer.Stop()
				return nil
			},
			expectedState: false,
			errorExpected: false,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// We don't use t.Parallel() here because we need to run these tests in sequence
			// to properly test the start/stop functionality on the same server instance

			// Perform the operation
			err := tc.operation(t, fs)

			// Check if error behavior matches expectations
			if tc.errorExpected {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Got unexpected error")
			}

			// Verify the server state
			require.NotNil(t, fs.dbusServer, "D-Bus server should be initialized")
			require.Equal(t, tc.expectedState, fs.dbusServer.started, 
				"D-Bus server state does not match expected state. Expected: %v, Got: %v", 
				tc.expectedState, fs.dbusServer.started)
		})
	}
}

// TestDBusServerFunctionality tests various functionality of the D-Bus server
func TestDBusServerFunctionality(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name        string
		description string
		setup       func(t *testing.T) (*Filesystem, *Inode, string, func())
		test        func(t *testing.T, testFS *Filesystem, inode *Inode, filePath string)
	}{
		{
			name:        "GetFileStatus_ShouldReturnCorrectStatus",
			description: "Tests that the GetFileStatus method returns the correct status for a file",
			setup: func(t *testing.T) (*Filesystem, *Inode, string, func()) {
				// Create a test file
				testFilePath := filepath.Join(TestDir, "dbus_test_file.txt")
				err := os.WriteFile(testFilePath, []byte("test content"), 0644)
				require.NoError(t, err, "Failed to create test file")

				// Wait for the file to be recognized by the filesystem
				var inode *Inode
				require.Eventually(t, func() bool {
					inode, err = fs.GetPath(testFilePath, auth)
					return err == nil && inode != nil
				}, 10*time.Second, 100*time.Millisecond, "Failed to get inode for test file")

				// Return cleanup function
				cleanup := func() {
					if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
						t.Logf("Failed to remove test file during cleanup: %v", err)
					}
				}

				return fs, inode, testFilePath, cleanup
			},
			test: func(t *testing.T, testFS *Filesystem, inode *Inode, filePath string) {
				// Call the GetFileStatus method directly
				status := testFS.GetFileStatus(inode.ID())
				statusStr := status.Status.String()

				// The status should be "LocalModified" since we just created the file with os.WriteFile
				assert.Equal(t, "LocalModified", statusStr, "File status should be 'LocalModified'")
			},
		},
		{
			name:        "FileStatusSignal_ShouldEmitCorrectSignal",
			description: "Tests that the FileStatusChanged signal is emitted with the correct data",
			setup: func(t *testing.T) (*Filesystem, *Inode, string, func()) {
				// Create a test file
				testFilePath := filepath.Join(TestDir, "dbus_test_signal.txt")
				err := os.WriteFile(testFilePath, []byte("test content"), 0644)
				require.NoError(t, err, "Failed to create test file")

				// Wait for the file to be recognized by the filesystem
				var inode *Inode
				require.Eventually(t, func() bool {
					inode, err = fs.GetPath(testFilePath, auth)
					return err == nil && inode != nil
				}, 10*time.Second, 100*time.Millisecond, "Failed to get inode for test file")

				// Return cleanup function
				cleanup := func() {
					if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
						t.Logf("Failed to remove test file during cleanup: %v", err)
					}
				}

				return fs, inode, testFilePath, cleanup
			},
			test: func(t *testing.T, testFS *Filesystem, inode *Inode, filePath string) {
				// Connect to the D-Bus service
				conn, err := dbus.SessionBus()
				require.NoError(t, err, "Failed to connect to D-Bus session bus")

				// Set up a signal handler
				signalChan := make(chan *dbus.Signal, 10)
				conn.Signal(signalChan)

				// Ensure connection is properly closed
				t.Cleanup(func() {
					if err := conn.Close(); err != nil {
						t.Logf("Failed to close D-Bus connection: %v", err)
					}
				})

				// Add a match rule for the FileStatusChanged signal
				err = conn.AddMatchSignal(
					dbus.WithMatchInterface(DBusInterface),
					dbus.WithMatchMember("FileStatusChanged"),
					dbus.WithMatchObjectPath(DBusObjectPath),
				)
				require.NoError(t, err, "Failed to add match rule for signal")

				// Trigger a file status update
				testFS.updateFileStatus(inode)

				// Wait for the signal
				var signal *dbus.Signal
				select {
				case signal = <-signalChan:
					// Check if this is the signal we're looking for
					if signal.Path != DBusObjectPath || signal.Name != DBusInterface+".FileStatusChanged" {
						t.Fatalf("Received unexpected signal: %v", signal)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("Timed out waiting for FileStatusChanged signal")
				}

				// Verify the signal
				// Convert dbus.ObjectPath to string for comparison
				assert.Equal(t, string(DBusObjectPath), string(signal.Path), "Signal path should match")
				assert.Equal(t, DBusInterface+".FileStatusChanged", signal.Name, "Signal name should match")
				assert.Len(t, signal.Body, 2, "Signal should have 2 arguments")
				// The path in the signal is the OneDrive API path, not the local filesystem path
				// Just check that it contains the filename
				assert.Contains(t, signal.Body[0].(string), "dbus_test_signal.txt", "Signal path should contain the test file name")
				assert.Equal(t, "LocalModified", signal.Body[1].(string), "Signal status should be 'LocalModified'")
			},
		},
		{
			name:        "ServerReconnect_ShouldWorkAfterRestart",
			description: "Tests that the D-Bus server can reconnect after being stopped",
			setup: func(t *testing.T) (*Filesystem, *Inode, string, func()) {
				// Create a temporary filesystem for testing
				tempDir := filepath.Join(testDBLoc, "test_dbus_reconnect_"+t.Name())
				err := os.RemoveAll(tempDir)
				require.NoError(t, err, "Failed to remove temp directory")

				err = os.MkdirAll(tempDir, 0755)
				require.NoError(t, err, "Failed to create temp directory")

				// Create a new filesystem
				testFS, err := NewFilesystem(auth, tempDir, 30)
				require.NoError(t, err, "Failed to create filesystem")

				// Create a test file
				testFilePath := filepath.Join(TestDir, "dbus_test_reconnect.txt")
				err = os.WriteFile(testFilePath, []byte("test content"), 0644)
				require.NoError(t, err, "Failed to create test file")

				// Wait for the file to be recognized by the global filesystem
				var inode *Inode
				require.Eventually(t, func() bool {
					inode, err = fs.GetPath(testFilePath, auth)
					return err == nil && inode != nil
				}, 10*time.Second, 100*time.Millisecond, "Failed to get inode for test file")

				// Return cleanup function
				cleanup := func() {
					if err := os.RemoveAll(tempDir); err != nil {
						t.Logf("Failed to remove temp directory during cleanup: %v", err)
					}
					if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
						t.Logf("Failed to remove test file during cleanup: %v", err)
					}
				}

				return testFS, inode, testFilePath, cleanup
			},
			test: func(t *testing.T, testFS *Filesystem, inode *Inode, filePath string) {
				// The D-Bus server should be started automatically
				assert.NotNil(t, testFS.dbusServer, "D-Bus server should be initialized")
				assert.True(t, testFS.dbusServer.started, "D-Bus server should be started")

				// Stop the D-Bus server
				testFS.dbusServer.Stop()
				assert.False(t, testFS.dbusServer.started, "D-Bus server should be stopped")

				// Updating file status should not cause an error even though the D-Bus server is stopped
				testFS.updateFileStatus(inode)

				// Start the D-Bus server again
				err := testFS.dbusServer.StartForTesting()
				assert.NoError(t, err, "Failed to start D-Bus server")
				assert.True(t, testFS.dbusServer.started, "D-Bus server should be started")

				// Connect to the D-Bus service
				conn, err := dbus.SessionBus()
				require.NoError(t, err, "Failed to connect to D-Bus session bus")
				t.Cleanup(func() {
					if err := conn.Close(); err != nil {
						t.Logf("Failed to close D-Bus connection: %v", err)
					}
				})

				// Get the D-Bus object
				obj := conn.Object(DBusServiceName, DBusObjectPath)
				require.NotNil(t, obj, "Failed to get D-Bus object")

				// Call the GetFileStatus method on the test filesystem's D-Bus server
				var status string
				err = obj.Call(DBusInterface+".GetFileStatus", 0, filePath).Store(&status)
				require.NoError(t, err, "Failed to call GetFileStatus method")

				// The status should be "LocalModified" since we just created the file with os.WriteFile
				assert.Equal(t, "LocalModified", status, "File status should be 'LocalModified'")
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup test resources
			testFS, inode, filePath, cleanup := tc.setup(t)

			// Register cleanup
			t.Cleanup(cleanup)

			// Run the test
			tc.test(t, testFS, inode, filePath)
		})
	}
}
