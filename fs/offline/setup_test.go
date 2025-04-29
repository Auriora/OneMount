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

	"github.com/bcherrington/onedriver/fs"
	"github.com/bcherrington/onedriver/fs/graph"
	"github.com/bcherrington/onedriver/testutil"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Use constants from testutil package
var (
	mountLoc  = testutil.TestMountPoint
	testDBLoc = testutil.TestDBLoc
	TestDir   = testutil.TestDir
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
	// First check if the mount point exists before attempting to unmount
	if _, err := os.Stat(mountLoc); err == nil {
		if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err != nil {
			fmt.Println("Warning: Failed to unmount:", err)
			// Continue anyway as it might not be mounted
		}
	} else {
		fmt.Println("Mount point does not exist, no need to unmount")
	}

	// Remove the mount directory if it exists, then recreate it
	// This ensures we start with a clean state
	if _, err := os.Stat(mountLoc); err == nil {
		if err := os.RemoveAll(mountLoc); err != nil {
			fmt.Println("Warning: Failed to remove existing mount directory:", err)
		}
	}

	// Create the mount directory
	if err := os.MkdirAll(mountLoc, 0755); err != nil {
		fmt.Println("Failed to create mount directory:", err)
		os.Exit(1)
	}

	var err error
	// Check if we should use mock authentication
	if os.Getenv("ONEDRIVER_MOCK_AUTH") == "1" {
		// Use mock authentication
		mockClient := graph.NewMockGraphClient()
		auth = &mockClient.Auth
		log.Info().Msg("Using mock authentication for tests")
	} else {
		// Use real authentication
		auth, err = graph.Authenticate(context.Background(), graph.AuthConfig{}, ".auth_tokens.json", false)
		if err != nil {
			fmt.Println("Authentication failed:", err)
			os.Exit(1)
		}
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

	var server *fuse.Server
	var sigChan chan os.Signal
	var unmountDone chan struct{}

	// Check if we should skip FUSE mounting
	if os.Getenv("ONEDRIVER_MOCK_AUTH") == "1" {
		// Skip FUSE mounting when using mock authentication
		log.Info().Msg("Skipping FUSE mounting for tests with mock authentication")

		// Create the mount directory structure manually
		if err := os.MkdirAll(mountLoc, 0755); err != nil && !os.IsExist(err) {
			log.Error().Err(err).Msg("Failed to create mount directory")
			os.Exit(1)
		}

		// Create test directories
		if err := os.MkdirAll(TestDir, 0755); err != nil && !os.IsExist(err) {
			log.Error().Err(err).Msg("Failed to create test directory")
			os.Exit(1)
		}

		// Create directories for specific tests
		testDirs := []string{
			filepath.Join(TestDir, "donuts_TestOfflineFileSystemOperations"),
			filepath.Join(TestDir, "modify_TestOfflineFileSystemOperations"),
			filepath.Join(TestDir, "delete_TestOfflineFileSystemOperations"),
			filepath.Join(TestDir, "dir_create_TestOfflineFileSystemOperations"),
			filepath.Join(TestDir, "dir_delete_TestOfflineFileSystemOperations"),
			filepath.Join(TestDir, "parent_dir_TestOfflineFileSystemOperations"),
		}

		for _, dir := range testDirs {
			if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
				log.Error().Err(err).Str("dir", dir).Msg("Failed to create test subdirectory")
				os.Exit(1)
			}
		}

		// Create the bagels file with the expected content
		bagelPath := filepath.Join(TestDir, "bagels")
		if err := os.WriteFile(bagelPath, []byte("bagels\n"), 0644); err != nil {
			log.Error().Err(err).Msg("Failed to create bagels file")
			os.Exit(1)
		}
	} else {
		server, err = fuse.NewServer(
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
		sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
		unmountDone = make(chan struct{})
		go fs.UnmountHandler(sigChan, server, filesystem, unmountDone)

		// mount fs in background thread
		go server.Serve()

		// Wait for the filesystem to be mounted with improved error handling
		log.Info().Msg("Waiting for filesystem to be mounted...")

		// Use WaitForCondition to wait for the filesystem to be mounted
		// This replaces the complex goroutine with a simpler, more reliable approach

		// Define a timeout for the mount operation
		timeout := 30 * time.Second
		pollInterval := 100 * time.Millisecond

		// Create a function that checks if the mount point is ready
		isMountReady := func() bool {
			// Check if mount point exists
			if _, err := os.Stat(mountLoc); err != nil {
				log.Debug().Err(err).Msg("tmp/mount point not accessible yet")
				return false
			}

			// Try to create a test file to verify the filesystem is working
			testFile := filepath.Join(mountLoc, ".test-mount-ready")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				log.Debug().Err(err).Msg("tmp/mount point exists but test file creation failed")
				return false
			}

			// Successfully created test file, filesystem is mounted
			if removeErr := os.Remove(testFile); removeErr != nil {
				log.Warn().Err(removeErr).Msg("Failed to remove test file, but mount is confirmed")
			}

			return true
		}

		// Use WaitForCondition to wait for the mount point to be ready
		// If the condition is not met within the timeout, it will fail the test
		testutil.WaitForCondition(nil, isMountReady, timeout, pollInterval, "Filesystem failed to mount within timeout")

		log.Info().Msg("Filesystem mounted successfully")
	}

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

	// Create directories for specific tests
	testDirs := []string{
		filepath.Join(TestDir, "donuts_TestOfflineFileSystemOperations"),
		filepath.Join(TestDir, "modify_TestOfflineFileSystemOperations"),
		filepath.Join(TestDir, "delete_TestOfflineFileSystemOperations"),
		filepath.Join(TestDir, "dir_create_TestOfflineFileSystemOperations"),
		filepath.Join(TestDir, "dir_delete_TestOfflineFileSystemOperations"),
		filepath.Join(TestDir, "parent_dir_TestOfflineFileSystemOperations"),
	}

	for _, dir := range testDirs {
		if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
			log.Error().Err(err).Str("dir", dir).Msg("Failed to create test subdirectory")
			os.Exit(1)
		}
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

	// Use WaitForCondition to wait for the bagels file to be accessible
	// This replaces the fixed sleep with a more reliable approach
	log.Info().Msg("Waiting for filesystem to stabilize in offline mode...")
	testutil.WaitForCondition(nil, func() bool {
		_, err := os.ReadFile(readPath)
		return err == nil
	}, 5*time.Second, 100*time.Millisecond, "Bagels file not accessible within timeout")

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

	// Create a channel to receive the test result
	resultChan := make(chan int)

	// Create a context with timeout for the tests
	testCtx, testCancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer testCancel()

	// Run the tests in a goroutine
	go func() {
		code := m.Run()
		select {
		case <-testCtx.Done():
			// Context was canceled, don't try to send on resultChan
			return
		default:
			resultChan <- code
		}
	}()

	// Declare the code variable before the select statement
	var code int

	// Wait for the tests to complete or timeout
	// Using 8 minutes to ensure we have time for cleanup before the default Go test timeout of 10 minutes
	select {
	case code = <-resultChan:
		log.Info().Int("code", code).Msg("Finish offline tests ------------------------------")
	case <-testCtx.Done():
		log.Error().Msg("Tests timed out after 8 minutes, forcing cleanup and exit")
		// Return a non-zero exit code to indicate failure
		code = 1
	}

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

	// Use a context with timeout to ensure we don't wait indefinitely
	log.Info().Msg("Waiting for file handles to be closed...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a ticker for periodic checks
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Wait for a short period to allow file handles to close
	// This is a best-effort approach as we can't directly check if all handles are closed
	select {
	case <-ctx.Done():
		log.Warn().Msg("Timeout waiting for file handles to close")
	case <-time.After(500 * time.Millisecond):
		// Wait for a longer period to ensure file handles are closed
		log.Info().Msg("Waited for file handles to close")
	}

	// Stop the UnmountHandler goroutine if it exists
	if unmountDone != nil {
		close(unmountDone)
	}

	// Wait for the UnmountHandler to exit
	// This is a best-effort approach as we can't directly check if the handler has exited
	select {
	case <-ctx.Done():
		log.Warn().Msg("Timeout waiting for UnmountHandler to exit")
	case <-time.After(500 * time.Millisecond):
		// Wait for a longer period to ensure the handler has exited
		log.Info().Msg("Waited for UnmountHandler to exit")
	}

	// Stop signal notifications if they exist
	if sigChan != nil {
		signal.Stop(sigChan)
	}

	unmountSuccess := false

	// Check if we're using mock authentication
	if os.Getenv("ONEDRIVER_MOCK_AUTH") == "1" {
		// Skip unmounting when using mock authentication
		log.Info().Msg("Skipping FUSE unmounting for tests with mock authentication")
		unmountSuccess = true
	} else {
		// Attempt to unmount with retries
		log.Info().Msg("Attempting to unmount filesystem...")

		// Check if the mount point exists before attempting to unmount
		if _, err := os.Stat(mountLoc); err != nil {
			log.Warn().Err(err).Msg("Mount point does not exist, no need to unmount")
			unmountSuccess = true
		} else {
			// First try normal unmount
			unmountErr := server.Unmount()
			if unmountErr == nil {
				unmountSuccess = true
				log.Info().Msg("Successfully unmounted filesystem")
			} else {
				log.Error().Err(unmountErr).Msg("Failed to unmount test fuse server, attempting lazy unmount")

				// Try lazy unmount with retries using exponential backoff
				err := testutil.RetryWithBackoff(nil, func() error {
					return exec.Command("fusermount3", "-uz", mountLoc).Run()
				}, 3, 500*time.Millisecond, 2*time.Second, "Lazy unmount")

				if err == nil {
					unmountSuccess = true
					log.Info().Msg("Successfully performed lazy unmount")
				} else {
					log.Error().Err(err).Msg("Failed to perform lazy unmount after retries")
				}
			}
		}
	}

	if unmountSuccess {
		fmt.Println("Successfully unmounted fuse server!")
	} else {
		fmt.Println("Warning: Failed to unmount fuse server. Continuing with exit anyway to prevent hanging.")
		// Make one final attempt with the most aggressive unmount option
		if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err != nil {
			log.Error().Err(err).Msg("Final attempt at lazy unmount failed")
		} else {
			log.Info().Msg("Final lazy unmount succeeded")
		}
	}

	os.Exit(code)
}
