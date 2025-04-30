package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// Filesystem is the actual FUSE filesystem implementation for onemount.
// It provides a native Linux filesystem for Microsoft OneDrive using the
// "low-level" FUSE API (https://github.com/libfuse/libfuse/blob/master/include/fuse_lowlevel.h).
// The Filesystem handles file operations, caching, synchronization with OneDrive,
// and offline mode functionality.
type Filesystem struct {
	fuse.RawFileSystem // Implements the base FUSE filesystem interface

	metadata             sync.Map        // In-memory cache of filesystem metadata
	db                   *bolt.DB        // Persistent database for filesystem state
	content              *LoopbackCache  // Cache for file contents
	thumbnails           *ThumbnailCache // Cache for file thumbnails
	auth                 *graph.Auth     // Authentication for Microsoft Graph API
	root                 string          // The ID of the filesystem's root item
	deltaLink            string          // Link for incremental synchronization with OneDrive
	subscribeChangesLink string
	uploads              *UploadManager   // Manages file uploads to OneDrive
	downloads            *DownloadManager // Manages file downloads from OneDrive

	// Cache cleanup configuration
	cacheExpirationDays  int            // Number of days after which cached files expire
	cacheCleanupStop     chan struct{}  // Channel to signal cache cleanup to stop
	cacheCleanupStopOnce sync.Once      // Ensures cleanup is stopped only once
	cacheCleanupWg       sync.WaitGroup // Wait group for cache cleanup goroutine

	// DeltaLoop stop channel and context
	deltaLoopStop     chan struct{}      // Channel to signal delta loop to stop
	deltaLoopWg       sync.WaitGroup     // Wait group for delta loop goroutine
	deltaLoopStopOnce sync.Once          // Ensures delta loop is stopped only once
	deltaLoopCtx      context.Context    // Context for delta loop cancellation
	deltaLoopCancel   context.CancelFunc // Function to cancel delta loop context

	sync.RWMutex          // Mutex for filesystem state
	offline      bool     // Whether the filesystem is in offline mode
	lastNodeID   uint64   // Last assigned node ID
	inodes       []string // List of inode IDs

	// Tracks currently open directories
	opendirsM sync.RWMutex        // Mutex for open directories map
	opendirs  map[uint64][]*Inode // Map of open directories by node ID

	// Track file statuses
	statusM  sync.RWMutex              // Mutex for file statuses map
	statuses map[string]FileStatusInfo // Map of file statuses by ID

	// D-Bus server for file status updates
	dbusServer *FileStatusDBusServer
}

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

// NewFilesystem creates a new filesystem instance for onemount.
// It initializes the filesystem with the provided authentication, cache directory,
// and cache expiration settings. The function sets up the database, content cache,
// and starts background processes for synchronization and cache management.
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
	// prepare cache directory
	if _, err := os.Stat(cacheDir); err != nil {
		if err = os.Mkdir(cacheDir, 0700); err != nil {
			log.Error().Err(err).Msg("Could not create cache directory.")
			return nil, fmt.Errorf("could not create cache directory: %w", err)
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
					log.Warn().Dur("age", lockAge).Msg("Found stale lock file (older than 5 minutes), attempting to remove it")
					if rmErr := os.Remove(lockPath); rmErr != nil {
						log.Warn().Err(rmErr).Msg("Failed to remove stale lock file")
					} else {
						log.Info().Msg("Successfully removed stale lock file")
					}
				} else {
					log.Warn().Dur("age", lockAge).Msg("Found recent lock file, another instance may be running")
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
			log.Debug().Int("attempt", attempt+1).Msg("Successfully opened database")
			break
		}

		// If this is the last attempt, don't wait
		if attempt == maxRetries-1 {
			log.Error().Err(err).Int("attempts", maxRetries).Msg("Could not open DB after multiple attempts. Is it already in use by another mount?")
			return nil, fmt.Errorf("could not open DB (is it already in use by another mount?): %w", err)
		}

		// Log the error and wait before retrying
		log.Warn().Err(err).Int("attempt", attempt+1).Dur("backoff", backoff).Msg("Failed to open database, retrying after backoff")
		time.Sleep(backoff)
	}

	// If we still have an error after all retries, return it
	if err != nil {
		log.Error().Err(err).Msg("Could not open DB. Is it already in use by another mount?")
		return nil, fmt.Errorf("could not open DB (is it already in use by another mount?): %w", err)
	}

	// Set up database options for better performance and reliability
	if err := db.Update(func(tx *bolt.Tx) error {
		// Set NoSync option to improve performance (we'll sync manually when needed)
		tx.DB().NoSync = true
		return nil
	}); err != nil {
		log.Warn().Err(err).Msg("Failed to set database options")
	}

	// Explicitly create content and thumbnail directories
	contentDir := filepath.Join(cacheDir, "content")
	thumbnailDir := filepath.Join(cacheDir, "thumbnails")

	// Create content directory
	if err := os.MkdirAll(contentDir, 0700); err != nil {
		log.Error().Err(err).Msg("Could not create content cache directory.")
		return nil, fmt.Errorf("could not create content cache directory: %w", err)
	}

	// Create thumbnail directory
	if err := os.MkdirAll(thumbnailDir, 0700); err != nil {
		log.Error().Err(err).Msg("Could not create thumbnail cache directory.")
		return nil, fmt.Errorf("could not create thumbnail cache directory: %w", err)
	}

	content := NewLoopbackCache(contentDir)
	thumbnails := NewThumbnailCache(thumbnailDir)
	db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketMetadata); err != nil {
			log.Error().Err(err).Msg("Failed to create metadata bucket")
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketDelta); err != nil {
			log.Error().Err(err).Msg("Failed to create delta bucket")
			return err
		}
		versionBucket, err := tx.CreateBucketIfNotExists(bucketVersion)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create version bucket")
			return err
		}

		// migrate old content bucket to the local filesystem
		b := tx.Bucket(bucketContent)
		if b != nil {
			oldVersion := "0"
			log.Info().
				Str("oldVersion", oldVersion).
				Str("version", fsVersion).
				Msg("Migrating to new db format.")
			err := b.ForEach(func(k []byte, v []byte) error {
				log.Info().Bytes("key", k).Msg("Migrating file content.")
				if err := content.Insert(string(k), v); err != nil {
					return err
				}
				return b.Delete(k)
			})
			if err != nil {
				log.Error().Err(err).Msg("Migration failed.")
			}
			if err := tx.DeleteBucket(bucketContent); err != nil {
				log.Error().Err(err).Msg("Failed to delete content bucket during migration")
			}
			log.Info().
				Str("oldVersion", oldVersion).
				Str("version", fsVersion).
				Msg("Migrations complete.")
		}
		return versionBucket.Put([]byte("version"), []byte(fsVersion))
	})

	// ok, ready to start fs
	ctx, cancel := context.WithCancel(context.Background())
	fs := &Filesystem{
		RawFileSystem:       fuse.NewDefaultRawFileSystem(),
		content:             content,
		thumbnails:          thumbnails,
		db:                  db,
		auth:                auth,
		opendirs:            make(map[uint64][]*Inode),
		statuses:            make(map[string]FileStatusInfo),
		cacheExpirationDays: cacheExpirationDays,
		cacheCleanupStop:    make(chan struct{}),
		deltaLoopStop:       make(chan struct{}),
		deltaLoopCtx:        ctx,
		deltaLoopCancel:     cancel,
	}

	rootItem, err := graph.GetItem("root", auth)
	root := NewInodeDriveItem(rootItem)
	if err != nil {
		if graph.IsOffline(err) {
			// no network, load from db if possible and go to read-only state
			fs.Lock()
			fs.offline = true
			fs.Unlock()
			if root = fs.GetID("root"); root == nil {
				log.Error().Msg(
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
					log.Error().Msg("Cannot perform an offline startup without a valid " +
						"delta link from a previous session.")
					deltaLinkErr = errors.New("cannot perform an offline startup without a valid delta link from a previous session")
				}
				return nil
			}); viewErr != nil {
				log.Error().Err(viewErr).Msg("Failed to read delta link from database")
				return nil, fmt.Errorf("failed to read delta link from database: %w", viewErr)
			}
			if deltaLinkErr != nil {
				return nil, deltaLinkErr
			}
		} else {
			log.Error().Err(err).Msg("Could not fetch root item of filesystem!")
			return nil, fmt.Errorf("could not fetch root item of filesystem: %w", err)
		}
	}
	// root inode is inode 1
	fs.root = root.ID()
	fs.InsertID(fs.root, root)

	fs.uploads = NewUploadManager(2*time.Second, db, fs, auth)

	// Initialize download manager with 4 worker threads
	fs.downloads = NewDownloadManager(fs, auth, 4)

	if !fs.IsOffline() {
		// .Trash-UID is used by "gio trash" for user trash, create it if it
		// does not exist
		trash := fmt.Sprintf(".Trash-%d", os.Getuid())
		if child, _ := fs.GetChild(fs.root, trash, auth); child == nil {
			item, err := graph.Mkdir(trash, fs.root, auth)
			if err != nil {
				log.Error().Err(err).
					Msg("Could not create trash folder. " +
						"Trashing items through the file browser may result in errors.")
			} else {
				fs.InsertID(item.ID, NewInodeDriveItem(item))
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
		log.Error().Err(err).Msg("Failed to start D-Bus server")
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
	methodName, startTime := LogMethodCall()
	f.RLock()
	defer f.RUnlock()

	result := f.offline
	defer LogMethodReturn(methodName, startTime, result)
	return result
}

// TrackOfflineChange records a change made while offline
func (f *Filesystem) TrackOfflineChange(change *OfflineChange) error {
	methodName, startTime := LogMethodCall()
	defer func() {
		// We can't capture the return value directly in a defer, so we'll just log completion
		LogMethodReturn(methodName, startTime)
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
func (f *Filesystem) ProcessOfflineChanges() {
	// Create a logging context
	ctx := LogContext{
		Operation: "process_offline_changes",
	}

	// Log method entry with context
	methodName, startTime, logger, ctx := LogMethodCallWithContext("ProcessOfflineChanges", ctx)
	defer LogMethodReturnWithContext(methodName, startTime, logger, ctx)

	logger.Info().Msg("Processing offline changes...")

	// Get all offline changes
	changes := make([]*OfflineChange, 0)
	if err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketOfflineChanges)
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			change := &OfflineChange{}
			if err := json.Unmarshal(v, change); err != nil {
				return err
			}
			changes = append(changes, change)
			return nil
		})
	}); err != nil {
		LogErrorWithContext(err, ctx, "Failed to read offline changes from database")
		return
	}

	// Sort changes by timestamp
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Timestamp.Before(changes[j].Timestamp)
	})

	// Process each change
	for _, change := range changes {
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
					LogErrorWithContext(err, ctx, "Failed to queue upload for offline change",
						FieldID, change.ID)
				}
			}
		case "delete":
			// Handle deletion
			if !isLocalID(change.ID) {
				if err := graph.Remove(change.ID, f.auth); err != nil {
					LogErrorWithContext(err, ctx, "Failed to remove item during offline change processing",
						FieldID, change.ID)
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
							LogErrorWithContext(err, ctx, "Failed to move item during offline change processing",
								FieldID, change.ID,
								"oldPath", change.OldPath,
								"newPath", change.NewPath)
						}
					}
				}
			}
		}

		// Remove the processed change
		if err := f.db.Batch(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketOfflineChanges)
			if b == nil {
				return nil
			}
			key := []byte(fmt.Sprintf("%s-%d", change.ID, change.Timestamp.UnixNano()))
			return b.Delete(key)
		}); err != nil {
			LogErrorWithContext(err, ctx, "Failed to remove processed offline change from database",
				FieldID, change.ID,
				"timestamp", change.Timestamp)
		}
	}

	logger.Info().Msg("Finished processing offline changes.")
}

// TranslateID returns the DriveItemID for a given NodeID
func (f *Filesystem) TranslateID(nodeID uint64) string {
	methodName, startTime := LogMethodCall()
	f.RLock()
	defer f.RUnlock()

	var result string
	if nodeID > f.lastNodeID || nodeID == 0 {
		result = ""
	} else {
		result = f.inodes[nodeID-1]
	}

	defer LogMethodReturn(methodName, startTime, result)
	return result
}

// GetNodeID fetches the inode for a particular inode ID.
func (f *Filesystem) GetNodeID(nodeID uint64) *Inode {
	methodName, startTime := LogMethodCall()

	id := f.TranslateID(nodeID)
	if id == "" {
		// Log the return value (nil) and return
		defer LogMethodReturn(methodName, startTime, nil)
		return nil
	}

	result := f.GetID(id)
	// Log the return value (could be nil or a pointer)
	defer LogMethodReturn(methodName, startTime, result)
	return result
}

// InsertNodeID assigns a numeric inode ID used by the kernel if one is not
// already assigned.
func (f *Filesystem) InsertNodeID(inode *Inode) uint64 {
	methodName, startTime := LogMethodCall()

	nodeID := inode.NodeID()
	if nodeID == 0 {
		// lock ordering is to satisfy deadlock detector
		inode.Lock()
		f.Lock()

		f.lastNodeID++
		f.inodes = append(f.inodes, inode.DriveItem.ID)
		nodeID = f.lastNodeID
		inode.nodeID = nodeID

		f.Unlock()
		inode.Unlock()
	}

	defer LogMethodReturn(methodName, startTime, nodeID)
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
	methodName, startTime := LogMethodCall()

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
			log.Error().Err(err).Str("id", id).Msg("Failed to read inode from database")
			defer LogMethodReturn(methodName, startTime, nil)
			return nil
		}
		if found != nil {
			f.InsertNodeID(found)
			f.metadata.Store(id, found) // move to memory for next time
		}
		defer LogMethodReturn(methodName, startTime, found)
		return found
	}

	result := entry.(*Inode)
	defer LogMethodReturn(methodName, startTime, result)
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
func (f *Filesystem) GetIDWithContext(id string, ctx LogContext) *Inode {
	// Log method entry with context
	methodName, startTime, logger, ctx := LogMethodCallWithContext("GetIDWithContext", ctx)

	// Call the regular GetID method
	result := f.GetID(id)

	// Log method exit with context
	defer LogMethodReturnWithContext(methodName, startTime, logger, ctx, result)
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
	methodName, startTime := LogMethodCall()

	f.metadata.Store(id, inode)
	nodeID := f.InsertNodeID(inode)

	if id != inode.ID() {
		// we update the inode IDs here in case they do not match/changed
		inode.Lock()
		inode.DriveItem.ID = id
		inode.Unlock()

		f.Lock()
		if nodeID <= f.lastNodeID {
			f.inodes[nodeID-1] = id
		} else {
			log.Error().
				Uint64("nodeID", nodeID).
				Uint64("lastNodeID", f.lastNodeID).
				Msg("NodeID exceeded maximum node ID! Ignoring ID change.")
		}
		f.Unlock()
	}

	parentID := inode.ParentID()
	if parentID == "" {
		// root item, or parent not set
		defer LogMethodReturn(methodName, startTime, nodeID)
		return nodeID
	}
	parent := f.GetID(parentID)
	if parent == nil {
		log.Error().
			Str("parentID", parentID).
			Str("childID", id).
			Str("childName", inode.Name()).
			Msg("Parent item could not be found when setting parent.")
		defer LogMethodReturn(methodName, startTime, nodeID)
		return nodeID
	}

	// check if the item has already been added to the parent
	// Lock order is super key here, must go parent->child or the deadlock
	// detector screams at us.
	parent.Lock()
	defer parent.Unlock()
	for _, child := range parent.children {
		if child == id {
			// exit early, child cannot be added twice
			defer LogMethodReturn(methodName, startTime, nodeID)
			return nodeID
		}
	}

	// add to parent
	if inode.IsDir() {
		parent.subdir++
	}
	parent.children = append(parent.children, id)

	defer LogMethodReturn(methodName, startTime, nodeID)
	return nodeID
}

// InsertChild adds an item as a child of a specified parent ID.
func (f *Filesystem) InsertChild(parentID string, child *Inode) uint64 {
	child.Lock()
	// should already be set, just double-checking here.
	child.DriveItem.Parent.ID = parentID
	id := child.DriveItem.ID
	child.Unlock()
	return f.InsertID(id, child)
}

// DeleteID deletes an item from the cache, and removes it from its parent. Must
// be called before InsertID if being used to rename/move an item.
func (f *Filesystem) DeleteID(id string) {
	if inode := f.GetID(id); inode != nil {
		// If this is a directory, recursively delete all its children first
		if inode.IsDir() && inode.HasChildren() {
			// Make a copy of the children slice to avoid concurrent modification issues
			inode.RLock()
			childrenCopy := make([]string, len(inode.children))
			copy(childrenCopy, inode.children)
			inode.RUnlock()

			// Delete each child
			for _, childID := range childrenCopy {
				f.DeleteID(childID)
			}
		}

		parent := f.GetID(inode.ParentID())
		parent.Lock()
		for i, childID := range parent.children {
			if childID == id {
				parent.children = append(parent.children[:i], parent.children[i+1:]...)
				if inode.IsDir() {
					parent.subdir--
				}
				break
			}
		}
		parent.Unlock()
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
	log.Debug().Str("id", id).Str("func", "GetChildrenID").Msg("Starting GetChildrenID")

	// fetch item and catch common errors
	inode := f.GetID(id)
	children := make(map[string]*Inode)
	if inode == nil {
		log.Error().Str("id", id).Msg("Inode not found in cache")
		return children, errors.New(id + " not found in cache")
	} else if !inode.IsDir() {
		// Normal files are treated as empty folders. This only gets called if
		// we messed up and tried to get the children of a plain-old file.
		log.Warn().
			Str("id", id).
			Str("path", inode.Path()).
			Msg("Attepted to get children of ordinary file")
		return children, nil
	}

	// Get the path before acquiring any locks to avoid potential deadlocks
	pathForLogs := inode.Path()
	log.Debug().Str("id", id).Str("path", pathForLogs).Msg("Checking if children are already cached")

	// If item.children is not nil, it means we have the item's children
	// already and can fetch them directly from the cache
	inode.RLock()
	if inode.children != nil {
		log.Debug().Str("id", id).Str("path", pathForLogs).Int("childCount", len(inode.children)).Msg("Children found in cache, retrieving them")
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
		inode.RUnlock()
		log.Debug().Str("id", id).Str("path", pathForLogs).Int("childCount", len(children)).Msg("Successfully retrieved children from cache")
		return children, nil
	}
	// Update path before unlocking to avoid potential deadlocks
	pathForLogs = inode.Path()
	inode.RUnlock()

	log.Debug().Str("id", id).Str("path", pathForLogs).Msg("Children not in cache, fetching from server")

	// We haven't fetched the children for this item yet, get them from the server.
	log.Debug().Str("id", id).Str("path", pathForLogs).Msg("About to call graph.GetItemChildren")
	fetched, err := graph.GetItemChildren(id, auth)
	log.Debug().Str("id", id).Str("path", pathForLogs).Err(err).Msg("Returned from graph.GetItemChildren")

	if err != nil {
		if graph.IsOffline(err) {
			log.Warn().Str("id", id).
				Msg("We are offline, and no children found in cache. " +
					"Pretending there are no children.")
			return children, nil
		}
		// something else happened besides being offline
		log.Error().Str("id", id).Str("path", pathForLogs).Err(err).Msg("Error fetching children from server")
		return nil, err
	}

	log.Debug().Str("id", id).Str("path", pathForLogs).Int("fetchedCount", len(fetched)).Msg("Processing fetched children")

	// Store the path before locking to avoid potential deadlocks
	processingPath := pathForLogs

	inode.Lock()
	inode.children = make([]string, 0)
	for i, item := range fetched {
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

		if i%50 == 0 && i > 0 {
			log.Debug().Str("id", id).Str("path", processingPath).Int("processedCount", i).Int("totalCount", len(fetched)).Msg("Processing children progress")
		}
	}
	log.Debug().Str("id", id).Str("path", processingPath).Int("childrenCount", len(children)).Uint32("subdirCount", inode.subdir).Msg("Finished processing all children")
	inode.Unlock()

	log.Debug().Str("id", id).Str("path", processingPath).Int("childrenCount", len(children)).Msg("GetChildrenID completed successfully")
	return children, nil
}

// GetChildrenPath grabs all DriveItems that are the children of the resource at
// the path. If items are not found, they are fetched.
func (f *Filesystem) GetChildrenPath(path string, auth *graph.Auth) (map[string]*Inode, error) {
	inode, err := f.GetPath(path, auth)
	if err != nil {
		return make(map[string]*Inode), err
	}
	return f.GetChildrenID(inode.ID(), auth)
}

// GetPath fetches a given DriveItem in the cache, if any items along the way are
// not found, they are fetched.
func (f *Filesystem) GetPath(path string, auth *graph.Auth) (*Inode, error) {
	lastID := f.root
	if path == "/" {
		return f.GetID(lastID), nil
	}

	// from the root directory, traverse the chain of items till we reach our
	// target ID.
	path = strings.TrimSuffix(strings.ToLower(path), "/")
	split := strings.Split(path, "/")[1:] //omit leading "/"
	var inode *Inode
	for i := 0; i < len(split); i++ {
		// fetches children
		children, err := f.GetChildrenID(lastID, auth)
		if err != nil {
			return nil, err
		}

		var exists bool // if we use ":=", item is shadowed
		inode, exists = children[split[i]]
		if !exists {
			// the item still doesn't exist after fetching from server. it
			// doesn't exist
			return nil, errors.New(strings.Join(split[:i+1], "/") +
				" does not exist on server or in local cache")
		}
		lastID = inode.ID()
	}
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
		log.Error().
			Str("key", key).
			Str("path", inode.Path()).
			Msg(errMsg)
		return 0, errors.New(errMsg)
	}

	// Coded this way to make sure locks are in the same order for the deadlock
	// detector (lock ordering needs to be the same as InsertID: Parent->Child).
	parentID := parent.ID()
	inode.Lock()
	inode.DriveItem.Parent.ID = parentID
	id := inode.DriveItem.ID
	inode.Unlock()

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
	parent := f.GetID(inode.ParentID())
	parent.Lock()
	for i, child := range parent.children {
		if child == oldID {
			parent.children[i] = newID
			break
		}
	}
	parent.Unlock()

	// now actually perform the metadata+content move
	f.DeleteID(oldID)
	f.InsertID(newID, inode)
	if inode.IsDir() {
		return nil
	}
	if err := f.content.Move(oldID, newID); err != nil {
		log.Error().Err(err).
			Str("oldID", oldID).
			Str("newID", newID).
			Msg("Failed to move file content")
		return fmt.Errorf("failed to move file content: %w", err)
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
	inode.Parent.ID = parent.DriveItem.ID
	f.InsertID(id, inode)
	return nil
}

// StartCacheCleanup starts a background goroutine that periodically cleans up
// the content cache by removing files that haven't been modified for the specified
// number of days.
func (f *Filesystem) StartCacheCleanup() {
	// Don't start cleanup if expiration days is 0 or negative
	if f.cacheExpirationDays <= 0 {
		log.Info().Msg("Cache cleanup disabled (expiration days <= 0)")
		return
	}

	log.Info().Int("expirationDays", f.cacheExpirationDays).Msg("Starting content cache cleanup routine")

	// Add to wait group to track this goroutine
	f.cacheCleanupWg.Add(1)

	// Run cleanup in a goroutine
	go func() {
		defer f.cacheCleanupWg.Done()
		// Run cleanup immediately on startup
		count, err := f.content.CleanupCache(f.cacheExpirationDays)
		if err != nil {
			log.Error().Err(err).Msg("Error during initial content cache cleanup")
		} else {
			log.Info().Int("removedFiles", count).Msg("Initial content cache cleanup completed")
		}

		// Set up ticker for periodic cleanup (once per day)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Run cleanup
				count, err := f.content.CleanupCache(f.cacheExpirationDays)
				if err != nil {
					log.Error().Err(err).Msg("Error during content cache cleanup")
				} else {
					log.Info().Int("removedFiles", count).Msg("Content cache cleanup completed")
				}
			case <-f.cacheCleanupStop:
				// Stop the cleanup routine
				log.Info().Msg("Stopping content cache cleanup routine")
				return
			}
		}
	}()
}

// StopCacheCleanup stops the background cache cleanup routine.
func (f *Filesystem) StopCacheCleanup() {
	log.Info().Msg("Stopping cache cleanup routine...")
	// Only send stop signal if expiration days is positive (cleanup is running)
	if f.cacheExpirationDays > 0 {
		f.cacheCleanupStopOnce.Do(func() {
			close(f.cacheCleanupStop)
		})
		f.cacheCleanupWg.Wait()
		log.Info().Msg("Cache cleanup routine stopped")
	}
}

// StopDeltaLoop stops the delta loop goroutine and waits for it to finish.
func (f *Filesystem) StopDeltaLoop() {
	log.Info().Msg("Stopping delta loop...")

	// Cancel the context to interrupt any in-progress network requests
	f.deltaLoopCancel()
	log.Debug().Msg("Cancelled delta loop context to interrupt network operations")

	// Close the stop channel to signal the delta loop to stop
	f.deltaLoopStopOnce.Do(func() {
		close(f.deltaLoopStop)
	})
	log.Debug().Msg("Closed delta loop stop channel")

	// Wait for delta loop to finish with a timeout
	done := make(chan struct{})
	go func() {
		f.deltaLoopWg.Wait()
		close(done)
	}()

	// Wait for delta loop to finish or timeout after 10 seconds
	select {
	case <-done:
		log.Info().Msg("Delta loop stopped successfully")
	case <-time.After(10 * time.Second):
		log.Warn().Msg("Timed out waiting for delta loop to stop - continuing shutdown anyway")
		// Log additional debug information
		log.Debug().Msg("Delta loop may be stuck in a network operation or processing a large batch of changes")
		log.Debug().Msg("This is not a critical error, but may indicate a potential issue with network operations")
	}
}

// StopDownloadManager stops the download manager and waits for all workers to finish.
func (f *Filesystem) StopDownloadManager() {
	log.Info().Msg("Stopping download manager...")
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
			log.Info().Msg("Download manager stopped successfully")
		case <-time.After(5 * time.Second):
			log.Warn().Msg("Timed out waiting for download manager to stop")
		}
	}
}

// StopUploadManager stops the upload manager and waits for all uploads to finish.
func (f *Filesystem) StopUploadManager() {
	log.Info().Msg("Stopping upload manager...")
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
			log.Info().Msg("Upload manager stopped successfully")
		case <-time.After(5 * time.Second):
			log.Warn().Msg("Timed out waiting for upload manager to stop")
		}
	}
}

// SerializeAll dumps all inode metadata currently in the cache to disk. This
// metadata is only used later if an item could not be found in memory AND the
// cache is offline. Old metadata is not removed, only overwritten (to avoid an
// offline session from wiping all metadata on a subsequent serialization).
func (f *Filesystem) SerializeAll() {
	log.Debug().Msg("Serializing cache metadata to disk.")

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
				return fmt.Errorf("failed to put item %s: %w", k, err)
			}
			if k == f.root {
				// root item must be updated manually (since there's actually
				// two copies)
				if err := b.Put([]byte("root"), v); err != nil {
					return fmt.Errorf("failed to put root item: %w", err)
				}
			}
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Msg("Failed to serialize metadata to database")
	}
}
