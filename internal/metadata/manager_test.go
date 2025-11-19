package metadata

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStateManagerHydrationLifecycle(t *testing.T) {
	store := newMemoryStore()
	entry := &Entry{
		ID:    "id-1",
		Name:  "example.bin",
		State: ItemStateGhost,
	}
	if err := store.Save(context.Background(), entry); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	manager, err := NewStateManager(store)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}

	start := time.Date(2025, time.November, 19, 10, 0, 0, 0, time.UTC)
	if _, err := manager.Transition(context.Background(), "id-1", ItemStateHydrating,
		WithHydrationEvent(),
		WithWorker("hydrator-1"),
		WithTransitionTimestamp(start),
	); err != nil {
		t.Fatalf("transition to hydrating: %v", err)
	}

	entryAfterStart, _ := store.Get(context.Background(), "id-1")
	if entryAfterStart.State != ItemStateHydrating {
		t.Fatalf("expected hydrating state, got %s", entryAfterStart.State)
	}
	if entryAfterStart.Hydration.WorkerID != "hydrator-1" {
		t.Fatalf("expected worker tracking")
	}
	if entryAfterStart.Hydration.StartedAt == nil || !entryAfterStart.Hydration.StartedAt.Equal(start) {
		t.Fatalf("expected hydration start timestamp")
	}

	finish := start.Add(2 * time.Minute)
	size := uint64(2048)
	if _, err := manager.Transition(context.Background(), "id-1", ItemStateHydrated,
		WithHydrationEvent(),
		WithWorker("hydrator-1"),
		WithETag("etag123"),
		WithSize(size),
		WithTransitionTimestamp(finish),
		ClearPendingRemote(),
	); err != nil {
		t.Fatalf("transition to hydrated: %v", err)
	}

	finalState, _ := store.Get(context.Background(), "id-1")
	if finalState.State != ItemStateHydrated {
		t.Fatalf("expected hydrated, got %s", finalState.State)
	}
	if finalState.ETag != "etag123" || finalState.Size != size {
		t.Fatalf("expected size & etag updates %+v", finalState)
	}
	if finalState.LastHydrated == nil || !finalState.LastHydrated.Equal(finish) {
		t.Fatalf("expected hydration timestamp recorded")
	}
	if finalState.Hydration.CompletedAt == nil || !finalState.Hydration.CompletedAt.Equal(finish) {
		t.Fatalf("expected hydration completed timestamp")
	}
	if finalState.PendingRemote {
		t.Fatalf("expected pending remote cleared")
	}
}

func TestStateManagerRejectsInvalidTransition(t *testing.T) {
	store := newMemoryStore()
	entry := &Entry{
		ID:    "id-2",
		Name:  "virtual.txt",
		State: ItemStateHydrated,
	}
	if err := store.Save(context.Background(), entry); err != nil {
		t.Fatalf("seed: %v", err)
	}
	manager, err := NewStateManager(store)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	if _, err := manager.Transition(context.Background(), "id-2", ItemStateHydrating); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
}

func TestStateManagerErrorTransition(t *testing.T) {
	store := newMemoryStore()
	entry := &Entry{
		ID:    "id-3",
		Name:  "file.txt",
		State: ItemStateHydrating,
	}
	if err := store.Save(context.Background(), entry); err != nil {
		t.Fatalf("seed: %v", err)
	}
	manager, err := NewStateManager(store)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	if _, err := manager.Transition(context.Background(), "id-3", ItemStateError,
		WithHydrationEvent(),
		WithTransitionError(errors.New("network timeout"), true),
	); err != nil {
		t.Fatalf("transition to error: %v", err)
	}
	current, _ := store.Get(context.Background(), "id-3")
	if current.LastError == nil || current.LastError.Message != "network timeout" || !current.LastError.Temporary {
		t.Fatalf("expected error metadata recorded %+v", current.LastError)
	}
	if current.Hydration.Error == nil {
		t.Fatalf("expected hydration error field")
	}
}

// memoryStore is a simple in-memory implementation of Store for unit tests.
type memoryStore struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	clock   Clock
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		entries: make(map[string]*Entry),
		clock:   systemClock{},
	}
}

func (m *memoryStore) Get(_ context.Context, id string) (*Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.entries[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *entry
	return &copy, nil
}

func (m *memoryStore) Save(_ context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = m.clock.Now()
	}
	entry.UpdatedAt = m.clock.Now()
	if err := entry.Validate(); err != nil {
		return err
	}
	copy := *entry
	m.entries[entry.ID] = &copy
	return nil
}

func (m *memoryStore) Update(_ context.Context, id string, fn func(*Entry) error) (*Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[id]
	if !ok {
		return nil, ErrNotFound
	}
	copy := *entry
	if err := fn(&copy); err != nil {
		return nil, err
	}
	copy.UpdatedAt = m.clock.Now()
	if err := copy.Validate(); err != nil {
		return nil, err
	}
	m.entries[id] = &copy
	return &copy, nil
}
