# Tasks: Cache Management

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

## Phase 9: Cache Management Verification

- [x] 11. Verify cache management
- [x] 11.1 Review cache code
  - Read and analyze `internal/fs/cache.go`
  - Review `internal/fs/content_cache.go`
  - Check cache cleanup implementation
  - Review bbolt database usage
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 11.2 Test content caching
  - Access several files
  - Verify content is stored in cache directory
  - Check cache directory structure
  - Verify cached content is correct
  - _Requirements: 7.1_

- [x] 11.3 Test cache hit/miss
  - Access a cached file (should be cache hit)
  - Access an uncached file (should be cache miss)
  - Verify cache statistics reflect hits and misses
  - _Requirements: 7.5_

- [x] 11.4 Test cache expiration with manual verification
  - Configure short cache expiration (e.g., 1 day)
  - Create files with old access times
  - Trigger cache cleanup
  - Verify old files are removed
  - Verify recent files are retained
  - **Retest**: Perform manual cache management verification in Docker
  - Set short cache expiration time in configuration
  - Access multiple files to populate cache
  - Monitor cache cleanup process
  - Verify cache statistics with large datasets
  - Test with different cache size limits
  - Run in Docker: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`
  - Document results in `docs/verification-tracking.md` Phase 9 section
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 11.5 Test cache statistics
  - Run `onemount --stats /mount/path`
  - Verify statistics show cache size
  - Check file count
  - Verify hit rate calculation
  - _Requirements: 7.5_

- [x] 11.6 Test metadata cache persistence
  - Access files to populate metadata cache
  - Unmount filesystem
  - Remount filesystem
  - Verify metadata is still cached (no refetch)
  - _Requirements: 7.1_

- [x] 11.7 Create cache management integration tests
  - Write test for cache storage and retrieval
  - Write test for cache expiration
  - Write test for cache cleanup
  - Write test for cache statistics
  - _Requirements: 7.1, 7.2, 7.3, 7.5_

- [x] 11.8 Implement cache management property-based tests
- [x] 11.8.1 Implement Property 28: ETag-Based Cache Storage
  - **Property 28: ETag-Based Cache Storage**
  - **Validates: Requirements 7.1**
  - Create `internal/fs/cache_property_test.go`
  - Generate random downloaded file scenarios
  - Verify content stored in cache directory with file's ETag
  - Test ETag association accuracy
  - _Requirements: 7.1_

- [x] 11.8.2 Implement Property 29: Cache Invalidation on Remote ETag Change
  - **Property 29: Cache Invalidation on Remote ETag Change**
  - **Validates: Requirements 7.3**
  - Generate random cached files with different remote ETags
  - Verify cache invalidation and new version download
  - Test invalidation trigger accuracy
  - _Requirements: 7.3_

- [x] 11.9 Document cache issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 16: ETag Cache Validation Verification ✅ COMPLETE

- [x] 29. Verify ETag-based cache validation with real OneDrive ✅ COMPLETE
- [x] 29.1 Review ETag implementation ✅ COMPLETE
- [x] 29.2 Test cache hit with valid ETag using real OneDrive ✅ COMPLETE
- [x] 29.3 Test cache miss with changed ETag using real OneDrive ✅ COMPLETE
- [x] 29.4 Test ETag updates from delta sync using real OneDrive ✅ COMPLETE
- [x] 29.5 Test conflict detection with ETags using real OneDrive ✅ COMPLETE
- [x] 29.6 Run ETag validation integration tests with real OneDrive ✅ COMPLETE

**Status**: ✅ **COMPLETE** (2025-11-13)  
**Requirements**: 3.4-3.6, 7.1-7.4, 8.1-8.3 all verified  
**Results**: ETag-based cache validation working correctly with real OneDrive API. All integration tests passing.

---

## Phase 17: State Management Verification

- [x] 30. Verify metadata state model implementation
- [x] 30.1 Review state model implementation
  - Read and analyze `internal/fs/state_manager.go`
  - Review `internal/fs/hydration.go` state transitions
  - Check state persistence in metadata database
  - Review state transition diagram implementation
  - Verify all 7 states are implemented (GHOST, HYDRATING, HYDRATED, DIRTY_LOCAL, DELETED_LOCAL, CONFLICT, ERROR)
  - _Requirements: 21.1-21.10_

- [x] 30.2 Test initial item state assignment
  - Test items discovered via delta are inserted with GHOST state
  - Verify no content download until required
  - Test state persistence in metadata database
  - Verify virtual entries use correct state and flags
  - _Requirements: 21.2, 21.10_

- [x] 30.3 Test hydration state transitions
  - Test GHOST → HYDRATING transition on user access
  - Test HYDRATING → HYDRATED transition on successful download
  - Test HYDRATING → ERROR transition on download failure
  - Test HYDRATING → GHOST transition on cancellation
  - Verify worker deduplication during hydration
  - _Requirements: 21.3, 21.4, 21.5_

- [x] 30.4 Test modification and upload state transitions
  - Test HYDRATED → DIRTY_LOCAL transition on local modification
  - Test DIRTY_LOCAL → HYDRATED transition on successful upload
  - Test DIRTY_LOCAL → ERROR transition on upload failure
  - Test ETag updates after successful upload
  - _Requirements: 21.6_

- [x] 30.5 Test deletion state transitions
  - Test HYDRATED → DELETED_LOCAL transition on local delete
  - Test DELETED_LOCAL → [REMOVED] transition on server confirmation
  - Test DELETED_LOCAL → CONFLICT transition on remote modification
  - Verify tombstone handling
  - _Requirements: 21.7_

- [x] 30.6 Test conflict state transitions
  - Test DIRTY_LOCAL → CONFLICT transition on remote changes
  - Test CONFLICT → HYDRATED transition on conflict resolution
  - Test CONFLICT → GHOST transition on local version deletion
  - Verify both versions are preserved during conflict
  - _Requirements: 21.8_

- [x] 30.7 Test eviction and error recovery
  - Test HYDRATED → GHOST transition on cache eviction
  - Test ERROR → HYDRATING transition on retry
  - Test ERROR → DIRTY_LOCAL transition on upload retry
  - Test ERROR → GHOST transition on error clearing
  - _Requirements: 21.9_

- [x] 30.8 Test virtual file state handling
  - Test virtual entries have item_state=HYDRATED
  - Test virtual entries have remote_id=NULL and is_virtual=TRUE
  - Verify virtual entries bypass sync/upload logic
  - Test virtual entries participate in directory listings
  - _Requirements: 21.10_

- [x] 30.9 Test state transition atomicity and consistency
  - Test state transitions are atomic
  - Test no intermediate inconsistent states
  - Test state persistence across restarts
  - Test concurrent state transition safety
  - _Requirements: 21.1-21.10_

- [x] 30.10 Create state model integration tests
  - Write test for complete state lifecycle
  - Write test for state transition edge cases
  - Write test for state persistence and recovery
  - Write test for concurrent state operations
  - _Requirements: 21.1-21.10_

- [x] 30.11 Implement metadata state model property-based tests
- [x] 30.11.1 Implement Property 40: Initial Item State
  - **Property 40: Initial Item State**
  - **Validates: Requirements 21.2**
  - Create `internal/fs/state_property_test.go`
  - Generate random drive items discovered via delta
  - Verify items are inserted with GHOST state
  - Verify no content download until required
  - _Requirements: 21.2_

- [x] 30.11.2 Implement Property 41: Successful Hydration State Transition
  - **Property 41: Successful Hydration State Transition**
  - **Validates: Requirements 21.4**
  - Generate random successful hydration scenarios
  - Verify transition to HYDRATED state
  - Verify content path recording and metadata updates
  - Verify error field clearing
  - _Requirements: 21.4_

- [x] 30.11.3 Implement Property 42: Local Modification State Transition
  - **Property 42: Local Modification State Transition**
  - **Validates: Requirements 21.6**
  - Generate random locally modified hydrated file scenarios
  - Verify transition to DIRTY_LOCAL state
  - Verify state persists until upload succeeds
  - _Requirements: 21.6_

---
