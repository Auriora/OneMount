package fs

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

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
		if parent.DriveItem.Parent != nil {
			itemParent.DriveID = parent.DriveItem.Parent.DriveID
			itemParent.DriveType = parent.DriveItem.Parent.DriveType
		}
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
// the API. NOTE: This is intentionally NOT implemented as MarshalJSON because
// that would break delta syncs for business accounts due to Graph API quirks.
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
// used with the API. NOTE: This is intentionally NOT implemented as UnmarshalJSON
// because that would break delta syncs for business accounts due to Graph API quirks.
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

// GetName returns the name of the Inode.
func (i *Inode) GetName() string {
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.Name
}

// GetNodeID returns the filesystem node ID of the Inode.
func (i *Inode) GetNodeID() uint64 {
	i.RLock()
	defer i.RUnlock()
	return i.nodeID
}

// GetID returns the ID of the Inode.
func (i *Inode) GetID() string {
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.ID
}

// GetParentID returns the ID of the parent Inode.
func (i *Inode) GetParentID() string {
	i.RLock()
	defer i.RUnlock()
	if i.DriveItem.Parent == nil {
		return ""
	}
	return i.DriveItem.Parent.ID
}

// GetPath returns the path of the Inode.
func (i *Inode) GetPath() string {
	i.RLock()
	defer i.RUnlock()
	if i.DriveItem.Parent == nil {
		return "/"
	}
	if i.DriveItem.Parent.Path == "" {
		return "/" + i.DriveItem.Name
	}
	return i.DriveItem.Parent.Path + "/" + i.DriveItem.Name
}

// SetHasChanges sets whether the Inode has changes that need to be uploaded.
func (i *Inode) SetHasChanges(hasChanges bool) {
	i.Lock()
	defer i.Unlock()
	i.hasChanges = hasChanges
}

// GetChildren returns the children of the Inode.
func (i *Inode) GetChildren() []string {
	i.RLock()
	defer i.RUnlock()
	return i.children
}

// SetChildren sets the children of the Inode.
func (i *Inode) SetChildren(children []string) {
	i.Lock()
	defer i.Unlock()
	i.children = children
}

// AddChild adds a child to the Inode.
func (i *Inode) AddChild(child string) {
	i.Lock()
	defer i.Unlock()
	i.children = append(i.children, child)
}

// MakeAttr creates a fuse.Attr from the Inode.
func (i *Inode) MakeAttr() fuse.Attr {
	i.RLock()
	defer i.RUnlock()
	attr := fuse.Attr{
		Ino:  i.nodeID,
		Mode: i.mode,
	}
	attr.Size = i.DriveItem.Size
	if i.DriveItem.ModTime != nil {
		attr.Mtime = uint64(i.DriveItem.ModTime.Unix())
	}
	attr.Nlink = i.GetNLink()
	return attr
}

// GetMode returns the mode of the Inode.
func (i *Inode) GetMode() uint32 {
	i.RLock()
	defer i.RUnlock()
	if i.mode != 0 {
		return i.mode
	}
	if i.DriveItem.Folder != nil {
		return uint32(os.ModeDir) | 0755
	}
	return 0644
}

// VerifyChecksum checks to see if the Inode's checksum matches what it's
// supposed to be. This delegates to the embedded DriveItem's VerifyChecksum method.
func (i *Inode) VerifyChecksum(checksum string) bool {
	if i == nil {
		return false
	}
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.VerifyChecksum(checksum)
}

// SetMode sets the mode of the Inode.
func (i *Inode) SetMode(mode uint32) {
	i.Lock()
	defer i.Unlock()
	i.mode = mode
}

// GetModTime returns the modification time of the Inode.
func (i *Inode) GetModTime() uint64 {
	i.RLock()
	defer i.RUnlock()
	if i.DriveItem.ModTime != nil {
		return uint64(i.DriveItem.ModTime.Unix())
	}
	return 0
}

// GetNLink returns the number of hard links to the Inode.
func (i *Inode) GetNLink() uint32 {
	i.RLock()
	defer i.RUnlock()
	if i.IsDir() {
		return 2 + i.subdir
	}
	return 1
}

// GetSubdir returns the number of subdirectories of the Inode.
func (i *Inode) GetSubdir() uint32 {
	i.RLock()
	defer i.RUnlock()
	return i.subdir
}

// SetSubdir sets the number of subdirectories of the Inode.
func (i *Inode) SetSubdir(subdir uint32) {
	i.Lock()
	defer i.Unlock()
	i.subdir = subdir
}

// GetSize returns the size of the Inode.
func (i *Inode) GetSize() uint64 {
	i.RLock()
	defer i.RUnlock()
	return i.DriveItem.Size
}

// GetXattrs returns the extended attributes of the Inode.
func (i *Inode) GetXattrs() map[string][]byte {
	i.RLock()
	defer i.RUnlock()
	return i.xattrs
}

// SetXattr sets an extended attribute of the Inode.
func (i *Inode) SetXattr(name string, value []byte) {
	i.Lock()
	defer i.Unlock()
	i.xattrs[name] = value
}

// GetXattr gets an extended attribute of the Inode.
func (i *Inode) GetXattr(name string) ([]byte, bool) {
	i.RLock()
	defer i.RUnlock()
	value, ok := i.xattrs[name]
	return value, ok
}

// RemoveXattr removes an extended attribute of the Inode.
func (i *Inode) RemoveXattr(name string) {
	i.Lock()
	defer i.Unlock()
	delete(i.xattrs, name)
}
