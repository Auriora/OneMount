package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/graph"
)

func ensureMockGraphRoot(tb testing.TB) *graph.MockGraphClient {
	tb.Helper()
	mockClient := graph.NewMockGraphClient()
	rootItem := &graph.DriveItem{
		ID:   "root",
		Name: "root",
		Folder: &graph.Folder{
			ChildCount: 0,
		},
	}
	mockClient.AddMockItem("/me/drive/root", rootItem)
	mockClient.AddMockItem("/me/drive/items/root", rootItem)
	mockClient.AddMockItems("/me/drive/items/root/children", []*graph.DriveItem{})
	tb.Cleanup(func() {
		mockClient.Cleanup()
	})
	return mockClient
}
