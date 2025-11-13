# Large File Upload Retry Logic Fix

**Date**: 2025-11-13  
**Component**: Upload Manager / Upload Session  
**Issue**: #010 - Large File Upload Retry Logic Not Working  
**Task**: 19.1  
**Status**: ✅ RESOLVED

## Problem

The large file upload retry logic was not functioning correctly. When a chunk upload failed during a chunked upload (files > 4MB), the retry mechanism did not properly retry the failed chunk. The upload would fail immediately on the first error without attempting retries with exponential backoff.

### Symptoms

1. Chunk upload failures caused immediate upload session failure
2. No retry attempts were made (only 1 attempt instead of 3+)
3. ETag remained at initial value (not updated after successful retry)
4. File status remained "Syncing" instead of transitioning to "Local"
5. Upload did not recover from transient network failures

### Test Failures

Test `TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry` was failing with:
- "Expected no error, but got failed to perform chunk upload"
- "First chunk should have been attempted at least 3 times (was 1)"
- "ETag should be updated after successful upload"
- "File status should be Local after successful upload"

## Root Cause

The chunk upload retry logic in `internal/fs/upload_session.go` had a critical flaw:

1. **Immediate failure on error**: When `uploadChunk()` returned an error (not just a status code), the code immediately returned without retrying (line 649)
2. **Retry loop only handled 5xx status codes**: The retry loop only handled HTTP 5xx status codes, but if `uploadChunk()` returned an error, it would exit immediately (line 677)
3. **No retry for network errors**: Network errors that prevented the HTTP request from completing were not retried

The original code structure:
```go
resp, status, err = u.uploadChunk(auth, uint64(i)*uploadChunkSize)
if err != nil {
    return u.setState(uploadErrored, errors.Wrap(err, "failed to perform chunk upload"))
}

// retry server-side failures with an exponential back-off strategy
for backoff := 1; status >= 500; backoff *= 2 {
    // ... retry logic only for 5xx status codes
    resp, status, err = u.uploadChunk(auth, uint64(i)*uploadChunkSize)
    if err != nil { // a serious, non 4xx/5xx error
        return u.setState(uploadErrored, errors.Wrap(err, "failed to perform chunk upload"))
    }
}
```

This meant that:
- If the first `uploadChunk()` call returned an error, the upload failed immediately
- The retry loop was never entered
- Even within the retry loop, if an error occurred, it would exit immediately

## Solution

Modified the chunk upload retry logic to handle both errors and 5xx status codes uniformly:

### Changes Made

1. **Unified retry logic**: Combined error handling and 5xx status code handling into a single retry loop
2. **Retry on errors**: Now retries when `uploadChunk()` returns an error, not just for 5xx status codes
3. **Exponential backoff**: Implements proper exponential backoff (1s, 2s, 4s, 8s, 16s)
4. **Configurable max retries**: Set max retries to 5 attempts to balance reliability with reasonable timeout
5. **Enhanced logging**: Added detailed logging to track retry attempts and reasons

### Code Changes

**File**: `internal/fs/upload_session.go`

**Before**:
```go
resp, status, err = u.uploadChunk(auth, uint64(i)*uploadChunkSize)
if err != nil {
    return u.setState(uploadErrored, errors.Wrap(err, "failed to perform chunk upload"))
}

for backoff := 1; status >= 500; backoff *= 2 {
    // ... retry only for 5xx
}
```

**After**:
```go
// Attempt chunk upload with retry logic for both errors and 5xx status codes
resp, status, err = u.uploadChunk(auth, uint64(i)*uploadChunkSize)

// Retry both errors and server-side failures (5xx) with exponential back-off strategy
const maxChunkRetries = 5
retryAttempt := 0
for (err != nil || status >= 500) && retryAttempt < maxChunkRetries {
    retryAttempt++
    backoff := 1 << uint(retryAttempt-1) // Exponential: 1, 2, 4, 8, 16 seconds
    if backoff > 16 {
        backoff = 16 // Cap at 16 seconds
    }

    // Check for context cancellation during retries
    select {
    case <-ctx.Done():
        if db != nil {
            u.persistProgress(db)
        }
        return u.setState(uploadErrored, errors.New("upload cancelled during retry"))
    default:
    }

    if err != nil {
        logging.Error().
            Str("id", u.ID).
            Str("name", u.Name).
            Int("chunk", i).
            Int("nchunks", nchunks).
            Int("retryAttempt", retryAttempt).
            Int("maxRetries", maxChunkRetries).
            Err(err).
            Msgf("Chunk upload failed with error, retrying in %ds.", backoff)
    } else {
        logging.Error().
            Str("id", u.ID).
            Str("name", u.Name).
            Int("chunk", i).
            Int("nchunks", nchunks).
            Int("status", status).
            Int("retryAttempt", retryAttempt).
            Int("maxRetries", maxChunkRetries).
            Msgf("The OneDrive server is having issues, retrying chunk upload in %ds.", backoff)
    }
    
    time.Sleep(time.Duration(backoff) * time.Second)
    resp, status, err = u.uploadChunk(auth, uint64(i)*uploadChunkSize)
}

// If we still have an error after all retries, fail the upload
if err != nil {
    return u.setState(uploadErrored, errors.Wrap(err, fmt.Sprintf("failed to perform chunk upload after %d retries", retryAttempt)))
}
```

### Key Improvements

1. **Retry condition**: `(err != nil || status >= 500)` - retries on both errors and 5xx status codes
2. **Max retries**: Limited to 5 attempts to prevent excessive delays (total max time: 1+2+4+8+16 = 31 seconds)
3. **Exponential backoff**: Doubles delay between retries (1s, 2s, 4s, 8s, 16s)
4. **Context cancellation**: Checks for context cancellation during retries to support graceful shutdown
5. **Detailed logging**: Logs retry attempts with context (chunk number, attempt number, error/status)
6. **Error message**: Final error message includes retry count for debugging

## Verification

### Test Results

Test `TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry` now passes:

```
=== RUN   TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry
    upload_retry_integration_test.go:394: First chunk retry delay: 1.000709966s
    upload_retry_integration_test.go:395: Second chunk retry delay: 2.000582888s
    upload_retry_integration_test.go:396: Total elapsed time: 4.985465739s
--- PASS: TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry (5.37s)
PASS
```

### Verified Behavior

✅ **Chunk upload retries**: First chunk is attempted 3+ times (initial + 2 retries)  
✅ **Exponential backoff**: Second retry delay (2s) > first retry delay (1s)  
✅ **ETag updated**: ETag is updated to "uploaded-etag-large-retry" after successful upload  
✅ **File status**: File status transitions to "Local" after successful upload  
✅ **Total time**: Upload completes in ~5 seconds (reasonable for 2 retries with backoff)

## Impact

### Requirements Satisfied

- **Requirement 4.4**: Upload retry with exponential backoff ✅
- **Requirement 4.5**: ETag updated after successful upload ✅
- **Requirement 8.1**: File status tracking ✅

### Benefits

1. **Improved reliability**: Large file uploads now recover from transient network failures
2. **Better user experience**: Files eventually upload successfully instead of failing permanently
3. **Proper error handling**: Network errors are retried, not just server errors
4. **Reasonable timeouts**: Max 5 retries with capped backoff prevents excessive delays
5. **Better observability**: Enhanced logging helps diagnose upload issues

### Scope

This fix applies to:
- ✅ Large file uploads (> 4MB) using chunked upload
- ✅ Chunk-level retry logic within a single upload session
- ❌ Small file uploads (< 4MB) - handled by upload manager retry logic
- ❌ Upload session-level retry logic - handled by upload manager

## Related Issues

- **Issue #011**: Upload Max Retries Exceeded Not Working (Task 19.2) - Separate issue for upload session state machine

## Testing

### Manual Testing

To manually test the fix:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  go test -v -run TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry ./internal/fs
```

### Integration Testing

The fix is covered by the existing integration test:
- `internal/fs/upload_retry_integration_test.go::TestIT_FS_09_04_02_LargeFileUploadFailureAndRetry`

## Notes

- The max retry count (5) and backoff cap (16s) are constants that can be adjusted if needed
- The fix only applies to chunked uploads (large files > 4MB)
- Small file uploads use a different code path and are handled by the upload manager's retry logic
- Context cancellation is properly handled during retries to support graceful shutdown
- Progress is persisted during retries for large files to enable recovery after crashes

## References

- **Task**: `.kiro/specs/system-verification-and-fix/tasks.md` - Task 19.1
- **Issue**: `docs/reports/verification-tracking.md` - Issue #010
- **Test**: `internal/fs/upload_retry_integration_test.go` - TestIT_FS_09_04_02
- **Code**: `internal/fs/upload_session.go` - UploadWithContext method
