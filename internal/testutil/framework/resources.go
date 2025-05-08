// Package framework provides testing utilities for the OneMount project.
//
// This file implements the FileSystemResource type, which represents a mounted filesystem
// resource that needs cleanup after tests. It implements the TestResource interface and
// provides methods for mounting and unmounting filesystems.
//
// # Overview
//
// The FileSystemResource type provides a convenient way to manage filesystem resources
// during tests. It handles mounting and unmounting filesystems, as well as cleaning up
// resources after tests. The implementation is thread-safe, making it suitable for use
// in concurrent tests.
//
// # Basic Usage
//
//	// Create a new FileSystemResource
//	resource := NewFileSystemResource("/path/to/mount/point")
//
//	// Mount the filesystem
//	err := resource.Mount("onemount", "--option1", "--option2", "/path/to/source")
//	if err != nil {
//		// Handle error
//	}
//
//	// Add the resource to the test framework for automatic cleanup
//	framework.AddResource(resource)
//
//	// Or manually clean up the resource
//	err = resource.Cleanup()
//	if err != nil {
//		// Handle error
//	}
//
// # Advanced Usage
//
// The FileSystemResource type provides additional methods for more advanced use cases:
//
// ## Checking if a filesystem is mounted
//
//	if resource.IsMounted() {
//		// Filesystem is mounted
//	} else {
//		// Filesystem is not mounted
//	}
//
// ## Getting the mount point
//
//	mountPoint := resource.GetMountPoint()
//	fmt.Printf("Filesystem is mounted at %s\n", mountPoint)
//
// ## Remounting a filesystem
//
//	// Remount the filesystem (unmount and then mount again)
//	err := resource.Remount()
//	if err != nil {
//		// Handle error
//	}
//
// ## Custom cleanup
//
//	// Set a custom cleanup function to be called during cleanup
//	resource.SetCleanupFunc(func() error {
//		// Perform custom cleanup operations
//		return nil
//	})
//
// # Thread Safety
//
// All methods of the FileSystemResource type are thread-safe. The implementation uses
// a mutex to protect concurrent access to the resource. This makes it safe to use
// in concurrent tests.
//
// # Error Handling
//
// All methods that can fail return an error. It's important to check these errors
// and handle them appropriately. The Cleanup method will attempt to unmount the
// filesystem and remove the mount point directory, even if the filesystem was not
// successfully mounted.
//
// # Integration with TestFramework
//
// The FileSystemResource type implements the TestResource interface, which allows
// it to be used with the TestFramework. When added to the TestFramework using the
// AddResource method, the resource will be automatically cleaned up when the
// CleanupResources method is called.
package framework

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// FileSystemResource represents a mounted filesystem resource that needs cleanup after tests.
// It implements the TestResource interface.
type FileSystemResource struct {
	// MountPoint is the directory where the filesystem is mounted
	MountPoint string
	// Mounted indicates whether the filesystem is currently mounted
	Mounted bool
	// Command is the command used to mount the filesystem
	Command string
	// Args are the arguments passed to the mount command
	Args []string
	// CleanupFunc is an optional function to run during cleanup
	CleanupFunc func() error
	// mu is a mutex to protect concurrent access to the resource
	mu sync.Mutex
}

// NewFileSystemResource creates a new FileSystemResource.
// It does not mount the filesystem; use Mount() to do that.
func NewFileSystemResource(mountPoint string) *FileSystemResource {
	return &FileSystemResource{
		MountPoint: mountPoint,
		Mounted:    false,
		mu:         sync.Mutex{},
	}
}

// Mount mounts the filesystem using the specified command and arguments.
// It creates the mount point directory if it doesn't exist.
func (r *FileSystemResource) Mount(command string, args ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already mounted
	if r.Mounted {
		return fmt.Errorf("filesystem already mounted at %s", r.MountPoint)
	}

	// Create the mount point if it doesn't exist
	if err := os.MkdirAll(r.MountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point directory: %w", err)
	}

	// Store the command and args for potential remounting
	r.Command = command
	r.Args = args

	// Execute the mount command
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount filesystem: %w, output: %s", err, string(output))
	}

	r.Mounted = true
	return nil
}

// Unmount unmounts the filesystem.
// It uses fusermount3 -u for FUSE filesystems, which is appropriate for OneMount.
func (r *FileSystemResource) Unmount() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.Mounted {
		return nil // Already unmounted, nothing to do
	}

	// Use fusermount3 -u to unmount FUSE filesystems
	cmd := exec.Command("fusermount3", "-u", r.MountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unmount filesystem: %w, output: %s", err, string(output))
	}

	r.Mounted = false
	return nil
}

// Cleanup implements the TestResource interface.
// It unmounts the filesystem and removes the mount point directory.
func (r *FileSystemResource) Cleanup() error {
	// First unmount the filesystem
	if err := r.Unmount(); err != nil {
		return err
	}

	// Run any additional cleanup function if provided
	if r.CleanupFunc != nil {
		if err := r.CleanupFunc(); err != nil {
			return fmt.Errorf("cleanup function failed: %w", err)
		}
	}

	// Remove the mount point directory
	if err := os.RemoveAll(r.MountPoint); err != nil {
		return fmt.Errorf("failed to remove mount point directory: %w", err)
	}

	return nil
}

// SetCleanupFunc sets an additional cleanup function to be called during Cleanup.
func (r *FileSystemResource) SetCleanupFunc(cleanupFunc func() error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CleanupFunc = cleanupFunc
}

// IsMounted returns whether the filesystem is currently mounted.
func (r *FileSystemResource) IsMounted() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Mounted
}

// GetMountPoint returns the mount point of the filesystem.
func (r *FileSystemResource) GetMountPoint() string {
	return r.MountPoint
}

// Remount unmounts and then remounts the filesystem using the same command and arguments.
func (r *FileSystemResource) Remount() error {
	if err := r.Unmount(); err != nil {
		return err
	}
	return r.Mount(r.Command, r.Args...)
}
