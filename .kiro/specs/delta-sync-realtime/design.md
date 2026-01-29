# Design: Delta Sync Realtime

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 6. Delta Synchronization Component

**Location**: `internal/fs/delta.go`, `internal/fs/sync.go`

**Verification Steps**:
1. Review delta sync loop implementation
2. Test initial sync
3. Test incremental sync
4. Test conflict detection
5. Test delta link persistence

**Expected Interfaces**:
- `DeltaLoop()` goroutine
- `FetchDeltas()` method
- `ApplyChanges()` method
- Delta link storage in bbolt database

**Verification Criteria**:
- Initial sync fetches all metadata
- Incremental syncs fetch only changes
- Remote changes update local cache
- Conflicts create conflict copies
- Delta link persists across restarts

### 13. Realtime Subscription Component

**Location**: `internal/fs/socket_subscription.go`, `internal/graph/socketio/`

**Verification Steps**:
1. Review the Socket.IO subscription manager to ensure it requests delegated endpoints, maintains health snapshots, and surfaces events to the delta loop.
2. Test initialization per mount (including polling-only mode) to confirm configuration is honored.
3. Simulate notifications to ensure they immediately trigger delta queries and preempt lower-priority metadata work.
4. Inject transport failures to verify degraded health is reported and polling cadence falls back to the Requirement 5 baseline.
5. Test renewal before expiration and reconnection logic, ensuring tokens persist in BBolt.
6. Test shutdown/unmount flows to confirm the transport disconnects cleanly within the timeout.

**Expected Interfaces**:
- `type ChangeNotifier interface { Start(ctx); Notifications() <-chan struct{}; Health() socketio.HealthState; Stop(ctx); IsActive() bool }` implemented by `SocketSubscriptionManager`.
- Health/expiration metrics exposed via `internal/metrics`/stats so polling cadence decisions remain deterministic.

**Verification Criteria**:
- Socket.IO subscriptions start successfully on mount (or skip when polling-only) and surface diagnostics on failure.
- Notifications trigger immediate delta queries; health snapshots reflect heartbeat timings and failures.
- Subscription IDs and expirations persist and renew before the 24 h Graph deadline.
- When the transport degrades, the notifier marks itself unhealthy so the delta loop remains in fallback mode until the channel recovers.
- Unmounting a drive cleanly tears down the socket connection and deletes stored subscription metadata.
- Personal drives can scope subscriptions to root or subfolders; business drives restrict to the root per Graph limits.

### Realtime Socket Architecture

```
┌─────────────────────────────────────────────────────────────┐
│               Socket.IO Realtime Flow (per mount)           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  OneDrive API                                               │
│       │                                                     │
│       │ 1. POST /subscriptions/socketIo                     │
│       ▼                                                     │
│  ┌──────────────┐                                           │
│  │ Socket Sub   │                                           │
│  │   Manager    │                                           │
│  └──────┬───────┘                                           │
│         │ health + expiry                                   │
│         ▼                                                   │
│  ┌──────────────────────────────────────────────────┐       │
│  │    RealtimeNotifier (events + health)            │       │
│  └──────┬──────────────────────┬────────────────────┘       │
│         │ notifications        │ health snapshot            │
│         ▼                      ▼                            │
│   Delta Sync Trigger    Interval Controller                 │
│         │                      │                            │
│         ▼                      ▼                            │
│     Metadata DB        Polling Interval Logic               │
│                                                             │
│  Background: renewal loop + reconnect/backoff               │
└─────────────────────────────────────────────────────────────┘
```

- Each mount instantiates a Socket.IO subscription manager that owns the Engine.IO transport, renewal tokens, and reconnection/backoff logic.
- The manager publishes notifications to the delta loop and a health snapshot (status, last heartbeat, retry counts) so Requirement 5 can lengthen/shorten polling cadences deterministically.
- Polling-only mode skips the connection but still exposes “disabled” mode through the same interface so diagnostics stay consistent.
- When the transport degrades, the manager marks itself unhealthy, emits diagnostics, and leaves the delta loop in 5-minute polling mode until the Socket.IO channel recovers.

### Tree Synchronization Behavior

The initial tree walk (and any subsequent full resync) runs entirely in the background. While the walk is running:

- Every directory fetched is immediately marked “warm” by populating the parent inode’s `children` slice so later interactive commands do not re-fetch it.
- Progress counters are exposed so the UI can show status without blocking commands.
- User-facing commands operate on whatever portion of the cache is already populated; only directories that have never been fetched block while their metadata is retrieved.
- When the refresh interval elapses, the walker revalidates directories asynchronously while continuing to serve cached data.

### Delta Synchronization Properties

**Property 20: Initial Delta Sync**
*For any* first filesystem mount, the system should fetch the complete directory structure using the delta API
**Validates: Requirements 5.1**

**Property 21: Metadata Cache Updates**
*For any* remote changes detected via delta query, the system should update the local metadata cache
**Validates: Requirements 5.8**

**Property 22: Conflict Copy Creation**
*For any* file with both local and remote changes, the system should create a conflict copy
**Validates: Requirements 5.11**

**Property 23: Delta Token Persistence**
*For any* completed delta sync, the system should store the @odata.deltaLink token for the next sync cycle
**Validates: Requirements 5.12**
