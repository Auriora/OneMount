package fs

import (
	"context"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
	bolt "go.etcd.io/bbolt"
	"sync"
	"time"
)

// FilesystemInterface defines the interface for the filesystem operations
// that are used by other packages. This interface is implemented by the
// Filesystem type in the cache package.
type FilesystemInterface interface {
	// GetFileStatus File status methods
	GetFileStatus(id string) FileStatusInfo
	SetFileStatus(id string, status FileStatusInfo)
	MarkFileDownloading(id string)
	MarkFileOutofSync(id string)
	MarkFileError(id string, err error)
	MarkFileConflict(id string, message string)
	UpdateFileStatus(inode *Inode)
	InodePath(inode *Inode) string

	// GetID Inode management methods
	GetID(id string) *Inode
	MoveID(oldID string, newID string) error
	GetInodeContent(inode *Inode) *[]byte

	// IsOffline Filesystem state methods
	IsOffline() bool
}

// FileStatusDBusServerInterface defines the interface for the D-Bus server
// that is used by other packages. This interface is implemented by the
// FileStatusDBusServer type in the fs package.
type FileStatusDBusServerInterface interface {
	// Start starts the D-Bus server
	Start() error

	// Stop stops the D-Bus server and cleans up all resources
	Stop()

	// SendFileStatusUpdate sends a D-Bus signal with the updated file status
	SendFileStatusUpdate(path string, status string)
}

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

	// Root context for all operations
	ctx    context.Context    // Root context for all operations
	cancel context.CancelFunc // Function to cancel the root context
	Wg     sync.WaitGroup     // Wait group for all goroutines

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

	// StatFs warning throttling
	statfsWarningM    sync.RWMutex // Mutex for StatFs warning state
	statfsWarningTime time.Time    // Last time StatFs warning was shown
}
