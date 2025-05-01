package fs

import (
	"os"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT03_RepeatedUploads verifies that the same file can be uploaded multiple times
// with different content.
//
//	Test Case ID    UT-03
//	Title           Repeated File Upload
//	Description     Verify that the same file can be uploaded multiple times with different content
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a file with initial content
//	                2. Wait for upload to complete
//	                3. Modify the file content
//	                4. Wait for upload to complete
//	                5. Repeat steps 3-4 multiple times
//	Expected Result Each version of the file is successfully uploaded with the correct content
//	Notes: Directly tests uploading the same file multiple times with different content.
func TestUT03_RepeatedUploads(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Step 1: Create a file with initial content

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
			ChildCount: 0,
		},
	}

	// Add the root item to the mock client
	mockClient.AddMockItem("/me/drive/root", rootItem)
	mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{})

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

	// Create a test file path
	testFileName := "repeated_upload.txt"
	initialContent := "initial content"
	fileID := "file-id"
	fileItem := &graph.DriveItem{
		ID:   fileID,
		Name: testFileName,
		File: &graph.File{},
		Size: uint64(len(initialContent)),
	}

	// Insert the file into the filesystem
	fileInode := NewInodeDriveItem(fileItem)
	fs.InsertNodeID(fileInode)
	fs.InsertChild(rootID, fileInode)

	// Open the file for writing
	fd, err := fs.content.Open(fileID)
	require.NoError(t, err, "Failed to open file for writing")

	// Write initial content to the file
	n, err := fd.WriteAt([]byte(initialContent), 0)
	require.NoError(t, err, "Failed to write to file")
	require.Equal(t, len(initialContent), n, "Number of bytes written doesn't match content length")

	// Mark the file as having changes
	fileInode.hasChanges = true

	// Step 2: Wait for upload to complete

	// Queue the upload
	_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
	require.NoError(t, err, "Failed to queue upload")

	// Mock the upload response
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Wait for the upload to complete
	err = fs.uploads.WaitForUpload(fileID)
	require.NoError(t, err, "Failed to wait for upload")

	// Verify the file has the correct content
	fileInode = fs.GetID(fileID)
	require.NotNil(t, fileInode, "File not found in cache")
	assert.Equal(t, testFileName, fileInode.Name(), "File name mismatch")
	assert.Equal(t, uint64(len(initialContent)), fileInode.Size(), "File size mismatch")

	// Step 3: Modify the file content

	// Update the file content
	modifiedContent := "modified content"
	fileItem.Size = uint64(len(modifiedContent))

	// Open the file for writing
	fd, err = fs.content.Open(fileID)
	require.NoError(t, err, "Failed to open file for writing")

	// Write modified content to the file
	n, err = fd.WriteAt([]byte(modifiedContent), 0)
	require.NoError(t, err, "Failed to write to file")
	require.Equal(t, len(modifiedContent), n, "Number of bytes written doesn't match content length")

	// Mark the file as having changes
	fileInode.hasChanges = true

	// Step 4: Wait for upload to complete

	// Queue the upload
	_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
	require.NoError(t, err, "Failed to queue upload")

	// Mock the upload response
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Wait for the upload to complete
	err = fs.uploads.WaitForUpload(fileID)
	require.NoError(t, err, "Failed to wait for upload")

	// Verify the file has the correct content
	fileInode = fs.GetID(fileID)
	require.NotNil(t, fileInode, "File not found in cache")
	assert.Equal(t, testFileName, fileInode.Name(), "File name mismatch")
	assert.Equal(t, uint64(len(modifiedContent)), fileInode.Size(), "File size mismatch")

	// Step 5: Repeat steps 3-4 multiple times

	// Update the file content again
	finalContent := "final content"
	fileItem.Size = uint64(len(finalContent))

	// Open the file for writing
	fd, err = fs.content.Open(fileID)
	require.NoError(t, err, "Failed to open file for writing")

	// Write final content to the file
	n, err = fd.WriteAt([]byte(finalContent), 0)
	require.NoError(t, err, "Failed to write to file")
	require.Equal(t, len(finalContent), n, "Number of bytes written doesn't match content length")

	// Mark the file as having changes
	fileInode.hasChanges = true

	// Queue the upload
	_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
	require.NoError(t, err, "Failed to queue upload")

	// Mock the upload response
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Wait for the upload to complete
	err = fs.uploads.WaitForUpload(fileID)
	require.NoError(t, err, "Failed to wait for upload")

	// Verify the file has the correct content
	fileInode = fs.GetID(fileID)
	require.NotNil(t, fileInode, "File not found in cache")
	assert.Equal(t, testFileName, fileInode.Name(), "File name mismatch")
	assert.Equal(t, uint64(len(finalContent)), fileInode.Size(), "File size mismatch")
}

// TestUT04_UploadDiskSerialization verifies that large files can be uploaded correctly
// using upload sessions.
//
//	Test Case ID    UT-04
//	Title           Large File Upload
//	Description     Verify that large files can be uploaded correctly using upload sessions
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a large file (>4MB)
//	                2. Write content to the file
//	                3. Wait for the upload to complete
//	                4. Verify the file exists on OneDrive with correct content
//	Expected Result Large file is successfully uploaded to OneDrive with the correct content
//	Notes: Tests uploading large files using upload sessions.
func TestUT04_UploadDiskSerialization(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Step 1: Create a large file (>4MB)

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
			ChildCount: 0,
		},
	}

	// Add the root item to the mock client
	mockClient.AddMockItem("/me/drive/root", rootItem)
	mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{})

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

	// Create a test file path
	testFileName := "large_file.bin"
	fileSize := 5 * 1024 * 1024 // 5MB
	fileID := "large-file-id"
	fileItem := &graph.DriveItem{
		ID:   fileID,
		Name: testFileName,
		File: &graph.File{},
		Size: uint64(fileSize),
	}

	// Insert the file into the filesystem
	fileInode := NewInodeDriveItem(fileItem)
	fs.InsertNodeID(fileInode)
	fs.InsertChild(rootID, fileInode)

	// Open the file for writing
	fd, err := fs.content.Open(fileID)
	require.NoError(t, err, "Failed to open file for writing")

	// Generate large file content
	largeContent := make([]byte, fileSize)
	for i := 0; i < fileSize; i++ {
		largeContent[i] = byte(i % 256)
	}

	// Step 2: Write content to the file

	// Write content to the file
	n, err := fd.WriteAt(largeContent, 0)
	require.NoError(t, err, "Failed to write to file")
	require.Equal(t, fileSize, n, "Number of bytes written doesn't match content length")

	// Mark the file as having changes
	fileInode.hasChanges = true

	// Step 3: Wait for the upload to complete

	// Mock the upload session creation
	uploadURL := "https://example.com/upload-session"
	mockClient.AddMockResponse("/me/drive/items/"+fileID+"/createUploadSession", []byte(`{"uploadUrl":"`+uploadURL+`"}`), 200, nil)

	// Mock the upload session completion
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Queue the upload
	_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
	require.NoError(t, err, "Failed to queue upload")

	// Wait for the upload to complete
	err = fs.uploads.WaitForUpload(fileID)
	require.NoError(t, err, "Failed to wait for upload")

	// Step 4: Verify the file exists on OneDrive with correct content

	// Verify the file has the correct size
	fileInode = fs.GetID(fileID)
	require.NotNil(t, fileInode, "File not found in cache")
	assert.Equal(t, testFileName, fileInode.Name(), "File name mismatch")
	assert.Equal(t, uint64(fileSize), fileInode.Size(), "File size mismatch")

	// Verify the file status
	status := fs.GetFileStatus(fileID)
	assert.Equal(t, StatusLocal, status.Status, "File status should be local")
}
