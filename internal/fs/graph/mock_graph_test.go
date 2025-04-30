package graph

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockGraphClient_ConfigurableResponses(t *testing.T) {
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

func TestMockGraphClient_RecordCalls(t *testing.T) {
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

func TestMockGraphClient_NetworkConditions(t *testing.T) {
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

func TestMockGraphClient_CustomBehavior(t *testing.T) {
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

func TestMockGraphClient_GetItemContentStream(t *testing.T) {
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

func TestMockGraphClient_DriveItemOperations(t *testing.T) {
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

func TestMockGraphClient_ContextCancellation(t *testing.T) {
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
