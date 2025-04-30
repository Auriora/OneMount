package offline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/fs"
	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/bcherrington/onemount/internal/testutil"
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
var profiler *fs.Profiler

// validateTestEnvironment checks that all required tools and resources are available
// for the offline tests to run properly. It logs warnings if any dependencies are missing.
func validateTestEnvironment() {
	log.Info().Msg("Validating offline test environment...")

	// Check for fusermount3 (required for mounting/unmounting)
	if _, err := exec.LookPath("fusermount3"); err != nil {
		log.Warn().Err(err).Msg("fusermount3 not found in PATH - unmounting may fail")
	}

	// Check for lsof (used for diagnostics)
	if _, err := exec.LookPath("lsof"); err != nil {
		log.Warn().Err(err).Msg("lsof not found in PATH - some diagnostics will be unavailable")
	}

	// Check for findmnt (used to check mount status)
	if _, err := exec.LookPath("findmnt"); err != nil {
		log.Warn().Err(err).Msg("findmnt not found in PATH - mount status checks may be incomplete")
	}

	// Check if we have permission to create the mount directory
	if err := os.MkdirAll(mountLoc, 0755); err != nil {
		log.Error().Err(err).Str("path", mountLoc).Msg("Cannot create mount directory")
	}

	// Check if we have permission to create the test database directory
	if err := os.MkdirAll(testDBLoc, 0755); err != nil {
		log.Error().Err(err).Str("path", testDBLoc).Msg("Cannot create test database directory")
	}

	// Check available disk space
	var stat syscall.Statfs_t
	if err := syscall.Statfs(testDBLoc, &stat); err == nil {
		// Calculate available space in MB
		availableMB := (stat.Bavail * uint64(stat.Bsize)) / (1024 * 1024)
		if availableMB < 100 {
			log.Warn().Uint64("available_mb", availableMB).Msg("Low disk space for offline tests (< 100MB)")
		} else {
			log.Info().Uint64("available_mb", availableMB).Msg("Sufficient disk space for offline tests")
		}
	}

	// Check if we're running as root (not recommended)
	if os.Geteuid() == 0 {
		log.Warn().Msg("Offline tests are running as root, which is not recommended")
	}

	// Check if we have enough memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	availableMemMB := (m.Sys - m.Alloc) / (1024 * 1024)
	if availableMemMB < 200 {
		log.Warn().Uint64("available_mem_mb", availableMemMB).Msg("Low memory for offline tests (< 200MB)")
	} else {
		log.Info().Uint64("available_mem_mb", availableMemMB).Msg("Sufficient memory for offline tests")
	}

	log.Info().Msg("Offline test environment validation complete")
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
	// Validate the test environment first
	validateTestEnvironment()

	// Setup test environment
	f, setupErr := testutil.SetupTestEnvironment("../../../", false)
	if setupErr != nil {
		log.Error().Err(setupErr).Msg("Failed to setup test environment")
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close log file")
		}
	}()

	// attempt to unmount regardless of what happens (in case previous tests
	// failed and didn't clean themselves up)
	// First check if the mount point exists before attempting to unmount
	if _, err := os.Stat(mountLoc); err == nil {
		if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err != nil {
			log.Error().Err(err).Msg("Warning: Failed to unmount:")
			// Continue anyway as it might not be mounted
		}
	} else {
		log.Warn().Msg("Mount point does not exist, no need to unmount")
	}

	// Remove the mount directory if it exists, then recreate it
	// This ensures we start with a clean state
	if _, err := os.Stat(mountLoc); err == nil {
		if err := os.RemoveAll(mountLoc); err != nil {
			log.Warn().Err(err).Msg("Warning: Failed to remove existing mount directory:")
		}
	}

	// Create the mount directory
	if err := os.MkdirAll(mountLoc, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create mount directory:")
		os.Exit(1)
	}

	// Check if we should use mock authentication
	isMock := os.Getenv("ONEMOUNT_MOCK_AUTH") == "1"

	// Create authenticator based on configuration
	authenticator := graph.NewAuthenticator(graph.AuthConfig{}, testutil.AuthTokensPath, false, isMock)

	// Perform authentication
	var authErr error
	auth, authErr = authenticator.Authenticate()
	if authErr != nil {
		log.Error().Err(authErr).Msg("Authentication failed:")
		os.Exit(1)
	}

	if isMock {
		log.Info().Msg("Using mock authentication for tests")
	}

	f, err := os.OpenFile(testutil.TestLogPath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open log file:")
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

	// Set a unique D-Bus service name prefix for this test run
	// This helps avoid conflicts with other test runs
	uniquePrefix := fmt.Sprintf("offline_test_%d", time.Now().UnixNano())
	fs.SetDBusServiceNamePrefix(uniquePrefix)
	log.Info().Str("dbusPrefix", uniquePrefix).Msg("Set unique D-Bus service name prefix")

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
	if isMock {
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
				Name:          "onemount",
				FsName:        "onemount",
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

		// Add a watchdog timer to detect hangs beyond the timeout
		watchdogTimer := time.AfterFunc(timeout+15*time.Second, func() {
			log.Error().Msg("WATCHDOG: Mount operation appears to be hanging beyond timeout!")
			// Dump goroutine stacks for debugging
			pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)

			// Check if FUSE process is running
			if out, err := exec.Command("ps", "-ef").Output(); err == nil {
				processes := string(out)
				if strings.Contains(processes, "fuse") {
					log.Info().Str("fuse_processes", processes).Msg("FUSE processes found running")
				} else {
					log.Warn().Msg("No FUSE processes found running")
				}
			}

			// Check mount status
			if out, err := exec.Command("mount").Output(); err == nil {
				log.Info().Str("mount_output", string(out)).Msg("Current mount points")
			}

			// Check for open file descriptors related to the mount point
			if out, err := exec.Command("lsof", mountLoc).Output(); err == nil {
				log.Info().Str("open_files", string(out)).Msg("Open files at mount point")
			}
		})
		defer watchdogTimer.Stop()

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

			log.Info().Msg("Mount point is ready and operational")
			return true
		}

		// Use WaitForCondition to wait for the mount point to be ready with enhanced error reporting
		mountCtx, mountCancel := context.WithTimeout(context.Background(), timeout)
		defer mountCancel()

		mountDone := make(chan bool, 1)
		go func() {
			// Use a defer and recover to catch any panics from WaitForCondition
			defer func() {
				if r := recover(); r != nil {
					log.Error().Interface("panic", r).Msg("WaitForCondition panicked")
					select {
					case <-mountCtx.Done():
						return
					default:
						mountDone <- false
					}
				}
			}()

			// WaitForCondition will panic if the condition is not met within the timeout
			// If it returns normally, the condition was met
			testutil.WaitForCondition(nil, isMountReady, timeout, pollInterval, "Filesystem failed to mount within timeout")

			// If we get here, the condition was met
			select {
			case <-mountCtx.Done():
				return
			default:
				mountDone <- true
			}
		}()

		// Wait for mounting to complete or timeout with additional diagnostics
		select {
		case success := <-mountDone:
			if success {
				log.Info().Msg("Mount operation completed successfully")
			} else {
				log.Error().Msg("Mount operation failed")
				log.Error().Msg("Dumping system state for diagnosis...")

				// Check if FUSE process is running
				if out, err := exec.Command("ps", "-ef").Output(); err == nil {
					processes := string(out)
					if strings.Contains(processes, "fuse") {
						log.Info().Str("fuse_processes", processes).Msg("FUSE processes found running")
					} else {
						log.Warn().Msg("No FUSE processes found running")
					}
				}

				// Check mount status
				if out, err := exec.Command("mount").Output(); err == nil {
					log.Info().Str("mount_output", string(out)).Msg("Current mount points")
				}

				// Check for open file descriptors related to the mount point
				if out, err := exec.Command("lsof", mountLoc).Output(); err == nil {
					log.Info().Str("open_files", string(out)).Msg("Open files at mount point")
				}

				os.Exit(1)
			}
		case <-mountCtx.Done():
			// Timeout or cancellation
			log.Error().Err(mountCtx.Err()).Msg("Mount operation timed out")
			log.Error().Msg("Dumping system state for diagnosis...")

			// Check if FUSE process is running
			if out, err := exec.Command("ps", "-ef").Output(); err == nil {
				processes := string(out)
				if strings.Contains(processes, "fuse") {
					log.Info().Str("fuse_processes", processes).Msg("FUSE processes found running")
				} else {
					log.Warn().Msg("No FUSE processes found running")
				}
			}

			// Check mount status
			if out, err := exec.Command("mount").Output(); err == nil {
				log.Info().Str("mount_output", string(out)).Msg("Current mount points")
			}

			// Check for open file descriptors related to the mount point
			if out, err := exec.Command("lsof", mountLoc).Output(); err == nil {
				log.Info().Str("open_files", string(out)).Msg("Open files at mount point")
			}

			os.Exit(1)
		}

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

	// Let the tests create their own directories as needed

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

	// Initialize the profiler for resource monitoring
	profilerDir := filepath.Join(testDBLoc, "offline_profiles")
	profiler = fs.NewProfiler(profilerDir)
	if profiler == nil {
		log.Error().Msg("Failed to create profiler")
	} else {
		log.Info().Str("dir", profilerDir).Msg("Created profiler for resource monitoring")

		// Create a channel to signal the monitoring goroutine to stop
		monitoringStopChan := make(chan struct{})

		// Start a goroutine to periodically capture resource usage
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			// Capture initial resource usage
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Info().
				Int("goroutines", runtime.NumGoroutine()).
				Uint64("alloc_mb", m.Alloc/1024/1024).
				Uint64("total_alloc_mb", m.TotalAlloc/1024/1024).
				Uint64("sys_mb", m.Sys/1024/1024).
				Uint32("gc_cycles", m.NumGC).
				Msg("Initial resource usage (offline mode)")

			for {
				select {
				case <-ticker.C:
					// Capture memory usage
					if err := profiler.Stop(fs.ProfileMemory); err != nil {
						log.Error().Err(err).Msg("Failed to capture memory profile")
					}

					// Capture goroutine count
					if err := profiler.Stop(fs.ProfileGoroutine); err != nil {
						log.Error().Err(err).Msg("Failed to capture goroutine profile")
					}

					// Log current resource usage
					runtime.ReadMemStats(&m)
					log.Info().
						Int("goroutines", runtime.NumGoroutine()).
						Uint64("alloc_mb", m.Alloc/1024/1024).
						Uint64("total_alloc_mb", m.TotalAlloc/1024/1024).
						Uint64("sys_mb", m.Sys/1024/1024).
						Uint32("gc_cycles", m.NumGC).
						Msg("Resource usage (offline mode)")
				case <-monitoringStopChan:
					// Stop monitoring when signaled
					log.Info().Msg("Stopping resource monitoring")
					return
				}
			}
		}()

		// Add a deferred function to stop the monitoring goroutine
		defer func() {
			close(monitoringStopChan)
			// Wait a moment for the goroutine to exit
			time.Sleep(100 * time.Millisecond)
			log.Info().Msg("Resource monitoring stopped")
		}()
	}

	// Capture the initial state of the filesystem before running tests
	initialState, initialStateErr := testutil.CaptureFileSystemState(mountLoc)
	if initialStateErr != nil {
		log.Error().Err(initialStateErr).Msg("Failed to capture initial filesystem state")
	} else {
		log.Info().Int("files", len(initialState)).Msg("Captured initial filesystem state")
	}

	// Create a channel to signal when cleanup is done
	cleanupDone := make(chan struct{})

	// Define a comprehensive cleanup function that will run even if tests panic
	cleanupFunc := func() {
		// Avoid running cleanup multiple times
		select {
		case <-cleanupDone:
			return // Already cleaned up
		default:
			defer close(cleanupDone)
		}

		log.Info().Msg("Running comprehensive cleanup...")

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

		// Wait for a short period to allow file handles to close
		select {
		case <-ctx.Done():
			log.Warn().Msg("Timeout waiting for file handles to close")
		case <-time.After(500 * time.Millisecond):
			log.Info().Msg("Waited for file handles to close")
		}

		// Stop the UnmountHandler goroutine if it exists
		if unmountDone != nil {
			select {
			case <-unmountDone:
				// Already closed
			default:
				close(unmountDone)
			}
		}

		// Wait for the UnmountHandler to exit
		select {
		case <-ctx.Done():
			log.Warn().Msg("Timeout waiting for UnmountHandler to exit")
		case <-time.After(500 * time.Millisecond):
			log.Info().Msg("Waited for UnmountHandler to exit")
		}

		// Stop signal notifications if they exist
		if sigChan != nil {
			signal.Stop(sigChan)
		}

		// Check if we're using mock authentication
		if isMock {
			// Skip unmounting when using mock authentication
			log.Info().Msg("Skipping FUSE unmounting for tests with mock authentication")
		} else {
			// Attempt to unmount with retries and more aggressive approach
			log.Info().Msg("Attempting to unmount filesystem with enhanced retry logic...")

			// Check if the mount point exists before attempting to unmount
			if _, err := os.Stat(mountLoc); err != nil {
				log.Warn().Err(err).Msg("Mount point does not exist, no need to unmount")
			} else {
				// First try normal unmount
				unmountErr := server.Unmount()
				if unmountErr == nil {
					log.Info().Msg("Successfully unmounted filesystem")
				} else {
					log.Error().Err(unmountErr).Msg("Failed to unmount test fuse server, attempting lazy unmount")

					// Try lazy unmount with more retries and longer waits
					for i := 0; i < 5; i++ {
						if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err == nil {
							log.Info().Msg("Successfully performed lazy unmount")
							break
						} else {
							log.Error().Err(err).Int("attempt", i+1).Msg("Failed to perform lazy unmount")
							// Longer wait between retries
							time.Sleep(1 * time.Second)
						}
					}

					// Final check if mount still exists
					if _, err := os.Stat(mountLoc); err == nil {
						// Try the most aggressive approach - kill any processes using the mount point
						log.Warn().Msg("Mount point still exists after all unmount attempts, trying to find and kill processes using it")

						// Find processes using the mount point
						findCmd := exec.Command("lsof", mountLoc)
						findOutput, _ := findCmd.CombinedOutput()
						log.Info().Str("lsof_output", string(findOutput)).Msg("Processes using mount point")

						// Try one final lazy unmount
						if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err != nil {
							log.Error().Err(err).Msg("Final unmount attempt failed")
						} else {
							log.Info().Msg("Final unmount attempt succeeded")
						}
					}
				}
			}
		}

		// Capture the final state of the filesystem after tests
		if initialStateErr == nil {
			finalState, finalStateErr := testutil.CaptureFileSystemState(mountLoc)
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

		log.Info().Msg("Comprehensive cleanup completed")
	}

	// Setup cleanup to run even if tests panic
	defer cleanupFunc()

	// Register the cleanup function with a goroutine that will be triggered on process exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("Received termination signal, running emergency cleanup...")
		cleanupFunc()
		os.Exit(1)
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
	if isMock {
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
		log.Info().Msg("Successfully unmounted fuse server!")
	} else {
		log.Warn().Msg("Warning: Failed to unmount fuse server. Continuing with exit anyway to prevent hanging.")
		// Make one final attempt with the most aggressive unmount option
		if err := exec.Command("fusermount3", "-uz", mountLoc).Run(); err != nil {
			log.Error().Err(err).Msg("Final attempt at lazy unmount failed")
		} else {
			log.Info().Msg("Final lazy unmount succeeded")
		}
	}

	os.Exit(code)
}
