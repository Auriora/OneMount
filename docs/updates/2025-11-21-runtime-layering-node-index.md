# Runtime Layering – Local-First FUSE Node Index (2025-11-21)

**Type**: Implementation Update  
**Status**: In Progress  
**Components**: `internal/fs/cache.go`, `internal/fs/filesystem_types.go`, `internal/fs/file_operations.go`, `internal/fs/dir_operations.go`, `internal/fs/xattr_operations.go`, `internal/fs/fuse_thumbnail_handlers.go`, `internal/fs/metadata_operations.go`, `internal/fs/metadata_store.go`

## Summary

- Finished wiring the node-index map introduced earlier through every FUSE touchpoint (Open/Read/Write/Fsync, directory ops, xattrs, thumbnails) so each syscall resolves the in-memory inode in O(1) before touching Drive IDs, satisfying Requirement 2.15 and Work Breakdown item 3 in `docs/plans/2025-11-18-runtime-layering-and-notifier-plan.md`.
- Ensured local metadata/state remains authoritative by short-circuiting virtual items and deleted entries inside `markEntryDeleted`/`ensureInodeFromMetadataStore`, matching the `.kiro/specs/system-verification-and-fix` state machine for `DELETED_LOCAL` transitions.
- Reworked directory helpers (`OpenDir`, `Lookup`, `Rmdir`) and thumbnail/xattr handlers to reuse `GetNodeID`, preventing redundant `TranslateID → GetID` scans while keeping overlay and virtual-file policies local-first.
- Documented and validated that the optimized node index brings `TestHighConcurrencyStress` back under the 25 ms budget, proving the per-op work reduction envisioned in the runtime-layering plan.

## Testing

- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestHighConcurrencyStress -count=1 -timeout 20m`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
