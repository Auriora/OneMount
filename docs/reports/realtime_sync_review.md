# OneMount Real-time Synchronization Review

OneMount goes beyond traditional polling sync clients by utilizing **Microsoft Graph's Socket.IO subscriptions**, integrated into the `internal/socketio` and `internal/graph` packages.

## The Problem with Polling
Polling a remote API every few minutes consumes network bandwidth, battery, and causes frustrating delays before files updated on another device reflect locally. Microsoft Graph offers webhook subscriptions, but webhooks require a publicly accessible URL, which is not feasible for a desktop Linux client behind a NAT/firewall.
As a workaround for clients, Microsoft Graph natively supports an undocumented/lightly-documented Socket.IO subscription endpoint that initiates a long-lived WebSocket connection.

## Socket.IO Implementation
Because mature native Go Socket.IO clients are scarce and often unmaintained, OneMount implements its own minimal Engine.IO v4 over WebSocket client internally (`internal/socketio/engine_transport.go` & `protocol/`).

1. **Endpoint Resolution:** `internal/graph/socket_subscription.go` queries the Microsoft Graph API (`/subscriptions/socketIo`) to receive a unique `notificationUrl`.
2. **EngineTransport:** The `EngineTransport` struct connects to this WebSocket endpoint. It strictly handles Engine.IO v4 handshakes, ping/pong heartbeats, and packet routing.
3. **Event Emitting:** The transport emits strongly typed payloads via a local listener registry (`listenerRegistry`), firing events like `EventConnected`, `EventNotification`, and `EventHealthChanged`.
4. **Resiliency & Health:**
   - It maintains a `HealthState` which tracks consecutive failures, missed pings, and reconnect attempts.
   - Reconnections utilize an exponential backoff algorithm with jitter to prevent server throttling (`nextBackoffDelay()`).
   - If the WebSocket completely fails or enters a `StatusDegraded` state, the filesystem (`internal/fs`) is notified via `EventHealthChanged`. The filesystem's `DeltaLoop` seamlessly catches this and falls back to traditional HTTP polling until the Socket.IO connection heals.

## Conclusion
This mechanism enables **push-based delta syncs**. When a file is altered remotely, the Graph API pushes a JSON notification frame down the WebSocket. OneMount sees this notification instantly, fires a local delta query, and updates its FUSE caches with the new ETag and file metadata in real-time, drastically reducing access latency without wasting daily bandwidth on polling.
