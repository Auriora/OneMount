# OneMount Test Utilities Documentation

This document describes the test utilities provided in the `internal/testutil` package and its subpackages. These utilities are designed to facilitate testing of the OneMount filesystem by providing common functionality for test setup, file operations, asynchronous operations, and more.

## Overview

The `internal/testutil` package contains a variety of utilities for testing different aspects of the OneMount filesystem:

- **Setup Utilities**: Functions for setting up the test environment, including changing to the project root directory, setting up logging, and ensuring test directories exist.
- **File Utilities**: Functions for creating, manipulating, and verifying files and directories during tests.
- **Asynchronous Utilities**: Functions for handling asynchronous operations, such as waiting for conditions, retrying operations with backoff, and running operations with timeouts.
- **UI Test Utilities**: Functions specifically designed for testing the UI components of OneMount.
- **Graph API Test Utilities**: Functions for creating test fixtures for the Microsoft Graph API integration.
- **Filesystem Test Utilities**: Functions for testing the filesystem implementation, including mounting, unmounting, and verifying filesystem state.

## Main Package (`internal/testutil`)

### Constants (`test_constants.go`)

**Purpose**: Defines constants used throughout the test utilities, such as paths for test directories, mount points, and log files.

**Constants**:
- `TestSandboxDir`: The directory used for test files.
- `TestSandboxTmpDir`: The directory used for temporary files.
- `TestMountPoint`: The location where the filesystem is mounted during tests.
- `TestDir`: The directory within the mount point used for tests.
- `TestDBLoc`: The location of the test database.
- `DeltaDir`: The directory used for delta tests.
- `DmelfaDir`: The path to the dmel.fa file used for tests.
- `AuthTokensPath`: The path to the authentication tokens file.
- `TestLogPath`: The path to the test log file.

**Usage Example**:
```
// Create a file in the test sandbox directory
err := os.WriteFile(filepath.Join(testutil.TestSandboxDir, "test.txt"), []byte("test"), 0644)
```

### Setup Utilities (`setup.go`)

**Purpose**: Provides functions for setting up the test environment, including changing to the project root directory, setting up logging, and ensuring test directories exist.

**Functions**:
- `SetupTestEnvironment`: Performs common setup tasks for tests, including changing to the project root directory, setting up logging, and ensuring test directories exist.
- `changeToProjectRoot`: Changes the current working directory to the project root.
- `setupLogging`: Sets up logging for tests.
- `ensureTestDirectories`: Ensures that all test directories exist and are clean.

**Usage Example**:
```
// Setup the test environment
logFile, err := testutil.SetupTestEnvironment("../..", false)
if err != nil {
    t.Fatalf("Failed to setup test environment: %v", err)
}
defer logFile.Close()
```

### File Utilities (`file.go`)

**Purpose**: Provides functions for creating, manipulating, and verifying files and directories during tests.

**Functions**:
- `CreateTestFile`: Creates a file with the given content and ensures it's cleaned up after the test.
- `CreateTestDir`: Creates a directory and ensures it's cleaned up after the test.
- `CreateTempDir`: Creates a temporary directory and ensures it's cleaned up after the test.
- `CreateTempFile`: Creates a temporary file with the given content and ensures it's cleaned up after the test.
- `FileExists`: Checks if a file exists at the given path.
- `FileContains`: Checks if a file contains the expected content.
- `AssertFileExists`: Asserts that a file exists at the given path.
- `AssertFileNotExists`: Asserts that a file does not exist at the given path.
- `AssertFileContains`: Asserts that a file contains the expected content.
- `CaptureFileSystemState`: Captures the current state of the filesystem by listing all files and directories.

**Usage Example**:
```
// Create a test file
filePath := testutil.CreateTestFile(t, testutil.TestSandboxTmpDir, "test.txt", []byte("test content"))

// Verify the file exists and contains the expected content
testutil.AssertFileExists(t, filePath)
testutil.AssertFileContains(t, filePath, []byte("test content"))
```

### Asynchronous Utilities (`async.go`)

**Purpose**: Provides functions for handling asynchronous operations, such as waiting for conditions, retrying operations with backoff, and running operations with timeouts.

**Functions**:
- `WaitForCondition`: Waits for a condition to be true with a configurable timeout and polling interval.
- `WaitForConditionWithContext`: Waits for a condition to be true with a context for cancellation.
- `RetryWithBackoff`: Retries an operation with exponential backoff until it succeeds or times out.
- `RunWithTimeout`: Runs an operation with a timeout.
- `RunConcurrently`: Runs multiple operations concurrently and waits for all to complete.
- `WaitForFileChange`: Waits for a file to change (by checking its modification time).
- `WaitForFileExistence`: Waits for a file to exist or not exist.

**Usage Example**:
```
// Wait for a file to be created
testutil.WaitForFileExistence(t, filePath, true, 5*time.Second, 100*time.Millisecond)

// Retry an operation with backoff
err := testutil.RetryWithBackoff(t, func() error {
    return someOperationThatMightFail()
}, 5, 100*time.Millisecond, 1*time.Second, "Operation failed")
```

### UI Test Utilities (`ui_test_utils.go`)

**Purpose**: Provides functions specifically designed for testing the UI components of OneMount.

**Functions**:
- `SetupUITest`: Performs common setup tasks for UI tests, including changing to the project root directory, setting up logging, and ensuring the mount directory exists and is clean.
- `EnsureMountPoint`: Checks if the mount point exists and is accessible, and creates or recreates it if necessary.

**Usage Example**:
```
// Setup UI test environment
logFile, err := testutil.SetupUITest("../..")
if err != nil {
    t.Fatalf("Failed to setup UI test environment: %v", err)
}
defer logFile.Close()

// Ensure mount point exists
err = testutil.EnsureMountPoint(testutil.TestMountPoint)
if err != nil {
    t.Fatalf("Failed to ensure mount point: %v", err)
}
```

### Dmelfa Generator (`dmelfa_generator.go`)

**Purpose**: Provides functions for generating a large test file (dmel.fa) with random DNA sequence data, used for testing large file operations.

**Functions**:
- `GenerateDmelfa`: Generates a dmel.fa file with random data if it doesn't exist.
- `EnsureDmelfaExists`: Ensures that the dmel.fa file exists before tests run.

**Usage Example**:
```
// Ensure dmel.fa file exists
testutil.EnsureDmelfaExists()

// Use the file in tests
file, err := os.Open(testutil.DmelfaDir)
if err != nil {
    t.Fatalf("Failed to open dmel.fa: %v", err)
}
defer file.Close()
```

## Subpackages

### Common Utilities (`internal/testutil/common`)

**Purpose**: Provides common utility functions used across different test packages.

**Functions**:
- `WaitForCondition`: Waits for a condition to be true with a timeout.
- `RetryWithBackoff`: Retries an operation with exponential backoff.
- `CheckAndUnmountMountPoint`: Checks if a mount point is in use and attempts to unmount it.
- `WaitForMount`: Waits for a filesystem to be mounted.
- `CleanupFilesystemState`: Cleans up the filesystem state after tests.

**Usage Example**:
```
// Wait for a condition to be true
common.WaitForCondition(t, func() bool {
    return someCondition()
}, 5*time.Second, 100*time.Millisecond, "Condition not met")

// Check and unmount a mount point
unmounted := common.CheckAndUnmountMountPoint(testutil.TestMountPoint)
if !unmounted {
    t.Fatalf("Failed to unmount mount point")
}
```

### Filesystem Test Utilities (`internal/testutil/fs`)

**Purpose**: Provides utility functions for testing the filesystem implementation.

**Functions**:
- `CheckAndUnmountMountPoint`: Checks if a mount point is in use and attempts to unmount it.
- `WaitForMount`: Waits for a filesystem to be mounted.
- `UnmountWithRetries`: Attempts to unmount a filesystem with retries.
- `CleanupFilesystemState`: Cleans up the filesystem state after tests.
- `StopFilesystemServices`: Stops all filesystem services.

**Usage Example**:
```
// Wait for the filesystem to be mounted
mounted, err := fs.WaitForMount(testutil.TestMountPoint, 5*time.Second)
if err != nil || !mounted {
    t.Fatalf("Failed to wait for mount: %v", err)
}

// Unmount the filesystem with retries
unmounted := fs.UnmountWithRetries(server, testutil.TestMountPoint)
if !unmounted {
    t.Fatalf("Failed to unmount filesystem")
}
```

### Graph API Test Utilities (`internal/testutil/graph`)

**Purpose**: Provides utility functions for creating test fixtures for the Microsoft Graph API integration.

**Functions**:
- `StandardTestFile`: Returns a standard test file content with predictable content.
- `CreateDriveItemFixture`: Creates a DriveItem fixture for testing.
- `CreateFileItemFixture`: Creates a DriveItem fixture representing a file.
- `CreateFolderItemFixture`: Creates a DriveItem fixture representing a folder.
- `CreateDeletedItemFixture`: Creates a DriveItem fixture representing a deleted item.
- `CreateChildrenFixture`: Creates a slice of DriveItem fixtures representing children of a folder.
- `CreateNestedFolderStructure`: Creates a nested folder structure for testing.
- `CreateDriveItemWithConflict`: Creates a DriveItem fixture with conflict behavior set.

**Usage Example**:
```
// Create a file item fixture
fileItem := graph.CreateFileItemFixture("test.txt", 1024, graph.StandardTestFile())

// Create a folder item fixture
folderItem := graph.CreateFolderItemFixture("test-folder", 5)

// Create children fixtures
children := graph.CreateChildrenFixture(folderItem.ID, 10)
```

## Best Practices for Using Test Utilities

1. **Always clean up after tests**: Use the cleanup functions provided by the test utilities to ensure that the test environment is clean after tests run.
2. **Use the appropriate utility for the task**: Choose the utility that best fits the task at hand, rather than reinventing the wheel.
3. **Handle errors appropriately**: Check for errors returned by the utility functions and handle them appropriately.
4. **Use the constants for file paths**: Use the constants defined in `test_constants.go` for file paths to ensure consistency across tests.
5. **Set up the test environment properly**: Use the `SetupTestEnvironment` function to set up the test environment before running tests.
6. **Use the asynchronous utilities for operations that might take time**: Use the asynchronous utilities for operations that might take time to complete, such as waiting for files to be created or conditions to be met.
7. **Use the UI test utilities for UI tests**: Use the UI test utilities for testing the UI components of OneMount.
8. **Use the Graph API test utilities for Graph API tests**: Use the Graph API test utilities for testing the Graph API integration.
9. **Use the filesystem test utilities for filesystem tests**: Use the filesystem test utilities for testing the filesystem implementation.

## Conclusion

The test utilities provided in the `internal/testutil` package and its subpackages are designed to facilitate testing of the OneMount filesystem by providing common functionality for test setup, file operations, asynchronous operations, and more. By using these utilities, you can write more concise, readable, and maintainable tests.
