package fs

import (
	"context"
	"github.com/auriora/onemount/internal/logging"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// HandleThumbnailRequest handles a request for a thumbnail.
// This is called when a file with a special extension is accessed.
// For example, if a file is accessed as "image.jpg.thumbnail.small",
// this function will return the small thumbnail for image.jpg.
//
// The cancel parameter is required by the FUSE interface but not used in this implementation
// as thumbnail operations are typically quick and don't need cancellation.
func (f *Filesystem) HandleThumbnailRequest(_ <-chan struct{}, in *fuse.OpenIn, name string, out *fuse.OpenOut) (fuse.Status, uint64) {
	// Parse the thumbnail request from the filename
	originalPath, size, ok := parseThumbnailRequest(name)
	if !ok {
		return fuse.ENOENT, 0
	}

	// Get the parent directory
	parent := f.GetNodeID(in.NodeId)
	if parent == nil {
		if f.TranslateID(in.NodeId) == "" {
			return fuse.EBADF, 0
		}
		return fuse.ENOENT, 0
	}

	// Construct the full path to the original file
	fullPath := filepath.Join(parent.Path(), originalPath)

	// Log the thumbnail request
	ctx := logging.DefaultLogger.With().
		Str("op", "HandleThumbnailRequest").
		Str("path", fullPath).
		Str("size", size).
		Logger()
	ctx.Debug().Msg("Handling thumbnail request")

	// Create a temporary file to store the thumbnail
	tempFile, err := os.CreateTemp("", "onemount-thumbnail-*")
	if err != nil {
		ctx.Error().Err(err).Msg("Failed to create temporary file for thumbnail")
		return fuse.EIO, 0
	}

	// Get the thumbnail and write it to the temporary file
	err = f.GetThumbnailStream(fullPath, size, tempFile)
	if err != nil {
		ctx.Error().Err(err).Msg("Failed to get thumbnail")
		if closeErr := tempFile.Close(); closeErr != nil {
			ctx.Error().Err(closeErr).Msg("Failed to close temporary file")
		}
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			ctx.Error().Err(removeErr).Msg("Failed to remove temporary file")
		}
		return fuse.EIO, 0
	}

	// Reset the file position to the beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		ctx.Error().Err(err).Msg("Failed to reset file position")
		if closeErr := tempFile.Close(); closeErr != nil {
			ctx.Error().Err(closeErr).Msg("Failed to close temporary file")
		}
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			ctx.Error().Err(removeErr).Msg("Failed to remove temporary file")
		}
		return fuse.EIO, 0
	}

	// Create a new file handle for the thumbnail
	fh := &ThumbnailFileHandle{
		file:     tempFile,
		path:     tempFile.Name(),
		size:     size,
		origPath: fullPath,
	}

	// Register the file handle
	handleID := f.RegisterFileHandle(fh)
	out.Fh = handleID

	return fuse.OK, handleID
}

// ThumbnailFileHandle represents a handle to a thumbnail file
type ThumbnailFileHandle struct {
	file     *os.File
	path     string
	size     string
	origPath string
}

// Read reads data from the thumbnail file
// The ctx parameter is required by the FUSE interface but not used in this implementation
func (fh *ThumbnailFileHandle) Read(_ context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	// Seek to the offset
	if _, err := fh.file.Seek(off, 0); err != nil {
		logging.Error().Err(err).
			Str("path", fh.path).
			Str("size", fh.size).
			Str("origPath", fh.origPath).
			Msg("Failed to seek in thumbnail file")
		return nil, syscall.EIO
	}

	// Read the data
	n, err := fh.file.Read(dest)
	if err != nil && err != io.EOF {
		logging.Error().Err(err).
			Str("path", fh.path).
			Str("size", fh.size).
			Str("origPath", fh.origPath).
			Msg("Failed to read from thumbnail file")
		return nil, syscall.EIO
	}

	return fuse.ReadResultData(dest[:n]), 0
}

// Release closes the thumbnail file and removes the temporary file
// The ctx parameter is required by the FUSE interface but not used in this implementation
func (fh *ThumbnailFileHandle) Release(_ context.Context) syscall.Errno {
	// Close the file
	if err := fh.file.Close(); err != nil {
		logging.Error().Err(err).
			Str("path", fh.path).
			Str("size", fh.size).
			Str("origPath", fh.origPath).
			Msg("Failed to close thumbnail file")
	}

	// Remove the temporary file
	if err := os.Remove(fh.path); err != nil {
		logging.Error().Err(err).
			Str("path", fh.path).
			Str("size", fh.size).
			Str("origPath", fh.origPath).
			Msg("Failed to remove temporary thumbnail file")
	}

	return 0
}

// parseThumbnailRequest parses a thumbnail request from a filename.
// Returns the original path, size, and a boolean indicating if the request is valid.
// Valid formats:
// - filename.ext.thumbnail.small
// - filename.ext.thumbnail.medium
// - filename.ext.thumbnail.large
func parseThumbnailRequest(name string) (string, string, bool) {
	// Check if the name ends with a thumbnail extension
	if !strings.HasSuffix(name, ".thumbnail.small") &&
		!strings.HasSuffix(name, ".thumbnail.medium") &&
		!strings.HasSuffix(name, ".thumbnail.large") {
		return "", "", false
	}

	// Extract the size
	var size string
	if strings.HasSuffix(name, ".thumbnail.small") {
		size = "small"
		name = strings.TrimSuffix(name, ".thumbnail.small")
	} else if strings.HasSuffix(name, ".thumbnail.medium") {
		size = "medium"
		name = strings.TrimSuffix(name, ".thumbnail.medium")
	} else if strings.HasSuffix(name, ".thumbnail.large") {
		size = "large"
		name = strings.TrimSuffix(name, ".thumbnail.large")
	}

	return name, size, true
}

// fileHandles stores thumbnail file handles by ID
var fileHandles sync.Map

// nextHandleID is the next handle ID to assign
var nextHandleID uint64 = 1

// handleIDLock protects nextHandleID
var handleIDLock sync.Mutex

// RegisterFileHandle registers a file handle and returns a handle ID
func (f *Filesystem) RegisterFileHandle(fh *ThumbnailFileHandle) uint64 {
	// Get a unique handle ID
	handleIDLock.Lock()
	handleID := nextHandleID
	nextHandleID++
	handleIDLock.Unlock()

	// Store the file handle
	fileHandles.Store(handleID, fh)

	return handleID
}

// GetFileHandle gets a file handle by ID
func (f *Filesystem) GetFileHandle(handleID uint64) *ThumbnailFileHandle {
	if handle, ok := fileHandles.Load(handleID); ok {
		return handle.(*ThumbnailFileHandle)
	}
	return nil
}

// ReleaseFileHandle releases a file handle
func (f *Filesystem) ReleaseFileHandle(handleID uint64) {
	fileHandles.Delete(handleID)
}
