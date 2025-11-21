# Runtime Layering – FUSE Metadata Persistence (2025-11-20)

**Type**: Implementation Update  
**Status**: In Progress  
**Components**: `internal/fs/cache.go`, `internal/fs/virtual_files.go`, `internal/fs/cache_test.go`, `docs/plans/2025-11-18-runtime-layering-and-notifier-plan.md`

## Summary

- Persisted parent/child relationships directly into `metadata_v2` whenever `InsertID`, `DeleteID`, or `cacheChildrenFromMap` mutate directory listings, so the metadata store is now the authoritative source for `readdir` overlays and pending-remote nodes instead of best-effort snapshots in memory.
- Added metadata-backed fallbacks to `GetChildrenID` (`tryPopulateChildrenFromMetadata`) plus `RegisterVirtualFile` persistence so local-only overlays (e.g., `.xdg-volume-info`) and prior directory enumerations can be resurrected instantly, even when the Graph client is offline or the in-memory cache was evicted.
- Hardened `DeleteID`/virtual-file registration to reload parents from the structured store when necessary and to write updates back after each mutation, eliminating cache divergence between BBolt and FUSE.
- Introduced `TestGetChildrenIDUsesMetadataStoreWhenOffline`, which deletes every cached child, forces offline mode, and proves that `GetChildrenID` now succeeds purely from structured metadata—locking in the runtime-layering guarantee that FUSE never blocks on Graph for directories it has already seen.
- Wired the content cache’s eviction path through structured metadata: pinned or dirty-local items are now skipped, successful evictions transition entries back to `GHOST`, and tree-sync persistence writes every discovered child into `metadata_v2` so later hydrations never depend on legacy JSON.
- `GetPath` now resolves components directly from `metadata_v2` (via `tryPopulateChildrenFromMetadata`), so offline path traversal rebuilds inodes locally without falling back to Graph.
- Delta change handling (`applyDelta`) now snapshots remote metadata into `metadata_v2`, updates item states when content is invalidated, and reuses structured entries so metadata-only updates never reach back to Graph from FUSE.
- Added `TestContentEvictionTransitionsMetadata` and `TestPinnedContentNotEvicted` to lock in the new eviction semantics (state-machine transitions plus pin-mode guardrails).

## Testing

- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestGetChildrenIDUsesMetadataStoreWhenOffline -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestGetPathUsesMetadataStoreWhenOffline -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestContentEvictionTransitionsMetadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestPinnedContentNotEvicted -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestApplyDelta -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
