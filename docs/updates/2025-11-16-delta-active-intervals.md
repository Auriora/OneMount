# Adaptive Delta Interval For Foreground Activity

**Date**: 2025-11-16  
**Type**: Enhancement  
**Component**: Filesystem / Delta Loop  
**Status**: Complete

## Summary

- Added an "active" delta tuning window so mounts without Graph webhooks can temporarily poll every 60 seconds after user activity while keeping the 5-minute baseline for idle periods.
- Foreground metadata requests now record activity, and the delta loop automatically logs when it deviates from the configured base interval, honoring Requirement 5.7.
- Configuration gained two new knobs (`activeDeltaInterval`, `activeDeltaWindow`) to customize the faster cadence without editing code.

## Changes

1. `cmd/common/config.go`, `cmd/onemount/main.go`: introduce/validate `activeDeltaInterval/activeDeltaWindow`, plumb them into the filesystem, and add unit tests.
2. `internal/fs/*`: track the last foreground metadata request via the `MetadataRequestManager`, add adaptive logic in `desiredDeltaInterval`, and include unit tests to cover the behavior.
3. `docs/updates/index.md`: reference this entry for Task 27 tracking.

## Testing

- `CGO_CFLAGS=-Wno-deprecated-declarations go test -tags $(bash scripts/detect-glib-build-tag.sh) ./cmd/common`
- `go test ./internal/fs -run "(Subscription|Delta)"`
- `go test ./internal/graph -run Subscription`

## Rules Consulted

- AGENT-GUIDE-Coding-Standards (priority 100)
- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
