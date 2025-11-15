# XDG Virtual File Deadlock Fix

**Date**: 2025-11-14  
**Type**: Bugfix  
**Component**: Filesystem Cache / Virtual Files  
**Status**: Complete

## Summary

`go test ./cmd/common` was hanging for 10 minutes inside `TestUT_CMD_01_01_XDGVolumeInfo_VirtualFileBehavior` and eventually timing out. The test fixture inserts a remote `.xdg-volume-info` inode and then calls `CreateXDGVolumeInfo`, which internally refreshes the cache by asking `GetChildrenID` for the root children. That code path calls `appendVirtualChildrenLocked` while holding `parent.mu` and `inode.ID()` was re-entering the same lock, causing a self-deadlock.

## Root Cause

- `appendVirtualChildrenLocked` accessed `parent.ID()` even though the caller already held `parent.mu`.  
- `Inode.ID()` acquires an `RLock` on the same mutex, so the goroutine blocked on itself and the test never progressed.  
- Because `GetChildrenID` was in the middle of rebuilding the root cache, the worker goroutine (and the entire test) stuck until the testing timeout fired.

## Changes

1. Capture the parent's ID while the caller still holds `parent.mu` and pass the raw value through the loop to avoid re-locking.  
2. Cache `inode.ID()` once per iteration before editing the parent's `children` slice to avoid repeating redundant lock acquisitions.  
3. Re-ran `go test ./cmd/common` (focused and full package) to confirm the fixture now completes in <100ms.

## Verification

- `go test ./cmd/common -run TestUT_CMD_01_01_XDGVolumeInfo_VirtualFileBehavior -count=1 -timeout 120s`
- `go test ./cmd/common -count=1`

## Follow-Ups

- None required; the virtual file helper no longer re-enters the parent's lock.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-GUIDE-Coding-Standards (priority 100)  
- AGENT-RULE-Documentation-Conventions (priority 20)
