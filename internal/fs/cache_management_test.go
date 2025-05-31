package fs

import (
	"fmt"
	"testing"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/testutil/framework"
	"github.com/auriora/onemount/pkg/testutil/helpers"
)

// TestUT_FS_Cache_01_CacheInvalidation_WorksCorrectly tests cache invalidation mechanisms.
//
//	Test Case ID    UT-FS-Cache-01
//	Title           Cache Invalidation
//	Description     Tests cache invalidation and cleanup mechanisms
//	Preconditions   None
//	Steps           1. Create files and populate cache
//	                2. Test cache cleanup operations
//	                3. Test cache invalidation
//	                4. Verify cache state after operations
//	Expected Result Cache invalidation works correctly
//	Notes: This test verifies that cache management operations work correctly.
func TestUT_FS_Cache_01_CacheInvalidation_WorksCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CacheInvalidationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem with a short cache TTL for testing
		fs, err := NewFilesystem(auth, mountPoint, 1) // 1 second TTL
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
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		// Step 1: Create files and populate cache

		// Create multiple test files
		testFiles := []struct {
			id   string
			name string
		}{
			{"cache-test-file-1", "cache_file_1.txt"},
			{"cache-test-file-2", "cache_file_2.txt"},
			{"cache-test-file-3", "cache_file_3.txt"},
		}

		for _, tf := range testFiles {
			fileItem := &graph.DriveItem{
				ID:   tf.id,
				Name: tf.name,
				Size: 1024,
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: "test-hash-" + tf.id,
					},
				},
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
			}

			// Insert the file into the cache
			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			// Verify file is in cache
			retrievedInode := fs.GetID(tf.id)
			assert.NotNil(retrievedInode, "File %s should be in cache", tf.name)
			assert.Equal(tf.id, retrievedInode.ID(), "File ID should match")
		}

		// Step 2: Test cache cleanup operations

		// Start cache cleanup
		fs.StartCacheCleanup()

		// Wait a moment for cleanup to potentially run
		time.Sleep(100 * time.Millisecond)

		// Verify files are still accessible (they shouldn't be cleaned up immediately)
		for _, tf := range testFiles {
			retrievedInode := fs.GetID(tf.id)
			assert.NotNil(retrievedInode, "File %s should still be in cache after cleanup start", tf.name)
		}

		// Step 3: Test manual cache operations

		// Test DeleteID operation
		firstFileID := testFiles[0].id
		fs.DeleteID(firstFileID)

		// Verify file is removed from cache
		deletedInode := fs.GetID(firstFileID)
		assert.Nil(deletedInode, "Deleted file should not be in cache")

		// Verify other files are still in cache
		for i := 1; i < len(testFiles); i++ {
			retrievedInode := fs.GetID(testFiles[i].id)
			assert.NotNil(retrievedInode, "Other files should still be in cache")
		}

		// Step 4: Test cache serialization

		// Test SerializeAll operation
		fs.SerializeAll()
		// SerializeAll doesn't return an error, so we just verify it doesn't panic

		// Verify files are still accessible after serialization
		for i := 1; i < len(testFiles); i++ {
			retrievedInode := fs.GetID(testFiles[i].id)
			assert.NotNil(retrievedInode, "Files should still be accessible after serialization")
		}

		// Step 5: Test cache cleanup stop

		// Stop cache cleanup
		fs.StopCacheCleanup()

		// Verify cache is still functional
		for i := 1; i < len(testFiles); i++ {
			retrievedInode := fs.GetID(testFiles[i].id)
			assert.NotNil(retrievedInode, "Files should still be accessible after cleanup stop")
		}
	})
}

// TestUT_FS_Cache_02_ContentCache_Operations tests content cache operations.
//
//	Test Case ID    UT-FS-Cache-02
//	Title           Content Cache Operations
//	Description     Tests content cache insertion, retrieval, and deletion
//	Preconditions   None
//	Steps           1. Insert content into cache
//	                2. Retrieve content from cache
//	                3. Test content deletion
//	                4. Verify cache consistency
//	Expected Result Content cache operations work correctly
//	Notes: This test verifies that content caching works correctly.
func TestUT_FS_Cache_02_ContentCache_Operations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ContentCacheFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		// Step 1: Create a file and insert content into cache

		testFileID := "content-cache-test-file"
		testFileName := "content_test.txt"
		testContent := "This is test content for cache testing."

		fileItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Size: uint64(len(testContent)),
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "test-content-hash",
				},
			},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Insert content into the content cache
		err := fs.content.Insert(testFileID, []byte(testContent))
		assert.NoError(err, "Content insertion should succeed")

		// Step 2: Retrieve content from cache

		retrievedContent := fs.content.Get(testFileID)
		assert.NotNil(retrievedContent, "Retrieved content should not be nil")
		assert.Equal(testContent, string(retrievedContent), "Retrieved content should match original")

		// Test GetInodeContent method
		inodeContent := fs.GetInodeContent(fileInode)
		assert.NotNil(inodeContent, "Inode content should not be nil")
		assert.Equal(testContent, string(*inodeContent), "Inode content should match original")

		// Step 3: Test content file operations

		// Open the content file for reading/writing
		fd, err := fs.content.Open(testFileID)
		assert.NoError(err, "Opening content file should succeed")
		assert.NotNil(fd, "File descriptor should not be nil")

		// Write additional content
		additionalContent := " Additional content."
		n, err := fd.WriteAt([]byte(additionalContent), int64(len(testContent)))
		assert.NoError(err, "Writing to content file should succeed")
		assert.Equal(len(additionalContent), n, "Should write all additional bytes")

		// Read the updated content
		totalLength := len(testContent) + len(additionalContent)
		readBuffer := make([]byte, totalLength)
		n, err = fd.ReadAt(readBuffer, 0)
		assert.NoError(err, "Reading from content file should succeed")
		assert.Equal(totalLength, n, "Should read all bytes")

		expectedContent := testContent + additionalContent
		assert.Equal(expectedContent, string(readBuffer), "Read content should match expected")

		// Close the file descriptor
		err = fd.Close()
		assert.NoError(err, "Closing content file should succeed")

		// Step 4: Test content deletion

		err = fs.content.Delete(testFileID)
		// Note: Delete might fail if the file is already closed/deleted, which is acceptable
		if err != nil {
			t.Logf("Content deletion returned error (may be expected): %v", err)
		}

		// Verify content is no longer available
		deletedContent := fs.content.Get(testFileID)
		assert.Len(deletedContent, 0, "Deleted content should be empty")

		// Step 5: Test cache consistency after deletion

		// Verify the inode still exists but content is gone
		retrievedInode := fs.GetID(testFileID)
		assert.NotNil(retrievedInode, "Inode should still exist after content deletion")
		assert.Equal(testFileID, retrievedInode.ID(), "Inode ID should still match")

		// Verify GetInodeContent returns empty for deleted content
		emptyContent := fs.GetInodeContent(retrievedInode)
		assert.NotNil(emptyContent, "GetInodeContent should return non-nil pointer")
		assert.Len(*emptyContent, 0, "Content should be empty after deletion")
	})
}

// TestUT_FS_Cache_03_CacheConsistency_MultipleOperations tests cache consistency across multiple operations.
//
//	Test Case ID    UT-FS-Cache-03
//	Title           Cache Consistency
//	Description     Tests cache consistency across multiple concurrent operations
//	Preconditions   None
//	Steps           1. Perform multiple cache operations
//	                2. Test cache state consistency
//	                3. Verify no cache corruption
//	Expected Result Cache remains consistent across operations
//	Notes: This test verifies that cache consistency is maintained.
func TestUT_FS_Cache_03_CacheConsistency_MultipleOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CacheConsistencyFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		// Step 1: Perform multiple cache operations

		// Create and insert multiple files
		numFiles := 10
		fileIDs := make([]string, numFiles)
		fileNames := make([]string, numFiles)

		for i := 0; i < numFiles; i++ {
			fileIDs[i] = fmt.Sprintf("consistency-test-file-%d", i)
			fileNames[i] = fmt.Sprintf("consistency_file_%d.txt", i)

			fileItem := &graph.DriveItem{
				ID:   fileIDs[i],
				Name: fileNames[i],
				Size: 1024,
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: fmt.Sprintf("hash-%d", i),
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
			assert.NotEqual(uint64(0), nodeID, "File %d should get valid node ID", i)
		}

		// Step 2: Test cache state consistency

		// Verify all files are accessible by ID
		for i := 0; i < numFiles; i++ {
			retrievedInode := fs.GetID(fileIDs[i])
			assert.NotNil(retrievedInode, "File %d should be retrievable by ID", i)
			assert.Equal(fileIDs[i], retrievedInode.ID(), "File %d ID should match", i)
			assert.Equal(fileNames[i], retrievedInode.Name(), "File %d name should match", i)
		}

		// Verify all files are accessible by NodeID
		for i := 0; i < numFiles; i++ {
			inode := fs.GetID(fileIDs[i])
			nodeID := inode.NodeID()
			retrievedByNodeID := fs.GetNodeID(nodeID)
			assert.NotNil(retrievedByNodeID, "File %d should be retrievable by NodeID", i)
			assert.Equal(fileIDs[i], retrievedByNodeID.ID(), "File %d ID should match when retrieved by NodeID", i)
		}

		// Step 3: Test operations that modify cache state

		// Delete every other file
		for i := 0; i < numFiles; i += 2 {
			fs.DeleteID(fileIDs[i])
		}

		// Verify deleted files are gone and remaining files are still accessible
		for i := 0; i < numFiles; i++ {
			retrievedInode := fs.GetID(fileIDs[i])
			if i%2 == 0 {
				assert.Nil(retrievedInode, "Deleted file %d should not be accessible", i)
			} else {
				assert.NotNil(retrievedInode, "Non-deleted file %d should still be accessible", i)
				assert.Equal(fileIDs[i], retrievedInode.ID(), "Non-deleted file %d ID should match", i)
			}
		}

		// Step 4: Verify cache pointers are consistent

		// Check that remaining files have consistent pointers
		for i := 1; i < numFiles; i += 2 {
			inode := fs.GetID(fileIDs[i])
			nodeID := inode.NodeID()

			// Verify the same inode is returned by both access methods
			byNodeID := fs.GetNodeID(nodeID)
			assert.Equal(inode, byNodeID, "File %d should return same inode via both access methods", i)
		}
	})
}
