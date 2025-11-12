# Cache Behavior for Deleted Files - Documentation Update

**Date:** 2025-11-12  
**Task:** 21.4 Document Cache Behavior for Deleted Files  
**Source:** `docs/verification-phase4-file-write-operations.md`

## Overview

This document summarizes the updates made to requirements and design documentation to clarify the expected behavior for cache management when files are deleted from the filesystem.

## Issue Background

During Phase 4 verification (file write operations), it was discovered that the content cache retains file data even after files are deleted from the filesystem. The verification document noted:

> **Issue 1: Content Cache Persistence After Deletion**
> 
> **Severity:** Low (Expected Behavior)  
> **Component:** Content Cache (`internal/fs/content_cache.go`)  
> **Description:** The content cache retains file data even after files are deleted from the filesystem.
> 
> **Analysis:**
> - The `LoopbackCache.Open()` method uses `os.O_CREATE` flag, which creates files if they don't exist
> - This means deleted files can still be opened from the cache
> - This is likely intentional for caching purposes and performance optimization

The task required documenting the expected behavior and clarifying that deleted files should be removed from the cache.

## Changes Made

### 1. Requirements Document Updates

**File:** `.kiro/specs/system-verification-and-fix/requirements.md`

Added two new acceptance criteria to Requirement 7 (Cache Management Verification):

8. **WHEN a file is deleted from the filesystem, THE OneMount System SHALL remove the corresponding cache entry to free disk space**
   - Ensures that deleted files don't consume disk space indefinitely
   - Clarifies that cache cleanup is expected for deleted files

9. **WHEN cache cleanup runs, THE OneMount System SHALL identify and remove cache entries for files that no longer exist in the filesystem metadata**
   - Addresses orphaned cache entries (files deleted but cache not cleaned up)
   - Ensures periodic cleanup maintains cache hygiene

### 2. Design Document Updates

**File:** `.kiro/specs/system-verification-and-fix/design.md`

#### Section 7: Cache Management Component

Added **Cache Cleanup Behavior** section with four key behaviors:

1. **Time-based expiration**: Files older than the configured expiration threshold are removed during periodic cleanup
2. **Deleted file cleanup**: When a file is deleted from the filesystem, the corresponding cache entry should be removed to free disk space
3. **Orphaned cache entries**: During cleanup, cache entries for files that no longer exist in the filesystem metadata should be identified and removed
4. **Cleanup frequency**: Cache cleanup runs periodically (default: every 24 hours) to maintain cache hygiene

Updated **Verification Criteria** to include:
- Cleanup removes cache entries for deleted files
- Cleanup removes orphaned cache entries (files not in metadata)

#### Cache Cleanup Implementation Section

Added new section after the ETag Cache Entry Data Model with detailed implementation guidance:

**Deleted File Handling** describes three approaches:

1. **Immediate cleanup on deletion**: 
   - Remove inode from filesystem tracking
   - Mark cache entry for deletion
   - Optionally remove cache entry immediately

2. **Periodic cleanup of orphaned entries**:
   - Iterate through all cache entries
   - Check if each cached item ID exists in metadata database
   - Remove cache entries for items that no longer exist
   - Log the number of orphaned entries removed

3. **Cache cleanup algorithm**:
   - Pseudocode for the cleanup process
   - Checks both metadata existence and expiration
   - Updates cache statistics after cleanup

**Implementation Notes**:
- Documents the `LoopbackCache.Open()` behavior with `os.O_CREATE` flag
- Explains why deferred cleanup is acceptable
- Specifies cleanup frequency and manual trigger options
- Notes that cache statistics should reflect actual state

**Rationale for Deferred Cleanup**:
- Immediate deletion on every `Unlink` could impact performance
- Periodic cleanup is more efficient for batch operations
- Allows for potential "undo" functionality if file is restored quickly
- Balances disk space usage with performance

## Implementation Impact

### Current Behavior
The current implementation allows deleted files to remain in the cache, which is acceptable for performance reasons but requires proper cleanup.

### Expected Behavior
After implementing these requirements:

1. **On file deletion**: The system should mark the cache entry for cleanup (immediate or deferred)
2. **During periodic cleanup**: The system should:
   - Remove cache entries for deleted files
   - Remove orphaned cache entries (not in metadata)
   - Update cache statistics
   - Log cleanup results

3. **Cache statistics**: Should accurately reflect the state after cleanup

### Implementation Tasks

To fully implement these requirements, the following code changes may be needed:

1. **Modify `Unlink` operation** (`internal/fs/file_operations.go`):
   - Add cache entry cleanup or marking for deletion
   - Consider immediate vs. deferred cleanup trade-offs

2. **Enhance cache cleanup** (`internal/fs/cache.go`):
   - Add orphaned entry detection (check metadata database)
   - Remove cache files for items not in metadata
   - Update cleanup statistics

3. **Add tests**:
   - Test cache cleanup after file deletion
   - Test orphaned entry removal
   - Test cache statistics accuracy

## Requirements Traceability

| Requirement | Description | Status |
|-------------|-------------|--------|
| 7.1 | Store content with ETag | ✅ Existing |
| 7.2 | Update last access time | ✅ Existing |
| 7.3 | Invalidate on ETag change | ✅ Existing |
| 7.4 | Invalidate on delta sync | ✅ Existing |
| 7.5 | Remove old files | ✅ Existing |
| 7.6 | Respect expiration config | ✅ Existing |
| 7.7 | Display cache statistics | ✅ Existing |
| **7.8** | **Remove cache on file deletion** | **✅ Added** |
| **7.9** | **Remove orphaned cache entries** | **✅ Added** |

## Verification

To verify this behavior is implemented correctly:

1. **Unit tests**:
   - Test that `Unlink` marks cache for cleanup
   - Test that cleanup removes orphaned entries
   - Test cache statistics after cleanup

2. **Integration tests**:
   - Delete files and verify cache is cleaned up
   - Run cleanup and verify orphaned entries are removed
   - Check cache statistics reflect actual state

3. **Manual testing**:
   - Delete files via filesystem operations
   - Check cache directory for orphaned files
   - Trigger cache cleanup manually
   - Verify cache statistics are accurate

## Related Issues

- **Issue #CACHE-001**: No Cache Size Limit Enforcement (Task 20.10)
- **Issue #CACHE-002**: No Explicit Cache Invalidation When ETag Changes (Task 20.11)
- **Issue #CACHE-003**: Statistics Collection Slow for Large Filesystems (Task 20.12)
- **Issue #CACHE-004**: Fixed 24-Hour Cleanup Interval (Task 20.13)

## Conclusion

The requirements and design documentation have been updated to clarify that:

1. Deleted files should be removed from the cache to free disk space
2. Periodic cleanup should remove orphaned cache entries
3. The implementation can use deferred cleanup for performance reasons
4. Cache statistics should reflect the actual state after cleanup

This documentation provides clear guidance for implementing proper cache cleanup behavior while maintaining performance and allowing for optimization strategies like deferred cleanup.

## Rules Consulted

- **testing-conventions.md**: Docker test environment requirements
- **operational-best-practices.md**: Documentation consistency and SRS alignment
- **coding-standards.md**: Documentation as code principles
- **general-preferences.md**: SOLID and DRY principles

## Rules Applied

- Updated requirements specification with new acceptance criteria (7.8, 7.9)
- Updated design documentation with implementation guidance
- Maintained consistency between requirements and design
- Documented rationale for design decisions
- Provided clear verification criteria
