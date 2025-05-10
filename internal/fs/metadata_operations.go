package fs

import (
	"math"
	"path/filepath"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/rs/zerolog/log"
)

// StatFs Statfs returns information about the filesystem. Mainly useful for checking
// quotas and storage limits.
func (f *Filesystem) StatFs(_ <-chan struct{}, _ *fuse.InHeader, out *fuse.StatfsOut) fuse.Status {
	ctx := log.With().Str("op", "StatFs").Logger()
	ctx.Debug().Msg("")
	drive, err := graph.GetDrive(f.auth)
	if err != nil {
		return fuse.EREMOTEIO
	}

	if drive.DriveType == graph.DriveTypePersonal {
		ctx.Warn().Msg("Personal OneDrive accounts do not show number of files, " +
			"inode counts reported by onemount will be bogus.")
	} else if drive.Quota.Total == 0 { // <-- check for if microsoft ever fixes their API
		ctx.Warn().Msg("OneDrive for Business accounts do not report quotas, " +
			"pretending the quota is 5TB and it's all unused.")
		drive.Quota.Total = 5 * uint64(math.Pow(1024, 4))
		drive.Quota.Remaining = 5 * uint64(math.Pow(1024, 4))
		drive.Quota.FileCount = 0
	}

	// limits are pasted from https://support.microsoft.com/en-us/help/3125202
	const blkSize uint64 = 4096 // default ext4 block size
	out.Bsize = uint32(blkSize)
	out.Blocks = drive.Quota.Total / blkSize
	out.Bfree = drive.Quota.Remaining / blkSize
	out.Bavail = drive.Quota.Remaining / blkSize
	out.Files = 100000
	out.Ffree = 100000 - drive.Quota.FileCount
	out.NameLen = 260
	return fuse.OK
}

// GetAttr Getattr returns a the Inode as a UNIX stat. Holds the read mutex for all of
// the "metadata fetch" operations.
func (f *Filesystem) GetAttr(_ <-chan struct{}, in *fuse.GetAttrIn, out *fuse.AttrOut) fuse.Status {
	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return fuse.ENOENT
	}
	log.Trace().
		Str("op", "GetAttr").
		Uint64("nodeID", in.NodeId).
		Str("id", id).
		Str("path", inode.Path()).
		Msg("")

	out.Attr = inode.makeAttr()
	out.SetTimeout(timeout)
	return fuse.OK
}

// SetAttr Setattr is the workhorse for setting filesystem attributes. Does the work of
// operations like utimens, chmod, chown (not implemented, FUSE is single-user),
// and truncate.
func (f *Filesystem) SetAttr(_ <-chan struct{}, in *fuse.SetAttrIn, out *fuse.AttrOut) fuse.Status {
	i := f.GetNodeID(in.NodeId)
	if i == nil {
		return fuse.ENOENT
	}
	path := i.Path()
	isDir := i.IsDir() // holds an rlock
	i.Lock()

	ctx := log.With().
		Str("op", "SetAttr").
		Uint64("nodeID", in.NodeId).
		Str("id", i.DriveItem.ID).
		Str("path", path).
		Logger()

	// utimens
	if mtime, valid := in.GetMTime(); valid {
		ctx.Info().
			Str("subop", "utimens").
			Time("oldMtime", *i.DriveItem.ModTime).
			Time("newMtime", *i.DriveItem.ModTime).
			Msg("")
		i.DriveItem.ModTime = &mtime
	}

	// chmod
	if mode, valid := in.GetMode(); valid {
		ctx.Info().
			Str("subop", "chmod").
			Str("oldMode", Octal(i.mode)).
			Str("newMode", Octal(mode)).
			Msg("")
		if isDir {
			i.mode = fuse.S_IFDIR | mode
		} else {
			i.mode = fuse.S_IFREG | mode
		}
	}

	// truncate
	if size, valid := in.GetSize(); valid {
		ctx.Info().
			Str("subop", "truncate").
			Uint64("oldSize", i.DriveItem.Size).
			Uint64("newSize", size).
			Msg("")
		fd, err := f.content.Open(i.DriveItem.ID)
		if err != nil {
			errors.LogError(err, "Failed to open file for truncation",
				errors.FieldID, i.DriveItem.ID,
				errors.FieldOperation, "SetAttr.truncate",
				errors.FieldPath, path)
			i.Unlock()
			return fuse.EIO
		}
		// the unix syscall does not update the seek position, so neither should we
		if err := fd.Truncate(int64(size)); err != nil {
			errors.LogError(err, "Failed to truncate file",
				errors.FieldID, i.DriveItem.ID,
				errors.FieldOperation, "SetAttr.truncate",
				errors.FieldPath, path,
				"size", size)
			i.Unlock()
			return fuse.EIO
		}
		i.DriveItem.Size = size
		i.hasChanges = true
	}

	i.Unlock()
	out.Attr = i.makeAttr()
	out.SetTimeout(timeout)
	return fuse.OK
}

// Rename renames and/or moves an inode.
func (f *Filesystem) Rename(_ <-chan struct{}, in *fuse.RenameIn, name string, newName string) fuse.Status {
	if isNameRestricted(newName) {
		return fuse.EINVAL
	}

	oldParentID := f.TranslateID(in.NodeId)
	oldParentItem := f.GetNodeID(in.NodeId)
	if oldParentID == "" || oldParentItem == nil {
		return fuse.EBADF
	}
	path := filepath.Join(oldParentItem.Path(), name)

	// we'll have the metadata for the dest inode already so it is not necessary
	// to use GetPath() to prefetch it. In order for the fs to know about this
	// inode, it has already fetched all of the inodes up to the new destination.
	newParentItem := f.GetNodeID(in.Newdir)
	if newParentItem == nil {
		return fuse.ENOENT
	}
	dest := filepath.Join(newParentItem.Path(), newName)

	inode, _ := f.GetChild(oldParentID, name, f.auth)
	id, err := f.remoteID(inode)
	newParentID := newParentItem.ID()

	ctx := log.With().
		Str("op", "Rename").
		Str("id", id).
		Str("parentID", newParentID).
		Str("path", path).
		Str("dest", dest).
		Logger()
	ctx.Info().
		Uint64("srcNodeID", in.NodeId).
		Uint64("dstNodeID", in.Newdir).
		Msg("")

	if isLocalID(id) || err != nil {
		// uploads will fail without an id
		errors.LogError(err, "ID of item to move cannot be local and we failed to obtain an ID",
			errors.FieldOperation, "Rename",
			errors.FieldPath, path,
			"dest", dest,
			errors.FieldID, id)
		return fuse.EREMOTEIO
	}

	// Check if there's already a file with the same name (case-insensitive) at the destination
	existingChild, _ := f.GetChild(newParentID, newName, f.auth)
	if existingChild != nil && existingChild.ID() != id {
		// There's already a different file with the same name (case-insensitive)
		// We need to remove it before we can rename our file to this name
		ctx.Info().
			Str("existingID", existingChild.ID()).
			Str("newName", newName).
			Msg("Found existing file with same name (case-insensitive) at destination, removing it first")

		// Remove the existing file
		if err = graph.Remove(existingChild.ID(), f.auth); err != nil {
			errors.LogError(err, "Failed to remove existing file at destination",
				errors.FieldOperation, "Rename.removeExisting",
				errors.FieldID, existingChild.ID(),
				errors.FieldPath, dest)
			return fuse.EREMOTEIO
		}

		// Also remove it from the local cache
		f.DeleteID(existingChild.ID())
	}

	// perform remote rename
	if err = graph.Rename(id, newName, newParentID, f.auth); err != nil {
		errors.LogError(err, "Failed to rename remote item",
			errors.FieldOperation, "Rename.remoteRename",
			errors.FieldID, id,
			errors.FieldPath, path,
			"dest", dest,
			"newName", newName,
			"newParentID", newParentID)
		return fuse.EREMOTEIO
	}

	// now rename local copy
	if err = f.MovePath(oldParentID, newParentID, name, newName, f.auth); err != nil {
		errors.LogError(err, "Failed to rename local item",
			errors.FieldOperation, "Rename.localRename",
			errors.FieldID, id,
			errors.FieldPath, path,
			"dest", dest,
			"oldParentID", oldParentID,
			"newParentID", newParentID,
			"name", name,
			"newName", newName)
		return fuse.EIO
	}

	// whew! item renamed
	return fuse.OK
}
