# Requirements Document: Directory Loading and Caching

## Introduction

This specification defines the directory loading and caching behavior in OneMount. It consolidates the lazy directory loading fix with the initial sync and cache policies from system verification. The goal is to ensure directories never appear empty on first access, recursive metadata prefetch runs in the background without blocking mount, stale cache refresh is bounded, and cache invalidation is scoped to the affected entry only.

OneMount uses a lazy-loading approach with metadata state management (GHOST, HYDRATING, HYDRATED states), metadata request prioritization (foreground vs background queues), a stale-cache policy (attempt short refresh, then serve stale data and continue refresh in background), and a structured metadata store (BBolt database for persistence).

## Glossary

- **OneMount System**: The complete OneDrive filesystem client for Linux including FUSE filesystem, Graph API integration, caching, and UI components
- **Lazy Loading**: Loading data on-demand rather than preloading everything at mount time
- **Prefetch**: Proactively loading directory metadata in the background before user access
- **Recursive Prefetch**: Prefetching directory metadata for all subdirectories starting from root
- **Metadata**: File and directory information (names, sizes, timestamps) without file content
- **File Content**: The actual data/bytes of a file, loaded separately from metadata
- **Cache Miss**: When requested data is not found in the cache and must be fetched from API
- **Cache Hit**: When requested data is found in the cache and returned immediately
- **Stale Cache**: Cached data that has exceeded its time-to-live (TTL) and may be outdated
- **Synchronous Fetch**: Blocking operation that waits for data to be retrieved before returning
- **Asynchronous Fetch**: Non-blocking operation that schedules data retrieval in the background
- **Metadata State**: Current status of metadata (GHOST, HYDRATING, HYDRATED, ERROR)
- **GHOST State**: Metadata is known to exist but not yet fetched
- **HYDRATING State**: Metadata fetch is currently in progress
- **HYDRATED State**: Metadata has been fetched and is cached
- **Metadata Store**: BBolt database for persisting metadata across restarts
- **Request Priority**: Classification of requests as foreground (user-initiated) or background (system-initiated)
- **TTL (Time-To-Live)**: Duration after which cached data is considered stale
- **FUSE**: Filesystem in Userspace - Linux kernel interface for implementing filesystems
- **Microsoft Graph API**: Microsoft's REST API for accessing OneDrive data

## Requirements

### Requirement 1: Never Return Empty Directories

**User Story:** As a Linux user, I want directories to always show their contents on first access so that I don't see empty folders that later populate.

#### Acceptance Criteria

1. WHEN a user accesses a directory for the first time, THE OneMount System SHALL block and fetch directory contents synchronously before returning
2. THE OneMount System SHALL NEVER return an empty directory listing when the directory contains files or subdirectories
3. WHEN fetching directory contents synchronously, THE OneMount System SHALL timeout after 10 seconds and return an error (not empty)
4. WHEN a synchronous fetch fails, THE OneMount System SHALL return an error status to the user (not empty)
5. WHEN a directory is accessed and data is already cached, THE OneMount System SHALL return cached data immediately (< 50ms)
6. WHEN file managers open directories, THE OneMount System SHALL display complete contents on first access

### Requirement 2: Recursive Metadata Prefetch

**User Story:** As a user, I want the system to prefetch directory metadata in the background so that directories are already cached when I navigate to them.

#### Acceptance Criteria

1. WHEN the filesystem is mounted, THE OneMount System SHALL start recursive prefetch from the root directory in the background without blocking mount or interactive operations
2. WHEN prefetching directories, THE OneMount System SHALL fetch metadata only (directory listings, file names, sizes, timestamps) and NOT file contents
3. WHEN prefetching a directory, THE OneMount System SHALL recursively prefetch all subdirectories found
4. THE OneMount System SHALL limit prefetch recursion depth to 100 levels to prevent infinite loops
5. WHEN prefetching, THE OneMount System SHALL use low priority (background) to not interfere with user operations
6. WHEN prefetch completes for a directory, THE OneMount System SHALL update metadata state to HYDRATED
7. WHEN prefetch is in progress for a directory, THE OneMount System SHALL set metadata state to HYDRATING
8. WHEN prefetch fails for a directory, THE OneMount System SHALL log the error and continue with other directories
9. THE OneMount System SHALL persist prefetched metadata to the metadata store for use across restarts

### Requirement 3: Prefetch-Aware Directory Access

**User Story:** As a user, I want the system to use prefetched data when available so that directory access is instant.

#### Acceptance Criteria

1. WHEN a user accesses a directory that has been prefetched, THE OneMount System SHALL return cached data immediately (< 50ms)
2. WHEN a user accesses a directory where prefetch is in progress, THE OneMount System SHALL wait for prefetch to complete (up to 5 seconds)
3. WHEN a user accesses a directory that has not been prefetched, THE OneMount System SHALL block and fetch synchronously (up to 10 seconds)
4. WHEN waiting for prefetch, THE OneMount System SHALL poll the cache every 100ms to check if data is available
5. WHEN prefetch timeout is reached, THE OneMount System SHALL fall back to synchronous fetch
6. THE OneMount System SHALL check metadata state (GHOST, HYDRATING, HYDRATED) to determine if prefetch is in progress

### Requirement 4: Stale Cache Refresh Policy

**User Story:** As a user, I want the system to serve cached data quickly while keeping it fresh in the background so that I get both speed and accuracy.

#### Acceptance Criteria

1. WHEN cached data is fresh (within TTL), THE OneMount System SHALL return it immediately (< 50ms)
2. WHEN cached data is stale (exceeded TTL), THE OneMount System SHALL attempt to refresh it synchronously with a 2 second timeout
3. WHEN stale cache refresh succeeds within timeout, THE OneMount System SHALL return fresh data
4. WHEN stale cache refresh times out, THE OneMount System SHALL serve stale data and continue refresh in background
5. WHEN stale cache refresh fails, THE OneMount System SHALL serve stale data and log the error
6. THE OneMount System SHALL NEVER return empty when stale cache is available
7. WHEN background refresh completes, THE OneMount System SHALL update the cache for next access

### Requirement 5: Scoped Cache Invalidation on Lookup Failure

**User Story:** As a user, I want failed directory lookups to invalidate only the affected entry so that the rest of the cache remains available.

#### Acceptance Criteria

1. WHEN a directory lookup fails (typos, case mismatches, or virtual file handling), THE OneMount System SHALL scope cache invalidation to the affected entry rather than clearing the entire parent directory cache
2. WHEN scoped invalidation occurs, THE OneMount System SHALL preserve unrelated cached entries in the parent directory
3. WHEN invalidating an entry, THE OneMount System SHALL mark it for refresh or revalidation without disrupting other cached children

### Requirement 6: File Content On-Demand Loading

**User Story:** As a user, I want file contents to be loaded only when I open files so that the system doesn't waste bandwidth prefetching data I may not need.

#### Acceptance Criteria

1. THE OneMount System SHALL NOT prefetch file contents during recursive metadata prefetch
2. WHEN a user opens a file, THE OneMount System SHALL check if content is already cached
3. WHEN file content is cached, THE OneMount System SHALL return it immediately
4. WHEN file content is not cached, THE OneMount System SHALL block and download it synchronously (up to 60 seconds)
5. WHEN file download fails, THE OneMount System SHALL return an error (not partial/empty file)
6. THE OneMount System SHALL use the existing download manager with foreground priority for file opens
7. THE OneMount System SHALL NEVER return partial or empty file content to the user

### Requirement 7: Performance Targets

**User Story:** As a user, I want directory access to be fast and responsive so that the filesystem feels native.

#### Acceptance Criteria

1. WHEN accessing a cached directory, THE OneMount System SHALL complete the operation in less than 50ms
2. WHEN accessing an uncached directory with prefetch complete, THE OneMount System SHALL complete in less than 50ms
3. WHEN accessing an uncached directory without prefetch, THE OneMount System SHALL complete within 10 seconds or timeout
4. WHEN refreshing stale cache, THE OneMount System SHALL attempt refresh for up to 2 seconds before serving stale data
5. WHEN prefetching metadata, THE OneMount System SHALL not significantly delay mount operation (mount completes quickly, prefetch runs in background)
6. THE OneMount System SHALL maintain memory usage increase below 20% compared to current implementation
7. THE OneMount System SHALL not significantly increase API request rate to avoid rate limiting

### Requirement 8: Test Updates and Validation

**User Story:** As a developer, I want tests to validate the new behavior so that regressions are caught early.

#### Acceptance Criteria

1. THE OneMount System SHALL update test `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` to expect blocking behavior (not quick return)
2. THE OneMount System SHALL add tests to verify directories are NEVER returned empty
3. THE OneMount System SHALL add tests to verify recursive prefetch fetches all directories
4. THE OneMount System SHALL add tests to verify prefetch only fetches metadata (not file contents)
5. THE OneMount System SHALL add tests to verify file open blocks until content is downloaded
6. THE OneMount System SHALL add tests to verify stale cache refresh with timeout
7. THE OneMount System SHALL add tests to verify prefetch-aware directory access (wait for prefetch in progress)
8. ALL existing tests SHALL pass with the new implementation

## Out of Scope

The following items are explicitly excluded from this specification:

- **Prefetching file contents**: Only metadata (directory listings, file names, sizes) is prefetched; file contents are loaded on-demand
- **Selective prefetch**: All directories are prefetched recursively; selective/partial prefetch is not supported
- **User configuration of prefetch behavior**: Prefetch is automatic with no user configuration options (future enhancement)
- **Metadata state machine changes**: The existing GHOST/HYDRATING/HYDRATED state machine is not modified
- **Metadata request prioritization changes**: The existing foreground/background priority system is not modified
- **Progress indication for prefetch**: No UI feedback for prefetch progress (future enhancement)
- **Prefetch cancellation**: Once started, prefetch runs to completion (future enhancement)
- **Cache eviction and size limits**: See cache-management spec
- **Virtual file policy definitions**: See virtual-file-management spec

## Dependencies

### Internal Dependencies

- **Cache Management Spec**: Prefetch populates cache and metadata store
- **File Download and Hydration Spec**: File content loading uses existing download manager
- **Delta Sync Spec**: Delta sync may populate metadata store, reducing prefetch work
- **Virtual File Management Spec**: Lookup failures and virtual file overlays inform scoped invalidation

### External Dependencies

- **Microsoft Graph API**: Source of directory metadata and file information
- **BBolt Metadata Store**: Persistence layer for prefetched metadata (ADR-001)
- **Metadata Request Manager**: Handles prioritization of prefetch vs user requests (ADR-003)

### Code Dependencies

- `internal/fs/cache.go`: GetChildrenID() implementation
- `internal/fs/fuse_metadata_local_test.go`: Tests for metadata operations
- `internal/fs/cache_test.go`: Tests for cache behavior
- `internal/metadata/store.go`: Metadata persistence
- `internal/fs/download_manager.go`: File content downloads

## Risks and Assumptions

### Risks

1. **Large directory trees**: Prefetching very large directory structures (10,000+ folders) may take significant time and memory
2. **API rate limiting**: Recursive prefetch may trigger Microsoft Graph API rate limits
3. **Network failures**: Prefetch failures may leave gaps in cached metadata
4. **Memory usage**: Caching entire directory tree metadata may increase memory consumption
5. **Mount delay perception**: Users may perceive mount as slow if they try to access directories before prefetch completes

### Assumptions

1. **Metadata store is reliable**: BBolt database correctly persists and retrieves metadata
2. **State transitions work correctly**: GHOST → HYDRATING → HYDRATED transitions are reliable
3. **Priority system works**: Low priority prefetch doesn't block high priority user requests
4. **Delta sync compatibility**: Prefetch and delta sync can coexist without conflicts
5. **Reasonable directory sizes**: Most OneDrive accounts have < 10,000 directories
6. **Network is available**: Prefetch assumes network connectivity at mount time

### Mitigation Strategies

- **Rate limiting**: Implement delays between prefetch requests to avoid API throttling
- **Graceful degradation**: If prefetch fails, fall back to synchronous fetch on user access
- **Memory monitoring**: Track memory usage and implement limits if needed
- **Timeout handling**: Prefetch requests have timeouts to prevent hanging
- **Error recovery**: Prefetch errors are logged but don't prevent mount or user access

## Success Metrics

### User Experience Metrics

- **Zero empty directories**: No directory ever appears empty on first access
- **Fast cached access**: Cached directory access < 50ms (99th percentile)
- **Reasonable first access**: Uncached directory access < 10 seconds (with timeout)
- **File manager compatibility**: Nautilus, Dolphin, Thunar display directories correctly on first open

### Performance Metrics

- **Memory usage**: < 20% increase compared to current implementation
- **API request rate**: No significant increase in API calls per minute
- **Mount time**: Mount operation completes quickly (< 2 seconds), prefetch runs in background
- **Cache hit rate**: > 95% of directory accesses served from cache after prefetch completes

### Quality Metrics

- **Test coverage**: All new code covered by unit and integration tests
- **Test pass rate**: 100% of tests pass with new implementation
- **No regressions**: All existing functionality continues to work
- **Error handling**: All error paths tested and logged appropriately

## References

### SRS Requirements

- **FR-FS-005**: The system shall cache file metadata to improve performance
- **FR-FS-006**: The system shall implement lazy loading for directory contents
- **NFR-PERF-001**: Directory listing operations shall complete within 100ms for cached data
- **NFR-PERF-002**: The system shall minimize API calls through effective caching

### Design Documents

- **ADR-001**: Structured Metadata Store (docs/2-architecture/decisions/ADR-001-structured-metadata-store.md)
- **ADR-003**: Metadata Request Prioritization (docs/2-architecture/decisions/ADR-003-metadata-request-prioritization.md)

### Issue Tracking

- **Issue**: Lazy Directory Loading Performance (docs/issues/lazy-directory-loading-performance.md)
- **Original Spec**: `.kiro/specs/archive/system-verification-and-fix/` (initial sync and cache policy)
- **Merged Spec**: `.kiro/specs/archive/lazy-directory-loading-fix/`

### Test Files

- `internal/fs/cache_test.go`: Cache behavior tests
- `internal/fs/fuse_metadata_local_test.go`: Metadata operation tests
- Test: `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` (needs update)
- Test: `TestIT_FS_Cache_GetChildrenIDDoesNotCallGraphWhenMetadataPresent` (should still pass)
