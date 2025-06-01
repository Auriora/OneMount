# OneMount System Tests

This directory contains comprehensive system tests for OneMount that use a real OneDrive account to verify end-to-end functionality.

## Overview

The system tests provide comprehensive testing of OneMount's functionality using a real OneDrive account. These tests verify that all operations work correctly in a production-like environment.

## Test Categories

### Comprehensive Tests (`TestSystemST_COMPREHENSIVE_01_AllOperations`)
- Basic file operations (create, read, write, delete)
- Directory operations (create, list, delete)
- Large file handling (up to 50MB)
- Special character file names
- Concurrent operations
- File permissions
- Streaming operations

### Performance Tests (`TestSystemST_PERFORMANCE_01_UploadDownloadSpeed`)
- Upload/download speed measurements for various file sizes
- Performance benchmarking and reporting

### Reliability Tests (`TestSystemST_RELIABILITY_01_ErrorRecovery`)
- Error handling and recovery scenarios
- Invalid file name handling
- Authentication token refresh
- Disk space handling

### Integration Tests (`TestSystemST_INTEGRATION_01_MountUnmount`)
- Mount/unmount operations
- Filesystem persistence across mount cycles

### Stress Tests (`TestSystemST_STRESS_01_HighLoad`)
- High load scenarios with many concurrent operations
- System behavior under stress

## Prerequisites

### Authentication Setup
You need a real OneDrive account with valid authentication tokens at:
```
~/.onemount-tests/.auth_tokens.json
```

To set up authentication:
```bash
# Build OneMount
make onemount

# Authenticate with your test OneDrive account
./build/onemount --auth-only

# Copy tokens to test location
mkdir -p ~/.onemount-tests
cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
```

### System Requirements
- Go 1.19+
- FUSE support
- Network access to OneDrive
- At least 1GB free disk space
- Dedicated test OneDrive account (not production data)

## Running Tests

### Using Make (Recommended)
```bash
# Run comprehensive tests
make system-test-real

# Run all test categories
make system-test-all

# Run specific categories
make system-test-performance
make system-test-reliability
make system-test-integration
make system-test-stress
```

### Using the Script
```bash
# Run comprehensive tests
./scripts/run-system-tests.sh --comprehensive

# Run all tests
./scripts/run-system-tests.sh --all

# Run with custom timeout
./scripts/run-system-tests.sh --comprehensive --timeout 60m

# Run with verbose output
./scripts/run-system-tests.sh --comprehensive --verbose
```

### Using Go Test Directly
```bash
# Run comprehensive tests
go test -v -timeout 30m ./tests/system -run TestSystemST_COMPREHENSIVE_01_AllOperations

# Run performance tests
go test -v -timeout 30m ./tests/system -run TestSystemST_PERFORMANCE_01_UploadDownloadSpeed

# Run all system tests
go test -v -timeout 30m ./tests/system -run "TestSystemST_.*"
```

## Test Data

### OneDrive Location
Tests create files in: `/onemount_system_tests/`

### Local Artifacts
- Mount point: `~/.onemount-tests/tmp/system-test-mount/`
- Cache: `~/.onemount-tests/system-test-data/cache/`
- Logs: `~/.onemount-tests/logs/system_tests.log`

### Automatic Cleanup
All test data is automatically cleaned up after tests complete, even if tests fail.

## Test Structure

### SystemTestSuite
The main test suite that handles:
- Authentication with real OneDrive account
- Filesystem mounting using FUSE
- Test data management
- Cleanup operations

### Test Methods
Each test method focuses on a specific aspect:
- `TestBasicFileOperations()` - File CRUD operations
- `TestDirectoryOperations()` - Directory operations
- `TestLargeFileOperations()` - Large file handling
- `TestSpecialCharacterFiles()` - Special character handling
- `TestConcurrentOperations()` - Concurrent operations
- `TestFilePermissions()` - Permission handling
- `TestStreamingOperations()` - Streaming I/O
- `TestPerformance()` - Performance measurements
- `TestInvalidFileNames()` - Error handling
- `TestAuthenticationRefresh()` - Auth token refresh
- `TestDiskSpaceHandling()` - Disk space scenarios
- `TestMountUnmountCycle()` - Mount/unmount operations
- `TestHighLoadOperations()` - High load scenarios

## Troubleshooting

### Common Issues

#### Authentication Errors
```
Error: Authentication tokens not found
```
**Solution**: Set up authentication tokens as described in Prerequisites

#### Mount Errors
```
Error: Failed to create FUSE server
```
**Solution**: 
- Check FUSE is installed: `sudo apt install fuse3` (Ubuntu/Debian)
- Ensure mount point is not in use
- Check permissions

#### Network Errors
```
Error: Failed to upload/download files
```
**Solution**:
- Check internet connection
- Verify OneDrive account access
- Check for rate limiting

### Debug Mode
Enable debug logging:
```bash
export ONEMOUNT_LOG_LEVEL=debug
go test -v ./tests/system -run TestSystemST_COMPREHENSIVE_01_AllOperations
```

### Manual Cleanup
If tests fail to clean up:
```bash
# Unmount if still mounted
fusermount3 -uz ~/.onemount-tests/tmp/system-test-mount

# Remove test artifacts
rm -rf ~/.onemount-tests/tmp/system-test-mount/
rm -rf ~/.onemount-tests/system-test-data/
```

## Contributing

When adding new system tests:
1. Follow the naming convention: `TestSystemST_CATEGORY_##_Description`
2. Add proper cleanup in the test suite
3. Handle network/auth errors gracefully
4. Document expected behavior
5. Update this README if adding new test categories

## Security

- **Test Account Only**: Never use production OneDrive accounts
- **Token Security**: Auth tokens are stored with restricted permissions
- **Data Isolation**: Test data is isolated in dedicated directories
- **Automatic Cleanup**: All test data is removed after tests
