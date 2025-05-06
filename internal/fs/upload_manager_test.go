package fs

import (
	"fmt"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"os"
	"testing"

	"github.com/auriora/onemount/internal/fs/graph"
	"github.com/auriora/onemount/internal/testutil"
)

// TestUT_FS_05_RepeatedUploads_OnlineMode_SuccessfulUpload verifies that the same file can be uploaded multiple times
// with different content when network connection is available.
//
//	Test Case ID    UT-FS-05-02
//	Title           Repeated File Upload (Online)
//	Description     Verify that the same file can be uploaded multiple times with different content
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a file with initial content
//	                2. Wait for upload to complete
//	                3. Modify the file content
//	                4. Wait for upload to complete
//	                5. Repeat steps 3-4 multiple times
//	Expected Result Each version of the file is successfully uploaded with the correct content
//	Notes: Directly tests uploading the same file multiple times with different content in online mode.
func TestUT_FS_05_02_RepeatedUploads_OnlineMode_SuccessfulUpload(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "RepeatedUploadsOnlineFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		fsFixture, err := helpers.SetupFSTest(t, "RepeatedUploadsOnlineFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create test file data
		testFileName := "repeated_upload.txt"
		initialContent := "initial content"
		fileID := "file-id"
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			File: &graph.File{},
			Size: uint64(len(initialContent)),
		}

		// Add the test data to the fixture
		fsFixture.Data["testFileName"] = testFileName
		fsFixture.Data["initialContent"] = initialContent
		fsFixture.Data["fileID"] = fileID
		fsFixture.Data["fileItem"] = fileItem

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
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID
		fs := fsFixture.FS.(*Filesystem)
		testFileName := fsFixture.Data["testFileName"].(string)
		initialContent := fsFixture.Data["initialContent"].(string)
		fileID := fsFixture.Data["fileID"].(string)
		fileItem := fsFixture.Data["fileItem"].(*graph.DriveItem)

		// Step 1: Create a file with initial content

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Open the file for writing
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		// Write initial content to the file
		n, err := fd.WriteAt([]byte(initialContent), 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(len(initialContent), n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes
		fileInode.hasChanges = true

		// Step 2: Wait for upload to complete

		// Queue the upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Verify the file has the correct content
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(len(initialContent)), fileInode.Size(), "File size mismatch")

		// Step 3: Modify the file content

		// Update the file content
		modifiedContent := "modified content"
		fileItem.Size = uint64(len(modifiedContent))

		// Open the file for writing
		fd, err = fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		// Write modified content to the file
		n, err = fd.WriteAt([]byte(modifiedContent), 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(len(modifiedContent), n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes
		fileInode.hasChanges = true

		// Step 4: Wait for upload to complete

		// Queue the upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Verify the file has the correct content
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(len(modifiedContent)), fileInode.Size(), "File size mismatch")

		// Step 5: Repeat steps 3-4 multiple times

		// Update the file content again
		finalContent := "final content"
		fileItem.Size = uint64(len(finalContent))

		// Open the file for writing
		fd, err = fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		// Write final content to the file
		n, err = fd.WriteAt([]byte(finalContent), 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(len(finalContent), n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes
		fileInode.hasChanges = true

		// Queue the upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Verify the file has the correct content
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(len(finalContent)), fileInode.Size(), "File size mismatch")
	})
}

// TestUT_FS_05_01_RepeatedUploads_OfflineMode_SuccessfulUpload verifies that the same file can be uploaded multiple times
// with different content when in offline mode.
//
//	Test Case ID    UT-FS-05-01
//	Title           Repeated File Upload (Offline)
//	Description     Verify that the same file can be uploaded multiple times with different content in offline mode
//	Preconditions   1. User is authenticated with valid credentials
//	                2. System is in operational offline mode
//	Steps           1. Create a file with initial content
//	                2. Wait for upload to complete
//	                3. Modify the file content
//	                4. Wait for upload to complete
//	                5. Repeat steps 3-4 multiple times
//	Expected Result Each version of the file is successfully uploaded with the correct content
//	Notes: Directly tests uploading the same file multiple times with different content in offline mode.
func TestUT_FS_05_01_RepeatedUploads_OfflineMode_SuccessfulUpload(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "RepeatedUploadsOfflineFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		fsFixture, err := helpers.SetupFSTest(t, "RepeatedUploadsOfflineFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Set operational offline mode to prevent real network requests
		graph.SetOperationalOffline(true)

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

		// Create test file data
		testFileName := "repeated_upload.txt"
		initialContent := "initial content"
		fileID := "file-id"
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			File: &graph.File{},
			Size: uint64(len(initialContent)),
		}

		// Add the test data to the fixture
		fsFixture.Data["testFileName"] = testFileName
		fsFixture.Data["initialContent"] = initialContent
		fsFixture.Data["fileID"] = fileID
		fsFixture.Data["fileItem"] = fileItem

		return fsFixture, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		// Reset operational offline mode
		graph.SetOperationalOffline(false)

		// The base teardown will be called automatically
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
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID
		fs := fsFixture.FS.(*Filesystem)
		testFileName := fsFixture.Data["testFileName"].(string)
		initialContent := fsFixture.Data["initialContent"].(string)
		fileID := fsFixture.Data["fileID"].(string)
		fileItem := fsFixture.Data["fileItem"].(*graph.DriveItem)

		// Step 1: Create a file with initial content

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Open the file for writing
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		// Write initial content to the file
		n, err := fd.WriteAt([]byte(initialContent), 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(len(initialContent), n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes
		fileInode.hasChanges = true

		// Step 2: Wait for upload to complete

		// Queue the upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Verify the file has the correct content
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(len(initialContent)), fileInode.Size(), "File size mismatch")

		// Step 3: Modify the file content

		// Update the file content
		modifiedContent := "modified content"
		fileItem.Size = uint64(len(modifiedContent))

		// Open the file for writing
		fd, err = fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		// Write modified content to the file
		n, err = fd.WriteAt([]byte(modifiedContent), 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(len(modifiedContent), n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes
		fileInode.hasChanges = true

		// Step 4: Wait for upload to complete

		// Queue the upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Verify the file has the correct content
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(len(modifiedContent)), fileInode.Size(), "File size mismatch")

		// Step 5: Repeat steps 3-4 multiple times

		// Update the file content again
		finalContent := "final content"
		fileItem.Size = uint64(len(finalContent))

		// Open the file for writing
		fd, err = fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		// Write final content to the file
		n, err = fd.WriteAt([]byte(finalContent), 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(len(finalContent), n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes
		fileInode.hasChanges = true

		// Queue the upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Verify the file has the correct content
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(len(finalContent)), fileInode.Size(), "File size mismatch")
	})
}

// TestUT_FS_06_UploadDiskSerialization_LargeFile_SuccessfulUpload verifies that large files can be uploaded correctly
// using upload sessions.
//
//	Test Case ID    UT-FS-06
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
func TestUT_FS_06_UploadDiskSerialization_LargeFile_SuccessfulUpload(t *testing.T) {
	// Mark the test for parallel execution
	t.Parallel()

	// Create a test fixture
	fixture := framework.NewUnitTestFixture("UploadDiskSerializationFixture")

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
				ChildCount: 0,
			},
		}

		// Add the root item to the mock client
		mockClient.AddMockItem("/me/drive/root", rootItem)
		mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{})

		// Get auth tokens, either from existing file or create mock
		auth := helpers.GetTestAuth()

		// Create the filesystem
		fs, err := NewFilesystem(auth, tempDir, 30)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}

		// Set the root ID
		fs.root = rootID

		// Create test file data
		testFileName := "large_file.bin"
		fileSize := 5 * 1024 * 1024 // 5MB
		fileID := "large-file-id"
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			File: &graph.File{},
			Size: uint64(fileSize),
		}

		// Generate large file content
		largeContent := make([]byte, fileSize)
		for i := 0; i < fileSize; i++ {
			largeContent[i] = byte(i % 256)
		}

		// Return the test data
		return map[string]interface{}{
			"tempDir":      tempDir,
			"mockClient":   mockClient,
			"rootID":       rootID,
			"auth":         auth,
			"fs":           fs,
			"testFileName": testFileName,
			"fileSize":     fileSize,
			"fileID":       fileID,
			"fileItem":     fileItem,
			"largeContent": largeContent,
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
		mockClient := data["mockClient"].(*graph.MockGraphClient)
		rootID := data["rootID"].(string)
		fs := data["fs"].(*Filesystem)
		testFileName := data["testFileName"].(string)
		fileSize := data["fileSize"].(int)
		fileID := data["fileID"].(string)
		fileItem := data["fileItem"].(*graph.DriveItem)
		largeContent := data["largeContent"].([]byte)

		// Step 1: Create a large file (>4MB)

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Open the file for writing
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		// Step 2: Write content to the file

		// Write content to the file
		n, err := fd.WriteAt(largeContent, 0)
		assert.NoError(err, "Failed to write to file")
		assert.Equal(fileSize, n, "Number of bytes written doesn't match content length")

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
		assert.NoError(err, "Failed to queue upload")

		// Wait for the upload to complete
		err = fs.uploads.WaitForUpload(fileID)
		assert.NoError(err, "Failed to wait for upload")

		// Step 4: Verify the file exists on OneDrive with correct content

		// Verify the file has the correct size
		fileInode = fs.GetID(fileID)
		assert.NotNil(fileInode, "File not found in cache")
		assert.Equal(testFileName, fileInode.Name(), "File name mismatch")
		assert.Equal(uint64(fileSize), fileInode.Size(), "File size mismatch")

		// Verify the file status
		status := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, status.Status, "File status should be local")
	})
}
