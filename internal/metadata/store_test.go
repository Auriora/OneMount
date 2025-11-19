package metadata

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestBoltStoreSaveAndGet(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "metadata.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		t.Fatalf("open bolt: %v", err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("metadata_v2"))
		return err
	}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	store, err := NewBoltStore(db, []byte("metadata_v2"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	entry := &Entry{
		ID:    "item-1",
		Name:  "file.txt",
		State: ItemStateHydrated,
	}
	if err := store.Save(context.Background(), entry); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.Get(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != entry.Name || got.State != ItemStateHydrated {
		t.Fatalf("unexpected entry: %+v", got)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set")
	}
}

func TestBoltStoreUpdate(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "metadata.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		t.Fatalf("open bolt: %v", err)
	}
	defer db.Close()
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("metadata_v2"))
		return err
	}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}
	store, err := NewBoltStore(db, []byte("metadata_v2"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	entry := &Entry{
		ID:    "item-2",
		Name:  "notes.docx",
		State: ItemStateGhost,
	}
	if err := store.Save(context.Background(), entry); err != nil {
		t.Fatalf("save: %v", err)
	}

	updated, err := store.Update(context.Background(), "item-2", func(e *Entry) error {
		e.State = ItemStateHydrated
		return nil
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.State != ItemStateHydrated {
		t.Fatalf("expected hydrated state, got %s", updated.State)
	}
}
