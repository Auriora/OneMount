package fs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/metadata"
	"github.com/stretchr/testify/require"
)

func TestQueueRemoteDeleteTransitionsToDeleted(t *testing.T) {
	fs := setupEvictionTestFS(t, 10)

	// Seed metadata entry
	parent := NewInode("parent", 0755, nil)
	parent.DriveItem.ID = "parent"
	registerHydratedEntry(t, fs, parent)

	child := NewInode("child.txt", 0644, parent)
	child.DriveItem.ID = "child-delete"
	registerHydratedEntry(t, fs, child)
	// Persist child so state manager has an entry to transition
	require.NoError(t, fs.SaveMetadataEntry(fs.metadataEntryFromInode(child.ID(), child, time.Now().UTC())))

	// Use test hook to avoid async retry timing
	require.NoError(t, fs.queueRemoteDeleteTestHook(child.ID()))

	entry, err := fs.metadataStore.Get(context.Background(), child.ID())
	require.NoError(t, err)
	require.Equal(t, metadata.ItemStateDeleted, entry.State)
}

func TestMutationQueueReturnsImmediatelyAndProcessesWork(t *testing.T) {
	fs := newTestFilesystemWithMetadata(t)
	fs.ctx = context.Background()
	fs.startMutationQueue()
	t.Cleanup(func() {
		fs.stopMutationQueue()
		fs.Wg.Wait()
	})

	var executed int32
	start := time.Now()
	fs.runMutationWithRetry("noop", "local", func() error {
		atomic.AddInt32(&executed, 1)
		return nil
	})
	if time.Since(start) > 20*time.Millisecond {
		t.Fatalf("mutation enqueue should not block callers")
	}
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&executed) == 1
	}, 500*time.Millisecond, 10*time.Millisecond, "mutation worker should execute job")
}
