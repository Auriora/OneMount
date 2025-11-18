package socketio

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// FakeTransport is a lightweight RealtimeTransport implementation for tests.
type FakeTransport struct {
	listeners *listenerRegistry

	mu        sync.Mutex
	connected bool
	endpoint  string
	health    HealthState
}

// NewFakeTransport returns a deterministic transport suitable for unit tests.
func NewFakeTransport() *FakeTransport {
	return &FakeTransport{
		listeners: newListenerRegistry(),
		health:    HealthState{Status: StatusUnknown},
	}
}

func (f *FakeTransport) Connect(ctx context.Context, endpoint string, headers http.Header) error {
	_ = headers
	f.mu.Lock()
	f.connected = true
	f.endpoint = endpoint
	f.health.Status = StatusHealthy
	f.health.LastError = nil
	f.health.LastHeartbeat = time.Now()
	f.mu.Unlock()

	f.listeners.emit(EventConnected, &ConnectedEvent{Endpoint: endpoint, Attempt: 1})
	f.emitHealthChange(StatusHealthy)
	return nil
}

func (f *FakeTransport) Close(ctx context.Context) error {
	_ = ctx
	f.mu.Lock()
	already := f.connected
	f.connected = false
	f.health.Status = StatusUnknown
	f.mu.Unlock()

	if already {
		f.listeners.emit(EventDisconnected, &DisconnectedEvent{})
	}
	f.emitHealthChange(StatusUnknown)
	return nil
}

func (f *FakeTransport) On(event EventType, handler Listener) {
	f.listeners.On(event, handler)
}

func (f *FakeTransport) Health() HealthState {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.health
}

// EmitNotification injects a Socket.IO notification with arbitrary payloads.
func (f *FakeTransport) EmitNotification(payloads ...interface{}) {
	f.listeners.emit(EventNotification, &NotificationEvent{Payloads: clonePayloads(payloads)})
}

// EmitError surfaces an error to listeners and marks the transport as failed.
func (f *FakeTransport) EmitError(err error) {
	if err == nil {
		return
	}
	f.mu.Lock()
	prev := f.health
	f.health.Status = StatusFailed
	f.health.LastError = err
	f.mu.Unlock()

	f.listeners.emit(EventError, &ErrorEvent{Err: err})
	f.listeners.emit(EventHealthChanged, &HealthChangeEvent{Previous: prev, Current: f.Health()})
}

// SetHealth overrides the underlying health snapshot (used in tests).
func (f *FakeTransport) SetHealth(state HealthState) {
	f.mu.Lock()
	prev := f.health
	f.health = state
	f.mu.Unlock()
	f.listeners.emit(EventHealthChanged, &HealthChangeEvent{Previous: prev, Current: state})
}

func (f *FakeTransport) emitHealthChange(status StatusCode) {
	f.mu.Lock()
	prev := f.health
	f.health.Status = status
	current := f.health
	f.mu.Unlock()
	f.listeners.emit(EventHealthChanged, &HealthChangeEvent{Previous: prev, Current: current})
}
