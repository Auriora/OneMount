package fs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	mountLoc     = "mount"
	testDBLoc    = "tmp"
	TestDir      = mountLoc + "/onedriver_tests"
	DeltaDir     = TestDir + "/delta"
	retrySeconds = 60 * time.Second //lint:ignore ST1011 a
)

var (
	auth *graph.Auth
	fs   *Filesystem
)

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
// It sets up the test environment, runs the tests, and cleans up afterward.
//
// The setup process:
// 1. Sets the ONEDRIVER_TEST environment variable to indicate we're in a test environment
// 2. Changes the working directory to the project root
// 3. Checks if the mount point is already in use and unmounts it if necessary
// 4. Creates the mount directory if it doesn't exist
// 5. Wipes all cached data from previous tests
// 6. Sets up logging
// 7. Authenticates with Microsoft Graph API
// 8. Initializes the filesystem
// 9. Mounts the filesystem with FUSE
// 10. Sets up signal handlers for graceful unmount
// 11. Creates test directories and files
// 12. Captures the initial state of the filesystem
//
// The teardown process:
// 1. Waits for any remaining uploads to complete
// 2. Stops the UnmountHandler goroutine
// 3. Stops signal notifications
// 4. Unmounts the filesystem with retries
// 5. Stops all filesystem services
// 6. Removes the test database directory
//
// Tests are done in the main project directory with a mounted filesystem to
// avoid having to repeatedly recreate auth_tokens.json and juggle multiple auth
// sessions.
func TestMain(m *testing.M) {
	// Set environment variable to indicate we're in a test environment
	if err := os.Setenv("ONEDRIVER_TEST", "1"); err != nil {
		fmt.Println("Failed to set ONEDRIVER_TEST environment variable:", err)
		os.Exit(1)
	}
	// We used to skip paging test setup for single tests, but that caused issues
	// when running TestListChildrenPaging individually

	// Check if we're already in the project root directory
	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		fmt.Println("Failed to get current working directory:", cwdErr)
		os.Exit(1)
	}

	if strings.HasSuffix(cwd, "/fs") {
		// If we're in the fs directory, change to the project root
		if cdErr := os.Chdir(".."); cdErr != nil {
			fmt.Println("Failed to change to project root directory:", cdErr)
			os.Exit(1)
		}
	} else if !strings.HasSuffix(cwd, "/onedriver") {
		// If we're not in the project root, try to find it
		// This handles the case where tests are run from GoLand with a different working directory
		if strings.Contains(cwd, "/onedriver") {
			// Extract the path up to and including "onedriver"
			index := strings.Index(cwd, "/onedriver")
			projectRoot := cwd[:index+len("/onedriver")]
			if cdErr := os.Chdir(projectRoot); cdErr != nil {
				fmt.Println("Failed to change to project root directory:", cdErr)
				os.Exit(1)
			}
		}
	}

	// Check if the mount point is already in use by another process
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

	// Attempt to unmount if necessary
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

	// Create mount directory if it doesn't exist
	if mkdirErr := os.Mkdir(mountLoc, 0755); mkdirErr != nil && !os.IsExist(mkdirErr) {
		fmt.Println("Failed to create mount directory:", mkdirErr)
		os.Exit(1)
	}
	// wipe all cached data from previous tests
	if rmErr := os.RemoveAll(testDBLoc); rmErr != nil {
		fmt.Println("Failed to remove test database location:", rmErr)
		os.Exit(1)
	}
	if mkdirErr := os.Mkdir(testDBLoc, 0755); mkdirErr != nil && !os.IsExist(mkdirErr) {
		fmt.Println("Failed to create test database directory:", mkdirErr)
		os.Exit(1)
	}

	f, openErr := os.OpenFile("fusefs_tests.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if openErr != nil {
		fmt.Println("Failed to open log file:", openErr)
		os.Exit(1)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05"})
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close log file")
		}
	}()

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
	var fsErr error
	fs, fsErr = NewFilesystem(auth, filepath.Join(testDBLoc, "test"), 30)
	if fsErr != nil {
		log.Error().Err(fsErr).Msg("Failed to initialize filesystem")
		os.Exit(1)
	}

	var server *fuse.Server
	var sigChan chan os.Signal
	var unmountDone chan struct{}
	var mounted bool = false

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
		if err := os.MkdirAll(DeltaDir, 0755); err != nil && !os.IsExist(err) {
			log.Error().Err(err).Msg("Failed to create delta directory")
			os.Exit(1)
		}
		if err := os.MkdirAll(filepath.Join(TestDir, "paging"), 0755); err != nil && !os.IsExist(err) {
			log.Error().Err(err).Msg("Failed to create paging directory")
			os.Exit(1)
		}
		if err := os.MkdirAll(filepath.Join(mountLoc, "Documents"), 0755); err != nil && !os.IsExist(err) {
			log.Error().Err(err).Msg("Failed to create Documents directory")
		}

		mounted = true
	} else {
		// Create mount options
		mountOptions := &fuse.MountOptions{
			Name:          "onedriver",
			FsName:        "onedriver",
			DisableXAttrs: false,
			MaxBackground: 1024,
		}

		// Create the FUSE server
		var err error
		server, err = fuse.NewServer(
			fs,
			mountLoc,
			mountOptions,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create FUSE server")
			os.Exit(1)
		}

		// setup sigint handler for graceful unmount on interrupt/terminate
		sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
		unmountDone = make(chan struct{})
		go UnmountHandler(sigChan, server, fs, unmountDone)

		// mount fs in background thread
		go server.Serve()

		// Wait for the filesystem to be mounted with improved error handling
		log.Info().Msg("Waiting for filesystem to be mounted...")
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
	}

	// cleanup from last run
	log.Info().Msg("Setup test environment ---------------------------------")
	if err := os.RemoveAll(TestDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if mkdirErr := os.Mkdir(TestDir, 0755); mkdirErr != nil && !os.IsExist(mkdirErr) {
		log.Error().Err(mkdirErr).Msg("Failed to create test directory")
		os.Exit(1)
	}
	if mkdirErr := os.Mkdir(DeltaDir, 0755); mkdirErr != nil && !os.IsExist(mkdirErr) {
		log.Error().Err(mkdirErr).Msg("Failed to create delta directory")
		os.Exit(1)
	}

	// create paging test files before the delta thread is created
	if mkdirErr := os.Mkdir(filepath.Join(TestDir, "paging"), 0755); mkdirErr != nil && !os.IsExist(mkdirErr) {
		log.Error().Err(mkdirErr).Msg("Failed to create paging directory")
		os.Exit(1)
	}
	createPagingTestFiles()
	go fs.DeltaLoop(5 * time.Second)

	// not created by default on onedrive for business
	if mkdirErr := os.Mkdir(mountLoc+"/Documents", 0755); mkdirErr != nil && !os.IsExist(mkdirErr) {
		log.Error().Err(mkdirErr).Msg("Failed to create Documents directory")
		// Not exiting here as this is not critical
	}

	// we do not cd into the mounted directory or it will hang indefinitely on
	// unmount with "device or resource busy"
	log.Info().Msg("Test session start ---------------------------------")

	// Ensure the filesystem is fully initialized before running tests
	// This helps prevent race conditions when tests start immediately
	log.Info().Msg("Ensuring filesystem is fully initialized before running tests...")

	// Create a readiness file to verify the filesystem is fully operational
	readinessFile := filepath.Join(TestDir, ".readiness-check")
	if err := os.WriteFile(readinessFile, []byte("readiness check"), 0644); err != nil {
		log.Error().Err(err).Msg("Failed to create readiness check file")
		os.Exit(1)
	}

	// Read the file back to ensure it's accessible
	if _, err := os.ReadFile(readinessFile); err != nil {
		log.Error().Err(err).Msg("Failed to read readiness check file")
		os.Exit(1)
	}

	// Clean up the readiness file
	if err := os.Remove(readinessFile); err != nil {
		log.Warn().Err(err).Msg("Failed to remove readiness check file")
		// Not fatal, continue with tests
	}

	// Give the filesystem a moment to stabilize after initialization
	time.Sleep(500 * time.Millisecond)

	log.Info().Msg("Filesystem is fully initialized, starting tests...")

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

	// Register a cleanup function that will run even if tests panic
	cleanupDone := make(chan struct{})
	cleanupFunc := func() {
		// Avoid running cleanup multiple times
		select {
		case <-cleanupDone:
			return // Already cleaned up
		default:
			defer close(cleanupDone)
		}

		log.Info().Msg("Running emergency cleanup handler...")

		// Check if we're using mock authentication
		if os.Getenv("ONEDRIVER_MOCK_AUTH") == "1" {
			// Skip unmounting when using mock authentication
			log.Info().Msg("Skipping FUSE unmounting for tests with mock authentication")
			return
		}

		// Stop the UnmountHandler goroutine if it exists
		if unmountDone != nil {
			select {
			case <-unmountDone: // Already closed
			default:
				close(unmountDone)
			}
		}

		// Give the UnmountHandler a moment to exit
		time.Sleep(100 * time.Millisecond)

		// Stop signal notifications if sigChan exists
		if sigChan != nil {
			signal.Stop(sigChan)
		}

		// Attempt to unmount with retries
		log.Info().Msg("Emergency cleanup: Attempting to unmount filesystem...")

		// First try normal unmount if server exists
		if server != nil {
			unmountErr := server.Unmount()
			if unmountErr == nil {
				log.Info().Msg("Emergency cleanup: Successfully unmounted filesystem")
				return
			}
			log.Error().Err(unmountErr).Msg("Emergency cleanup: Failed to unmount test fuse server, attempting lazy unmount")
		}

		// Try lazy unmount with retries
		for i := 0; i < 5; i++ { // Increased retry count for emergency cleanup
			if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err == nil {
				log.Info().Msg("Emergency cleanup: Successfully performed lazy unmount")
				return
			} else {
				log.Error().Err(err).Int("attempt", i+1).Msg("Emergency cleanup: Failed to perform lazy unmount")
				time.Sleep(1 * time.Second) // Longer wait time for emergency cleanup
			}
		}

		log.Error().Msg("Emergency cleanup: All unmount attempts failed. Mount point may still be active.")

		// Even if unmount failed, try to clean up filesystem resources
		if fs != nil {
			log.Info().Msg("Emergency cleanup: Stopping filesystem services...")
			fs.StopCacheCleanup()
			fs.StopDeltaLoop()
			fs.StopDownloadManager()
			fs.StopUploadManager()
			fs.SerializeAll()

			// Wait a moment to ensure all file handles are closed
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Register the cleanup function with a goroutine that will be triggered on process exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("Received termination signal, running emergency cleanup...")
		cleanupFunc()
		os.Exit(1)
	}()

	// Ensure cleanup runs even if tests panic or fail
	defer cleanupFunc()

	// run tests
	code := m.Run()

	// Normal cleanup path
	log.Info().Msg("Test session end -----------------------------------")
	fmt.Printf("Waiting 5 seconds for any remaining uploads to complete")
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		fmt.Printf(".")
	}
	fmt.Printf("\n")

	unmountSuccess := false

	// Check if we're using mock authentication
	if os.Getenv("ONEDRIVER_MOCK_AUTH") == "1" {
		// Skip unmounting when using mock authentication
		log.Info().Msg("Skipping FUSE unmounting for tests with mock authentication")
		unmountSuccess = true
	} else {
		// Stop the UnmountHandler goroutine
		close(unmountDone)

		// Give the UnmountHandler a moment to exit
		time.Sleep(100 * time.Millisecond)

		// Stop signal notifications
		signal.Stop(sigChan)

		// Attempt to unmount with retries
		log.Info().Msg("Attempting to unmount filesystem...")

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
	}

	if unmountSuccess {
		fmt.Println("Successfully unmounted fuse server!")
	} else {
		fmt.Println("Warning: Failed to unmount fuse server. You may need to manually unmount with 'fusermount3 -uz mount'")
	}

	// Clean up the test database directory by stopping all services
	fs.StopCacheCleanup()
	fs.StopDeltaLoop()
	fs.StopDownloadManager()
	fs.StopUploadManager()
	fs.SerializeAll()

	// Wait a moment to ensure all file handles are closed
	time.Sleep(100 * time.Millisecond)

	// Remove the test database directory
	if rmErr := os.RemoveAll(testDBLoc); rmErr != nil {
		log.Error().Err(rmErr).Msg("Failed to remove test database location")
	}

	os.Exit(code)
}

// Apparently 200 reqests is the default paging limit.
// Upload at least this many for a later test before the delta thread is created.
func createPagingTestFiles() {
	fmt.Println("Setting up paging test files.")
	var group sync.WaitGroup
	var errCounter int64

	// Create a semaphore to limit concurrent goroutines
	// This prevents resource exhaustion and potential deadlocks
	semaphore := make(chan struct{}, 20) // Limit to 20 concurrent goroutines

	for i := 0; i < 250; i++ {
		group.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		go func(n int, wg *sync.WaitGroup) {
			defer func() {
				<-semaphore // Release semaphore
				wg.Done()
			}()

			_, err := graph.Put(
				graph.ResourcePath(fmt.Sprintf("/onedriver_tests/paging/%d.txt", n))+":/content",
				auth,
				strings.NewReader("test\n"),
			)
			if err != nil {
				log.Error().Err(err).Msg("Paging upload fail.")
				atomic.AddInt64(&errCounter, 1)
			}
		}(i, &group)
	}
	group.Wait()
	log.Info().Msgf("%d failed paging uploads.\n", errCounter)
	fmt.Println("Finished with paging test setup.")
}
