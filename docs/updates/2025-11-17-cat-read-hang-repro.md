# Cat Read Hang – Container Reproduction (2025-11-17)

**Type**: Investigation / Live Repro  
**Status**: In Progress  
**Component**: Filesystem runtime (FUSE, inode cache)  
**Owner**: AI pair (Codex)  
**Log Artifacts**: `/tmp/onemount-run.log`, goroutine dump from `kill -QUIT 449961`

## Test Matrix

| # | Scenario | Commands | Result | Notes |
|---|----------|----------|--------|-------|
| 1 | Fresh mount with debug logging | `LOG_LEVEL=debug build/onemount ~/OneMountTest > /tmp/onemount-run.log 2>&1 &` | Mount succeeded (`fuse.onemount` visible via `mount`) | Confirms auth/config already present in container. |
| 2 | Directory listing sanity check | `ls ~/OneMountTest | head` | Succeeds immediately | Displays expected mix of `Documents`, `dbus-test-file.txt`, and stress files. |
| 3 | Reproduce `cat` hang | `timeout 15s cat ~/OneMountTest/dbus-test-file.txt` | Prints file content but never exits; `timeout` kills after ~124 s | Log shows file served entirely from cache (no network I/O) at `2025-11-16T20:06:21Z`. |
| 4 | Reproduce `ls` hang | `cd ~/OneMountTest && pwd && ls | head` (wrapped in `timeout 120s`) | Hangs until timeout | Triggers another `OpenDir /` at `20:09:04Z` that never completes. |
| 5 | Capture diagnostics | `kill -QUIT <pid>` while hang active | Generates stack dump appended to `/tmp/onemount-run.log` | Unmounted afterward with `fusermount3 -u ~/OneMountTest`. |

## Key Findings

1. **Hang occurs without systemd involvement** – Running the current `build/onemount` binary directly inside the devcontainer reproduces the user-visible freeze for both `cat` and `ls`, ruling out service-unit flapping as a prerequisite.
2. **Reads are served from cache but syscall never returns** – At `2025-11-16T20:06:21Z` the log sequence for `/dbus-test-file.txt` shows `Found content in cache`, `Read`, `Flush`, and `Fsync` completing successfully. The client-side process still blocks forever, indicating the FUSE reply is not sent once the subsequent deadlock hits.
3. **Directory iteration wedges in `GetChildrenID`** – When the `ls` reproduction runs (~`20:09:04Z`), logs show `Checking if children are already cached` followed by `Children found in cache, retrieving them` for the root directory with `childCount=272`; the corresponding `LogMethodExit` never appears.
4. **Stack dump captures the stuck lock ordering** – `goroutine 1` (the FUSE dispatcher) is blocked in `(*Filesystem).getChildrenID` while holding the parent inode’s `RLock` and trying to call `child.Name()` (which acquires the child inode’s `RLock`). Other goroutines (e.g., delta loop serialization, metadata walkers) concurrently hold child locks and attempt to touch parent metadata, producing a classic read-lock inversion. Once this happens, every foreground request (`cat`, `ls`, etc.) is starved.
5. **Download manager is idle** – No `QueueDownload` or `WaitForDownload` entries appear around the failures, confirming the hang is unrelated to stalled blob transfers; it’s purely an inode/metadata locking issue.

## Next Steps

1. **Fix lock ordering in `getChildrenID`**  
   - Copy `inode.children` under the parent lock, release it, then resolve each child ID outside the parent critical section.  
   - Audit other callers (e.g., `refreshChildrenAsync`, serialization) to ensure they follow a consistent ordering (parent before child) or avoid nested locks entirely.

2. **Add contention instrumentation**  
   - Wrap inode `RWMutex` acquisition with debug counters/timers (behind `LOG_LEVEL=debug`) so we can detect when Name()/Path() calls block for >N ms and log the offender.

3. **Regression tests**  
   - Introduce a concurrency stress test that repeatedly enumerates directories while another goroutine serializes the tree (mirroring the delta loop) to ensure future refactors do not reintroduce the deadlock.  
   - Consider enabling Go’s `-race` build for critical suites touching inode locking.

4. **Document & monitor**  
   - Reference this update in `docs/updates/index.md` once the fix lands.  
   - Keep `/tmp/onemount-run.log` artifacts until the locking fix is merged for regression analysis.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-GUIDE-Operational-Best-Practices (priority 40)

