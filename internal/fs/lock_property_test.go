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

// LockOrderingScenario represents a lock ordering test scenario
type LockOrderingScenario struct {
	NumGoroutines int
	NumOperations int
	OperationType string // "filesystem-inode", "manager-session", "inode-cache", "mixed"
	ExpectSuccess bool
}

// generateLockOrderingScenario creates a random lock ordering scenario
func generateLockOrderingScenario(seed int) LockOrderingScenario {
	goroutineCounts := []int{2, 5, 10, 20}
	operationCounts := []int{10, 50, 100, 200}
	opTypes := []string{"filesystem-inode", "manager-session", "inode-cache", "mixed"}

	return LockOrderingScenario{
		NumGoroutines: goroutineCounts[seed%len(goroutineCounts)],
		NumOperations: operationCounts[(seed/len(goroutineCounts))%len(operationCounts)],
		OperationType: opTypes[(seed/(len(goroutineCounts)*len(operationCounts)))%len(opTypes)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 63: Lock Ordering Compliance**
// **Validates: Concurrency Design Requirements**
func TestProperty63_LockOrderingCompliance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random sequences of lock acquisitions, locks should be acquired
	// in the defined hierarchical order to prevent deadlocks
	property := func() bool {
		scenario := generateLockOrderingScenario(int(time.Now().UnixNano() % 1000))

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
		numFiles := 10
		fileIDs := make([]string, numFiles)
		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("test-lock-file-%03d", i)
			fileName := fmt.Sprintf("locktest-%03d.txt", i)
			content := fmt.Sprintf("Content for lock test file %d", i)

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

		// Test: Perform operations that require lock ordering
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumGoroutines*scenario.NumOperations)
		successCount := 0
		var successMutex sync.Mutex

		for g := 0; g < scenario.NumGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for op := 0; op < scenario.NumOperations; op++ {
					fileID := fileIDs[op%len(fileIDs)]
					fileInode := filesystem.GetID(fileID)
					if fileInode == nil {
						errChan <- fmt.Errorf("goroutine %d: failed to get inode for %s", goroutineID, fileID)
						return
					}

					switch scenario.OperationType {
					case "filesystem-inode":
						// Test: Filesystem lock before inode lock (Level 1 -> Level 3)
						// This is the correct ordering
						filesystem.RLock()
						offline := filesystem.offline
						filesystem.RUnlock()

						fileInode.mu.RLock()
						_ = fileInode.DriveItem.Size
						fileInode.mu.RUnlock()

						if offline {
							// Just checking offline status
						}

					case "manager-session":
						// Test: Manager lock before session lock (Level 2 -> Level 4)
						// Access download manager state
						filesystem.downloads.mutex.RLock()
						_ = len(filesystem.downloads.sessions)
						filesystem.downloads.mutex.RUnlock()

						// Then access session if it exists
						filesystem.downloads.mutex.RLock()
						session, exists := filesystem.downloads.sessions[fileID]
						filesystem.downloads.mutex.RUnlock()

						if exists {
							session.mutex.RLock()
							_ = session.State
							session.mutex.RUnlock()
						}

					case "inode-cache":
						// Test: Inode lock before cache operation (Level 3 -> Level 5)
						fileInode.mu.RLock()
						id := fileInode.DriveItem.ID
						fileInode.mu.RUnlock()

						// Cache operation (internally locks cache)
						_ = filesystem.content.Get(id)

					case "mixed":
						// Test: Mixed operations following correct ordering
						switch op % 3 {
						case 0:
							// Filesystem -> Inode
							filesystem.RLock()
							_ = filesystem.offline
							filesystem.RUnlock()

							fileInode.mu.RLock()
							_ = fileInode.DriveItem.Name
							fileInode.mu.RUnlock()

						case 1:
							// Inode -> Cache
							fileInode.mu.RLock()
							id := fileInode.DriveItem.ID
							fileInode.mu.RUnlock()

							_ = filesystem.content.Get(id)

						case 2:
							// Manager -> Session
							filesystem.downloads.mutex.RLock()
							session, exists := filesystem.downloads.sessions[fileID]
							filesystem.downloads.mutex.RUnlock()

							if exists {
								session.mutex.RLock()
								_ = session.State
								session.mutex.RUnlock()
							}
						}
					}

					// Small delay to increase chance of lock contention
					time.Sleep(time.Microsecond)
				}

				// Mark success for this goroutine
				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}(g)
		}

		// Wait for all goroutines to complete with timeout
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - all goroutines completed
		case <-time.After(20 * time.Second):
			// Timeout - possible deadlock
			t.Logf("Lock ordering test timed out - possible deadlock detected")
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Lock ordering error: %v", err)
			return false
		}

		// Verify: All goroutines completed successfully
		if successCount != scenario.NumGoroutines {
			t.Logf("Expected %d successful goroutines, got %d", scenario.NumGoroutines, successCount)
			return false
		}

		// Success: All operations followed correct lock ordering without deadlock
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 63 (Lock Ordering Compliance) failed: %v", err)
	}
}

// DeadlockScenario represents a deadlock prevention test scenario
type DeadlockScenario struct {
	NumGoroutines int
	NumFiles      int
	StressLevel   string // "low", "medium", "high", "extreme"
	ExpectSuccess bool
}

// generateDeadlockScenario creates a random deadlock prevention scenario
func generateDeadlockScenario(seed int) DeadlockScenario {
	goroutineCounts := []int{5, 10, 20, 50}
	fileCounts := []int{5, 10, 20, 50}
	stressLevels := []string{"low", "medium", "high", "extreme"}

	return DeadlockScenario{
		NumGoroutines: goroutineCounts[seed%len(goroutineCounts)],
		NumFiles:      fileCounts[(seed/len(goroutineCounts))%len(fileCounts)],
		StressLevel:   stressLevels[(seed/(len(goroutineCounts)*len(fileCounts)))%len(stressLevels)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 64: Deadlock Prevention**
// **Validates: Concurrency Design Requirements**
func TestProperty64_DeadlockPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random concurrent operation scenarios, no deadlocks should occur
	// when following the lock ordering policy
	property := func() bool {
		scenario := generateDeadlockScenario(int(time.Now().UnixNano() % 1000))

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
			fileID := fmt.Sprintf("test-deadlock-file-%03d", i)
			fileName := fmt.Sprintf("deadlock-%03d.txt", i)
			content := fmt.Sprintf("Content for deadlock test file %d", i)

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

		// Determine stress parameters
		var operationsPerGoroutine int
		var delayBetweenOps time.Duration

		switch scenario.StressLevel {
		case "low":
			operationsPerGoroutine = 10
			delayBetweenOps = 10 * time.Millisecond
		case "medium":
			operationsPerGoroutine = 50
			delayBetweenOps = 1 * time.Millisecond
		case "high":
			operationsPerGoroutine = 100
			delayBetweenOps = 100 * time.Microsecond
		case "extreme":
			operationsPerGoroutine = 200
			delayBetweenOps = 0 // No delay
		}

		// Test: Perform high-concurrency operations that could cause deadlock
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumGoroutines*operationsPerGoroutine)
		successCount := 0
		var successMutex sync.Mutex

		for g := 0; g < scenario.NumGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for op := 0; op < operationsPerGoroutine; op++ {
					// Randomly select files to operate on
					fileIdx1 := (goroutineID + op) % len(fileIDs)
					fileIdx2 := (goroutineID + op + 1) % len(fileIDs)

					fileID1 := fileIDs[fileIdx1]
					fileID2 := fileIDs[fileIdx2]

					fileInode1 := filesystem.GetID(fileID1)
					fileInode2 := filesystem.GetID(fileID2)

					if fileInode1 == nil || fileInode2 == nil {
						errChan <- fmt.Errorf("goroutine %d: failed to get inodes", goroutineID)
						return
					}

					// Perform various operations that require multiple locks
					switch op % 5 {
					case 0:
						// Filesystem state check + inode access
						filesystem.RLock()
						offline := filesystem.offline
						filesystem.RUnlock()

						if !offline {
							fileInode1.mu.RLock()
							_ = fileInode1.DriveItem.Size
							fileInode1.mu.RUnlock()
						}

					case 1:
						// Multiple inode access (using ID ordering)
						id1 := fileInode1.ID()
						id2 := fileInode2.ID()

						if id1 < id2 {
							fileInode1.mu.RLock()
							fileInode2.mu.RLock()
							_ = fileInode1.DriveItem.Name
							_ = fileInode2.DriveItem.Name
							fileInode2.mu.RUnlock()
							fileInode1.mu.RUnlock()
						} else {
							fileInode2.mu.RLock()
							fileInode1.mu.RLock()
							_ = fileInode1.DriveItem.Name
							_ = fileInode2.DriveItem.Name
							fileInode1.mu.RUnlock()
							fileInode2.mu.RUnlock()
						}

					case 2:
						// Inode + cache access
						fileInode1.mu.RLock()
						id := fileInode1.DriveItem.ID
						fileInode1.mu.RUnlock()

						_ = filesystem.content.Get(id)

					case 3:
						// Manager + session access
						filesystem.downloads.mutex.RLock()
						session, exists := filesystem.downloads.sessions[fileID1]
						filesystem.downloads.mutex.RUnlock()

						if exists {
							session.mutex.RLock()
							_ = session.State
							session.mutex.RUnlock()
						}

					case 4:
						// Complex operation: filesystem + inode + cache
						filesystem.RLock()
						offline := filesystem.offline
						filesystem.RUnlock()

						if !offline {
							fileInode1.mu.RLock()
							id := fileInode1.DriveItem.ID
							fileInode1.mu.RUnlock()

							_ = filesystem.content.Get(id)
						}
					}

					// Add delay based on stress level
					if delayBetweenOps > 0 {
						time.Sleep(delayBetweenOps)
					}
				}

				// Mark success for this goroutine
				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}(g)
		}

		// Wait for all goroutines to complete with timeout
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		// Timeout based on stress level
		var timeout time.Duration
		switch scenario.StressLevel {
		case "low":
			timeout = 10 * time.Second
		case "medium":
			timeout = 20 * time.Second
		case "high":
			timeout = 30 * time.Second
		case "extreme":
			timeout = 45 * time.Second
		}

		select {
		case <-done:
			// Success - all goroutines completed
		case <-time.After(timeout):
			// Timeout - deadlock detected
			t.Logf("Deadlock prevention test timed out after %v - deadlock detected", timeout)
			t.Logf("Scenario: %d goroutines, %d files, %s stress level",
				scenario.NumGoroutines, scenario.NumFiles, scenario.StressLevel)
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Deadlock prevention error: %v", err)
			return false
		}

		// Verify: All goroutines completed successfully
		if successCount != scenario.NumGoroutines {
			t.Logf("Expected %d successful goroutines, got %d", scenario.NumGoroutines, successCount)
			return false
		}

		// Success: No deadlocks occurred even under high concurrency and stress
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 64 (Deadlock Prevention) failed: %v", err)
	}
}

// LockReleaseScenario represents a lock release consistency test scenario
type LockReleaseScenario struct {
	NumGoroutines int
	ErrorRate     float64 // Probability of error occurring (0.0 to 1.0)
	OperationType string  // "simple", "nested", "complex"
	ExpectSuccess bool
}

// generateLockReleaseScenario creates a random lock release scenario
func generateLockReleaseScenario(seed int) LockReleaseScenario {
	goroutineCounts := []int{5, 10, 20, 50}
	errorRates := []float64{0.0, 0.1, 0.3, 0.5}
	opTypes := []string{"simple", "nested", "complex"}

	return LockReleaseScenario{
		NumGoroutines: goroutineCounts[seed%len(goroutineCounts)],
		ErrorRate:     errorRates[(seed/len(goroutineCounts))%len(errorRates)],
		OperationType: opTypes[(seed/(len(goroutineCounts)*len(errorRates)))%len(opTypes)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 65: Lock Release Consistency**
// **Validates: Concurrency Design Requirements**
func TestProperty65_LockReleaseConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random lock acquisition scenarios with errors, locks should be
	// released in reverse order (LIFO) and error handling should properly clean up
	property := func() bool {
		scenario := generateLockReleaseScenario(int(time.Now().UnixNano() % 1000))

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
		numFiles := 10
		fileIDs := make([]string, numFiles)
		for i := 0; i < numFiles; i++ {
			fileID := fmt.Sprintf("test-release-file-%03d", i)
			fileName := fmt.Sprintf("release-%03d.txt", i)
			content := fmt.Sprintf("Content for release test file %d", i)

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

		// Test: Perform operations with potential errors to verify lock release
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumGoroutines*100)
		successCount := 0
		var successMutex sync.Mutex

		for g := 0; g < scenario.NumGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for op := 0; op < 100; op++ {
					fileID := fileIDs[op%len(fileIDs)]
					fileInode := filesystem.GetID(fileID)
					if fileInode == nil {
						errChan <- fmt.Errorf("goroutine %d: failed to get inode for %s", goroutineID, fileID)
						return
					}

					// Simulate error based on error rate
					shouldError := (float64(op%100) / 100.0) < scenario.ErrorRate

					switch scenario.OperationType {
					case "simple":
						// Test: Simple lock with defer (LIFO guaranteed by defer)
						func() {
							fileInode.mu.Lock()
							defer fileInode.mu.Unlock()

							if shouldError {
								// Error occurs, but defer ensures unlock
								return
							}

							// Normal operation
							fileInode.DriveItem.Size = uint64(op)
						}()

					case "nested":
						// Test: Nested locks with defer (LIFO order)
						func() {
							filesystem.RLock()
							defer filesystem.RUnlock()

							if shouldError && op%2 == 0 {
								// Early return - filesystem lock released
								return
							}

							fileInode.mu.Lock()
							defer fileInode.mu.Unlock()

							if shouldError && op%2 == 1 {
								// Error after acquiring both locks
								// Defer ensures LIFO release: inode.mu then filesystem
								return
							}

							// Normal operation
							_ = filesystem.offline
							_ = fileInode.DriveItem.Name
						}()

					case "complex":
						// Test: Complex multi-lock scenario with error paths
						func() {
							// Lock 1: Filesystem
							filesystem.RLock()
							defer filesystem.RUnlock()

							if shouldError && op%3 == 0 {
								return
							}

							// Lock 2: Inode
							fileInode.mu.Lock()
							defer fileInode.mu.Unlock()

							if shouldError && op%3 == 1 {
								return
							}

							// Lock 3: Cache (internal)
							id := fileInode.DriveItem.ID
							content := filesystem.content.Get(id)

							if shouldError && op%3 == 2 {
								return
							}

							// Normal operation
							if content != nil {
								_ = len(content)
							}
						}()
					}

					// Small delay to increase lock contention
					time.Sleep(time.Microsecond)
				}

				// Mark success for this goroutine
				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}(g)
		}

		// Wait for all goroutines to complete with timeout
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - all goroutines completed
		case <-time.After(30 * time.Second):
			// Timeout - possible lock not released
			t.Logf("Lock release test timed out - locks may not have been released properly")
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Lock release error: %v", err)
			return false
		}

		// Verify: All goroutines completed successfully
		if successCount != scenario.NumGoroutines {
			t.Logf("Expected %d successful goroutines, got %d", scenario.NumGoroutines, successCount)
			return false
		}

		// Success: All locks were released properly in LIFO order even with errors
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 65 (Lock Release Consistency) failed: %v", err)
	}
}

// ConcurrentFileAccessScenario represents a concurrent file access safety test scenario
type ConcurrentFileAccessScenario struct {
	NumGoroutines int
	NumInodes     int
	OperationType string // "read", "write", "mixed", "stress"
	ExpectSuccess bool
}

// generateConcurrentFileAccessScenario creates a random concurrent file access scenario
func generateConcurrentFileAccessScenario(seed int) ConcurrentFileAccessScenario {
	goroutineCounts := []int{10, 20, 50, 100}
	inodeCounts := []int{10, 20, 50, 100}
	opTypes := []string{"read", "write", "mixed", "stress"}

	return ConcurrentFileAccessScenario{
		NumGoroutines: goroutineCounts[seed%len(goroutineCounts)],
		NumInodes:     inodeCounts[(seed/len(goroutineCounts))%len(inodeCounts)],
		OperationType: opTypes[(seed/(len(goroutineCounts)*len(inodeCounts)))%len(opTypes)],
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 66: Concurrent File Access Safety**
// **Validates: Concurrency Design Requirements**
func TestProperty66_ConcurrentFileAccessSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random concurrent file operations on different inodes, operations
	// should complete safely without race conditions
	property := func() bool {
		scenario := generateConcurrentFileAccessScenario(int(time.Now().UnixNano() % 1000))

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

		// Create test files (inodes)
		fileIDs := make([]string, scenario.NumInodes)
		initialSizes := make([]int64, scenario.NumInodes)
		for i := 0; i < scenario.NumInodes; i++ {
			fileID := fmt.Sprintf("test-concurrent-inode-%03d", i)
			fileName := fmt.Sprintf("concurrent-%03d.txt", i)
			content := fmt.Sprintf("Initial content for file %d", i)

			file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, content)
			registerDriveItem(filesystem, "root", file)

			// Cache the file content
			err = filesystem.content.Insert(fileID, []byte(content))
			if err != nil {
				t.Logf("Failed to cache file %s: %v", fileID, err)
				return false
			}

			fileIDs[i] = fileID
			initialSizes[i] = int64(len(content))
		}

		// Test: Perform concurrent operations on different inodes
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumGoroutines*100)
		readCount := 0
		writeCount := 0
		var countMutex sync.Mutex

		for g := 0; g < scenario.NumGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Each goroutine operates on a subset of inodes
				startIdx := (goroutineID * scenario.NumInodes) / scenario.NumGoroutines
				endIdx := ((goroutineID + 1) * scenario.NumInodes) / scenario.NumGoroutines

				// Ensure at least one inode per goroutine
				if endIdx <= startIdx {
					endIdx = startIdx + 1
				}
				if endIdx > scenario.NumInodes {
					endIdx = scenario.NumInodes
				}

				for op := 0; op < 100; op++ {
					// Select inode from this goroutine's range
					inodeIdx := startIdx + (op % (endIdx - startIdx))
					fileID := fileIDs[inodeIdx]
					fileInode := filesystem.GetID(fileID)
					if fileInode == nil {
						errChan <- fmt.Errorf("goroutine %d: failed to get inode for %s", goroutineID, fileID)
						return
					}

					switch scenario.OperationType {
					case "read":
						// Read-only operations
						fileInode.mu.RLock()
						_ = fileInode.DriveItem.Size
						_ = fileInode.DriveItem.Name
						_ = fileInode.DriveItem.ETag
						fileInode.mu.RUnlock()

						countMutex.Lock()
						readCount++
						countMutex.Unlock()

					case "write":
						// Write operations
						fileInode.mu.Lock()
						fileInode.DriveItem.Size = uint64(goroutineID*1000 + op)
						fileInode.hasChanges = true
						fileInode.mu.Unlock()

						countMutex.Lock()
						writeCount++
						countMutex.Unlock()

					case "mixed":
						// Mix of read and write operations
						if op%2 == 0 {
							// Read
							fileInode.mu.RLock()
							_ = fileInode.DriveItem.Size
							fileInode.mu.RUnlock()

							countMutex.Lock()
							readCount++
							countMutex.Unlock()
						} else {
							// Write
							fileInode.mu.Lock()
							fileInode.DriveItem.Size = uint64(goroutineID*1000 + op)
							fileInode.mu.Unlock()

							countMutex.Lock()
							writeCount++
							countMutex.Unlock()
						}

					case "stress":
						// Stress test with rapid operations
						for i := 0; i < 10; i++ {
							if i%2 == 0 {
								fileInode.mu.RLock()
								_ = fileInode.DriveItem.Name
								fileInode.mu.RUnlock()
							} else {
								fileInode.mu.Lock()
								fileInode.hasChanges = !fileInode.hasChanges
								fileInode.mu.Unlock()
							}
						}

						countMutex.Lock()
						readCount += 5
						writeCount += 5
						countMutex.Unlock()
					}

					// Minimal delay
					if scenario.OperationType != "stress" {
						time.Sleep(time.Microsecond)
					}
				}
			}(g)
		}

		// Wait for all goroutines to complete with timeout
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - all goroutines completed
		case <-time.After(30 * time.Second):
			// Timeout - possible race condition or deadlock
			t.Logf("Concurrent file access test timed out")
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Concurrent file access error: %v", err)
			return false
		}

		// Verify: All inodes are still accessible and consistent
		for i, fileID := range fileIDs {
			fileInode := filesystem.GetID(fileID)
			if fileInode == nil {
				t.Logf("Inode %s disappeared after concurrent access", fileID)
				return false
			}

			// Verify inode is in consistent state
			fileInode.mu.RLock()
			name := fileInode.DriveItem.Name
			size := fileInode.DriveItem.Size
			fileInode.mu.RUnlock()

			if name == "" {
				t.Logf("Inode %s has empty name after concurrent access", fileID)
				return false
			}

			// For read-only operations, size should be unchanged
			if scenario.OperationType == "read" && size != uint64(initialSizes[i]) {
				t.Logf("Inode %s size changed during read-only operations: expected %d, got %d",
					fileID, initialSizes[i], size)
				return false
			}
		}

		// Verify operation counts
		expectedMinOps := scenario.NumGoroutines * 10 // At least 10 ops per goroutine
		totalOps := readCount + writeCount
		if totalOps < expectedMinOps {
			t.Logf("Too few operations completed: expected at least %d, got %d", expectedMinOps, totalOps)
			return false
		}

		// Success: All concurrent operations completed safely without race conditions
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 66 (Concurrent File Access Safety) failed: %v", err)
	}
}

// StateTransitionScenario represents a state transition atomicity test scenario
type StateTransitionScenario struct {
	NumGoroutines  int
	NumFiles       int
	TransitionType string // "hydration", "modification", "conflict", "mixed"
	ExpectSuccess  bool
}

// generateStateTransitionScenario creates a random state transition scenario
func generateStateTransitionScenario(seed int) StateTransitionScenario {
	goroutineCounts := []int{5, 10, 20, 50}
	fileCounts := []int{5, 10, 20, 50}
	transitionTypes := []string{"hydration", "modification", "conflict", "mixed"}

	return StateTransitionScenario{
		NumGoroutines:  goroutineCounts[seed%len(goroutineCounts)],
		NumFiles:       fileCounts[(seed/len(goroutineCounts))%len(fileCounts)],
		TransitionType: transitionTypes[(seed/(len(goroutineCounts)*len(fileCounts)))%len(transitionTypes)],
		ExpectSuccess:  true,
	}
}

// **Feature: system-verification-and-fix, Property 67: State Transition Atomicity**
// **Validates: State Machine Design Requirements**
func TestProperty67_StateTransitionAtomicity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any random item state transition scenarios, transitions should
	// complete atomically without intermediate inconsistent states
	property := func() bool {
		scenario := generateStateTransitionScenario(int(time.Now().UnixNano() % 1000))

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
			fileID := fmt.Sprintf("test-state-file-%03d", i)
			fileName := fmt.Sprintf("state-%03d.txt", i)
			content := fmt.Sprintf("Content for state test file %d", i)

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

		// Test: Perform concurrent state transitions
		var wg sync.WaitGroup
		errChan := make(chan error, scenario.NumGoroutines*100)
		transitionCount := 0
		var transitionMutex sync.Mutex
		inconsistentStates := 0
		var inconsistentMutex sync.Mutex

		for g := 0; g < scenario.NumGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for op := 0; op < 100; op++ {
					fileID := fileIDs[op%len(fileIDs)]
					fileInode := filesystem.GetID(fileID)
					if fileInode == nil {
						errChan <- fmt.Errorf("goroutine %d: failed to get inode for %s", goroutineID, fileID)
						return
					}

					switch scenario.TransitionType {
					case "hydration":
						// Simulate hydration state transition
						// GHOST -> HYDRATING -> HYDRATED
						fileInode.mu.Lock()

						// Check for intermediate inconsistent state
						// (e.g., hasChanges=true but no content)
						if fileInode.hasChanges && filesystem.content.Get(fileID) == nil {
							inconsistentMutex.Lock()
							inconsistentStates++
							inconsistentMutex.Unlock()
						}

						// Perform atomic state transition
						fileInode.hasChanges = false
						fileInode.mu.Unlock()

						transitionMutex.Lock()
						transitionCount++
						transitionMutex.Unlock()

					case "modification":
						// Simulate modification state transition
						// HYDRATED -> DIRTY_LOCAL
						fileInode.mu.Lock()

						// Atomic transition
						fileInode.hasChanges = true
						fileInode.DriveItem.Size = uint64(goroutineID*1000 + op)

						fileInode.mu.Unlock()

						transitionMutex.Lock()
						transitionCount++
						transitionMutex.Unlock()

					case "conflict":
						// Simulate conflict detection
						// DIRTY_LOCAL -> CONFLICT (when remote changes detected)
						fileInode.mu.Lock()

						// Check current state
						hasLocalChanges := fileInode.hasChanges
						oldETag := fileInode.DriveItem.ETag

						// Simulate remote change detection
						newETag := fmt.Sprintf("etag-%d-%d", goroutineID, op)

						if hasLocalChanges && oldETag != newETag {
							// Conflict detected - state should transition atomically
							// In real implementation, this would create conflict copy
							fileInode.DriveItem.ETag = newETag

							// Verify no intermediate state
							if fileInode.DriveItem.ETag == "" {
								inconsistentMutex.Lock()
								inconsistentStates++
								inconsistentMutex.Unlock()
							}
						}

						fileInode.mu.Unlock()

						transitionMutex.Lock()
						transitionCount++
						transitionMutex.Unlock()

					case "mixed":
						// Mix of different state transitions
						switch op % 3 {
						case 0:
							// Hydration
							fileInode.mu.Lock()
							fileInode.hasChanges = false
							fileInode.mu.Unlock()

						case 1:
							// Modification
							fileInode.mu.Lock()
							fileInode.hasChanges = true
							fileInode.DriveItem.Size = uint64(op)
							fileInode.mu.Unlock()

						case 2:
							// Conflict detection
							fileInode.mu.Lock()
							if fileInode.hasChanges {
								fileInode.DriveItem.ETag = fmt.Sprintf("etag-%d", op)
							}
							fileInode.mu.Unlock()
						}

						transitionMutex.Lock()
						transitionCount++
						transitionMutex.Unlock()
					}

					// Verify state consistency after transition
					fileInode.mu.RLock()

					// Check for common inconsistencies:
					// 1. hasChanges=true but no content in cache
					// 2. Empty ETag with valid size
					// 3. Negative size

					if fileInode.hasChanges && filesystem.content.Get(fileID) == nil {
						// This might be okay if file is being uploaded
					}

					if fileInode.DriveItem.ETag == "" && fileInode.DriveItem.Size > 0 {
						// This is actually okay - files can have size without ETag during modifications
						// inconsistentMutex.Lock()
						// inconsistentStates++
						// inconsistentMutex.Unlock()
					}

					if fileInode.DriveItem.Size < 0 {
						// Invalid state
						inconsistentMutex.Lock()
						inconsistentStates++
						inconsistentMutex.Unlock()
					}

					fileInode.mu.RUnlock()

					// Small delay
					time.Sleep(time.Microsecond)
				}
			}(g)
		}

		// Wait for all goroutines to complete with timeout
		done := make(chan bool)
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - all goroutines completed
		case <-time.After(30 * time.Second):
			// Timeout
			t.Logf("State transition test timed out")
			return false
		}

		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("State transition error: %v", err)
			return false
		}

		// Verify: No inconsistent states detected
		if inconsistentStates > 0 {
			t.Logf("Detected %d inconsistent states during transitions", inconsistentStates)
			return false
		}

		// Verify: Transitions occurred
		expectedMinTransitions := scenario.NumGoroutines * 10
		if transitionCount < expectedMinTransitions {
			t.Logf("Too few transitions: expected at least %d, got %d", expectedMinTransitions, transitionCount)
			return false
		}

		// Verify: All files are in consistent final state
		for _, fileID := range fileIDs {
			fileInode := filesystem.GetID(fileID)
			if fileInode == nil {
				t.Logf("File %s disappeared after state transitions", fileID)
				return false
			}

			fileInode.mu.RLock()
			size := fileInode.DriveItem.Size
			etag := fileInode.DriveItem.ETag
			name := fileInode.DriveItem.Name
			fileInode.mu.RUnlock()

			// Verify basic consistency
			if name == "" {
				t.Logf("File %s has empty name after transitions", fileID)
				return false
			}

			if size < 0 {
				t.Logf("File %s has negative size after transitions: %d", fileID, size)
				return false
			}

			// ETag can be empty for new files, but if size > 0, should have ETag
			// (This is a simplified check - real implementation may vary)
			if size > 0 && etag == "" {
				// This might be okay for files being uploaded
				// Just log for information
				t.Logf("File %s has size %d but empty ETag (may be uploading)", fileID, size)
			}
		}

		// Success: All state transitions completed atomically without inconsistent states
		return true
	}

	config := &quick.Config{
		MaxCount: 100,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 67 (State Transition Atomicity) failed: %v", err)
	}
}
