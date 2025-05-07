# OneMount Test Implementation Details

## Overview

This document provides comprehensive details for implementing the testing framework and test cases for the OneMount project. It includes detailed task descriptions, implementation recommendations, Junie AI prompts for implementation, and project review recommendations. This document complements the [Test Implementation Execution Plan](test-implementation-execution-plan.md), which focuses on the work breakdown, priority, and schedule.

## 1. Core Framework Implementation Details

### 1.1 Enhanced Resource Management

**Description**: Implement the `FileSystemResource` type and related functionality to enhance the TestFramework's ability to handle complex resources like mounted filesystems.

**Implementation Details**:
- Create a `FileSystemResource` struct that implements the `Resource` interface
- Add methods for mounting and unmounting filesystems
- Implement proper cleanup mechanisms for all resources
- Ensure thread-safety with appropriate mutex usage

**Files to Modify**:
- `internal/testutil/framework/framework.go`
- `internal/testutil/framework/resources.go`

### 1.2 Signal Handling

**Description**: Add signal handling capabilities to the TestFramework to ensure proper cleanup when tests are interrupted.

**Implementation Details**:
- Add a `SetupSignalHandling` method to the TestFramework
- Register signal handlers for SIGINT and SIGTERM
- Ensure all resources are properly cleaned up when signals are received
- Use a channel to coordinate signal handling and cleanup

**Files to Modify**:
- `internal/testutil/framework/framework.go`

### 1.3 Upload API Race Condition Fix

**Description**: Fix the race condition in the `UploadManager` by enhancing the `WaitForUpload` method to handle cases where a session hasn't been added to the sessions map yet.

**Implementation Details**:
- Add a new `GetSession` method to provide thread-safe access to session information
- Enhance the `WaitForUpload` method to wait for session creation with a timeout
- Improve error messages to help diagnose issues
- Ensure thread-safety with appropriate mutex usage

**Files to Modify**:
- `internal/fs/upload_manager.go`

**Junie AI Prompt**:

```
# Junie Prompt: Redesign Upload API for Robust Session Handling

## Task Overview
Implement Solution 5 from the race condition analysis to redesign the Upload API, making it more robust by enhancing the `WaitForUpload` method to handle cases where a session hasn't been added to the sessions map yet.

## Current Issue
There's a race condition in the `UploadManager` between queuing an upload and waiting for it. The `WaitForUpload` method checks if the upload session exists in the `sessions` map, but this map is only populated when the session is processed by the `uploadLoop`, which runs on a ticker. This causes test failures when `WaitForUpload` is called immediately after `QueueUploadWithPriority`.

## Proposed Changes

### 1. Requirements Impact Analysis
- No changes to the Software Requirements Specification (SRS) are needed
- This change improves the robustness of the API without changing its functional requirements
- The change maintains backward compatibility with existing code
- The change addresses a race condition that could affect reliability in production environments

### 2. Architecture Document Changes
Add the following to the architecture document:

```
#### Upload Manager Session Handling
The UploadManager now includes enhanced session handling to prevent race conditions between queuing uploads and waiting for them to complete. The `WaitForUpload` method has been improved to handle cases where a session hasn't been processed by the upload loop yet, making the API more resilient to timing issues.
```

### 3. Design Documentation Changes
Update the design documentation with:

```
#### Upload API Robustness Improvements
The Upload API has been enhanced to handle race conditions between session creation and waiting:

1. `WaitForUpload` now includes a waiting period for session creation
2. A new helper method `GetSession` provides thread-safe access to session information
3. Timeout mechanisms prevent indefinite waiting for sessions that may never be created
4. Error messages are more descriptive to help diagnose issues
```

### 4. Implementation Details

Modify `upload_manager.go` to include:

1. Add a new `GetSession` method
2. Enhance the `WaitForUpload` method

### 5. Refactoring Dependent Code
No changes to dependent code are required as this implementation maintains the same API signature and behavior, only making it more robust.

### 6. Testing Strategy

#### Unit Tests
1. Test `WaitForUpload` with a session that already exists
2. Test `WaitForUpload` with a session that doesn't exist yet but is added shortly after
3. Test `WaitForUpload` with a session that is never added (should timeout)
4. Test `WaitForUpload` with a session that is removed during waiting

#### Integration Tests
1. Fix the existing `TestUT_FS_05_02_RepeatedUploads_OnlineMode_SuccessfulUpload` test without adding delays
2. Add a stress test that rapidly queues and waits for multiple uploads

### 7. Implementation Considerations
- The timeout value (5 seconds) should be configurable or at least carefully chosen
- Error messages should be descriptive to help diagnose issues
- Consider adding logging to help debug timing issues
- The implementation should be thread-safe and handle concurrent calls to `WaitForUpload`

### 8. Backward Compatibility
This change maintains backward compatibility with existing code that uses `WaitForUpload`, as it only adds functionality without changing the method signature or expected behavior.

## Expected Outcome
After implementing these changes:
1. The race condition in tests will be resolved without adding unreliable delays
2. The API will be more robust for all users, not just in tests
3. Error messages will be more descriptive
4. The code will handle edge cases more gracefully

## Acceptance Criteria
1. All existing tests pass without modifications
2. New tests for the enhanced functionality pass
3. No regression in performance or functionality
4. Code review confirms thread safety and proper error handling
```

## 2. Test Utilities Implementation Details

### 2.1 File Utilities

**Description**: Create a comprehensive set of file utilities for testing in a dedicated `file.go` file.

**Implementation Details**:
- Create functions for file creation, verification, and state capture
- Ensure proper cleanup of created files and directories
- Implement assertion functions for file existence and content
- Ensure thread-safety with appropriate mutex usage

**Files to Create**:
- `internal/testutil/helpers/file.go`

**Key Functions to Implement**:
- `CreateTestFile` - Creates a file with the given content and ensures it's cleaned up after the test
- `CreateTestDir` - Creates a directory and ensures it's cleaned up after the test
- `CreateTempDir` - Creates a temporary directory and ensures it's cleaned up after the test
- `CreateTempFile` - Creates a temporary file with the given content and ensures it's cleaned up after the test
- `FileExists` - Checks if a file exists at the given path
- `FileContains` - Checks if a file contains the expected content
- `AssertFileExists` - Asserts that a file exists at the given path
- `AssertFileNotExists` - Asserts that a file does not exist at the given path
- `AssertFileContains` - Asserts that a file contains the expected content
- `CaptureFileSystemState` - Captures the current state of the filesystem by listing all files and directories

**Junie AI Prompt**:

```
I need to implement a comprehensive set of file utilities for testing in the OneMount project. Please create a new file at `internal/testutil/helpers/file.go` that includes the following functions:

1. `CreateTestFile(t *testing.T, dir, name string, content []byte) string` - Creates a file with the given content and ensures it's cleaned up after the test
2. `CreateTestDir(t *testing.T, parent, name string) string` - Creates a directory and ensures it's cleaned up after the test
3. `CreateTempDir(t *testing.T, prefix string) string` - Creates a temporary directory and ensures it's cleaned up after the test
4. `CreateTempFile(t *testing.T, dir, prefix string, content []byte) string` - Creates a temporary file with the given content and ensures it's cleaned up after the test
5. `FileExists(path string) bool` - Checks if a file exists at the given path
6. `FileContains(path string, expected []byte) (bool, error)` - Checks if a file contains the expected content
7. `AssertFileExists(t *testing.T, path string)` - Asserts that a file exists at the given path
8. `AssertFileNotExists(t *testing.T, path string)` - Asserts that a file does not exist at the given path
9. `AssertFileContains(t *testing.T, path string, expected []byte)` - Asserts that a file contains the expected content
10. `CaptureFileSystemState(dir string) (map[string]os.FileInfo, error)` - Captures the current state of the filesystem by listing all files and directories

Each function should be properly documented with godoc comments and include appropriate error handling. The functions should integrate with the TestFramework for automatic cleanup of resources.
```

### 2.2 Asynchronous Utilities

**Description**: Create a set of asynchronous utilities for testing in a dedicated `async.go` file.

**Implementation Details**:
- Create functions for waiting, retrying, and handling timeouts
- Implement condition-based waiting with configurable timeouts
- Add support for retrying operations with exponential backoff
- Ensure proper context handling and cancellation

**Files to Create**:
- `internal/testutil/helpers/async.go`

**Key Functions to Implement**:
- `WaitForCondition` - Waits for a condition to be true with a configurable timeout and polling interval
- `WaitForConditionWithContext` - Waits for a condition to be true with a context for cancellation
- `RetryWithBackoff` - Retries an operation with exponential backoff until it succeeds or times out
- `RunWithTimeout` - Runs an operation with a timeout
- `RunConcurrently` - Runs multiple operations concurrently and waits for all to complete
- `WaitForFileChange` - Waits for a file to change (by checking its modification time)
- `WaitForFileExistence` - Waits for a file to exist or not exist

**Junie AI Prompt**:

```
I need to implement asynchronous utilities for testing in the OneMount project. Please create a new file at `internal/testutil/helpers/async.go` that includes the following functions:

1. `WaitForCondition(t *testing.T, condition func() bool, timeout, interval time.Duration, message string) error` - Waits for a condition to be true with a configurable timeout and polling interval
2. `WaitForConditionWithContext(ctx context.Context, condition func() bool, interval time.Duration, message string) error` - Waits for a condition to be true with a context for cancellation
3. `RetryWithBackoff(t *testing.T, operation func() error, maxRetries int, initialDelay, maxDelay time.Duration, message string) error` - Retries an operation with exponential backoff until it succeeds or times out
4. `RunWithTimeout(t *testing.T, operation func() error, timeout time.Duration) error` - Runs an operation with a timeout
5. `RunConcurrently(t *testing.T, operations []func() error) []error` - Runs multiple operations concurrently and waits for all to complete
6. `WaitForFileChange(t *testing.T, path string, timeout, interval time.Duration) error` - Waits for a file to change (by checking its modification time)
7. `WaitForFileExistence(t *testing.T, path string, shouldExist bool, timeout, interval time.Duration) error` - Waits for a file to exist or not exist

Each function should be properly documented with godoc comments and include appropriate error handling. The functions should be designed to work with the TestFramework and use the context provided by it when appropriate.
```

### 2.3 Dmelfa Generator

**Description**: Create a utility for generating large test files with random DNA sequence data.

**Implementation Details**:
- Create functions for generating large test files with random DNA sequence data
- Implement a FASTA file format generator
- Add support for configurable file sizes
- Ensure proper error handling and logging

**Files to Create**:
- `internal/testutil/helpers/dmelfa_generator.go`

**Key Functions to Implement**:
- `GenerateDmelfa` - Generates a dmel.fa file with random DNA sequence data of the specified size
- `EnsureDmelfaExists` - Ensures that the dmel.fa file exists at the path specified in testutil.DmelfaDir

**Junie AI Prompt**:

```
I need to implement a utility for generating large test files with random DNA sequence data for the OneMount project. Please create a new file at `internal/testutil/helpers/dmelfa_generator.go` that includes the following functions:

1. `GenerateDmelfa(path string, size int64) error` - Generates a dmel.fa file with random DNA sequence data of the specified size if it doesn't exist
2. `EnsureDmelfaExists() error` - Ensures that the dmel.fa file exists at the path specified in `testutil.DmelfaDir` before tests run, generating it if necessary

The generator should create a file with a format similar to a FASTA file, with header lines starting with '>' followed by sequence identifier information, and sequence lines containing DNA sequences (A, C, G, T). The file should be large enough to test performance with large files (at least 100MB by default).

Each function should be properly documented with godoc comments and include appropriate error handling. The functions should be designed to work with the TestFramework and be integrated as a test resource.
```

### 2.4 Graph API Test Fixtures

**Description**: Extend the existing `mock_graph.go` file to include more comprehensive fixture creation utilities.

**Implementation Details**:
- Create functions for creating various types of DriveItem fixtures
- Implement utilities for creating nested folder structures
- Add support for creating items with specific properties
- Ensure proper integration with the existing MockGraphProvider

**Files to Modify**:
- `internal/testutil/mock/mock_graph.go`

**Key Functions to Implement**:
- `StandardTestFile` - Returns a standard test file content with predictable content
- `CreateDriveItemFixture` - Creates a DriveItem fixture for testing
- `CreateFileItemFixture` - Creates a DriveItem fixture representing a file
- `CreateFolderItemFixture` - Creates a DriveItem fixture representing a folder
- `CreateDeletedItemFixture` - Creates a DriveItem fixture representing a deleted item
- `CreateChildrenFixture` - Creates a slice of DriveItem fixtures representing children of a folder
- `CreateNestedFolderStructure` - Creates a nested folder structure for testing
- `CreateDriveItemWithConflict` - Creates a DriveItem fixture with conflict behavior set

**Junie AI Prompt**:

```
I need to enhance the Graph API test fixtures in the OneMount project. Please extend the existing file at `internal/testutil/mock/mock_graph.go` to include the following functions:

1. `StandardTestFile() []byte` - Returns a standard test file content with predictable content
2. `CreateDriveItemFixture(id, name string, size uint64, content []byte) *graph.DriveItem` - Creates a DriveItem fixture for testing
3. `CreateFileItemFixture(name string, size uint64, content []byte) *graph.DriveItem` - Creates a DriveItem fixture representing a file
4. `CreateFolderItemFixture(name string, childCount int) *graph.DriveItem` - Creates a DriveItem fixture representing a folder
5. `CreateDeletedItemFixture(name string) *graph.DriveItem` - Creates a DriveItem fixture representing a deleted item
6. `CreateChildrenFixture(parentID string, count int) []*graph.DriveItem` - Creates a slice of DriveItem fixtures representing children of a folder
7. `CreateNestedFolderStructure(parentID, baseName string, depth, width int) []*graph.DriveItem` - Creates a nested folder structure for testing
8. `CreateDriveItemWithConflict(name string, conflictBehavior string) *graph.DriveItem` - Creates a DriveItem fixture with conflict behavior set

Each function should be properly documented with godoc comments. The functions should be designed to work with the existing MockGraphProvider and be used to create consistent test data.
```

## 3. Advanced Framework Implementation Details

### 3.1 Specialized Framework Extensions

**Description**: Create specialized framework extensions for specific components like the Graph API and filesystem.

**Implementation Details**:
- Create specialized TestFramework extensions for different components
- Implement the specialized setup logic from the old TestMain functions
- Add support for component-specific configuration and utilities
- Ensure proper integration with the core TestFramework

**Files to Create**:
- `internal/testutil/framework/graph_framework.go`
- `internal/testutil/framework/fs_framework.go`

**Key Types and Functions to Implement**:
- `GraphTestFramework` - A specialized TestFramework for testing Graph API functionality
- `NewGraphTestFramework` - Creates a new GraphTestFramework
- `SetupGraphTest` - Sets up a graph test with the given configuration
- `FSTestFramework` - A specialized TestFramework for testing filesystem functionality
- `NewFSTestFramework` - Creates a new FSTestFramework
- `SetupFSTest` - Sets up a filesystem test with the given configuration

### 3.2 Environment Validation

**Description**: Add environment validation capabilities to the TestFramework to verify prerequisites before running tests.

**Implementation Details**:
- Create an `EnvironmentValidator` interface and `DefaultEnvironmentValidator` implementation
- Add methods for validating the test environment
- Implement checks for required tools, permissions, and configuration
- Ensure proper error handling and reporting

**Files to Modify**:
- `internal/testutil/framework/framework.go`
- `internal/testutil/framework/validator.go`

**Key Types and Functions to Implement**:
- `EnvironmentValidator` - An interface for validating the test environment
- `DefaultEnvironmentValidator` - The default implementation of EnvironmentValidator
- `NewDefaultEnvironmentValidator` - Creates a new DefaultEnvironmentValidator
- `Validate` - Validates the test environment
- `ValidateWithContext` - Validates the test environment with a context for cancellation

### 3.3 Enhanced Network Simulation

**Description**: Enhance the NetworkSimulator to support more realistic network scenarios.

**Implementation Details**:
- Implement methods for simulating intermittent connections and network partitions
- Add support for selective network rules applied to specific API endpoints
- Implement bandwidth throttling for realistic testing of large file transfers
- Simulate real-world network error patterns like intermittent failures and partial responses

**Files to Modify**:
- `internal/testutil/framework/network.go`

**Key Types and Functions to Implement**:
- `NetworkSimulator` - A simulator for network conditions
- `NewNetworkSimulator` - Creates a new NetworkSimulator
- `SetLatency` - Sets the latency for network operations
- `SetBandwidthLimit` - Sets the bandwidth limit for network operations
- `SimulateDisconnection` - Simulates a network disconnection
- `SimulateIntermittentConnection` - Simulates an intermittent network connection
- `SimulateNetworkPartition` - Simulates a network partition
- `ApplyToEndpoint` - Applies network conditions to a specific API endpoint

## 4. Project Review Recommendations

### 4.1 Architecture Recommendations

1. **Adopt a Standard Go Layout**
   - Introduce `internal/` for private packages and `pkg/` for public libraries
   - Align with community practices for Go project structure

2. **Refactor Core Functions**
   - Break down large `main.go` routines into discrete services (e.g., AuthService, FilesystemService)
   - Improve readability and testability of the codebase

3. **Implement Dependency Injection**
   - Define interfaces for external dependencies (Graph API, DB)
   - Inject implementations for easier mocking in tests

### 4.2 Testing Recommendations

1. **Improve Test Coverage**
   - Target ≥80% coverage by adding table-driven unit tests
   - Focus on filesystem operations, error conditions, and concurrency scenarios

2. **Use Context for Concurrency**
   - Replace raw goroutines with `context.Context` management and `sync.WaitGroup`
   - Handle cancellations and orderly shutdowns properly

3. **Standardize Error Handling**
   - Adopt a uniform error-wrapping strategy across modules
   - Leverage Go's `errors` package or a chosen wrapper for clarity and consistency

### 4.3 Documentation Recommendations

1. **Enhance Documentation**
   - Add a table of contents, contribution guidelines, and code-of-conduct to `README.md`
   - Provide an architecture overview in `docs/DEVELOPMENT.md`

2. **Create Test Framework Documentation**
   - Document the test framework architecture
   - Create API documentation for test framework components
   - Add examples of using the test framework

3. **Create Test Writing Guidelines**
   - Document best practices for writing tests
   - Create templates for different types of tests
   - Add examples of good test design

## 5. Conclusion

This document provides comprehensive details for implementing the testing framework and test cases for the OneMount project. By following these implementation details, the development team can create a robust testing framework that ensures the quality, reliability, and security of the OneMount system.

The implementation details are designed to be flexible and adaptable, allowing for adjustments as the project evolves. The Junie AI prompts provide guidance for implementing specific components, and the project review recommendations provide a roadmap for improving the overall architecture, testing, and documentation of the project.