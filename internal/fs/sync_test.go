package fs

import (
	"os"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT01_SyncDirectoryTree verifies that the filesystem can successfully synchronize
// the directory tree from the root.
//
//	Test Case ID    UT-01
//	Title           Directory Tree Synchronization
//	Description     Verify that the filesystem can successfully synchronize the directory tree from the root
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Initialize the filesystem
//	                2. Call SyncDirectoryTree method
//	                3. Verify all directories are cached
//	Expected Result All directory metadata is successfully cached without errors
//	Notes: Directly tests the SyncDirectoryTree function to verify directory tree synchronization.
func TestUT01_SyncDirectoryTree(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Step 1: Initialize the filesystem

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "onemount-test-*")
	require.NoError(t, err, "Failed to create temporary directory")

	// Register cleanup function
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to clean up temporary directory %s: %v", tempDir, err)
		}
	})

	// Create a mock graph client
	mockClient := graph.NewMockGraphClient()

	// Set up the mock directory structure
	rootID := "root-id"
	rootItem := &graph.DriveItem{
		ID:   rootID,
		Name: "root",
		Folder: &graph.Folder{
			ChildCount: 2,
		},
	}

	// Add the root item to the mock client
	mockClient.AddMockItem("/me/drive/root", rootItem)

	// Create child items
	child1ID := "child1-id"
	child1Item := &graph.DriveItem{
		ID:   child1ID,
		Name: "child1",
		Folder: &graph.Folder{
			ChildCount: 1,
		},
	}

	child2ID := "child2-id"
	child2Item := &graph.DriveItem{
		ID:   child2ID,
		Name: "child2",
		Folder: &graph.Folder{
			ChildCount: 0,
		},
	}

	// Add child items to the mock client
	mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{child1Item, child2Item})

	// Create grandchild item
	grandchildID := "grandchild-id"
	grandchildItem := &graph.DriveItem{
		ID:   grandchildID,
		Name: "grandchild",
		Folder: &graph.Folder{
			ChildCount: 0,
		},
	}

	// Add grandchild item to the mock client
	mockClient.AddMockItems("/me/drive/items/"+child1ID+"/children", []*graph.DriveItem{grandchildItem})

	// Add empty children for the leaf nodes
	mockClient.AddMockItems("/me/drive/items/"+child2ID+"/children", []*graph.DriveItem{})
	mockClient.AddMockItems("/me/drive/items/"+grandchildID+"/children", []*graph.DriveItem{})

	// Create a mock auth object
	auth := &graph.Auth{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		Account:      "mock@example.com",
	}

	// Create the filesystem
	fs, err := NewFilesystem(auth, tempDir, 30)
	require.NoError(t, err, "Failed to create filesystem")

	// Set the root ID
	fs.root = rootID

	// Step 2: Call SyncDirectoryTree method
	err = fs.SyncDirectoryTree(auth)
	require.NoError(t, err, "SyncDirectoryTree failed")

	// Step 3: Verify all directories are cached

	// Verify root directory
	rootInode := fs.GetID(rootID)
	require.NotNil(t, rootInode, "Root directory not found in cache")
	assert.Equal(t, "root", rootInode.Name(), "Root directory name mismatch")

	// Verify child directories
	children, err := fs.GetChildrenID(rootID, auth)
	require.NoError(t, err, "Failed to get children of root directory")
	assert.Equal(t, 2, len(children), "Root directory should have 2 children")

	// Verify child1 directory
	child1Inode := fs.GetID(child1ID)
	require.NotNil(t, child1Inode, "Child1 directory not found in cache")
	assert.Equal(t, "child1", child1Inode.Name(), "Child1 directory name mismatch")

	// Verify child2 directory
	child2Inode := fs.GetID(child2ID)
	require.NotNil(t, child2Inode, "Child2 directory not found in cache")
	assert.Equal(t, "child2", child2Inode.Name(), "Child2 directory name mismatch")

	// Verify grandchild directory
	grandchildInode := fs.GetID(grandchildID)
	require.NotNil(t, grandchildInode, "Grandchild directory not found in cache")
	assert.Equal(t, "grandchild", grandchildInode.Name(), "Grandchild directory name mismatch")

	// Verify child1 has one child
	child1Children, err := fs.GetChildrenID(child1ID, auth)
	require.NoError(t, err, "Failed to get children of child1 directory")
	assert.Equal(t, 1, len(child1Children), "Child1 directory should have 1 child")

	// Verify child2 has no children
	child2Children, err := fs.GetChildrenID(child2ID, auth)
	require.NoError(t, err, "Failed to get children of child2 directory")
	assert.Equal(t, 0, len(child2Children), "Child2 directory should have no children")

	// Verify grandchild has no children
	grandchildChildren, err := fs.GetChildrenID(grandchildID, auth)
	require.NoError(t, err, "Failed to get children of grandchild directory")
	assert.Equal(t, 0, len(grandchildChildren), "Grandchild directory should have no children")
}
