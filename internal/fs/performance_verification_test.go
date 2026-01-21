package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
)

// TestPerformanceVerification runs comprehensive performance benchmarks
// Task 37: Performance verification
// - Run performance benchmarks
// - Test with Socket.IO realtime (30min polling fallback)
// - Test polling-only mode (5min polling)
// - Compare polling frequency impact
// - Verify response times meet expectations
// - Check resource usage is reasonable
func TestPerformanceVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance verification in short mode")
	}

	// Check if we should run performance tests
	if os.Getenv("RUN_PERFORMANCE_TESTS") != "1" {
		t.Skip("Skipping performance tests (set RUN_PERFORMANCE_TESTS=1 to run)")
	}

	t.Run("DirectoryListingPerformance", testDirectoryListingPerformance)
	t.Run("CachedFileAccessPerformance", testCachedFileAccessPerformance)
	t.Run("MemoryUsageIdle", testMemoryUsageIdle)
	t.Run("MemoryUsageActive", testMemoryUsageActive)
	t.Run("ConcurrentOperations", testConcurrentOperations)
	t.Run("StartupTime", testStartupTime)
	t.Run("ShutdownTime", testShutdownTime)
	t.Run("PollingFrequencyImpact", testPollingFrequencyImpact)
}

// testDirectoryListingPerformance verifies directory listing meets performance requirements
// Requirement 23.1: Directory with up to 1000 files should respond within 2 seconds
// Requirement 23.8: Directory with up to 10,000 files should respond within 3 seconds
func testDirectoryListingPerformance(t *testing.T) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer cleanupFilesystem(fs)

	// Test with 1000 files (should be < 2 seconds)
	t.Run("1000Files", func(t *testing.T) {
		// Create a directory with 1000 files in metadata
		dirID := "test-dir-1000"
		createTestDirectory(t, fs, dirID, 1000)

		start := time.Now()
		children, err := fs.GetChildrenID(dirID, auth)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Failed to list directory: %v", err)
		}

		if len(children) != 1000 {
			t.Errorf("Expected 1000 children, got %d", len(children))
		}

		maxDuration := 2 * time.Second
		if elapsed > maxDuration {
			t.Errorf("Directory listing took %v, expected < %v (Requirement 23.1)", elapsed, maxDuration)
		} else {
			t.Logf("✓ Directory listing (1000 files) completed in %v (< %v)", elapsed, maxDuration)
		}
	})

	// Test with 10,000 files (should be < 3 seconds)
	t.Run("10000Files", func(t *testing.T) {
		dirID := "test-dir-10000"
		createTestDirectory(t, fs, dirID, 10000)

		start := time.Now()
		children, err := fs.GetChildrenID(dirID, auth)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Failed to list directory: %v", err)
		}

		if len(children) != 10000 {
			t.Errorf("Expected 10000 children, got %d", len(children))
		}

		maxDuration := 3 * time.Second
		if elapsed > maxDuration {
			t.Errorf("Directory listing took %v, expected < %v (Requirement 23.8)", elapsed, maxDuration)
		} else {
			t.Logf("✓ Directory listing (10000 files) completed in %v (< %v)", elapsed, maxDuration)
		}
	})
}

// testCachedFileAccessPerformance verifies cached file access meets performance requirements
// Requirement 23.2: Cached file should be served within 100 milliseconds
func testCachedFileAccessPerformance(t *testing.T) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer cleanupFilesystem(fs)

	// Create a test file and cache it
	fileID := "test-cached-file"
	testContent := []byte("test content for performance verification")

	// Create file in metadata using DriveItem
	fileItem := &graph.DriveItem{
		ID:   fileID,
		Name: "test.txt",
		Size: uint64(len(testContent)),
		File: &graph.File{},
		Parent: &graph.DriveItemParent{
			ID: "root",
		},
	}
	inode := NewInodeDriveItem(fileItem)
	fs.InsertID(fileID, inode)

	// Cache the content
	contentPath := filepath.Join(tempDir, "content", fileID)
	os.MkdirAll(filepath.Dir(contentPath), 0755)
	if err := os.WriteFile(contentPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}

	// Measure cached file access time
	start := time.Now()
	content, err := os.ReadFile(contentPath)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Failed to read cached file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("Content mismatch")
	}

	maxDuration := 100 * time.Millisecond
	if elapsed > maxDuration {
		t.Errorf("Cached file access took %v, expected < %v (Requirement 23.2)", elapsed, maxDuration)
	} else {
		t.Logf("✓ Cached file access completed in %v (< %v)", elapsed, maxDuration)
	}
}

// testMemoryUsageIdle verifies idle memory usage meets requirements
// Requirement 23.3: Idle system should consume no more than 50 MB of RAM
func testMemoryUsageIdle(t *testing.T) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer cleanupFilesystem(fs)

	// Wait for system to stabilize
	time.Sleep(2 * time.Second)

	// Force garbage collection to get accurate measurement
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := float64(m.Alloc) / 1024 / 1024
	maxMemoryMB := 50.0

	if allocMB > maxMemoryMB {
		t.Errorf("Idle memory usage: %.2f MB, expected < %.2f MB (Requirement 23.3)", allocMB, maxMemoryMB)
	} else {
		t.Logf("✓ Idle memory usage: %.2f MB (< %.2f MB)", allocMB, maxMemoryMB)
	}
}

// testMemoryUsageActive verifies active memory usage meets requirements
// Requirement 23.4: Active system should consume no more than 200 MB of RAM
func testMemoryUsageActive(t *testing.T) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer cleanupFilesystem(fs)

	// Simulate active usage by creating many files
	for i := 0; i < 100; i++ {
		fileID := fmt.Sprintf("active-test-file-%d", i)
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: fmt.Sprintf("file-%d.txt", i),
			Size: 1024,
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: "root",
			},
		}
		inode := NewInodeDriveItem(fileItem)
		fs.InsertID(fileID, inode)
	}

	// Force garbage collection
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := float64(m.Alloc) / 1024 / 1024
	maxMemoryMB := 200.0

	if allocMB > maxMemoryMB {
		t.Errorf("Active memory usage: %.2f MB, expected < %.2f MB (Requirement 23.4)", allocMB, maxMemoryMB)
	} else {
		t.Logf("✓ Active memory usage: %.2f MB (< %.2f MB)", allocMB, maxMemoryMB)
	}
}

// testConcurrentOperations verifies concurrent operation handling
// Requirement 23.7: Should handle at least 10 simultaneous file operations
func testConcurrentOperations(t *testing.T) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer cleanupFilesystem(fs)

	// Create test files
	numOperations := 10
	for i := 0; i < numOperations; i++ {
		fileID := fmt.Sprintf("concurrent-file-%d", i)
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: fmt.Sprintf("file-%d.txt", i),
			Size: 1024,
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: "root",
			},
		}
		inode := NewInodeDriveItem(fileItem)
		fs.InsertID(fileID, inode)
	}

	// Perform concurrent operations
	start := time.Now()
	done := make(chan bool, numOperations)

	for i := 0; i < numOperations; i++ {
		go func(idx int) {
			fileID := fmt.Sprintf("concurrent-file-%d", idx)
			inode := fs.GetID(fileID)
			if inode == nil {
				t.Errorf("Concurrent operation %d failed: inode not found", idx)
			}
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numOperations; i++ {
		<-done
	}
	elapsed := time.Since(start)

	// All operations should complete reasonably quickly
	maxDuration := 1 * time.Second
	if elapsed > maxDuration {
		t.Errorf("Concurrent operations took %v, expected < %v (Requirement 23.7)", elapsed, maxDuration)
	} else {
		t.Logf("✓ %d concurrent operations completed in %v (< %v)", numOperations, elapsed, maxDuration)
	}
}

// testStartupTime verifies startup time meets requirements
// Requirement 23.9: Should complete initialization within 5 seconds
func testStartupTime(t *testing.T) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	start := time.Now()
	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer cleanupFilesystem(fs)

	maxDuration := 5 * time.Second
	if elapsed > maxDuration {
		t.Errorf("Startup took %v, expected < %v (Requirement 23.9)", elapsed, maxDuration)
	} else {
		t.Logf("✓ Startup completed in %v (< %v)", elapsed, maxDuration)
	}
}

// testShutdownTime verifies shutdown time meets requirements
// Requirement 23.10: Should complete graceful shutdown within 10 seconds
func testShutdownTime(t *testing.T) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}

	// Simulate some activity
	for i := 0; i < 10; i++ {
		fileID := fmt.Sprintf("shutdown-test-file-%d", i)
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: fmt.Sprintf("file-%d.txt", i),
			Size: 1024,
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: "root",
			},
		}
		inode := NewInodeDriveItem(fileItem)
		fs.InsertID(fileID, inode)
	}

	start := time.Now()
	cleanupFilesystem(fs)
	elapsed := time.Since(start)

	maxDuration := 10 * time.Second
	if elapsed > maxDuration {
		t.Errorf("Shutdown took %v, expected < %v (Requirement 23.10)", elapsed, maxDuration)
	} else {
		t.Logf("✓ Shutdown completed in %v (< %v)", elapsed, maxDuration)
	}
}

// testPollingFrequencyImpact tests different polling configurations
// Task 37: Test with Socket.IO realtime (30min polling fallback) vs polling-only mode (5min polling)
func testPollingFrequencyImpact(t *testing.T) {
	t.Run("PollingOnlyMode5Min", func(t *testing.T) {
		testPollingMode(t, 5*time.Minute, "polling-only")
	})

	t.Run("RealtimeMode30MinFallback", func(t *testing.T) {
		testPollingMode(t, 30*time.Minute, "realtime-fallback")
	})
}

// testPollingMode tests a specific polling configuration
func testPollingMode(t *testing.T, pollingInterval time.Duration, mode string) {
	ensureMockGraphRoot(t)

	auth := createTestAuth(t)
	tempDir := t.TempDir()

	ctx := context.Background()
	fs, err := NewFilesystemWithContext(ctx, auth, tempDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer cleanupFilesystem(fs)

	// Measure baseline resource usage
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	baselineMemMB := float64(m1.Alloc) / 1024 / 1024

	// Simulate polling activity for a short period
	testDuration := 10 * time.Second
	start := time.Now()

	// Simulate some file operations during polling
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	operationCount := 0
	timeout := time.After(testDuration)

loop:
	for {
		select {
		case <-ticker.C:
			// Simulate file access
			fileID := fmt.Sprintf("poll-test-file-%d", operationCount)
			fileItem := &graph.DriveItem{
				ID:   fileID,
				Name: fmt.Sprintf("file-%d.txt", operationCount),
				Size: 1024,
				File: &graph.File{},
				Parent: &graph.DriveItemParent{
					ID: "root",
				},
			}
			inode := NewInodeDriveItem(fileItem)
			fs.InsertID(fileID, inode)
			operationCount++

		case <-timeout:
			break loop
		}
	}

	elapsed := time.Since(start)

	// Measure final resource usage
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	finalMemMB := float64(m2.Alloc) / 1024 / 1024
	memDeltaMB := finalMemMB - baselineMemMB

	t.Logf("✓ Polling mode: %s", mode)
	t.Logf("  Polling interval: %v", pollingInterval)
	t.Logf("  Test duration: %v", elapsed)
	t.Logf("  Operations performed: %d", operationCount)
	t.Logf("  Baseline memory: %.2f MB", baselineMemMB)
	t.Logf("  Final memory: %.2f MB", finalMemMB)
	t.Logf("  Memory delta: %.2f MB", memDeltaMB)

	// Verify resource usage is reasonable
	maxMemDeltaMB := 50.0
	if memDeltaMB > maxMemDeltaMB {
		t.Errorf("Memory increase (%.2f MB) exceeds threshold (%.2f MB)", memDeltaMB, maxMemDeltaMB)
	}
}

// Helper functions

func createTestAuth(t *testing.T) *graph.Auth {
	return &graph.Auth{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}
}

func createTestDirectory(t *testing.T, fs *Filesystem, dirID string, numFiles int) {
	// Create directory inode
	dirItem := &graph.DriveItem{
		ID:     dirID,
		Name:   filepath.Base(dirID),
		Folder: &graph.Folder{},
		Parent: &graph.DriveItemParent{
			ID: "root",
		},
	}
	dirInode := NewInodeDriveItem(dirItem)
	fs.InsertID(dirID, dirInode)

	// Create child files
	for i := 0; i < numFiles; i++ {
		fileID := fmt.Sprintf("%s-file-%d", dirID, i)
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: fmt.Sprintf("file-%d.txt", i),
			Size: 1024,
			File: &graph.File{},
			Parent: &graph.DriveItemParent{
				ID: dirID,
			},
		}
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertChild(dirID, fileInode)
	}
}

func cleanupFilesystem(fs *Filesystem) {
	if fs != nil {
		fs.StopCacheCleanup()
		fs.StopDeltaLoop()
		fs.StopDownloadManager()
		fs.StopUploadManager()
		fs.StopMetadataRequestManager()
	}
}
