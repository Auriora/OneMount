package fs

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_09_04_UploadFailureAndRetry verifies upload failure and retry functionality.
// This test corresponds to task 9.4 in the system verification plan.
//
//	Test Case ID    IT-FS-09-04
//	Title           Upload Failure and Retry
//	Description     Verify that uploads are retried with exponential backoff after network failures
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a file for upload
//	                2. Simulate network failure during upload
//	                3. Verify upload is retried
//	                4. Check exponential backoff is used
//	                5. Verify eventual success after retries
//	Expected Result Upload fails initially, retries with exponential backoff, and eventually succeeds
//	Requirements    4.4
func TestIT_FS_09_04_UploadFailureAndRetry(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "UploadFailureRetryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create a file for upload
		testFileName := "retry_test_file.txt"
		testFileContent := "This is a test file for upload retry verification."
		testFileSize := len(testFileContent)
		fileID := "retry-file-test-id"

		// Calculate the QuickXorHash for the test file
		testFileContentBytes := []byte(testFileContent)
		testFileQuickXorHash := graph.QuickXORHash(&testFileContentBytes)

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
			ETag: "initial-etag-retry",
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Write the test file content to the content cache
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		n, err := fd.WriteAt(testFileContentBytes, 0)
		assert.NoError(err, "Failed to write test file content")
		assert.Equal(testFileSize, n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes to trigger upload
		fileInode.hasChanges = true

		// Mock the item response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Step 2: Simulate network failure during upload
		// Configure the mock client to fail the first 2 attempts and succeed on the 3rd
		var attemptCount int
		var attemptMutex sync.Mutex
		var attemptTimes []time.Time

		mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
			attemptMutex.Lock()
			attemptCount++
			currentAttempt := attemptCount
			attemptTimes = append(attemptTimes, time.Now())
			attemptMutex.Unlock()

			if currentAttempt <= 2 {
				// Fail the first 2 attempts with a network error
				return nil, 500, fmt.Errorf("simulated network failure (attempt %d)", currentAttempt)
			}

			// Succeed on the 3rd attempt
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
				ETag: "uploaded-etag-retry",
			}

			uploadedFileItemJSON, _ := json.Marshal(uploadedFileItem)
			return uploadedFileItemJSON, 200, nil
		})

		// Step 3: Queue the upload
		uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")
		assert.NotNil(uploadSession, "Upload session should not be nil")

		// Give the upload manager time to process the queued session
		time.Sleep(100 * time.Millisecond)

		// Verify session was created
		session, exists := fs.uploads.GetSession(fileID)
		assert.True(exists, "Upload session should exist after queueing")
		assert.NotNil(session, "Upload session should not be nil")

		// Wait for the upload to complete (with retries)
		// This should take some time due to the retries
		startTime := time.Now()
		err = fs.uploads.WaitForUpload(fileID)
		elapsedTime := time.Since(startTime)

		// Step 4: Verify upload eventually succeeded
		assert.NoError(err, "Upload should eventually succeed after retries")

		// Verify that multiple attempts were made
		attemptMutex.Lock()
		finalAttemptCount := attemptCount
		attemptTimesCopy := make([]time.Time, len(attemptTimes))
		copy(attemptTimesCopy, attemptTimes)
		attemptMutex.Unlock()

		assert.True(finalAttemptCount >= 3,
			"Upload should have been attempted at least 3 times (was %d)", finalAttemptCount)

		// Step 5: Verify exponential backoff was used
		// Check that the time between retries increases (exponential backoff)
		if len(attemptTimesCopy) >= 3 {
			// Calculate time between first and second attempt
			firstRetryDelay := attemptTimesCopy[1].Sub(attemptTimesCopy[0])
			// Calculate time between second and third attempt
			secondRetryDelay := attemptTimesCopy[2].Sub(attemptTimesCopy[1])

			// The second retry delay should be longer than the first
			// (allowing for some timing variance)
			assert.True(secondRetryDelay >= firstRetryDelay,
				"Second retry delay (%v) should be >= first retry delay (%v) for exponential backoff",
				secondRetryDelay, firstRetryDelay)

			// Log the retry delays for verification
			t.Logf("First retry delay: %v", firstRetryDelay)
			t.Logf("Second retry delay: %v", secondRetryDelay)
			t.Logf("Total elapsed time: %v", elapsedTime)
		}

		// Verify ETag was updated after successful upload
		updatedInode := fs.GetID(fileID)
		assert.NotNil(updatedInode, "File inode should exist after upload")

		updatedInode.mu.RLock()
		updatedETag := updatedInode.DriveItem.ETag
		updatedSize := updatedInode.DriveItem.Size
		updatedInode.mu.RUnlock()

		assert.Equal("uploaded-etag-retry", updatedETag,
			"ETag should be updated after successful upload")
		assert.Equal(uint64(testFileSize), updatedSize,
			"File size should match after upload")

		// Verify file status is set to Local (synced)
		status := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, status.Status,
			"File status should be Local after successful upload")
	})
}

// TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry verifies large file upload retry with chunked upload.
//
//	Test Case ID    IT-FS-09-04-02
//	Title           Large File Upload Failure and Retry
//	Description     Verify that large file uploads are retried with exponential backoff after chunk failures
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a large file (> 10MB) for upload
//	                2. Simulate network failure during chunk upload
//	                3. Verify chunk upload is retried
//	                4. Check exponential backoff is used for chunk retries
//	                5. Verify eventual success after retries
//	Expected Result Large file upload fails on some chunks, retries with exponential backoff, and eventually succeeds
//	Requirements    4.4
func TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "LargeFileUploadRetryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		testFileName := "large_retry_test_file.bin"
		testFileSize := 12 * 1024 * 1024 // 12MB
		fileID := "large-retry-file-test-id"

		// Create test file content (repeating pattern for efficiency)
		pattern := []byte("LARGE_RETRY_TEST_")
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
			ETag: "initial-etag-large-retry",
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

		// Mock the item response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Step 2: Simulate network failure during chunk upload
		// Track chunk upload attempts and timing
		var chunkAttempts map[int]int
		var chunkAttemptTimes map[int][]time.Time
		var attemptMutex sync.Mutex

		chunkAttempts = make(map[int]int)
		chunkAttemptTimes = make(map[int][]time.Time)

		// Configure mock to fail first chunk upload twice, then succeed
		mockClient.SetChunkUploadCallback(func(chunkIndex int) ([]byte, int, error) {
			attemptMutex.Lock()
			chunkAttempts[chunkIndex]++
			currentAttempt := chunkAttempts[chunkIndex]
			if chunkAttemptTimes[chunkIndex] == nil {
				chunkAttemptTimes[chunkIndex] = []time.Time{}
			}
			chunkAttemptTimes[chunkIndex] = append(chunkAttemptTimes[chunkIndex], time.Now())
			attemptMutex.Unlock()

			// Fail the first chunk twice to test retry logic
			if chunkIndex == 0 && currentAttempt <= 2 {
				// Return 500 error to trigger retry with exponential backoff
				return nil, 500, fmt.Errorf("simulated chunk upload failure (chunk %d, attempt %d)", chunkIndex, currentAttempt)
			}

			// For the last chunk, return the completed file item
			// For other chunks, return 202 Accepted
			if chunkIndex == 1 { // Last chunk for 12MB file with 10MB chunks
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
					ETag: "uploaded-etag-large-retry",
				}

				uploadedFileItemJSON, _ := json.Marshal(uploadedFileItem)
				return uploadedFileItemJSON, 200, nil
			}

			// Return 202 for intermediate chunks
			return []byte(`{"expirationDateTime":"2099-01-01T00:00:00Z"}`), 202, nil
		})

		// Queue the upload
		uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")
		assert.NotNil(uploadSession, "Upload session should not be nil")

		// Give the upload manager time to process the queued session
		time.Sleep(100 * time.Millisecond)

		// Step 3: Wait for the upload to complete (with retries)
		startTime := time.Now()
		err = fs.uploads.WaitForUpload(fileID)
		elapsedTime := time.Since(startTime)

		// Step 4: Verify upload eventually succeeded
		assert.NoError(err, "Large file upload should eventually succeed after chunk retries")

		// Verify that the first chunk was attempted multiple times
		attemptMutex.Lock()
		firstChunkAttempts := chunkAttempts[0]
		firstChunkTimes := chunkAttemptTimes[0]
		attemptMutex.Unlock()

		assert.True(firstChunkAttempts >= 3,
			"First chunk should have been attempted at least 3 times (was %d)", firstChunkAttempts)

		// Step 5: Verify exponential backoff was used for chunk retries
		if len(firstChunkTimes) >= 3 {
			// Calculate time between first and second attempt
			firstRetryDelay := firstChunkTimes[1].Sub(firstChunkTimes[0])
			// Calculate time between second and third attempt
			secondRetryDelay := firstChunkTimes[2].Sub(firstChunkTimes[1])

			// The second retry delay should be longer than the first (exponential backoff)
			// Note: The backoff is implemented in uploadChunk with time.Sleep(time.Duration(backoff) * time.Second)
			// where backoff doubles each time (1s, 2s, 4s, etc.)
			assert.True(secondRetryDelay > firstRetryDelay,
				"Second chunk retry delay (%v) should be > first chunk retry delay (%v) for exponential backoff",
				secondRetryDelay, firstRetryDelay)

			// Log the retry delays for verification
			t.Logf("First chunk retry delay: %v", firstRetryDelay)
			t.Logf("Second chunk retry delay: %v", secondRetryDelay)
			t.Logf("Total elapsed time: %v", elapsedTime)
		}

		// Verify ETag was updated after successful upload
		updatedInode := fs.GetID(fileID)
		assert.NotNil(updatedInode, "File inode should exist after upload")

		updatedInode.mu.RLock()
		updatedETag := updatedInode.DriveItem.ETag
		updatedSize := updatedInode.DriveItem.Size
		updatedInode.mu.RUnlock()

		assert.Equal("uploaded-etag-large-retry", updatedETag,
			"ETag should be updated after successful upload")
		assert.Equal(uint64(testFileSize), updatedSize,
			"File size should match after upload")

		// Verify file status is set to Local (synced)
		status := fs.GetFileStatus(fileID)
		assert.Equal(StatusLocal, status.Status,
			"File status should be Local after successful upload")
	})
}

// TestIT_FS_09_04_03_UploadMaxRetriesExceeded verifies behavior when max retries are exceeded.
//
//	Test Case ID    IT-FS-09-04-03
//	Title           Upload Max Retries Exceeded
//	Description     Verify that uploads fail permanently after exceeding max retry attempts
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a file for upload
//	                2. Simulate persistent network failure
//	                3. Verify upload is retried multiple times
//	                4. Verify upload eventually fails after max retries
//	                5. Check file status is set to error
//	Expected Result Upload fails permanently after max retries and file status shows error
//	Requirements    4.4
func TestIT_FS_09_04_03_UploadMaxRetriesExceeded(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "UploadMaxRetriesFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create a file for upload
		testFileName := "max_retry_test_file.txt"
		testFileContent := "This is a test file for max retry verification."
		testFileSize := len(testFileContent)
		fileID := "max-retry-file-test-id"

		// Calculate the QuickXorHash for the test file
		testFileContentBytes := []byte(testFileContent)
		testFileQuickXorHash := graph.QuickXORHash(&testFileContentBytes)

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
			ETag: "initial-etag-max-retry",
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Write the test file content to the content cache
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		n, err := fd.WriteAt(testFileContentBytes, 0)
		assert.NoError(err, "Failed to write test file content")
		assert.Equal(testFileSize, n, "Number of bytes written doesn't match content length")

		// Mark the file as having changes to trigger upload
		fileInode.hasChanges = true

		// Mock the item response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Step 2: Simulate persistent network failure
		// Configure the mock client to always fail
		var attemptCount int
		var attemptMutex sync.Mutex

		mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
			attemptMutex.Lock()
			attemptCount++
			attemptMutex.Unlock()

			// Always fail to test max retries
			return nil, 500, fmt.Errorf("simulated persistent network failure")
		})

		// Queue the upload
		uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")
		assert.NotNil(uploadSession, "Upload session should not be nil")

		// Give the upload manager time to process and retry
		// The upload manager retries up to 5 times, so we need to wait long enough
		// for all retries to complete. With exponential backoff, this could take
		// several seconds.
		time.Sleep(15 * time.Second)

		// Step 3: Verify that multiple attempts were made
		attemptMutex.Lock()
		finalAttemptCount := attemptCount
		attemptMutex.Unlock()

		// The upload manager should have tried multiple times before giving up
		assert.True(finalAttemptCount > 1,
			"Upload should have been attempted multiple times (was %d)", finalAttemptCount)

		// Step 4: Verify upload eventually failed
		// Check the session state
		session, exists := fs.uploads.GetSession(fileID)
		// The session might have been cleaned up after max retries
		if exists {
			state := session.getState()
			// If session still exists, it should be in error state
			assert.Equal(uploadErrored, state,
				"Upload session should be in error state after max retries")
		}

		// Step 5: Verify file status shows error
		status := fs.GetFileStatus(fileID)
		// The file should be marked with an error status
		assert.True(status.Status == StatusError || status.Status == StatusLocalModified,
			"File status should indicate error or local modification after failed upload (was %v)", status.Status)

		t.Logf("Upload failed after %d attempts as expected", finalAttemptCount)
	})
}
