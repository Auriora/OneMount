package fs

import (
	"encoding/json"
	"fmt"
	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/testutil/framework"
	"github.com/auriora/onemount/pkg/testutil/helpers"
	"net/http"
	"testing"
)

// TestUT_FS_07_01_DownloadManager_QueueDownload_NonExistentFile tests that QueueDownload returns an error when the file doesn't exist.
//
//	Test Case ID    UT-FS-07-01
//	Title           Download Manager - Queue Non-Existent File
//	Description     Verify that QueueDownload returns an error when the file doesn't exist
//	Preconditions   1. User is authenticated with valid credentials
//	Steps           1. Attempt to queue a download for a non-existent file
//	Expected Result QueueDownload returns a NotFoundError
//	Notes: Tests error handling when attempting to download a non-existent file.
func TestUT_FS_07_01_DownloadManager_QueueDownload_NonExistentFile(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DownloadManagerNonExistentFileFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *testutil.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Attempt to queue a download for a non-existent file
		nonExistentFileID := "non-existent-file-id"
		session, err := fs.downloads.QueueDownload(nonExistentFileID)

		// Verify that QueueDownload returns a NotFoundError
		assert.Error(err, "QueueDownload should return an error for non-existent file")
		assert.True(errors.IsNotFoundError(err), "Error should be a NotFoundError")
		assert.Nil(session, "Session should be nil for non-existent file")
	})
}

// TestUT_FS_07_02_DownloadManager_GetDownloadStatus_NonExistentSession tests that GetDownloadStatus returns an error when the session doesn't exist.
//
//	Test Case ID    UT-FS-07-02
//	Title           Download Manager - Get Status of Non-Existent Session
//	Description     Verify that GetDownloadStatus returns an error when the session doesn't exist
//	Preconditions   1. User is authenticated with valid credentials
//	Steps           1. Attempt to get the status of a non-existent download session
//	Expected Result GetDownloadStatus returns a NotFoundError
//	Notes: Tests error handling when attempting to get the status of a non-existent download session.
func TestUT_FS_07_02_DownloadManager_GetDownloadStatus_NonExistentSession(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DownloadManagerNonExistentSessionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *testutil.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Attempt to get the status of a non-existent download session
		nonExistentSessionID := "non-existent-session-id"
		status, err := fs.downloads.GetDownloadStatus(nonExistentSessionID)

		// Verify that GetDownloadStatus returns a NotFoundError
		assert.Error(err, "GetDownloadStatus should return an error for non-existent session")
		assert.True(errors.IsNotFoundError(err), "Error should be a NotFoundError")
		assert.Equal(DownloadState(0), status, "Status should be the zero value for non-existent session")
	})
}

// TestUT_FS_07_03_DownloadManager_WaitForDownload_NonExistentSession tests that WaitForDownload returns an error when the session doesn't exist.
//
//	Test Case ID    UT-FS-07-03
//	Title           Download Manager - Wait for Non-Existent Session
//	Description     Verify that WaitForDownload returns an error when the session doesn't exist
//	Preconditions   1. User is authenticated with valid credentials
//	Steps           1. Attempt to wait for a non-existent download session
//	Expected Result WaitForDownload returns a NotFoundError
//	Notes: Tests error handling when attempting to wait for a non-existent download session.
func TestUT_FS_07_03_DownloadManager_WaitForDownload_NonExistentSession(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DownloadManagerWaitNonExistentSessionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *testutil.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Attempt to wait for a non-existent download session
		nonExistentSessionID := "non-existent-session-id"
		err := fs.downloads.WaitForDownload(nonExistentSessionID)

		// Verify that WaitForDownload returns a NotFoundError
		assert.Error(err, "WaitForDownload should return an error for non-existent session")
		assert.True(errors.IsNotFoundError(err), "Error should be a NotFoundError")
	})
}

// TestUT_FS_07_04_DownloadManager_ProcessDownload_NetworkError tests that processDownload handles network errors correctly.
//
//	Test Case ID    UT-FS-07-04
//	Title           Download Manager - Process Download with Network Error
//	Description     Verify that processDownload handles network errors correctly
//	Preconditions   1. User is authenticated with valid credentials
//	Steps           1. Set up a mock that returns a network error
//	                2. Queue a download
//	                3. Wait for the download to complete
//	Expected Result The download session is marked as errored with a NetworkError
//	Notes: Tests error handling when a network error occurs during download.
func TestUT_FS_07_04_DownloadManager_ProcessDownload_NetworkError(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DownloadManagerNetworkErrorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		fsFixture, err := helpers.SetupFSTest(t, "DownloadManagerNetworkErrorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		testFileName := "network_error_test.txt"
		fileID := "network-error-file-id"

		// Create a file item
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "hash",
				},
			},
			Size: 100,
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(fsFixture.RootID, fileInode)

		// Set up the mock to return the file metadata first (required by GetItemContentStream)
		fileItemJSON, _ := json.Marshal(fileItem)
		fsFixture.MockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)

		// Set up the mock to return a network error for content download
		networkErr := errors.NewNetworkError("network error during download", nil)
		fsFixture.MockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", nil, http.StatusServiceUnavailable, networkErr)

		// Add the test data to the fixture
		fsFixture.Data["testFileName"] = testFileName
		fsFixture.Data["fileID"] = fileID

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
		fs := fsFixture.FS.(*Filesystem)
		fileID := fsFixture.Data["fileID"].(string)

		// Queue the download
		session, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(session, "Download session should not be nil")

		// Wait for the download to complete or error out
		err = fs.downloads.WaitForDownload(fileID)

		// Verify that the download failed with a network error
		assert.Error(err, "Download should fail with a network error")
		assert.True(errors.IsNetworkError(err), "Error should be a NetworkError")

		// Verify the session state
		status, err := fs.downloads.GetDownloadStatus(fileID)
		assert.NoError(err, "Failed to get download status")
		assert.Equal(downloadErrored, status, "Download status should be downloadErrored")

		// Verify the file status
		fileStatus := fs.GetFileStatus(fileID)
		assert.Equal(StatusError, fileStatus.Status, "File status should be StatusError")
	})
}

// TestUT_FS_07_05_DownloadManager_ProcessDownload_ChecksumError tests that processDownload handles checksum verification errors correctly.
//
//	Test Case ID    UT-FS-07-05
//	Title           Download Manager - Process Download with Checksum Error
//	Description     Verify that processDownload handles checksum verification errors correctly
//	Preconditions   1. User is authenticated with valid credentials
//	Steps           1. Set up a file with an expected checksum
//	                2. Set up a mock that returns content with a different checksum
//	                3. Queue a download
//	                4. Wait for the download to complete
//	Expected Result The download session is marked as errored with a ValidationError
//	Notes: Tests error handling when a checksum verification error occurs during download.
func TestUT_FS_07_05_DownloadManager_ProcessDownload_ChecksumError(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DownloadManagerChecksumErrorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		fsFixture, err := helpers.SetupFSTest(t, "DownloadManagerChecksumErrorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		testFileName := "checksum_error_test.txt"
		fileID := "checksum-error-file-id"
		expectedContent := "expected content"
		expectedContentBytes := []byte(expectedContent)
		expectedChecksum := graph.QuickXORHash(&expectedContentBytes)

		// Create a file item with the expected checksum
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: expectedChecksum,
				},
			},
			Size: uint64(len(expectedContent)),
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(fsFixture.RootID, fileInode)

		// Set up the mock to return the file metadata first (required by GetItemContentStream)
		fileItemJSON, _ := json.Marshal(fileItem)
		fsFixture.MockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)

		// Set up the mock to return different content (which will cause a checksum mismatch)
		differentContent := "different content that will cause a checksum mismatch"
		fsFixture.MockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", []byte(differentContent), http.StatusOK, nil)

		// Add the test data to the fixture
		fsFixture.Data["testFileName"] = testFileName
		fsFixture.Data["fileID"] = fileID
		fsFixture.Data["expectedContent"] = expectedContent
		fsFixture.Data["differentContent"] = differentContent

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
		fs := fsFixture.FS.(*Filesystem)
		fileID := fsFixture.Data["fileID"].(string)

		// Queue the download
		session, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(session, "Download session should not be nil")

		// Wait for the download to complete or error out
		err = fs.downloads.WaitForDownload(fileID)

		// Verify that the download failed with a validation error
		assert.Error(err, "Download should fail with a validation error")
		assert.True(errors.IsValidationError(err), "Error should be a ValidationError")

		// Verify the session state
		status, err := fs.downloads.GetDownloadStatus(fileID)
		assert.NoError(err, "Failed to get download status")
		assert.Equal(downloadErrored, status, "Download status should be downloadErrored")

		// Verify the file status
		fileStatus := fs.GetFileStatus(fileID)
		assert.Equal(StatusError, fileStatus.Status, "File status should be StatusError")
	})
}
