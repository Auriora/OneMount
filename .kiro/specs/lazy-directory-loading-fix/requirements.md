# Requirements: Lazy Directory Loading Performance Fix

## 1. Overview

### 1.1 Purpose
This specification addresses the lazy directory loading performance issue where directories appear empty on first access and populate after 5-10 seconds. The investigation will determine if this is a bug in the implementation or a gap in the requirements/design.

### 1.2 Background
OneMount uses a lazy-loading approach with:
- **Metadata state management** (GHOST, HYDRATING, HYDRATED states)
- **Metadata request prioritization** (foreground vs background queues)
- **Stale-cache policy** (serve stale data immediately, refresh async)
- **Structured metadata store** (BBolt database for persistence)

### 1.3 Problem Statement
When navigating directories in the mounted OneDrive filesystem:
1. Initial directory listing shows empty (no subdirectories or files)
2. After 5-10 seconds, contents appear
3. File managers show empty folders initially, causing confusion
4. Subsequent access is fast (< 100ms) due to caching

## 2. Investigation Findings

### 2.1 Current Implementation (INCORRECT)
The system currently:
1. Returns empty immediately on cache miss (< 50ms)
2. Schedules background refresh
3. Populates cache asynchronously
4. Requires second access to see contents

This is **confirmed by test**: `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached`

**This is BAD UX** - users see empty directories on first access!

### 2.2 Desired Behavior (CORRECT)
The system **should NEVER return empty**:
1. **First access**: Block and fetch data synchronously (up to 10 second timeout)
2. Return complete data to user on first access
3. Cache the data for subsequent fast access (< 50ms)
4. Background refresh keeps data fresh

**No exceptions** - always wait for data to be available before returning.

### 2.3 Stale-Cache Policy (for subsequent accesses)
After first access, when cache becomes stale:
1. Check if directory already cached
2. Check if cache is fresh (< TTL, e.g., 5 minutes)
3. If fresh, serve from cache immediately (< 50ms)
4. **If stale, block and try to refresh (2 second timeout)**:
   - If refresh succeeds within timeout: Return fresh data
   - If refresh times out: Serve stale data and continue refresh in background
5. Background refresh updates cache for next access

**Key point**: Try to get fresh data first, but don't wait too long - stale data is better than long waits!

### 2.4 Prefetch Strategy (Critical for Performance)
To ensure directories are never empty, implement **recursive prefetch on mount**:

1. **On filesystem mount**:
   - Start prefetch from root directory
   - Fetch all children (files and folders) - metadata only, not content
   - For each subdirectory found, recursively prefetch its children
   - Continue until entire directory tree is prefetched

2. **What to prefetch**:
   - ✅ Directory listings (folder names and metadata)
   - ✅ File names and metadata (size, modified time, etc.)
   - ❌ File contents (only loaded when file is opened)

3. **Prefetch behavior**:
   - Run in background after mount
   - Use low priority to not interfere with user operations
   - Populate cache and metadata store
   - Track progress (GHOST → HYDRATING → HYDRATED states)

4. **User access during prefetch**:
   - If directory already prefetched: Return immediately from cache
   - If directory prefetch in progress: Wait for prefetch to complete
   - If directory not yet prefetched: Block and fetch synchronously

**This ensures**: By the time user navigates to any directory, it's already cached!

### 2.5 File Content Loading
File contents are handled separately:
- **NOT prefetched** - only metadata is prefetched
- **Loaded on-demand** when file is opened
- **Block until download complete** - never return partial/empty file
- Use existing download manager with hydration workers

### 2.5 Root Cause
The current implementation has **two fundamental design flaws**:
1. **Returns empty on cache miss** instead of blocking
2. **No recursive prefetch** to populate cache on mount

**The fix requires**:
1. Implement recursive prefetch on mount (metadata only)
2. Update `GetChildrenID()` to block if not cached
3. Update tests to expect blocking behavior
4. Track prefetch progress with metadata states

## 3. Acceptance Criteria

### 3.1 Investigation Phase
- [ ] **AC-1.1**: Verify current behavior returns empty on first access
- [ ] **AC-1.2**: Identify all code paths that return empty directory listings
- [ ] **AC-1.3**: Review test `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached` expectations
- [ ] **AC-1.4**: Document why current implementation returns empty

### 3.2 Fix Phase
- [ ] **AC-2.1**: First directory access **NEVER returns empty** - always blocks until data available
- [ ] **AC-2.2**: First directory access completes within 10 seconds (with timeout)
- [ ] **AC-2.3**: Subsequent accesses remain fast (< 50ms from cache)
- [ ] **AC-2.4**: Stale cache is served immediately (never empty)
- [ ] **AC-2.5**: File managers display directories correctly on first open
- [ ] **AC-2.6**: No increase in API rate limit errors
- [ ] **AC-2.7**: Memory usage remains acceptable (< 20% increase)
- [ ] **AC-2.8**: All existing tests updated to reflect new behavior
- [ ] **AC-2.9**: New tests verify **NEVER empty** behavior

## 4. Requirements

### 4.1 Functional Requirements

| ID | Requirement | Priority | Rationale |
|----|-------------|----------|-----------|
| **FR-1** | The system shall NEVER return an empty directory listing | Must-have | Core UX requirement - users must see complete data |
| **FR-2** | The system shall block on first directory access until data is fetched from API | Must-have | Required to satisfy FR-1 |
| **FR-3** | The system shall cache directory contents after first fetch | Must-have | Performance requirement for subsequent accesses |
| **FR-4** | The system shall serve cached data immediately on subsequent accesses | Must-have | Performance requirement |
| **FR-5** | The system shall attempt to refresh stale cached data synchronously with 2 second timeout | Must-have | Try to get fresh data without long waits |
| **FR-6** | The system shall serve stale cached data if refresh times out | Must-have | Better UX than blocking indefinitely |
| **FR-7** | The system shall continue refresh in background after serving stale data | Must-have | Keep data fresh for next access |
| **FR-8** | The system shall timeout synchronous fetches after 10 seconds | Must-have | Prevent indefinite blocking |
| **FR-9** | The system shall return error (not empty) if fetch fails | Must-have | Clear error indication to user |
| **FR-10** | The system shall prefetch all directory metadata recursively on mount | Must-have | Ensure directories are cached before user access |
| **FR-11** | The system shall prefetch directory listings (not file contents) | Must-have | Metadata only for performance |
| **FR-12** | The system shall use low priority for prefetch operations | Must-have | Don't interfere with user operations |
| **FR-13** | The system shall track prefetch progress using metadata states | Must-have | Know what's cached vs in-progress |
| **FR-14** | The system shall load file contents only when file is opened | Must-have | On-demand content loading |
| **FR-15** | The system shall block file open until content is downloaded | Must-have | Never return partial/empty file |

### 4.2 Non-Functional Requirements

| ID | Requirement | Priority | Rationale |
|----|-------------|----------|-----------|
| **NFR-1** | First directory access shall complete within 10 seconds (with timeout) | Must-have | User experience - reasonable wait time |
| **NFR-2** | Cached directory access shall complete within 50ms | Must-have | Performance - instant feel |
| **NFR-3** | Stale cache refresh shall be attempted within 2 seconds before fallback | Must-have | Balance freshness vs responsiveness |
| **NFR-4** | Memory usage shall not increase by more than 20% | Should-have | Resource efficiency |
| **NFR-5** | API usage shall not increase significantly | Should-have | Rate limit compliance |
| **NFR-6** | System shall handle network errors gracefully | Must-have | Reliability |
| **NFR-7** | Prefetch shall not significantly delay mount operation | Should-have | Mount should complete quickly |
| **NFR-8** | Prefetch shall use reasonable API rate limits | Must-have | Avoid throttling |
| **NFR-9** | File content download shall show progress indication | Should-have | User feedback |

## 5. Out of Scope

- Prefetching file contents (only metadata is prefetched)
- Selective prefetch (all directories are prefetched)
- User configuration of prefetch behavior (future enhancement)
- Changing the metadata state machine
- Modifying the metadata request prioritization system

## 6. Dependencies

- ADR-001: Structured Metadata Store
- ADR-003: Metadata Request Prioritization
- FR-FS-005: The system shall cache file metadata to improve performance
- Existing tests: `TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached`
- Existing tests: `TestIT_FS_Cache_GetChildrenIDDoesNotCallGraphWhenMetadataPresent`

## 7. Risks and Assumptions

### 7.1 Risks
- Metadata store may not be properly populated during delta sync
- `tryPopulateChildrenFromMetadata()` may have a bug
- Metadata store queries may be inefficient
- State transitions may not be updating metadata store correctly

### 7.2 Assumptions
- The design intent (stale-cache policy) is correct
- The structured metadata store is the right solution
- The existing tests accurately reflect requirements
- Delta sync is running and populating metadata

## 8. Success Metrics

- **NEVER** returns empty directory listing
- First directory access shows complete contents (blocks until available)
- Subsequent accesses are fast (< 50ms from cache)
- Stale cache is served immediately (< 50ms) while refreshing
- No regression in existing functionality
- All tests updated to reflect new behavior
- New tests verify **NEVER empty** requirement

## 9. References

- Issue: `docs/issues/lazy-directory-loading-performance.md`
- ADR-001: `docs/2-architecture/decisions/ADR-001-structured-metadata-store.md`
- ADR-003: `docs/2-architecture/decisions/ADR-003-metadata-request-prioritization.md`
- Tests: `internal/fs/cache_test.go`
- Tests: `internal/fs/fuse_metadata_local_test.go`
