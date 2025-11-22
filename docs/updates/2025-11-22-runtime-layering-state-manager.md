# 2025-11-22 – Runtime Layering: State manager in mutation hot paths

## Summary
- Routed file writes, truncation, rename, upload queueing, and conflict resolution through the metadata `StateManager` helpers (`markDirtyLocalState`/`markCleanLocalState`) to eliminate direct inode state flips in runtime paths.
- Ensured upload completions and remote-ID swaps clear dirty state via validated transitions; upload queue backpressure no longer clears dirty flags.
- Added shared state helper implementations to keep runtime `hasChanges` flags in sync with `metadata_v2` transitions.
- Added state-manager coverage for eviction: content eviction now transitions entries to `GHOST` and pinned items still auto-hydrate; remote delete mutations mark entries `DELETED`.
- Hydration error paths now respect fast retry caps in tests via `ONEMOUNT_TEST_FAST_RETRY`, reducing backoff while still exercising ERROR transitions.
- Offline change processing now marks state via helpers (create/modify → DIRTY, delete → DELETED, rename → DIRTY), eliminating direct flag flips.
- Runtime `hasChanges` writes are now centralized; remaining direct writes are limited to inode reconstruction from persisted metadata.
- Added delete regression test (`mutation_queue_test.go`) ensuring queued remote delete drives `DELETED` state; network-error hydration tests run with fast retry.

## Testing
- `go test ./internal/fs -run TestMetadataEntryFromInodeStateInference -count=1`
- `go test ./internal/fs -run "MetadataStore|StateManager" -count=1`
