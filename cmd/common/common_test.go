package common

import (
	"net/http"
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/fs"
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestUT_CMD_01_01_XDGVolumeInfo_VirtualFileBehavior verifies that
// CreateXDGVolumeInfo replaces an existing cloud copy with a local-only virtual
// file and refreshes the cached content.
func TestUT_CMD_01_01_XDGVolumeInfo_VirtualFileBehavior(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "XDGVolumeInfoVirtualFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return fs.NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*fs.Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		remoteID := "remote-xdg-id"
		remoteItem := &graph.DriveItem{
			ID:   remoteID,
			Name: ".xdg-volume-info",
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{QuickXorHash: "remotehash=="},
			},
			Size: 85,
		}
		inode := fs.NewInodeDriveItem(remoteItem)
		filesystem.InsertChild(rootID, inode)

		mockClient.AddMockResponse("/me", []byte(`{"userPrincipalName":"virtual@example.com"}`), http.StatusOK, nil)
		mockClient.AddMockResponse("/me/drive/items/"+remoteID, nil, http.StatusNoContent, nil)

		CreateXDGVolumeInfo(filesystem, fsFixture.Auth)

		virtualInode, err := filesystem.GetPath("/.xdg-volume-info", fsFixture.Auth)
		assert.NoError(err, "Virtual .xdg-volume-info should resolve without error")
		if assert.NotNil(virtualInode, "Virtual inode should exist") {
			assert.True(strings.HasPrefix(virtualInode.ID(), "local-"), "Virtual inode must use local ID, got %s", virtualInode.ID())
			assert.True(virtualInode.IsVirtual(), "Inode should be marked virtual")
			expected := TemplateXDGVolumeInfo("virtual@example.com")
			actual := string(virtualInode.ReadVirtualContent(0, len(expected)))
			assert.Equal(expected, actual, "Virtual file content should match template")
		}
		assert.Nil(filesystem.GetID(remoteID), "Remote inode should be removed from cache")

		calls := mockClient.GetRecorder().GetCalls()
		deleted := false
		for _, call := range calls {
			if call.Method != "RoundTrip" || len(call.Args) == 0 {
				continue
			}
			req, ok := call.Args[0].(*http.Request)
			if !ok {
				continue
			}
			if req.Method == http.MethodDelete && strings.Contains(req.URL.Path, "/me/drive/items/"+remoteID) {
				deleted = true
				break
			}
		}
		assert.True(deleted, "Remote .xdg-volume-info should be deleted from OneDrive")
	})
}
