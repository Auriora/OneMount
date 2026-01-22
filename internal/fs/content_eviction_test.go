package fs

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/metadata"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func setupEvictionTestFS(t *testing.T, maxSize int64) *Filesystem {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "meta.db")

	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketMetadataV2)
		return err
	}))

	store, err := metadata.NewBoltStore(db, bucketMetadataV2)
	require.NoError(t, err)
	stateMgr, err := metadata.NewStateManager(store)
	require.NoError(t, err)

	cacheDir := filepath.Join(dir, "content")
	fs := &Filesystem{
		db:            db,
		content:       NewLoopbackCacheWithSize(cacheDir, maxSize),
		metadataStore: store,
		stateManager:  stateMgr,
		timeoutConfig: DefaultTimeoutConfig(),
	}
	fs.content.SetEvictionGuard(fs.shouldEvictContent)
	fs.content.SetEvictionHandler(fs.handleContentEvicted)

	return fs
}

func registerHydratedEntry(t *testing.T, fs *Filesystem, inode *Inode) {
	t.Helper()
	fs.InsertNodeID(inode)
	fs.metadata.Store(inode.ID(), inode)
	fs.persistMetadataEntry(inode.ID(), inode)
	fs.transitionItemState(inode.ID(), metadata.ItemStateHydrated)
}

func TestUT_FS_ContentEviction_TransitionsMetadata(t *testing.T) {
	fs := setupEvictionTestFS(t, 10)

	parent := NewInode("parent", fuse.S_IFDIR|0755, nil)
	parent.DriveItem.ID = "parent"
	registerHydratedEntry(t, fs, parent)

	fileOld := NewInode("old.txt", fuse.S_IFREG|0644, parent)
	fileOld.DriveItem.ID = "file-old"
	registerHydratedEntry(t, fs, fileOld)

	require.NoError(t, fs.content.Insert(fileOld.ID(), []byte("123456"))) // 6 bytes

	fileNew := NewInode("new.txt", fuse.S_IFREG|0644, parent)
	fileNew.DriveItem.ID = "file-new"
	registerHydratedEntry(t, fs, fileNew)

	require.NoError(t, fs.content.Insert(fileNew.ID(), []byte("abcdef"))) // triggers eviction of old entry

	entry, err := fs.GetMetadataEntry(fileOld.ID())
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateGhost, entry.State, "evicted content should transition to GHOST")
}

func TestUT_FS_ContentEviction_PinnedContentNotEvicted(t *testing.T) {
	fs := setupEvictionTestFS(t, 8)

	parent := NewInode("parent", fuse.S_IFDIR|0755, nil)
	parent.DriveItem.ID = "parent"
	registerHydratedEntry(t, fs, parent)

	pinned := NewInode("pinned.txt", fuse.S_IFREG|0644, parent)
	pinned.DriveItem.ID = "file-pinned"
	registerHydratedEntry(t, fs, pinned)
	_, err := fs.UpdateMetadataEntry(pinned.ID(), func(entry *metadata.Entry) error {
		entry.Pin.Mode = metadata.PinModeAlways
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, fs.content.Insert(pinned.ID(), []byte("123456")))

	other := NewInode("other.txt", fuse.S_IFREG|0644, parent)
	other.DriveItem.ID = "file-other"
	registerHydratedEntry(t, fs, other)

	insertErr := fs.content.Insert(other.ID(), []byte("abcd"))
	require.Error(t, insertErr, "eviction should fail when only pinned content is available")

	_, statErr := os.Stat(fs.content.contentPath(pinned.ID()))
	require.NoError(t, statErr, "pinned file should remain in cache")

	entry, err := fs.GetMetadataEntry(pinned.ID())
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateHydrated, entry.State, "pinned file should remain hydrated")
}

func TestUT_FS_ContentEviction_PinnedContentAutoHydratesAfterEviction(t *testing.T) {
	fs := setupEvictionTestFS(t, 10)

	parent := NewInode("parent", fuse.S_IFDIR|0755, nil)
	parent.DriveItem.ID = "parent"
	registerHydratedEntry(t, fs, parent)

	pinned := NewInode("pinned.txt", fuse.S_IFREG|0644, parent)
	pinned.DriveItem.ID = "file-pinned-auto"
	registerHydratedEntry(t, fs, pinned)
	_, err := fs.UpdateMetadataEntry(pinned.ID(), func(entry *metadata.Entry) error {
		entry.Pin.Mode = metadata.PinModeAlways
		entry.ItemType = metadata.ItemKindFile
		return nil
	})
	require.NoError(t, err)

	var autoHydrateCount int32
	fs.SetTestHooks(&FilesystemTestHooks{
		AutoHydrateHook: func(_ *Filesystem, id string) bool {
			if id == pinned.ID() {
				atomic.AddInt32(&autoHydrateCount, 1)
			}
			return true
		},
	})
	defer fs.ClearTestHooks()

	fs.handleContentEvicted(pinned.ID())
	require.Equal(t, int32(1), atomic.LoadInt32(&autoHydrateCount), "pinned item should trigger auto hydration")
}

func TestUT_FS_ContentEviction_TransitionsToGhost(t *testing.T) {
	fs := setupEvictionTestFS(t, 10)
	parent := NewInode("parent", fuse.S_IFDIR|0755, nil)
	parent.DriveItem.ID = "parent"
	registerHydratedEntry(t, fs, parent)

	file := NewInode("evict-me.txt", fuse.S_IFREG|0644, parent)
	file.DriveItem.ID = "file-evict"
	registerHydratedEntry(t, fs, file)

	fs.handleContentEvicted(file.ID())

	entry, err := fs.GetMetadataEntry(file.ID())
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateGhost, entry.State, "evicted content should mark entry ghost")
}
