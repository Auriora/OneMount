// Package framework provides testing utilities for the OneMount project.
package framework

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// testLogger is a simple implementation of the Logger interface for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, args ...interface{}) {
	l.t.Logf("DEBUG: "+msg, args...)
}

func (l *testLogger) Info(msg string, args ...interface{}) {
	l.t.Logf("INFO: "+msg, args...)
}

func (l *testLogger) Warn(msg string, args ...interface{}) {
	l.t.Logf("WARN: "+msg, args...)
}

func (l *testLogger) Error(msg string, args ...interface{}) {
	l.t.Logf("ERROR: "+msg, args...)
}

func TestFileSystemResource_Basic(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem-resource-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mount point within the temporary directory
	mountPoint := filepath.Join(tempDir, "mount-point")

	// Create a new FileSystemResource
	resource := NewFileSystemResource(mountPoint)

	// Verify initial state
	if resource.IsMounted() {
		t.Errorf("Expected resource to not be mounted initially")
	}

	if resource.GetMountPoint() != mountPoint {
		t.Errorf("Expected mount point to be %s, got %s", mountPoint, resource.GetMountPoint())
	}

	// Test SetCleanupFunc
	cleanupCalled := false
	resource.SetCleanupFunc(func() error {
		cleanupCalled = true
		return nil
	})

	// Test Cleanup (without actually mounting)
	if err := resource.Cleanup(); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// Verify cleanup function was called
	if !cleanupCalled {
		t.Errorf("Expected cleanup function to be called")
	}

	// Verify mount point was removed
	if _, err := os.Stat(mountPoint); !os.IsNotExist(err) {
		t.Errorf("Expected mount point to be removed after cleanup")
	}
}

func TestFileSystemResource_WithTestFramework(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem-resource-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mount point within the temporary directory
	mountPoint := filepath.Join(tempDir, "mount-point")

	// Create a new FileSystemResource
	resource := NewFileSystemResource(mountPoint)

	// Create a test framework
	logger := &testLogger{t: t}
	framework := NewTestFramework(TestConfig{}, logger)

	// Add the resource to the framework
	framework.AddResource(resource)

	// Verify the resource was added
	if len(framework.resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(framework.resources))
	}

	// Test CleanupResources
	if err := framework.CleanupResources(); err != nil {
		t.Errorf("CleanupResources failed: %v", err)
	}

	// Verify resources were cleared
	if len(framework.resources) != 0 {
		t.Errorf("Expected 0 resources after cleanup, got %d", len(framework.resources))
	}

	// Verify mount point was removed
	if _, err := os.Stat(mountPoint); !os.IsNotExist(err) {
		t.Errorf("Expected mount point to be removed after cleanup")
	}
}

// TestFileSystemResource_MountUnmount tests the mounting and unmounting functionality
// without actually executing system commands.
func TestFileSystemResource_MountUnmount(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem-resource-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mount point within the temporary directory
	mountPoint := filepath.Join(tempDir, "mount-point")

	// Create a new FileSystemResource
	resource := NewFileSystemResource(mountPoint)

	// Verify initial state
	if resource.IsMounted() {
		t.Errorf("Expected resource to not be mounted initially")
	}

	// Save the original exec.Command function
	origExecCommand := exec.Command
	defer func() { exec.Command = origExecCommand }()

	// Mock the exec.Command function to avoid actually executing commands
	execCommandCalled := false
	execCommandArgs := []string{}
	exec.Command = func(command string, args ...string) *exec.Cmd {
		execCommandCalled = true
		execCommandArgs = append([]string{command}, args...)
		// Return a dummy command that does nothing
		return &exec.Cmd{
			Path: command,
			Args: append([]string{command}, args...),
		}
	}

	// Test Mount
	err = resource.Mount("onemount", "--option1", "--option2", "/path/to/source")
	if err != nil {
		t.Errorf("Mount failed: %v", err)
	}

	// Verify Mount called exec.Command with the correct arguments
	if !execCommandCalled {
		t.Errorf("Expected exec.Command to be called")
	}
	expectedArgs := []string{"onemount", "--option1", "--option2", "/path/to/source"}
	if len(execCommandArgs) != len(expectedArgs) {
		t.Errorf("Expected %d arguments, got %d", len(expectedArgs), len(execCommandArgs))
	} else {
		for i, arg := range expectedArgs {
			if execCommandArgs[i] != arg {
				t.Errorf("Expected argument %d to be %s, got %s", i, arg, execCommandArgs[i])
			}
		}
	}

	// Verify the resource is now marked as mounted
	if !resource.IsMounted() {
		t.Errorf("Expected resource to be mounted after Mount")
	}

	// Verify the command and args were stored
	if resource.Command != "onemount" {
		t.Errorf("Expected Command to be 'onemount', got '%s'", resource.Command)
	}
	expectedArgs = []string{"--option1", "--option2", "/path/to/source"}
	if len(resource.Args) != len(expectedArgs) {
		t.Errorf("Expected %d arguments, got %d", len(expectedArgs), len(resource.Args))
	} else {
		for i, arg := range expectedArgs {
			if resource.Args[i] != arg {
				t.Errorf("Expected argument %d to be %s, got %s", i, arg, resource.Args[i])
			}
		}
	}

	// Reset the mock
	execCommandCalled = false
	execCommandArgs = []string{}

	// Test Unmount
	err = resource.Unmount()
	if err != nil {
		t.Errorf("Unmount failed: %v", err)
	}

	// Verify Unmount called exec.Command with the correct arguments
	if !execCommandCalled {
		t.Errorf("Expected exec.Command to be called")
	}
	expectedArgs = []string{"fusermount3", "-u", mountPoint}
	if len(execCommandArgs) != len(expectedArgs) {
		t.Errorf("Expected %d arguments, got %d", len(expectedArgs), len(execCommandArgs))
	} else {
		for i, arg := range expectedArgs {
			if execCommandArgs[i] != arg {
				t.Errorf("Expected argument %d to be %s, got %s", i, arg, execCommandArgs[i])
			}
		}
	}

	// Verify the resource is now marked as unmounted
	if resource.IsMounted() {
		t.Errorf("Expected resource to not be mounted after Unmount")
	}

	// Test Remount
	execCommandCalled = false
	execCommandArgs = []string{}

	// First mount the resource
	err = resource.Mount("onemount", "--option1", "--option2", "/path/to/source")
	if err != nil {
		t.Errorf("Mount failed: %v", err)
	}

	// Reset the mock
	execCommandCalled = false
	execCommandArgs = []string{}

	// Then remount it
	err = resource.Remount()
	if err != nil {
		t.Errorf("Remount failed: %v", err)
	}

	// Verify Remount called exec.Command twice (once for Unmount, once for Mount)
	if !execCommandCalled {
		t.Errorf("Expected exec.Command to be called")
	}

	// Verify the resource is still marked as mounted
	if !resource.IsMounted() {
		t.Errorf("Expected resource to be mounted after Remount")
	}
}
