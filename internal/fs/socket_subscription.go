package fs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/socketio"
)

type subscriptionLookup func(context.Context, *graph.Auth, string) (*graph.SocketSubscription, error)

// SocketSubscriptionManager streams change notifications via Microsoft Graph's Socket.IO endpoint.
type SocketSubscriptionManager struct {
	opts WebhookOptions
	auth *graph.Auth

	transport socketio.RealtimeTransport
	lookup    subscriptionLookup

	notifications chan struct{}
	closeOnce     sync.Once

	mu             sync.RWMutex
	subscriptionID string

	health atomic.Value // stores socketio.HealthState
}

// NewSocketSubscriptionManager creates a Socket.IO-based notification manager.
func NewSocketSubscriptionManager(opts WebhookOptions, auth *graph.Auth, transport socketio.RealtimeTransport) *SocketSubscriptionManager {
	if transport == nil {
		transport = socketio.NewEngineTransport(socketio.EngineTransportOptions{})
	}
	mgr := &SocketSubscriptionManager{
		opts:          opts,
		auth:          auth,
		transport:     transport,
		lookup:        graph.GetSocketSubscription,
		notifications: make(chan struct{}, 1),
	}
	mgr.health.Store(socketio.HealthState{Status: socketio.StatusUnknown})
	mgr.bindTransportEvents()
	return mgr
}

// Start establishes the Socket.IO channel and begins listening for notifications.
func (m *SocketSubscriptionManager) Start(ctx context.Context) error {
	if !m.opts.Enabled {
		return errors.New("socket subscription manager started while disabled")
	}
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if m.auth == nil {
		return errors.New("auth cannot be nil for socket subscriptions")
	}

	resource := m.opts.Resource
	if strings.TrimSpace(resource) == "" {
		resource = "/me/drive/root"
	}

	sub, err := m.lookup(ctx, m.auth, resource)
	if err != nil {
		return fmt.Errorf("get socket subscription: %w", err)
	}

	m.mu.Lock()
	m.subscriptionID = sub.ID
	m.mu.Unlock()

	logging.Info().
		Str("subscriptionID", sub.ID).
		Str("resource", resource).
		Str("notificationURL", sub.NotificationURL).
		Msg("Socket.IO subscription endpoint acquired")

	headers, err := m.buildHeaders(ctx)
	if err != nil {
		return err
	}

	if err := m.transport.Connect(ctx, sub.NotificationURL, headers); err != nil {
		m.mu.Lock()
		m.subscriptionID = ""
		m.mu.Unlock()
		return fmt.Errorf("connect realtime transport: %w", err)
	}

	m.health.Store(m.transport.Health())
	return nil
}

// Notifications returns the notification channel.
func (m *SocketSubscriptionManager) Notifications() <-chan struct{} {
	return m.notifications
}

// Stop terminates the Socket.IO connection.
func (m *SocketSubscriptionManager) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var err error
	if m.transport != nil {
		err = m.transport.Close(ctx)
	}

	m.mu.Lock()
	m.subscriptionID = ""
	m.mu.Unlock()

	m.health.Store(socketio.HealthState{Status: socketio.StatusUnknown})

	m.closeOnce.Do(func() {
		close(m.notifications)
	})

	return err
}

// IsActive returns true if a Socket.IO endpoint is connected and healthy.
func (m *SocketSubscriptionManager) IsActive() bool {
	if state, ok := m.health.Load().(socketio.HealthState); ok {
		return state.Status == socketio.StatusHealthy
	}
	return false
}

func (m *SocketSubscriptionManager) bindTransportEvents() {
	if m.transport == nil {
		return
	}

	m.transport.On(socketio.EventConnected, func(payload interface{}) {
		evt, _ := payload.(*socketio.ConnectedEvent)
		entry := logging.Info().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Str("endpoint", safeEndpoint(evt)).
			Int("attempt", attemptOf(evt))
		if evt != nil && evt.Handshake != nil {
			entry = entry.
				Dur("pingInterval", time.Duration(evt.Handshake.PingInterval)*time.Millisecond).
				Dur("pingTimeout", time.Duration(evt.Handshake.PingTimeout)*time.Millisecond)
		}
		entry.Msg("Socket.IO channel connected")
	})

	m.transport.On(socketio.EventReconnected, func(payload interface{}) {
		evt, _ := payload.(*socketio.ReconnectedEvent)
		logging.Info().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Str("endpoint", safeEndpoint(evt)).
			Dur("backoff", backoffOf(evt)).
			Int("attempt", attemptOf(evt)).
			Msg("Socket.IO channel reconnected")
	})

	m.transport.On(socketio.EventDisconnected, func(payload interface{}) {
		evt, _ := payload.(*socketio.DisconnectedEvent)
		entry := logging.Warn().
			Str("subscriptionID", m.subscriptionIDSafe())
		if evt != nil {
			entry = entry.Err(evt.Err)
		}
		entry.Msg("Socket.IO channel disconnected")
	})

	m.transport.On(socketio.EventError, func(payload interface{}) {
		evt, _ := payload.(*socketio.ErrorEvent)
		entry := logging.Warn().
			Str("subscriptionID", m.subscriptionIDSafe())
		if evt != nil && evt.Err != nil {
			entry = entry.Err(evt.Err)
		}
		entry.Msg("Socket.IO channel error")
	})

	m.transport.On(socketio.EventNotification, func(payload interface{}) {
		evt, _ := payload.(*socketio.NotificationEvent)
		logging.Debug().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Interface("payload", evt).
			Msg("Socket.IO notification received")
		m.trigger()
	})

	m.transport.On(socketio.EventPacketTrace, func(payload interface{}) {
		evt, _ := payload.(*socketio.PacketTraceEvent)
		if evt == nil || evt.Trace == nil || !logging.IsTraceEnabled() {
			return
		}
		logging.Trace().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Str("direction", evt.Trace.Direction).
			Str("packetType", evt.Trace.Type).
			Str("payload", evt.Trace.Payload).
			Msg("Socket.IO engine packet")
	})

	m.transport.On(socketio.EventMessageTrace, func(payload interface{}) {
		evt, _ := payload.(*socketio.MessageTraceEvent)
		if evt == nil || evt.Trace == nil || !logging.IsTraceEnabled() {
			return
		}
		logging.Trace().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Str("messageType", evt.Trace.Type).
			Str("namespace", evt.Trace.Namespace).
			Str("event", evt.Trace.Event).
			Interface("payloads", evt.Trace.Payloads).
			Msg("Socket.IO engine message decoded")
	})

	m.transport.On(socketio.EventHealthChanged, func(payload interface{}) {
		evt, _ := payload.(*socketio.HealthChangeEvent)
		if evt == nil {
			return
		}
		m.health.Store(evt.Current)
		level := logging.Debug
		if evt.Current.Status == socketio.StatusFailed {
			level = logging.Warn
		} else if evt.Current.Status == socketio.StatusDegraded {
			level = logging.Info
		}
		level().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Str("status", string(evt.Current.Status)).
			Int("missedHeartbeats", evt.Current.MissedHeartbeats).
			Int("consecutiveFailures", evt.Current.ConsecutiveFailures).
			Err(evt.Current.LastError).
			Msg("Socket.IO transport health changed")
	})
}

func (m *SocketSubscriptionManager) buildHeaders(_ context.Context) (http.Header, error) {
	if m.auth == nil {
		return nil, errors.New("auth cannot be nil")
	}
	token := strings.TrimSpace(m.auth.AccessToken)
	if token == "" {
		return nil, errors.New("access token missing")
	}
	headers := make(http.Header)
	headers.Set("Authorization", "Bearer "+token)
	return headers, nil
}

func (m *SocketSubscriptionManager) trigger() {
	select {
	case m.notifications <- struct{}{}:
	default:
	}
}

func (m *SocketSubscriptionManager) subscriptionIDSafe() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.subscriptionID
}

// HealthSnapshot returns the current transport health.
func (m *SocketSubscriptionManager) HealthSnapshot() socketio.HealthState {
	if state, ok := m.health.Load().(socketio.HealthState); ok {
		return state
	}
	return socketio.HealthState{Status: socketio.StatusUnknown}
}

// RealtimeMode identifies the transport mechanism used by the manager.
func (m *SocketSubscriptionManager) RealtimeMode() string {
	return "socketio"
}

func safeEndpoint(evt interface{}) string {
	switch v := evt.(type) {
	case *socketio.ConnectedEvent:
		if v != nil {
			return v.Endpoint
		}
	case *socketio.ReconnectedEvent:
		if v != nil {
			return v.Endpoint
		}
	}
	return ""
}

func backoffOf(evt *socketio.ReconnectedEvent) time.Duration {
	if evt == nil {
		return 0
	}
	return evt.Backoff
}

func attemptOf(evt interface{}) int {
	switch v := evt.(type) {
	case *socketio.ConnectedEvent:
		if v != nil {
			return v.Attempt
		}
	case *socketio.ReconnectedEvent:
		if v != nil {
			return v.Attempt
		}
	}
	return 0
}
