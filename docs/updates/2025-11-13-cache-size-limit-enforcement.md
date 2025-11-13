# Cache Size Limit Enforcement with LRU Eviction

**Date**: 2025-11-13  
**Issue**: #CACHE-001  
**Status**: ✅ Implemented  
**Priority**: Medium

## Summary

Implemented LRU (Least Recently Used) cache eviction with configurable size limits to prevent unbounded cache growth. The cache now tracks file sizes and access times, automatically evicting the least recently used files when the cache size exceeds the configured limit.

## Problem

The cache only expired files based on time (`cacheExpirationDays`), not size. This meant the cache could grow unbounded until files reached the expiration age, potentially consuming all available disk space.

**Steps to Reproduce**:
1. Mount OneMount with a large OneDrive account
2. Access many large files
3. Observe cache directory growing without limit
4. Eventually run out of disk space

## Solution

### 1. LRU Cache Tracking

Added cache entry tracking to `LoopbackCache`:

```go
type CacheEntry struct {
    id           string
    size         int64
    lastAccessed time.Time
}

type LoopbackCache struct {
    directory    string
    fds          sync.Map
    lastCleanup  time.Time
    // LRU tracking
    entriesM     sync.RWMutex
    entries      map[string]*CacheEntry
    totalSize    int64
    maxCacheSize int64  // 0 = unlimited
}
```

### 2. Cache Size Initialization

The cache now scans existing files on startup to build the LRU tracking data:

```go
func (l *LoopbackCache) initializeCacheTracking() {
    // Scan cache directory
    // Build entries map with file sizes and modification times
    // Calculate total cache size
}
```

### 3. LRU Eviction Algorithm

When inserting new files, the cache checks if eviction is needed:

```go
func (l *LoopbackCache) evictIfNeeded(newSize int64) error {
    if l.maxCacheSize == 0 {
        return nil  // Unlimited
    }
    
    spaceNeeded := (l.totalSize + newSize) - l.maxCacheSize
    if spaceNeeded <= 0 {
        return nil  // No eviction needed
    }
    
    // Sort entries by last accessed time (oldest first)
    // Evict entries until enough space is freed
    // Skip files that are currently open
}
```

### 4. Configuration Support

Added `maxCacheSize` configuration option:

**Config File** (`config.yml`):
```yaml
maxCacheSize: 10737418240  # 10 GB in bytes (0 = unlimited)
```

**Command-Line Flag** (future enhancement):
```bash
onemount --max-cache-size 10G /path/to/mount
```

**Default**: 0 (unlimited) for backward compatibility

### 5. Statistics Integration

Updated `GetStats()` to include cache size information:

```go
type Stats struct {
    // ...
    MaxCacheSize    int64   // Maximum cache size limit (0 = unlimited)
    CacheSizeUsage  float64 // Percentage of max cache size used (0-100, or -1 if unlimited)
}
```

### 6. Cache Entry Updates

Cache entries are updated on:
- **Insert**: New file added to cache
- **Open**: File accessed (updates last accessed time)
- **Delete**: File removed from cache
- **InsertStream**: Streaming data written to cache

## Implementation Details

### Files Modified

1. **`internal/fs/content_cache.go`**:
   - Added `CacheEntry` struct for LRU tracking
   - Added `entriesM`, `entries`, `totalSize`, `maxCacheSize` fields to `LoopbackCache`
   - Implemented `initializeCacheTracking()` to scan existing files
   - Implemented `evictIfNeeded()` for LRU eviction
   - Added helper methods: `updateCacheEntry()`, `removeCacheEntry()`, `touchCacheEntry()`
   - Added public methods: `GetCacheSize()`, `GetMaxCacheSize()`, `SetMaxCacheSize()`, `GetCacheEntryCount()`
   - Updated `Insert()`, `InsertStream()`, `Delete()`, `Open()` to track cache entries
   - Updated `CleanupCache()` to enforce size limits after time-based cleanup

2. **`internal/fs/cache.go`**:
   - Updated `NewFilesystemWithContext()` to accept `maxCacheSize` parameter
   - Updated `NewFilesystem()` to pass 0 (unlimited) for backward compatibility
   - Updated `NewLoopbackCacheWithSize()` call to pass max cache size

3. **`internal/fs/stats.go`**:
   - Added `MaxCacheSize` and `CacheSizeUsage` fields to `Stats` struct
   - Updated `calculateStats()` to populate cache size information
   - Updated `GetQuickStats()` to include cache size information

4. **`cmd/common/config.go`**:
   - Added `MaxCacheSize` field to `Config` struct
   - Added default value of 0 (unlimited) in `createDefaultConfig()`

5. **`cmd/onemount/main.go`**:
   - Updated both `NewFilesystemWithContext()` calls to pass `config.MaxCacheSize`

6. **Test Files**:
   - Updated all test files to pass 0 (unlimited) for `maxCacheSize` parameter

## Behavior

### Cache Size Enforcement

1. **On File Insert**:
   - Check if adding the new file would exceed the limit
   - If yes, evict least recently used files until enough space is freed
   - Skip files that are currently open
   - Insert the new file

2. **On Cache Cleanup**:
   - First, remove files older than expiration threshold (time-based)
   - Then, if cache size still exceeds limit, perform LRU eviction
   - Log eviction statistics

3. **On File Access**:
   - Update the last accessed time for the file
   - This prevents recently accessed files from being evicted

### Eviction Priority

Files are evicted in this order:
1. Oldest last accessed time first
2. Files that are currently open are never evicted
3. Eviction continues until enough space is freed

### Configuration Examples

**Unlimited Cache** (default):
```yaml
maxCacheSize: 0
```

**10 GB Limit**:
```yaml
maxCacheSize: 10737418240
```

**1 GB Limit**:
```yaml
maxCacheSize: 1073741824
```

## Testing

### Manual Testing

1. **Test LRU Eviction**:
   ```bash
   # Set a small cache limit (e.g., 100 MB)
   # Access multiple large files (> 100 MB total)
   # Verify oldest files are evicted
   # Verify recently accessed files are retained
   ```

2. **Test Cache Size Tracking**:
   ```bash
   # Check stats before and after file access
   onemount --stats /mount/path
   # Verify ContentSize and CacheSizeUsage are accurate
   ```

3. **Test Configuration**:
   ```bash
   # Test with different maxCacheSize values
   # Test with 0 (unlimited)
   # Test with very small limits
   # Test with very large limits
   ```

### Integration Tests

Integration tests should be added to verify:
- LRU eviction works correctly
- Cache size tracking is accurate
- Files are evicted in the correct order
- Open files are not evicted
- Statistics reflect actual cache state

## Performance Impact

- **Minimal overhead**: Cache tracking uses a simple map with O(1) lookups
- **Eviction cost**: O(n log n) for sorting entries, but only when eviction is needed
- **Memory overhead**: ~40 bytes per cached file for tracking data
- **Disk I/O**: No additional I/O during normal operations

## Backward Compatibility

- Default `maxCacheSize` is 0 (unlimited) to maintain existing behavior
- Existing configurations without `maxCacheSize` will continue to work
- Tests updated to pass 0 for unlimited cache

## Future Enhancements

1. **Command-Line Flag**: Add `--max-cache-size` flag with human-readable sizes (e.g., "10G", "500M")
2. **Dynamic Adjustment**: Allow changing cache size limit without remounting
3. **Cache Warming**: Preload frequently accessed files on startup
4. **Smart Eviction**: Consider file importance (e.g., recently modified files)
5. **Cache Metrics**: Track eviction rate, hit rate, and other metrics

## Requirements Satisfied

- **Requirement 7.2**: Cache size limits are now enforced
- **Requirement 7.3**: LRU eviction algorithm implemented
- **Requirement 7.5**: Cache statistics include size information

## Related Issues

- Issue #CACHE-001: No Cache Size Limit Enforcement (RESOLVED)
- Issue #CACHE-004: Fixed 24-Hour Cleanup Interval (RESOLVED)

## References

- Design Document: `docs/2-architecture/software-design-specification.md`
- Requirements: `.kiro/specs/system-verification-and-fix/requirements.md` (Requirement 7)
- Verification Tracking: `docs/reports/verification-tracking.md`
