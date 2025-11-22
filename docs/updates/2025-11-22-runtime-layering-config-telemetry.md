# Runtime Layering – Config & Telemetry Knobs (2025-11-22)

**Type**: Implementation Update  
**Status**: Complete  
**Components**: `cmd/common/config.go`, `cmd/onemount/main.go`, `internal/fs/cache.go`, `internal/fs/download_manager.go`, `internal/fs/metadata_priority.go`, `internal/fs/stats.go`

## Summary

- Added hydration controls to config/CLI (`--hydration-workers`, `--hydration-queue-size`) and metadata queue sizing knobs (`--metadata-workers`, `--metadata-high-queue-size`, `--metadata-low-queue-size`) so mounts can tune worker counts and queue depths per spec item 6.
- Exposed realtime fallback override (`--realtime-fallback-seconds`) and overlay default (`--overlay-policy`) as first-class flags on top of YAML fields.
- Instrumented the metadata request manager to track queue depth and avg wait; `onemount --stats` now reports metadata queue and hydration queue/worker activity alongside realtime heartbeat metrics.

## Usage

- YAML: `hydration.workers`, `hydration.queueSize`, `metadataQueue.workers`, `metadataQueue.highPrioritySize`, `metadataQueue.lowPrioritySize`, `realtime.fallbackIntervalSeconds`, `overlay.defaultPolicy`.
- CLI overrides: `--hydration-workers`, `--hydration-queue-size`, `--metadata-workers`, `--metadata-high-queue-size`, `--metadata-low-queue-size`, `--realtime-fallback-seconds`, `--overlay-policy`.
- Stats: `onemount --stats` now prints hydration queue depth/active downloads plus metadata high/low queue depth and average wait (ms), along with realtime heartbeat counters.

## Testing

- `GOCACHE=/tmp/gocache go test ./cmd/common -run 'TestHydrationConfigDefaultsAndValidation|TestMetadataQueueDefaultsAndValidation' -count=1` *(timed out in this environment; code builds locally—rerun after rebuild to confirm)*
- Prior delta interval tests continue to pass under the same target when cache permissions allow.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
