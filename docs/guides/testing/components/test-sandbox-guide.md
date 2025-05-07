`# Test Sandbox Guidelines for OneMount

## Overview

This document provides guidelines for using the test-sandbox directory in the OneMount project. It outlines best practices for test working folders, including proper usage, cleanup procedures, isolation between tests, naming conventions, and resource management.

## Current Structure and Usage

The test-sandbox directory is currently used as the main test working folder for the OneMount project. It contains various test artifacts, including:

1. **Log Files**: `fusefs_tests.log` - Contains logs from test runs
2. **Mount Point**: `tmp/mount` - Where the filesystem is mounted during tests
3. **Test Files**: `dmel.fa` - A large test file used for upload session tests
4. **Authentication Tokens**: `.auth_tokens.json` - Contains authentication tokens for tests
5. **Test Database**: `tmp` - Contains the test database and other temporary files
6. **Content and Thumbnails**: `tmp/test/content` and `tmp/test/thumbnails` - Directories for test content and thumbnails

The test-sandbox directory is defined in `internal/testutil/test_constants.go` with the following structure:

```
test-sandbox/                  (TestSandboxDir)
├── .auth_tokens.json          (AuthTokensPath)
├── dmel.fa                    (DmelfaDir)
├── fusefs_tests.log           (TestLogPath)
└── tmp/                       (TestSandboxTmpDir)
    ├── test/
    │   ├── content/
    │   └── thumbnails/
    └── mount/                 (TestMountPoint)
        └── onemount_tests/    (TestDir)
            └── delta/         (DeltaDir)
```

## Recommended Structure Outside the Project

To improve test isolation and avoid cluttering the project directory, we recommend moving the test-sandbox directory outside of the project. This can be achieved by:

1. Creating a dedicated directory for test artifacts outside the project directory
2. Updating the constants in `internal/testutil/test_constants.go` to use this external directory
3. Ensuring all tests use the constants from `testutil` rather than hardcoded paths

The recommended structure is:

```
$HOME/.onemount-tests/                  (TestSandboxDir)
├── .auth_tokens.json                   (AuthTokensPath)
├── dmel.fa                             (DmelfaDir)
├── logs/
│   └── fusefs_tests.log                (TestLogPath)
├── tmp/                                (TestSandboxTmpDir)
│   ├── test/
│   │   ├── content/
│   │   └── thumbnails/
│   └── mount/                          (TestMountPoint)
│       └── onemount_tests/             (TestDir)
│           └── delta/                  (DeltaDir)
└── graph_test_dir/                     (New directory for graph tests)
```

## Best Practices for Test Working Folders

### Proper Usage in Tests

1. **Use Constants**: Always use the constants defined in `internal/testutil/test_constants.go` rather than hardcoded paths.
2. **Avoid Direct Manipulation**: Do not directly manipulate the test-sandbox directory in tests. Use the provided utility functions in `internal/testutil/setup.go`.
3. **Respect Directory Structure**: Maintain the directory structure defined in the constants. Do not create additional directories or files in the test-sandbox directory unless necessary.
4. **Test Isolation**: Each test should operate in its own subdirectory to avoid conflicts with other tests.
5. **Resource Limits**: Be mindful of resource usage, especially when creating large files or many small files.

### Cleanup Procedures

1. **Clean Up After Tests**: Always clean up any files or directories created during tests.
2. **Use t.Cleanup()**: Use the `t.Cleanup()` function to register cleanup functions that will be called even if tests fail.
3. **Temporary Files**: Store temporary files in the `tmp` directory, which is cleaned up between test runs.
4. **Persistent Files**: Store files that need to persist between test runs (e.g., authentication tokens) in the root of the test-sandbox directory.
5. **Unmount Before Cleanup**: Always unmount the filesystem before attempting to clean up the mount point.

### Isolation Between Tests

1. **Unique Test Directories**: Each test should use a unique directory to avoid conflicts with other tests.
2. **Parallel Tests**: When running tests in parallel, ensure they do not share resources.
3. **Clean State**: Start each test with a clean state by removing and recreating test directories.
4. **Independent Tests**: Tests should not depend on the state created by other tests.
5. **Mock Dependencies**: Use mock implementations of external dependencies to improve isolation.

### Naming Conventions

1. **Descriptive Names**: Use descriptive names for test files and directories.
2. **Test-Specific Prefixes**: Prefix test files and directories with the test name to avoid conflicts.
3. **Temporary File Suffix**: Use a `.tmp` suffix for temporary files.
4. **Test Data Files**: Store test data files in a `testdata` directory.
5. **Log Files**: Store log files in a `logs` directory with descriptive names.

### Resource Management

1. **Limit File Sizes**: Keep test files as small as possible while still being useful for testing.
2. **Clean Up Resources**: Always clean up resources after tests, especially large files.
3. **Reuse Test Files**: Reuse test files when possible instead of creating new ones.
4. **Monitor Resource Usage**: Use the profiler to monitor resource usage during tests.
5. **Limit Concurrent Operations**: Use semaphores to limit concurrent operations and prevent resource exhaustion.

## Specific Recommendations for Test Artifacts

### fusefs_tests.log

- Move to `$HOME/.onemount-tests/logs/fusefs_tests.log`
- Implement log rotation to prevent the log file from growing too large
- Add timestamps to log entries for better debugging

### mount-point

- Move to `$HOME/.onemount-tests/tmp/mount`
- Ensure it's unmounted and cleaned up after tests
- Use a unique mount point for each test run to avoid conflicts

### dmel.fa

- Move to `$HOME/.onemount-tests/dmel.fa`
- Consider generating this file on demand instead of storing it
- Implement a mechanism to verify the file's integrity before using it

### graph_test_dir

- Create a new directory at `$HOME/.onemount-tests/graph_test_dir`
- Use this directory for graph API tests
- Implement proper cleanup procedures for this directory

### test/

- Move to `$HOME/.onemount-tests/tmp/test`
- Ensure it's cleaned up between test runs
- Use subdirectories for different types of tests (e.g., content, thumbnails)

## Implementation Considerations

When implementing these recommendations, consider the following:

1. **Backward Compatibility**: Ensure that existing tests continue to work with the new structure.
2. **Environment Variables**: Use environment variables to allow overriding the test-sandbox location.
3. **Documentation**: Update documentation to reflect the new structure and best practices.
4. **CI/CD Integration**: Ensure that CI/CD pipelines are updated to use the new structure.
5. **Test Helpers**: Create helper functions to simplify working with the new structure.

## Conclusion

Moving the test-sandbox directory outside of the project will improve test isolation, reduce clutter in the project directory, and make it easier to manage test artifacts. By following the best practices outlined in this document, we can ensure that tests are reliable, maintainable, and efficient.