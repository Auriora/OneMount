# Design: Lazy Directory Loading Performance Fix

## 1. Problem Analysis

### 1.1 Current Behavior (INCORRECT)
When a directory is accessed for the first time:
1. `OpenDir()` calls `GetChildrenID(id, auth)`
2. `GetChildrenID()` checks in-memory cache (`inode.children`) - **MISS**
3. Returns **empty immediately** and schedules background refresh
4. Background refresh fetches from API (5-10 seconds)
5. Next access shows contents from cache

**This is wrong** - users see empty directories! (Violates Requirement 1)

### 1.2 Desired Behavior (CORRECT)
When a directory is accessed:
1. `OpenDir()` calls `GetChildrenID(id, auth)`
2. Check in-memory cache - if present, return immediately (< 50ms)
3. If not cached, **block and fetch from API synchronously** (up to 10 seconds)
4. Return complete data (NEVER empty)
5. Cache for subsequent fast access
6. On subsequent access, serve from cache immediately
7. If cache is stale, attempt 2-second refresh, then serve stale data if timeout

**Key principle**: NEVER return empty - always wait for data or serve stale cache. (Requirements 1, 4)

### 1.3 Root Cause
The current implementation has a **fundamental design flaw**:
- It prioritizes "responsiveness" (return quickly) over "correctness" (return complete data)
- The test `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` codifies this wrong behavior
- This creates terrible UX where directories appear empty on first access
- No prefetch mechanism exists to populate cache before user access

## 2. Solution Design

### 2.1 Recursive Prefetch on Mount (Primary Solution)
**Addresses: Requirement 2 (Recursive Metadata Prefetch)**

Implement recursive metadata prefetch to populate cache before user access:

```go
func (f *Filesystem) StartPrefetch() {
    // Start prefetch in background after mount (Req 2.1)
    // Non-blocking to ensure mount completes quickly (Req 6.5)
    go f.prefetchRecursive(f.root, 0)
}

func (f *Filesystem) prefetchRecursive(dirID string, depth int) {
    // Prevent runaway recursion in a single batch (safety limit) (Req 2.4)
    // Directories beyond the limit are queued for a later batch so the full
    // tree is eventually covered.
    if depth > 100 {
        enqueueForNextBatch(dirID)
        return
    }
    
    // Set metadata state to HYDRATING (Req 2.7)
    f.setMetadataState(dirID, metadata.ItemStateHydrating)
    
    // Fetch children with low priority (don't interfere with user operations) (Req 2.5)
    children, err := f.metadataRequestManager.FetchChildrenSync(
        dirID, 
        f.auth, 
        30*time.Second,
        PriorityBackground,
    )
    if err != nil {
        // Log error and continue with other directories (Req 2.8)
        logging.Debug().Err(err).Str("dirID", dirID).Msg("Prefetch failed for directory")
        f.setMetadataState(dirID, metadata.ItemStateError)
        return
    }
    
    // Cache the results in memory (metadata only) (Req 2.2)
    f.cacheChildren(dirID, children)
    
    // Persist to metadata store for use across restarts (Req 2.9)
    f.persistMetadataEntry(dirID, children)
    
    // Update metadata state to HYDRATED (Req 2.6)
    f.setMetadataState(dirID, metadata.ItemStateHydrated)
    
    // Recursively prefetch subdirectories (Req 2.3)
    for _, child := range children {
        if child.IsDir() {
            f.prefetchRecursive(child.ID(), depth+1)
        }
        // Note: We do NOT prefetch file contents, only metadata (Req 2.2, 5.1)
    }
}

// Design Decision: Use existing metadata state machine
// Rationale: Leverages existing GHOST/HYDRATING/HYDRATED states to track
// prefetch progress without introducing new state management complexity.
// This allows GetChildrenID to determine if prefetch is in progress.
```
```

### 2.2 GetChildrenID with Prefetch Awareness
**Addresses: Requirements 1, 3, 4 (Never Empty, Prefetch-Aware Access, Stale Cache)**

Update `GetChildrenID()` to work with prefetch and implement stale cache policy:

```go
func (f *Filesystem) GetChildrenID(id string, auth *graph.Auth) (map[string]*Inode, error) {
    // Check in-memory cache first
    if cachedChildren := f.getCachedChildren(id); cachedChildren != nil {
        // Cache hit - check freshness (Req 4.1)
        if f.isCacheFresh(id) {
            // Fresh cache - return immediately (< 50ms) (Req 4.1, 6.1)
            return cachedChildren, nil
        }
        
        // Cache is stale - attempt refresh with 2-second timeout (Req 4.2)
        refreshed, err := f.refreshChildrenWithTimeout(id, auth, 2*time.Second)
        if err == nil {
            // Refresh succeeded - return fresh data (Req 4.3)
            return refreshed, nil
        }
        
        // Refresh timed out or failed - serve stale data as fallback (Req 4.4, 4.5, 4.6)
        // Continue refresh in background (Req 4.7)
        go f.refreshChildrenAsync(id, auth)
        return cachedChildren, nil
    }
    
    // Cache miss - check if prefetch is in progress (Req 3.6)
    if f.isPrefetchInProgress(id) {
        // Wait for prefetch to complete (up to 5 seconds) (Req 3.2, 3.4)
        if children := f.waitForPrefetch(id, 5*time.Second); children != nil {
            return children, nil
        }
        // Prefetch timeout - fall back to synchronous fetch (Req 3.5)
    }
    
    // Not cached and not being prefetched - BLOCK and fetch synchronously (Req 1.1, 3.3)
    // Timeout after 10 seconds and return error (not empty) (Req 1.3)
    children, err := f.fetchChildrenSync(id, auth, 10*time.Second)
    if err != nil {
        // Return error status (not empty) (Req 1.4)
        return nil, err
    }
    return children, nil
}

func (f *Filesystem) isPrefetchInProgress(id string) bool {
    // Check metadata state - if HYDRATING, prefetch is in progress (Req 3.6)
    entry, err := f.metadataStore.Get(context.Background(), id)
    if err != nil {
        return false
    }
    return entry.State == metadata.ItemStateHydrating
}

func (f *Filesystem) waitForPrefetch(id string, timeout time.Duration) map[string]*Inode {
    // Poll cache until data appears or timeout (Req 3.4)
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if children := f.getCachedChildren(id); children != nil {
            return children
        }
        time.Sleep(100 * time.Millisecond) // Poll interval (Req 3.4)
    }
    return nil
}

// Design Decision: Stale cache with timeout-based refresh
// Rationale: Balances freshness with responsiveness. Attempts quick refresh
// (2 seconds) but falls back to stale data if refresh is slow. This ensures
// users never see empty directories while still getting fresh data when possible.
// Background refresh ensures next access has fresh data.
```
```

### 2.3 File Content Loading (Separate from Prefetch)
**Addresses: Requirement 5 (File Content On-Demand Loading)**

File contents are loaded on-demand when opened, NOT during prefetch:

```go
func (f *Filesystem) Open(cancel <-chan struct{}, in *fuse.OpenIn, out *fuse.OpenOut) fuse.Status {
    inode := f.GetNodeID(in.NodeId)
    if inode == nil {
        return fuse.ENOENT
    }
    
    // Check if content is already cached (Req 5.2, 5.3)
    if f.content.Has(inode.ID()) {
        return fuse.OK
    }
    
    // Content not cached - BLOCK and download (Req 5.4)
    // Use download manager with foreground priority (Req 5.6)
    session := f.downloads.QueueDownload(inode.ID(), PriorityForeground)
    
    // Block until download completes (60 second timeout) (Req 5.4)
    err := f.downloads.WaitForDownload(inode.ID(), 60*time.Second)
    if err != nil {
        // Return error if download fails (not partial/empty file) (Req 5.5, 5.7)
        return fuse.EIO
    }
    
    return fuse.OK
}

// Design Decision: Separate metadata prefetch from content download
// Rationale: Metadata is small (KB) and benefits from prefetch, while file
// contents can be large (GB) and should only be downloaded when needed.
// This minimizes bandwidth usage and memory consumption while still providing
// fast directory navigation. (Req 5.1)
```

**Key points**:
- Prefetch only fetches metadata (directory listings, file names, sizes, etc.) (Req 2.2)
- File contents are downloaded on-demand when file is opened (Req 5.1)
- Both operations block until complete (NEVER return empty/partial data) (Req 1.2, 5.7)

### 2.6 Mount + Offline Behavior (Implementation Detail)

- **Online mount**: synchronously hydrate root children before returning, then launch recursive prefetch in a goroutine.
- **Offline mount**: only proceed if cached root children exist; otherwise fail mount.
- **Offline → online transition**: resume prefetch when connectivity returns.

### 2.4 Metadata State Tracking
**Addresses: Requirement 2.6, 2.7, 3.6**

Use existing metadata states to track prefetch progress:

- **GHOST**: Metadata known but not yet fetched
- **HYDRATING**: Prefetch in progress (Req 2.7)
- **HYDRATED**: Metadata cached and available (Req 2.6)
- **ERROR**: Prefetch failed

This allows `GetChildrenID()` to know if it should wait for prefetch or fetch synchronously (Req 3.6).

**Design Decision: Reuse existing metadata state machine**
**Rationale**: The existing state machine already tracks metadata lifecycle. By using HYDRATING state for prefetch, we avoid introducing new state management complexity and leverage existing infrastructure. This also ensures consistency with other metadata operations like delta sync.

### 2.5 Priority Management
**Addresses: Requirement 2.5, 5.6**

- **Prefetch**: Low priority (PriorityBackground) - doesn't interfere with user operations (Req 2.5)
- **User access**: High priority (PriorityForeground) - immediate response
- **Stale refresh**: Medium priority - balance between freshness and responsiveness
- **File downloads**: High priority (PriorityForeground) when user opens file (Req 5.6)

This ensures user operations are never blocked by prefetch.

**Design Decision: Leverage existing priority system**
**Rationale**: The existing metadata request manager (ADR-003) already implements priority-based request queuing. By using PriorityBackground for prefetch, we ensure it doesn't interfere with user-initiated operations while still populating the cache proactively.

### 2.6 Test Updates
**Addresses: Requirement 7 (Test Updates and Validation)**

Update tests to reflect new behavior:

```go
func TestIT_FS_Cache_PrefetchRecursive(t *testing.T) {
    // Test that prefetch recursively fetches all directories (Req 7.3)
    // Verify metadata is cached
    // Verify file contents are NOT prefetched (Req 7.4)
}

func TestIT_FS_Cache_GetChildrenIDReturnsPrefetchedData(t *testing.T) {
    // Test that GetChildrenID returns immediately if prefetched (Req 3.1, 6.2)
    // Should be < 50ms
}

func TestIT_FS_Cache_GetChildrenIDWaitsForPrefetch(t *testing.T) {
    // Test that GetChildrenID waits if prefetch in progress (Req 3.2, 7.7)
    // Should wait up to 5 seconds
}

func TestIT_FS_Cache_GetChildrenIDBlocksIfNotPrefetched(t *testing.T) {
    // Test that GetChildrenID blocks and fetches if not prefetched (Req 1.1, 3.3)
    // Should block up to 10 seconds
    // NEVER returns empty (Req 1.2, 7.2)
}

func TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached(t *testing.T) {
    // UPDATE: This test needs to be modified (Req 7.1)
    // Old behavior: Expected quick return with empty
    // New behavior: Expected blocking until data fetched (up to 10s)
}

func TestIT_FS_Cache_StaleCacheRefreshWithTimeout(t *testing.T) {
    // Test stale cache refresh with 2-second timeout (Req 4.2, 7.6)
    // Verify serves stale data if refresh times out (Req 4.4)
    // Verify background refresh continues (Req 4.7)
}

func TestIT_FS_Cache_FileOpenBlocksUntilDownloaded(t *testing.T) {
    // Test that file open blocks until content downloaded (Req 5.4, 7.5)
    // Should block up to 60 seconds
    // NEVER returns partial/empty file (Req 5.7)
}

// Design Decision: Update existing test rather than delete
// Rationale: TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached codifies
// the wrong behavior. Rather than delete it, we update it to verify the correct
// blocking behavior. This maintains test coverage while fixing the expectation.
```
```

## 3. Design Rationale

### 3.1 Why Recursive Prefetch?
**Problem**: Directories appear empty on first access because data isn't cached.
**Solution**: Proactively fetch all directory metadata after mount.
**Rationale**: 
- Most users navigate through directory trees sequentially
- Metadata is small (KB per directory) compared to file contents (MB-GB)
- Background prefetch doesn't delay mount operation (Req 6.5)
- Prefetched data persists across restarts via metadata store (Req 2.9)
- Addresses Requirement 2 completely

### 3.2 Why Block on Cache Miss?
**Problem**: Current implementation returns empty immediately.
**Solution**: Block and fetch synchronously when data not cached.
**Rationale**:
- Correctness over speed: Better to wait 5-10 seconds than show empty directory
- Users expect directories to show contents on first access
- File managers (Nautilus, Dolphin) expect complete data
- Timeout (10s) prevents indefinite hangs (Req 1.3)
- Error return (not empty) on failure (Req 1.4)
- Addresses Requirement 1 completely

### 3.3 Why Stale Cache with Timeout?
**Problem**: Need balance between freshness and responsiveness.
**Solution**: Attempt 2-second refresh, fall back to stale data if timeout.
**Rationale**:
- Most refreshes complete quickly (< 2s)
- Stale data is better than empty or slow response
- Background refresh ensures next access has fresh data (Req 4.7)
- Never returns empty when stale cache available (Req 4.6)
- Addresses Requirement 4 completely

### 3.4 Why Separate Metadata and Content?
**Problem**: Prefetching file contents wastes bandwidth and memory.
**Solution**: Prefetch metadata only, download content on-demand.
**Rationale**:
- Metadata is small (KB), content can be large (GB)
- Users don't open every file, but do navigate directories
- Bandwidth efficiency: Only download what's needed
- Memory efficiency: Don't cache unused file contents
- Addresses Requirement 5 completely

### 3.5 Why Use Existing State Machine?
**Problem**: Need to track prefetch progress.
**Solution**: Reuse existing GHOST/HYDRATING/HYDRATED states.
**Rationale**:
- Avoids introducing new state management complexity
- Consistent with existing metadata operations (delta sync)
- Leverages existing infrastructure (metadata store, state transitions)
- No changes to fundamental architecture (out of scope)
- Addresses Requirements 2.6, 2.7, 3.6

### 3.6 Why Priority-Based Queuing?
**Problem**: Prefetch shouldn't block user operations.
**Solution**: Use PriorityBackground for prefetch, PriorityForeground for user access.
**Rationale**:
- Existing metadata request manager (ADR-003) supports priorities
- Background prefetch runs when system is idle
- User operations always take precedence
- No changes to existing priority system (out of scope)
- Addresses Requirements 2.5, 5.6

## 4. Design Constraints

### 4.1 Must Maintain
- Existing metadata state machine (GHOST, HYDRATING, HYDRATED, etc.) - no modifications (out of scope)
- Existing metadata request prioritization (foreground/background) - no modifications (out of scope)
- Existing stale-cache policy design - enhanced with timeout-based refresh
- Backward compatibility with existing metadata store
- All existing tests must pass (Req 7.8)

### 4.2 Must Not Change
- Fundamental lazy-loading architecture
- FUSE operation contracts
- Metadata state transitions
- API request patterns (no significant increase in rate) (Req 6.7)

### 4.3 Performance Constraints
- Memory usage increase < 20% (Req 6.6)
- Cached directory access < 50ms (Req 6.1, 6.2)
- Uncached directory access < 10 seconds (Req 6.3)
- Stale cache refresh timeout: 2 seconds (Req 6.4)
- Mount operation completes quickly (< 2 seconds) (Req 6.5)

### 4.4 Out of Scope (Explicitly Excluded)
- Prefetching file contents (only metadata)
- Selective/partial prefetch (all directories prefetched)
- User configuration of prefetch behavior
- Progress indication for prefetch
- Prefetch cancellation
- Changes to metadata state machine
- Changes to priority system

## 5. Testing Strategy

### 5.1 Unit Tests
**Addresses: Requirement 7 (Test Updates and Validation)**

- Test `prefetchRecursive()` fetches all directories (Req 7.3)
- Test prefetch only fetches metadata, not file contents (Req 7.4)
- Test `isPrefetchInProgress()` checks metadata state correctly (Req 3.6)
- Test `waitForPrefetch()` polls cache with 100ms interval (Req 3.4)
- Test `isCacheFresh()` checks TTL correctly (Req 4.1)
- Test `refreshChildrenWithTimeout()` respects 2-second timeout (Req 4.2)
- Test state transitions (GHOST → HYDRATING → HYDRATED) (Req 2.6, 2.7)
- Test error handling for prefetch failures (Req 2.8)

### 5.2 Integration Tests
**Addresses: Requirement 7 (Test Updates and Validation)**

- Test `GetChildrenID()` returns immediately if prefetched (< 50ms) (Req 3.1, 7.3)
- Test `GetChildrenID()` waits for prefetch in progress (up to 5s) (Req 3.2, 7.7)
- Test `GetChildrenID()` blocks if not prefetched (up to 10s) (Req 3.3, 7.2)
- Test `GetChildrenID()` NEVER returns empty (Req 1.2, 7.2)
- Test stale cache refresh with timeout (Req 4.2, 7.6)
- Test file open blocks until content downloaded (Req 5.4, 7.5)
- Test prefetch persists to metadata store (Req 2.9)
- Update `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` to expect blocking (Req 7.1)

### 5.3 System Tests
**Addresses: Requirement 6 (Performance Targets)**

- Test real directory access patterns with file managers
- Test performance under load (1000+ directories)
- Test memory usage increase < 20% (Req 6.6)
- Test API request rate doesn't significantly increase (Req 6.7)
- Test mount operation completes quickly (< 2s) (Req 6.5)
- Test with slow network conditions
- Test with network errors during prefetch

## 6. Performance Considerations

### 6.1 Targets
**Addresses: Requirement 6 (Performance Targets)**

- Cached directory access: < 50ms (Req 6.1, 6.2)
- Uncached directory access: < 10 seconds with timeout (Req 6.3)
- Stale cache refresh timeout: 2 seconds (Req 6.4)
- Mount operation: < 2 seconds (prefetch runs in background) (Req 6.5)
- Memory usage increase: < 20% compared to current (Req 6.6)
- API request rate: No significant increase (Req 6.7)

### 6.2 Optimizations
- Use PriorityBackground for prefetch to avoid blocking user operations (Req 2.5)
- Persist prefetched metadata to metadata store for reuse across restarts (Req 2.9)
- Poll cache every 100ms when waiting for prefetch (Req 3.4)
- Implement depth limit (100 levels) to prevent infinite recursion (Req 2.4)
- Use existing priority system for request management (ADR-003)

### 6.3 Risk Mitigation
**Addresses: Requirements document Risks section**

- **Large directory trees**: Depth limit prevents excessive recursion
- **API rate limiting**: Background priority and error handling prevent throttling
- **Network failures**: Graceful degradation - prefetch errors logged but don't block mount
- **Memory usage**: Monitor and stay within 20% increase constraint
- **Mount delay perception**: Prefetch runs in background, mount completes quickly

## 7. Implementation Dependencies

### 7.1 Internal Dependencies
**From Requirements document**

- **Cache Management Spec**: Prefetch populates cache and metadata store
- **File Download and Hydration Spec**: File content loading uses existing download manager (Req 5.6)
- **Delta Sync Spec**: Delta sync may populate metadata store, reducing prefetch work

### 7.2 External Dependencies
**From Requirements document**

- **Microsoft Graph API**: Source of directory metadata and file information
- **BBolt Metadata Store**: Persistence layer for prefetched metadata (ADR-001, Req 2.9)
- **Metadata Request Manager**: Handles prioritization of prefetch vs user requests (ADR-003, Req 2.5)

### 7.3 Code Dependencies
**From Requirements document**

- `internal/fs/cache.go`: GetChildrenID() implementation
- `internal/fs/fuse_metadata_local_test.go`: Tests for metadata operations
- `internal/fs/cache_test.go`: Tests for cache behavior
- `internal/metadata/store.go`: Metadata persistence
- `internal/fs/download_manager.go`: File content downloads

## 8. Success Metrics

### 8.1 User Experience Metrics
**From Requirements document**

- **Zero empty directories**: No directory ever appears empty on first access (Req 1.2)
- **Fast cached access**: Cached directory access < 50ms (99th percentile) (Req 6.1)
- **Reasonable first access**: Uncached directory access < 10 seconds (with timeout) (Req 6.3)
- **File manager compatibility**: Nautilus, Dolphin, Thunar display directories correctly on first open (Req 1.6)

### 8.2 Performance Metrics
**From Requirements document**

- **Memory usage**: < 20% increase compared to current implementation (Req 6.6)
- **API request rate**: No significant increase in API calls per minute (Req 6.7)
- **Mount time**: Mount operation completes quickly (< 2 seconds), prefetch runs in background (Req 6.5)
- **Cache hit rate**: > 95% of directory accesses served from cache after prefetch completes

### 8.3 Quality Metrics
**From Requirements document**

- **Test coverage**: All new code covered by unit and integration tests (Req 7)
- **Test pass rate**: 100% of tests pass with new implementation (Req 7.8)
- **No regressions**: All existing functionality continues to work (Req 7.8)
- **Error handling**: All error paths tested and logged appropriately (Req 2.8)

## 9. Monitoring and Logging

### 9.1 Metrics to Track
- Prefetch progress (directories fetched, depth reached)
- Prefetch errors and failures
- Cache hit/miss rates
- Directory access latency (cached vs uncached)
- Stale cache refresh success/timeout rates
- API request count and rate
- Memory usage during prefetch

### 9.2 Logging
- Log prefetch start and completion
- Log prefetch errors (Req 2.8)
- Log depth limit reached (Req 2.4)
- Log cache hits/misses
- Log stale cache refresh timeouts (Req 4.4, 4.5)
- Log performance metrics for monitoring

## 10. Rollback and Safety

### 10.1 Rollback Plan
- Feature flag for new behavior (optional)
- Revert code changes if issues arise
- Restore previous behavior
- Investigate root cause
- Re-test before re-deploying

### 10.2 Safety Measures
- Depth limit prevents infinite recursion (Req 2.4)
- Timeouts prevent indefinite hangs (Req 1.3, 3.2, 3.3, 4.2, 5.4)
- Error handling prevents crashes (Req 2.8)
- Background priority prevents blocking user operations (Req 2.5)
- Graceful degradation on prefetch failures (Req 2.8)

## 11. Open Questions and Future Enhancements

### 11.1 Resolved Questions
1. ✅ Should prefetch be recursive? **Yes** - Requirement 2.3
2. ✅ Should file contents be prefetched? **No** - Requirement 5.1
3. ✅ What timeout for synchronous fetch? **10 seconds** - Requirement 1.3
4. ✅ What timeout for stale cache refresh? **2 seconds** - Requirement 4.2
5. ✅ Should we modify metadata state machine? **No** - Out of scope

### 11.2 Future Enhancements (Out of Scope)
- User configuration of prefetch behavior
- Progress indication for prefetch
- Prefetch cancellation
- Selective/partial prefetch
- Adaptive prefetch based on usage patterns

## 12. Migration Strategy

### 12.1 Backward Compatibility
- Existing metadata store format must be supported
- Migration path for existing deployments
- Graceful handling of missing prefetch data
- No breaking changes to existing APIs

### 12.2 Deployment Steps
1. Deploy code with prefetch implementation
2. Monitor prefetch progress and errors
3. Verify directory access behavior (no empty directories)
4. Monitor performance metrics (memory, API rate)
5. Verify all tests pass (Req 7.8)
6. Rollback if issues detected

## 13. Next Steps

1. **Phase 1: Implement Prefetch Infrastructure** (Tasks 1-2)
   - Create `StartPrefetch()` and `prefetchRecursive()` methods
   - Integrate with filesystem initialization
   - Implement metadata state tracking

2. **Phase 2: Update GetChildrenID** (Tasks 3-4)
   - Implement prefetch-aware logic
   - Add stale cache refresh with timeout
   - Implement helper methods

3. **Phase 3: File Content Loading** (Tasks 5-6)
   - Verify file content is NOT prefetched
   - Update file open to block until downloaded

4. **Phase 4: Testing** (Tasks 7-10)
   - Add prefetch tests
   - Update GetChildrenID tests
   - Add file content tests
   - Integration and system testing

5. **Phase 5: Validation** (Tasks 11-12)
   - Performance testing
   - Manual testing with file managers

6. **Phase 6: Documentation** (Task 13)
   - Update code comments
   - Create fix document
   - Update ADRs

## 14. References

### Requirements
- Requirements: `.kiro/specs/lazy-directory-loading-fix/requirements.md`
- All 7 requirements addressed in this design

### Architecture Decisions
- ADR-001: Structured Metadata Store (`docs/2-architecture/decisions/ADR-001-structured-metadata-store.md`)
- ADR-003: Metadata Request Prioritization (`docs/2-architecture/decisions/ADR-003-metadata-request-prioritization.md`)

### Implementation
- Implementation: `internal/fs/cache.go`
- Tests: `internal/fs/cache_test.go`
- Metadata: `internal/metadata/store.go`
- Downloads: `internal/fs/download_manager.go`

### Issue Tracking
- Issue: Lazy Directory Loading Performance (`docs/issues/lazy-directory-loading-performance.md`)

### Test Files
- `internal/fs/cache_test.go`: Cache behavior tests
- `internal/fs/fuse_metadata_local_test.go`: Metadata operation tests
- Test to update: `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` (Req 7.1)
