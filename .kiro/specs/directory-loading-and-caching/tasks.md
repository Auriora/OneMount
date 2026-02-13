# Tasks: Directory Loading And Caching

**References:**
- Requirements: `.kiro/specs/directory-loading-and-caching/requirements.md`
- Design: `.kiro/specs/directory-loading-and-caching/design.md`

## Implementation Status Summary

All core functionality has been implemented and tested. The remaining tasks focus on performance validation, manual testing, and documentation updates.

**Completed:**
- ✅ GetChildrenID blocking behavior with stale cache refresh policy
- ✅ Recursive metadata prefetch infrastructure
- ✅ Prefetch-aware directory access with wait logic
- ✅ File content on-demand loading (separate from metadata)
- ✅ Comprehensive unit and integration tests

**Remaining:**
- ⏳ Performance validation and benchmarking
- ⏳ Manual testing with file managers
- ⏳ Documentation updates

---

## Phase 1: Fix GetChildrenID Blocking Behavior (Req 1, 3, 4)

- [x] 1. Fix GetChildrenID to Block on Cache Miss
**Addresses:** Requirements 1.1, 1.2, 1.3, 1.4, 3.3
**Status:** ✅ Complete - Implemented in `internal/fs/cache.go`
- [x] 1.1 Remove undefined `syncOnMiss` variable reference in `GetChildrenID()`
- [x] 1.2 Implement synchronous blocking when cache miss occurs (no async return)
- [x] 1.3 Add 10-second timeout for synchronous fetch with error return (not empty)
- [x] 1.4 Ensure NEVER returns empty directory listing when data exists
- [x] 1.5 Update error handling to return proper error status on timeout/failure

[x] 2. Implement Stale Cache Refresh Policy
**Addresses:** Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7
**Status:** ✅ Complete - Implemented in `internal/fs/cache.go`
- [x] 2.1 Create `isCacheFresh()` helper to check cache TTL
- [x] 2.2 Implement 2-second timeout for stale cache refresh attempt
- [x] 2.3 Serve stale data if refresh times out (never return empty)
- [x] 2.4 Continue refresh in background after serving stale data
- [x] 2.5 Add logging for stale cache refresh behavior

[x] 3. Update GetChildrenID Test Expectations
**Addresses:** Requirement 7.1
**Status:** ✅ Complete - Updated in `internal/fs/cache_test.go`
- [x] 3.1 Update `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` to expect blocking (not quick return)
- [x] 3.2 Verify test expects complete data (not empty) on first access
- [x] 3.3 Add timeout expectation (up to 10 seconds for uncached)

## Phase 2: Implement Recursive Prefetch (Req 2)

[x] 4. Create Prefetch Infrastructure
**Addresses:** Requirements 2.1, 2.2, 2.3, 2.4, 2.5
**Status:** ✅ Complete - Implemented in `internal/fs/prefetch.go`
- [x] 4.1 Create `StartPrefetch()` method to initiate background prefetch
- [x] 4.2 Create `prefetchRecursive()` method with depth limit (100 levels)
- [x] 4.3 Use PriorityBackground for all prefetch requests
- [x] 4.4 Fetch metadata only (NOT file contents)
- [x] 4.5 Add logging to track prefetch progress and errors

[x] 5. Integrate Prefetch with Metadata State Machine
**Addresses:** Requirements 2.6, 2.7, 2.8, 2.9
**Status:** ✅ Complete - Integrated in `internal/fs/prefetch.go`
- [x] 5.1 Set metadata state to HYDRATING when prefetch starts
- [x] 5.2 Set metadata state to HYDRATED when prefetch completes
- [x] 5.3 Set metadata state to ERROR on prefetch failure (log and continue)
- [x] 5.4 Persist prefetched metadata to metadata store
- [x] 5.5 Handle graceful degradation on prefetch errors

[x] 6. Call Prefetch After Mount
**Addresses:** Requirements 2.1, 6.5
**Status:** ✅ Complete - Integrated in `internal/fs/cache.go` (NewFilesystem)
- [x] 6.1 Call `StartPrefetch()` in `NewFilesystem()` after initialization
- [x] 6.2 Ensure prefetch runs in goroutine (non-blocking mount)
- [x] 6.3 Verify mount completes quickly (< 2 seconds)
- [x] 6.4 Add integration point in filesystem initialization

## Phase 3: Prefetch-Aware Directory Access (Req 3)

[x] 7. Implement Prefetch Detection Helpers
**Addresses:** Requirements 3.2, 3.4, 3.6
**Status:** ✅ Complete - Implemented in `internal/fs/cache.go`
- [x] 7.1 Create `isPrefetchInProgress()` to check if state is HYDRATING
- [x] 7.2 Create `waitForPrefetch()` to poll cache with 100ms interval
- [x] 7.3 Add 5-second timeout for waiting on prefetch
- [x] 7.4 Fall back to synchronous fetch if prefetch times out

[x] 8. Update GetChildrenID for Prefetch Awareness
**Addresses:** Requirements 3.1, 3.2, 3.3, 3.5
**Status:** ✅ Complete - Implemented in `internal/fs/cache.go`
- [x] 8.1 Check if prefetch in progress before synchronous fetch
- [x] 8.2 Wait for prefetch completion if HYDRATING state detected
- [x] 8.3 Return immediately if data already prefetched (< 50ms)
- [x] 8.4 Block and fetch synchronously if not prefetched
- [x] 8.5 Add logging for prefetch-aware behavior

## Phase 4: File Content Loading (Req 5)

[x] 9. Verify File Content Separation
**Addresses:** Requirements 5.1, 5.2, 5.3
**Status:** ✅ Complete - Verified in `internal/fs/prefetch.go` and `internal/fs/file_operations.go`
- [x] 9.1 Verify prefetch ONLY fetches metadata (not file contents)
- [x] 9.2 Add code comments documenting metadata vs content separation
- [x] 9.3 Add assertions to prevent accidental content prefetch
- [x] 9.4 Verify file content loaded only in `Open()` method

[x] 10. Update File Open to Block Until Downloaded
**Addresses:** Requirements 5.2, 5.3, 5.4, 5.5, 5.6, 5.7
**Status:** ✅ Complete - Implemented in `internal/fs/file_operations.go`
- [x] 10.1 Verify content cache check in `Open()` (already exists)
- [x] 10.2 Ensure download uses PriorityForeground (already exists)
- [x] 10.3 Add blocking wait for download completion (60 second timeout)
- [x] 10.4 Return error on download failure (not partial/empty file)
- [x] 10.5 Verify NEVER returns partial content to user

## Phase 5: Testing (Req 7)

[x] 11. Add Prefetch Tests
**Addresses:** Requirements 7.3, 7.4
**Status:** ✅ Complete - Implemented in `internal/fs/prefetch_test.go`
- [x] 11.1 Test `prefetchRecursive()` fetches all directories recursively
- [x] 11.2 Test prefetch ONLY fetches metadata (not file contents)
- [x] 11.3 Test prefetch handles errors gracefully (log and continue)
- [x] 11.4 Test prefetch respects 100-level depth limit
- [x] 11.5 Test prefetch uses PriorityBackground

[x] 12. Add GetChildrenID Blocking Tests
**Addresses:** Requirements 7.2, 7.7
**Status:** ✅ Complete - Implemented in `internal/fs/cache_test.go`
- [x] 12.1 Test returns immediately if prefetched (< 50ms)
- [x] 12.2 Test waits for prefetch if HYDRATING (< 5s)
- [x] 12.3 Test blocks and fetches if not prefetched (< 10s)
- [x] 12.4 Test NEVER returns empty directory listing
- [x] 12.5 Test stale cache refresh with 2-second timeout

[x] 13. Add File Content Tests
**Addresses:** Requirements 7.5
**Status:** ✅ Complete - Verified in existing tests
- [x] 13.1 Test file open blocks until content downloaded
- [x] 13.2 Test file open returns error on download failure
- [x] 13.3 Test file content is NOT prefetched during metadata prefetch
- [x] 13.4 Test multiple concurrent file opens
- [x] 13.5 Test large file downloads with 60-second timeout

[x] 14. Integration Testing
**Addresses:** Requirements 6.1, 6.2, 6.3, 6.4, 6.5
**Status:** ✅ Complete - Verified in integration tests
- [x] 14.1 Test complete mount → prefetch → user access flow
- [x] 14.2 Test cached directory access < 50ms
- [x] 14.3 Test uncached directory access < 10s with timeout
- [x] 14.4 Test stale cache refresh < 2s timeout
- [x] 14.5 Test mount completes quickly (< 2s) with background prefetch

## Phase 6: Performance Validation (Req 6)

- [ ] 15. Performance Testing
**Addresses:** Requirements 6.6, 6.7
**Status:** ⏳ Pending - Requires dedicated performance test suite
**Implementation Notes:**
- Use existing performance test framework in `internal/testutil/framework/performance_test.go`
- Create dedicated performance test for prefetch behavior
- Measure baseline vs. prefetch-enabled metrics
- Document results in test artifacts

**Sub-tasks:**
- [ ] 15.1 Measure memory usage increase (must be < 20%)
  - Create benchmark comparing memory before/after prefetch
  - Use Go's runtime.MemStats to track heap allocation
  - Test with various directory tree sizes (100, 1000, 10000 items)
  
- [ ] 15.2 Measure API request rate (no significant increase)
  - Count Graph API calls during prefetch vs. on-demand loading
  - Verify prefetch doesn't trigger rate limiting
  - Compare total API calls for same user workflow
  
- [ ] 15.3 Measure prefetch time for various directory sizes
  - Small tree: < 100 directories (should complete in seconds)
  - Medium tree: 100-1000 directories (should complete in minutes)
  - Large tree: > 1000 directories (verify depth limit works)
  
- [ ] 15.4 Measure directory access latency after prefetch
  - Verify cached access < 50ms (99th percentile)
  - Compare prefetched vs. non-prefetched access times
  - Test with concurrent directory access
  
- [ ] 15.5 Verify no performance regression in existing operations
  - Run full test suite and compare execution times
  - Verify file operations (read/write) not affected
  - Check mount time remains < 2 seconds

- [ ] 16. Manual Testing
**Addresses:** Requirements 1.6, 6.1, 6.2
**Status:** ⏳ Pending - Requires manual verification with real OneDrive account
**Implementation Notes:**
- Test with real OneDrive account (not mocks)
- Use various file managers to verify behavior
- Document observations in test artifacts

**Sub-tasks:**
- [ ] 16.1 Mount filesystem and verify prefetch starts in background
  - Mount with `onemount` command
  - Check logs for "Prefetch completed" message
  - Verify mount returns immediately (< 2s)
  
- [ ] 16.2 Navigate directories and verify instant access (< 50ms)
  - Use `ls` command on various directories
  - Measure response time with `time ls /path/to/dir`
  - Verify no visible delay after prefetch completes
  
- [ ] 16.3 Open files and verify content downloads on-demand
  - Open files with `cat`, `less`, or text editor
  - Verify download happens on first open (not during prefetch)
  - Check logs for download manager activity
  
- [ ] 16.4 Test with file managers (Nautilus, Dolphin, Thunar)
  - Open mounted directory in Nautilus (GNOME)
  - Open mounted directory in Dolphin (KDE)
  - Open mounted directory in Thunar (XFCE)
  - Verify directories show contents immediately
  - Verify no "empty folder" flicker
  
- [ ] 16.5 Verify directories NEVER appear empty on first access
  - Navigate to uncached directory
  - Verify contents appear (may take up to 10s)
  - Verify no empty state shown
  - Test with slow network conditions

## Phase 7: Documentation

- [ ] 17. Update Documentation
**Status:** ⏳ Pending - Documentation needs to reflect implementation
**Implementation Notes:**
- Update all relevant documentation files
- Ensure consistency across design, ADRs, and code comments
- Create comprehensive fix document

**Sub-tasks:**
- [ ] 17.1 Document prefetch behavior in code comments
  - Add detailed comments to `StartPrefetch()` and `prefetchRecursive()`
  - Document prefetch state transitions
  - Explain depth limit and batching strategy
  - Document offline behavior and pending prefetch flag
  
- [ ] 17.2 Update design document with implementation details
  - Update `.kiro/specs/directory-loading-and-caching/design.md`
  - Add actual implementation details vs. design
  - Document any deviations from original design
  - Add performance characteristics observed
  
- [ ] 17.3 Create fix document in `docs/fixes/lazy-directory-loading-fix.md`
  - Document the original problem (empty directories)
  - Explain the solution (prefetch + blocking)
  - Include before/after behavior comparison
  - Add troubleshooting section
  
- [ ] 17.4 Update ADR-003 with prefetch strategy details
  - Document prefetch as part of metadata request strategy
  - Explain priority handling (background for prefetch)
  - Document interaction with existing metadata manager
  - Add rationale for design decisions
  
- [ ] 17.5 Document any configuration options added
  - Document prefetch depth limit (100 levels)
  - Document timeout values (10s sync, 2s stale refresh, 5s prefetch wait)
  - Explain when prefetch is skipped (offline mode)
  - Document prefetch pending flag behavior

## Implementation Notes

**Architecture:**
- ✅ Metadata state machine (GHOST, HYDRATING, HYDRATED) - leveraged existing
- ✅ Metadata request manager with priority queuing - leveraged existing
- ✅ Download manager - leveraged existing
- ✅ Prefetch infrastructure - created in `internal/fs/prefetch.go`
- ✅ GetChildrenID blocking behavior - updated in `internal/fs/cache.go`

**Key Files:**
- `internal/fs/prefetch.go` - Prefetch implementation
- `internal/fs/cache.go` - GetChildrenID with blocking and stale cache refresh
- `internal/fs/file_operations.go` - File content on-demand loading
- `internal/fs/prefetch_test.go` - Prefetch unit tests
- `internal/fs/cache_test.go` - GetChildrenID and cache tests

**Performance Characteristics:**
- Prefetch runs in background (non-blocking mount)
- Depth limit prevents runaway recursion (100 levels per batch)
- Background priority ensures user operations not blocked
- Stale cache policy balances freshness and responsiveness
- Synchronous fetch ensures directories never appear empty

**Next Steps:**
1. Run performance benchmarks (Task 15)
2. Conduct manual testing with file managers (Task 16)
3. Update documentation (Task 17)
