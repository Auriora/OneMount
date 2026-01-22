package socketio

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestUT_SocketIO_ToEngineIOURL(t *testing.T) {
	t.Parallel()

	raw := "https://graph.microsoft.com/me/drive/root"
	converted, err := toEngineIOURL(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(converted, "wss://") {
		t.Fatalf("expected wss scheme, got %s", converted)
	}
	if !strings.Contains(converted, "/socket.io/") {
		t.Fatalf("expected socket.io path, got %s", converted)
	}
	if !strings.Contains(converted, "EIO=4") || !strings.Contains(converted, "transport=websocket") {
		t.Fatalf("expected engine query parameters, got %s", converted)
	}
}

func TestUT_SocketIO_ToEngineIOURLStripsCallback(t *testing.T) {
	t.Parallel()

	raw := "https://f3hb0mpua.svc.ms/zbaehwg/callback?sn=token"
	converted, err := toEngineIOURL(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(converted, "/callback") {
		t.Fatalf("expected callback segment to be removed, got %s", converted)
	}
	if !strings.Contains(converted, "/socket.io/") {
		t.Fatalf("expected socket.io path, got %s", converted)
	}
	if !strings.Contains(converted, "sn=token") {
		t.Fatalf("opaque query token should be preserved, got %s", converted)
	}
}

func TestUT_SocketIO_EngineTransportBackoffRespectsCap(t *testing.T) {
	t.Parallel()

	tr := NewEngineTransport(EngineTransportOptions{
		InitialBackoff:           time.Second,
		MaxBackoff:               4 * time.Second,
		RandSource:               rand.NewSource(1),
		MissedHeartbeatThreshold: 2,
	})
	tr.opts.BackoffJitter = 0

	type expectation struct {
		failures int
		expected time.Duration
	}

	cases := []expectation{
		{failures: 1, expected: time.Second},
		{failures: 2, expected: 2 * time.Second},
		{failures: 3, expected: 4 * time.Second},
		{failures: 6, expected: 4 * time.Second}, // capped at max
	}

	for _, tc := range cases {
		tr.healthMu.Lock()
		tr.health.ConsecutiveFailures = tc.failures
		tr.healthMu.Unlock()
		if got := tr.nextBackoffDelay(); got != tc.expected {
			t.Fatalf("failures=%d expected %s got %s", tc.failures, tc.expected, got)
		}
	}
}
