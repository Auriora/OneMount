package fs

import (
	"math"
	"path/filepath"
	"time"

	"github.com/auriora/onemount/internal/logging"

	"github.com/auriora/onemount/internal/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// StatFs Statfs returns information about the filesystem. Mainly useful for checking
// quotas and storage limits.
func (f *Filesystem) StatFs(_ <-chan struct{}, _ *fuse.InHeader, out *fuse.StatfsOut) fuse.Status {
	ctx := logging.DefaultLogger.With().Str("op", "StatFs").Logger()
	ctx.Debug().Msg("")
	drive, err := graph.GetDrive(f.auth)
	if err != nil {
		return fuse.EREMOTEIO
	}

	// Estimate file count from cached metadata
	estimatedFileCount := f.getEstimatedFileCount()

	if drive.DriveType == graph.DriveTypePersonal {
		// Throttle the warning to show only once every 5 minutes
		f.statfsWarningM.RLock()
		lastWarning := f.statfsWarningTime
		f.statfsWarningM.RUnlock()

		if time.Since(lastWarning) > 5*time.Minute {
			f.statfsWarningM.Lock()
			// Double-check in case another goroutine updated it
			if time.Since(f.statfsWarningTime) > 5*time.Minute {
				ctx.Warn().
					Uint64("estimatedFiles", estimatedFileCount).
					Msg("Personal OneDrive accounts do not show number of files, " +
						"using estimated count from local cache.")
				f.statfsWarningTime = time.Now()
			}
			f.statfsWarningM.Unlock()
		}
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

	// Use estimated file count for Personal OneDrive, actual count for Business
	if drive.DriveType == graph.DriveTypePersonal {
		out.Files = estimatedFileCount
		// Reserve some inodes for new files (10% or minimum 1000)
		reserved := estimatedFileCount / 10
		if reserved < 1000 {
			reserved = 1000
		}
		out.Ffree = reserved
	} else {
		out.Files = 100000
		out.Ffree = 100000 - drive.Quota.FileCount
	}

	out.NameLen = 260
	return fuse.OK
}

// getEstimatedFileCount estimates the total number of files and directories
// in the filesystem based on cached metadata. This provides a reasonable
// approximation for Personal OneDrive accounts where the API doesn't provide
// file counts.
func (f *Filesystem) getEstimatedFileCount() uint64 {
	var count uint64

	// Count items in the in-memory metadata cache
	f.metadata.Range(func(_, value interface{}) bool {
		count++
		return true
	})

	// If we have very few items in memory cache, fall back to a reasonable default
	// This can happen during startup or with limited cache
	if count < 10 {
		count = 1000 // Conservative default
	}

	return count
}

// GetAttr Getattr returns a the Inode as a UNIX stat. Holds the read mutex for all of
// the "metadata fetch" operations.

func (f *Filesystem) GetAttr(_ <-chan struct{}, in *fuse.GetAttrIn, out *fuse.AttrOut) fuse.Status {
	inode := f.GetNodeID(in.NodeId)
	if inode == nil {
		return fuse.ENOENT
	}
	id := inode.ID()
	logging.Trace().
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
	i.mu.Lock()

	ctx := logging.DefaultLogger.With().
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
		if i.IsVirtual() {
			if err := i.TruncateVirtualContent(size); err != nil {
				i.mu.Unlock()
				logging.LogError(err, "Failed to truncate virtual file",
					logging.FieldOperation, "SetAttr.truncate",
					logging.FieldID, i.DriveItem.ID,
					logging.FieldPath, path)
				return fuse.EIO
			}
			i.mu.Unlock()
			out.Attr = i.makeAttr()
			out.SetTimeout(timeout)
			return fuse.OK
		}
		fd, err := f.content.Open(i.DriveItem.ID)
		if err != nil {
			logging.LogError(err, "Failed to open file for truncation",
				logging.FieldID, i.DriveItem.ID,
				logging.FieldOperation, "SetAttr.truncate",
				logging.FieldPath, path)
			i.mu.Unlock()
			return fuse.EIO
		}
		// the unix syscall does not update the seek position, so neither should we
		if err := fd.Truncate(int64(size)); err != nil {
			logging.LogError(err, "Failed to truncate file",
				logging.FieldID, i.DriveItem.ID,
				logging.FieldOperation, "SetAttr.truncate",
				logging.FieldPath, path,
				"size", size)
			i.mu.Unlock()
			return fuse.EIO
		}
		i.DriveItem.Size = size
		i.hasChanges = true
	}

	i.mu.Unlock()
	out.Attr = i.makeAttr()
	out.SetTimeout(timeout)
	return fuse.OK
}

// Rename renames and/or moves an inode.
func (f *Filesystem) Rename(_ <-chan struct{}, in *fuse.RenameIn, name string, newName string) fuse.Status {
	if isNameRestricted(newName) {
		return fuse.EINVAL
	}

	oldParentItem := f.GetNodeID(in.NodeId)
	if oldParentItem == nil {
		return fuse.EBADF
	}
	oldParentID := oldParentItem.ID()
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
	if inode == nil {
		return fuse.ENOENT
	}

	id := inode.ID()
	newParentID := newParentItem.ID()

	ctx := logging.DefaultLogger.With().
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

	remoteID := ""
	if !isLocalID(id) {
		var err error
		remoteID, err = f.remoteID(inode)
		if err != nil {
			logging.LogError(err, "Failed to obtain remote ID for rename operation",
				logging.FieldOperation, "Rename",
				logging.FieldPath, path,
				"dest", dest,
				logging.FieldID, id)
			return fuse.EREMOTEIO
		}
	}

	// Check if there's already a file with the same name (case-insensitive) at the destination
	existingChild, _ := f.GetChild(newParentID, newName, f.auth)
	if existingChild != nil && existingChild.ID() != id {
		ctx.Info().
			Str("existingID", existingChild.ID()).
			Str("newName", newName).
			Msg("Found existing file with same name at destination, removing local entry")
		conflictID := existingChild.ID()
		f.DeleteID(conflictID)
		if !isLocalID(conflictID) && !f.IsOffline() {
			f.queueRemoteDelete(conflictID)
		}
	}

	if err := f.MovePath(oldParentID, newParentID, name, newName, f.auth); err != nil {
		logging.LogError(err, "Failed to rename local item",
			logging.FieldOperation, "Rename.localRename",
			logging.FieldID, id,
			logging.FieldPath, path,
			"dest", dest,
			"oldParentID", oldParentID,
			"newParentID", newParentID,
			"name", name,
			"newName", newName)
		return fuse.EIO
	}
	f.markDirtyLocalState(id)

	if f.IsOffline() || isLocalID(id) || isLocalID(newParentID) {
		change := &OfflineChange{
			ID:        id,
			Type:      "rename",
			Timestamp: time.Now(),
			OldPath:   path,
			NewPath:   dest,
		}
		if err := f.TrackOfflineChange(change); err != nil {
			logging.Warn().Err(err).Msg("Failed to record offline rename change")
		}
		return fuse.OK
	}

	if remoteID != "" {
		f.queueRemoteRename(remoteID, newParentID, newName)
	}

	return fuse.OK
}
