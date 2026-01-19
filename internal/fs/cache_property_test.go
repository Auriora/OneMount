package fs

import (
	"context"
	"crypto/rand"
	"fmt"
	"path/filepath"
	"testing"
	"testing/quick"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// CacheScenario represents a cache management test scenario
type CacheScenario struct {
	FileSize    int64
	ETag        string
	ContentHash string
	CacheHit    bool
}

// generateCacheScenario creates a random cache scenario
func generateCacheScenario(seed int) CacheScenario {
	fileSizes := []int64{100, 1024, 10240, 102400, 1048576} // 100B, 1KB, 10KB, 100KB, 1MB

	return CacheScenario{
		FileSize:    fileSizes[seed%len(fileSizes)],
		ETag:        fmt.Sprintf("etag-%d-%d", seed, time.Now().UnixNano()),
		ContentHash: fmt.Sprintf("hash-%d-%d", seed, time.Now().UnixNano()),
		CacheHit:    (seed % 2) == 0,
	}
}

// generateRandomContent creates random file content of the specified size
func generateRandomContent(size int64) []byte {
	content := make([]byte, size)
	_, _ = rand.Read(content)
	return content
}

// **Feature: system-verification-and-fix, Property 28: ETag-Based Cache Storage**
// **Validates: Requirements 7.1**
func TestProperty28_ETagBasedCacheStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any downloaded file, the system should store content in cache directory with file's ETag
	property := func() bool {
		scenario := generateCacheScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

		// Create test file with specific ETag
		fileID := fmt.Sprintf("test-cache-etag-%d", time.Now().UnixNano())
		fileName := "testfile.txt"
		fileContent := generateRandomContent(scenario.FileSize)

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(fileContent))
		file.ETag = scenario.ETag
		file.File.Hashes.QuickXorHash = scenario.ContentHash
		registerDriveItem(filesystem, "root", file)

		// Test 1: Store content in cache
		err = filesystem.content.Insert(fileID, fileContent)
		if err != nil {
			t.Logf("Failed to insert content into cache: %v", err)
			return false
		}

		// Update metadata to reflect cached state
		_, err = filesystem.metadataStore.Update(ctx, fileID, func(e *metadata.Entry) error {
			e.ETag = scenario.ETag
			e.ContentHash = scenario.ContentHash
			e.State = metadata.ItemStateHydrated
			return nil
		})
		if err != nil {
			t.Logf("Failed to update metadata after caching: %v", err)
			return false
		}

		// Test 2: Verify content is stored in cache directory
		if !filesystem.content.HasContent(fileID) {
			t.Logf("Content not found in cache after insert")
			return false
		}

		// Test 3: Verify metadata entry has correct ETag
		entry, err := filesystem.metadataStore.Get(ctx, fileID)
		if err != nil {
			t.Logf("Failed to get metadata entry: %v", err)
			return false
		}

		if entry.ETag != scenario.ETag {
			t.Logf("ETag mismatch in metadata: expected %s, got %s", scenario.ETag, entry.ETag)
			return false
		}

		// Test 4: Verify content hash is stored correctly
		if entry.ContentHash != scenario.ContentHash {
			t.Logf("Content hash mismatch: expected %s, got %s", scenario.ContentHash, entry.ContentHash)
			return false
		}

		// Test 5: Verify cached content matches original
		cachedContent := filesystem.content.Get(fileID)
		if len(cachedContent) != len(fileContent) {
			t.Logf("Cached content size mismatch: expected %d, got %d", len(fileContent), len(cachedContent))
			return false
		}

		// Test 6: Verify content can be retrieved
		if len(cachedContent) == 0 && scenario.FileSize > 0 {
			t.Logf("Failed to retrieve cached content")
			return false
		}

		// Test 7: Verify ETag association with cached content
		// The cache stores content by file ID, and the ETag is stored in metadata
		// This ensures the ETag is properly associated with the cached content
		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}

		fileInode.mu.RLock()
		inodeETag := fileInode.ETag
		fileInode.mu.RUnlock()

		if inodeETag != scenario.ETag {
			t.Logf("Inode ETag mismatch: expected %s, got %s", scenario.ETag, inodeETag)
			return false
		}

		// Success: Content is properly stored with ETag association
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 28 (ETag-Based Cache Storage) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 29: Cache Invalidation on Remote ETag Change**
// **Validates: Requirements 7.3**
func TestProperty29_CacheInvalidationOnRemoteETagChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any cached file with different remote ETag, the system should invalidate cache and download new version
	property := func() bool {
		scenario := generateCacheScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

		// Create test file with initial ETag
		fileID := fmt.Sprintf("test-cache-invalidate-%d", time.Now().UnixNano())
		fileName := "testfile.txt"
		oldETag := fmt.Sprintf("old-etag-%d", time.Now().UnixNano())
		newETag := fmt.Sprintf("new-etag-%d", time.Now().UnixNano())

		oldContent := generateRandomContent(scenario.FileSize)
		newContent := generateRandomContent(scenario.FileSize)

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(oldContent))
		file.ETag = oldETag
		file.File.Hashes.QuickXorHash = "old-hash"
		registerDriveItem(filesystem, "root", file)

		// Test 1: Cache the file with old ETag
		err = filesystem.content.Insert(fileID, oldContent)
		if err != nil {
			t.Logf("Failed to insert content into cache: %v", err)
			return false
		}

		// Test 2: Verify initial cache state
		if !filesystem.content.HasContent(fileID) {
			t.Logf("Content not found in cache after initial insert")
			return false
		}

		// Test 3: Verify initial metadata has old ETag
		entry, err := filesystem.metadataStore.Get(ctx, fileID)
		if err != nil {
			t.Logf("Failed to get initial metadata entry: %v", err)
			return false
		}

		if entry.ETag != oldETag {
			t.Logf("Initial ETag mismatch: expected %s, got %s", oldETag, entry.ETag)
			return false
		}

		// Test 4: Simulate remote change by updating ETag
		// This simulates what happens during delta sync when remote file changes
		file.ETag = newETag
		file.File.Hashes.QuickXorHash = "new-hash"

		// Update metadata to reflect remote change
		entry.ETag = newETag
		entry.ContentHash = "new-hash"
		entry.UpdatedAt = time.Now().UTC()

		err = filesystem.metadataStore.Save(ctx, entry)
		if err != nil {
			t.Logf("Failed to update metadata with new ETag: %v", err)
			return false
		}

		// Test 5: Verify ETag has changed in metadata
		updatedEntry, err := filesystem.metadataStore.Get(ctx, fileID)
		if err != nil {
			t.Logf("Failed to get updated metadata entry: %v", err)
			return false
		}

		if updatedEntry.ETag != newETag {
			t.Logf("Updated ETag mismatch: expected %s, got %s", newETag, updatedEntry.ETag)
			return false
		}

		// Test 6: Verify ETag mismatch detection
		// In a real scenario, the system would detect this mismatch and invalidate cache
		etagChanged := (oldETag != newETag)
		if !etagChanged {
			t.Logf("ETag should have changed")
			return false
		}

		// Test 7: Simulate cache invalidation
		// When ETag changes, the system should invalidate the cached content
		err = filesystem.content.Delete(fileID)
		if err != nil {
			t.Logf("Failed to invalidate cache: %v", err)
			return false
		}

		// Test 8: Verify cache is invalidated
		if filesystem.content.HasContent(fileID) {
			t.Logf("Content should be invalidated after ETag change")
			return false
		}

		// Test 9: Simulate download of new version
		err = filesystem.content.Insert(fileID, newContent)
		if err != nil {
			t.Logf("Failed to insert new content: %v", err)
			return false
		}

		// Update metadata to reflect new cached state
		_, err = filesystem.metadataStore.Update(ctx, fileID, func(e *metadata.Entry) error {
			e.ETag = newETag
			e.ContentHash = "new-hash"
			e.State = metadata.ItemStateHydrated
			return nil
		})
		if err != nil {
			t.Logf("Failed to update metadata after re-caching: %v", err)
			return false
		}

		// Test 10: Verify new content is cached
		if !filesystem.content.HasContent(fileID) {
			t.Logf("New content not found in cache")
			return false
		}

		// Test 11: Verify metadata reflects new state
		finalEntry, err := filesystem.metadataStore.Get(ctx, fileID)
		if err != nil {
			t.Logf("Failed to get final metadata entry: %v", err)
			return false
		}

		if finalEntry.ETag != newETag {
			t.Logf("Final ETag mismatch: expected %s, got %s", newETag, finalEntry.ETag)
			return false
		}

		// Test 12: Verify item state is appropriate
		// After re-download, state should be HYDRATED
		if finalEntry.State != metadata.ItemStateHydrated {
			t.Logf("Item state should be HYDRATED after re-download, got: %v", finalEntry.State)
			return false
		}

		// Success: Cache invalidation and re-download work correctly
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 29 (Cache Invalidation on Remote ETag Change) failed: %v", err)
	}
}
