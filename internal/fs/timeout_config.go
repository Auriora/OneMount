package fs

import (
	"time"
)

// TimeoutConfig holds all timeout values used across the filesystem components.
// This provides a centralized location for timeout configuration and ensures
// consistency across all managers and background processes.
//
// Timeout Policy:
// - Short operations (< 5s): Used for quick checks and lightweight operations
// - Medium operations (5-30s): Used for network requests and file operations
// - Long operations (30s-2m): Used for large file uploads/downloads
// - Graceful shutdown (30s-60s): Used for clean shutdown of background processes
//
// All timeouts are configurable via command-line flags or configuration file.
type TimeoutConfig struct {
	// Download manager timeouts
	DownloadWorkerShutdown time.Duration // Time to wait for download workers to finish

	// Upload manager timeouts
	UploadGracefulShutdown time.Duration // Time to wait for active uploads to complete

	// Filesystem shutdown timeouts
	FilesystemShutdown time.Duration // Time to wait for all filesystem goroutines to stop

	// Network feedback timeouts
	NetworkCallbackShutdown time.Duration // Time to wait for network feedback callbacks

	// Metadata request timeouts
	MetadataRequestTimeout time.Duration // Time to wait for metadata requests

	// Content cache statistics timeout
	ContentStatsTimeout time.Duration // Time to wait for content cache statistics
}

// DefaultTimeoutConfig returns the default timeout configuration.
// These values are chosen to balance responsiveness with reliability:
// - Short timeouts (5s) for operations that should complete quickly
// - Medium timeouts (10-30s) for network operations
// - Long timeouts (30-60s) for graceful shutdown of complex operations
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		// Download manager: 5 seconds for workers to finish
		DownloadWorkerShutdown: 5 * time.Second,

		// Upload manager: 30 seconds for large uploads to complete
		UploadGracefulShutdown: 30 * time.Second,

		// Filesystem: 10 seconds for all goroutines to stop
		FilesystemShutdown: 10 * time.Second,

		// Network feedback: 5 seconds for callbacks to complete
		NetworkCallbackShutdown: 5 * time.Second,

		// Metadata requests: 30 seconds for metadata fetch
		MetadataRequestTimeout: 30 * time.Second,

		// Content stats: 5 seconds for statistics collection
		ContentStatsTimeout: 5 * time.Second,
	}
}

// Validate checks that all timeout values are reasonable.
// Returns an error if any timeout is invalid.
func (tc *TimeoutConfig) Validate() error {
	// All timeouts must be positive
	if tc.DownloadWorkerShutdown <= 0 {
		return &InvalidConfigError{Field: "DownloadWorkerShutdown", Reason: "must be positive"}
	}
	if tc.UploadGracefulShutdown <= 0 {
		return &InvalidConfigError{Field: "UploadGracefulShutdown", Reason: "must be positive"}
	}
	if tc.FilesystemShutdown <= 0 {
		return &InvalidConfigError{Field: "FilesystemShutdown", Reason: "must be positive"}
	}
	if tc.NetworkCallbackShutdown <= 0 {
		return &InvalidConfigError{Field: "NetworkCallbackShutdown", Reason: "must be positive"}
	}
	if tc.MetadataRequestTimeout <= 0 {
		return &InvalidConfigError{Field: "MetadataRequestTimeout", Reason: "must be positive"}
	}
	if tc.ContentStatsTimeout <= 0 {
		return &InvalidConfigError{Field: "ContentStatsTimeout", Reason: "must be positive"}
	}

	// Warn if timeouts are unreasonably short (< 1 second)
	if tc.DownloadWorkerShutdown < time.Second {
		return &InvalidConfigError{Field: "DownloadWorkerShutdown", Reason: "should be at least 1 second"}
	}
	if tc.NetworkCallbackShutdown < time.Second {
		return &InvalidConfigError{Field: "NetworkCallbackShutdown", Reason: "should be at least 1 second"}
	}
	if tc.ContentStatsTimeout < time.Second {
		return &InvalidConfigError{Field: "ContentStatsTimeout", Reason: "should be at least 1 second"}
	}

	// Warn if timeouts are unreasonably long (> 5 minutes)
	if tc.DownloadWorkerShutdown > 5*time.Minute {
		return &InvalidConfigError{Field: "DownloadWorkerShutdown", Reason: "should not exceed 5 minutes"}
	}
	if tc.UploadGracefulShutdown > 5*time.Minute {
		return &InvalidConfigError{Field: "UploadGracefulShutdown", Reason: "should not exceed 5 minutes"}
	}
	if tc.FilesystemShutdown > 5*time.Minute {
		return &InvalidConfigError{Field: "FilesystemShutdown", Reason: "should not exceed 5 minutes"}
	}
	if tc.NetworkCallbackShutdown > 5*time.Minute {
		return &InvalidConfigError{Field: "NetworkCallbackShutdown", Reason: "should not exceed 5 minutes"}
	}
	if tc.MetadataRequestTimeout > 5*time.Minute {
		return &InvalidConfigError{Field: "MetadataRequestTimeout", Reason: "should not exceed 5 minutes"}
	}
	if tc.ContentStatsTimeout > 5*time.Minute {
		return &InvalidConfigError{Field: "ContentStatsTimeout", Reason: "should not exceed 5 minutes"}
	}

	return nil
}

// InvalidConfigError represents an invalid configuration error
type InvalidConfigError struct {
	Field  string
	Reason string
}

func (e *InvalidConfigError) Error() string {
	return "invalid timeout configuration for " + e.Field + ": " + e.Reason
}
