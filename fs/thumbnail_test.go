package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jstaf/onedriver/fs/graph"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestThumbnailCache(t *testing.T) {
	// Create a temporary directory for the thumbnail cache
	tempDir, err := os.MkdirTemp("", "onedriver-thumbnail-test-*")
	assert.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create a thumbnail cache
	cache := NewThumbnailCache(tempDir)
	assert.NotNil(t, cache)

	// Test inserting and retrieving a thumbnail
	id := "test-id"
	size := "small"
	content := []byte("test thumbnail content")
	err = cache.Insert(id, size, content)
	assert.NoError(t, err)

	// Check if the thumbnail exists
	assert.True(t, cache.HasThumbnail(id, size))

	// Retrieve the thumbnail
	retrieved := cache.Get(id, size)
	assert.Equal(t, content, retrieved)

	// Delete the thumbnail
	err = cache.Delete(id, size)
	assert.NoError(t, err)

	// Check that the thumbnail no longer exists
	assert.False(t, cache.HasThumbnail(id, size))
}

func TestThumbnailCacheMultipleSizes(t *testing.T) {
	// Create a temporary directory for the thumbnail cache
	tempDir, err := os.MkdirTemp("", "onedriver-thumbnail-test-*")
	assert.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create a thumbnail cache
	cache := NewThumbnailCache(tempDir)
	assert.NotNil(t, cache)

	// Test inserting and retrieving thumbnails of different sizes
	id := "test-id"
	sizes := []string{"small", "medium", "large"}
	contents := map[string][]byte{
		"small":  []byte("small thumbnail content"),
		"medium": []byte("medium thumbnail content"),
		"large":  []byte("large thumbnail content"),
	}

	// Insert thumbnails of different sizes
	for _, size := range sizes {
		err = cache.Insert(id, size, contents[size])
		assert.NoError(t, err)
	}

	// Check if the thumbnails exist
	for _, size := range sizes {
		assert.True(t, cache.HasThumbnail(id, size))
	}

	// Retrieve the thumbnails
	for _, size := range sizes {
		retrieved := cache.Get(id, size)
		assert.Equal(t, contents[size], retrieved)
	}

	// Delete all thumbnails
	err = cache.DeleteAll(id)
	assert.NoError(t, err)

	// Check that the thumbnails no longer exist
	for _, size := range sizes {
		assert.False(t, cache.HasThumbnail(id, size))
	}
}

func TestThumbnailCacheCleanup(t *testing.T) {
	// Create a temporary directory for the thumbnail cache
	tempDir, err := os.MkdirTemp("", "onedriver-thumbnail-test-*")
	assert.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create a thumbnail cache
	cache := NewThumbnailCache(tempDir)
	assert.NotNil(t, cache)

	// Test inserting thumbnails
	id := "test-id"
	size := "small"
	content := []byte("test thumbnail content")
	err = cache.Insert(id, size, content)
	assert.NoError(t, err)

	// Check if the thumbnail exists
	assert.True(t, cache.HasThumbnail(id, size))

	// Set the last cleanup time to a long time ago to force cleanup
	cache.lastCleanup = cache.lastCleanup.AddDate(-1, 0, 0)

	// Run cleanup with a short expiration time
	count, err := cache.CleanupCache(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Check that the thumbnail no longer exists
	assert.False(t, cache.HasThumbnail(id, size))
}

func TestThumbnailOperations(t *testing.T) {
	// Skip this test if we're not running with a valid auth token
	auth, err := graph.Authenticate(context.Background(), graph.AuthConfig{}, ".auth_tokens.json", false)
	if err != nil {
		t.Skip("Skipping test because no valid auth token is available")
	}

	// Create a temporary directory for the filesystem
	tempDir, err := os.MkdirTemp("", "onedriver-thumbnail-test-*")
	assert.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temporary directory: %v", err)
		}
	}()

	// Create a filesystem
	fs, err := NewFilesystem(auth, tempDir, 30)
	assert.NoError(t, err)
	// No Close method on Filesystem

	// Find an image file to test with
	var imagePath string
	var imageID string
	if err := fs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMetadata)
		return b.ForEach(func(k, v []byte) error {
			var item graph.DriveItem
			if err := json.Unmarshal(v, &item); err != nil {
				return nil
			}
			if item.File != nil && filepath.Ext(item.Name) == ".jpg" {
				// Create an Inode from the DriveItem to get the path
				inode := NewInodeDriveItem(&item)
				imagePath = inode.Path()
				imageID = item.ID
				return nil
			}
			return nil
		})
	}); err != nil {
		t.Fatalf("Failed to search for image files: %v", err)
	}

	if imagePath == "" {
		t.Skip("Skipping test because no image file was found")
	}

	// Test getting a thumbnail
	thumbnail, err := fs.GetThumbnail(imagePath, "small")
	assert.NoError(t, err)
	assert.NotEmpty(t, thumbnail)

	// Check that the thumbnail was cached
	assert.True(t, fs.thumbnails.HasThumbnail(imageID, "small"))

	// Delete the thumbnail
	err = fs.DeleteThumbnail(imagePath, "small")
	assert.NoError(t, err)

	// Check that the thumbnail no longer exists
	assert.False(t, fs.thumbnails.HasThumbnail(imageID, "small"))
}
