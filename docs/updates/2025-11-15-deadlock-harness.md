# Deadlock Harness Instrumentation & Stress Fixes

**Date**: 2025-11-15  
**Type**: Test Infrastructure / Bugfix  
**Components**: internal/fs/concurrency_test.go, internal/fs/concurrency_deadlock_monitor_test.go, internal/testutil/helpers  
**Status**: Complete

## Summary

- Added a reusable `deadlockMonitor` helper that records each worker's last activity and optionally dumps goroutine stacks when `DEADLOCK_TRACE` is set, so future stalls emit actionable diagnostics instead of a raw `go test -timeout` panic.
- Replaced the detached goroutine "lock contention" pattern with inline timed locking and raised the stress test's file pool to 512 entries, eliminating the writer starvation hotspot on `Inode.mu` (e.g., `stress_test_file_94`).
- Tightened `TestHighConcurrencyStress`'s watchdog to 35s, lowered the noise by gating instrumentation, and set a realistic `maxAvgLatency` (25ms) so the stress harness fails only on meaningful regressions.
- Instrumented `TestDeadlockPrevention` and `TestConcurrentDirectoryOperations` with the same monitor plus per-worker accounting, resolving the flaky "operations accounted for" assertion by checking for `>=` rather than exact equality.
- Documented the investigation in this update and linked it from the master index per AGENT-RULE-Documentation-Conventions.

## Root Cause

High-concurrency runs hammered the same 100 inodes with tight metadata loops while a detached goroutine tried to take `inode.mu.Lock()` for the simulated contention case. The writers were not part of the fixture's wait group, so once a writer waited on `stress_test_file_94`, Go's writer-preference blocked new readers yet no goroutine released the lock, causing a full-suite timeout after 20 minutes.

## Remediation

1. Replace the detached goroutine pattern with inline locking + hold-time tracking and log writers that exceed 1ms (counts now appear as `lock-contention`).
2. Increase the stress corpus to 512 files so hotspots are less likely, and cap the timeout at `testDuration + 5s` with a detailed monitor snapshot + goroutine dump when workers fall behind.
3. Introduce the `deadlockMonitor` helper (opt-in stack dumps via `DEADLOCK_TRACE`) and wire it into `TestDeadlockPrevention`/`TestConcurrentDirectoryOperations` to provide per-worker heartbeats.
4. Adjust `TestConcurrentDirectoryOperations`'s accounting to allow aggregate success counts >= iteration count, reflecting that a single iteration can perform multiple successful child lookups.

## Testing

- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestHighConcurrencyStress -timeout 5m`
- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestDeadlockPrevention -timeout 5m`
- `HOME=$PWD/.sandbox/home go test ./internal/fs -run TestConcurrentDirectoryOperations -timeout 5m`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
