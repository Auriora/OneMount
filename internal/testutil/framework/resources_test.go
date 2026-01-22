// Package framework provides testing utilities for the OneMount project.
package framework

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUT_Framework_Resources_FileSystemResource_Basic(t *testing.T) {
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

func TestUT_Framework_Resources_FileSystemResource_WithTestFramework(t *testing.T) {
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
// by directly manipulating the Mounted flag without executing system commands.
func TestUT_Framework_Resources_FileSystemResource_MountUnmount(t *testing.T) {
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

	// Manually set the mounted flag and command/args
	resource.mu.Lock()
	resource.Mounted = true
	resource.Command = "onemount"
	resource.Args = []string{"--option1", "--option2", "/path/to/source"}
	resource.mu.Unlock()

	// Verify the resource is now marked as mounted
	if !resource.IsMounted() {
		t.Errorf("Expected resource to be mounted after setting Mounted flag")
	}

	// Verify the command and args were stored
	if resource.Command != "onemount" {
		t.Errorf("Expected Command to be 'onemount', got '%s'", resource.Command)
	}
	expectedArgs := []string{"--option1", "--option2", "/path/to/source"}
	if len(resource.Args) != len(expectedArgs) {
		t.Errorf("Expected %d arguments, got %d", len(expectedArgs), len(resource.Args))
	} else {
		for i, arg := range expectedArgs {
			if resource.Args[i] != arg {
				t.Errorf("Expected argument %d to be %s, got %s", i, arg, resource.Args[i])
			}
		}
	}

	// Manually set the mounted flag to false
	resource.mu.Lock()
	resource.Mounted = false
	resource.mu.Unlock()

	// Verify the resource is now marked as unmounted
	if resource.IsMounted() {
		t.Errorf("Expected resource to not be mounted after setting Mounted flag to false")
	}

	// Manually set the mounted flag to true again
	resource.mu.Lock()
	resource.Mounted = true
	resource.mu.Unlock()

	// Verify the resource is now marked as mounted again
	if !resource.IsMounted() {
		t.Errorf("Expected resource to be mounted after setting Mounted flag to true")
	}
}
