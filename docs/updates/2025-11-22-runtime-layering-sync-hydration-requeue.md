# 2025-11-22 – Runtime Layering: Sync/hydration snapshots & pinned requeue

## Summary
- Routed sync walker writes through the validated state manager so directory/file entries record canonical state transitions and timestamps instead of raw metadata mutations.
- Persisted hydration/upload error snapshots via state-manager transitions; hydration/upload errors now capture `LastError`, completion time, and worker IDs in `metadata_v2`.
- Added pinned invalidation regression: delta invalidations on pinned items trigger deterministic auto-hydration (guarded by test hook).

## Testing
- `go test ./internal/fs -run "HydrationError|UploadError|ApplyDeltaPinnedItemRequeuesHydration" -count=1`
