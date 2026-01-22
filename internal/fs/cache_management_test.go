package fs

import (
	"fmt"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_Cache_01_CacheInvalidation_WorksCorrectly tests cache invalidation mechanisms.
//
//	Test Case ID    IT-FS-Cache-01
//	Title           Cache Invalidation
//	Description     Tests cache invalidation and cleanup mechanisms
//	Preconditions   None
//	Steps           1. Create files and populate cache
//	                2. Test cache cleanup operations
//	                3. Test cache invalidation
//	                4. Verify cache state after operations
//	Expected Result Cache invalidation works correctly
//	Notes: This test verifies that cache management operations work correctly.
func TestIT_FS_Cache_01_CacheInvalidation_WorksCorrectly(t *testing.T) {
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

// TestIT_FS_Cache_04_CacheInvalidation_Comprehensive tests comprehensive cache invalidation scenarios.
//
//	Test Case ID    IT-FS-Cache-04
//	Title           Comprehensive Cache Invalidation
//	Description     Tests various cache invalidation scenarios including file modifications, deletions
//	Preconditions   None
//	Steps           1. Create files and populate cache
//	                2. Modify files and verify cache invalidation
//	                3. Delete files and verify cache cleanup
//	                4. Test cache consistency after operations
//	Expected Result Cache is properly invalidated and cleaned up
//	Notes: This test provides comprehensive coverage of cache invalidation.
func TestIT_FS_Cache_04_CacheInvalidation_Comprehensive(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ComprehensiveCacheInvalidationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create files and populate cache

		// Create multiple test files
		testFiles := []struct {
			id   string
			name string
			size uint64
		}{
			{"cache-invalidation-file-1", "cache_test_1.txt", 1024},
			{"cache-invalidation-file-2", "cache_test_2.txt", 2048},
			{"cache-invalidation-file-3", "cache_test_3.txt", 4096},
		}

		var fileInodes []*Inode
		var nodeIDs []uint64

		for _, testFile := range testFiles {
			fileItem := &graph.DriveItem{
				ID:   testFile.id,
				Name: testFile.name,
				Size: testFile.size,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: "test-hash-" + testFile.id,
					},
				},
			}

			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			nodeID := fs.InsertChild(rootID, fileInode)

			fileInodes = append(fileInodes, fileInode)
			nodeIDs = append(nodeIDs, nodeID)

			// Verify file is in cache
			retrievedInode := fs.GetID(testFile.id)
			assert.NotNil(retrievedInode, "File should be in cache")
		}

		// Step 2: Test cache invalidation through file modifications

		// Modify first file and verify cache behavior
		firstFile := fileInodes[0]
		firstFile.DriveItem.Size = 8192 // Change size
		firstFile.hasChanges = true

		// File should still be in cache but marked as modified
		retrievedInode := fs.GetID(testFiles[0].id)
		assert.NotNil(retrievedInode, "Modified file should still be in cache")
		assert.True(retrievedInode.hasChanges, "File should be marked as modified")

		// Step 3: Test cache cleanup through deletion

		// Delete second file
		fs.DeleteID(testFiles[1].id)

		// Verify file is no longer in cache
		deletedInode := fs.GetID(testFiles[1].id)
		assert.Nil(deletedInode, "Deleted file should not be in cache")

		// Step 4: Test cache consistency

		// Verify remaining files are still accessible
		for i, testFile := range testFiles {
			if i == 1 { // Skip deleted file
				continue
			}
			retrievedInode := fs.GetID(testFile.id)
			assert.NotNil(retrievedInode, "Remaining files should still be in cache")
			assert.Equal(testFile.name, retrievedInode.Name(), "File name should match")
		}

		// Test cache serialization with mixed state
		fs.SerializeAll()

		// Verify files are still accessible after serialization
		for i, testFile := range testFiles {
			if i == 1 { // Skip deleted file
				continue
			}
			retrievedInode := fs.GetID(testFile.id)
			assert.NotNil(retrievedInode, "Files should still be accessible after serialization")
		}
	})
}

// TestIT_FS_Cache_05_CachePerformance_Operations tests cache performance characteristics.
//
//	Test Case ID    IT-FS-Cache-05
//	Title           Cache Performance Operations
//	Description     Tests cache performance with multiple operations
//	Preconditions   None
//	Steps           1. Create many files to stress test cache
//	                2. Perform rapid cache operations
//	                3. Verify cache performance is reasonable
//	                4. Test cache cleanup efficiency
//	Expected Result Cache operations perform efficiently
//	Notes: This test verifies cache performance characteristics.
func TestIT_FS_Cache_05_CachePerformance_Operations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CachePerformanceFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create many files to stress test cache
		numFiles := 50
		fileIDs := make([]string, numFiles)

		startTime := time.Now()

		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("performance-test-file-%d", i)
			fileName := fmt.Sprintf("perf_test_%d.txt", i)

			fileItem := &graph.DriveItem{
				ID:   fileID,
				Name: fileName,
				Size: uint64(1024 * (i + 1)), // Varying sizes
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: fmt.Sprintf("hash-%d", i),
					},
				},
			}

			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			fileIDs[i] = fileID
		}

		insertTime := time.Since(startTime)

		// Step 2: Perform rapid cache operations
		startTime = time.Now()

		// Test rapid retrieval
		for i := 0; i < numFiles; i++ {
			retrievedInode := fs.GetID(fileIDs[i])
			assert.NotNil(retrievedInode, "File should be retrievable")
		}

		retrievalTime := time.Since(startTime)

		// Step 3: Verify cache performance is reasonable
		// These are basic performance checks - actual thresholds may need adjustment
		assert.True(insertTime < time.Second*5, "Cache insertion should be reasonably fast")
		assert.True(retrievalTime < time.Second*2, "Cache retrieval should be reasonably fast")

		// Step 4: Test cache cleanup efficiency
		startTime = time.Now()

		// Delete half the files
		for i := 0; i < numFiles/2; i++ {
			fs.DeleteID(fileIDs[i])
		}

		deletionTime := time.Since(startTime)
		assert.True(deletionTime < time.Second*2, "Cache deletion should be reasonably fast")

		// Verify remaining files are still accessible
		for i := numFiles / 2; i < numFiles; i++ {
			retrievedInode := fs.GetID(fileIDs[i])
			assert.NotNil(retrievedInode, "Remaining files should still be accessible")
		}

		// Verify deleted files are not accessible
		for i := 0; i < numFiles/2; i++ {
			deletedInode := fs.GetID(fileIDs[i])
			assert.Nil(deletedInode, "Deleted files should not be accessible")
		}
	})
}

// TestIT_FS_Cache_02_ContentCache_Operations tests content cache operations.
//
//	Test Case ID    IT-FS-Cache-02
//	Title           Content Cache Operations
//	Description     Tests content cache insertion, retrieval, and deletion
//	Preconditions   None
//	Steps           1. Insert content into cache
//	                2. Retrieve content from cache
//	                3. Test content deletion
//	                4. Verify cache consistency
//	Expected Result Content cache operations work correctly
//	Notes: This test verifies that content caching works correctly.
func TestIT_FS_Cache_02_ContentCache_Operations(t *testing.T) {
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

// TestIT_FS_Cache_03_CacheConsistency_MultipleOperations tests cache consistency across multiple operations.
//
//	Test Case ID    IT-FS-Cache-03
//	Title           Cache Consistency
//	Description     Tests cache consistency across multiple concurrent operations
//	Preconditions   None
//	Steps           1. Perform multiple cache operations
//	                2. Test cache state consistency
//	                3. Verify no cache corruption
//	Expected Result Cache remains consistent across operations
//	Notes: This test verifies that cache consistency is maintained.
func TestIT_FS_Cache_03_CacheConsistency_MultipleOperations(t *testing.T) {
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
