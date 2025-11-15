# Directory Listing Cache Refresh

**Date**: 2025-11-14  
**Type**: Bugfix  
**Component**: Filesystem Metadata Cache  
**Status**: Complete

## Summary

- `TestUT_FS_FileRead_03_DirectoryListing` exposed a regression where `GetChild` would keep returning cached results even after new remote children were added through the Graph mock, leading to panics when `InsertNodeID` received a nil inode.
- Added a refresh path to `Filesystem.GetChild` that invalidates and re-fetches the parent directory’s children when the requested name is missing but cached metadata exists, ensuring directory listings stay in sync without waiting for a delta loop.
- Preserves the previous cache snapshot if the refresh attempt fails so offline directory listings remain available.
- Directory opens now short-circuit inside `Filesystem.Open` before touching the content cache or queuing downloads, preventing `ls`-style operations from spawning bogus download sessions or blocking on file-status updates.
- Updated `TestUT_FS_FileRead_03_DirectoryListing` to rely on `LoopbackCache.HasContent` when asserting that no file content was downloaded; previously it used `LoopbackCache.Open`, which created empty files and produced false positives once the test made it past earlier failures.

## Testing

- `HOME=$(pwd)/.home GOCACHE=$(pwd)/.gocache go test ./internal/fs -run TestUT_FS_FileRead_03_DirectoryListing`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Coding-Standards (priority 100)
- AGENT-RULE-Testing-Conventions (priority 25)
