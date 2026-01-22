package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	bolt "go.etcd.io/bbolt"
)

// TestDocumentPerformanceCharacteristics documents current performance characteristics
// Task 44.1: Profile status determination performance
func TestDocumentPerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance documentation test in short mode")
	}

	fixture := helpers.SetupFSTestFixture(t, "PerfCharacteristicsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Get the filesystem from the fixture
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Initialize status cache
		fs.statusCache = newStatusCache(5 * time.Second)

		results := make(map[string]map[string]interface{})

		// Test with different file counts
		fileCounts := []int{100, 1000, 10000}

		for _, count := range fileCounts {
			t.Run(fmt.Sprintf("files_%d", count), func(t *testing.T) {
				ids := setupTestFilesForProfiling(t, fs, count)

				// Measure without cache
				start := time.Now()
				for _, id := range ids {
					_ = fs.determineFileStatus(id)
				}
				noCacheDuration := time.Since(start)

				// Measure with cache (second pass)
				start = time.Now()
				for _, id := range ids {
					_ = fs.GetFileStatus(id)
				}
				withCacheDuration := time.Since(start)

				// Measure batch operation
				start = time.Now()
				_ = fs.GetFileStatusBatch(ids)
				batchDuration := time.Since(start)

				// Store results
				results[fmt.Sprintf("%d_files", count)] = map[string]interface{}{
					"no_cache_total":      noCacheDuration,
					"no_cache_per_file":   noCacheDuration / time.Duration(count),
					"with_cache_total":    withCacheDuration,
					"with_cache_per_file": withCacheDuration / time.Duration(count),
					"batch_total":         batchDuration,
					"batch_per_file":      batchDuration / time.Duration(count),
					"cache_speedup":       float64(noCacheDuration) / float64(withCacheDuration),
					"batch_speedup":       float64(noCacheDuration) / float64(batchDuration),
				}

				t.Logf("Performance characteristics for %d files:", count)
				t.Logf("  No cache: %v total, %v per file", noCacheDuration, noCacheDuration/time.Duration(count))
				t.Logf("  With cache: %v total, %v per file", withCacheDuration, withCacheDuration/time.Duration(count))
				t.Logf("  Batch: %v total, %v per file", batchDuration, batchDuration/time.Duration(count))
				t.Logf("  Cache speedup: %.2fx", float64(noCacheDuration)/float64(withCacheDuration))
				t.Logf("  Batch speedup: %.2fx", float64(noCacheDuration)/float64(batchDuration))
			})
		}

		// Save results to file
		resultsPath := filepath.Join(os.TempDir(), "status_determination_performance.txt")
		f, err := os.Create(resultsPath)
		if err != nil {
			t.Fatalf("Failed to create results file: %v", err)
		}
		defer f.Close()

		fmt.Fprintf(f, "Status Determination Performance Characteristics\n")
		fmt.Fprintf(f, "==============================================\n\n")

		for key, metrics := range results {
			fmt.Fprintf(f, "%s:\n", key)
			for metric, value := range metrics {
				fmt.Fprintf(f, "  %s: %v\n", metric, value)
			}
			fmt.Fprintf(f, "\n")
		}

		t.Logf("Performance characteristics saved to: %s", resultsPath)
	})
}

// TestIdentifyBottlenecks runs targeted tests to identify specific bottlenecks
// Task 44.1: Identify bottlenecks (database queries, hash calculation)
func TestIdentifyBottlenecks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bottleneck identification test in short mode")
	}

	fixture := helpers.SetupFSTestFixture(t, "BottleneckFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Get the filesystem from the fixture
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Test 1: Database query performance
		t.Run("database_queries", func(t *testing.T) {
			count := 1000
			ids := setupTestFilesForProfiling(t, fs, count)

			// Add some offline changes
			err := fs.db.Update(func(tx *bolt.Tx) error {
				b, err := tx.CreateBucketIfNotExists(bucketOfflineChanges)
				if err != nil {
					return err
				}

				for i := 0; i < count/10; i++ {
					key := fmt.Sprintf("%s-%d", ids[i], time.Now().Unix())
					if err := b.Put([]byte(key), []byte("change")); err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				t.Fatalf("Failed to setup offline changes: %v", err)
			}

			// Measure database query time
			start := time.Now()
			for _, id := range ids {
				_ = fs.determineFileStatus(id)
			}
			duration := time.Since(start)

			t.Logf("Database query performance:")
			t.Logf("  Total time: %v", duration)
			t.Logf("  Average per query: %v", duration/time.Duration(count))
			t.Logf("  Queries per second: %.2f", float64(count)/duration.Seconds())
		})

		// Test 2: Upload session check performance
		t.Run("upload_session_checks", func(t *testing.T) {
			count := 1000
			ids := setupTestFilesForProfiling(t, fs, count)

			// Create some upload sessions
			if fs.uploads == nil {
				fs.uploads = NewUploadManager(5*time.Second, fs.db, fs, fs.auth)
			}
			for i := 0; i < count/10; i++ {
				session := &UploadSession{
					ID:    ids[i],
					state: uploadStarted,
				}
				fs.uploads.sessions[ids[i]] = session
			}

			// Measure upload session check time
			start := time.Now()
			for _, id := range ids {
				_ = fs.determineFileStatus(id)
			}
			duration := time.Since(start)

			t.Logf("Upload session check performance:")
			t.Logf("  Total time: %v", duration)
			t.Logf("  Average per check: %v", duration/time.Duration(count))
			t.Logf("  Checks per second: %.2f", float64(count)/duration.Seconds())
		})
	})
}

// TestProfileMemoryUsage profiles memory usage during status determination
// Task 44.1: Measure performance with various file counts
func TestProfileMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping profiling test in short mode")
	}

	fixture := helpers.SetupFSTestFixture(t, "ProfileMemoryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Get the filesystem from the fixture
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Initialize status cache
		fs.statusCache = newStatusCache(5 * time.Second)

		// Create test files
		count := 10000 // Reduced from 100000 for faster testing
		ids := setupTestFilesForProfiling(t, fs, count)

		// Force GC and get baseline memory
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Populate cache
		for _, id := range ids {
			_ = fs.GetFileStatus(id)
		}

		// Force GC and get final memory
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Calculate memory usage
		allocDiff := m2.Alloc - m1.Alloc
		totalAllocDiff := m2.TotalAlloc - m1.TotalAlloc

		t.Logf("Memory usage for %d cached statuses:", count)
		t.Logf("  Allocated: %d bytes (%.2f MB)", allocDiff, float64(allocDiff)/(1024*1024))
		t.Logf("  Total allocated: %d bytes (%.2f MB)", totalAllocDiff, float64(totalAllocDiff)/(1024*1024))
		t.Logf("  Per entry: %.2f bytes", float64(allocDiff)/float64(count))

		// Memory heap profile
		profilePath := filepath.Join(os.TempDir(), "mem_profile_status_cache.prof")
		f, err := os.Create(profilePath)
		if err != nil {
			t.Fatalf("Failed to create memory profile: %v", err)
		}
		defer f.Close()

		if err := pprof.WriteHeapProfile(f); err != nil {
			t.Fatalf("Failed to write memory profile: %v", err)
		}

		t.Logf("Memory profile saved to: %s", profilePath)
	})
}

// setupTestFilesForProfiling creates test files for profiling
func setupTestFilesForProfiling(t *testing.T, fs *Filesystem, count int) []string {
	t.Helper()

	ids := make([]string, count)

	// Create test files in database
	err := fs.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketMetadata)
		if err != nil {
			return err
		}

		for i := 0; i < count; i++ {
			id := fmt.Sprintf("test-file-%d", i)
			ids[i] = id

			// Store metadata in database
			data := []byte(fmt.Sprintf(`{"id":"%s","name":"file%d.txt"}`, id, i))
			if err := b.Put([]byte(id), data); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to setup test files: %v", err)
	}

	return ids
}
