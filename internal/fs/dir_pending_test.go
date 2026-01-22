package fs

import (
	"strings"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

func TestIT_FS_DirPending_RemoteVisibilitySurvivesRefresh(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "PendingRemoteDirectoryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, data interface{}) {
		assert := framework.NewAssert(t)
		fsFixture := getFSTestFixture(t, data)
		if fsFixture.MockClient == nil {
			t.Skip("pending visibility test requires mock graph client")
		}

		filesystem, ok := fsFixture.FS.(*Filesystem)
		if !ok {
			t.Fatalf("expected *Filesystem, got %T", fsFixture.FS)
		}

		parentID := fsFixture.RootID
		childID := "remote-child-id"
		childName := "pkg"

		childItem := &graph.DriveItem{
			ID:   childID,
			Name: childName,
			Parent: &graph.DriveItemParent{
				ID: parentID,
			},
			Folder:  &graph.Folder{},
			ModTime: func() *time.Time { now := time.Now(); return &now }(),
		}
		childInode := NewInodeDriveItem(childItem)
		filesystem.InsertChild(parentID, childInode)
		filesystem.markChildPendingRemote(childID)

		resource := "/me/drive/items/" + parentID + "/children"
		fsFixture.MockClient.AddMockItems(resource, []*graph.DriveItem{})
		origManager := filesystem.metadataRequestManager
		filesystem.metadataRequestManager = nil
		defer func() {
			filesystem.metadataRequestManager = origManager
		}()

		children, err := filesystem.getChildrenID(parentID, fsFixture.Auth, true)
		assert.NoError(err, "force refresh should succeed")
		_, exists := children[strings.ToLower(childName)]
		assert.True(exists, "pending directory should remain accessible")
		assert.True(filesystem.isChildPendingRemote(childID), "pending flag should still be set while Graph misses entry")

		remoteVisible := &graph.DriveItem{
			ID:   childID,
			Name: childName,
			Parent: &graph.DriveItemParent{
				ID: parentID,
			},
			Folder:  &graph.Folder{},
			ModTime: childItem.ModTime,
		}
		fsFixture.MockClient.AddMockItems(resource, []*graph.DriveItem{remoteVisible})

		children, err = filesystem.getChildrenID(parentID, fsFixture.Auth, true)
		assert.NoError(err, "second refresh should succeed once Graph lists the child")
		_, exists = children[strings.ToLower(childName)]
		assert.True(exists, "directory should remain accessible after Graph lists it")
		assert.False(filesystem.isChildPendingRemote(childID), "pending flag should clear when Graph reports the child")
	})
}
