# Runtime Layering Recovery Plan (2025-11-22)

**Type**: Plan  
**Status**: Draft  
**Owner**: Runtime Layering team  
**Purpose**: Close the remaining gaps in items 1‑7 of the 2025-11-18 plan so FUSE is strictly local-first and the metadata state machine governs hydration, uploads, eviction, and notifier cadence.

## Goals
- Eliminate legacy metadata paths; make `metadata_v2` the single source of truth.
- Ensure every mutation and reconciliation flows through the state manager.
- Guarantee FUSE callbacks never block on Graph.
- Align sync/hydration, notifier cadence, config, telemetry, and tests with the specs in `.kiro/specs/system-verification-and-fix/{requirements,design}.md`.

## Work Items (DoD per item)
1) **Metadata schema refactor**  
   - Remove legacy bucket reads/writes; add one-time migrator/validator.  
   - DoD: `SerializeAll` writes only `metadata_v2`; validator tool exists; legacy code paths removed.
2) **State machine everywhere**  
   - Replace `hasChanges`; wire `applyDelta`, eviction, conflict/offline, uploads/hydration to `transitionItemState`.  
   - DoD: State graph covered by unit tests; no direct inode state mutation remains.
3) **Local-first FUSE**  
   - Strip sync Graph calls from FUSE entry points; add mutation queue; `GetChildrenID` returns cached view and only queues background fetch.  
   - DoD: Regression proving `Lookup/OpenDir` succeeds offline when metadata exists; no synchronous Graph in FUSE code.
4) **Sync & hydration pipeline**  
   - Walker/delta write via `metadata_v2` + state manager; hydration/upload snapshots persisted; `LastError` surfaced.  
   - DoD: Delta tests cover invalidation/deletion/pinned requeue; snapshots visible in metadata entries.
5) **ChangeNotifier cadence**  
   - Delta loop uses notifier health for interval; 10s recovery window; degraded state logged and exposed via stats/DBus.  
   - DoD: Unit tests stub healthy/degraded paths; stats show notifier state; logs include transitions.
6) **Config & telemetry**  
   - Add knobs for hydration queue thresholds, notifier fallback, overlay per mount; expose metadata queue depth/latency and notifier heartbeat in `--stats`; docs updated.  
   - DoD: CLI/config validated; `onemount --stats` shows new metrics; README/config docs updated.
7) **Testing & documentation closure**  
   - Add regressions for (a) GetChildrenID no-Graph when cached, (b) mutation state graph, (c) delta cadence respects notifier health.  
   - Update `docs/updates/2025-11-19…21` to “Complete” with links to tests/results and refresh `docs/updates/index.md`.  
   - DoD: Tests green; updates edited with test references.

## Execution Order
1. Metadata refactor (1)  
2. State machine retrofit (2)  
3. FUSE local-first + mutation queue (3)  
4. Sync/hydration fixes (4)  
5. Notifier cadence (5)  
6. Config/telemetry (6)  
7. Final tests + documentation closure (7)

## Engineering Instructions
- Keep edits minimal and localized; favor helper functions over inline changes.  
- Use `metadata.StateManager` for all state transitions; no raw inode flag mutations.  
- Background work only: FUSE paths must return from cached metadata immediately.  
- When adding config, update validation, defaults, and docs together.  
- For hydrations/uploads, always persist snapshots and `LastError` in `metadata_v2`.

## Testing Instructions
- Add/extend unit tests per DoD bullets; prefer existing fixtures (`helpers.SetupFSTestFixture`).  
- Integration: run targeted `go test ./internal/fs -run "<new test|affected suite>" -count=1`.  
- Keep perf/long tests opt-in; mark with build tags or `-short` guards.

## Review Checklist
- No synchronous Graph calls in FUSE entry points.  
- All state changes go through `transitionItemState`.  
- `metadata_v2` is the only persisted store; legacy paths removed.  
- New config options validated and documented.  
- Stats/DBus show notifier health and metadata queue metrics.  
- Tests added/updated and referenced in updates docs.

## Progress Logging
- For each completed item, add a short entry to `docs/updates/` (use the existing date scheme) summarizing changes and linking to tests.  
- Update `docs/updates/index.md` when items 1‑7 close.

## Definition of Done (overall)
- All DoD bullets for items 1‑7 met.  
- Test suite (unit + targeted integration) passes with `-count=1`.  
- Relevant docs (plan, config/README, updates index, dated updates) reflect the final state.  
- On a cold mount, `ls` and `cat` rely solely on local metadata, and notifier health drives delta cadence per requirements.  
