package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestIT_FS_ETag_DiagnosticWithProgress tests the hanging behavior with detailed progress tracking
func TestIT_FS_ETag_DiagnosticWithProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create progress indicator
	progress := NewTestProgressIndicator()
	defer progress.Complete()

	progress.Step("Loading authentication tokens")
	authPath, err := testutil.GetAuthTokenPath()
	if err != nil {
		progress.Fail("Authentication not configured")
		t.Fatalf("Authentication not configured: %v", err)
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		progress.Fail("Cannot load auth tokens")
		t.Skipf("Skipping test: cannot load auth tokens from %s: %v", authPath, err)
	}
	progress.Substep("Auth tokens loaded successfully")

	progress.Step("Creating temporary directories")
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	mountPoint := filepath.Join(tempDir, "mount")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		progress.Fail("Failed to create mount point")
		t.Fatalf("Failed to create mount point: %v", err)
	}
	progress.Substep("Directories created successfully")

	progress.Step("Creating context with timeout")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	progress.Step("Creating filesystem instance")
	progress.Substep("Calling NewFilesystemWithContext...")
	filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		progress.Fail("Failed to create filesystem")
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	progress.Substep("Filesystem created successfully")

	// Setup cleanup with progress tracking
	defer func() {
		progress.Step("Cleaning up filesystem")
		progress.Substep("Stopping cache cleanup...")
		filesystem.StopCacheCleanup()
		progress.Substep("Stopping delta loop...")
		filesystem.StopDeltaLoop()
		progress.Substep("Stopping download manager...")
		filesystem.StopDownloadManager()
		progress.Substep("Stopping upload manager...")
		filesystem.StopUploadManager()
		progress.Substep("Stopping metadata request manager...")
		filesystem.StopMetadataRequestManager()
		progress.Substep("All cleanup completed")
	}()

	progress.Step("Creating FUSE server")
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-progress-test",
		FsName:        "onemount-progress-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(filesystem, mountPoint, mountOptions)
	if err != nil {
		progress.Fail("Failed to create FUSE server")
		t.Fatalf("Failed to create FUSE server: %v", err)
	}
	progress.Substep("FUSE server created")

	// Setup server cleanup with progress tracking
	defer func() {
		progress.Step("Unmounting FUSE server")
		progress.Substep("Calling server.Unmount()...")
		server.Unmount()
		progress.Substep("Unmount completed")
	}()

	progress.Step("Starting FUSE server")
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		progress.Substep("Server.Serve() started in goroutine")
		server.Serve()
		progress.Substep("Server.Serve() completed")
	}()

	progress.Step("Waiting for mount to complete")
	progress.Substep("Calling server.WaitMount()...")

	// Add timeout for mount operation
	mountDone := make(chan error, 1)
	go func() {
		mountDone <- server.WaitMount()
	}()

	select {
	case err := <-mountDone:
		if err != nil {
			progress.Fail("Mount failed")
			t.Fatalf("Failed to wait for mount: %v", err)
		}
		progress.Substep("Mount completed successfully")
	case <-time.After(15 * time.Second):
		progress.Fail("Mount timed out after 15 seconds")
		t.Fatal("Mount operation timed out")
	}

	progress.Step("Waiting for initial sync")
	for i := 0; i < 6; i++ {
		progress.Heartbeat(fmt.Sprintf("Initial sync wait %d/6 seconds", i+1))
		time.Sleep(1 * time.Second)
	}

	progress.Step("Testing basic file operations")
	testFileName := "progress_test.txt"
	testContent := []byte("test content for progress tracking")
	testFilePath := filepath.Join(mountPoint, testFileName)

	progress.Substep("Creating test file...")

	// Add timeout for file write
	writeDone := make(chan error, 1)
	go func() {
		writeDone <- os.WriteFile(testFilePath, testContent, 0644)
	}()

	select {
	case err := <-writeDone:
		if err != nil {
			progress.Fail("File write failed")
			t.Fatalf("Failed to create test file: %v", err)
		}
		progress.Substep("File created successfully")
	case <-time.After(10 * time.Second):
		progress.Fail("File write timed out after 10 seconds")
		t.Fatal("File write operation timed out - likely deadlock here")
	}

	progress.Step("Waiting for upload to complete")
	for i := 0; i < 5; i++ {
		progress.Heartbeat(fmt.Sprintf("Upload wait %d/5 seconds", i+1))
		time.Sleep(1 * time.Second)
	}

	progress.Step("Testing file read operation")
	progress.Substep("Reading test file...")

	// Add timeout for file read
	readDone := make(chan error, 1)
	var readContent []byte
	go func() {
		var err error
		readContent, err = os.ReadFile(testFilePath)
		readDone <- err
	}()

	select {
	case err := <-readDone:
		if err != nil {
			progress.Fail("File read failed")
			t.Fatalf("Failed to read test file: %v", err)
		}
		progress.Substep("File read successfully")
	case <-time.After(10 * time.Second):
		progress.Fail("File read timed out after 10 seconds")
		t.Fatal("File read operation timed out - likely deadlock here")
	}

	if string(readContent) != string(testContent) {
		progress.Fail("Content mismatch")
		t.Fatalf("Content mismatch: expected %q, got %q", testContent, readContent)
	}

	progress.Step("Testing the problematic GetChildrenID call")
	progress.Substep("This is where the original test hangs...")

	// Add timeout for the problematic call
	childrenDone := make(chan error, 1)
	go func() {
		progress.Heartbeat("Calling filesystem.GetChildrenID...")
		_, err := filesystem.GetChildrenID(filesystem.root, auth)
		childrenDone <- err
	}()

	select {
	case err := <-childrenDone:
		if err != nil {
			progress.Substep(fmt.Sprintf("GetChildrenID failed: %v", err))
		} else {
			progress.Substep("GetChildrenID completed successfully!")
		}
	case <-time.After(20 * time.Second):
		progress.Fail("GetChildrenID timed out after 20 seconds - THIS IS THE DEADLOCK!")
		t.Fatal("GetChildrenID operation timed out - confirmed deadlock location")
	}

	progress.Step("All operations completed successfully")
	t.Log("✓ Test completed without deadlock")
}
