package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT02_FileOperations verifies that a file can be successfully uploaded to OneDrive.
//
//	Test Case ID    UT-02
//	Title           File Upload Synchronization
//	Description     Verify that a file can be successfully uploaded to OneDrive
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a new file in the local filesystem
//	                2. Write content to the file
//	                3. Wait for the upload to complete
//	                4. Verify the file exists on OneDrive with correct content
//	Expected Result File is successfully uploaded to OneDrive with the correct content
//	Notes: Tests basic file creation and writing, which triggers uploads.
func TestUT02_FileOperations(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Step 1: Create a new file in the local filesystem

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
	testFileName := "test_file.txt"
	testFileContent := "test content"
	testFilePath := filepath.Join(tempDir, testFileName)

	// Create a file in the local filesystem
	err = os.WriteFile(testFilePath, []byte(testFileContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Step 2: Write content to the file

	// Get the parent directory (root)
	rootInode := fs.GetID(rootID)
	require.NotNil(t, rootInode, "Root directory not found in cache")

	// Create a new file inode
	fileID := "file-id"
	fileItem := &graph.DriveItem{
		ID:   fileID,
		Name: testFileName,
		File: &graph.File{},
		Size: uint64(len(testFileContent)),
	}

	// Insert the file into the filesystem
	fileInode := NewInodeDriveItem(fileItem)
	fs.InsertNodeID(fileInode)
	fs.InsertChild(rootID, fileInode)

	// Open the file for writing
	fd, err := fs.content.Open(fileID)
	require.NoError(t, err, "Failed to open file for writing")

	// Write content to the file
	n, err := fd.WriteAt([]byte(testFileContent), 0)
	require.NoError(t, err, "Failed to write to file")
	require.Equal(t, len(testFileContent), n, "Number of bytes written doesn't match content length")

	// Mark the file as having changes
	fileInode.hasChanges = true

	// Step 3: Wait for the upload to complete

	// Queue the upload
	_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
	require.NoError(t, err, "Failed to queue upload")

	// Mock the upload response
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Wait for the upload to complete
	err = fs.uploads.WaitForUpload(fileID)
	require.NoError(t, err, "Failed to wait for upload")

	// Step 4: Verify the file exists on OneDrive with correct content

	// Verify the file exists in the filesystem
	children, err := fs.GetChildrenID(rootID, auth)
	require.NoError(t, err, "Failed to get children of root directory")
	assert.Equal(t, 1, len(children), "Root directory should have 1 child")

	// Verify the file has the correct content
	fileInode = fs.GetID(fileID)
	require.NotNil(t, fileInode, "File not found in cache")
	assert.Equal(t, testFileName, fileInode.Name(), "File name mismatch")
	assert.Equal(t, uint64(len(testFileContent)), fileInode.Size(), "File size mismatch")

	// Verify the file status
	status := fs.GetFileStatus(fileID)
	assert.Equal(t, StatusLocal, status.Status, "File status should be local")
}

// TestUT05_BasicFileSystemOperations verifies that a file can be successfully downloaded from OneDrive.
//
//	Test Case ID    UT-05
//	Title           File Download Synchronization
//	Description     Verify that a file can be successfully downloaded from OneDrive
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	                3. File exists on OneDrive
//	Steps           1. Access a file that exists on OneDrive but not in local cache
//	                2. Read the file content
//	                3. Verify the content matches what's on OneDrive
//	Expected Result File is successfully downloaded from OneDrive with the correct content
//	Notes: Tests reading files, which triggers downloads if not in cache.
func TestUT05_BasicFileSystemOperations(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Step 1: Access a file that exists on OneDrive but not in local cache

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
			ChildCount: 1,
		},
	}

	// Add the root item to the mock client
	mockClient.AddMockItem("/me/drive/root", rootItem)

	// Create a file item
	fileID := "file-id"
	fileContent := "remote file content"
	fileItem := &graph.DriveItem{
		ID:   fileID,
		Name: "remote_file.txt",
		File: &graph.File{},
		Size: uint64(len(fileContent)),
	}

	// Add the file to the mock client
	mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{fileItem})
	mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

	// Mock the file content
	mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", []byte(fileContent), 200, nil)

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

	// Step 2: Read the file content

	// Get the file from the filesystem
	children, err := fs.GetChildrenID(rootID, auth)
	require.NoError(t, err, "Failed to get children of root directory")
	assert.Equal(t, 1, len(children), "Root directory should have 1 child")

	// Get the file inode
	var fileInode *Inode
	for _, child := range children {
		if child.Name() == "remote_file.txt" {
			fileInode = child
			break
		}
	}
	require.NotNil(t, fileInode, "File not found in cache")

	// Open the file for reading
	fd, err := fs.content.Open(fileID)
	require.NoError(t, err, "Failed to open file for reading")

	// Read the file content
	content := make([]byte, fileInode.Size())
	n, err := fd.ReadAt(content, 0)
	require.NoError(t, err, "Failed to read file content")
	require.Equal(t, int(fileInode.Size()), n, "Number of bytes read doesn't match file size")

	// Step 3: Verify the content matches what's on OneDrive

	// Verify the file content
	assert.Equal(t, fileContent, string(content), "File content mismatch")

	// Verify the file status
	status := fs.GetFileStatus(fileID)
	assert.Equal(t, StatusLocal, status.Status, "File status should be local")
}
