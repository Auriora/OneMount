# Runtime Layering & ChangeNotifier Implementation Plan (2025-11-18)

This document captures the gap analysis and step-by-step plan for aligning the codebase with the newly updated requirements/design specs (runtime-local FUSE operations, metadata state machine, and realtime Socket.IO notifications). No code changes are included here—this is purely planning.

## 0. Review Findings Snapshot

- **FUSE still blocks on Graph** – `dir_operations.go` (Mkdir), `file_operations.go` (`remoteID`), `cache.go:GetChildrenID`, and other handlers invoke Graph synchronously, stalling `readdir/getattr` in violation of Requirement 2.15 and the runtime-layering design.
- **No explicit metadata state machine** – Inodes only wrap `graph.DriveItem`; eviction/hydration toggle ad-hoc booleans (`hasChanges`, pending-remote flags) instead of well-defined `GHOST/HYDRATED/...` states demanded by Requirements 2.16, 3.22‑3.23, 6.*, 8.*, 21.
- **Virtual overlays are scattered** – `.xdg-volume-info` is special-cased in cache fetch logic, pending-local entries aren’t persisted, and overlay precedence isn’t encoded in the DB, contradicting Requirement 2.16 and the “Virtual Items and Overlay Policies” design.
- **Hydration/upload pipeline bypasses shared state** – Cache eviction deletes local files without updating metadata; uploads can run inline from FUSE (`remoteID`), making behavior nondeterministic and hard to test.
- **Change notification layer still references webhooks** – Legacy HTTP listener/config/docs remain even though the intended architecture is Socket.IO-only; requirements/design must drop webhook-specific behavior.
- **Realtime config/logging inconsistent** – Mixture of `webhook.*` flags, HTTPS validation, and CLI/man-page wording implies webhooks; stats report “webhook” mode. Needs consolidation around Socket.IO vs polling.
- **Metadata request manager still synchronous** – `GetChildrenID` blocks on queued fetch results; there’s minimal telemetry for queue depth/latency, so diagnosing “ls” stalls is hard.
- **Graph client usage is ad hoc** – Many packages import `internal/graph` directly without service interfaces, leading to duplicated throttling/backoff logic and harder testing/mocking.
- **Blocking uploads from hot path** – `remoteID` can fully upload a file during `open/create`, which violates local-first responsiveness and can wedge the FUSE thread during slow uploads.
- **Notifier health not surfaced everywhere** – Stats/DBus lack consistent health reporting; fallback decisions aren’t centrally logged, making realtime issues opaque.
- **Additional quality gaps** – Missing metrics for hydration/eviction queues, limited unit coverage of state transitions, and scattered offline/sync logic all hurt maintainability/debuggability.

These findings inform the work breakdown below.

## 1. Inventory & Gaps

| Area | Current State | Gap vs Spec |
| --- | --- | --- |
| Metadata persistence | Inodes store raw drive items with ad-hoc flags | Need explicit `item_state`, `virtual`, `overlay_policy`, pinning metadata, hydration error details (Requirements 2.15‑2.16, 3.22‑3.23, 21; Design §“Item State Model”, §“Virtual Items and Overlay Policies”) |
| FUSE callbacks | `OpenDir`, `GetChildrenID`, mkdir/etc. often call Graph synchronously | Must be strictly local per Requirement 2.15 and Design §“Runtime Layering”; Graph IO belongs to background workers |
| Sync/hydration | Download/upload paths manipulate cache directly without shared state machine | Need deterministic transitions and eviction semantics (Requirements 3.22‑3.23, 6.*, 8.*, 21; Design §“Item State Model”) |
| Change notifications | Socket path coexists with deprecated webhook HTTP server | Spec now requires a single Socket.IO transport with healthy fallback behaviour (Requirements 5.2‑5.14, 17, 20; Design §“Realtime Socket Architecture”, §“Realtime Subscription Component”) |
| Config/logging | No knobs for overlay policies, notifier preference, or state logging | Need CLI/config additions + diagnostics (Requirements 8.6‑8.9, 9.*, 17.2‑17.9; Design §“Runtime Layering” + notification sections) |

## 2. Work Breakdown

1. **Metadata Schema Refactor** *(Requirements 2.15‑2.16, 3.22‑3.23, 21; Design §“Item State Model”, §“Virtual Items and Overlay Policies”)*
   - Define `ItemState` enum + struct representing metadata row.
   - Extend BBolt buckets or add new bucket to store structured entries.
   - Write shims to load existing raw entries and re-materialize them into the new struct (one-time rebuild acceptable per user guidance).

2. **State Machine Implementation** *(Requirements 3.22‑3.23, 6.*, 8.*, 21; Design §“Item State Model”)*
   - Central helper (e.g., `itemstate.Manager`) for atomic transitions + validation.
   - Update hydration, upload, eviction, conflict code paths to invoke helper.
   - Emit tracing/logs + instrumentation for new states.

3. **Local-First FUSE Refactor** *(Requirement 2.15, Design §“Runtime Layering”, §“Virtual Items and Overlay Policies”)*
   - Audit `internal/fs/dir_operations.go`, `file_operations.go`, `lookup.go`, etc. and remove direct Graph calls.
   - Enhance metadata request manager to queue background fetches; FUSE returns cached data immediately and relies on states to decide hydration/refresh.
   - Model virtual files entirely via metadata entries and overlay policies.

4. **Sync & Hydration Pipeline Update** *(Requirements 3.*, 5.*, 6.*, 8.*, 21; Design §“Runtime Layering”, §“Item State Model”)*
   - Delta loop writes state transitions instead of mutating cache inline.
   - Hydration workers respect `HYDRATING` lockouts and update states on success/failure.
   - Eviction and pinning logic manipulates state without deleting metadata.

5. **ChangeNotifier Layer** *(Requirements 5.2‑5.14, 17, 20; Design §“Change Notifier Architecture”, §“Change Notifier Component”)*
	- Introduce/retain a lightweight `ChangeNotifier` facade around the Socket.IO manager (webhook transport removed).
	- Health reporting, fallback to polling, diagnostics, and signaling to the sync loop.
	- Config surface (CLI/env) limited to enabling/disabling realtime + polling-only.

6. **Config & Telemetry** *(Requirements 8.*, 9.*, 17.*, 20.*; Design §“Runtime Layering”, notification sections)*
   - Add options for overlay policy defaults, notifier selection, hydration thresholds.
   - Ensure logging/metrics expose state transitions, notifier health, hydration queue.

7. **Testing & Docs** *(Requirements 3.*, 5.*, 6.*, 8‑10, 17‑20; Design §“Architecture Enhancements”, §“Change Notifier Component”)*
   - Unit tests for state transitions, notifier fallback, hydration/eviction behavior.
   - Integration tests verifying FUSE no longer blocks on network.
   - Update relevant docs once implementation completes.

## 3. Sequencing & Dependencies

1. Metadata schema work must precede state-machine + FUSE refactor (everything reads from new struct).
2. After metadata/state changes, refactor FUSE callbacks and hydration pipeline simultaneously (they share state helpers).
3. ChangeNotifier work can occur in parallel once metadata refactor is underway, but sync loop integration waits until state machine is ready so health signals can drive polling cadence.
4. Config/logging updates happen alongside the functional changes they expose.

## 4. Risks & Mitigations

- **Wide blast radius:** Refactors touch most filesystem components. Mitigate by introducing helper layers (state manager, notifier) to keep call sites simple and testable.
- **BBolt data reset:** Schema overhaul may require wiping old caches; acceptable per user instruction (no backward-compat needed) but document the behavior for developers.
- **FUSE latency regression during transition:** Ensure incremental refactor keeps temporary shims (e.g., metadata fetch queue) so we don’t stall user operations while reworking code.

## 5. Next Steps

1. Draft `ItemState` enum + metadata struct, update BBolt persistence helpers.
2. Design the state-transition helper + integration points (hydration, uploads, eviction).
3. Finalize the Socket.IO ChangeNotifier interface (webhook transport removed) and validate it in isolation before wiring into the delta loop.
4. Review with maintainers before executing refactor.

## 6. Reality Check (2025-11-22)

Field verification against the current code shows the plan items above remain largely unfinished despite “Complete” statuses in the 2025-11-19…21 update notes. Gaps by work breakdown item:

- **1) Metadata schema refactor**: Legacy fallback and dual-bucket writes are still active (`loadLegacyMetadataEntry` in `internal/fs/metadata_store.go`; `SerializeAll` writes legacy buckets). No validator/repair pass exists for `metadata_v2`.
- **2) State machine implementation**: `inode.hasChanges` remains authoritative; write paths don’t consistently call `transitionItemState`; `applyDelta` still mutates inodes directly; conflict/offline/eviction paths bypass the state manager.
- **3) Local-first FUSE callbacks**: `GetChildrenID` still issues synchronous Graph calls when cold; mkdir/unlink/rename call Graph inline; no mutation queue exists; offline-friendly tests only cover a subset of cases.
- **4) Sync & hydration pipeline**: Tree sync still materializes `DriveItem`s into inodes; delta reconciliation does not drive validated state transitions for all cases; hydration/upload snapshots and `LastError` plumbing are absent.
- **5) ChangeNotifier layer**: Delta loop ignores notifier health; 10-second recovery window and degraded-state telemetry are not wired; D-Bus/stats exposure missing.
- **6) Config & telemetry**: Only the overlay default knob landed. No per-mount hydration/notifier tuning or queue metrics surfaced in `--stats`; README/config docs not updated for the promised knobs.
- **7) Testing & documentation closure**: Required regressions (no Graph from GetChildrenID when cached, mutation state graph, delta cadence vs notifier health) are missing; `docs/updates/index.md` and Nov 19–21 update files still declare “Complete” contrary to code state.

## 7. Revised Task List to Complete the Plan

1. **Kill legacy metadata**: Remove legacy bucket reads/writes, add one-time migrator/repair, and wire validator over `metadata_v2`.
2. **State machine everywhere**: Replace `hasChanges` with `transitionItemState`; refit `applyDelta`, eviction, conflict/offline handlers, and uploads/hydration to the state manager with regression tests.
3. **Local-first FUSE**: Strip synchronous Graph from FUSE paths; add mutation queue; ensure `GetChildrenID` returns cached data and only queues background fetches; add offline/blocking regressions.
4. **Sync/hydration correctness**: Run walker/delta through metadata_v2 + state manager; persist hydration/upload snapshots and `LastError`; keep pinned/ghost transitions consistent.
5. **Notifier-driven cadence**: Feed `ChangeNotifier` health into delta loop cadence, implement 10s recovery/backoff logging, expose degraded state via stats/DBus, and add stubbed tests.
6. **Config/telemetry surface**: Add per-mount knobs (hydration queue thresholds, notifier fallbacks, overlay policies), expose metadata-queue/notifier metrics in `--stats`, and document/validate.
7. **Close documentation/tests**: Add the promised regressions, then update Nov 19–21 update files and `docs/updates/index.md` to reflect actual completion with links to tests/results.
