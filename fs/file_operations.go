package fs

import (
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog/log"
)

// Mknod creates a regular file. The server doesn't have this yet.
func (f *Filesystem) Mknod(cancel <-chan struct{}, in *fuse.MknodIn, name string, out *fuse.EntryOut) fuse.Status {
	if isNameRestricted(name) {
		return fuse.EINVAL
	}

	parentID := f.TranslateID(in.NodeId)
	if parentID == "" {
		return fuse.EBADF
	}

	parent := f.GetID(parentID)
	if parent == nil {
		return fuse.ENOENT
	}

	path := filepath.Join(parent.Path(), name)
	ctx := log.With().
		Str("op", "Mknod").
		Uint64("nodeID", in.NodeId).
		Str("path", path).
		Logger()
	if f.IsOffline() {
		// Instead of returning EROFS, log that we're in offline mode but allowing file creation
		ctx.Info().Msg("Operating in offline mode with write access. File creation will sync when online.")

		// Track this change for later synchronization
		change := &OfflineChange{
			ID:        parent.ID() + "-" + name, // Temporary ID until we get a real one
			Type:      "create",
			Timestamp: time.Now(),
			Path:      filepath.Join(parent.Path(), name),
		}
		f.TrackOfflineChange(change)
	}

	if child, _ := f.GetChild(parentID, name, f.auth); child != nil {
		return fuse.Status(syscall.EEXIST)
	}

	inode := NewInode(name, in.Mode, parent)
	ctx.Debug().
		Str("childID", inode.ID()).
		Str("mode", Octal(in.Mode)).
		Msg("Creating inode.")
	out.NodeId = f.InsertChild(parentID, inode)
	out.Attr = inode.makeAttr()
	out.SetAttrTimeout(timeout)
	out.SetEntryTimeout(timeout)
	return fuse.OK
}

// Create creates a regular file and opens it. The server doesn't have this yet.
func (f *Filesystem) Create(cancel <-chan struct{}, in *fuse.CreateIn, name string, out *fuse.CreateOut) fuse.Status {
	// we reuse mknod here
	result := f.Mknod(
		cancel,
		// we don't actually use the umask or padding here, so they don't get passed
		&fuse.MknodIn{
			InHeader: in.InHeader,
			Mode:     in.Mode,
		},
		name,
		&out.EntryOut,
	)
	if result == fuse.Status(syscall.EEXIST) {
		// if the inode already exists, we should truncate the existing file and
		// return the existing file inode as per "man creat"
		parentID := f.TranslateID(in.NodeId)
		child, _ := f.GetChild(parentID, name, f.auth)
		log.Debug().
			Str("op", "Create").
			Uint64("nodeID", in.NodeId).
			Str("id", parentID).
			Str("childID", child.ID()).
			Str("path", child.Path()).
			Str("mode", Octal(in.Mode)).
			Msg("Child inode already exists, truncating.")
		f.content.Delete(child.ID())
		f.content.Open(child.ID())
		child.DriveItem.Size = 0
		child.hasChanges = true
		return fuse.OK
	}
	// no further initialized required to open the file, it's empty
	return result
}

// Open fetches a Inodes's content and initializes the .Data field with actual
// data from the server.
func (f *Filesystem) Open(cancel <-chan struct{}, in *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return fuse.ENOENT
	}

	path := inode.Path()
	ctx := log.With().
		Str("op", "Open").
		Uint64("nodeID", in.NodeId).
		Str("id", id).
		Str("path", path).
		Logger()

	flags := int(in.Flags)
	if flags&os.O_RDWR+flags&os.O_WRONLY > 0 && f.IsOffline() {
		// Instead of returning EROFS, log that we're in offline mode but allowing writes
		ctx.Info().
			Bool("readWrite", flags&os.O_RDWR > 0).
			Bool("writeOnly", flags&os.O_WRONLY > 0).
			Msg("Operating in offline mode with write access. Changes will sync when online.")
	}

	ctx.Debug().Msg("")

	// we have something on disk-
	// verify content against what we're supposed to have
	inode.Lock()

	// try grabbing from disk
	fd, err := f.content.Open(id)
	if err != nil {
		inode.Unlock()
		ctx.Error().Err(err).Msg("Could not create cache file.")
		return fuse.EIO
	}

	if isLocalID(id) {
		// just use whatever's present if we're the only ones who have it
		inode.Unlock()
		return fuse.OK
	}

	if inode.VerifyChecksum(graph.QuickXORHashStream(fd)) {
		// disk content is only used if the checksums match
		ctx.Info().Msg("Found content in cache.")

		// we check size ourselves in case the API file sizes are WRONG (it happens)
		st, err := fd.Stat()
		if err != nil {
			inode.Unlock()
			ctx.Error().Err(err).Msg("Could not fetch file stats.")
			return fuse.EIO
		}
		inode.DriveItem.Size = uint64(st.Size())
		inode.Unlock()
		return fuse.OK
	}

	// Release the lock before network operations
	inode.Unlock()

	ctx.Info().Msg(
		"Not using cached item due to file hash mismatch, fetching content from API.",
	)

	// Queue the download in the background
	if _, err := f.downloads.QueueDownload(id); err != nil {
		ctx.Error().Err(err).Msg("Failed to queue download.")
		f.MarkFileError(id, err)
		return fuse.EIO
	}

	// For directory listing operations (like 'ls'), we don't want to block waiting for downloads
	// Check if this is a directory - if so, return immediately without waiting for download
	// This prevents the 'ls' command from hanging when there are pending downloads
	if inode.IsDir() {
		ctx.Debug().Msg("Non-blocking open for directory")
		// Update file status attributes but don't wait for download
		f.updateFileStatus(inode)
		return fuse.OK
	}

	// For actual file read/write operations, wait for the download to complete
	// This ensures we don't return until the file is available
	if err := f.downloads.WaitForDownload(id); err != nil {
		ctx.Error().Err(err).Msg("Download failed.")
		return fuse.EREMOTEIO
	}

	// Update file status attributes
	f.updateFileStatus(inode)

	return fuse.OK
}

// Unlink deletes a child file.
func (f *Filesystem) Unlink(cancel <-chan struct{}, in *fuse.InHeader, name string) fuse.Status {
	parentID := f.TranslateID(in.NodeId)
	child, _ := f.GetChild(parentID, name, nil)
	if child == nil {
		// the file we are unlinking never existed
		return fuse.ENOENT
	}

	id := child.ID()
	path := child.Path()
	ctx := log.With().
		Str("op", "Unlink").
		Uint64("nodeID", in.NodeId).
		Str("id", parentID).
		Str("childID", id).
		Str("path", path).
		Logger()

	if f.IsOffline() {
		// Instead of returning EROFS, log that we're in offline mode but allowing file deletion
		ctx.Info().Msg("Operating in offline mode with write access. File deletion will sync when online.")

		// Track this change for later synchronization
		change := &OfflineChange{
			ID:        id,
			Type:      "delete",
			Timestamp: time.Now(),
			Path:      path,
		}
		f.TrackOfflineChange(change)
	}

	ctx.Debug().Msg("Unlinking inode.")

	// if no ID, the item is local-only, and does not need to be deleted on the
	// server
	if !isLocalID(id) {
		if err := graph.Remove(id, f.auth); err != nil {
			ctx.Err(err).Msg("Failed to delete item on server. Aborting op.")
			return fuse.EREMOTEIO
		}
	}

	f.DeleteID(id)
	f.content.Delete(id)
	return fuse.OK
}

// Read an inode's data like a file.
func (f *Filesystem) Read(cancel <-chan struct{}, in *fuse.ReadIn, buf []byte) (fuse.ReadResult, fuse.Status) {
	inode := f.GetNodeID(in.NodeId)
	if inode == nil {
		return fuse.ReadResultData(make([]byte, 0)), fuse.EBADF
	}

	id := inode.ID()
	path := inode.Path()
	ctx := log.With().
		Str("op", "Read").
		Uint64("nodeID", in.NodeId).
		Str("id", id).
		Str("path", path).
		Int("bufsize", len(buf)).
		Logger()
	ctx.Trace().Msg("")

	fd, err := f.content.Open(id)
	if err != nil {
		ctx.Error().Err(err).Msg("Cache Open() failed.")
		return fuse.ReadResultData(make([]byte, 0)), fuse.EIO
	}

	// we are locked for the remainder of this op
	inode.RLock()
	defer inode.RUnlock()
	return fuse.ReadResultFd(fd.Fd(), int64(in.Offset), int(in.Size)), fuse.OK
}

// Write to an Inode like a file. Note that changes are 100% local until
// Flush() is called. Returns the number of bytes written and the status of the
// op.
func (f *Filesystem) Write(cancel <-chan struct{}, in *fuse.WriteIn, data []byte) (uint32, fuse.Status) {
	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return 0, fuse.EBADF
	}

	nWrite := len(data)
	offset := int(in.Offset)
	ctx := log.With().
		Str("op", "Write").
		Str("id", id).
		Uint64("nodeID", in.NodeId).
		Str("path", inode.Path()).
		Int("bufsize", nWrite).
		Int("offset", offset).
		Logger()
	ctx.Trace().Msg("")

	fd, err := f.content.Open(id)
	if err != nil {
		ctx.Error().Msg("Cache Open() failed.")
		return 0, fuse.EIO
	}

	// Get the path before acquiring the lock to avoid potential deadlocks
	path := inode.Path()

	inode.Lock()
	n, err := fd.WriteAt(data, int64(offset))
	if err != nil {
		inode.Unlock()
		ctx.Error().Err(err).Msg("Error during write")
		return uint32(n), fuse.EIO
	}

	st, _ := fd.Stat()
	inode.DriveItem.Size = uint64(st.Size())
	inode.hasChanges = true
	inode.Unlock()

	// Mark file as locally modified
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusLocalModified,
		Timestamp: time.Now(),
	})

	// Track this change if we're offline
	if f.IsOffline() {
		change := &OfflineChange{
			ID:        id,
			Type:      "modify",
			Timestamp: time.Now(),
			Path:      path,
		}
		f.TrackOfflineChange(change)
	}

	return uint32(n), fuse.OK
}

// Fsync is a signal to ensure writes to the Inode are flushed to stable
// storage. This method is used to trigger uploads of file content.
func (f *Filesystem) Fsync(cancel <-chan struct{}, in *fuse.FsyncIn) fuse.Status {
	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return fuse.EBADF
	}

	ctx := log.With().
		Str("op", "Fsync").
		Str("id", id).
		Uint64("nodeID", in.NodeId).
		Str("path", inode.Path()).
		Logger()
	ctx.Debug().Msg("")
	if inode.HasChanges() {
		inode.Lock()
		inode.hasChanges = false

		// recompute hashes when saving new content
		inode.DriveItem.File = &graph.File{}
		fd, err := f.content.Open(id)
		if err != nil {
			ctx.Error().Err(err).Msg("Could not get fd.")
		}
		fd.Sync()
		inode.DriveItem.File.Hashes.QuickXorHash = graph.QuickXORHashStream(fd)
		inode.Unlock()

		// Queue the upload in the background with high priority since it's a mount point request
		_, err = f.uploads.QueueUploadWithPriority(inode, PriorityHigh)
		if err != nil {
			ctx.Error().Err(err).Msg("Error creating upload session.")
			return fuse.EREMOTEIO
		}

		// Don't wait for the upload to complete, return immediately
		ctx.Debug().Str("id", id).Msg("File upload queued in background with high priority")
		return fuse.OK
	}
	return fuse.OK
}

// Flush is called when a file descriptor is closed. Uses Fsync() to perform file
// uploads. (Release not implemented because all cleanup is already done here).
func (f *Filesystem) Flush(cancel <-chan struct{}, in *fuse.FlushIn) fuse.Status {
	inode := f.GetNodeID(in.NodeId)
	if inode == nil {
		return fuse.EBADF
	}

	id := inode.ID()
	log.Trace().
		Str("op", "Flush").
		Str("id", id).
		Str("path", inode.Path()).
		Uint64("nodeID", in.NodeId).
		Msg("")
	f.Fsync(cancel, &fuse.FsyncIn{InHeader: in.InHeader})

	// grab a lock to prevent a race condition closing an opened file prior to its use (use after free segfault)
	inode.Lock()
	f.content.Close(id)
	inode.Unlock()

	// Update file status attributes after releasing the lock
	f.updateFileStatus(inode)
	return 0
}
