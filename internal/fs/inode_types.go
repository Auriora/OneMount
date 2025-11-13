package fs

import (
	"sync"

	"github.com/auriora/onemount/internal/graph"
)

// Inode represents a file or folder in the onemount filesystem.
// It wraps a DriveItem from the Microsoft Graph API and adds filesystem-specific
// metadata and functionality. The Inode struct is thread-safe, with all methods
// properly handling concurrent access through its mutex pointer.
//
// The embedded DriveItem's fields should never be accessed directly, as they are
// not safe for concurrent access. Instead, use the provided methods to access
// and modify the Inode's properties.
//
// Reads and writes are performed directly on DriveItems rather than implementing
// a separate file handle interface to minimize the complexity of operations like
// Flush. All modifications to the Inode are tracked and synchronized with OneDrive
// when appropriate.
//
// Note: The mutex is a pointer to avoid issues with struct copying and to improve
// performance by reducing the size of the Inode struct when passed by value.
type Inode struct {
	mu              *sync.RWMutex     // Protects access to all fields
	graph.DriveItem                   // The underlying OneDrive item
	nodeID          uint64            // Filesystem node ID used by the kernel
	children        []string          // Slice of child item IDs, nil when uninitialized
	hasChanges      bool              // Flag to trigger an upload on flush
	subdir          uint32            // Number of subdirectories, used by NLink()
	mode            uint32            // File mode/permissions, do not set manually
	xattrs          map[string][]byte // Extended attributes
}

// SerializeableInode is like a Inode, but can be serialized for local storage
// to disk
type SerializeableInode struct {
	graph.DriveItem
	Children []string
	Subdir   uint32
	Mode     uint32
	Xattrs   map[string][]byte
}

// InodeInfo is an interface that provides basic information about an inode
// This is used to avoid circular dependencies between packages
type InodeInfo interface {
	// ID returns the unique identifier of the inode
	ID() string

	// Name returns the name of the inode
	Name() string

	// IsDir returns true if the inode represents a directory
	IsDir() bool
}
