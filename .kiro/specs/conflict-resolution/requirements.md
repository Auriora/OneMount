# Requirements Document: Conflict Resolution

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/requirements.md`. Sections below are verbatim.

## Requirements

### Requirement 8: Conflict Resolution Verification

**User Story:** As a user, I want conflicts between local and remote changes to be handled gracefully so that I don't lose any work.

#### Acceptance Criteria

1. WHEN a file has been modified both locally and remotely, THE OneMount System SHALL detect the conflict by comparing ETags
2. WHEN uploading a file with local changes, THE OneMount System SHALL check if the remote ETag has changed since last sync
3. IF the remote ETag differs from the cached ETag, THEN THE OneMount System SHALL detect a conflict
4. WHEN a conflict is detected, THE OneMount System SHALL preserve the local version with its original name
5. WHEN a conflict is detected, THE OneMount System SHALL create a conflict copy with a timestamp suffix
6. WHEN a conflict is detected, THE OneMount System SHALL download the remote version as the conflict copy
7. WHEN a conflict is resolved, THE OneMount System SHALL log the conflict details including file path, ETags, and timestamps
8. WHERE multiple conflict resolution strategies are available, THE OneMount System SHALL use the configured strategy (last-writer-wins, keep-both, user-choice, merge, or rename)
9. WHEN the user accesses a file with unresolved conflicts, THE OneMount System SHALL display both versions
10. WHERE the user configures a default conflict resolution strategy, THE OneMount System SHALL use the specified strategy for automatic conflict resolution
11. IF no conflict resolution strategy is configured, THEN THE OneMount System SHALL use the keep-both strategy as default
12. WHEN using the last-writer-wins strategy, THE OneMount System SHALL compare modification timestamps and preserve the most recent version
13. WHEN using the user-choice strategy, THE OneMount System SHALL present resolution options to the user
14. WHEN using the merge strategy, THE OneMount System SHALL attempt automatic merging for compatible changes
15. WHEN using the rename strategy, THE OneMount System SHALL create separate versions with conflict indicators
16. WHEN using the keep-both strategy, THE OneMount System SHALL create separate versions for both local and remote changes
