package fs

import (
	"github.com/auriora/onemount/pkg/testutil/framework"
	"github.com/auriora/onemount/pkg/testutil/helpers"
	"testing"

	"github.com/auriora/onemount/pkg/graph"
)

// TestIT_FS_01_01_Cache_BasicOperations_WorkCorrectly tests various cache operations.
//
//	Test Case ID    IT-FS-01-01
//	Title           Cache Operations
//	Description     Tests various cache operations
//	Preconditions   None
//	Steps           1. Create a filesystem cache
//	                2. Perform operations on the cache (get path, get children, check pointers)
//	                3. Verify the results of each operation
//	Expected Result Cache operations work correctly
//	Notes: This test verifies that the cache operations work correctly.
func TestIT_FS_01_01_Cache_BasicOperations_WorkCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CacheOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Get the test data
		fsFixture := fixture.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		// Step 1: Test basic cache operations

		// Test GetPath operation
		rootInodeByPath, err := fs.GetPath("/", fs.auth)
		assert.NoError(err, "GetPath should not return error")
		assert.NotNil(rootInodeByPath, "Root inode should exist")
		assert.Equal("/", rootInodeByPath.Path(), "Root path should be /")

		// Test GetID operation
		rootInode := fs.GetID(rootID)
		assert.NotNil(rootInode, "Root inode should exist")
		assert.Equal(rootID, rootInode.ID(), "Root inode ID should match")

		// Step 2: Test cache insertion and retrieval

		// Create a test file item
		testFileID := "test-cache-file-id"
		testFileName := "cache_test_file.txt"
		fileItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Size: 1024,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "test-hash",
				},
			},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Insert the file into the cache
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		nodeID := fs.InsertChild(rootID, fileInode)

		// Verify insertion
		assert.NotEqual(uint64(0), nodeID, "Node ID should be assigned")
		assert.Equal(nodeID, fileInode.NodeID(), "Node ID should match inode")

		// Test retrieval by ID
		retrievedInode := fs.GetID(testFileID)
		assert.NotNil(retrievedInode, "File should be retrievable by ID")
		assert.Equal(testFileID, retrievedInode.ID(), "Retrieved inode ID should match")
		assert.Equal(testFileName, retrievedInode.Name(), "Retrieved inode name should match")

		// Test retrieval by NodeID
		retrievedByNodeID := fs.GetNodeID(nodeID)
		assert.NotNil(retrievedByNodeID, "File should be retrievable by NodeID")
		assert.Equal(testFileID, retrievedByNodeID.ID(), "Retrieved inode ID should match")

		// Step 3: Test GetChild operation
		childInode, err := fs.GetChild(rootID, testFileName, fs.auth)
		assert.NoError(err, "GetChild should not return error")
		assert.NotNil(childInode, "Child should be found")
		assert.Equal(testFileID, childInode.ID(), "Child ID should match")

		// Step 4: Test GetChildrenID operation
		children, err := fs.GetChildrenID(rootID, fs.auth)
		assert.NoError(err, "GetChildrenID should not return error")
		assert.NotNil(children, "Children map should not be nil")
		assert.Contains(children, testFileName, "Children should contain our test file")
		assert.Equal(testFileID, children[testFileName].ID(), "Child in map should have correct ID")

		// Step 5: Test path operations
		expectedPath := "/" + testFileName
		assert.Equal(expectedPath, fileInode.Path(), "File path should be correct")

		// Test that the file can be found by path
		foundInode, err := fs.GetPath(expectedPath, fs.auth)
		assert.NoError(err, "GetPath should find the file")
		assert.NotNil(foundInode, "Found inode should not be nil")
		assert.Equal(testFileID, foundInode.ID(), "Found inode ID should match")

		// Step 6: Test cache cleanup and management

		// Test that cache pointers are working correctly
		assert.Equal(fileInode, fs.GetNodeID(nodeID), "Node ID pointer should be consistent")
		assert.Equal(fileInode, fs.GetID(testFileID), "ID pointer should be consistent")
	})
}
