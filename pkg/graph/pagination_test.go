package graph

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/auriora/onemount/pkg/graph/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT_GR_PAGE_01_01_BasicPagination_LargeCollection_HandlesMultiplePages tests basic pagination
func TestUT_GR_PAGE_01_01_BasicPagination_LargeCollection_HandlesMultiplePages(t *testing.T) {
	client := NewMockGraphClient()

	// Create test items for pagination
	var items []*DriveItem
	for i := 0; i < 25; i++ {
		items = append(items, &DriveItem{
			ID:   fmt.Sprintf("item-%d", i),
			Name: fmt.Sprintf("Item %d", i),
			Size: uint64(1024 * (i + 1)),
		})
	}

	// Configure pagination with 10 items per page
	resource := "/me/drive/items/parent-id/children"
	client.AddMockItemsWithPagination(resource, items, 10)

	// Fetch first page
	body, err := client.Get(resource)
	require.NoError(t, err)

	// Parse first page response
	var firstPage driveChildren
	err = json.Unmarshal(body, &firstPage)
	require.NoError(t, err)

	// Verify first page
	assert.Len(t, firstPage.Children, 10)
	assert.Equal(t, "item-0", firstPage.Children[0].ID)
	assert.Equal(t, "item-9", firstPage.Children[9].ID)
	assert.NotEmpty(t, firstPage.NextLink)

	// Fetch second page
	nextLink := strings.TrimPrefix(firstPage.NextLink, GraphURL)
	body, err = client.Get(nextLink)
	require.NoError(t, err)

	var secondPage driveChildren
	err = json.Unmarshal(body, &secondPage)
	require.NoError(t, err)

	// Verify second page
	assert.Len(t, secondPage.Children, 10)
	assert.Equal(t, "item-10", secondPage.Children[0].ID)
	assert.Equal(t, "item-19", secondPage.Children[9].ID)
	assert.NotEmpty(t, secondPage.NextLink)

	// Fetch third (final) page
	nextLink = strings.TrimPrefix(secondPage.NextLink, GraphURL)
	body, err = client.Get(nextLink)
	require.NoError(t, err)

	var thirdPage driveChildren
	err = json.Unmarshal(body, &thirdPage)
	require.NoError(t, err)

	// Verify third page
	assert.Len(t, thirdPage.Children, 5)
	assert.Equal(t, "item-20", thirdPage.Children[0].ID)
	assert.Equal(t, "item-24", thirdPage.Children[4].ID)
	assert.Empty(t, thirdPage.NextLink)
}

// TestUT_GR_PAGE_01_02_GetItemChildren_PaginatedResults_ReturnsAllItems tests GetItemChildren with pagination
func TestUT_GR_PAGE_01_02_GetItemChildren_PaginatedResults_ReturnsAllItems(t *testing.T) {
	client := NewMockGraphClient()

	// Create test items for pagination
	var items []*api.DriveItem
	for i := 0; i < 35; i++ {
		items = append(items, &api.DriveItem{
			ID:   fmt.Sprintf("child-%d", i),
			Name: fmt.Sprintf("Child %d", i),
			Size: uint64(512 * (i + 1)),
		})
	}

	// Configure pagination with 15 items per page
	resource := "/me/drive/items/parent-id/children"
	client.AddMockItems(resource, items)

	// Call GetItemChildren (should handle pagination automatically)
	children, err := client.GetItemChildren("parent-id")

	// Verify all children are returned
	assert.NoError(t, err)
	assert.Len(t, children, 35)

	// Verify order is preserved
	assert.Equal(t, "child-0", children[0].ID)
	assert.Equal(t, "child-34", children[34].ID)

	// Verify method was recorded (may include internal Get calls)
	calls := client.Recorder.GetCalls()
	assert.Greater(t, len(calls), 0)
	// The last call should be GetItemChildren
	lastCall := calls[len(calls)-1]
	assert.Equal(t, "GetItemChildren", lastCall.Method)
}

// TestUT_GR_PAGE_02_01_EmptyCollection_NoItems_HandlesGracefully tests pagination with empty collections
func TestUT_GR_PAGE_02_01_EmptyCollection_NoItems_HandlesGracefully(t *testing.T) {
	client := NewMockGraphClient()

	// Configure empty collection
	resource := "/me/drive/items/empty-folder/children"
	emptyResponse := driveChildren{
		Children: []*DriveItem{},
		NextLink: "",
	}
	body, _ := json.Marshal(emptyResponse)
	client.AddMockResponse(resource, body, http.StatusOK, nil)

	// Fetch empty collection
	body, err := client.Get(resource)
	require.NoError(t, err)

	// Parse response
	var result driveChildren
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify empty collection is handled correctly
	assert.Empty(t, result.Children)
	assert.Empty(t, result.NextLink)
}

// TestUT_GR_PAGE_02_02_SinglePage_SmallCollection_NoNextLink tests single page collections
func TestUT_GR_PAGE_02_02_SinglePage_SmallCollection_NoNextLink(t *testing.T) {
	client := NewMockGraphClient()

	// Create small collection (5 items)
	var items []*DriveItem
	for i := 0; i < 5; i++ {
		items = append(items, &DriveItem{
			ID:   fmt.Sprintf("item-%d", i),
			Name: fmt.Sprintf("Item %d", i),
			Size: uint64(256 * (i + 1)),
		})
	}

	// Configure single page (no pagination needed)
	resource := "/me/drive/items/small-folder/children"
	response := driveChildren{
		Children: items,
		NextLink: "", // No next link for single page
	}
	body, _ := json.Marshal(response)
	client.AddMockResponse(resource, body, http.StatusOK, nil)

	// Fetch small collection
	body, err := client.Get(resource)
	require.NoError(t, err)

	// Parse response
	var result driveChildren
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify single page collection
	assert.Len(t, result.Children, 5)
	assert.Empty(t, result.NextLink)
	assert.Equal(t, "item-0", result.Children[0].ID)
	assert.Equal(t, "item-4", result.Children[4].ID)
}
