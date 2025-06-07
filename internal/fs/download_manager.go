// Package fs provides the filesystem implementation for onemount.
package fs

// The download_manager.go file implements a background download manager for OneDrive files.
// This helps decouple the local file system logic from the OneDrive sync logic by running
// file downloads in separate worker threads. This improves performance by handling the
// OneDrive cloud file sync in the background, except when waiting for a file or folder to download.

import (
	"context"
	"encoding/json"
	"github.com/auriora/onemount/pkg/logging"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/retry"
	bolt "go.etcd.io/bbolt"
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

const (
	// Default chunk size for downloads (1MB)
	downloadChunkSize uint64 = 1024 * 1024
)

var bucketDownloads = []byte("downloads")

// DownloadSession represents a file download session with recovery capabilities
type DownloadSession struct {
	ID        string
	Path      string
	State     DownloadState
	Error     error
	StartTime time.Time
	EndTime   time.Time

	// Recovery and progress tracking fields
	Size                uint64    `json:"size"`
	BytesDownloaded     uint64    `json:"bytesDownloaded"`
	LastSuccessfulChunk int       `json:"lastSuccessfulChunk"`
	TotalChunks         int       `json:"totalChunks"`
	ChunkSize           uint64    `json:"chunkSize"`
	LastProgressTime    time.Time `json:"lastProgressTime"`
	RecoveryAttempts    int       `json:"recoveryAttempts"`
	CanResume           bool      `json:"canResume"`
	DownloadURL         string    `json:"downloadUrl"`
	ETag                string    `json:"eTag"`

	mutex sync.RWMutex
}

// updateProgress updates the download progress
func (ds *DownloadSession) updateProgress(chunkIndex int, bytesDownloaded uint64) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	ds.LastSuccessfulChunk = chunkIndex
	ds.BytesDownloaded = bytesDownloaded
	ds.LastProgressTime = time.Now()
}

// canResumeDownload checks if the download can be resumed from the last checkpoint
func (ds *DownloadSession) canResumeDownload() bool {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.CanResume && ds.LastSuccessfulChunk >= 0 && ds.DownloadURL != ""
}

// getResumeOffset returns the byte offset from which to resume the download
func (ds *DownloadSession) getResumeOffset() uint64 {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	if ds.LastSuccessfulChunk < 0 {
		return 0
	}
	return uint64(ds.LastSuccessfulChunk+1) * ds.ChunkSize
}

// markAsResumable marks the session as resumable and calculates total chunks
func (ds *DownloadSession) markAsResumable(size uint64, chunkSize uint64) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	ds.CanResume = true
	ds.Size = size
	ds.ChunkSize = chunkSize
	ds.TotalChunks = int(math.Ceil(float64(size) / float64(chunkSize)))
	ds.LastSuccessfulChunk = -1 // No chunks downloaded yet
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
	db         *bolt.DB
}

// NewDownloadManager creates a new download manager
func NewDownloadManager(fs *Filesystem, auth *graph.Auth, numWorkers int, db *bolt.DB) *DownloadManager {
	dm := &DownloadManager{
		fs:         fs,
		auth:       auth,
		sessions:   make(map[string]*DownloadSession),
		queue:      make(chan string, 500), // Buffer for 500 download requests
		numWorkers: numWorkers,
		stopChan:   make(chan struct{}),
		db:         db,
	}

	// Restore any incomplete download sessions from disk
	dm.restoreDownloadSessions()

	// Start worker goroutines
	dm.startWorkers()

	return dm
}

// restoreDownloadSessions restores incomplete download sessions from the database
func (dm *DownloadManager) restoreDownloadSessions() {
	if dm.db == nil {
		return
	}

	dm.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDownloads)
		if b == nil {
			return nil
		}
		return b.ForEach(func(key []byte, val []byte) error {
			session := &DownloadSession{}
			err := json.Unmarshal(val, session)
			if err != nil {
				logging.Error().Err(err).Msg("Failed to restore download session from disk")
				return err
			}

			// Reset state to queued for recovery
			session.State = downloadQueued
			session.RecoveryAttempts++

			dm.mutex.Lock()
			dm.sessions[session.ID] = session
			dm.mutex.Unlock()

			logging.Info().
				Str("id", session.ID).
				Str("path", session.Path).
				Int("recoveryAttempts", session.RecoveryAttempts).
				Msg("Restored download session for recovery")

			return nil
		})
	})
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

	// Create a context for the download operation
	ctx := context.Background()

	// Create a retry config for the download operation
	retryConfig := retry.DefaultConfig()

	// Download the file content with retry
	var size uint64
	err = retry.Do(ctx, func() error {
		// Reset the file position before each attempt
		if _, err := temp.Seek(0, 0); err != nil {
			return errors.Wrap(err, "failed to reset file position")
		}

		// Truncate the file before each attempt
		if err := temp.Truncate(0); err != nil {
			return errors.Wrap(err, "failed to truncate temporary file")
		}

		// Download the file content
		var downloadErr error
		size, downloadErr = graph.GetItemContentStream(id, dm.auth, temp)
		if downloadErr != nil {
			return errors.Wrap(downloadErr, "failed to download file content")
		}

		// Verify checksum
		if !inode.DriveItem.VerifyChecksum(graph.QuickXORHashStream(temp)) {
			return errors.NewValidationError("checksum verification failed", nil)
		}

		return nil
	}, retryConfig)

	if err != nil {
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

	// Copy content from temp file to destination with retry
	err = retry.Do(ctx, func() error {
		// Reset file positions before each attempt
		if _, err := temp.Seek(0, 0); err != nil {
			return errors.Wrap(err, "failed to reset temp file position")
		}

		if _, err := fd.Seek(0, 0); err != nil {
			return errors.Wrap(err, "failed to reset destination file position")
		}

		if err := fd.Truncate(0); err != nil {
			return errors.Wrap(err, "failed to truncate destination file")
		}

		// Copy content
		_, copyErr := copyBuffer(fd, temp)
		if copyErr != nil {
			return errors.Wrap(copyErr, "failed to copy file content")
		}

		// Ensure data is flushed to disk
		if syncErr := fd.Sync(); syncErr != nil {
			return errors.Wrap(syncErr, "failed to sync file to disk")
		}

		return nil
	}, retryConfig)

	if err != nil {
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

	// Update session state and progress tracking
	session.mutex.Lock()
	session.State = downloadCompleted
	session.EndTime = time.Now()
	// Update progress tracking to reflect completed download
	session.BytesDownloaded = size
	if session.CanResume && session.TotalChunks > 0 {
		session.LastSuccessfulChunk = session.TotalChunks - 1 // All chunks completed
	}
	session.mutex.Unlock()

	logging.Info().
		Str("id", id).
		Str("path", session.Path).
		Msg("File download completed")

	// Note: Session cleanup is deferred to allow status checking after completion
	// The session will be cleaned up when a new download is queued or during shutdown
}

// setSessionError updates a session with an error
func (dm *DownloadManager) setSessionError(session *DownloadSession, err error) {
	session.mutex.Lock()
	session.State = downloadErrored
	session.Error = err
	session.EndTime = time.Now()
	session.RecoveryAttempts++
	session.mutex.Unlock()

	// Update file status
	dm.fs.MarkFileError(session.ID, err)

	// Persist updated session state for potential recovery
	if dm.db != nil && session.RecoveryAttempts <= 3 {
		contents, _ := json.Marshal(session)
		dm.db.Batch(func(tx *bolt.Tx) error {
			b, _ := tx.CreateBucketIfNotExists(bucketDownloads)
			return b.Put([]byte(session.ID), contents)
		})
	}

	logging.LogError(err, "File download failed",
		logging.FieldOperation, "setSessionError",
		logging.FieldID, session.ID,
		logging.FieldPath, session.Path,
		"recoveryAttempts", session.RecoveryAttempts)
}

// finishDownloadSession removes a completed download session from memory and database
func (dm *DownloadManager) finishDownloadSession(id string) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Remove from memory
	delete(dm.sessions, id)

	// Remove from database
	if dm.db != nil {
		dm.db.Batch(func(tx *bolt.Tx) error {
			if b := tx.Bucket(bucketDownloads); b != nil {
				b.Delete([]byte(id))
			}
			return nil
		})
	}
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

	// Create a new session with recovery capabilities
	session = &DownloadSession{
		ID:                  id,
		Path:                path,
		State:               downloadQueued,
		LastSuccessfulChunk: -1,
		BytesDownloaded:     0,
		RecoveryAttempts:    0,
		CanResume:           false,
	}

	// Initialize session for large files that support resumable downloads
	if inode.DriveItem.Size > downloadChunkSize {
		session.markAsResumable(inode.DriveItem.Size, downloadChunkSize)
	}

	// Persist session to database for recovery
	if dm.db != nil {
		contents, _ := json.Marshal(session)
		dm.db.Batch(func(tx *bolt.Tx) error {
			b, _ := tx.CreateBucketIfNotExists(bucketDownloads)
			return b.Put([]byte(session.ID), contents)
		})
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

	// Clean up completed sessions after returning their status
	if state == downloadCompleted {
		// Use a goroutine to avoid blocking the caller
		go dm.finishDownloadSession(id)
	}

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

// GetID returns the ID of the item being downloaded
func (ds *DownloadSession) GetID() string {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.ID
}

// GetPath returns the path of the item being downloaded
func (ds *DownloadSession) GetPath() string {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.Path
}

// GetState returns the current state of the download
func (ds *DownloadSession) GetState() DownloadState {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.State
}

// GetError returns any error that occurred during download
func (ds *DownloadSession) GetError() error {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.Error
}

// IsComplete returns true if the download has completed successfully
func (ds *DownloadSession) IsComplete() bool {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.State == downloadCompleted
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
