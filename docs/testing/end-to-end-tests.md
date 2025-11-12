# End-to-End Workflow Tests

## Overview

This document describes the end-to-end (E2E) workflow tests for OneMount. These tests verify complete user workflows from authentication through file operations, ensuring the system works correctly in real-world scenarios.

## Test Files

- **Location**: `internal/fs/end_to_end_workflow_test.go`
- **Helper Functions**: `internal/testutil/helpers/e2e_helpers.go`

## Test Cases

### E2E-17-01: Complete User Workflow

**Purpose**: Verify the complete user workflow from authentication to file operations and state persistence.

**Test Steps**:
1. Authenticate with Microsoft account
2. Mount OneDrive filesystem
3. Create new files
4. Modify existing files
5. Delete files
6. Verify changes sync to OneDrive
7. Unmount filesystem
8. Remount filesystem
9. Verify state is preserved

**Requirements**: All requirements

### E2E-17-02: Multi-File Operations

**Purpose**: Verify copying entire directories with multiple files to/from OneDrive.

**Test Steps**:
1. Create directory with multiple files locally
2. Copy directory to OneDrive mount point
3. Verify all files upload correctly
4. Copy directory from OneDrive to local
5. Verify all files download correctly

**Requirements**: 3.2, 4.3, 10.1, 10.2

### E2E-17-03: Long-Running Operations

**Purpose**: Verify large file upload (1GB+) with progress monitoring.

**Test Steps**:
1. Create a very large file (1GB)
2. Start upload to OneDrive
3. Monitor upload progress
4. Verify upload completes successfully
5. Test interruption and resume (optional)

**Requirements**: 4.3, 4.4

**Note**: This test takes 20+ minutes to complete.

### E2E-17-04: Stress Scenarios

**Purpose**: Verify system stability under many concurrent operations.

**Test Steps**:
1. Perform many concurrent file operations (20 workers × 50 operations each)
2. Monitor resource usage (CPU, memory, goroutines)
3. Verify system remains stable
4. Check for memory leaks

**Requirements**: 10.1, 10.2

## Running the Tests

### Prerequisites

1. **Authentication**: Valid OneDrive credentials stored in auth tokens file
2. **Environment**: Docker test environment with FUSE support
3. **Network**: Active internet connection to OneDrive

### Environment Variables

- `RUN_E2E_TESTS=1` - Enable end-to-end tests (required for all E2E tests)
- `RUN_LONG_TESTS=1` - Enable long-running tests (E2E-17-03)
- `RUN_STRESS_TESTS=1` - Enable stress tests (E2E-17-04)
- `ONEMOUNT_AUTH_PATH` - Path to auth tokens file (default: `test-artifacts/.auth_tokens.json`)

### Running in Docker

The recommended way to run E2E tests is in Docker containers:

```bash
# Run all E2E tests (except long-running and stress tests)
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  system-tests go test -v -run TestE2E ./internal/fs

# Run specific E2E test
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  system-tests go test -v -run TestE2E_17_01 ./internal/fs

# Run long-running tests (1GB file upload)
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  -e RUN_LONG_TESTS=1 \
  system-tests go test -v -run TestE2E_17_03 ./internal/fs

# Run stress tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  -e RUN_STRESS_TESTS=1 \
  system-tests go test -v -run TestE2E_17_04 ./internal/fs

# Run all E2E tests including long-running and stress tests
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -e RUN_E2E_TESTS=1 \
  -e RUN_LONG_TESTS=1 \
  -e RUN_STRESS_TESTS=1 \
  system-tests go test -v -timeout 60m -run TestE2E ./internal/fs
```

### Running Locally (Not Recommended)

If you must run tests locally (not recommended due to potential system impact):

```bash
# Ensure you have valid auth tokens
export ONEMOUNT_AUTH_PATH="test-artifacts/.auth_tokens.json"

# Run basic E2E tests
RUN_E2E_TESTS=1 go test -v -run TestE2E_17_01 ./internal/fs
RUN_E2E_TESTS=1 go test -v -run TestE2E_17_02 ./internal/fs

# Run long-running test (takes 20+ minutes)
RUN_E2E_TESTS=1 RUN_LONG_TESTS=1 go test -v -timeout 60m -run TestE2E_17_03 ./internal/fs

# Run stress test
RUN_E2E_TESTS=1 RUN_STRESS_TESTS=1 go test -v -timeout 30m -run TestE2E_17_04 ./internal/fs
```

## Test Artifacts

Test artifacts are stored in `test-artifacts/`:

- `test-artifacts/.auth_tokens.json` - Authentication tokens (gitignored)
- `test-artifacts/logs/` - Test execution logs
- `test-artifacts/.onemount-tests/` - Test cache and metadata

## Expected Results

### Success Criteria

- **E2E-17-01**: All file operations complete successfully, state persists across mount/unmount
- **E2E-17-02**: All files in directory are copied correctly in both directions
- **E2E-17-03**: Large file uploads successfully (may take 20+ minutes)
- **E2E-17-04**: Success rate > 90%, no memory leaks, goroutine count < 100

### Common Issues

1. **Authentication Failures**
   - Ensure auth tokens are valid and not expired
   - Re-authenticate if necessary: `onemount --auth-only`

2. **Mount Failures**
   - Check if mount point is already in use
   - Verify FUSE device is accessible: `ls -l /dev/fuse`
   - Ensure proper capabilities in Docker

3. **Network Issues**
   - Verify internet connectivity
   - Check OneDrive service status
   - Review rate limiting in logs

4. **Timeout Issues**
   - Increase test timeout for slow connections
   - Use `-timeout` flag: `go test -timeout 60m`

## Helper Functions

The following helper functions are available in `internal/testutil/helpers/e2e_helpers.go`:

- `GenerateRandomString(length int) string` - Generate random test data
- `CopyDirectory(src, dst string) error` - Recursively copy directories
- `CopyFile(src, dst string) error` - Copy single file
- `GetFileStatus(filePath string) (string, error)` - Get file sync status from extended attributes
- `SetFileStatus(filePath, status string) error` - Set file sync status
- `GetFileETag(filePath string) (string, error)` - Get file ETag from extended attributes
- `WaitForFileStatus(filePath, expectedStatus string, timeout, checkInterval time.Duration) error` - Wait for file to reach specific status

## Integration with CI/CD

These tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run E2E Tests
  run: |
    docker compose -f docker/compose/docker-compose.test.yml run --rm \
      -e RUN_E2E_TESTS=1 \
      -e ONEMOUNT_AUTH_PATH=/secrets/auth_tokens.json \
      system-tests go test -v -timeout 30m -run TestE2E_17_01 ./internal/fs
```

**Note**: Long-running and stress tests should be run separately or on a schedule due to their duration.

## Troubleshooting

### Test Hangs During Mount

If a test hangs during filesystem mount:

1. Check FUSE device availability
2. Verify mount point is not already in use
3. Check for kernel FUSE module: `lsmod | grep fuse`
4. Review FUSE debug logs if enabled

### File Operations Fail

If file operations fail during tests:

1. Check filesystem is properly mounted
2. Verify network connectivity to OneDrive
3. Check authentication tokens are valid
4. Review error logs in `test-artifacts/logs/`

### Memory or Resource Issues

If tests fail due to resource constraints:

1. Increase Docker container resources
2. Reduce number of concurrent workers in stress test
3. Monitor system resources during test execution
4. Check for memory leaks using profiling tools

## Future Enhancements

Potential improvements for E2E tests:

1. **Automated Verification**: Verify files actually exist on OneDrive via API
2. **Interruption Testing**: Implement safe interruption and resume testing
3. **Performance Benchmarks**: Add performance metrics collection
4. **Multi-Account Testing**: Test with multiple OneDrive accounts simultaneously
5. **Offline Mode Testing**: Test offline/online transitions
6. **Conflict Resolution**: Test conflict scenarios with concurrent modifications

## References

- [Test Setup Guide](../TEST_SETUP.md)
- [Docker Test Environment](docker-test-environment.md)
- [System Tests Guide](system-tests-guide.md)
- [Verification Tracking](../verification-tracking.md)
