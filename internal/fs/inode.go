package fs

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Inode represents a file or folder in the onemount filesystem.
// It wraps a DriveItem from the Microsoft Graph API and adds filesystem-specific
// metadata and functionality. The Inode struct is thread-safe, with all methods
// properly handling concurrent access through its embedded RWMutex.
//
// The embedded DriveItem's fields should never be accessed directly, as they are
// not safe for concurrent access. Instead, use the provided methods to access
// and modify the Inode's properties.
//
// Reads and writes are performed directly on DriveItems rather than implementing
// a separate file handle interface to minimize the complexity of operations like
// Flush. All modifications to the Inode are tracked and synchronized with OneDrive
// when appropriate.
type Inode struct {
	sync.RWMutex                      // Protects access to all fields
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

// NewInode creates a new Inode with the specified name, mode, and parent.
// This constructor is typically used when creating new files or directories
// that don't yet exist in OneDrive. It assigns a local ID to the new Inode,
// which will be replaced with a OneDrive ID when the item is uploaded.
//
// Parameters:
//   - name: The name of the file or directory
//   - mode: The file mode/permissions (e.g., directory or regular file)
//   - parent: The parent Inode, or nil if this is the root
//
// Returns:
//   - A new Inode instance with initialized fields
func NewInode(name string, mode uint32, parent *Inode) *Inode {
	itemParent := &graph.DriveItemParent{ID: "", Path: ""}
	if parent != nil {
		itemParent.Path = parent.Path()
		parent.RLock()
		itemParent.ID = parent.DriveItem.ID
		itemParent.DriveID = parent.DriveItem.Parent.DriveID
		itemParent.DriveType = parent.DriveItem.Parent.DriveType
		parent.RUnlock()
	}

	currentTime := time.Now()
	return &Inode{
		DriveItem: graph.DriveItem{
			ID:      localID(),
			Name:    name,
			Parent:  itemParent,
			ModTime: &currentTime,
		},
		children: make([]string, 0),
		mode:     mode,
		xattrs:   make(map[string][]byte),
	}
}

// AsJSON converts a DriveItem to JSON for use with local storage. Not used with
// the API. FIXME: If implemented as MarshalJSON, this will break delta syncs
// for business accounts. Don't ask me why.
func (i *Inode) AsJSON() []byte {
	i.RLock()
	defer i.RUnlock()
	data, _ := json.Marshal(SerializeableInode{
		DriveItem: i.DriveItem,
		Children:  i.children,
		Subdir:    i.subdir,
		Mode:      i.mode,
		Xattrs:    i.xattrs,
	})
	return data
}

// NewInodeJSON converts JSON to a *DriveItem when loading from local storage. Not
// used with the API. FIXME: If implemented as UnmarshalJSON, this will break
// delta syncs for business accounts. Don't ask me why.
func NewInodeJSON(data []byte) (*Inode, error) {
	var raw SerializeableInode
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return &Inode{
		DriveItem: raw.DriveItem,
		children:  raw.Children,
		mode:      raw.Mode,
		subdir:    raw.Subdir,
		xattrs:    raw.Xattrs,
	}, nil
}

// NewInodeDriveItem creates a new Inode from an existing DriveItem.
// This constructor is typically used when retrieving items from OneDrive
// through the Microsoft Graph API. It wraps the DriveItem in an Inode
// to provide filesystem functionality.
//
// Parameters:
//   - item: The DriveItem from OneDrive to wrap in an Inode
//
// Returns:
//   - A new Inode instance containing the DriveItem, or nil if item is nil
func NewInodeDriveItem(item *graph.DriveItem) *Inode {
	if item == nil {
		return nil
	}
	return &Inode{
		DriveItem: *item,
		xattrs:    make(map[string][]byte),
	}
}

// String is only used for debugging by go-fuse
func (i *Inode) String() string {
	return i.Name()
}

// Name is used to ensure thread-safe access to the NameInternal field.
func (i *Inode) Name() string {
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.Name
}

// SetName sets the name of the item in a thread-safe manner.
func (i *Inode) SetName(name string) {
	i.Lock()
	i.DriveItem.Name = name
	i.Unlock()
}

// NodeID returns the inodes ID in the filesystem
func (i *Inode) NodeID() uint64 {
	i.RLock()
	defer i.RUnlock()
	return i.nodeID
}

// SetNodeID sets the inode ID for an inode if not already set. Does nothing if
// the Inode already has an ID.
func (i *Inode) SetNodeID(id uint64) uint64 {
	i.Lock()
	defer i.Unlock()
	if i.nodeID == 0 {
		i.nodeID = id
	}
	return i.nodeID
}

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(length int) string {
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		out[i] = charset[rand.Intn(len(charset))]
	}
	return string(out)
}

func localID() string {
	return "local-" + randString(20)
}

func isLocalID(id string) bool {
	return strings.HasPrefix(id, "local-") || id == ""
}

// ID returns the internal ID of the item
func (i *Inode) ID() string {
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.ID
}

// ParentID returns the ID of this item's parent.
func (i *Inode) ParentID() string {
	i.RLock()
	defer i.RUnlock()
	if i.DriveItem.Parent == nil {
		return ""
	}
	return i.DriveItem.Parent.ID
}

// Path returns an inode's full Path
func (i *Inode) Path() string {
	i.RLock()
	defer i.RUnlock()

	// special case when it's the root item
	name := i.DriveItem.Name
	if i.DriveItem.Parent == nil || i.DriveItem.Parent.ID == "" {
		if name == "root" {
			return "/"
		}
		return name
	}

	// all paths come prefixed with "/drive/root:"
	prepath := strings.TrimPrefix(i.DriveItem.Parent.Path+"/"+name, "/drive/root:")
	return strings.Replace(prepath, "//", "/", -1)
}

// HasChanges returns true if the file has local changes that haven't been
// uploaded yet.
func (i *Inode) HasChanges() bool {
	i.RLock()
	defer i.RUnlock()
	return i.hasChanges
}

// HasChildren returns true if the item has more than 0 children
func (i *Inode) HasChildren() bool {
	i.RLock()
	defer i.RUnlock()
	return len(i.children) > 0
}

// makeattr is a convenience function to create a set of filesystem attrs for
// use with syscalls that use or modify attrs.
func (i *Inode) makeAttr() fuse.Attr {
	mtime := i.ModTime()
	return fuse.Attr{
		Ino:   i.NodeID(),
		Size:  i.Size(),
		Nlink: i.NLink(),
		Ctime: mtime,
		Mtime: mtime,
		Atime: mtime,
		Mode:  i.Mode(),
		// whatever user is running the filesystem is the owner
		Owner: fuse.Owner{
			Uid: uint32(os.Getuid()),
			Gid: uint32(os.Getgid()),
		},
	}
}

// IsDir returns if it is a directory (true) or file (false).
func (i *Inode) IsDir() bool {
	if i == nil {
		return false
	}
	// 0 if the dir bit is not set
	return i.Mode()&fuse.S_IFDIR > 0
}

// Mode returns the permissions/mode of the file.
func (i *Inode) Mode() uint32 {
	if i == nil {
		return 0
	}
	i.RLock()
	defer i.RUnlock()
	if i.mode == 0 { // only 0 if fetched from Graph API
		if i.DriveItem.IsDir() {
			return fuse.S_IFDIR | 0755
		}
		return fuse.S_IFREG | 0644
	}
	return i.mode
}

// ModTime returns the Unix timestamp of last modification (to get a time.Time
// struct, use time.Unix(int64(d.ModTime()), 0))
func (i *Inode) ModTime() uint64 {
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.ModTimeUnix()
}

// NLink gives the number of hard links to an inode (or child count if a
// directory)
func (i *Inode) NLink() uint32 {
	if i.IsDir() {
		i.RLock()
		defer i.RUnlock()
		// we precompute subdir due to mutex lock contention between NLink and
		// other ops. subdir is modified by cache Insert/Delete and GetChildren.
		return 2 + i.subdir
	}
	return 1
}

// Size pretends that folders are 4096 bytes, even though they're 0 (since
// they actually don't exist).
func (i *Inode) Size() uint64 {
	if i == nil {
		return 0
	}
	if i.IsDir() {
		return 4096
	}
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.Size
}

// Octal converts a number to its octal representation in string form.
func Octal(i uint32) string {
	return strconv.FormatUint(uint64(i), 8)
}
