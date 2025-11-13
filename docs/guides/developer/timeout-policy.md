# Timeout Policy

## Overview

This document describes the timeout policy for OneMount, including standardized timeout values across all components and guidelines for configuring timeouts.

## Centralized Timeout Configuration

All timeout values are centralized in the `TimeoutConfig` struct (`internal/fs/timeout_config.go`). This ensures consistency across all managers and background processes.

## Timeout Categories

### Short Operations (< 5 seconds)

Used for quick checks and lightweight operations that should complete quickly:

- **Download Worker Shutdown**: 5 seconds
  - Time to wait for download workers to finish during shutdown
  - Workers should complete current chunk downloads and exit
  
- **Network Callback Shutdown**: 5 seconds
  - Time to wait for network feedback callbacks to complete
  - Callbacks should be lightweight and complete quickly
  
- **Content Stats Timeout**: 5 seconds
  - Time to wait for content cache statistics collection
  - Statistics collection should be fast or use sampling

### Medium Operations (5-30 seconds)

Used for network requests and file operations:

- **Metadata Request Timeout**: 30 seconds
  - Time to wait for metadata fetch operations from OneDrive
  - Includes network latency and API processing time
  - May need to be increased for slow connections

### Long Operations (30 seconds - 2 minutes)

Used for large file uploads/downloads and complex operations:

- **Upload Graceful Shutdown**: 30 seconds
  - Time to wait for active uploads to complete during shutdown
  - Large file uploads may need more time to complete
  - Configurable for environments with slow upload speeds

### Graceful Shutdown (10-60 seconds)

Used for clean shutdown of background processes:

- **Filesystem Shutdown**: 10 seconds
  - Time to wait for all filesystem goroutines to stop
  - Includes stopping all managers and background processes
  - Should be sufficient for normal shutdown scenarios

## Configuration

### Default Values

The default timeout configuration is defined in `DefaultTimeoutConfig()`:

```go
&TimeoutConfig{
    DownloadWorkerShutdown:  5 * time.Second,
    UploadGracefulShutdown:  30 * time.Second,
    FilesystemShutdown:      10 * time.Second,
    NetworkCallbackShutdown: 5 * time.Second,
    MetadataRequestTimeout:  30 * time.Second,
    ContentStatsTimeout:     5 * time.Second,
}
```

### Command-Line Configuration

Timeout values can be configured via command-line flags (future enhancement):

```bash
onemount --upload-shutdown-timeout=60s \
         --download-shutdown-timeout=10s \
         --filesystem-shutdown-timeout=30s \
         /mnt/onedrive
```

### Configuration File

Timeout values can be configured in the configuration file (future enhancement):

```yaml
timeouts:
  upload_shutdown: 60s
  download_shutdown: 10s
  filesystem_shutdown: 30s
  metadata_request: 45s
  content_stats: 10s
  network_callback: 5s
```

## Validation

All timeout values are validated on startup:

- **Positive Values**: All timeouts must be positive (> 0)
- **Minimum Values**: Most timeouts should be at least 1 second
- **Maximum Values**: Timeouts should not exceed 5 minutes to prevent indefinite hangs

Invalid timeout values will result in an error on startup with a clear message indicating the problem.

## Usage Guidelines

### When to Increase Timeouts

Consider increasing timeouts in the following scenarios:

1. **Slow Network Connections**: Increase metadata request and upload/download timeouts
2. **Large File Operations**: Increase upload graceful shutdown timeout
3. **High Latency Environments**: Increase all network-related timeouts
4. **Resource-Constrained Systems**: Increase shutdown timeouts to allow more time for cleanup

### When to Decrease Timeouts

Consider decreasing timeouts in the following scenarios:

1. **Fast Network Connections**: Decrease metadata request timeouts for faster failure detection
2. **Testing Environments**: Decrease all timeouts to speed up test execution
3. **Development Environments**: Decrease shutdown timeouts for faster iteration

### Best Practices

1. **Start with Defaults**: Use default timeout values unless you have a specific reason to change them
2. **Monitor Logs**: Watch for timeout warnings in logs to identify operations that need more time
3. **Test Changes**: Test timeout changes in a non-production environment first
4. **Document Rationale**: Document why you changed timeout values from defaults
5. **Consider Trade-offs**: Longer timeouts improve reliability but may delay error detection

## Component-Specific Timeouts

### Download Manager

- **Worker Shutdown**: Time to wait for download workers to finish
- **Current Value**: 5 seconds
- **Rationale**: Workers should complete current chunk downloads quickly
- **When to Increase**: If workers frequently timeout during shutdown

### Upload Manager

- **Graceful Shutdown**: Time to wait for active uploads to complete
- **Current Value**: 30 seconds
- **Rationale**: Large file uploads may need time to complete
- **When to Increase**: If large uploads frequently timeout during shutdown

### Filesystem

- **Shutdown**: Time to wait for all goroutines to stop
- **Current Value**: 10 seconds
- **Rationale**: Should be sufficient for normal shutdown scenarios
- **When to Increase**: If shutdown frequently times out

### Metadata Requests

- **Request Timeout**: Time to wait for metadata fetch operations
- **Current Value**: 30 seconds
- **Rationale**: Includes network latency and API processing time
- **When to Increase**: If metadata requests frequently timeout on slow connections

### Content Statistics

- **Stats Timeout**: Time to wait for statistics collection
- **Current Value**: 5 seconds
- **Rationale**: Statistics should be fast or use sampling
- **When to Increase**: If statistics collection frequently times out with large datasets

### Network Callbacks

- **Callback Shutdown**: Time to wait for callbacks to complete
- **Current Value**: 5 seconds
- **Rationale**: Callbacks should be lightweight and complete quickly
- **When to Increase**: If callbacks perform complex operations

## Troubleshooting

### Timeout Warnings in Logs

If you see timeout warnings in logs:

1. **Identify the Component**: Check which component is timing out
2. **Check Network Conditions**: Verify network connectivity and latency
3. **Review System Resources**: Check CPU, memory, and disk usage
4. **Increase Timeout**: If appropriate, increase the timeout value
5. **Investigate Root Cause**: Determine why the operation is taking longer than expected

### Frequent Timeouts

If timeouts occur frequently:

1. **Network Issues**: Check for network connectivity problems
2. **Resource Constraints**: Verify system has sufficient resources
3. **API Rate Limiting**: Check if OneDrive API is rate limiting requests
4. **Configuration Issues**: Verify timeout values are appropriate for your environment

### Shutdown Hangs

If shutdown hangs or takes too long:

1. **Check Active Operations**: Verify no long-running operations are blocking shutdown
2. **Review Goroutine Leaks**: Check for goroutines that aren't being tracked properly
3. **Increase Shutdown Timeout**: If operations legitimately need more time
4. **Fix Blocking Code**: Identify and fix code that blocks shutdown

## Future Enhancements

1. **Dynamic Timeout Adjustment**: Automatically adjust timeouts based on network conditions
2. **Per-Operation Timeouts**: Allow different timeouts for different types of operations
3. **Timeout Metrics**: Collect metrics on timeout occurrences and durations
4. **Adaptive Timeouts**: Learn optimal timeout values based on historical data
5. **Timeout Profiles**: Predefined timeout profiles for different environments (fast, normal, slow)

## Related Documentation

- [Concurrency Guidelines](concurrency-guidelines.md) - Guidelines for goroutine management
- [Error Handling](../../2-architecture/software-design-specification.md#error-handling) - Error handling patterns
- [Configuration](../../guides/user/configuration.md) - Configuration file format (future)
