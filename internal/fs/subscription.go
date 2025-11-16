package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
)

// SubscriptionManager handles Microsoft Graph webhook subscriptions and notification delivery.
type SubscriptionManager struct {
	opts WebhookOptions
	auth *graph.Auth

	notifications chan struct{}
	stopCh        chan struct{}
	server        *http.Server
	wg            sync.WaitGroup

	mu             sync.RWMutex
	subscriptionID string
	expiration     time.Time
	listenAddr     string

	ctx context.Context
}

var subscriptionRenewalCheckInterval = time.Hour

// NewSubscriptionManager creates a manager for the provided options.
func NewSubscriptionManager(opts WebhookOptions, auth *graph.Auth) *SubscriptionManager {
	return &SubscriptionManager{
		opts:          opts,
		auth:          auth,
		notifications: make(chan struct{}, 1),
		stopCh:        make(chan struct{}),
	}
}

// Start launches the HTTP server (if required) and creates the webhook subscription.
func (m *SubscriptionManager) Start(ctx context.Context) error {
	if !m.opts.Enabled {
		return errors.New("webhook manager started while disabled")
	}
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if m.auth == nil {
		return errors.New("auth cannot be nil for webhook subscriptions")
	}
	if err := m.startServer(); err != nil {
		return err
	}

	m.ctx = ctx

	if err := m.createSubscription(ctx); err != nil {
		m.shutdownServer()
		return err
	}

	m.wg.Add(1)
	go m.renewalLoop()
	return nil
}

// Notifications returns the channel that fires when a webhook notification arrives.
func (m *SubscriptionManager) Notifications() <-chan struct{} {
	return m.notifications
}

// Stop shuts down the HTTP server and deletes the subscription.
func (m *SubscriptionManager) Stop(ctx context.Context) error {
	close(m.stopCh)
	m.shutdownServer()

	m.deleteSubscription(ctx)

	m.wg.Wait()
	close(m.notifications)
	return nil
}

func (m *SubscriptionManager) startServer() error {
	path := m.opts.Path
	if path == "" {
		path = "/onemount/webhook"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, m.handleWebhook)

	addr := m.opts.ListenAddress
	if addr == "" {
		addr = "127.0.0.1:8787"
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen webhook addr: %w", err)
	}
	m.mu.Lock()
	m.listenAddr = listener.Addr().String()
	m.mu.Unlock()

	m.server = &http.Server{
		Handler: mux,
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		var serveErr error
		if m.opts.TLSCertFile != "" && m.opts.TLSKeyFile != "" {
			serveErr = m.server.ServeTLS(listener, m.opts.TLSCertFile, m.opts.TLSKeyFile)
		} else {
			serveErr = m.server.Serve(listener)
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logging.Error().Err(serveErr).Msg("Webhook server terminated unexpectedly")
		}
	}()

	logging.Info().
		Str("listenAddr", m.listenAddr).
		Str("path", path).
		Msg("Webhook server started")
	return nil
}

func (m *SubscriptionManager) shutdownServer() {
	if m.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.Warn().Err(err).Msg("Failed to shut down webhook server gracefully")
	}
}

func (m *SubscriptionManager) handleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		m.handleValidation(w, r)
	case http.MethodPost:
		m.handleNotification(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *SubscriptionManager) handleValidation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("validationToken")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(token)); err != nil {
		logging.Warn().Err(err).Msg("Failed to write validation token response")
	}
}

type webhookNotificationPayload struct {
	Value []struct {
		SubscriptionID string `json:"subscriptionId"`
		ClientState    string `json:"clientState"`
		ChangeType     string `json:"changeType"`
		Resource       string `json:"resource"`
		Expiration     string `json:"subscriptionExpirationDateTime"`
	} `json:"value"`
}

func (m *SubscriptionManager) handleNotification(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var payload webhookNotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logging.Warn().Err(err).Msg("Failed to decode webhook notification payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	valid := false
	for _, entry := range payload.Value {
		if m.opts.ClientState != "" && entry.ClientState != "" && entry.ClientState != m.opts.ClientState {
			logging.Warn().
				Str("subscriptionID", entry.SubscriptionID).
				Msg("Webhook notification rejected due to clientState mismatch")
			continue
		}
		if entry.Expiration != "" {
			if exp, err := time.Parse(time.RFC3339, entry.Expiration); err == nil {
				m.mu.Lock()
				m.expiration = exp
				m.mu.Unlock()
			}
		}
		valid = true
	}

	w.WriteHeader(http.StatusAccepted)
	if valid {
		m.trigger()
	}
}

func (m *SubscriptionManager) trigger() {
	select {
	case m.notifications <- struct{}{}:
	default:
	}
}

func (m *SubscriptionManager) createSubscription(ctx context.Context) error {
	publicURL := strings.TrimSuffix(m.opts.PublicURL, "/")
	if publicURL == "" {
		return errors.New("webhook public URL must be set")
	}
	path := m.opts.Path
	if path == "" {
		path = "/onemount/webhook"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	expiration := graph.BuildExpiration(48 * time.Hour)

	req := graph.SubscriptionRequest{
		ChangeType:         m.opts.ChangeType,
		NotificationURL:    publicURL + path,
		Resource:           m.opts.Resource,
		ExpirationDateTime: expiration,
		ClientState:        m.opts.ClientState,
	}

	logging.Info().
		Str("resource", req.Resource).
		Str("notificationUrl", req.NotificationURL).
		Msg("Creating webhook subscription")

	sub, err := graph.CreateSubscription(ctx, m.auth, req)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	m.mu.Lock()
	m.subscriptionID = sub.ID
	m.expiration = sub.ExpirationDateTime
	m.mu.Unlock()

	graph.LogSubscription("Webhook subscription established", sub)
	return nil
}

func (m *SubscriptionManager) attemptRecreation(reason string) {
	ctx := m.ctx
	if ctx == nil {
		logging.Debug().
			Str("reason", reason).
			Msg("Skipping webhook subscription recreation; context is nil")
		return
	}
	if err := ctx.Err(); err != nil {
		logging.Debug().
			Str("reason", reason).
			Err(err).
			Msg("Skipping webhook subscription recreation; context cancelled")
		return
	}
	if err := m.createSubscription(ctx); err != nil {
		logging.Error().
			Str("reason", reason).
			Err(err).
			Msg("Failed to recreate webhook subscription; continuing with polling only")
		return
	}
	m.mu.RLock()
	id := m.subscriptionID
	m.mu.RUnlock()
	logging.Info().
		Str("reason", reason).
		Str("subscriptionID", id).
		Msg("Webhook subscription recreated")
}

func (m *SubscriptionManager) deleteSubscription(ctx context.Context) {
	m.mu.Lock()
	id := m.subscriptionID
	m.subscriptionID = ""
	m.mu.Unlock()
	if id == "" {
		return
	}
	if err := graph.DeleteSubscription(ctx, m.auth, id); err != nil {
		logging.Warn().Err(err).Str("subscriptionID", id).Msg("Failed to delete webhook subscription")
	} else {
		logging.Info().Str("subscriptionID", id).Msg("Webhook subscription deleted")
	}
}

func (m *SubscriptionManager) renewalLoop() {
	defer m.wg.Done()
	ticker := time.NewTicker(subscriptionRenewalCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.IsActive() {
				m.attemptRecreation("inactive")
				continue
			}
			m.mu.RLock()
			expiration := m.expiration
			m.mu.RUnlock()
			if time.Until(expiration) <= 24*time.Hour {
				if err := m.renewSubscription(); err != nil {
					logging.Error().Err(err).Msg("Webhook subscription renewal failed; falling back to polling")
					m.mu.Lock()
					m.subscriptionID = ""
					m.mu.Unlock()
					m.attemptRecreation("renewal-failed")
				}
			}
		case <-m.stopCh:
			return
		}
	}
}

func (m *SubscriptionManager) renewSubscription() error {
	m.mu.RLock()
	id := m.subscriptionID
	m.mu.RUnlock()
	if id == "" {
		return errors.New("no active subscription to renew")
	}

	expiration := graph.BuildExpiration(48 * time.Hour)
	sub, err := graph.RenewSubscription(m.ctx, m.auth, id, expiration)
	if err != nil {
		return fmt.Errorf("renew subscription: %w", err)
	}

	m.mu.Lock()
	m.expiration = sub.ExpirationDateTime
	m.mu.Unlock()

	logging.Info().
		Str("subscriptionID", sub.ID).
		Time("expiration", sub.ExpirationDateTime).
		Msg("Webhook subscription renewed")
	return nil
}

// IsActive returns true when a subscription is currently active.
func (m *SubscriptionManager) IsActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.subscriptionID != ""
}

// ListenAddress returns the address the webhook server is bound to. Primarily used for tests.
func (m *SubscriptionManager) ListenAddress() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listenAddr
}
