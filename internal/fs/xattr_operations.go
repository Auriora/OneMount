package fs

import (
	"syscall"
	"time"

	"github.com/auriora/onemount/internal/logging"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// GetXAttr retrieves the value of an extended attribute.
func (f *Filesystem) GetXAttr(_ <-chan struct{}, header *fuse.InHeader, name string, buf []byte) (uint32, fuse.Status) {
	methodName, startTime := logging.LogMethodEntry("GetXAttr", header.NodeId, name, len(buf))

	id := f.TranslateID(header.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), fuse.ENOENT)
		return 0, fuse.ENOENT
	}

	// Create a context for this operation with request ID and user ID
	ctx := logging.NewLogContextWithRequestAndUserID("xattr_operations").
		WithPath(inode.Path()).
		With("id", id).
		With("nodeID", header.NodeId).
		With("name", name)

	// Get a logger with the context
	logger := ctx.Logger()

	inode.RLock()
	defer inode.RUnlock()

	value, exists := inode.xattrs[name]
	if !exists {
		logger.Debug().Msg("Xattr not found")
		logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), fuse.Status(syscall.ENODATA))
		return 0, fuse.Status(syscall.ENODATA)
	}

	logger.Debug().
		Int("valueLen", len(value)).
		Msg("Retrieved xattr")

	// If this is just a size query, return the size
	if len(buf) == 0 {
		logging.LogMethodExit(methodName, time.Since(startTime), uint32(len(value)), fuse.OK)
		return uint32(len(value)), fuse.OK
	}

	// If the buffer is too small, return ERANGE
	if len(buf) < len(value) {
		logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), fuse.Status(syscall.ERANGE))
		return 0, fuse.Status(syscall.ERANGE)
	}

	// Copy the value to the output buffer
	copy(buf, value)
	result := uint32(len(value))
	logging.LogMethodExit(methodName, time.Since(startTime), result, fuse.OK)
	return result, fuse.OK
}

// SetXAttr sets the value of an extended attribute.
func (f *Filesystem) SetXAttr(_ <-chan struct{}, in *fuse.SetXAttrIn, name string, value []byte) fuse.Status {
	methodName, startTime := logging.LogMethodEntry("SetXAttr", in.NodeId, name, len(value))

	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		logging.LogMethodExit(methodName, time.Since(startTime), fuse.ENOENT)
		return fuse.ENOENT
	}

	// Create a context for this operation with request ID and user ID
	ctx := logging.NewLogContextWithRequestAndUserID("xattr_operations").
		WithPath(inode.Path()).
		With("id", id).
		With("nodeID", in.NodeId).
		With("name", name).
		With("valueLen", len(value))

	// Get a logger with the context
	logger := ctx.Logger()

	inode.Lock()
	defer inode.Unlock()

	// Initialize the xattrs map if it's nil
	if inode.xattrs == nil {
		inode.xattrs = make(map[string][]byte)
	}

	// Store a copy of the value
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	inode.xattrs[name] = valueCopy
	logger.Debug().Msg("Set xattr")

	logging.LogMethodExit(methodName, time.Since(startTime), fuse.OK)
	return fuse.OK
}

// ListXAttr lists all extended attributes for a file.
func (f *Filesystem) ListXAttr(_ <-chan struct{}, header *fuse.InHeader, buf []byte) (uint32, fuse.Status) {
	methodName, startTime := logging.LogMethodEntry("ListXAttr", header.NodeId, len(buf))

	id := f.TranslateID(header.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), fuse.ENOENT)
		return 0, fuse.ENOENT
	}

	// Create a context for this operation with request ID and user ID
	ctx := logging.NewLogContextWithRequestAndUserID("xattr_operations").
		WithPath(inode.Path()).
		With("id", id).
		With("nodeID", header.NodeId)

	// Get a logger with the context
	logger := ctx.Logger()

	inode.RLock()
	defer inode.RUnlock()

	// Calculate total size needed for all attribute names
	var totalSize uint32
	for name := range inode.xattrs {
		// +1 for null terminator
		totalSize += uint32(len(name) + 1)
	}

	// If this is just a size query, return the size
	if len(buf) == 0 {
		logger.Debug().
			Uint32("size", totalSize).
			Msg("Returning xattr list size")
		logging.LogMethodExit(methodName, time.Since(startTime), totalSize, fuse.OK)
		return totalSize, fuse.OK
	}

	// If the buffer is too small, return ERANGE
	if len(buf) < int(totalSize) {
		logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), fuse.Status(syscall.ERANGE))
		return 0, fuse.Status(syscall.ERANGE)
	}

	// Build the list of attribute names
	var offset int
	for name := range inode.xattrs {
		nameBytes := []byte(name)
		// Ensure we don't exceed buffer bounds
		if offset+len(nameBytes)+1 <= len(buf) {
			copy(buf[offset:], nameBytes)
			offset += len(nameBytes)
			buf[offset] = 0 // null terminator
			offset++
		} else {
			// Buffer too small, shouldn't happen due to earlier check but added for safety
			logging.LogMethodExit(methodName, time.Since(startTime), uint32(0), fuse.Status(syscall.ERANGE))
			return 0, fuse.Status(syscall.ERANGE)
		}
	}

	logger.Debug().
		Int("count", len(inode.xattrs)).
		Msg("Listed xattrs")

	logging.LogMethodExit(methodName, time.Since(startTime), totalSize, fuse.OK)
	return totalSize, fuse.OK
}

// RemoveXAttr removes an extended attribute.
func (f *Filesystem) RemoveXAttr(_ <-chan struct{}, header *fuse.InHeader, name string) fuse.Status {
	methodName, startTime := logging.LogMethodEntry("RemoveXAttr", header.NodeId, name)

	id := f.TranslateID(header.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		logging.LogMethodExit(methodName, time.Since(startTime), fuse.ENOENT)
		return fuse.ENOENT
	}

	// Create a context for this operation with request ID and user ID
	ctx := logging.NewLogContextWithRequestAndUserID("xattr_operations").
		WithPath(inode.Path()).
		With("id", id).
		With("nodeID", header.NodeId).
		With("name", name)

	// Get a logger with the context
	logger := ctx.Logger()

	inode.Lock()
	defer inode.Unlock()

	if inode.xattrs == nil {
		logger.Debug().Msg("Xattr map is nil")
		logging.LogMethodExit(methodName, time.Since(startTime), fuse.Status(syscall.ENODATA))
		return fuse.Status(syscall.ENODATA)
	}

	if _, exists := inode.xattrs[name]; !exists {
		logger.Debug().Msg("Xattr not found")
		logging.LogMethodExit(methodName, time.Since(startTime), fuse.Status(syscall.ENODATA))
		return fuse.Status(syscall.ENODATA)
	}

	delete(inode.xattrs, name)
	logger.Debug().Msg("Removed xattr")

	logging.LogMethodExit(methodName, time.Since(startTime), fuse.OK)
	return fuse.OK
}
