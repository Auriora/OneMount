package fs

import (
	"encoding/json"
	"math"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	bolt "go.etcd.io/bbolt"
)

// TestUT_FS_Signal_Integration_01_SignalHandling tests actual signal handling
func TestUT_FS_Signal_Integration_01_SignalHandling(t *testing.T) {
	// Create a temporary database file
	tmpFile := filepath.Join(t.TempDir(), "test.db")
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

	// Verify initial state
	if manager.IsShuttingDown() {
		t.Error("Upload manager should not be shutting down initially")
	}

	// Send a SIGTERM signal to the signal channel
	t.Log("Sending SIGTERM signal to upload manager")
	go func() {
		time.Sleep(100 * time.Millisecond) // Give the signal handler time to start
		manager.signalChan <- syscall.SIGTERM
	}()

	// Wait for signal handling to complete
	time.Sleep(500 * time.Millisecond)

	// Verify that the upload manager is shutting down
	if !manager.IsShuttingDown() {
		t.Error("Upload manager should be shutting down after SIGTERM")
	}

	// Clean shutdown
	manager.Stop()

	t.Log("Signal handling integration test completed successfully")
}

// TestUT_FS_Signal_Integration_02_GracefulShutdown tests graceful shutdown behavior
func TestUT_FS_Signal_Integration_02_GracefulShutdown(t *testing.T) {
	// Create a temporary database file
	tmpFile := filepath.Join(t.TempDir(), "test.db")
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

	// Test that graceful shutdown methods work
	t.Log("Testing hasActiveUploads method")
	hasActive := manager.hasActiveUploads()
	if hasActive {
		t.Log("hasActiveUploads() returned true (unexpected but not an error)")
	} else {
		t.Log("hasActiveUploads() returned false (expected)")
	}

	t.Log("Testing persistActiveUploads method")
	manager.persistActiveUploads()

	t.Log("Testing logActiveUploads method")
	manager.logActiveUploads()

	// Test graceful shutdown timeout
	if manager.gracefulTimeout != 30*time.Second {
		t.Errorf("Expected graceful timeout to be 30s, got %v", manager.gracefulTimeout)
	}

	// Clean shutdown
	manager.Stop()

	t.Log("Graceful shutdown integration test completed successfully")
}

// TestUT_FS_Signal_Integration_03_ContextSupport tests context support in upload sessions
func TestUT_FS_Signal_Integration_03_ContextSupport(t *testing.T) {
	// Create a simple upload session for testing
	session := &UploadSession{
		ID:                  "test-context-integration",
		Name:                "test_context.txt",
		Size:                1024,
		LastSuccessfulChunk: -1,
		TotalChunks:         0,
		BytesUploaded:       0,
		LastProgressTime:    time.Now(),
		RecoveryAttempts:    0,
		CanResume:           false,
	}

	// Test that the session can be marshaled to JSON
	t.Log("Testing JSON marshaling for upload session")
	_, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal session: %v", err)
	}

	// Test progress tracking methods
	t.Log("Testing updateProgress method")
	session.updateProgress(2, 512)

	if session.LastSuccessfulChunk != 2 {
		t.Errorf("Expected LastSuccessfulChunk to be 2, got %d", session.LastSuccessfulChunk)
	}

	if session.BytesUploaded != 512 {
		t.Errorf("Expected BytesUploaded to be 512, got %d", session.BytesUploaded)
	}

	t.Log("Testing markAsResumable method")
	session.markAsResumable()

	if !session.CanResume {
		t.Error("Expected CanResume to be true after markAsResumable")
	}

	expectedChunks := int(math.Ceil(float64(session.Size) / float64(uploadChunkSize)))
	if session.TotalChunks != expectedChunks {
		t.Errorf("Expected TotalChunks to be %d, got %d", expectedChunks, session.TotalChunks)
	}

	t.Log("Context support integration test completed successfully")
}
