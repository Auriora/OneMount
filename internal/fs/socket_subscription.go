package fs

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/socketio"
)

// SocketSubscriptionManager streams change notifications via Microsoft Graph's Socket.IO endpoint.
type SocketSubscriptionManager struct {
	opts WebhookOptions
	auth *graph.Auth

	notifications chan struct{}
	stopCh        chan struct{}
	wg            sync.WaitGroup

	mu             sync.RWMutex
	subscriptionID string
	client         *socketio.Client
}

// NewSocketSubscriptionManager creates a Socket.IO-based notification manager.
func NewSocketSubscriptionManager(opts WebhookOptions, auth *graph.Auth) *SocketSubscriptionManager {
	return &SocketSubscriptionManager{
		opts:          opts,
		auth:          auth,
		notifications: make(chan struct{}, 1),
		stopCh:        make(chan struct{}),
	}
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
	if resource == "" {
		resource = "/me/drive/root"
	}

	sub, err := graph.GetSocketSubscription(ctx, m.auth, resource)
	if err != nil {
		return fmt.Errorf("get socket subscription: %w", err)
	}

	m.mu.Lock()
	m.subscriptionID = sub.ID
	m.mu.Unlock()

	logging.Info().
		Str("subscriptionID", sub.ID).
		Str("resource", resource).
		Msg("Socket.IO subscription endpoint acquired")

	if err := m.connect(sub.NotificationURL); err != nil {
		return err
	}

	return nil
}

func (m *SocketSubscriptionManager) connect(endpoint string) error {
	client, err := socketio.Socket(endpoint)
	if err != nil {
		return fmt.Errorf("initialize socket.io client: %w", err)
	}

	client.On(socketio.EventConnect, func(args ...interface{}) {
		logging.Info().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Msg("Socket.IO channel connected")
	})
	client.On(socketio.EventReconnect, func(args ...interface{}) {
		logging.Info().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Msg("Socket.IO channel reconnected")
	})
	client.On(socketio.EventError, func(args ...interface{}) {
		logging.Warn().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Interface("error", args).
			Msg("Socket.IO channel error")
	})
	client.On("notification", func(args ...interface{}) {
		logging.Debug().
			Str("subscriptionID", m.subscriptionIDSafe()).
			Interface("payload", args).
			Msg("Socket.IO notification received")
		m.trigger()
	})

	m.mu.Lock()
	m.client = client
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		client.Connect(nil)
		<-m.stopCh
		client.Disconnect()
	}()

	return nil
}

// Notifications returns the notification channel.
func (m *SocketSubscriptionManager) Notifications() <-chan struct{} {
	return m.notifications
}

// Stop terminates the Socket.IO connection.
func (m *SocketSubscriptionManager) Stop(ctx context.Context) error {
	_ = ctx
	close(m.stopCh)

	m.mu.Lock()
	client := m.client
	m.client = nil
	m.subscriptionID = ""
	m.mu.Unlock()

	if client != nil {
		client.Disconnect()
	}

	m.wg.Wait()
	close(m.notifications)
	return nil
}

// IsActive returns true if a Socket.IO endpoint is connected.
func (m *SocketSubscriptionManager) IsActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.subscriptionID != ""
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
