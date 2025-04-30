package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/bcherrington/onemount/internal/testutil/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSyncDirectoryTree tests the SyncDirectoryTree function (TC-01)
// This test verifies that the filesystem can successfully synchronize the directory tree from the root
func TestSyncDirectoryTree(t *testing.T) {
	// Skip if using mock auth since we need real directory structure
	// TODO: In the future, this should be updated to use the interface-based approach
	// instead of checking the environment variable directly
	if os.Getenv("ONEMOUNT_MOCK_AUTH") == "1" {
		t.Skip("Skipping test with mock authentication")
	}

	// Create a test directory structure if it doesn't exist
	testDirPath := filepath.Join(mountLoc, "onemount_sync_test")
	subDirPath := filepath.Join(testDirPath, "subdir")
	subSubDirPath := filepath.Join(subDirPath, "subsubdir")

	// Create the test directories
	t.Log("Creating test directory structure")
	createTestDir(t, testDirPath)
	createTestDir(t, subDirPath)
	createTestDir(t, subSubDirPath)

	// Setup cleanup to remove test directories after test completes
	t.Cleanup(func() {
		t.Log("Cleaning up test directories")
		removeTestDir(t, subSubDirPath)
		removeTestDir(t, subDirPath)
		removeTestDir(t, testDirPath)
	})

	// Get the root directory ID
	rootItem, err := graph.GetItemPath("/", auth)
	require.NoError(t, err, "Failed to get root directory")
	require.NotNil(t, rootItem, "Root directory is nil")

	// Clear the filesystem metadata cache to ensure we're testing actual synchronization
	// Create a new empty sync.Map and assign it to fs.metadata
	fs.metadata = sync.Map{}

	// Call SyncDirectoryTree
	t.Log("Starting directory tree synchronization")
	err = fs.SyncDirectoryTree(auth)
	require.NoError(t, err, "SyncDirectoryTree failed")

	// Verify that the test directories are cached in the filesystem metadata
	verifyDirectoryCached(t, "/onemount_sync_test")
	verifyDirectoryCached(t, "/onemount_sync_test/subdir")
	verifyDirectoryCached(t, "/onemount_sync_test/subdir/subsubdir")

	// Verify that other known directories are also cached
	verifyDirectoryCached(t, "/OneMount-Documents")
	verifyDirectoryCached(t, "/onemount_tests")
}

// Helper function to create a test directory
func createTestDir(t *testing.T, path string) {
	// Check if directory already exists
	if _, err := os.Stat(path); err == nil {
		return // Directory already exists
	}

	// Create the directory
	err := os.MkdirAll(path, 0755)
	require.NoError(t, err, "Failed to create directory: %s", path)

	// Wait for the directory to be recognized by the filesystem
	common.WaitForCondition(t, func() bool {
		fsPath := strings.TrimPrefix(path, mountLoc)
		if fsPath == "" {
			fsPath = "/"
		}
		_, err := fs.GetPath(fsPath, auth)
		return err == nil
	}, 10*time.Second, 500*time.Millisecond, fmt.Sprintf("Directory was not recognized by filesystem: %s", path))
}

// Helper function to remove a test directory
func removeTestDir(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Logf("Directory already removed or doesn't exist: %s", path)
		return
	}

	err := os.RemoveAll(path)
	if err != nil {
		t.Logf("Failed to remove directory: %s, error: %v", path, err)
	}
}

// Helper function to verify that a directory is cached in the filesystem metadata
func verifyDirectoryCached(t *testing.T, path string) {
	t.Logf("Verifying directory is cached: %s", path)

	// Get the directory from the filesystem
	dir, err := fs.GetPath(path, nil) // Pass nil auth to ensure we're using cached data
	assert.NoError(t, err, "Failed to get directory from cache: %s", path)
	assert.NotNil(t, dir, "Directory is nil: %s", path)
	assert.True(t, dir.IsDir(), "Not a directory: %s", path)
}
