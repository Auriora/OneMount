package fs

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/metadata"
	"github.com/hanwen/go-fuse/v2/fuse"
	bolt "go.etcd.io/bbolt"
)

func TestMetadataEntryFromInodeStateInference(t *testing.T) {
	cacheDir := t.TempDir()
	fs := &Filesystem{
		content: NewLoopbackCacheWithSize(filepath.Join(cacheDir, "content"), 0),
	}

	parent := NewInode("parent", fuse.S_IFDIR|0755, nil)
	child := NewInode("child.txt", fuse.S_IFREG|0644, parent)
	child.virtual = true

	now := time.Now().UTC()
	entry := fs.metadataEntryFromInode("local-1", child, now)
	if entry == nil {
		t.Fatalf("expected entry")
	}
	if !entry.Virtual {
		t.Fatalf("expected virtual flag set")
	}
	if entry.State != metadata.ItemStateHydrated {
		t.Fatalf("expected hydrated state, got %s", entry.State)
	}
	if entry.OverlayPolicy != metadata.OverlayPolicyLocalWins {
		t.Fatalf("expected overlay local wins, got %s", entry.OverlayPolicy)
	}
}

func TestBootstrapMetadataStoreMigratesLegacyEntries(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "metadata.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		t.Fatalf("failed to open bolt db: %v", err)
	}
	defer db.Close()

	inode := NewInode("legacy.txt", fuse.S_IFREG|0644, nil)
	legacyPayload := inode.AsJSON()

	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketMetadata); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bucketMetadataV2); err != nil {
			return err
		}
		return tx.Bucket(bucketMetadata).Put([]byte(inode.ID()), legacyPayload)
	}); err != nil {
		t.Fatalf("failed to seed legacy metadata: %v", err)
	}

	fs := &Filesystem{
		db:      db,
		content: NewLoopbackCacheWithSize(filepath.Join(dir, "content"), 0),
	}

	if err := fs.bootstrapMetadataStore(); err != nil {
		t.Fatalf("bootstrapMetadataStore returned error: %v", err)
	}

	if err := db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucketMetadataV2); b != nil {
			if got := b.Get([]byte(inode.ID())); len(got) == 0 {
				t.Fatalf("expected migrated entry for %s", inode.ID())
			}
		} else {
			t.Fatalf("metadata_v2 bucket missing")
		}
		return nil
	}); err != nil {
		t.Fatalf("failed to verify migration: %v", err)
	}
}

func TestInodeFromMetadataEntry(t *testing.T) {
	fs := &Filesystem{}
	lastModified := time.Date(2025, time.November, 19, 12, 0, 0, 0, time.UTC)
	entry := &metadata.Entry{
		ID:            "item-3",
		Name:          "local.txt",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateDirtyLocal,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		ParentID:      "parent-1",
		Size:          1024,
		ETag:          "etag-123",
		LastModified:  &lastModified,
		Children:      []string{},
		Mode:          fuse.S_IFREG | 0644,
	}

	inode := fs.inodeFromMetadataEntry(entry)
	if inode == nil {
		t.Fatalf("expected inode from metadata entry")
	}
	if inode.ID() != entry.ID {
		t.Fatalf("expected inode ID %s got %s", entry.ID, inode.ID())
	}
	if !inode.hasChanges {
		t.Fatalf("expected hasChanges inferred from DIRTY_LOCAL state")
	}
	if inode.DriveItem.Parent == nil || inode.DriveItem.Parent.ID != entry.ParentID {
		t.Fatalf("expected parent relationship on inode")
	}
	if inode.DriveItem.ETag != entry.ETag {
		t.Fatalf("expected etag propagation")
	}
}
