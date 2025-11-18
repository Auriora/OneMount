# `du` Zero Usage Investigation (2025-11-18)

**Type**: Bugfix / Investigation  
**Status**: Complete  
**Components**: `internal/fs/inode.go`, `internal/fs/inode_attr_test.go`

## Summary

- Confirmed that OneDrive metadata already persists each item's logical size, but our FUSE replies never populated `Attr.Blocks`/`Attr.Blksize`, so every `stat(2)` reported zero allocated blocks and `du` always printed `0`.
- Added explicit block accounting derived from the cached OneDrive size (including the 4 KiB placeholder for directories) and exposed it through both `makeAttr` and the legacy `MakeAttr` helper so all kernel responses now contain realistic block counts.
- Documented the behavior with a focused unit test that asserts `inode.makeAttr()` copies OneDrive sizes verbatim and emits the rounded-up `Blocks` value expected by tools like `du`.

## Testing

- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestInodeMakeAttrReportsBlocksUsingMetadata`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
