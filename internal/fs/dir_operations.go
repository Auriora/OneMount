package fs

import (
	"github.com/auriora/onemount/internal/logging"
	"math"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Mkdir creates a directory.
func (f *Filesystem) Mkdir(_ <-chan struct{}, in *fuse.MkdirIn, name string, out *fuse.EntryOut) fuse.Status {
	if isNameRestricted(name) {
		return fuse.EINVAL
	}

	inode := f.GetNodeID(in.NodeId)
	if inode == nil {
		return fuse.ENOENT
	}
	id := inode.ID()
	path := filepath.Join(inode.Path(), name)
	if existing, _ := f.GetChild(id, name, f.auth); existing != nil {
		return fuse.Status(syscall.EEXIST)
	}
	currentTime := time.Now()
	ctx := logging.DefaultLogger.With().
		Str("op", "Mkdir").
		Uint64("nodeID", in.NodeId).
		Str("id", id).
		Str("path", path).
		Str("mode", Octal(in.Mode)).
		Logger()
	ctx.Debug().Msg("")

	var item *graph.DriveItem
	var err error

	if f.IsOffline() {
		// In offline mode, create a local directory that will be synced when online
		ctx.Info().Msg("Directory creation in offline mode will be cached locally")
		item = &graph.DriveItem{
			ID:     localID(),
			Name:   name,
			Folder: &graph.Folder{},
			Parent: &graph.DriveItemParent{
				ID: id,
			},
			ModTime: &currentTime,
		}
	} else {
		// create the new directory on the server
		item, err = graph.Mkdir(name, id, f.auth)
		if err != nil {
			logging.LogError(err, "Could not create remote directory",
				logging.FieldOperation, "Mkdir",
				logging.FieldID, id,
				logging.FieldPath, path,
				"name", name)
			return fuse.EREMOTEIO
		}
		if item.ModTime == nil {
			item.ModTime = &currentTime
		}
	}

	newInode := NewInodeDriveItem(item)
	newInode.mode = in.Mode | fuse.S_IFDIR
	if !f.IsOffline() {
		f.markChildPendingRemote(newInode.ID())
	}

	out.NodeId = f.InsertChild(id, newInode)
	out.Attr = newInode.makeAttr()
	out.SetAttrTimeout(timeout)
	out.SetEntryTimeout(timeout)
	if f.IsOffline() || isLocalID(newInode.ID()) {
		f.markDirtyLocalState(newInode.ID())
	} else {
		f.markHydratedState(newInode.ID())
	}
	return fuse.OK
}

// Rmdir removes a directory if it's empty.
func (f *Filesystem) Rmdir(cancel <-chan struct{}, in *fuse.InHeader, name string) fuse.Status {
	parent := f.GetNodeID(in.NodeId)
	if parent == nil {
		return fuse.ENOENT
	}
	parentID := parent.ID()
	child, _ := f.GetChild(parentID, name, f.auth)
	if child == nil {
		return fuse.ENOENT
	}
	if child.HasChildren() {
		return fuse.Status(syscall.ENOTEMPTY)
	}
	return f.Unlink(cancel, in, name)
}

// OpenDir provides a list of all the entries in the directory
func (f *Filesystem) OpenDir(_ <-chan struct{}, in *fuse.OpenIn, _ *fuse.OpenOut) fuse.Status {
	dir := f.GetNodeID(in.NodeId)
	if dir == nil {
		logging.Debug().Uint64("nodeID", in.NodeId).Msg("OpenDir: Directory not found")
		return fuse.ENOENT
	}
	id := dir.ID()
	if !dir.IsDir() {
		logging.Debug().Uint64("nodeID", in.NodeId).Str("id", id).Msg("OpenDir: Not a directory")
		return fuse.ENOTDIR
	}
	path := dir.Path()
	ctx := logging.DefaultLogger.With().
		Str("op", "OpenDir").
		Uint64("nodeID", in.NodeId).
		Str("id", id).
		Str("path", path).Logger()
	ctx.Debug().Msg("Starting OpenDir operation")

	ctx.Debug().Msg("About to call GetChildrenID")
	children, err := f.GetChildrenID(id, f.auth)
	ctx.Debug().Err(err).Int("childrenCount", len(children)).Msg("Returned from GetChildrenID")

	if err != nil {
		// not an item not found error (Lookup/Getattr will always be called
		// before Readdir()), something has happened to our connection
		logging.LogError(err, "Could not fetch children",
			logging.FieldOperation, "OpenDir",
			logging.FieldID, id,
			logging.FieldPath, path)
		return fuse.EREMOTEIO
	}

	ctx.Debug().Msg("Getting parent directory")
	parent := f.GetID(dir.ParentID())
	if parent == nil {
		// This is the parent of the mountpoint. The FUSE kernel module discards
		// this info, so what we put here doesn't actually matter.
		ctx.Debug().Msg("Parent is nil, creating dummy parent")
		parent = NewInode("..", 0755|fuse.S_IFDIR, nil)
		parent.nodeID = math.MaxUint64
	}

	ctx.Debug().Msg("Creating entries array")
	entries := make([]*Inode, 2)
	entries[0] = dir
	entries[1] = parent

	ctx.Debug().Int("childrenCount", len(children)).Msg("Adding children to entries")
	for _, child := range children {
		entries = append(entries, child)
	}

	ctx.Debug().Int("totalEntries", len(entries)).Msg("Storing entries in opendirs map")
	f.opendirsM.Lock()
	f.opendirs[in.NodeId] = entries
	f.opendirsM.Unlock()

	ctx.Debug().Msg("OpenDir operation completed successfully")
	return fuse.OK
}

// ReleaseDir closes a directory and purges it from memory
func (f *Filesystem) ReleaseDir(in *fuse.ReleaseIn) {
	f.opendirsM.Lock()
	delete(f.opendirs, in.NodeId)
	f.opendirsM.Unlock()
}

// readDirCommon contains the common code for ReadDir and ReadDirPlus
func (f *Filesystem) readDirCommon(cancel <-chan struct{}, in *fuse.ReadIn) ([]*Inode, fuse.Status) {
	f.opendirsM.RLock()
	entries, ok := f.opendirs[in.NodeId]
	f.opendirsM.RUnlock()
	if !ok {
		// readdir can sometimes arrive before the corresponding opendir, so we force it
		status := f.OpenDir(cancel, &fuse.OpenIn{InHeader: in.InHeader}, nil)
		if status != fuse.OK {
			return nil, status
		}
		f.opendirsM.RLock()
		entries, ok = f.opendirs[in.NodeId]
		f.opendirsM.RUnlock()
		if !ok {
			return nil, fuse.EBADF
		}
	}

	if in.Offset >= uint64(len(entries)) {
		// just tried to seek past end of directory, we're all done!
		return nil, fuse.OK
	}

	return entries, fuse.OK
}

// ReadDirPlus reads an individual directory entry AND does a lookup.
func (f *Filesystem) ReadDirPlus(cancel <-chan struct{}, in *fuse.ReadIn, out *fuse.DirEntryList) fuse.Status {
	entries, status := f.readDirCommon(cancel, in)
	if status != fuse.OK {
		return status
	}

	// Check if entries is nil or if offset is out of range
	if entries == nil || in.Offset >= uint64(len(entries)) {
		return fuse.OK
	}

	inode := entries[in.Offset]
	entry := fuse.DirEntry{
		Ino:  inode.NodeID(),
		Mode: inode.Mode(),
	}
	// first two entries will always be "." and ".."
	switch in.Offset {
	case 0:
		entry.Name = "."
	case 1:
		entry.Name = ".."
	default:
		entry.Name = inode.Name()
	}
	entryOut := out.AddDirLookupEntry(entry)
	if entryOut == nil {
		// Buffer is full, return OK to indicate we've provided as many entries as possible
		// The kernel will call ReadDirPlus again with a higher offset to get more entries
		logging.Debug().
			Uint64("nodeID", in.NodeId).
			Uint64("offset", in.Offset).
			Str("entryName", entry.Name).
			Msg("Directory entry buffer full, returning partial results")
		return fuse.OK
	}
	entryOut.NodeId = entry.Ino
	entryOut.Attr = inode.makeAttr()
	entryOut.SetAttrTimeout(timeout)
	entryOut.SetEntryTimeout(timeout)
	return fuse.OK
}

// ReadDir reads a directory entry. Usually doesn't get called (ReadDirPlus is
// typically used).
func (f *Filesystem) ReadDir(cancel <-chan struct{}, in *fuse.ReadIn, out *fuse.DirEntryList) fuse.Status {
	entries, status := f.readDirCommon(cancel, in)
	if status != fuse.OK {
		return status
	}

	// Check if entries is nil or if offset is out of range
	if entries == nil || in.Offset >= uint64(len(entries)) {
		return fuse.OK
	}

	inode := entries[in.Offset]
	entry := fuse.DirEntry{
		Ino:  inode.NodeID(),
		Mode: inode.Mode(),
	}
	// first two entries will always be "." and ".."
	switch in.Offset {
	case 0:
		entry.Name = "."
	case 1:
		entry.Name = ".."
	default:
		entry.Name = inode.Name()
	}

	out.AddDirEntry(entry)
	return fuse.OK
}

// Lookup is called by the kernel when the VFS wants to know about a file inside
// a directory.
func (f *Filesystem) Lookup(_ <-chan struct{}, in *fuse.InHeader, name string, out *fuse.EntryOut) fuse.Status {
	parent := f.GetNodeID(in.NodeId)
	if parent == nil {
		return fuse.ENOENT
	}
	id := parent.ID()
	logging.Trace().
		Str("op", "Lookup").
		Uint64("nodeID", in.NodeId).
		Str("id", id).
		Str("name", name).
		Msg("")

	child, _ := f.GetChild(id, strings.ToLower(name), f.auth)
	if child == nil {
		return fuse.ENOENT
	}

	out.NodeId = child.NodeID()
	out.Attr = child.makeAttr()
	out.SetAttrTimeout(timeout)
	out.SetEntryTimeout(timeout)
	return fuse.OK
}
