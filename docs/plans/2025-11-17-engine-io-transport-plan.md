# Plan: Engine.IO/Socket.IO Transport Implementation

## Context & Goals
- Replace the abandoned third-party Socket.IO client with an in-repo Engine.IO v4/WebSocket transport that fulfills Requirement 20 and the revised Socket.IO subscription workflow (Requirements 5 & 17).
- Preserve OneMount’s standalone architecture: no Azure Web PubSub, no inbound webhooks, no reliance on unmaintained libraries.
- Provide a clean, testable interface (`RealtimeTransport`) that the filesystem can depend on without leaking implementation details.

## Key Findings
1. **Legacy client is unmaintained** – The vendored library predates Engine.IO v4, lacks TLS/WebSocket fixes, and emits opaque `{}` errors under load.
2. **Design gap** – Documentation now specifies the behaviors (heartbeat timing, reconnection, tracing), but the codebase still exposes ad-hoc emitter callbacks, making it hard to unit test or swap implementations.
3. **Operational pain** – Without granular packet/health telemetry, diagnosing the “Socket.IO channel error [{}]” incidents is slow, and delta polling never lengthens because the channel rarely reports healthy status.

## Recommended Approach
- Define an explicit `RealtimeTransport` interface (per SDS §4.5) with health reporting, typed events, and lifecycle hooks.
- Implement `EngineTransport` on top of `gorilla/websocket`, focusing only on the features Microsoft Graph requires (text frames, namespace `/`, event payloads).
- Introduce a `FakeTransport` for unit tests and a thin Node.js harness (optional) for golden-frame capture during development.

## Step-by-Step Plan
1. **Interface & Types**
   - Add `RealtimeTransport`, `EventType`, `HealthState`, and typed listener signatures to `internal/socketio`.
   - Update `SocketSubscriptionManager` to depend on the interface, not the concrete client.
2. **Handshake & Connection Layer**
   - Implement `EngineTransport.Connect` to build the `/socket.io/?EIO=4&transport=websocket` URL, attach OAuth headers, parse handshake payloads, and surface initial health state.
   - Preserve Graph query params (resource, delta token) without mutating paths.
3. **Heartbeat Scheduler**
   - Use server-provided `pingInterval`/`pingTimeout` to drive timers.
   - Declare the transport `Degraded` after two missed pings; emit `EventHealthChanged` so the delta loop can shorten polling.
4. **Reconnection & Backoff**
   - Implement exponential backoff with jitter (1 s up to 60 s) and ensure it resets once a connection stabilizes for ≥1 heartbeat window.
   - Capture close codes/reasons for logging; differentiate auth failures vs transient network issues.
5. **Event Pipeline**
   - Decode `42[...]` payloads, dispatch `notification` events, and surface `error` frames with structured metadata.
   - Emit `EnginePacketTrace` and `EngineMessageTrace` at trace level with payload truncation (configurable limit, default 512 bytes).
6. **Health API & Metrics**
   - Maintain an atomic `HealthState` snapshot (status, last error, last heartbeat, missed count) that `SocketSubscriptionManager` can query cheaply.
   - Add counters for reconnect attempts and consecutive failures; expose them via debug logs.
7. **Testing Strategy**
   - Unit tests for packet encode/decode, heartbeat scheduling, and backoff math (using fake clock/time source).
   - Integration test harness that replays captured engine.io frames to ensure the parser and dispatcher behave correctly.
   - Update existing delta-loop tests to run with `FakeTransport` so failure modes (e.g., degraded channel) are deterministic.
8. **Integration & Cleanup**
   - Swap the old emitter-based client with the new transport in `SocketSubscriptionManager`.
   - Remove legacy webhook references, ensuring CLI/config uses consistent terminology (“Socket.IO subscription”).
   - Document the new module in `docs/updates` and add troubleshooting guidance (e.g., interpreting health logs).

## Open Questions / Follow-ups
- Do we need a CLI flag to force polling-only mode for troubleshooting? (Nice-to-have, not a blocker.)
- Should we ship the optional Node.js harness for golden traces as part of `scripts/` or keep it out-of-tree?
- How do we expose health metrics to users (e.g., `onemount --stats`)? Might defer until after transport lands.
