# Socket.IO Delta Notifications

**Date**: 2025-11-17  
**Type**: Feature  
**Component**: Delta Loop / Subscription Manager  
**Status**: Complete

## Summary
- Added delegated Socket.IO change-notification support so foreground mounts no longer need an externally reachable HTTPS webhook.
- Introduced a shared subscription-manager interface that now supports either webhook + HTTP server or Socket.IO streaming.
- Copied the permissively licensed Ali IoT socket.io client into `internal/socketio/` and tweaked it for arbitrary Graph endpoints (path preservation, TLS/wss forcing, Engine.IO v4 parameters).
- Extended the user config (`webhook.useSocketIo`) and default YAML to enable the new transport without breaking existing webhook deployments.

## Testing
1. `LOG_LEVEL=debug go test ./cmd/common -run WebhookConfig -count=1`
2. `LOG_LEVEL=debug go test ./internal/graph -run SocketSubscription -count=1`
3. Manual: `useSocketIo: true` in config, run `build/onemount ~/OneMountTest` with delegated tokens and observe `Socket.IO channel connected` logs plus immediate delta wakeups on `console` writes to OneDrive (no webhook server started).

## Follow-Up
- Add integration coverage for Socket.IO once we expand the mock Graph harness to emit engine.io frames.
- Consider surfacing basic telemetry (connect time, reconnect count) via `onemount --stats` so users can tell when the socket channel is unhealthy.
