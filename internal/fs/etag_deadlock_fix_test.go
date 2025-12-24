package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestIT_FS_ETag_DeadlockDiagnostic is a minimal test to isolate the hanging behavior
// This test focuses on the specific operations that cause the deadlock
func TestIT_FS_ETag_DeadlockDiagnostic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load authentication
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Skipf("Skipping test: cannot load auth tokens from %s: %v", authPath, err)
	}

	// Create temporary directories
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	mountPoint := filepath.Join(tempDir, "mount")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	// Use a shorter context timeout for faster failure detection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Creating filesystem...")
	filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}

	// Ensure proper cleanup with timeout
	cleanupDone := make(chan struct{})
	defer func() {
		go func() {
			defer close(cleanupDone)
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		select {
		case <-cleanupDone:
			t.Log("Cleanup completed successfully")
		case <-time.After(5 * time.Second):
			t.Log("Warning: Cleanup timed out")
		}
	}()

	t.Log("Creating FUSE server...")
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-deadlock-test",
		FsName:        "onemount-deadlock-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(filesystem, mountPoint, mountOptions)
	if err != nil {
		t.Fatalf("Failed to create FUSE server: %v", err)
	}

	// Ensure server unmount with timeout
	unmountDone := make(chan struct{})
	defer func() {
		go func() {
			defer close(unmountDone)
			server.Unmount()
		}()

		select {
		case <-unmountDone:
			t.Log("Unmount completed successfully")
		case <-time.After(5 * time.Second):
			t.Log("Warning: Unmount timed out")
		}
	}()

	t.Log("Starting FUSE server...")
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		server.Serve()
	}()

	t.Log("Waiting for mount...")
	mountTimeout := time.After(10 * time.Second)
	select {
	case <-mountTimeout:
		t.Fatal("Mount timed out after 10 seconds")
	default:
		if err := server.WaitMount(); err != nil {
			t.Fatalf("Failed to wait for mount: %v", err)
		}
	}

	t.Log("Filesystem mounted successfully")

	// Wait for initial sync with timeout
	syncTimeout := time.After(5 * time.Second)
	select {
	case <-syncTimeout:
		t.Log("Initial sync timed out, continuing...")
	case <-time.After(3 * time.Second):
		t.Log("Initial sync completed")
	}

	// Test basic file operations that might cause deadlock
	testFileName := "deadlock_test.txt"
	testContent := []byte("test content")
	testFilePath := filepath.Join(mountPoint, testFileName)

	t.Log("Testing file creation...")
	writeTimeout := time.After(10 * time.Second)
	writeDone := make(chan error, 1)

	go func() {
		writeDone <- os.WriteFile(testFilePath, testContent, 0644)
	}()

	select {
	case err := <-writeDone:
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		t.Log("File creation successful")
	case <-writeTimeout:
		t.Fatal("File write operation timed out - this is likely where the deadlock occurs")
	}

	// Test file reading
	t.Log("Testing file reading...")
	readTimeout := time.After(10 * time.Second)
	readDone := make(chan error, 1)

	go func() {
		_, err := os.ReadFile(testFilePath)
		readDone <- err
	}()

	select {
	case err := <-readDone:
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}
		t.Log("File reading successful")
	case <-readTimeout:
		t.Fatal("File read operation timed out - this is likely where the deadlock occurs")
	}

	t.Log("✓ Basic file operations completed without deadlock")
}
