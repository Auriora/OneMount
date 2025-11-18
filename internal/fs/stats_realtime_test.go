package fs

import (
	"context"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/socketio"
)

type stubRealtimeManager struct {
	mode          string
	state         socketio.HealthState
	notifications chan struct{}
}

func (s *stubRealtimeManager) RealtimeMode() string                 { return s.mode }
func (s *stubRealtimeManager) HealthSnapshot() socketio.HealthState { return s.state }
func (s *stubRealtimeManager) Start(context.Context) error          { return nil }
func (s *stubRealtimeManager) Stop(context.Context) error           { return nil }
func (s *stubRealtimeManager) Notifications() <-chan struct{} {
	if s.notifications == nil {
		s.notifications = make(chan struct{})
	}
	return s.notifications
}
func (s *stubRealtimeManager) IsActive() bool { return true }

func TestFilesystemAugmentRealtimeStatsFromManager(t *testing.T) {
	f := &Filesystem{
		webhookOptions: &WebhookOptions{Enabled: true, UseSocketIO: true},
	}
	state := socketio.HealthState{
		Status:              socketio.StatusHealthy,
		MissedHeartbeats:    1,
		ConsecutiveFailures: 0,
		LastHeartbeat:       time.Unix(100, 0),
		ReconnectCount:      3,
	}
	f.subscriptionManager = &stubRealtimeManager{mode: "socketio", state: state}

	stats := &Stats{}
	f.augmentRealtimeStats(stats)

	if stats.RealtimeMode != "socketio" {
		t.Fatalf("expected mode socketio, got %s", stats.RealtimeMode)
	}
	if stats.RealtimeStatus != socketio.StatusHealthy {
		t.Fatalf("expected status healthy, got %s", stats.RealtimeStatus)
	}
	if stats.RealtimeReconnectCount != 3 {
		t.Fatalf("expected reconnect count 3, got %d", stats.RealtimeReconnectCount)
	}
	if stats.RealtimeMissedHeartbeats != 1 {
		t.Fatalf("expected missed heartbeats 1, got %d", stats.RealtimeMissedHeartbeats)
	}
	if stats.RealtimeLastHeartbeat.IsZero() {
		t.Fatalf("expected last heartbeat to be set")
	}
}

func TestFilesystemAugmentRealtimeStatsPollingOnly(t *testing.T) {
	f := &Filesystem{
		webhookOptions: &WebhookOptions{Enabled: true, PollingOnly: true},
	}
	stats := &Stats{}
	f.augmentRealtimeStats(stats)
	if stats.RealtimeMode != "polling-only" {
		t.Fatalf("expected polling-only mode, got %s", stats.RealtimeMode)
	}
}
