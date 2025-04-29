package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bcherrington/onedriver/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/rs/zerolog/log"
)

// CaptureFileSystemState captures the current state of the filesystem
// by listing all files and directories in the mount location
func CaptureFileSystemState(mountLoc string) (map[string]os.FileInfo, error) {
	state := make(map[string]os.FileInfo)

	err := filepath.Walk(mountLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the mount point itself
		if path == mountLoc {
			return nil
		}
		// Store the file info in the state map
		state[path] = info
		return nil
	})

	return state, err
}

// CheckAndUnmountMountPoint checks if the mount point is already in use by another process
// and attempts to unmount it if necessary.
func CheckAndUnmountMountPoint(mountLoc string) bool {
	isMounted := false
	if _, err := os.Stat(mountLoc); err == nil {
		// Check if it's a mount point by trying to read from it
		if _, err := os.ReadDir(mountLoc); err != nil {
			// If we can't read the directory, it might be a stale mount point
			log.Warn().Err(err).Msg("Mount point exists but can't be read, attempting to unmount")
			isMounted = true
		} else {
			// Check if it's a mount point using findmnt
			cmd := exec.Command("findmnt", "--noheadings", "--output", "TARGET", mountLoc)
			if output, err := cmd.Output(); err == nil && len(output) > 0 {
				log.Warn().Msg("Mount point is already mounted, attempting to unmount")
				isMounted = true
			}
		}
	}

	if isMounted {
		log.Info().Msg("Attempting to unmount previous instance")
		// Try normal unmount first
		if unmountErr := exec.Command("fusermount3", "-u", mountLoc).Run(); unmountErr != nil {
			log.Warn().Err(unmountErr).Msg("Normal unmount failed, trying lazy unmount")
			// Try lazy unmount
			if lazyErr := exec.Command("fusermount3", "-uz", mountLoc).Run(); lazyErr != nil {
				log.Error().Err(lazyErr).Msg("Lazy unmount also failed, mount point may be in use by another process")
				// Continue anyway, but warn the user
				fmt.Println("WARNING: Failed to unmount existing filesystem. Tests may fail if mount point is in use.")
			} else {
				log.Info().Msg("Successfully performed lazy unmount")
			}
		} else {
			log.Info().Msg("Successfully unmounted previous instance")
		}
	}

	return isMounted
}

// WaitForMount waits for the filesystem to be mounted and ready.
// It returns true if the mount was successful, false otherwise.
func WaitForMount(mountLoc string, timeout time.Duration) (bool, error) {
	log.Info().Msg("Waiting for filesystem to be mounted...")
	mounted := false
	var lastError error

	// Define a context with timeout for mount operation
	mountCtx, mountCancel := context.WithTimeout(context.Background(), timeout)
	defer mountCancel()

	// Create a ticker for periodic checks
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Use a channel to signal when mounting is complete
	mountDone := make(chan bool, 1)

	// Start a goroutine to check mount status
	go func() {
		for {
			select {
			case <-mountCtx.Done():
				// Context timeout or cancellation
				return
			case <-ticker.C:
				// Check if mount point exists
				if _, err := os.Stat(mountLoc); err == nil {
					// Try to create a test file to verify the filesystem is working
					testFile := filepath.Join(mountLoc, ".test-mount-ready")
					if err := os.WriteFile(testFile, []byte("test"), 0644); err == nil {
						// Successfully created test file, filesystem is mounted
						if removeErr := os.Remove(testFile); removeErr != nil {
							log.Warn().Err(removeErr).Msg("Failed to remove test file, but mount is confirmed")
						}
						mountDone <- true
						return
					} else {
						lastError = err
						log.Debug().Err(err).Msg("Mount point exists but test file creation failed")
					}
				} else {
					lastError = err
					log.Debug().Err(err).Msg("Mount point not accessible yet")
				}
			}
		}
	}()

	// Wait for mounting to complete or timeout
	select {
	case <-mountDone:
		mounted = true
	case <-mountCtx.Done():
		// Timeout or cancellation
		log.Error().Err(mountCtx.Err()).Msg("Mount operation timed out")
	}

	if !mounted {
		log.Error().Err(lastError).Msg("Filesystem failed to mount within timeout")
		return false, lastError
	}

	log.Info().Msg("Filesystem mounted successfully")
	return true, nil
}

// UnmountWithRetries attempts to unmount the filesystem with retries.
// It returns true if the unmount was successful, false otherwise.
func UnmountWithRetries(server *fuse.Server, mountLoc string) bool {
	log.Info().Msg("Attempting to unmount filesystem...")
	unmountSuccess := false

	// First try normal unmount
	unmountErr := server.Unmount()
	if unmountErr == nil {
		unmountSuccess = true
		log.Info().Msg("Successfully unmounted filesystem")
	} else {
		log.Error().Err(unmountErr).Msg("Failed to unmount test fuse server, attempting lazy unmount")

		// Try lazy unmount with retries
		for i := 0; i < 3; i++ {
			if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err == nil {
				unmountSuccess = true
				log.Info().Msg("Successfully performed lazy unmount")
				break
			} else {
				log.Error().Err(err).Int("attempt", i+1).Msg("Failed to perform lazy unmount")
				time.Sleep(500 * time.Millisecond) // Wait before retrying
			}
		}
	}

	if unmountSuccess {
		fmt.Println("Successfully unmounted fuse server!")
	} else {
		fmt.Println("Warning: Failed to unmount fuse server. You may need to manually unmount with 'fusermount3 -uz mount'")
	}

	return unmountSuccess
}

// CleanupFilesystemState compares the initial and final state of the filesystem
// and attempts to clean up any files that were created during the tests.
func CleanupFilesystemState(initialState map[string]os.FileInfo) {
	log.Info().Msg("Running filesystem state cleanup...")

	// Capture the final state of the filesystem
	finalState, finalStateErr := CaptureFileSystemState("tmp/mount")
	if finalStateErr != nil {
		log.Error().Err(finalStateErr).Msg("Failed to capture final filesystem state")
		return
	}
	log.Info().Int("files", len(finalState)).Msg("Captured final filesystem state")

	// Check for files that exist in the final state but not in the initial state
	for path, info := range finalState {
		if _, exists := initialState[path]; !exists {
			log.Warn().Str("path", path).Bool("isDir", info.IsDir()).Msg("File created during tests but not cleaned up")

			// Attempt to clean up the file/directory
			if info.IsDir() {
				// Only remove empty directories to avoid accidentally deleting important content
				if entries, err := os.ReadDir(path); err == nil && len(entries) == 0 {
					if err := os.Remove(path); err != nil {
						log.Error().Err(err).Str("path", path).Msg("Failed to clean up directory")
					} else {
						log.Info().Str("path", path).Msg("Successfully cleaned up directory")
					}
				}
			} else {
				// Remove files
				if err := os.Remove(path); err != nil {
					log.Error().Err(err).Str("path", path).Msg("Failed to clean up file")
				} else {
					log.Info().Str("path", path).Msg("Successfully cleaned up file")
				}
			}
		}
	}
}

// StopFilesystemServices stops all filesystem services in preparation for unmounting.
func StopFilesystemServices(filesystem *fs.Filesystem) {
	log.Info().Msg("Stopping all filesystem services...")
	filesystem.StopCacheCleanup()
	filesystem.StopDeltaLoop()
	filesystem.StopDownloadManager()
	filesystem.StopUploadManager()
	filesystem.SerializeAll()

	// Wait a moment to ensure all file handles are closed
	time.Sleep(100 * time.Millisecond)
}
