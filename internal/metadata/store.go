package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// ErrNotFound indicates the requested metadata entry was not present in the store.
var ErrNotFound = errors.New("metadata: entry not found")

// Store defines the persistence contract required by the state manager.
type Store interface {
	// Get returns the entry for the provided ID or ErrNotFound.
	Get(ctx context.Context, id string) (*Entry, error)
	// Save persists the given entry, overwriting any existing record.
	Save(ctx context.Context, entry *Entry) error
	// Update atomically loads, mutates via fn, and persists the entry.
	Update(ctx context.Context, id string, fn func(*Entry) error) (*Entry, error)
}

// Clock abstracts time retrieval for deterministic testing.
type Clock interface {
	Now() time.Time
}

// systemClock implements Clock using time.Now.
type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

// BoltStore implements Store using a BBolt bucket.
type BoltStore struct {
	db     *bolt.DB
	bucket []byte
	clock  Clock
}

// BoltStoreOption controls BoltStore construction.
type BoltStoreOption func(*BoltStore)

// WithClock overrides the default system clock.
func WithClock(clock Clock) BoltStoreOption {
	return func(store *BoltStore) {
		if clock != nil {
			store.clock = clock
		}
	}
}

// NewBoltStore constructs a metadata store backed by the provided bucket.
func NewBoltStore(db *bolt.DB, bucket []byte, opts ...BoltStoreOption) (*BoltStore, error) {
	if db == nil {
		return nil, fmt.Errorf("metadata: bolt DB is required")
	}
	if len(bucket) == 0 {
		return nil, fmt.Errorf("metadata: bucket name is required")
	}
	store := &BoltStore{
		db:     db,
		bucket: bucket,
		clock:  systemClock{},
	}
	for _, opt := range opts {
		opt(store)
	}
	return store, nil
}

// Get returns a deep copy of the stored entry.
func (s *BoltStore) Get(_ context.Context, id string) (*Entry, error) {
	var entry *Entry
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return fmt.Errorf("metadata: bucket %q missing", string(s.bucket))
		}
		raw := b.Get([]byte(id))
		if len(raw) == 0 {
			return ErrNotFound
		}
		var decoded Entry
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return err
		}
		entry = &decoded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// Save validates and writes the entry to the bucket.
func (s *BoltStore) Save(_ context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("metadata: entry is nil")
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = s.clock.Now()
	}
	entry.UpdatedAt = s.clock.Now()
	if err := entry.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return fmt.Errorf("metadata: bucket %q missing", string(s.bucket))
		}
		return b.Put([]byte(entry.ID), data)
	})
}

// Update executes fn atomically against the entry identified by id.
func (s *BoltStore) Update(_ context.Context, id string, fn func(*Entry) error) (*Entry, error) {
	var result *Entry
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return fmt.Errorf("metadata: bucket %q missing", string(s.bucket))
		}
		raw := b.Get([]byte(id))
		if len(raw) == 0 {
			return ErrNotFound
		}
		var entry Entry
		if err := json.Unmarshal(raw, &entry); err != nil {
			return err
		}
		if err := fn(&entry); err != nil {
			return err
		}
		entry.UpdatedAt = s.clock.Now()
		if err := entry.Validate(); err != nil {
			return err
		}
		data, err := json.Marshal(&entry)
		if err != nil {
			return err
		}
		if err := b.Put([]byte(entry.ID), data); err != nil {
			return err
		}
		result = &entry
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
