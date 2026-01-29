package fs

import (
	"context"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/stretchr/testify/require"
)

// TestIT_FS_LazyDirectoryLoading_IntegrationFlow covers mount, prefetch, and access behaviors.
func TestIT_FS_LazyDirectoryLoading_IntegrationFlow(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "LazyDirectoryLoadingIntegrationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		fsFixture := getFSTestFixture(t, fixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		if mockClient == nil {
			t.Skip("Skipping lazy directory integration tests: requires mock graph client")
			return
		}

		waitForPrefetchIdle(t, fs)

		dirItem := helpers.CreateMockDirectory(mockClient, rootID, "lazy-dir", "lazy-dir-id")
		require.NotNil(t, dirItem)
		fileItem := helpers.CreateMockFile(mockClient, dirItem.ID, "lazy-file.txt", "lazy-file-id", "lazy")
		require.NotNil(t, fileItem)
		mockClient.AddMockItems("/me/drive/items/"+dirItem.ID+"/children", []*graph.DriveItem{fileItem})

		dirInode := NewInodeDriveItem(dirItem)
		fs.InsertChild(rootID, dirInode)
		fs.persistMetadataEntry(dirItem.ID, dirInode)
		fs.persistMetadataEntry(rootID, fs.GetID(rootID))

		t.Run("MountPrefetchAccessFlow", func(t *testing.T) {
			fs.runPrefetch(context.Background())
			children, err := fs.GetChildrenID(dirItem.ID, fs.auth)
			require.NoError(t, err)
			require.NotEmpty(t, children, "Prefetch flow should return children")
		})

		t.Run("CachedAccessPerformance", func(t *testing.T) {
			start := time.Now()
			children, err := fs.GetChildrenID(dirItem.ID, fs.auth)
			elapsed := time.Since(start)
			require.NoError(t, err)
			require.NotEmpty(t, children)
			require.Less(t, elapsed, 50*time.Millisecond, "Cached access should be < 50ms")
		})

		t.Run("UncachedAccessTimeout", func(t *testing.T) {
			uncachedDir := helpers.CreateMockDirectory(mockClient, rootID, "uncached-dir", "uncached-dir-id")
			require.NotNil(t, uncachedDir)
			uncachedFile := helpers.CreateMockFile(mockClient, uncachedDir.ID, "uncached-file.txt", "uncached-file-id", "uncached")
			require.NotNil(t, uncachedFile)
			mockClient.AddMockItems("/me/drive/items/"+uncachedDir.ID+"/children", []*graph.DriveItem{uncachedFile})

			uncachedInode := NewInodeDriveItem(uncachedDir)
			fs.InsertChild(rootID, uncachedInode)
			fs.persistMetadataEntry(uncachedDir.ID, uncachedInode)

			if inode := fs.GetID(uncachedDir.ID); inode != nil {
				inode.mu.Lock()
				inode.children = nil
				inode.subdir = 0
				inode.mu.Unlock()
			}

			resource := "/me/drive/items/" + uncachedDir.ID + "/children"
			if response, ok := mockClient.RequestResponses[resource]; ok {
				mockClient.SetResponseCallback(resource, func() ([]byte, int, error) {
					time.Sleep(200 * time.Millisecond)
					return response.Body, response.StatusCode, response.Error
				})
			}

			start := time.Now()
			children, err := fs.GetChildrenID(uncachedDir.ID, fs.auth)
			elapsed := time.Since(start)
			require.NoError(t, err)
			require.NotEmpty(t, children)
			require.Less(t, elapsed, 10*time.Second, "Uncached access should complete within 10 seconds")
		})

		t.Run("StaleCacheRefreshTimeout", func(t *testing.T) {
			staleDir := helpers.CreateMockDirectory(mockClient, rootID, "stale-dir", "stale-dir-id")
			require.NotNil(t, staleDir)
			file1 := helpers.CreateMockFile(mockClient, staleDir.ID, "stale1.txt", "stale-file-1", "stale1")
			file2 := helpers.CreateMockFile(mockClient, staleDir.ID, "stale2.txt", "stale-file-2", "stale2")
			require.NotNil(t, file1)
			require.NotNil(t, file2)
			mockClient.AddMockItems("/me/drive/items/"+staleDir.ID+"/children", []*graph.DriveItem{file1, file2})

			staleInode := NewInodeDriveItem(staleDir)
			fs.InsertChild(rootID, staleInode)
			fs.persistMetadataEntry(staleDir.ID, staleInode)

			_, err := fs.GetChildrenID(staleDir.ID, fs.auth)
			require.NoError(t, err)

			_, err = fs.UpdateMetadataEntry(staleDir.ID, func(entry *metadata.Entry) error {
				if entry == nil {
					return metadata.ErrNotFound
				}
				entry.UpdatedAt = time.Now().Add(-10 * time.Minute)
				return nil
			})
			require.NoError(t, err)

			file3 := helpers.CreateMockFile(mockClient, staleDir.ID, "stale3.txt", "stale-file-3", "stale3")
			require.NotNil(t, file3)
			mockClient.AddMockItems("/me/drive/items/"+staleDir.ID+"/children", []*graph.DriveItem{file1, file2, file3})

			start := time.Now()
			children, err := fs.GetChildrenID(staleDir.ID, fs.auth)
			elapsed := time.Since(start)
			require.NoError(t, err)
			require.NotEmpty(t, children)
			require.Less(t, elapsed, 2*time.Second, "Stale refresh should return within 2 seconds")
		})

		t.Run("MountCompletesQuickly", func(t *testing.T) {
			mountDir := t.TempDir()
			start := time.Now()
			freshFS, err := NewFilesystem(fsFixture.Auth, mountDir, 30)
			elapsed := time.Since(start)
			require.NoError(t, err)
			require.Less(t, elapsed, 2*time.Second, "Mount should complete within 2 seconds")
			if freshFS != nil {
				freshFS.Stop()
			}
		})
	})
}
