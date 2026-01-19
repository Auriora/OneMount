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

// ConcurrentAccessScenario represents a concurrent file access test scenario
type ConcurrentAccessScenario struct {
	NumGoroutines int
	NumFiles      int
	OperationType string // "read", "stat", "list", "mixed"
	ExpectSuccess bool
}

// generateConcurrentAccessScenario creates a random concurrent access scenario
func generateConcurrentAccessScenario(seed int) ConcurrentAccessScenario {
	goroutineCounts := []int{2, 5, 10, 20}
	fileCounts := []int{1, 5, 10, 20}
	opTypes := []string{"read", "stat", "list", "mixed"}

	return ConcurrentAccessScenario{
		NumGoroutines: goroutineCounts[seed%len(goroutineCounts)],
		NumFiles:      fileCounts[(seed/len(goroutineCounts))%len(fileCounts)],
		OperationType: opTypes[(seed/(len(goroutineCounts)*len(fileCounts)))%len(opTypes)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 33: Safe Concurrent File Access**
// **Validates: Requirements 12.1**
func TestProperty33_SafeConcurrentFileAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any simultaneous file access operations, the system should handle
	// concurrent operations safely without race conditions
	property := func() bool {
		scenario := generateConcurrentAccessScenario(int(time.Now().UnixNano() % 1000))

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
			fileID := fmt.Sprintf("test-concurrent-file-%03d", i)
			fileName := fmt.Sprintf("testfile-%03d.txt", i)
			content := fmt.Sprintf("Content for file %d", i)

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

		// Test: Perform concurrent operations
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumGoroutines)
		successCount := 0
		var successMutex sync.Mutex

		for g := 0; g < scenario.NumGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Each goroutine performs operations on files
				for _, fileID := range fileIDs {
					fileInode := filesystem.GetID(fileID)
					if fileInode == nil {
						errChan <- fmt.Errorf("goroutine %d: failed to get inode for %s", goroutineID, fileID)
						return
					}

					switch scenario.OperationType {
					case "read":
						// Simulate read operation by accessing file content
						content := filesystem.content.Get(fileID)
						if content == nil {
							errChan <- fmt.Errorf("goroutine %d: failed to read %s: content is nil", goroutineID, fileID)
							return
						}

					case "stat":
						// Simulate stat operation by accessing file metadata
						fileInode.mu.RLock()
						_ = fileInode.DriveItem.Size
						_ = fileInode.DriveItem.Name
						fileInode.mu.RUnlock()

					case "list":
						// Simulate directory listing by accessing parent
						rootInode := filesystem.GetID("root")
						if rootInode != nil {
							rootInode.mu.RLock()
							_ = len(rootInode.GetChildren())
							rootInode.mu.RUnlock()
						}

					case "mixed":
						// Mix of operations
						switch goroutineID % 3 {
						case 0:
							_ = filesystem.content.Get(fileID)
						case 1:
							fileInode.mu.RLock()
							_ = fileInode.DriveItem.Size
							fileInode.mu.RUnlock()
						case 2:
							rootInode := filesystem.GetID("root")
							if rootInode != nil {
								rootInode.mu.RLock()
								_ = len(rootInode.GetChildren())
								rootInode.mu.RUnlock()
							}
						}
					}
				}

				// Mark success for this goroutine
				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}(g)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Concurrent operation error: %v", err)
			return false
		}

		// Verify: All goroutines completed successfully
		if successCount != scenario.NumGoroutines {
			t.Logf("Expected %d successful goroutines, got %d", scenario.NumGoroutines, successCount)
			return false
		}

		// Verify: No data corruption - check that all files still exist and have correct content
		for i, fileID := range fileIDs {
			fileInode := filesystem.GetID(fileID)
			if fileInode == nil {
				t.Logf("File %s disappeared after concurrent access", fileID)
				return false
			}

			// Verify content is still correct
			content := filesystem.content.Get(fileID)
			if content == nil {
				t.Logf("Failed to get content for %s after concurrent access: content is nil", fileID)
				return false
			}

			expectedContent := fmt.Sprintf("Content for file %d", i)
			if string(content) != expectedContent {
				t.Logf("Content corruption detected for %s: expected %q, got %q", fileID, expectedContent, string(content))
				return false
			}
		}

		// Success: All concurrent operations completed safely without race conditions
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 33 (Safe Concurrent File Access) failed: %v", err)
	}
}

// DownloadScenario represents a download test scenario
type DownloadScenario struct {
	NumDownloads      int
	NumOtherOps       int
	DownloadSize      int64
	ExpectNonBlocking bool
}

// generateDownloadScenario creates a random download scenario
func generateDownloadScenario(seed int) DownloadScenario {
	downloadCounts := []int{1, 3, 5, 10}
	otherOpCounts := []int{5, 10, 20, 50}
	downloadSizes := []int64{1024, 10240, 102400, 1024000} // 1KB, 10KB, 100KB, 1MB

	return DownloadScenario{
		NumDownloads:      downloadCounts[seed%len(downloadCounts)],
		NumOtherOps:       otherOpCounts[(seed/len(downloadCounts))%len(otherOpCounts)],
		DownloadSize:      downloadSizes[(seed/(len(downloadCounts)*len(otherOpCounts)))%len(downloadSizes)],
		ExpectNonBlocking: true,
	}
}

// **Feature: system-verification-and-fix, Property 34: Non-blocking Downloads**
// **Validates: Requirements 12.2**
func TestProperty34_NonBlockingDownloads(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any ongoing download operations, other file operations should be
	// allowed to proceed without blocking
	property := func() bool {
		scenario := generateDownloadScenario(int(time.Now().UnixNano() % 1000))

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

		// Create files for download (not cached)
		downloadFileIDs := make([]string, scenario.NumDownloads)
		for i := 0; i < scenario.NumDownloads; i++ {
			fileID := fmt.Sprintf("test-download-file-%03d", i)
			fileName := fmt.Sprintf("download-%03d.txt", i)
			content := make([]byte, scenario.DownloadSize)
			for j := range content {
				content[j] = byte('A' + (j % 26))
			}

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
			registerDriveItem(filesystem, "root", file)

			// Do NOT cache the file - it will need to be downloaded
			downloadFileIDs[i] = fileID
		}

		// Create files for other operations (already cached)
		otherFileIDs := make([]string, scenario.NumOtherOps)
		for i := 0; i < scenario.NumOtherOps; i++ {
			fileID := fmt.Sprintf("test-other-file-%03d", i)
			fileName := fmt.Sprintf("other-%03d.txt", i)
			content := fmt.Sprintf("Content for other file %d", i)

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, content)
			registerDriveItem(filesystem, "root", file)

			// Cache these files so they can be accessed immediately
			err = filesystem.content.Insert(fileID, []byte(content))
			if err != nil {
				t.Logf("Failed to cache file %s: %v", fileID, err)
				return false
			}

			otherFileIDs[i] = fileID
		}

		// Test: Start downloads in background
		var downloadWg sync.WaitGroup
		downloadStarted := make(chan bool, scenario.NumDownloads)

		for _, fileID := range downloadFileIDs {
			downloadWg.Add(1)
			go func(id string) {
				defer downloadWg.Done()
				downloadStarted <- true

				// Simulate download by queuing it
				fileInode := filesystem.GetID(id)
				if fileInode != nil {
					// Queue download (this would normally trigger actual download)
					filesystem.downloads.QueueDownload(id)
				}

				// Simulate download time
				time.Sleep(100 * time.Millisecond)
			}(fileID)
		}

		// Wait for downloads to start
		for i := 0; i < scenario.NumDownloads; i++ {
			<-downloadStarted
		}

		// Test: Perform other operations while downloads are in progress
		otherOpsStartTime := time.Now()
		var otherOpsWg sync.WaitGroup
		otherOpsSuccess := 0
		var otherOpsMutex sync.Mutex

		for _, fileID := range otherFileIDs {
			otherOpsWg.Add(1)
			go func(id string) {
				defer otherOpsWg.Done()

				// Perform read operation on cached file
				fileInode := filesystem.GetID(id)
				if fileInode == nil {
					return
				}

				// Access file metadata (should not block)
				fileInode.mu.RLock()
				_ = fileInode.DriveItem.Size
				_ = fileInode.DriveItem.Name
				fileInode.mu.RUnlock()

				// Access file content (should not block since it's cached)
				content := filesystem.content.Get(id)
				if content != nil {
					otherOpsMutex.Lock()
					otherOpsSuccess++
					otherOpsMutex.Unlock()
				}
			}(fileID)
		}

		// Wait for other operations to complete
		otherOpsWg.Wait()
		otherOpsDuration := time.Since(otherOpsStartTime)

		// Wait for downloads to complete
		downloadWg.Wait()

		// Verify: Other operations completed successfully
		if otherOpsSuccess != scenario.NumOtherOps {
			t.Logf("Expected %d successful other operations, got %d", scenario.NumOtherOps, otherOpsSuccess)
			return false
		}

		// Verify: Other operations completed quickly (not blocked by downloads)
		// They should complete in less than 5 seconds even with downloads running
		maxExpectedDuration := 5 * time.Second
		if otherOpsDuration > maxExpectedDuration {
			t.Logf("Other operations took too long (%v), may have been blocked by downloads", otherOpsDuration)
			return false
		}

		// Success: Other operations proceeded without blocking during downloads
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 34 (Non-blocking Downloads) failed: %v", err)
	}
}
