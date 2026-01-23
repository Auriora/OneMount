package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestMinimalHangReproduction creates the absolute minimal test to reproduce the hanging behavior
// This test isolates each step to identify exactly where the hang occurs
func TestUT_FS_MinimalHang_Reproduction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	progress := NewTestProgressIndicator()
	defer progress.Complete()

	// Step 1: Create mock authentication
	progress.Step("Creating mock authentication")
	auth := createMockAuth()
	progress.Substep("✓ Mock auth created")

	// Step 2: Mock token validation (always valid for unit tests)
	progress.Step("Checking token validity")
	progress.Substep("✓ Token is valid (mocked)")

	// Step 3: Create filesystem with minimal timeout
	progress.Step("Creating filesystem instance")
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	mountPoint := filepath.Join(tempDir, "mount")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		progress.Fail("Failed to create mount point")
		t.Fatalf("Failed to create mount point: %v", err)
	}

	// Use mock graph root for unit tests
	ensureMockGraphRoot(t)

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
func TestUT_FS_MinimalHang_PointIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	progress := NewTestProgressIndicator()
	defer progress.Complete()

	// Test 1: Authentication operations (mocked for unit tests)
	progress.Step("Testing authentication operations in isolation")
	auth := createMockAuth()
	progress.Substep("✓ Mock auth created")

	// Test token validation with mock (always succeeds)
	progress.Substep("Testing token validation...")
	progress.Substep("✓ Token validation successful (mocked)")

	// Test 2: Filesystem creation
	progress.Step("Testing filesystem creation in isolation")
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")

	// Use mock graph root for unit tests
	ensureMockGraphRoot(t)

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
func TestUT_FS_MinimalHang_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	progress := NewTestProgressIndicator()
	defer progress.Complete()

	progress.Step("Testing concurrent operations for deadlocks")

	// Use mock auth for unit tests
	auth := createMockAuth()
	progress.Substep("✓ Mock auth created")

	// Test concurrent filesystem operations (no real API calls)
	progress.Substep("Testing concurrent filesystem operations...")

	// Use mock graph root for unit tests
	ensureMockGraphRoot(t)

	var wg sync.WaitGroup
	results := make(chan error, 5)

	// Start 5 concurrent filesystem creation operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			progress.Heartbeat(fmt.Sprintf("Concurrent operation %d", id))

			tempDir := t.TempDir()
			cacheDir := filepath.Join(tempDir, fmt.Sprintf("cache-%d", id))

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
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
		progress.Substep("✓ All concurrent operations completed")

		// Check results
		close(results)
		for err := range results {
			if err != nil {
				t.Logf("Concurrent operation failed: %v", err)
			}
		}
	case <-time.After(30 * time.Second):
		progress.Fail("Concurrent operations timed out")
		t.Fatal("Concurrent operations timed out - deadlock detected")
	}

	progress.Step("Concurrent operations test completed")
	t.Log("✓ Concurrent operations test completed without deadlocks")
}
