package fs

import (
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog/log"
)

// Mkdir creates a directory
func (f *Filesystem) Mkdir(cancel <-chan struct{}, in *fuse.MkdirIn, name string, out *fuse.EntryOut) fuse.Status {
	parent := f.GetInodeForHeader(in.Header)
	if parent == nil {
		return fuse.ENOENT
	}

	if isNameRestricted(name) {
		log.Error().Str("name", name).Msg("Attempted to create directory with restricted name")
		return fuse.Status(syscall.EINVAL)
	}

	// create directory on remote
	id, err := graph.CreateDir(parent.ID(), name, f.auth)
	if err != nil {
		log.Error().
			Err(err).
			Str("parent", parent.ID()).
			Str("name", name).
			Msg("Failed to create directory")
		return fuse.EIO
	}

	// create directory locally
	now := time.Now()
	inode := NewInode(name, 0755|os.ModeDir, parent)
	inode.DriveItem.ID = id
	inode.DriveItem.CreatedDateTime = &now
	inode.DriveItem.LastModifiedDateTime = &now
	inode.DriveItem.CTag = "1"
	inode.DriveItem.ETag = "0"
	inode.DriveItem.File = nil
	inode.DriveItem.Folder = &graph.Folder{ChildCount: 0}

	parent.Lock()
	parent.Children[name] = inode
	parent.Unlock()

	f.InsertID(id, inode)
	out.Attr.FromInodeAttr(inode.InodeAttr())
	out.NodeId = inode.NodeID()
	out.Generation = 1
	out.SetEntryTimeout(timeout)
	out.SetAttrTimeout(timeout)
	return fuse.OK
}

// Rmdir removes a directory
func (f *Filesystem) Rmdir(cancel <-chan struct{}, in *fuse.InHeader, name string) fuse.Status {
	parent := f.GetInodeForHeader(in)
	if parent == nil {
		return fuse.ENOENT
	}

	parent.Lock()
	defer parent.Unlock()
	child, ok := parent.Children[name]
	if !ok {
		return fuse.ENOENT
	}

	if !child.IsDir() {
		return fuse.Status(syscall.ENOTDIR)
	}

	// delete from remote
	if err := graph.Delete(child.ID(), f.auth); err != nil {
		log.Error().
			Err(err).
			Str("id", child.ID()).
			Str("name", name).
			Msg("Failed to delete directory")
		return fuse.EIO
	}

	// delete from local
	delete(parent.Children, name)
	f.RemoveID(child.ID())
	return fuse.OK
}

// OpenDir opens a directory for reading
func (f *Filesystem) OpenDir(cancel <-chan struct{}, in *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
	inode := f.GetInodeForHeader(in.Header)
	if inode == nil {
		return fuse.ENOENT
	}

	if !inode.IsDir() {
		return fuse.Status(syscall.ENOTDIR)
	}

	// if we're offline, just return what we have
	if f.auth.AccessToken == "" {
		return fuse.OK
	}

	// if we're online, fetch the latest children
	children, err := graph.GetItemChildren(inode.ID(), f.auth)
	if err != nil {
		log.Error().
			Err(err).
			Str("id", inode.ID()).
			Msg("Failed to get directory children")
		return fuse.EIO
	}

	// update the directory
	inode.Lock()
	defer inode.Unlock()

	// create a map of existing children by ID for quick lookup
	existingByID := make(map[string]*Inode)
	for _, child := range inode.Children {
		existingByID[child.ID()] = child
	}

	// create a map of new children by name
	newChildren := make(map[string]*Inode)
	for _, item := range children {
		// if the child already exists, update it
		if existing, ok := existingByID[item.ID]; ok {
			existing.Lock()
			existing.DriveItem = item
			existing.Unlock()
			newChildren[item.Name] = existing
		} else {
			// otherwise create a new inode
			child := NewInodeFromDriveItem(item, inode)
			newChildren[item.Name] = child
			f.InsertID(item.ID, child)
		}
	}

	// replace the children map
	inode.Children = newChildren
	return fuse.OK
}

// ReleaseDir releases a directory
func (f *Filesystem) ReleaseDir(in *fuse.ReleaseIn) {
	// nothing to do
}

// ReadDirPlus reads a directory and returns the entries plus their attributes
func (f *Filesystem) ReadDirPlus(cancel <-chan struct{}, in *fuse.ReadIn, out *fuse.DirEntryList) fuse.Status {
	inode := f.GetInodeForHeader(in.Header)
	if inode == nil {
		return fuse.ENOENT
	}

	if !inode.IsDir() {
		return fuse.Status(syscall.ENOTDIR)
	}

	// if we're at the end of the directory, return OK
	if in.Offset > 0 && uint64(in.Offset) >= uint64(len(inode.Children)+2) {
		return fuse.OK
	}

	// add . and .. entries
	if in.Offset == 0 {
		out.AddDirLookupEntry(".", inode.NodeID(), inode.InodeAttr())
	}
	if in.Offset <= 1 {
		var parent *Inode
		if inode.Parent != nil {
			parent = inode.Parent
		} else {
			parent = inode
		}
		out.AddDirLookupEntry("..", parent.NodeID(), parent.InodeAttr())
	}

	// add children
	inode.RLock()
	defer inode.RUnlock()

	offset := int(in.Offset)
	if offset < 2 {
		offset = 2
	}

	idx := 0
	for name, child := range inode.Children {
		if idx+2 >= offset {
			if !out.AddDirLookupEntry(name, child.NodeID(), child.InodeAttr()) {
				break
			}
		}
		idx++
	}

	return fuse.OK
}

// ReadDir reads a directory and returns the entries
func (f *Filesystem) ReadDir(cancel <-chan struct{}, in *fuse.ReadIn, out *fuse.DirEntryList) fuse.Status {
	inode := f.GetInodeForHeader(in.Header)
	if inode == nil {
		return fuse.ENOENT
	}

	if !inode.IsDir() {
		return fuse.Status(syscall.ENOTDIR)
	}

	// if we're at the end of the directory, return OK
	if in.Offset > 0 && uint64(in.Offset) >= uint64(len(inode.Children)+2) {
		return fuse.OK
	}

	// add . and .. entries
	if in.Offset == 0 {
		out.AddDirEntry(".", inode.NodeID(), inode.Mode())
	}
	if in.Offset <= 1 {
		var parent *Inode
		if inode.Parent != nil {
			parent = inode.Parent
		} else {
			parent = inode
		}
		out.AddDirEntry("..", parent.NodeID(), parent.Mode())
	}

	// add children
	inode.RLock()
	defer inode.RUnlock()

	offset := int(in.Offset)
	if offset < 2 {
		offset = 2
	}

	idx := 0
	for name, child := range inode.Children {
		if idx+2 >= offset {
			if !out.AddDirEntry(name, child.NodeID(), child.Mode()) {
				break
			}
		}
		idx++
	}

	return fuse.OK
}

// Lookup looks up a file or directory by name
func (f *Filesystem) Lookup(cancel <-chan struct{}, in *fuse.InHeader, name string, out *fuse.EntryOut) fuse.Status {
	parent := f.GetInodeForHeader(in)
	if parent == nil {
		return fuse.ENOENT
	}

	parent.RLock()
	child, ok := parent.Children[name]
	parent.RUnlock()
	if !ok {
		return fuse.ENOENT
	}

	out.Attr.FromInodeAttr(child.InodeAttr())
	out.NodeId = child.NodeID()
	out.Generation = 1
	out.SetEntryTimeout(timeout)
	out.SetAttrTimeout(timeout)
	return fuse.OK
}