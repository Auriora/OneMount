# Download Recovery Requeue Plan (2025-11-16)

**Summary**

- Restored download sessions remain in the `downloadQueued` state and are never re-enqueued, causing every `Open()` that triggers `WaitForDownload()` to hang indefinitely (observed via `cat ~/OneMountTest/etag_304_test.txt`).
- Fix requires re-enqueueing recovered sessions, teaching `QueueDownload` to handle existing-but-stuck sessions, and adding regression tests to cover both flows.

**Code Changes**

1. `internal/fs/download_manager.go`
   - In `restoreDownloadSessions()`, after inserting each recovered session into `dm.sessions`, call a new helper (e.g., `enqueueSession(id, session)`) that pushes IDs back onto `dm.queue` unless the session is already `downloadStarted`/`downloadCompleted`.
   - Update `QueueDownload()` so that when an existing session is found:
     - `downloadCompleted` → return as-is (no enqueue).
     - `downloadStarted` → return as-is (workers already running).
     - `downloadQueued` or `downloadErrored` → pass through `enqueueSession` to restart processing.
   - Implement `enqueueSession` with the same non-blocking semantics (log and keep going if queue is full) so both the restore path and duplicate `QueueDownload` calls share logic.

**Tests**

Add regression coverage in `internal/fs/download_manager_test.go`:

1. `TestDownloadManager_RestoredSessionsAreRequeued`
   - Seed the Bolt DB with a `downloadQueued` session.
   - Instantiate `DownloadManager`; expect it to enqueue the session and drive it to `downloadCompleted` via existing hooks/mocks.

2. `TestDownloadManager_QueueDownload_RequeuesExistingQueuedSession`
   - Create a manager, queue a file, manually reset its state to `downloadQueued`, call `QueueDownload(id)` again, and assert the helper attempts another enqueue (can be verified by consuming from a test queue channel or exposing a counter).

**Validation Checklist**

1. Restart `onemount@home-bcherrington-OneMountTest.service` and confirm the journal shows `File queued for download` followed by `File download completed` for previously stuck files.
2. Run `timeout 5s cat ~/OneMountTest/etag_304_test.txt` and `ls ~/OneMountTest` to ensure they complete immediately.
3. Reboot the service again to verify restored sessions now drain automatically (no more infinite `WaitForDownload`).

**Logging Level Reminder**

- When running tests, set `LOG_LEVEL` (e.g., `LOG_LEVEL=error go test ./internal/fs -run TestDownloadManager_RestoredSessionsAreRequeued`) per `docs/updates/2025-11-15-graph-http-client-reset.md`.

