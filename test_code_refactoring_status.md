# Test Code Refactoring Status

This document summarizes the status of the recommendations from the [test_code_refactoring.md](test_code_refactoring.md) document.

## Recommendations Status

### 1. Create Test Utilities Package (COMPLETED)

**Implementation**: Created and enhanced the `testutil` package with common test utilities:
- Added file operation utilities in `testutil/file.go` with functions for creating test files, directories, and checking file existence and content
- Added test fixtures in `testutil/fixtures.go` with functions for creating standard test data and fixtures for DriveItem and related types
- Added async operation utilities in `testutil/async.go` with functions for waiting for conditions, retrying operations with backoff, and handling timeouts

### 2. Standardize Test Patterns (IN PROGRESS)

**Implementation**: Update tests to use consistent patterns:
- Implement consistent use of `t.Parallel()` (PARTIALLY COMPLETED)
  - Added `t.Parallel()` to all appropriate tests in the fs/graph package
  - Added `t.Parallel()` to all appropriate tests in the cmd/common package
  - Identified tests that should not use `t.Parallel()` due to shared state (e.g., offline_test.go, fs_test.go)
  - Reviewed ui package tests and found that some tests already use `t.Parallel()` while others need to be updated
  - Reviewed fs/offline package tests and found that most tests already use `t.Parallel()` with comments explaining why some don't
- Implement consistent resource cleanup with `t.Cleanup()` (PARTIALLY COMPLETED)
  - Replaced all `defer` statements with `t.Cleanup()` in the fs/graph package tests
  - Reviewed fs package tests and found they already use `t.Cleanup()` for most resource cleanup
  - Added improved error handling in cleanup functions
  - Updated cmd/common package tests to use `t.Cleanup()` for resource cleanup
  - Identified that ui package tests need to be updated to use `t.Cleanup()` instead of defer
  - Reviewed fs/offline package tests and found they already use `t.Cleanup()` for resource cleanup in all test functions
  - Verified that defer is only used in the TestMain function where `t.Cleanup()` is not available
- Implement consistent assertion style using `require` and `assert` (PARTIALLY COMPLETED)
  - Updated all tests in the fs/graph package to use `require` for critical assertions and `assert` for non-critical assertions
  - Reviewed fs package tests and found they already use a mix of `require` and `assert` appropriately
  - Updated cmd/common package tests to use `require` for critical assertions and `assert` for non-critical assertions
  - Added clear error messages to all assertions
  - Identified that ui package tests need to be updated to use `require` for critical assertions
  - Identified that fs/offline package tests mostly use `require` but some still use t.Fatal/t.Error

### 3. Improve Test Reliability (IN PROGRESS)

**Implementation**: Make tests more reliable:
- Replace fixed timeouts with dynamic waiting (COMPLETED)
  - Added WaitForCondition utility in testutil/async.go to replace fixed timeouts with dynamic waiting
  - Updated TestUnitActive in ui/systemd/systemd_test.go to use WaitForCondition instead of fixed timeout
  - Updated setup_test.go in fs/offline to use WaitForCondition for mount point checks and other waiting operations
  - Added RetryWithBackoff utility in testutil/async.go for operations that need multiple attempts
  - Replaced all fixed sleeps in fs/fs_test.go with WaitForCondition:
    - Updated TestTouchUpdateTime to wait for file modification time to change
    - Updated TestMkdirRmdir to wait for directory removal
    - Updated TestNTFSIsABadFilesystem and its variants to wait for file operations
    - Updated TestEchoWritesToFile to wait for file content
    - Updated TestStat to wait for directory creation
    - Updated TestGIOTrash to wait for file creation and deletion
- Fix race conditions in tests (COMPLETED)
  - Fixed race conditions in TestUploadDiskSerialization by making the test more deterministic:
    - Used synchronous file copy to avoid race conditions
    - Added WaitForCondition to wait for the file to be recognized by the filesystem
    - Added WaitForCondition to wait for the upload session to be created
    - Added proper cleanup with t.Cleanup()
  - Improved TestRepeatedUploads to use dynamic waiting instead of fixed sleeps:
    - Added WaitForCondition to wait for the file to be uploaded
    - Used proper assertions to verify the upload was successful
  - Fixed potential race conditions in ui/systemd/systemd_test.go:
    - Added comments explaining why some subtests cannot use t.Parallel()
    - Used WaitForCondition to wait for unit state changes
    - Added proper cleanup to ensure resources are released
- Isolate tests from each other (PARTIALLY COMPLETED)
  - Converted TestChmod to TestFilePermissions using table-driven tests with subtests
  - Added unique filenames for each subtest to avoid conflicts
  - Added proper cleanup for each subtest
  - Converted TestMountpointIsValid in ui/onedriver_test.go to use table-driven tests with subtests
  - Converted TestHomeEscapeUnescape in ui/onedriver_test.go to use table-driven tests with subtests
  - Added nested subtests for testing related operations
  - Converted TestUnitEnabled and TestUnitActive in ui/systemd/systemd_test.go to use table-driven tests with subtests
  - Added proper cleanup using t.Cleanup() to ensure resources are released
  - Added comments explaining why some subtests cannot use t.Parallel()

### 4. Improve Error Handling (IN PROGRESS)

**Implementation**: Enhance error handling in tests:
- Add context to error messages (PARTIALLY COMPLETED)
  - Updated TestOfflineReaddir in fs/offline/offline_test.go to use require with descriptive error messages
  - Updated TestOfflineBagelDetection in fs/offline/offline_test.go to use require with descriptive error messages
  - Added better error reporting with context about what's being tested
  - Added proper error handling for previously ignored errors
- Test error conditions explicitly (PARTIALLY COMPLETED)
  - Updated TestGetAccountName in ui/onedriver_test.go to test error conditions:
    - Added test case for nonexistent auth_tokens.json file
    - Added test case for invalid JSON in auth_tokens.json file
    - Added test case for empty Account field in auth_tokens.json file
  - Converted TestGetAccountName to use table-driven tests with subtests
  - Added proper error handling and descriptive error messages

### 5. Improve Test Organization (IN PROGRESS)

**Implementation**: Better organize tests:
- Convert appropriate tests to table-driven tests (PARTIALLY COMPLETED)
  - Reviewed fs/graph package tests and found that many tests already use table-driven approach:
    - TestIDPath, TestChildrenPath, and TestChildrenPathID in path_test.go
    - TestSHA256Hash, TestSHA256HashStream, TestSHA1Hash, TestSHA1HashStream, TestQuickXORHash, and TestQuickXORHashStream in hash_functions_test.go
  - Reviewed ui package tests and found that several tests already use table-driven approach:
    - TestMountpointIsValid, TestHomeEscapeUnescape, and TestGetAccountName in ui/onedriver_test.go
  - Converted TestChmod to TestFilePermissions using table-driven tests with subtests in fs/fs_test.go
  - Converted delta tests in fs/delta_test.go to a table-driven approach:
    - Combined TestDeltaMkdir, TestDeltaRmdir, TestDeltaRename, and TestDeltaMoveParent into a single TestDeltaOperations test
    - Used descriptive test case names that clearly indicate what is being tested
    - Added proper setup and cleanup for each test case
    - Improved error handling with descriptive error messages
  - Converted file operation tests in fs/fs_test.go to a table-driven approach:
    - Combined TestReadWrite, TestAppend, and TestTruncate into a single TestFileOperations test
    - Used descriptive test case names that clearly indicate what is being tested
    - Added parallel execution for each test case
    - Created unique filenames for each test case to avoid conflicts
    - Added proper cleanup for each test case
    - Improved error handling with descriptive error messages
  - Converted TestReadWriteMidfile in fs/fs_test.go to a table-driven test:
    - Created TestPositionalFileOperations with multiple test cases for different scenarios
    - Added test cases for writing to the middle, beginning, end, and beyond the end of a file
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added parallel execution for each test case
    - Created unique filenames for each test case to avoid conflicts
    - Added proper cleanup for each test case
    - Improved error handling with descriptive error messages
  - Converted TestRenameMove and TestCopy in fs/fs_test.go to a table-driven test:
    - Combined them into a single TestFileMovementOperations test
    - Added test cases for renaming, moving to a subdirectory, and copying
    - Used a flexible structure with setupFunc, operationFunc, and verifyFunc
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added parallel execution for each test case
    - Created unique filenames for each test case to avoid conflicts
    - Added proper cleanup with special handling for different operation types
    - Improved error handling with descriptive error messages
  - Improved xattr operation tests in fs/xattr_operations_test.go:
    - Added descriptive test names using the format "Operation_ShouldExpectedResult"
    - Added proper cleanup for test resources
    - Made the test structure more consistent
    - Added better error messages with more context
    - Improved verification to provide more detailed information
  - Converted TestUntemplateUnit in ui/systemd/systemd_test.go to a table-driven test:
    - Added multiple test cases with descriptive names
    - Used parallel execution for each test case
    - Improved error handling with descriptive error messages
  - Converted TestTemplateUnit in ui/systemd/systemd_test.go to a table-driven test:
    - Added multiple test cases with descriptive names
    - Used parallel execution for each test case
    - Added test case for path with special characters
  - Converted TestNTFSIsABadFilesystem and its variants to a table-driven test:
    - Combined three related tests into a single TestCaseSensitivityHandling test
    - Added descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Created unique filenames for each test to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Added comments explaining why parallel execution is not used
  - Converted TestDisallowedFilenames to use subtests:
    - Added descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Used parallel execution for each test case
    - Created unique filenames for each test to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
  - Converted TestChildrenAreCasedProperly to TestFilenameCase with table-driven approach:
    - Added multiple test cases for different filename case patterns (uppercase, lowercase, mixed case)
    - Used parallel execution for each test case
    - Created unique filenames for each test to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
  - Converted TestEchoWritesToFile to TestShellFileOperations with table-driven approach:
    - Added multiple test cases for different shell commands and content types
    - Used parallel execution for each test case
    - Created unique filenames for each test to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
    - Added special handling for multiline content
  - Converted TestStat to TestFileInfo with table-driven approach:
    - Added multiple test cases for different file types (regular files, directories, executable files)
    - Created unique filenames for each test to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
    - Added comprehensive verification of file attributes
  - Converted TestNoQuestionMarks in fs/fs_test.go to a table-driven test:
    - Added multiple test cases for different directories and ls options
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Improved error handling with descriptive error messages
    - Added better error output for debugging
  - Converted TestGIOTrash in fs/fs_test.go to a table-driven test:
    - Added multiple test cases for different file types (regular text file, binary file, executable file, empty directory)
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Created unique filenames for each test case to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
  - Converted TestListChildrenPaging in fs/fs_test.go to a table-driven test:
    - Added multiple test cases for different directories (paging directory and root directory)
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Improved error handling with descriptive error messages
    - Added more detailed logging for debugging
    - Added comprehensive assertions to verify both API and filesystem results
  - Converted TestReaddir in fs/fs_test.go to a table-driven test:
    - Added multiple test cases for different directories (root directory, test directory, documents directory)
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added comprehensive checks, including verifying item types (file vs. directory)
    - Improved error handling with descriptive error messages
    - Added more detailed logging for debugging
  - Converted TestLs in fs/fs_test.go to a table-driven test:
    - Added multiple test cases for different directories and ls options
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added support for checking both expected and unexpected items
    - Improved error handling with descriptive error messages
    - Added more detailed logging for debugging
  - Converted TestLibreOfficeSavePattern in fs/fs_test.go to a table-driven test:
    - Added multiple test cases for different file types, content, and conversion formats
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Created unique filenames for each test case to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
    - Added more detailed logging for debugging
    - Added support for checking different expected file sizes
  - Combined TestTouchCreate and TestTouchUpdateTime in fs/fs_test.go into a single table-driven test TestTouchOperations:
    - Added test cases for creating empty files and updating modification times
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Created unique filenames for each test case to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
    - Added better verification of file properties
  - Converted TestRmdirNonempty in fs/fs_test.go to a table-driven test TestDirectoryRemoval:
    - Added test cases for removing empty directories and non-empty directories with different methods
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Created unique directory names for each test case to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
    - Added waiting for directory creation and removal with WaitForCondition
  - Converted TestMkdirRmdir in fs/fs_test.go to a table-driven test TestDirectoryOperations:
    - Added test cases for creating directories, removing directories, and the mkdir->rmdir->mkdir chain
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Created unique directory names for each test case to avoid conflicts
    - Added proper cleanup with t.Cleanup()
    - Improved error handling with descriptive error messages
    - Added waiting for directory creation and removal with WaitForCondition
    - Added a test case for creating directories with specific permissions
  - Combined TestStatFs and TestUnlink in fs/fs_test.go into a single table-driven test TestBasicFileSystemOperations:
    - Added test cases for filesystem statistics and file deletion operations
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added parallel execution for the unlink test case (but not for statfs since it operates on the shared TestDir)
    - Created unique filenames for the unlink test case to avoid conflicts
    - Added proper cleanup with t.Cleanup() for the unlink test case
    - Improved error handling with descriptive error messages
    - Added waiting for file creation and deletion with WaitForCondition
    - Added more detailed logging for debugging
  - Improved TestDisallowedFilenames in fs/fs_test.go:
    - Kept the existing table-driven structure for file/directory creation tests
    - Improved error checking by using require.Error() instead of manual if/else
    - Added more descriptive error messages
    - Enhanced the rename test with proper error handling
    - Maintained parallel execution, unique paths, and cleanup
  - Combined tests in fs/cache_test.go into a single table-driven test TestCacheOperations:
    - Merged TestRootGet, TestRootChildrenUpdate, TestSubdirGet, TestSubdirChildrenUpdate, and TestSamePointer
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Created unique database paths for each test case to avoid conflicts
    - Improved error handling with descriptive error messages
    - Added more detailed logging for debugging
    - Better organized with clear separation of setup and verification
  - Combined TestMode and TestIsDir in fs/inode_test.go into a single table-driven test TestInodeProperties:
    - Added test cases for both directories and files
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added parallel execution for each test case
    - Created separate setup and cleanup functions for each test case
    - Used t.Cleanup() for proper resource cleanup
    - Improved error handling with descriptive error messages
    - Added comprehensive assertions to verify both Mode() and IsDir() functionality
  - Refactored TestDBusServerStartStop in fs/dbus_test.go into a table-driven test TestDBusServerOperations:
    - Added test cases for checking initial state, stopping the server, starting the server, and stopping again
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added parallel execution at the top level (but not for subtests due to sequential nature)
    - Created operation functions for each test case
    - Used t.Cleanup() for proper resource cleanup
    - Improved error handling with descriptive error messages
    - Added comments explaining why parallel execution is not used in subtests
  - Combined TestThumbnailCache and TestThumbnailCacheMultipleSizes in fs/thumbnail_test.go into a single table-driven test TestThumbnailCacheOperations:
    - Added test cases for single thumbnail operations, multiple size operations, and individual deletion
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added parallel execution for both top level and subtests
    - Created separate setup and test functions for each test case
    - Used t.Cleanup() for proper resource cleanup
    - Improved error handling with descriptive error messages
    - Added a description field for each test case
  - Combined TestUploadSession, TestUploadSessionSmallFS, and TestUploadSessionLargeFS in fs/upload_session_test.go into a single table-driven test TestUploadSessionOperations:
    - Added test cases for direct uploads, small file uploads, and large file uploads
    - Used descriptive test case names following the "Operation_ShouldExpectedResult" pattern
    - Added parallel execution for each test case
    - Created separate setup and test functions for each test case
    - Used t.Cleanup() for proper resource cleanup
    - Improved error handling with descriptive error messages
    - Added a description field for each test case
    - Added a skipCheck function to conditionally skip tests (e.g., when large test file is not available)
    - Created unique filenames for each test to avoid conflicts
- Group related tests (PARTIALLY COMPLETED)
  - Grouped related tests in ui/systemd/systemd_test.go:
    - Standardized the structure of TestTemplateUnit and TestUntemplateUnit
    - Used consistent naming conventions for test cases
    - Organized test cases in a logical order
- Use clear test names (PARTIALLY COMPLETED)
  - Improved test names in fs/delta_test.go:
    - Used descriptive names like "CreateDirectoryOnServer_ShouldSyncToClient" instead of generic names
    - Added clear indication of what is being tested and what the expected outcome is
  - Improved test names in ui/systemd/systemd_test.go:
    - Used descriptive names like "StandardTemplate_ShouldCreateCorrectUnitName"
    - Added clear indication of what is being tested and what the expected outcome is

## Next Steps

1. Continue implementing test reliability improvements:
   - Isolate tests from each other by using subtests and proper cleanup:
     - Convert more tests to use table-driven tests with subtests where appropriate
     - Add parallel execution to subtests where possible
     - Ensure proper cleanup for all tests

2. Continue implementing test organization improvements:
   - Convert more appropriate tests to table-driven tests:
     - Focus on tests in fs package that test similar functionality with different inputs
     - Note: All tests in the ui package have already been refactored to use a table-driven approach
     - Potential candidates include:
       - Tests in fs/inode_test.go (TestConstructor, TestMode, etc.)
       - Tests in fs/dbus_test.go (TestDBusServerStartStop, etc.)
       - Tests in fs/thumbnail_test.go (TestThumbnailCache, etc.)
       - Tests in fs/upload_session_test.go (TestUploadSession, etc.)
       - Tests in fs/offline/offline_test.go (TestOfflineReaddir, etc.)
   - Group related tests:
     - Organize tests by functionality rather than by implementation details
     - Use clear naming conventions for test functions
   - Improve test naming conventions:
     - Use descriptive names that indicate what is being tested
     - Follow a consistent naming pattern across all tests

3. Review remaining tests for consistency:
   - Ensure all tests follow the established patterns and best practices
   - Check for any remaining fixed sleeps or timeouts that could be replaced with dynamic waiting
   - Verify that all tests have proper cleanup mechanisms
   - Consider adding more descriptive comments to explain test purpose and behavior
   - Focus on tests in the fs/graph and ui packages that haven't been reviewed yet

4. Document best practices for future test development (COMPLETED):
   - Created test_best_practices.md document outlining the patterns and practices established during this refactoring
   - Included examples of table-driven tests, proper cleanup, and error handling
   - Provided guidelines for when to use t.Parallel() and when not to
   - Added sections on test naming conventions, dynamic waiting, and test isolation

5. Consider additional improvements:
   - Add more comprehensive test coverage for error conditions
   - Improve test documentation with more detailed comments
   - Consider creating more test utilities for common operations
   - Review test performance and optimize where necessary
