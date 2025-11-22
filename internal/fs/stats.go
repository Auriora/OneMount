package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/socketio"

	bolt "go.etcd.io/bbolt"
)

// Stats represents statistics about the filesystem
type Stats struct {
	// Metadata statistics
	MetadataCount int

	// Content cache statistics
	ContentCount   int
	ContentSize    int64
	ContentDir     string
	Expiration     int
	MaxCacheSize   int64   // Maximum cache size limit (0 = unlimited)
	CacheSizeUsage float64 // Percentage of max cache size used (0-100, or -1 if unlimited)

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

	// Cache metadata
	CachedAt  time.Time // When these statistics were cached
	IsSampled bool      // Whether statistics were calculated using sampling

	// Extended attributes support
	XAttrSupported bool // Whether extended attributes are supported on this filesystem

	// Realtime transport insights
	RealtimeMode                string
	RealtimeStatus              socketio.StatusCode
	RealtimeLastHeartbeat       time.Time
	RealtimeMissedHeartbeats    int
	RealtimeConsecutiveFailures int
	RealtimeReconnectCount      int
	RealtimeLastError           string
	RealtimeRecoveryWindowOpen  bool
	RealtimeRecoverySince       time.Time

	// Metadata/hydration telemetry
	MetadataStateCounts      map[string]int
	HydrationHydrating       int
	HydrationHydrated        int
	HydrationGhost           int
	HydrationDirtyLocal      int
	HydrationErrored         int
	HydrationQueueDepth      int
	HydrationActiveDownloads int
	MetadataQueueHighDepth   int
	MetadataQueueLowDepth    int
	MetadataQueueAvgWaitMs   float64
}

// CachedStats holds cached statistics with TTL
type CachedStats struct {
	stats     *Stats
	expiresAt time.Time
	mu        sync.RWMutex
}

// StatsConfig holds configuration for statistics collection
type StatsConfig struct {
	// CacheTTL is the time-to-live for cached statistics
	CacheTTL time.Duration
	// SamplingThreshold is the number of items above which sampling is used
	SamplingThreshold int
	// SamplingRate is the percentage of items to sample (0.0-1.0)
	SamplingRate float64
	// UseBackgroundCalculation enables background goroutines for expensive calculations
	UseBackgroundCalculation bool
}

// DefaultStatsConfig returns the default statistics configuration
func DefaultStatsConfig() *StatsConfig {
	return &StatsConfig{
		CacheTTL:                 5 * time.Minute, // Cache stats for 5 minutes
		SamplingThreshold:        10000,           // Use sampling for >10k items
		SamplingRate:             0.1,             // Sample 10% of items
		UseBackgroundCalculation: true,
	}
}

// GetStats returns statistics about the filesystem with caching and optimization
func (f *Filesystem) GetStats() (*Stats, error) {
	return f.GetStatsWithConfig(nil)
}

// GetStatsWithConfig returns statistics about the filesystem with custom configuration
func (f *Filesystem) GetStatsWithConfig(config *StatsConfig) (*Stats, error) {
	// Use default config if none provided
	if config == nil {
		if f.statsConfig == nil {
			f.statsConfig = DefaultStatsConfig()
		}
		config = f.statsConfig
	}

	// Check if we have cached stats that are still valid
	if f.cachedStats != nil {
		f.cachedStats.mu.RLock()
		if f.cachedStats.stats != nil && time.Now().Before(f.cachedStats.expiresAt) {
			stats := f.cachedStats.stats
			f.augmentRealtimeStats(stats)
			f.cachedStats.mu.RUnlock()
			logging.Debug().Msg("Returning cached statistics")
			return stats, nil
		}
		f.cachedStats.mu.RUnlock()
	}

	// Calculate new statistics
	logging.Debug().Msg("Calculating fresh statistics")
	stats, err := f.calculateStats(config)
	if err != nil {
		return nil, err
	}

	stats.HydrationHydrating = stats.MetadataStateCounts[string(metadata.ItemStateHydrating)]
	stats.HydrationHydrated = stats.MetadataStateCounts[string(metadata.ItemStateHydrated)]
	stats.HydrationGhost = stats.MetadataStateCounts[string(metadata.ItemStateGhost)]
	stats.HydrationDirtyLocal = stats.MetadataStateCounts[string(metadata.ItemStateDirtyLocal)]
	stats.HydrationErrored = stats.MetadataStateCounts[string(metadata.ItemStateError)]

	if f.downloads != nil {
		snap := f.downloads.Snapshot()
		stats.HydrationQueueDepth = snap.QueueDepth
		stats.HydrationActiveDownloads = snap.Active
	}

	if f.metadataRequestManager != nil {
		q := f.metadataRequestManager.Snapshot()
		stats.MetadataQueueHighDepth = q.HighDepth
		stats.MetadataQueueLowDepth = q.LowDepth
		stats.MetadataQueueAvgWaitMs = q.AvgWaitMs
	}

	// Cache the statistics
	if f.cachedStats == nil {
		f.cachedStats = &CachedStats{}
	}
	f.cachedStats.mu.Lock()
	f.cachedStats.stats = stats
	f.cachedStats.expiresAt = time.Now().Add(config.CacheTTL)
	f.cachedStats.mu.Unlock()
	f.augmentRealtimeStats(stats)

	// Trigger background update if enabled
	if config.UseBackgroundCalculation && f.statsUpdateCh != nil {
		select {
		case f.statsUpdateCh <- struct{}{}:
		default:
			// Channel full, skip background update
		}
	}

	return stats, nil
}

// InvalidateStatsCache invalidates the cached statistics, forcing a recalculation on next request
func (f *Filesystem) InvalidateStatsCache() {
	if f.cachedStats != nil {
		f.cachedStats.mu.Lock()
		f.cachedStats.stats = nil
		f.cachedStats.expiresAt = time.Time{}
		f.cachedStats.mu.Unlock()
		logging.Debug().Msg("Statistics cache invalidated")
	}
}

// calculateStats performs the actual statistics calculation
func (f *Filesystem) calculateStats(config *StatsConfig) (*Stats, error) {
	// Get xattr support status
	f.xattrSupportedM.RLock()
	xattrSupported := f.xattrSupported
	f.xattrSupportedM.RUnlock()

	stats := &Stats{
		Expiration:     f.cacheExpirationDays,
		IsOffline:      f.IsOffline(),
		DeltaLink:      f.deltaLink,
		FileExtensions: make(map[string]int),
		FileSizeRanges: make(map[string]int),
		FileAgeRanges:  make(map[string]int),
		CachedAt:       time.Now(),
		XAttrSupported: xattrSupported,
	}

	// Count metadata items (fast operation, no optimization needed)
	f.metadata.Range(func(_, _ interface{}) bool {
		stats.MetadataCount++
		return true
	})

	// Determine if we should use sampling based on metadata count
	useSampling := stats.MetadataCount > config.SamplingThreshold
	stats.IsSampled = useSampling

	if useSampling {
		logging.Debug().
			Int("metadataCount", stats.MetadataCount).
			Int("threshold", config.SamplingThreshold).
			Float64("samplingRate", config.SamplingRate).
			Msg("Using sampling for statistics calculation")
	}

	// Content cache statistics - use background goroutine if enabled
	contentDir := filepath.Join(filepath.Dir(f.db.Path()), "content")
	stats.ContentDir = contentDir

	// Get cache size limits from content cache
	stats.MaxCacheSize = f.content.GetMaxCacheSize()
	currentCacheSize := f.content.GetCacheSize()
	if stats.MaxCacheSize > 0 {
		stats.CacheSizeUsage = (float64(currentCacheSize) / float64(stats.MaxCacheSize)) * 100.0
	} else {
		stats.CacheSizeUsage = -1 // Unlimited
	}

	if config.UseBackgroundCalculation {
		// Use a channel to collect results from background goroutine
		type contentStats struct {
			count int
			size  int64
			err   error
		}
		resultCh := make(chan contentStats, 1)

		go func() {
			var count int
			var size int64
			err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() {
					count++
					size += info.Size()
				}
				return nil
			})
			resultCh <- contentStats{count: count, size: size, err: err}
		}()

		// Wait for result with timeout
		select {
		case result := <-resultCh:
			if result.err != nil {
				logging.Warn().Err(result.err).Msg("Error walking content directory in background")
			}
			stats.ContentCount = result.count
			stats.ContentSize = result.size
		case <-time.After(f.timeoutConfig.ContentStatsTimeout):
			logging.Warn().
				Dur("timeout", f.timeoutConfig.ContentStatsTimeout).
				Msg("Timeout waiting for content cache statistics, using partial results")
			// Continue with empty content stats
		}
	} else {
		// Synchronous calculation
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
		metadataBucket := tx.Bucket(bucketMetadataV2)
		if metadataBucket != nil {
			stats.DBMetadataCount = metadataBucket.Stats().KeyN

			// Analyze metadata for additional statistics
			dirDepths := make(map[string]int)        // Map of directory ID to its depth
			dirChildren := make(map[string][]string) // Map of directory ID to its children
			fileCount := 0                           // Total number of files
			now := time.Now()

			var sampleEveryN = 1
			if useSampling {
				sampleEveryN = int(1.0 / config.SamplingRate)
				if sampleEveryN < 1 {
					sampleEveryN = 1
				}
			}

			itemIndex := 0
			err := metadataBucket.ForEach(func(k, v []byte) error {
				itemIndex++
				if useSampling && itemIndex%sampleEveryN != 0 {
					return nil
				}
				var entry metadata.Entry
				if err := json.Unmarshal(v, &entry); err != nil {
					return nil
				}
				inode := f.inodeFromMetadataEntry(&entry)
				if inode == nil {
					return nil
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

			// If sampling was used, extrapolate the counts
			if useSampling && sampleEveryN > 1 {
				scaleFactor := float64(sampleEveryN)

				// Scale file type statistics
				for ext, count := range stats.FileExtensions {
					stats.FileExtensions[ext] = int(float64(count) * scaleFactor)
				}

				// Scale file size statistics
				for sizeRange, count := range stats.FileSizeRanges {
					stats.FileSizeRanges[sizeRange] = int(float64(count) * scaleFactor)
				}

				// Scale file age statistics
				for ageRange, count := range stats.FileAgeRanges {
					stats.FileAgeRanges[ageRange] = int(float64(count) * scaleFactor)
				}

				// Scale directory statistics
				stats.DirCount = int(float64(stats.DirCount) * scaleFactor)
				stats.EmptyDirCount = int(float64(stats.EmptyDirCount) * scaleFactor)

				logging.Debug().
					Float64("scaleFactor", scaleFactor).
					Msg("Extrapolated statistics from sampled data")
			}
		}

		if metadataV2 := tx.Bucket(bucketMetadataV2); metadataV2 != nil {
			if stats.MetadataStateCounts == nil {
				stats.MetadataStateCounts = make(map[string]int)
			}
			if err := metadataV2.ForEach(func(k, v []byte) error {
				var entry metadata.Entry
				if err := json.Unmarshal(v, &entry); err != nil {
					return nil
				}
				state := string(entry.State)
				if state == "" {
					state = "UNKNOWN"
				}
				stats.MetadataStateCounts[state]++
				return nil
			}); err != nil {
				return err
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
	if f.uploads != nil {
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
	}

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

// StartBackgroundStatsUpdater starts a background goroutine that periodically updates statistics
func (f *Filesystem) StartBackgroundStatsUpdater(ctx context.Context, interval time.Duration) {
	if f.statsUpdateCh == nil {
		f.statsUpdateCh = make(chan struct{}, 1)
	}

	if f.statsConfig == nil {
		f.statsConfig = DefaultStatsConfig()
	}

	f.Wg.Add(1)
	go func() {
		defer f.Wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logging.Info().
			Dur("interval", interval).
			Msg("Started background statistics updater")

		for {
			select {
			case <-ctx.Done():
				logging.Info().Msg("Stopping background statistics updater")
				return
			case <-ticker.C:
				// Periodic update
				logging.Debug().Msg("Triggering periodic statistics update")
				f.updateStatsInBackground()
			case <-f.statsUpdateCh:
				// Triggered update
				logging.Debug().Msg("Triggering on-demand statistics update")
				f.updateStatsInBackground()
			}
		}
	}()
}

// updateStatsInBackground updates statistics in the background without blocking
func (f *Filesystem) updateStatsInBackground() {
	go func() {
		startTime := time.Now()
		_, err := f.calculateStats(f.statsConfig)
		duration := time.Since(startTime)

		if err != nil {
			logging.Error().
				Err(err).
				Dur("duration", duration).
				Msg("Background statistics update failed")
		} else {
			logging.Debug().
				Dur("duration", duration).
				Msg("Background statistics update completed")
		}
	}()
}

// GetStatsPage returns a paginated view of statistics for large datasets
func (f *Filesystem) GetStatsPage(category string, page int, pageSize int) (map[string]int, error) {
	stats, err := f.GetStats()
	if err != nil {
		return nil, err
	}

	var data map[string]int
	switch category {
	case "extensions":
		data = stats.FileExtensions
	case "sizes":
		data = stats.FileSizeRanges
	case "ages":
		data = stats.FileAgeRanges
	default:
		return nil, fmt.Errorf("unknown category: %s", category)
	}

	// Convert map to sorted slice for pagination
	type kv struct {
		key   string
		value int
	}
	var sorted []kv
	for k, v := range data {
		sorted = append(sorted, kv{k, v})
	}

	// Sort by value (descending)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].value > sorted[i].value {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate pagination
	start := page * pageSize
	end := start + pageSize
	if start >= len(sorted) {
		return make(map[string]int), nil
	}
	if end > len(sorted) {
		end = len(sorted)
	}

	// Build result map
	result := make(map[string]int)
	for i := start; i < end; i++ {
		result[sorted[i].key] = sorted[i].value
	}

	return result, nil
}

// GetStatsWithSampling returns statistics using a specific sampling rate
func (f *Filesystem) GetStatsWithSampling(samplingRate float64) (*Stats, error) {
	config := &StatsConfig{
		CacheTTL:                 0, // Don't cache sampled stats
		SamplingThreshold:        0, // Always use sampling
		SamplingRate:             samplingRate,
		UseBackgroundCalculation: false,
	}
	return f.GetStatsWithConfig(config)
}

// GetQuickStats returns a minimal set of statistics quickly without expensive calculations
func (f *Filesystem) GetQuickStats() (*Stats, error) {
	// Get xattr support status
	f.xattrSupportedM.RLock()
	xattrSupported := f.xattrSupported
	f.xattrSupportedM.RUnlock()

	stats := &Stats{
		Expiration:     f.cacheExpirationDays,
		IsOffline:      f.IsOffline(),
		DeltaLink:      f.deltaLink,
		FileExtensions: make(map[string]int),
		FileSizeRanges: make(map[string]int),
		FileAgeRanges:  make(map[string]int),
		CachedAt:       time.Now(),
		IsSampled:      false,
		XAttrSupported: xattrSupported,
	}

	// Count metadata items (fast)
	f.metadata.Range(func(_, _ interface{}) bool {
		stats.MetadataCount++
		return true
	})

	// Get cache size limits from content cache (fast)
	stats.MaxCacheSize = f.content.GetMaxCacheSize()
	currentCacheSize := f.content.GetCacheSize()
	if stats.MaxCacheSize > 0 {
		stats.CacheSizeUsage = (float64(currentCacheSize) / float64(stats.MaxCacheSize)) * 100.0
	} else {
		stats.CacheSizeUsage = -1 // Unlimited
	}

	// Get database statistics (fast)
	if f.db != nil {
		dbPath := f.db.Path()
		stats.DBPath = dbPath

		// Get database file size
		if dbInfo, err := os.Stat(dbPath); err == nil {
			stats.DBSize = dbInfo.Size()
		}

		// Count items in each bucket (fast)
		if err := f.db.View(func(tx *bolt.Tx) error {
			if b := tx.Bucket(bucketMetadataV2); b != nil {
				stats.DBMetadataCount = b.Stats().KeyN
			}
			if b := tx.Bucket(bucketDelta); b != nil {
				stats.DBDeltaCount = b.Stats().KeyN
			}
			if b := tx.Bucket(bucketOfflineChanges); b != nil {
				stats.DBOfflineCount = b.Stats().KeyN
			}
			if b := tx.Bucket(bucketUploads); b != nil {
				stats.DBUploadsCount = b.Stats().KeyN
			}
			return nil
		}); err != nil {
			logging.Error().Err(err).Msg("Error reading database statistics")
		}
	}

	// Upload queue statistics (fast)
	if f.uploads != nil {
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
	}

	// File status statistics (fast)
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

	f.augmentRealtimeStats(stats)
	return stats, nil
}

func (f *Filesystem) augmentRealtimeStats(stats *Stats) {
	if stats == nil {
		return
	}
	stats.RealtimeMode = "disabled"
	stats.RealtimeStatus = socketio.StatusUnknown
	stats.RealtimeLastError = ""
	stats.RealtimeLastHeartbeat = time.Time{}
	stats.RealtimeMissedHeartbeats = 0
	stats.RealtimeConsecutiveFailures = 0
	stats.RealtimeReconnectCount = 0

	if f.realtimeOptions != nil && f.realtimeOptions.Enabled {
		if f.realtimeOptions.PollingOnly {
			stats.RealtimeMode = "polling-only"
		} else {
			stats.RealtimeMode = "socketio"
		}
	}

	if f.subscriptionManager == nil {
		return
	}

	type realtimeModeProvider interface {
		RealtimeMode() string
	}

	type realtimeHealthProvider interface {
		HealthSnapshot() socketio.HealthState
	}

	if modeProvider, ok := f.subscriptionManager.(realtimeModeProvider); ok {
		if mode := modeProvider.RealtimeMode(); mode != "" {
			stats.RealtimeMode = mode
		}
	}

	if healthProvider, ok := f.subscriptionManager.(realtimeHealthProvider); ok {
		health := healthProvider.HealthSnapshot()
		stats.RealtimeStatus = health.Status
		stats.RealtimeMissedHeartbeats = health.MissedHeartbeats
		stats.RealtimeConsecutiveFailures = health.ConsecutiveFailures
		stats.RealtimeLastHeartbeat = health.LastHeartbeat
		stats.RealtimeReconnectCount = health.ReconnectCount
		if health.LastError != nil {
			stats.RealtimeLastError = health.LastError.Error()
		} else {
			stats.RealtimeLastError = ""
		}
		stats.RealtimeRecoveryWindowOpen = isNotifierFailed(health.Status)
		if since := f.notifierRecoverySince.Load(); since != 0 {
			stats.RealtimeRecoverySince = time.Unix(0, since)
		} else {
			stats.RealtimeRecoverySince = time.Time{}
		}
	}
}
