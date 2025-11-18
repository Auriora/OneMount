# Engine.IO Transport Module

**Date**: 2025-11-17  
**Type**: Feature  
**Component**: Realtime Notifications / Socket.IO  
**Status**: Complete

## Summary
- Replaced the legacy emitter-style Socket.IO client with a first-class `RealtimeTransport` interface and `EngineTransport` implementation that performs Engine.IO v4 handshakes, heartbeats, exponential backoff with jitter, and structured packet/message tracing.
- Added a `FakeTransport` plus deterministic helper utilities so unit and integration tests can simulate health transitions and notification bursts without a live Graph endpoint.
- Refactored `SocketSubscriptionManager` to depend solely on the new interface, surface transport health to the delta loop, log typed lifecycle events, and emit notifications via the shared channel.
- Updated the Socket.IO module tests to cover URL generation/backoff math and introduced manager-level tests that assert realtime notifications toggle delta scheduling as expected.
- Surfaced realtime health telemetry in `onemount --stats` (mode, status, heartbeats, reconnect count) and added a `--polling-only`/`webhook.pollingOnly` override for troubleshooting sessions that must disable the Engine.IO channel.

## Rules Consulted
- `AGENT-GUIDE-General-Preferences` (priority 50) – repository-wide change coordination and transparency requirements.
- `AGENT-GUIDE-Planning-Protocol` (priority 30) – followed the approved step-by-step plan for multi-file feature work.
- `AGENT-RULE-Documentation-Conventions` (priority 20) – ensured change history is recorded under `docs/updates/` and linked from the index.

## Testing
1. `LOG_LEVEL=debug go test ./internal/socketio`
2. `LOG_LEVEL=debug go test ./internal/fs -run SocketSubscription`

## Follow-Up
- Plumb the transport health snapshot into CLI diagnostics (e.g., `onemount --stats`) so operators can inspect realtime channel state without enabling trace logs.
- Consider exposing a polling-only toggle for troubleshooting sessions where the Engine.IO channel must be disabled temporarily.
