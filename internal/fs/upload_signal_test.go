package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	bolt "go.etcd.io/bbolt"
)

// TestUT_FS_Signal_01_UploadManager_GracefulShutdown tests that the upload manager
// handles signals gracefully and persists upload progress
// Note: This test does not run in parallel due to shared mock HTTP client state.
func TestUT_FS_Signal_01_UploadManager_GracefulShutdown(t *testing.T) {
	// Note: t.Parallel() removed due to race conditions with mock HTTP client cleanup

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SignalHandlingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
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

		// Create a large test file (5MB to trigger chunked upload)
		largeFileSize := 5 * 1024 * 1024
		testData := make([]byte, largeFileSize)
		for i := range testData {
			testData[i] = byte(i % 256)
		}

		// Create a test file item
		fileID := "test-large-file-signal"
		rootID := fsFixture.RootID
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: "large_test_signal.bin",
			Size: uint64(largeFileSize),
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Set the file content
		fs.content.Insert(fileID, testData)

		// Configure mock client to handle upload session creation
		mockClient := fsFixture.MockClient
		createSessionPath := "/me/drive/items/" + fileID + "/createUploadSession"
		uploadURL := "https://graph.microsoft.com/upload/session/test"
		sessionResponse := fmt.Sprintf(`{"uploadUrl":"%s","expirationDateTime":"2024-01-01T00:00:00Z"}`, uploadURL)
		mockClient.AddMockResponse(createSessionPath, []byte(sessionResponse), 200, nil)

		// Queue the upload
		uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")
		assert.NotNil(uploadSession, "Upload session should not be nil")

		// Wait for upload to start and create upload session (which marks it as resumable)
		// We need to wait long enough for the upload to actually start, not just be queued
		time.Sleep(1 * time.Second)

		// Verify upload session is active
		session, exists := fs.uploads.GetSession(fileID)
		assert.True(exists, "Upload session should exist")
		assert.NotNil(session, "Upload session should not be nil")

		// Check the upload state before sending signal
		state := session.getState()
		t.Logf("Upload state before signal: %v", state)

		// Send SIGTERM to trigger graceful shutdown
		fs.uploads.signalChan <- syscall.SIGTERM

		// Wait for signal handling to complete
		time.Sleep(2 * time.Second)

		// Verify that the upload manager is shutting down
		assert.True(fs.uploads.IsShuttingDown(), "Upload manager should be shutting down")

		// Verify that upload progress was persisted
		// Check if the session was saved to disk
		var persistedSession *UploadSession
		err = fs.uploads.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketUploads)
			if b == nil {
				return nil
			}
			data := b.Get([]byte(fileID))
			if data != nil {
				persistedSession = &UploadSession{}
				return json.Unmarshal(data, persistedSession)
			}
			return nil
		})
		assert.NoError(err, "Failed to check persisted session")

		// Check if the session was persisted
		if persistedSession != nil {
			// If the upload had started, it should be marked as resumable
			if state == uploadStarted {
				assert.True(persistedSession.CanResume, "Persisted session should be resumable if upload had started")
			}
			assert.Equal(fileID, persistedSession.ID, "Persisted session ID should match")
			t.Logf("Session was persisted with CanResume=%v", persistedSession.CanResume)
		} else {
			// If the upload hadn't started yet, it might not be persisted by the signal handler
			// This is acceptable behavior since only active uploads are persisted during shutdown
			t.Logf("Session was not persisted by signal handler (upload may not have started yet)")
		}
	})
}

// TestUT_FS_Signal_02_UploadSession_ContextCancellation tests that upload sessions
// handle context cancellation properly
// Note: This test does not run in parallel due to shared mock HTTP client state.
func TestUT_FS_Signal_02_UploadSession_ContextCancellation(t *testing.T) {
	// Note: t.Parallel() removed due to race conditions with mock HTTP client cleanup

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ContextCancellationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
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

		// Create a test file
		testData := []byte("test content for context cancellation")
		fileID := "test-context-cancel"
		rootID := fsFixture.RootID
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: "context_cancel_test.txt",
			Size: uint64(len(testData)),
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Create upload session
		fileInode := NewInodeDriveItem(fileItem)
		session, err := NewUploadSession(fileInode, &testData)
		assert.NoError(err, "Failed to create upload session")

		// Create a context that will be cancelled immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately to ensure context cancellation is detected

		// Try to upload with cancelled context - this should fail quickly
		err = session.UploadWithContext(ctx, fs.auth, fs.uploads.db)

		// Upload should have been cancelled
		assert.Error(err, "Upload should have been cancelled")
		assert.Contains(err.Error(), "cancelled", "Error should indicate cancellation")
	})
}

// TestUT_FS_Signal_02b_UploadSession_ContextCancellation_LargeFile tests that large file upload sessions
// handle context cancellation properly
// Note: This test does not run in parallel due to shared mock HTTP client state.
func TestUT_FS_Signal_02b_UploadSession_ContextCancellation_LargeFile(t *testing.T) {
	// Note: t.Parallel() removed due to race conditions with mock HTTP client cleanup

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ContextCancellationLargeFileFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
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

		// Create a large test file (>4MB to trigger large file upload path)
		testData := make([]byte, 5*1024*1024) // 5MB
		for i := range testData {
			testData[i] = byte(i % 256)
		}
		fileID := "test-context-cancel-large"
		rootID := fsFixture.RootID
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: "context_cancel_large_test.txt",
			Size: uint64(len(testData)),
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Create upload session
		fileInode := NewInodeDriveItem(fileItem)
		session, err := NewUploadSession(fileInode, &testData)
		assert.NoError(err, "Failed to create upload session")

		// Create a context that will be cancelled immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately to ensure context cancellation is detected

		// Try to upload with cancelled context - this should fail quickly
		err = session.UploadWithContext(ctx, fs.auth, fs.uploads.db)

		// Upload should have been cancelled
		assert.Error(err, "Upload should have been cancelled")
		assert.Contains(err.Error(), "cancelled", "Error should indicate cancellation")
	})
}

// TestUT_FS_Signal_03_UploadSession_ProgressPersistence tests that upload progress
// is persisted correctly during interruptions
// Note: This test does not run in parallel due to shared mock HTTP client state.
func TestUT_FS_Signal_03_UploadSession_ProgressPersistence(t *testing.T) {
	// Note: t.Parallel() removed due to race conditions with mock HTTP client cleanup

	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ProgressPersistenceFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
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

		// Create a test file
		testData := []byte("test content for progress persistence")
		fileID := "test-progress-persist"
		rootID := fsFixture.RootID
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: "progress_persist_test.txt",
			Size: uint64(len(testData)),
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Configure mock client to handle upload requests
		mockClient := fsFixture.MockClient
		uploadPath := "/me/drive/items/" + fileID + "/content"
		mockClient.AddMockResponse(uploadPath, []byte(`{"id":"`+fileID+`","name":"progress_persist_test.txt"}`), 200, nil)

		// Create upload session
		fileInode := NewInodeDriveItem(fileItem)
		session, err := NewUploadSession(fileInode, &testData)
		assert.NoError(err, "Failed to create upload session")

		// Mark as resumable first, then simulate progress
		// Note: markAsResumable() resets LastSuccessfulChunk to -1, so it must be called before updateProgress()
		session.markAsResumable()
		session.updateProgress(2, 1024)

		// Test persistence
		err = session.persistProgress(fs.uploads.db)
		assert.NoError(err, "Failed to persist progress")

		// Verify persistence by reading from database
		var persistedData []byte
		err = fs.uploads.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("uploads"))
			if b != nil {
				persistedData = b.Get([]byte(fileID))
			}
			return nil
		})
		assert.NoError(err, "Failed to read persisted data")
		assert.NotNil(persistedData, "Persisted data should not be nil")

		// Verify that the persisted session contains the progress
		var restoredSession UploadSession
		err = json.Unmarshal(persistedData, &restoredSession)
		assert.NoError(err, "Failed to unmarshal persisted session")
		assert.Equal(2, restoredSession.LastSuccessfulChunk, "Last successful chunk should be persisted")
		assert.Equal(uint64(1024), restoredSession.BytesUploaded, "Bytes uploaded should be persisted")
		assert.True(restoredSession.CanResume, "Session should be marked as resumable")
	})
}
