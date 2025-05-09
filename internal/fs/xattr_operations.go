package fs

import (
	"bytes"
	"syscall"

	"github.com/auriora/onemount/internal/common/errors"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/rs/zerolog/log"
)

// GetXAttr retrieves the value of an extended attribute.
func (f *Filesystem) GetXAttr(_ <-chan struct{}, header *fuse.InHeader, name string, buf []byte) (uint32, fuse.Status) {
	id := f.TranslateID(header.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return 0, fuse.ENOENT
	}

	ctx := log.With().
		Str("op", "GetXAttr").
		Uint64("nodeID", header.NodeId).
		Str("id", id).
		Str("path", inode.Path()).
		Str("name", name).
		Logger()

	inode.RLock()
	defer inode.RUnlock()

	value, exists := inode.xattrs[name]
	if !exists {
		ctx.Debug().Msg("Xattr not found")
		return 0, fuse.Status(syscall.ENODATA)
	}

	ctx.Debug().
		Int("valueLen", len(value)).
		Msg("Retrieved xattr")

	// If this is just a size query, return the size
	if len(buf) == 0 {
		return uint32(len(value)), fuse.OK
	}

	// If the buffer is too small, return ERANGE
	if len(buf) < len(value) {
		return 0, fuse.Status(syscall.ERANGE)
	}

	// Copy the value to the output buffer
	copy(buf, value)
	return uint32(len(value)), fuse.OK
}

// SetXAttr sets the value of an extended attribute.
func (f *Filesystem) SetXAttr(_ <-chan struct{}, in *fuse.SetXAttrIn, name string, value []byte) fuse.Status {
	id := f.TranslateID(in.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return fuse.ENOENT
	}

	ctx := log.With().
		Str("op", "SetXAttr").
		Uint64("nodeID", in.NodeId).
		Str("id", id).
		Str("path", inode.Path()).
		Str("name", name).
		Int("valueLen", len(value)).
		Logger()

	inode.Lock()
	defer inode.Unlock()

	// Initialize the xattrs map if it's nil
	if inode.xattrs == nil {
		inode.xattrs = make(map[string][]byte)
	}

	// Store a copy of the value
	inode.xattrs[name] = bytes.Clone(value)
	ctx.Debug().Msg("Set xattr")

	return fuse.OK
}

// ListXAttr lists all extended attributes for a file.
func (f *Filesystem) ListXAttr(_ <-chan struct{}, header *fuse.InHeader, buf []byte) (uint32, fuse.Status) {
	id := f.TranslateID(header.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return 0, fuse.ENOENT
	}

	ctx := log.With().
		Str("op", "ListXAttr").
		Uint64("nodeID", header.NodeId).
		Str("id", id).
		Str("path", inode.Path()).
		Logger()

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
		ctx.Debug().
			Uint32("size", totalSize).
			Msg("Returning xattr list size")
		return totalSize, fuse.OK
	}

	// If the buffer is too small, return ERANGE
	if len(buf) < int(totalSize) {
		return 0, fuse.Status(syscall.ERANGE)
	}

	// Build the list of attribute names
	var offset int
	for name := range inode.xattrs {
		nameBytes := []byte(name)
		copy(buf[offset:], nameBytes)
		offset += len(nameBytes)
		buf[offset] = 0 // null terminator
		offset++
	}

	ctx.Debug().
		Int("count", len(inode.xattrs)).
		Msg("Listed xattrs")

	return totalSize, fuse.OK
}

// RemoveXAttr removes an extended attribute.
func (f *Filesystem) RemoveXAttr(_ <-chan struct{}, header *fuse.InHeader, name string) fuse.Status {
	id := f.TranslateID(header.NodeId)
	inode := f.GetID(id)
	if inode == nil {
		return fuse.ENOENT
	}

	ctx := log.With().
		Str("op", "RemoveXAttr").
		Uint64("nodeID", header.NodeId).
		Str("id", id).
		Str("path", inode.Path()).
		Str("name", name).
		Logger()

	inode.Lock()
	defer inode.Unlock()

	if inode.xattrs == nil || inode.xattrs[name] == nil {
		ctx.Debug().Msg("Xattr not found")
		return fuse.Status(syscall.ENODATA)
	}

	delete(inode.xattrs, name)
	ctx.Debug().Msg("Removed xattr")

	return fuse.OK
}
