package fs

import (
	"context"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/socketio"
)

type socketNotifier interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Notifications() <-chan struct{}
	IsActive() bool
	HealthSnapshot() socketio.HealthState
}

// ChangeNotifier provides a facade over the realtime subscription manager, allowing the
// filesystem to receive change notifications without directly depending on Socket.IO implementation details.
// It abstracts the complexity of managing Socket.IO connections, health monitoring, and fallback logic,
// providing a simple interface for the delta loop to consume realtime notifications.
type ChangeNotifier struct {
	opts    RealtimeOptions
	auth    *graph.Auth
	manager socketNotifier
	factory func(RealtimeOptions, *graph.Auth) socketNotifier
}

// NewChangeNotifier creates a new change notifier with the specified realtime options and authentication.
// The notifier manages Socket.IO subscriptions to Microsoft Graph and provides change notifications
// to the filesystem's delta loop. When realtime is disabled or polling-only mode is enabled,
// the notifier operates in a passive mode that doesn't establish Socket.IO connections.
func NewChangeNotifier(opts RealtimeOptions, auth *graph.Auth) *ChangeNotifier {
	return newChangeNotifierWithFactory(opts, auth, func(o RealtimeOptions, a *graph.Auth) socketNotifier {
		return NewSocketSubscriptionManager(o, a, nil)
	})
}

func newChangeNotifierWithFactory(opts RealtimeOptions, auth *graph.Auth, factory func(RealtimeOptions, *graph.Auth) socketNotifier) *ChangeNotifier {
	return &ChangeNotifier{opts: opts, auth: auth, factory: factory}
}

func (n *ChangeNotifier) ensureManager() socketNotifier {
	if n.manager == nil && n.factory != nil {
		n.manager = n.factory(n.opts, n.auth)
	}
	return n.manager
}

func (n *ChangeNotifier) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !n.opts.Enabled {
		logging.Info().Msg("Realtime notifier disabled; delta loop will rely on polling")
		return nil
	}
	if n.opts.PollingOnly {
		logging.Info().Msg("Realtime notifier running in polling-only mode")
		return nil
	}
	mgr := n.ensureManager()
	if mgr == nil {
		return nil
	}
	if err := mgr.Start(ctx); err != nil {
		return err
	}
	return nil
}

func (n *ChangeNotifier) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if n.manager == nil {
		return nil
	}
	err := n.manager.Stop(ctx)
	n.manager = nil
	return err
}

func (n *ChangeNotifier) Notifications() <-chan struct{} {
	if n.manager == nil {
		return nil
	}
	return n.manager.Notifications()
}

func (n *ChangeNotifier) IsActive() bool {
	if !n.opts.Enabled || n.opts.PollingOnly {
		return false
	}
	if n.manager == nil {
		return false
	}
	return n.manager.IsActive()
}

func (n *ChangeNotifier) RealtimeMode() string {
	if !n.opts.Enabled {
		return "disabled"
	}
	if n.opts.PollingOnly {
		return "polling-only"
	}
	if n.IsActive() {
		return "socketio"
	}
	return "socketio-inactive"
}

func (n *ChangeNotifier) HealthSnapshot() socketio.HealthState {
	if n.manager == nil {
		return socketio.HealthState{Status: socketio.StatusUnknown}
	}
	return n.manager.HealthSnapshot()
}
