package socketio

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/socketio/protocol"
)

// RealtimeTransport exposes the minimal interface needed by the filesystem to
// consume Engine.IO/Socket.IO notifications from Microsoft Graph.
// This interface abstracts the Socket.IO transport layer, allowing the filesystem
// to receive realtime change notifications without depending on specific transport implementations.
type RealtimeTransport interface {
	// Connect establishes a Socket.IO connection to the specified endpoint.
	// The endpoint should be a WebSocket URL obtained from Microsoft Graph's subscription API.
	// Headers typically include authentication tokens and other required Graph API headers.
	Connect(ctx context.Context, endpoint string, headers http.Header) error

	// Close gracefully terminates the Socket.IO connection.
	// This should clean up all resources and stop background goroutines.
	Close(ctx context.Context) error

	// On registers an event handler for the specified event type.
	// Handlers receive strongly-typed payloads for transport lifecycle and notification events.
	On(event EventType, handler Listener)

	// Health returns the current health state of the transport connection.
	// This includes connection status, error information, and heartbeat metrics
	// used by the delta loop to adjust polling frequency.
	Health() HealthState
}

// EventType enumerates transport lifecycle and diagnostic events that can be
// observed from the Socket.IO realtime transport. These events allow consumers
// to monitor connection health, receive notifications, and implement custom
// logging or diagnostic behavior.
type EventType string

const (
	// EventConnected is emitted when the Socket.IO connection is successfully established.
	EventConnected EventType = "transport:connected"

	// EventReconnected is emitted when the connection is re-established after a failure.
	EventReconnected EventType = "transport:reconnected"

	// EventDisconnected is emitted when the Socket.IO connection is closed or lost.
	EventDisconnected EventType = "transport:disconnected"

	// EventNotification is emitted when a change notification is received from Microsoft Graph.
	EventNotification EventType = "transport:notification"

	// EventError is emitted when transport-level errors occur (connection failures, protocol errors, etc.).
	EventError EventType = "transport:error"

	// EventPacketTrace is emitted for low-level packet debugging (when trace logging is enabled).
	EventPacketTrace EventType = "transport:packet_trace"

	// EventMessageTrace is emitted for Socket.IO message debugging (when trace logging is enabled).
	EventMessageTrace EventType = "transport:message_trace"

	// EventHealthChanged is emitted when the transport health state changes.
	EventHealthChanged EventType = "transport:health_changed"
)

// StatusCode represents the coarse health of the realtime channel, used by the
// delta loop to determine appropriate polling frequencies and fallback behavior.
type StatusCode string

const (
	// StatusUnknown indicates the transport health is not yet determined.
	// This is the initial state before any connection attempts.
	StatusUnknown StatusCode = "unknown"

	// StatusHealthy indicates the Socket.IO connection is working normally.
	// Delta polling occurs infrequently (every 30+ minutes) in this state.
	StatusHealthy StatusCode = "healthy"

	// StatusDegraded indicates connection issues but the transport is still functional.
	// Delta polling frequency increases to provide more reliable change detection.
	StatusDegraded StatusCode = "degraded"

	// StatusFailed indicates the Socket.IO connection is not working.
	// Delta polling falls back to the configured fallback interval (typically 30 minutes).
	StatusFailed StatusCode = "failed"
)

// HealthState captures the observable channel health metrics required by the
// delta loop to adjust polling cadence and provide diagnostic information.
// The filesystem uses these metrics to determine when to fall back to polling
// and when to resume relying on realtime notifications.
type HealthState struct {
	// Status indicates the overall health of the realtime connection.
	// Used by the delta loop to determine polling frequency.
	Status StatusCode

	// LastError contains the most recent transport error, if any.
	// Useful for diagnostics and troubleshooting connection issues.
	LastError error

	// LastHeartbeat is the timestamp of the most recent successful heartbeat.
	// Used to detect connection staleness and trigger reconnection attempts.
	LastHeartbeat time.Time

	// MissedHeartbeats counts consecutive heartbeat failures.
	// High values indicate connection degradation.
	MissedHeartbeats int

	// ConsecutiveFailures counts consecutive connection or operation failures.
	// Used to implement exponential backoff in reconnection logic.
	ConsecutiveFailures int

	// ReconnectCount tracks the total number of reconnection attempts.
	// Useful for monitoring connection stability over time.
	ReconnectCount int
}

// Listener receives strongly-typed payloads for the subscribed event.
type Listener func(payload interface{})

// ConnectedEvent describes the initial Engine.IO handshake for a session.
type ConnectedEvent struct {
	Endpoint  string
	Handshake *protocol.Handshake
	Attempt   int
}

// ReconnectedEvent captures reconnect attempts that eventually succeed.
type ReconnectedEvent struct {
	Endpoint  string
	Handshake *protocol.Handshake
	Attempt   int
	Backoff   time.Duration
}

// DisconnectedEvent provides context for channel closure.
type DisconnectedEvent struct {
	Code   int
	Reason string
	Err    error
}

// NotificationEvent contains the payloads from 42["notification", ...] frames.
type NotificationEvent struct {
	Payloads []interface{}
}

// ErrorEvent wraps transport-level errors.
type ErrorEvent struct {
	Err error
}

// HealthChangeEvent reports health transitions so consumers can react promptly.
type HealthChangeEvent struct {
	Previous HealthState
	Current  HealthState
}

// PacketTraceEvent and MessageTraceEvent reuse EnginePacketTrace and
// EngineMessageTrace and exist primarily for type clarity.
type PacketTraceEvent struct {
	Trace *EnginePacketTrace
}

type MessageTraceEvent struct {
	Trace *EngineMessageTrace
}

// listenerRegistry manages event handlers for transport implementations.
type listenerRegistry struct {
	mu        sync.RWMutex
	listeners map[EventType][]Listener
}

func newListenerRegistry() *listenerRegistry {
	return &listenerRegistry{listeners: make(map[EventType][]Listener)}
}

func (r *listenerRegistry) On(event EventType, handler Listener) {
	if handler == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners[event] = append(r.listeners[event], handler)
}

func (r *listenerRegistry) emit(event EventType, payload interface{}) {
	r.mu.RLock()
	handlers := append([]Listener(nil), r.listeners[event]...)
	r.mu.RUnlock()
	for _, handler := range handlers {
		safeInvoke(handler, payload)
	}
}

func safeInvoke(handler Listener, payload interface{}) {
	defer func() {
		_ = recover()
	}()
	handler(payload)
}
