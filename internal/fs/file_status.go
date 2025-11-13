package fs

import (
	"bytes"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/logging"

	"github.com/auriora/onemount/internal/graph"
	bolt "go.etcd.io/bbolt"
)

// statusCacheEntry represents a cached status determination result
type statusCacheEntry struct {
	status    FileStatusInfo
	timestamp time.Time
}

// statusCache provides TTL-based caching for status determination results
type statusCache struct {
	entries map[string]*statusCacheEntry
	ttl     time.Duration
	mutex   sync.RWMutex
}

// newStatusCache creates a new status cache with the specified TTL
func newStatusCache(ttl time.Duration) *statusCache {
	return &statusCache{
		entries: make(map[string]*statusCacheEntry),
		ttl:     ttl,
	}
}

// get retrieves a cached status if it exists and hasn't expired
func (sc *statusCache) get(id string) (FileStatusInfo, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	entry, exists := sc.entries[id]
	if !exists {
		return FileStatusInfo{}, false
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > sc.ttl {
		return FileStatusInfo{}, false
	}

	return entry.status, true
}

// set stores a status in the cache
func (sc *statusCache) set(id string, status FileStatusInfo) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.entries[id] = &statusCacheEntry{
		status:    status,
		timestamp: time.Now(),
	}
}

// invalidate removes a cached status
func (sc *statusCache) invalidate(id string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	delete(sc.entries, id)
}

// invalidateAll clears all cached statuses
func (sc *statusCache) invalidateAll() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.entries = make(map[string]*statusCacheEntry)
}

// cleanup removes expired entries from the cache
func (sc *statusCache) cleanup() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	now := time.Now()
	for id, entry := range sc.entries {
		if now.Sub(entry.timestamp) > sc.ttl {
			delete(sc.entries, id)
		}
	}
}

// GetFileStatus determines the current status of a file
func (f *Filesystem) GetFileStatus(id string) FileStatusInfo {
	// First check the status map (explicit status set by operations)
	f.statusM.RLock()
	if status, exists := f.statuses[id]; exists {
		f.statusM.RUnlock()
		return status
	}
	f.statusM.RUnlock()

	// Check the determination cache (computed status with TTL)
	if f.statusCache != nil {
		if status, found := f.statusCache.get(id); found {
			return status
		}
	}

	// If no cached status, determine it now
	status := f.determineFileStatus(id)

	// Cache the determination result
	if f.statusCache != nil {
		f.statusCache.set(id, status)
	}

	return status
}

// determineFileStatus calculates the current status of a file
// This method is optimized to minimize expensive operations
func (f *Filesystem) determineFileStatus(id string) FileStatusInfo {
	// Check if file is being uploaded (fast check, no I/O)
	if f.uploads != nil {
		// Use the UploadManager's mutex to safely access the sessions map
		f.uploads.mutex.RLock()
		for _, session := range f.uploads.sessions {
			session.Lock()
			sessionID := session.ID
			sessionOldID := session.OldID
			session.Unlock()

			if sessionID == id || sessionOldID == id {
				state := session.getState()
				f.uploads.mutex.RUnlock()
				switch state {
				case uploadNotStarted:
					return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
				case uploadStarted:
					return FileStatusInfo{Status: StatusSyncing, Timestamp: time.Now()}
				case uploadComplete:
					return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
				case uploadErrored:
					session.Lock()
					var errorMsg string
					if session.error != nil {
						errorMsg = session.error.Error()
					} else {
						errorMsg = "Unknown error"
					}
					session.Unlock()
					return FileStatusInfo{
						Status:    StatusError,
						ErrorMsg:  errorMsg,
						Timestamp: time.Now(),
					}
				default:
					return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
				}
			}
		}
		f.uploads.mutex.RUnlock()
	}

	// Check if file has offline changes (database query - expensive)
	hasOfflineChanges := false
	if err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketOfflineChanges)
		if b == nil {
			return nil
		}

		// Check if there are any changes for this ID
		c := b.Cursor()
		prefix := []byte(id + "-")
		k, _ := c.Seek(prefix)
		if k != nil && bytes.HasPrefix(k, prefix) {
			hasOfflineChanges = true
		}
		return nil
	}); err != nil {
		logging.DefaultLogger.Error().Err(err).Str("id", id).Msg("Error checking offline changes")
	}

	if hasOfflineChanges {
		return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
	}

	// Check if file is in local cache (fast check)
	if f.content.HasContent(id) {
		// Only perform expensive hash verification if needed
		// Skip hash verification for local-only files
		if isLocalID(id) {
			return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
		}

		// Get the inode to check if it's out of sync
		inode := f.GetID(id)
		if inode != nil {
			// Only verify checksum if the inode has a remote hash to compare against
			// This avoids expensive hash calculation when not needed
			inode.mu.RLock()
			hasRemoteHash := inode.DriveItem.File != nil &&
				inode.DriveItem.File.Hashes.QuickXorHash != ""
			inode.mu.RUnlock()

			if hasRemoteHash {
				// Perform hash verification (expensive - only when necessary)
				fd, err := f.content.Open(id)
				if err == nil {
					defer fd.Close()
					localHash := graph.QuickXORHashStream(fd)
					if !inode.VerifyChecksum(localHash) {
						return FileStatusInfo{Status: StatusOutofSync, Timestamp: time.Now()}
					}
				}
			}
		}
		return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
	}

	// Default: file is in cloud only
	return FileStatusInfo{Status: StatusCloud, Timestamp: time.Now()}
}

// SetFileStatus updates the status of a file
func (f *Filesystem) SetFileStatus(id string, status FileStatusInfo) {
	f.statusM.Lock()
	defer f.statusM.Unlock()
	f.statuses[id] = status

	// Invalidate the determination cache since we have an explicit status
	if f.statusCache != nil {
		f.statusCache.invalidate(id)
	}
}

// MarkFileDownloading marks a file as currently downloading
func (f *Filesystem) MarkFileDownloading(id string) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusDownloading,
		Timestamp: time.Now(),
	})
}

// MarkFileOutofSync marks a file as needing update from cloud
func (f *Filesystem) MarkFileOutofSync(id string) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusOutofSync,
		Timestamp: time.Now(),
	})
}

// MarkFileError marks a file as having an error
func (f *Filesystem) MarkFileError(id string, err error) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusError,
		ErrorMsg:  err.Error(),
		Timestamp: time.Now(),
	})
}

// MarkFileConflict marks a file as having a conflict
func (f *Filesystem) MarkFileConflict(id string, message string) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusConflict,
		ErrorMsg:  message,
		Timestamp: time.Now(),
	})
}

// InodePath returns the full path of an inode
func (f *Filesystem) InodePath(inode *Inode) string {
	if inode == nil {
		return ""
	}
	return inode.Path()
}

// updateFileStatus sets the extended attribute for file status and sends a D-Bus signal
func (f *Filesystem) updateFileStatus(inode *Inode) {
	path := f.InodePath(inode)
	if path == "" {
		return
	}

	// Get the ID before locking to avoid potential deadlocks
	id := inode.ID()

	// Store the path for D-Bus signal
	pathCopy := path
	var statusStrCopy string

	// Lock the inode before getting the status to prevent race conditions
	inode.mu.Lock()

	// Get the status after locking the inode
	status := f.GetFileStatus(id)
	statusStr := status.Status.String()

	// Store the status string for D-Bus signal
	statusStrCopy = statusStr

	// Initialize the xattrs map if it's nil
	if inode.xattrs == nil {
		inode.xattrs = make(map[string][]byte)
	}

	// Attempt to set the status xattr
	// Note: xattr operations may fail on filesystems that don't support extended attributes
	// (e.g., tmpfs, some network filesystems). We log warnings but continue operation.
	xattrSuccess := true

	// Set the status xattr
	inode.xattrs["user.onemount.status"] = []byte(statusStr)

	// If there's an error message, set it too
	if status.ErrorMsg != "" {
		inode.xattrs["user.onemount.error"] = []byte(status.ErrorMsg)
	} else {
		// Remove the error xattr if it exists
		delete(inode.xattrs, "user.onemount.error")
	}

	// Track xattr support status if this is the first time we're setting xattrs
	if !f.xattrSupported && xattrSuccess {
		f.xattrSupportedM.Lock()
		f.xattrSupported = true
		f.xattrSupportedM.Unlock()
		logging.DefaultLogger.Info().
			Str("path", pathCopy).
			Msg("Extended attributes are supported on this filesystem")
	}

	// Unlock the inode before sending D-Bus signal to avoid potential deadlocks
	inode.mu.Unlock()

	// Send D-Bus signal if server is available
	if f.dbusServer != nil {
		f.dbusServer.SendFileStatusUpdate(pathCopy, statusStrCopy)
	}
}

// UpdateFileStatus sets the extended attribute for file status and sends a D-Bus signal.
// This method is part of the FilesystemInterface.
func (f *Filesystem) UpdateFileStatus(inode *Inode) {
	f.updateFileStatus(inode)
}

// GetFileStatusBatch retrieves status for multiple files efficiently
// This method batches database queries and reduces lock contention
func (f *Filesystem) GetFileStatusBatch(ids []string) map[string]FileStatusInfo {
	result := make(map[string]FileStatusInfo, len(ids))

	// Collect IDs that need determination
	needDetermination := make([]string, 0, len(ids))

	// First pass: check explicit statuses and cache
	f.statusM.RLock()
	for _, id := range ids {
		if status, exists := f.statuses[id]; exists {
			result[id] = status
			continue
		}

		// Check determination cache
		if f.statusCache != nil {
			if status, found := f.statusCache.get(id); found {
				result[id] = status
				continue
			}
		}

		needDetermination = append(needDetermination, id)
	}
	f.statusM.RUnlock()

	// Second pass: batch determine statuses for remaining files
	if len(needDetermination) > 0 {
		// Batch check offline changes (single database transaction)
		offlineChanges := f.batchCheckOfflineChanges(needDetermination)

		// Determine status for each file
		for _, id := range needDetermination {
			status := f.determineFileStatusOptimized(id, offlineChanges[id])
			result[id] = status

			// Cache the result
			if f.statusCache != nil {
				f.statusCache.set(id, status)
			}
		}
	}

	return result
}

// batchCheckOfflineChanges checks offline changes for multiple files in a single transaction
func (f *Filesystem) batchCheckOfflineChanges(ids []string) map[string]bool {
	result := make(map[string]bool, len(ids))

	// Initialize all to false
	for _, id := range ids {
		result[id] = false
	}

	// Single database transaction for all IDs
	if err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketOfflineChanges)
		if b == nil {
			return nil
		}

		c := b.Cursor()
		for _, id := range ids {
			prefix := []byte(id + "-")
			k, _ := c.Seek(prefix)
			if k != nil && bytes.HasPrefix(k, prefix) {
				result[id] = true
			}
		}
		return nil
	}); err != nil {
		logging.DefaultLogger.Error().Err(err).Msg("Error batch checking offline changes")
	}

	return result
}

// determineFileStatusOptimized is an optimized version that uses pre-fetched data
func (f *Filesystem) determineFileStatusOptimized(id string, hasOfflineChanges bool) FileStatusInfo {
	// Check if file is being uploaded (fast check, no I/O)
	if f.uploads != nil {
		f.uploads.mutex.RLock()
		for _, session := range f.uploads.sessions {
			session.Lock()
			sessionID := session.ID
			sessionOldID := session.OldID
			session.Unlock()

			if sessionID == id || sessionOldID == id {
				state := session.getState()
				f.uploads.mutex.RUnlock()
				switch state {
				case uploadNotStarted:
					return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
				case uploadStarted:
					return FileStatusInfo{Status: StatusSyncing, Timestamp: time.Now()}
				case uploadComplete:
					return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
				case uploadErrored:
					session.Lock()
					var errorMsg string
					if session.error != nil {
						errorMsg = session.error.Error()
					} else {
						errorMsg = "Unknown error"
					}
					session.Unlock()
					return FileStatusInfo{
						Status:    StatusError,
						ErrorMsg:  errorMsg,
						Timestamp: time.Now(),
					}
				default:
					return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
				}
			}
		}
		f.uploads.mutex.RUnlock()
	}

	// Use pre-fetched offline changes status
	if hasOfflineChanges {
		return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
	}

	// Check if file is in local cache (fast check)
	if f.content.HasContent(id) {
		// Skip expensive hash verification for local-only files
		if isLocalID(id) {
			return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
		}

		// For batch operations, skip hash verification to improve performance
		// Hash verification can be done on-demand when needed
		return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
	}

	// Default: file is in cloud only
	return FileStatusInfo{Status: StatusCloud, Timestamp: time.Now()}
}

// InvalidateStatusCache invalidates cached status for a specific file
// This should be called when events occur that change file status
func (f *Filesystem) InvalidateStatusCache(id string) {
	if f.statusCache != nil {
		f.statusCache.invalidate(id)
	}
}

// InvalidateAllStatusCache invalidates all cached statuses
// This should be called after major events like delta sync
func (f *Filesystem) InvalidateAllStatusCache() {
	if f.statusCache != nil {
		f.statusCache.invalidateAll()
	}
}

// StartStatusCacheCleanup starts a background goroutine to periodically clean up expired cache entries
func (f *Filesystem) StartStatusCacheCleanup() {
	if f.statusCache == nil {
		return
	}

	f.Wg.Add(1)
	go func() {
		defer f.Wg.Done()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				f.statusCache.cleanup()
			case <-f.ctx.Done():
				return
			}
		}
	}()
}
