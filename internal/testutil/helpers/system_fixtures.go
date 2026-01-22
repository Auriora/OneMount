// Package helpers provides testing utilities for the OneMount project.
package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
)

// SystemTestFixture represents a full end-to-end system test fixture with mounting.
type SystemTestFixture struct {
	// TempDir is the temporary directory for the test
	TempDir string
	// MountPoint is the directory where the filesystem is mounted
	MountPoint string
	// Auth is the authentication object
	Auth *graph.Auth
	// FS is the filesystem object
	FS interface{}
	// IsMounted indicates whether the filesystem is currently mounted
	IsMounted bool
	// Additional data for the test
	Data map[string]interface{}
}

// SetupSystemTestFixture creates a test fixture for full end-to-end system tests.
// This fixture is designed for system tests that require complete mounting and real OneDrive access.
//
// Features:
//   - Full end-to-end setup with FUSE mounting
//   - Uses real auth tokens and real OneDrive
//   - Tests complete user workflows
//   - Requires FUSE device access
//
// Requirements:
//   - Auth tokens must be available in test-artifacts/.auth_tokens.json
//   - FUSE device must be accessible (/dev/fuse)
//   - Network connectivity required
//   - Root or appropriate permissions for mounting
//   - Should be run in Docker container for isolation
//
// Usage:
//
//	fixture := helpers.SetupSystemTestFixture(t, "MySystemTest", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
//	    return fs.NewFilesystem(auth, mountPoint, cacheTTL)
//	})
//	defer fixture.Teardown(t)
//
//	// Access the fixture data
//	sysFixture := fixture.GetFixture(t).(*SystemTestFixture)
//	// ... test code with mounted filesystem ...
func SetupSystemTestFixture(t *testing.T, fixtureName string, newFilesystem func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error)) *framework.UnitTestFixture {
	fixture := framework.NewUnitTestFixture(fixtureName)

	// Set up the fixture with full mounting
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Ensure we're in online mode for test setup
		graph.SetOperationalOffline(false)

		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "onemount-system-test-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create a temporary directory: %w", err)
		}

		// Create a mount point directory
		mountPoint := filepath.Join(tempDir, "mount")
		if err := os.MkdirAll(mountPoint, 0755); err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("failed to create mount point: %w", err)
		}

		// Load real authentication tokens
		auth := GetTestAuth()

		// Verify we have real tokens (not mock)
		if auth.AccessToken == "mock-access-token" {
			// Clean up temp directory
			os.RemoveAll(tempDir)
			t.Skip("Skipping system test: real auth tokens not available")
			return nil, fmt.Errorf("real auth tokens required for system tests")
		}

		// Check if FUSE device is available
		if _, err := os.Stat("/dev/fuse"); os.IsNotExist(err) {
			os.RemoveAll(tempDir)
			t.Skip("Skipping system test: FUSE device not available")
			return nil, fmt.Errorf("FUSE device required for system tests")
		}

		// Create the filesystem with real OneDrive connection
		fs, err := newFilesystem(auth, mountPoint, 30)
		if err != nil {
			// Clean up the temporary directory if filesystem creation fails
			if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
				t.Logf("Warning: Failed to clean up temporary directory %s: %v", tempDir, cleanupErr)
			}
			return nil, fmt.Errorf("failed to create filesystem for system test: %w", err)
		}

		// Create the fixture
		sysFixture := &SystemTestFixture{
			TempDir:    tempDir,
			MountPoint: mountPoint,
			Auth:       auth,
			FS:         fs,
			IsMounted:  false, // Will be set to true after mounting
			Data:       make(map[string]interface{}),
		}

		// Note: Actual mounting would be done by the test code if needed
		// This fixture just sets up the environment

		return sysFixture, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		sysFixture := fixture.(*SystemTestFixture)

		// Ensure we reset to online mode after the test
		graph.SetOperationalOffline(false)

		// Unmount if still mounted
		if sysFixture.IsMounted {
			t.Logf("Unmounting filesystem at %s", sysFixture.MountPoint)
			// Try to unmount using fusermount3
			cmd := exec.Command("fusermount3", "-uz", sysFixture.MountPoint)
			if err := cmd.Run(); err != nil {
				t.Logf("Warning: Failed to unmount filesystem: %v", err)
			}
			// Wait a bit for unmount to complete
			time.Sleep(100 * time.Millisecond)
		}

		// Stop the filesystem if it has a Stop method (to prevent goroutine leaks)
		if sysFixture.FS != nil {
			fsValue := reflect.ValueOf(sysFixture.FS)
			if fsValue.IsValid() && !fsValue.IsNil() {
				stopMethod := fsValue.MethodByName("Stop")
				if stopMethod.IsValid() {
					t.Logf("Calling Stop() on filesystem to clean up background goroutines")
					stopMethod.Call(nil)
				}
			}
		}

		// Clean up the temporary directory
		if err := os.RemoveAll(sysFixture.TempDir); err != nil {
			t.Logf("Warning: Failed to clean up temporary directory %s: %v", sysFixture.TempDir, err)
			return err
		}
		return nil
	})

	return fixture
}
