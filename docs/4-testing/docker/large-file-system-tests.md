# Large File System Tests

## Overview

This document describes the large file system tests that have been added to OneMount to test files larger than 2.5GB. These tests were created in response to crashes that occurred when copying large files, specifically a 2.5GB file that would pause at 2.4GB and then crash.

## Test Coverage Gap

Prior to these tests, OneMount's system tests only covered files up to:
- **10MB** in `TestLargeFileOperations()`
- **50MB** in `TestDiskSpaceHandling()`
- **10MB** in performance tests

This left a significant gap for multi-gigabyte files that are common in real-world usage.

## New Test Cases

### ST-LARGE-FILES-01: Multi-Gigabyte File Operations Test

**Test ID**: `TestSystemST_LARGE_FILES_01_MultiGigabyteFiles`

**Purpose**: Test system behavior with files larger than 2.5GB

**File Sizes Tested**:
- **1GB** - Baseline large file test (15-minute timeout)
- **2.5GB** - The specific size that was causing crashes (30-minute timeout)
- **5GB** - Stress test for very large files (60-minute timeout)

**Operations Tested**:
1. **File Creation**: Create large files using chunked writes (64MB chunks)
2. **Upload Operations**: Test upload to OneDrive with progress monitoring
3. **Download Operations**: Test download from OneDrive
4. **Integrity Verification**: Verify file content integrity using deterministic patterns
5. **Performance Monitoring**: Measure upload/download speeds

**Key Features**:
- **Memory Efficient**: Uses 64MB chunks to avoid memory exhaustion
- **Progress Monitoring**: Logs progress every 500MB during operations
- **Integrity Checking**: Verifies every byte using deterministic patterns
- **Performance Metrics**: Reports upload/download speeds in MB/s
- **Timeout Handling**: Different timeouts based on file size

### ST-LARGE-FILES-02: Streaming Large File Operations Test

**Test ID**: `TestSystemST_LARGE_FILES_02_StreamingLargeFiles`

**Purpose**: Test streaming read/write operations with large files

**Operations Tested**:
1. **Streaming Writes**: Write large files in 1MB chunks with delays
2. **Streaming Reads**: Read large files in chunks
3. **Partial Operations**: Test seek and partial read/write operations
4. **Real-time Monitoring**: Monitor operations as they progress

**Key Features**:
- **Streaming Simulation**: Uses smaller chunks (1MB) with delays
- **Real-time Progress**: Logs progress every 100MB
- **Random Data**: Uses random data to test different patterns
- **Upload Verification**: Ensures streaming uploads complete correctly

## Running the Tests

### Prerequisites

1. **Disk Space**: At least 10GB free disk space
2. **Time**: Tests can take 1+ hours to complete
3. **Network**: Stable internet connection for OneDrive operations
4. **Authentication**: Valid OneDrive test account credentials

### Using Make (Recommended)

```bash
# Run large file tests with confirmation prompt
make system-test-large-files
```

This will:
- Show a warning about disk space and time requirements
- Prompt for confirmation before proceeding
- Run all large file tests with a 2-hour timeout

### Using Go Test Directly

```bash
# Run all large file tests
go test -v -timeout 2h ./tests/system -run "TestSystemST_LARGE_FILES_.*"

# Run specific test
go test -v -timeout 2h ./tests/system -run "TestSystemST_LARGE_FILES_01_MultiGigabyteFiles"
```

### Test Output Example

```
=== RUN   TestSystemST_LARGE_FILES_01_MultiGigabyteFiles
=== RUN   TestSystemST_LARGE_FILES_01_MultiGigabyteFiles/LargeFile_1GB
    large_file_system_test.go:XXX: Creating 1GB file (1073741824 bytes)
    large_file_system_test.go:XXX: Written 500 MB / 1024 MB (48.8%) in 2m15s
    large_file_system_test.go:XXX: Large file creation completed in 4m30s
    large_file_system_test.go:XXX: Waiting up to 15m0s for upload to complete
    large_file_system_test.go:XXX: Upload completed in 8m45s
    large_file_system_test.go:XXX: Starting integrity verification for 1GB file
    large_file_system_test.go:XXX: Verified 500 MB / 1024 MB (48.8%)
    large_file_system_test.go:XXX: Integrity verification completed in 3m20s
    large_file_system_test.go:XXX: Performance metrics for 1GB:
    large_file_system_test.go:XXX:   Upload speed: 2.05 MB/s
    large_file_system_test.go:XXX:   Read speed: 5.12 MB/s
    large_file_system_test.go:XXX: Cleaning up 1GB file
--- PASS: TestSystemST_LARGE_FILES_01_MultiGigabyteFiles/LargeFile_1GB (16m35s)
```

## Test Implementation Details

### Memory Management

- **Chunked Operations**: All operations use chunks (64MB for creation, 1MB for streaming)
- **No Full File Loading**: Never loads entire files into memory
- **Progress Monitoring**: Regular progress updates prevent timeout issues

### Error Handling

- **Timeout Protection**: Different timeouts based on file size
- **Cancellation Support**: Tests can be cancelled gracefully
- **Cleanup Guarantee**: Files are always cleaned up, even on failure
- **Detailed Error Messages**: Clear error messages with context

### Performance Monitoring

- **Upload Speed**: Measures actual upload performance to OneDrive
- **Download Speed**: Measures read performance from mounted filesystem
- **Progress Tracking**: Real-time progress updates during operations
- **Time Breakdown**: Separate timing for creation, upload, and verification

## Integration with Existing Tests

### Test Categories

The large file tests are integrated into the existing system test framework:

- **ST-LARGE-FILES-01**: Multi-gigabyte file operations
- **ST-LARGE-FILES-02**: Streaming large file operations

### Test Environment

- Uses the same `SystemTestSuite` as other system tests
- Shares authentication and mount point setup
- Follows the same cleanup and error handling patterns

### CI/CD Considerations

These tests are **not included** in regular CI/CD pipelines due to:
- **Time Requirements**: Can take 1+ hours to complete
- **Disk Space**: Requires significant temporary disk space
- **Network Usage**: Uploads/downloads multiple gigabytes of data

Instead, they should be run:
- **Before Releases**: As part of release validation
- **After Major Changes**: When filesystem code is modified
- **Issue Investigation**: When investigating large file problems

## Troubleshooting

### Common Issues

1. **Disk Space**: Ensure at least 10GB free space
2. **Network Timeouts**: Use stable internet connection
3. **Authentication**: Verify OneDrive test credentials
4. **Memory**: Monitor system memory during tests

### Performance Expectations

- **Upload Speed**: 1-5 MB/s (depends on network)
- **Download Speed**: 5-20 MB/s (depends on disk)
- **Total Time**: 30-90 minutes for all tests

### Debugging

Enable verbose logging to see detailed progress:
```bash
go test -v -timeout 2h ./tests/system -run "TestSystemST_LARGE_FILES_.*" -args -test.v
```

## Future Enhancements

1. **Parallel Operations**: Test multiple large files simultaneously
2. **Network Interruption**: Test behavior during network failures
3. **Disk Full Scenarios**: Test behavior when disk space runs out
4. **Very Large Files**: Test files larger than 10GB
5. **Performance Benchmarking**: Establish performance baselines
