package fs

import (
	"github.com/auriora/onemount/internal/logging"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// isMountpointMounted checks if a filesystem is mounted at the given mountpoint
func isMountpointMounted(mountpoint string) bool {
	if mountpoint == "" {
		return false
	}

	// Check if it's a mount point using findmnt
	cmd := exec.Command("findmnt", "--noheadings", "--output", "TARGET", mountpoint)
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		return true
	}

	return false
}

// UnmountHandler should be used as goroutine that will handle sigint then exit gracefully
// It accepts a done channel that can be used to signal it to stop
func UnmountHandler(signal <-chan os.Signal, server *fuse.Server, filesystem *Filesystem, done <-chan struct{}) {
	select {
	case sig := <-signal: // block until signal
		logging.Info().Str("signal", strings.ToUpper(sig.String())).
			Msg("Signal received, cleaning up and unmounting filesystem.")

		// Stop all background processes
		if filesystem != nil {
			// Stop the cache cleanup routine
			filesystem.StopCacheCleanup()

			// Stop the delta loop
			filesystem.StopDeltaLoop()

			// Stop the download manager
			filesystem.StopDownloadManager()

			// Stop the upload manager
			filesystem.StopUploadManager()

			// Stop the metadata request manager
			filesystem.StopMetadataRequestManager()

			// Give the system a moment to release all resources
			logging.Info().Msg("Waiting for all resources to be released before unmounting...")
			time.Sleep(500 * time.Millisecond)
		}

		// Unmount the filesystem with retries
		maxRetries := 3
		retryDelay := 5000 * time.Millisecond
		var err error

		// Check if the server is nil, which would indicate it's not mounted
		if server == nil {
			logging.Warn().Msg("FUSE server is nil, skipping unmount operation")
		} else {
			for i := 0; i < maxRetries; i++ {
				err = server.Unmount()
				if err == nil {
					break
				}

				if i < maxRetries-1 {
					logging.Warn().Err(err).
						Int("retry", i+1).
						Dur("delay", retryDelay).
						Msg("Failed to unmount filesystem, retrying after delay...")
					time.Sleep(retryDelay)
					retryDelay *= 2 // Exponential backoff
				}
			}
		}

		if err != nil {
			logging.Error().Err(err).Msg("Failed to unmount filesystem cleanly after multiple attempts! " +
				"Run \"fusermount3 -uz /MOUNTPOINT/GOES/HERE\" to unmount.")
			os.Exit(1) // Exit with error code 1 to indicate failure
		} else {
			logging.Info().Msg("Filesystem unmounted successfully.")
			os.Exit(0) // Exit with success code 0
		}
	case <-done: // Exit if done channel is closed
		logging.Debug().Msg("UnmountHandler stopped")
		return
	}
}
