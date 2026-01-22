package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestDeadlockRootCauseAnalysis systematically tests each component that could cause the hang
func TestDeadlockRootCauseAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	progress := NewTestProgressIndicator()
	defer progress.Complete()

	progress.Step("Phase 1: Testing authentication")
	authPath, err := testutil.GetAuthTokenPath()
	if err != nil {
		progress.Fail("Authentication not configured")
		t.Fatalf("Authentication not configured: %v", err)
	}

	// Test 1: Can we load auth tokens?
	progress.Substep("Loading auth tokens...")
	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		progress.Fail("Cannot load auth tokens")
		t.Fatalf("Cannot load auth tokens: %v", err)
	}

	// Test 2: Are tokens expired?
	progress.Substep("Checking token expiration...")
	expiresAt := time.Unix(auth.ExpiresAt, 0)
	if time.Now().After(expiresAt) {
		progress.Substep("⚠️ Tokens are EXPIRED - this could cause hangs!")
		t.Logf("Token expired at: %v, current time: %v", expiresAt, time.Now())
	} else {
		progress.Substep("✓ Tokens are valid")
	}

	// Test 3: Can we make a simple API call?
	progress.Substep("Testing basic API connectivity...")
	apiTestDone := make(chan error, 1)
	go func() {
		_, err := graph.Get("/me", auth)
		apiTestDone <- err
	}()

	select {
	case err := <-apiTestDone:
		if err != nil {
			progress.Substep(fmt.Sprintf("⚠️ API call failed: %v - this could cause hangs!", err))
		} else {
			progress.Substep("✓ API connectivity working")
		}
	case <-time.After(15 * time.Second):
		progress.Substep("❌ API call timed out - THIS IS LIKELY THE ROOT CAUSE!")
		t.Log("API calls are timing out - this will cause filesystem operations to hang")
	}

	progress.Step("Phase 2: Testing filesystem creation")
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")

	progress.Substep("Creating filesystem instance...")
	filesystem, err := NewFilesystem(auth, cacheDir, 30)
	if err != nil {
		progress.Fail("Failed to create filesystem")
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer func() {
		progress.Substep("Cleaning up filesystem...")
		filesystem.StopCacheCleanup()
		filesystem.StopDeltaLoop()
		filesystem.StopDownloadManager()
		filesystem.StopUploadManager()
		filesystem.StopMetadataRequestManager()
	}()

	progress.Step("Phase 3: Testing cache operations")
	progress.Substep("Testing cache file creation...")

	testID := "test-cache-file-123"
	cacheTestDone := make(chan error, 1)
	go func() {
		fd, err := filesystem.content.Open(testID)
		if err != nil {
			cacheTestDone <- err
			return
		}

		_, err = fd.WriteAt([]byte("test content"), 0)
		if err != nil {
			cacheTestDone <- err
			return
		}

		err = filesystem.content.Close(testID)
		cacheTestDone <- err
	}()

	select {
	case err := <-cacheTestDone:
		if err != nil {
			progress.Substep(fmt.Sprintf("❌ Cache operations failed: %v", err))
		} else {
			progress.Substep("✓ Cache operations working")
		}
	case <-time.After(10 * time.Second):
		progress.Fail("Cache operations timed out")
		t.Fatal("Cache operations are hanging")
	}

	progress.Step("Phase 4: Testing FUSE mount")
	mountPoint := filepath.Join(tempDir, "mount")
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		progress.Fail("Failed to create mount point")
		t.Fatalf("Failed to create mount point: %v", err)
	}

	progress.Substep("Creating FUSE server...")
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-deadlock-analysis",
		FsName:        "onemount-deadlock-analysis",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(filesystem, mountPoint, mountOptions)
	if err != nil {
		progress.Fail("Failed to create FUSE server")
		t.Fatalf("Failed to create FUSE server: %v", err)
	}
	defer server.Unmount()

	progress.Substep("Starting FUSE server...")
	go server.Serve()

	progress.Substep("Waiting for mount...")
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
		progress.Substep("✓ Mount successful")
	case <-time.After(15 * time.Second):
		progress.Fail("Mount timed out")
		t.Fatal("FUSE mount is hanging")
	}

	progress.Step("Phase 5: Testing file operations")
	testFileName := "deadlock_analysis.txt"
	testContent := []byte("test content")
	testFilePath := filepath.Join(mountPoint, testFileName)

	progress.Substep("Attempting file creation...")

	writeDone := make(chan error, 1)
	go func() {
		t.Logf("About to call os.WriteFile for %s", testFilePath)
		err := os.WriteFile(testFilePath, testContent, 0644)
		t.Logf("os.WriteFile completed with error: %v", err)
		writeDone <- err
	}()

	writeTimeout := time.After(20 * time.Second)
	heartbeatTicker := time.NewTicker(2 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case err := <-writeDone:
			if err != nil {
				progress.Substep(fmt.Sprintf("❌ File write failed: %v", err))
				t.Fatalf("File write failed: %v", err)
			} else {
				progress.Substep("✓ File write successful!")
				goto writeComplete
			}
		case <-heartbeatTicker.C:
			progress.Heartbeat("File write still in progress...")
		case <-writeTimeout:
			progress.Fail("File write timed out - ROOT CAUSE IDENTIFIED!")
			t.Fatal("File write operation is hanging")
		}
	}

writeComplete:
	progress.Step("Phase 6: Testing file read")
	progress.Substep("Attempting file read...")

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
			progress.Substep(fmt.Sprintf("❌ File read failed: %v", err))
		} else {
			progress.Substep("✓ File read successful!")
			if string(readContent) == string(testContent) {
				progress.Substep("✓ File content matches")
			}
		}
	case <-time.After(10 * time.Second):
		progress.Fail("File read timed out")
		t.Fatal("File read operation is hanging")
	}

	progress.Step("All tests completed")
	t.Log("✓ No deadlock detected")
}
