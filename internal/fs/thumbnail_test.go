package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bcherrington/onedriver/internal/fs/graph"
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

// TestThumbnailCacheCleanup tests the cleanup functionality of the thumbnail cache
func TestThumbnailCacheCleanup(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name           string
		setupFunc      func(t *testing.T, cache *ThumbnailCache) (string, string, []byte)
		expirationTime int
		expectedCount  int
		description    string
	}{
		{
			name: "ExpiredThumbnail_ShouldBeRemoved",
			setupFunc: func(t *testing.T, cache *ThumbnailCache) (string, string, []byte) {
				id := "test-id-expired"
				size := "small"
				content := []byte("test thumbnail content for expired item")
				err := cache.Insert(id, size, content)
				assert.NoError(t, err, "Failed to insert thumbnail")

				// Set the last cleanup time to a long time ago to force cleanup
				cache.lastCleanup = cache.lastCleanup.AddDate(-1, 0, 0)

				return id, size, content
			},
			expirationTime: 0, // Immediate expiration
			expectedCount:  1, // One thumbnail should be removed
			description:    "Tests that expired thumbnails are removed during cleanup",
		},
		{
			name: "NonExpiredThumbnail_ShouldRemain",
			setupFunc: func(t *testing.T, cache *ThumbnailCache) (string, string, []byte) {
				id := "test-id-not-expired"
				size := "medium"
				content := []byte("test thumbnail content for non-expired item")
				err := cache.Insert(id, size, content)
				assert.NoError(t, err, "Failed to insert thumbnail")

				// Set the last cleanup time to a long time ago to force cleanup
				cache.lastCleanup = cache.lastCleanup.AddDate(-1, 0, 0)

				return id, size, content
			},
			expirationTime: 3600, // 1 hour expiration
			expectedCount:  0,    // No thumbnails should be removed
			description:    "Tests that non-expired thumbnails are not removed during cleanup",
		},
		{
			name: "MultipleThumbnails_ShouldRemoveOnlyExpired",
			setupFunc: func(t *testing.T, cache *ThumbnailCache) (string, string, []byte) {
				// Insert first thumbnail (will be expired)
				id1 := "test-id-multi-expired"
				size1 := "small"
				content1 := []byte("test thumbnail content for expired item in multi-test")
				err := cache.Insert(id1, size1, content1)
				assert.NoError(t, err, "Failed to insert first thumbnail")

				// Insert second thumbnail (will not be expired)
				id2 := "test-id-multi-not-expired"
				size2 := "large"
				content2 := []byte("test thumbnail content for non-expired item in multi-test")
				err = cache.Insert(id2, size2, content2)
				assert.NoError(t, err, "Failed to insert second thumbnail")

				// Set the last cleanup time to a long time ago to force cleanup
				cache.lastCleanup = cache.lastCleanup.AddDate(-1, 0, 0)

				// Set the modification time of the first thumbnail to a long time ago
				thumbnailPath := cache.thumbnailPath(id1, size1)
				oldTime := time.Now().AddDate(0, 0, -2) // 2 days ago
				err = os.Chtimes(thumbnailPath, oldTime, oldTime)
				assert.NoError(t, err, "Failed to set old modification time")

				return id1, size1, content1
			},
			expirationTime: 86400, // 1 day expiration
			expectedCount:  1,     // One thumbnail should be removed
			description:    "Tests that only expired thumbnails are removed during cleanup with multiple thumbnails",
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
			id, size, _ := tc.setupFunc(t, cache)

			// Verify the thumbnail exists before cleanup
			assert.True(t, cache.HasThumbnail(id, size),
				"Thumbnail should exist before cleanup")

			// Run cleanup with the specified expiration time
			count, err := cache.CleanupCache(tc.expirationTime)
			assert.NoError(t, err, "Cleanup should not return an error")
			assert.Equal(t, tc.expectedCount, count,
				"Number of removed thumbnails should match expected count")

			// Verify the thumbnail exists or not based on expected count
			if tc.expectedCount > 0 {
				assert.False(t, cache.HasThumbnail(id, size),
					"Thumbnail should not exist after cleanup")
			} else {
				assert.True(t, cache.HasThumbnail(id, size),
					"Thumbnail should still exist after cleanup")
			}
		})
	}
}

// TestThumbnailOperations tests various operations on thumbnails in the filesystem
func TestThumbnailOperations(t *testing.T) {
	// Skip this test if we're not running with a valid auth token
	auth, err := graph.Authenticate(context.Background(), graph.AuthConfig{}, testutil.AuthTokensPath, false)
	if err != nil {
		t.Skip("Skipping test because no valid auth token is available")
	}

	t.Parallel()

	// Define test cases
	testCases := []struct {
		name        string
		size        string
		description string
		testFunc    func(t *testing.T, fs *Filesystem, imagePath string, imageID string, size string)
	}{
		{
			name:        "GetThumbnail_ShouldReturnAndCacheThumbnail",
			size:        "small",
			description: "Tests that getting a thumbnail returns the thumbnail and caches it",
			testFunc: func(t *testing.T, fs *Filesystem, imagePath string, imageID string, size string) {
				// Test getting a thumbnail
				thumbnail, err := fs.GetThumbnail(imagePath, size)
				assert.NoError(t, err, "Failed to get thumbnail")
				assert.NotEmpty(t, thumbnail, "Thumbnail should not be empty")

				// Check that the thumbnail was cached
				assert.True(t, fs.thumbnails.HasThumbnail(imageID, size),
					"Thumbnail should be cached after getting it")
			},
		},
		{
			name:        "DeleteThumbnail_ShouldRemoveThumbnail",
			size:        "medium",
			description: "Tests that deleting a thumbnail removes it from the cache",
			testFunc: func(t *testing.T, fs *Filesystem, imagePath string, imageID string, size string) {
				// First get the thumbnail to ensure it's cached
				thumbnail, err := fs.GetThumbnail(imagePath, size)
				assert.NoError(t, err, "Failed to get thumbnail")
				assert.NotEmpty(t, thumbnail, "Thumbnail should not be empty")

				// Check that the thumbnail was cached
				assert.True(t, fs.thumbnails.HasThumbnail(imageID, size),
					"Thumbnail should be cached after getting it")

				// Delete the thumbnail
				err = fs.DeleteThumbnail(imagePath, size)
				assert.NoError(t, err, "Failed to delete thumbnail")

				// Check that the thumbnail no longer exists
				assert.False(t, fs.thumbnails.HasThumbnail(imageID, size),
					"Thumbnail should not exist after deletion")
			},
		},
		{
			name:        "GetMultipleSizes_ShouldCacheEachSize",
			size:        "large",
			description: "Tests that getting thumbnails of different sizes caches each size separately",
			testFunc: func(t *testing.T, fs *Filesystem, imagePath string, imageID string, size string) {
				// Get thumbnails of different sizes
				sizes := []string{"small", "medium", size}
				for _, s := range sizes {
					thumbnail, err := fs.GetThumbnail(imagePath, s)
					assert.NoError(t, err, "Failed to get thumbnail of size %s", s)
					assert.NotEmpty(t, thumbnail, "Thumbnail of size %s should not be empty", s)

					// Check that the thumbnail was cached
					assert.True(t, fs.thumbnails.HasThumbnail(imageID, s),
						"Thumbnail of size %s should be cached after getting it", s)
				}

				// Delete all thumbnails
				for _, s := range sizes {
					err := fs.DeleteThumbnail(imagePath, s)
					assert.NoError(t, err, "Failed to delete thumbnail of size %s", s)

					// Check that the thumbnail no longer exists
					assert.False(t, fs.thumbnails.HasThumbnail(imageID, s),
						"Thumbnail of size %s should not exist after deletion", s)
				}
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for the filesystem
			tempDir, err := os.MkdirTemp("", "onedriver-thumbnail-test-*")
			assert.NoError(t, err, "Failed to create temporary directory")

			// Setup cleanup to remove the temp directory after test completes or fails
			t.Cleanup(func() {
				if err := os.RemoveAll(tempDir); err != nil {
					t.Logf("Warning: Failed to clean up temp directory %s: %v", tempDir, err)
				}
			})

			// Create a filesystem
			fs, err := NewFilesystem(auth, tempDir, 30)
			assert.NoError(t, err, "Failed to create filesystem")

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

			// Run the test
			tc.testFunc(t, fs, imagePath, imageID, tc.size)
		})
	}
}
