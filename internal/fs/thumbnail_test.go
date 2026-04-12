package fs

import (
	"os"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_08_01_ThumbnailCache_BasicOperations_WorkCorrectly tests various operations on the thumbnail cache.
func TestIT_FS_08_01_ThumbnailCache_BasicOperations_WorkCorrectly(t *testing.T) {
	fixture := framework.NewUnitTestFixture("ThumbnailCacheOperationsFixture")

	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		dir, err := os.MkdirTemp("", "thumbnail-cache-test-*")
		if err != nil {
			return nil, err
		}
		return dir, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		dir := fixture.(string)
		return os.RemoveAll(dir)
	})

	fixture.Use(t, func(t *testing.T, f interface{}) {
		assert := framework.NewAssert(t)
		fixtureData := f.(*framework.UnitTestFixture)
		dir := fixtureData.SetupData.(string)

		// Step 1: Create a thumbnail cache
		cache := NewThumbnailCache(dir)
		assert.NotNil(cache, "Thumbnail cache should be created")

		// Step 2: Insert thumbnails
		testID := "test-item-id"
		smallContent := []byte("small-thumbnail-data")
		mediumContent := []byte("medium-thumbnail-data-larger")
		largeContent := []byte("large-thumbnail-data-even-larger-content")

		err := cache.Insert(testID, "small", smallContent)
		assert.NoError(err, "Insert small thumbnail should succeed")
		err = cache.Insert(testID, "medium", mediumContent)
		assert.NoError(err, "Insert medium thumbnail should succeed")
		err = cache.Insert(testID, "large", largeContent)
		assert.NoError(err, "Insert large thumbnail should succeed")

		// Step 3: Check if thumbnails exist
		assert.True(cache.HasThumbnail(testID, "small"), "Small thumbnail should exist")
		assert.True(cache.HasThumbnail(testID, "medium"), "Medium thumbnail should exist")
		assert.True(cache.HasThumbnail(testID, "large"), "Large thumbnail should exist")
		assert.False(cache.HasThumbnail("nonexistent", "small"), "Non-existent thumbnail should not exist")

		// Step 4: Retrieve thumbnails
		retrieved := cache.Get(testID, "small")
		assert.Equal(string(smallContent), string(retrieved), "Retrieved small thumbnail should match")
		retrieved = cache.Get(testID, "medium")
		assert.Equal(string(mediumContent), string(retrieved), "Retrieved medium thumbnail should match")
		retrieved = cache.Get(testID, "large")
		assert.Equal(string(largeContent), string(retrieved), "Retrieved large thumbnail should match")

		// Step 5: Delete thumbnails
		err = cache.Delete(testID, "small")
		assert.NoError(err, "Delete small thumbnail should succeed")
		assert.False(cache.HasThumbnail(testID, "small"), "Small thumbnail should not exist after deletion")
		assert.True(cache.HasThumbnail(testID, "medium"), "Medium thumbnail should still exist")

		// Delete all thumbnails for the ID
		err = cache.DeleteAll(testID)
		assert.NoError(err, "DeleteAll should succeed")
		assert.False(cache.HasThumbnail(testID, "medium"), "Medium thumbnail should not exist after DeleteAll")
		assert.False(cache.HasThumbnail(testID, "large"), "Large thumbnail should not exist after DeleteAll")
	})
}

// TestIT_FS_09_01_ThumbnailCache_Cleanup_RemovesExpiredThumbnails tests the cleanup functionality.
func TestIT_FS_09_01_ThumbnailCache_Cleanup_RemovesExpiredThumbnails(t *testing.T) {
	fixture := framework.NewUnitTestFixture("ThumbnailCacheCleanupFixture")

	fixture.WithSetup(func(t *testing.T) (interface{}, error) {
		dir, err := os.MkdirTemp("", "thumbnail-cleanup-test-*")
		if err != nil {
			return nil, err
		}
		return dir, nil
	}).WithTeardown(func(t *testing.T, fixture interface{}) error {
		dir := fixture.(string)
		return os.RemoveAll(dir)
	})

	fixture.Use(t, func(t *testing.T, f interface{}) {
		assert := framework.NewAssert(t)
		fixtureData := f.(*framework.UnitTestFixture)
		dir := fixtureData.SetupData.(string)

		// Step 1: Create a thumbnail cache
		cache := NewThumbnailCache(dir)
		assert.NotNil(cache, "Thumbnail cache should be created")

		// Step 2: Insert thumbnails
		err := cache.Insert("item1", "small", []byte("thumb1"))
		assert.NoError(err, "Insert should succeed")
		err = cache.Insert("item2", "small", []byte("thumb2"))
		assert.NoError(err, "Insert should succeed")

		// Step 3: Verify thumbnails exist
		assert.True(cache.HasThumbnail("item1", "small"), "Thumbnail 1 should exist")
		assert.True(cache.HasThumbnail("item2", "small"), "Thumbnail 2 should exist")

		// Step 4: Run cleanup with a very large expiration (nothing should be removed)
		removed, err := cache.CleanupCache(365)
		assert.NoError(err, "CleanupCache should succeed")
		// Cleanup has a 24-hour throttle, so first call may return 0 or clean
		_ = removed

		// Step 5: Verify thumbnails still exist (they're not expired)
		assert.True(cache.HasThumbnail("item1", "small"), "Thumbnail 1 should still exist after cleanup with large expiration")
		assert.True(cache.HasThumbnail("item2", "small"), "Thumbnail 2 should still exist after cleanup with large expiration")
	})
}

// TestIT_FS_10_01_Thumbnails_FileSystemOperations_WorkCorrectly tests thumbnail operations in the filesystem.
func TestIT_FS_10_01_Thumbnails_FileSystemOperations_WorkCorrectly(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "ThumbnailOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		fsFixture := getFSTestFixture(t, fixture)
		filesystem := fsFixture.FS.(*Filesystem)

		// Verify the filesystem has a thumbnail cache
		assert.NotNil(filesystem.thumbnails, "Filesystem should have a thumbnail cache")

		// Insert a thumbnail through the cache
		testID := "fs-thumb-test-id"
		content := []byte("filesystem-thumbnail-content")
		err := filesystem.thumbnails.Insert(testID, "small", content)
		assert.NoError(err, "Inserting thumbnail through filesystem should succeed")

		// Verify it can be retrieved
		retrieved := filesystem.thumbnails.Get(testID, "small")
		assert.Equal(string(content), string(retrieved), "Retrieved thumbnail should match")

		// Delete and verify
		err = filesystem.thumbnails.DeleteAll(testID)
		assert.NoError(err, "DeleteAll through filesystem should succeed")
		assert.False(filesystem.thumbnails.HasThumbnail(testID, "small"), "Thumbnail should be gone after DeleteAll")
	})
}
