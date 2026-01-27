# Design: Lazy Directory Loading Performance Fix

## 1. Problem Analysis

### 1.1 Current Behavior (INCORRECT)
When a directory is accessed for the first time:
1. `OpenDir()` calls `GetChildrenID(id, auth)`
2. `GetChildrenID()` checks in-memory cache (`inode.children`) - **MISS**
3. Returns **empty immediately** and schedules background refresh
4. Background refresh fetches from API (5-10 seconds)
5. Next access shows contents from cache

**This is wrong** - users see empty directories!

### 1.2 Desired Behavior (CORRECT)
When a directory is accessed:
1. `OpenDir()` calls `GetChildrenID(id, auth)`
2. Check in-memory cache - if present, return immediately
3. If not cached, **block and fetch from API synchronously**
4. Return complete data (NEVER empty)
5. Cache for subsequent fast access
6. On subsequent access, serve from cache immediately
7. If cache is stale, serve stale data and refresh in background

**Key principle**: NEVER return empty - always wait for data or serve stale cache.

### 1.3 Root Cause
The current implementation has a **fundamental design flaw**:
- It prioritizes "responsiveness" (return quickly) over "correctness" (return complete data)
- The test `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` codifies this wrong behavior
- This creates terrible UX where directories appear empty on first access

## 2. Solution Design

### 2.1 Recursive Prefetch on Mount (Primary Solution)
Implement recursive metadata prefetch to populate cache before user access:

```go
func (f *Filesystem) StartPrefetch() {
    // Start prefetch in background after mount
    go f.prefetchRecursive(f.root, 0)
}

func (f *Filesystem) prefetchRecursive(dirID string, depth int) {
    // Prevent infinite recursion (safety limit)
    if depth > 100 {
        logging.Warn().Int("depth", depth).Msg("Prefetch depth limit reached")
        return
    }
    
    // Fetch children with low priority (don't interfere with user operations)
    children, err := f.metadataRequestManager.FetchChildrenSync(
        dirID, 
        f.auth, 
        30*time.Second, // Longer timeout for background operation
        PriorityBackground,
    )
    if err != nil {
        logging.Debug().Err(err).Str("dirID", dirID).Msg("Prefetch failed for directory")
        return
    }
    
    // Cache the results
    f.cacheChildren(dirID, children)
    
    // Recursively prefetch subdirectories
    for _, child := range children {
        if child.IsDir() {
            f.prefetchRecursive(child.ID(), depth+1)
        }
        // Note: We do NOT prefetch file contents, only metadata
    }
}
```

### 2.2 GetChildrenID with Prefetch Awareness
Update `GetChildrenID()` to work with prefetch:

```go
func (f *Filesystem) GetChildrenID(id string, auth *graph.Auth) (map[string]*Inode, error) {
    // Check in-memory cache first
    if cachedChildren := f.getCachedChildren(id); cachedChildren != nil {
        // Cache hit - check freshness
        if f.isCacheFresh(id) {
            // Fresh cache - return immediately
            return cachedChildren, nil
        }
        
        // Cache is stale - BLOCK and refresh with timeout
        refreshed, err := f.refreshChildrenWithTimeout(id, auth, 2*time.Second)
        if err == nil {
            // Refresh succeeded - return fresh data
            return refreshed, nil
        }
        
        // Refresh timed out or failed - serve stale data as fallback
        // Continue refresh in background
        go f.refreshChildrenAsync(id, auth)
        return cachedChildren, nil
    }
    
    // Cache miss - check if prefetch is in progress
    if f.isPrefetchInProgress(id) {
        // Wait for prefetch to complete (with timeout)
        if children := f.waitForPrefetch(id, 5*time.Second); children != nil {
            return children, nil
        }
    }
    
    // Not cached and not being prefetched - BLOCK and fetch synchronously
    return f.fetchChildrenSync(id, auth)
}

func (f *Filesystem) isPrefetchInProgress(id string) bool {
    // Check metadata state - if HYDRATING, prefetch is in progress
    entry, err := f.metadataStore.Get(context.Background(), id)
    if err != nil {
        return false
    }
    return entry.State == metadata.ItemStateHydrating
}

func (f *Filesystem) waitForPrefetch(id string, timeout time.Duration) map[string]*Inode {
    // Poll cache until data appears or timeout
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if children := f.getCachedChildren(id); children != nil {
            return children
        }
        time.Sleep(100 * time.Millisecond)
    }
    return nil
}
```

### 2.3 File Content Loading (Separate from Prefetch)
File contents are loaded on-demand when opened:

```go
func (f *Filesystem) Open(cancel <-chan struct{}, in *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
    inode := f.GetNodeID(in.NodeId)
    if inode == nil {
        return fuse.ENOENT
    }
    
    // Check if content is already cached
    if f.content.Has(inode.ID()) {
        return fuse.OK
    }
    
    // Content not cached - BLOCK and download
    // Use download manager with foreground priority
    session := f.downloads.QueueDownload(inode.ID())
    err := f.downloads.WaitForDownload(inode.ID(), 60*time.Second)
    if err != nil {
        return fuse.EIO
    }
    
    return fuse.OK
}
```

**Key points**:
- Prefetch only fetches metadata (directory listings, file names, sizes, etc.)
- File contents are downloaded on-demand when file is opened
- Both operations block until complete (NEVER return empty/partial data)

### 2.4 Metadata State Tracking
Use existing metadata states to track prefetch progress:

- **GHOST**: Metadata known but not yet fetched
- **HYDRATING**: Prefetch in progress
- **HYDRATED**: Metadata cached and available
- **ERROR**: Prefetch failed

This allows `GetChildrenID()` to know if it should wait for prefetch or fetch synchronously.

### 2.5 Priority Management
- **Prefetch**: Low priority (background)
- **User access**: High priority (foreground)
- **Stale refresh**: Medium priority

This ensures user operations are never blocked by prefetch.

### 2.6 Test Updates
Update tests to reflect new behavior:

```go
func TestIT_FS_Cache_PrefetchRecursive(t *testing.T) {
    // Test that prefetch recursively fetches all directories
    // Verify metadata is cached
    // Verify file contents are NOT prefetched
}

func TestIT_FS_Cache_GetChildrenIDReturnsPrefetchedData(t *testing.T) {
    // Test that GetChildrenID returns immediately if prefetched
    // Should be < 50ms
}

func TestIT_FS_Cache_GetChildrenIDWaitsForPrefetch(t *testing.T) {
    // Test that GetChildrenID waits if prefetch in progress
    // Should wait up to 5 seconds
}

func TestIT_FS_Cache_GetChildrenIDBlocksIfNotPrefetched(t *testing.T) {
    // Test that GetChildrenID blocks and fetches if not prefetched
    // Should block up to 10 seconds
    // NEVER returns empty
}

func TestIT_FS_Cache_FileOpenBlocksUntilDownloaded(t *testing.T) {
    // Test that file open blocks until content downloaded
    // Should block up to 60 seconds
    // NEVER returns partial/empty file
}
```

## 4. Design Constraints

### 4.1 Must Maintain
- Existing metadata state machine (GHOST, HYDRATING, HYDRATED, etc.)
- Existing metadata request prioritization
- Existing stale-cache policy design
- Backward compatibility with existing metadata store
- All existing tests must pass

### 4.2 Must Not Change
- Fundamental lazy-loading architecture
- FUSE operation contracts
- Metadata state transitions
- API request patterns

## 5. Testing Strategy

### 5.1 Unit Tests
- Test `persistMetadataEntry()` saves children correctly
- Test `tryPopulateChildrenFromMetadata()` restores children correctly
- Test metadata store queries return correct data
- Test state transitions persist children

### 5.2 Integration Tests
- Test directory listing after delta sync
- Test directory listing after cold start
- Test directory listing with metadata store populated
- Test background refresh updates metadata store

### 5.3 System Tests
- Test real directory access patterns
- Test file manager behavior
- Test performance under load
- Test with large directories (1000+ items)

## 6. Performance Considerations

### 6.1 Targets
- First directory access: < 100ms (from metadata store)
- Subsequent access: < 50ms (from memory cache)
- Metadata store query: < 10ms
- Background refresh: 5-10 seconds (acceptable)

### 6.2 Optimizations
- Index metadata store by parent_id for fast children queries
- Batch metadata store updates during delta sync
- Use transactions for atomic updates
- Cache metadata store queries in memory

## 7. Migration Strategy

### 7.1 Backward Compatibility
- Existing metadata store format must be supported
- Migration path for existing deployments
- Graceful handling of missing children data

### 7.2 Migration Steps
1. Detect old metadata store format
2. Populate children during next delta sync
3. Verify children are present
4. Enable new restoration logic

## 8. Monitoring and Metrics

### 8.1 Metrics to Track
- Metadata store hit rate
- Metadata store query latency
- Directory listing latency
- Background refresh latency
- API request count

### 8.2 Logging
- Log metadata store population
- Log metadata store restoration
- Log cache hits/misses
- Log performance metrics

## 9. Rollback Plan

### 9.1 If Fix Causes Issues
- Revert code changes
- Restore previous behavior
- Investigate root cause
- Re-test before re-deploying

### 9.2 Safety Measures
- Feature flag for new behavior
- Gradual rollout
- Monitor metrics closely
- Quick rollback capability

## 10. Next Steps

1. **Investigation**: Run Phase 1-2 investigation
2. **Root Cause**: Identify specific root cause
3. **Solution Design**: Design specific fix based on root cause
4. **Implementation**: Implement fix with tests
5. **Validation**: Verify fix resolves issue
6. **Documentation**: Update documentation

## 11. Open Questions

1. Is the metadata store being populated with children during delta sync?
2. Is `tryPopulateChildrenFromMetadata()` being called correctly?
3. Is the metadata store query returning correct data?
4. Are state transitions persisting children correctly?
5. Is there a performance issue with metadata store queries?

## 12. References

- Requirements: `.kiro/specs/lazy-directory-loading-fix/requirements.md`
- ADR-001: `docs/2-architecture/decisions/ADR-001-structured-metadata-store.md`
- ADR-003: `docs/2-architecture/decisions/ADR-003-metadata-request-prioritization.md`
- Implementation: `internal/fs/cache.go`
- Tests: `internal/fs/cache_test.go`
