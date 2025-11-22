# Runtime Layering – Config & Telemetry (2025-11-21)

**Type**: Implementation Update  
**Status**: Complete (validated 2025-11-22)  
**Components**: `cmd/common/config.go`, `cmd/onemount/main.go`, `internal/fs/cache.go`, `internal/fs/metadata_store.go`, `internal/fs/download_manager.go`, `internal/fs/stats.go`

## Summary

- Added an `overlay.defaultPolicy` knob in `config.yml` (and CLI defaults) so developers can choose whether new metadata rows favor `REMOTE_WINS`, `LOCAL_WINS`, or `MERGED` overlays without patching code. The filesystem now enforces the selected policy when persisting entries and normalizes invalid values back to `REMOTE_WINS`.
- Surfaced richer telemetry in `onemount --stats`: metadata state histograms (GHOST/HYDRATING/HYDRATED/etc.) and hydration queue insight (active download count + queue depth) pull directly from the structured metadata store and download manager snapshot, closing the observability gap called out in Work Breakdown item 6.
- Hooked the new config + telemetry plumbing together by exposing a `SetDefaultOverlayPolicy` setter on the filesystem, wiring it during initialization, and teaching the stats code to parse `metadata_v2` entries so state transitions are visible without enabling debug logging.

## Testing

- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./cmd/common -run TestValidateConfigOverlayPolicy -count=1`
- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestStatsReportsMetadataStates -count=1`
- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestChangeNotifier -count=1`
- Validation addendum (2025-11-22): `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestDeltaIntervalRespectsNotifierHealth -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
