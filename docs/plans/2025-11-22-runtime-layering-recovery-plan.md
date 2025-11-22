# Runtime Layering Recovery Plan (2025-11-22)

**Type**: Plan  
**Status**: Ready for execution  
**Owner**: Runtime Layering team  
**Last updated**: 2025-11-22  
**Purpose**: Close the remaining gaps in items 1‑7 of the 2025-11-18 plan so FUSE is strictly local-first and the metadata state machine governs hydration, uploads, eviction, and notifier cadence.

## Situation Summary
- `metadata_v2` coexists with legacy buckets; `SerializeAll` still writes both and `loadLegacyMetadataEntry` remains in `internal/fs/metadata_store.go`.
- `inode.hasChanges` and ad-hoc flag mutations still control eviction/offline/conflict logic instead of the state manager; `applyDelta` edits inodes directly.
- FUSE entry points (`GetChildrenID`, mkdir/unlink/rename) still make synchronous Graph calls and lack a mutation queue, breaking offline readiness.
- Sync/hydration paths do not persist snapshots or `LastError`, and delta reconciliation bypasses validated state transitions.
- ChangeNotifier health is not driving delta cadence; telemetry and DBus/stats lack notifier health and metadata queue visibility; promised configs/README updates are missing.

## Goals
- Eliminate legacy metadata paths; make `metadata_v2` the single source of truth.
- Ensure every mutation and reconciliation flows through the state manager.
- Guarantee FUSE callbacks never block on Graph.
- Align sync/hydration, notifier cadence, config, telemetry, and tests with `.kiro/specs/system-verification-and-fix/{requirements,design}.md`.

## Workstreams and Deliverables
1) **Metadata schema refactor**  
   - Remove legacy bucket reads/writes; add one-time migrator + validator CLI that repairs or flags corrupt rows.  
   - Update `SerializeAll`/load paths to emit only `metadata_v2`; delete `loadLegacyMetadataEntry`.  
   - Evidence: migration log shows row counts; validator exposes `--check` output; unit test covering migrate/validate; DoD met when only `metadata_v2` persists and legacy helpers are gone.
2) **State machine everywhere**  
   - Replace `hasChanges`/inline flag flips with `StateManager.transitionItemState`; ensure eviction, conflict/offline, uploads/hydration, and `applyDelta` route through it.  
   - Add table-driven tests for allowed/blocked transitions (hydrating→hydrated, ghost→hydrating, pinned→hydrating, error paths).  
   - Evidence: no direct inode state mutation in grep; state graph unit tests green; DoD met when mutation helpers only accept state-manager calls.
3) **Local-first FUSE**  
   - Strip synchronous Graph calls from FUSE entry points; introduce mutation queue for create/unlink/rename; `GetChildrenID` returns cached view and enqueues background fetch only.  
   - Add regression proving `Lookup/OpenDir/GetChildrenID` succeed offline when metadata exists; ensure hot path uses cached metadata + queued background fetches.  
   - Evidence: tests demonstrate offline success; code search shows no Graph calls in FUSE handlers; DoD met when queue + regressions land.
4) **Sync & hydration pipeline**  
   - Walker/delta writes via `metadata_v2` + state manager; persist hydration/upload snapshots and `LastError` in metadata entries; requeue pinned/invalidated items deterministically.  
   - Extend delta tests to cover invalidation/deletion/pinned requeue and snapshot persistence.  
   - Evidence: metadata rows include snapshot + last error fields after hydration/upload; test suite covers requeue/error cases.
5) **ChangeNotifier cadence**  
   - Drive delta-loop interval from notifier health with a 10s recovery window; log and expose degraded state via stats/DBus.  
   - Add stubbed tests for healthy/degraded paths; telemetry surfaces notifier state transitions.  
   - Evidence: stats/DBus show notifier health; logs contain transition messages; cadence shortens during degraded tests.
6) **Config & telemetry**  
   - Add per-mount knobs for hydration queue thresholds, notifier fallback, overlay policy; validate defaults; document in CLI/README/config reference.  
   - Expose metadata queue depth/latency and notifier heartbeat in `onemount --stats`; ensure config validation rejects invalid ranges.  
   - Evidence: CLI help + README updated; `--stats` output shows new metrics; validation tests cover bad inputs.
7) **Testing & documentation closure**  
   - Add regressions: (a) cached `GetChildrenID` avoids Graph, (b) mutation state graph, (c) delta cadence respects notifier health.  
   - Update `docs/updates/2025-11-19…21` from “Complete” to accurate status with links to new tests/results; refresh `docs/updates/index.md`.  
   - Evidence: new tests pass; updates docs list test names/paths and current status.

## Execution Order & Milestones
1. Metadata refactor → 2. State machine retrofit → 3. FUSE local-first + mutation queue → 4. Sync/hydration fixes → 5. Notifier cadence → 6. Config/telemetry → 7. Tests + documentation closure.  
- **Milestone A** (end of steps 1–3): FUSE operates local-first with state manager enforcing mutations.  
- **Milestone B** (after steps 4–5): Sync/hydration and notifier cadence aligned; telemetry visible.  
- **Milestone C** (after steps 6–7): Config/docs/tests finalized and updates files corrected.

## Engineering Instructions
- Keep edits minimal and localized; favor helpers over inline changes.  
- All state changes must go through `metadata.StateManager`; no raw inode flag mutations.  
- FUSE paths return immediately from cached metadata; Graph work is strictly background.  
- When adding config, update validation, defaults, CLI help, and docs together.  
- For hydrations/uploads, always persist snapshots and `LastError` in `metadata_v2` rows.

## Testing Plan
- Unit: table-driven state transition tests; notifier health cadence tests with fakes; validator/migrator tests; queue depth/latency metrics surfaced.  
- Integration: targeted `go test ./internal/fs -run "<new test|affected suite>" -count=1` for FUSE offline regressions and delta requeue paths.  
- Performance/long: keep opt-in via build tags or `-short` guards; add minimal perf sampling around new queue/telemetry if needed.

## Review Checklist
- No synchronous Graph calls in FUSE entry points.  
- All state changes go through `transitionItemState`; legacy metadata helpers removed.  
- `metadata_v2` is the only persisted store; migrations/validator documented.  
- New config options validated and documented; stats/DBus show notifier health + metadata queue metrics.  
- Tests added/updated and referenced in updates docs.

## Progress Logging
- After each completed workstream, add a short entry to `docs/updates/` (date-slug) summarizing changes and linking to tests.  
- Update `docs/updates/index.md` once items 1‑7 close.

## Definition of Done (overall)
- All DoD bullets for items 1‑7 met with evidence recorded in `docs/updates/`.  
- Unit + targeted integration suites pass with `-count=1`.  
- Relevant docs (plan, config/README, updates index, dated updates) reflect the final state.  
- On a cold mount, `ls`/`cat` rely solely on local metadata; notifier health drives delta cadence per requirements.
