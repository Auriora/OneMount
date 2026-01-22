package fs

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/metadata"
	"github.com/hanwen/go-fuse/v2/fuse"
	bolt "go.etcd.io/bbolt"
)

func TestUT_FS_MetadataStore_EntryFromInodeStateInference(t *testing.T) {
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

func TestUT_FS_MetadataStore_BootstrapMigratesLegacyEntries(t *testing.T) {
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

	if err := fs.bootstrapMetadataStore(); err == nil {
		t.Fatalf("expected bootstrap to require migration, got nil")
	} else if !errors.Is(err, ErrLegacyMetadataPresent) {
		t.Fatalf("expected ErrLegacyMetadataPresent, got %v", err)
	}

	report, err := MigrateLegacyMetadata(db)
	if err != nil {
		t.Fatalf("migrate legacy: %v", err)
	}
	if report.Migrated != 1 {
		t.Fatalf("expected 1 migrated entry, got %d", report.Migrated)
	}

	if err := fs.bootstrapMetadataStore(); err != nil {
		t.Fatalf("bootstrap after migration returned error: %v", err)
	}
}

func TestUT_FS_MetadataStore_InodeFromMetadataEntry(t *testing.T) {
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

func TestUT_FS_MetadataStore_PendingRemoteMetadataUpdates(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "meta.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		t.Fatalf("open bolt: %v", err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketMetadataV2)
		return err
	}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	store, err := metadata.NewBoltStore(db, bucketMetadataV2)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	fs := &Filesystem{
		content:       NewLoopbackCacheWithSize(filepath.Join(dir, "content"), 0),
		metadataStore: store,
	}

	entry := &metadata.Entry{
		ID:    "file-42",
		Name:  "pending.txt",
		State: metadata.ItemStateGhost,
	}
	if err := fs.SaveMetadataEntry(entry); err != nil {
		t.Fatalf("save entry: %v", err)
	}

	fs.markChildPendingRemote("file-42")
	stored, err := fs.GetMetadataEntry("file-42")
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if !stored.PendingRemote {
		t.Fatalf("expected pending remote flag to be set")
	}

	fs.clearChildPendingRemote("file-42")
	stored, err = fs.GetMetadataEntry("file-42")
	if err != nil {
		t.Fatalf("get entry second time: %v", err)
	}
	if stored.PendingRemote {
		t.Fatalf("expected pending remote flag cleared")
	}
}

func TestUT_FS_MetadataStore_GetIDLoadsFromMetadataStore(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "meta.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		t.Fatalf("open bolt: %v", err)
	}
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketMetadataV2)
		return err
	}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	store, err := metadata.NewBoltStore(db, bucketMetadataV2)
	if err != nil {
		t.Fatalf("new bolt store: %v", err)
	}

	fs := &Filesystem{
		content:       NewLoopbackCacheWithSize(filepath.Join(dir, "content"), 0),
		metadataStore: store,
		db:            db,
	}

	entry := &metadata.Entry{
		ID:       "cached-item",
		Name:     "cached.txt",
		ItemType: metadata.ItemKindFile,
		State:    metadata.ItemStateHydrated,
		Mode:     fuse.S_IFREG | 0644,
	}
	if err := fs.SaveMetadataEntry(entry); err != nil {
		t.Fatalf("save metadata entry: %v", err)
	}

	inode := fs.GetID(entry.ID)
	if inode == nil {
		t.Fatalf("expected inode retrieved from metadata store")
	}
	if inode.Name() != entry.Name {
		t.Fatalf("expected inode name %s, got %s", entry.Name, inode.Name())
	}
}
