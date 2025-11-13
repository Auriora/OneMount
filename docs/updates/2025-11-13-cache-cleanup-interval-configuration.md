# Cache Cleanup Interval Configuration

**Date**: 2025-11-13  
**Component**: Cache Management  
**Issue**: #CACHE-004 - Fixed 24-Hour Cleanup Interval  
**Status**: ✅ COMPLETED

## Summary

Made the cache cleanup interval configurable to allow users to adjust how frequently the cache cleanup process runs. Previously, the cleanup interval was hardcoded to 24 hours.

## Changes Made

### 1. Configuration Structure (`cmd/common/config.go`)

Added `CacheCleanupInterval` field to the `Config` struct:

```go
type Config struct {
    // ... existing fields ...
    CacheCleanupInterval int    `yaml:"cacheCleanupInterval"` // Cache cleanup interval in hours
    // ... existing fields ...
}
```

**Default Value**: 24 hours  
**Valid Range**: 1-720 hours (1 hour to 30 days)

### 2. Validation

Added validation in `validateConfig()` to ensure the cleanup interval is within reasonable bounds:

```go
// Validate CacheCleanupInterval (1 hour to 30 days = 720 hours)
if config.CacheCleanupInterval < 1 || config.CacheCleanupInterval > 720 {
    logging.Warn().
        Int("cacheCleanupInterval", config.CacheCleanupInterval).
        Msg("Cache cleanup interval must be between 1 and 720 hours (1 hour to 30 days), using default.")
    config.CacheCleanupInterval = 24
}
```

### 3. Command-Line Flag (`cmd/onemount/main.go`)

Added `--cache-cleanup-interval` flag:

```go
cacheCleanupInterval := flag.IntP("cache-cleanup-interval", "", 0,
    "Set the interval in hours between cache cleanup runs. "+
        "Default is 24 hours. Valid range: 1-720 hours (1 hour to 30 days). Set to 0 to use the default.")
```

### 4. Filesystem Structure (`internal/fs/filesystem_types.go`)

Added `cacheCleanupInterval` field to the `Filesystem` struct:

```go
// Cache cleanup configuration
cacheExpirationDays   int            // Number of days after which cached files expire
cacheCleanupInterval  time.Duration  // Interval between cache cleanup runs
cacheCleanupStop      chan struct{}  // Channel to signal cache cleanup to stop
cacheCleanupStopOnce  sync.Once      // Ensures cleanup is stopped only once
cacheCleanupWg        sync.WaitGroup // Wait group for cache cleanup goroutine
```

### 5. Filesystem Initialization (`internal/fs/cache.go`)

Updated `NewFilesystemWithContext()` to accept and validate the cleanup interval:

```go
func NewFilesystemWithContext(ctx context.Context, auth *graph.Auth, cacheDir string, 
    cacheExpirationDays int, cacheCleanupIntervalHours int) (*Filesystem, error) {
    
    // Validate and set cache cleanup interval (default to 24 hours if invalid)
    cleanupInterval := time.Duration(cacheCleanupIntervalHours) * time.Hour
    if cacheCleanupIntervalHours < 1 || cacheCleanupIntervalHours > 720 {
        logging.Warn().
            Int("cacheCleanupIntervalHours", cacheCleanupIntervalHours).
            Msg("Invalid cache cleanup interval, using default of 24 hours")
        cleanupInterval = 24 * time.Hour
    }
    
    // ... initialize filesystem with cleanupInterval ...
}
```

### 6. Cache Cleanup Routine (`internal/fs/cache.go`)

Updated `StartCacheCleanup()` to use the configured interval:

```go
func (f *Filesystem) StartCacheCleanup() {
    // ... validation ...
    
    logging.Info().
        Int("expirationDays", f.cacheExpirationDays).
        Dur("cleanupInterval", f.cacheCleanupInterval).
        Msg("Starting content cache cleanup routine")
    
    // ... goroutine setup ...
    
    // Set up ticker for periodic cleanup using configured interval
    ticker := time.NewTicker(f.cacheCleanupInterval)
    defer ticker.Stop()
    
    // ... cleanup loop ...
}
```

### 7. Test Updates

Updated all test files to pass the cleanup interval parameter (default 24 hours):

- `internal/fs/mount_integration_test.go`
- `internal/fs/mount_integration_real_test.go`
- `internal/fs/etag_validation_integration_test.go`
- `internal/fs/end_to_end_workflow_test.go`

## Usage

### Command-Line

```bash
# Use default 24-hour interval
onemount /path/to/mountpoint

# Set cleanup to run every 6 hours
onemount --cache-cleanup-interval 6 /path/to/mountpoint

# Set cleanup to run every 48 hours (2 days)
onemount --cache-cleanup-interval 48 /path/to/mountpoint
```

### Configuration File

Add to `~/.config/onemount/config.yml`:

```yaml
cacheCleanupInterval: 12  # Run cleanup every 12 hours
```

## Benefits

1. **Flexibility**: Users can adjust cleanup frequency based on their needs
2. **Performance**: More frequent cleanup for systems with limited disk space
3. **Efficiency**: Less frequent cleanup for systems with ample storage
4. **Testing**: Easier to test cache cleanup behavior with shorter intervals

## Validation

- Minimum interval: 1 hour (prevents excessive cleanup overhead)
- Maximum interval: 720 hours / 30 days (prevents cache from growing indefinitely)
- Invalid values default to 24 hours with a warning message

## Requirements Addressed

- **Requirement 7.2**: Cache cleanup configuration and management

## Related Issues

- Issue #CACHE-004: Fixed 24-Hour Cleanup Interval

## Testing

All existing tests pass with the new parameter. The cleanup interval can be tested by:

1. Setting a short interval (e.g., 1 hour)
2. Monitoring logs for cleanup messages
3. Verifying cleanup runs at the configured interval

## Backward Compatibility

The `NewFilesystem()` wrapper function maintains backward compatibility by using the default 24-hour interval for existing code that doesn't specify the cleanup interval.
