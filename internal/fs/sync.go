package fs

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/logging"
)

// SyncProgress tracks the progress of directory tree synchronization
type SyncProgress struct {
	TotalDirectories     int64     // Total directories discovered
	ProcessedDirectories int64     // Directories processed so far
	TotalFiles           int64     // Total files discovered
	ProcessedFiles       int64     // Files processed so far
	StartTime            time.Time // When sync started
	LastUpdateTime       time.Time // Last progress update
	IsComplete           bool      // Whether sync is complete
	mutex                sync.RWMutex
}

// GetProgress returns a copy of the current sync progress
func (sp *SyncProgress) GetProgress() SyncProgress {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()
	return *sp
}

// UpdateProgress atomically updates the progress counters
func (sp *SyncProgress) UpdateProgress(processedDirs, processedFiles int64) {
	atomic.AddInt64(&sp.ProcessedDirectories, processedDirs)
	atomic.AddInt64(&sp.ProcessedFiles, processedFiles)

	sp.mutex.Lock()
	sp.LastUpdateTime = time.Now()
	sp.mutex.Unlock()
}

// AddDiscovered atomically adds to the discovered counters
func (sp *SyncProgress) AddDiscovered(dirs, files int64) {
	atomic.AddInt64(&sp.TotalDirectories, dirs)
	atomic.AddInt64(&sp.TotalFiles, files)
}

// MarkComplete marks the sync as complete
func (sp *SyncProgress) MarkComplete() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()
	sp.IsComplete = true
	sp.LastUpdateTime = time.Now()
}

// SyncDirectoryTree recursively traverses the filesystem from the root
// and caches all directory metadata (not file contents) with progress tracking
func (f *Filesystem) SyncDirectoryTree(auth *graph.Auth) error {
	return f.SyncDirectoryTreeWithContext(context.Background(), auth)
}

// SyncDirectoryTreeWithContext recursively traverses the filesystem from the root
// with context support for cancellation
func (f *Filesystem) SyncDirectoryTreeWithContext(ctx context.Context, auth *graph.Auth) error {
	logging.Info().Msg("Starting full directory tree synchronization...")

	// Initialize sync progress
	progress := &SyncProgress{
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
	}

	// Store progress in filesystem for external access
	f.Lock()
	f.syncProgress = progress
	f.Unlock()

	// Create a map to track visited directories to prevent cycles
	visited := make(map[string]bool)
	// Set a reasonable maximum depth to prevent excessive recursion
	const maxDepth = 20

	err := f.syncDirectoryTreeRecursiveWithContext(ctx, f.root, auth, visited, 0, maxDepth, progress)

	// Mark sync as complete
	progress.MarkComplete()

	if err != nil {
		logging.Error().Err(err).Msg("Directory tree synchronization completed with errors")
	} else {
		logging.Info().
			Int64("totalDirectories", atomic.LoadInt64(&progress.TotalDirectories)).
			Int64("totalFiles", atomic.LoadInt64(&progress.TotalFiles)).
			Dur("duration", time.Since(progress.StartTime)).
			Msg("Directory tree synchronization completed successfully")
	}

	return err
}

// syncDirectoryTreeRecursiveWithContext recursively traverses the filesystem
// starting from the given directory ID with context support and progress tracking
func (f *Filesystem) syncDirectoryTreeRecursiveWithContext(ctx context.Context, dirID string, auth *graph.Auth, visited map[string]bool, depth int, maxDepth int, progress *SyncProgress) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		logging.Debug().Str("dirID", dirID).Msg("Directory sync cancelled due to context cancellation")
		return ctx.Err()
	default:
		// Continue processing
	}

	// Check if we've already visited this directory to prevent cycles
	if visited[dirID] {
		logging.Debug().Str("dirID", dirID).Msg("Skipping already visited directory to prevent cycle")
		return nil
	}

	// Check if we've reached the maximum depth
	if depth >= maxDepth {
		logging.Warn().Str("dirID", dirID).Int("depth", depth).Int("maxDepth", maxDepth).Msg("Reached maximum recursion depth, stopping")
		return nil
	}

	// Mark this directory as visited
	visited[dirID] = true

	// Use prioritized metadata request for background sync
	var children map[string]*Inode
	var err error

	// Create a channel to receive the result
	resultChan := make(chan struct {
		children map[string]*Inode
		err      error
	}, 1)

	// Queue the metadata request with background priority
	if f.metadataRequestManager != nil {
		err = f.metadataRequestManager.QueueChildrenRequest(dirID, auth, PriorityBackground, func(items []*graph.DriveItem, reqErr error) {
			if reqErr != nil {
				resultChan <- struct {
					children map[string]*Inode
					err      error
				}{nil, reqErr}
				return
			}

			// Convert DriveItems to Inodes and cache them
			childrenMap := make(map[string]*Inode)
			var dirCount, fileCount int64

			for _, item := range items {
				child := NewInodeDriveItem(item)
				f.InsertNodeID(child)
				f.metadata.Store(child.DriveItem.ID, child)
				childrenMap[strings.ToLower(child.Name())] = child

				if child.IsDir() {
					dirCount++
				} else {
					fileCount++
				}
			}

			// Update progress counters
			progress.AddDiscovered(dirCount, fileCount)

			resultChan <- struct {
				children map[string]*Inode
				err      error
			}{childrenMap, nil}
		})

		if err != nil {
			// Fallback to direct call if queue is full
			logging.Debug().Str("dirID", dirID).Msg("Metadata queue full, falling back to direct call")
			children, err = f.GetChildrenID(dirID, auth)
		} else {
			// Wait for the result with timeout
			select {
			case result := <-resultChan:
				children = result.children
				err = result.err
			case <-time.After(30 * time.Second):
				err = context.DeadlineExceeded
				logging.Warn().Str("dirID", dirID).Msg("Metadata request timed out")
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	} else {
		// Fallback if metadata request manager is not available
		children, err = f.GetChildrenID(dirID, auth)
	}

	if err != nil {
		logging.Error().Err(err).Str("dirID", dirID).Msg("Failed to get children during sync")
		return err
	}

	// Update progress - processed one directory
	progress.UpdateProgress(1, int64(len(children)))

	// Log progress periodically
	if depth <= 2 { // Only log for top-level directories to avoid spam
		processedDirs := atomic.LoadInt64(&progress.ProcessedDirectories)
		totalFiles := atomic.LoadInt64(&progress.TotalFiles)
		logging.Info().
			Str("dirID", dirID).
			Int("depth", depth).
			Int64("processedDirectories", processedDirs).
			Int64("totalFiles", totalFiles).
			Int("childrenCount", len(children)).
			Msg("Sync progress update")
	}

	// Recursively process all subdirectories
	for _, child := range children {
		if child.IsDir() {
			if err := f.syncDirectoryTreeRecursiveWithContext(ctx, child.ID(), auth, visited, depth+1, maxDepth, progress); err != nil {
				// Check if it's a cancellation error
				if err == context.Canceled || err == context.DeadlineExceeded {
					return err
				}
				// Log the error but continue with other directories
				logging.Warn().Err(err).Str("dirID", child.ID()).Msg("Error syncing directory")
			}
		}
	}

	return nil
}

// Legacy function for backward compatibility
func (f *Filesystem) syncDirectoryTreeRecursive(dirID string, auth *graph.Auth, visited map[string]bool, depth int, maxDepth int) error {
	return f.syncDirectoryTreeRecursiveWithContext(context.Background(), dirID, auth, visited, depth, maxDepth, &SyncProgress{})
}
