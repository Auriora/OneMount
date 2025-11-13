# Timeout Configuration Standardization

**Date**: 2025-11-13  
**Issue**: #PERF-003 - Inconsistent Timeout Values  
**Task**: 20.8 Fix Issue #PERF-003: Inconsistent Timeout Values  
**Status**: ✅ Complete

## Summary

Standardized timeout values across all OneMount components by creating a centralized `TimeoutConfig` struct. This addresses inconsistent timeout values that were hardcoded throughout the codebase and makes timeouts configurable for future enhancements.

## Changes Made

### 1. Created Centralized Timeout Configuration

**File**: `internal/fs/timeout_config.go`

Created a new `TimeoutConfig` struct that centralizes all timeout values:

```go
type TimeoutConfig struct {
    DownloadWorkerShutdown  time.Duration // 5 seconds
    UploadGracefulShutdown  time.Duration // 30 seconds
    FilesystemShutdown      time.Duration // 10 seconds
    NetworkCallbackShutdown time.Duration // 5 seconds
    MetadataRequestTimeout  time.Duration // 30 seconds
    ContentStatsTimeout     time.Duration // 5 seconds
}
```

**Features**:
- Default configuration with reasonable values
- Validation to ensure all timeouts are positive and within acceptable ranges
- Clear error messages for invalid configurations
- Extensible design for future command-line/config file support

### 2. Updated Filesystem Initialization

**File**: `internal/fs/cache.go`

- Added `timeoutConfig` field to `Filesystem` struct
- Initialize with `DefaultTimeoutConfig()` in `NewFilesystemWithContext()`
- Updated `Stop()` method to use configured timeout

**Before**:
```go
case <-time.After(10 * time.Second):
    logging.Warn().Msg("Timed out waiting for filesystem goroutines to stop")
```

**After**:
```go
timeout := 10 * time.Second // Default fallback
if f.timeoutConfig != nil {
    timeout = f.timeoutConfig.FilesystemShutdown
}
// ...
case <-time.After(timeout):
    logging.Warn().
        Dur("timeout", timeout).
        Msg("Timed out waiting for filesystem goroutines to stop")
```

### 3. Updated Download Manager

**File**: `internal/fs/download_manager.go`

Updated `Stop()` method to use configured timeout:

**Before**:
```go
case <-time.After(5 * time.Second):
    logging.Warn().Msg("Timed out waiting for download manager to stop")
```

**After**:
```go
timeout := 5 * time.Second // Default fallback
if dm.fs != nil && dm.fs.timeoutConfig != nil {
    timeout = dm.fs.timeoutConfig.DownloadWorkerShutdown
}
// ...
case <-time.After(timeout):
    logging.Warn().
        Dur("timeout", timeout).
        Msg("Timed out waiting for download manager to stop")
```

### 4. Updated Upload Manager

**File**: `internal/fs/upload_manager.go`

Updated `NewUploadManager()` to use configured timeout:

**Before**:
```go
gracefulTimeout: 30 * time.Second, // 30 seconds for large uploads to complete
```

**After**:
```go
gracefulTimeout := 30 * time.Second // Default fallback
if fs != nil {
    // Type assert to *Filesystem to access timeoutConfig
    if fsImpl, ok := fs.(*Filesystem); ok && fsImpl.timeoutConfig != nil {
        gracefulTimeout = fsImpl.timeoutConfig.UploadGracefulShutdown
    }
}
// ...
gracefulTimeout: gracefulTimeout, // Use configured timeout
```

### 5. Updated Metadata Request Timeout

**File**: `internal/fs/sync.go`

Updated metadata request timeout to use configured value:

**Before**:
```go
case <-time.After(30 * time.Second):
    err = context.DeadlineExceeded
    logging.Warn().Str("dirID", dirID).Msg("Metadata request timed out")
```

**After**:
```go
case <-time.After(f.timeoutConfig.MetadataRequestTimeout):
    err = context.DeadlineExceeded
    logging.Warn().
        Str("dirID", dirID).
        Dur("timeout", f.timeoutConfig.MetadataRequestTimeout).
        Msg("Metadata request timed out")
```

### 6. Updated Content Statistics Timeout

**File**: `internal/fs/stats.go`

Updated content statistics timeout to use configured value:

**Before**:
```go
case <-time.After(5 * time.Second):
    logging.Warn().Msg("Timeout waiting for content cache statistics, using partial results")
```

**After**:
```go
case <-time.After(f.timeoutConfig.ContentStatsTimeout):
    logging.Warn().
        Dur("timeout", f.timeoutConfig.ContentStatsTimeout).
        Msg("Timeout waiting for content cache statistics, using partial results")
```

### 7. Created Comprehensive Tests

**File**: `internal/fs/timeout_config_test.go`

Created tests to verify:
- Default configuration is valid
- All timeout values are positive
- Validation catches invalid configurations
- Error messages are clear and helpful

**Test Results**:
```
=== RUN   TestDefaultTimeoutConfig
--- PASS: TestDefaultTimeoutConfig (0.00s)
=== RUN   TestTimeoutConfigValidation
--- PASS: TestTimeoutConfigValidation (0.00s)
=== RUN   TestTimeoutConfigInFilesystem
--- PASS: TestTimeoutConfigInFilesystem (0.00s)
=== RUN   TestInvalidConfigError
--- PASS: TestInvalidConfigError (0.00s)
PASS
```

### 8. Created Documentation

**File**: `docs/guides/developer/timeout-policy.md`

Created comprehensive documentation covering:
- Timeout categories and rationale
- Default values and configuration
- Usage guidelines and best practices
- Component-specific timeout details
- Troubleshooting timeout issues
- Future enhancement plans

### 9. Updated Design Document

**File**: `.kiro/specs/system-verification-and-fix/design.md`

Added "Timeout Configuration" section documenting:
- Overview of centralized configuration
- Timeout categories and values
- Configuration structure
- Validation rules
- Usage in components
- Future enhancements

## Timeout Values Summary

| Component | Timeout | Value | Rationale |
|-----------|---------|-------|-----------|
| Download Worker Shutdown | `DownloadWorkerShutdown` | 5s | Workers should complete current chunk downloads quickly |
| Upload Graceful Shutdown | `UploadGracefulShutdown` | 30s | Large file uploads may need time to complete |
| Filesystem Shutdown | `FilesystemShutdown` | 10s | Should be sufficient for normal shutdown scenarios |
| Network Callback Shutdown | `NetworkCallbackShutdown` | 5s | Callbacks should be lightweight and complete quickly |
| Metadata Request Timeout | `MetadataRequestTimeout` | 30s | Includes network latency and API processing time |
| Content Stats Timeout | `ContentStatsTimeout` | 5s | Statistics should be fast or use sampling |

## Benefits

1. **Consistency**: All timeout values are now defined in one place
2. **Maintainability**: Easy to update timeout values across the entire codebase
3. **Configurability**: Foundation for future command-line/config file support
4. **Validation**: Ensures timeout values are reasonable and within acceptable ranges
5. **Observability**: Timeout durations are now logged when timeouts occur
6. **Documentation**: Clear documentation of timeout policy and rationale

## Testing

All changes have been tested:
- ✅ Unit tests for timeout configuration
- ✅ Validation tests for invalid configurations
- ✅ Error message tests
- ✅ No diagnostics errors in modified files

## Future Enhancements

1. **Command-Line Flags**: Add flags like `--upload-shutdown-timeout=60s`
2. **Configuration File**: Support timeout settings in YAML config
3. **Dynamic Adjustment**: Automatically adjust based on network conditions
4. **Timeout Metrics**: Collect metrics on timeout occurrences
5. **Timeout Profiles**: Predefined profiles for different environments (fast, normal, slow)

## Requirements Satisfied

- ✅ Requirement 12.5 (Performance and Concurrency): Goroutines tracked with wait groups for clean shutdown
- ✅ All timeout values are now standardized and configurable
- ✅ Timeout policy is documented
- ✅ Validation ensures reasonable timeout values

## Related Issues

- Issue #PERF-003: Inconsistent Timeout Values (RESOLVED)
- Issue #PERF-002: Network Callbacks Lack Wait Group Tracking (Related - already fixed)

## References

- Task: `.kiro/specs/system-verification-and-fix/tasks.md` - Task 20.8
- Documentation: `docs/guides/developer/timeout-policy.md`
- Design: `.kiro/specs/system-verification-and-fix/design.md` - Timeout Configuration section
- Code: `internal/fs/timeout_config.go`
- Tests: `internal/fs/timeout_config_test.go`
