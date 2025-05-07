package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT_GR_01_02_MockGraphClient_ConfigurableResponses_ReturnsExpectedValues tests that the mock graph client returns configured responses.
//
//	Test Case ID    UT-GR-01-01
//	Title           Mock Graph Client Configurable Responses
//	Description     Tests that the mock graph client returns configured responses
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Configure a success response for a specific resource
//	                3. Verify the configured response is returned
//	                4. Configure an error response for another resource
//	                5. Verify the configured error is returned
//	Expected Result The mock client returns the configured responses and errors
func TestUT_GR_01_02_MockGraphClient_ConfigurableResponses_ReturnsExpectedValues(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Configure a response for a specific resource
	resource := "/me/drive/items/test-id"
	expectedBody := []byte(`{"id":"test-id","name":"test-item"}`)
	client.AddMockResponse(resource, expectedBody, http.StatusOK, nil)

	// Test that the configured response is returned
	body, err := client.Get(resource)
	assert.NoError(t, err)
	assert.Equal(t, expectedBody, body)

	// Configure an error response
	errorResource := "/me/drive/items/error-id"
	expectedError := errors.New("test error")
	client.AddMockResponse(errorResource, nil, http.StatusBadRequest, expectedError)

	// Test that the configured error is returned
	_, err = client.Get(errorResource)
	assert.Equal(t, expectedError, err)
}

// TestUT_GR_02_02_MockGraphClient_RecordCalls_TracksMethodCalls tests that the mock graph client records API calls.
//
//	Test Case ID    UT-GR-02-01
//	Title           Mock Graph Client Call Recording
//	Description     Tests that the mock graph client records API calls
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Make several API calls (Get, Post, Delete)
//	                3. Get the recorded calls
//	                4. Verify all calls were recorded in the correct order
//	                5. Verify specific call counts using VerifyCall
//	Expected Result All API calls are recorded correctly and can be verified
func TestUT_GR_02_02_MockGraphClient_RecordCalls_TracksMethodCalls(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Make some calls to the client
	resource := "/me/drive/items/test-id"
	client.Get(resource)
	client.Post(resource, strings.NewReader(`{"name":"test"}`))
	client.Delete(resource)

	// Get the recorded calls
	recorder := client.GetRecorder()
	calls := recorder.GetCalls()

	// Verify that all calls were recorded
	assert.Equal(t, 3, len(calls))
	assert.Equal(t, "Get", calls[0].Method)
	assert.Equal(t, "Post", calls[1].Method)
	assert.Equal(t, "Delete", calls[2].Method)

	// Verify specific calls
	assert.True(t, recorder.VerifyCall("Get", 1))
	assert.True(t, recorder.VerifyCall("Post", 1))
	assert.True(t, recorder.VerifyCall("Delete", 1))
	assert.False(t, recorder.VerifyCall("Put", 1))
}

// TestUT_GR_03_02_MockGraphClient_NetworkConditions_SimulatesNetworkIssues tests that the mock graph client can simulate network conditions.
//
//	Test Case ID    UT-GR-03-01
//	Title           Mock Graph Client Network Conditions
//	Description     Tests that the mock graph client can simulate network conditions like latency and packet loss
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Configure network conditions with high latency
//	                3. Make a request and measure the time
//	                4. Verify the request took at least the configured latency
//	                5. Configure network conditions with 100% packet loss
//	                6. Make another request
//	                7. Verify the request fails due to packet loss
//	Expected Result The mock client correctly simulates network conditions
func TestUT_GR_03_02_MockGraphClient_NetworkConditions_SimulatesNetworkIssues(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Configure network conditions with high latency
	client.SetNetworkConditions(100*time.Millisecond, 0, 0)

	// Measure the time it takes to make a request
	start := time.Now()
	_, err := client.Get("/me/drive/items/test-id")
	duration := time.Since(start)

	// Verify that the request took at least the configured latency
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)

	// Configure network conditions with packet loss
	client.SetNetworkConditions(0, 1.0, 0) // 100% packet loss

	// Verify that the request fails due to packet loss
	_, err = client.Get("/me/drive/items/test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated packet loss")
}

// TestUT_GR_04_02_MockGraphClient_CustomBehavior_ConfiguresClientBehavior tests that the mock graph client can be configured with custom behavior.
//
//	Test Case ID    UT-GR-04-01
//	Title           Mock Graph Client Custom Behavior
//	Description     Tests that the mock graph client can be configured with custom behavior
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Configure custom behavior including latency, response delay, and custom parameters
//	                3. Make a request and measure the time
//	                4. Verify the request took at least the configured time
//	                5. Verify the custom behavior parameters were set correctly
//	Expected Result The mock client correctly applies custom behavior configuration
func TestUT_GR_04_02_MockGraphClient_CustomBehavior_ConfiguresClientBehavior(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Configure custom behavior
	client.SetConfig(MockConfig{
		Latency:       50 * time.Millisecond,
		ErrorRate:     0.0,
		ResponseDelay: 50 * time.Millisecond,
		CustomBehavior: map[string]interface{}{
			"retryCount": 3,
		},
	})

	// Measure the time it takes to make a request
	start := time.Now()
	_, err := client.Get("/me/drive/items/test-id")
	duration := time.Since(start)

	// Verify that the request took at least the configured latency + response delay
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)

	// Verify that custom behavior was set
	assert.Equal(t, 3, client.Config.CustomBehavior["retryCount"])
}

// TestUT_GR_05_02_MockGraphClient_GetItemContentStream_LimitsBandwidth tests that the mock graph client can stream content with bandwidth limitations.
//
//	Test Case ID    UT-GR-05-01
//	Title           Mock Graph Client Content Streaming with Bandwidth Limitation
//	Description     Tests that the mock graph client can stream content with bandwidth limitations
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Configure a response for a content resource
//	                3. Set network conditions with bandwidth limitation
//	                4. Call GetItemContentStream and measure the time
//	                5. Verify the content is correctly streamed
//	                6. Verify the operation took time due to bandwidth limitation
//	Expected Result The mock client correctly streams content with bandwidth limitations
func TestUT_GR_05_02_MockGraphClient_GetItemContentStream_LimitsBandwidth(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Configure a response for a specific resource
	resource := "/me/drive/items/test-id/content"
	expectedContent := []byte("test content")
	client.AddMockResponse(resource, expectedContent, http.StatusOK, nil)

	// Test GetItemContentStream with bandwidth limitation
	client.SetNetworkConditions(0, 0, 10) // 10 KB/s

	var buf bytes.Buffer
	start := time.Now()
	size, err := client.GetItemContentStream("test-id", &buf)
	duration := time.Since(start)

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, uint64(len(expectedContent)), size)
	assert.Equal(t, expectedContent, buf.Bytes())

	// Verify that the request took some time due to bandwidth limitation
	// This is a simple check and might be flaky in CI environments
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond)
}

// TestUT_GR_06_02_MockGraphClient_DriveItemOperations_PerformsBasicOperations tests that the mock graph client can perform basic drive item operations.
//
//	Test Case ID    UT-GR-06-01
//	Title           Mock Graph Client Drive Item Operations
//	Description     Tests that the mock graph client can perform basic drive item operations
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Test GetItem operation
//	                3. Test GetItemPath operation
//	                4. Test GetItemChildren operation
//	                5. Test GetItemChildrenPath operation
//	                6. Test Mkdir operation
//	                7. Test Rename operation
//	                8. Test Remove operation
//	                9. Verify all operations were recorded
//	Expected Result All drive item operations work correctly and are recorded
func TestUT_GR_06_02_MockGraphClient_DriveItemOperations_PerformsBasicOperations(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Test GetItem
	item, err := client.GetItem("test-id")
	assert.NoError(t, err)
	assert.Equal(t, "mock-id", item.ID)
	assert.Equal(t, "mock-item", item.Name)

	// Test GetItemPath
	item, err = client.GetItemPath("/test/path")
	assert.NoError(t, err)
	assert.Equal(t, "mock-id", item.ID)
	assert.Equal(t, "mock-item", item.Name)

	// Test GetItemChildren
	children, err := client.GetItemChildren("test-id")
	assert.NoError(t, err)
	assert.Empty(t, children)

	// Test GetItemChildrenPath
	children, err = client.GetItemChildrenPath("/test/path")
	assert.NoError(t, err)
	assert.Empty(t, children)

	// Test Mkdir
	folder, err := client.Mkdir("test-folder", "parent-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-folder", folder.Name)
	assert.NotNil(t, folder.Folder)

	// Test Rename
	err = client.Rename("test-id", "new-name", "parent-id")
	assert.NoError(t, err)

	// Test Remove
	err = client.Remove("test-id")
	assert.NoError(t, err)

	// Verify that all operations were recorded
	recorder := client.GetRecorder()
	assert.True(t, recorder.VerifyCall("GetItem", 1))
	assert.True(t, recorder.VerifyCall("GetItemPath", 1))
	assert.True(t, recorder.VerifyCall("GetItemChildren", 1))
	assert.True(t, recorder.VerifyCall("GetItemChildrenPath", 1))
	assert.True(t, recorder.VerifyCall("Mkdir", 1))
	assert.True(t, recorder.VerifyCall("Rename", 1))
	assert.True(t, recorder.VerifyCall("Remove", 1))
}

// TestUT_GR_07_02_MockGraphClient_ContextCancellation_FailsRequestsWithCanceledContext tests that the mock graph client handles context cancellation.
//
//	Test Case ID    UT-GR-07-01
//	Title           Mock Graph Client Context Cancellation
//	Description     Tests that the mock graph client handles context cancellation correctly
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Create a context and cancel it
//	                3. Make a request with the cancelled context
//	                4. Verify the request fails with context.Canceled error
//	                5. Verify the call was recorded with the error
//	Expected Result Requests with cancelled contexts fail appropriately and are recorded
func TestUT_GR_07_02_MockGraphClient_ContextCancellation_FailsRequestsWithCanceledContext(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test that the request fails due to context cancellation
	_, err := client.GetWithContext(ctx, "/me/drive/items/test-id")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Verify that the call was recorded with the error
	recorder := client.GetRecorder()
	calls := recorder.GetCalls()
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, "GetWithContext", calls[0].Method)
	assert.Equal(t, context.Canceled, calls[0].Result)
}

// TestUT_GR_08_02_MockGraphClient_APIThrottling_SimulatesThrottling tests that the mock graph client can simulate API throttling.
//
//	Test Case ID    UT-GR-08-01
//	Title           Mock Graph Client API Throttling
//	Description     Tests that the mock graph client can simulate API throttling
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Configure API throttling with 100% throttle rate
//	                3. Make a request and measure the time
//	                4. Verify the request fails with a throttling error
//	                5. Verify the request took at least the configured throttle delay
//	Expected Result The mock client correctly simulates API throttling
func TestUT_GR_08_02_MockGraphClient_APIThrottling_SimulatesThrottling(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Configure API throttling
	client.SetConfig(MockConfig{
		ThrottleRate:  1.0, // 100% throttling rate
		ThrottleDelay: 50 * time.Millisecond,
	})

	// Measure the time it takes to make a request
	start := time.Now()
	_, err := client.Get("/me/drive/items/test-id")
	duration := time.Since(start)

	// Verify that the request failed due to throttling
	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated API throttling")

	// Verify that the request took at least the configured throttle delay
	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
}

// TestUT_GR_09_02_MockGraphClient_Pagination_ReturnsPagedResults tests that the mock graph client supports pagination.
//
//	Test Case ID    UT-GR-09-01
//	Title           Mock Graph Client Pagination
//	Description     Tests that the mock graph client supports pagination for large result sets
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Create a large list of items (25 items)
//	                3. Add the items with pagination (10 items per page)
//	                4. Get the first page and verify its contents
//	                5. Extract the next page URL and get the second page
//	                6. Verify the second page contents
//	                7. Extract the next page URL and get the third page
//	                8. Verify the third page contents and that there are no more pages
//	Expected Result The mock client correctly implements pagination for large result sets
func TestUT_GR_09_02_MockGraphClient_Pagination_ReturnsPagedResults(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Create a large list of items
	items := make([]*DriveItem, 0, 25)
	for i := 0; i < 25; i++ {
		items = append(items, &DriveItem{
			ID:   fmt.Sprintf("item-%d", i),
			Name: fmt.Sprintf("Item %d", i),
		})
	}

	// Add the items with pagination (10 items per page)
	resource := "/me/drive/items/parent-id/children"
	client.AddMockItemsWithPagination(resource, items, 10)

	// Get the first page
	body, err := client.Get(resource)
	require.NoError(t, err)

	// Parse the response
	var result driveChildren
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify the first page
	assert.Len(t, result.Children, 10)
	assert.Equal(t, "item-0", result.Children[0].ID)
	assert.Equal(t, "item-9", result.Children[9].ID)
	assert.NotEmpty(t, result.NextLink)

	// Extract the next page URL
	nextLink := strings.TrimPrefix(result.NextLink, GraphURL)
	assert.Contains(t, nextLink, "skiptoken=10")

	// Get the second page
	body, err = client.Get(nextLink)
	require.NoError(t, err)

	// Parse the response
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify the second page
	assert.Len(t, result.Children, 10)
	assert.Equal(t, "item-10", result.Children[0].ID)
	assert.Equal(t, "item-19", result.Children[9].ID)
	assert.NotEmpty(t, result.NextLink)

	// Extract the next page URL
	nextLink = strings.TrimPrefix(result.NextLink, GraphURL)
	assert.Contains(t, nextLink, "skiptoken=20")

	// Get the third page
	body, err = client.Get(nextLink)
	require.NoError(t, err)

	// Parse the response
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	// Verify the third page
	assert.Len(t, result.Children, 5)
	assert.Equal(t, "item-20", result.Children[0].ID)
	assert.Equal(t, "item-24", result.Children[4].ID)
	assert.Empty(t, result.NextLink) // No more pages
}

// TestUT_GR_10_02_MockGraphClient_PaginationWithGetItemChildrenPath_RetrievesAllItems tests that the mock graph client supports pagination with GetItemChildrenPath.
//
//	Test Case ID    UT-GR-10-01
//	Title           Mock Graph Client Pagination with GetItemChildrenPath
//	Description     Tests that the mock graph client supports pagination when using GetItemChildrenPath
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Create a large collection of items (30 items)
//	                3. Add the items with pagination (10 items per page)
//	                4. Retrieve all items using GetItemChildrenPath
//	                5. Verify that all items are retrieved correctly
//	                6. Verify the items are in the correct order
//	                7. Verify that the recorder has the correct calls
//	Expected Result GetItemChildrenPath correctly handles pagination and retrieves all items
func TestUT_GR_10_02_MockGraphClient_PaginationWithGetItemChildrenPath_RetrievesAllItems(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Create a large collection of items (>25)
	items := make([]*DriveItem, 0, 30)
	for i := 0; i < 30; i++ {
		items = append(items, &DriveItem{
			ID:   fmt.Sprintf("item-%d", i),
			Name: fmt.Sprintf("Item %d", i),
		})
	}

	// Add the items with pagination (10 items per page)
	path := "/path/to/folder"
	resource := childrenPath(path)
	client.AddMockItemsWithPagination(resource, items, 10)

	// Retrieve the items using GetItemChildrenPath
	retrievedItems, err := client.GetItemChildrenPath(path)
	require.NoError(t, err)

	// Verify that all items are retrieved correctly
	assert.Len(t, retrievedItems, 30, "Should retrieve all 30 items")

	// Verify the items are in the correct order
	for i := 0; i < 30; i++ {
		assert.Equal(t, fmt.Sprintf("item-%d", i), retrievedItems[i].ID)
		assert.Equal(t, fmt.Sprintf("Item %d", i), retrievedItems[i].Name)
	}

	// Verify that the recorder has the correct calls
	calls := client.Recorder.GetCalls()

	// There should be at least one call to Get and one to GetItemChildrenPath
	getCallFound := false
	getItemChildrenPathCallFound := false

	for _, call := range calls {
		if call.Method == "Get" {
			getCallFound = true
		}
		if call.Method == "GetItemChildrenPath" {
			getItemChildrenPathCallFound = true
		}
	}

	assert.True(t, getCallFound, "Should have a call to Get")
	assert.True(t, getItemChildrenPathCallFound, "Should have a call to GetItemChildrenPath")
}

// TestUT_GR_11_02_MockGraphClient_PaginationWithGetItemChildren_RetrievesAllItems tests that the mock graph client supports pagination with GetItemChildren.
//
//	Test Case ID    UT-GR-11-01
//	Title           Mock Graph Client Pagination with GetItemChildren
//	Description     Tests that the mock graph client supports pagination when using GetItemChildren
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Create a large collection of items (30 items)
//	                3. Add the items with pagination (10 items per page)
//	                4. Retrieve all items using GetItemChildren
//	                5. Verify that all items are retrieved correctly
//	                6. Verify the items are in the correct order
//	                7. Verify that the recorder has the correct calls
//	Expected Result GetItemChildren correctly handles pagination and retrieves all items
func TestUT_GR_11_02_MockGraphClient_PaginationWithGetItemChildren_RetrievesAllItems(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Create a large collection of items (>25)
	items := make([]*DriveItem, 0, 30)
	for i := 0; i < 30; i++ {
		items = append(items, &DriveItem{
			ID:   fmt.Sprintf("item-%d", i),
			Name: fmt.Sprintf("Item %d", i),
		})
	}

	// Add the items with pagination (10 items per page)
	parentID := "parent-id"
	resource := childrenPathID(parentID)
	client.AddMockItemsWithPagination(resource, items, 10)

	// Retrieve the items using GetItemChildren
	retrievedItems, err := client.GetItemChildren(parentID)
	require.NoError(t, err)

	// Verify that all items are retrieved correctly
	assert.Len(t, retrievedItems, 30, "Should retrieve all 30 items")

	// Verify the items are in the correct order
	for i := 0; i < 30; i++ {
		assert.Equal(t, fmt.Sprintf("item-%d", i), retrievedItems[i].ID)
		assert.Equal(t, fmt.Sprintf("Item %d", i), retrievedItems[i].Name)
	}

	// Verify that the recorder has the correct calls
	calls := client.Recorder.GetCalls()

	// There should be at least one call to Get and one to GetItemChildren
	getCallFound := false
	getItemChildrenCallFound := false

	for _, call := range calls {
		if call.Method == "Get" {
			getCallFound = true
		}
		if call.Method == "GetItemChildren" {
			getItemChildrenCallFound = true
		}
	}

	assert.True(t, getCallFound, "Should have a call to Get")
	assert.True(t, getItemChildrenCallFound, "Should have a call to GetItemChildren")
}

// TestUT_GR_12_02_MockGraphClient_ThreadSafety_HandlesMultipleConcurrentRequests tests that the mock graph client is thread-safe.
//
//	Test Case ID    UT-GR-12-01
//	Title           Mock Graph Client Thread Safety
//	Description     Tests that the mock graph client can handle multiple concurrent requests
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Configure responses for several items
//	                3. Start multiple goroutines to make concurrent requests
//	                4. Collect results from all goroutines
//	                5. Verify that all requests were successful
//	                6. Verify that the recorder recorded all calls
//	Expected Result The mock client correctly handles concurrent requests without race conditions
func TestUT_GR_12_02_MockGraphClient_ThreadSafety_HandlesMultipleConcurrentRequests(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Configure some responses
	client.AddMockItem("/me/drive/items/item1", &DriveItem{ID: "item1", Name: "Item 1"})
	client.AddMockItem("/me/drive/items/item2", &DriveItem{ID: "item2", Name: "Item 2"})
	client.AddMockItem("/me/drive/items/item3", &DriveItem{ID: "item3", Name: "Item 3"})

	// Number of concurrent goroutines
	numGoroutines := 10
	// Number of requests per goroutine
	numRequests := 10

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Use a mutex to protect access to the results slice
	var mu sync.Mutex
	results := make([]string, 0, numGoroutines*numRequests)

	// Start multiple goroutines to make concurrent requests
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				// Choose a random item ID
				itemID := fmt.Sprintf("item%d", (id+j)%3+1)
				resource := "/me/drive/items/" + itemID

				// Make the request
				body, err := client.Get(resource)
				if err != nil {
					t.Errorf("Error in goroutine %d, request %d: %v", id, j, err)
					continue
				}

				// Parse the response
				var item DriveItem
				err = json.Unmarshal(body, &item)
				if err != nil {
					t.Errorf("Error parsing response in goroutine %d, request %d: %v", id, j, err)
					continue
				}

				// Add the result to the results slice
				mu.Lock()
				results = append(results, item.ID)
				mu.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify that we got the expected number of results
	assert.Len(t, results, numGoroutines*numRequests)

	// Verify that the recorder recorded all calls
	recorder := client.GetRecorder()
	calls := recorder.GetCalls()
	assert.Len(t, calls, numGoroutines*numRequests)
}

// TestUT_GR_13_02_MockGraphClient_GetUser_ReturnsUserInformation tests that the mock graph client can return user information.
//
//	Test Case ID    UT-GR-13-01
//	Title           Mock Graph Client GetUser Method
//	Description     Tests that the mock graph client can return user information
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Test default behavior of GetUser
//	                3. Configure a custom user response
//	                4. Test GetUser with the custom response
//	                5. Configure an error response
//	                6. Test GetUser with the error response
//	                7. Verify that all calls were recorded
//	Expected Result The mock client correctly returns user information and handles errors
func TestUT_GR_13_02_MockGraphClient_GetUser_ReturnsUserInformation(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Test default behavior
	user, err := client.GetUser()
	assert.NoError(t, err)
	assert.Equal(t, "mock@example.com", user.UserPrincipalName)

	// Test with custom response
	customUser := User{
		UserPrincipalName: "custom@example.com",
	}
	userBytes, _ := json.Marshal(customUser)
	client.AddMockResponse("/me", userBytes, http.StatusOK, nil)

	user, err = client.GetUser()
	assert.NoError(t, err)
	assert.Equal(t, "custom@example.com", user.UserPrincipalName)

	// Test with error response
	expectedError := errors.New("user not found")
	client.AddMockResponse("/me", nil, http.StatusNotFound, expectedError)

	user, err = client.GetUser()
	assert.Equal(t, expectedError, err)
	assert.Empty(t, user.UserPrincipalName)

	// Verify that all calls were recorded
	recorder := client.GetRecorder()
	assert.True(t, recorder.VerifyCall("GetUser", 3))
}

// TestUT_GR_14_02_MockGraphClient_GetUserWithContext_HandlesContextCancellation tests that the mock graph client handles context cancellation with GetUserWithContext.
//
//	Test Case ID    UT-GR-14-01
//	Title           Mock Graph Client GetUserWithContext Method
//	Description     Tests that the mock graph client handles context cancellation with GetUserWithContext
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Test GetUserWithContext with a valid context
//	                3. Test GetUserWithContext with a cancelled context
//	                4. Verify that all calls were recorded
//	Expected Result The mock client correctly returns user information with a valid context and handles cancelled contexts
func TestUT_GR_14_02_MockGraphClient_GetUserWithContext_HandlesContextCancellation(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Test with valid context
	ctx := context.Background()
	user, err := client.GetUserWithContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "mock@example.com", user.UserPrincipalName)

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	user, err = client.GetUserWithContext(cancelledCtx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, user.UserPrincipalName)

	// Verify that all calls were recorded
	recorder := client.GetRecorder()
	assert.True(t, recorder.VerifyCall("GetUserWithContext", 2))
}

// TestUT_GR_15_02_MockGraphClient_GetDrive_ReturnsDriveInformation tests that the mock graph client can return drive information.
//
//	Test Case ID    UT-GR-15-01
//	Title           Mock Graph Client GetDrive Method
//	Description     Tests that the mock graph client can return drive information
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Test default behavior of GetDrive
//	                3. Configure a custom drive response
//	                4. Test GetDrive with the custom response
//	                5. Configure an error response
//	                6. Test GetDrive with the error response
//	                7. Verify that all calls were recorded
//	Expected Result The mock client correctly returns drive information and handles errors
func TestUT_GR_15_02_MockGraphClient_GetDrive_ReturnsDriveInformation(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Test default behavior
	drive, err := client.GetDrive()
	assert.NoError(t, err)
	assert.Equal(t, "mock-drive-id", drive.ID)
	assert.Equal(t, DriveTypePersonal, drive.DriveType)
	assert.Equal(t, uint64(1024*1024*1024*10), drive.Quota.Total) // 10 GB
	assert.Equal(t, uint64(1024*1024*1024*2), drive.Quota.Used)   // 2 GB
	assert.Equal(t, "normal", drive.Quota.State)

	// Test with custom response
	customDrive := Drive{
		ID:        "custom-drive-id",
		DriveType: "business",
		Quota: DriveQuota{
			Total:     1024 * 1024 * 1024 * 100, // 100 GB
			Used:      1024 * 1024 * 1024 * 50,  // 50 GB
			Remaining: 1024 * 1024 * 1024 * 50,  // 50 GB
			State:     "nearing",
		},
	}
	driveBytes, _ := json.Marshal(customDrive)
	client.AddMockResponse("/me/drive", driveBytes, http.StatusOK, nil)

	drive, err = client.GetDrive()
	assert.NoError(t, err)
	assert.Equal(t, "custom-drive-id", drive.ID)
	assert.Equal(t, "business", drive.DriveType)
	assert.Equal(t, uint64(1024*1024*1024*100), drive.Quota.Total) // 100 GB
	assert.Equal(t, uint64(1024*1024*1024*50), drive.Quota.Used)   // 50 GB
	assert.Equal(t, "nearing", drive.Quota.State)

	// Test with error response
	expectedError := errors.New("drive not found")
	client.AddMockResponse("/me/drive", nil, http.StatusNotFound, expectedError)

	drive, err = client.GetDrive()
	assert.Equal(t, expectedError, err)
	assert.Empty(t, drive.ID)

	// Verify that all calls were recorded
	recorder := client.GetRecorder()
	assert.True(t, recorder.VerifyCall("GetDrive", 3))
}

// TestUT_GR_16_02_MockGraphClient_GetItemChild_ReturnsChildItem tests that the mock graph client can return child items.
//
//	Test Case ID    UT-GR-16-01
//	Title           Mock Graph Client GetItemChild Method
//	Description     Tests that the mock graph client can return child items
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Test default behavior of GetItemChild
//	                3. Configure a custom child item response
//	                4. Test GetItemChild with the custom response
//	                5. Configure an error response
//	                6. Test GetItemChild with the error response
//	                7. Verify that all calls were recorded
//	Expected Result The mock client correctly returns child items and handles errors
func TestUT_GR_16_02_MockGraphClient_GetItemChild_ReturnsChildItem(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Test default behavior
	item, err := client.GetItemChild("parent-id", "child-name")
	assert.NoError(t, err)
	assert.Equal(t, "mock-child-id", item.ID)
	assert.Equal(t, "child-name", item.Name)

	// Test with custom response
	customItem := DriveItem{
		ID:   "custom-child-id",
		Name: "custom-child-name",
		Size: 1024,
	}
	itemBytes, _ := json.Marshal(customItem)
	resource := fmt.Sprintf("%s:/%s", IDPath("parent-id"), url.PathEscape("custom-child"))
	client.AddMockResponse(resource, itemBytes, http.StatusOK, nil)

	item, err = client.GetItemChild("parent-id", "custom-child")
	assert.NoError(t, err)
	assert.Equal(t, "custom-child-id", item.ID)
	assert.Equal(t, "custom-child-name", item.Name)
	assert.Equal(t, uint64(1024), item.Size)

	// Test with error response
	expectedError := errors.New("child not found")
	resource = fmt.Sprintf("%s:/%s", IDPath("parent-id"), url.PathEscape("missing-child"))
	client.AddMockResponse(resource, nil, http.StatusNotFound, expectedError)

	item, err = client.GetItemChild("parent-id", "missing-child")
	assert.Equal(t, expectedError, err)
	assert.Nil(t, item)

	// Verify that all calls were recorded
	recorder := client.GetRecorder()
	assert.True(t, recorder.VerifyCall("GetItemChild", 3))
}

// TestUT_GR_17_01_MockGraphClient_MethodCallRecording_TracksMethodCallsAndArguments tests that the mock graph client records method calls and their arguments.
//
//	Test Case ID    UT-GR-17-01
//	Title           Mock Graph Client Method Call Recording
//	Description     Tests that the mock graph client records method calls and their arguments
//	Preconditions   None
//	Steps           1. Create a new MockGraphClient
//	                2. Add mock items for testing
//	                3. Perform several operations (GetItem, GetItemChildren, GetItemPath, Mkdir, Rename, Remove)
//	                4. Retrieve the recorder and verify the expected methods were called
//	                5. Verify the total number of calls
//	                6. Check the number of calls for each method
//	                7. Verify the order of calls
//	                8. Verify the arguments for some key calls
//	Expected Result The mock client correctly records all method calls, their order, and their arguments
func TestUT_GR_17_01_MockGraphClient_MethodCallRecording_TracksMethodCallsAndArguments(t *testing.T) {
	// Create a new MockGraphClient
	client := NewMockGraphClient()

	// Add some mock items for testing
	rootItem := &DriveItem{
		ID:   "root-id",
		Name: "root",
	}
	client.AddMockItem("/me/drive/items/root-id", rootItem)

	childItem1 := &DriveItem{
		ID:   "child-id-1",
		Name: "child1",
	}
	childItem2 := &DriveItem{
		ID:   "child-id-2",
		Name: "child2",
	}
	children := []*DriveItem{childItem1, childItem2}
	client.AddMockItems("/me/drive/items/root-id/children", children)

	// Perform several operations
	// 1. Get the root item
	rootResult, err := client.GetItem("root-id")
	assert.NoError(t, err)
	assert.Equal(t, rootItem.ID, rootResult.ID)

	// 2. Get the children of the root item
	childrenResult, err := client.GetItemChildren("root-id")
	assert.NoError(t, err)
	assert.Len(t, childrenResult, 2)

	// 3. Get a specific child by path
	childPathResult, err := client.GetItemPath("/root/child1")
	assert.NoError(t, err)
	assert.NotNil(t, childPathResult)

	// 4. Create a new folder
	newFolder, err := client.Mkdir("new-folder", "root-id")
	assert.NoError(t, err)
	assert.Equal(t, "new-folder", newFolder.Name)

	// 5. Rename an item
	err = client.Rename("child-id-1", "renamed-child", "root-id")
	assert.NoError(t, err)

	// 6. Remove an item
	err = client.Remove("child-id-2")
	assert.NoError(t, err)

	// Retrieve the recorder and verify the expected methods were called
	recorder := client.GetRecorder()
	calls := recorder.GetCalls()

	// Verify the total number of calls
	// Note: Each high-level method may make multiple low-level calls
	// We're focusing on the high-level methods here
	assert.True(t, len(calls) >= 6, "Expected at least 6 calls, got %d", len(calls))

	// Check the number of calls for each method matches expectations
	assert.True(t, recorder.VerifyCall("GetItem", 1), "Expected 1 call to GetItem")
	assert.True(t, recorder.VerifyCall("GetItemChildren", 1), "Expected 1 call to GetItemChildren")
	assert.True(t, recorder.VerifyCall("GetItemPath", 1), "Expected 1 call to GetItemPath")
	assert.True(t, recorder.VerifyCall("Mkdir", 1), "Expected 1 call to Mkdir")
	assert.True(t, recorder.VerifyCall("Rename", 1), "Expected 1 call to Rename")
	assert.True(t, recorder.VerifyCall("Remove", 1), "Expected 1 call to Remove")

	// Verify the order of calls
	methodCalls := []string{}
	for _, call := range calls {
		if call.Method == "GetItem" ||
			call.Method == "GetItemChildren" ||
			call.Method == "GetItemPath" ||
			call.Method == "Mkdir" ||
			call.Method == "Rename" ||
			call.Method == "Remove" {
			methodCalls = append(methodCalls, call.Method)
		}
	}

	// Check that the high-level methods were called in the expected order
	expectedMethods := []string{"GetItem", "GetItemChildren", "GetItemPath", "Mkdir", "Rename", "Remove"}
	for i, method := range expectedMethods {
		assert.Contains(t, methodCalls, method, "Method %s should be in the call list", method)
		if i > 0 {
			// Find the index of the current and previous method
			currentIndex := -1
			previousIndex := -1
			for j, call := range methodCalls {
				if call == method {
					currentIndex = j
				}
				if call == expectedMethods[i-1] {
					previousIndex = j
				}
			}
			// Verify the order if both methods were found
			if currentIndex != -1 && previousIndex != -1 {
				assert.Greater(t, currentIndex, previousIndex,
					"Method %s should be called after %s", method, expectedMethods[i-1])
			}
		}
	}

	// Verify the arguments for some key calls
	for _, call := range calls {
		if call.Method == "GetItem" {
			assert.Equal(t, "root-id", call.Args[0], "GetItem should be called with root-id")
		} else if call.Method == "GetItemChildren" {
			assert.Equal(t, "root-id", call.Args[0], "GetItemChildren should be called with root-id")
		} else if call.Method == "Mkdir" {
			assert.Equal(t, "new-folder", call.Args[0], "Mkdir should be called with new-folder")
			assert.Equal(t, "root-id", call.Args[1], "Mkdir should be called with root-id as parent")
		}
	}
}
