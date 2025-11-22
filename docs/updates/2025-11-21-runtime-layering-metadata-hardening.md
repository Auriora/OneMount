# Runtime Layering – Metadata Hardening (2025-11-21)

**Type**: Implementation Update  
**Status**: Complete  
**Components**: `internal/fs/cache.go`, `internal/fs/metadata_store.go`, `internal/fs/stats.go`, `internal/fs/file_operations.go`, `internal/fs/metadata_operations.go`, `internal/fs/mutation_queue.go`, `internal/fs/delta.go`, `internal/fs/cache_test.go`, `internal/fs/metadata_operations_test.go`, `docs/plans/2025-11-18-runtime-layering-and-notifier-plan.md`

## Summary

- Removed every FUSE-time dependency on the legacy `metadata` bucket so Requirement 2.15’s local-first rule now flows through `metadata_v2`; `GetID`, offline bootstrap, and SerializeAll exclusively read/write the structured store while keeping one-time migrations available.
- Added a structured fallback loader that rebuilds the root inode from `metadata_v2` when Microsoft Graph is unavailable, preventing offline startups from scanning stale JSON blobs and ensuring the runtime-layering design’s single-source-of-truth guarantee.
- Reworked stats/telemetry to analyze `metadata_v2` entries (state counts, directory depth, file histograms) so operators see the same view the metadata state machine enforces, aligning with Work Breakdown items 1 & 6 of the runtime-layering plan.
- Hooked file creation/write paths and metadata-only delta reconciliation into the `metadata.StateManager`, so new local work immediately transitions to `DIRTY_LOCAL` and stable remote updates push entries back to `HYDRATED`, advancing Work Breakdown item 2.
- Applied the same state-machine wiring to directory creation (online vs offline) and added regression tests to lock the expected `HYDRATED`/`DIRTY_LOCAL` transitions in place.
- Taught `GetChildrenID` to return immediately with whatever metadata is cached while queuing an async refresh through the request manager, so FUSE never blocks on Graph; added coverage that the call no longer stalls when a directory has never been enumerated.
- Swapped directory create/delete paths to pure local mutations: `Mkdir` now inserts a local inode, marks it dirty/pending-remote, and queues an async Graph create, while `Unlink` immediately removes the inode and queues remote deletion so foreground FUSE threads never wait on Graph.
- Renames now run entirely against local metadata, mark entries `DIRTY_LOCAL`, and either record an offline rename change or queue a background Graph rename so FUSE `Rename` calls never wait on the network.

## Testing

- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestMetadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/metadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestFallbackRootFromMetadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestFileCreationMarksMetadataDirty -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestMkdirStateReflectsConnectivity -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestGetChildrenIDReturnsQuicklyWhenUncached -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestRenameRecordsOfflineChange -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
