package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/logging"
)

// CacheEntry represents a cached file with LRU tracking
type CacheEntry struct {
	id           string
	size         int64
	lastAccessed time.Time
}

// LoopbackCache stores the content for files under a folder as regular files
type LoopbackCache struct {
	directory string
	fds       sync.Map
	// lastCleanup tracks when the last cache cleanup was performed
	lastCleanup time.Time
	// LRU tracking
	entriesM     sync.RWMutex
	entries      map[string]*CacheEntry // Map of file ID to cache entry
	totalSize    int64                  // Total size of all cached files
	maxCacheSize int64                  // Maximum cache size in bytes (0 = unlimited)

	evictionHandler func(string)
	evictionGuard   func(string) bool
}

// NewLoopbackCache creates a new LoopbackCache with optional size limit
// maxCacheSize: Maximum cache size in bytes (0 = unlimited)
func NewLoopbackCache(directory string) *LoopbackCache {
	return NewLoopbackCacheWithSize(directory, 0)
}

// NewLoopbackCacheWithSize creates a new LoopbackCache with a specified size limit
// maxCacheSize: Maximum cache size in bytes (0 = unlimited)
func NewLoopbackCacheWithSize(directory string, maxCacheSize int64) *LoopbackCache {
	if err := os.MkdirAll(directory, 0700); err != nil {
		// Log the error properly
		logging.Error().Err(err).Str("directory", directory).Msg("Failed to create content cache directory")
		// Try to create parent directories if they don't exist
		parentDir := filepath.Dir(directory)
		if err := os.MkdirAll(parentDir, 0700); err != nil {
			logging.Error().Err(err).Str("parentDir", parentDir).Msg("Failed to create parent directory for content cache")
		}
		// Try again to create the content directory
		if err := os.MkdirAll(directory, 0700); err != nil {
			logging.Error().Err(err).Str("directory", directory).Msg("Second attempt to create content cache directory failed")
		}
	}

	cache := &LoopbackCache{
		directory:    directory,
		fds:          sync.Map{},
		lastCleanup:  time.Now(),
		entries:      make(map[string]*CacheEntry),
		totalSize:    0,
		maxCacheSize: maxCacheSize,
	}

	// Initialize cache size tracking by scanning existing files
	cache.initializeCacheTracking()

	return cache
}

// SetEvictionHandler registers a callback invoked after a cache entry is evicted.
func (l *LoopbackCache) SetEvictionHandler(fn func(string)) {
	l.evictionHandler = fn
}

// SetEvictionGuard registers a guard invoked before evicting an entry.
// Returning false skips the eviction attempt for that entry.
func (l *LoopbackCache) SetEvictionGuard(fn func(string) bool) {
	l.evictionGuard = fn
}

// initializeCacheTracking scans the cache directory and builds the LRU tracking data
func (l *LoopbackCache) initializeCacheTracking() {
	l.entriesM.Lock()
	defer l.entriesM.Unlock()

	err := filepath.Walk(l.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the file ID from the path
		id := filepath.Base(path)

		// Add to tracking
		l.entries[id] = &CacheEntry{
			id:           id,
			size:         info.Size(),
			lastAccessed: info.ModTime(),
		}
		l.totalSize += info.Size()

		return nil
	})

	if err != nil {
		logging.Error().Err(err).Msg("Error initializing cache tracking")
	}

	logging.Info().
		Int64("totalSize", l.totalSize).
		Int("fileCount", len(l.entries)).
		Int64("maxCacheSize", l.maxCacheSize).
		Msg("Initialized cache tracking")
}

// contentPath returns the path for the given content file
func (l *LoopbackCache) contentPath(id string) string {
	return filepath.Join(l.directory, id)
}

// Get reads a file's content from disk.
func (l *LoopbackCache) Get(id string) []byte {
	content, err := os.ReadFile(l.contentPath(id))
	if err != nil {
		// Return empty content if file doesn't exist or can't be read
		// This matches the previous behavior but is more explicit
		return []byte{}
	}
	return content
}

// Insert InsertContent writes file content to disk in a single bulk insert.
func (l *LoopbackCache) Insert(id string, content []byte) error {
	size := int64(len(content))

	// Calculate the actual space needed (accounting for replacement)
	l.entriesM.RLock()
	existingSize := int64(0)
	if entry, exists := l.entries[id]; exists {
		existingSize = entry.size
	}
	l.entriesM.RUnlock()

	// Calculate net new space needed
	netNewSize := size - existingSize

	// Evict old entries if necessary to make room
	if err := l.evictIfNeeded(netNewSize); err != nil {
		return err
	}

	// Write the file
	if err := os.WriteFile(l.contentPath(id), content, 0600); err != nil {
		return err
	}

	// Update tracking
	l.updateCacheEntry(id, size)

	return nil
}

// InsertStream inserts a stream of data
func (l *LoopbackCache) InsertStream(id string, reader io.Reader) (int64, error) {
	fd, err := l.Open(id)
	if err != nil {
		return 0, err
	}

	// Copy the data from the reader to the file
	// We don't reset position or truncate here to maintain compatibility with existing code
	n, err := io.Copy(fd, reader)
	if err != nil {
		return n, err
	}

	// Update tracking with the actual size written
	l.updateCacheEntry(id, n)

	return n, nil
}

// Delete closes the fd AND deletes content from disk.
func (l *LoopbackCache) Delete(id string) error {
	// Try to close the file first
	closeErr := l.Close(id)

	// Try to remove the file regardless of close error
	removeErr := os.Remove(l.contentPath(id))

	// Update tracking - remove from cache size tracking
	l.removeCacheEntry(id)

	// Handle remove error - ignore "file not found" errors
	if removeErr != nil && !os.IsNotExist(removeErr) {
		return removeErr
	}

	// If we got here, the remove succeeded or the file didn't exist
	// Return any close error that might have occurred
	return closeErr
}

// Move moves content from one ID to another
func (l *LoopbackCache) Move(oldID string, newID string) error {
	// Close both files to ensure they're not open during the move
	// Capture errors but continue with the move operation
	oldCloseErr := l.Close(oldID)
	newCloseErr := l.Close(newID)

	// Make sure the destination directory exists
	destDir := filepath.Dir(l.contentPath(newID))
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return err
	}

	// Check if source file exists
	if _, err := os.Stat(l.contentPath(oldID)); os.IsNotExist(err) {
		return err
	}

	// Remove destination file if it exists to avoid "file exists" errors
	// Ignore any errors from this operation
	_ = os.Remove(l.contentPath(newID))

	// Perform the rename operation
	renameErr := os.Rename(l.contentPath(oldID), l.contentPath(newID))
	if renameErr != nil {
		return renameErr
	}

	// If we got here, the rename succeeded
	// Return any close errors that might have occurred
	if oldCloseErr != nil {
		return oldCloseErr
	}
	return newCloseErr
}

// IsOpen returns true if the file is already opened somewhere
func (l *LoopbackCache) IsOpen(id string) bool {
	_, ok := l.fds.Load(id)
	return ok
}

// HasContent is used to find if we have a file or not in cache (in any state)
func (l *LoopbackCache) HasContent(id string) bool {
	// is it already open?
	_, ok := l.fds.Load(id)
	if ok {
		return ok
	}
	// is it on disk?
	_, err := os.Stat(l.contentPath(id))
	return err == nil
}

// Open returns a filehandle for subsequent access
func (l *LoopbackCache) Open(id string) (*os.File, error) {
	if fd, ok := l.fds.Load(id); ok {
		// already opened, return existing fd
		// Touch the cache entry to update last accessed time
		l.touchCacheEntry(id)
		return fd.(*os.File), nil
	}

	// Ensure the parent directory exists before opening the file
	filePath := l.contentPath(id)
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		logging.Error().Err(err).Str("directory", dirPath).Msg("Failed to create parent directory for content file")
		return nil, err
	}

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Since we explicitly want to store *os.Files, we need to prevent the Go
	// GC from trying to be "helpful" and closing files for us behind the
	// scenes.
	// https://github.com/hanwen/go-fuse/issues/371#issuecomment-694799535
	runtime.SetFinalizer(fd, nil)
	l.fds.Store(id, fd)

	// Touch the cache entry to update last accessed time
	l.touchCacheEntry(id)

	return fd, nil
}

// Close closes the currently open fd
func (l *LoopbackCache) Close(id string) error {
	if fd, ok := l.fds.Load(id); ok {
		file := fd.(*os.File)

		// Try to sync the file, but don't fail if it doesn't work
		// We still want to try to close the file even if sync fails
		syncErr := file.Sync()

		// Close the file and capture any error
		closeErr := file.Close()

		// Remove from the map regardless of errors
		l.fds.Delete(id)

		// Return the first error encountered
		if syncErr != nil {
			return syncErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

// CleanupCache removes files from the content cache that haven't been modified
// for the specified number of days. Also enforces cache size limits if configured.
// Returns the number of files removed and any error.
func (l *LoopbackCache) CleanupCache(expirationDays int) (int, error) {
	// Update the last cleanup time
	l.lastCleanup = time.Now()

	// Calculate the cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -expirationDays)

	// Count of removed files
	removedCount := 0

	// Walk through the content directory
	err := filepath.Walk(l.directory, func(path string, info os.FileInfo, err error) error {
		// Skip the root directory
		if path == l.directory {
			return nil
		}

		// Skip if there was an error accessing the file
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if the file's modification time is older than the cutoff
		if info.ModTime().Before(cutoffTime) {
			// Get the file ID from the path
			id := filepath.Base(path)

			// Check if the file is currently open
			if l.IsOpen(id) {
				// Skip files that are currently open
				return nil
			}

			// Remove the file
			if err := os.Remove(path); err != nil {
				// Log the error but continue with other files
				return nil
			}

			// Update tracking
			l.removeCacheEntry(id)

			removedCount++
		}

		return nil
	})

	// After time-based cleanup, enforce size limits if configured
	if l.maxCacheSize > 0 {
		l.entriesM.RLock()
		currentSize := l.totalSize
		l.entriesM.RUnlock()

		if currentSize > l.maxCacheSize {
			logging.Info().
				Int64("currentSize", currentSize).
				Int64("maxCacheSize", l.maxCacheSize).
				Msg("Cache size exceeds limit after time-based cleanup, performing LRU eviction")

			// Evict entries to get under the limit
			spaceToFree := currentSize - l.maxCacheSize
			if evictErr := l.evictIfNeeded(0); evictErr != nil {
				logging.Error().Err(evictErr).Msg("Failed to enforce cache size limit during cleanup")
			} else {
				logging.Info().
					Int64("freedSpace", spaceToFree).
					Msg("Successfully enforced cache size limit")
			}
		}
	}

	return removedCount, err
}

// updateCacheEntry updates the cache entry for a file
func (l *LoopbackCache) updateCacheEntry(id string, size int64) {
	l.entriesM.Lock()
	defer l.entriesM.Unlock()

	// Remove old size if entry exists
	if entry, exists := l.entries[id]; exists {
		l.totalSize -= entry.size
	}

	// Add new entry
	l.entries[id] = &CacheEntry{
		id:           id,
		size:         size,
		lastAccessed: time.Now(),
	}
	l.totalSize += size

	logging.Debug().
		Str("id", id).
		Int64("size", size).
		Int64("totalSize", l.totalSize).
		Int64("maxCacheSize", l.maxCacheSize).
		Msg("Updated cache entry")
}

// removeCacheEntry removes a cache entry
func (l *LoopbackCache) removeCacheEntry(id string) {
	l.entriesM.Lock()
	defer l.entriesM.Unlock()

	if entry, exists := l.entries[id]; exists {
		l.totalSize -= entry.size
		delete(l.entries, id)

		logging.Debug().
			Str("id", id).
			Int64("size", entry.size).
			Int64("totalSize", l.totalSize).
			Msg("Removed cache entry")
	}
}

// touchCacheEntry updates the last accessed time for a cache entry
func (l *LoopbackCache) touchCacheEntry(id string) {
	l.entriesM.Lock()
	defer l.entriesM.Unlock()

	if entry, exists := l.entries[id]; exists {
		entry.lastAccessed = time.Now()
	}
}

// evictIfNeeded evicts old entries if the cache size would exceed the limit
func (l *LoopbackCache) evictIfNeeded(newSize int64) error {
	// If no size limit is set, no eviction needed
	if l.maxCacheSize == 0 {
		return nil
	}

	l.entriesM.Lock()
	defer l.entriesM.Unlock()

	// Calculate how much space we need
	spaceNeeded := (l.totalSize + newSize) - l.maxCacheSize
	if spaceNeeded <= 0 {
		return nil // No eviction needed
	}

	logging.Info().
		Int64("currentSize", l.totalSize).
		Int64("newSize", newSize).
		Int64("maxCacheSize", l.maxCacheSize).
		Int64("spaceNeeded", spaceNeeded).
		Msg("Cache size limit exceeded, evicting old entries")

	// Build a sorted list of entries by last accessed time (oldest first)
	type entryWithTime struct {
		id           string
		size         int64
		lastAccessed time.Time
	}

	entries := make([]entryWithTime, 0, len(l.entries))
	for id, entry := range l.entries {
		// Skip files that are currently open
		if l.IsOpen(id) {
			continue
		}

		entries = append(entries, entryWithTime{
			id:           id,
			size:         entry.size,
			lastAccessed: entry.lastAccessed,
		})
	}

	// Sort by last accessed time (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].lastAccessed.Before(entries[j].lastAccessed)
	})

	// Evict entries until we have enough space
	var evictedSize int64
	var evictedCount int
	var skipped int

	for _, entry := range entries {
		if evictedSize >= spaceNeeded {
			break
		}

		if l.evictionGuard != nil && !l.evictionGuard(entry.id) {
			skipped++
			continue
		}

		// Remove the file
		if err := os.Remove(l.contentPath(entry.id)); err != nil && !os.IsNotExist(err) {
			logging.Warn().Err(err).Str("id", entry.id).Msg("Failed to evict cache entry")
			continue
		}

		// Update tracking
		delete(l.entries, entry.id)
		l.totalSize -= entry.size
		evictedSize += entry.size
		evictedCount++

		if l.evictionHandler != nil {
			l.evictionHandler(entry.id)
		}

		logging.Debug().
			Str("id", entry.id).
			Int64("size", entry.size).
			Time("lastAccessed", entry.lastAccessed).
			Msg("Evicted cache entry")
	}

	logging.Info().
		Int("evictedCount", evictedCount).
		Int64("evictedSize", evictedSize).
		Int64("newTotalSize", l.totalSize).
		Msg("Cache eviction completed")

	// Check if we freed enough space
	if l.totalSize+newSize > l.maxCacheSize {
		if skipped > 0 {
			return fmt.Errorf("unable to free cache space: %d entries skipped by guard, need %d bytes, freed %d bytes", skipped, spaceNeeded, evictedSize)
		}
		return fmt.Errorf("unable to free enough cache space: need %d bytes, freed %d bytes", spaceNeeded, evictedSize)
	}

	return nil
}

// GetCacheSize returns the current total size of cached files
func (l *LoopbackCache) GetCacheSize() int64 {
	l.entriesM.RLock()
	defer l.entriesM.RUnlock()
	return l.totalSize
}

// GetMaxCacheSize returns the maximum cache size limit
func (l *LoopbackCache) GetMaxCacheSize() int64 {
	l.entriesM.RLock()
	defer l.entriesM.RUnlock()
	return l.maxCacheSize
}

// SetMaxCacheSize sets the maximum cache size limit
func (l *LoopbackCache) SetMaxCacheSize(maxSize int64) {
	l.entriesM.Lock()
	l.maxCacheSize = maxSize
	l.entriesM.Unlock()

	logging.Info().
		Int64("maxCacheSize", maxSize).
		Msg("Updated maximum cache size")

	// Trigger eviction if we're over the new limit
	if maxSize > 0 {
		if err := l.evictIfNeeded(0); err != nil {
			logging.Error().Err(err).Msg("Failed to evict entries after setting new cache size limit")
		}
	}
}

// GetCacheEntryCount returns the number of cached files
func (l *LoopbackCache) GetCacheEntryCount() int {
	l.entriesM.RLock()
	defer l.entriesM.RUnlock()
	return len(l.entries)
}
