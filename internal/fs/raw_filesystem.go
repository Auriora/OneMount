package fs

import (
	"github.com/hanwen/go-fuse/v2/fuse"
)

// CustomRawFileSystem is a custom implementation of the fuse.RawFileSystem interface.
// Note: POLL opcode support has been removed as go-fuse intentionally disables it
// to prevent deadlocks. See: https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs
type CustomRawFileSystem struct {
	fuse.RawFileSystem
	fs FilesystemInterface
}

// NewCustomRawFileSystem creates a new CustomRawFileSystem that wraps the default
// RawFileSystem implementation.
func NewCustomRawFileSystem(fs FilesystemInterface) *CustomRawFileSystem {
	return &CustomRawFileSystem{
		RawFileSystem: fuse.NewDefaultRawFileSystem(),
		fs:            fs,
	}
}
