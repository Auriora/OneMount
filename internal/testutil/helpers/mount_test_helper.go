// Package helpers provides testing utilities for the OneMount project.
package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
)

// FilesystemInterface defines the interface for filesystem operations needed by the mount helper
type FilesystemInterface interface {
	IsOffline() bool
}

// MountTestHelper provides utilities for testing mount/unmount operations
type MountTestHelper struct {
	t          *testing.T
	mountPoint string
	filesystem FilesystemInterface
	auth       *graph.Auth
	cleanup    []func() error
}

// NewMountTestHelper creates a new mount test helper
func NewMountTestHelper(t *testing.T) *MountTestHelper {
	// Create a unique mount point for this test
	mountPoint := filepath.Join(testutil.TestSandboxTmpDir, "mount", fmt.Sprintf("test_%d", time.Now().UnixNano()))

	return &MountTestHelper{
		t:          t,
		mountPoint: mountPoint,
		cleanup:    make([]func() error, 0),
	}
}

// FilesystemFactory is a function that creates a filesystem instance
type FilesystemFactory func(auth *graph.Auth, mountPoint string, cacheTTL int) (FilesystemInterface, error)

// SetupMountWithFactory creates and mounts a filesystem for testing using the provided factory
func (h *MountTestHelper) SetupMountWithFactory(factory FilesystemFactory) error {
	// Ensure the mount point directory exists
	if err := os.MkdirAll(h.mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Create authentication for testing
	auth := GetTestAuth()
	h.auth = auth

	// Create the filesystem using the factory
	filesystem, err := factory(auth, h.mountPoint, 300) // 5 minute cache TTL
	if err != nil {
		return fmt.Errorf("failed to create filesystem: %w", err)
	}
	h.filesystem = filesystem

	// Note: Actual mounting is handled by FUSE server in real usage
	// For testing, we just create the filesystem instance

	// Add cleanup for removing mount point
	h.cleanup = append(h.cleanup, func() error {
		return os.RemoveAll(h.mountPoint)
	})

	return nil
}

// GetMountPoint returns the mount point path
func (h *MountTestHelper) GetMountPoint() string {
	return h.mountPoint
}

// GetFilesystem returns the mounted filesystem
func (h *MountTestHelper) GetFilesystem() FilesystemInterface {
	return h.filesystem
}

// GetAuth returns the authentication object
func (h *MountTestHelper) GetAuth() *graph.Auth {
	return h.auth
}

// IsMounted checks if the filesystem is currently available for testing
func (h *MountTestHelper) IsMounted() bool {
	if h.filesystem == nil {
		return false
	}

	// For testing purposes, we consider the filesystem "mounted" if it's created
	// In real usage, this would check if the FUSE mount is active
	return true
}

// Unmount simulates unmounting the filesystem for testing
func (h *MountTestHelper) Unmount() error {
	if h.filesystem == nil {
		return fmt.Errorf("filesystem not initialized")
	}

	// For testing purposes, we just mark it as unmounted
	// In real usage, this would unmount the FUSE filesystem
	return nil
}

// Remount simulates remounting the filesystem for testing
func (h *MountTestHelper) Remount() error {
	// For testing purposes, remount is a no-op since we don't actually mount/unmount
	// In real usage, this would unmount and remount the FUSE filesystem
	return nil
}

// WaitForMount waits for the filesystem to be ready for testing
func (h *MountTestHelper) WaitForMount(_ time.Duration) error {
	// For testing purposes, the filesystem is immediately ready
	// In real usage, this would wait for the FUSE mount to be accessible
	if h.filesystem == nil {
		return fmt.Errorf("filesystem not initialized")
	}
	return nil
}

// WaitForUnmount waits for the filesystem to be unmounted
func (h *MountTestHelper) WaitForUnmount(_ time.Duration) error {
	// For testing purposes, unmount is immediate
	// In real usage, this would wait for the FUSE unmount to complete
	return nil
}

// CreateTestFile creates a test file in the mounted filesystem
func (h *MountTestHelper) CreateTestFile(relativePath string, content []byte) error {
	if !h.IsMounted() {
		return fmt.Errorf("filesystem not mounted")
	}

	fullPath := filepath.Join(h.mountPoint, relativePath)

	// Ensure the directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	return os.WriteFile(fullPath, content, 0644)
}

// ReadTestFile reads a test file from the mounted filesystem
func (h *MountTestHelper) ReadTestFile(relativePath string) ([]byte, error) {
	if !h.IsMounted() {
		return nil, fmt.Errorf("filesystem not mounted")
	}

	fullPath := filepath.Join(h.mountPoint, relativePath)
	return os.ReadFile(fullPath)
}

// VerifyFileExists checks if a file exists in the mounted filesystem
func (h *MountTestHelper) VerifyFileExists(relativePath string) bool {
	if !h.IsMounted() {
		return false
	}

	fullPath := filepath.Join(h.mountPoint, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetMountStats returns statistics about the mounted filesystem
func (h *MountTestHelper) GetMountStats() (*syscall.Statfs_t, error) {
	if !h.IsMounted() {
		return nil, fmt.Errorf("filesystem not mounted")
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(h.mountPoint, &stat); err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	return &stat, nil
}

// Cleanup performs cleanup operations
func (h *MountTestHelper) Cleanup() error {
	var lastErr error

	// Run cleanup functions in reverse order
	for i := len(h.cleanup) - 1; i >= 0; i-- {
		if err := h.cleanup[i](); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// SetupMountTestFixtureWithFactory creates a test fixture for mount/unmount testing with a custom factory
func SetupMountTestFixtureWithFactory(_ *testing.T, fixtureName string, factory FilesystemFactory) *framework.UnitTestFixture {
	return framework.NewUnitTestFixture(fixtureName).
		WithSetup(func(t *testing.T) (interface{}, error) {
			helper := NewMountTestHelper(t)
			if err := helper.SetupMountWithFactory(factory); err != nil {
				return nil, err
			}

			// Wait for mount to be ready
			if err := helper.WaitForMount(10 * time.Second); err != nil {
				if cleanupErr := helper.Cleanup(); cleanupErr != nil {
					// Log cleanup error but return the original error
					t.Logf("Warning: cleanup failed: %v", cleanupErr)
				}
				return nil, err
			}

			return helper, nil
		}).
		WithTeardown(func(_ *testing.T, fixture interface{}) error {
			helper := fixture.(*MountTestHelper)
			return helper.Cleanup()
		})
}
