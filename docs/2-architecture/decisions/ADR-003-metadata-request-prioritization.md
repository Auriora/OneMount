# ADR-003: Metadata Request Prioritization

## Status

Accepted

## Context

In the original design, all metadata requests (directory listings, file lookups, etc.) were handled equally. This caused problems:

1. **User-Facing Blocking**: Interactive operations (ls, cd) blocked on background work
2. **Poor Responsiveness**: Users experienced delays during background sync
3. **Resource Contention**: Background and foreground work competed for resources
4. **Duplicate Requests**: Multiple requests for same directory caused redundant API calls
5. **Stale Cache Issues**: No mechanism to serve stale data while refreshing

The simple approach had limitations:
- No prioritization of user-facing operations
- No deduplication of in-flight requests
- No stale-cache serving with async refresh
- Background work could starve foreground operations

## Decision

We implemented a metadata request manager with prioritization and deduplication:

1. **Priority Queues**: Separate queues for foreground and background work
2. **Worker Pools**: Dedicated workers with at least one for foreground
3. **In-Flight Deduplication**: Share results for duplicate requests
4. **Stale-Cache Policy**: Serve stale data immediately, refresh async
5. **Preemption**: Foreground requests preempt background work

**Key Components**:

```go
// Metadata Request Manager
type MetadataRequestManager struct {
    foregroundQueue chan *metadataRequest
    backgroundQueue chan *metadataRequest
    inFlight        map[string]*inflightRequest
    workers         []*metadataWorker
    numWorkers      int
    minForeground   int
}

// Request Priority
type RequestPriority int

const (
    PriorityForeground RequestPriority = 1  // User-facing operations
    PriorityBackground RequestPriority = 2  // Background sync
)

// Request Types
type RequestType int

const (
    RequestTypeDirectory RequestType = 1  // Directory listing
    RequestTypeLookup    RequestType = 2  // File/folder lookup
    RequestTypeMetadata  RequestType = 3  // Metadata refresh
)
```

**Architecture**:

```
┌─────────────────────────────────────────────────────────────┐
│           Metadata Request Prioritization Flow              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User Operation (ls, cd, open)                             │
│       │                                                     │
│       ▼                                                     │
│  ┌──────────────────┐                                       │
│  │ Foreground Queue │ (Priority 1)                         │
│  └────────┬─────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌──────────────────────────────────────┐                  │
│  │   Metadata Request Manager           │                  │
│  │   - In-flight deduplication          │                  │
│  │   - Stale-cache policy               │                  │
│  │   - Worker pool management           │                  │
│  └────────┬─────────────────────────────┘                  │
│           │                                                 │
│           ▼                                                 │
│  ┌──────────────────┐                                       │
│  │ Worker Pool      │                                       │
│  │ - Foreground (1+)│                                       │
│  │ - Background (N) │                                       │
│  └────────┬─────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  Microsoft Graph API                                        │
│                                                             │
│  Background Sync                                            │
│       │                                                     │
│       ▼                                                     │
│  ┌──────────────────┐                                       │
│  │ Background Queue │ (Priority 2)                         │
│  └──────────────────┘                                       │
└─────────────────────────────────────────────────────────────┘
```

## Consequences

### Positive

1. **Better Responsiveness**: User operations never block on background work
2. **Efficient Resource Use**: Deduplication reduces redundant API calls
3. **Improved UX**: Stale-cache policy provides instant responses
4. **Predictable Performance**: Foreground operations have bounded latency
5. **Lower API Usage**: Fewer duplicate requests
6. **Graceful Degradation**: System remains responsive under load

### Negative

1. **Increased Complexity**: More components to understand and maintain
2. **Memory Overhead**: In-flight tracking uses memory
3. **Stale Data Risk**: Users may see slightly stale data briefly
4. **Testing Challenges**: Harder to test priority behavior
5. **Debugging Difficulty**: More complex request flow

### Neutral

1. **Worker Pool Size**: Configurable based on workload
2. **Cache TTL**: Configurable stale-cache threshold
3. **Queue Sizes**: Configurable queue depths

## Alternatives Considered

### 1. Single Queue (FIFO)

**Pros**:
- Simple implementation
- Easy to understand
- Predictable ordering

**Cons**:
- No prioritization
- User operations block on background work
- Poor responsiveness
- No deduplication

**Rejected**: Insufficient for good user experience

### 2. Separate Threads for Foreground/Background

**Pros**:
- Clear separation
- No priority inversion
- Simple to implement

**Cons**:
- No deduplication
- Resource inefficiency
- No stale-cache policy
- Fixed resource allocation

**Rejected**: Less flexible than worker pool approach

### 3. Async/Await with Priorities

**Pros**:
- Modern approach
- Good for I/O-bound work
- Built-in cancellation

**Cons**:
- Go doesn't have async/await
- Would require significant refactoring
- Unclear benefits over goroutines

**Rejected**: Not idiomatic Go

## Implementation Notes

### Priority Handling

1. **Foreground Priority**: Always processed first
2. **Background Priority**: Processed when no foreground work
3. **Preemption**: Workers check for foreground work between operations
4. **Starvation Prevention**: Background work eventually processed

### In-Flight Deduplication

1. **Request Key**: Generate key from request type and parameters
2. **Check In-Flight**: Check if request already in flight
3. **Share Result**: Multiple callers wait for same result
4. **Cleanup**: Remove from in-flight map when complete

### Stale-Cache Policy

1. **Check Cache**: Check if directory already cached
2. **Check Freshness**: Check if cache is fresh (< TTL)
3. **Serve Stale**: If stale, serve immediately and trigger refresh
4. **Async Refresh**: Refresh in background, update cache when complete

### Worker Pool Management

1. **Minimum Foreground**: At least one worker for foreground
2. **Dynamic Allocation**: Workers switch between queues
3. **Graceful Shutdown**: Wait for in-flight requests to complete
4. **Error Handling**: Retry failed requests with backoff

## Performance Considerations

- **Latency**: Foreground operations have bounded latency
- **Throughput**: Background operations processed efficiently
- **Memory**: In-flight tracking uses O(N) memory
- **API Usage**: Deduplication reduces API calls significantly

## Configuration

```go
type MetadataRequestConfig struct {
    NumWorkers      int           // Total number of workers
    MinForeground   int           // Minimum workers for foreground
    ForegroundQueue int           // Foreground queue size
    BackgroundQueue int           // Background queue size
    StaleCacheTTL   time.Duration // Stale cache threshold
}
```

## Monitoring

Track metrics:
- Foreground request latency
- Background request latency
- Queue depths
- Deduplication rate
- Cache hit rate
- Worker utilization

## References

- [Requirement 2D: FUSE Operation Performance](../../1-requirements/software-requirements-specification.md#requirement-2d)
- [Metadata Request Manager Implementation](../../../internal/fs/metadata_request_manager.go)
- [Filesystem Integration](../../../internal/fs/filesystem_types.go)

## Related ADRs

- ADR-001: Structured Metadata Store
- ADR-002: Socket.IO Realtime Subscriptions
- ADR-006: Status Caching

## Revision History

| Date | Version | Author | Changes |
|------|---------|--------|---------|
| 2025-01-21 | 1.0 | AI Agent | Initial version |
