package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
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
}

// GetStats returns statistics about the filesystem
func (f *Filesystem) GetStats() (*Stats, error) {
	stats := &Stats{
		Expiration: f.cacheExpirationDays,
		IsOffline:  f.IsOffline(),
		DeltaLink:  f.deltaLink,
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

	// Count items in each bucket
	err = f.db.View(func(tx *bolt.Tx) error {
		// Count metadata items
		if b := tx.Bucket(bucketMetadata); b != nil {
			stats.DBMetadataCount = b.Stats().KeyN
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
		log.Error().Err(err).Msg("Error counting items in bbolt buckets")
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
