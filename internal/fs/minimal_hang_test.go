package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestMinimalHangReproduction creates the absolute minimal test to reproduce the hanging behavior
// This test isolates each step to identify exactly where the hang occurs
func TestMinimalHangReproduction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	progress := NewTestProgressIndicator()
	defer progress.Complete()

	// Step 1: Load authentication with timeout
	progress.Step("Loading authentication tokens")
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		progress.Fail("Cannot load auth tokens")
		t.Skipf("Skipping test: cannot load auth tokens from %s: %v", authPath, err)
	}
	progress.Substep("✓ Auth tokens loaded")

	// Step 2: Check token expiration
	progress.Step("Checking token validity")
	safeAuth := NewSafeAuthWrapper(auth, DefaultAuthTimeoutConfig())
	if safeAuth.IsTokenExpired() {
		progress.Substep("⚠️ Token is expired - this may cause hangs")
		t.Logf("Token is expired, this is likely the cause of hangs")
	} else {
		progress.Substep("✓ Token is valid")
	}

	// Step 3: Create filesystem with minimal timeout
	progress.Step("Creating filesystem instance")
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	mountPoint := filepath.Join(tempDir, "mount")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		progress.Fail("Failed to create mount point")
		t.Fatalf("Failed to create mount point: %v", err)
	}

	// Use very short timeout to fail fast
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	progress.Substep("Creating filesystem with 15s timeout...")
	filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		progress.Fail("Failed to create filesystem")
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	progress.Substep("✓ Filesystem created")

	// Ensure cleanup
	defer func() {
		progress.Substep("Cleaning up filesystem...")
		filesystem.StopCacheCleanup()
		filesystem.StopDeltaLoop()
		filesystem.StopDownloadManager()
		filesystem.StopUploadManager()
		filesystem.StopMetadataRequestManager()
	}()

	// Step 4: Create FUSE server
	progress.Step("Creating FUSE server")
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-minimal-hang-test",
		FsName:        "onemount-minimal-hang-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(filesystem, mountPoint, mountOptions)
	if err != nil {
		progress.Fail("Failed to create FUSE server")
		t.Fatalf("Failed to create FUSE server: %v", err)
	}
	progress.Substep("✓ FUSE server created")

	defer server.Unmount()

	// Step 5: Start server in background
	progress.Step("Starting FUSE server")
	go server.Serve()
	progress.Substep("✓ Server started in background")

	// Step 6: Wait for mount with timeout and detailed monitoring
	progress.Step("Waiting for filesystem mount")
	mountDone := make(chan error, 1)
	go func() {
		progress.Substep("Calling server.WaitMount()...")
		err := server.WaitMount()
		progress.Substep(fmt.Sprintf("server.WaitMount() returned: %v", err))
		mountDone <- err
	}()

	select {
	case err := <-mountDone:
		if err != nil {
			progress.Fail("Mount failed")
			t.Fatalf("Failed to wait for mount: %v", err)
		}
		progress.Substep("✓ Filesystem mounted successfully")
	case <-time.After(10 * time.Second):
		progress.Fail("Mount timed out")
		t.Fatal("Mount operation timed out - this is where the hang likely occurs")
	}

	// Step 7: Test basic directory listing (this often hangs)
	progress.Step("Testing basic directory operations")
	progress.Substep("Attempting to list mount point directory...")

	listDone := make(chan error, 1)
	go func() {
		_, err := os.ReadDir(mountPoint)
		listDone <- err
	}()

	select {
	case err := <-listDone:
		if err != nil {
			progress.Substep("❌ Directory listing failed")
			t.Logf("Directory listing failed: %v", err)
		} else {
			progress.Substep("✓ Directory listing successful")
		}
	case <-time.After(10 * time.Second):
		progress.Fail("Directory listing timed out")
		t.Fatal("Directory listing timed out - this is likely where the hang occurs")
	}

	// Step 8: Test file creation (this is where most hangs occur)
	progress.Step("Testing file creation")
	testFileName := "minimal_hang_test.txt"
	testContent := []byte("minimal test content")
	testFilePath := filepath.Join(mountPoint, testFileName)

	progress.Substep("Attempting to create test file...")
	createDone := make(chan error, 1)
	go func() {
		err := os.WriteFile(testFilePath, testContent, 0644)
		createDone <- err
	}()

	select {
	case err := <-createDone:
		if err != nil {
			progress.Substep("❌ File creation failed")
			t.Logf("File creation failed: %v", err)
		} else {
			progress.Substep("✓ File creation successful")
		}
	case <-time.After(15 * time.Second):
		progress.Fail("File creation timed out")
		t.Fatal("File creation timed out - THIS IS THE HANG POINT")
	}

	progress.Step("All operations completed successfully")
	t.Log("✓ Minimal hang reproduction test completed without hanging")
}

// TestHangPointIsolation tests each operation individually to isolate the exact hang point
func TestHangPointIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	progress := NewTestProgressIndicator()
	defer progress.Complete()

	// Test 1: Authentication operations
	progress.Step("Testing authentication operations in isolation")
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		progress.Fail("Cannot load auth tokens")
		t.Skipf("Skipping test: cannot load auth tokens from %s: %v", authPath, err)
	}

	// Test token validation with timeout
	progress.Substep("Testing token validation...")
	safeAuth := NewSafeAuthWrapper(auth, DefaultAuthTimeoutConfig())

	validationDone := make(chan error, 1)
	go func() {
		validationDone <- safeAuth.ValidateConnection()
	}()

	select {
	case err := <-validationDone:
		if err != nil {
			progress.Substep("⚠️ Token validation failed - this will cause hangs")
			t.Logf("Token validation failed: %v", err)
		} else {
			progress.Substep("✓ Token validation successful")
		}
	case <-time.After(15 * time.Second):
		progress.Substep("❌ Token validation timed out")
		t.Log("Token validation timed out - authentication is the hang point")
	}

	// Test 2: Filesystem creation
	progress.Step("Testing filesystem creation in isolation")
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createDone := make(chan error, 1)
	var filesystem FilesystemInterface
	go func() {
		fs, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		filesystem = fs
		createDone <- err
	}()

	select {
	case err := <-createDone:
		if err != nil {
			progress.Substep("❌ Filesystem creation failed")
			t.Logf("Filesystem creation failed: %v", err)
			return
		} else {
			progress.Substep("✓ Filesystem creation successful")
		}
	case <-time.After(12 * time.Second):
		progress.Substep("❌ Filesystem creation timed out")
		t.Log("Filesystem creation timed out - this is a hang point")
		return
	}

	// Clean up filesystem if created
	if fs, ok := filesystem.(*Filesystem); ok {
		defer func() {
			fs.StopCacheCleanup()
			fs.StopDeltaLoop()
			fs.StopDownloadManager()
			fs.StopUploadManager()
			fs.StopMetadataRequestManager()
		}()
	}

	progress.Step("Isolation test completed")
	t.Log("✓ Hang point isolation test completed")
}

// TestConcurrentOperations tests if concurrent operations cause deadlocks
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	progress := NewTestProgressIndicator()
	defer progress.Complete()

	progress.Step("Testing concurrent operations for deadlocks")

	// Load auth
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		progress.Fail("Cannot load auth tokens")
		t.Skipf("Skipping test: cannot load auth tokens from %s: %v", authPath, err)
	}

	// Test concurrent auth operations
	progress.Substep("Testing concurrent authentication operations...")
	safeAuth := NewSafeAuthWrapper(auth, DefaultAuthTimeoutConfig())

	var wg sync.WaitGroup
	results := make(chan error, 5)

	// Start 5 concurrent auth operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			progress.Heartbeat(fmt.Sprintf("Concurrent auth operation %d", id))
			err := safeAuth.ValidateConnection()
			results <- err
		}(i)
	}

	// Wait for all operations with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		progress.Substep("✓ All concurrent auth operations completed")

		// Check results
		close(results)
		for err := range results {
			if err != nil {
				t.Logf("Concurrent auth operation failed: %v", err)
			}
		}
	case <-time.After(30 * time.Second):
		progress.Fail("Concurrent operations timed out")
		t.Fatal("Concurrent auth operations timed out - deadlock detected")
	}

	progress.Step("Concurrent operations test completed")
	t.Log("✓ Concurrent operations test completed without deadlocks")
}
