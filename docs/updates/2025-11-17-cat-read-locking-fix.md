# Cat Read Hang – Locking Fix & Stress Coverage (2025-11-17)

**Type**: Implementation / Testing  
**Status**: Complete  
**Components**: `internal/fs/cache.go`, `internal/fs/child_snapshot.go`, `internal/fs/virtual_files.go`, `internal/fs/inode.go`, `internal/fs/concurrency_test.go`

## Summary

- Refactored the metadata cache to snapshot child lists before iterating, eliminating the parent→child lock inversion that blocked FUSE replies during `cat`/`ls`.
- Introduced reusable child snapshot helpers and updated virtual file handling so callers add virtual entries without touching child locks while a parent lock is held.
- Added debug-only lock-hold instrumentation (triggered when `LOG_LEVEL=debug`) around the longest-running inode-parent critical sections to surface future regressions quickly.
- Enhanced `cacheChildrenFromMap` and other helpers to precompute child metadata outside parent locks while still ensuring virtual entries are reflected in caller maps.
- Bumped concurrency coverage with `TestDirectoryEnumerationWhileRefreshing`, which mimics the original hang by running `GetChildrenID` alongside concurrent metadata refreshes using the `deadlockMonitor`.

## Testing

- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestDirectoryEnumerationWhileRefreshing -timeout 5m`
- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestDeadlockPrevention -timeout 5m`
- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestHighConcurrencyStress -timeout 5m`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Testing-Conventions (priority 25)
- docs/guides/developer/concurrency-guidelines.md (locking guidance)
