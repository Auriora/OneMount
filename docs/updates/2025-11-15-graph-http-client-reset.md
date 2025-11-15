# Graph HTTP Client Reset Fix

**Date**: 2025-11-15  
**Type**: Bugfix / Testing Stability  
**Components**: internal/graph, internal/fs tests  
**Status**: Complete

## Summary

- Stored the shared `http.Client` instance in a package-level variable so `SetHTTPClient(nil)` reliably repopulates the Graph transport after mock cleanups; fixes the `httpClient.Do` nil panic seen in `TestStatsPagination`.
- Taught `MockGraphClient.Cleanup` to only release the HTTP client when it still owns the override, preventing concurrent download tests from falling back to the real network mid-run.
- Updated `TestStatsPagination` to call `ensureMockGraphRoot`, ensuring the filesystem bootstraps from the Graph mock rather than the real network and preventing offline test failures.
- Relaxed `DownloadSession.canResumeDownload` to key off chunk progress (not download URLs) and backfilled `DownloadSession.Size` after each transfer so multi-download tests observe accurate progress metadata.
- Verified the scenarios by re-running the targeted tests with a workspace-local `GOCACHE`.
- Hardened `Filesystem.GetPath`/`GetChildrenPath` to reject empty paths and bubble errors consistently, which unblocked `TestUT_FS_Path_02_PathValidation_InvalidPaths`.
- Serialized the download-manager integration fixtures (removed `t.Parallel`) and updated `TestUT_UR_02_02` to the new chunk-based resume semantics so the shared Graph mock no longer flips mid-suite.

## Testing

- `GOCACHE=$(pwd)/.gocache go test ./internal/fs -run TestStatsPagination -count=1`
- `HOME=$(pwd)/.home GOCACHE=$(pwd)/.gocache LOG_LEVEL=error go test ./internal/fs -run TestUT_FS_10_DownloadManager_ChunkBasedDownload_LargeFile -count=1`
- `HOME=$(pwd)/.home GOCACHE=$(pwd)/.gocache LOG_LEVEL=error go test ./internal/fs -run TestUT_FS_11_DownloadManager_ResumeDownload_InterruptedTransfer -count=1`
- `HOME=$(pwd)/.home GOCACHE=$(pwd)/.gocache LOG_LEVEL=error go test ./internal/fs -run TestUT_FS_12_DownloadManager_ConcurrentDownloads_QueueManagement -count=1`
- `GOCACHE=$(pwd)/.gocache go test ./internal/fs -run TestUT_FS_Path_02_PathValidation_InvalidPaths -count=1`
- `GOCACHE=$(pwd)/.gocache go test ./internal/fs -run 'Test(UT_FS_1(0|1|2)|IT_FS_08_0[1-5]|UT_UR_02_02)' -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Testing-Conventions (priority 25)
- AGENT-RULE-Documentation-Conventions (priority 20)
