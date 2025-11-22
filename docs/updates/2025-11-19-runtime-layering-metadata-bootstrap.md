# Runtime Layering Metadata Bootstrap (2025-11-19)

**Type**: Implementation Update  
**Status**: Complete  
**Components**: `internal/metadata`, `internal/fs`, `docs/plans/2025-11-18-runtime-layering-and-notifier-plan.md`

## Summary

- Introduced the `internal/metadata` package with formal `ItemState`, overlay, pin-mode, and metadata entry definitions so future runtime-layering work can reference a single schema that mirrors `.kiro/specs/system-verification-and-fix/{requirements,design}.md`.
- Added the `metadata_v2` BBolt bucket, legacy-migration bootstrap, and snapshot helpers that emit explicit item-state records (hydrated, dirty, ghost, virtual) alongside the existing inode serialization, unlocking local-first FUSE changes without breaking offline starts.
- Captured initial heuristics (virtual overlays, pending-remote markers, cache-backed hydration detection) plus migration/unit tests to ensure metadata rebuilds remain deterministic before wiring the new state machine into FUSE/runtime layers.
- Implemented the Bolt-backed metadata store and `metadata.StateManager`, Filesystem helper APIs, pending-remote propagation, and offline root loading built on the structured records so future FUSE/runtime code can consume consistent state (legacy JSON now acts as a fallback source only).
- Updated `GetID`/metadata loaders to instantiate inodes from structured metadata before legacy JSON, ensuring offline metadata requests (including metadata request manager shims) stay entirely local without re-fetching Graph content.
- Hooked DownloadManager and UploadManager into the state machine so hydration/upload lifecycles emit `HYDRATING`, `DIRTY_LOCAL`, `HYDRATED`, and `ERROR` transitions, and added regression tests that verify queueing, completion, and failure paths against the structured metadata store.

## Testing

- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/metadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestMetadata -count=1`
- `GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestUploadManager -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Planning-Protocol (priority 30)
- AGENT-RULE-Documentation-Conventions (priority 20)
