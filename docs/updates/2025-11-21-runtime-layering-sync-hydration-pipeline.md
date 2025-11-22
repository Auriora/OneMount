# Runtime Layering – Sync & Hydration Pipeline (2025-11-21)

**Type**: Implementation Update  
**Status**: Complete  
**Components**: `internal/fs/sync.go`, `internal/fs/delta.go`, `internal/fs/metadata_store.go`, `internal/fs/delta_state_manager_test.go`

## Summary

- Reworked the tree sync walker to populate `metadata_v2` via the structured store and state manager instead of instantiating inodes directly; parent/child links now update in a single metadata transaction so FUSE stays local-first per Requirement 2.15.
- Routed delta reconciliation through validated state transitions: remote deletions now flow through `markEntryDeleted`, metadata updates are written via `metadata_v2`, and content invalidations move items to `GHOST` (or `CONFLICT` when local edits exist) while successful reconciliations force `HYDRATED`.
- Added helper state transitions that safely reapply options (e.g., `ClearPendingRemote`) even when already in the target state, keeping hydration/upload metadata snapshots consistent with the design’s Item State Model.
- Introduced focused regression tests to lock in the new delta-driven state behavior, cache eviction path for remote changes, and the metadata-only FUSE readdir path when the in-memory cache is cold.

## Testing

- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run 'TestApplyDelta(Set|Hydrates|Marks)' -count=1`
- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestGetChildrenIDUsesMetadataStoreWhenCold -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-RULE-Documentation-Conventions (priority 20)  
- AGENT-RULE-Testing-Conventions (priority 25)
