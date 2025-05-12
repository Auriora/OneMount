package fs

import (
	"time"

	"github.com/auriora/onemount/pkg/logging"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// CustomRawFileSystem is a custom implementation of the fuse.RawFileSystem interface
// that adds support for the POLL opcode.
type CustomRawFileSystem struct {
	fuse.RawFileSystem
	fs FilesystemInterface
}

// NewCustomRawFileSystem creates a new CustomRawFileSystem that wraps the default
// RawFileSystem implementation and adds support for the POLL opcode.
func NewCustomRawFileSystem(fs FilesystemInterface) *CustomRawFileSystem {
	return &CustomRawFileSystem{
		RawFileSystem: fuse.NewDefaultRawFileSystem(),
		fs:            fs,
	}
}

// Implement the POLL opcode handler
func (c *CustomRawFileSystem) Poll(cancel <-chan struct{}, in *fuse.InHeader, out *fuse.OutHeader) fuse.Status {
	methodName, startTime := logging.LogMethodEntry("Poll", in.NodeId)

	// Call the Poll method on the filesystem
	result := c.fs.Poll(cancel, in, out)

	defer logging.LogMethodExit(methodName, time.Since(startTime), result)
	return result
}
