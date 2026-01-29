package fs

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/stretchr/testify/require"
)

func waitForPrefetchIdle(t *testing.T, fs *Filesystem) {
	t.Helper()
	if fs == nil {
		return
	}
	require.Eventually(t, func() bool {
		return !fs.prefetchActive.Load()
	}, time.Second, 10*time.Millisecond)
}

// TestUT_FS_Prefetch_RecursiveFetchesDirectories verifies recursive prefetch hydrates directories and children.
func TestUT_FS_Prefetch_RecursiveFetchesDirectories(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "PrefetchRecursiveFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		fsFixture := getFSTestFixture(t, fixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		if mockClient == nil {
			t.Skip("Skipping prefetch test: requires mock graph client")
			return
		}

		waitForPrefetchIdle(t, fs)

		dirA := helpers.CreateMockDirectory(mockClient, rootID, "dirA", "prefetch-dir-a")
		require.NotNil(t, dirA)
		dirB := helpers.CreateMockDirectory(mockClient, rootID, "dirB", "prefetch-dir-b")
		require.NotNil(t, dirB)
		rootFile := helpers.CreateMockFile(mockClient, rootID, "root.txt", "prefetch-root-file", "root content")
		require.NotNil(t, rootFile)

		dirA1 := helpers.CreateMockDirectory(mockClient, dirA.ID, "dirA1", "prefetch-dir-a1")
		require.NotNil(t, dirA1)
		fileA := helpers.CreateMockFile(mockClient, dirA.ID, "fileA.txt", "prefetch-file-a", "file A")
		require.NotNil(t, fileA)
		fileA1 := helpers.CreateMockFile(mockClient, dirA1.ID, "fileA1.txt", "prefetch-file-a1", "file A1")
		require.NotNil(t, fileA1)
		fileB := helpers.CreateMockFile(mockClient, dirB.ID, "fileB.txt", "prefetch-file-b", "file B")
		require.NotNil(t, fileB)

		fs.runPrefetch(context.Background())

		rootChildren := fs.getCachedChildrenSnapshot(rootID)
		require.NotNil(t, rootChildren)
		require.Contains(t, rootChildren, strings.ToLower(dirA.Name))
		require.Contains(t, rootChildren, strings.ToLower(dirB.Name))
		require.Contains(t, rootChildren, strings.ToLower(rootFile.Name))

		dirAChildren := fs.getCachedChildrenSnapshot(dirA.ID)
		require.NotNil(t, dirAChildren)
		require.Contains(t, dirAChildren, strings.ToLower(dirA1.Name))
		require.Contains(t, dirAChildren, strings.ToLower(fileA.Name))

		dirA1Children := fs.getCachedChildrenSnapshot(dirA1.ID)
		require.NotNil(t, dirA1Children)
		require.Contains(t, dirA1Children, strings.ToLower(fileA1.Name))

		dirBChildren := fs.getCachedChildrenSnapshot(dirB.ID)
		require.NotNil(t, dirBChildren)
		require.Contains(t, dirBChildren, strings.ToLower(fileB.Name))

		entry, err := fs.GetMetadataEntry(dirA.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, entry.State)
	})
}

// TestUT_FS_Prefetch_DoesNotFetchFileContent ensures metadata prefetch never hydrates file data.
func TestUT_FS_Prefetch_DoesNotFetchFileContent(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "PrefetchMetadataOnlyFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		fsFixture := getFSTestFixture(t, fixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		if mockClient == nil {
			t.Skip("Skipping prefetch test: requires mock graph client")
			return
		}

		waitForPrefetchIdle(t, fs)

		file := helpers.CreateMockFile(mockClient, rootID, "metadata-only.txt", "prefetch-meta-only", "metadata only")
		require.NotNil(t, file)

		fs.runPrefetch(context.Background())

		require.False(t, fs.content.HasContent(file.ID), "Prefetch should not download file content")
	})
}

// TestUT_FS_Prefetch_ContinuesAfterErrors verifies prefetch logs errors and continues on other directories.
func TestUT_FS_Prefetch_ContinuesAfterErrors(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "PrefetchErrorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		fsFixture := getFSTestFixture(t, fixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		if mockClient == nil {
			t.Skip("Skipping prefetch test: requires mock graph client")
			return
		}

		waitForPrefetchIdle(t, fs)

		errDir := helpers.CreateMockDirectory(mockClient, rootID, "error-dir", "prefetch-error-dir")
		require.NotNil(t, errDir)
		okDir := helpers.CreateMockDirectory(mockClient, rootID, "ok-dir", "prefetch-ok-dir")
		require.NotNil(t, okDir)
		okFile := helpers.CreateMockFile(mockClient, okDir.ID, "ok-file.txt", "prefetch-ok-file", "ok")
		require.NotNil(t, okFile)

		errResource := "/me/drive/items/" + errDir.ID + "/children"
		mockClient.SetResponseCallback(errResource, func() ([]byte, int, error) {
			return nil, 0, errors.New("prefetch failure")
		})

		fs.runPrefetch(context.Background())

		okEntry, err := fs.GetMetadataEntry(okDir.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, okEntry.State)

		errEntry, err := fs.GetMetadataEntry(errDir.ID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateError, errEntry.State)
	})
}

// TestUT_FS_Prefetch_RespectsDepthLimit verifies deep trees are still prefetched without runaway recursion.
func TestUT_FS_Prefetch_RespectsDepthLimit(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "PrefetchDepthLimitFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		fsFixture := getFSTestFixture(t, fixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		if mockClient == nil {
			t.Skip("Skipping prefetch test: requires mock graph client")
			return
		}

		waitForPrefetchIdle(t, fs)

		parentID := rootID
		deepestID := ""
		for i := 0; i < prefetchDepthLimit+2; i++ {
			dirID := fmt.Sprintf("prefetch-deep-%d", i)
			dirName := fmt.Sprintf("deep-%d", i)
			dirItem := helpers.CreateMockDirectory(mockClient, parentID, dirName, dirID)
			require.NotNil(t, dirItem)
			parentID = dirID
			deepestID = dirID
		}

		fs.runPrefetch(context.Background())

		deepEntry, err := fs.GetMetadataEntry(deepestID)
		require.NoError(t, err)
		require.Equal(t, metadata.ItemStateHydrated, deepEntry.State)
	})
}

// TestUT_FS_Prefetch_UsesBackgroundPriority ensures prefetch queues children requests with background priority.
func TestUT_FS_Prefetch_UsesBackgroundPriority(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "PrefetchPriorityFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		fsFixture := getFSTestFixture(t, fixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		if mockClient == nil {
			t.Skip("Skipping prefetch test: requires mock graph client")
			return
		}

		waitForPrefetchIdle(t, fs)

		if fs.metadataRequestManager != nil {
			fs.metadataRequestManager.Stop()
		}
		fs.metadataRequestManager = NewMetadataRequestManager(fs, 1, 10, 10)
		fs.metadataRequestManager.Start()

		blockerID := "prefetch-blocker"
		blockerResource := "/me/drive/items/" + blockerID + "/children"
		blockerStarted := make(chan struct{})
		blockerRelease := make(chan struct{})
		mockClient.SetResponseCallback(blockerResource, func() ([]byte, int, error) {
			select {
			case <-blockerStarted:
			default:
				close(blockerStarted)
			}
			<-blockerRelease
			return []byte(`{"value":[]}`), 200, nil
		})

		err := fs.metadataRequestManager.QueueChildrenRequest(blockerID, fs.auth, PriorityForeground, func(items []*graph.DriveItem, err error) {})
		require.NoError(t, err)
		<-blockerStarted

		prefetchDir := helpers.CreateMockDirectory(mockClient, rootID, "prefetch-dir", "prefetch-priority-dir")
		require.NotNil(t, prefetchDir)
		prefetchFile := helpers.CreateMockFile(mockClient, prefetchDir.ID, "prefetch.txt", "prefetch-priority-file", "prefetch")
		require.NotNil(t, prefetchFile)

		prefetchErrCh := make(chan error, 1)
		go func() {
			_, err := fs.prefetchDirectory(context.Background(), prefetchDir.ID, 0)
			prefetchErrCh <- err
		}()

		require.Eventually(t, func() bool {
			high, low := fs.metadataRequestManager.GetQueueStats()
			return low > 0 && high == 0
		}, time.Second, 10*time.Millisecond)

		close(blockerRelease)

		select {
		case err := <-prefetchErrCh:
			require.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("prefetch did not complete after releasing blocked worker")
		}
	})
}
