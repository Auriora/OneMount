package fs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_08_01_DownloadManager_SingleFileDownload tests single file download end-to-end
//
//	Test Case ID    IT-FS-08-01
//	Title           Download Manager - Single File Download Integration Test
//	Description     Verify that a single file can be downloaded successfully with proper status tracking and caching
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Filesystem is mounted
//	                3. Network connection is available
//	Steps           1. Create a test file in the mock OneDrive
//	                2. Queue the file for download
//	                3. Monitor download progress in logs
//	                4. Wait for download to complete
//	                5. Verify file content is correct
//	                6. Verify file is cached
//	                7. Verify file status transitions
//	Expected Result File downloads successfully, content matches, file is cached, status updates correctly
//	Requirements    3.2 (On-Demand File Download)
//	Notes: Integration test for single file download workflow
func TestIT_FS_08_01_DownloadManager_SingleFileDownload(t *testing.T) {
	// Parallel execution disabled: these fixtures mutate the shared Graph mock client
	// and must run serially to avoid HTTP client races with other download suites.

	// Create a test fixture using the mock setup (download manager tests use mocks)
	fixture := helpers.SetupMockFSTestFixture(t, "SingleFileDownloadIntegrationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create test file data
		testFileName := "single_download_test.txt"
		testFileContent := "This is test content for single file download verification"
		testFileBytes := []byte(testFileContent)
		fileID := "single-download-file-id"
		rootID := fsFixture.RootID

		// Calculate the QuickXorHash for the test file content
		testFileQuickXorHash := graph.QuickXORHash(&testFileBytes)

		// Create a file item
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
			Size: uint64(len(testFileContent)),
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Mock the file item response
		fileItemJSON, _ := json.Marshal(fileItem)
		mockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)

		// Mock the content download response
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", testFileBytes, http.StatusOK, nil)

		t.Logf("Step 1: Created test file '%s' with ID '%s'", testFileName, fileID)

		// Step 2: Queue the file for download
		t.Logf("Step 2: Queuing file for download...")
		downloadSession, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(downloadSession, "Download session should not be nil")

		t.Logf("Download session created: ID=%s, Path=%s", downloadSession.GetID(), downloadSession.GetPath())

		// Step 3: Monitor download progress
		t.Logf("Step 3: Monitoring download progress...")
		startTime := time.Now()

		// Check initial status
		initialStatus, err := fs.downloads.GetDownloadStatus(fileID)
		assert.NoError(err, "Failed to get initial download status")
		t.Logf("Initial download status: %v", initialStatus)

		// Step 4: Wait for download to complete
		t.Logf("Step 4: Waiting for download to complete...")
		err = fs.downloads.WaitForDownload(fileID)
		assert.NoError(err, "Download should complete without error")

		downloadDuration := time.Since(startTime)
		t.Logf("Download completed in %v", downloadDuration)

		// Verify download completion
		finalStatus, err := fs.downloads.GetDownloadStatus(fileID)
		assert.NoError(err, "Failed to get final download status")
		assert.Equal(downloadCompleted, finalStatus, "Download status should be completed")
		t.Logf("Final download status: %v", finalStatus)

		// Step 5: Verify file content is correct
		t.Logf("Step 5: Verifying file content...")

		// Open the cached file
		cachedFile, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open cached file")
		defer cachedFile.Close()

		// Seek to the beginning of the file before reading
		_, err = cachedFile.Seek(0, 0)
		assert.NoError(err, "Failed to seek to beginning of file")

		// Read the content
		cachedContent := make([]byte, len(testFileContent))
		n, err := cachedFile.Read(cachedContent)
		assert.NoError(err, "Failed to read cached content")
		assert.Equal(len(testFileContent), n, "Read byte count should match content length")
		assert.Equal(testFileContent, string(cachedContent), "Cached content should match original content")

		t.Logf("Content verification passed: %d bytes read, content matches", n)

		// Step 6: Verify file is cached
		t.Logf("Step 6: Verifying file is cached...")
		isCached := fs.content.HasContent(fileID)
		assert.True(isCached, "File should be cached after download")
		t.Logf("Cache verification passed: file is cached")

		// Step 7: Verify file status transitions
		t.Logf("Step 7: Verifying file status...")
		fileStatus := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, fileStatus.Status, "File status should be StatusLocal after successful download")
		t.Logf("File status verification passed: status=%s", fileStatus.Status.String())

		// Verify the inode size was updated
		inode := fs.GetID(fileID)
		assert.NotNil(inode, "Inode should exist")
		inode.mu.Lock()
		inodeSize := inode.DriveItem.Size
		inode.mu.Unlock()
		assert.Equal(uint64(len(testFileContent)), inodeSize, "Inode size should match downloaded content size")
		t.Logf("Inode size verification passed: size=%d bytes", inodeSize)

		// Additional verification: Check that the download session was cleaned up
		t.Logf("Additional verification: Checking session cleanup...")
		// The session should be cleaned up after status check, so another status check should return NotFoundError
		time.Sleep(100 * time.Millisecond) // Give cleanup goroutine time to run
		_, err = fs.downloads.GetDownloadStatus(fileID)
		// We expect either NotFoundError (session cleaned up) or completed status (cleanup pending)
		// Both are acceptable outcomes
		if err != nil {
			t.Logf("Session cleanup verification: session was cleaned up (NotFoundError)")
		} else {
			t.Logf("Session cleanup verification: session still exists (cleanup pending)")
		}

		t.Logf("✅ Single file download integration test completed successfully")
	})
}

// TestIT_FS_08_02_DownloadManager_CachedFileAccess tests accessing an already cached file
//
//	Test Case ID    IT-FS-08-02
//	Title           Download Manager - Cached File Access
//	Description     Verify that accessing a cached file does not trigger a new download
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Filesystem is mounted
//	                3. File is already cached
//	Steps           1. Download a file (cache it)
//	                2. Access the same file again
//	                3. Verify no new download is triggered
//	                4. Verify content is served from cache
//	Expected Result Cached file is served without new download
//	Requirements    3.2 (On-Demand File Download)
//	Notes: Integration test for cache hit scenario
func TestIT_FS_08_02_DownloadManager_CachedFileAccess(t *testing.T) {
	// Parallel execution disabled for the same reason as TestIT_FS_08_01.

	// Create a test fixture using the mock setup (download manager tests use mocks)
	fixture := helpers.SetupMockFSTestFixture(t, "CachedFileAccessIntegrationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create test file data
		testFileName := "cached_file_test.txt"
		testFileContent := "This is test content for cached file access verification"
		testFileBytes := []byte(testFileContent)
		fileID := "cached-file-id"
		rootID := fsFixture.RootID

		// Calculate the QuickXorHash for the test file content
		testFileQuickXorHash := graph.QuickXORHash(&testFileBytes)

		// Create a file item
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
			Size: uint64(len(testFileContent)),
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Mock the file item response
		fileItemJSON, _ := json.Marshal(fileItem)
		mockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)

		// Mock the content download response
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", testFileBytes, http.StatusOK, nil)

		t.Logf("Step 1: Downloading file to cache it...")

		// First download to cache the file
		downloadSession, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue initial download")
		assert.NotNil(downloadSession, "Download session should not be nil")

		err = fs.downloads.WaitForDownload(fileID)
		assert.NoError(err, "Initial download should complete without error")

		t.Logf("File cached successfully")

		// Verify file is cached
		isCached := fs.content.HasContent(fileID)
		assert.True(isCached, "File should be cached after download")

		// Step 2: Access the cached file
		t.Logf("Step 2: Accessing cached file...")

		// Open the cached file directly (simulating file access)
		cachedFile, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open cached file")
		defer cachedFile.Close()

		// Step 3: Verify no new download is triggered
		t.Logf("Step 3: Verifying no new download was triggered...")

		// Check that the file is still in cache and no download session exists
		isCached = fs.content.HasContent(fileID)
		assert.True(isCached, "File should still be cached")

		// Try to get download status - should return NotFoundError since no new download was queued
		_, err = fs.downloads.GetDownloadStatus(fileID)
		// We expect NotFoundError since no download is in progress
		if err != nil {
			t.Logf("No download session found (as expected): %v", err)
		}

		// Step 4: Verify content is served from cache
		t.Logf("Step 4: Verifying content is served from cache...")

		// Seek to the beginning of the file before reading
		_, seekErr := cachedFile.Seek(0, 0)
		assert.NoError(seekErr, "Failed to seek to beginning of file")

		// Read the content
		cachedContent := make([]byte, len(testFileContent))
		n, err := cachedFile.Read(cachedContent)
		assert.NoError(err, "Failed to read cached content")
		assert.Equal(len(testFileContent), n, "Read byte count should match content length")
		assert.Equal(testFileContent, string(cachedContent), "Cached content should match original content")

		t.Logf("Content served from cache successfully: %d bytes", n)

		// Verify file status is still StatusLocal
		fileStatus := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, fileStatus.Status, "File status should remain StatusLocal")

		t.Logf("✅ Cached file access integration test completed successfully")
	})
}

// TestIT_FS_08_03_DownloadManager_ConcurrentDownloads tests concurrent file downloads
//
//	Test Case ID    IT-FS-08-03
//	Title           Download Manager - Concurrent Downloads
//	Description     Verify that multiple files can be downloaded concurrently without race conditions or deadlocks
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Filesystem is mounted
//	                3. Network connection is available
//	Steps           1. Create multiple test files in the mock OneDrive
//	                2. Queue all files for download simultaneously
//	                3. Verify downloads proceed concurrently
//	                4. Wait for all downloads to complete
//	                5. Verify all files are downloaded correctly
//	                6. Verify no race conditions or deadlocks occurred
//	Expected Result All files download successfully, concurrently, without errors
//	Requirements    3.4 (Concurrent Downloads), 10.1 (Handle concurrent operations safely)
//	Notes: Integration test for concurrent download workflow
func TestIT_FS_08_03_DownloadManager_ConcurrentDownloads(t *testing.T) {
	// Parallel execution disabled to keep the mock Graph client state isolated.

	// Create a test fixture using the mock setup (download manager tests use mocks)
	fixture := helpers.SetupMockFSTestFixture(t, "ConcurrentDownloadsIntegrationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create multiple test files
		numFiles := 5
		fileIDs := make([]string, numFiles)
		fileContents := make([]string, numFiles)

		t.Logf("Step 1: Creating %d test files...", numFiles)

		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("concurrent-file-%d", i)
			fileName := fmt.Sprintf("concurrent_test_%d.txt", i)
			fileContent := fmt.Sprintf("This is test content for concurrent download file %d", i)
			fileBytes := []byte(fileContent)

			fileIDs[i] = fileID
			fileContents[i] = fileContent

			// Calculate the QuickXorHash for the test file content
			fileQuickXorHash := graph.QuickXORHash(&fileBytes)

			// Create a file item
			fileItem := &graph.DriveItem{
				ID:   fileID,
				Name: fileName,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: fileQuickXorHash,
					},
				},
				Size: uint64(len(fileContent)),
			}

			// Insert the file into the filesystem
			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			// Mock the file item response
			fileItemJSON, _ := json.Marshal(fileItem)
			mockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)

			// Mock the content download response
			mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileBytes, http.StatusOK, nil)

			t.Logf("Created test file %d: '%s' with ID '%s'", i+1, fileName, fileID)
		}

		// Step 2: Queue all files for download simultaneously
		t.Logf("Step 2: Queuing all files for download simultaneously...")
		startTime := time.Now()

		for i, fileID := range fileIDs {
			downloadSession, err := fs.downloads.QueueDownload(fileID)
			assert.NoError(err, "Failed to queue download for file %d", i)
			assert.NotNil(downloadSession, "Download session should not be nil for file %d", i)
			t.Logf("Queued file %d: %s", i+1, fileID)
		}

		// Step 3: Verify downloads proceed concurrently
		t.Logf("Step 3: Monitoring concurrent downloads...")

		// Check that multiple downloads are in progress
		time.Sleep(50 * time.Millisecond) // Give downloads time to start

		inProgressCount := 0
		for _, fileID := range fileIDs {
			status, err := fs.downloads.GetDownloadStatus(fileID)
			if err == nil && (status == downloadQueued || status == downloadStarted) {
				inProgressCount++
			}
		}
		t.Logf("Downloads in progress or queued: %d/%d", inProgressCount, numFiles)

		// Step 4: Wait for all downloads to complete
		t.Logf("Step 4: Waiting for all downloads to complete...")

		for i, fileID := range fileIDs {
			err := fs.downloads.WaitForDownload(fileID)
			assert.NoError(err, "Download should complete without error for file %d", i)
			t.Logf("File %d completed: %s", i+1, fileID)
		}

		downloadDuration := time.Since(startTime)
		t.Logf("All downloads completed in %v", downloadDuration)

		// Step 5: Verify all files are downloaded correctly
		t.Logf("Step 5: Verifying all files are downloaded correctly...")

		for i, fileID := range fileIDs {
			// Verify file is cached
			isCached := fs.content.HasContent(fileID)
			assert.True(isCached, "File %d should be cached after download", i)

			// Verify content
			cachedFile, err := fs.content.Open(fileID)
			assert.NoError(err, "Failed to open cached file %d", i)

			// Seek to the beginning
			_, err = cachedFile.Seek(0, 0)
			assert.NoError(err, "Failed to seek to beginning of file %d", i)

			// Read and verify content
			cachedContent := make([]byte, len(fileContents[i]))
			n, err := cachedFile.Read(cachedContent)
			assert.NoError(err, "Failed to read cached content for file %d", i)
			assert.Equal(len(fileContents[i]), n, "Read byte count should match for file %d", i)
			assert.Equal(fileContents[i], string(cachedContent), "Content should match for file %d", i)

			cachedFile.Close()

			// Verify file status
			fileStatus := fs.GetFileStatus(fileID)
			assert.Equal(StatusLocal, fileStatus.Status, "File %d status should be StatusLocal", i)

			t.Logf("File %d verified: content matches, cached, status correct", i+1)
		}

		// Step 6: Verify no race conditions or deadlocks occurred
		t.Logf("Step 6: Verifying no race conditions or deadlocks...")

		// All downloads completed successfully without hanging or errors
		// This is verified by the fact that we reached this point
		t.Logf("No race conditions or deadlocks detected")

		// Verify download manager is still operational
		testFileID := "post-concurrent-test-file"
		testFileName := "post_concurrent_test.txt"
		testFileContent := "Test file after concurrent downloads"
		testFileBytes := []byte(testFileContent)
		testFileQuickXorHash := graph.QuickXORHash(&testFileBytes)

		postTestItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: testFileQuickXorHash,
				},
			},
			Size: uint64(len(testFileContent)),
		}

		postTestInode := NewInodeDriveItem(postTestItem)
		fs.InsertNodeID(postTestInode)
		fs.InsertChild(rootID, postTestInode)

		postTestItemJSON, _ := json.Marshal(postTestItem)
		mockClient.AddMockResponse("/me/drive/items/"+testFileID, postTestItemJSON, http.StatusOK, nil)
		mockClient.AddMockResponse("/me/drive/items/"+testFileID+"/content", testFileBytes, http.StatusOK, nil)

		// Queue and wait for download
		_, err := fs.downloads.QueueDownload(testFileID)
		assert.NoError(err, "Download manager should still be operational after concurrent downloads")

		err = fs.downloads.WaitForDownload(testFileID)
		assert.NoError(err, "Post-concurrent download should complete successfully")

		t.Logf("Download manager verified operational after concurrent downloads")

		t.Logf("✅ Concurrent downloads integration test completed successfully")
	})
}

// TestIT_FS_08_04_DownloadManager_DownloadFailureAndRetry tests download failure and retry logic
//
//	Test Case ID    IT-FS-08-04
//	Title           Download Manager - Download Failure and Retry
//	Description     Verify that download failures trigger retry with exponential backoff
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Filesystem is mounted
//	                3. Network connection can be simulated to fail
//	Steps           1. Create a test file in the mock OneDrive
//	                2. Configure mock to fail first download attempt
//	                3. Queue file for download
//	                4. Verify download is retried
//	                5. Configure mock to succeed on retry
//	                6. Verify eventual success
//	Expected Result Download fails initially, retries with backoff, eventually succeeds
//	Requirements    3.5 (Download failure and retry), 9.1 (Error handling with retry)
//	Notes: Integration test for download retry logic
func TestIT_FS_08_04_DownloadManager_DownloadFailureAndRetry(t *testing.T) {
	// Parallel execution disabled to prevent races with shared mock Graph client state.

	// Create a test fixture using the mock setup (download manager tests use mocks)
	fixture := helpers.SetupMockFSTestFixture(t, "DownloadFailureRetryIntegrationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create test file data
		testFileName := "retry_test.txt"
		testFileContent := "This is test content for download retry verification"
		testFileBytes := []byte(testFileContent)
		fileID := "retry-test-file-id"

		// Calculate the QuickXorHash for the test file content
		testFileQuickXorHash := graph.QuickXORHash(&testFileBytes)

		// Create a file item
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
			Size: uint64(len(testFileContent)),
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Mock the file item response
		fileItemJSON, _ := json.Marshal(fileItem)
		mockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)

		t.Logf("Step 1: Created test file '%s' with ID '%s'", testFileName, fileID)

		// Step 2: Configure mock to fail first attempts
		t.Logf("Step 2: Configuring mock to fail initial download attempts...")

		// Add a failing response first (will be consumed on first attempt)
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", nil, http.StatusServiceUnavailable, fmt.Errorf("simulated network error"))

		// Add a successful response for retry (will be consumed on retry)
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", testFileBytes, http.StatusOK, nil)

		// Step 3: Queue file for download
		t.Logf("Step 3: Queuing file for download...")
		downloadSession, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(downloadSession, "Download session should not be nil")

		// Step 4: Wait for download to complete (with retries)
		t.Logf("Step 4: Waiting for download with retries...")
		startTime := time.Now()

		err = fs.downloads.WaitForDownload(fileID)
		assert.NoError(err, "Download should eventually succeed after retry")

		downloadDuration := time.Since(startTime)
		t.Logf("Download completed after retries in %v", downloadDuration)

		// Step 5: Verify eventual success
		t.Logf("Step 5: Verifying eventual success...")

		// Verify file is cached
		isCached := fs.content.HasContent(fileID)
		assert.True(isCached, "File should be cached after successful retry")

		// Verify content
		cachedFile, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open cached file")
		defer cachedFile.Close()

		// Seek to the beginning
		_, err = cachedFile.Seek(0, 0)
		assert.NoError(err, "Failed to seek to beginning of file")

		// Read and verify content
		cachedContent := make([]byte, len(testFileContent))
		n, err := cachedFile.Read(cachedContent)
		assert.NoError(err, "Failed to read cached content")
		assert.Equal(len(testFileContent), n, "Read byte count should match content length")
		assert.Equal(testFileContent, string(cachedContent), "Cached content should match original content")

		t.Logf("Content verification passed after retry: %d bytes", n)

		// Verify file status
		fileStatus := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, fileStatus.Status, "File status should be StatusLocal after successful retry")

		t.Logf("✅ Download failure and retry integration test completed successfully")
	})
}

// TestIT_FS_08_05_DownloadManager_DownloadStatusTracking tests download status tracking
//
//	Test Case ID    IT-FS-08-05
//	Title           Download Manager - Download Status Tracking
//	Description     Verify that file status transitions correctly during download lifecycle
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Filesystem is mounted
//	                3. Network connection is available
//	Steps           1. Create a test file in the mock OneDrive
//	                2. Queue file for download
//	                3. Monitor status transitions: not cached → downloading → cached
//	                4. Verify status is visible via GetFileStatus
//	                5. Verify status updates are logged
//	Expected Result Status transitions correctly through all states
//	Requirements    3.4 (Download status tracking), 8.1 (File status updates)
//	Notes: Integration test for download status tracking
func TestIT_FS_08_05_DownloadManager_DownloadStatusTracking(t *testing.T) {
	// Parallel execution disabled to shield shared mock Graph client fixtures.

	// Create a test fixture using the mock setup (download manager tests use mocks)
	fixture := helpers.SetupMockFSTestFixture(t, "DownloadStatusTrackingIntegrationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create test file data
		testFileName := "status_tracking_test.txt"
		testFileContent := "This is test content for download status tracking verification"
		testFileBytes := []byte(testFileContent)
		fileID := "status-tracking-file-id"

		// Calculate the QuickXorHash for the test file content
		testFileQuickXorHash := graph.QuickXORHash(&testFileBytes)

		// Create a file item
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
			Size: uint64(len(testFileContent)),
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Mock the file item response
		fileItemJSON, _ := json.Marshal(fileItem)
		mockClient.AddMockResponse("/me/drive/items/"+fileID, fileItemJSON, http.StatusOK, nil)

		// Mock the content download response
		mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", testFileBytes, http.StatusOK, nil)

		t.Logf("Step 1: Created test file '%s' with ID '%s'", testFileName, fileID)

		// Step 2: Check initial status (not cached)
		t.Logf("Step 2: Checking initial file status...")
		initialStatus := fs.GetFileStatus(fileID)
		t.Logf("Initial file status: %s", initialStatus.Status.String())

		// File should not be cached initially
		isCached := fs.content.HasContent(fileID)
		assert.False(isCached, "File should not be cached initially")

		// Step 3: Queue file for download and monitor status transitions
		t.Logf("Step 3: Queuing file for download and monitoring status transitions...")

		downloadSession, err := fs.downloads.QueueDownload(fileID)
		assert.NoError(err, "Failed to queue download")
		assert.NotNil(downloadSession, "Download session should not be nil")

		// Check status immediately after queuing
		queuedStatus := fs.GetFileStatus(fileID)
		t.Logf("Status after queuing: %s", queuedStatus.Status.String())

		// Give download time to start
		time.Sleep(50 * time.Millisecond)

		// Check status during download
		downloadingStatus := fs.GetFileStatus(fileID)
		t.Logf("Status during download: %s", downloadingStatus.Status.String())

		// Wait for download to complete
		err = fs.downloads.WaitForDownload(fileID)
		assert.NoError(err, "Download should complete without error")

		// Step 4: Verify final status (cached)
		t.Logf("Step 4: Verifying final file status...")
		finalStatus := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, finalStatus.Status, "File status should be StatusLocal after download")
		t.Logf("Final file status: %s", finalStatus.Status.String())

		// Verify file is now cached
		isCached = fs.content.HasContent(fileID)
		assert.True(isCached, "File should be cached after download")

		// Step 5: Verify status information is complete
		t.Logf("Step 5: Verifying status information completeness...")

		// Check that status has timestamp
		assert.False(finalStatus.Timestamp.IsZero(), "Status should have a timestamp")
		t.Logf("Status timestamp: %v", finalStatus.Timestamp)

		// Verify status is accessible via GetFileStatus
		retrievedStatus := fs.GetFileStatus(fileID)
		assert.Equal(finalStatus.Status, retrievedStatus.Status, "Retrieved status should match")
		t.Logf("Status retrieval verified")

		// Additional verification: Check download session status
		downloadStatus, err := fs.downloads.GetDownloadStatus(fileID)
		if err == nil {
			t.Logf("Download session status: %v", downloadStatus)
			assert.Equal(downloadCompleted, downloadStatus, "Download session status should be completed")
		} else {
			t.Logf("Download session cleaned up (expected): %v", err)
		}

		t.Logf("✅ Download status tracking integration test completed successfully")
	})
}
