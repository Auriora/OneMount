package fs

import (
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/stretchr/testify/assert"
)

// createMockAuth creates a mock authentication object for testing
func createMockAuthForStatusTests() *graph.Auth {
	return &graph.Auth{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}
}

// TestUT_FS_FileStatus_01_CachePerformance tests the performance of status determination with caching
func TestUT_FS_FileStatus_01_CachePerformance(t *testing.T) {
	t.Run("Status cache basic operations", func(t *testing.T) {
		// Create a status cache with 5 second TTL
		cache := newStatusCache(5 * time.Second)

		// Test set and get
		testID := "test-file-id"
		testStatus := FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		}

		// Set status in cache
		cache.set(testID, testStatus)

		// Get status from cache
		cachedStatus, found := cache.get(testID)
		assert.True(t, found, "Status should be found in cache")
		assert.Equal(t, testStatus.Status, cachedStatus.Status)

		// Test cache miss
		_, found = cache.get("non-existent-id")
		assert.False(t, found, "Non-existent ID should not be found in cache")
	})

	t.Run("Cache invalidation", func(t *testing.T) {
		// Create a status cache
		cache := newStatusCache(5 * time.Second)

		// Set a status
		testID := "test-file-id"
		testStatus := FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		}
		cache.set(testID, testStatus)

		// Verify it's cached
		_, found := cache.get(testID)
		assert.True(t, found, "Status should be in cache")

		// Invalidate the cache entry
		cache.invalidate(testID)

		// Verify it's no longer cached
		_, found = cache.get(testID)
		assert.False(t, found, "Status should not be in cache after invalidation")
	})

	t.Run("Cache TTL expiration", func(t *testing.T) {
		// Create cache with very short TTL
		cache := newStatusCache(100 * time.Millisecond)

		// Set a status
		testID := "test-file-id"
		testStatus := FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Now(),
		}
		cache.set(testID, testStatus)

		// Verify it's cached
		_, found := cache.get(testID)
		assert.True(t, found, "Status should be in cache")

		// Wait for TTL to expire
		time.Sleep(150 * time.Millisecond)

		// Verify it's no longer cached (expired)
		_, found = cache.get(testID)
		assert.False(t, found, "Status should not be in cache after TTL expiration")
	})
}

// TestUT_FS_FileStatus_02_CacheInvalidation tests cache invalidation scenarios
func TestUT_FS_FileStatus_02_CacheInvalidation(t *testing.T) {
	t.Run("Invalidate all clears entire cache", func(t *testing.T) {
		// Create cache
		cache := newStatusCache(5 * time.Second)

		// Add multiple entries
		cache.set("file1", FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()})
		cache.set("file2", FileStatusInfo{Status: StatusCloud, Timestamp: time.Now()})
		cache.set("file3", FileStatusInfo{Status: StatusDownloading, Timestamp: time.Now()})

		// Verify all are cached
		_, found := cache.get("file1")
		assert.True(t, found)
		_, found = cache.get("file2")
		assert.True(t, found)
		_, found = cache.get("file3")
		assert.True(t, found)

		// Invalidate all
		cache.invalidateAll()

		// Verify all are removed
		_, found = cache.get("file1")
		assert.False(t, found)
		_, found = cache.get("file2")
		assert.False(t, found)
		_, found = cache.get("file3")
		assert.False(t, found)
	})
}

// TestUT_FS_FileStatus_04_CacheCleanup tests status cache cleanup
func TestUT_FS_FileStatus_04_CacheCleanup(t *testing.T) {
	t.Run("Cleanup removes expired entries", func(t *testing.T) {
		// Create cache with short TTL
		cache := newStatusCache(100 * time.Millisecond)

		// Add some entries
		cache.set("file1", FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()})
		cache.set("file2", FileStatusInfo{Status: StatusCloud, Timestamp: time.Now()})
		cache.set("file3", FileStatusInfo{Status: StatusDownloading, Timestamp: time.Now()})

		// Verify entries exist
		_, exists := cache.get("file1")
		assert.True(t, exists)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Run cleanup
		cache.cleanup()

		// Verify entries are removed
		_, exists = cache.get("file1")
		assert.False(t, exists)
		_, exists = cache.get("file2")
		assert.False(t, exists)
		_, exists = cache.get("file3")
		assert.False(t, exists)
	})

	t.Run("Cleanup preserves non-expired entries", func(t *testing.T) {
		// Create cache with longer TTL
		cache := newStatusCache(1 * time.Second)

		// Add some entries
		cache.set("file1", FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()})
		time.Sleep(100 * time.Millisecond)
		cache.set("file2", FileStatusInfo{Status: StatusCloud, Timestamp: time.Now()})

		// Run cleanup (file1 is older but not expired)
		cache.cleanup()

		// Verify both entries still exist
		_, exists := cache.get("file1")
		assert.True(t, exists)
		_, exists = cache.get("file2")
		assert.True(t, exists)
	})
}
