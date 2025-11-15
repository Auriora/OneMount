package fs

import (
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// registerDriveItem seeds the filesystem cache with a mock DriveItem so tests
// can observe it without performing Graph calls.
func registerDriveItem(fs *Filesystem, parentID string, item *graph.DriveItem) *Inode {
	inode := NewInodeDriveItem(item)
	fs.InsertNodeID(inode)
	fs.InsertChild(parentID, inode)
	return inode
}

func createAndRegisterMockFile(fs *Filesystem, mockClient *graph.MockGraphClient, parentID, name, id, content string) *graph.DriveItem {
	item := helpers.CreateMockFile(mockClient, parentID, name, id, content)
	if item != nil {
		registerDriveItem(fs, parentID, item)
	}
	return item
}

func createAndRegisterMockDirectory(fs *Filesystem, mockClient *graph.MockGraphClient, parentID, name, id string) *graph.DriveItem {
	item := helpers.CreateMockDirectory(mockClient, parentID, name, id)
	if item != nil {
		registerDriveItem(fs, parentID, item)
	}
	return item
}
