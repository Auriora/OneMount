package fs

import (
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/graph/api"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/require"
)

// Test that GetChildrenID serves from structured metadata without issuing Graph calls when
// in-memory child cache is cold. This verifies the local-first contract for FUSE readdir.
func TestUT_FS_FUSEMetadata_GetChildrenIDUsesMetadataStoreWhenCold(t *testing.T) {
	now := time.Now().UTC()

	// Build a filesystem with an initialized metadata store/state manager.
	fs := newTestFilesystemWithMetadata(t)
	fs.statuses = make(map[string]FileStatusInfo)

	// Set a mock Graph client and keep a recorder to assert zero calls.
	mockClient := graph.NewMockGraphClient()
	recorder := mockClient.GetRecorder()

	parentEntry := &metadata.Entry{
		ID:            "parent",
		Name:          "parent",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		Children:      []string{"child"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	childEntry := &metadata.Entry{
		ID:            "child",
		Name:          "child.txt",
		ParentID:      "parent",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Persist structured metadata.
	require.NoError(t, fs.metadataStore.Save(fs.ctx, parentEntry))
	require.NoError(t, fs.metadataStore.Save(fs.ctx, childEntry))

	// Hydrate in-memory inodes from metadata.
	parentInode := fs.ensureInodeFromMetadataStore("parent")
	require.NotNil(t, parentInode)
	childInode := fs.ensureInodeFromMetadataStore("child")
	require.NotNil(t, childInode)

	// Clear cached children to simulate a cold inode cache while metadata_v2 is warm.
	parentInode.mu.Lock()
	parentInode.children = nil
	parentInode.mu.Unlock()

	// Calling GetChildrenID should rebuild from metadata_v2 without touching Graph
	// (metadataRequestManager is nil, so any Graph attempt would error).
	children, err := fs.GetChildrenID("parent", &graph.Auth{})
	require.NoError(t, err)
	require.Len(t, children, 1)
	entry := children["child.txt"]
	require.NotNil(t, entry)
	require.Equal(t, childInode, entry)

	// Assert no Graph calls were made.
	if recorder != nil {
		calls := recorder.GetCalls()
		require.Len(t, calls, 0, "expected no Graph calls when metadata_v2 is warm")
	} else {
		t.Log("mock recorder unavailable; skipping Graph call count assertion")
	}

	// Ensure mock client cleaned up.
	t.Cleanup(func() {
		if c, ok := recorder.(interface{ GetCalls() []api.MockCall }); ok && len(c.GetCalls()) > 0 {
			t.Fatalf("unexpected Graph calls: %+v", c.GetCalls())
		}
		mockClient.Cleanup()
	})
}

// Lookup-by-name should be satisfied from metadata_v2 without Graph calls when child exists.
func TestUT_FS_FUSEMetadata_GetChildUsesMetadataStoreWhenCold(t *testing.T) {
	now := time.Now().UTC()
	fs := newTestFilesystemWithMetadata(t)

	mockClient := graph.NewMockGraphClient()
	recorder := mockClient.GetRecorder()

	parent := &metadata.Entry{
		ID:            "dir",
		Name:          "dir",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		Children:      []string{"file"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	file := &metadata.Entry{
		ID:            "file",
		Name:          "file.txt",
		ParentID:      "dir",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, fs.metadataStore.Save(fs.ctx, parent))
	require.NoError(t, fs.metadataStore.Save(fs.ctx, file))

	dirInode := fs.ensureInodeFromMetadataStore("dir")
	require.NotNil(t, dirInode)
	fileInode := fs.ensureInodeFromMetadataStore("file")
	require.NotNil(t, fileInode)

	// Clear cached children to force lookup to rebuild from metadata store.
	dirInode.mu.Lock()
	dirInode.children = nil
	dirInode.mu.Unlock()

	found, err := fs.GetChild("dir", "file.txt", &graph.Auth{})
	require.NoError(t, err)
	require.Equal(t, fileInode, found)

	if recorder != nil {
		require.Len(t, recorder.GetCalls(), 0, "expected no Graph calls for lookup served from metadata")
	}
	mockClient.Cleanup()
}

// Negative lookup should not hit Graph when parent metadata is present but empty.
func TestUT_FS_FUSEMetadata_GetChildMissingDoesNotHitGraph(t *testing.T) {
	now := time.Now().UTC()
	fs := newTestFilesystemWithMetadata(t)

	mockClient := graph.NewMockGraphClient()
	recorder := mockClient.GetRecorder()

	parent := &metadata.Entry{
		ID:            "dir",
		Name:          "dir",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		Children:      []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, fs.metadataStore.Save(fs.ctx, parent))

	dirInode := fs.ensureInodeFromMetadataStore("dir")
	require.NotNil(t, dirInode)

	// Clear cached children to keep inode cold.
	dirInode.mu.Lock()
	dirInode.children = nil
	dirInode.mu.Unlock()

	child, err := fs.GetChild("dir", "missing.txt", &graph.Auth{})
	require.Nil(t, child)
	require.Error(t, err)

	if recorder != nil {
		require.Len(t, recorder.GetCalls(), 0, "expected no Graph calls on negative lookup with warm metadata")
	}
	mockClient.Cleanup()
}

// OpenDir should succeed offline when metadata exists, rebuilding children from metadata_v2.
func TestUT_FS_FUSEMetadata_OpenDirUsesMetadataOffline(t *testing.T) {
	now := time.Now().UTC()
	fs := newTestFilesystemWithMetadata(t)
	fs.opendirs = make(map[uint64][]*Inode)
	fs.auth = &graph.Auth{}

	parent := &metadata.Entry{
		ID:            "dir",
		Name:          "dir",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		Children:      []string{"child"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	child := &metadata.Entry{
		ID:            "child",
		Name:          "child.txt",
		ParentID:      "dir",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, fs.metadataStore.Save(fs.ctx, parent))
	require.NoError(t, fs.metadataStore.Save(fs.ctx, child))

	directory := fs.ensureInodeFromMetadataStore(parent.ID)
	require.NotNil(t, directory)
	childInode := fs.ensureInodeFromMetadataStore(child.ID)
	require.NotNil(t, childInode)

	// Clear cached children to force rebuild from metadata_v2
	directory.mu.Lock()
	directory.children = nil
	directory.mu.Unlock()

	graph.SetOperationalOffline(true)
	t.Cleanup(func() { graph.SetOperationalOffline(false) })

	status := fs.OpenDir(nil, &fuse.OpenIn{InHeader: fuse.InHeader{NodeId: directory.NodeID()}}, &fuse.OpenOut{})
	require.Equal(t, fuse.OK, status)

	fs.opendirsM.RLock()
	entries := fs.opendirs[directory.NodeID()]
	fs.opendirsM.RUnlock()
	require.NotNil(t, entries)
	require.True(t, len(entries) >= 3)
	require.Equal(t, childInode.NodeID(), entries[2].NodeID())
}

// Lookup should be satisfied from metadata without triggering Graph when offline.
func TestUT_FS_FUSEMetadata_LookupUsesMetadataOffline(t *testing.T) {
	now := time.Now().UTC()
	fs := newTestFilesystemWithMetadata(t)
	fs.auth = &graph.Auth{}

	parent := &metadata.Entry{
		ID:            "dir",
		Name:          "dir",
		ItemType:      metadata.ItemKindDirectory,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		Children:      []string{"child"},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	child := &metadata.Entry{
		ID:            "child",
		Name:          "child.txt",
		ParentID:      "dir",
		ItemType:      metadata.ItemKindFile,
		State:         metadata.ItemStateHydrated,
		OverlayPolicy: metadata.OverlayPolicyRemoteWins,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, fs.metadataStore.Save(fs.ctx, parent))
	require.NoError(t, fs.metadataStore.Save(fs.ctx, child))

	directory := fs.ensureInodeFromMetadataStore(parent.ID)
	require.NotNil(t, directory)
	childInode := fs.ensureInodeFromMetadataStore(child.ID)
	require.NotNil(t, childInode)

	// Clear cached children to force rebuild from metadata store on lookup
	directory.mu.Lock()
	directory.children = nil
	directory.mu.Unlock()

	graph.SetOperationalOffline(true)
	t.Cleanup(func() { graph.SetOperationalOffline(false) })

	out := &fuse.EntryOut{}
	status := fs.Lookup(nil, &fuse.InHeader{NodeId: directory.NodeID()}, "child.txt", out)
	require.Equal(t, fuse.OK, status)
	require.Equal(t, childInode.NodeID(), out.NodeId)
}
