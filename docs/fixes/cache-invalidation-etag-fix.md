# Cache Invalidation on ETag Change - Fix Documentation

**Issue ID**: #CACHE-002  
**Component**: Cache Management / Delta Sync  
**Severity**: Medium  
**Status**: ✅ RESOLVED (2025-11-13)  
**Fixed By**: Task 20.11

## Problem Description

When delta sync detected that a file's ETag had changed (indicating remote modification), the cached content was not being explicitly invalidated with proper status updates. While the cache deletion was happening, the file status was not being updated to reflect that the file was out of sync, which could lead to confusion in the UI and potential issues with file status tracking.

### Steps to Reproduce

1. Access a file to cache it locally
2. Modify the file remotely (via web or another client)
3. Delta sync detects ETag change
4. Cached content was deleted but file status was not updated
5. File status extended attributes were not updated for UI integration

### Expected Behavior

- Delta sync should explicitly invalidate cached content when ETag changes
- Cached file should be marked as OutofSync
- File status extended attributes should be updated
- Next access should trigger fresh download
- No stale content served to user

### Actual Behavior (Before Fix)

- ETag stored in metadata was updated
- Cached content was deleted (implicit invalidation)
- File status was not explicitly set to OutofSync
- Extended attributes were not updated
- Potential for UI to show incorrect status

## Solution

### Code Changes

**File**: `internal/fs/delta.go`

Enhanced the `applyDelta` function to:

1. **Explicit Cache Invalidation**: Added error handling for cache deletion
2. **File Status Update**: Call `MarkFileOutofSync(id)` to explicitly set the file status
3. **Extended Attributes Update**: Call `updateFileStatus(local)` to update xattrs for UI integration
4. **Improved Logging**: Changed log level to Info and added more descriptive messages

```go
// Before (line ~547):
logger.Debug().Msg("Content has changed, invalidating cache")
// invalidate the cache
f.content.Delete(id)
// update the metadata
local.mu.Lock()
local.DriveItem.ModTime = delta.ModTime
local.DriveItem.Size = delta.Size
local.DriveItem.ETag = delta.ETag
local.DriveItem.File = delta.File
local.hasChanges = false
local.mu.Unlock()
logger.Debug().Msg("Updated metadata and invalidated content cache")

// After:
logger.Info().Str("delta", "invalidate").
    Msg("Content has changed, invalidating cache and marking file as out of sync")
// Explicitly invalidate the cache by deleting cached content
// This ensures stale content is not served to users
if err := f.content.Delete(id); err != nil {
    logger.Warn().Err(err).Msg("Failed to delete cached content during invalidation")
}
// Mark file status as OutofSync to indicate it needs to be re-downloaded
f.MarkFileOutofSync(id)
// update the metadata with new ETag and size
local.mu.Lock()
local.DriveItem.ModTime = delta.ModTime
local.DriveItem.Size = delta.Size
local.DriveItem.ETag = delta.ETag
local.DriveItem.File = delta.File
local.hasChanges = false
local.mu.Unlock()
// Update file status extended attributes for UI integration
f.updateFileStatus(local)
logger.Debug().Msg("Updated metadata, invalidated content cache, and marked file as OutofSync")
```

### Integration Test

**File**: `internal/fs/delta_sync_integration_test.go`

Added comprehensive integration test `TestIT_Delta_10_07_ETagCacheInvalidation` that verifies:

1. File can be cached locally
2. Remote modification (ETag change) is detected
3. Delta sync triggers cache invalidation
4. File status is updated to OutofSync
5. Metadata is updated with new ETag and size
6. Extended attributes are updated for UI integration

**Test Coverage**:
- Requirements 7.3: ETag-based cache invalidation
- Requirements 7.4: Delta sync cache invalidation
- Requirements 5.3: Remotely modified files download new version

## Benefits

1. **Explicit Status Tracking**: File status is now explicitly set to OutofSync when ETag changes
2. **UI Integration**: Extended attributes are updated, allowing file managers to show correct sync status
3. **Better Observability**: Improved logging makes it easier to track cache invalidation events
4. **Error Handling**: Added error handling for cache deletion failures
5. **Test Coverage**: Integration test ensures the behavior is maintained

## Verification

### Manual Testing

1. Mount OneMount filesystem
2. Access a file to cache it
3. Modify the file remotely (OneDrive web interface)
4. Wait for delta sync to run (or trigger manually)
5. Verify file status shows OutofSync
6. Access the file again to trigger download
7. Verify new content is downloaded

### Automated Testing

Run the integration test:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  go test -v -run TestIT_Delta_10_07_ETagCacheInvalidation ./internal/fs
```

Expected result: Test passes, verifying:
- Cache invalidation on ETag change
- File status update to OutofSync
- Extended attributes update
- Metadata update with new ETag

## Related Requirements

- **Requirement 7.3**: Cache invalidation when ETag differs from remote
- **Requirement 7.4**: Delta sync invalidates affected cache entries
- **Requirement 5.3**: Remotely modified files download new version
- **Requirement 8.1**: File status reflects sync state

## Related Issues

- Issue #CACHE-001: No cache size limit enforcement (separate issue)
- Issue #CACHE-003: Statistics collection slow for large filesystems (separate issue)
- Issue #CACHE-004: Fixed 24-hour cleanup interval (separate issue)

## Implementation Notes

### Cache Invalidation Flow

1. **Delta Sync Detection**: Delta sync detects ETag mismatch
2. **Content Comparison**: Checks if content hash is the same (metadata-only change)
3. **Cache Deletion**: If content changed, deletes cached file
4. **Status Update**: Marks file as OutofSync in status map
5. **Metadata Update**: Updates inode with new ETag, size, and modification time
6. **Extended Attributes**: Updates xattrs for UI integration
7. **D-Bus Signal**: Sends status update signal if D-Bus is available

### File Status States

- **StatusCloud**: File exists in cloud but not in local cache
- **StatusLocal**: File is cached and up-to-date
- **StatusOutofSync**: File needs to be updated from cloud (ETag changed)
- **StatusDownloading**: File is currently being downloaded
- **StatusLocalModified**: File has local changes not yet uploaded
- **StatusSyncing**: File is currently being uploaded
- **StatusError**: Error occurred during sync

### Extended Attributes

The following extended attributes are set on files:

- `user.onemount.status`: Current sync status (e.g., "OutofSync")
- `user.onemount.error`: Error message (if status is Error)

These attributes are used by file manager extensions (Nemo, Nautilus) to display sync status icons.

## Future Enhancements

1. **Cache Size Limits**: Implement LRU eviction with configurable size limit (Issue #CACHE-001)
2. **Proactive Invalidation**: Consider invalidating cache proactively when offline changes are detected
3. **Partial Invalidation**: For large files, consider invalidating only changed chunks
4. **Status Notifications**: Add user notifications when files become out of sync

## References

- Design Document: `docs/2-architecture-and-design/software-design-specification.md`
- Requirements: `.kiro/specs/system-verification-and-fix/requirements.md`
- Verification Tracking: `docs/reports/verification-tracking.md`
- Task List: `.kiro/specs/system-verification-and-fix/tasks.md` (Task 20.11)
