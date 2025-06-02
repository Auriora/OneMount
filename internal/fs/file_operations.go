package fs

import (
	"context"
	"github.com/auriora/onemount/pkg/logging"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Mknod creates a regular file. The server doesn't have this yet.
func (f *Filesystem) Mknod(_ <-chan struct{}, in *fuse.MknodIn, name string, out *fuse.EntryOut) fuse.Status {
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
	ctx := logging.DefaultLogger.With().
		Str("op", "Mknod").
		Uint64("nodeID", in.NodeId).
		Str("path", path).
		Logger()
	if f.IsOffline() {
		ctx.Info().Msg("File creation in offline mode will be cached locally")
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
		logger := logging.Debug().
			Str("op", "Create").
			Uint64("nodeID", in.NodeId).
			Str("id", parentID).
			Str("childID", child.ID()).
			Str("path", child.Path()).
			Str("mode", Octal(in.Mode))
		logger.Msg("Child inode already exists, truncating.")

		if err := f.content.Delete(child.ID()); err != nil {
			logging.Error().Err(err).Str("id", child.ID()).Msg("Failed to delete existing file content")
		}
		if _, err := f.content.Open(child.ID()); err != nil {
			logging.Error().Err(err).Str("id", child.ID()).Msg("Failed to open file for writing")
		}
		child.DriveItem.Size = 0
		child.hasChanges = true
		return fuse.OK
	}
	// no further initialized required to open the file, it's empty
	return result
}

// Open handles file open operations for the FUSE filesystem.
// This method is called when a file is opened by the kernel. It verifies if the file
// content is available locally, and if not or if the checksum doesn't match, it queues
// a download from OneDrive. For directories, it performs a non-blocking open to prevent
// 'ls' commands from hanging when there are pending downloads.
//
// The method handles both online and offline modes. In offline mode, it allows write
// operations but logs that changes will sync when online.
//
// Parameters:
//   - cancel: Channel that signals if the operation should be canceled
//   - in: Input parameters for the open operation, including node ID and flags
//   - out: Output parameters for the open operation
//
// Returns:
//   - fuse.OK if the file was opened successfully
//   - fuse.ENOENT if the file doesn't exist
//   - fuse.EIO if there was an error creating the cache file
//   - fuse.EREMOTEIO if the download failed
func (f *Filesystem) Open(cancel <-chan struct{}, in *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
	methodName, startTime := logging.LogMethodEntry("Open", in.NodeId)

	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), fuse.ENOENT)
		}()
		return fuse.ENOENT
	}

	// Check if this is a thumbnail request
	name := inode.Name()
	if _, _, ok := parseThumbnailRequest(name); ok {
		// This is a thumbnail request, handle it
		status, handleID := f.HandleThumbnailRequest(cancel, in, name, out)
		if status == fuse.OK {
			out.Fh = handleID
		}
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), status)
		}()
		return status
	}

	path := inode.Path()
	// Create a context for this operation with request ID, user ID, and path
	logCtx := logging.NewLogContextWithRequestAndUserID("file_open").
		WithPath(path)

	logger := logging.WithLogContext(logCtx)

	flags := int(in.Flags)
	if flags&os.O_RDWR+flags&os.O_WRONLY > 0 && f.IsOffline() {
		logger.Info().
			Bool("readWrite", flags&os.O_RDWR > 0).
			Bool("writeOnly", flags&os.O_WRONLY > 0).
			Msg("Write operations in offline mode will be cached locally")
	}

	if logging.IsDebugEnabled() {
		logger.Debug().
			Uint64("nodeID", in.NodeId).
			Str(logging.FieldID, id).
			Msg("Opening file")
	}

	// we have something on disk-
	// verify content against what we're supposed to have
	inode.Lock()

	// try grabbing from disk
	fd, err := f.content.Open(id)
	if err != nil {
		inode.Unlock()
		logging.LogErrorWithContext(err, logCtx, "Could not create cache file",
			logging.FieldID, id,
			logging.FieldPath, path)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), fuse.EIO)
		}()
		return fuse.EIO
	}

	if isLocalID(id) {
		// just use whatever's present if we're the only ones who have it
		inode.Unlock()
		return fuse.OK
	}

	// If we're in offline mode, use the cached content regardless of checksum
	if f.IsOffline() {
		logger.Info().Msg("Using cached content in offline mode regardless of checksum")

		// we check size ourselves in case the API file sizes are WRONG (it happens)
		st, err := fd.Stat()
		if err != nil {
			inode.Unlock()
			logging.LogErrorWithContext(err, logCtx, "Could not fetch file stats",
				logging.FieldID, id,
				logging.FieldPath, path)
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), fuse.EIO)
			}()
			return fuse.EIO
		}
		inode.DriveItem.Size = uint64(st.Size())
		inode.Unlock()
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), fuse.OK)
		}()
		return fuse.OK
	}

	if inode.VerifyChecksum(graph.QuickXORHashStream(fd)) {
		// disk content is only used if the checksums match
		logger.Info().Msg("Found content in cache")

		// we check size ourselves in case the API file sizes are WRONG (it happens)
		st, err := fd.Stat()
		if err != nil {
			inode.Unlock()
			logging.LogErrorWithContext(err, logCtx, "Could not fetch file stats",
				logging.FieldID, id,
				logging.FieldPath, path)
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), fuse.EIO)
			}()
			return fuse.EIO
		}
		inode.DriveItem.Size = uint64(st.Size())
		inode.Unlock()
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), fuse.OK)
		}()
		return fuse.OK
	}

	// Release the lock before network operations
	inode.Unlock()

	logger.Info().Msg("Not using cached item due to file hash mismatch, fetching content from API")

	// Queue the download in the background
	if _, err := f.downloads.QueueDownload(id); err != nil {
		logging.LogErrorWithContext(err, logCtx, "Failed to queue download",
			logging.FieldID, id,
			logging.FieldPath, path)
		f.MarkFileError(id, err)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), fuse.EIO)
		}()
		return fuse.EIO
	}

	// For directory listing operations (like 'ls'), we don't want to block waiting for downloads
	// Check if this is a directory - if so, return immediately without waiting for download
	// This prevents the 'ls' command from hanging when there are pending downloads
	if inode.IsDir() {
		if logging.IsDebugEnabled() {
			logger.Debug().Msg("Non-blocking open for directory")
		}
		// Update file status attributes but don't wait for download
		f.updateFileStatus(inode)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), fuse.OK)
		}()
		return fuse.OK
	}

	// For actual file read/write operations, wait for the download to complete
	// This ensures we don't return until the file is available
	if err := f.downloads.WaitForDownload(id); err != nil {
		logging.LogErrorWithContext(err, logCtx, "Download failed",
			logging.FieldID, id,
			logging.FieldPath, path)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), fuse.EREMOTEIO)
		}()
		return fuse.EREMOTEIO
	}

	// Update file status attributes
	f.updateFileStatus(inode)

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), fuse.OK)
	}()
	return fuse.OK
}

// Unlink deletes a child file.
func (f *Filesystem) Unlink(_ <-chan struct{}, in *fuse.InHeader, name string) fuse.Status {
	parentID := f.TranslateID(in.NodeId)
	child, _ := f.GetChild(parentID, name, f.auth)
	if child == nil {
		// the file we are unlinking never existed
		return fuse.ENOENT
	}

	id := child.ID()
	path := child.Path()
	ctx := logging.DefaultLogger.With().
		Str("op", "Unlink").
		Uint64("nodeID", in.NodeId).
		Str("id", parentID).
		Str("childID", id).
		Str("path", path).
		Logger()

	if f.IsOffline() {
		ctx.Info().Msg("File deletion in offline mode will be cached locally")
	}

	ctx.Debug().Msg("Unlinking inode.")

	// if no ID, the item is local-only, and does not need to be deleted on the
	// server
	if !isLocalID(id) {
		if err := graph.Remove(id, f.auth); err != nil {
			ctx.Error().Err(err).Msg("Failed to delete item on server. Aborting op.")
			return fuse.EREMOTEIO
		}
	}

	f.DeleteID(id)
	if err := f.content.Delete(id); err != nil {
		ctx.Error().Err(err).Str("id", id).Msg("Failed to delete file content")
	}
	return fuse.OK
}

// Read handles file read operations for the FUSE filesystem.
// This method is called when a file's content is read by the kernel. It retrieves
// the file content from the local cache and returns it to the caller. The method
// uses file descriptor passing for efficient data transfer to the kernel.
//
// Parameters:
//   - cancel: Channel that signals if the operation should be canceled
//   - in: Input parameters for the read operation, including node ID, offset, and size
//   - buf: Buffer to store the read data
//
// Returns:
//   - fuse.ReadResult: A read result object containing the data or file descriptor
//   - fuse.OK if the read was successful
//   - fuse.EBADF if the inode doesn't exist
//   - fuse.EIO if there was an error opening the cache file
func (f *Filesystem) Read(cancel <-chan struct{}, in *fuse.ReadIn, buf []byte) (fuse.ReadResult, fuse.Status) {
	methodName, startTime := logging.LogMethodEntry("Read", in.NodeId, len(buf))

	// Check for cancellation before starting the read operation
	select {
	case <-cancel:
		logging.Debug().Msg("Read operation cancelled")
		emptyResult := fuse.ReadResultData(make([]byte, 0))
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), emptyResult, fuse.EINTR)
		}()
		return emptyResult, fuse.EINTR
	default:
	}

	// Check if this is a thumbnail file handle
	if in.Fh != 0 {
		// Get the file handle
		fh := f.GetFileHandle(in.Fh)
		if fh != nil {
			// This is a thumbnail file handle, use its Read method
			ctx := context.Background()
			result, errno := fh.Read(ctx, buf, int64(in.Offset))
			if errno != 0 {
				defer func() {
					logging.LogMethodExit(methodName, time.Since(startTime), nil, fuse.Status(errno))
				}()
				return nil, fuse.Status(errno)
			}
			defer func() {
				logging.LogMethodExit(methodName, time.Since(startTime), result, fuse.OK)
			}()
			return result, fuse.OK
		}
	}

	// Regular file read
	inode := f.GetNodeID(in.NodeId)
	if inode == nil {
		emptyResult := fuse.ReadResultData(make([]byte, 0))
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), emptyResult, fuse.EBADF)
		}()
		return emptyResult, fuse.EBADF
	}

	id := inode.ID()
	path := inode.Path()

	// Create a context for this operation with request ID, user ID, and path
	logCtx := logging.NewLogContextWithRequestAndUserID("file_read").
		WithPath(path)

	logger := logging.WithLogContext(logCtx)

	if logging.IsTraceEnabled() {
		logger.Trace().
			Uint64("nodeID", in.NodeId).
			Str(logging.FieldID, id).
			Int(logging.FieldSize, len(buf)).
			Int(logging.FieldOffset, int(in.Offset)).
			Msg("Reading file")
	}

	fd, err := f.content.Open(id)
	if err != nil {
		logging.LogErrorWithContext(err, logCtx, "Cache Open() failed",
			logging.FieldOperation, "file_read",
			logging.FieldID, id,
			logging.FieldPath, path)
		emptyResult := fuse.ReadResultData(make([]byte, 0))
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), emptyResult, fuse.EIO)
		}()
		return emptyResult, fuse.EIO
	}

	// we are locked for the remainder of this op
	inode.RLock()
	defer inode.RUnlock()

	result := fuse.ReadResultFd(fd.Fd(), int64(in.Offset), int(in.Size))
	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), result, fuse.OK)
	}()
	return result, fuse.OK
}

// Write handles file write operations for the FUSE filesystem.
// This method is called when data is written to a file by the kernel. It writes
// the data to the local cache and marks the file as modified. The changes remain
// local until Flush() is called, which triggers synchronization with OneDrive.
// In offline mode, the changes are tracked for later synchronization when the
// filesystem goes online.
//
// Parameters:
//   - cancel: Channel that signals if the operation should be canceled
//   - in: Input parameters for the write operation, including node ID and offset
//   - data: The data to be written to the file
//
// Returns:
//   - uint32: The number of bytes written
//   - fuse.OK if the write was successful
//   - fuse.EBADF if the inode doesn't exist
//   - fuse.EIO if there was an error writing to the cache file
func (f *Filesystem) Write(cancel <-chan struct{}, in *fuse.WriteIn, data []byte) (uint32, fuse.Status) {
	methodName, startTime := logging.LogMethodEntry("Write", in.NodeId, len(data), in.Offset)

	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), int32(fuse.EBADF))
		}()
		return 0, fuse.EBADF
	}

	nWrite := len(data)
	offset := int(in.Offset)
	path := inode.Path()

	// Create a context for this operation with request ID, user ID, and path
	logCtx := logging.NewLogContextWithRequestAndUserID("file_write").
		WithPath(path)

	logger := logging.WithLogContext(logCtx)

	// Check for cancellation before starting the write operation
	select {
	case <-cancel:
		logger.Debug().Msg("Write operation cancelled")
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), int32(fuse.EINTR))
		}()
		return 0, fuse.EINTR
	default:
	}

	if logging.IsTraceEnabled() {
		logger.Trace().
			Uint64("nodeID", in.NodeId).
			Str(logging.FieldID, id).
			Int(logging.FieldSize, nWrite).
			Int(logging.FieldOffset, offset).
			Msg("Writing to file")
	}

	// In offline mode, we allow writes but they will be cached locally
	if f.IsOffline() {
		logger.Info().Msg("Write operations in offline mode will be cached locally")
	}

	// Check for large file operations and log warnings
	const largeFileThreshold = 1024 * 1024 * 1024 // 1GB
	if nWrite > largeFileThreshold {
		logger.Warn().
			Int("writeSize", nWrite).
			Msg("Large write operation detected - this may take some time")
	}

	fd, err := f.content.Open(id)
	if err != nil {
		logging.LogErrorWithContext(err, logCtx, "Cache Open() failed",
			logging.FieldOperation, "file_write",
			logging.FieldID, id,
			logging.FieldPath, path)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), int32(fuse.EIO))
		}()
		return 0, fuse.EIO
	}

	inode.Lock()
	n, err := fd.WriteAt(data, int64(offset))
	if err != nil {
		inode.Unlock()
		logging.LogErrorWithContext(err, logCtx, "Error during write",
			logging.FieldOperation, "file_write",
			logging.FieldID, id,
			logging.FieldPath, path,
			logging.FieldOffset, offset,
			logging.FieldSize, nWrite)
		defer func() {
			logging.LogMethodExit(methodName, time.Since(startTime), uint32(n), int32(fuse.EIO))
		}()
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

	if logging.IsDebugEnabled() {
		logger.Debug().
			Int("bytesWritten", n).
			Uint64("newSize", inode.DriveItem.Size).
			Msg("File write completed successfully")
	}

	defer func() {
		logging.LogMethodExit(methodName, time.Since(startTime), uint32(n), int32(fuse.OK))
	}()
	return uint32(n), fuse.OK
}

// Fsync is a signal to ensure writes to the Inode are flushed to stable
// storage. This method is used to trigger uploads of file content.
func (f *Filesystem) Fsync(_ <-chan struct{}, in *fuse.FsyncIn) fuse.Status {
	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return fuse.EBADF
	}

	ctx := logging.DefaultLogger.With().
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
			logging.LogError(err, "Could not get fd",
				logging.FieldOperation, "Fsync",
				logging.FieldID, id,
				logging.FieldPath, inode.Path())
		} else {
			if err := fd.Sync(); err != nil {
				logging.LogError(err, "Failed to sync file to disk",
					logging.FieldOperation, "Fsync",
					logging.FieldID, id,
					logging.FieldPath, inode.Path())
			}
		}
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
	// Check if this is a thumbnail file handle
	if in.Fh != 0 {
		// Get the file handle
		fh := f.GetFileHandle(in.Fh)
		if fh != nil {
			// This is a thumbnail file handle, but we don't need to do anything special
			// The file will be cleaned up in Release
			return fuse.OK
		}
	}

	// Regular file flush
	inode := f.GetNodeID(in.NodeId)
	if inode == nil {
		return fuse.EBADF
	}

	id := inode.ID()
	logging.Trace().
		Str("op", "Flush").
		Str("id", id).
		Str("path", inode.Path()).
		Uint64("nodeID", in.NodeId).
		Msg("")
	f.Fsync(cancel, &fuse.FsyncIn{InHeader: in.InHeader})

	// grab a lock to prevent a race condition closing an opened file prior to its use (use after free segfault)
	inode.Lock()
	if err := f.content.Close(id); err != nil {
		logging.Error().Err(err).Str("id", id).Str("path", inode.Path()).Msg("Failed to close file")
	}
	inode.Unlock()

	// Update file status attributes after releasing the lock
	f.updateFileStatus(inode)
	return 0
}
