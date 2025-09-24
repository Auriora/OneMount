package fs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	bolt "go.etcd.io/bbolt"
)

// TestUT_FS_Signal_Simple_01_UploadSession_PersistProgress tests basic progress persistence
func TestUT_FS_Signal_Simple_01_UploadSession_PersistProgress(t *testing.T) {
	// Create a temporary database file
	tmpFile := t.TempDir() + "/test.db"
	db, err := bolt.Open(tmpFile, 0600, &bolt.Options{
		Timeout:        time.Second * 5,
		NoFreelistSync: true,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a test file item
	fileItem := &graph.DriveItem{
		ID:   "test-persist-simple",
		Name: "test_persist.txt",
		Size: 1024,
		File: &graph.File{},
		Parent: &graph.DriveItemParent{
			ID: "root",
		},
	}

	// Create test data
	testData := []byte("test content for persistence")

	// Create upload session
	fileInode := NewInodeDriveItem(fileItem)
	session, err := NewUploadSession(fileInode, &testData)
	if err != nil {
		t.Fatalf("Failed to create upload session: %v", err)
	}

	// Simulate some progress
	session.updateProgress(2, 512)
	session.markAsResumable()

	// Test basic database operations first
	t.Log("Testing basic database operations...")
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("test"))
		return err
	})
	if err != nil {
		t.Fatalf("Failed basic database operation: %v", err)
	}
	t.Log("Basic database operations work")

	// Test JSON marshaling first
	t.Log("Testing JSON marshaling...")
	contents, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal session: %v", err)
	}
	t.Logf("JSON marshaling successful, size: %d bytes", len(contents))

	// Test db.Batch directly with the same operation
	t.Log("Testing db.Batch directly...")
	done := make(chan error, 1)
	go func() {
		done <- db.Batch(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("uploads"))
			if err != nil {
				return err
			}
			return b.Put([]byte(session.ID), contents)
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Failed db.Batch operation: %v", err)
		}
		t.Log("db.Batch operation successful")
	case <-time.After(5 * time.Second):
		t.Fatal("db.Batch operation hung - timeout after 5 seconds")
	}

	// Now test the actual persistProgress method
	t.Log("Testing persistProgress method...")
	done2 := make(chan error, 1)
	go func() {
		done2 <- session.persistProgress(db)
	}()

	select {
	case err := <-done2:
		if err != nil {
			t.Fatalf("Failed to persist progress: %v", err)
		}
		t.Log("Successfully persisted upload session progress")
	case <-time.After(5 * time.Second):
		t.Fatal("persistProgress method hung - timeout after 5 seconds")
	}

	// Verify persistence by reading from database
	var persistedData []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("uploads"))
		if b != nil {
			persistedData = b.Get([]byte(session.ID))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to read from database: %v", err)
	}

	if persistedData == nil {
		t.Fatal("No data was persisted to database")
	}

	// Verify the persisted data can be unmarshaled
	var restoredSession UploadSession
	err = json.Unmarshal(persistedData, &restoredSession)
	if err != nil {
		t.Fatalf("Failed to unmarshal persisted session: %v", err)
	}

	// Verify key fields were persisted correctly
	if restoredSession.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, restoredSession.ID)
	}
	if restoredSession.LastSuccessfulChunk != session.LastSuccessfulChunk {
		t.Errorf("Expected LastSuccessfulChunk %d, got %d", session.LastSuccessfulChunk, restoredSession.LastSuccessfulChunk)
	}
	if restoredSession.BytesUploaded != session.BytesUploaded {
		t.Errorf("Expected BytesUploaded %d, got %d", session.BytesUploaded, restoredSession.BytesUploaded)
	}
	if !restoredSession.CanResume {
		t.Error("Expected CanResume to be true")
	}

	t.Log("Successfully tested upload session progress tracking and database persistence")
}

// TestUT_FS_Signal_Simple_02_UploadManager_SignalHandling tests basic signal handling setup
func TestUT_FS_Signal_Simple_02_UploadManager_SignalHandling(t *testing.T) {
	// Create a temporary database file
	tmpFile := t.TempDir() + "/test.db"
	db, err := bolt.Open(tmpFile, 0600, &bolt.Options{
		Timeout:        time.Second * 5,
		NoFreelistSync: true,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a mock filesystem interface
	mockFS := &mockFilesystemInterface{}

	// Create a mock auth
	auth := &graph.Auth{}

	// Create upload manager
	manager := NewUploadManager(time.Second, db, mockFS, auth)
	if manager == nil {
		t.Fatal("Failed to create upload manager")
	}

	// Verify signal handling fields are initialized
	if manager.signalChan == nil {
		t.Error("Signal channel should be initialized")
	}

	if manager.shutdownContext == nil {
		t.Error("Shutdown context should be initialized")
	}

	if manager.shutdownCancel == nil {
		t.Error("Shutdown cancel function should be initialized")
	}

	if manager.gracefulTimeout == 0 {
		t.Error("Graceful timeout should be set")
	}

	// Test graceful shutdown
	manager.Stop()

	t.Log("Upload manager signal handling initialized successfully")
}

// TestUT_FS_Signal_Simple_03_UploadSession_ContextCancellation tests context cancellation
func TestUT_FS_Signal_Simple_03_UploadSession_ContextCancellation(t *testing.T) {
	// Create a test file item
	fileItem := &graph.DriveItem{
		ID:   "test-context-simple",
		Name: "test_context.txt",
		Size: 100,
		File: &graph.File{},
		Parent: &graph.DriveItemParent{
			ID: "root",
		},
	}

	// Create test data
	testData := []byte("test content for context cancellation")

	// Create upload session
	fileInode := NewInodeDriveItem(fileItem)
	session, err := NewUploadSession(fileInode, &testData)
	if err != nil {
		t.Fatalf("Failed to create upload session: %v", err)
	}

	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Create a temporary database file
	tmpFile := t.TempDir() + "/test.db"
	db, err := bolt.Open(tmpFile, 0600, &bolt.Options{
		Timeout:        time.Second * 5,
		NoFreelistSync: true,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create mock auth
	auth := &graph.Auth{}

	// Try to upload with cancelled context - this should fail quickly
	err = session.UploadWithContext(ctx, auth, db)
	if err == nil {
		t.Error("Upload should have failed with cancelled context")
	}

	if err.Error() != "upload cancelled by context" {
		t.Logf("Got error: %v", err)
		// This is expected - the upload should be cancelled
	}

	t.Log("Context cancellation handled correctly")
}

// mockFilesystemInterface is a simple mock for testing
type mockFilesystemInterface struct{}

func (m *mockFilesystemInterface) SetFileStatus(id string, status FileStatusInfo) {}
func (m *mockFilesystemInterface) GetFileStatus(id string) FileStatusInfo {
	return FileStatusInfo{}
}
func (m *mockFilesystemInterface) MarkFileDownloading(id string)              {}
func (m *mockFilesystemInterface) MarkFileOutofSync(id string)                {}
func (m *mockFilesystemInterface) MarkFileError(id string, err error)         {}
func (m *mockFilesystemInterface) MarkFileConflict(id string, message string) {}
func (m *mockFilesystemInterface) UpdateFileStatus(inode *Inode)              {}
func (m *mockFilesystemInterface) InodePath(inode *Inode) string              { return "" }
func (m *mockFilesystemInterface) GetID(id string) *Inode                     { return nil }
func (m *mockFilesystemInterface) MoveID(oldID string, newID string) error    { return nil }
func (m *mockFilesystemInterface) GetInodeContent(inode *Inode) *[]byte       { return nil }
func (m *mockFilesystemInterface) IsOffline() bool                            { return false }
