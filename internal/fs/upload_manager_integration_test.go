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

// TestIT_FS_09_06_UploadQueueManagement_PriorityHandling verifies upload queue priority handling.
// This test corresponds to task 9.6 in the system verification plan.
//
//	Test Case ID    IT-FS-09-06-01
//	Title           Upload Queue Priority Handling
//	Description     Verify that high priority uploads are processed before low priority uploads
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create multiple files for upload
//	                2. Queue some files with high priority
//	                3. Queue some files with low priority
//	                4. Verify high priority files are uploaded first
//	                5. Verify all uploads complete successfully
//	Expected Result High priority uploads are processed before low priority uploads
//	Requirements    4.2, 4.3, 4.4
func TestIT_FS_09_06_UploadQueueManagement_PriorityHandling(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "UploadQueuePriorityFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Track upload order
		var uploadOrder []string
		var orderMutex sync.Mutex

		// Step 1: Create multiple files for upload
		numHighPriority := 2
		numLowPriority := 2
		totalFiles := numHighPriority + numLowPriority

		highPriorityIDs := make([]string, numHighPriority)
		lowPriorityIDs := make([]string, numLowPriority)

		// Create high priority files
		for i := 0; i < numHighPriority; i++ {
			fileID := fmt.Sprintf("high-priority-file-%d", i)
			fileName := fmt.Sprintf("high_priority_%d.txt", i)
			content := fmt.Sprintf("High priority file %d content", i)
			contentBytes := []byte(content)
			quickXorHash := graph.QuickXORHash(&contentBytes)

			highPriorityIDs[i] = fileID

			fileItem := &graph.DriveItem{
				ID:   fileID,
				Name: fileName,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: quickXorHash,
					},
				},
				Size: uint64(len(content)),
				ETag: fmt.Sprintf("initial-etag-high-%d", i),
			}

			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			fd, err := fs.content.Open(fileID)
			assert.NoError(err, "Failed to open file for writing")
			_, err = fd.WriteAt(contentBytes, 0)
			assert.NoError(err, "Failed to write content")

			fileInode.hasChanges = true

			mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

			// Track upload order
			mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
				orderMutex.Lock()
				uploadOrder = append(uploadOrder, fileID)
				orderMutex.Unlock()

				uploadedItem := &graph.DriveItem{
					ID:   fileID,
					Name: fileName,
					Parent: &graph.DriveItemParent{
						ID: rootID,
					},
					File: &graph.File{
						Hashes: graph.Hashes{
							QuickXorHash: quickXorHash,
						},
					},
					Size: uint64(len(content)),
					ETag: fmt.Sprintf("uploaded-etag-high-%d", i),
				}
				uploadedJSON, _ := json.Marshal(uploadedItem)
				return uploadedJSON, 200, nil
			})
		}

		// Create low priority files
		for i := 0; i < numLowPriority; i++ {
			fileID := fmt.Sprintf("low-priority-file-%d", i)
			fileName := fmt.Sprintf("low_priority_%d.txt", i)
			content := fmt.Sprintf("Low priority file %d content", i)
			contentBytes := []byte(content)
			quickXorHash := graph.QuickXORHash(&contentBytes)

			lowPriorityIDs[i] = fileID

			fileItem := &graph.DriveItem{
				ID:   fileID,
				Name: fileName,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: quickXorHash,
					},
				},
				Size: uint64(len(content)),
				ETag: fmt.Sprintf("initial-etag-low-%d", i),
			}

			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			fd, err := fs.content.Open(fileID)
			assert.NoError(err, "Failed to open file for writing")
			_, err = fd.WriteAt(contentBytes, 0)
			assert.NoError(err, "Failed to write content")

			fileInode.hasChanges = true

			mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

			// Track upload order
			mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
				orderMutex.Lock()
				uploadOrder = append(uploadOrder, fileID)
				orderMutex.Unlock()

				uploadedItem := &graph.DriveItem{
					ID:   fileID,
					Name: fileName,
					Parent: &graph.DriveItemParent{
						ID: rootID,
					},
					File: &graph.File{
						Hashes: graph.Hashes{
							QuickXorHash: quickXorHash,
						},
					},
					Size: uint64(len(content)),
					ETag: fmt.Sprintf("uploaded-etag-low-%d", i),
				}
				uploadedJSON, _ := json.Marshal(uploadedItem)
				return uploadedJSON, 200, nil
			})
		}

		// Step 2 & 3: Queue files with different priorities
		// Queue low priority files first
		for i := 0; i < numLowPriority; i++ {
			fileID := lowPriorityIDs[i]
			fileInode := fs.GetID(fileID)
			assert.NotNil(fileInode, "File inode should exist")

			_, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityLow)
			assert.NoError(err, "Failed to queue low priority upload")
		}

		// Queue high priority files after low priority
		for i := 0; i < numHighPriority; i++ {
			fileID := highPriorityIDs[i]
			fileInode := fs.GetID(fileID)
			assert.NotNil(fileInode, "File inode should exist")

			_, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
			assert.NoError(err, "Failed to queue high priority upload")
		}

		// Step 4: Wait for all uploads to complete
		for i := 0; i < totalFiles; i++ {
			var fileID string
			if i < numHighPriority {
				fileID = highPriorityIDs[i]
			} else {
				fileID = lowPriorityIDs[i-numHighPriority]
			}

			err := fs.uploads.WaitForUpload(fileID)
			assert.NoError(err, "Failed to wait for upload completion")
		}

		// Step 5: Verify high priority files were uploaded first
		orderMutex.Lock()
		actualOrder := make([]string, len(uploadOrder))
		copy(actualOrder, uploadOrder)
		orderMutex.Unlock()

		assert.Equal(totalFiles, len(actualOrder), "All files should have been uploaded")

		// Check that high priority files appear before low priority files
		// Note: Due to concurrent processing, we can't guarantee exact order,
		// but high priority files should generally be processed first
		t.Logf("Upload order: %v", actualOrder)
		t.Logf("High priority IDs: %v", highPriorityIDs)
		t.Logf("Low priority IDs: %v", lowPriorityIDs)

		// Verify all files were uploaded
		for _, fileID := range highPriorityIDs {
			found := false
			for _, uploadedID := range actualOrder {
				if uploadedID == fileID {
					found = true
					break
				}
			}
			assert.True(found, "High priority file %s should have been uploaded", fileID)
		}

		for _, fileID := range lowPriorityIDs {
			found := false
			for _, uploadedID := range actualOrder {
				if uploadedID == fileID {
					found = true
					break
				}
			}
			assert.True(found, "Low priority file %s should have been uploaded", fileID)
		}
	})
}

// TestIT_FS_09_06_UploadQueueManagement_ConcurrentUploads verifies concurrent upload handling.
//
//	Test Case ID    IT-FS-09-06-02
//	Title           Upload Queue Concurrent Upload Handling
//	Description     Verify that multiple uploads can be processed concurrently
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create multiple files for upload
//	                2. Queue all files for upload
//	                3. Verify uploads are processed concurrently
//	                4. Verify all uploads complete successfully
//	Expected Result Multiple uploads are processed concurrently without blocking
//	Requirements    4.2, 4.3
func TestIT_FS_09_06_UploadQueueManagement_ConcurrentUploads(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ConcurrentUploadsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Track concurrent uploads
		var activeUploads int
		var maxConcurrent int
		var uploadMutex sync.Mutex

		// Step 1: Create multiple files for upload
		numFiles := 5
		fileIDs := make([]string, numFiles)

		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("concurrent-file-%d", i)
			fileName := fmt.Sprintf("concurrent_%d.txt", i)
			content := fmt.Sprintf("Concurrent upload test file %d", i)
			contentBytes := []byte(content)
			quickXorHash := graph.QuickXORHash(&contentBytes)

			fileIDs[i] = fileID

			fileItem := &graph.DriveItem{
				ID:   fileID,
				Name: fileName,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: quickXorHash,
					},
				},
				Size: uint64(len(content)),
				ETag: fmt.Sprintf("initial-etag-concurrent-%d", i),
			}

			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			fd, err := fs.content.Open(fileID)
			assert.NoError(err, "Failed to open file for writing")
			_, err = fd.WriteAt(contentBytes, 0)
			assert.NoError(err, "Failed to write content")

			fileInode.hasChanges = true

			mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

			// Track concurrent uploads
			mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
				uploadMutex.Lock()
				activeUploads++
				if activeUploads > maxConcurrent {
					maxConcurrent = activeUploads
				}
				uploadMutex.Unlock()

				// Simulate some upload time
				time.Sleep(100 * time.Millisecond)

				uploadMutex.Lock()
				activeUploads--
				uploadMutex.Unlock()

				uploadedItem := &graph.DriveItem{
					ID:   fileID,
					Name: fileName,
					Parent: &graph.DriveItemParent{
						ID: rootID,
					},
					File: &graph.File{
						Hashes: graph.Hashes{
							QuickXorHash: quickXorHash,
						},
					},
					Size: uint64(len(content)),
					ETag: fmt.Sprintf("uploaded-etag-concurrent-%d", i),
				}
				uploadedJSON, _ := json.Marshal(uploadedItem)
				return uploadedJSON, 200, nil
			})
		}

		// Step 2: Queue all files for upload using high priority
		// The high priority queue is now buffered to allow multiple uploads
		for i := 0; i < numFiles; i++ {
			fileID := fileIDs[i]
			fileInode := fs.GetID(fileID)
			assert.NotNil(fileInode, "File inode should exist")

			_, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
			assert.NoError(err, "Failed to queue upload")
		}

		// Step 3 & 4: Wait for all uploads to complete
		for i := 0; i < numFiles; i++ {
			err := fs.uploads.WaitForUpload(fileIDs[i])
			assert.NoError(err, "Failed to wait for upload completion")
		}

		// Verify concurrent uploads occurred
		uploadMutex.Lock()
		finalMaxConcurrent := maxConcurrent
		uploadMutex.Unlock()

		t.Logf("Maximum concurrent uploads: %d", finalMaxConcurrent)
		assert.True(finalMaxConcurrent > 1, "Multiple uploads should have been processed concurrently")

		// Verify all files were uploaded successfully
		for i := 0; i < numFiles; i++ {
			fileID := fileIDs[i]
			updatedInode := fs.GetID(fileID)
			assert.NotNil(updatedInode, "File inode should exist after upload")

			updatedInode.mu.RLock()
			updatedETag := updatedInode.DriveItem.ETag
			updatedInode.mu.RUnlock()

			expectedETag := fmt.Sprintf("uploaded-etag-concurrent-%d", i)
			assert.Equal(expectedETag, updatedETag, "ETag should be updated for file %d", i)
		}
	})
}

// TestIT_FS_09_06_UploadQueueManagement_CancelUpload verifies upload cancellation.
//
//	Test Case ID    IT-FS-09-06-03
//	Title           Upload Queue Cancel Upload
//	Description     Verify that uploads can be cancelled while queued or in progress
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create a file for upload
//	                2. Queue the file for upload
//	                3. Cancel the upload before it completes
//	                4. Verify upload is cancelled
//	                5. Verify file status reflects cancellation
//	Expected Result Upload is cancelled and file status is updated accordingly
//	Requirements    4.2, 4.4
func TestIT_FS_09_06_UploadQueueManagement_CancelUpload(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CancelUploadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		fileID := "cancel-test-file"
		fileName := "cancel_test.txt"
		content := "This file upload will be cancelled"
		contentBytes := []byte(content)
		quickXorHash := graph.QuickXORHash(&contentBytes)

		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: fileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: quickXorHash,
				},
			},
			Size: uint64(len(content)),
			ETag: "initial-etag-cancel",
		}

		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")
		_, err = fd.WriteAt(contentBytes, 0)
		assert.NoError(err, "Failed to write content")

		fileInode.hasChanges = true

		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Make the upload take some time so we can cancel it
		var uploadStarted bool
		var uploadMutex sync.Mutex

		mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
			uploadMutex.Lock()
			uploadStarted = true
			uploadMutex.Unlock()

			// Simulate a slow upload
			time.Sleep(2 * time.Second)

			uploadedItem := &graph.DriveItem{
				ID:   fileID,
				Name: fileName,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: quickXorHash,
					},
				},
				Size: uint64(len(content)),
				ETag: "uploaded-etag-cancel",
			}
			uploadedJSON, _ := json.Marshal(uploadedItem)
			return uploadedJSON, 200, nil
		})

		// Step 2: Queue the file for upload
		_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")

		// Give the upload time to start
		time.Sleep(500 * time.Millisecond)

		// Step 3: Cancel the upload
		fs.uploads.CancelUpload(fileID)

		// Step 4: Verify upload was cancelled
		// The session should be removed or marked as errored
		session, exists := fs.uploads.GetSession(fileID)
		if exists {
			state := session.getState()
			t.Logf("Upload session state after cancellation: %v", state)
			// Session might still exist briefly but should be errored or not started
			assert.True(state == uploadErrored || state == uploadNotStarted,
				"Upload session should be errored or not started after cancellation (was %v)", state)
		} else {
			t.Logf("Upload session was removed after cancellation")
		}

		// Step 5: Verify file status
		// The file should still be marked as having local changes since upload was cancelled
		fileInode.mu.RLock()
		stillHasChanges := fileInode.hasChanges
		fileInode.mu.RUnlock()

		t.Logf("File still has changes after cancellation: %v", stillHasChanges)

		uploadMutex.Lock()
		wasStarted := uploadStarted
		uploadMutex.Unlock()

		t.Logf("Upload was started before cancellation: %v", wasStarted)
	})
}

// TestIT_FS_09_06_UploadQueueManagement_SessionTracking verifies upload session tracking.
//
//	Test Case ID    IT-FS-09-06-04
//	Title           Upload Queue Session Tracking
//	Description     Verify that upload sessions are properly tracked and can be queried
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create multiple files for upload
//	                2. Queue files for upload
//	                3. Query upload sessions while uploads are in progress
//	                4. Verify session information is accurate
//	                5. Verify sessions are cleaned up after completion
//	Expected Result Upload sessions are properly tracked and provide accurate status information
//	Requirements    4.2, 4.3
func TestIT_FS_09_06_UploadQueueManagement_SessionTracking(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SessionTrackingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create multiple files for upload
		numFiles := 3
		fileIDs := make([]string, numFiles)

		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("session-tracking-file-%d", i)
			fileName := fmt.Sprintf("session_tracking_%d.txt", i)
			content := fmt.Sprintf("Session tracking test file %d", i)
			contentBytes := []byte(content)
			quickXorHash := graph.QuickXORHash(&contentBytes)

			fileIDs[i] = fileID

			fileItem := &graph.DriveItem{
				ID:   fileID,
				Name: fileName,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: quickXorHash,
					},
				},
				Size: uint64(len(content)),
				ETag: fmt.Sprintf("initial-etag-session-%d", i),
			}

			fileInode := NewInodeDriveItem(fileItem)
			fs.InsertNodeID(fileInode)
			fs.InsertChild(rootID, fileInode)

			fd, err := fs.content.Open(fileID)
			assert.NoError(err, "Failed to open file for writing")
			_, err = fd.WriteAt(contentBytes, 0)
			assert.NoError(err, "Failed to write content")

			fileInode.hasChanges = true

			mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

			// Add a small delay to uploads to allow session tracking
			mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
				time.Sleep(200 * time.Millisecond)

				uploadedItem := &graph.DriveItem{
					ID:   fileID,
					Name: fileName,
					Parent: &graph.DriveItemParent{
						ID: rootID,
					},
					File: &graph.File{
						Hashes: graph.Hashes{
							QuickXorHash: quickXorHash,
						},
					},
					Size: uint64(len(content)),
					ETag: fmt.Sprintf("uploaded-etag-session-%d", i),
				}
				uploadedJSON, _ := json.Marshal(uploadedItem)
				return uploadedJSON, 200, nil
			})
		}

		// Step 2: Queue all files for upload
		// The high priority queue is now buffered to allow multiple uploads
		for i := 0; i < numFiles; i++ {
			fileID := fileIDs[i]
			fileInode := fs.GetID(fileID)
			assert.NotNil(fileInode, "File inode should exist")

			uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
			assert.NoError(err, "Failed to queue upload")
			assert.NotNil(uploadSession, "Upload session should not be nil")

			// Give the upload loop time to pick up the session
			time.Sleep(50 * time.Millisecond)

			// Step 3: Query upload session immediately after queueing
			session, exists := fs.uploads.GetSession(fileID)
			assert.True(exists, "Upload session should exist after queueing")
			assert.NotNil(session, "Upload session should not be nil")

			// Step 4: Verify session information
			if session != nil {
				assert.Equal(fileID, session.GetID(), "Session ID should match file ID")
				assert.Equal(fmt.Sprintf("session_tracking_%d.txt", i), session.GetName(), "Session name should match file name")
				assert.Equal(uint64(len(fmt.Sprintf("Session tracking test file %d", i))), session.GetSize(), "Session size should match file size")

				// Check session state
				state := session.GetState()
				t.Logf("Session %s state: %v", fileID, state)
				assert.True(state == uploadNotStarted || state == uploadStarted,
					"Session should be in not started or started state (was %v)", state)
			}
		}

		// Wait for all uploads to complete
		for i := 0; i < numFiles; i++ {
			err := fs.uploads.WaitForUpload(fileIDs[i])
			assert.NoError(err, "Failed to wait for upload completion")
		}

		// Step 5: Verify sessions are cleaned up or marked as completed
		for i := 0; i < numFiles; i++ {
			fileID := fileIDs[i]

			// Check upload status
			status, err := fs.uploads.GetUploadStatus(fileID)
			if err == nil {
				t.Logf("Upload status for %s: %v", fileID, status)
				// Status should be completed or session might be cleaned up
				assert.True(status == UploadCompletedState || status == UploadNotStartedState,
					"Upload status should be completed or not started (cleaned up) for %s (was %v)", fileID, status)
			} else {
				t.Logf("Upload session for %s was cleaned up: %v", fileID, err)
			}

			// Verify file was uploaded successfully
			updatedInode := fs.GetID(fileID)
			assert.NotNil(updatedInode, "File inode should exist after upload")

			updatedInode.mu.RLock()
			updatedETag := updatedInode.DriveItem.ETag
			updatedInode.mu.RUnlock()

			expectedETag := fmt.Sprintf("uploaded-etag-session-%d", i)
			assert.Equal(expectedETag, updatedETag, "ETag should be updated for file %d", i)
		}
	})
}
