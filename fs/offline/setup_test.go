package offline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/fs"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	mountLoc  = "mount"
	testDBLoc = "tmp"
	TestDir   = mountLoc + "/onedriver_tests"
)

var auth *graph.Auth

// captureFileSystemState captures the current state of the filesystem
// by listing all files and directories in the mount location
func captureFileSystemState() (map[string]os.FileInfo, error) {
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

// TestMain is the entry point for all tests in this package.
// It sets up the test environment for offline testing, runs the tests, and cleans up afterward.
//
// The setup process:
// 1. Changes the working directory to the project root
// 2. Attempts to unmount any existing filesystem
// 3. Creates the mount directory if it doesn't exist
// 4. Authenticates with Microsoft Graph API
// 5. Sets up logging
// 6. Initializes the filesystem with cached data from previous tests
// 7. Mounts the filesystem with FUSE
// 8. Sets up signal handlers for graceful unmount
// 9. Waits for the filesystem to be mounted
// 10. Creates test files before entering offline mode
// 11. Sets the operational offline state to true to simulate offline mode
// 12. Sets the filesystem's offline mode to ReadWrite
// 13. Verifies that files are accessible in offline mode
// 14. Captures the initial state of the filesystem
//
// The teardown process:
// 1. Resets the operational offline state to false
// 2. Stops all filesystem services
// 3. Stops the UnmountHandler goroutine
// 4. Stops signal notifications
// 5. Unmounts the filesystem with retries
//
// This package is designed for running tests in offline mode to verify that the filesystem
// works correctly when network access is unavailable.
func TestMain(m *testing.M) {
	if wd, _ := os.Getwd(); strings.HasSuffix(wd, "/offline") {
		// depending on how this test gets launched, the working directory can be wrong
		if err := os.Chdir("../.."); err != nil {
			fmt.Println("Failed to change directory:", err)
			os.Exit(1)
		}
	}

	// attempt to unmount regardless of what happens (in case previous tests
	// failed and didn't clean themselves up)
	if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err != nil {
		fmt.Println("Warning: Failed to unmount:", err)
		// Continue anyway as it might not be mounted
	}
	if err := os.Mkdir(mountLoc, 0755); err != nil && !os.IsExist(err) {
		fmt.Println("Failed to create mount directory:", err)
		os.Exit(1)
	}

	var err error
	auth, err = graph.Authenticate(context.Background(), graph.AuthConfig{}, ".auth_tokens.json", false)
	if err != nil {
		fmt.Println("Authentication failed:", err)
		os.Exit(1)
	}

	f, err := os.OpenFile("fusefs_tests.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log file")
		}
	}()
	log.Info().Msg("Setup offline tests ------------------------------")

	// reuses the cached data from the previous tests
	filesystem, err := fs.NewFilesystem(auth, filepath.Join(testDBLoc, "test"), 30)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize filesystem")
		os.Exit(1)
	}

	server, err := fuse.NewServer(
		filesystem,
		mountLoc,
		&fuse.MountOptions{
			Name:          "onedriver",
			FsName:        "onedriver",
			DisableXAttrs: false,
			MaxBackground: 1024,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create FUSE server")
		os.Exit(1)
	}

	// setup sigint handler for graceful unmount on interrupt/terminate
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	unmountDone := make(chan struct{})
	go fs.UnmountHandler(sigChan, server, filesystem, unmountDone)

	// mount fs in background thread
	go server.Serve()

	// Wait for the filesystem to be mounted with improved error handling
	log.Info().Msg("Waiting for filesystem to be mounted...")
	mounted := false
	var lastError error

	// Define a context with timeout for mount operation
	mountCtx, mountCancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		// Attempt to clean up
		if unmountErr := exec.Command("fusermount3", "-uz", mountLoc).Run(); unmountErr != nil {
			log.Error().Err(unmountErr).Msg("Failed to unmount during cleanup after mount failure")
		}
		os.Exit(1)
	}

	log.Info().Msg("Filesystem mounted successfully")

	// Create the test directory and files before setting offline mode
	// This is necessary because file creation is not allowed in offline mode
	log.Info().Msg("Creating test files before entering offline mode")
	if err := os.MkdirAll(TestDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create test directory")
		os.Exit(1)
	}

	// Create the bagels file with the expected content
	bagelPath := filepath.Join(TestDir, "bagels")
	if err := os.WriteFile(bagelPath, []byte("bagels\n"), 0644); err != nil {
		log.Error().Err(err).Msg("Failed to create bagels file")
		os.Exit(1)
	}

	// Set operational offline state to true to simulate offline mode
	log.Info().Msg("Setting operational offline state to true")
	graph.SetOperationalOffline(true)

	// Also set the filesystem's offline mode
	log.Info().Msg("Setting filesystem offline mode to ReadWrite")
	filesystem.SetOfflineMode(fs.OfflineModeReadWrite)

	// Ensure the filesystem is fully initialized before running tests
	// This helps prevent race conditions when tests start immediately
	log.Info().Msg("Ensuring filesystem is fully initialized before running tests...")

	// Verify that the bagels file is accessible in offline mode
	readPath := filepath.Join(TestDir, "bagels")
	if _, err := os.ReadFile(readPath); err != nil {
		log.Error().Err(err).Msg("Failed to read bagels file in offline mode")
		os.Exit(1)
	}

	// Give the filesystem a moment to stabilize after initialization
	time.Sleep(500 * time.Millisecond)

	log.Info().Msg("Filesystem is fully initialized in offline mode, starting tests...")

	// Capture the initial state of the filesystem before running tests
	initialState, initialStateErr := captureFileSystemState()
	if initialStateErr != nil {
		log.Error().Err(initialStateErr).Msg("Failed to capture initial filesystem state")
	} else {
		log.Info().Int("files", len(initialState)).Msg("Captured initial filesystem state")
	}

	// Setup cleanup to run even if tests panic
	defer func() {
		log.Info().Msg("Running deferred cleanup...")

		// Capture the final state of the filesystem after tests
		if initialStateErr == nil {
			finalState, finalStateErr := captureFileSystemState()
			if finalStateErr != nil {
				log.Error().Err(finalStateErr).Msg("Failed to capture final filesystem state")
			} else {
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
		}
	}()

	log.Info().Msg("Start offline tests ------------------------------")
	code := m.Run()
	log.Info().Msg("Finish offline tests ------------------------------")

	// Reset operational offline state to false before exiting
	log.Info().Msg("Resetting operational offline state to false")
	graph.SetOperationalOffline(false)

	// Clean up the test database directory by stopping all services first
	// This is important to do before unmounting to ensure no active operations
	log.Info().Msg("Stopping all filesystem services...")
	filesystem.StopCacheCleanup()
	filesystem.StopDeltaLoop()
	filesystem.StopDownloadManager()
	filesystem.StopUploadManager()
	filesystem.SerializeAll()

	// Wait a moment to ensure all file handles are closed
	time.Sleep(500 * time.Millisecond)

	// Stop the UnmountHandler goroutine
	close(unmountDone)

	// Give the UnmountHandler a moment to exit
	time.Sleep(100 * time.Millisecond)

	// Stop signal notifications
	signal.Stop(sigChan)

	// Attempt to unmount with retries
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

	os.Exit(code)
}
