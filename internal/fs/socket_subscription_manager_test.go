package fs

import (
	"context"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/socketio"
)

func TestSocketSubscriptionManagerTriggersNotifications(t *testing.T) {
	t.Parallel()

	fake := socketio.NewFakeTransport()
	mgr := NewSocketSubscriptionManager(RealtimeOptions{Enabled: true, Resource: "/me/drive/root"}, &graph.Auth{AccessToken: "test-token"}, fake)
	mgr.lookup = func(context.Context, *graph.Auth, string) (*graph.SocketSubscription, error) {
		return &graph.SocketSubscription{ID: "sub-123", NotificationURL: "https://graph.test/notifications"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("start manager: %v", err)
	}
	defer mgr.Stop(context.Background())

	// Healthy transport should mark manager active.
	if !mgr.IsActive() {
		t.Fatalf("expected manager to be active after transport connect")
	}

	fake.EmitNotification(map[string]any{"id": "abc"})

	select {
	case <-mgr.Notifications():
		// success
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected notification trigger within deadline")
	}

	// Degraded health must flip IsActive to false so the delta loop can fall back to polling.
	fake.SetHealth(socketio.HealthState{Status: socketio.StatusDegraded, MissedHeartbeats: 2})
	if mgr.IsActive() {
		t.Fatalf("expected manager to report inactive when transport is degraded")
	}
	health := mgr.HealthSnapshot()
	if health.Status != socketio.StatusDegraded {
		t.Fatalf("expected degraded health snapshot, got %s", health.Status)
	}
}
