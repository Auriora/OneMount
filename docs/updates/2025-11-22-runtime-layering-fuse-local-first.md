# 2025-11-22 – Runtime Layering: FUSE local-first queueing

## Summary
- Added a bounded mutation queue with worker goroutines so Mkdir/Unlink/Rename enqueue remote Graph mutations instead of spawning unbounded goroutines from FUSE entry points.
- Kept GetChildrenID strictly cache-first by default and ensured background refreshes are isolated from FUSE call stacks.
- Added offline regressions for FUSE Lookup/OpenDir to prove cached metadata_v2 satisfies directory listing and lookups without Graph calls.
- Hardened mutation queue coverage with a non-blocking enqueue test to verify FUSE callers return immediately while mutations retry in the background.

## Testing
- `go test ./internal/fs -run "(MutationQueue|FuseMetadata)" -count=1`
