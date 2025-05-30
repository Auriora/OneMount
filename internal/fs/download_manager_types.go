package fs

// DownloadState represents the state of a download
type DownloadState int

const (
	// DownloadNotStarted indicates the download is queued but not started
	DownloadNotStartedState DownloadState = iota
	// DownloadStarted indicates the download is in progress
	DownloadStartedState
	// DownloadCompleted indicates the download completed successfully
	DownloadCompletedState
	// DownloadErrored indicates the download failed
	DownloadErroredState
)

// DownloadSessionInterface defines the interface for a download session
type DownloadSessionInterface interface {
	// GetID returns the ID of the item being downloaded
	GetID() string

	// GetPath returns the path of the item being downloaded
	GetPath() string

	// GetState returns the current state of the download
	GetState() DownloadState

	// GetError returns any error that occurred during download
	GetError() error

	// IsComplete returns true if the download has completed successfully
	IsComplete() bool
}

// DownloadManagerInterface defines the interface for the download manager
// that is used by other packages. This interface is implemented by the
// DownloadManager type in the fs package.
type DownloadManagerInterface interface {
	// Queue a download
	QueueDownload(id string) (DownloadSessionInterface, error)

	// Get the status of a download
	GetDownloadStatus(id string) (DownloadState, error)

	// Wait for a download to complete
	WaitForDownload(id string) error

	// Stop the download manager
	Stop()
}
