# Download Manager Foreground Priority Queue

**Date**: 2026-01-29  
**Time**: 12:30:00  
**Type**: Implementation  
**Status**: ✅ Complete

## Summary

Added a minimal foreground queue to the download manager so user-initiated file opens preempt background hydration without changing session tracking or wait behavior.

## Changes Made

1. **Foreground queue support**
   - Added a dedicated foreground queue and `QueueDownloadWithPriority`.
   - Workers check foreground first, then background, to prioritize user opens.

2. **Open path uses foreground**
   - `Open()` now queues foreground downloads for file opens.

3. **Telemetry snapshot**
   - Queue depth now aggregates foreground + background queues.

## Tests

- `docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests` (failed: `TestSystemST_Auth_01_01_InteractiveAuthentication` attempted interactive auth and returned `invalid_grant`)
- `docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner go test -v -short -run '^TestUT_' ./...` (pass)

## Files Touched

- `internal/fs/download_manager.go`
- `internal/fs/download_manager_types.go`
- `internal/fs/file_operations.go`

## Rules Applied

Rules consulted: `.kiro/steering/general-preferences.md` (Priority 50), `.kiro/steering/operational-best-practices.md` (Priority 40), `.kiro/steering/documentation-conventions.md` (Priority 20), `docs/guides/ai-agent/AGENT-GUIDE-General-Preferences.md` (Priority 50), `docs/guides/ai-agent/AGENT-GUIDE-Operational-Best-Practices.md` (Priority 40), `docs/guides/ai-agent/AGENT-RULE-Documentation-Conventions.md` (Priority 20) — Rules applied: minimal scoped edits, documentation update in docs/updates — Overrides: None
