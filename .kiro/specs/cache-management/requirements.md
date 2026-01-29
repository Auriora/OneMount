# Requirements Document: Cache Management

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/requirements.md`. Sections below are verbatim.

## Requirements

### Requirement 7: Cache Management Verification

**User Story:** As a user, I want the cache to be managed efficiently so that it doesn't consume excessive disk space and always reflects the latest remote state.

#### Acceptance Criteria

1. WHEN files are downloaded, THE OneMount System SHALL store content in the cache directory with the file's ETag
2. WHEN files are accessed, THE OneMount System SHALL update the last access time in the cache
3. WHEN a cached file's ETag differs from the remote ETag, THE OneMount System SHALL invalidate the cache entry and download the new version
4. WHEN delta sync detects remote changes, THE OneMount System SHALL invalidate affected cache entries to prevent stale data
5. WHILE the cache cleanup process runs, THE OneMount System SHALL remove files older than the expiration threshold
6. WHERE cache expiration is configured, THE OneMount System SHALL respect the configured number of days
7. WHEN the user requests cache statistics, THE OneMount System SHALL display cache size, file count, and hit rate
8. WHEN a file is deleted from the filesystem, THE OneMount System SHALL remove the corresponding cache entry to free disk space
9. WHEN cache cleanup runs, THE OneMount System SHALL identify and remove cache entries for files that no longer exist in the filesystem metadata

### Requirement 21: Metadata State Model Verification

**User Story:** As a developer, I want a clearly defined metadata state machine so that every file or folder transitions predictably between cloud-only, hydrated, dirty, or deleted states.

#### Acceptance Criteria

1. THE metadata database SHALL persist an `item_state` field whose value is one of: `GHOST`, `HYDRATING`, `HYDRATED`, `DIRTY_LOCAL`, `DELETED_LOCAL`, `CONFLICT`, or `ERROR`.
2. WHEN a drive item is discovered via delta for the first time, THE OneMount System SHALL insert it with state `GHOST` and SHALL not download content until a user action requires it or a pinning policy hydrates it.
3. WHEN a user or policy triggers hydration, THE OneMount System SHALL transition the item to `HYDRATING` while the download is in flight and SHALL record the worker responsible so duplicate hydrations can be deduplicated.
4. WHEN hydration completes successfully, THE OneMount System SHALL transition the item to `HYDRATED`, record the content path, update size/mtime metadata, and clear any hydration error fields.
5. WHEN hydration fails, THE OneMount System SHALL transition the item to `ERROR`, capture the failure reason, and keep the previous state metadata so that the user can retry.
6. WHEN a hydrated file is modified locally, THE OneMount System SHALL transition it to `DIRTY_LOCAL` until the upload succeeds, at which point it SHALL return to `HYDRATED` with the new remote ETag.
7. WHEN a local delete occurs, THE OneMount System SHALL transition the item to `DELETED_LOCAL` and queue the delete operation; after Graph confirms the delete, the item SHALL be removed (or left as a tombstone if required for conflict resolution).
8. WHEN delta detects conflicting remote changes for an item that is `DIRTY_LOCAL`, THE OneMount System SHALL transition it to `CONFLICT`, persist both versions’ metadata, and emit the conflict notification defined in Requirement 8.
9. WHEN pinning or eviction policies remove local content for disk-space reasons, THE OneMount System SHALL transition the item back to `GHOST` (or `HYDRATED` if immediately rehydrated) without deleting its metadata entry.
10. ALL virtual-only entries (Requirement 2.16) SHALL set `item_state=HYDRATED`, `remote_id=NULL`, and `is_virtual=TRUE`, ensuring they bypass sync/upload logic while still participating in directory listings.
