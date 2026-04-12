# OneMount Offline Conflict Resolution Review

OneMount features a robust, network-resilient offline mode that allows local modifications while disconnected, reconciling them smoothly when connectivity is restored.

## The Reconciliation Flow
When the filesystem detects that network connectivity has returned, the `SyncManager` (in `internal/fs/sync_manager.go`) coordinates the sync process:
1. **Fetch Offline Changes:** It queries the local metadata store for any operations (Create, Modify, Delete, Rename) performed while offline.
2. **Fetch Remote State:** For each changed item, it queries the Microsoft Graph API to determine the current remote state of that file.
3. **Detect Conflicts:** It delegates comparison to the `ConflictResolver` (`internal/fs/conflict_resolution.go`).

## Conflict Detection Rules
The `ConflictResolver` checks for four distinct types of conflicts:
- **Existence Conflicts:** A file was deleted locally but modified remotely, or created/modified locally but deleted remotely.
- **Content Conflicts:** Both local and remote versions were modified, and their `QuickXorHash`es differ.
- **Metadata Conflicts:** The file's name or parent directory was changed concurrently on both sides.

## Resolution Strategies
The resolver supports multiple strategies mapped tightly to user preference or application defaults. By default, `SyncManager` initializes the resolver using **`StrategyKeepBoth`**.
- **Keep Both (Default):** The local modifications are kept and queued for upload. To preserve the remote changes, a new "Conflict Copy" of the remote file is downloaded and inserted into the filesystem side-by-side with the local file, timestamped for clarity.
- **Last Writer Wins:** Compares the local modification time with the remote modification time. If remote is newer, the local cache is invalidated and overwritten; if local is newer, it forces an upload to overwrite the remote logic.
- **Rename:** Currently aliases to the "Keep Both" mechanism.

## Error Handling
The `SyncManager` uses an advanced `retry.Config` package wrapping its reconciliation fetches. It implements exponential backoffs, jitter, and targets specific retryable network/server errors to ensure that intermittent connection drops during the recovery phase do not silently fail or skip user modifications.
