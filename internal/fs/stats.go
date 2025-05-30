package fs

import (
	"fmt"
	"github.com/auriora/onemount/pkg/logging"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Stats represents statistics about the filesystem
type Stats struct {
	// Metadata statistics
	MetadataCount int

	// Content cache statistics
	ContentCount int
	ContentSize  int64
	ContentDir   string
	Expiration   int

	// Upload queue statistics
	UploadCount       int
	UploadsNotStarted int
	UploadsInProgress int
	UploadsCompleted  int
	UploadsErrored    int

	// File status statistics
	StatusCloud         int
	StatusLocal         int
	StatusLocalModified int
	StatusSyncing       int
	StatusDownloading   int
	StatusOutofSync     int
	StatusError         int
	StatusConflict      int

	// Delta link information
	DeltaLink string

	// Offline status
	IsOffline bool

	// BBolt database statistics
	DBPath          string
	DBSize          int64
	DBPageCount     int
	DBPageSize      int
	DBMetadataCount int
	DBDeltaCount    int
	DBOfflineCount  int
	DBUploadsCount  int

	// File type statistics
	FileExtensions map[string]int // Count of files by extension

	// Directory statistics
	MaxDirDepth     int     // Maximum directory depth
	AvgDirDepth     float64 // Average directory depth
	DirCount        int     // Number of directories
	EmptyDirCount   int     // Number of empty directories
	AvgFilesPerDir  float64 // Average number of files per directory
	MaxFilesInDir   int     // Maximum number of files in a directory
	MaxFilesInDirID string  // ID of directory with maximum files

	// File size statistics
	FileSizeRanges map[string]int // Count of files in different size ranges

	// Age statistics
	FileAgeRanges map[string]int // Count of files by age range
}

// GetStats returns statistics about the filesystem
func (f *Filesystem) GetStats() (*Stats, error) {
	// TODO: Optimize statistics collection for large filesystems (Issue #11, #10, #9, #8, #7)
	// Current implementation performs full traversal of metadata and content directories
	// which can be slow for large filesystems (>100k files).
	// Performance optimizations to implement in v1.1:
	// 1. Implement incremental statistics updates instead of full recalculation
	// 2. Cache frequently accessed statistics with TTL
	// 3. Use background goroutines for expensive calculations
	// 4. Implement sampling for very large datasets
	// 5. Add pagination support for statistics display
	// 6. Optimize database queries with better indexing
	// 7. Consider using separate statistics database/table
	// Target: v1.1 release
	// Priority: Medium (acceptable performance for typical use cases)

	stats := &Stats{
		Expiration:     f.cacheExpirationDays,
		IsOffline:      f.IsOffline(),
		DeltaLink:      f.deltaLink,
		FileExtensions: make(map[string]int),
		FileSizeRanges: make(map[string]int),
		FileAgeRanges:  make(map[string]int),
	}

	// Count metadata items
	f.metadata.Range(func(_, _ interface{}) bool {
		stats.MetadataCount++
		return true
	})

	// Content cache statistics
	contentDir := filepath.Join(filepath.Dir(f.db.Path()), "content")
	stats.ContentDir = contentDir
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			stats.ContentCount++
			stats.ContentSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking content directory: %w", err)
	}

	// BBolt database statistics
	dbPath := f.db.Path()
	stats.DBPath = dbPath

	// Get database file size
	dbInfo, err := os.Stat(dbPath)
	if err == nil {
		stats.DBSize = dbInfo.Size()
	}

	// Count items in each bucket and analyze metadata
	err = f.db.View(func(tx *bolt.Tx) error {
		// Count metadata items
		metadataBucket := tx.Bucket(bucketMetadata)
		if metadataBucket != nil {
			stats.DBMetadataCount = metadataBucket.Stats().KeyN

			// Analyze metadata for additional statistics
			dirDepths := make(map[string]int)        // Map of directory ID to its depth
			dirChildren := make(map[string][]string) // Map of directory ID to its children
			fileCount := 0                           // Total number of files
			now := time.Now()

			// First pass: collect basic information about each item
			err := metadataBucket.ForEach(func(k, v []byte) error {
				inode, err := NewInodeJSON(v)
				if err != nil {
					return nil // Skip items that can't be parsed
				}

				id := string(k)
				parentID := inode.ParentID()

				// Skip the root item in some calculations
				if id == "root" || id == f.root {
					dirDepths[id] = 0
					return nil
				}

				// Initialize children slice for this parent if it doesn't exist
				if _, exists := dirChildren[parentID]; !exists {
					dirChildren[parentID] = make([]string, 0)
				}

				// Add this item to its parent's children
				dirChildren[parentID] = append(dirChildren[parentID], id)

				if inode.IsDir() {
					stats.DirCount++

					// Calculate directory depth (parent's depth + 1)
					if parentDepth, exists := dirDepths[parentID]; exists {
						dirDepths[id] = parentDepth + 1
						if dirDepths[id] > stats.MaxDirDepth {
							stats.MaxDirDepth = dirDepths[id]
						}
					} else {
						// If parent depth is unknown, assume it's 1 (under root)
						dirDepths[id] = 1
					}
				} else {
					fileCount++

					// File extension statistics
					ext := filepath.Ext(inode.Name())
					if ext != "" {
						ext = strings.ToLower(ext)
						stats.FileExtensions[ext]++
					} else {
						stats.FileExtensions["(no extension)"]++
					}

					// File size statistics
					size := inode.Size()
					switch {
					case size == 0:
						stats.FileSizeRanges["Empty (0 bytes)"]++
					case size < 1024:
						stats.FileSizeRanges["< 1 KB"]++
					case size < 1024*1024:
						stats.FileSizeRanges["1 KB - 1 MB"]++
					case size < 10*1024*1024:
						stats.FileSizeRanges["1 MB - 10 MB"]++
					case size < 100*1024*1024:
						stats.FileSizeRanges["10 MB - 100 MB"]++
					case size < 1024*1024*1024:
						stats.FileSizeRanges["100 MB - 1 GB"]++
					default:
						stats.FileSizeRanges["> 1 GB"]++
					}

					// File age statistics
					if inode.ModTime() > 0 {
						modTime := time.Unix(int64(inode.ModTime()), 0)
						ageInDays := int(now.Sub(modTime).Hours() / 24)

						switch {
						case ageInDays < 1:
							stats.FileAgeRanges["Today"]++
						case ageInDays < 7:
							stats.FileAgeRanges["This week"]++
						case ageInDays < 30:
							stats.FileAgeRanges["This month"]++
						case ageInDays < 90:
							stats.FileAgeRanges["Last 3 months"]++
						case ageInDays < 365:
							stats.FileAgeRanges["This year"]++
						default:
							stats.FileAgeRanges["Older than a year"]++
						}
					}
				}

				return nil
			})

			if err != nil {
				return err
			}

			// Second pass: calculate directory statistics
			var totalDepth int
			var totalDirs int

			for dirID, children := range dirChildren {
				fileCountInDir := 0
				for _, childID := range children {
					child := f.GetID(childID)
					if child != nil && !child.IsDir() {
						fileCountInDir++
					}
				}

				// Check for empty directories
				if len(children) == 0 {
					stats.EmptyDirCount++
				}

				// Track directory with most files
				if fileCountInDir > stats.MaxFilesInDir {
					stats.MaxFilesInDir = fileCountInDir
					stats.MaxFilesInDirID = dirID
				}

				// Add to totals for averages
				if depth, exists := dirDepths[dirID]; exists && dirID != "root" && dirID != f.root {
					totalDepth += depth
					totalDirs++
				}
			}

			// Calculate averages
			if totalDirs > 0 {
				stats.AvgDirDepth = float64(totalDepth) / float64(totalDirs)
			}

			if stats.DirCount > 0 {
				stats.AvgFilesPerDir = float64(fileCount) / float64(stats.DirCount)
			}
		}

		// Count delta items
		if b := tx.Bucket(bucketDelta); b != nil {
			stats.DBDeltaCount = b.Stats().KeyN
		}

		// Count offline changes
		if b := tx.Bucket(bucketOfflineChanges); b != nil {
			stats.DBOfflineCount = b.Stats().KeyN
		}

		// Count uploads
		if b := tx.Bucket(bucketUploads); b != nil {
			stats.DBUploadsCount = b.Stats().KeyN
		}

		return nil
	})
	if err != nil {
		logging.Error().Err(err).Msg("Error analyzing metadata in bbolt database")
	}

	// Get database page statistics
	dbStats := f.db.Stats()
	stats.DBPageCount = dbStats.FreePageN + dbStats.PendingPageN + dbStats.FreeAlloc
	stats.DBPageSize = os.Getpagesize() // Use system page size as bbolt doesn't expose page size directly

	// Upload queue statistics
	f.uploads.mutex.RLock()
	stats.UploadCount = len(f.uploads.sessions)
	for _, session := range f.uploads.sessions {
		state := session.getState()
		switch state {
		case uploadNotStarted:
			stats.UploadsNotStarted++
		case uploadStarted:
			stats.UploadsInProgress++
		case uploadComplete:
			stats.UploadsCompleted++
		case uploadErrored:
			stats.UploadsErrored++
		}
	}
	f.uploads.mutex.RUnlock()

	// File status statistics
	f.statusM.RLock()
	for _, status := range f.statuses {
		switch status.Status {
		case StatusCloud:
			stats.StatusCloud++
		case StatusLocal:
			stats.StatusLocal++
		case StatusLocalModified:
			stats.StatusLocalModified++
		case StatusSyncing:
			stats.StatusSyncing++
		case StatusDownloading:
			stats.StatusDownloading++
		case StatusOutofSync:
			stats.StatusOutofSync++
		case StatusError:
			stats.StatusError++
		case StatusConflict:
			stats.StatusConflict++
		}
	}
	f.statusM.RUnlock()

	return stats, nil
}

// FormatSize formats a size in bytes to a human-readable string
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}
