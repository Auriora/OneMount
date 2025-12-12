# Inode Truncate Lock Reduction

**Date**: 2025-11-23  
**Type**: Bugfix  
**Component**: Filesystem (SetAttr)

## Summary

`SetAttr` held `inode.mu` while performing on-disk truncation, allowing slow I/O to block all readers and writers on that inode and causing hangs during directory lookups. The lock scope has been narrowed so the truncate I/O runs outside the mutex.

## Key Changes

1. `internal/fs/metadata_operations.go`: capture truncate intent under lock, release the inode mutex before calling `fd.Truncate`, then briefly re-lock to update size/dirty state. Virtual-file truncation remains unchanged.

## Verification

- Pending: rebuild, re-run the hanging scenario (truncate then dir traversal). The mount should no longer stall; directory listings should respond while truncation proceeds.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
