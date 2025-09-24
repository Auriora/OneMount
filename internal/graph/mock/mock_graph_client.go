// Package mock provides mock implementations for testing.
package mock

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/graph/api"
	"github.com/auriora/onemount/internal/logging"
)

// MockGraphClient is a mock implementation for testing Graph API interactions.
type MockGraphClient struct {
	// Recorder for method calls
	recorder api.MockRecorder

	// Mutex for thread safety
	mu sync.Mutex

	// Mock responses for specific resources
	responses map[string]mockResponse

	// Mock items for specific resources
	items map[string]*api.DriveItem

	// Mock item collections for specific resources
	itemCollections map[string][]*api.DriveItem

	// Network conditions
	networkConditions struct {
		latency    time.Duration
		packetLoss float64
		bandwidth  int
	}

	// Configuration
	config struct {
		errorRate      float64
		responseDelay  time.Duration
		throttleRate   float64
		throttleDelay  time.Duration
		customBehavior map[string]interface{}
	}
}

// mockResponse represents a mock response for a specific resource.
type mockResponse struct {
	body       []byte
	statusCode int
	err        error
}

// NewMockGraphClient creates a new MockGraphClient with default values
func NewMockGraphClient() *MockGraphClient {
	mock := &MockGraphClient{
		recorder:        newBasicMockRecorder(),
		responses:       make(map[string]mockResponse),
		items:           make(map[string]*api.DriveItem),
		itemCollections: make(map[string][]*api.DriveItem),
	}
	mock.config.customBehavior = make(map[string]interface{})
	logging.Debug().Msg("Setting up MockGraphClient as the test HTTP client")
	return mock
}

// newBasicMockRecorder creates a new basic mock recorder.
func newBasicMockRecorder() api.MockRecorder {
	return &basicMockRecorder{
		calls: make([]api.MockCall, 0),
	}
}

// basicMockRecorder is a simple implementation of the api.MockRecorder interface.
type basicMockRecorder struct {
	mu    sync.Mutex
	calls []api.MockCall
}

// RecordCall records a method call.
func (r *basicMockRecorder) RecordCall(method string, args ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls = append(r.calls, api.MockCall{
		Method: method,
		Args:   args,
		Time:   time.Now(),
	})
}

// RecordCallWithResult records a method call with a result and error.
func (r *basicMockRecorder) RecordCallWithResult(method string, result interface{}, err error, args ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls = append(r.calls, api.MockCall{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
		Time:   time.Now(),
	})
}

// GetCalls returns all recorded calls.
func (r *basicMockRecorder) GetCalls() []api.MockCall {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Return a copy to avoid race conditions
	calls := make([]api.MockCall, len(r.calls))
	copy(calls, r.calls)
	return calls
}

// VerifyCall verifies a method was called a specific number of times.
func (r *basicMockRecorder) VerifyCall(method string, times int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for _, call := range r.calls {
		if call.Method == method {
			count++
		}
	}
	return count == times
}

// SetNetworkConditions sets the simulated network conditions
func (m *MockGraphClient) SetNetworkConditions(latency time.Duration, packetLoss float64, bandwidth int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.networkConditions.latency = latency
	m.networkConditions.packetLoss = packetLoss
	m.networkConditions.bandwidth = bandwidth
}

// SetConfig sets the mock configuration
func (m *MockGraphClient) SetConfig(config struct {
	Latency        time.Duration
	ErrorRate      float64
	ResponseDelay  time.Duration
	ThrottleRate   float64
	ThrottleDelay  time.Duration
	CustomBehavior map[string]interface{}
}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.errorRate = config.ErrorRate
	m.config.responseDelay = config.ResponseDelay
	m.config.throttleRate = config.ThrottleRate
	m.config.throttleDelay = config.ThrottleDelay
	m.config.customBehavior = config.CustomBehavior
}

// GetRecorder returns the mock recorder
func (m *MockGraphClient) GetRecorder() api.MockRecorder {
	return m.recorder
}

// Cleanup cleans up the mock client
func (m *MockGraphClient) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.responses = make(map[string]mockResponse)
	m.items = make(map[string]*api.DriveItem)
	m.itemCollections = make(map[string][]*api.DriveItem)
	logging.Debug().Msg("Cleaning up MockGraphClient, resetting HTTP client to default")
}

// AddMockResponse adds a predefined response for a specific resource path
func (m *MockGraphClient) AddMockResponse(resource string, body []byte, statusCode int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("AddMockResponse", resource, body, statusCode, err)
	m.responses[resource] = mockResponse{
		body:       body,
		statusCode: statusCode,
		err:        err,
	}
}

// AddMockItem adds a predefined DriveItem response for a specific resource path
func (m *MockGraphClient) AddMockItem(resource string, item *api.DriveItem) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("AddMockItem", resource, item)
	m.items[resource] = item
}

// AddMockItems adds a predefined collection of DriveItems for a specific resource path
func (m *MockGraphClient) AddMockItems(resource string, items []*api.DriveItem) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("AddMockItems", resource, items)
	m.itemCollections[resource] = items
}

// RequestWithContext performs a request to the Microsoft Graph API with context
func (m *MockGraphClient) RequestWithContext(ctx context.Context, resource string, method string, content io.Reader, headers ...api.Header) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("RequestWithContext", resource, method, content, headers)

	// Check if we have a mock response for this resource
	if resp, ok := m.responses[resource]; ok {
		return resp.body, resp.err
	}

	// Default response
	return nil, nil
}

// Get performs a GET request to the Microsoft Graph API
func (m *MockGraphClient) Get(resource string, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("Get", resource, headers)
	return m.RequestWithContext(context.Background(), resource, "GET", nil, headers...)
}

// GetWithContext performs a GET request to the Microsoft Graph API with context
func (m *MockGraphClient) GetWithContext(ctx context.Context, resource string, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("GetWithContext", ctx, resource, headers)
	return m.RequestWithContext(ctx, resource, "GET", nil, headers...)
}

// GetItem fetches a DriveItem by ID
func (m *MockGraphClient) GetItem(id string) (*api.DriveItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("GetItem", id)

	// Check if we have a mock item for this ID
	if item, ok := m.items[id]; ok {
		return item, nil
	}

	// Default response
	return nil, nil
}

// GetItemPath fetches a DriveItem by path
func (m *MockGraphClient) GetItemPath(path string) (*api.DriveItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("GetItemPath", path)

	// Check if we have a mock response (including error responses) for this path
	if resp, ok := m.responses[path]; ok {
		if resp.err != nil {
			return nil, resp.err
		}
		// If there's a response but no error, try to unmarshal it as a DriveItem
		// For now, just return the error if any
	}

	// Check if we have a mock item for this path
	if item, ok := m.items[path]; ok {
		return item, nil
	}

	// Default response
	return nil, nil
}

// GetItemChildren fetches the children of a DriveItem by ID
func (m *MockGraphClient) GetItemChildren(id string) ([]*api.DriveItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("GetItemChildren", id)

	// Check if we have mock items for this ID
	if items, ok := m.itemCollections[id]; ok {
		return items, nil
	}

	// Default response
	return nil, nil
}

// GetItemChildrenPath fetches the children of a DriveItem by path
func (m *MockGraphClient) GetItemChildrenPath(path string) ([]*api.DriveItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("GetItemChildrenPath", path)

	// Check if we have mock items for this path
	if items, ok := m.itemCollections[path]; ok {
		return items, nil
	}

	// Default response
	return nil, nil
}

// GetItemContent retrieves an item's content from the Graph endpoint
func (m *MockGraphClient) GetItemContent(id string) ([]byte, uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("GetItemContent", id)

	// Check if we have a mock response for this ID
	if resp, ok := m.responses[id]; ok {
		return resp.body, uint64(len(resp.body)), resp.err
	}

	// Default response
	return nil, 0, nil
}

// GetItemContentStream retrieves an item's content and writes it to the provided writer
func (m *MockGraphClient) GetItemContentStream(id string, output io.Writer) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("GetItemContentStream", id, output)

	// Check if we have a mock response for this ID
	if resp, ok := m.responses[id]; ok {
		if resp.err != nil {
			return 0, resp.err
		}
		n, err := output.Write(resp.body)
		return uint64(n), err
	}

	// Default response
	return 0, nil
}

// Patch performs a PATCH request to the Microsoft Graph API
func (m *MockGraphClient) Patch(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("Patch", resource, content, headers)
	return m.RequestWithContext(context.Background(), resource, "PATCH", content, headers...)
}

// Post performs a POST request to the Microsoft Graph API
func (m *MockGraphClient) Post(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("Post", resource, content, headers)
	return m.RequestWithContext(context.Background(), resource, "POST", content, headers...)
}

// Put performs a PUT request to the Microsoft Graph API
func (m *MockGraphClient) Put(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("Put", resource, content, headers)
	return m.RequestWithContext(context.Background(), resource, "PUT", content, headers...)
}

// Delete performs a DELETE request to the Microsoft Graph API
func (m *MockGraphClient) Delete(resource string, headers ...api.Header) error {
	m.recorder.RecordCall("Delete", resource, headers)
	_, err := m.RequestWithContext(context.Background(), resource, "DELETE", nil, headers...)
	return err
}

// Mkdir creates a new directory
func (m *MockGraphClient) Mkdir(name string, parentID string) (*api.DriveItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("Mkdir", name, parentID)

	// Create a new mock item for the directory
	item := &api.DriveItem{
		ID:   "mock-dir-" + name,
		Name: name,
		Folder: &api.Folder{
			ChildCount: 0,
		},
	}

	// Add the item to our mock items
	m.items[item.ID] = item

	return item, nil
}

// Rename renames an item
func (m *MockGraphClient) Rename(itemID string, itemName string, parentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("Rename", itemID, itemName, parentID)

	// Check if we have a mock item for this ID
	if item, ok := m.items[itemID]; ok {
		// Update the item's name
		item.Name = itemName
		return nil
	}

	// Default response
	return nil
}

// Remove removes an item
func (m *MockGraphClient) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.recorder.RecordCall("Remove", id)

	// Check if we have a mock item for this ID
	if _, ok := m.items[id]; ok {
		// Remove the item
		delete(m.items, id)
		return nil
	}

	// Default response
	return nil
}
