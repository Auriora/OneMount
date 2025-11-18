# Directory Copy Consistency Fix (2025-11-18)

**Type**: Bugfix / Filesystem Behavior  
**Status**: Complete  
**Components**: `internal/fs/cache.go`, `internal/fs/dir_operations.go`, `internal/fs/dir_pending_test.go`, `internal/fs/filesystem_types.go`

## Summary

- Identified that newly created directories disappeared immediately after a foreground metadata refresh because the refresh pipeline dropped any child whose OneDrive ID already existed but hadn’t been returned by `GetItemChildren` yet.
- Added a short-lived “pending remote” tracker so directories created via `Mkdir` stay in the cache (and remain discoverable by `cp`, `stat`, etc.) until a Graph listing confirms them, preventing cascading `ENOENT` errors during recursive copies.
- Cleared the pending flag once the remote listing surfaces the directory, ensuring we don’t mask real deletions, and covered the regression with a fixture-based test that simulates Graph lag.

## Testing

- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestDirectoryPendingRemoteVisibilitySurvivesRefresh`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
