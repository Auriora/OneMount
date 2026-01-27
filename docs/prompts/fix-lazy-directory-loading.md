# Prompt: Fix Lazy Directory Loading Performance Issue

## Context

The OneMount FUSE filesystem has a performance issue where directories appear empty on first access, then populate after a 5-10 second delay. This affects user experience when browsing the mounted OneDrive filesystem.

## Problem Statement

When navigating directories in the mounted filesystem:
1. Initial directory listing shows empty (no subdirectories or files)
2. After 5-10 seconds, contents appear
3. File managers show empty folders initially, causing confusion
4. Subsequent is also slow when accessing the same directory again.

**Example:**
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

## Investigation Required

### Phase 1: Understand Current Behavior

1. **Read and analyze** `docs/issues/lazy-directory-loading-performance.md` for detailed context
2. **Examine** the current caching implementation:
   - `internal/fs/fs.go` - FUSE `Readdir` and `Lookup` operations
   - `internal/graph/graph.go` - `GetChildrenID()` function
   - `internal/fs/cache.go` - Metadata caching logic
   - `internal/metadata/manager.go` - Metadata queue and workers

3. **Trace the code flow** for directory access:
   - What happens when `Readdir` is called?
   - How does `GetChildrenID()` handle cache misses?
   - When is background refresh scheduled vs. synchronous fetch?
   - How long does the metadata queue take to process requests?

4. **Add instrumentation** to measure:
   - Time from `Readdir` call to data availability
   - Cache hit/miss rates
   - API request latency
   - Metadata queue depth and processing time

### Phase 2: Identify Root Cause

Determine which hypothesis is correct:

**Hypothesis A: Cache Miss Strategy**
- Current: Returns empty immediately, schedules background fetch
- Problem: User sees empty directory until next access
- Evidence: Logs show "Children not in cache; scheduling background refresh"

**Hypothesis B: Metadata Queue Backlog**
- Current: Background requests wait in queue
- Problem: Delay increases with queue depth
- Evidence: Need to check queue metrics

**Hypothesis C: API Latency**
- Current: Microsoft Graph API takes 5-10 seconds to respond
- Problem: Network/API performance bottleneck
- Evidence: Need to measure actual API response times

### Phase 3: Implement Solution

Based on root cause, implement one of these solutions:

#### Option 1: Hybrid Loading (Recommended)
**When to use**: If cache miss strategy is the issue

```go
func GetChildrenID(id string, firstAccess bool) (map[string]*Inode, error) {
    cached := cache.Get(id)
    
    // If first access and no cache, fetch synchronously
    if firstAccess && cached == nil {
        return fetchChildrenSync(id)
    }
    
    // If cached but expired, return stale and refresh async
    if cached != nil && cached.IsExpired() {
        go refreshChildrenAsync(id)
        return cached, nil
    }
    
    // If no cache, schedule async and return empty
    if cached == nil {
        go refreshChildrenAsync(id)
        return map[string]*Inode{}, nil
    }
    
    return cached, nil
}
```

**Changes needed:**
- Modify `Readdir` to detect first access vs. subsequent
- Add synchronous fetch path for first access
- Keep async refresh for subsequent accesses
- Add timeout to prevent hanging (max 10 seconds)

#### Option 2: Prefetch Common Directories
**When to use**: If specific directories are frequently accessed

```go
func initializeFilesystem() error {
    // Prefetch root directory
    go prefetchDirectory(rootID)
    
    // Prefetch common directories
    commonDirs := []string{"Documents", "Pictures", "Desktop", "Downloads"}
    for _, dir := range commonDirs {
        if inode := lookupPath(dir); inode != nil {
            go prefetchDirectory(inode.ID)
        }
    }
    
    return nil
}

func prefetchDirectory(id string) {
    children := GetChildrenID(id, true)
    
    // Optionally prefetch grandchildren for folders
    for _, child := range children {
        if child.IsFolder() && shouldPrefetch(child) {
            GetChildrenID(child.ID, false)
        }
    }
}
```

**Changes needed:**
- Add prefetch logic to filesystem initialization
- Implement heuristics for which directories to prefetch
- Add configuration option to enable/disable prefetching
- Monitor memory usage impact

#### Option 3: Increase Metadata Workers
**When to use**: If metadata queue is the bottleneck

```go
// In internal/metadata/manager.go
const (
    DefaultWorkers = 3  // Current
    OptimalWorkers = 10 // Proposed
)

func NewManager(workers int) *Manager {
    if workers == 0 {
        workers = OptimalWorkers
    }
    // ... rest of initialization
}
```

**Changes needed:**
- Increase worker pool size from 3 to 8-10
- Add configuration option for worker count
- Monitor memory and API rate limit impact
- Add metrics for queue depth and worker utilization

### Phase 4: Testing

1. **Benchmark current performance:**
   ```bash
   # Measure directory access time
   time ls -la /path/to/mount/.Trash-1000/
   
   # Test with various directory sizes
   time ls -la /path/to/mount/Documents/
   time ls -la /path/to/mount/  # Root with 277 items
   ```

2. **Test solution:**
   - Verify first access shows complete data
   - Measure performance improvement
   - Check cache hit rates
   - Monitor API usage
   - Test with file managers (Nautilus, Dolphin)

3. **Regression testing:**
   - Ensure no performance degradation elsewhere
   - Verify cache consistency
   - Test with large directories (1000+ items)
   - Check memory usage

### Phase 5: Documentation

1. Update `docs/issues/lazy-directory-loading-performance.md` with:
   - Root cause analysis results
   - Solution implemented
   - Performance measurements (before/after)
   - Configuration options added

2. Create `docs/fixes/lazy-directory-loading-fix.md` documenting:
   - Problem summary
   - Solution approach
   - Code changes made
   - Performance improvements
   - Testing results

## Success Criteria

- ✅ First directory access shows complete contents within 2 seconds
- ✅ Subsequent accesses remain fast (< 50ms)
- ✅ File managers display directories correctly on first open
- ✅ No increase in API rate limit errors
- ✅ Memory usage remains acceptable (< 20% increase)
- ✅ Cache consistency maintained
- ✅ Configuration options documented

## Key Files to Modify

1. `internal/fs/fs.go` - FUSE operations
2. `internal/graph/graph.go` - `GetChildrenID()` logic
3. `internal/fs/cache.go` - Caching strategy
4. `internal/metadata/manager.go` - Worker pool size
5. `cmd/common/config.go` - Add configuration options
6. `configs/default-config.yml` - Update default config

## Testing Commands

```bash
# Build and install
go build -o onemount ./cmd/onemount
sudo cp onemount /usr/bin/onemount

# Restart service
systemctl --user restart "onemount@home-bcherrington-onmount\\x2dauriora.service"

# Test directory access
time ls -la /home/bcherrington/onmount-auriora/.Trash-1000/
time ls -la /home/bcherrington/onmount-auriora/Documents/

# Monitor logs
journalctl --user -u "onemount@*.service" -f | grep -i "children\|cache\|fetch"

# Check cache database
sqlite3 ~/.cache/onemount/accounts/*/onemount.db "SELECT COUNT(*) FROM metadata;"
```

## References

- Issue documentation: `docs/issues/lazy-directory-loading-performance.md`
- FUSE documentation: https://libfuse.github.io/doxygen/
- Microsoft Graph API: https://docs.microsoft.com/en-us/graph/api/driveitem-list-children
- Similar issues in rclone: https://github.com/rclone/rclone/issues/2975

## Notes

- This is a performance optimization, not a bug fix
- Balance between responsiveness and completeness
- Consider API rate limits when implementing solution
- User experience is priority - directories should appear populated on first access
- Maintain backward compatibility with existing cache
