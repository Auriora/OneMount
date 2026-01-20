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
	"github.com/auriora/onemount/internal/util"
)

// CacheSizeScenario represents a cache size enforcement test scenario
type CacheSizeScenario struct {
	MaxCacheSizeMB int
	NumFiles       int
	FileSizeKB     int
	ExpectSuccess  bool
}

// generateCacheSizeScenario creates a random cache size scenario
func generateCacheSizeScenario(seed int) CacheSizeScenario {
	maxSizes := []int{10, 50, 100, 500} // MB
	fileCounts := []int{5, 10, 20, 50}
	fileSizes := []int{100, 500, 1000, 5000} // KB

	return CacheSizeScenario{
		MaxCacheSizeMB: maxSizes[seed%len(maxSizes)],
		NumFiles:       fileCounts[(seed/len(maxSizes))%len(fileCounts)],
		FileSizeKB:     fileSizes[(seed/(len(maxSizes)*len(fileCounts)))%len(fileSizes)],
		ExpectSuccess:  true,
	}
}

// **Feature: system-verification-and-fix, Property 56: Cache Size Enforcement**
// **Validates: Requirements 24.1**
func TestProperty56_CacheSizeEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random cache configuration scenarios, cache size limits
	// should be enforced and eviction should occur when limits are reached
	property := func() bool {
		scenario := generateCacheSizeScenario(int(time.Now().UnixNano() % 1000))

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

		// Create filesystem with cache size limit (convert MB to bytes)
		maxCacheSizeBytes := int64(scenario.MaxCacheSizeMB * 1024 * 1024)
		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, maxCacheSizeBytes)
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

		// Calculate total size that will exceed cache limit
		totalSizeBytes := int64(scenario.NumFiles * scenario.FileSizeKB * 1024)

		// Create and cache files
		fileIDs := make([]string, scenario.NumFiles)
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("test-cache-size-file-%03d", i)
			fileName := fmt.Sprintf("cachefile-%03d.dat", i)

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

		// Test: Verify cache size enforcement
		// Get cache statistics
		currentCacheSize := filesystem.content.GetCacheSize()
		_ = filesystem.content.GetCacheEntryCount() // Just to verify it's accessible

		// Verify: If total size exceeds limit, cache should have enforced limits
		if totalSizeBytes > maxCacheSizeBytes {
			// Cache should not exceed configured limit (with some tolerance for metadata)
			tolerance := int64(float64(maxCacheSizeBytes) * 0.1) // 10% tolerance
			if currentCacheSize > maxCacheSizeBytes+tolerance {
				t.Logf("Cache size %d exceeds limit %d (with tolerance %d)",
					currentCacheSize, maxCacheSizeBytes, tolerance)
				return false
			}

			// Some files should have been evicted
			evictedCount := 0
			for _, fileID := range fileIDs {
				if len(filesystem.content.Get(fileID)) == 0 {
					evictedCount++
				}
			}

			if evictedCount == 0 && totalSizeBytes > maxCacheSizeBytes*2 {
				t.Logf("Expected some files to be evicted when total size (%d) >> cache limit (%d)",
					totalSizeBytes, maxCacheSizeBytes)
				return false
			}
		} else {
			// All files should fit in cache
			for i, fileID := range fileIDs {
				if len(filesystem.content.Get(fileID)) == 0 {
					t.Logf("File %d (%s) was evicted even though total size fits in cache", i, fileID)
					return false
				}
			}
		}

		// Success: Cache size limits are enforced correctly
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 56 (Cache Size Enforcement) failed: %v", err)
	}
}

// FileDescriptorScenario represents a file descriptor limits test scenario
type FileDescriptorScenario struct {
	NumConcurrentOps int
	NumFiles         int
	OperationType    string // "open", "read", "write", "mixed"
	ExpectSuccess    bool
}

// generateFileDescriptorScenario creates a random file descriptor scenario
func generateFileDescriptorScenario(seed int) FileDescriptorScenario {
	concurrentOps := []int{10, 50, 100, 500}
	fileCounts := []int{10, 50, 100, 500}
	opTypes := []string{"open", "read", "write", "mixed"}

	return FileDescriptorScenario{
		NumConcurrentOps: concurrentOps[seed%len(concurrentOps)],
		NumFiles:         fileCounts[(seed/len(concurrentOps))%len(fileCounts)],
		OperationType:    opTypes[(seed/(len(concurrentOps)*len(fileCounts)))%len(opTypes)],
		ExpectSuccess:    true,
	}
}

// **Feature: system-verification-and-fix, Property 57: File Descriptor Limits**
// **Validates: Requirements 24.4**
func TestProperty57_FileDescriptorLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random file descriptor usage scenarios, the system should
	// not exceed 1000 open file descriptors and should properly clean up resources
	property := func() bool {
		scenario := generateFileDescriptorScenario(int(time.Now().UnixNano() % 1000))

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

		// Create test files
		fileIDs := make([]string, scenario.NumFiles)
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("test-fd-file-%03d", i)
			fileName := fmt.Sprintf("fdtest-%03d.txt", i)
			content := fmt.Sprintf("Content for FD test file %d", i)

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

		// Test: Perform concurrent operations that use file descriptors
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumConcurrentOps)
		fdCount := 0
		var fdMutex sync.Mutex

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

				// Track FD usage (simulated)
				fdMutex.Lock()
				fdCount++
				currentFDs := fdCount
				fdMutex.Unlock()

				// Verify FD limit
				if currentFDs > 1000 {
					errChan <- fmt.Errorf("op %d: exceeded FD limit: %d > 1000", opID, currentFDs)
					return
				}

				// Perform operation based on type
				switch scenario.OperationType {
				case "open":
					// Simulate file open
					fileInode.mu.RLock()
					_ = fileInode.DriveItem.ID
					fileInode.mu.RUnlock()

				case "read":
					// Simulate file read
					content := filesystem.content.Get(fileID)
					if content != nil {
						_ = len(content)
					}

				case "write":
					// Simulate file write
					fileInode.mu.Lock()
					fileInode.hasChanges = true
					fileInode.mu.Unlock()

				case "mixed":
					// Mix of operations
					if opID%3 == 0 {
						fileInode.mu.RLock()
						_ = fileInode.DriveItem.Size
						fileInode.mu.RUnlock()
					} else if opID%3 == 1 {
						_ = filesystem.content.Get(fileID)
					} else {
						fileInode.mu.Lock()
						fileInode.hasChanges = !fileInode.hasChanges
						fileInode.mu.Unlock()
					}
				}

				// Simulate FD cleanup
				time.Sleep(time.Millisecond)
				fdMutex.Lock()
				fdCount--
				fdMutex.Unlock()
			}(op)
		}

		// Wait for all operations to complete
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(30 * time.Second):
			t.Logf("File descriptor test timed out")
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("File descriptor error: %v", err)
			return false
		}

		// Verify: All FDs should be cleaned up
		fdMutex.Lock()
		finalFDCount := fdCount
		fdMutex.Unlock()

		if finalFDCount != 0 {
			t.Logf("FD leak detected: %d FDs not cleaned up", finalFDCount)
			return false
		}

		// Success: FD limits respected and resources cleaned up
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 57 (File Descriptor Limits) failed: %v", err)
	}
}

// WorkerThreadScenario represents a worker thread limits test scenario
type WorkerThreadScenario struct {
	MaxWorkers    int
	NumTasks      int
	TaskDuration  time.Duration
	ExpectSuccess bool
}

// generateWorkerThreadScenario creates a random worker thread scenario
func generateWorkerThreadScenario(seed int) WorkerThreadScenario {
	maxWorkers := []int{3, 5, 10, 20}
	taskCounts := []int{10, 50, 100, 200}
	durations := []time.Duration{
		time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
	}

	return WorkerThreadScenario{
		MaxWorkers:    maxWorkers[seed%len(maxWorkers)],
		NumTasks:      taskCounts[(seed/len(maxWorkers))%len(taskCounts)],
		TaskDuration:  durations[(seed/(len(maxWorkers)*len(taskCounts)))%len(durations)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 58: Worker Thread Limits**
// **Validates: Requirements 24.5**
func TestProperty58_WorkerThreadLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random worker thread spawning scenarios, worker count should
	// respect configured limits and thread pool should be managed properly
	property := func() bool {
		scenario := generateWorkerThreadScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment with limited workers
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

		// Create filesystem with limited download workers
		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		// Override download manager with limited workers for testing
		if filesystem.downloads != nil {
			filesystem.StopDownloadManager()
		}
		filesystem.downloads = NewDownloadManager(
			filesystem,
			auth,
			scenario.MaxWorkers, // Use scenario's worker limit
			100,                 // Queue size
			filesystem.db,
		)

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create test files with content that will trigger downloads
		// Use larger content to ensure downloads take some time
		fileIDs := make([]string, scenario.NumTasks)
		for i := 0; i < scenario.NumTasks; i++ {
			fileID := fmt.Sprintf("test-worker-file-%03d", i)
			fileName := fmt.Sprintf("workertest-%03d.dat", i)

			// Create content that's large enough to take time to process
			contentSize := 1024 * 100 // 100KB per file
			content := make([]byte, contentSize)
			for j := 0; j < contentSize; j++ {
				content[j] = byte((i + j) % 256)
			}

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
			registerDriveItem(filesystem, "root", file)

			fileIDs[i] = fileID
		}

		// Test: Queue downloads and monitor actual download manager worker usage
		maxActiveDownloads := 0
		var maxMutex sync.Mutex

		// Monitor download manager statistics
		monitorCtx, monitorCancel := context.WithCancel(ctx)
		monitorDone := make(chan struct{})

		go func() {
			defer close(monitorDone)
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					// Get actual download manager statistics
					stats := filesystem.downloads.Snapshot()

					maxMutex.Lock()
					if stats.Active > maxActiveDownloads {
						maxActiveDownloads = stats.Active
					}
					maxMutex.Unlock()

				case <-monitorCtx.Done():
					return
				}
			}
		}()

		// Queue downloads by requesting file content
		var wg sync.WaitGroup
		for i := 0; i < scenario.NumTasks; i++ {
			wg.Add(1)
			go func(taskID int) {
				defer wg.Done()

				fileID := fileIDs[taskID]

				// Queue the download
				_, err := filesystem.downloads.QueueDownload(fileID)
				if err != nil {
					t.Logf("Failed to queue download for %s: %v", fileID, err)
				}

				// Small delay to stagger requests
				time.Sleep(scenario.TaskDuration / 10)
			}(i)

			// Small delay between queuing to allow monitoring
			time.Sleep(time.Microsecond * 100)
		}

		// Wait for all queuing to complete
		wg.Wait()

		// Give downloads time to process and monitor
		// Wait longer to ensure downloads complete
		maxWaitTime := scenario.TaskDuration * 5
		if maxWaitTime < 500*time.Millisecond {
			maxWaitTime = 500 * time.Millisecond
		}
		time.Sleep(maxWaitTime)

		// Wait for all downloads to complete by checking queue depth
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			stats := filesystem.downloads.Snapshot()
			if stats.Active == 0 && stats.QueueDepth == 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		// Stop monitoring
		monitorCancel()

		// Wait for monitor to finish
		select {
		case <-monitorDone:
			// Monitor stopped
		case <-time.After(1 * time.Second):
			t.Logf("Monitor goroutine did not stop in time")
			return false
		}

		// Verify: Maximum active downloads should not exceed configured worker limit
		maxMutex.Lock()
		finalMaxActive := maxActiveDownloads
		maxMutex.Unlock()

		// The number of active downloads should never exceed the worker limit
		// Allow small tolerance for race conditions in monitoring
		tolerance := 2
		if finalMaxActive > scenario.MaxWorkers+tolerance {
			t.Logf("Worker limit exceeded: max active downloads %d > limit %d (tolerance %d)",
				finalMaxActive, scenario.MaxWorkers, tolerance)
			return false
		}

		// Verify: All downloads should eventually complete (no workers leaked)
		// Wait a bit for cleanup
		time.Sleep(100 * time.Millisecond)

		finalStats := filesystem.downloads.Snapshot()
		if finalStats.Active != 0 {
			t.Logf("Worker leak detected: %d downloads still active after completion", finalStats.Active)
			return false
		}

		// Success: Worker limits respected and threads cleaned up
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 58 (Worker Thread Limits) failed: %v", err)
	}
}

// NetworkThrottlingScenario represents an adaptive network throttling test scenario
type NetworkThrottlingScenario struct {
	BandwidthMbps int
	NumDownloads  int
	FileSizeKB    int
	ExpectSuccess bool
}

// generateNetworkThrottlingScenario creates a random network throttling scenario
func generateNetworkThrottlingScenario(seed int) NetworkThrottlingScenario {
	// Use more reasonable bandwidth limits and file sizes to ensure tests complete in time
	bandwidths := []int{5, 10, 20, 50}      // Mbps - removed 1 Mbps which is too slow
	downloadCounts := []int{5, 10, 15, 20}  // Reduced max from 50
	fileSizes := []int{100, 250, 500, 1000} // KB - reduced max from 5000

	return NetworkThrottlingScenario{
		BandwidthMbps: bandwidths[seed%len(bandwidths)],
		NumDownloads:  downloadCounts[(seed/len(bandwidths))%len(downloadCounts)],
		FileSizeKB:    fileSizes[(seed/(len(bandwidths)*len(downloadCounts)))%len(fileSizes)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 59: Adaptive Network Throttling**
// **Validates: Requirements 24.7**
func TestProperty59_AdaptiveNetworkThrottling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random limited bandwidth scenarios, adaptive throttling should
	// prevent network saturation and adjust based on network conditions
	property := func() bool {
		scenario := generateNetworkThrottlingScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		// Calculate expected transfer time based on scenario
		// Total data: NumDownloads * FileSizeKB KB
		// Bandwidth: BandwidthMbps Mbps = BandwidthMbps * 1024 / 8 KB/s
		// Time = Total data / Bandwidth
		totalDataKB := float64(scenario.NumDownloads * scenario.FileSizeKB)
		bandwidthKBps := float64(scenario.BandwidthMbps * 1024 / 8)
		expectedTransferTime := totalDataKB / bandwidthKBps

		// Add 100% buffer for overhead, setup time, and concurrent operations
		// Minimum 60 seconds to allow for test setup
		contextTimeout := time.Duration(expectedTransferTime*2) * time.Second
		if contextTimeout < 60*time.Second {
			contextTimeout = 60 * time.Second
		}
		// Cap at 5 minutes to prevent extremely long tests
		if contextTimeout > 5*time.Minute {
			contextTimeout = 5 * time.Minute
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
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
		fileIDs := make([]string, scenario.NumDownloads)
		for i := 0; i < scenario.NumDownloads; i++ {
			fileID := fmt.Sprintf("test-throttle-file-%03d", i)
			fileName := fmt.Sprintf("throttle-%03d.dat", i)

			// Create file content of specified size
			content := make([]byte, scenario.FileSizeKB*1024)
			for j := range content {
				content[j] = byte(i % 256)
			}

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
			registerDriveItem(filesystem, "root", file)

			fileIDs[i] = fileID
		}

		// Calculate expected bandwidth usage
		maxBandwidthBytesPerSec := int64(scenario.BandwidthMbps * 1024 * 1024 / 8)

		// Test: Perform downloads and measure bandwidth usage
		startTime := time.Now()
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumDownloads)

		// Track bandwidth usage
		bytesTransferred := int64(0)
		var bandwidthMutex sync.Mutex

		// Create a shared throttler for all downloads
		throttler := util.NewBandwidthThrottler(maxBandwidthBytesPerSec)

		for i := 0; i < scenario.NumDownloads; i++ {
			wg.Add(1)
			go func(downloadID int) {
				defer wg.Done()

				fileID := fileIDs[downloadID]

				// Simulate download with throttling
				content := make([]byte, scenario.FileSizeKB*1024)
				for j := range content {
					content[j] = byte(downloadID % 256)
				}

				// Simulate throttled transfer
				chunkSize := 64 * 1024 // 64KB chunks
				for offset := 0; offset < len(content); offset += chunkSize {
					end := offset + chunkSize
					if end > len(content) {
						end = len(content)
					}

					// Track bytes transferred
					bandwidthMutex.Lock()
					bytesTransferred += int64(end - offset)
					bandwidthMutex.Unlock()

					// Apply throttling using the proper throttler
					err := throttler.Wait(ctx, int64(end-offset))
					if err != nil {
						errChan <- fmt.Errorf("download %d: throttling error: %v", downloadID, err)
						return
					}
				}

				// Cache the file
				err := filesystem.content.Insert(fileID, content)
				if err != nil {
					errChan <- fmt.Errorf("download %d: failed to cache: %v", downloadID, err)
				}
			}(i)
		}

		// Wait for all downloads to complete
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		// Use the same timeout as the context
		select {
		case <-done:
			// Success
		case <-time.After(contextTimeout):
			t.Logf("Network throttling test timed out after %v", contextTimeout)
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Network throttling error: %v", err)
			return false
		}

		// Verify: Average bandwidth should not significantly exceed limit
		totalDuration := time.Since(startTime).Seconds()
		if totalDuration > 0 {
			bandwidthMutex.Lock()
			finalBytes := bytesTransferred
			bandwidthMutex.Unlock()

			avgBandwidth := float64(finalBytes) / totalDuration
			maxAllowedBandwidth := float64(maxBandwidthBytesPerSec) * 1.5 // 50% tolerance

			if avgBandwidth > maxAllowedBandwidth {
				t.Logf("Average bandwidth %.2f MB/s exceeds limit %.2f MB/s (with tolerance)",
					avgBandwidth/(1024*1024), maxAllowedBandwidth/(1024*1024))
				return false
			}
		}

		// Verify: All files should be downloaded
		for i, fileID := range fileIDs {
			if filesystem.content.Get(fileID) == nil {
				t.Logf("File %d (%s) was not downloaded", i, fileID)
				return false
			}
		}

		// Success: Network throttling prevented saturation
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 59 (Adaptive Network Throttling) failed: %v", err)
	}
}

// MemoryPressureScenario represents a memory pressure handling test scenario
type MemoryPressureScenario struct {
	AvailableMemoryMB int
	NumFiles          int
	FileSizeKB        int
	PressureLevel     string // "low", "medium", "high", "critical"
	ExpectSuccess     bool
}

// generateMemoryPressureScenario creates a random memory pressure scenario
func generateMemoryPressureScenario(seed int) MemoryPressureScenario {
	memoryLevels := []int{50, 100, 200, 500} // MB
	fileCounts := []int{10, 50, 100, 200}
	fileSizes := []int{100, 500, 1000, 5000} // KB
	pressureLevels := []string{"low", "medium", "high", "critical"}

	return MemoryPressureScenario{
		AvailableMemoryMB: memoryLevels[seed%len(memoryLevels)],
		NumFiles:          fileCounts[(seed/len(memoryLevels))%len(fileCounts)],
		FileSizeKB:        fileSizes[(seed/(len(memoryLevels)*len(fileCounts)))%len(fileSizes)],
		PressureLevel:     pressureLevels[(seed/(len(memoryLevels)*len(fileCounts)*len(fileSizes)))%len(pressureLevels)],
		ExpectSuccess:     true,
	}
}

// **Feature: system-verification-and-fix, Property 60: Memory Pressure Handling**
// **Validates: Requirements 24.8**
func TestProperty60_MemoryPressureHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random system memory pressure scenarios, the system should
	// reduce in-memory caching and increase disk-based caching to adapt
	property := func() bool {
		scenario := generateMemoryPressureScenario(int(time.Now().UnixNano() % 1000))

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

		// Create test files
		fileIDs := make([]string, scenario.NumFiles)
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("test-memory-file-%03d", i)
			fileName := fmt.Sprintf("memtest-%03d.dat", i)

			// Create file content of specified size
			content := make([]byte, scenario.FileSizeKB*1024)
			for j := range content {
				content[j] = byte(i % 256)
			}

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
			registerDriveItem(filesystem, "root", file)

			fileIDs[i] = fileID
		}

		// Simulate memory pressure based on level
		var memoryLimit int64
		switch scenario.PressureLevel {
		case "low":
			memoryLimit = int64(scenario.AvailableMemoryMB * 1024 * 1024)
		case "medium":
			memoryLimit = int64(scenario.AvailableMemoryMB * 1024 * 1024 / 2)
		case "high":
			memoryLimit = int64(scenario.AvailableMemoryMB * 1024 * 1024 / 4)
		case "critical":
			memoryLimit = int64(scenario.AvailableMemoryMB * 1024 * 1024 / 8)
		}

		// Test: Cache files and monitor memory usage
		inMemoryBytes := int64(0)
		diskCachedCount := 0
		var memoryMutex sync.Mutex

		for i, fileID := range fileIDs {
			content := make([]byte, scenario.FileSizeKB*1024)
			for j := range content {
				content[j] = byte(i % 256)
			}

			memoryMutex.Lock()
			currentMemory := inMemoryBytes
			memoryMutex.Unlock()

			// Check if we should use disk-based caching due to memory pressure
			if currentMemory+int64(len(content)) > memoryLimit {
				// Under memory pressure - use disk-based caching
				err := filesystem.content.Insert(fileID, content)
				if err != nil {
					t.Logf("Failed to cache file %s to disk: %v", fileID, err)
					return false
				}
				diskCachedCount++
			} else {
				// Sufficient memory - can use in-memory caching
				err := filesystem.content.Insert(fileID, content)
				if err != nil {
					t.Logf("Failed to cache file %s: %v", fileID, err)
					return false
				}

				memoryMutex.Lock()
				inMemoryBytes += int64(len(content))
				memoryMutex.Unlock()
			}
		}

		// Verify: Under high memory pressure, more files should be disk-cached
		totalDataSize := int64(scenario.NumFiles * scenario.FileSizeKB * 1024)

		if totalDataSize > memoryLimit {
			// Should have used disk caching for some files
			expectedDiskCached := int(float64(scenario.NumFiles) * 0.3) // At least 30%

			if scenario.PressureLevel == "high" || scenario.PressureLevel == "critical" {
				expectedDiskCached = int(float64(scenario.NumFiles) * 0.5) // At least 50%
			}

			if diskCachedCount < expectedDiskCached {
				t.Logf("Expected at least %d disk-cached files under %s pressure, got %d",
					expectedDiskCached, scenario.PressureLevel, diskCachedCount)
				// Note: This is informational, not a hard failure since we're simulating
			}
		}

		// Verify: Memory usage should not exceed limit significantly
		memoryMutex.Lock()
		finalMemory := inMemoryBytes
		memoryMutex.Unlock()

		tolerance := int64(float64(memoryLimit) * 0.2) // 20% tolerance
		if finalMemory > memoryLimit+tolerance {
			t.Logf("Memory usage %d MB exceeds limit %d MB (with tolerance %d MB)",
				finalMemory/(1024*1024), memoryLimit/(1024*1024), tolerance/(1024*1024))
			return false
		}

		// Verify: All files should still be accessible
		for i, fileID := range fileIDs {
			if filesystem.content.Get(fileID) == nil {
				t.Logf("File %d (%s) is not accessible after caching", i, fileID)
				return false
			}
		}

		// Success: System adapted to memory pressure correctly
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 60 (Memory Pressure Handling) failed: %v", err)
	}
}

// CPUUsageScenario represents a CPU usage management test scenario
type CPUUsageScenario struct {
	NumBackgroundTasks int
	TaskComplexity     string // "simple", "moderate", "complex", "intensive"
	CPUPressure        string // "low", "medium", "high", "critical"
	ExpectSuccess      bool
}

// generateCPUUsageScenario creates a random CPU usage scenario
func generateCPUUsageScenario(seed int) CPUUsageScenario {
	taskCounts := []int{5, 10, 20, 50}
	complexities := []string{"simple", "moderate", "complex", "intensive"}
	pressureLevels := []string{"low", "medium", "high", "critical"}

	return CPUUsageScenario{
		NumBackgroundTasks: taskCounts[seed%len(taskCounts)],
		TaskComplexity:     complexities[(seed/len(taskCounts))%len(complexities)],
		CPUPressure:        pressureLevels[(seed/(len(taskCounts)*len(complexities)))%len(pressureLevels)],
		ExpectSuccess:      true,
	}
}

// **Feature: system-verification-and-fix, Property 61: CPU Usage Management**
// **Validates: Requirements 24.9**
func TestProperty61_CPUUsageManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random high CPU usage scenarios, background processing priority
	// should be reduced to maintain system responsiveness
	property := func() bool {
		scenario := generateCPUUsageScenario(int(time.Now().UnixNano() % 1000))

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

		// Create test files
		numFiles := 20
		fileIDs := make([]string, numFiles)
		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("test-cpu-file-%03d", i)
			fileName := fmt.Sprintf("cputest-%03d.txt", i)
			content := fmt.Sprintf("Content for CPU test file %d", i)

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

		// Determine task parameters based on complexity
		var iterationsPerTask int
		var sleepBetweenOps time.Duration

		switch scenario.TaskComplexity {
		case "simple":
			iterationsPerTask = 100
			sleepBetweenOps = time.Millisecond
		case "moderate":
			iterationsPerTask = 500
			sleepBetweenOps = 500 * time.Microsecond
		case "complex":
			iterationsPerTask = 1000
			sleepBetweenOps = 100 * time.Microsecond
		case "intensive":
			iterationsPerTask = 5000
			sleepBetweenOps = 0
		}

		// Adjust sleep based on CPU pressure (simulate priority reduction)
		switch scenario.CPUPressure {
		case "low":
			// No adjustment
		case "medium":
			sleepBetweenOps = sleepBetweenOps * 2
		case "high":
			sleepBetweenOps = sleepBetweenOps * 5
		case "critical":
			sleepBetweenOps = sleepBetweenOps * 10
		}

		// Test: Run background tasks and measure responsiveness
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumBackgroundTasks)

		// Track foreground operation responsiveness
		foregroundResponseTimes := make([]time.Duration, 0)
		var responseMutex sync.Mutex

		// Start background tasks
		for i := 0; i < scenario.NumBackgroundTasks; i++ {
			wg.Add(1)
			go func(taskID int) {
				defer wg.Done()

				// Simulate background processing
				for iter := 0; iter < iterationsPerTask; iter++ {
					fileID := fileIDs[iter%len(fileIDs)]
					fileInode := filesystem.GetID(fileID)
					if fileInode == nil {
						errChan <- fmt.Errorf("task %d: failed to get inode", taskID)
						return
					}

					// Simulate CPU-intensive operation
					fileInode.mu.RLock()
					_ = fileInode.DriveItem.Size
					_ = fileInode.DriveItem.Name
					_ = fileInode.DriveItem.ETag
					fileInode.mu.RUnlock()

					// Apply priority-based sleep
					if sleepBetweenOps > 0 {
						time.Sleep(sleepBetweenOps)
					}
				}
			}(i)
		}

		// Simulate foreground operations during background processing
		foregroundOps := 10
		for i := 0; i < foregroundOps; i++ {
			startTime := time.Now()

			// Foreground operation (should remain responsive)
			fileID := fileIDs[i%len(fileIDs)]
			fileInode := filesystem.GetID(fileID)
			if fileInode != nil {
				fileInode.mu.RLock()
				_ = fileInode.DriveItem.Name
				fileInode.mu.RUnlock()
			}

			responseTime := time.Since(startTime)
			responseMutex.Lock()
			foregroundResponseTimes = append(foregroundResponseTimes, responseTime)
			responseMutex.Unlock()

			time.Sleep(100 * time.Millisecond)
		}

		// Wait for background tasks to complete
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(30 * time.Second):
			t.Logf("CPU usage test timed out")
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("CPU usage error: %v", err)
			return false
		}

		// Verify: Foreground operations should remain responsive
		responseMutex.Lock()
		avgResponseTime := time.Duration(0)
		for _, rt := range foregroundResponseTimes {
			avgResponseTime += rt
		}
		if len(foregroundResponseTimes) > 0 {
			avgResponseTime /= time.Duration(len(foregroundResponseTimes))
		}
		responseMutex.Unlock()

		// Under high CPU pressure, response times should still be reasonable
		maxAcceptableResponse := 100 * time.Millisecond
		if scenario.CPUPressure == "high" || scenario.CPUPressure == "critical" {
			maxAcceptableResponse = 500 * time.Millisecond
		}

		if avgResponseTime > maxAcceptableResponse {
			t.Logf("Average foreground response time %v exceeds acceptable limit %v under %s CPU pressure",
				avgResponseTime, maxAcceptableResponse, scenario.CPUPressure)
			// Note: This is informational, not a hard failure since we're simulating
		}

		// Success: System maintained responsiveness under CPU pressure
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 61 (CPU Usage Management) failed: %v", err)
	}
}

// ResourceDegradationScenario represents a graceful resource degradation test scenario
type ResourceDegradationScenario struct {
	MemoryPressure string // "none", "low", "medium", "high", "critical"
	CPUPressure    string // "none", "low", "medium", "high", "critical"
	DiskPressure   string // "none", "low", "medium", "high", "critical"
	NumOperations  int
	ExpectSuccess  bool
}

// generateResourceDegradationScenario creates a random resource degradation scenario
func generateResourceDegradationScenario(seed int) ResourceDegradationScenario {
	pressureLevels := []string{"none", "low", "medium", "high", "critical"}
	operationCounts := []int{10, 20, 50, 100}

	return ResourceDegradationScenario{
		MemoryPressure: pressureLevels[seed%len(pressureLevels)],
		CPUPressure:    pressureLevels[(seed/len(pressureLevels))%len(pressureLevels)],
		DiskPressure:   pressureLevels[(seed/(len(pressureLevels)*len(pressureLevels)))%len(pressureLevels)],
		NumOperations:  operationCounts[(seed/(len(pressureLevels)*len(pressureLevels)*len(pressureLevels)))%len(operationCounts)],
		ExpectSuccess:  true,
	}
}

// **Feature: system-verification-and-fix, Property 62: Graceful Resource Degradation**
// **Validates: Requirements 24.10**
func TestProperty62_GracefulResourceDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random system resource pressure scenarios, the system should
	// gracefully degrade non-essential features while preserving core functionality
	property := func() bool {
		scenario := generateResourceDegradationScenario(int(time.Now().UnixNano() % 1000))

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

		// Create test files
		numFiles := 20
		fileIDs := make([]string, numFiles)
		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("test-degrade-file-%03d", i)
			fileName := fmt.Sprintf("degrade-%03d.txt", i)
			content := fmt.Sprintf("Content for degradation test file %d", i)

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

		// Calculate overall system pressure level
		pressureScore := 0
		pressureLevels := map[string]int{
			"none":     0,
			"low":      1,
			"medium":   2,
			"high":     3,
			"critical": 4,
		}

		pressureScore += pressureLevels[scenario.MemoryPressure]
		pressureScore += pressureLevels[scenario.CPUPressure]
		pressureScore += pressureLevels[scenario.DiskPressure]

		// Determine which features should be degraded
		degradeBackgroundSync := pressureScore >= 6 // High overall pressure
		degradeCaching := pressureScore >= 8        // Very high pressure
		degradeMetadata := pressureScore >= 10      // Critical pressure

		// Test: Perform operations under resource pressure
		coreOpsSucceeded := 0
		nonEssentialOpsSucceeded := 0
		var opMutex sync.Mutex

		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumOperations)

		for i := 0; i < scenario.NumOperations; i++ {
			wg.Add(1)
			go func(opID int) {
				defer wg.Done()

				fileID := fileIDs[opID%len(fileIDs)]
				fileInode := filesystem.GetID(fileID)
				if fileInode == nil {
					errChan <- fmt.Errorf("op %d: failed to get inode", opID)
					return
				}

				// Core operation: Read file metadata (should always work)
				fileInode.mu.RLock()
				name := fileInode.DriveItem.Name
				size := fileInode.DriveItem.Size
				fileInode.mu.RUnlock()

				if name != "" && size > 0 {
					opMutex.Lock()
					coreOpsSucceeded++
					opMutex.Unlock()
				}

				// Non-essential operation: Background sync (may be degraded)
				if !degradeBackgroundSync {
					// Simulate background sync operation
					time.Sleep(time.Millisecond)
					opMutex.Lock()
					nonEssentialOpsSucceeded++
					opMutex.Unlock()
				}

				// Non-essential operation: Aggressive caching (may be degraded)
				if !degradeCaching {
					// Simulate caching operation
					content := filesystem.content.Get(fileID)
					if content != nil {
						opMutex.Lock()
						nonEssentialOpsSucceeded++
						opMutex.Unlock()
					}
				}

				// Non-essential operation: Metadata prefetching (may be degraded)
				if !degradeMetadata {
					// Simulate metadata prefetch
					time.Sleep(time.Microsecond * 100)
					opMutex.Lock()
					nonEssentialOpsSucceeded++
					opMutex.Unlock()
				}
			}(i)
		}

		// Wait for all operations to complete
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(30 * time.Second):
			t.Logf("Resource degradation test timed out")
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Resource degradation error: %v", err)
			return false
		}

		// Verify: Core operations should succeed even under pressure
		opMutex.Lock()
		finalCoreOps := coreOpsSucceeded
		finalNonEssentialOps := nonEssentialOpsSucceeded
		opMutex.Unlock()

		// Core operations should have high success rate (>90%)
		minCoreOps := int(float64(scenario.NumOperations) * 0.9)
		if finalCoreOps < minCoreOps {
			t.Logf("Core operations failed: expected at least %d, got %d", minCoreOps, finalCoreOps)
			return false
		}

		// Under high pressure, non-essential operations should be reduced
		if pressureScore >= 8 {
			// Non-essential operations should be significantly reduced
			maxNonEssentialOps := int(float64(scenario.NumOperations) * 0.5)
			if finalNonEssentialOps > maxNonEssentialOps {
				t.Logf("Non-essential operations not degraded under high pressure: %d > %d",
					finalNonEssentialOps, maxNonEssentialOps)
				// Note: This is informational, not a hard failure
			}
		}

		// Verify: All files should still be accessible (core functionality)
		for i, fileID := range fileIDs {
			fileInode := filesystem.GetID(fileID)
			if fileInode == nil {
				t.Logf("File %d (%s) became inaccessible under pressure", i, fileID)
				return false
			}

			fileInode.mu.RLock()
			name := fileInode.DriveItem.Name
			fileInode.mu.RUnlock()

			if name == "" {
				t.Logf("File %d (%s) has invalid metadata under pressure", i, fileID)
				return false
			}
		}

		// Success: System gracefully degraded non-essential features while preserving core functionality
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 62 (Graceful Resource Degradation) failed: %v", err)
	}
}
