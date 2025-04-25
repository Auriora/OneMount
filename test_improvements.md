# Test Setup and Teardown Improvements

This document outlines recommendations for improving the test setup and teardown code across the onedriver project. These recommendations aim to ensure consistent setup and teardown across all tests, proper state checking before tests start, proper cleanup after tests finish, recovery from failures, and the ability to reset to a known state.

## Current Issues

After reviewing the test setup and teardown code across the project, the following issues were identified:

1. **Inconsistent error handling**:
   - `fs/setup_test.go` and `fs/offline/setup_test.go` have thorough error handling
   - `fs/graph/setup_test.go`, `cmd/common/setup_test.go`, `ui/setup_test.go`, and `ui/systemd/setup_test.go` have minimal or no error handling

2. **Inconsistent cleanup**:
   - `fs/setup_test.go` and `fs/offline/setup_test.go` capture initial and final filesystem state and attempt to clean up resources
   - `fs/graph/setup_test.go`, `cmd/common/setup_test.go`, `ui/setup_test.go`, and `ui/systemd/setup_test.go` don't have cleanup mechanisms

3. **Inconsistent state checking**:
   - `fs/setup_test.go` and `fs/offline/setup_test.go` check if the mount point is available and working
   - `ui/setup_test.go` and `ui/systemd/setup_test.go` check if the mount directory exists but not if it's in use
   - `fs/graph/setup_test.go` and `cmd/common/setup_test.go` don't check the state at all

4. **Inconsistent recovery from failures**:
   - `fs/setup_test.go` and `fs/offline/setup_test.go` have mechanisms to recover from failures
   - The other setup files don't

5. **Inconsistent use of t.Cleanup()**:
   - `fs/offline/offline_test.go` consistently uses `t.Cleanup()` for resource cleanup
   - `fs/fs_test.go` has a mix of `t.Cleanup()` and manual cleanup

6. **Duplication**:
   - `ui/setup_test.go` and `ui/systemd/setup_test.go` are almost identical
   - There's duplication of code between `fs/setup_test.go` and `fs/offline/setup_test.go`

7. **Potential bugs**:
   - No check if the mount point is already in use by another process
   - No timeout for unmounting in some cases
   - No handling of concurrent test runs
   - No handling of test interruption (e.g., Ctrl+C)

8. **Documentation**:
   - Limited or no documentation in most setup files
   - No explanation of the test environment requirements
   - No explanation of the test setup and teardown process

## Recommendations

### 1. Consistent Error Handling (COMPLETED)

All test setup files should have consistent error handling:

```go
// Example of good error handling
if err := someOperation(); err != nil {
    log.Error().Err(err).Msg("Failed to perform operation")
    os.Exit(1)
}
```

**Implementation**: Added consistent error handling to all test setup files:
- `fs/graph/setup_test.go`: Added error handling for Chdir and OpenFile operations
- `cmd/common/setup_test.go`: Added error handling for Chdir, RemoveAll, and OpenFile operations
- `ui/setup_test.go`: Added zerolog import, logging setup, and error handling for Chdir, Remove, and Mkdir operations
- `ui/systemd/setup_test.go`: Added zerolog import, logging setup, and error handling for Chdir, Remove, and Mkdir operations

### 2. Proper State Checking Before Tests Start (IN PROGRESS)

Before starting tests, check if the environment is in the expected state:

```go
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
```

**Implementation Status**: Attempted to implement in `fs/setup_test.go` and `fs/offline/setup_test.go`, but encountered compilation errors related to the `UnmountHandler` function. The error message "Too many arguments in call to 'UnmountHandler'" suggests a mismatch between the function signature and the calls in the test files. Further investigation is needed to resolve this issue before proceeding with the implementation.

### 3. Proper Cleanup After Tests Finish

Ensure all resources are cleaned up after tests finish, even if they fail:

```go
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
```

### 4. Recovery from Failures

Implement robust recovery mechanisms to handle failures during test setup and teardown:

```go
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
```

### 5. Consistent Use of t.Cleanup()

Use `t.Cleanup()` consistently across all tests to ensure resources are cleaned up even if tests fail:

```go
func TestSomething(t *testing.T) {
    // Create a resource
    resource := createResource()

    // Setup cleanup to run after test completes or fails
    t.Cleanup(func() {
        if err := cleanupResource(resource); err != nil {
            t.Logf("Warning: Failed to clean up resource: %v", err)
        }
    })

    // Test logic here
}
```

### 6. Reduce Duplication

Extract common code into shared functions or packages:

```go
// Common function for checking and unmounting a mount point
func checkAndUnmountMountPoint(mountLoc string) bool {
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
        // Unmount logic here
    }

    return isMounted
}
```

### 7. Improve Documentation

Add comprehensive documentation to explain the test setup and teardown process:

```go
// TestMain is the entry point for all tests in this package.
// It sets up the test environment, runs the tests, and cleans up afterward.
//
// The setup process:
// 1. Changes the working directory to the project root
// 2. Checks if the mount point is already in use and unmounts it if necessary
// 3. Creates the mount directory if it doesn't exist
// 4. Sets up logging
// 5. Authenticates with Microsoft Graph API
// 6. Initializes the filesystem
// 7. Mounts the filesystem with FUSE
// 8. Sets up signal handlers for graceful unmount
// 9. Creates test directories and files
// 10. Captures the initial state of the filesystem
//
// The teardown process:
// 1. Captures the final state of the filesystem
// 2. Compares it with the initial state to identify files that weren't cleaned up
// 3. Attempts to clean up those files
// 4. Stops all filesystem services
// 5. Unmounts the filesystem
// 6. Removes the test database directory
func TestMain(m *testing.M) {
    // Setup code here

    // Run tests
    code := m.Run()

    // Teardown code here

    os.Exit(code)
}
```

## Implementation Plan

1. **Update fs/setup_test.go**:
   - Add proper state checking before tests start
   - Improve cleanup after tests finish
   - Add recovery from failures
   - Add comprehensive documentation

2. **Update fs/offline/setup_test.go**:
   - Add proper state checking before tests start
   - Improve cleanup after tests finish
   - Add recovery from failures
   - Add comprehensive documentation

3. **Update fs/graph/setup_test.go**:
   - Add error handling
   - Add cleanup mechanisms
   - Add documentation

4. **Update cmd/common/setup_test.go**:
   - Add error handling
   - Add cleanup mechanisms
   - Add documentation

5. **Update ui/setup_test.go and ui/systemd/setup_test.go**:
   - Extract common code to reduce duplication
   - Add error handling
   - Add cleanup mechanisms
   - Add documentation

6. **Update fs/fs_test.go and fs/offline/offline_test.go**:
   - Ensure consistent use of t.Cleanup() for resource cleanup
   - Add documentation

## Conclusion

Implementing these recommendations will ensure consistent setup and teardown across all tests, proper state checking before tests start, proper cleanup after tests finish, recovery from failures, and the ability to reset to a known state. This will make the tests more reliable, easier to maintain, and less prone to leaving behind artifacts that could interfere with subsequent test runs.
