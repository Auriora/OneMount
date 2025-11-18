package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/graph"
)

func TestInodeMakeAttrReportsBlocksUsingMetadata(t *testing.T) {
	t.Run("regular file", func(t *testing.T) {
		inode := NewInodeDriveItem(&graph.DriveItem{
			ID:   "file-id",
			Name: "file.bin",
			Size: 1536,
			File: &graph.File{},
		})
		inode.SetNodeID(1)

		attr := inode.makeAttr()
		expectedBlocks := blocksForSize(inode.Size())
		if attr.Blocks != expectedBlocks {
			t.Fatalf("expected %d blocks, got %d", expectedBlocks, attr.Blocks)
		}
		if attr.Size != inode.Size() {
			t.Fatalf("expected size %d, got %d", inode.Size(), attr.Size)
		}
		if attr.Blksize != preferredIOBlockSize {
			t.Fatalf("expected blksize %d, got %d", preferredIOBlockSize, attr.Blksize)
		}
	})

	t.Run("directory placeholder size", func(t *testing.T) {
		inode := NewInodeDriveItem(&graph.DriveItem{
			ID:     "dir-id",
			Name:   "dir",
			Folder: &graph.Folder{},
		})
		inode.SetNodeID(2)

		attr := inode.makeAttr()
		if attr.Size != placeholderDirSize {
			t.Fatalf("expected directory size %d, got %d", placeholderDirSize, attr.Size)
		}
		expectedBlocks := blocksForSize(placeholderDirSize)
		if attr.Blocks != expectedBlocks {
			t.Fatalf("expected %d directory blocks, got %d", expectedBlocks, attr.Blocks)
		}
	})
}
