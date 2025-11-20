package metadata

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrInvalidTransition indicates an unsupported state change was requested.
var ErrInvalidTransition = errors.New("metadata: invalid state transition")

// StateManager coordinates validated metadata state transitions.
type StateManager struct {
	store   Store
	clock   Clock
	allowed map[ItemState]map[ItemState]struct{}
}

// StateManagerOption customizes manager construction.
type StateManagerOption func(*StateManager)

// WithStateManagerClock overrides the default clock.
func WithStateManagerClock(clock Clock) StateManagerOption {
	return func(m *StateManager) {
		if clock != nil {
			m.clock = clock
		}
	}
}

// NewStateManager returns a manager using the provided store.
func NewStateManager(store Store, opts ...StateManagerOption) (*StateManager, error) {
	if store == nil {
		return nil, fmt.Errorf("metadata: store is required")
	}
	manager := &StateManager{
		store: store,
		clock: systemClock{},
		allowed: map[ItemState]map[ItemState]struct{}{
			ItemStateGhost: stateSet(
				ItemStateHydrating,
				ItemStateHydrated,
				ItemStateDeleted,
				ItemStateDirtyLocal,
			),
			ItemStateHydrating: stateSet(
				ItemStateHydrated,
				ItemStateError,
			),
			ItemStateHydrated: stateSet(
				ItemStateDirtyLocal,
				ItemStateGhost,
				ItemStateDeleted,
			),
			ItemStateDirtyLocal: stateSet(
				ItemStateHydrated,
				ItemStateConflict,
				ItemStateError,
				ItemStateDeleted,
			),
			ItemStateConflict: stateSet(
				ItemStateHydrated,
				ItemStateDirtyLocal,
			),
			ItemStateError: stateSet(
				ItemStateHydrating,
				ItemStateDeleted,
			),
			ItemStateDeleted: {},
		},
	}
	for _, opt := range opts {
		opt(manager)
	}
	return manager, nil
}

// TransitionOption configures metadata transition behavior.
type TransitionOption func(*transitionConfig)

type transitionConfig struct {
	workerID        string
	err             error
	errTemporary    bool
	force           bool
	hydrationEvent  bool
	uploadEvent     bool
	newETag         string
	newSize         *uint64
	pinState        *PinState
	clearPending    bool
	customTimestamp *time.Time
}

// WithWorker assigns a worker ID to hydration/upload bookkeeping.
func WithWorker(id string) TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.workerID = id
	}
}

// WithHydrationEvent annotates the transition as part of hydration pipelines.
func WithHydrationEvent() TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.hydrationEvent = true
	}
}

// WithUploadEvent annotates the transition as part of upload workflows.
func WithUploadEvent() TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.uploadEvent = true
	}
}

// WithTransitionError records an error for ERROR transitions.
func WithTransitionError(err error, temporary bool) TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.err = err
		cfg.errTemporary = temporary
	}
}

// WithETag updates the entry's ETag when the transition succeeds.
func WithETag(etag string) TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.newETag = etag
	}
}

// WithSize updates the entry's size on transition.
func WithSize(size uint64) TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.newSize = &size
	}
}

// WithPinState replaces the entry's pin metadata.
func WithPinState(pin PinState) TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.pinState = &pin
	}
}

// ForceTransition bypasses the default transition table (use sparingly).
func ForceTransition() TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.force = true
	}
}

// ClearPendingRemote clears pending-remote markers post-transition.
func ClearPendingRemote() TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.clearPending = true
	}
}

// WithTransitionTimestamp overrides the default clock timestamp.
func WithTransitionTimestamp(ts time.Time) TransitionOption {
	return func(cfg *transitionConfig) {
		cfg.customTimestamp = &ts
	}
}

// Transition validates and applies a state change.
func (m *StateManager) Transition(ctx context.Context, id string, to ItemState, opts ...TransitionOption) (*Entry, error) {
	if to == "" {
		return nil, fmt.Errorf("metadata: transition requires a target state")
	}
	cfg := transitionConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	return m.store.Update(ctx, id, func(entry *Entry) error {
		if entry == nil {
			return ErrNotFound
		}
		if err := m.validateTransition(entry, to, cfg.force); err != nil {
			return err
		}
		m.applyTransition(entry, to, cfg)
		return nil
	})
}

func (m *StateManager) validateTransition(entry *Entry, to ItemState, force bool) error {
	// Virtual entries remain HYDRATED by definition.
	if entry.Virtual {
		if to != ItemStateHydrated {
			return fmt.Errorf("%w: virtual entries must remain HYDRATED (requested %s)", ErrInvalidTransition, to)
		}
		return nil
	}

	if force {
		return nil
	}

	current := entry.State
	targets, ok := m.allowed[current]
	if !ok {
		return fmt.Errorf("%w: no transitions defined for %s", ErrInvalidTransition, current)
	}
	if _, allowed := targets[to]; !allowed {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current, to)
	}
	return nil
}

func (m *StateManager) applyTransition(entry *Entry, to ItemState, cfg transitionConfig) {
	now := m.clock.Now()
	if cfg.customTimestamp != nil {
		now = cfg.customTimestamp.UTC()
	}

	switch to {
	case ItemStateHydrating:
		entry.State = ItemStateHydrating
		entry.Hydration.WorkerID = cfg.workerID
		entry.Hydration.StartedAt = &now
		entry.Hydration.CompletedAt = nil
		entry.Hydration.Error = nil
		entry.LastError = nil
	case ItemStateHydrated:
		entry.State = ItemStateHydrated
		entry.LastHydrated = &now
		entry.LastError = nil
		if cfg.hydrationEvent {
			entry.Hydration.CompletedAt = &now
			entry.Hydration.WorkerID = cfg.workerID
			entry.Hydration.Error = nil
		}
		if cfg.uploadEvent {
			entry.LastUploaded = &now
			entry.Upload.CompletedAt = &now
			entry.Upload.LastError = nil
		}
		if cfg.newETag != "" {
			entry.ETag = cfg.newETag
		}
		if cfg.newSize != nil {
			entry.Size = *cfg.newSize
		}
	case ItemStateDirtyLocal:
		entry.State = ItemStateDirtyLocal
		if cfg.uploadEvent {
			entry.Upload.SessionID = cfg.workerID
			entry.Upload.StartedAt = &now
			entry.Upload.CompletedAt = nil
			entry.Upload.LastError = nil
		}
	case ItemStateGhost:
		entry.State = ItemStateGhost
	case ItemStateDeleted:
		entry.State = ItemStateDeleted
	case ItemStateConflict:
		entry.State = ItemStateConflict
	case ItemStateError:
		entry.State = ItemStateError
		errMsg := ""
		if cfg.err != nil {
			errMsg = cfg.err.Error()
		}
		entry.LastError = &OperationError{
			Message:    errMsg,
			Temporary:  cfg.errTemporary,
			OccurredAt: now,
		}
		if cfg.hydrationEvent {
			entry.Hydration.Error = entry.LastError
			entry.Hydration.CompletedAt = &now
		}
		if cfg.uploadEvent {
			entry.Upload.LastError = entry.LastError
			entry.Upload.CompletedAt = &now
		}
	}

	if cfg.pinState != nil {
		entry.Pin = *cfg.pinState
	}
	if cfg.clearPending {
		entry.PendingRemote = false
	}
}

func stateSet(states ...ItemState) map[ItemState]struct{} {
	set := make(map[ItemState]struct{}, len(states))
	for _, st := range states {
		set[st] = struct{}{}
	}
	return set
}
