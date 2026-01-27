# Tasks: Lazy Directory Loading Performance Fix

## Phase 1: Implement Recursive Prefetch

### 1. Implement Prefetch Infrastructure
- [ ] 1.1 Create `StartPrefetch()` method to initiate prefetch on mount
- [ ] 1.2 Create `prefetchRecursive()` method to recursively fetch directory metadata
- [ ] 1.3 Add depth limit (100 levels) to prevent infinite recursion
- [ ] 1.4 Use low priority (PriorityBackground) for prefetch requests
- [ ] 1.5 Add logging to track prefetch progress

### 2. Integrate Prefetch with Filesystem Initialization
- [ ] 2.1 Call `StartPrefetch()` after filesystem mount completes
- [ ] 2.2 Ensure prefetch runs in background (non-blocking)
- [ ] 2.3 Handle errors gracefully (log and continue)
- [ ] 2.4 Track prefetch state using metadata states (GHOST → HYDRATING → HYDRATED)
- [ ] 2.5 Persist prefetched metadata to metadata store

### 3. Update GetChildrenID for Prefetch Awareness
- [ ] 3.1 Check cache first - return immediately if fresh
- [ ] 3.2 If stale, attempt refresh with 2 second timeout
- [ ] 3.3 If cache miss, check if prefetch in progress (`isPrefetchInProgress()`)
- [ ] 3.4 If prefetch in progress, wait for it to complete (`waitForPrefetch()`)
- [ ] 3.5 If not prefetched, block and fetch synchronously

### 4. Implement Prefetch Helper Methods
- [ ] 4.1 Create `isPrefetchInProgress()` to check metadata state
- [ ] 4.2 Create `waitForPrefetch()` to poll cache with timeout
- [ ] 4.3 Create `isCacheFresh()` to check cache TTL
- [ ] 4.4 Create `cacheChildren()` to store results
- [ ] 4.5 Add proper error handling for all methods

## Phase 2: File Content Loading

### 5. Ensure File Content is NOT Prefetched
- [ ] 5.1 Verify prefetch only fetches metadata (not content)
- [ ] 5.2 Verify file content is only loaded on file open
- [ ] 5.3 Add assertions to prevent accidental content prefetch
- [ ] 5.4 Document the separation clearly in code comments

### 6. Update File Open to Block Until Downloaded
- [ ] 6.1 Check if content is already cached in `Open()`
- [ ] 6.2 If not cached, queue download with foreground priority
- [ ] 6.3 Block until download completes (60 second timeout)
- [ ] 6.4 Return error if download fails (not partial/empty file)
- [ ] 6.5 Add progress indication if possible

## Phase 3: Testing

### 7. Add Prefetch Tests
- [ ] 7.1 Test `prefetchRecursive()` fetches all directories
- [ ] 7.2 Test prefetch only fetches metadata (not content)
- [ ] 7.3 Test prefetch handles errors gracefully
- [ ] 7.4 Test prefetch respects depth limit
- [ ] 7.5 Test prefetch uses low priority

### 8. Update GetChildrenID Tests
- [ ] 8.1 Test returns immediately if prefetched (< 50ms)
- [ ] 8.2 Test waits for prefetch if in progress (< 5s)
- [ ] 8.3 Test blocks and fetches if not prefetched (< 10s)
- [ ] 8.4 Test NEVER returns empty
- [ ] 8.5 Test stale cache refresh with timeout

### 9. Add File Content Tests
- [ ] 9.1 Test file open blocks until content downloaded
- [ ] 9.2 Test file open returns error on download failure
- [ ] 9.3 Test file content is NOT prefetched
- [ ] 9.4 Test multiple concurrent file opens
- [ ] 9.5 Test large file downloads with timeout

### 10. Integration Testing
- [ ] 10.1 Test complete mount → prefetch → user access flow
- [ ] 10.2 Test with real OneDrive account
- [ ] 10.3 Test with large directory trees (1000+ folders)
- [ ] 10.4 Test with slow network conditions
- [ ] 10.5 Test with network errors during prefetch

## Phase 4: Performance and Validation

### 11. Performance Testing
- [ ] 11.1 Measure prefetch time for various directory sizes
- [ ] 11.2 Measure directory access latency after prefetch
- [ ] 11.3 Measure API usage during prefetch
- [ ] 11.4 Measure memory usage during prefetch
- [ ] 11.5 Verify no performance regression

### 12. Manual Testing
- [ ] 12.1 Mount filesystem and verify prefetch starts
- [ ] 12.2 Navigate directories and verify instant access
- [ ] 12.3 Open files and verify content downloads
- [ ] 12.4 Test with file managers (Nautilus, Dolphin)
- [ ] 12.5 Verify error messages are clear

## Phase 5: Documentation

### 13. Update Documentation
- [ ] 13.1 Document prefetch behavior in code comments
- [ ] 13.2 Update design document with implementation details
- [ ] 13.3 Create fix document in `docs/fixes/`
- [ ] 13.4 Update ADR-003 with prefetch strategy
- [ ] 13.5 Document configuration options (if any)

## Notes

- **Prefetch is critical** - without it, directories will be empty on first access
- **Prefetch only metadata** - file contents are loaded on-demand
- **NEVER return empty** - always block until data is available
- **Use metadata states** - track prefetch progress (GHOST → HYDRATING → HYDRATED)
- **Priority management** - prefetch uses low priority, user operations use high priority
