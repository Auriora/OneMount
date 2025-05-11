// Package fs provides the filesystem implementation for onemount.
package fs

// The download_manager.go file implements a background download manager for OneDrive files.
// This helps decouple the local file system logic from the OneDrive sync logic by running
// file downloads in separate worker threads. This improves performance by handling the
// OneDrive cloud file sync in the background, except when waiting for a file or folder to download.

import (
	"github.com/auriora/onemount/pkg/logging"
	"io"
	"os"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/graph"
)

const (
	// downloadQueued indicates the download is queued but not started
	downloadQueued DownloadState = iota
	// downloadStarted indicates the download is in progress
	downloadStarted
	// downloadCompleted indicates the download completed successfully
	downloadCompleted
	// downloadErrored indicates the download failed
	downloadErrored
)

// DownloadSession represents a file download session
type DownloadSession struct {
	ID        string
	Path      string
	State     DownloadState
	Error     error
	StartTime time.Time
	EndTime   time.Time
	mutex     sync.RWMutex
}

// DownloadManager handles background file downloads
type DownloadManager struct {
	fs         *Filesystem
	auth       *graph.Auth
	sessions   map[string]*DownloadSession
	queue      chan string
	mutex      sync.RWMutex
	workerWg   sync.WaitGroup
	numWorkers int
	stopChan   chan struct{}
}

// NewDownloadManager creates a new download manager
func NewDownloadManager(fs *Filesystem, auth *graph.Auth, numWorkers int) *DownloadManager {
	dm := &DownloadManager{
		fs:         fs,
		auth:       auth,
		sessions:   make(map[string]*DownloadSession),
		queue:      make(chan string, 500), // Buffer for 500 download requests
		numWorkers: numWorkers,
		stopChan:   make(chan struct{}),
	}

	// Start worker goroutines
	dm.startWorkers()

	return dm
}

// startWorkers starts the download worker goroutines
func (dm *DownloadManager) startWorkers() {
	for i := 0; i < dm.numWorkers; i++ {
		dm.workerWg.Add(1)
		go dm.worker()
	}
}

// worker processes download requests from the queue
func (dm *DownloadManager) worker() {
	defer dm.workerWg.Done()

	for {
		select {
		case id := <-dm.queue:
			dm.processDownload(id)
		case <-dm.stopChan:
			return
		}
	}
}

// processDownload handles the actual download of a file
func (dm *DownloadManager) processDownload(id string) {
	// Get the session
	dm.mutex.RLock()
	session, exists := dm.sessions[id]
	dm.mutex.RUnlock()

	if !exists {
		logging.LogError(errors.New("download session not found"), "Failed to process download",
			logging.FieldOperation, "processDownload",
			logging.FieldID, id)
		return
	}

	// Update session state
	session.mutex.Lock()
	session.State = downloadStarted
	session.StartTime = time.Now()
	session.mutex.Unlock()

	// Get the inode
	inode := dm.fs.GetID(id)
	if inode == nil {
		err := errors.NewNotFoundError("inode not found", nil)
		dm.setSessionError(session, err)
		return
	}

	// Update file status
	dm.fs.SetFileStatus(id, FileStatusInfo{
		Status:    StatusDownloading,
		Timestamp: time.Now(),
	})

	// Get file content
	// Access content field directly
	fd, err := dm.fs.content.Open(id)
	if err != nil {
		dm.setSessionError(session, err)
		return
	}

	// Create a temporary file for download
	tempID := "temp-" + id
	temp, err := dm.fs.content.Open(tempID)
	if err != nil {
		dm.setSessionError(session, err)
		return
	}
	defer func() {
		if err := dm.fs.content.Delete(tempID); err != nil {
			logging.LogError(err, "Failed to delete temporary file",
				logging.FieldOperation, "processDownload.cleanup",
				"tempID", tempID)
		}
	}()

	// Download the file content
	size, err := graph.GetItemContentStream(id, dm.auth, temp)
	if err != nil {
		dm.setSessionError(session, err)
		return
	}

	// Verify checksum
	if !inode.VerifyChecksum(graph.QuickXORHashStream(temp)) {
		err := errors.NewValidationError("checksum verification failed", nil)
		dm.setSessionError(session, err)
		return
	}

	// Reset file positions
	if _, err := temp.Seek(0, 0); err != nil {
		dm.setSessionError(session, err)
		return
	}

	if _, err := fd.Seek(0, 0); err != nil {
		dm.setSessionError(session, err)
		return
	}

	if err := fd.Truncate(0); err != nil {
		dm.setSessionError(session, err)
		return
	}

	// Copy content from temp file to destination
	if _, err := copyBuffer(fd, temp); err != nil {
		dm.setSessionError(session, err)
		return
	}

	// Ensure data is flushed to disk
	if err := fd.Sync(); err != nil {
		dm.setSessionError(session, err)
		return
	}

	// Update inode size
	inode.Lock()
	inode.DriveItem.Size = size
	inode.Unlock()

	// Update file status
	dm.fs.SetFileStatus(id, FileStatusInfo{
		Status:    StatusLocal,
		Timestamp: time.Now(),
	})

	// Update session state
	session.mutex.Lock()
	session.State = downloadCompleted
	session.EndTime = time.Now()
	session.mutex.Unlock()

	logging.Info().
		Str("id", id).
		Str("path", session.Path).
		Msg("File download completed")
}

// setSessionError updates a session with an error
func (dm *DownloadManager) setSessionError(session *DownloadSession, err error) {
	session.mutex.Lock()
	session.State = downloadErrored
	session.Error = err
	session.EndTime = time.Now()
	session.mutex.Unlock()

	// Update file status
	dm.fs.MarkFileError(session.ID, err)

	logging.LogError(err, "File download failed",
		logging.FieldOperation, "setSessionError",
		logging.FieldID, session.ID,
		logging.FieldPath, session.Path)
}

// QueueDownload adds a file to the download queue
func (dm *DownloadManager) QueueDownload(id string) (*DownloadSession, error) {
	// Check if the file is already being downloaded
	dm.mutex.RLock()
	session, exists := dm.sessions[id]
	dm.mutex.RUnlock()

	if exists {
		// Return the existing session
		return session, nil
	}

	// Get the inode to get the path
	inode := dm.fs.GetID(id)
	if inode == nil {
		return nil, errors.NewNotFoundError("inode not found", nil)
	}

	path := inode.Path()

	// Create a new session
	session = &DownloadSession{
		ID:    id,
		Path:  path,
		State: downloadQueued,
	}

	// Add to sessions map
	dm.mutex.Lock()
	dm.sessions[id] = session
	dm.mutex.Unlock()

	// Add to download queue
	select {
	case dm.queue <- id:
		logging.Info().
			Str("id", id).
			Str("path", path).
			Msg("File queued for download")
	default:
		// Queue is full, return error
		dm.mutex.Lock()
		delete(dm.sessions, id)
		dm.mutex.Unlock()
		return nil, errors.NewResourceBusyError("download queue is full", nil)
	}

	return session, nil
}

// GetDownloadStatus returns the status of a download
func (dm *DownloadManager) GetDownloadStatus(id string) (DownloadState, error) {
	dm.mutex.RLock()
	session, exists := dm.sessions[id]
	dm.mutex.RUnlock()

	if !exists {
		return 0, errors.NewNotFoundError("download session not found", nil)
	}

	session.mutex.RLock()
	state := session.State
	session.mutex.RUnlock()

	return state, nil
}

// WaitForDownload waits for a download to complete
func (dm *DownloadManager) WaitForDownload(id string) error {
	for {
		dm.mutex.RLock()
		session, exists := dm.sessions[id]
		dm.mutex.RUnlock()

		if !exists {
			return errors.NewNotFoundError("download session not found", nil)
		}

		session.mutex.RLock()
		state := session.State
		err := session.Error
		session.mutex.RUnlock()

		switch state {
		case downloadCompleted:
			return nil
		case downloadErrored:
			return err
		default:
			// Still in progress, wait a bit
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop stops the download manager and waits for all workers to finish
func (dm *DownloadManager) Stop() {
	logging.Info().Msg("Stopping download manager...")
	close(dm.stopChan)

	// Wait for all workers to finish with a timeout
	done := make(chan struct{})
	go func() {
		dm.workerWg.Wait()
		close(done)
	}()

	// Wait for workers to finish or timeout after 5 seconds
	select {
	case <-done:
		logging.Info().Msg("Download manager stopped successfully")
	case <-time.After(5 * time.Second):
		logging.Warn().Msg("Timed out waiting for download manager to stop")
	}
}

// copyBuffer copies from src to dst using a buffer
func copyBuffer(dst, src *os.File) (int64, error) {
	buf := make([]byte, 32*1024)
	var written int64

	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			nw, err := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if err == nil {
					err = errors.NewOperationError("invalid write result", nil)
				}
			}
			written += int64(nw)
			if err != nil {
				return written, err
			}
			if nr != nw {
				return written, errors.NewOperationError("short write", nil)
			}
		}
		if err != nil {
			if err == io.EOF {
				return written, nil
			}
			return written, err
		}
	}
}
