# OneMount Test Setup Documentation

This document describes the test setup routines used in the OneMount project. These routines are responsible for setting up the test environment before tests are run and cleaning up after tests complete.

## Setup Files and Their Locations

The following table lists all setup_test.go files in the project and their locations:

| Setup File | Location |
|------------|----------|
| cmd/common/setup_test.go | /home/bcherrington/Projects/Goland/tmp/OneMount/cmd/common/setup_test.go |
| internal/fs/graph/setup_test.go | /home/bcherrington/Projects/Goland/tmp/OneMount/internal/fs/graph/setup_test.go |
| internal/fs/offline/setup_test.go | /home/bcherrington/Projects/Goland/tmp/OneMount/internal/fs/offline/setup_test.go |
| internal/fs/setup_test.go | /home/bcherrington/Projects/Goland/tmp/OneMount/internal/fs/setup_test.go |
| internal/ui/setup_test.go | /home/bcherrington/Projects/Goland/tmp/OneMount/internal/ui/setup_test.go |
| internal/ui/systemd/setup_test.go | /home/bcherrington/Projects/Goland/tmp/OneMount/internal/ui/systemd/setup_test.go |

## Detailed Setup Documentation

### cmd/common/setup_test.go

**Purpose**: Sets up the test environment for the cmd/common package tests.

**Setup Routine**: TestMain

**Description**: 
The TestMain function is a special function recognized by the Go testing package. It's called before any tests in the package are run and is responsible for setting up the test environment and cleaning up after all tests have completed.

**Implementation Details**:
1. Calls `testutil.SetupTestEnvironment("../..", false)` to set up the test environment:
   - The first parameter ("../..") is the relative path to the project root
   - The second parameter (false) indicates that this is not an offline test
2. The SetupTestEnvironment function returns a file handle for the log file
3. Defers closing the log file to ensure it's closed after all tests complete
4. Calls `os.Exit(m.Run())` to run all the tests in the package and exit with the appropriate status code

**Error Handling**:
- If setting up the test environment fails, logs an error message and exits with status code 1
- If closing the log file fails, logs an error message

**Dependencies**:
- github.com/bcherrington/onemount/internal/testutil
- github.com/rs/zerolog/log
- os
- testing

### internal/fs/graph/setup_test.go

**Purpose**: Sets up the test environment for the graph package tests, which interact with the Microsoft Graph API.

**Setup Routine**: TestMain

**Description**: 
This setup routine prepares the environment for testing the Graph API integration. It handles authentication, creates necessary test directories, and ensures proper cleanup after tests.

**Implementation Details**:
1. Calls `testutil.SetupTestEnvironment("../../../", false)` to set up the test environment
2. Authenticates with Microsoft Graph API (or uses mock authentication if ONEMOUNT_MOCK_AUTH=1)
3. Retrieves user and drive information (or creates mock versions if using mock authentication)
4. Logs account and drive type information for debugging
5. Creates a test directory for capturing filesystem state
6. Ensures the dmel.fa file exists for hash tests
7. Captures the initial state of the filesystem
8. Runs the tests
9. Performs cleanup after tests, including:
   - Capturing the final state of the filesystem
   - Cleaning up any files created during tests
   - Removing the test directory

**Error Handling**:
- If setting up the test environment fails, logs an error message and exits with status code 1
- If authentication fails, logs an error message and exits with status code 1
- If creating the test directory fails, logs an error message and exits with status code 1
- If closing the log file fails, logs an error message

**Dependencies**:
- github.com/bcherrington/onemount/internal/testutil
- github.com/rs/zerolog/log
- os
- path/filepath
- testing

### internal/fs/offline/setup_test.go

**Purpose**: Sets up the test environment for offline mode testing, simulating scenarios where network connectivity is unavailable.

**Setup Routine**: TestMain

**Description**: 
This setup routine creates a comprehensive test environment for testing the filesystem's behavior in offline mode. It mounts a filesystem, creates test files, then simulates offline mode by setting the appropriate flags.

**Implementation Details**:
1. Validates the test environment by checking for required tools and resources
2. Calls `testutil.SetupTestEnvironment("../../../", false)` to set up the test environment
3. Attempts to unmount any existing filesystem and recreates the mount directory
4. Authenticates with Microsoft Graph API (or uses mock authentication)
5. Sets up logging
6. Initializes the filesystem with cached data from previous tests
7. Mounts the filesystem with FUSE (or skips mounting when using mock authentication)
8. Creates test files before entering offline mode
9. Sets the operational offline state to true to simulate offline mode
10. Sets the filesystem's offline mode to ReadWrite
11. Verifies that files are accessible in offline mode
12. Captures the initial state of the filesystem
13. Runs the tests with a timeout
14. Performs cleanup after tests, including:
    - Resetting the offline state
    - Stopping all filesystem services
    - Unmounting the filesystem

**Error Handling**:
- If setting up the test environment fails, logs an error message and exits with status code 1
- If authentication fails, logs an error message and exits with status code 1
- If mounting the filesystem fails, logs diagnostic information and exits with status code 1
- If creating test files fails, logs an error message and exits with status code 1
- If tests timeout after 8 minutes, forces cleanup and exits with status code 1
- If unmounting fails, makes multiple attempts with different strategies

**Dependencies**:
- github.com/bcherrington/onemount/internal/fs
- github.com/bcherrington/onemount/internal/fs/graph
- github.com/bcherrington/onemount/internal/testutil
- github.com/hanwen/go-fuse/v2/fuse
- github.com/rs/zerolog
- github.com/rs/zerolog/log
- context
- fmt
- os
- os/exec
- os/signal
- path/filepath
- runtime
- runtime/pprof
- strings
- syscall
- testing
- time

### internal/fs/setup_test.go

**Purpose**: Sets up the test environment for the filesystem package tests, which test the core filesystem functionality.

**Setup Routine**: TestMain

**Description**: 
This setup routine creates a comprehensive test environment for testing the filesystem. It mounts a filesystem with FUSE, creates test directories and files, and sets up resource monitoring.

**Implementation Details**:
1. Validates the test environment by checking for required tools and resources
2. Limits the number of CPUs that can execute simultaneously to reduce resource contention
3. Calls `testutil.SetupTestEnvironment("..", true)` to set up the test environment
4. Sets a unique D-Bus service name prefix for this test run
5. Authenticates with Microsoft Graph API (or uses mock authentication)
6. Initializes the filesystem
7. Mounts the filesystem with FUSE (or skips mounting when using mock authentication)
8. Sets up signal handlers for graceful unmount
9. Creates test directories and files, including special files for paging tests
10. Ensures the filesystem is fully initialized before running tests
11. Sets up resource monitoring using a profiler
12. Captures the initial state of the filesystem
13. Runs the tests
14. Performs cleanup after tests, including:
    - Waiting for any remaining uploads to complete
    - Unmounting the filesystem
    - Stopping all filesystem services
    - Removing the test database directory

**Error Handling**:
- If setting up the test environment fails, logs an error message and exits with status code 1
- If authentication fails, logs an error message and exits with status code 1
- If initializing the filesystem fails, logs an error message and exits with status code 1
- If creating the FUSE server fails, logs an error message and exits with status code 1
- If mounting the filesystem fails, logs an error message and exits with status code 1
- If creating test directories fails, logs an error message and exits with status code 1
- If unmounting fails, makes multiple attempts with different strategies

**Dependencies**:
- github.com/bcherrington/onemount/internal/fs/graph
- github.com/bcherrington/onemount/internal/testutil
- github.com/hanwen/go-fuse/v2/fuse
- github.com/rs/zerolog/log
- context
- fmt
- os
- os/exec
- os/signal
- path/filepath
- runtime
- strings
- sync
- sync/atomic
- syscall
- testing
- time

### internal/ui/setup_test.go

**Purpose**: Sets up the test environment for the UI package tests.

**Setup Routine**: TestMain

**Description**: 
This setup routine prepares the environment for testing the UI components. It's a simpler setup compared to the filesystem tests since it doesn't involve mounting a filesystem.

**Implementation Details**:
1. Calls `testutil.SetupUITest("../")` to set up the UI test environment
2. Defers closing the log file to ensure it's closed after all tests complete
3. Calls `os.Exit(m.Run())` to run all the tests in the package and exit with the appropriate status code

**Error Handling**:
- If setting up the test environment fails, exits with status code 1
- If closing the log file fails, panics

**Dependencies**:
- github.com/bcherrington/onemount/internal/testutil
- os
- testing

### internal/ui/systemd/setup_test.go

**Purpose**: Sets up the test environment for the systemd package tests, which test the integration with systemd.

**Setup Routine**: TestMain

**Description**: 
This setup routine prepares the environment for testing the systemd integration. It's similar to the UI setup but with a different path to the project root.

**Implementation Details**:
1. Calls `testutil.SetupUITest("../..")` to set up the UI test environment
2. Defers closing the log file to ensure it's closed after all tests complete
3. Calls `os.Exit(m.Run())` to run all the tests in the package and exit with the appropriate status code

**Error Handling**:
- If setting up the test environment fails, exits with status code 1
- If closing the log file fails, panics

**Dependencies**:
- github.com/bcherrington/onemount/internal/testutil
- os
- testing
