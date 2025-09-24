package fs

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/stretchr/testify/require"
)

// TestConcurrentFileAccess tests concurrent file access scenarios
//
//	Test Case ID    CT-FS-01-01
//	Title           Concurrent File Access
//	Description     Tests concurrent access to files by multiple goroutines
//	Preconditions   Filesystem is initialized and online
//	Steps           1. Create test files
//	                2. Launch multiple goroutines to access files concurrently
//	                3. Verify no data corruption or race conditions
//	Expected Result All operations complete successfully without corruption
func TestConcurrentFileAccess(t *testing.T) {
	// Create a test fixture
	fixture := helpers.SetupFSTestFixture(t, "ConcurrentFileAccessFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Test parameters
		const (
			numGoroutines          = 10
			numFiles               = 5
			operationsPerGoroutine = 20
		)

		// Create test files
		testFiles := make([]*Inode, numFiles)
		for i := 0; i < numFiles; i++ {
			fileName := fmt.Sprintf("test_file_%d.txt", i)
			fileItem := &graph.DriveItem{
				ID:   fmt.Sprintf("file_%d", i),
				Name: fileName,
				Size: 1024,
				File: &graph.File{},
			}
			testFiles[i] = NewInodeDriveItem(fileItem)
			fs.InsertID(fileItem.ID, testFiles[i])
		}

		// Counters for tracking operations
		var (
			readCount    int64
			writeCount   int64
			successCount int64
			errorCount   int64
		)

		// Wait group for synchronization
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Launch concurrent goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					// Select a random file
					fileIndex := rand.Intn(numFiles)
					inode := testFiles[fileIndex]

					// Perform random operation
					switch rand.Intn(3) {
					case 0: // Read operation
						content := fs.GetInodeContent(inode)
						if content != nil {
							atomic.AddInt64(&readCount, 1)
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&errorCount, 1)
						}

					case 1: // Write operation (simulate)
						inode.Lock()
						inode.hasChanges = true
						inode.Unlock()
						atomic.AddInt64(&writeCount, 1)
						atomic.AddInt64(&successCount, 1)

					case 2: // Metadata access
						_ = inode.Name()
						_ = inode.Size()
						_ = inode.NodeID()
						atomic.AddInt64(&successCount, 1)
					}

					// Small delay to increase chance of race conditions
					time.Sleep(time.Microsecond * time.Duration(rand.Intn(100)))
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify results
		totalOperations := int64(numGoroutines * operationsPerGoroutine)
		actualOperations := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&errorCount)

		assert.Equal(totalOperations, actualOperations, "All operations should be accounted for")
		assert.True(atomic.LoadInt64(&successCount) > 0, "Should have successful operations")
		assert.True(atomic.LoadInt64(&readCount) > 0, "Should have read operations")
		assert.True(atomic.LoadInt64(&writeCount) > 0, "Should have write operations")

		// Verify filesystem integrity
		for i, inode := range testFiles {
			assert.NotNil(fs.GetID(fmt.Sprintf("file_%d", i)), "File should still exist in filesystem")
			assert.Equal(fmt.Sprintf("test_file_%d.txt", i), inode.Name(), "File name should be intact")
		}
	})
}

// TestConcurrentCacheOperations tests cache consistency under concurrent load
//
//	Test Case ID    CT-FS-02-01
//	Title           Cache Consistency Under Load
//	Description     Tests cache operations with multiple concurrent readers/writers
//	Preconditions   Filesystem is initialized with cache
//	Steps           1. Create cache entries
//	                2. Launch concurrent cache operations
//	                3. Verify cache consistency
//	Expected Result Cache remains consistent under concurrent access
func TestConcurrentCacheOperations(t *testing.T) {
	// Create a test fixture
	fixture := helpers.SetupFSTestFixture(t, "ConcurrentCacheFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Test parameters
		const (
			numGoroutines          = 15
			numCacheEntries        = 20
			operationsPerGoroutine = 30
		)

		// Create cache entries
		cacheEntries := make(map[string]*Inode)
		for i := 0; i < numCacheEntries; i++ {
			entryID := fmt.Sprintf("cache_entry_%d", i)
			entryItem := &graph.DriveItem{
				ID:   entryID,
				Name: fmt.Sprintf("cache_file_%d.txt", i),
				Size: uint64(1024 + i*100),
				File: &graph.File{},
			}
			inode := NewInodeDriveItem(entryItem)
			cacheEntries[entryID] = inode
			fs.InsertID(entryID, inode)
		}

		// Counters for tracking operations
		var (
			insertCount  int64
			lookupCount  int64
			successCount int64
			errorCount   int64
		)

		// Wait group for synchronization
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Launch concurrent goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					entryIndex := rand.Intn(numCacheEntries)
					entryID := fmt.Sprintf("cache_entry_%d", entryIndex)

					// Perform random cache operation
					switch rand.Intn(4) {
					case 0: // Cache lookup
						inode := fs.GetID(entryID)
						if inode != nil {
							atomic.AddInt64(&lookupCount, 1)
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&errorCount, 1)
						}

					case 1: // Cache insert (re-insert existing)
						if originalInode, exists := cacheEntries[entryID]; exists {
							fs.InsertID(entryID, originalInode)
							atomic.AddInt64(&insertCount, 1)
							atomic.AddInt64(&successCount, 1)
						}

					case 2: // Metadata access
						if inode := fs.GetID(entryID); inode != nil {
							_ = inode.Name()
							_ = inode.Size()
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&errorCount, 1)
						}

					case 3: // Node ID operations
						if inode := fs.GetID(entryID); inode != nil {
							nodeID := inode.NodeID()
							retrievedInode := fs.GetNodeID(nodeID)
							if retrievedInode == inode {
								atomic.AddInt64(&successCount, 1)
							} else {
								atomic.AddInt64(&errorCount, 1)
							}
						} else {
							atomic.AddInt64(&errorCount, 1)
						}
					}

					// Small delay to increase chance of race conditions
					time.Sleep(time.Microsecond * time.Duration(rand.Intn(50)))
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify results
		totalOperations := int64(numGoroutines * operationsPerGoroutine)
		actualOperations := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&errorCount)

		assert.Equal(totalOperations, actualOperations, "All operations should be accounted for")
		assert.True(atomic.LoadInt64(&successCount) > 0, "Should have successful operations")
		assert.True(atomic.LoadInt64(&lookupCount) > 0, "Should have lookup operations")

		// Verify cache integrity - all original entries should still exist
		for entryID, originalInode := range cacheEntries {
			cachedInode := fs.GetID(entryID)
			assert.NotNil(cachedInode, "Cache entry should still exist: %s", entryID)
			if cachedInode != nil {
				assert.Equal(originalInode.Name(), cachedInode.Name(), "Cache entry name should be intact")
				assert.Equal(originalInode.Size(), cachedInode.Size(), "Cache entry size should be intact")
			}
		}
	})
}

// TestDeadlockPrevention tests for potential deadlocks in filesystem operations
//
//	Test Case ID    CT-FS-03-01
//	Title           Deadlock Prevention
//	Description     Tests filesystem operations for potential deadlocks
//	Preconditions   Filesystem is initialized
//	Steps           1. Create scenarios that could cause deadlocks
//	                2. Execute operations with timeouts
//	                3. Verify no deadlocks occur
//	Expected Result All operations complete within timeout, no deadlocks
func TestDeadlockPrevention(t *testing.T) {
	// Create a test fixture
	fixture := helpers.SetupFSTestFixture(t, "DeadlockPreventionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		requireAssert := require.New(t)
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Test parameters
		const (
			numGoroutines    = 8
			testDuration     = 10 * time.Second
			operationTimeout = 5 * time.Second
		)

		// Create test files for potential lock contention
		testFiles := make([]*Inode, 5)
		for i := 0; i < 5; i++ {
			fileItem := &graph.DriveItem{
				ID:   fmt.Sprintf("deadlock_test_file_%d", i),
				Name: fmt.Sprintf("deadlock_file_%d.txt", i),
				Size: 1024,
				File: &graph.File{},
			}
			testFiles[i] = NewInodeDriveItem(fileItem)
			fs.InsertID(fileItem.ID, testFiles[i])
		}

		// Context with timeout for the entire test
		ctx, cancel := context.WithTimeout(context.Background(), testDuration)
		defer cancel()

		// Counters for tracking operations
		var (
			completedOperations int64
			timeoutOperations   int64
			errorOperations     int64
		)

		// Wait group for synchronization
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Launch goroutines that could potentially deadlock
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				for {
					select {
					case <-ctx.Done():
						return
					default:
						// Create operation context with timeout
						opCtx, opCancel := context.WithTimeout(ctx, operationTimeout)

						// Perform operation that could potentially deadlock
						err := func() error {
							// Channel to signal operation completion
							done := make(chan error, 1)

							go func() {
								defer func() {
									if r := recover(); r != nil {
										done <- fmt.Errorf("panic in operation: %v", r)
									}
								}()

								// Simulate complex operations that involve multiple locks
								switch rand.Intn(4) {
								case 0: // Multiple file access
									for _, file := range testFiles {
										_ = file.Name()
										_ = file.Size()
									}

								case 1: // Cache operations with filesystem operations
									for _, file := range testFiles {
										cachedFile := fs.GetID(file.ID())
										if cachedFile != nil {
											_ = cachedFile.NodeID()
										}
									}

								case 2: // Mixed read/write operations
									file := testFiles[rand.Intn(len(testFiles))]
									file.RLock()
									_ = file.Name()
									file.RUnlock()

									file.Lock()
									file.hasChanges = true
									file.Unlock()

								case 3: // Filesystem metadata operations
									for _, file := range testFiles {
										nodeID := file.NodeID()
										retrievedFile := fs.GetNodeID(nodeID)
										if retrievedFile != nil {
											_ = retrievedFile.Name()
										}
									}
								}

								done <- nil
							}()

							// Wait for operation to complete or timeout
							select {
							case err := <-done:
								return err
							case <-opCtx.Done():
								return opCtx.Err()
							}
						}()

						opCancel()

						// Track operation result
						if err != nil {
							if err == context.DeadlineExceeded {
								atomic.AddInt64(&timeoutOperations, 1)
							} else {
								atomic.AddInt64(&errorOperations, 1)
							}
						} else {
							atomic.AddInt64(&completedOperations, 1)
						}

						// Small delay between operations
						time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete or timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// All goroutines completed successfully
		case <-time.After(testDuration + 5*time.Second):
			t.Fatal("Test timed out - possible deadlock detected")
		}

		// Verify results
		completed := atomic.LoadInt64(&completedOperations)
		timeouts := atomic.LoadInt64(&timeoutOperations)
		errors := atomic.LoadInt64(&errorOperations)

		assert.True(completed > 0, "Should have completed some operations")

		// Allow some timeouts but not too many (which could indicate deadlocks)
		timeoutRatio := float64(timeouts) / float64(completed+timeouts+errors)
		assert.True(timeoutRatio < 0.1, "Timeout ratio should be less than 10%%, got %.2f%%", timeoutRatio*100)

		// Verify filesystem is still functional after the test
		for i, file := range testFiles {
			cachedFile := fs.GetID(fmt.Sprintf("deadlock_test_file_%d", i))
			requireAssert.NotNil(cachedFile, "File should still be accessible after deadlock test")
			assert.Equal(file.Name(), cachedFile.Name(), "File name should be intact")
		}
	})
}

// TestHighConcurrencyStress tests system behavior under extreme concurrent load
//
//	Test Case ID    CT-FS-04-01
//	Title           High-Concurrency Stress Test
//	Description     Tests filesystem under extreme concurrent load
//	Preconditions   Filesystem is initialized
//	Steps           1. Create many concurrent operations
//	                2. Monitor system resources and performance
//	                3. Verify system stability under load
//	Expected Result System remains stable and responsive under high load
func TestHighConcurrencyStress(t *testing.T) {
	// Skip this test in short mode as it's resource intensive
	if testing.Short() {
		t.Skip("Skipping high-concurrency stress test in short mode")
	}

	// Create a test fixture
	fixture := helpers.SetupFSTestFixture(t, "HighConcurrencyStressFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Test parameters for stress testing
		const (
			numGoroutines       = 50
			numFiles            = 100
			testDuration        = 30 * time.Second
			operationsPerSecond = 100
		)

		// Create many test files for stress testing
		testFiles := make([]*Inode, numFiles)
		for i := 0; i < numFiles; i++ {
			fileItem := &graph.DriveItem{
				ID:   fmt.Sprintf("stress_test_file_%d", i),
				Name: fmt.Sprintf("stress_file_%d.txt", i),
				Size: uint64(1024 + i*10),
				File: &graph.File{},
			}
			testFiles[i] = NewInodeDriveItem(fileItem)
			fs.InsertID(fileItem.ID, testFiles[i])
		}

		// Context with timeout for the stress test
		ctx, cancel := context.WithTimeout(context.Background(), testDuration)
		defer cancel()

		// Counters for tracking operations and performance
		var (
			totalOperations int64
			successfulOps   int64
			failedOps       int64
			cacheHits       int64
			cacheMisses     int64
			lockContentions int64
		)

		// Performance tracking
		var (
			minLatency   = int64(^uint64(0) >> 1) // Max int64
			maxLatency   int64
			totalLatency int64
		)

		// Wait group for synchronization
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Launch high-concurrency stress test goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				// Rate limiting to control operations per second
				ticker := time.NewTicker(time.Second / operationsPerSecond)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						// Measure operation latency
						startTime := time.Now()

						// Perform random high-stress operation
						err := func() error {
							switch rand.Intn(6) {
							case 0: // Intensive file access pattern
								for j := 0; j < 10; j++ {
									file := testFiles[rand.Intn(numFiles)]
									_ = file.Name()
									_ = file.Size()
									_ = file.NodeID()
								}

							case 1: // Cache thrashing pattern
								for j := 0; j < 5; j++ {
									fileID := fmt.Sprintf("stress_test_file_%d", rand.Intn(numFiles))
									cachedFile := fs.GetID(fileID)
									if cachedFile != nil {
										atomic.AddInt64(&cacheHits, 1)
									} else {
										atomic.AddInt64(&cacheMisses, 1)
									}
								}

							case 2: // Lock contention pattern
								file := testFiles[rand.Intn(numFiles)]

								// Try to acquire lock with timeout
								lockAcquired := make(chan bool, 1)
								go func() {
									file.Lock()
									lockAcquired <- true
									time.Sleep(time.Microsecond * time.Duration(rand.Intn(100)))
									file.Unlock()
								}()

								select {
								case <-lockAcquired:
									// Lock acquired successfully
								case <-time.After(time.Millisecond):
									atomic.AddInt64(&lockContentions, 1)
								}

							case 3: // Metadata operations burst
								for j := 0; j < 20; j++ {
									file := testFiles[rand.Intn(numFiles)]
									nodeID := file.NodeID()
									retrievedFile := fs.GetNodeID(nodeID)
									if retrievedFile == nil {
										return fmt.Errorf("node ID lookup failed")
									}
								}

							case 4: // Mixed read/write simulation
								file := testFiles[rand.Intn(numFiles)]

								// Simulate read
								file.RLock()
								_ = file.Name()
								_ = file.Size()
								file.RUnlock()

								// Simulate write
								file.Lock()
								file.hasChanges = true
								file.Unlock()

							case 5: // Filesystem traversal simulation
								for j := 0; j < 5; j++ {
									fileID := fmt.Sprintf("stress_test_file_%d", rand.Intn(numFiles))
									file := fs.GetID(fileID)
									if file != nil {
										_ = file.Path()
									}
								}
							}
							return nil
						}()

						// Record operation latency
						latency := time.Since(startTime).Nanoseconds()
						atomic.AddInt64(&totalLatency, latency)

						// Update min/max latency atomically
						for {
							currentMin := atomic.LoadInt64(&minLatency)
							if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
								break
							}
						}
						for {
							currentMax := atomic.LoadInt64(&maxLatency)
							if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
								break
							}
						}

						// Track operation result
						atomic.AddInt64(&totalOperations, 1)
						if err != nil {
							atomic.AddInt64(&failedOps, 1)
						} else {
							atomic.AddInt64(&successfulOps, 1)
						}
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Analyze stress test results
		total := atomic.LoadInt64(&totalOperations)
		successful := atomic.LoadInt64(&successfulOps)
		_ = atomic.LoadInt64(&failedOps) // Suppress unused variable warning
		hits := atomic.LoadInt64(&cacheHits)
		misses := atomic.LoadInt64(&cacheMisses)
		contentions := atomic.LoadInt64(&lockContentions)

		// Calculate performance metrics
		successRate := float64(successful) / float64(total) * 100
		avgLatency := time.Duration(atomic.LoadInt64(&totalLatency) / total)
		minLat := time.Duration(atomic.LoadInt64(&minLatency))
		maxLat := time.Duration(atomic.LoadInt64(&maxLatency))

		// Verify stress test results
		assert.True(total > 0, "Should have performed operations")
		assert.True(successRate >= 95.0, "Success rate should be at least 95%%, got %.2f%%", successRate)
		assert.True(avgLatency < 10*time.Millisecond, "Average latency should be reasonable, got %v", avgLatency)

		// Log performance metrics for analysis
		t.Logf("Stress Test Results:")
		t.Logf("  Total Operations: %d", total)
		t.Logf("  Success Rate: %.2f%%", successRate)
		t.Logf("  Cache Hit Rate: %.2f%%", float64(hits)/float64(hits+misses)*100)
		t.Logf("  Lock Contentions: %d", contentions)
		t.Logf("  Latency - Min: %v, Max: %v, Avg: %v", minLat, maxLat, avgLatency)

		// Verify filesystem integrity after stress test
		for i := 0; i < 10; i++ { // Check a sample of files
			fileID := fmt.Sprintf("stress_test_file_%d", i)
			file := fs.GetID(fileID)
			assert.NotNil(file, "File should still exist after stress test: %s", fileID)
			if file != nil {
				assert.Equal(fmt.Sprintf("stress_file_%d.txt", i), file.Name(), "File name should be intact")
			}
		}
	})
}

// TestConcurrentDirectoryOperations tests concurrent directory operations
//
//	Test Case ID    CT-FS-05-01
//	Title           Concurrent Directory Operations
//	Description     Tests concurrent directory creation, deletion, and traversal
//	Preconditions   Filesystem is initialized
//	Steps           1. Create concurrent directory operations
//	                2. Verify directory structure integrity
//	                3. Test concurrent file operations within directories
//	Expected Result Directory operations complete successfully without corruption
func TestConcurrentDirectoryOperations(t *testing.T) {
	// Create a test fixture
	fixture := helpers.SetupFSTestFixture(t, "ConcurrentDirectoryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		// Test parameters
		const (
			numGoroutines          = 12
			numDirectories         = 8
			numFilesPerDir         = 5
			operationsPerGoroutine = 25
		)

		// Create test directories
		testDirs := make([]*Inode, numDirectories)
		for i := 0; i < numDirectories; i++ {
			dirItem := &graph.DriveItem{
				ID:     fmt.Sprintf("test_dir_%d", i),
				Name:   fmt.Sprintf("test_directory_%d", i),
				Folder: &graph.Folder{},
				Parent: &graph.DriveItemParent{ID: rootID},
			}
			testDirs[i] = NewInodeDriveItem(dirItem)
			fs.InsertID(dirItem.ID, testDirs[i])
			fs.InsertChild(rootID, testDirs[i])
		}

		// Create files within directories
		testFiles := make(map[string][]*Inode)
		for i, dir := range testDirs {
			files := make([]*Inode, numFilesPerDir)
			for j := 0; j < numFilesPerDir; j++ {
				fileItem := &graph.DriveItem{
					ID:     fmt.Sprintf("dir_%d_file_%d", i, j),
					Name:   fmt.Sprintf("file_%d.txt", j),
					Size:   1024,
					File:   &graph.File{},
					Parent: &graph.DriveItemParent{ID: dir.ID()},
				}
				files[j] = NewInodeDriveItem(fileItem)
				fs.InsertID(fileItem.ID, files[j])
				fs.InsertChild(dir.ID(), files[j])
			}
			testFiles[dir.ID()] = files
		}

		// Counters for tracking operations
		var (
			dirTraversals  int64
			fileAccesses   int64
			childLookups   int64
			pathOperations int64
			successCount   int64
			errorCount     int64
		)

		// Wait group for synchronization
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Launch concurrent directory operation goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					// Select a random directory
					dirIndex := rand.Intn(numDirectories)
					dir := testDirs[dirIndex]

					// Perform random directory operation
					switch rand.Intn(5) {
					case 0: // Directory traversal
						children, err := fs.GetChildrenID(dir.ID(), fs.auth)
						if err == nil && len(children) > 0 {
							atomic.AddInt64(&dirTraversals, 1)
							atomic.AddInt64(&successCount, 1)
						} else {
							atomic.AddInt64(&errorCount, 1)
						}

					case 1: // File access within directory
						if files, exists := testFiles[dir.ID()]; exists && len(files) > 0 {
							file := files[rand.Intn(len(files))]
							_ = file.Name()
							_ = file.Size()
							atomic.AddInt64(&fileAccesses, 1)
							atomic.AddInt64(&successCount, 1)
						}

					case 2: // Child lookup operations
						for k := 0; k < 3; k++ {
							fileName := fmt.Sprintf("file_%d.txt", k)
							child, err := fs.GetChild(dir.ID(), fileName, fs.auth)
							if err == nil && child != nil {
								atomic.AddInt64(&childLookups, 1)
								atomic.AddInt64(&successCount, 1)
							}
						}

					case 3: // Path operations
						dirPath := dir.Path()
						if dirPath != "" {
							// Try to get directory by path
							foundDir, err := fs.GetPath(dirPath, fs.auth)
							if err == nil && foundDir != nil {
								atomic.AddInt64(&pathOperations, 1)
								atomic.AddInt64(&successCount, 1)
							} else {
								atomic.AddInt64(&errorCount, 1)
							}
						}

					case 4: // Mixed operations on directory and its files
						// Access directory metadata
						_ = dir.Name()
						_ = dir.NodeID()

						// Access some files in the directory
						if files, exists := testFiles[dir.ID()]; exists {
							for k := 0; k < minInt(3, len(files)); k++ {
								file := files[k]
								file.RLock()
								_ = file.Name()
								file.RUnlock()
							}
						}
						atomic.AddInt64(&successCount, 1)
					}

					// Small delay to increase chance of race conditions
					time.Sleep(time.Microsecond * time.Duration(rand.Intn(50)))
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Verify results
		totalOperations := int64(numGoroutines * operationsPerGoroutine)
		actualOperations := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&errorCount)

		assert.Equal(totalOperations, actualOperations, "All operations should be accounted for")
		assert.True(atomic.LoadInt64(&successCount) > 0, "Should have successful operations")
		assert.True(atomic.LoadInt64(&dirTraversals) > 0, "Should have directory traversals")
		assert.True(atomic.LoadInt64(&fileAccesses) > 0, "Should have file accesses")

		// Verify directory structure integrity
		for i, dir := range testDirs {
			// Verify directory still exists
			cachedDir := fs.GetID(fmt.Sprintf("test_dir_%d", i))
			assert.NotNil(cachedDir, "Directory should still exist: test_dir_%d", i)

			if cachedDir != nil {
				assert.Equal(fmt.Sprintf("test_directory_%d", i), cachedDir.Name(), "Directory name should be intact")

				// Verify children still exist
				children, err := fs.GetChildrenID(cachedDir.ID(), fs.auth)
				if err == nil {
					assert.Equal(numFilesPerDir, len(children), "Directory should have correct number of children")
				}

				// Verify each file in the directory
				if files, exists := testFiles[dir.ID()]; exists {
					for j, file := range files {
						cachedFile := fs.GetID(fmt.Sprintf("dir_%d_file_%d", i, j))
						assert.NotNil(cachedFile, "File should still exist: dir_%d_file_%d", i, j)
						if cachedFile != nil {
							assert.Equal(file.Name(), cachedFile.Name(), "File name should be intact")
						}
					}
				}
			}
		}

		// Log operation statistics
		t.Logf("Directory Operations Results:")
		t.Logf("  Directory Traversals: %d", atomic.LoadInt64(&dirTraversals))
		t.Logf("  File Accesses: %d", atomic.LoadInt64(&fileAccesses))
		t.Logf("  Child Lookups: %d", atomic.LoadInt64(&childLookups))
		t.Logf("  Path Operations: %d", atomic.LoadInt64(&pathOperations))
		t.Logf("  Success Rate: %.2f%%", float64(atomic.LoadInt64(&successCount))/float64(actualOperations)*100)
	})
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
