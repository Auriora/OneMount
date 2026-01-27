# Lazy Directory Loading Performance Issue

**Date**: 2026-01-27  
**Status**: 🔍 Investigation Required  
**Severity**: Medium  
**Component**: Filesystem / Metadata Caching

## Problem Description

When navigating directories in the mounted OneDrive filesystem, there is a noticeable delay before subdirectories and files appear. This affects user experience, particularly when browsing through folder hierarchies.

### Observed Behavior

1. **Initial Directory Listing**: When first accessing a directory (e.g., `.Trash-1000`), it appears empty
2. **Delayed Population**: After 5-10 seconds, subdirectories appear (e.g., `files/` and `info/`)
3. **Inconsistent Loading**: Some directories load quickly, others have significant delays
4. **File Manager Impact**: File managers (Nautilus, Dolphin, etc.) show empty folders initially, then populate after delay

### Example

```bash
# First access - appears empty
$ ls -la /home/bcherrington/onmount-auriora/.Trash-1000/
total 8
drwxr-xr-x 2 bcherrington bcherrington 4096 Nov 12 06:37 .
drwxr-xr-x 8 bcherrington bcherrington 4096 Jan 26 22:53 ..

# After 5-10 seconds - subdirectories appear
$ ls -la /home/bcherrington/onmount-auriora/.Trash-1000/
total 16
drwxr-xr-x 4 bcherrington bcherrington 4096 Nov 12 06:37 .
drwxr-xr-x 8 bcherrington bcherrington 4096 Jan 26 22:53 ..
drwxr-xr-x 2 bcherrington bcherrington 4096 Nov 12 06:37 files
drwxr-xr-x 2 bcherrington bcherrington 4096 Nov 12 06:37 info
```

## Technical Context

### Current Implementation

OneMount uses a lazy-loading approach for directory contents:

1. **On-Demand Fetching**: Directory contents are fetched from Microsoft Graph API when accessed
2. **Background Refresh**: `GetChildrenID()` schedules background refresh if children not in cache
3. **Cache Strategy**: Metadata is cached in SQLite database with expiration times
4. **Metadata Queue**: Background workers process metadata requests asynchronously

### Relevant Code Locations

- `internal/fs/fs.go`: FUSE filesystem operations (`Readdir`, `Lookup`)
- `internal/graph/graph.go`: `GetChildrenID()` - fetches children from API
- `internal/fs/cache.go`: Metadata caching logic
- `internal/metadata/manager.go`: Metadata request queue and workers

### Configuration Parameters

From `cmd/common/config.go`:

```go
DeltaInterval        int  // Delta sync interval (seconds)
ActiveDeltaInterval  int  // Active delta interval (seconds)
ActiveDeltaWindow    int  // Active delta window (seconds)
CacheExpiration      int  // Cache expiration time
MetadataQueue        MetadataQueueConfig
```

## Root Cause Analysis (Preliminary)

### Hypothesis 1: Cache Miss on First Access

When a directory is accessed for the first time:
1. FUSE `Readdir` operation is called
2. `GetChildrenID()` checks cache - MISS
3. Background refresh is scheduled
4. Empty result returned immediately
5. Background worker fetches from API (5-10 seconds)
6. Next `Readdir` call returns cached results

**Evidence**: Logs show "Children not in cache; scheduling background refresh"

### Hypothesis 2: Synchronous vs Asynchronous Loading

The current implementation prioritizes responsiveness over completeness:
- Returns immediately with cached data (even if empty)
- Schedules background fetch
- Requires second access to see results

**Trade-off**: Fast initial response vs. complete data on first access

### Hypothesis 3: Metadata Queue Backlog

If metadata queue is backed up:
- Background refresh requests wait in queue
- Delay increases with queue depth
- Workers may be busy with other requests

**Evidence**: Need to check queue depth and worker utilization

## Investigation Steps

### 1. Analyze Current Caching Behavior

```bash
# Check cache database
sqlite3 ~/.cache/onemount/accounts/<hash>/onemount.db "SELECT COUNT(*) FROM metadata;"
sqlite3 ~/.cache/onemount/accounts/<hash>/onemount.db "SELECT id, name, parent_id FROM metadata WHERE name = '.Trash-1000';"

# Monitor cache hits/misses
journalctl --user -u "onemount@*.service" -f | grep -i "cache\|children"
```

### 2. Measure API Response Times

Add timing instrumentation to:
- `GetChildrenID()` - time from call to completion
- Graph API requests - network latency
- Cache operations - database query time

### 3. Test Different Cache Strategies

Experiment with:
- **Eager Loading**: Pre-fetch children when parent is accessed
- **Prefetching**: Load common directories on mount
- **Increased Cache TTL**: Reduce cache expiration frequency
- **Synchronous First Load**: Block on first access until data available

### 4. Profile Metadata Queue

Monitor:
- Queue depth over time
- Worker utilization
- Request processing time
- Backlog accumulation

### 5. Compare with Other FUSE Filesystems

Benchmark against:
- `rclone mount` (OneDrive)
- `google-drive-ocamlfuse`
- `s3fs`

## Potential Solutions

### Option 1: Hybrid Loading Strategy

**Approach**: Synchronous load on first access, async refresh thereafter

```go
func GetChildrenID(id string) (map[string]*Inode, error) {
    // Check cache
    cached := cache.Get(id)
    if cached != nil && !cached.IsExpired() {
        return cached, nil
    }
    
    // First access or expired - fetch synchronously
    if cached == nil {
        return fetchChildrenSync(id)
    }
    
    // Expired but have stale data - return stale, refresh async
    go refreshChildrenAsync(id)
    return cached, nil
}
```

**Pros**: Complete data on first access  
**Cons**: Slower initial response (5-10s)

### Option 2: Prefetch Common Directories

**Approach**: Pre-load root and common directories on mount

```go
func initializeFilesystem() {
    // Prefetch root
    GetChildrenID(rootID)
    
    // Prefetch common directories
    for _, commonDir := range []string{"Documents", "Pictures", "Desktop"} {
        if inode := LookupPath(commonDir); inode != nil {
            go GetChildrenID(inode.ID)
        }
    }
}
```

**Pros**: Fast access to common directories  
**Cons**: Increased mount time, may prefetch unused directories

### Option 3: Readahead / Predictive Loading

**Approach**: When directory is accessed, prefetch its children's children

```go
func GetChildrenID(id string) (map[string]*Inode, error) {
    children := fetchChildren(id)
    
    // Prefetch grandchildren in background
    for _, child := range children {
        if child.IsFolder() {
            go GetChildrenID(child.ID)
        }
    }
    
    return children, nil
}
```

**Pros**: Smooth navigation experience  
**Cons**: Increased API usage, may fetch unnecessary data

### Option 4: Increase Metadata Worker Pool

**Approach**: Add more workers to process metadata requests faster

```go
// Current: 3 workers
// Proposed: 8-10 workers

metadataQueue.Workers = 10
```

**Pros**: Faster background refresh  
**Cons**: Increased memory usage, more concurrent API requests

### Option 5: Implement FUSE Readdirplus

**Approach**: Use `Readdirplus` to return directory entries with attributes in single call

```go
func (fs *Filesystem) Readdirplus(ctx context.Context, op *fuseops.ReaddirplusOp) error {
    children := GetChildrenID(op.Inode)
    
    // Return entries with full attributes
    for _, child := range children {
        op.Entries = append(op.Entries, fuseutil.Dirent{
            Inode: child.Inode,
            Name:  child.Name,
            Type:  child.Type,
            // Include full attributes
            Attributes: child.Attributes,
        })
    }
}
```

**Pros**: Reduces round trips, better performance  
**Cons**: Requires FUSE protocol support, more complex implementation

## Performance Targets

### Current Performance
- First directory access: 5-10 seconds delay
- Subsequent access: < 100ms (cached)
- Root directory: ~2 seconds (277 items)

### Target Performance
- First directory access: < 2 seconds
- Subsequent access: < 50ms (cached)
- Root directory: < 1 second
- Prefetched directories: < 100ms

## Testing Plan

1. **Benchmark Current Performance**
   - Measure directory access times across various scenarios
   - Profile cache hit/miss rates
   - Monitor API request latency

2. **Test Each Solution**
   - Implement in feature branch
   - Run performance benchmarks
   - Measure user experience impact

3. **A/B Testing**
   - Compare solutions side-by-side
   - Gather user feedback
   - Measure API usage impact

4. **Regression Testing**
   - Ensure no performance degradation in other areas
   - Verify cache consistency
   - Test with large directories (1000+ items)

## Related Issues

- Metadata caching strategy
- Background refresh timing
- API rate limiting considerations
- Memory usage with increased caching

## References

- FUSE documentation: https://libfuse.github.io/doxygen/
- Microsoft Graph API: https://docs.microsoft.com/en-us/graph/api/driveitem-list-children
- Similar issues in other FUSE filesystems:
  - rclone: https://github.com/rclone/rclone/issues/2975
  - google-drive-ocamlfuse: https://github.com/astrada/google-drive-ocamlfuse/issues/367
