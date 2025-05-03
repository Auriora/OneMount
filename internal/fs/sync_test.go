package fs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/bcherrington/onemount/internal/testutil"
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

	// Create a test fixture
	fixture := testutil.NewUnitTestFixture("SyncDirectoryTreeFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "onemount-test-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary directory: %w", err)
		}

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

		// Set operational offline mode to prevent real network requests
		graph.SetOperationalOffline(true)
		defer graph.SetOperationalOffline(false) // Reset when test is done

		// Create the filesystem
		fs, err := NewFilesystem(auth, tempDir, 30)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}

		// Set the root ID
		fs.root = rootID

		// Manually set up the root item
		rootInode := NewInodeDriveItem(rootItem)
		fs.InsertID(rootID, rootInode)

		// Insert the root item into the database to avoid the "offline and could not fetch the filesystem root item from disk" error
		fs.InsertNodeID(rootInode)

		// Return the test data
		return map[string]interface{}{
			"tempDir":      tempDir,
			"mockClient":   mockClient,
			"rootID":       rootID,
			"child1ID":     child1ID,
			"child2ID":     child2ID,
			"grandchildID": grandchildID,
			"auth":         auth,
			"fs":           fs,
		}, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Clean up the temporary directory
		data := fixture.(map[string]interface{})
		tempDir := data["tempDir"].(string)
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to clean up temporary directory %s: %v", tempDir, err)
		}
		return nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := testutil.NewAssert(t)

		// Get the test data
		data := fixture.(map[string]interface{})
		rootID := data["rootID"].(string)
		child1ID := data["child1ID"].(string)
		child2ID := data["child2ID"].(string)
		grandchildID := data["grandchildID"].(string)
		auth := data["auth"].(*graph.Auth)
		fs := data["fs"].(*Filesystem)

		// Step 2: Call SyncDirectoryTree method
		err := fs.SyncDirectoryTree(auth)
		assert.NoError(err, "SyncDirectoryTree failed")

		// Step 3: Verify all directories are cached

		// Verify root directory
		rootInode := fs.GetID(rootID)
		assert.NotNil(rootInode, "Root directory not found in cache")
		assert.Equal("root", rootInode.Name(), "Root directory name mismatch")

		// Verify child directories
		children, err := fs.GetChildrenID(rootID, auth)
		assert.NoError(err, "Failed to get children of root directory")
		assert.Equal(2, len(children), "Root directory should have 2 children")

		// Verify child1 directory
		child1Inode := fs.GetID(child1ID)
		assert.NotNil(child1Inode, "Child1 directory not found in cache")
		assert.Equal("child1", child1Inode.Name(), "Child1 directory name mismatch")

		// Verify child2 directory
		child2Inode := fs.GetID(child2ID)
		assert.NotNil(child2Inode, "Child2 directory not found in cache")
		assert.Equal("child2", child2Inode.Name(), "Child2 directory name mismatch")

		// Verify grandchild directory
		grandchildInode := fs.GetID(grandchildID)
		assert.NotNil(grandchildInode, "Grandchild directory not found in cache")
		assert.Equal("grandchild", grandchildInode.Name(), "Grandchild directory name mismatch")

		// Verify child1 has one child
		child1Children, err := fs.GetChildrenID(child1ID, auth)
		assert.NoError(err, "Failed to get children of child1 directory")
		assert.Equal(1, len(child1Children), "Child1 directory should have 1 child")

		// Verify child2 has no children
		child2Children, err := fs.GetChildrenID(child2ID, auth)
		assert.NoError(err, "Failed to get children of child2 directory")
		assert.Equal(0, len(child2Children), "Child2 directory should have no children")

		// Verify grandchild has no children
		grandchildChildren, err := fs.GetChildrenID(grandchildID, auth)
		assert.NoError(err, "Failed to get children of grandchild directory")
		assert.Equal(0, len(grandchildChildren), "Grandchild directory should have no children")
	})
}
