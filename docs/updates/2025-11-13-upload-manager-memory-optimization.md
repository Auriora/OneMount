# Upload Manager Memory Optimization

**Date**: 2025-11-13  
**Component**: Upload Manager  
**Issue**: #008 - Memory Usage for Large Files

## Problem

The upload manager was loading entire file contents into memory (`Data []byte` field in `UploadSession`), causing excessive memory usage for large files (>100MB). This could lead to:
- High memory consumption when uploading multiple large files
- Potential out-of-memory errors on systems with limited RAM
- Poor performance due to memory pressure

## Solution

Implemented streaming upload for large files (>= 100MB) to reduce memory footprint:

### Changes Made

1. **UploadSession Structure** (`internal/fs/upload_session.go`):
   - Added `ContentPath string` field to store the path to cached file content
   - Modified `Data []byte` to be used only for small files (<100MB)
   - Updated `MarshalJSON` to include `ContentPath`

2. **NewUploadSessionWithPath** (`internal/fs/upload_session.go`):
   - New constructor that creates upload sessions using file paths instead of loading data into memory
   - Calculates hash by reading file once, then closes it
   - More memory-efficient for large files

3. **uploadChunk Method** (`internal/fs/upload_session.go`):
   - Updated to support both in-memory (`Data []byte`) and streaming (`ContentPath`) uploads
   - For streaming uploads, opens file, seeks to offset, reads only the required chunk
   - Closes file after reading chunk to minimize open file descriptors

4. **Upload Method** (`internal/fs/upload_session.go`):
   - Updated small file upload path to support both in-memory and streaming
   - Creates appropriate reader based on whether `Data` or `ContentPath` is set

5. **FilesystemInterface** (`internal/fs/filesystem_types.go`):
   - Added `GetInodeContentPath(inode *Inode) string` method
   - Allows upload manager to get file path without loading content

6. **Filesystem Implementation** (`internal/fs/fs.go`):
   - Implemented `GetInodeContentPath` to return cache file path
   - Uses existing `content.contentPath()` method

7. **QueueUploadWithPriority** (`internal/fs/upload_manager.go`):
   - Added logic to determine upload strategy based on file size
   - Files >= 100MB use `NewUploadSessionWithPath` (streaming)
   - Files < 100MB use `NewUploadSession` (in-memory, faster)

## Memory Requirements

### Before Optimization
- **Small file (10MB)**: ~10MB memory per upload
- **Large file (500MB)**: ~500MB memory per upload
- **Multiple large files**: Memory usage multiplies (e.g., 5 files = 2.5GB)

### After Optimization
- **Small file (10MB)**: ~10MB memory per upload (unchanged, optimized for speed)
- **Large file (500MB)**: ~10MB memory per upload (only chunk size in memory)
- **Multiple large files**: ~10MB per upload regardless of file size

### Chunk Size
- Upload chunk size: 10MB (as per Microsoft Graph API recommendations)
- Only one chunk is held in memory at a time during upload
- Memory usage is constant regardless of file size

## Testing

### Manual Testing
To test with large files:

```bash
# Create a test file > 100MB
dd if=/dev/urandom of=/tmp/large_test_file bs=1M count=150

# Mount filesystem in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm shell

# Copy large file to mount point
cp /tmp/large_test_file /mnt/onedrive/

# Monitor memory usage during upload
watch -n 1 'ps aux | grep onemount'
```

### Integration Tests
Run existing upload tests to verify no regressions:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests \
  go test -v -run TestIT_FS.*Upload ./internal/fs
```

## Performance Considerations

### Trade-offs
1. **Small files (<100MB)**:
   - Still use in-memory approach for faster uploads
   - No performance degradation
   - Optimized for common case (most files are small)

2. **Large files (>=100MB)**:
   - Slightly slower due to disk I/O for each chunk
   - Significantly reduced memory usage
   - Better system stability and scalability

### Threshold Selection
- 100MB threshold chosen based on:
  - Typical system memory constraints
  - Microsoft Graph API chunk size (10MB)
  - Balance between performance and memory usage
  - Most files are < 100MB, so common case remains fast

## Future Improvements

1. **Configurable Threshold**:
   - Add command-line flag or config option for large file threshold
   - Allow users to tune based on their system resources

2. **Memory Usage Metrics**:
   - Add metrics to track actual memory usage during uploads
   - Expose via stats API or logging

3. **Adaptive Strategy**:
   - Monitor system memory pressure
   - Dynamically adjust threshold based on available memory

4. **Parallel Chunk Upload**:
   - For very large files, consider uploading multiple chunks in parallel
   - Would require careful memory management to avoid exceeding limits

## Requirements Addressed

- **Requirement 4.3**: Large file upload using chunked upload sessions
- **Requirement 11.1**: Integration test coverage for upload workflows

## Related Issues

- Issue #008: Upload Manager - Memory Usage for Large Files (RESOLVED)

## References

- Microsoft Graph API Documentation: https://docs.microsoft.com/en-us/graph/api/driveitem-createuploadsession
- Upload Session Best Practices: https://docs.microsoft.com/en-us/graph/api/driveitem-createuploadsession#best-practices
