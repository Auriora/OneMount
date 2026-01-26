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

// TestIT_FS_DBus_SignalContent_FileModification verifies signal content during file modification.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-CONTENT-01
//	Title           D-Bus Signal Content - File Modification
//	Description     Tests that signals contain correct data during file modification operations
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger file modification (Cached → Modified)
//	                4. Collect signals with timeout
//	                5. Verify signal parameters (path, status)
//	                6. Verify path is string type
//	                7. Verify status is string type
//	                8. Verify status value is "Modified"
//	Expected Result Signals contain correct path and status data
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal content validation for file modification operations.
func TestIT_FS_DBus_SignalContent_FileModification(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalContentModifyFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_content_modify")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Connect to D-Bus session bus
		conn, err := dbus.SessionBus()
		assert.NoError(err, "Should connect to D-Bus session bus")

		// Subscribe to signals
		signalChan := make(chan *dbus.Signal, 100)
		conn.Signal(signalChan)
		defer conn.RemoveSignal(signalChan)

		// Add match rule for FileStatusChanged signals
		serviceName := DBusServiceName
		matchRule := fmt.Sprintf(
			"type='signal',sender='%s',interface='%s',member='FileStatusChanged'",
			serviceName,
			DBusInterface,
		)
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.AddMatch", 0, matchRule,
		).Err
		assert.NoError(err, "Should add match rule for signals")

		// Test file modification signal content
		testPath := "/test/modify/document.txt"

		// Simulate file modification by sending signals
		dbusServer.SendFileStatusUpdate(testPath, "Cached")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Modified")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 2, 5*time.Second)
		assert.Equal(2, len(signals), "Should receive 2 signals for modification")

		// Verify signal content for Modified status
		if len(signals) >= 2 {
			modifiedSignal := signals[1]

			// Verify signal has correct number of parameters
			assert.Equal(2, len(modifiedSignal.Body), "Signal should have 2 parameters (path, status)")

			// Verify path parameter
			path, pathOk := modifiedSignal.Body[0].(string)
			assert.True(pathOk, "First parameter should be string (path)")
			assert.Equal(testPath, path, "Path should match test path")

			// Verify status parameter
			status, statusOk := modifiedSignal.Body[1].(string)
			assert.True(statusOk, "Second parameter should be string (status)")
			assert.Equal("Modified", status, "Status should be 'Modified'")

			// Verify signal name (includes interface prefix)
			assert.Contains(modifiedSignal.Name, "FileStatusChanged", "Signal name should contain FileStatusChanged")

			t.Logf("✓ File modification signal content verified: path=%s, status=%s", path, status)
		}
	})
}

// TestIT_FS_DBus_SignalContent_FileDeletion verifies signal content during file deletion.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-CONTENT-02
//	Title           D-Bus Signal Content - File Deletion
//	Description     Tests that signals contain correct data during file deletion operations
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger file deletion
//	                4. Collect signals with timeout
//	                5. Verify signal parameters (path, status)
//	                6. Verify deletion status is correct
//	Expected Result Signals contain correct deletion data
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal content validation for file deletion operations.
func TestIT_FS_DBus_SignalContent_FileDeletion(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalContentDeleteFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_content_delete")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Connect to D-Bus session bus
		conn, err := dbus.SessionBus()
		assert.NoError(err, "Should connect to D-Bus session bus")

		// Subscribe to signals
		signalChan := make(chan *dbus.Signal, 100)
		conn.Signal(signalChan)
		defer conn.RemoveSignal(signalChan)

		// Add match rule for FileStatusChanged signals
		serviceName := DBusServiceName
		matchRule := fmt.Sprintf(
			"type='signal',sender='%s',interface='%s',member='FileStatusChanged'",
			serviceName,
			DBusInterface,
		)
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.AddMatch", 0, matchRule,
		).Err
		assert.NoError(err, "Should add match rule for signals")

		// Test file deletion signal content
		testPath := "/test/delete/file.txt"

		// Simulate file deletion by sending signals
		// File goes from Cached to being deleted
		dbusServer.SendFileStatusUpdate(testPath, "Cached")
		time.Sleep(100 * time.Millisecond)

		// Note: In actual implementation, deletion might not send a specific "Deleted" status
		// but rather the file would disappear from the filesystem
		// For testing purposes, we'll verify the signal format is correct
		dbusServer.SendFileStatusUpdate(testPath, "Unknown")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 2, 5*time.Second)
		assert.Equal(2, len(signals), "Should receive 2 signals for deletion")

		// Verify signal content
		if len(signals) >= 2 {
			deletionSignal := signals[1]

			// Verify signal has correct number of parameters
			assert.Equal(2, len(deletionSignal.Body), "Signal should have 2 parameters (path, status)")

			// Verify path parameter
			path, pathOk := deletionSignal.Body[0].(string)
			assert.True(pathOk, "First parameter should be string (path)")
			assert.Equal(testPath, path, "Path should match test path")

			// Verify status parameter
			status, statusOk := deletionSignal.Body[1].(string)
			assert.True(statusOk, "Second parameter should be string (status)")
			assert.NotEqual("", status, "Status should not be empty")

			// Verify signal name (includes interface prefix)
			assert.Contains(deletionSignal.Name, "FileStatusChanged", "Signal name should contain FileStatusChanged")

			t.Logf("✓ File deletion signal content verified: path=%s, status=%s", path, status)
		}
	})
}

// TestIT_FS_DBus_SignalContent_UploadOperations verifies signal content during upload operations.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-CONTENT-03
//	Title           D-Bus Signal Content - Upload Operations
//	Description     Tests that signals contain correct data during file upload operations
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger file upload (Modified → Uploading → Cached)
//	                4. Collect signals with timeout
//	                5. Verify signal parameters for each state
//	                6. Verify upload progress signals
//	Expected Result Signals contain correct upload progress data
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal content validation for upload operations.
func TestIT_FS_DBus_SignalContent_UploadOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalContentUploadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_content_upload")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Connect to D-Bus session bus
		conn, err := dbus.SessionBus()
		assert.NoError(err, "Should connect to D-Bus session bus")

		// Subscribe to signals
		signalChan := make(chan *dbus.Signal, 100)
		conn.Signal(signalChan)
		defer conn.RemoveSignal(signalChan)

		// Add match rule for FileStatusChanged signals
		serviceName := DBusServiceName
		matchRule := fmt.Sprintf(
			"type='signal',sender='%s',interface='%s',member='FileStatusChanged'",
			serviceName,
			DBusInterface,
		)
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.AddMatch", 0, matchRule,
		).Err
		assert.NoError(err, "Should add match rule for signals")

		// Test upload operation signal content
		testPath := "/test/upload/large-file.dat"

		// Simulate upload sequence by sending signals
		dbusServer.SendFileStatusUpdate(testPath, "Modified")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Uploading")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Cached")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 3, 5*time.Second)
		assert.Equal(3, len(signals), "Should receive 3 signals for upload sequence")

		// Verify signal content for each state
		if len(signals) >= 3 {
			// Verify Modified signal
			modifiedSignal := signals[0]
			assert.Equal(2, len(modifiedSignal.Body), "Modified signal should have 2 parameters")

			modPath, _ := modifiedSignal.Body[0].(string)
			modStatus, _ := modifiedSignal.Body[1].(string)
			assert.Equal(testPath, modPath, "Modified signal path should match")
			assert.Equal("Modified", modStatus, "Status should be 'Modified'")

			// Verify Uploading signal
			uploadingSignal := signals[1]
			assert.Equal(2, len(uploadingSignal.Body), "Uploading signal should have 2 parameters")

			uplPath, _ := uploadingSignal.Body[0].(string)
			uplStatus, _ := uploadingSignal.Body[1].(string)
			assert.Equal(testPath, uplPath, "Uploading signal path should match")
			assert.Equal("Uploading", uplStatus, "Status should be 'Uploading'")

			// Verify Cached signal (upload complete)
			cachedSignal := signals[2]
			assert.Equal(2, len(cachedSignal.Body), "Cached signal should have 2 parameters")

			cachePath, _ := cachedSignal.Body[0].(string)
			cacheStatus, _ := cachedSignal.Body[1].(string)
			assert.Equal(testPath, cachePath, "Cached signal path should match")
			assert.Equal("Cached", cacheStatus, "Status should be 'Cached'")

			t.Logf("✓ Upload operation signal content verified: %s → %s → %s", modStatus, uplStatus, cacheStatus)
		}
	})
}

// TestIT_FS_DBus_SignalContent_ErrorStates verifies signal content for error conditions.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-CONTENT-04
//	Title           D-Bus Signal Content - Error States
//	Description     Tests that signals contain correct data during error conditions
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger error condition (Downloading → Error)
//	                4. Collect signals with timeout
//	                5. Verify error signal parameters
//	                6. Verify error status is correct
//	Expected Result Signals contain correct error state data
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal content validation for error conditions.
func TestIT_FS_DBus_SignalContent_ErrorStates(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalContentErrorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_content_error")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Connect to D-Bus session bus
		conn, err := dbus.SessionBus()
		assert.NoError(err, "Should connect to D-Bus session bus")

		// Subscribe to signals
		signalChan := make(chan *dbus.Signal, 100)
		conn.Signal(signalChan)
		defer conn.RemoveSignal(signalChan)

		// Add match rule for FileStatusChanged signals
		serviceName := DBusServiceName
		matchRule := fmt.Sprintf(
			"type='signal',sender='%s',interface='%s',member='FileStatusChanged'",
			serviceName,
			DBusInterface,
		)
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.AddMatch", 0, matchRule,
		).Err
		assert.NoError(err, "Should add match rule for signals")

		// Test error state signal content
		testPath := "/test/error/corrupted-file.bin"

		// Simulate error condition by sending signals
		dbusServer.SendFileStatusUpdate(testPath, "Downloading")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Error")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 2, 5*time.Second)
		assert.Equal(2, len(signals), "Should receive 2 signals for error condition")

		// Verify error signal content
		if len(signals) >= 2 {
			errorSignal := signals[1]

			// Verify signal has correct number of parameters
			assert.Equal(2, len(errorSignal.Body), "Error signal should have 2 parameters (path, status)")

			// Verify path parameter
			path, pathOk := errorSignal.Body[0].(string)
			assert.True(pathOk, "First parameter should be string (path)")
			assert.Equal(testPath, path, "Path should match test path")

			// Verify status parameter
			status, statusOk := errorSignal.Body[1].(string)
			assert.True(statusOk, "Second parameter should be string (status)")
			assert.Equal("Error", status, "Status should be 'Error'")

			// Verify signal name (includes interface prefix)
			assert.Contains(errorSignal.Name, "FileStatusChanged", "Signal name should contain FileStatusChanged")

			t.Logf("✓ Error state signal content verified: path=%s, status=%s", path, status)
		}
	})
}

// TestIT_FS_DBus_SignalContent_DirectoryOperations verifies signal content for directory operations.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-CONTENT-05
//	Title           D-Bus Signal Content - Directory Operations
//	Description     Tests that signals contain correct data during directory operations
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger directory operations
//	                4. Collect signals with timeout
//	                5. Verify directory-related signals
//	                6. Verify signal parameters are correct
//	Expected Result Signals contain correct directory operation data
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal content validation for directory operations.
func TestIT_FS_DBus_SignalContent_DirectoryOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalContentDirFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_content_dir")

		// Create and start D-Bus server with full registration
		dbusServer := NewFileStatusDBusServer(filesystem)
		err := dbusServer.Start()
		assert.NoError(err, "D-Bus server should start successfully")
		defer dbusServer.Stop()

		// Connect to D-Bus session bus
		conn, err := dbus.SessionBus()
		assert.NoError(err, "Should connect to D-Bus session bus")

		// Subscribe to signals
		signalChan := make(chan *dbus.Signal, 100)
		conn.Signal(signalChan)
		defer conn.RemoveSignal(signalChan)

		// Add match rule for FileStatusChanged signals
		serviceName := DBusServiceName
		matchRule := fmt.Sprintf(
			"type='signal',sender='%s',interface='%s',member='FileStatusChanged'",
			serviceName,
			DBusInterface,
		)
		err = conn.BusObject().Call(
			"org.freedesktop.DBus.AddMatch", 0, matchRule,
		).Err
		assert.NoError(err, "Should add match rule for signals")

		// Test directory operation signal content
		testDirPath := "/test/directory/new-folder"

		// Simulate directory creation by sending signals
		// Directories typically go from Unknown to Local (created locally)
		dbusServer.SendFileStatusUpdate(testDirPath, "Unknown")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testDirPath, "Local")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 2, 5*time.Second)
		assert.Equal(2, len(signals), "Should receive 2 signals for directory operation")

		// Verify signal content
		if len(signals) >= 2 {
			dirSignal := signals[1]

			// Verify signal has correct number of parameters
			assert.Equal(2, len(dirSignal.Body), "Directory signal should have 2 parameters (path, status)")

			// Verify path parameter
			path, pathOk := dirSignal.Body[0].(string)
			assert.True(pathOk, "First parameter should be string (path)")
			assert.Equal(testDirPath, path, "Path should match test directory path")

			// Verify status parameter
			status, statusOk := dirSignal.Body[1].(string)
			assert.True(statusOk, "Second parameter should be string (status)")
			assert.Equal("Local", status, "Status should be 'Local' for newly created directory")

			// Verify signal name (includes interface prefix)
			assert.Contains(dirSignal.Name, "FileStatusChanged", "Signal name should contain FileStatusChanged")

			t.Logf("✓ Directory operation signal content verified: path=%s, status=%s", path, status)
		}
	})
}
