# Runtime Layering – Metadata Hardening (2025-11-21)

**Type**: Implementation Update  
**Status**: In Progress  
**Components**: `internal/fs/cache.go`, `internal/fs/metadata_store.go`, `internal/fs/stats.go`, `internal/fs/file_operations.go`, `internal/fs/delta.go`, `internal/fs/cache_test.go`, `docs/plans/2025-11-18-runtime-layering-and-notifier-plan.md`

## Summary

- Removed every FUSE-time dependency on the legacy `metadata` bucket so Requirement 2.15’s local-first rule now flows through `metadata_v2`; `GetID`, offline bootstrap, and SerializeAll exclusively read/write the structured store while keeping one-time migrations available.
- Added a structured fallback loader that rebuilds the root inode from `metadata_v2` when Microsoft Graph is unavailable, preventing offline startups from scanning stale JSON blobs and ensuring the runtime-layering design’s single-source-of-truth guarantee.
- Reworked stats/telemetry to analyze `metadata_v2` entries (state counts, directory depth, file histograms) so operators see the same view the metadata state machine enforces, aligning with Work Breakdown items 1 & 6 of the runtime-layering plan.
- Hooked file creation/write paths and metadata-only delta reconciliation into the `metadata.StateManager`, so new local work immediately transitions to `DIRTY_LOCAL` and stable remote updates push entries back to `HYDRATED`, advancing Work Breakdown item 2.
- Applied the same state-machine wiring to directory creation (online vs offline) and added regression tests to lock the expected `HYDRATED`/`DIRTY_LOCAL` transitions in place.

## Testing

- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestMetadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/metadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestFallbackRootFromMetadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestFileCreationMarksMetadataDirty -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestMkdirStateReflectsConnectivity -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
