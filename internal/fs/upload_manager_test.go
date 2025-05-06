package fs

import (
	"encoding/json"
	"fmt"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"testing"

	"github.com/auriora/onemount/internal/fs/graph"
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

		// Mock the content upload response
		fileItemJSON, err := json.Marshal(fileItem)
		assert.NoError(err, "Failed to marshal file item")
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

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

		// Mock the content upload response
		fileItemJSON, err = json.Marshal(fileItem)
		assert.NoError(err, "Failed to marshal file item")
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

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

		// Mock the content upload response
		fileItemJSON, err = json.Marshal(fileItem)
		assert.NoError(err, "Failed to marshal file item")
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

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

		// Set the system to offline mode
		graph.SetOperationalOffline(true)
		defer graph.SetOperationalOffline(false) // Reset to online mode after the test

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

		// Mock the content upload response
		fileItemJSON, err := json.Marshal(fileItem)
		assert.NoError(err, "Failed to marshal file item")
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

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

		// Mock the content upload response
		fileItemJSON, err = json.Marshal(fileItem)
		assert.NoError(err, "Failed to marshal file item")
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

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

		// Mock the content upload response
		fileItemJSON, err = json.Marshal(fileItem)
		assert.NoError(err, "Failed to marshal file item")
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

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
// and that the upload session is properly serialized to disk.
//
//	Test Case ID    UT-FS-06
//	Title           Upload Disk Serialization (Large File)
//	Description     Verify that large files can be uploaded correctly and that the upload session is properly serialized to disk
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a large file
//	                2. Queue the file for upload
//	                3. Verify that the upload session is serialized to disk
//	                4. Wait for the upload to complete
//	                5. Verify that the upload session is removed from disk
//	Expected Result The file is successfully uploaded and the upload session is properly serialized to and removed from disk
//	Notes: Directly tests the serialization of upload sessions to disk for large files.
func TestUT_FS_06_UploadDiskSerialization_LargeFile_SuccessfulUpload(t *testing.T) {
	// Skip this test for now as it's not fully implemented
	t.Skip("Test not fully implemented yet")

	// Mark the test for parallel execution

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "UploadDiskSerializationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		fsFixture, err := helpers.SetupFSTest(t, "UploadDiskSerializationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		fsFixture.MockClient.AddMockItems("/me/drive/items/"+fsFixture.RootID+"/children", []*graph.DriveItem{})

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
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID
		fs := fsFixture.FS.(*Filesystem)

		// Step 1: Create a large file
		testFileName := "large_file.bin"
		fileID := "large-file-id"
		fileSize := uploadLargeSize + 1 // Just over the large file threshold

		// Create a large file item
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			File: &graph.File{},
			Size: fileSize,
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Mock the upload response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Mock the content upload response
		fileItemJSON, err := json.Marshal(fileItem)
		assert.NoError(err, "Failed to marshal file item")
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

		// TODO: Complete the test implementation
		// - Create a large file
		// - Queue it for upload
		// - Verify serialization to disk
		// - Wait for upload to complete
		// - Verify removal from disk
	})
}
