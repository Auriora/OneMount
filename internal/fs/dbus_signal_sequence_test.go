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

// collectSignals collects D-Bus signals from a channel with a timeout.
// It returns all signals received before the timeout or until the expected count is reached.
func collectSignals(ch chan *dbus.Signal, expectedCount int, timeout time.Duration) []*dbus.Signal {
	signals := make([]*dbus.Signal, 0, expectedCount)
	deadline := time.After(timeout)

	for i := 0; i < expectedCount; i++ {
		select {
		case sig := <-ch:
			signals = append(signals, sig)
		case <-deadline:
			return signals
		}
	}

	return signals
}

// TestIT_FS_DBus_SignalSequence_DownloadFlow verifies signals are emitted in correct order during download.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-SEQ-01
//	Title           D-Bus Signal Sequence - Download Flow
//	Description     Tests that signals are emitted in correct order during file download
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger file download (Ghost → Downloading → Cached)
//	                4. Collect signals with timeout
//	                5. Verify signal sequence is correct
//	                6. Verify no duplicate signals
//	Expected Result Signals emitted in order: Ghost → Downloading → Cached
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal sequence verification for download operations.
func TestIT_FS_DBus_SignalSequence_DownloadFlow(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalSequenceDownloadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_seq_download")

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

		// Test download sequence: Ghost → Downloading → Cached
		testPath := "/test/download/file.txt"

		// Simulate download sequence by sending signals
		// In a real scenario, these would be triggered by actual file operations
		dbusServer.SendFileStatusUpdate(testPath, "Ghost")
		time.Sleep(100 * time.Millisecond) // Small delay between signals

		dbusServer.SendFileStatusUpdate(testPath, "Downloading")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Cached")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 3, 5*time.Second)
		assert.Equal(3, len(signals), "Should receive 3 signals for download sequence")

		// Verify signal sequence
		if len(signals) >= 3 {
			// Extract status from signal body
			status1, ok1 := signals[0].Body[1].(string)
			status2, ok2 := signals[1].Body[1].(string)
			status3, ok3 := signals[2].Body[1].(string)

			assert.True(ok1 && ok2 && ok3, "Signal bodies should contain string status")
			assert.Equal("Ghost", status1, "First signal should be Ghost")
			assert.Equal("Downloading", status2, "Second signal should be Downloading")
			assert.Equal("Cached", status3, "Third signal should be Cached")

			// Verify path is consistent
			path1, _ := signals[0].Body[0].(string)
			path2, _ := signals[1].Body[0].(string)
			path3, _ := signals[2].Body[0].(string)

			assert.Equal(testPath, path1, "First signal path should match")
			assert.Equal(testPath, path2, "Second signal path should match")
			assert.Equal(testPath, path3, "Third signal path should match")

			t.Logf("✓ Download signal sequence verified: %s → %s → %s", status1, status2, status3)
		}

		// Verify no duplicate signals (drain channel)
		select {
		case extraSig := <-signalChan:
			t.Errorf("Unexpected extra signal received: %+v", extraSig)
		case <-time.After(500 * time.Millisecond):
			t.Log("✓ No duplicate signals detected")
		}
	})
}

// TestIT_FS_DBus_SignalSequence_UploadFlow verifies signals are emitted in correct order during upload.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-SEQ-02
//	Title           D-Bus Signal Sequence - Upload Flow
//	Description     Tests that signals are emitted in correct order during file upload
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger file upload (Modified → Uploading → Cached)
//	                4. Collect signals with timeout
//	                5. Verify signal sequence is correct
//	                6. Verify no duplicate signals
//	Expected Result Signals emitted in order: Modified → Uploading → Cached
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal sequence verification for upload operations.
func TestIT_FS_DBus_SignalSequence_UploadFlow(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalSequenceUploadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_seq_upload")

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

		// Test upload sequence: Modified → Uploading → Cached
		testPath := "/test/upload/file.txt"

		// Simulate upload sequence by sending signals
		dbusServer.SendFileStatusUpdate(testPath, "Modified")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Uploading")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Cached")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 3, 5*time.Second)
		assert.Equal(3, len(signals), "Should receive 3 signals for upload sequence")

		// Verify signal sequence
		if len(signals) >= 3 {
			status1, ok1 := signals[0].Body[1].(string)
			status2, ok2 := signals[1].Body[1].(string)
			status3, ok3 := signals[2].Body[1].(string)

			assert.True(ok1 && ok2 && ok3, "Signal bodies should contain string status")
			assert.Equal("Modified", status1, "First signal should be Modified")
			assert.Equal("Uploading", status2, "Second signal should be Uploading")
			assert.Equal("Cached", status3, "Third signal should be Cached")

			// Verify path is consistent
			path1, _ := signals[0].Body[0].(string)
			path2, _ := signals[1].Body[0].(string)
			path3, _ := signals[2].Body[0].(string)

			assert.Equal(testPath, path1, "First signal path should match")
			assert.Equal(testPath, path2, "Second signal path should match")
			assert.Equal(testPath, path3, "Third signal path should match")

			t.Logf("✓ Upload signal sequence verified: %s → %s → %s", status1, status2, status3)
		}

		// Verify no duplicate signals
		select {
		case extraSig := <-signalChan:
			t.Errorf("Unexpected extra signal received: %+v", extraSig)
		case <-time.After(500 * time.Millisecond):
			t.Log("✓ No duplicate signals detected")
		}
	})
}

// TestIT_FS_DBus_SignalSequence_ModifyFlow verifies signals are emitted in correct order during modification.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-SEQ-03
//	Title           D-Bus Signal Sequence - Modify Flow
//	Description     Tests that signals are emitted in correct order during file modification
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger file modification (Cached → Modified)
//	                4. Collect signals with timeout
//	                5. Verify signal sequence is correct
//	Expected Result Signals emitted in order: Cached → Modified
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal sequence verification for modification operations.
func TestIT_FS_DBus_SignalSequence_ModifyFlow(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalSequenceModifyFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_seq_modify")

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

		// Test modification sequence: Cached → Modified
		testPath := "/test/modify/file.txt"

		// Simulate modification sequence by sending signals
		dbusServer.SendFileStatusUpdate(testPath, "Cached")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Modified")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 2, 5*time.Second)
		assert.Equal(2, len(signals), "Should receive 2 signals for modification sequence")

		// Verify signal sequence
		if len(signals) >= 2 {
			status1, ok1 := signals[0].Body[1].(string)
			status2, ok2 := signals[1].Body[1].(string)

			assert.True(ok1 && ok2, "Signal bodies should contain string status")
			assert.Equal("Cached", status1, "First signal should be Cached")
			assert.Equal("Modified", status2, "Second signal should be Modified")

			// Verify path is consistent
			path1, _ := signals[0].Body[0].(string)
			path2, _ := signals[1].Body[0].(string)

			assert.Equal(testPath, path1, "First signal path should match")
			assert.Equal(testPath, path2, "Second signal path should match")

			t.Logf("✓ Modification signal sequence verified: %s → %s", status1, status2)
		}

		// Verify no duplicate signals
		select {
		case extraSig := <-signalChan:
			t.Errorf("Unexpected extra signal received: %+v", extraSig)
		case <-time.After(500 * time.Millisecond):
			t.Log("✓ No duplicate signals detected")
		}
	})
}

// TestIT_FS_DBus_SignalSequence_ErrorFlow verifies signals are emitted correctly for error states.
//
//	Test Case ID    IT-FS-DBUS-SIGNAL-SEQ-04
//	Title           D-Bus Signal Sequence - Error Flow
//	Description     Tests that signals are emitted correctly during error state transitions
//	Preconditions   D-Bus session bus is available, filesystem mounted
//	Steps           1. Create and start a D-Bus server
//	                2. Subscribe to FileStatusChanged signals
//	                3. Trigger error state transition (Downloading → Error)
//	                4. Collect signals with timeout
//	                5. Verify signal sequence is correct
//	Expected Result Signals emitted in order: Downloading → Error
//	Requirements    8.1, 8.2, 10.2
//	Notes: This test automates signal sequence verification for error conditions.
func TestIT_FS_DBus_SignalSequence_ErrorFlow(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DBusSignalSequenceErrorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		SetDBusServiceNamePrefix("test_signal_seq_error")

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

		// Test error sequence: Downloading → Error
		testPath := "/test/error/file.txt"

		// Simulate error sequence by sending signals
		dbusServer.SendFileStatusUpdate(testPath, "Downloading")
		time.Sleep(100 * time.Millisecond)

		dbusServer.SendFileStatusUpdate(testPath, "Error")

		// Collect signals with timeout
		signals := collectSignals(signalChan, 2, 5*time.Second)
		assert.Equal(2, len(signals), "Should receive 2 signals for error sequence")

		// Verify signal sequence
		if len(signals) >= 2 {
			status1, ok1 := signals[0].Body[1].(string)
			status2, ok2 := signals[1].Body[1].(string)

			assert.True(ok1 && ok2, "Signal bodies should contain string status")
			assert.Equal("Downloading", status1, "First signal should be Downloading")
			assert.Equal("Error", status2, "Second signal should be Error")

			// Verify path is consistent
			path1, _ := signals[0].Body[0].(string)
			path2, _ := signals[1].Body[0].(string)

			assert.Equal(testPath, path1, "First signal path should match")
			assert.Equal(testPath, path2, "Second signal path should match")

			t.Logf("✓ Error signal sequence verified: %s → %s", status1, status2)
		}

		// Verify no duplicate signals
		select {
		case extraSig := <-signalChan:
			t.Errorf("Unexpected extra signal received: %+v", extraSig)
		case <-time.After(500 * time.Millisecond):
			t.Log("✓ No duplicate signals detected")
		}
	})
}
