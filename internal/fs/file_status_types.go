package fs

import (
	"time"
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
	ErrorCode string    // Error code for more specific error handling
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
