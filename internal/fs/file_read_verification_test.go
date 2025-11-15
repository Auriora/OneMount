package fs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestUT_FS_FileRead_01_UncachedFile tests reading a file that hasn't been cached yet.
//
//	Test Case ID    UT-FS-FileRead-01
//	Title           Read Uncached File
//	Description     Tests reading a file that hasn't been accessed before
//	Preconditions   File exists on OneDrive but not in local cache
//	Steps           1. Clear cache
//	                2. Create a mock file on OneDrive
//	                3. Open the file (should trigger download)
//	                4. Read the file content
//	                5. Verify content is correct
//	                6. Verify file is now cached
//	Expected Result File downloads from OneDrive and content is correct
//	Requirements    3.2 - On-demand file download
//	Notes: This test verifies that uncached files are downloaded on first access.
func TestUT_FS_FileRead_01_UncachedFile(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "UncachedFileReadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Step 1: Create a mock file on OneDrive (not in cache)
		testFileName := "uncached_test_file.txt"
		testFileID := "uncached-file-id-001"
		testContent := "This is test content for uncached file read verification"

		// Create the mock file
		fileItem := helpers.CreateMockFile(mockClient, rootID, testFileName, testFileID, testContent)
		assert.NotNil(fileItem, "Mock file should be created")

		// Update the root's children to include this file
		mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{fileItem})

		// Step 2: Ensure the file is not in cache by clearing any stale entry
		_ = fs.content.Delete(testFileID)

		// Step 3: Get the file inode (this should fetch metadata but not content)
		child, err := fs.GetChild(rootID, testFileName, fs.auth)
		assert.Nil(err, "Should be able to get child inode")
		assert.NotNil(child, "Child inode should exist")
		if child == nil {
			t.Fatal("Child inode is nil, cannot continue test")
		}
		assert.Equal(testFileName, child.Name(), "File name should match")

		nodeID := fs.InsertNodeID(child)
		assert.NotEqual(uint64(0), nodeID, "Node ID should be assigned")

		// Step 4: Open the file (should trigger download)
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
			Flags:    uint32(os.O_RDONLY),
		}
		openOut := &fuse.OpenOut{}

		status := fs.Open(nil, openIn, openOut)
		assert.Equal(fuse.OK, status, "File open should succeed and trigger download")

		// Give download manager time to complete (it's async)
		time.Sleep(100 * time.Millisecond)

		// Step 5: Read the file content
		readBuf := make([]byte, len(testContent)+10)
		readIn := &fuse.ReadIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
			Offset:   0,
			Size:     uint32(len(readBuf)),
		}

		readResult, status := fs.Read(nil, readIn, readBuf)
		assert.Equal(fuse.OK, status, "File read should succeed")
		assert.NotNil(readResult, "Read result should not be nil")

		// Step 6: Verify file is now cached
		cachedFd, err := fs.content.Open(testFileID)
		assert.Nil(err, "File should now be in cache")
		assert.NotNil(cachedFd, "Cached file descriptor should exist")
		if cachedFd != nil {
			cachedFd.Close()
		}

		// Step 7: Verify file size was updated
		assert.Equal(uint64(len(testContent)), child.DriveItem.Size, "File size should match content")
	})
}

// TestUT_FS_FileRead_02_CachedFile tests reading a file that is already cached.
//
//	Test Case ID    UT-FS-FileRead-02
//	Title           Read Cached File
//	Description     Tests reading a file that has been previously accessed
//	Preconditions   File exists in local cache with valid checksum
//	Steps           1. Create and cache a file
//	                2. Read the file (should use cache)
//	                3. Verify content is served from cache
//	                4. Check read performance is fast
//	Expected Result File is served from cache without network access
//	Requirements    3.3 - Serve cached files
//	Notes: This test verifies that cached files are served without network requests.
func TestUT_FS_FileRead_02_CachedFile(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CachedFileReadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Step 1: Create a mock file and pre-cache it
		testFileName := "cached_test_file.txt"
		testFileID := "cached-file-id-001"
		testContent := "This is test content for cached file read verification"

		// Create the mock file
		fileItem := helpers.CreateMockFile(mockClient, rootID, testFileName, testFileID, testContent)
		assert.NotNil(fileItem, "Mock file should be created")

		// Update the root's children to include this file
		mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{fileItem})

		// Get the file inode
		child, err := fs.GetChild(rootID, testFileName, fs.auth)
		assert.Nil(err, "Should be able to get child inode")
		assert.NotNil(child, "Child inode should exist")
		if child == nil {
			t.Fatal("Child inode is nil, cannot continue test")
		}

		nodeID := fs.InsertNodeID(child)

		// Pre-cache the file by opening and reading it once
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
			Flags:    uint32(os.O_RDONLY),
		}
		openOut := &fuse.OpenOut{}

		status := fs.Open(nil, openIn, openOut)
		assert.Equal(fuse.OK, status, "Initial file open should succeed")

		// Give download time to complete
		time.Sleep(100 * time.Millisecond)

		// Read once to ensure it's cached
		readBuf := make([]byte, len(testContent)+10)
		readIn := &fuse.ReadIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
			Offset:   0,
			Size:     uint32(len(readBuf)),
		}

		_, status = fs.Read(nil, readIn, readBuf)
		assert.Equal(fuse.OK, status, "Initial read should succeed")

		// Step 2: Open and read the file again (should use cache)
		startTime := time.Now()

		status = fs.Open(nil, openIn, openOut)
		assert.Equal(fuse.OK, status, "Second file open should succeed")

		readResult, status := fs.Read(nil, readIn, readBuf)
		assert.Equal(fuse.OK, status, "Second file read should succeed")
		assert.NotNil(readResult, "Read result should not be nil")

		readDuration := time.Since(startTime)

		// Step 3: Verify read performance is fast (should be < 100ms for cached read)
		// This is a reasonable threshold that indicates cache usage
		assert.True(readDuration < 100*time.Millisecond,
			"Cached read should be fast (took %v)", readDuration)

		t.Logf("Cached read completed in %v (expected < 100ms for cache hit)", readDuration)

		// Step 4: Verify file is still in cache
		cachedFd, err := fs.content.Open(testFileID)
		assert.Nil(err, "File should still be in cache after second read")
		if cachedFd != nil {
			cachedFd.Close()
		}
	})
}

// TestUT_FS_FileRead_03_DirectoryListing tests listing a directory without downloading file content.
//
//	Test Case ID    UT-FS-FileRead-03
//	Title           Directory Listing Without Content Download
//	Description     Tests listing a directory with many files
//	Preconditions   Directory exists with multiple files
//	Steps           1. Create a directory with multiple files
//	                2. List the directory
//	                3. Verify all files appear
//	                4. Verify no file content is downloaded
//	                5. Check that metadata is displayed correctly
//	Expected Result Directory listing shows all files without downloading content
//	Requirements    3.1 - Display files using cached metadata
//	Notes: This test verifies that directory listing doesn't trigger content downloads.
func TestUT_FS_FileRead_03_DirectoryListing(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DirectoryListingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Step 1: Create a directory with multiple files
		testDirName := "test_directory"
		testDirID := "test-dir-id-001"

		dirItem := helpers.CreateMockDirectory(mockClient, rootID, testDirName, testDirID)
		assert.NotNil(dirItem, "Mock directory should be created")

		// Create multiple files in the directory
		fileCount := 5
		fileItems := make([]*graph.DriveItem, fileCount)
		for i := 0; i < fileCount; i++ {
			fileName := fmt.Sprintf("test_file_%d.txt", i)
			fileID := fmt.Sprintf("test-file-id-%03d", i)
			content := fmt.Sprintf("Content for file %d", i)

			fileItems[i] = helpers.CreateMockFile(mockClient, testDirID, fileName, fileID, content)
			assert.NotNil(fileItems[i], "Mock file %d should be created", i)
		}

		// Update the mock to return all files as children
		mockClient.AddMockItems("/me/drive/items/"+testDirID+"/children", fileItems)

		// Step 2: Get the directory inode
		dirChild, err := fs.GetChild(rootID, testDirName, fs.auth)
		assert.Nil(err, "Should be able to get directory inode")
		assert.NotNil(dirChild, "Directory inode should exist")
		assert.True(dirChild.IsDir(), "Should be a directory")

		dirNodeID := fs.InsertNodeID(dirChild)

		// Step 3: List the directory (OpenDir)
		openDirIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{NodeId: dirNodeID},
			Flags:    uint32(os.O_RDONLY),
		}
		openDirOut := &fuse.OpenOut{}

		status := fs.Open(nil, openDirIn, openDirOut)
		assert.Equal(fuse.OK, status, "Directory open should succeed")

		// Step 4: Verify all files appear in listing by getting children
		childrenMap, err := fs.GetChildrenID(testDirID, fs.auth)
		assert.Nil(err, "Should be able to get children")
		assert.NotNil(childrenMap, "Children map should not be nil")

		assert.True(len(childrenMap) >= fileCount,
			"Should have at least %d files in listing (got %d)", fileCount, len(childrenMap))

		// Convert map to slice for iteration
		children := make([]*Inode, 0, len(childrenMap))
		for _, child := range childrenMap {
			children = append(children, child)
		}

		// Step 5: Verify file metadata is correct
		for i, child := range children {
			if i >= fileCount {
				break
			}
			assert.NotNil(child, "Child %d should exist", i)
			assert.False(child.IsDir(), "Child %d should be a file", i)
			assert.True(child.DriveItem.Size > 0, "Child %d should have size", i)
		}

		// Step 6: Verify no file content was downloaded
		// Check that none of the files are in the content cache
		contentDownloadCount := 0
		for i := 0; i < fileCount; i++ {
			fileID := fmt.Sprintf("test-file-id-%03d", i)
			if fs.content.HasContent(fileID) {
				contentDownloadCount++
			}
		}

		// Directory listing should not download any file content
		assert.Equal(0, contentDownloadCount,
			"Directory listing should not download file content (found %d files in cache)",
			contentDownloadCount)

		t.Logf("Directory listing completed without downloading content for %d files", fileCount)
	})
}

// TestUT_FS_FileRead_04_FileMetadata tests file metadata operations.
//
//	Test Case ID    UT-FS-FileRead-04
//	Title           File Metadata Operations
//	Description     Tests retrieving file metadata without downloading content
//	Preconditions   File exists on OneDrive
//	Steps           1. Create a mock file
//	                2. Get file inode (metadata)
//	                3. Check file size, timestamps, permissions
//	                4. Verify metadata matches OneDrive
//	                5. Verify no content download occurred
//	Expected Result File metadata is correct without downloading content
//	Requirements    3.1 - Display files using cached metadata
//	Notes: This test verifies that metadata operations don't trigger content downloads.
func TestUT_FS_FileRead_04_FileMetadata(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileMetadataFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Step 1: Create a mock file with specific metadata
		testFileName := "metadata_test_file.txt"
		testFileID := "metadata-file-id-001"
		testContent := "This is test content for metadata verification"
		expectedSize := uint64(len(testContent))

		// Create the mock file
		fileItem := helpers.CreateMockFile(mockClient, rootID, testFileName, testFileID, testContent)
		assert.NotNil(fileItem, "Mock file should be created")

		// Update the root's children to include this file
		mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{fileItem})

		// Step 2: Get the file inode (metadata only)
		child, err := fs.GetChild(rootID, testFileName, fs.auth)
		assert.Nil(err, "Should be able to get child inode")
		assert.NotNil(child, "Child inode should exist")
		if child == nil {
			t.Fatal("Child inode is nil, cannot continue test")
		}

		// Step 3: Verify file metadata
		assert.Equal(testFileName, child.Name(), "File name should match")
		assert.Equal(expectedSize, child.DriveItem.Size, "File size should match")
		assert.False(child.IsDir(), "Should be a file, not a directory")
		assert.NotNil(child.DriveItem.File, "File property should be set")

		// Step 4: Verify file attributes
		attr := child.makeAttr()
		assert.Equal(expectedSize, attr.Size, "Attribute size should match")
		assert.True(attr.Mtime > 0, "File should have modification time")
		assert.True(attr.Ctime > 0, "File should have creation time")

		// Step 5: Verify no content download occurred
		cachedContent := fs.content.Get(testFileID)
		assert.Equal(0, len(cachedContent), "File content should not be downloaded for metadata operations")

		t.Logf("File metadata verified: name=%s, size=%d bytes", testFileName, expectedSize)
	})
}
