package graph

import (
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
)

// StandardTestFile returns a standard test file content with predictable content
func StandardTestFile() []byte {
	return []byte("This is a standard test file content")
}

// CreateDriveItemFixture creates a DriveItem fixture for testing
func CreateDriveItemFixture(name string, isFolder bool) *graph.DriveItem {
	now := time.Now()
	item := &graph.DriveItem{
		ID:      "test-id-" + name,
		Name:    name,
		Size:    1024,
		ModTime: &now,
		Parent: &graph.DriveItemParent{
			ID:        "parent-id",
			DriveID:   "drive-id",
			DriveType: graph.DriveTypePersonal,
		},
		ETag: "etag-" + name,
	}

	if isFolder {
		item.Folder = &graph.Folder{
			ChildCount: 0,
		}
	} else {
		item.File = &graph.File{
			Hashes: graph.Hashes{
				SHA1Hash:     "sha1-hash-value",
				QuickXorHash: "quickxor-hash-value",
			},
		}
	}

	return item
}

// CreateFileItemFixture creates a DriveItem fixture representing a file
func CreateFileItemFixture(name string, size uint64, content []byte) *graph.DriveItem {
	item := CreateDriveItemFixture(name, false)
	item.Size = size

	// If content is provided, update the hash values
	if content != nil && len(content) > 0 {
		// In a real implementation, we would calculate actual hashes here
		// For testing purposes, we just use placeholder values
		item.File.Hashes.SHA1Hash = "sha1-" + name
		item.File.Hashes.QuickXorHash = "qxh-" + name
	}

	return item
}

// CreateFolderItemFixture creates a DriveItem fixture representing a folder
func CreateFolderItemFixture(name string, childCount uint32) *graph.DriveItem {
	item := CreateDriveItemFixture(name, true)
	item.Folder.ChildCount = childCount
	return item
}

// CreateDeletedItemFixture creates a DriveItem fixture representing a deleted item
func CreateDeletedItemFixture(name string, isFolder bool) *graph.DriveItem {
	item := CreateDriveItemFixture(name, isFolder)
	item.Deleted = &graph.Deleted{
		State: "deleted",
	}
	return item
}

// CreateChildrenFixture creates a slice of DriveItem fixtures representing children of a folder
func CreateChildrenFixture(parentID string, count int) []*graph.DriveItem {
	children := make([]*graph.DriveItem, count)

	for i := 0; i < count; i++ {
		isFolder := i%2 == 0 // Alternate between files and folders
		name := ""
		if isFolder {
			name = "folder-" + string(rune('A'+i))
		} else {
			name = "file-" + string(rune('A'+i)) + ".txt"
		}

		child := CreateDriveItemFixture(name, isFolder)
		child.Parent.ID = parentID

		children[i] = child
	}

	return children
}

// CreateNestedFolderStructure creates a nested folder structure for testing
func CreateNestedFolderStructure(depth int) *graph.DriveItem {
	root := CreateFolderItemFixture("root", uint32(depth))

	if depth <= 0 {
		return root
	}

	current := root
	for i := 0; i < depth; i++ {
		child := CreateFolderItemFixture("folder-"+string(rune('A'+i)), uint32(depth-i-1))
		child.Parent.ID = current.ID
		current = child
	}

	return root
}

// CreateDriveItemWithConflict creates a DriveItem fixture with conflict behavior set
func CreateDriveItemWithConflict(name string, isFolder bool, conflictBehavior string) *graph.DriveItem {
	item := CreateDriveItemFixture(name, isFolder)
	item.ConflictBehavior = conflictBehavior
	return item
}
