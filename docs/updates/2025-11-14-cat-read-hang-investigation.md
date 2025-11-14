# Cat Read Hang Investigation

**Date**: 2025-11-14  
**Type**: Investigation  
**Component**: Filesystem Runtime / Download Pipeline  
**Status**: In Progress

## Summary

Initial review of the most recent user-space run (`onemount-1411250639.log`) shows only two files were downloaded during the session (`.xdg-volume-info` and `Getting started with OneDrive.pdf`), yet the service was terminated and restarted by systemd roughly every 90 seconds. Each shutdown attempt failed with `fusermount3: Device or resource busy`, implying open FUSE handles (likely the `cat` process the user mentioned) continued to block the mount during restarts.

## Findings

1. **Limited file activity before the hang** – The log only records downloads for `.xdg-volume-info` and `/Getting started with OneDrive.pdf`, and both completed successfully (`onemount-1411250639.log:4082-19084` and `880678-881857`). No other `Open`/`Read` entries appear for the `e2e_*` stress files the user attempted to `cat`, which suggests the problematic access happened after the captured log window or while the service was already shutting down.
2. **Repeated forced restarts** – Systemd delivered `SIGTERM` to the mount roughly every 90 seconds (`onemount-1411250639.log:1078763`, `1100283`, `1119452`, `1138647`, `1157297`, `1176492`, `1195885`). Each shutdown reports `fusermount3: failed to unmount ... Device or resource busy` (`onemount-1411250639.log:1100311-1100313`, `1119479-1119481`, `1138674-1138676`, `1157324-1157326`, `1176519-1176521`, `1195913-1195915`), confirming open file handles—consistent with a `cat` that never completed.
3. **Download wait loop cannot be cancelled** – `WaitForDownload` simply polls until a session flips to `downloadCompleted` or `downloadErrored` with no timeout or context hook (`internal/fs/download_manager.go:528-554`). If `processDownload` blocks (e.g., network stall), the reader thread blocks indefinitely and `cat` appears frozen.
4. **Network requests lack caller-provided context** – `graph.GetItemContentStream` issues metadata and content downloads via `graph.Get(...)`, which always uses `context.Background()` (`internal/graph/drive_item.go:69-118`). The only guard is the global HTTP client timeout (60s); when Graph or the Azure blob endpoint stalls longer than systemd's `TimeoutStopSec`, the request keeps running even while the service is trying to exit, leaving the mount busy.

## Live Reproduction (Nov 14, 2025 @ ~11:30 GMT)

- `systemctl --user start onemount@home-bcherrington-OneMountTest.service` never reaches **active**; it sits in `activating` until systemd emits `start operation timed out. Terminating.` and then kills the mount (latest sample at `Nov 14 11:31:24`, `11:33:37`, etc.). Each termination triggers the familiar `fusermount3: ... Device or resource busy` warnings because a user-space client still has the filesystem open.
- While the unit is flapping, `timeout 10s cat /home/bcherrington/OneMountTest/e2e_stress_w0_op0.txt` consistently exits with code 124 after ~45s, reproducing the user-visible hang: FUSE stops responding once systemd sends SIGTERM, so `cat` sits in `D` state until the entire service restarts.
- `busctl --user list` shows only dynamically generated names such as `org.onemount.FileStatus.instance_2890645_2674`; the canonical `org.onemount.FileStatus.mnt_home-bcherrington-OneMountTest` that the systemd unit waits for is never registered. As a result, Type=dbus jobs never report readiness.
- The installed binary reports `onemount v0.1.0rc1 e86524e7` and logs `Using unique D-Bus service name` even for `--version`. This predates the deterministic service-name work that the repo documents (`docs/updates/2025-11-13-dbus-service-discovery-fix.md`) and therefore cannot satisfy the newer systemd unit.
- The rendered systemd unit at `~/.config/systemd/user/onemount@.service` still sets `BusName=org.onemount.FileStatus.mnt_%i`, so there's a hard contract mismatch: the binary never owns that name, and systemd consequently restarts it every `TimeoutStartSec` even when the filesystem is otherwise healthy.

**Implication**: The `cat` hang is a symptom of the start-timeout loop, not a content-download stall. Any foreground read will be interrupted every ~90 seconds when systemd tears down the daemon, and the open FD keeps the FUSE unmount busy, feeding back into the restart storm.

**Immediate Mitigations**:

1. Rebuild / reinstall `onemount` from the current repo (which calls `SetDBusServiceNameForMount`) or cherry-pick that change into the deployed binary, then restart the user service. Once the daemon registers `org.onemount.FileStatus.mnt_<escaped mount>`, systemd will consider the start successful and stop killing it mid-read.
2. As a stop-gap, edit the user unit to either (a) remove `Type=dbus`/`BusName=` (fall back to `Type=notify` or `simple`) or (b) set `BusName=org.onemount.FileStatus.instance_%i` to match the running build. This at least prevents spurious restarts while the binary remains old.

After the unit/binary mismatch is addressed we can re-run the reproduction to assess whether any genuine download deadlocks remain.

## Next Steps

1. Capture a new log with `ONEMOUNT_DEBUG=1` while reproducing the hang on a specific file, then preserve both the onemount log and `journalctl --user -u onemount@<mount>` output for correlation.
2. Propagate cancellation: pass a `context.Context` with deadline from the filesystem operation into the download manager so `WaitForDownload` can abort when the caller disconnects or when shutdown begins.
3. Add progress/timeouts for large downloads (e.g., per-chunk deadlines plus a watchdog in `WaitForDownload`) and surface that status via `onemount --stats` or D-Bus so users see that `cat` is waiting on network I/O rather than silently hanging.
4. Consider lowering systemd's `TimeoutStopSec` or pausing restarts while readers are active to avoid repeated `Device or resource busy` loops when users manually restart.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-GUIDE-Planning-Protocol (priority 30) – evaluated applicability; full protocol not invoked because the task was diagnostic rather than a multi-file implementation.
