# OneMount System Tests Guide

This guide explains how to run comprehensive system tests for OneMount using a real OneDrive account.

## Overview

The OneMount system tests provide comprehensive end-to-end testing using a real OneDrive account to verify all functionality works correctly in a production-like environment. These tests cover:

- **Basic Operations**: File creation, reading, writing, deletion
- **Directory Operations**: Directory creation, listing, deletion
- **Large File Handling**: Upload/download of large files (up to 50MB)
- **Special Characters**: Files with special characters in names
- **Concurrent Operations**: Multiple simultaneous file operations
- **Performance Testing**: Upload/download speed measurements
- **Error Recovery**: Handling of various error conditions
- **Mount/Unmount**: Filesystem mount and unmount operations
- **High Load**: System behavior under stress conditions

## Prerequisites

### 1. Authentication Setup

You need a real OneDrive account with valid authentication tokens. The system tests will look for authentication tokens at:

```
~/.onemount-tests/.auth_tokens.json
```

To set up authentication:

1. **Option 1: Use existing OneMount installation**
   ```bash
   # Authenticate with your test OneDrive account
   onemount --auth-only
   
   # Copy the auth tokens to the test location
   mkdir -p ~/.onemount-tests
   cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
   ```

2. **Option 2: Authenticate directly for tests**
   ```bash
   # Build OneMount
   make onemount
   
   # Authenticate and save tokens to test location
   mkdir -p ~/.onemount-tests
   ./build/onemount --auth-only --config-file /dev/null
   # Follow the authentication flow
   # Tokens will be saved to the test location
   ```

### 2. OneDrive Account Requirements

- **Dedicated Test Account**: Use a dedicated OneDrive account for testing (not your production account)
- **Storage Space**: Ensure the account has at least 500MB of free space
- **Network Access**: Stable internet connection required

### 3. System Requirements

- **Go 1.19+**: Required for running tests
- **FUSE Support**: Required for filesystem mounting
- **Sufficient Disk Space**: At least 1GB free space for test artifacts

## Running System Tests

### Quick Start

Run the comprehensive system test suite:

```bash
# Using Make
make system-test-real

# Or directly using the script
./scripts/run-system-tests.sh --comprehensive
```

### Test Categories

#### 1. Comprehensive Tests (Default)
Tests all basic functionality:
```bash
make system-test-real
# or
./scripts/run-system-tests.sh --comprehensive
```

#### 2. Performance Tests
Measures upload/download speeds for various file sizes:
```bash
make system-test-performance
# or
./scripts/run-system-tests.sh --performance
```

#### 3. Reliability Tests
Tests error handling and recovery:
```bash
make system-test-reliability
# or
./scripts/run-system-tests.sh --reliability
```

#### 4. Integration Tests
Tests mount/unmount operations:
```bash
make system-test-integration
# or
./scripts/run-system-tests.sh --integration
```

#### 5. Stress Tests
Tests system behavior under high load:
```bash
make system-test-stress
# or
./scripts/run-system-tests.sh --stress
```

#### 6. All Tests
Run all test categories:
```bash
make system-test-all
# or
./scripts/run-system-tests.sh --all
```

### Advanced Usage

#### Custom Timeout
Set a custom timeout for long-running tests:
```bash
./scripts/run-system-tests.sh --comprehensive --timeout 60m
```

#### Verbose Output
Enable verbose logging:
```bash
./scripts/run-system-tests.sh --comprehensive --verbose
```

#### Direct Go Test Execution
Run specific tests directly with Go:
```bash
# Run comprehensive tests
go test -v -timeout 30m ./pkg/testutil -run TestSystemST_COMPREHENSIVE_01_AllOperations

# Run performance tests
go test -v -timeout 30m ./pkg/testutil -run TestSystemST_PERFORMANCE_01_UploadDownloadSpeed

# Run all system tests
go test -v -timeout 30m ./pkg/testutil -run "TestSystemST_.*"
```

## Test Data and Cleanup

### Test Data Location

System tests create files in the following locations:

- **OneDrive**: `/onemount_system_tests/` directory
- **Local Mount**: `~/.onemount-tests/tmp/system-test-mount/`
- **Test Artifacts**: `~/.onemount-tests/system-test-data/`
- **Logs**: `~/.onemount-tests/logs/system_tests.log`

### Automatic Cleanup

The system tests automatically clean up test data:

1. **During Tests**: Each test cleans up its own files
2. **After Tests**: The test suite removes the test directory
3. **On Failure**: Cleanup runs even if tests fail

### Manual Cleanup

If you need to manually clean up test data:

```bash
# Remove local test artifacts
rm -rf ~/.onemount-tests/tmp/system-test-mount/
rm -rf ~/.onemount-tests/system-test-data/

# Remove OneDrive test directory (if mounted)
rm -rf /path/to/mount/onemount_system_tests/
```

## Understanding Test Results

### Success Indicators

- All tests pass without errors
- Performance metrics are within reasonable ranges
- No resource leaks or crashes
- Clean test data cleanup

### Common Issues

#### Authentication Errors
```
Error: Authentication tokens not found
```
**Solution**: Set up authentication tokens as described in Prerequisites

#### Mount Errors
```
Error: Failed to mount filesystem
```
**Solution**: 
- Check FUSE is installed and working
- Ensure mount point is not already in use
- Verify sufficient permissions

#### Network Errors
```
Error: Failed to upload/download files
```
**Solution**:
- Check internet connection
- Verify OneDrive account is accessible
- Check for rate limiting

#### Timeout Errors
```
Error: Test timeout exceeded
```
**Solution**:
- Increase timeout with `--timeout` option
- Check network speed
- Verify OneDrive account performance

### Performance Expectations

Typical performance ranges (may vary based on network and OneDrive performance):

- **Small files (1KB-100KB)**: 10-100 KB/s
- **Medium files (1MB)**: 100KB/s - 1MB/s  
- **Large files (10MB+)**: 500KB/s - 5MB/s

## Troubleshooting

### Debug Mode

Enable debug logging by setting the log level:
```bash
export ONEMOUNT_LOG_LEVEL=debug
./scripts/run-system-tests.sh --comprehensive --verbose
```

### Check Logs

Review detailed logs:
```bash
# System test logs
tail -f ~/.onemount-tests/logs/system_tests.log

# OneMount logs (if running)
tail -f ~/.var/log/onemount.log
```

### Verify Authentication

Test authentication manually:
```bash
# Check if auth tokens are valid
python3 -c "
import json
import time
data = json.load(open('$HOME/.onemount-tests/.auth_tokens.json'))
expires_at = data.get('expires_at', 0)
now = int(time.time())
print(f'Token expires at: {expires_at}')
print(f'Current time: {now}')
print(f'Token valid: {expires_at > now}')
print(f'Account: {data.get(\"account\", \"N/A\")}')
"
```

### Reset Test Environment

If tests are consistently failing, reset the test environment:
```bash
# Stop any running OneMount instances
pkill -f onemount

# Clean up test directories
rm -rf ~/.onemount-tests/tmp/
rm -rf ~/.onemount-tests/system-test-data/

# Re-run authentication
./build/onemount --auth-only
```

## Contributing

When adding new system tests:

1. **Follow Naming Convention**: Use `TestSystemST_CATEGORY_##_Description` format
2. **Add Cleanup**: Ensure all test data is cleaned up
3. **Handle Errors**: Test should handle network/auth errors gracefully
4. **Document**: Add test description and expected behavior
5. **Update Scripts**: Add new test categories to the run script if needed

## Security Considerations

- **Test Account Only**: Never use production OneDrive accounts for testing
- **Token Security**: Auth tokens are stored with restricted permissions (0600)
- **Data Isolation**: Test data is isolated in dedicated directories
- **Cleanup**: All test data is automatically removed after tests
