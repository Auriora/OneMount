package fs

import (
	"bytes"
	"time"

	"github.com/auriora/onemount/internal/fs/graph"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// FileStatus represents the synchronization status of a file or folder
type FileStatus int

const (
	// StatusCloud indicates the file exists in the cloud but not in local cache
	StatusCloud FileStatus = iota

	// StatusLocal indicates the file exists in the local cache
	StatusLocal

	// StatusLocalModified indicates the file has been modified locally but not synced
	StatusLocalModified

	// StatusSyncing indicates the file is currently being synchronized (uploaded)
	StatusSyncing

	// StatusDownloading indicates the file is currently being downloaded
	StatusDownloading

	// StatusOutofSync indicates the file needs to be updated from OneDrive cloud
	StatusOutofSync

	// StatusError indicates there was an error synchronizing the file
	StatusError

	// StatusConflict indicates there is a conflict between local and remote versions
	StatusConflict
)

// FileStatusInfo contains detailed information about a file's status
type FileStatusInfo struct {
	Status    FileStatus
	ErrorMsg  string    // Only populated for StatusError
	Timestamp time.Time // When the status was last updated
}

// String returns a human-readable representation of the file status
func (s FileStatus) String() string {
	switch s {
	case StatusCloud:
		return "Cloud"
	case StatusLocal:
		return "Local"
	case StatusLocalModified:
		return "LocalModified"
	case StatusSyncing:
		return "Syncing"
	case StatusDownloading:
		return "Downloading"
	case StatusOutofSync:
		return "OutofSync"
	case StatusError:
		return "Error"
	case StatusConflict:
		return "Conflict"
	default:
		return "Unknown"
	}
}

// GetFileStatus determines the current status of a file
func (f *Filesystem) GetFileStatus(id string) FileStatusInfo {
	f.statusM.RLock()
	if status, exists := f.statuses[id]; exists {
		f.statusM.RUnlock()
		return status
	}
	f.statusM.RUnlock()

	// If no cached status, determine it now
	status := f.determineFileStatus(id)
	f.SetFileStatus(id, status)
	return status
}

// determineFileStatus calculates the current status of a file
func (f *Filesystem) determineFileStatus(id string) FileStatusInfo {
	// Check if file is being uploaded
	if f.uploads != nil {
		// Use the UploadManager's mutex to safely access the sessions map
		f.uploads.mutex.RLock()
		for _, session := range f.uploads.sessions {
			session.Lock()
			sessionID := session.ID
			sessionOldID := session.OldID
			session.Unlock()

			if sessionID == id || sessionOldID == id {
				state := session.getState()
				f.uploads.mutex.RUnlock()
				switch state {
				case uploadNotStarted:
					return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
				case uploadStarted:
					return FileStatusInfo{Status: StatusSyncing, Timestamp: time.Now()}
				case uploadComplete:
					return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
				case uploadErrored:
					session.Lock()
					var errorMsg string
					if session.error != nil {
						errorMsg = session.error.Error()
					} else {
						errorMsg = "Unknown error"
					}
					session.Unlock()
					return FileStatusInfo{
						Status:    StatusError,
						ErrorMsg:  errorMsg,
						Timestamp: time.Now(),
					}
				default:
					return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
				}
			}
		}
		f.uploads.mutex.RUnlock()
	}

	// Check if file has offline changes
	hasOfflineChanges := false
	if err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketOfflineChanges)
		if b == nil {
			return nil
		}

		// Check if there are any changes for this ID
		c := b.Cursor()
		prefix := []byte(id + "-")
		k, _ := c.Seek(prefix)
		if k != nil && bytes.HasPrefix(k, prefix) {
			hasOfflineChanges = true
		}
		return nil
	}); err != nil {
		log.Error().Err(err).Str("id", id).Msg("Error checking offline changes")
	}

	if hasOfflineChanges {
		return FileStatusInfo{Status: StatusLocalModified, Timestamp: time.Now()}
	}

	// Check if file is in local cache
	if f.content.HasContent(id) {
		// Get the inode to check if it's out of sync
		inode := f.GetID(id)
		if inode != nil && !isLocalID(id) {
			// Check if the file needs to be updated from cloud
			// This happens when the local hash doesn't match the remote hash
			fd, err := f.content.Open(id)
			if err == nil {
				localHash := graph.QuickXORHashStream(fd)
				if !inode.VerifyChecksum(localHash) {
					return FileStatusInfo{Status: StatusOutofSync, Timestamp: time.Now()}
				}
			}
		}
		return FileStatusInfo{Status: StatusLocal, Timestamp: time.Now()}
	}

	// Default: file is in cloud only
	return FileStatusInfo{Status: StatusCloud, Timestamp: time.Now()}
}

// SetFileStatus updates the status of a file
func (f *Filesystem) SetFileStatus(id string, status FileStatusInfo) {
	f.statusM.Lock()
	defer f.statusM.Unlock()
	f.statuses[id] = status
}

// MarkFileDownloading marks a file as currently downloading
func (f *Filesystem) MarkFileDownloading(id string) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusDownloading,
		Timestamp: time.Now(),
	})
}

// MarkFileOutofSync marks a file as needing update from cloud
func (f *Filesystem) MarkFileOutofSync(id string) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusOutofSync,
		Timestamp: time.Now(),
	})
}

// MarkFileError marks a file as having an error
func (f *Filesystem) MarkFileError(id string, err error) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusError,
		ErrorMsg:  err.Error(),
		Timestamp: time.Now(),
	})
}

// MarkFileConflict marks a file as having a conflict
func (f *Filesystem) MarkFileConflict(id string, message string) {
	f.SetFileStatus(id, FileStatusInfo{
		Status:    StatusConflict,
		ErrorMsg:  message,
		Timestamp: time.Now(),
	})
}

// InodePath returns the full path of an inode
func (f *Filesystem) InodePath(inode *Inode) string {
	if inode == nil {
		return ""
	}
	return inode.Path()
}

// updateFileStatus sets the extended attribute for file status and sends a D-Bus signal
func (f *Filesystem) updateFileStatus(inode *Inode) {
	path := f.InodePath(inode)
	if path == "" {
		return
	}

	// Get the ID before locking to avoid potential deadlocks
	id := inode.ID()

	// Store the path for D-Bus signal
	pathCopy := path
	var statusStrCopy string

	// Lock the inode before getting the status to prevent race conditions
	inode.Lock()

	// Get the status after locking the inode
	status := f.GetFileStatus(id)
	statusStr := status.Status.String()

	// Store the status string for D-Bus signal
	statusStrCopy = statusStr

	// Initialize the xattrs map if it's nil
	if inode.xattrs == nil {
		inode.xattrs = make(map[string][]byte)
	}

	// Set the status xattr
	inode.xattrs["user.onemount.status"] = []byte(statusStr)

	// If there's an error message, set it too
	if status.ErrorMsg != "" {
		inode.xattrs["user.onemount.error"] = []byte(status.ErrorMsg)
	} else {
		// Remove the error xattr if it exists
		delete(inode.xattrs, "user.onemount.error")
	}

	// Unlock the inode before sending D-Bus signal to avoid potential deadlocks
	inode.Unlock()

	// Send D-Bus signal if server is available
	if f.dbusServer != nil {
		f.dbusServer.SendFileStatusUpdate(pathCopy, statusStrCopy)
	}
}
