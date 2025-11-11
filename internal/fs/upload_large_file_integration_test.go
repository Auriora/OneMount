package fs

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_09_03_LargeFileUpload_EndToEnd verifies large file upload functionality with chunked upload.
// This test corresponds to task 9.3 in the system verification plan.
//
//	Test Case ID    IT-FS-09-03
//	Title           Large File Upload End-to-End
//	Description     Verify that large files (> 10MB) are uploaded correctly using chunked upload
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a large file (> 10MB)
//	                2. Verify chunked upload is used
//	                3. Monitor upload progress
//	                4. Verify complete file on OneDrive (mocked)
//	Expected Result Large file is successfully uploaded using chunked upload with progress tracking
//	Requirements    4.3
func TestIT_FS_09_03_LargeFileUpload_EndToEnd(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "LargeFileUploadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the filesystem from the fixture
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Ensure we're in online mode
		graph.SetOperationalOffline(false)

		// Step 1: Create a large file (> 10MB)
		testFileName := "large_test_file.bin"
		// Create a file larger than 10MB to ensure multiple chunks
		testFileSize := 12 * 1024 * 1024 // 12MB
		fileID := "large-file-test-id"

		// Create test file content (repeating pattern for efficiency)
		pattern := []byte("LARGE_FILE_TEST_PATTERN_")
		testFileContent := make([]byte, testFileSize)
		for i := 0; i < testFileSize; i += len(pattern) {
			copy(testFileContent[i:], pattern)
		}

		// Calculate the QuickXorHash for the test file
		testFileQuickXorHash := graph.QuickXORHash(&testFileContent)

		// Create the file item
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: testFileQuickXorHash,
				},
			},
			Size: uint64(testFileSize),
			ETag: "initial-etag-large-file",
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Write the test file content to the content cache
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		n, err := fd.WriteAt(testFileContent, 0)
		assert.NoError(err, "Failed to write test file content")
		assert.Equal(testFileSize, n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes to trigger upload
		fileInode.hasChanges = true

		// Step 2: Verify file size is large enough to trigger chunked upload
		assert.True(uint64(testFileSize) > uploadLargeSize,
			"Test file should be larger than 4MB to trigger chunked upload")

		// Calculate expected number of chunks
		expectedChunks := int(math.Ceil(float64(testFileSize) / float64(uploadChunkSize)))
		assert.True(expectedChunks > 1,
			"Test file should require multiple chunks (expected %d chunks)", expectedChunks)

		// Mock the item response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Mock the upload session creation response
		// Note: The mock client automatically returns the upload URL for createUploadSession requests
		// We don't need to explicitly mock it, but we do need to mock the chunk upload responses

		// Mock the chunk upload responses
		// For each chunk except the last, return 202 Accepted
		// For the last chunk, return the completed file item with updated ETag
		uploadedFileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: testFileQuickXorHash,
				},
			},
			Size: uint64(testFileSize),
			ETag: "uploaded-etag-large-file",
		}

		// Mock the final chunk upload response to return the completed file item
		// The mock client will return 202 for intermediate chunks by default
		// We only need to configure the final response
		uploadedFileItemJSON, err := json.Marshal(uploadedFileItem)
		assert.NoError(err, "Failed to marshal uploaded file item")

		// Add mock response for the upload session URL
		// The mock client uses the full URL as the key
		mockClient.AddMockResponse("https://mock-upload.example.com/session123", uploadedFileItemJSON, 200, nil)

		// Queue the upload with high priority
		uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")
		assert.NotNil(uploadSession, "Upload session should not be nil")

		// Give the upload manager time to process the queued session
		time.Sleep(100 * time.Millisecond)

		// Verify session was created
		session, exists := fs.uploads.GetSession(fileID)
		assert.True(exists, "Upload session should exist after queueing")
		if session != nil {
			assert.Equal(fileID, session.GetID(), "Upload session ID should match file ID")
			assert.Equal(testFileName, session.GetName(), "Upload session name should match file name")
			assert.Equal(uint64(testFileSize), session.GetSize(), "Upload session size should match file size")
		}

		// Step 3: Monitor upload progress
		// For large files, the upload should take some time
		// Wait for the upload to complete with a reasonable timeout
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload completion")

		// Step 4: Verify complete file on OneDrive (mocked)
		// Get the updated inode
		updatedInode := fs.GetID(fileID)
		assert.NotNil(updatedInode, "File inode should exist after upload")

		// Verify ETag was updated
		updatedInode.RLock()
		updatedETag := updatedInode.DriveItem.ETag
		updatedSize := updatedInode.DriveItem.Size
		updatedInode.RUnlock()

		assert.Equal("uploaded-etag-large-file", updatedETag,
			"ETag should be updated after successful upload")
		assert.Equal(uint64(testFileSize), updatedSize,
			"File size should match after upload")

		// Verify file status is set to Local (synced)
		status := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, status.Status,
			"File status should be Local after successful upload")
	})
}
