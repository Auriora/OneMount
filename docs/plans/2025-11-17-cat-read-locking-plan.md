
# Cat Read Hang Locking Remediation Plan

## 1. Background & Findings
- `docs/updates/2025-11-17-cat-read-hang-repro.md` captures a reproducible hang where `cat`/`ls` never return because `getChildrenID` holds the parent inode `RLock` while calling child accessors, leading to parent↔child lock inversion.
- Multiple helpers (`InsertID`, `DeleteID`, `cacheChildrenFromMap`, parent-child slice iteration in `file_operations.go`) share the same pattern, so fixing only `getChildrenID` will not stabilize the system.
- Instrumentation from `docs/updates/2025-11-15-deadlock-harness.md` and the new guidance added to `docs/guides/developer/concurrency-guidelines.md` demonstrate that we need both code changes and better observability to prevent regressions.

## 2. Objectives
- Eliminate parent-child lock inversion across the filesystem cache layer.
- Provide lightweight diagnostics that flag long-held inode locks during debug builds.
- Add automated coverage that exercises directory enumeration while metadata refreshes mutate child lists.
- Preserve documented lock-order exceptions by keeping their critical sections non-overlapping and clearly commented.

## 3. Scope
Included:
1. Refactors inside `internal/fs/cache.go`, `internal/fs/file_operations.go`, and related helpers that touch parent/child slices or call child methods while holding parent locks.
2. Debug-only lock timing instrumentation around high-contention inode locks (parent and child).
3. Concurrency tests and/or extensions to `internal/fs/concurrency_test.go` using the `deadlockMonitor`.
4. Documentation updates (if new rules emerge) plus summary entry in `docs/updates/`.

Excluded:
- Any redesign of the metadata request manager or delta loop scheduling.
- Performance tuning outside of the lock-order fixes (e.g., cache eviction policies).

## 4. Detailed Plan
1. **Audit Current Lock Sites**
   - Enumerate every `parent.mu` usage in `internal/fs/cache.go`, `internal/fs/file_operations.go`, `internal/fs/virtual_files.go`, etc.
   - Label each location as **safe**, **needs snapshotting**, or **needs reordering**. Record notes in the task branch (or `docs/updates` entry) for traceability.

2. **Introduce Snapshot Helpers**
   - Add internal helpers (e.g., `snapshotChildIDs(parent *Inode) []string`) that copy the `children` slice while holding `parent.mu` and release the lock before iterating.
   - Update `getChildrenID`, `cacheChildrenFromMap`, `InsertID`, `DeleteID`, and any other sites identified during the audit to use the snapshot approach and avoid calling `child.Name()`/`child.IsDir()` under the parent lock.

3. **Normalize Parent Mutation Blocks**
   - Ensure each mutation of `parent.children` and `parent.subdir` performs only the minimum required work inside the lock (append IDs, update counts) with all expensive operations computed beforehand.
   - Document the lock-order expectations inline (referencing `docs/guides/developer/concurrency-guidelines.md`) where exceptions still exist.

4. **Add Debug Instrumentation**
   - Wrap parent and child lock acquisition in debug-gated timing (e.g., `time.Since(start)`) that logs when a lock is held longer than a small threshold. Keep the instrumentation behind `logging.IsDebugEnabled()` or a build tag to avoid runtime overhead.
   - Emit structured log fields (`lock=parent`, `hold_ms`, `path`) so the data is easy to correlate with hangs.

5. **Extend Concurrency Tests**
   - Enhance `internal/fs/concurrency_test.go` (or create a new test) to repeatedly call `GetChildrenID` while another goroutine triggers `cacheChildrenFromMap` and deletions, using the existing `deadlockMonitor` to detect stalls.
   - Run targeted stress cases via `go test ./internal/fs -run TestDeadlockPrevention -timeout 5m` and capture monitor snapshots if the test trips.

6. **Manual Reproduction & Verification**
   - Rebuild (`GOFLAGS='' go build ./cmd/onemount`) and run the devcontainer repro from the update doc (`timeout 15s cat ~/OneMountTest/dbus-test-file.txt`, repeated `ls`) to confirm the hang no longer occurs.
   - Capture `/tmp/onemount-run.log` before/after for documentation.

7. **Documentation & Notes**
   - Add/extend the relevant `docs/updates/` entry summarizing the fix, instrumentation, and tests.
   - If new locking rules emerge during implementation, update `docs/guides/developer/concurrency-guidelines.md` accordingly.

## 5. Testing & Validation Checklist
- [ ] `go test ./internal/fs -run TestGetChildrenID` (or targeted unit tests touching cache helpers).
- [ ] `go test ./internal/fs -run TestDeadlockPrevention -timeout 5m`.
- [ ] `go test ./internal/fs -run TestHighConcurrencyStress -timeout 5m`.
- [ ] `go test -race ./internal/fs`.
- [ ] Manual devcontainer repro steps (cat/ls) succeed without timeouts.

## 6. Risks & Mitigations
| Risk | Mitigation |
| --- | --- |
| Refactor introduces performance regressions due to extra allocations when snapshotting child slices. | Use `make([]string, len(children))` + `copy` and consider pooling (`sync.Pool`) if profiling shows regressions; verify with benchmarks before and after. |
| Instrumentation generates noisy logs in production. | Gate logging behind `LOG_LEVEL=debug` and ensure zero-cost when disabled. |
| Tests still miss certain lock patterns. | Incorporate randomization in the new concurrency test and leverage the `deadlockMonitor` stack dumps to tune scenarios iteratively. |
| Existing documented exceptions accidentally gain overlapping critical sections. | During code review, double-check that inode and filesystem locks are never held simultaneously in the exception helpers; update the guide if unavoidable. |

## 7. References
- `docs/updates/2025-11-17-cat-read-hang-repro.md`
- `docs/updates/2025-11-15-deadlock-harness.md`
- `docs/guides/developer/concurrency-guidelines.md`
