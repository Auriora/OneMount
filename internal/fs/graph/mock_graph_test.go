package graph

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockGraphClient(t *testing.T) {
	t.Parallel()

	t.Run("basic functionality", func(t *testing.T) {
		client := NewMockGraphClient()

		// Test default values
		assert.Equal(t, "mock-access-token", client.Auth.AccessToken)
		assert.Equal(t, "mock@example.com", client.Auth.Account)
		assert.Greater(t, client.Auth.ExpiresAt, time.Now().Unix())
	})

	t.Run("mock item responses", func(t *testing.T) {
		client := NewMockGraphClient()

		// Create a mock item
		mockItem := &DriveItem{
			ID:   "test-id",
			Name: "test-file.txt",
			Size: 1024,
			File: &File{},
		}

		// Add the mock item response
		resource := ResourcePath("/test-file.txt")
		client.AddMockItem(resource, mockItem)

		// Test GetItemPath
		item, err := client.GetItemPath("/test-file.txt")
		require.NoError(t, err)
		assert.Equal(t, "test-id", item.ID)
		assert.Equal(t, "test-file.txt", item.Name)
		assert.Equal(t, uint64(1024), item.Size)
		assert.False(t, item.IsDir())
	})

	t.Run("mock children responses", func(t *testing.T) {
		client := NewMockGraphClient()

		// Create mock items
		mockItems := []*DriveItem{
			{
				ID:     "folder-id",
				Name:   "test-folder",
				Folder: &Folder{},
			},
			{
				ID:   "file-id",
				Name: "test-file.txt",
				File: &File{},
			},
		}

		// Add the mock children response
		resource := childrenPath("/test-parent")
		client.AddMockItems(resource, mockItems)

		// Test GetItemChildrenPath
		children, err := client.GetItemChildrenPath("/test-parent")
		require.NoError(t, err)
		require.Len(t, children, 2)

		assert.Equal(t, "folder-id", children[0].ID)
		assert.Equal(t, "test-folder", children[0].Name)
		assert.True(t, children[0].IsDir())

		assert.Equal(t, "file-id", children[1].ID)
		assert.Equal(t, "test-file.txt", children[1].Name)
		assert.False(t, children[1].IsDir())
	})

	t.Run("mock content responses", func(t *testing.T) {
		client := NewMockGraphClient()

		// Create mock content
		mockContent := []byte("This is test content")
		resource := "/me/drive/items/content-id/content"
		client.AddMockResponse(resource, mockContent, http.StatusOK, nil)

		// Test GetItemContent
		content, size, err := client.GetItemContent("content-id")
		require.NoError(t, err)
		assert.Equal(t, mockContent, content)
		assert.Equal(t, uint64(len(mockContent)), size)

		// Test GetItemContentStream
		var buf bytes.Buffer
		size, err = client.GetItemContentStream("content-id", &buf)
		require.NoError(t, err)
		assert.Equal(t, uint64(len(mockContent)), size)
		assert.Equal(t, string(mockContent), buf.String())
	})

	t.Run("mock error responses", func(t *testing.T) {
		client := NewMockGraphClient()

		// Create a mock error response
		mockError := errors.New("item not found")
		resource := ResourcePath("/nonexistent-file.txt")
		client.AddMockResponse(resource, nil, http.StatusNotFound, mockError)

		// Test error handling
		_, err := client.GetItemPath("/nonexistent-file.txt")
		assert.Error(t, err)
		assert.Equal(t, "item not found", err.Error())
	})

	t.Run("mock mkdir", func(t *testing.T) {
		client := NewMockGraphClient()

		// Mock the response for the POST request
		mockFolder := &DriveItem{
			ID:     "new-folder-id",
			Name:   "new-folder",
			Folder: &Folder{},
		}
		resource := childrenPathID("parent-id")
		client.AddMockItem(resource, mockFolder)

		// Test Mkdir
		folder, err := client.Mkdir("new-folder", "parent-id")
		require.NoError(t, err)
		assert.Equal(t, "new-folder-id", folder.ID)
		assert.Equal(t, "new-folder", folder.Name)
		assert.True(t, folder.IsDir())
	})

	t.Run("mock rename", func(t *testing.T) {
		client := NewMockGraphClient()

		// Test Rename (just verifying it doesn't error with default responses)
		err := client.Rename("item-id", "new-name", "parent-id")
		assert.NoError(t, err)
	})

	t.Run("mock remove", func(t *testing.T) {
		client := NewMockGraphClient()

		// Test Remove (just verifying it doesn't error with default responses)
		err := client.Remove("item-id")
		assert.NoError(t, err)
	})

	t.Run("global request failure", func(t *testing.T) {
		client := NewMockGraphClient()
		client.ShouldFailRequest = true

		// All requests should fail
		_, err := client.GetItemPath("/any-path")
		assert.Error(t, err)
		assert.Equal(t, "mock request failure", err.Error())

		_, err = client.GetItem("any-id")
		assert.Error(t, err)
		assert.Equal(t, "mock request failure", err.Error())

		_, err = client.GetItemChildren("any-id")
		assert.Error(t, err)
		assert.Equal(t, "mock request failure", err.Error())
	})
}
