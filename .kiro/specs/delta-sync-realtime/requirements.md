# Requirements Document: Delta Sync Realtime

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/requirements.md`. Sections below are verbatim.

## Requirements

### Requirement 5: Delta Synchronization Verification

**User Story:** As a user, I want local changes from OneDrive to be reflected automatically so that I always see the latest version of files.

#### Acceptance Criteria

1. WHEN the filesystem is first mounted, THE OneMount System SHALL fetch the complete directory structure from OneDrive using the delta API
2. WHEN the filesystem is mounted and realtime sync is enabled, THE OneMount System SHALL attempt to establish a Microsoft Graph Socket.IO change-notification subscription for the mounted drive so that real-time events can wake the delta loop
3. WHEN creating the Socket.IO subscription for personal OneDrive, THE OneMount System SHALL target the root folder or selected subfolders consistent with Microsoft Graph’s supported resources; WHEN targeting OneDrive for Business, THE OneMount System SHALL limit the scope to the drive root as required by Graph
4. WHEN the Socket.IO subscription is healthy, THE OneMount System SHALL run delta polling no more frequently than every 30 minutes (configurable but never lower than 5 minutes) and SHALL log any deviation from that cadence
5. WHEN a Socket.IO notification payload is received, THE OneMount System SHALL immediately trigger a delta query to fetch changes and SHALL preempt lower-priority metadata work so user-facing operations do not stall
6. WHEN the Socket.IO subscription is unavailable or unhealthy, THE OneMount System SHALL automatically fall back to delta polling every 5 minutes by default and SHALL log the degraded state
7. IF the subscription continues to fail or error, THEN THE OneMount System MAY temporarily shorten the polling interval down to 10 seconds to recover, but MUST return to the configured fallback cadence within one interval after the subscription is restored and SHALL log the entire degraded period
8. WHEN remote changes are detected via delta query, THE OneMount System SHALL update the local metadata cache
9. WHEN a remotely modified file is accessed, THE OneMount System SHALL download the new version
10. WHEN a cached file has been modified remotely, THE OneMount System SHALL invalidate the local cache entry using ETag comparison
11. IF a file has both local and remote changes, THEN THE OneMount System SHALL create a conflict copy
12. WHEN delta sync completes, THE OneMount System SHALL store the @odata.deltaLink token for the next sync cycle
13. WHEN the Socket.IO subscription approaches expiration (per Graph limits), THE OneMount System SHALL renew it proactively and log the attempt
14. IF subscription renewal or reconnection fails, THEN THE OneMount System SHALL continue using the shorter polling interval until the subscription is restored and SHALL raise diagnostics for the operator

### Requirement 17: Realtime Subscription Management

**User Story:** As a system, I want a resilient Microsoft Graph Socket.IO subscription layer so that realtime notifications stay healthy without requiring inbound webhooks.

#### Acceptance Criteria

1. WHEN mounting a drive and realtime sync is enabled, THE OneMount System SHALL instantiate a single Socket.IO subscription manager that exposes a unified stream of events to the delta loop.
2. THE subscription manager SHALL surface health, expiration, and last-success timestamps so that Requirement 5 can adjust polling cadences deterministically and display status via `onemount --stats`.
3. WHERE the user enables polling-only mode, THE OneMount System SHALL skip establishing the Socket.IO connection but SHALL continue to report that realtime mode is disabled.
4. WHEN shutting down or unmounting, THE subscription manager SHALL gracefully disconnect from the Socket.IO endpoint and release all resources.
5. THE realtime implementation SHALL remain fully standalone (no webhooks, proxies, or managed relays such as Azure Web PubSub) unless explicitly approved in configuration.

### Requirement 20: Engine.IO / Socket.IO Transport Implementation

**User Story:** As a OneMount developer, I want clear requirements for the optional Engine.IO/Socket.IO transport so that, when this transport is selected, it behaves predictably without relying on unmaintained third-party libraries or external services.

#### Acceptance Criteria

1. WHEN the Socket.IO transport is enabled, THE OneMount System SHALL implement the Microsoft Graph notification channel using Engine.IO v4 over WebSocket only, setting `EIO=4` and `transport=websocket` query parameters and joining the default namespace (`/`).
2. WHEN establishing or refreshing the connection, THE transport SHALL attach the current OAuth access token (Authorization bearer header) and any additional headers Graph requires, and SHALL refresh the connection whenever the token is rotated.
3. WHEN an Engine.IO handshake frame is received, THE transport SHALL parse the ping interval/timeout values, log them at debug level, and configure its heartbeat timers accordingly.
4. WHILE the connection is active, THE transport SHALL send ping/pong frames per the negotiated interval, detect two consecutive missed heartbeats as a failure, and immediately surface the unhealthy state to the ChangeNotifier so Requirement 5 can fall back to polling.
5. WHEN the connection closes or errors, THE transport SHALL attempt reconnection with exponential backoff (starting at 1 s, doubling each attempt, capped at 60 s, with ±10 % jitter) and SHALL reset the backoff after a successful reconnect.
6. THE implementation SHALL stream decoded Socket.IO events (e.g., `notification`, `error`) through strongly typed callbacks, and SHALL expose a health indicator that the ChangeNotifier and delta sync loop can query in constant time.
7. WHEN running with verbose logging enabled, THE transport SHALL emit structured trace logs for handshake data, ping/pong timing, packet read/write summaries (payload truncated to a configurable limit), and close/error codes sufficient for supportability.
8. THE transport SHALL include automated tests covering packet encode/decode, heartbeat scheduling, reconnection backoff, and error propagation so regressions can be caught without live Graph access.
9. THE transport SHALL remain self-contained within the OneMount codebase—no third-party Socket.IO client libraries, proxies, or managed relays (e.g., Azure Web PubSub) are permitted unless explicitly whitelisted via configuration for troubleshooting.
