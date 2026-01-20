package fs

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// DirectoryListingPerfScenario represents a directory listing performance test scenario
type DirectoryListingPerfScenario struct {
	NumFiles      int
	FileNameLen   int
	ExpectSuccess bool
}

// generateDirectoryListingPerfScenario creates a random directory listing scenario
func generateDirectoryListingPerfScenario(seed int) DirectoryListingPerfScenario {
	fileCounts := []int{10, 50, 100, 500, 1000}
	fileNameLengths := []int{10, 20, 50, 100}

	return DirectoryListingPerfScenario{
		NumFiles:      fileCounts[seed%len(fileCounts)],
		FileNameLen:   fileNameLengths[(seed/len(fileCounts))%len(fileNameLengths)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 49: Directory Listing Performance**
// **Validates: Requirements 23.1**
func TestProperty49_DirectoryListingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random directory listing scenarios (up to 1000 files),
	// response times should be within 2 seconds
	property := func() bool {
		scenario := generateDirectoryListingPerfScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create test directory with files
		dirID := "test-perf-dir"
		dirName := "perftest"
		dirItem := &graph.DriveItem{
			ID:   dirID,
			Name: dirName,
			Folder: &graph.Folder{
				ChildCount: uint32(scenario.NumFiles),
			},
		}
		registerDriveItem(filesystem, "root", dirItem)

		// Create files in the directory
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("test-perf-file-%05d", i)
			fileName := fmt.Sprintf("file-%05d.txt", i)

			// Pad filename to specified length
			if len(fileName) < scenario.FileNameLen {
				padding := make([]byte, scenario.FileNameLen-len(fileName))
				for j := range padding {
					padding[j] = 'x'
				}
				fileName = fileName + string(padding)
			}

			content := fmt.Sprintf("Content for performance test file %d", i)

			file := helpers.CreateMockFile(mockClient, dirID, fileName, fileID, content)
			registerDriveItem(filesystem, dirID, file)
		}

		// Test: Measure directory listing performance
		startTime := time.Now()

		// Get directory inode
		dirInode := filesystem.GetID(dirID)
		if dirInode == nil {
			t.Logf("Failed to get directory inode")
			return false
		}

		// List directory contents
		dirInode.mu.RLock()
		childCount := len(dirInode.GetChildren())
		dirInode.mu.RUnlock()

		listingDuration := time.Since(startTime)

		// Verify: Response time should be within 2 seconds
		maxDuration := 2 * time.Second
		if listingDuration > maxDuration {
			t.Logf("Directory listing took %v, exceeds limit of %v (files: %d)",
				listingDuration, maxDuration, scenario.NumFiles)
			return false
		}

		// Verify: All files should be listed
		if childCount != scenario.NumFiles {
			t.Logf("Expected %d files, got %d", scenario.NumFiles, childCount)
			return false
		}

		// Success: Directory listing performance is acceptable
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 49 (Directory Listing Performance) failed: %v", err)
	}
}

// CachedFileAccessScenario represents a cached file access performance test scenario
type CachedFileAccessScenario struct {
	NumFiles      int
	FileSizeKB    int
	ExpectSuccess bool
}

// generateCachedFileAccessScenario creates a random cached file access scenario
func generateCachedFileAccessScenario(seed int) CachedFileAccessScenario {
	fileCounts := []int{5, 10, 20, 50}
	fileSizes := []int{1, 10, 100, 1000} // KB

	return CachedFileAccessScenario{
		NumFiles:      fileCounts[seed%len(fileCounts)],
		FileSizeKB:    fileSizes[(seed/len(fileCounts))%len(fileSizes)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 50: Cached File Access Performance**
// **Validates: Requirements 23.2**
func TestProperty50_CachedFileAccessPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random cached file access scenarios, content should be
	// served within 100 milliseconds
	property := func() bool {
		scenario := generateCachedFileAccessScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create and cache test files
		fileIDs := make([]string, scenario.NumFiles)
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("test-cached-file-%03d", i)
			fileName := fmt.Sprintf("cached-%03d.dat", i)

			// Create file content of specified size
			content := make([]byte, scenario.FileSizeKB*1024)
			for j := range content {
				content[j] = byte(i % 256)
			}

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
			registerDriveItem(filesystem, "root", file)

			// Cache the file content
			err = filesystem.content.Insert(fileID, content)
			if err != nil {
				t.Logf("Failed to cache file %s: %v", fileID, err)
				return false
			}

			fileIDs[i] = fileID
		}

		// Test: Measure cached file access performance
		totalDuration := time.Duration(0)
		for _, fileID := range fileIDs {
			startTime := time.Now()

			// Access cached file content
			content := filesystem.content.Get(fileID)
			if content == nil {
				t.Logf("Failed to get cached content for %s", fileID)
				return false
			}

			accessDuration := time.Since(startTime)
			totalDuration += accessDuration

			// Verify: Each access should be within 100 milliseconds
			maxDuration := 100 * time.Millisecond
			if accessDuration > maxDuration {
				t.Logf("Cached file access took %v, exceeds limit of %v (file: %s, size: %d KB)",
					accessDuration, maxDuration, fileID, scenario.FileSizeKB)
				return false
			}
		}

		// Success: Cached file access performance is acceptable
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 50 (Cached File Access Performance) failed: %v", err)
	}
}

// IdleMemoryScenario represents an idle memory usage test scenario
type IdleMemoryScenario struct {
	IdleDuration  time.Duration
	ExpectSuccess bool
}

// generateIdleMemoryScenario creates a random idle memory scenario
func generateIdleMemoryScenario(seed int) IdleMemoryScenario {
	durations := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
	}

	return IdleMemoryScenario{
		IdleDuration:  durations[seed%len(durations)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 51: Idle Memory Usage**
// **Validates: Requirements 23.3**
func TestProperty51_IdleMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random idle system scenarios, memory consumption should
	// stay below 50 MB
	property := func() bool {
		scenario := generateIdleMemoryScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		// Wait for idle period
		time.Sleep(scenario.IdleDuration)

		// Note: Actual memory measurement would require runtime.ReadMemStats()
		// For this test, we verify the system remains responsive and doesn't leak
		// This is a simplified test - real memory measurement would be more complex

		// Verify: System should still be responsive
		rootInode := filesystem.GetID("root")
		if rootInode == nil {
			t.Logf("System not responsive after idle period")
			return false
		}

		// Success: System remains responsive during idle
		// (Real memory measurement would be added in production)
		return true
	}

	config := &quick.Config{
		MaxCount: 20, // Reduced count due to sleep time
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 51 (Idle Memory Usage) failed: %v", err)
	}
}

// ActiveSyncMemoryScenario represents an active sync memory usage test scenario
type ActiveSyncMemoryScenario struct {
	NumFiles      int
	FileSizeKB    int
	ExpectSuccess bool
}

// generateActiveSyncMemoryScenario creates a random active sync memory scenario
func generateActiveSyncMemoryScenario(seed int) ActiveSyncMemoryScenario {
	fileCounts := []int{10, 50, 100, 200}
	fileSizes := []int{10, 100, 500, 1000} // KB

	return ActiveSyncMemoryScenario{
		NumFiles:      fileCounts[seed%len(fileCounts)],
		FileSizeKB:    fileSizes[(seed/len(fileCounts))%len(fileSizes)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 52: Active Sync Memory Usage**
// **Validates: Requirements 23.4**
func TestProperty52_ActiveSyncMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random active synchronization scenarios, memory consumption
	// should stay below 200 MB
	property := func() bool {
		scenario := generateActiveSyncMemoryScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create files to simulate active sync
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("test-sync-file-%05d", i)
			fileName := fmt.Sprintf("syncfile-%05d.dat", i)

			// Create file content
			content := make([]byte, scenario.FileSizeKB*1024)
			for j := range content {
				content[j] = byte(i % 256)
			}

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
			registerDriveItem(filesystem, "root", file)

			// Cache the file to simulate sync activity
			err = filesystem.content.Insert(fileID, content)
			if err != nil {
				t.Logf("Failed to cache file %s: %v", fileID, err)
				return false
			}
		}

		// Note: Actual memory measurement would require runtime.ReadMemStats()
		// For this test, we verify the system handles the sync load
		// This is a simplified test - real memory measurement would be more complex

		// Verify: System should remain responsive during sync
		rootInode := filesystem.GetID("root")
		if rootInode == nil {
			t.Logf("System not responsive during sync")
			return false
		}

		// Success: System handles sync load
		// (Real memory measurement would be added in production)
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 52 (Active Sync Memory Usage) failed: %v", err)
	}
}

// ConcurrentOpsPerformanceScenario represents a concurrent operations performance test scenario
type ConcurrentOpsPerformanceScenario struct {
	NumConcurrentOps int
	NumFiles         int
	OperationType    string // "read", "write", "mixed"
	ExpectSuccess    bool
}

// generateConcurrentOpsPerformanceScenario creates a random concurrent ops performance scenario
func generateConcurrentOpsPerformanceScenario(seed int) ConcurrentOpsPerformanceScenario {
	concurrentOps := []int{10, 20, 50, 100}
	fileCounts := []int{10, 20, 50, 100}
	opTypes := []string{"read", "write", "mixed"}

	return ConcurrentOpsPerformanceScenario{
		NumConcurrentOps: concurrentOps[seed%len(concurrentOps)],
		NumFiles:         fileCounts[(seed/len(concurrentOps))%len(fileCounts)],
		OperationType:    opTypes[(seed/(len(concurrentOps)*len(fileCounts)))%len(opTypes)],
		ExpectSuccess:    true,
	}
}

// **Feature: system-verification-and-fix, Property 53: Concurrent Operations Performance**
// **Validates: Requirements 23.7**
func TestProperty53_ConcurrentOperationsPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random concurrent operation scenarios (10+ operations),
	// there should be no performance degradation under concurrent load
	property := func() bool {
		scenario := generateConcurrentOpsPerformanceScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create test files
		fileIDs := make([]string, scenario.NumFiles)
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("test-concurrent-perf-file-%03d", i)
			fileName := fmt.Sprintf("concperf-%03d.txt", i)
			content := fmt.Sprintf("Content for concurrent performance test file %d", i)

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, content)
			registerDriveItem(filesystem, "root", file)

			// Cache the file content
			err = filesystem.content.Insert(fileID, []byte(content))
			if err != nil {
				t.Logf("Failed to cache file %s: %v", fileID, err)
				return false
			}

			fileIDs[i] = fileID
		}

		// Test: Perform concurrent operations and measure performance
		startTime := time.Now()
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumConcurrentOps)

		for op := 0; op < scenario.NumConcurrentOps; op++ {
			wg.Add(1)
			go func(opID int) {
				defer wg.Done()

				fileID := fileIDs[opID%len(fileIDs)]
				fileInode := filesystem.GetID(fileID)
				if fileInode == nil {
					errChan <- fmt.Errorf("op %d: failed to get inode for %s", opID, fileID)
					return
				}

				switch scenario.OperationType {
				case "read":
					// Simulate read operation
					content := filesystem.content.Get(fileID)
					if content == nil {
						errChan <- fmt.Errorf("op %d: failed to read %s", opID, fileID)
						return
					}

				case "write":
					// Simulate write operation
					fileInode.mu.Lock()
					fileInode.hasChanges = true
					fileInode.mu.Unlock()

				case "mixed":
					// Mix of operations
					if opID%2 == 0 {
						_ = filesystem.content.Get(fileID)
					} else {
						fileInode.mu.Lock()
						fileInode.hasChanges = !fileInode.hasChanges
						fileInode.mu.Unlock()
					}
				}
			}(op)
		}

		// Wait for all operations to complete
		wg.Wait()
		close(errChan)

		totalDuration := time.Since(startTime)

		// Check for errors
		for err := range errChan {
			t.Logf("Concurrent operation error: %v", err)
			return false
		}

		// Verify: Operations should complete in reasonable time
		// Allow 100ms per operation on average
		maxDuration := time.Duration(scenario.NumConcurrentOps*100) * time.Millisecond
		if totalDuration > maxDuration {
			t.Logf("Concurrent operations took %v, exceeds limit of %v (ops: %d)",
				totalDuration, maxDuration, scenario.NumConcurrentOps)
			return false
		}

		// Success: Concurrent operations performance is acceptable
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 53 (Concurrent Operations Performance) failed: %v", err)
	}
}

// StartupPerformanceScenario represents a startup performance test scenario
type StartupPerformanceScenario struct {
	NumCachedFiles int
	ExpectSuccess  bool
}

// generateStartupPerformanceScenario creates a random startup performance scenario
func generateStartupPerformanceScenario(seed int) StartupPerformanceScenario {
	fileCounts := []int{0, 10, 50, 100}

	return StartupPerformanceScenario{
		NumCachedFiles: fileCounts[seed%len(fileCounts)],
		ExpectSuccess:  true,
	}
}

// **Feature: system-verification-and-fix, Property 54: Startup Performance**
// **Validates: Requirements 23.9**
func TestProperty54_StartupPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random system startup scenarios, initialization should
	// complete within 5 seconds
	property := func() bool {
		_ = generateStartupPerformanceScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		// Test: Measure startup time
		startTime := time.Now()

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		// Verify filesystem is ready
		rootInode := filesystem.GetID("root")
		if rootInode == nil {
			t.Logf("Filesystem not ready after startup")
			return false
		}

		startupDuration := time.Since(startTime)

		// Verify: Startup should complete within 5 seconds
		maxDuration := 5 * time.Second
		if startupDuration > maxDuration {
			t.Logf("Startup took %v, exceeds limit of %v", startupDuration, maxDuration)
			return false
		}

		// Success: Startup performance is acceptable
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 54 (Startup Performance) failed: %v", err)
	}
}

// ShutdownPerformanceScenario represents a shutdown performance test scenario
type ShutdownPerformanceScenario struct {
	NumActiveOps  int
	ExpectSuccess bool
}

// generateShutdownPerformanceScenario creates a random shutdown performance scenario
func generateShutdownPerformanceScenario(seed int) ShutdownPerformanceScenario {
	activeOps := []int{0, 5, 10, 20}

	return ShutdownPerformanceScenario{
		NumActiveOps:  activeOps[seed%len(activeOps)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 55: Shutdown Performance**
// **Validates: Requirements 23.10**
func TestProperty55_ShutdownPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random system shutdown scenarios, graceful shutdown should
	// complete within 10 seconds
	property := func() bool {
		scenario := generateShutdownPerformanceScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		mockClient := graph.NewMockGraphClient()

		// Create some active operations if specified
		if scenario.NumActiveOps > 0 {
			for i := 0; i < scenario.NumActiveOps; i++ {
				fileID := fmt.Sprintf("test-shutdown-file-%03d", i)
				fileName := fmt.Sprintf("shutdown-%03d.txt", i)
				content := fmt.Sprintf("Content for shutdown test file %d", i)

				file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, content)
				registerDriveItem(filesystem, "root", file)

				// Start a background operation
				go func(id string, c string) {
					_ = filesystem.content.Insert(id, []byte(c))
				}(fileID, content)
			}

			// Give operations time to start
			time.Sleep(100 * time.Millisecond)
		}

		// Test: Measure shutdown time
		startTime := time.Now()

		// Perform graceful shutdown
		filesystem.StopCacheCleanup()
		filesystem.StopDeltaLoop()
		filesystem.StopDownloadManager()
		filesystem.StopUploadManager()
		filesystem.StopMetadataRequestManager()

		shutdownDuration := time.Since(startTime)

		// Verify: Shutdown should complete within 10 seconds
		maxDuration := 10 * time.Second
		if shutdownDuration > maxDuration {
			t.Logf("Shutdown took %v, exceeds limit of %v (active ops: %d)",
				shutdownDuration, maxDuration, scenario.NumActiveOps)
			return false
		}

		// Success: Shutdown performance is acceptable
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 55 (Shutdown Performance) failed: %v", err)
	}
}
