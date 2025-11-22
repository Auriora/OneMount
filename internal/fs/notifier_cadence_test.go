package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/socketio"
)

func TestDeltaIntervalRespectsNotifierHealth(t *testing.T) {
	fs := &Filesystem{realtimeOptions: &RealtimeOptions{Enabled: true, FallbackInterval: defaultRealtimeFallbackInterval}}
	fs.notifierLastStatus.Store(socketio.StatusHealthy)

	// healthy → fallback realtime interval
	fs.subscriptionManager = &fakeDeltaNotifier{health: socketio.HealthState{Status: socketio.StatusHealthy}, active: true}
	interval, ok := fs.deltaIntervalFromNotifier()
	if !ok || interval != defaultRealtimeFallbackInterval {
		t.Fatalf("expected realtime fallback interval, got %v ok=%v", interval, ok)
	}

	// degraded → polling interval
	fs.subscriptionManager = &fakeDeltaNotifier{health: socketio.HealthState{Status: socketio.StatusDegraded}, active: true}
	interval, ok = fs.deltaIntervalFromNotifier()
	if !ok || interval != defaultPollingInterval {
		t.Fatalf("expected polling interval for degraded, got %v", interval)
	}

	// failed → 10s recovery window, set recoverySince
	fs.notifierRecoverySince.Store(0)
	fs.subscriptionManager = &fakeDeltaNotifier{health: socketio.HealthState{Status: socketio.StatusFailed}, active: true}
	interval, ok = fs.deltaIntervalFromNotifier()
	if !ok || interval != defaultRecoveryInterval {
		t.Fatalf("expected recovery interval for failed, got %v", interval)
	}
	if fs.notifierRecoverySince.Load() == 0 {
		t.Fatalf("expected recovery window to be opened")
	}

	// healthy again should clear recovery window
	fs.subscriptionManager = &fakeDeltaNotifier{health: socketio.HealthState{Status: socketio.StatusHealthy}, active: true}
	_, _ = fs.deltaIntervalFromNotifier()
	if since := fs.notifierRecoverySince.Load(); since != 0 {
		t.Fatalf("expected recovery window cleared, still %d", since)
	}
}

// alias to reuse fakeNotifier defined in delta_test.go
type fakeNotifier = fakeDeltaNotifier
