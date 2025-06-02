package fs

import (
	"context"
	"encoding/json"
	"syscall"
	"testing"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/testutil/framework"
	"github.com/auriora/onemount/pkg/testutil/helpers"
	bolt "go.etcd.io/bbolt"
)

// TestUT_FS_Signal_01_UploadManager_GracefulShutdown tests that the upload manager
// handles signals gracefully and persists upload progress
func TestUT_FS_Signal_01_UploadManager_GracefulShutdown(t *testing.T) {
	t.Parallel()

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

		// Queue the upload
		uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")
		assert.NotNil(uploadSession, "Upload session should not be nil")

		// Wait a moment for upload to start
		time.Sleep(100 * time.Millisecond)

		// Verify upload session is active
		session, exists := fs.uploads.GetSession(fileID)
		assert.True(exists, "Upload session should exist")
		assert.NotNil(session, "Upload session should not be nil")

		// Send SIGTERM to trigger graceful shutdown
		fs.uploads.signalChan <- syscall.SIGTERM

		// Wait for signal handling to complete
		time.Sleep(2 * time.Second)

		// Verify that the upload manager is shutting down
		assert.True(fs.uploads.isShuttingDown, "Upload manager should be shutting down")

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

		if persistedSession != nil {
			assert.True(persistedSession.CanResume, "Persisted session should be resumable")
			assert.Equal(fileID, persistedSession.ID, "Persisted session ID should match")
		}
	})
}

// TestUT_FS_Signal_02_UploadSession_ContextCancellation tests that upload sessions
// handle context cancellation properly
func TestUT_FS_Signal_02_UploadSession_ContextCancellation(t *testing.T) {
	t.Parallel()

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

		// Create a context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())

		// Start upload in a goroutine
		uploadDone := make(chan error, 1)
		go func() {
			uploadDone <- session.UploadWithContext(ctx, fs.auth, fs.uploads.db)
		}()

		// Cancel the context after a short delay
		time.Sleep(50 * time.Millisecond)
		cancel()

		// Wait for upload to complete or timeout
		select {
		case err := <-uploadDone:
			// Upload should have been cancelled
			assert.Error(err, "Upload should have been cancelled")
			assert.Contains(err.Error(), "cancelled", "Error should indicate cancellation")

		case <-time.After(5 * time.Second):
			t.Fatal("Upload did not complete within timeout")
		}
	})
}

// TestUT_FS_Signal_03_UploadSession_ProgressPersistence tests that upload progress
// is persisted correctly during interruptions
func TestUT_FS_Signal_03_UploadSession_ProgressPersistence(t *testing.T) {
	t.Parallel()

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

		// Create upload session
		fileInode := NewInodeDriveItem(fileItem)
		session, err := NewUploadSession(fileInode, &testData)
		assert.NoError(err, "Failed to create upload session")

		// Simulate some progress
		session.updateProgress(2, 1024)
		session.markAsResumable()

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
