# Runtime Layering – Hydration Pipeline & Pinning (2025-11-21)

**Type**: Implementation Update  
**Status**: In Progress  
**Components**: `internal/fs/cache.go`, `internal/fs/test_hooks.go`, `internal/fs/content_eviction_test.go`, `internal/fs/delta_test.go`

## Summary

- Added a metadata-driven auto-hydration helper that re-queues downloads for `PinModeAlways` items whenever eviction or delta invalidation transitions them back to `GHOST`, keeping pinned entries compliant with `.kiro/specs/system-verification-and-fix` Requirement 7 and the plan’s Work Breakdown #4.
- Extended filesystem test hooks with an `AutoHydrateHook` so unit suites can assert the behavior without spinning real downloads, and wired the eviction handler to invoke the helper immediately after updating structured metadata.
- Broadened regression coverage: new eviction and delta tests prove pinned items refuse to stay ghosted (even when delta invalidates them) while still honoring virtual/download wiring from earlier runtime layering slices.
- Taught `applyDelta` to update `metadata_v2` entries (name/parent/ETag/size/LastModified) via the structured store so delta processing no longer relies on mutating in-memory inodes alone, closing the remaining gap in Work Breakdown item 4.

## Testing

- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestPinnedContentAutoHydratesAfterEviction -count=1`
- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestApplyDeltaPinnedFileQueuesHydration -count=1`
- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestApplyDelta -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
