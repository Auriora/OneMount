package fs

import (
	"context"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/socketio"
)

type stubSocketNotifier struct {
	ch     chan struct{}
	active bool
	health socketio.HealthState
}

func (s *stubSocketNotifier) Start(context.Context) error { return nil }
func (s *stubSocketNotifier) Stop(context.Context) error  { return nil }
func (s *stubSocketNotifier) Notifications() <-chan struct{} {
	if s.ch == nil {
		s.ch = make(chan struct{})
	}
	return s.ch
}
func (s *stubSocketNotifier) IsActive() bool { return s.active }
func (s *stubSocketNotifier) HealthSnapshot() socketio.HealthState {
	if s.health.Status == "" {
		s.health.Status = socketio.StatusUnknown
	}
	return s.health
}

func TestUT_FS_ChangeNotifier_Disabled(t *testing.T) {
	notifier := NewChangeNotifier(RealtimeOptions{Enabled: false}, nil)
	if err := notifier.Start(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if notifier.Notifications() != nil {
		t.Fatalf("expected nil notifications when disabled")
	}
	if mode := notifier.RealtimeMode(); mode != "disabled" {
		t.Fatalf("expected disabled mode, got %s", mode)
	}
}

func TestUT_FS_ChangeNotifier_PollingOnly(t *testing.T) {
	notifier := NewChangeNotifier(RealtimeOptions{Enabled: true, PollingOnly: true}, nil)
	if err := notifier.Start(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mode := notifier.RealtimeMode(); mode != "polling-only" {
		t.Fatalf("expected polling-only mode, got %s", mode)
	}
	if notifier.IsActive() {
		t.Fatalf("polling-only notifier should not be active")
	}
}

func TestUT_FS_ChangeNotifier_DelegatesToSocketManager(t *testing.T) {
	state := socketio.HealthState{Status: socketio.StatusHealthy, LastHeartbeat: time.Unix(123, 0)}
	stub := &stubSocketNotifier{active: true, health: state}
	notifier := newChangeNotifierWithFactory(RealtimeOptions{Enabled: true}, nil, func(RealtimeOptions, *graph.Auth) socketNotifier {
		return stub
	})
	if err := notifier.Start(context.Background()); err != nil {
		t.Fatalf("expected no error starting stub notifier, got %v", err)
	}
	if !notifier.IsActive() {
		t.Fatalf("expected notifier to report active state")
	}
	if notifier.RealtimeMode() != "socketio" {
		t.Fatalf("expected socketio mode, got %s", notifier.RealtimeMode())
	}
	health := notifier.HealthSnapshot()
	if health.Status != socketio.StatusHealthy {
		t.Fatalf("expected healthy status, got %s", health.Status)
	}
}
