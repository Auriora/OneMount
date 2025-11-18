package socketio

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/socketio/protocol"
)

// RealtimeTransport exposes the minimal surface needed by the filesystem to
// consume Engine.IO/Socket.IO notifications.
type RealtimeTransport interface {
	Connect(ctx context.Context, endpoint string, headers http.Header) error
	Close(ctx context.Context) error
	On(event EventType, handler Listener)
	Health() HealthState
}

// EventType enumerates transport lifecycle and diagnostic events.
type EventType string

const (
	EventConnected     EventType = "transport:connected"
	EventReconnected   EventType = "transport:reconnected"
	EventDisconnected  EventType = "transport:disconnected"
	EventNotification  EventType = "transport:notification"
	EventError         EventType = "transport:error"
	EventPacketTrace   EventType = "transport:packet_trace"
	EventMessageTrace  EventType = "transport:message_trace"
	EventHealthChanged EventType = "transport:health_changed"
)

// StatusCode represents the coarse health of the realtime channel.
type StatusCode string

const (
	StatusUnknown  StatusCode = "unknown"
	StatusHealthy  StatusCode = "healthy"
	StatusDegraded StatusCode = "degraded"
	StatusFailed   StatusCode = "failed"
)

// HealthState captures the observable channel health metrics required by the
// delta loop to adjust polling cadence.
type HealthState struct {
	Status              StatusCode
	LastError           error
	LastHeartbeat       time.Time
	MissedHeartbeats    int
	ConsecutiveFailures int
	ReconnectCount      int
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
