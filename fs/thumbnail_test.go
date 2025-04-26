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

// TestThumbnailCacheOperations tests various operations on the thumbnail cache
func TestThumbnailCacheOperations(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T, cache *ThumbnailCache) (string, []string, map[string][]byte)
		testFunc    func(t *testing.T, cache *ThumbnailCache, id string, sizes []string, contents map[string][]byte)
		description string
	}{
		{
			name: "SingleThumbnail_BasicOperations",
			setupFunc: func(t *testing.T, cache *ThumbnailCache) (string, []string, map[string][]byte) {
				id := "test-id-single"
				sizes := []string{"small"}
				contents := map[string][]byte{
					"small": []byte("test thumbnail content"),
				}
				return id, sizes, contents
			},
			testFunc: func(t *testing.T, cache *ThumbnailCache, id string, sizes []string, contents map[string][]byte) {
				size := sizes[0]
				content := contents[size]

				// Insert the thumbnail
				err := cache.Insert(id, size, content)
				assert.NoError(t, err, "Failed to insert thumbnail")

				// Check if the thumbnail exists
				assert.True(t, cache.HasThumbnail(id, size), "Thumbnail should exist after insertion")

				// Retrieve the thumbnail
				retrieved := cache.Get(id, size)
				assert.Equal(t, content, retrieved, "Retrieved thumbnail content should match inserted content")

				// Delete the thumbnail
				err = cache.Delete(id, size)
				assert.NoError(t, err, "Failed to delete thumbnail")

				// Check that the thumbnail no longer exists
				assert.False(t, cache.HasThumbnail(id, size), "Thumbnail should not exist after deletion")
			},
			description: "Tests basic operations (insert, has, get, delete) on a single thumbnail",
		},
		{
			name: "MultipleSizes_AllOperations",
			setupFunc: func(t *testing.T, cache *ThumbnailCache) (string, []string, map[string][]byte) {
				id := "test-id-multiple"
				sizes := []string{"small", "medium", "large"}
				contents := map[string][]byte{
					"small":  []byte("small thumbnail content"),
					"medium": []byte("medium thumbnail content"),
					"large":  []byte("large thumbnail content"),
				}
				return id, sizes, contents
			},
			testFunc: func(t *testing.T, cache *ThumbnailCache, id string, sizes []string, contents map[string][]byte) {
				// Insert thumbnails of different sizes
				for _, size := range sizes {
					err := cache.Insert(id, size, contents[size])
					assert.NoError(t, err, "Failed to insert thumbnail of size %s", size)
				}

				// Check if the thumbnails exist
				for _, size := range sizes {
					assert.True(t, cache.HasThumbnail(id, size), 
						"Thumbnail of size %s should exist after insertion", size)
				}

				// Retrieve the thumbnails
				for _, size := range sizes {
					retrieved := cache.Get(id, size)
					assert.Equal(t, contents[size], retrieved, 
						"Retrieved thumbnail content for size %s should match inserted content", size)
				}

				// Delete all thumbnails
				err := cache.DeleteAll(id)
				assert.NoError(t, err, "Failed to delete all thumbnails")

				// Check that the thumbnails no longer exist
				for _, size := range sizes {
					assert.False(t, cache.HasThumbnail(id, size), 
						"Thumbnail of size %s should not exist after deletion", size)
				}
			},
			description: "Tests operations on thumbnails of multiple sizes, including DeleteAll",
		},
		{
			name: "MultipleSizes_IndividualDelete",
			setupFunc: func(t *testing.T, cache *ThumbnailCache) (string, []string, map[string][]byte) {
				id := "test-id-individual-delete"
				sizes := []string{"small", "medium", "large"}
				contents := map[string][]byte{
					"small":  []byte("small thumbnail content"),
					"medium": []byte("medium thumbnail content"),
					"large":  []byte("large thumbnail content"),
				}
				return id, sizes, contents
			},
			testFunc: func(t *testing.T, cache *ThumbnailCache, id string, sizes []string, contents map[string][]byte) {
				// Insert thumbnails of different sizes
				for _, size := range sizes {
					err := cache.Insert(id, size, contents[size])
					assert.NoError(t, err, "Failed to insert thumbnail of size %s", size)
				}

				// Delete thumbnails individually
				for _, size := range sizes {
					err := cache.Delete(id, size)
					assert.NoError(t, err, "Failed to delete thumbnail of size %s", size)

					// Check that this thumbnail no longer exists
					assert.False(t, cache.HasThumbnail(id, size), 
						"Thumbnail of size %s should not exist after deletion", size)

					// Check that other thumbnails still exist (if any)
					for _, otherSize := range sizes {
						if otherSize == size {
							continue
						}
						// If we've already deleted this size in a previous iteration, it shouldn't exist
						alreadyDeleted := false
						for _, deletedSize := range sizes {
							if deletedSize == otherSize && deletedSize < size {
								alreadyDeleted = true
								break
							}
						}

						if alreadyDeleted {
							assert.False(t, cache.HasThumbnail(id, otherSize), 
								"Thumbnail of size %s should not exist after deletion", otherSize)
						} else {
							assert.True(t, cache.HasThumbnail(id, otherSize), 
								"Thumbnail of size %s should still exist", otherSize)
						}
					}
				}
			},
			description: "Tests deleting thumbnails of multiple sizes individually",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for the thumbnail cache
			tempDir, err := os.MkdirTemp("", "onedriver-thumbnail-test-*")
			assert.NoError(t, err, "Failed to create temporary directory")

			// Setup cleanup to remove the temp directory after test completes or fails
			t.Cleanup(func() {
				if err := os.RemoveAll(tempDir); err != nil {
					t.Logf("Warning: Failed to clean up temp directory %s: %v", tempDir, err)
				}
			})

			// Create a thumbnail cache
			cache := NewThumbnailCache(tempDir)
			assert.NotNil(t, cache, "Thumbnail cache should not be nil")

			// Setup test data
			id, sizes, contents := tc.setupFunc(t, cache)

			// Run the test
			tc.testFunc(t, cache, id, sizes, contents)
		})
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
