package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/logging"
	"github.com/pkg/errors"

	"github.com/auriora/onemount/internal/graph"
	bolt "go.etcd.io/bbolt"
)

// boltdb buckets
var (
	bucketContent        = []byte("content")
	bucketMetadata       = []byte("metadata")
	bucketDelta          = []byte("delta")
	bucketVersion        = []byte("version")
	bucketOfflineChanges = []byte("offline_changes") // New bucket for offline changes
)

// so we can tell what format the db has
const fsVersion = "1"

// OfflineChange represents a change made while offline
type OfflineChange struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "create", "modify", "delete", "rename", etc.
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path,omitempty"`
	OldPath   string    `json:"old_path,omitempty"` // For rename operations
	NewPath   string    `json:"new_path,omitempty"` // For rename operations
}

// NewFilesystemWithContext creates a new filesystem instance for onemount with a context.
// It initializes the filesystem with the provided authentication, cache directory,
// and cache expiration settings. The function sets up the database, content cache,
// and starts background processes for synchronization and cache management.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - auth: Authentication information for Microsoft Graph API
//   - cacheDir: Directory where filesystem data will be cached
//   - cacheExpirationDays: Number of days after which cached files expire
//   - cacheCleanupIntervalHours: Interval in hours between cache cleanup runs (1-720 hours)
//   - maxCacheSize: Maximum cache size in bytes (0 = unlimited)
//
// Returns:
//   - A new Filesystem instance and nil error on success
//   - nil and an error if initialization fails
func NewFilesystemWithContext(ctx context.Context, auth *graph.Auth, cacheDir string, cacheExpirationDays int, cacheCleanupIntervalHours int, maxCacheSize int64) (*Filesystem, error) {
	// prepare cache directory
	if _, err := os.Stat(cacheDir); err != nil {
		if err = os.Mkdir(cacheDir, 0700); err != nil {
			logging.LogError(err, "Could not create cache directory",
				logging.FieldOperation, "NewFilesystem",
				logging.FieldPath, cacheDir)
			return nil, errors.Wrap(err, "could not create cache directory")
		}
	}
	// Try to open the database with retries and exponential backoff
	var db *bolt.DB
	var err error
	dbPath := filepath.Join(cacheDir, "onemount.db")

	// Check if the database file exists
	if _, statErr := os.Stat(dbPath); statErr == nil {
		// Check for lock files
		lockPath := dbPath + ".lock"
		if _, lockErr := os.Stat(lockPath); lockErr == nil {
			// Check if the lock file is stale by checking its age
			if lockInfo, infoErr := os.Stat(lockPath); infoErr == nil {
				lockAge := time.Since(lockInfo.ModTime())
				if lockAge > 5*time.Minute {
					logging.Warn().Dur("age", lockAge).Msg("Found stale lock file (older than 5 minutes), attempting to remove it")
					if rmErr := os.Remove(lockPath); rmErr != nil {
						logging.Warn().Err(rmErr).Msg("Failed to remove stale lock file")
					} else {
						logging.Info().Msg("Successfully removed stale lock file")
					}
				} else {
					logging.Warn().Dur("age", lockAge).Msg("Found recent lock file, another instance may be running")
				}
			}
		}
	}

	// Define retry parameters
	maxRetries := 10                         // Increased from 5 to 10
	initialBackoff := 200 * time.Millisecond // Increased from 100ms to 200ms
	maxBackoff := 5 * time.Second            // Increased from 2s to 5s

	// Attempt to open the database with retries
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Calculate backoff duration with exponential increase
		backoff := initialBackoff * time.Duration(1<<uint(attempt))
		if backoff > maxBackoff {
			backoff = maxBackoff
		}

		// Try to open the database with increased timeout
		db, err = bolt.Open(
			dbPath,
			0600,
			&bolt.Options{
				Timeout: time.Second * 10, // Increased from 5s to 10s
				// Add NoFreelistSync for better performance
				NoFreelistSync: true,
			},
		)

		if err == nil {
			// Successfully opened the database
			logging.Debug().Int("attempt", attempt+1).Msg("Successfully opened database")
			break
		}

		// If this is the last attempt, don't wait
		if attempt == maxRetries-1 {
			logging.LogError(err, "Could not open DB after multiple attempts",
				logging.FieldOperation, "NewFilesystem",
				logging.FieldPath, dbPath,
				"attempts", maxRetries)
			return nil, errors.Wrap(err, "could not open DB (is it already in use by another mount?)")
		}

		// Log the error and wait before retrying
		logging.Warn().Err(err).Int("attempt", attempt+1).Dur("backoff", backoff).Msg("Failed to open database, retrying after backoff")
		time.Sleep(backoff)
	}

	// If we still have an error after all retries, return it
	if err != nil {
		logging.LogError(err, "Could not open DB",
			logging.FieldOperation, "NewFilesystem",
			logging.FieldPath, dbPath)
		return nil, errors.Wrap(err, "could not open DB (is it already in use by another mount?)")
	}

	// Set up database options for better performance and reliability
	if err := db.Update(func(tx *bolt.Tx) error {
		// Set NoSync option to improve performance (we'll sync manually when needed)
		tx.DB().NoSync = true
		return nil
	}); err != nil {
		logging.Warn().Err(err).Msg("Failed to set database options")
	}

	// Explicitly create content and thumbnail directories
	contentDir := filepath.Join(cacheDir, "content")
	thumbnailDir := filepath.Join(cacheDir, "thumbnails")

	// Create content directory
	if err := os.MkdirAll(contentDir, 0700); err != nil {
		logging.LogError(err, "Could not create content cache directory",
			logging.FieldOperation, "NewFilesystem",
			logging.FieldPath, contentDir)
		return nil, errors.Wrap(err, "could not create content cache directory")
	}

	// Create thumbnail directory
	if err := os.MkdirAll(thumbnailDir, 0700); err != nil {
		logging.LogError(err, "Could not create thumbnail cache directory",
			logging.FieldOperation, "NewFilesystem",
			logging.FieldPath, thumbnailDir)
		return nil, errors.Wrap(err, "could not create thumbnail cache directory")
	}

	content := NewLoopbackCacheWithSize(contentDir, maxCacheSize)
	thumbnails := NewThumbnailCache(thumbnailDir)
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketMetadata); err != nil {
			logging.Error().Err(err).Msg("Failed to create metadata bucket")
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketDelta); err != nil {
			logging.Error().Err(err).Msg("Failed to create delta bucket")
			return err
		}
		versionBucket, err := tx.CreateBucketIfNotExists(bucketVersion)
		if err != nil {
			logging.Error().Err(err).Msg("Failed to create version bucket")
			return err
		}

		// migrate old content bucket to the local filesystem
		b := tx.Bucket(bucketContent)
		if b != nil {
			oldVersion := "0"
			logging.Info().
				Str("oldVersion", oldVersion).
				Str("version", fsVersion).
				Msg("Migrating to new db format.")
			err := b.ForEach(func(k []byte, v []byte) error {
				logging.Info().Str("key", string(k)).Msg("Migrating file content.")
				if err := content.Insert(string(k), v); err != nil {
					return err
				}
				return b.Delete(k)
			})
			if err != nil {
				logging.Error().Err(err).Msg("Migration failed.")
			}
			if err := tx.DeleteBucket(bucketContent); err != nil {
				logging.Error().Err(err).Msg("Failed to delete content bucket during migration")
			}
			logging.Info().
				Str("oldVersion", oldVersion).
				Str("version", fsVersion).
				Msg("Migrations complete.")
		}
		return versionBucket.Put([]byte("version"), []byte(fsVersion))
	})
	if err != nil {
		return nil, err
	}

	// Validate and set cache cleanup interval (default to 24 hours if invalid)
	cleanupInterval := time.Duration(cacheCleanupIntervalHours) * time.Hour
	if cacheCleanupIntervalHours < 1 || cacheCleanupIntervalHours > 720 {
		logging.Warn().
			Int("cacheCleanupIntervalHours", cacheCleanupIntervalHours).
			Msg("Invalid cache cleanup interval, using default of 24 hours")
		cleanupInterval = 24 * time.Hour
	}

	// ok, ready to start fs
	fsCtx, fsCancel := context.WithCancel(ctx)
	deltaCtx, deltaCancel := context.WithCancel(fsCtx)
	fs := &Filesystem{
		content:              content,
		thumbnails:           thumbnails,
		db:                   db,
		auth:                 auth,
		opendirs:             make(map[uint64][]*Inode),
		statuses:             make(map[string]FileStatusInfo),
		statusCache:          newStatusCache(5 * time.Second), // 5 second TTL for status determination cache
		statusCacheTTL:       5 * time.Second,
		ctx:                  fsCtx,
		cancel:               fsCancel,
		cacheExpirationDays:  cacheExpirationDays,
		cacheCleanupInterval: cleanupInterval,
		cacheCleanupStop:     make(chan struct{}),
		deltaLoopStop:        make(chan struct{}),
		deltaLoopCtx:         deltaCtx,
		deltaLoopCancel:      deltaCancel,
		timeoutConfig:        DefaultTimeoutConfig(), // Initialize with default timeout values
	}

	// Initialize with our custom RawFileSystem implementation
	fs.RawFileSystem = NewCustomRawFileSystem(fs)

	// Initialize metadata request manager with 3 workers
	fs.metadataRequestManager = NewMetadataRequestManager(fs, 3)
	fs.metadataRequestManager.Start()

	rootItem, err := graph.GetItem("root", auth)
	root := NewInodeDriveItem(rootItem)
	if err != nil {
		if graph.IsOffline(err) {
			// no network, load from db if possible and go to read-only state
			fs.Lock()
			fs.offline = true
			fs.Unlock()

			// Try to get the root item from the database using the special ID "root"
			root = fs.GetID("root")

			// If that fails, try to find any item in the database that looks like a root folder
			if root == nil {
				// Look for items in the database that have folder properties and no parent
				if err := fs.db.View(func(tx *bolt.Tx) error {
					b := tx.Bucket(bucketMetadata)
					return b.ForEach(func(k, v []byte) error {
						if v != nil {
							inode, err := NewInodeJSON(v)
							if err == nil && inode.IsDir() && inode.ParentID() == "" {
								root = inode
								// Store this item with the special ID "root" for future lookups
								fs.metadata.Store("root", inode)
								return errors.New("found root item, stopping iteration")
							}
						}
						return nil
					})
				}); err != nil && err.Error() != "found root item, stopping iteration" {
					logging.Error().Err(err).Msg("Error searching for root item in database")
				}
			}

			// If we still couldn't find a root item, return an error
			if root == nil {
				logging.Error().Msg(
					"We are offline and could not fetch the filesystem root item from disk.",
				)
				return nil, errors.New("offline and could not fetch the filesystem root item from disk")
			}
			// when offline, we load the cache deltaLink from disk
			var deltaLinkErr error
			if viewErr := fs.db.View(func(tx *bolt.Tx) error {
				if link := tx.Bucket(bucketDelta).Get([]byte("deltaLink")); link != nil {
					fs.deltaLink = string(link)
				} else {
					// Only reached if a previous online session never survived
					// long enough to save its delta link. We explicitly disallow these
					// types of startups as it's possible for things to get out of sync
					// this way.
					logging.Error().Msg("Cannot perform an offline startup without a valid " +
						"delta link from a previous session.")
					deltaLinkErr = errors.New("cannot perform an offline startup without a valid delta link from a previous session")
				}
				return nil
			}); viewErr != nil {
				logging.LogError(viewErr, "Failed to read delta link from database",
					logging.FieldOperation, "NewFilesystem",
					logging.FieldPath, dbPath)
				return nil, errors.Wrap(viewErr, "failed to read delta link from database")
			}
			if deltaLinkErr != nil {
				return nil, deltaLinkErr
			}
		} else {
			logging.LogError(err, "Could not fetch root item of filesystem",
				logging.FieldOperation, "NewFilesystem")
			return nil, errors.Wrap(err, "could not fetch root item of filesystem")
		}
	}
	// root inode is inode 1
	fs.root = root.ID()
	fs.InsertID(fs.root, root)

	fs.uploads = NewUploadManager(2*time.Second, db, fs, auth)

	// Initialize download manager with 4 worker threads and database support
	fs.downloads = NewDownloadManager(fs, auth, 4, db)

	if !fs.IsOffline() {
		// .Trash-UID is used by "gio trash" for user trash, create it if it
		// does not exist
		trash := fmt.Sprintf(".Trash-%d", os.Getuid())
		if child, _ := fs.GetChild(fs.root, trash, auth); child == nil {
			item, err := graph.Mkdir(trash, fs.root, auth)
			if err != nil {
				logging.Error().Err(err).
					Msg("Could not create the trash folder. " +
						"Trashing items through the file browser may result in errors.")
			} else {
				trashInode := NewInodeDriveItem(item)
				fs.InsertID(item.ID, trashInode)

				// Create the required subdirectories for GIO trash
				infoDir := "info"
				filesDir := "files"

				// Create info directory
				if infoChild, _ := fs.GetChild(item.ID, infoDir, auth); infoChild == nil {
					infoItem, err := graph.Mkdir(infoDir, item.ID, auth)
					if err != nil {
						logging.Error().Err(err).Str("dir", infoDir).
							Msg("Could not create trash info directory")
					} else {
						fs.InsertID(infoItem.ID, NewInodeDriveItem(infoItem))
					}
				}

				// Create files directory
				if filesChild, _ := fs.GetChild(item.ID, filesDir, auth); filesChild == nil {
					filesItem, err := graph.Mkdir(filesDir, item.ID, auth)
					if err != nil {
						logging.Error().Err(err).Str("dir", filesDir).
							Msg("Could not create trash files directory")
					} else {
						fs.InsertID(filesItem.ID, NewInodeDriveItem(filesItem))
					}
				}
			}
		}

		// using token=latest because we don't care about existing items - they'll
		// be downloaded on-demand by the cache
		fs.deltaLink = "/me/drive/root/delta?token=latest"
		fs.subscribeChangesLink = "/me/drive/root/subscriptions/socketIo"
	}

	// deltaloop is started manually

	// Initialize D-Bus server
	fs.dbusServer = NewFileStatusDBusServer(fs)
	// Use StartForTesting in test environment
	if err := fs.dbusServer.Start(); err != nil {
		logging.Error().Err(err).Msg("Failed to start D-Bus server")
		// Continue even if D-Bus server fails to start
	}

	return fs, nil
}

// IsOffline returns whether the filesystem is currently in offline mode.
// In offline mode, the filesystem operates without network connectivity,
// using only locally cached content.
//
// Returns:
//   - true if the filesystem is in offline mode
//   - false if the filesystem is in online mode
func (f *Filesystem) IsOffline() bool {
	methodName, startTime := logging.LogMethodEntry("IsOffline")
	f.RLock()
	defer f.RUnlock()

	result := f.offline
	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), result)
	}()
	return result
}

// TrackOfflineChange records a change made while offline
func (f *Filesystem) TrackOfflineChange(change *OfflineChange) error {
	methodName, startTime := logging.LogMethodEntry("TrackOfflineChange", change)
	defer func() {
		// We can't capture the return value directly in a defer, so we'll just log completion
		logging.LogMethodExit(methodName, time.Since(startTime))
	}()

	if !f.IsOffline() {
		return nil // No need to track if we're online
	}

	return f.db.Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketOfflineChanges)
		if err != nil {
			return err
		}

		// Generate a unique key for this change
		key := []byte(fmt.Sprintf("%s-%d", change.ID, change.Timestamp.UnixNano()))

		data, err := json.Marshal(change)
		if err != nil {
			return err
		}

		return b.Put(key, data)
	})
}

// ProcessOfflineChanges processes all changes made while offline
// This is a backward-compatible version that uses a background context
func (f *Filesystem) ProcessOfflineChanges() {
	// Call the context-aware version with a background context
	f.ProcessOfflineChangesWithContext(context.Background())
}

// ProcessOfflineChangesWithSyncManager processes offline changes using the enhanced sync manager
func (f *Filesystem) ProcessOfflineChangesWithSyncManager(ctx context.Context) (*SyncResult, error) {
	syncManager := NewSyncManager(f)
	return syncManager.ProcessOfflineChangesWithRetry(ctx)
}

// getOfflineChanges retrieves all offline changes from the database
func (f *Filesystem) getOfflineChanges(ctx context.Context) ([]*OfflineChange, error) {
	changes := make([]*OfflineChange, 0)

	err := f.db.View(func(tx *bolt.Tx) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue with normal operation
		}

		b := tx.Bucket(bucketOfflineChanges)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			// Check for context cancellation periodically
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Continue with normal operation
			}

			change := &OfflineChange{}
			if err := json.Unmarshal(v, change); err != nil {
				return err
			}
			changes = append(changes, change)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	// Sort changes by timestamp
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Timestamp.Before(changes[j].Timestamp)
	})

	return changes, nil
}

// ProcessOfflineChangesWithContext processes all changes made while offline with context support
// This allows the operation to be cancelled if the filesystem is being shut down
func (f *Filesystem) ProcessOfflineChangesWithContext(goCtx context.Context) {
	// Create a logging context
	ctx := logging.LogContext{
		Operation: "process_offline_changes",
	}

	// Log method entry with context
	methodName, startTime, logger, ctx := logging.LogMethodEntryWithContext("ProcessOfflineChangesWithContext", ctx)
	defer logging.LogMethodExitWithContext(methodName, startTime, logger, ctx)

	logger.Info().Msg("Processing offline changes...")

	// Get all offline changes
	changes := make([]*OfflineChange, 0)
	if err := f.db.View(func(tx *bolt.Tx) error {
		// Check for context cancellation
		select {
		case <-goCtx.Done():
			return goCtx.Err()
		default:
			// Continue with normal operation
		}

		b := tx.Bucket(bucketOfflineChanges)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			// Check for context cancellation periodically
			select {
			case <-goCtx.Done():
				return goCtx.Err()
			default:
				// Continue with normal operation
			}

			change := &OfflineChange{}
			if err := json.Unmarshal(v, change); err != nil {
				return err
			}
			changes = append(changes, change)
			return nil
		})
	}); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			logger.Debug().Msg("Processing offline changes cancelled due to context cancellation")
			return
		}
		logging.LogErrorWithContext(err, ctx, "Failed to read offline changes from database")
		return
	}

	// Sort changes by timestamp
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Timestamp.Before(changes[j].Timestamp)
	})

	// Process each change
	for _, change := range changes {
		// Check for context cancellation before processing each change
		select {
		case <-goCtx.Done():
			logger.Debug().Msg("Processing offline changes cancelled due to context cancellation")
			return
		default:
			// Continue with normal operation
		}

		logger.Info().
			Str("id", change.ID).
			Str("type", change.Type).
			Str("path", change.Path).
			Msg("Processing offline change")

		switch change.Type {
		case "create", "modify":
			// Queue upload with low priority since it's a background task
			if inode := f.GetIDWithContext(change.ID, ctx); inode != nil {
				_, err := f.uploads.QueueUploadWithPriority(inode, PriorityLow)
				if err != nil {
					logging.LogErrorWithContext(err, ctx, "Failed to queue upload for offline change",
						logging.FieldID, change.ID)
				}
			}
		case "delete":
			// Handle deletion
			if !isLocalID(change.ID) {
				if err := graph.Remove(change.ID, f.auth); err != nil {
					logging.LogErrorWithContext(err, ctx, "Failed to remove item during offline change processing",
						logging.FieldID, change.ID)
				}
			}
		case "rename":
			// Handle rename
			if inode := f.GetIDWithContext(change.ID, ctx); inode != nil {
				// Implementation depends on how renames are tracked
				if change.OldPath != "" && change.NewPath != "" {
					oldDir := filepath.Dir(change.OldPath)
					newDir := filepath.Dir(change.NewPath)
					oldName := filepath.Base(change.OldPath)
					newName := filepath.Base(change.NewPath)

					// Get parent IDs
					oldParent, _ := f.GetPath(oldDir, f.auth)
					newParent, _ := f.GetPath(newDir, f.auth)

					if oldParent != nil && newParent != nil {
						if err := f.MovePath(oldParent.ID(), newParent.ID(), oldName, newName, f.auth); err != nil {
							logging.LogErrorWithContext(err, ctx, "Failed to move item during offline change processing",
								logging.FieldID, change.ID,
								"oldPath", change.OldPath,
								"newPath", change.NewPath)
						}
					}
				}
			}
		}

		// Remove the processed change
		if err := f.db.Batch(func(tx *bolt.Tx) error {
			// Check for context cancellation
			select {
			case <-goCtx.Done():
				return goCtx.Err()
			default:
				// Continue with normal operation
			}

			b := tx.Bucket(bucketOfflineChanges)
			if b == nil {
				return nil
			}
			key := []byte(fmt.Sprintf("%s-%d", change.ID, change.Timestamp.UnixNano()))
			return b.Delete(key)
		}); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				logger.Debug().Msg("Removing processed offline change cancelled due to context cancellation")
				return
			}
			logging.LogErrorWithContext(err, ctx, "Failed to remove processed offline change from database",
				logging.FieldID, change.ID,
				"timestamp", change.Timestamp)
		}
	}

	logger.Info().Msg("Finished processing offline changes.")
}

// TranslateID returns the DriveItemID for a given NodeID
func (f *Filesystem) TranslateID(nodeID uint64) string {
	methodName, startTime := logging.LogMethodEntry("TranslateID", nodeID)
	f.RLock()
	defer f.RUnlock()

	var result string
	if nodeID > f.lastNodeID || nodeID == 0 {
		result = ""
	} else {
		result = f.inodes[nodeID-1]
	}

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), result)
	}()
	return result
}

// GetNodeID fetches the inode for a particular inode ID.
func (f *Filesystem) GetNodeID(nodeID uint64) *Inode {
	methodName, startTime := logging.LogMethodEntry("GetNodeID", nodeID)

	id := f.TranslateID(nodeID)
	if id == "" {
		// Log the return value (nil) and return
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), nil)
		}()
		return nil
	}

	result := f.GetID(id)
	// Log the return value (could be nil or a pointer)
	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), result)
	}()
	return result
}

// InsertNodeID assigns a numeric inode ID used by the kernel if one is not
// already assigned.
func (f *Filesystem) InsertNodeID(inode *Inode) uint64 {
	methodName, startTime := logging.LogMethodEntry("InsertNodeID", inode)

	nodeID := inode.NodeID()
	if nodeID == 0 {
		// Lock ordering: inode.mu -> filesystem.RWMutex
		// This violates the standard hierarchy (filesystem before inode) but is safe here
		// because we're only modifying the inode's nodeID field and the filesystem's
		// lastNodeID/inodes, which don't create circular dependencies.
		// See docs/guides/developer/concurrency-guidelines.md for lock ordering policy.
		inode.mu.Lock()
		f.Lock()

		f.lastNodeID++
		f.inodes = append(f.inodes, inode.DriveItem.ID)
		nodeID = f.lastNodeID
		inode.nodeID = nodeID

		f.Unlock()
		inode.mu.Unlock()
	}

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), nodeID)
	}()
	return nodeID
}

// GetID retrieves an inode from the cache by its OneDrive ID.
// This method only checks the in-memory cache and local database; it does not
// perform any API requests to fetch the item from OneDrive.
//
// Parameters:
//   - id: The OneDrive ID of the item to retrieve
//
// Returns:
//   - The Inode if found in memory or database
//   - nil if the item is not found in the cache
func (f *Filesystem) GetID(id string) *Inode {
	methodName, startTime := logging.LogMethodEntry("GetID", id)

	entry, exists := f.metadata.Load(id)
	if !exists {
		// we allow fetching from disk as a fallback while offline (and it's also
		// necessary while transitioning from offline->online)
		var found *Inode
		if err := f.db.View(func(tx *bolt.Tx) error {
			data := tx.Bucket(bucketMetadata).Get([]byte(id))
			var err error
			if data != nil {
				found, err = NewInodeJSON(data)
			}
			return err
		}); err != nil {
			logging.Error().Err(err).Str("id", id).Msg("Failed to read inode from database")
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), nil)
			}()
			return nil
		}
		if found != nil {
			f.InsertNodeID(found)
			f.metadata.Store(id, found) // move to memory for next time
		}
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), found)
		}()
		return found
	}

	result := entry.(*Inode)
	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), result)
	}()
	return result
}

// GetIDWithContext retrieves an inode from the cache by its OneDrive ID with context propagation.
// This method only checks the in-memory cache and local database; it does not
// perform any API requests to fetch the item from OneDrive.
//
// Parameters:
//   - id: The OneDrive ID of the item to retrieve
//   - ctx: The logging context to use
//
// Returns:
//   - The Inode if found in memory or database
//   - nil if the item is not found in the cache
func (f *Filesystem) GetIDWithContext(id string, ctx logging.LogContext) *Inode {
	// Log method entry with context
	methodName, startTime, logger, ctx := logging.LogMethodEntryWithContext("GetIDWithContext", ctx)

	// Call the regular GetID method
	result := f.GetID(id)

	// Log method exit with context
	defer logging.LogMethodExitWithContext(methodName, startTime, logger, ctx, result)
	return result
}

// InsertID adds or updates an item in the filesystem by its OneDrive ID.
// This method stores the inode in the in-memory cache, assigns it a numeric node ID
// for the kernel, and establishes the parent-child relationship if a parent ID is set.
// When used for renaming or moving an item, DeleteID must be called first.
//
// Parameters:
//   - id: The OneDrive ID to associate with the inode
//   - inode: The inode to insert into the filesystem
//
// Returns:
//   - The numeric node ID assigned to the inode for kernel operations
func (f *Filesystem) InsertID(id string, inode *Inode) uint64 {
	methodName, startTime := logging.LogMethodEntry("InsertID", id, inode)

	f.metadata.Store(id, inode)
	nodeID := f.InsertNodeID(inode)

	if id != inode.ID() {
		// Lock ordering: inode.mu first, then filesystem.RWMutex
		// This violates the standard hierarchy but is safe because locks are
		// acquired and released separately (no overlapping lock holds).
		// See docs/guides/developer/concurrency-guidelines.md for lock ordering policy.
		inode.mu.Lock()
		inode.DriveItem.ID = id
		inode.mu.Unlock()

		f.Lock()
		if nodeID <= f.lastNodeID {
			f.inodes[nodeID-1] = id
		} else {
			logging.Error().
				Uint64("nodeID", nodeID).
				Uint64("lastNodeID", f.lastNodeID).
				Msg("NodeID exceeded maximum node ID! Ignoring ID change.")
		}
		f.Unlock()
	}

	parentID := inode.ParentID()
	if parentID == "" {
		// root item, or parent not set
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), nodeID)
		}()
		return nodeID
	}
	parent := f.GetID(parentID)
	if parent == nil {
		// Check if the parent ID is the root ID
		if parentID == f.root {
			// Create a dummy root item if it doesn't exist
			logging.Warn().
				Str("parentID", parentID).
				Str("childID", id).
				Str("childName", inode.Name()).
				Msg("Root item not found in cache, creating dummy root item.")

			// Create a dummy root item
			rootItem := &graph.DriveItem{
				ID:   parentID,
				Name: "root",
				Folder: &graph.Folder{
					ChildCount: 0,
				},
			}
			rootInode := NewInodeDriveItem(rootItem)

			// Insert the root item into the cache
			f.metadata.Store(parentID, rootInode)
			parent = rootInode
		} else {
			logging.Error().
				Str("parentID", parentID).
				Str("childID", id).
				Str("childName", inode.Name()).
				Msg("Parent item could not be found when setting parent.")
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), nodeID)
			}()
			return nodeID
		}
	}

	// check if the item has already been added to the parent
	// Lock ordering: parent inode before child inode (when both needed)
	// For multiple inodes at same level, use ID-based ordering.
	// See docs/guides/developer/concurrency-guidelines.md for lock ordering policy.
	parent.mu.Lock()
	defer parent.mu.Unlock()
	for _, child := range parent.children {
		if child == id {
			// exit early, child cannot be added twice
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), nodeID)
			}()
			return nodeID
		}
	}

	// add to parent
	if inode.IsDir() {
		parent.subdir++
	}
	parent.children = append(parent.children, id)

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), nodeID)
	}()
	return nodeID
}

// InsertChild adds an item as a child of a specified parent ID.
// Lock ordering: child inode only (parent locked in InsertID)
// See docs/guides/developer/concurrency-guidelines.md for lock ordering policy.
func (f *Filesystem) InsertChild(parentID string, child *Inode) uint64 {
	child.mu.Lock()
	// Initialize Parent if it's nil to avoid nil pointer dereference
	if child.DriveItem.Parent == nil {
		child.DriveItem.Parent = &graph.DriveItemParent{}
	}
	// should already be set, just double-checking here.
	child.DriveItem.Parent.ID = parentID
	id := child.DriveItem.ID
	child.mu.Unlock()
	return f.InsertID(id, child)
}

// DeleteID deletes an item from the cache, and removes it from its parent. Must
// be called before InsertID if being used to rename/move an item.
func (f *Filesystem) DeleteID(id string) {
	if inode := f.GetID(id); inode != nil {
		// If this is a directory, recursively delete all its children first
		if inode.IsDir() && inode.HasChildren() {
			// Make a copy of the children slice to avoid concurrent modification issues
			inode.mu.RLock()
			childrenCopy := make([]string, len(inode.children))
			copy(childrenCopy, inode.children)
			inode.mu.RUnlock()

			// Delete each child
			for _, childID := range childrenCopy {
				f.DeleteID(childID)
			}
		}

		// Lock ordering: parent inode only (child already processed)
		// See docs/guides/developer/concurrency-guidelines.md
		parent := f.GetID(inode.ParentID())
		parent.mu.Lock()
		for i, childID := range parent.children {
			if childID == id {
				parent.children = append(parent.children[:i], parent.children[i+1:]...)
				if inode.IsDir() {
					parent.subdir--
				}
				break
			}
		}
		parent.mu.Unlock()
	}
	f.metadata.Delete(id)
	f.uploads.CancelUpload(id)
}

// GetChild fetches a named child of an item. Wraps GetChildrenID.
func (f *Filesystem) GetChild(id string, name string, auth *graph.Auth) (*Inode, error) {
	children, err := f.GetChildrenID(id, auth)
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if strings.EqualFold(child.Name(), name) {
			return child, nil
		}
	}
	return nil, errors.New("child does not exist")
}

// GetChildrenID grabs all DriveItems that are the children of the given ID. If
// items are not found, they are fetched.
func (f *Filesystem) GetChildrenID(id string, auth *graph.Auth) (map[string]*Inode, error) {
	methodName, startTime := logging.LogMethodEntry("GetChildrenID", id)

	// Create a context for this operation with request ID and user ID
	ctx := logging.NewLogContextWithRequestAndUserID("get_children")

	logger := logging.WithLogContext(ctx)

	// fetch item and catch common errors
	inode := f.GetID(id)
	children := make(map[string]*Inode)
	if inode == nil {
		logger.Error().Str(logging.FieldID, id).Msg("Inode not found in cache")
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), children, errors.New(id+" not found in cache"))
		}()
		return children, errors.New(id + " not found in cache")
	} else if !inode.IsDir() {
		// Normal files are treated as empty folders. This only gets called if
		// we messed up and tried to get the children of a plain-old file.
		logger.Warn().
			Str(logging.FieldID, id).
			Str(logging.FieldPath, inode.Path()).
			Msg("Attempted to get children of ordinary file")
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), children, nil)
		}()
		return children, nil
	}

	// Get the path before acquiring any locks to avoid potential deadlocks
	pathForLogs := inode.Path()

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldID, id).
			Str(logging.FieldPath, pathForLogs).
			Msg("Checking if children are already cached")
	}

	// If item.children is not nil, it means we have the item's children
	// already and can fetch them directly from the cache
	inode.mu.RLock()
	if inode.children != nil {
		if logging.IsDebugEnabled() {
			logger.Debug().
				Str(logging.FieldID, id).
				Str(logging.FieldPath, pathForLogs).
				Int("childCount", len(inode.children)).
				Msg("Children found in cache, retrieving them")
		}

		// can potentially have out-of-date child metadata if started offline, but since
		// changes are disallowed while offline, the children will be back in sync after
		// the first successful delta fetch (which also brings the fs back online)
		for _, childID := range inode.children {
			child := f.GetID(childID)
			if child == nil {
				// will be nil if deleted or never existed
				continue
			}
			children[strings.ToLower(child.Name())] = child
		}
		inode.mu.RUnlock()

		if logging.IsDebugEnabled() {
			logger.Debug().
				Str(logging.FieldID, id).
				Str(logging.FieldPath, pathForLogs).
				Int("childCount", len(children)).
				Msg("Successfully retrieved children from cache")
		}

		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), children, nil)
		}()
		return children, nil
	}
	// Update path before unlocking to avoid potential deadlocks
	pathForLogs = inode.Path()
	inode.mu.RUnlock()

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldID, id).
			Str(logging.FieldPath, pathForLogs).
			Msg("Children not in cache, fetching from server")
	}

	// We haven't fetched the children for this item yet, get them from the server.
	// Use prioritized metadata request for foreground operations
	var fetched []*graph.DriveItem
	var err error

	if f.metadataRequestManager != nil {
		// Create a channel to receive the result
		resultChan := make(chan struct {
			items []*graph.DriveItem
			err   error
		}, 1)

		// Queue the metadata request with foreground priority
		reqErr := f.metadataRequestManager.QueueChildrenRequest(id, auth, PriorityForeground, func(items []*graph.DriveItem, reqErr error) {
			resultChan <- struct {
				items []*graph.DriveItem
				err   error
			}{items, reqErr}
		})

		if reqErr != nil {
			// Fallback to direct call if queue is full
			if logging.IsDebugEnabled() {
				logger.Debug().
					Str(logging.FieldID, id).
					Str(logging.FieldPath, pathForLogs).
					Msg("Metadata queue full, falling back to direct call")
			}
			fetched, err = graph.GetItemChildren(id, auth)
		} else {
			// Wait for the result with timeout
			select {
			case result := <-resultChan:
				fetched = result.items
				err = result.err
			case <-time.After(30 * time.Second):
				err = context.DeadlineExceeded
				logger.Warn().
					Str(logging.FieldID, id).
					Str(logging.FieldPath, pathForLogs).
					Msg("Foreground metadata request timed out, falling back to direct call")
				fetched, err = graph.GetItemChildren(id, auth)
			}
		}
	} else {
		// Fallback if metadata request manager is not available
		if logging.IsDebugEnabled() {
			logger.Debug().
				Str(logging.FieldID, id).
				Str(logging.FieldPath, pathForLogs).
				Msg("About to call graph.GetItemChildren (no metadata manager)")
		}
		fetched, err = graph.GetItemChildren(id, auth)
	}

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldID, id).
			Str(logging.FieldPath, pathForLogs).
			Err(err).
			Int("itemCount", len(fetched)).
			Msg("Completed metadata request")
	}

	if err != nil {
		if graph.IsOffline(err) {
			logger.Warn().
				Str(logging.FieldID, id).
				Msg("We are offline, and no children found in cache. " +
					"Pretending there are no children.")
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), children, nil)
			}()
			return children, nil
		}
		// something else happened besides being offline
		logging.LogErrorWithContext(err, ctx, "Error fetching children from server",
			logging.FieldID, id,
			logging.FieldPath, pathForLogs)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), nil, err)
		}()
		return nil, err
	}

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldID, id).
			Str(logging.FieldPath, pathForLogs).
			Int("fetchedCount", len(fetched)).
			Msg("Processing fetched children")
	}

	// Store the path before locking to avoid potential deadlocks
	processingPath := pathForLogs

	inode.mu.Lock()
	inode.children = make([]string, 0)
	for i, item := range fetched {
		if strings.EqualFold(item.Name, xdgVolumeInfoName) {
			if logging.IsDebugEnabled() {
				logger.Debug().
					Str(logging.FieldID, item.ID).
					Str(logging.FieldPath, processingPath).
					Msg("Skipping remote .xdg-volume-info entry in favor of virtual file")
			}
			continue
		}
		// we will always have an id after fetching from the server
		child := NewInodeDriveItem(item)
		f.InsertNodeID(child)
		f.metadata.Store(child.DriveItem.ID, child)

		// store in result map
		children[strings.ToLower(child.Name())] = child

		// store id in parent item and increment parents subdirectory count
		inode.children = append(inode.children, child.DriveItem.ID)
		if child.IsDir() {
			inode.subdir++
		}

		if logging.IsDebugEnabled() && i%50 == 0 && i > 0 {
			logger.Debug().
				Str(logging.FieldID, id).
				Str(logging.FieldPath, processingPath).
				Int("processedCount", i).
				Int("totalCount", len(fetched)).
				Msg("Processing children progress")
		}
	}

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldID, id).
			Str(logging.FieldPath, processingPath).
			Int("childrenCount", len(children)).
			Uint64("subdirCount", uint64(inode.subdir)).
			Msg("Finished processing all children")
	}

	inode.mu.Unlock()

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldID, id).
			Str(logging.FieldPath, processingPath).
			Int("childrenCount", len(children)).
			Msg("GetChildrenID completed successfully")
	}

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), children, nil)
	}()
	return children, nil
}

// GetChildrenPath grabs all DriveItems that are the children of the resource at
// the path. If items are not found, they are fetched.
func (f *Filesystem) GetChildrenPath(path string, auth *graph.Auth) (map[string]*Inode, error) {
	methodName, startTime := logging.LogMethodEntry("GetChildrenPath", path)

	// Create a context for this operation with request ID and user ID
	ctx := logging.NewLogContextWithRequestAndUserID("get_children_path").
		WithPath(path)

	logger := logging.WithLogContext(ctx)

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldPath, path).
			Msg("Getting children for path")
	}

	inode, err := f.GetPath(path, auth)
	if err != nil {
		logging.LogErrorWithContext(err, ctx, "Error getting path",
			logging.FieldPath, path)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), nil, err)
		}()
		return make(map[string]*Inode), err
	}

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldPath, path).
			Str(logging.FieldID, inode.ID()).
			Bool("isDir", inode.IsDir()).
			Msg("Found path, getting children")
	}

	children, err := f.GetChildrenID(inode.ID(), auth)
	if err != nil {
		logging.LogErrorWithContext(err, ctx, "Error getting children for path",
			logging.FieldPath, path,
			logging.FieldID, inode.ID())
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), nil, err)
		}()
		return make(map[string]*Inode), err
	}

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldPath, path).
			Str(logging.FieldID, inode.ID()).
			Int("childCount", len(children)).
			Msg("Successfully got children for path")
	}

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), children, nil)
	}()
	return children, nil
}

// GetPath fetches a given DriveItem in the cache, if any items along the way are
// not found, they are fetched.
func (f *Filesystem) GetPath(path string, auth *graph.Auth) (*Inode, error) {
	methodName, startTime := logging.LogMethodEntry("GetPath", path)

	// Create a context for this operation with request ID and user ID
	ctx := logging.NewLogContextWithRequestAndUserID("get_path").
		WithPath(path)

	logger := logging.WithLogContext(ctx)

	lastID := f.root
	if path == "/" {
		result := f.GetID(lastID)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), result, nil)
		}()
		return result, nil
	}

	// from the root directory, traverse the chain of items till we reach our
	// target ID.
	path = strings.TrimSuffix(strings.ToLower(path), "/")
	split := strings.Split(path, "/")[1:] //omit leading "/"

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldPath, path).
			Strs("pathComponents", split).
			Msg("Traversing path components")
	}

	var inode *Inode
	for i := 0; i < len(split); i++ {
		// fetches children
		if logging.IsDebugEnabled() {
			logger.Debug().
				Str(logging.FieldID, lastID).
				Str("component", split[i]).
				Int("componentIndex", i).
				Msg("Fetching children for path component")
		}

		children, err := f.GetChildrenID(lastID, auth)
		if err != nil {
			logging.LogErrorWithContext(err, ctx, "Error fetching children for path component",
				logging.FieldID, lastID,
				logging.FieldPath, path,
				"component", split[i],
				"componentIndex", i)
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), nil, err)
			}()
			return nil, err
		}

		var exists bool // if we use ":=", item is shadowed
		inode, exists = children[split[i]]
		if !exists {
			// the item still doesn't exist after fetching from server. it
			// doesn't exist
			errMsg := strings.Join(split[:i+1], "/") + " does not exist on server or in local cache"
			err := errors.New(errMsg)
			logging.LogErrorWithContext(err, ctx, "Path component not found",
				logging.FieldPath, path,
				"component", split[i],
				"componentIndex", i)
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), nil, err)
			}()
			return nil, err
		}

		lastID = inode.ID()

		if logging.IsDebugEnabled() {
			logger.Debug().
				Str(logging.FieldID, lastID).
				Str("component", split[i]).
				Int("componentIndex", i).
				Bool("isDir", inode.IsDir()).
				Msg("Found path component")
		}
	}

	if logging.IsDebugEnabled() {
		logger.Debug().
			Str(logging.FieldPath, path).
			Str(logging.FieldID, inode.ID()).
			Bool("isDir", inode.IsDir()).
			Msg("Successfully found path")
	}

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), inode, nil)
	}()
	return inode, nil
}

// DeletePath an item from the cache by path. Must be called before Insert if
// being used to move/rename an item.
func (f *Filesystem) DeletePath(key string) {
	inode, _ := f.GetPath(strings.ToLower(key), nil)
	if inode != nil {
		f.DeleteID(inode.ID())
	}
}

// InsertPath lets us manually insert an item to the cache (like if it was
// created locally). Overwrites a cached item if present. Must be called after
// delete if being used to move/rename an item.
func (f *Filesystem) InsertPath(key string, auth *graph.Auth, inode *Inode) (uint64, error) {
	key = strings.ToLower(key)

	// set the item.Parent.ID properly if the item hasn't been in the cache
	// before or is being moved.
	parent, err := f.GetPath(filepath.Dir(key), auth)
	if err != nil {
		return 0, err
	} else if parent == nil {
		const errMsg string = "parent of key was nil"
		logging.Error().
			Str("key", key).
			Str("path", inode.Path()).
			Msg(errMsg)
		return 0, errors.New(errMsg)
	}

	// Coded this way to make sure locks are in the same order for the deadlock
	// detector (lock ordering needs to be the same as InsertID: Parent->Child).
	parentID := parent.ID()
	inode.mu.Lock()
	inode.DriveItem.Parent.ID = parentID
	id := inode.DriveItem.ID
	inode.mu.Unlock()

	return f.InsertID(id, inode), nil
}

// MoveID moves an item to a new ID name. Also responsible for handling the
// actual overwrite of the item's IDInternal field
func (f *Filesystem) MoveID(oldID string, newID string) error {
	inode := f.GetID(oldID)
	if inode == nil {
		// It may have already been renamed. This is not an error. We assume
		// that IDs will never collide. Re-perform the op if this is the case.
		if inode = f.GetID(newID); inode == nil {
			// nope, it just doesn't exist
			return errors.New("Could not get item: " + oldID)
		}
	}

	// need to rename the child under the parent
	// Lock ordering: parent inode only (child not locked here)
	// See docs/guides/developer/concurrency-guidelines.md
	parent := f.GetID(inode.ParentID())
	parent.mu.Lock()
	for i, child := range parent.children {
		if child == oldID {
			parent.children[i] = newID
			break
		}
	}
	parent.mu.Unlock()

	// now actually perform the metadata+content move
	f.DeleteID(oldID)
	f.InsertID(newID, inode)
	if inode.IsDir() {
		return nil
	}
	if err := f.content.Move(oldID, newID); err != nil {
		logging.LogError(err, "Failed to move file content",
			logging.FieldOperation, "MoveID",
			logging.FieldID, oldID,
			"newID", newID)
		return errors.Wrap(err, "failed to move file content")
	}
	return nil
}

// MovePath moves an item to a new position.
func (f *Filesystem) MovePath(oldParent, newParent, oldName, newName string, auth *graph.Auth) error {
	inode, err := f.GetChild(oldParent, oldName, auth)
	if err != nil {
		return err
	}

	id := inode.ID()
	f.DeleteID(id)

	// this is the actual move op
	inode.SetName(newName)
	parent := f.GetID(newParent)
	inode.DriveItem.Parent.ID = parent.DriveItem.ID
	f.InsertID(id, inode)
	return nil
}

// StartCacheCleanup starts a background goroutine that periodically cleans up
// the content cache by removing files that haven't been modified for the specified
// number of days. The cleanup runs at the configured interval.
func (f *Filesystem) StartCacheCleanup() {
	// Don't start cleanup if expiration days is 0 or negative
	if f.cacheExpirationDays <= 0 {
		logging.Info().Msg("Cache cleanup disabled (expiration days <= 0)")
		return
	}

	logging.Info().
		Int("expirationDays", f.cacheExpirationDays).
		Dur("cleanupInterval", f.cacheCleanupInterval).
		Msg("Starting content cache cleanup routine")

	// Add to wait group to track this goroutine
	f.cacheCleanupWg.Add(1)
	f.Wg.Add(1)

	// Run cleanup in a goroutine
	go func() {
		defer f.cacheCleanupWg.Done()
		defer f.Wg.Done()

		// Run cleanup immediately on startup
		count, err := f.content.CleanupCache(f.cacheExpirationDays)
		if err != nil {
			logging.Error().Err(err).Msg("Error during initial content cache cleanup")
		} else {
			logging.Info().Int("removedFiles", count).Msg("Initial content cache cleanup completed")
		}

		// Set up ticker for periodic cleanup using configured interval
		ticker := time.NewTicker(f.cacheCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Run cleanup
				count, err := f.content.CleanupCache(f.cacheExpirationDays)
				if err != nil {
					logging.Error().Err(err).Msg("Error during content cache cleanup")
				} else {
					logging.Info().Int("removedFiles", count).Msg("Content cache cleanup completed")
				}
			case <-f.cacheCleanupStop:
				// Stop the cleanup routine
				logging.Info().Msg("Stopping content cache cleanup routine via stop channel")
				return
			case <-f.ctx.Done():
				// Context cancelled, stop the cleanup routine
				logging.Info().Msg("Stopping content cache cleanup routine via context cancellation")
				return
			}
		}
	}()
}

// StopCacheCleanup stops the background cache cleanup routine.
func (f *Filesystem) StopCacheCleanup() {
	logging.Info().Msg("Stopping cache cleanup routine...")
	// Only send stop signal if expiration days is positive (cleanup is running)
	if f.cacheExpirationDays > 0 {
		f.cacheCleanupStopOnce.Do(func() {
			close(f.cacheCleanupStop)
		})
		f.cacheCleanupWg.Wait()
		logging.Info().Msg("Cache cleanup routine stopped")
	}
}

// StopDeltaLoop stops the delta loop goroutine and waits for it to finish.
func (f *Filesystem) StopDeltaLoop() {
	logging.Info().Msg("Stopping delta loop...")

	// Cancel the context to interrupt any in-progress network requests
	f.deltaLoopCancel()
	logging.Debug().Msg("Cancelled delta loop context to interrupt network operations")

	// Close the stop channel to signal the delta loop to stop
	f.deltaLoopStopOnce.Do(func() {
		close(f.deltaLoopStop)
	})
	logging.Debug().Msg("Closed delta loop stop channel")

	// Wait for delta loop to finish with a timeout
	done := make(chan struct{})
	go func() {
		f.deltaLoopWg.Wait()
		close(done)
	}()

	// Wait for delta loop to finish or timeout after 10 seconds
	select {
	case <-done:
		logging.Info().Msg("Delta loop stopped successfully")
	case <-time.After(10 * time.Second):
		logging.Warn().Msg("Timed out waiting for delta loop to stop - continuing shutdown anyway")
		// Log additional debug information
		logging.Debug().Msg("Delta loop may be stuck in a network operation or processing a large batch of changes")
		logging.Debug().Msg("This is not a critical error, but may indicate a potential issue with network operations")
	}
}

// StopDownloadManager stops the download manager and waits for all workers to finish.
func (f *Filesystem) StopDownloadManager() {
	logging.Info().Msg("Stopping download manager...")
	if f.downloads != nil {
		// Create a channel to signal when the download manager has stopped
		done := make(chan struct{})

		// Start a goroutine to call Stop and signal when done
		go func() {
			f.downloads.Stop()
			close(done)
		}()

		// Wait for download manager to stop or timeout after 5 seconds
		select {
		case <-done:
			logging.Info().Msg("Download manager stopped successfully")
		case <-time.After(5 * time.Second):
			logging.Warn().Msg("Timed out waiting for download manager to stop")
		}
	}
}

// StopUploadManager stops the upload manager and waits for all uploads to finish.
func (f *Filesystem) StopUploadManager() {
	logging.Info().Msg("Stopping upload manager...")
	if f.uploads != nil {
		// Create a channel to signal when the upload manager has stopped
		done := make(chan struct{})

		// Start a goroutine to call Stop and signal when done
		go func() {
			f.uploads.Stop()
			close(done)
		}()

		// Wait for upload manager to stop or timeout after 5 seconds
		select {
		case <-done:
			logging.Info().Msg("Upload manager stopped successfully")
		case <-time.After(5 * time.Second):
			logging.Warn().Msg("Timed out waiting for upload manager to stop")
		}
	}
}

// StopMetadataRequestManager stops the metadata request manager and waits for all workers to finish.
func (f *Filesystem) StopMetadataRequestManager() {
	logging.Info().Msg("Stopping metadata request manager...")
	if f.metadataRequestManager != nil {
		// Create a channel to signal when the metadata request manager has stopped
		done := make(chan struct{})

		// Start a goroutine to call Stop and signal when done
		go func() {
			f.metadataRequestManager.Stop()
			close(done)
		}()

		// Wait for metadata request manager to stop or timeout after 5 seconds
		select {
		case <-done:
			logging.Info().Msg("Metadata request manager stopped successfully")
		case <-time.After(5 * time.Second):
			logging.Warn().Msg("Timed out waiting for metadata request manager to stop")
		}
	}
}

// Stop stops all background processes and cleans up the filesystem.
// This method should be called when the filesystem is no longer needed,
// especially in tests to prevent goroutine leaks.
func (f *Filesystem) Stop() {
	// Use a sync.Once to ensure Stop is only called once
	f.stopOnce.Do(func() {
		logging.Info().Msg("Stopping filesystem and all background processes...")

		// Cancel the root context to signal all operations to stop
		if f.cancel != nil {
			f.cancel()
		}

		// Stop all background processes in the correct order
		f.StopCacheCleanup()
		f.StopDeltaLoop()
		f.StopDownloadManager()
		f.StopUploadManager()
		f.StopMetadataRequestManager()

		// Stop the D-Bus server if it exists
		if f.dbusServer != nil {
			f.dbusServer.Stop()
		}

		// Wait for all goroutines to finish with a timeout
		done := make(chan struct{})
		go func() {
			f.Wg.Wait()
			close(done)
		}()

		// Get timeout from configuration
		timeout := 10 * time.Second // Default fallback
		if f.timeoutConfig != nil {
			timeout = f.timeoutConfig.FilesystemShutdown
		}

		// Wait for all goroutines to finish or timeout
		select {
		case <-done:
			logging.Info().Msg("All filesystem goroutines stopped successfully")
		case <-time.After(timeout):
			logging.Warn().
				Dur("timeout", timeout).
				Msg("Timed out waiting for filesystem goroutines to stop")
		}

		// Close the database connection
		if f.db != nil {
			if err := f.db.Close(); err != nil {
				logging.Warn().Err(err).Msg("Failed to close database connection")
			}
		}

		logging.Info().Msg("Filesystem stopped successfully")
	})
}

// GetSyncProgress returns the current sync progress, if available
func (f *Filesystem) GetSyncProgress() *SyncProgress {
	f.RLock()
	defer f.RUnlock()
	return f.syncProgress
}

// SerializeAll dumps all inode metadata currently in the cache to disk. This
// metadata is only used later if an item could not be found in memory AND the
// cache is offline. Old metadata is not removed, only overwritten (to avoid an
// offline session from wiping all metadata on a subsequent serialization).
func (f *Filesystem) SerializeAll() {
	logging.Debug().Msg("Serializing cache metadata to disk.")

	allItems := make(map[string][]byte)
	f.metadata.Range(func(k interface{}, v interface{}) bool {
		// cannot occur within bolt transaction because acquiring the inode lock
		// with AsJSON locks out other boltdb transactions
		id := fmt.Sprint(k)
		allItems[id] = v.(*Inode).AsJSON()
		return true
	})

	/*
		One transaction to serialize them all,
		One transaction to find them,
		One transaction to bring them all
		and in the darkness write them.
	*/
	if err := f.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMetadata)
		for k, v := range allItems {
			if err := b.Put([]byte(k), v); err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to put item %s", k))
			}
			if k == f.root {
				// root item must be updated manually (since there's actually
				// two copies)
				if err := b.Put([]byte("root"), v); err != nil {
					return errors.Wrap(err, "failed to put root item")
				}
			}
		}
		return nil
	}); err != nil {
		logging.Error().Err(err).Msg("Failed to serialize metadata to database")
	}
}

// NewFilesystem is provided for backward compatibility with existing tests and should not be used in new code.
//
// Parameters:
//   - auth: Authentication information for Microsoft Graph API
//   - cacheDir: Directory where filesystem data will be cached
//   - cacheExpirationDays: Number of days after which cached files expire
//
// Returns:
//   - A new Filesystem instance and nil error on success
//   - nil and an error if initialization fails
func NewFilesystem(auth *graph.Auth, cacheDir string, cacheExpirationDays int) (*Filesystem, error) {
	// Create a background context for backward compatibility
	ctx := context.Background()
	// Use default cleanup interval of 24 hours and unlimited cache size for backward compatibility
	return NewFilesystemWithContext(ctx, auth, cacheDir, cacheExpirationDays, 24, 0)
}
