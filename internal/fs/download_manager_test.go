package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/auriora/onemount/internal/errors"
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
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

// TestUT_FS_07_06_DownloadManager_ProcessDownload_MissingHashSeedsChecksum verifies that downloads succeed
// when the remote item omits checksum metadata and that the computed hash is persisted for future reads.
func TestUT_FS_07_06_DownloadManager_ProcessDownload_MissingHashSeedsChecksum(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "DownloadManagerMissingHashFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		fsFixture, err := helpers.SetupFSTest(t, "DownloadManagerMissingHashFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
			if err != nil {
				return nil, fmt.Errorf("failed to create filesystem: %w", err)
			}
			return fs, nil
		})
		if err != nil {
			return nil, err
		}

		fs := fsFixture.FS.(*Filesystem)
		fs.root = fsFixture.RootID

		rootItem := &graph.DriveItem{
			ID:   fsFixture.RootID,
			Name: "root",
			Folder: &graph.Folder{
				ChildCount: 1,
			},
		}
		fsFixture.MockClient.AddMockItem("/me/drive/root", rootItem)
		rootInode := NewInodeDriveItem(rootItem)
		fs.InsertID(fsFixture.RootID, rootInode)
		fs.InsertNodeID(rootInode)

		content := []byte("content without remote hash")
		actualHash := graph.QuickXORHash(&content)
		fileID := "missing-hash-file-id"
		fileName := "missing_hash.txt"

		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: fileName,
			Parent: &graph.DriveItemParent{
				ID: rootItem.ID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{}, // no QuickXorHash provided by service
			},
			Size: uint64(len(content)),
		}

		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(fsFixture.RootID, fileInode)

		// Persist a metadata entry that claims the item is hydrated but lacks a hash.
		entry := &metadata.Entry{
			ID:            fileID,
			Name:          fileName,
			ItemType:      metadata.ItemKindFile,
			State:         metadata.ItemStateHydrated,
			OverlayPolicy: metadata.OverlayPolicyRemoteWins,
			Size:          uint64(len(content)),
			Pin: metadata.PinState{
				Mode: metadata.PinModeUnset,
			},
		}
		_ = fs.metadataStore.Save(context.Background(), entry)

		fileItemJSON, _ := json.Marshal(fileItem)
		fsFixture.MockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)
		fsFixture.MockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", content, http.StatusOK, nil)

		fsFixture.Data["fileID"] = fileID
		fsFixture.Data["expectedHash"] = actualHash

		return fsFixture, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		fileID := fsFixture.Data["fileID"].(string)
		expectedHash := fsFixture.Data["expectedHash"].(string)

		session, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(session, "Download session should not be nil")

		err = fs.downloads.WaitForDownload(fileID)
		assert.NoError(err, "Download should succeed even without remote hash")

		status, err := fs.downloads.GetDownloadStatus(fileID)
		assert.NoError(err, "Failed to get download status")
		assert.Equal(DownloadCompletedState, status, "Download status should be completed")

		entry, err := fs.metadataStore.Get(context.Background(), fileID)
		assert.NoError(err, "Metadata entry should be persisted")
		assert.Equal(metadata.ItemStateHydrated, entry.State, "Entry should be hydrated after download")
		assert.Equal(expectedHash, entry.ContentHash, "Computed hash should be stored in metadata")

		inode := fs.GetID(fileID)
		if inode == nil {
			t.Fatalf("inode should exist after download")
		}
		inode.mu.RLock()
		hashOnInode := ""
		if inode.DriveItem.File != nil {
			hashOnInode = inode.DriveItem.File.Hashes.QuickXorHash
		}
		inode.mu.RUnlock()
		assert.Equal(expectedHash, hashOnInode, "Inode should have seeded checksum")
	})
}

// TestUT_FS_10_DownloadManager_ChunkBasedDownload_LargeFile verifies that large files are downloaded using chunk-based operations
//
//	Test Case ID    UT-FS-10
//	Title           Download Manager - Chunk-based Download (Large File)
//	Description     Verify that large files are downloaded correctly using chunk-based operations
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a large file item in the filesystem
//	                2. Queue the file for download
//	                3. Verify that the download uses chunked operations
//	                4. Monitor chunk download progress
//	                5. Wait for the download to complete
//	Expected Result The file is successfully downloaded using chunk-based operations with proper progress tracking
//	Notes: Tests chunk-based download operations for large files.
func TestUT_FS_10_DownloadManager_ChunkBasedDownload_LargeFile(t *testing.T) {
	// Note: parallel execution disabled to avoid mock HTTP client cleanup races

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ChunkBasedDownloadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a large test file (10MB to trigger chunked download)
		largeFileSize := 10 * 1024 * 1024 // 10MB
		largeFileContent := make([]byte, largeFileSize)
		for i := range largeFileContent {
			largeFileContent[i] = byte(i % 256)
		}

		// Calculate the QuickXorHash for the large file
		largeFileQuickXorHash := graph.QuickXORHash(&largeFileContent)

		// Create a test file item
		testFileName := "large_download_test_file.bin"
		fileID := "large-download-file-id"
		rootID := fsFixture.RootID

		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: largeFileQuickXorHash,
				},
			},
			Size: uint64(largeFileSize),
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Mock the file item response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Mock the content download response with chunked content
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", largeFileContent, 200, nil)

		// Queue the download
		downloadSession, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(downloadSession, "Download session should not be nil")

		// Verify that the download session is configured for chunked download
		assert.True(downloadSession.Size > uint64(downloadChunkSize), "File should be large enough to trigger chunked download")
		assert.True(downloadSession.CanResume, "Large file download should support resume")

		// Wait for the download to complete
		err = fs.downloads.WaitForDownload(fileID)
		assert.NoError(err, "Failed to wait for download")

		// Verify download completion
		status, err := fs.downloads.GetDownloadStatus(fileID)
		assert.NoError(err, "Failed to get download status")
		assert.Equal(DownloadCompletedState, status, "Download should be completed")

		// Verify that progress was tracked during download
		assert.True(downloadSession.BytesDownloaded > 0, "Bytes downloaded should be greater than 0")
		assert.Equal(downloadSession.Size, downloadSession.BytesDownloaded, "All bytes should be downloaded")
		assert.True(downloadSession.TotalChunks > 1, "Large file should be downloaded in multiple chunks")
	})
}

// TestUT_FS_11_DownloadManager_ResumeDownload_InterruptedTransfer verifies that interrupted downloads can be resumed
//
//	Test Case ID    UT-FS-11
//	Title           Download Manager - Download Resume Functionality
//	Description     Verify that interrupted downloads can be resumed from the last successful chunk
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a large file item in the filesystem
//	                2. Start the download and simulate interruption
//	                3. Verify download session persistence
//	                4. Resume the download from the last checkpoint
//	                5. Wait for the download to complete
//	Expected Result The download resumes correctly from the last successful chunk
//	Notes: Tests download resume functionality for interrupted transfers.
func TestUT_FS_11_DownloadManager_ResumeDownload_InterruptedTransfer(t *testing.T) {
	// Note: parallel execution disabled to avoid mock HTTP client cleanup races

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ResumeDownloadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a large test file (8MB to trigger chunked download)
		largeFileSize := 8 * 1024 * 1024 // 8MB
		largeFileContent := make([]byte, largeFileSize)
		for i := range largeFileContent {
			largeFileContent[i] = byte(i % 256)
		}

		// Calculate the QuickXorHash for the large file
		largeFileQuickXorHash := graph.QuickXORHash(&largeFileContent)

		// Create a test file item
		testFileName := "resume_download_test_file.bin"
		fileID := "resume-download-file-id"
		rootID := fsFixture.RootID

		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: largeFileQuickXorHash,
				},
			},
			Size: uint64(largeFileSize),
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Mock the file item response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Mock the content download response with chunked content
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", largeFileContent, 200, nil)

		// Queue the download
		downloadSession, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(downloadSession, "Download session should not be nil")

		// Verify that the download session is configured for resumable download
		assert.True(downloadSession.Size > uint64(downloadChunkSize), "File should be large enough to trigger chunked download")
		assert.True(downloadSession.CanResume, "Large file download should support resume")

		// Simulate partial download by setting progress
		chunkSize := uint64(downloadChunkSize)
		chunksCompleted := 2 // Simulate 2 chunks completed
		bytesDownloaded := uint64(chunksCompleted) * chunkSize
		downloadSession.updateProgress(chunksCompleted-1, bytesDownloaded)

		// Verify progress state before resume
		assert.Equal(bytesDownloaded, downloadSession.BytesDownloaded, "Bytes downloaded should match simulated progress")
		assert.Equal(chunksCompleted-1, downloadSession.LastSuccessfulChunk, "Last successful chunk should be set")
		assert.True(downloadSession.canResumeDownload(), "Download should be resumable")

		// Wait for the download to complete (it should resume from the last chunk)
		err = fs.downloads.WaitForDownload(fileID)
		assert.NoError(err, "Failed to wait for download")

		// Verify download completion
		status, err := fs.downloads.GetDownloadStatus(fileID)
		assert.NoError(err, "Failed to get download status")
		assert.Equal(DownloadCompletedState, status, "Download should be completed")

		// Verify that all bytes were downloaded
		assert.Equal(downloadSession.Size, downloadSession.BytesDownloaded, "All bytes should be downloaded")
	})
}

// TestUT_FS_12_DownloadManager_ConcurrentDownloads_QueueManagement verifies concurrent download management
//
//	Test Case ID    UT-FS-12
//	Title           Download Manager - Concurrent Download Management
//	Description     Verify that multiple concurrent downloads are managed correctly with proper queue management
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create multiple test files
//	                2. Queue all files for download simultaneously
//	                3. Verify queue management and worker allocation
//	                4. Monitor concurrent download execution
//	                5. Wait for all downloads to complete
//	Expected Result Multiple downloads are managed concurrently with proper queue management
//	Notes: Tests concurrent transfer management and queue handling for downloads.
func TestUT_FS_12_DownloadManager_ConcurrentDownloads_QueueManagement(t *testing.T) {
	// Note: parallel execution disabled to avoid mock HTTP client cleanup races

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ConcurrentDownloadsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create multiple test files
		numFiles := 4
		fileSize := 1024 * 1024 // 1MB each
		fileIDs := make([]string, numFiles)
		downloadSessions := make([]*DownloadSession, numFiles)

		for i := 0; i < numFiles; i++ {
			// Create test file content
			testFileContent := make([]byte, fileSize)
			for j := range testFileContent {
				testFileContent[j] = byte((i + j) % 256)
			}

			// Calculate the QuickXorHash for the test file content
			testFileQuickXorHash := graph.QuickXORHash(&testFileContent)

			// Create file item
			testFileName := fmt.Sprintf("concurrent_download_test_file_%d.bin", i)
			fileID := fmt.Sprintf("concurrent-download-file-id-%d", i)
			fileIDs[i] = fileID
			rootID := fsFixture.RootID

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
				Size: uint64(fileSize),
			}

			// Insert the file into the filesystem
			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			// Mock the file item response
			mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

			// Mock the content download response
			mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", testFileContent, 200, nil)
		}

		// Queue all downloads simultaneously
		for i := 0; i < numFiles; i++ {
			downloadSession, err := fs.downloads.QueueDownload(fileIDs[i])
			assert.NoError(err, "Failed to queue download")
			assert.NotNil(downloadSession, "Download session should not be nil")
			downloadSessions[i] = downloadSession
		}

		// Wait for all downloads to complete
		for i := 0; i < numFiles; i++ {
			err := fs.downloads.WaitForDownload(fileIDs[i])
			assert.NoError(err, "Failed to wait for download")

			// Verify download completion
			status, err := fs.downloads.GetDownloadStatus(fileIDs[i])
			assert.NoError(err, "Failed to get download status")
			assert.Equal(DownloadCompletedState, status, "Download should be completed")

			// Verify progress tracking
			assert.Equal(downloadSessions[i].Size, downloadSessions[i].BytesDownloaded, "All bytes should be downloaded")
		}
	})
}
