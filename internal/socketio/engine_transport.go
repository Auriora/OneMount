package socketio

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/socketio/protocol"
)

var errTransportClosed = errors.New("engine transport closed")
var errHeartbeatTimeout = errors.New("engine transport heartbeat timeout")

// EngineTransportOptions customises EngineTransport behaviour for production and tests.
type EngineTransportOptions struct {
	Dialer                   protocol.Transport
	PacketTraceLimit         int
	InitialBackoff           time.Duration
	MaxBackoff               time.Duration
	BackoffJitter            float64
	MissedHeartbeatThreshold int
	RandSource               rand.Source
}

// EngineTransport implements RealtimeTransport using Engine.IO v4 over gorilla/websocket.
type EngineTransport struct {
	opts      EngineTransportOptions
	listeners *listenerRegistry

	stateMu sync.Mutex
	state   transportState

	endpoint string
	headers  http.Header

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	ready  chan error

	conn   protocol.Conn
	connMu sync.Mutex

	pongCh chan struct{}

	jitter   *rand.Rand
	lastBack time.Duration

	healthMu sync.RWMutex
	health   HealthState
}

type transportState int

const (
	stateIdle transportState = iota
	stateRunning
)

// NewEngineTransport constructs a transport with sensible defaults.
func NewEngineTransport(opts EngineTransportOptions) *EngineTransport {
	defaults := EngineTransportOptions{
		Dialer:                   protocol.NewWebSocketTransport(),
		PacketTraceLimit:         defaultPacketTracePayloadLimit,
		InitialBackoff:           time.Second,
		MaxBackoff:               60 * time.Second,
		BackoffJitter:            0.10,
		MissedHeartbeatThreshold: 2,
		RandSource:               rand.NewSource(time.Now().UnixNano()),
	}
	if opts.Dialer == nil {
		opts.Dialer = defaults.Dialer
	}
	if opts.PacketTraceLimit <= 0 {
		opts.PacketTraceLimit = defaults.PacketTraceLimit
	}
	if opts.InitialBackoff <= 0 {
		opts.InitialBackoff = defaults.InitialBackoff
	}
	if opts.MaxBackoff <= 0 {
		opts.MaxBackoff = defaults.MaxBackoff
	}
	if opts.BackoffJitter <= 0 {
		opts.BackoffJitter = defaults.BackoffJitter
	}
	if opts.MissedHeartbeatThreshold <= 0 {
		opts.MissedHeartbeatThreshold = defaults.MissedHeartbeatThreshold
	}
	if opts.RandSource == nil {
		opts.RandSource = defaults.RandSource
	}

	return &EngineTransport{
		opts:      opts,
		listeners: newListenerRegistry(),
		jitter:    rand.New(opts.RandSource),
		health:    HealthState{Status: StatusUnknown},
	}
}

// On attaches a listener for the specified event.
func (t *EngineTransport) On(event EventType, handler Listener) {
	t.listeners.On(event, handler)
}

// Connect establishes the Engine.IO session and waits for the initial handshake.
func (t *EngineTransport) Connect(ctx context.Context, endpoint string, headers http.Header) error {
	t.stateMu.Lock()
	defer t.stateMu.Unlock()

	if t.state == stateRunning {
		return nil
	}

	t.endpoint = endpoint
	t.headers = cloneHeaders(headers)
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.done = make(chan struct{})
	t.ready = make(chan error, 1)
	t.pongCh = make(chan struct{}, 1)
	t.state = stateRunning

	go t.run()

	select {
	case err := <-t.ready:
		if err != nil {
			t.cancel()
			<-t.done
			t.state = stateIdle
			return err
		}
		return nil
	case <-ctx.Done():
		t.cancel()
		<-t.done
		t.state = stateIdle
		return ctx.Err()
	}
}

// Close terminates the underlying Engine.IO session and stops reconnection attempts.
func (t *EngineTransport) Close(ctx context.Context) error {
	t.stateMu.Lock()
	if t.state != stateRunning {
		t.stateMu.Unlock()
		return nil
	}
	cancel := t.cancel
	done := t.done
	t.stateMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	t.stateMu.Lock()
	t.state = stateIdle
	t.stateMu.Unlock()
	return nil
}

// Health returns the current health snapshot.
func (t *EngineTransport) Health() HealthState {
	t.healthMu.RLock()
	defer t.healthMu.RUnlock()
	return t.health
}

func (t *EngineTransport) run() {
	defer close(t.done)
	defer close(t.ready)

	ctx := t.ctx
	connected := false
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			if !connected {
				t.signalReady(ctx.Err())
			}
			t.emitDisconnected(nil)
			return
		default:
		}

		attempt++
		conn, handshake, err := t.openConnection()
		if err != nil {
			t.registerFailure(err)
			if !connected {
				t.signalReady(err)
				return
			}
			if !t.waitForNextAttempt(ctx) {
				t.emitDisconnected(err)
				return
			}
			continue
		}

		t.setConn(conn)
		t.drainPong()
		isReconnect := connected
		t.markHealthy(isReconnect)

		currentAttempt := attempt
		if !connected {
			connected = true
			t.signalReady(nil)
			t.listeners.emit(EventConnected, &ConnectedEvent{Endpoint: t.endpoint, Handshake: handshake, Attempt: currentAttempt})
		} else {
			t.listeners.emit(EventReconnected, &ReconnectedEvent{Endpoint: t.endpoint, Handshake: handshake, Attempt: currentAttempt, Backoff: t.lastBack})
		}
		attempt = 0

		err = t.serve(ctx, conn, handshake)
		t.clearConn()

		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			t.emitDisconnected(nil)
			return
		}

		t.registerFailure(err)
		if !t.waitForNextAttempt(ctx) {
			t.emitDisconnected(err)
			return
		}
	}
}

func (t *EngineTransport) openConnection() (protocol.Conn, *protocol.Handshake, error) {
	wsURL, err := toEngineIOURL(t.endpoint)
	if err != nil {
		return nil, nil, err
	}
	conn, err := t.opts.Dialer.Dial(wsURL, t.headers)
	if err != nil {
		return nil, nil, err
	}
	pkt, err := conn.Read()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	if pkt.Type != protocol.PacketTypeOpen {
		conn.Close()
		return nil, nil, fmt.Errorf("engine transport expected handshake, got %s", pkt.Type)
	}
	handshake, err := pkt.DecodeHandshake()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	if trace := newEnginePacketTrace("inbound", t.opts.PacketTraceLimit, pkt); trace != nil {
		t.listeners.emit(EventPacketTrace, &PacketTraceEvent{Trace: trace})
	}
	return conn, handshake, nil
}

func (t *EngineTransport) serve(ctx context.Context, conn protocol.Conn, handshake *protocol.Handshake) error {
	errCh := make(chan error, 2)
	done := make(chan struct{})

	go func() {
		errCh <- t.readLoop(ctx, conn, done)
	}()
	go func() {
		errCh <- t.heartbeatLoop(ctx, conn, handshake, done)
	}()

	err := <-errCh
	close(done)
	other := <-errCh
	if err == nil {
		err = other
	}
	return err
}

func (t *EngineTransport) readLoop(ctx context.Context, conn protocol.Conn, done <-chan struct{}) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return errTransportClosed
		default:
		}

		pkt, err := conn.Read()
		if err != nil {
			return err
		}
		if trace := newEnginePacketTrace("inbound", t.opts.PacketTraceLimit, pkt); trace != nil {
			t.listeners.emit(EventPacketTrace, &PacketTraceEvent{Trace: trace})
		}

		switch pkt.Type {
		case protocol.PacketTypePing:
			t.markHeartbeat()
			if err := t.writePacket(conn, protocol.NewPongPacket()); err != nil {
				return err
			}
		case protocol.PacketTypePong:
			if ch := t.pongCh; ch != nil {
				select {
				case ch <- struct{}{}:
				default:
				}
			}
			t.markHeartbeat()
		case protocol.PacketTypeMessage:
			msg, err := pkt.DecodeMessage()
			if err != nil {
				t.listeners.emit(EventError, &ErrorEvent{Err: err})
				continue
			}
			if trace := newEngineMessageTrace(t.opts.PacketTraceLimit, msg); trace != nil {
				t.listeners.emit(EventMessageTrace, &MessageTraceEvent{Trace: trace})
			}
			t.handleMessage(msg)
		case protocol.PacketTypeClose:
			return errTransportClosed
		}
	}
}

func (t *EngineTransport) handleMessage(msg *protocol.Message) {
	if msg == nil {
		return
	}
	switch msg.Event {
	case "notification":
		t.listeners.emit(EventNotification, &NotificationEvent{Payloads: clonePayloads(msg.Payloads)})
	case "error":
		t.listeners.emit(EventError, &ErrorEvent{Err: fmt.Errorf("socket.io error: %v", msg.Payloads)})
	default:
		// Unhandled events are only logged via trace
	}
}

func (t *EngineTransport) heartbeatLoop(ctx context.Context, conn protocol.Conn, handshake *protocol.Handshake, done <-chan struct{}) error {
	interval := time.Duration(handshake.PingInterval) * time.Millisecond
	if interval <= 0 {
		interval = 25 * time.Second
	}
	timeout := time.Duration(handshake.PingTimeout) * time.Millisecond
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return errTransportClosed
		case <-ticker.C:
			t.drainPong()
			if err := t.writePacket(conn, protocol.NewPingPacket()); err != nil {
				return err
			}
			if err := t.awaitPong(ctx, timeout, done); err != nil {
				t.noteMissedHeartbeat(err)
				if errors.Is(err, errHeartbeatTimeout) {
					return err
				}
			}
		}
	}
}

func (t *EngineTransport) awaitPong(ctx context.Context, timeout time.Duration, done <-chan struct{}) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ch := t.pongCh
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return errTransportClosed
	case <-timer.C:
		return errHeartbeatTimeout
	case <-ch:
		return nil
	}
}

func (t *EngineTransport) writePacket(conn protocol.Conn, packet *protocol.Packet) error {
	if packet == nil {
		return nil
	}
	t.connMu.Lock()
	defer t.connMu.Unlock()
	if conn == nil {
		return errTransportClosed
	}
	if err := conn.Write(packet); err != nil {
		return err
	}
	if trace := newEnginePacketTrace("outbound", t.opts.PacketTraceLimit, packet); trace != nil {
		t.listeners.emit(EventPacketTrace, &PacketTraceEvent{Trace: trace})
	}
	return nil
}

func (t *EngineTransport) signalReady(err error) {
	select {
	case t.ready <- err:
	default:
	}
}

func (t *EngineTransport) waitForNextAttempt(ctx context.Context) bool {
	delay := t.nextBackoffDelay()
	t.lastBack = delay
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (t *EngineTransport) nextBackoffDelay() time.Duration {
	t.healthMu.RLock()
	failures := t.health.ConsecutiveFailures
	t.healthMu.RUnlock()
	if failures <= 0 {
		failures = 1
	}
	shift := failures - 1
	if shift > 10 {
		shift = 10
	}
	base := t.opts.InitialBackoff * time.Duration(1<<shift)
	if base > t.opts.MaxBackoff {
		base = t.opts.MaxBackoff
	}
	jitter := (t.jitter.Float64()*2 - 1) * t.opts.BackoffJitter
	next := time.Duration(float64(base) * (1 + jitter))
	if next < t.opts.InitialBackoff {
		next = t.opts.InitialBackoff
	}
	return next
}

func (t *EngineTransport) registerFailure(err error) {
	t.healthMu.Lock()
	prev := t.health
	t.health.Status = StatusFailed
	t.health.LastError = err
	t.health.MissedHeartbeats = 0
	t.health.ConsecutiveFailures++
	current := t.health
	t.healthMu.Unlock()
	t.emitHealthChange(prev, current)
	if err != nil {
		t.listeners.emit(EventError, &ErrorEvent{Err: err})
	}
}

func (t *EngineTransport) markHealthy(isReconnect bool) {
	t.healthMu.Lock()
	prev := t.health
	t.health.Status = StatusHealthy
	t.health.LastError = nil
	t.health.MissedHeartbeats = 0
	t.health.ConsecutiveFailures = 0
	t.health.LastHeartbeat = time.Now()
	if isReconnect {
		t.health.ReconnectCount++
	} else {
		t.health.ReconnectCount = 0
	}
	current := t.health
	t.healthMu.Unlock()
	t.emitHealthChange(prev, current)
}

func (t *EngineTransport) markHeartbeat() {
	t.healthMu.Lock()
	prev := t.health
	t.health.LastHeartbeat = time.Now()
	t.health.MissedHeartbeats = 0
	if t.health.Status != StatusHealthy {
		t.health.Status = StatusHealthy
	}
	current := t.health
	t.healthMu.Unlock()
	t.emitHealthChange(prev, current)
}

func (t *EngineTransport) noteMissedHeartbeat(err error) {
	t.healthMu.Lock()
	prev := t.health
	t.health.LastError = err
	t.health.MissedHeartbeats++
	if t.health.MissedHeartbeats >= t.opts.MissedHeartbeatThreshold {
		t.health.Status = StatusDegraded
	}
	current := t.health
	t.healthMu.Unlock()
	t.emitHealthChange(prev, current)
}

func (t *EngineTransport) emitDisconnected(err error) {
	t.listeners.emit(EventDisconnected, &DisconnectedEvent{Err: err})
}

func (t *EngineTransport) emitHealthChange(prev, current HealthState) {
	if prev.Status == current.Status && prev.MissedHeartbeats == current.MissedHeartbeats && prev.ConsecutiveFailures == current.ConsecutiveFailures && errors.Is(prev.LastError, current.LastError) {
		return
	}
	t.listeners.emit(EventHealthChanged, &HealthChangeEvent{Previous: prev, Current: current})
}

func (t *EngineTransport) setConn(conn protocol.Conn) {
	t.connMu.Lock()
	t.conn = conn
	t.connMu.Unlock()
}

func (t *EngineTransport) clearConn() {
	t.connMu.Lock()
	if t.conn != nil {
		_ = t.conn.Close()
	}
	t.conn = nil
	t.connMu.Unlock()
}

func (t *EngineTransport) drainPong() {
	if t.pongCh == nil {
		return
	}
	for {
		select {
		case <-t.pongCh:
		default:
			return
		}
	}
}

func cloneHeaders(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	cloned := make(http.Header, len(h))
	for k, vv := range h {
		cp := make([]string, len(vv))
		copy(cp, vv)
		cloned[k] = cp
	}
	return cloned
}

func toEngineIOURL(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("empty endpoint")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", "4")
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	return u.String(), nil
}
