package fs

import "context"

// subscriptionManager defines the minimal functionality required by the delta loop.
type subscriptionManager interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Notifications() <-chan struct{}
    IsActive() bool
}
