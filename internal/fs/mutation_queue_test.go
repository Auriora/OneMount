package fs

import (
	"context"
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
