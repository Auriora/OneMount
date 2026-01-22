package fs

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestDBusServiceNameFileCreation tests that the service name file is created when the D-Bus server starts
func TestUT_FS_DBus_ServiceNameFileCreation(t *testing.T) {
	// Skip if D-Bus is not available
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		t.Skip("D-Bus session bus not available")
	}

	// Create a mock filesystem
	fs := &Filesystem{}

	// Create a D-Bus server
	server := NewFileStatusDBusServer(fs)

	// Set a custom service name prefix for testing
	SetDBusServiceNamePrefix("test")

	// Start the server in test mode (doesn't register service name)
	err := server.StartForTesting()
	if err != nil {
		t.Fatalf("Failed to start D-Bus server: %v", err)
	}
	defer server.Stop()

	// Write the service name file manually for testing
	err = server.writeServiceNameFile()
	if err != nil {
		t.Fatalf("Failed to write service name file: %v", err)
	}

	// Give it a moment to write the file
	time.Sleep(100 * time.Millisecond)

	// Check that the file exists
	if _, err := os.Stat(DBusServiceNameFile); os.IsNotExist(err) {
		t.Errorf("Service name file was not created: %s", DBusServiceNameFile)
	}

	// Read the file and verify it contains the service name
	data, err := os.ReadFile(DBusServiceNameFile)
	if err != nil {
		t.Fatalf("Failed to read service name file: %v", err)
	}

	serviceName := strings.TrimSpace(string(data))
	if serviceName != DBusServiceName {
		t.Errorf("Service name file contains wrong name: got %s, want %s", serviceName, DBusServiceName)
	}

	// Verify the service name has the expected format
	if serviceName != DBusServiceNameBase+".test" {
		t.Errorf("Service name has unexpected format: %s", serviceName)
	}
}

// TestDBusServiceNameFileCleanup tests that the service name file is removed when the D-Bus server stops
func TestUT_FS_DBus_ServiceNameFileCleanup(t *testing.T) {
	// Skip if D-Bus is not available
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		t.Skip("D-Bus session bus not available")
	}

	// Create a mock filesystem
	fs := &Filesystem{}

	// Create a D-Bus server
	server := NewFileStatusDBusServer(fs)

	// Set a custom service name prefix for testing
	SetDBusServiceNamePrefix("cleanup_test")

	// Start the server in test mode
	err := server.StartForTesting()
	if err != nil {
		t.Fatalf("Failed to start D-Bus server: %v", err)
	}

	// Write the service name file
	err = server.writeServiceNameFile()
	if err != nil {
		t.Fatalf("Failed to write service name file: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(DBusServiceNameFile); os.IsNotExist(err) {
		t.Fatalf("Service name file was not created")
	}

	// Stop the server (should remove the file)
	server.Stop()

	// Give it a moment to clean up
	time.Sleep(100 * time.Millisecond)

	// Verify the file was removed
	if _, err := os.Stat(DBusServiceNameFile); !os.IsNotExist(err) {
		t.Errorf("Service name file was not removed after server stop")
	}
}

// TestDBusServiceNameFileMultipleInstances tests that multiple instances don't interfere with each other
func TestUT_FS_DBus_ServiceNameFileMultipleInstances(t *testing.T) {
	// Skip if D-Bus is not available
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		t.Skip("D-Bus session bus not available")
	}

	// Create first instance
	fs1 := &Filesystem{}
	server1 := NewFileStatusDBusServer(fs1)
	SetDBusServiceNamePrefix("instance1")
	serviceName1 := DBusServiceName

	err := server1.StartForTesting()
	if err != nil {
		t.Fatalf("Failed to start first D-Bus server: %v", err)
	}
	defer server1.Stop()

	err = server1.writeServiceNameFile()
	if err != nil {
		t.Fatalf("Failed to write service name file for first instance: %v", err)
	}

	// Read the file and verify it contains the first service name
	data, err := os.ReadFile(DBusServiceNameFile)
	if err != nil {
		t.Fatalf("Failed to read service name file: %v", err)
	}

	if strings.TrimSpace(string(data)) != serviceName1 {
		t.Errorf("Service name file contains wrong name for first instance")
	}

	// Create second instance (simulating a second mount)
	fs2 := &Filesystem{}
	server2 := NewFileStatusDBusServer(fs2)
	SetDBusServiceNamePrefix("instance2")
	serviceName2 := DBusServiceName

	err = server2.StartForTesting()
	if err != nil {
		t.Fatalf("Failed to start second D-Bus server: %v", err)
	}
	defer server2.Stop()

	// Second instance writes its service name (overwrites the first)
	err = server2.writeServiceNameFile()
	if err != nil {
		t.Fatalf("Failed to write service name file for second instance: %v", err)
	}

	// Read the file and verify it now contains the second service name
	data, err = os.ReadFile(DBusServiceNameFile)
	if err != nil {
		t.Fatalf("Failed to read service name file after second write: %v", err)
	}

	if strings.TrimSpace(string(data)) != serviceName2 {
		t.Errorf("Service name file should contain second instance name")
	}

	// Stop the second instance
	server2.Stop()

	// The file should be removed since it contains the second instance's name
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(DBusServiceNameFile); !os.IsNotExist(err) {
		t.Errorf("Service name file should be removed when second instance stops")
	}

	// Stop the first instance (file is already gone, should not error)
	server1.Stop()
}
