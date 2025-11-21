package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/metadata"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/require"
)

func TestStatsReportsMetadataStates(t *testing.T) {
	fs := setupEvictionTestFS(t, 10)

	parent := NewInode("parent", fuse.S_IFDIR|0755, nil)
	parent.DriveItem.ID = "parent"
	registerHydratedEntry(t, fs, parent)

	file := NewInode("file.txt", fuse.S_IFREG|0644, parent)
	file.DriveItem.ID = "file-id"
	registerHydratedEntry(t, fs, file)

	stats, err := fs.GetStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	if stats.MetadataStateCounts == nil {
		t.Fatalf("expected metadata state counts to be populated")
	}
	if stats.MetadataStateCounts[string(metadata.ItemStateHydrated)] == 0 {
		t.Fatalf("expected hydrated count > 0")
	}
	if stats.HydrationHydrated == 0 {
		t.Fatalf("expected hydration hydrated summary to be populated")
	}
}
