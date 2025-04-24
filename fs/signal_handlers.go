package fs

import (
	"os"
	"strings"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/rs/zerolog/log"
)

// UnmountHandler should be used as goroutine that will handle sigint then exit gracefully
func UnmountHandler(signal <-chan os.Signal, server *fuse.Server) {
	sig := <-signal // block until signal
	log.Info().Str("signal", strings.ToUpper(sig.String())).
		Msg("Signal received, unmounting filesystem.")

	// Unmount the filesystem with retries
	maxRetries := 3
	retryDelay := 500 * time.Millisecond
	var err error

	for i := 0; i < maxRetries; i++ {
		err = server.Unmount()
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			log.Warn().Err(err).
				Int("retry", i+1).
				Dur("delay", retryDelay).
				Msg("Failed to unmount filesystem, retrying after delay...")
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to unmount filesystem cleanly after multiple attempts! " +
			"Run \"fusermount3 -uz /MOUNTPOINT/GOES/HERE\" to unmount.")
	}

	os.Exit(128)
}
