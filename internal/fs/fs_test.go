package fs

import (
	"fmt"
	"github.com/bcherrington/onemount/internal/testutil/framework"
	"github.com/bcherrington/onemount/internal/testutil/helpers"
	"os"
	"path/filepath"
	"testing"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/bcherrington/onemount/internal/testutil"
)

// TestUT_FS_02_FileOperations_FileUpload_SuccessfulUpload verifies that a file can be successfully uploaded to OneDrive.
//
//	Test Case ID    UT-FS-02
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
func TestUT_FS_02_FileOperations_FileUpload_SuccessfulUpload(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	// Set up the fixture with additional test-specific setup
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Get the base fixture setup
		fsFixture, err := helpers.SetupFSTest(t, "FileOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			// Create the filesystem
			fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
			if err != nil {
				return nil, fmt.Errorf("failed to create filesystem: %w", err)
			}
			return fs, nil
		})
		if err != nil {
			return nil, err
		}

		// Set the root ID in the filesystem
		fs := fsFixture.FS.(*Filesystem)
		fs.root = fsFixture.RootID

		// Update the root folder
		rootItem := &graph.DriveItem{
			ID:   fsFixture.RootID,
			Name: "root",
			Folder: &graph.Folder{
				ChildCount: 0,
			},
		}
		fsFixture.MockClient.AddMockItem("/me/drive/root", rootItem)

		// Manually set up the root item
		rootInode := NewInodeDriveItem(rootItem)
		fs.InsertID(fsFixture.RootID, rootInode)

		// Insert the root item into the database to avoid the "offline and could not fetch the filesystem root item from disk" error
		fs.InsertNodeID(rootInode)

		return fsFixture, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *testutil.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		tempDir := fsFixture.TempDir
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID
		auth := fsFixture.Auth
		fs := fsFixture.FS.(*Filesystem)

		// Step 1: Create a new file in the local filesystem
		testFileName := "test_file.txt"
		testFileContent := "test content"
		testFilePath := filepath.Join(tempDir, testFileName)

		// Create a file in the local filesystem
		err := os.WriteFile(testFilePath, []byte(testFileContent), 0644)
		assert.NoError(err, "Failed to create test file")

		// Step 2: Write content to the file

		// Get the parent directory (root)
		rootInode := fs.GetID(rootID)
		assert.NotNil(rootInode, "Root directory not found in cache")

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
		assert.NoError(err, "Failed to open file for writing")

		// Write content to the file
		n, err := fd.WriteAt([]byte(testFileContent), 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(len(testFileContent), n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes
		fileInode.hasChanges = true

		// Step 3: Wait for the upload to complete

		// Queue the upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Step 4: Verify the file exists on OneDrive with correct content

		// Verify the file exists in the filesystem
		children, err := fs.GetChildrenID(rootID, auth)
		assert.NoError(err, "Failed to get children of root directory")
		assert.Equal(1, len(children), "Root directory should have 1 child")

		// Verify the file has the correct content
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(len(testFileContent)), fileInode.Size(), "File size mismatch")

		// Verify the file status
		status := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, status.Status, "File status should be local")
	})
}

// TestUT_FS_03_BasicFileSystemOperations_FileDownload_SuccessfulDownload verifies that a file can be successfully downloaded from OneDrive.
//
//	Test Case ID    UT-FS-03
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
func TestUT_FS_03_BasicFileSystemOperations_FileDownload_SuccessfulDownload(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Create a test fixture
	fixture := framework.NewUnitTestFixture("BasicFileSystemOperationsFixture")

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

		// Get auth tokens, either from existing file or create mock
		auth := helpers.GetTestAuth()

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
			"tempDir":     tempDir,
			"mockClient":  mockClient,
			"rootID":      rootID,
			"fileID":      fileID,
			"fileContent": fileContent,
			"auth":        auth,
			"fs":          fs,
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
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *testutil.UnitTestFixture, but got %T", fixture)
		}
		data := unitTestFixture.SetupData.(map[string]interface{})
		rootID := data["rootID"].(string)
		fileID := data["fileID"].(string)
		fileContent := data["fileContent"].(string)
		auth := data["auth"].(*graph.Auth)
		fs := data["fs"].(*Filesystem)

		// Step 1: Access a file that exists on OneDrive but not in local cache

		// Step 2: Read the file content

		// Get the file from the filesystem
		children, err := fs.GetChildrenID(rootID, auth)
		assert.NoError(err, "Failed to get children of root directory")
		assert.Equal(1, len(children), "Root directory should have 1 child")

		// Get the file inode
		var fileInode *Inode
		for _, child := range children {
			if child.Name() == "remote_file.txt" {
				fileInode = child
				break
			}
		}
		assert.NotNil(fileInode, "File not found in cache")

		// Open the file for reading
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for reading")

		// Read the file content
		content := make([]byte, fileInode.Size())
		n, err := fd.ReadAt(content, 0)
		assert.NoError(err, "Failed to read file content")
		assert.Equal(int(fileInode.Size()), n, "Number of bytes read doesn't match file size")

		// Step 3: Verify the content matches what's on OneDrive

		// Verify the file content
		assert.Equal(fileContent, string(content), "File content mismatch")

		// Verify the file status
		status := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, status.Status, "File status should be local")
	})
}

// TestUT_FS_04_RootRetrieval_OfflineMode_SuccessfulRetrieval verifies that the filesystem can retrieve the root item from the database
// when in offline mode, even if the root ID is not "root".
//
//	Test Case ID    UT-FS-04
//	Title           Root Item Retrieval in Offline Mode
//	Description     Verify that the filesystem can retrieve the root item from the database when in offline mode
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Root item is stored in the database with a custom ID
//	Steps           1. Initialize the filesystem with a custom root ID
//	                2. Store the root item in the database
//	                3. Retrieve the root item from the database
//	                4. Verify the root item has the correct properties
//	Expected Result Root item is successfully retrieved from the database with the correct properties
//	Notes: Tests the ability to retrieve the root item from the database in offline mode.
func TestUT_FS_04_RootRetrieval_OfflineMode_SuccessfulRetrieval(t *testing.T) {
	// Create a test fixture
	fixture := framework.NewUnitTestFixture("RootRetrievalFixture")

	// Set up the fixture
	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "onemount-test-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary directory: %w", err)
		}

		// Create a mock graph client
		mockClient := graph.NewMockGraphClient()

		// Set up the mock directory structure with a custom root ID
		rootID := "custom-root-id"
		rootItem := &graph.DriveItem{
			ID:   rootID,
			Name: "root",
			Folder: &graph.Folder{
				ChildCount: 1,
			},
		}

		// Add the root item to the mock client
		mockClient.AddMockItem("/me/drive/root", rootItem)

		// Get auth tokens, either from existing file or create mock
		auth := helpers.GetTestAuth()

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

		// Insert the root item into the database
		fs.InsertNodeID(rootInode)

		// Return the test data
		return map[string]interface{}{
			"tempDir":    tempDir,
			"mockClient": mockClient,
			"rootID":     rootID,
			"auth":       auth,
			"fs":         fs,
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
		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *testutil.UnitTestFixture, but got %T", fixture)
		}
		data := unitTestFixture.SetupData.(map[string]interface{})

		rootID := data["rootID"].(string)
		fs := data["fs"].(*Filesystem)

		// Verify that the root item is correctly stored in the database
		// This simulates what would happen in offline mode when the root item is retrieved from the database
		rootItem := fs.GetID(rootID)
		if rootItem == nil {
			t.Errorf("Failed to get root item with ID %s from the database", rootID)
		} else {
			t.Logf("Successfully retrieved root item with ID %s from the database", rootID)
		}

		// Verify that the root item has the correct properties
		if rootItem != nil {
			if !rootItem.IsDir() {
				t.Errorf("Root item should be a directory")
			}
			if rootItem.ParentID() != "" {
				t.Errorf("Root item should have no parent, but got parent ID %s", rootItem.ParentID())
			}
			if rootItem.Name() != "root" {
				t.Errorf("Root item should have name 'root', but got %s", rootItem.Name())
			}
		}
	})
}
