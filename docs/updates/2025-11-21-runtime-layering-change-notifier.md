# Runtime Layering – ChangeNotifier Facade (2025-11-21)

**Type**: Implementation Update  
**Status**: In Progress  
**Components**: `internal/fs/change_notifier.go`, `internal/fs/socket_subscription.go`, `internal/fs/delta.go`, `docs/2-architecture/resources/delta-sync-sequence-diagram.puml`, `docs/updates/2025-11-17-socketio-realtime.md`

## Summary

- Introduced a dedicated `ChangeNotifier` facade that wraps the Socket.IO subscription manager so the delta loop, stats surface, and telemetry only speak in terms of “socketio” vs “polling-only” modes—no more direct webhook references or transport-specific logic sprinkled across the filesystem.
- Exposed health snapshots from the Socket.IO manager and bubbled them through the notifier, allowing `onemount --stats` and runtime layering instrumentation to report realtime status without peeking into Engine.IO internals.
- Updated the architecture diagram and docs to describe Socket.IO notifications (not webhooks) and aligned the configuration guidance with the `realtime.*` block instead of legacy `webhook.*` flags.

## Testing

- `HOME=/workspaces/OneMount GOCACHE=/workspaces/OneMount/.gocache go test ./internal/fs -run TestChangeNotifier -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Documentation-Conventions (priority 20)
- AGENT-RULE-Testing-Conventions (priority 25)
