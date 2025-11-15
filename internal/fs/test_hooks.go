package fs

import "github.com/hanwen/go-fuse/v2/fuse"

// FilesystemTestHooks contains optional overrides that test code can install
// to force specific results from high-level filesystem operations. Production
// code never sets these hooks; they are only used in unit/integration tests to
// simulate error scenarios that would be difficult to reproduce deterministically.
type FilesystemTestHooks struct {
	// OpenHook, when set, can short-circuit Filesystem.Open. Returning handled=false
	// allows the normal implementation to run.
	OpenHook func(fs *Filesystem, in *fuse.OpenIn, out *fuse.OpenOut) (status fuse.Status, handled bool)

	// WriteHook can intercept Filesystem.Write. The bool return indicates whether
	// the hook handled the write.
	WriteHook func(fs *Filesystem, in *fuse.WriteIn, data []byte) (bytes uint32, status fuse.Status, handled bool)

	// CreateHook can intercept Filesystem.Create. Returning handled=false lets the
	// default behavior proceed.
	CreateHook func(fs *Filesystem, in *fuse.CreateIn, name string, out *fuse.CreateOut) (status fuse.Status, handled bool)
}

// SetTestHooks installs the provided hooks for the filesystem instance.
func (f *Filesystem) SetTestHooks(hooks *FilesystemTestHooks) {
	f.testHooks = hooks
}

// ClearTestHooks removes any previously installed test hooks.
func (f *Filesystem) ClearTestHooks() {
	f.testHooks = nil
}
