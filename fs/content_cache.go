package fs

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// LoopbackCache stores the content for files under a folder as regular files
type LoopbackCache struct {
	directory string
	fds       sync.Map
}

func NewLoopbackCache(directory string) *LoopbackCache {
	if err := os.MkdirAll(directory, 0700); err != nil {
		// Log error but continue - the directory might already exist
		// or we might be able to create files directly
		// This is a best-effort approach
		// Using MkdirAll instead of Mkdir to create parent directories if needed
	}
	return &LoopbackCache{
		directory: directory,
		fds:       sync.Map{},
	}
}

// contentPath returns the path for the given content file
func (l *LoopbackCache) contentPath(id string) string {
	return filepath.Join(l.directory, id)
}

// Get reads a file's content from disk.
func (l *LoopbackCache) Get(id string) []byte {
	content, err := os.ReadFile(l.contentPath(id))
	if err != nil {
		// Return empty content if file doesn't exist or can't be read
		// This matches the previous behavior but is more explicit
		return []byte{}
	}
	return content
}

// InsertContent writes file content to disk in a single bulk insert.
func (l *LoopbackCache) Insert(id string, content []byte) error {
	return os.WriteFile(l.contentPath(id), content, 0600)
}

// InsertStream inserts a stream of data
func (l *LoopbackCache) InsertStream(id string, reader io.Reader) (int64, error) {
	fd, err := l.Open(id)
	if err != nil {
		return 0, err
	}

	// Copy the data from the reader to the file
	// We don't reset position or truncate here to maintain compatibility with existing code
	n, err := io.Copy(fd, reader)
	if err != nil {
		return n, err
	}

	return n, nil
}

// Delete closes the fd AND deletes content from disk.
func (l *LoopbackCache) Delete(id string) error {
	// Try to close the file first
	closeErr := l.Close(id)

	// Try to remove the file regardless of close error
	removeErr := os.Remove(l.contentPath(id))

	// Handle remove error - ignore "file not found" errors
	if removeErr != nil && !os.IsNotExist(removeErr) {
		return removeErr
	}

	// If we got here, the remove succeeded or the file didn't exist
	// Return any close error that might have occurred
	return closeErr
}

// Move moves content from one ID to another
func (l *LoopbackCache) Move(oldID string, newID string) error {
	// Close both files to ensure they're not open during the move
	// Capture errors but continue with the move operation
	oldCloseErr := l.Close(oldID)
	newCloseErr := l.Close(newID)

	// Make sure the destination directory exists
	destDir := filepath.Dir(l.contentPath(newID))
	if err := os.MkdirAll(destDir, 0700); err != nil {
		return err
	}

	// Check if source file exists
	if _, err := os.Stat(l.contentPath(oldID)); os.IsNotExist(err) {
		return err
	}

	// Remove destination file if it exists to avoid "file exists" errors
	// Ignore any errors from this operation
	_ = os.Remove(l.contentPath(newID))

	// Perform the rename operation
	renameErr := os.Rename(l.contentPath(oldID), l.contentPath(newID))
	if renameErr != nil {
		return renameErr
	}

	// If we got here, the rename succeeded
	// Return any close errors that might have occurred
	if oldCloseErr != nil {
		return oldCloseErr
	}
	return newCloseErr
}

// IsOpen returns true if the file is already opened somewhere
func (l *LoopbackCache) IsOpen(id string) bool {
	_, ok := l.fds.Load(id)
	return ok
}

// HasContent is used to find if we have a file or not in cache (in any state)
func (l *LoopbackCache) HasContent(id string) bool {
	// is it already open?
	_, ok := l.fds.Load(id)
	if ok {
		return ok
	}
	// is it on disk?
	_, err := os.Stat(l.contentPath(id))
	return err == nil
}

// Open returns a filehandle for subsequent access
func (l *LoopbackCache) Open(id string) (*os.File, error) {
	if fd, ok := l.fds.Load(id); ok {
		// already opened, return existing fd
		return fd.(*os.File), nil
	}

	fd, err := os.OpenFile(l.contentPath(id), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Since we explicitly want to store *os.Files, we need to prevent the Go
	// GC from trying to be "helpful" and closing files for us behind the
	// scenes.
	// https://github.com/hanwen/go-fuse/issues/371#issuecomment-694799535
	runtime.SetFinalizer(fd, nil)
	l.fds.Store(id, fd)
	return fd, nil
}

// Close closes the currently open fd
func (l *LoopbackCache) Close(id string) error {
	if fd, ok := l.fds.Load(id); ok {
		file := fd.(*os.File)

		// Try to sync the file, but don't fail if it doesn't work
		// We still want to try to close the file even if sync fails
		syncErr := file.Sync()

		// Close the file and capture any error
		closeErr := file.Close()

		// Remove from the map regardless of errors
		l.fds.Delete(id)

		// Return the first error encountered
		if syncErr != nil {
			return syncErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
