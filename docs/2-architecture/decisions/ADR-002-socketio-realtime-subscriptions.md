# ADR-002: Socket.IO Realtime Subscriptions

## Status

Accepted

## Context

The original design used a generic `ChangeNotifier` interface without specifying the implementation. We needed a way to receive realtime notifications from Microsoft Graph API about file changes without:

1. **Inbound Connectivity**: No need for public IP or port forwarding
2. **Webhook Infrastructure**: No need for webhook endpoints or SSL certificates
3. **Polling Overhead**: Reduce unnecessary API calls
4. **Deployment Complexity**: Keep deployment simple for end users
5. **Reliability**: Handle connection failures gracefully

Traditional webhook-based approaches had limitations:
- Require inbound connectivity (firewall issues)
- Need public endpoints with SSL
- Complex deployment and configuration
- Difficult for home users and laptops
- No built-in reconnection handling

## Decision

We implemented Socket.IO-based realtime subscriptions using Microsoft Graph's notification channel:

1. **Socket.IO Client**: Custom Engine.IO v4 implementation
2. **Subscription Manager**: Manages subscription lifecycle
3. **Health Monitoring**: Tracks connection health and fallback
4. **Automatic Reconnection**: Exponential backoff with jitter
5. **Graceful Degradation**: Falls back to polling when unhealthy

**Key Components**:

```go
// Subscription Manager Interface
type subscriptionManager interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Notifications() <-chan struct{}
    Health() HealthState
    IsActive() bool
    Stats() NotifierStats
}

// Socket.IO Transport
type RealtimeTransport interface {
    Connect(url string, auth *graph.Auth) error
    Disconnect() error
    Send(message []byte) error
    Receive() <-chan []byte
    Health() HealthState
}

// Health States
const (
    HealthHealthy   HealthState = "healthy"
    HealthDegraded  HealthState = "degraded"
    HealthUnhealthy HealthState = "unhealthy"
)
```

**Architecture**:

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

## Consequences

### Positive

1. **No Inbound Connectivity**: Works behind firewalls and NAT
2. **Simple Deployment**: No webhook configuration needed
3. **Better User Experience**: Faster sync with realtime notifications
4. **Reliable**: Automatic reconnection with exponential backoff
5. **Graceful Degradation**: Falls back to polling when unhealthy
6. **Lower API Usage**: Reduces unnecessary delta queries
7. **Laptop Friendly**: Handles sleep/wake cycles gracefully

### Negative

1. **Implementation Complexity**: Custom Engine.IO implementation required
2. **Maintenance Burden**: Need to maintain Socket.IO client
3. **Testing Challenges**: Harder to test realtime behavior
4. **Debugging Difficulty**: Connection issues harder to diagnose
5. **Microsoft Dependency**: Relies on Microsoft Graph Socket.IO support

### Neutral

1. **Fallback to Polling**: System still works without Socket.IO
2. **Optional Feature**: Can be disabled if problematic
3. **Resource Usage**: Persistent connection uses some resources

## Alternatives Considered

### 1. Webhooks

**Pros**:
- Standard approach
- Well-documented
- Supported by Microsoft Graph

**Cons**:
- Requires inbound connectivity
- Needs public endpoint with SSL
- Complex deployment
- Difficult for home users
- Firewall issues

**Rejected**: Too complex for end-user deployment

### 2. Polling Only

**Pros**:
- Simple implementation
- No persistent connection
- Easy to test

**Cons**:
- Higher API usage
- Slower sync
- Wastes bandwidth
- Poor user experience

**Rejected**: Insufficient for good user experience

### 3. Azure Web PubSub

**Pros**:
- Managed service
- Reliable
- Scalable

**Cons**:
- External dependency
- Additional cost
- Requires Azure account
- Overkill for use case

**Rejected**: Too complex and costly

### 4. Third-Party Socket.IO Library

**Pros**:
- Less implementation work
- Community support
- Battle-tested

**Cons**:
- External dependency
- May not support Engine.IO v4
- Less control
- Potential maintenance issues

**Rejected**: Need full control for reliability

## Implementation Notes

### Engine.IO v4 Protocol

1. **Handshake**: Negotiate ping interval and timeout
2. **Heartbeat**: Send ping/pong frames per negotiated interval
3. **Messages**: Send/receive Socket.IO messages
4. **Close**: Graceful disconnection

### Reconnection Strategy

1. **Exponential Backoff**: Start at 1s, double each attempt, cap at 60s
2. **Jitter**: Add ±10% randomness to prevent thundering herd
3. **Reset**: Reset backoff after successful reconnect
4. **Health Monitoring**: Track consecutive failures

### Fallback to Polling

1. **Health Check**: Monitor connection health
2. **Degraded State**: Two consecutive missed heartbeats
3. **Polling Interval**: Shorten to 5 minutes when degraded
4. **Recovery**: Return to 30-minute polling when healthy

### Token Refresh

1. **Monitor Expiration**: Track OAuth token expiration
2. **Proactive Refresh**: Refresh before expiration
3. **Reconnect**: Reconnect with new token
4. **Error Handling**: Handle refresh failures gracefully

## Performance Considerations

- **Connection Overhead**: Persistent connection uses some resources
- **Heartbeat Traffic**: Regular ping/pong frames
- **Reconnection Cost**: Exponential backoff limits reconnection attempts
- **Polling Fallback**: Ensures system remains responsive

## Security Considerations

- **Token Security**: OAuth token sent with each connection
- **TLS**: All connections use TLS
- **Token Refresh**: Tokens refreshed before expiration
- **Connection Validation**: Validate connection before use

## References

- [Requirement 17: Realtime Subscription Management](../../1-requirements/software-requirements-specification.md#requirement-17)
- [Requirement 20: Engine.IO / Socket.IO Transport Implementation](../../1-requirements/software-requirements-specification.md#requirement-20)
- [Socket.IO Implementation](../../../internal/socketio/)
- [Subscription Manager Implementation](../../../internal/fs/filesystem_types.go)
- [Microsoft Graph Socket.IO Documentation](https://docs.microsoft.com/en-us/graph/api/subscription-post-subscriptions)

## Related ADRs

- ADR-001: Structured Metadata Store
- ADR-003: Metadata Request Prioritization

## Revision History

| Date | Version | Author | Changes |
|------|---------|--------|---------|
| 2025-01-21 | 1.0 | AI Agent | Initial version |
