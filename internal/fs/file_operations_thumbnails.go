package fs

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Release handles file release operations for the FUSE filesystem.
// This method is called when a file is closed by the kernel. For thumbnail files,
// it releases the file handle and cleans up any temporary files.
//
// Parameters:
//   - cancel: Channel that signals if the operation should be canceled (required by FUSE interface but not used)
//   - in: Input parameters for the release operation, including node ID and file handle
func (f *Filesystem) Release(_ <-chan struct{}, in *fuse.ReleaseIn) {
	// Check if this is a thumbnail file handle
	if in.Fh != 0 {
		// Get the file handle
		fh := f.GetFileHandle(in.Fh)
		if fh != nil {
			// This is a thumbnail file handle, use its Release method
			// Create a background context for the thumbnail release
			ctx := context.Background()
			_ = fh.Release(ctx) // Ignore the return value

			// Release the file handle
			f.ReleaseFileHandle(in.Fh)
		}
	}

	// For regular files, we don't need to do anything special
	// The content cache handles closing files automatically
}

// CleanupThumbnails cleans up the thumbnail cache by removing thumbnails
// that haven't been accessed in a while.
//
// Returns:
//   - The number of thumbnails removed
//   - An error if the cleanup failed
func (f *Filesystem) CleanupThumbnails() (int, error) {
	return f.thumbnails.CleanupCache(f.cacheExpirationDays)
}
