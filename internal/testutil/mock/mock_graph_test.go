package mock

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT_MG_01_01_MockGraphProvider_ApiResponses_ConfiguresAndUsesResponses tests that the mock graph provider can be configured with responses and used for operations.
//
//	Test Case ID    UT-MG-01-01
//	Title           Mock Graph Provider API Responses
//	Description     Tests that the mock graph provider can be configured with responses and used for operations
//	Preconditions   None
//	Steps           1. Create a new MockGraphProvider
//	                2. Setup mock responses for item retrieval
//	                3. Setup mock responses for content download
//	                4. Setup mock responses for children listing
//	                5. Perform operations using the mock responses
//	                6. Verify that all operations were recorded
//	Expected Result The mock provider correctly returns configured responses and records operations
func TestUT_MG_01_01_MockGraphProvider_ApiResponses_ConfiguresAndUsesResponses(t *testing.T) {
	// Create a new MockGraphProvider
	provider := NewMockGraphProvider()

	// 1. Setup mock responses for item retrieval
	// Create a mock drive item
	mockItem := &graph.DriveItem{
		ID:   "test-item-id",
		Name: "test-item",
		Size: 1024,
		File: &graph.File{
			Hashes: graph.Hashes{
				SHA1Hash:     "test-sha1",
				QuickXorHash: "test-quickxor",
			},
		},
	}

	// Add the mock item response
	itemPath := "/me/drive/items/test-item-id"
	provider.AddMockItem(itemPath, mockItem)

	// 2. Setup mock responses for content download
	// Create mock content
	mockContent := []byte("This is the content of the test file")
	contentPath := "/me/drive/items/test-item-id/content"
	provider.AddMockResponse(contentPath, mockContent, 200, nil)

	// 3. Setup mock responses for children listing
	// Create mock children items
	mockChildren := []*graph.DriveItem{
		{
			ID:   "child-1",
			Name: "child-1",
			File: &graph.File{
				Hashes: graph.Hashes{
					SHA1Hash:     "child-sha1",
					QuickXorHash: "child-quickxor",
				},
			},
		},
		{
			ID:   "child-2",
			Name: "child-2",
			Folder: &graph.Folder{
				ChildCount: 5,
			},
		},
	}

	// Add the mock children response
	childrenPath := "/me/drive/items/test-item-id/children"
	provider.AddMockItems(childrenPath, mockChildren)

	// 4. Perform operations using the mock responses

	// 4.1 Get the item
	item, err := provider.GetItem("test-item-id")
	require.NoError(t, err)
	assert.Equal(t, "test-item-id", item.ID)
	assert.Equal(t, "test-item", item.Name)
	assert.Equal(t, uint64(1024), item.Size)
	assert.Equal(t, "test-sha1", item.File.Hashes.SHA1Hash)
	assert.Equal(t, "test-quickxor", item.File.Hashes.QuickXorHash)

	// 4.2 Get the item content
	content, size, err := provider.GetItemContent("test-item-id")
	require.NoError(t, err)
	assert.Equal(t, uint64(len(mockContent)), size)
	assert.Equal(t, mockContent, content)

	// 4.3 Get the item content using stream
	var buf bytes.Buffer
	size, err = provider.GetItemContentStream("test-item-id", &buf)
	require.NoError(t, err)
	assert.Equal(t, uint64(len(mockContent)), size)
	assert.Equal(t, mockContent, buf.Bytes())

	// 4.4 Get the item children
	children, err := provider.GetItemChildren("test-item-id")
	require.NoError(t, err)
	assert.Len(t, children, 2)
	assert.Equal(t, "child-1", children[0].ID)
	assert.Equal(t, "child-2", children[1].ID)
	assert.NotNil(t, children[0].File)
	assert.NotNil(t, children[1].Folder)
	assert.Equal(t, uint32(5), children[1].Folder.ChildCount)

	// 5. Verify that all operations were recorded
	recorder := provider.GetRecorder()
	calls := recorder.GetCalls()

	// Log all recorded calls for debugging
	t.Logf("Recorded calls: %d", len(calls))
	for i, call := range calls {
		t.Logf("Call %d: %s", i, call.Method)
	}

	// Count the occurrences of each method
	getItemCount := 0
	getItemContentCount := 0
	getItemContentStreamCount := 0
	getItemChildrenCount := 0

	for _, call := range calls {
		switch call.Method {
		case "GetItem":
			getItemCount++
		case "GetItemContent":
			getItemContentCount++
		case "GetItemContentStream":
			getItemContentStreamCount++
		case "GetItemChildren":
			getItemChildrenCount++
		}
	}

	// Verify that each method was called at least once
	assert.GreaterOrEqual(t, getItemCount, 1, "Expected at least 1 call to GetItem")
	assert.GreaterOrEqual(t, getItemContentCount, 1, "Expected at least 1 call to GetItemContent")
	assert.GreaterOrEqual(t, getItemContentStreamCount, 1, "Expected at least 1 call to GetItemContentStream")
	assert.GreaterOrEqual(t, getItemChildrenCount, 1, "Expected at least 1 call to GetItemChildren")
}

// TestUT_MG_02_01_MockGraphProvider_NetworkErrors_SimulatesNetworkConditions tests that the mock graph provider can simulate network errors.
//
//	Test Case ID    UT-MG-02-01
//	Title           Mock Graph Provider Network Errors
//	Description     Tests that the mock graph provider can simulate network errors and that operations handle these errors correctly
//	Preconditions   None
//	Steps           1. Create a new MockGraphProvider
//	                2. Setup mock responses for various operations
//	                3. Test network latency by configuring high latency and verifying request timing
//	                4. Test random errors by configuring a high error rate and verifying failures
//	                5. Test API throttling by configuring throttling and verifying throttled responses
//	                6. Test error handling with retries by implementing a retry mechanism
//	                7. Test combined error conditions with a mix of latency, errors, and throttling
//	Expected Result The mock provider correctly simulates various network conditions and error scenarios
func TestUT_MG_02_01_MockGraphProvider_NetworkErrors_SimulatesNetworkConditions(t *testing.T) {
	// Create a new MockGraphProvider
	provider := NewMockGraphProvider()

	// Setup mock responses for item retrieval
	mockItem := &graph.DriveItem{
		ID:   "test-item-id",
		Name: "test-item",
		Size: 1024,
	}
	itemPath := "/me/drive/items/test-item-id"
	provider.AddMockItem(itemPath, mockItem)

	// Setup mock responses for content download
	mockContent := []byte("This is the content of the test file")
	contentPath := "/me/drive/items/test-item-id/content"
	provider.AddMockResponse(contentPath, mockContent, 200, nil)

	// Setup mock responses for children listing
	mockChildren := []*graph.DriveItem{
		{
			ID:   "child-1",
			Name: "child-1",
		},
		{
			ID:   "child-2",
			Name: "child-2",
		},
	}
	childrenPath := "/me/drive/items/test-item-id/children"
	provider.AddMockItems(childrenPath, mockChildren)

	// Test 1: Configure network latency
	t.Run("NetworkLatency", func(t *testing.T) {
		// Reset the provider for this test
		provider.Reset()
		provider.AddMockItem(itemPath, mockItem)

		// Configure network conditions with high latency
		provider.SetNetworkConditions(100*time.Millisecond, 0, 0)

		// Measure the time it takes to make a request
		start := time.Now()
		item, err := provider.GetItem("test-item-id")
		duration := time.Since(start)

		// Verify that the request succeeded but took at least the configured latency
		require.NoError(t, err)
		assert.Equal(t, "test-item-id", item.ID)
		assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
	})

	// Test 2: Configure random errors
	t.Run("RandomErrors", func(t *testing.T) {
		// Reset the provider for this test
		provider.Reset()
		provider.AddMockItem(itemPath, mockItem)

		// Configure a high error rate (100% for deterministic testing)
		provider.SetConfig(MockConfig{
			ErrorRate: 1.0, // 100% error rate
		})

		// Attempt to get the item, which should fail due to the error rate
		_, err := provider.GetItem("test-item-id")

		// Verify that the request failed with the expected error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "EOF")
	})

	// Test 3: Configure API throttling
	t.Run("APIThrottling", func(t *testing.T) {
		// Reset the provider for this test
		provider.Reset()
		provider.AddMockItem(itemPath, mockItem)

		// Configure API throttling (100% for deterministic testing)
		provider.SetConfig(MockConfig{
			ThrottleRate:  1.0, // 100% throttle rate
			ThrottleDelay: 50 * time.Millisecond,
		})

		// Measure the time it takes to make a request
		start := time.Now()
		_, err := provider.GetItem("test-item-id")
		duration := time.Since(start)

		// Verify that the request failed due to throttling
		require.Error(t, err)
		assert.Contains(t, err.Error(), "throttling")

		// Verify that the request took at least the configured throttle delay
		assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
	})

	// Test 4: Test error handling with retries
	t.Run("ErrorHandlingWithRetries", func(t *testing.T) {
		// Reset the provider for this test
		provider.Reset()
		provider.AddMockItem(itemPath, mockItem)

		// Configure a moderate error rate (50% for testing retries)
		provider.SetConfig(MockConfig{
			ErrorRate: 0.5, // 50% error rate
		})

		// Create a function that retries getting an item up to 5 times
		getItemWithRetries := func(id string, maxRetries int) (*graph.DriveItem, error) {
			var lastErr error
			for i := 0; i < maxRetries; i++ {
				item, err := provider.GetItem(id)
				if err == nil {
					return item, nil
				}
				lastErr = err
				// In a real implementation, you might add a backoff delay here
			}
			return nil, lastErr
		}

		// Attempt to get the item with retries
		item, err := getItemWithRetries("test-item-id", 10)

		// With a 50% error rate and 10 retries, we should eventually succeed
		require.NoError(t, err)
		assert.Equal(t, "test-item-id", item.ID)

		// Verify that multiple GetItem calls were made due to retries
		recorder := provider.GetRecorder()
		calls := recorder.GetCalls()
		getItemCalls := 0
		for _, call := range calls {
			if call.Method == "GetItem" {
				getItemCalls++
			}
		}
		assert.Greater(t, getItemCalls, 1, "Expected multiple GetItem calls due to retries")
	})

	// Test 5: Test combined error conditions
	t.Run("CombinedErrorConditions", func(t *testing.T) {
		// Reset the provider for this test
		provider.Reset()
		provider.AddMockItem(itemPath, mockItem)
		provider.AddMockResponse(contentPath, mockContent, 200, nil)

		// Configure both latency and error rate
		provider.SetNetworkConditions(50*time.Millisecond, 0, 0)

		// Set very high error and throttle rates to ensure we get failures
		provider.SetConfig(MockConfig{
			ErrorRate:    0.7, // 70% error rate
			ThrottleRate: 0.7, // 70% throttle rate
		})

		// Perform multiple operations to test different error scenarios
		successCount := 0
		errorCount := 0
		throttleCount := 0
		totalOperations := 20

		for i := 0; i < totalOperations; i++ {
			// Use a different approach to get a mix of successes and failures
			// Instead of resetting the client, we'll directly manipulate the config
			// to force some operations to succeed and others to fail
			if i%3 == 0 {
				// Every third operation should succeed
				provider.SetConfig(MockConfig{
					ErrorRate:    0.0,
					ThrottleRate: 0.0,
				})
			} else if i%3 == 1 {
				// Every third+1 operation should fail with an error
				provider.SetConfig(MockConfig{
					ErrorRate:    1.0,
					ThrottleRate: 0.0,
				})
			} else {
				// Every third+2 operation should fail with throttling
				provider.SetConfig(MockConfig{
					ErrorRate:    0.0,
					ThrottleRate: 1.0,
				})
			}

			_, err := provider.GetItem("test-item-id")
			if err == nil {
				successCount++
			} else if err != nil && strings.Contains(err.Error(), "throttling") {
				throttleCount++
			} else if err != nil {
				errorCount++
			}
		}

		// Verify that we got a mix of successes and failures
		t.Logf("Success: %d, Errors: %d, Throttled: %d", successCount, errorCount, throttleCount)

		// We should have at least some successes and some failures (either errors or throttling)
		assert.Greater(t, successCount, 0, "Expected some successful operations")
		assert.Greater(t, errorCount+throttleCount, 0, "Expected some failed operations")
	})
}
