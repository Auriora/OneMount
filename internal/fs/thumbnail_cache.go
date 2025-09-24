package fs

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/logging"
)

// ThumbnailCache stores thumbnails for files under a folder as regular files
type ThumbnailCache struct {
	directory string
	fds       sync.Map
	// lastCleanup tracks when the last cache cleanup was performed
	lastCleanup time.Time
}

// NewThumbnailCache creates a new thumbnail cache
func NewThumbnailCache(directory string) *ThumbnailCache {
	if err := os.MkdirAll(directory, 0700); err != nil {
		// Log error but continue - the directory might already exist
		// or we might be able to create files directly
		// This is a best-effort approach
		// Using MkdirAll instead of Mkdir to create parent directories if needed
		logging.Error().Err(err).Str("directory", directory).Msg("Failed to create thumbnail cache directory")
	}
	return &ThumbnailCache{
		directory:   directory,
		fds:         sync.Map{},
		lastCleanup: time.Now(),
	}
}

// thumbnailPath returns the path for the given thumbnail file
// size can be "small", "medium", or "large"
func (t *ThumbnailCache) thumbnailPath(id string, size string) string {
	return filepath.Join(t.directory, id+"-"+size)
}

// Get reads a thumbnail from disk.
func (t *ThumbnailCache) Get(id string, size string) []byte {
	content, err := os.ReadFile(t.thumbnailPath(id, size))
	if err != nil {
		// Return empty content if file doesn't exist or can't be read
		return []byte{}
	}
	return content
}

// Insert writes thumbnail content to disk.
func (t *ThumbnailCache) Insert(id string, size string, content []byte) error {
	return os.WriteFile(t.thumbnailPath(id, size), content, 0600)
}

// InsertStream inserts a stream of thumbnail data
func (t *ThumbnailCache) InsertStream(id string, size string, reader io.Reader) (int64, error) {
	fd, err := t.Open(id, size)
	if err != nil {
		return 0, err
	}

	// Copy the data from the reader to the file
	n, err := io.Copy(fd, reader)
	if err != nil {
		return n, err
	}

	return n, nil
}

// Delete closes the fd AND deletes thumbnail from disk.
func (t *ThumbnailCache) Delete(id string, size string) error {
	// Try to close the file first
	closeErr := t.Close(id, size)

	// Try to remove the file regardless of close error
	removeErr := os.Remove(t.thumbnailPath(id, size))

	// Handle remove error - ignore "file not found" errors
	if removeErr != nil && !os.IsNotExist(removeErr) {
		return removeErr
	}

	// If we got here, the remove succeeded or the file didn't exist
	// Return any close error that might have occurred
	return closeErr
}

// DeleteAll deletes all thumbnails for a given ID
func (t *ThumbnailCache) DeleteAll(id string) error {
	sizes := []string{"small", "medium", "large"}
	var lastErr error
	for _, size := range sizes {
		if err := t.Delete(id, size); err != nil && !os.IsNotExist(err) {
			lastErr = err
		}
	}
	return lastErr
}

// HasThumbnail checks if a thumbnail exists in the cache
func (t *ThumbnailCache) HasThumbnail(id string, size string) bool {
	_, err := os.Stat(t.thumbnailPath(id, size))
	return err == nil
}

// Open opens a thumbnail file for reading/writing
func (t *ThumbnailCache) Open(id string, size string) (*os.File, error) {
	path := t.thumbnailPath(id, size)

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		logging.Error().Err(err).Str("directory", filepath.Dir(path)).Msg("Failed to create parent directory for thumbnail file")
		return nil, err
	}

	// Check if we already have an open file descriptor
	if fd, ok := t.fds.Load(id + "-" + size); ok {
		return fd.(*os.File), nil
	}

	// Open the file with read/write access, create if it doesn't exist
	fd, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	// Store the file descriptor
	t.fds.Store(id+"-"+size, fd)
	return fd, nil
}

// Close closes a thumbnail file
func (t *ThumbnailCache) Close(id string, size string) error {
	// Check if we have an open file descriptor
	if fd, ok := t.fds.Load(id + "-" + size); ok {
		// Remove from the map
		t.fds.Delete(id + "-" + size)

		// Close the file
		return fd.(*os.File).Close()
	}

	// No open file descriptor, nothing to do
	return nil
}

// CleanupCache removes thumbnails that haven't been accessed in a while
func (t *ThumbnailCache) CleanupCache(expirationDays int) (int, error) {
	// Only run cleanup once per day
	if time.Since(t.lastCleanup) < 24*time.Hour {
		return 0, nil
	}
	t.lastCleanup = time.Now()

	// Calculate the cutoff time
	cutoff := time.Now().AddDate(0, 0, -expirationDays)

	// Count of deleted files
	count := 0

	// Walk the directory and delete old files
	err := filepath.Walk(t.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if the file is older than the cutoff
		if info.ModTime().Before(cutoff) {
			// Try to remove the file
			if err := os.Remove(path); err != nil {
				// Log the error but continue
				logging.Error().Err(err).Str("path", path).Msg("Failed to remove old thumbnail file during cleanup")
				return nil
			}
			count++
		}

		return nil
	})

	// Force garbage collection after cleanup
	runtime.GC()

	return count, err
}
