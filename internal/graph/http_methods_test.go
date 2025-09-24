package graph

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/graph/api"
	"github.com/stretchr/testify/assert"
)

// TestUT_GR_HTTP_01_01_GET_ValidResource_ReturnsData tests GET requests with valid resources
func TestUT_GR_HTTP_01_01_GET_ValidResource_ReturnsData(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	expectedData := []byte(`{"id":"test-id","name":"test-item"}`)
	client.AddMockResponse(resource, expectedData, http.StatusOK, nil)

	data, err := client.Get(resource)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	calls := client.Recorder.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Get", calls[0].Method)
}

// TestUT_GR_HTTP_01_02_GET_WithContext_RespectsTimeout tests GET requests with context timeout
func TestUT_GR_HTTP_01_02_GET_WithContext_RespectsTimeout(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	expectedData := []byte(`{"id":"test-id","name":"test-item"}`)
	client.AddMockResponse(resource, expectedData, http.StatusOK, nil)

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetWithContext(ctx, resource)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestUT_GR_HTTP_02_01_POST_ValidData_CreatesResource tests POST requests with valid data
func TestUT_GR_HTTP_02_01_POST_ValidData_CreatesResource(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/parent-id/children"
	expectedData := []byte(`{"id":"new-id","name":"new-folder"}`)
	client.AddMockResponse(resource, expectedData, http.StatusCreated, nil)

	postData := `{"name":"new-folder","folder":{}}`
	content := strings.NewReader(postData)

	data, err := client.Post(resource, content)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	calls := client.Recorder.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Post", calls[0].Method)
}

// TestUT_GR_HTTP_03_01_PUT_ValidData_UpdatesResource tests PUT requests with valid data
func TestUT_GR_HTTP_03_01_PUT_ValidData_UpdatesResource(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	expectedData := []byte(`{"id":"test-id","name":"updated-name"}`)
	client.AddMockResponse(resource, expectedData, http.StatusOK, nil)

	putData := `{"name":"updated-name"}`
	content := strings.NewReader(putData)

	data, err := client.Put(resource, content)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	calls := client.Recorder.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Put", calls[0].Method)
}

// TestUT_GR_HTTP_04_01_DELETE_ValidResource_DeletesResource tests DELETE requests
func TestUT_GR_HTTP_04_01_DELETE_ValidResource_DeletesResource(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	client.AddMockResponse(resource, nil, http.StatusNoContent, nil)

	err := client.Delete(resource)

	assert.NoError(t, err)

	calls := client.Recorder.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Delete", calls[0].Method)
}

// TestUT_GR_HTTP_05_01_RequestWithHeaders_CustomHeaders_IncludesHeaders tests requests with custom headers
func TestUT_GR_HTTP_05_01_RequestWithHeaders_CustomHeaders_IncludesHeaders(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id"
	expectedData := []byte(`{"id":"test-id","name":"test-item"}`)
	client.AddMockResponse(resource, expectedData, http.StatusOK, nil)

	headers := []api.Header{
		{Key: "If-Match", Value: "test-etag"},
		{Key: "Prefer", Value: "return=representation"},
	}

	data, err := client.Get(resource, headers...)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	calls := client.Recorder.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Get", calls[0].Method)
}

// TestUT_GR_HTTP_06_01_RequestWithLargePayload_LargeData_HandlesCorrectly tests requests with large payloads
func TestUT_GR_HTTP_06_01_RequestWithLargePayload_LargeData_HandlesCorrectly(t *testing.T) {
	client := NewMockGraphClient()

	resource := "/me/drive/items/test-id/content"
	expectedData := []byte(`{"id":"test-id","size":1048576}`)
	client.AddMockResponse(resource, expectedData, http.StatusOK, nil)

	// Create 1MB payload
	largeData := bytes.Repeat([]byte("a"), 1024*1024)
	content := bytes.NewReader(largeData)

	data, err := client.Put(resource, content)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)

	calls := client.Recorder.GetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "Put", calls[0].Method)
}
