// Package graph provides mocks for testing without actual API calls.
// This package contains a mock implementation of the Microsoft Graph API client
// that can be used in tests to simulate API behavior without making actual network
// requests. The mock implementation supports various features to make testing more
// realistic and comprehensive:
//
// - Simulating network conditions like latency, packet loss, and bandwidth limitations
// - Simulating error conditions like random errors and API throttling
// - Recording and verifying method calls
// - Pagination support for large collections
// - Thread-safety for concurrent tests
//
// The mock implementation is designed to be used both directly in unit tests and
// through the higher-level mock in the testutil package for integration tests.
package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MockCall represents a record of a method call on a mock
type MockCall struct {
	Method    string
	Resource  string
	Args      []interface{}
	Result    interface{}
	Timestamp time.Time
}

// NetworkConditions simulates different network scenarios
type NetworkConditions struct {
	Latency    time.Duration // Simulated network latency
	PacketLoss float64       // Probability of packet loss (0.0-1.0)
	Bandwidth  int           // Simulated bandwidth in KB/s (0 = unlimited)
}

// MockRecorder records and verifies mock interactions
type MockRecorder interface {
	RecordCall(method string, args ...interface{})
	RecordCallWithResult(method string, result interface{}, args ...interface{})
	GetCalls() []MockCall
	VerifyCall(method string, times int) bool
	Clear()
}

// DefaultMockRecorder is a basic implementation of MockRecorder
type DefaultMockRecorder struct {
	calls []MockCall
	mu    sync.Mutex
}

// RecordCall records a method call
func (r *DefaultMockRecorder) RecordCall(method string, args ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, MockCall{
		Method:    method,
		Args:      args,
		Timestamp: time.Now(),
	})
}

// RecordCallWithResult records a method call with a specific result
func (r *DefaultMockRecorder) RecordCallWithResult(method string, result interface{}, args ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, MockCall{
		Method:    method,
		Args:      args,
		Result:    result,
		Timestamp: time.Now(),
	})
}

// GetCalls returns all recorded calls
func (r *DefaultMockRecorder) GetCalls() []MockCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]MockCall, len(r.calls))
	copy(result, r.calls)
	return result
}

// VerifyCall checks if a method was called a specific number of times
func (r *DefaultMockRecorder) VerifyCall(method string, times int) bool {
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

// Clear clears all recorded calls
func (r *DefaultMockRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = []MockCall{}
}

// MockConfig defines configuration for mock behavior
type MockConfig struct {
	Latency        time.Duration          // Default latency for all requests
	ErrorRate      float64                // Probability of random errors (0.0-1.0)
	ResponseDelay  time.Duration          // Additional delay before responding
	ThrottleRate   float64                // Probability of throttling (0.0-1.0)
	ThrottleDelay  time.Duration          // Delay to simulate when throttled
	CustomBehavior map[string]interface{} // Custom behavior configuration
}

// MockResponse represents a predefined response for a specific request
type MockResponse struct {
	Body       []byte
	StatusCode int
	Error      error
}

// MockGraphClient is a mock implementation for testing Graph API interactions.
// It simulates the behavior of the real Graph API client without making actual
// network requests, allowing for faster and more reliable tests.
//
// The mock client provides several features to make testing more realistic:
// - Predefined responses for specific API requests
// - Simulated network conditions (latency, packet loss, bandwidth)
// - Simulated error conditions (random errors, API throttling)
// - Recording of method calls for verification in tests
// - Thread-safety for concurrent tests
// - Pagination support for large collections
//
// Usage example:
//
//	client := NewMockGraphClient()
//	client.SetNetworkConditions(100*time.Millisecond, 0.1, 1024)
//	client.SetConfig(MockConfig{ErrorRate: 0.2, ThrottleRate: 0.1})
//	client.AddMockItem("/me/drive/root", &DriveItem{ID: "root", Name: "root"})
//	item, err := client.GetItem("root")
type MockGraphClient struct {
	// Auth is the authentication information
	Auth Auth

	// Mock behavior controls
	ShouldFailRefresh bool
	ShouldFailRequest bool
	RequestResponses  map[string]MockResponse

	// Simulated network conditions
	NetworkConditions NetworkConditions

	// Mock recorder for verification
	Recorder MockRecorder

	// Configuration for mock behavior
	Config MockConfig

	// Mutex for thread safety
	mu sync.Mutex

	// HTTP client that uses this mock
	httpClient *http.Client
}

// RoundTrip implements the http.RoundTripper interface
// This allows the MockGraphClient to intercept HTTP requests and provide mock responses
func (m *MockGraphClient) RoundTrip(req *http.Request) (*http.Response, error) {
	// Record the call
	m.Recorder.RecordCall("RoundTrip", req)

	// Check if we're in operational offline mode
	if GetOperationalOffline() {
		log.Debug().Msg("Mock client in operational offline mode, returning network error")
		return nil, errors.New("operational offline mode is enabled")
	}

	// Extract the resource path from the URL
	resource := strings.TrimPrefix(req.URL.Path, "/v1.0")
	if req.URL.RawQuery != "" {
		resource += "?" + req.URL.RawQuery
	}

	// Log the request details for debugging
	log.Debug().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Str("resource", resource).
		Str("client", "MockGraphClient").
		Msg("Mock client intercepted HTTP request")

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		return nil, err
	}

	// Check if we have a mock response for this resource
	m.mu.Lock()
	mockResponse, ok := m.RequestResponses[resource]

	// If not found, try with unescaped resource path
	if !ok {
		unescapedResource, err := url.PathUnescape(resource)
		if err == nil && unescapedResource != resource {
			mockResponse, ok = m.RequestResponses[unescapedResource]
		}
	}

	// If still not found, check if this is a content request and try with different path formats
	if !ok && strings.Contains(resource, "/content") {
		// Try different path formats for content requests
		alternateResources := []string{}

		// Extract the item ID and name from the resource path
		parts := strings.Split(resource, "/")

		// Format 1: /me/drive/items/{id}/content
		if len(parts) >= 4 {
			itemID := parts[len(parts)-2]
			alternateResources = append(alternateResources, "/me/drive/items/"+itemID+"/content")
		}

		// Format 2: /me/drive/items/{parentId}:/{name}:/content
		// This handles paths like /me/drive/items/parent-id:/file.txt:/content
		if strings.Contains(resource, ":/") {
			colonIndex := strings.Index(resource, ":/")
			if colonIndex > 0 {
				parentPath := resource[:colonIndex]
				remainingPath := resource[colonIndex+2:]
				// Extract parent ID
				parentParts := strings.Split(parentPath, "/")
				if len(parentParts) > 0 {
					parentID := parentParts[len(parentParts)-1]
					// Extract file name
					nameParts := strings.Split(remainingPath, ":/")
					if len(nameParts) > 0 {
						fileName := nameParts[0]
						alternateResources = append(alternateResources,
							"/me/drive/items/"+parentID+":/"+fileName+":/content")
					}
				}
			}
		}

		// Try each alternate resource path
		for _, altResource := range alternateResources {
			if mockResp, found := m.RequestResponses[altResource]; found {
				mockResponse = mockResp
				ok = true
				break
			}
		}
	}
	m.mu.Unlock()

	if !ok {
		// No mock response found, return a 404
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"error":{"code":"itemNotFound","message":"Item not found"}}`)),
			Header:     make(http.Header),
		}, nil
	}

	// If the mock response has an error, return it
	if mockResponse.Error != nil {
		return nil, mockResponse.Error
	}

	// Create and return the mock response
	return &http.Response{
		StatusCode: mockResponse.StatusCode,
		Body:       io.NopCloser(bytes.NewReader(mockResponse.Body)),
		Header:     make(http.Header),
	}, nil
}

// NewMockGraphClient creates a new MockGraphClient with default values
func NewMockGraphClient() *MockGraphClient {
	mock := &MockGraphClient{
		Auth: Auth{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			Account:      "mock@example.com",
		},
		RequestResponses: make(map[string]MockResponse),
		NetworkConditions: NetworkConditions{
			Latency:    0,
			PacketLoss: 0,
			Bandwidth:  0,
		},
		Recorder: &DefaultMockRecorder{
			calls: []MockCall{},
		},
		Config: MockConfig{
			Latency:        0,
			ErrorRate:      0,
			ResponseDelay:  0,
			ThrottleRate:   0,
			ThrottleDelay:  0,
			CustomBehavior: make(map[string]interface{}),
		},
	}

	// Create an HTTP client that uses this mock as its transport
	mock.httpClient = &http.Client{
		Transport: mock,
		Timeout:   defaultRequestTimeout,
	}

	// Set this mock's HTTP client as the test HTTP client
	log.Debug().Msg("Setting up MockGraphClient as the test HTTP client")
	SetHTTPClient(mock.httpClient)

	return mock
}

// SetNetworkConditions configures the network simulation conditions
func (m *MockGraphClient) SetNetworkConditions(latency time.Duration, packetLoss float64, bandwidth int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.NetworkConditions = NetworkConditions{
		Latency:    latency,
		PacketLoss: packetLoss,
		Bandwidth:  bandwidth,
	}
}

// SetConfig configures the mock behavior
func (m *MockGraphClient) SetConfig(config MockConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Config = config
}

// GetRecorder returns the mock recorder
func (m *MockGraphClient) GetRecorder() MockRecorder {
	return m.Recorder
}

// Cleanup resets the test HTTP client when the mock is no longer needed
// This ensures that tests don't interfere with each other
func (m *MockGraphClient) Cleanup() {
	log.Debug().Msg("Cleaning up MockGraphClient, resetting HTTP client to default")
	// Reset the test HTTP client
	SetHTTPClient(nil)
}

// simulateNetworkConditions applies the configured network conditions to a request.
// This method is used internally by other methods to simulate realistic network behavior.
//
// It simulates various network conditions and error scenarios:
//
// 1. Network Latency: Adds a delay to simulate network latency. The delay is the sum of:
//   - The latency from NetworkConditions (simulating base network latency)
//   - The latency from Config (simulating additional latency for specific tests)
//
// 2. Response Delay: Adds an additional delay to simulate slow server processing.
//
//  3. Packet Loss: Randomly fails requests based on the PacketLoss probability (0.0-1.0).
//     This simulates network packets being lost during transmission.
//
//  4. Random Errors: Randomly fails requests based on the ErrorRate probability (0.0-1.0).
//     This simulates various random errors that can occur during API calls.
//
//  5. API Throttling: Randomly fails requests with a throttling error based on the
//     ThrottleRate probability (0.0-1.0). If ThrottleDelay is set, it also adds a delay
//     before returning the error to simulate the backoff behavior of the real API.
//
//  6. Bandwidth Limitation: Simulates limited bandwidth by adding delays proportional
//     to the amount of data being transferred and inversely proportional to the
//     configured bandwidth.
//
// Returns:
//   - nil if no error is simulated
//   - An error describing the simulated failure otherwise
func (m *MockGraphClient) simulateNetworkConditions() error {
	m.mu.Lock()
	conditions := m.NetworkConditions
	config := m.Config
	m.mu.Unlock()

	// Apply latency from both network conditions and config
	latency := conditions.Latency
	if config.Latency > 0 {
		latency += config.Latency
	}
	if latency > 0 {
		time.Sleep(latency)
	}

	// Apply response delay from config
	if config.ResponseDelay > 0 {
		time.Sleep(config.ResponseDelay)
	}

	// Simulate packet loss
	if conditions.PacketLoss > 0 && rand.Float64() < conditions.PacketLoss {
		return errors.New("simulated packet loss")
	}

	// Simulate random errors based on error rate
	if config.ErrorRate > 0 && rand.Float64() < config.ErrorRate {
		return errors.New("simulated random error")
	}

	// Simulate API throttling
	if config.ThrottleRate > 0 && rand.Float64() < config.ThrottleRate {
		// If throttling is configured, simulate a throttling response
		if config.ThrottleDelay > 0 {
			time.Sleep(config.ThrottleDelay)
		}
		return errors.New("simulated API throttling: request rate exceeded")
	}

	// Simulate bandwidth limitation
	if conditions.Bandwidth > 0 {
		// Simple bandwidth simulation - sleep based on bandwidth
		// This is a very simplified model
		time.Sleep(time.Duration(1000/conditions.Bandwidth) * time.Millisecond)
	}

	return nil
}

// AddMockResponse adds a predefined response for a specific resource path
func (m *MockGraphClient) AddMockResponse(resource string, body []byte, statusCode int, err error) {
	// Check if this is a content resource
	if strings.Contains(resource, "/content") && statusCode == http.StatusOK {
		// Extract the item ID from the resource path
		var itemID string

		// Handle different path formats
		if strings.Contains(resource, ":/") {
			// Format: /me/drive/items/{parentId}:/{name}:/content
			colonIndex := strings.Index(resource, ":/")
			if colonIndex > 0 {
				parentPath := resource[:colonIndex]
				// Extract parent ID
				parentParts := strings.Split(parentPath, "/")
				if len(parentParts) > 0 {
					parentID := parentParts[len(parentParts)-1]

					// Extract file name
					remainingPath := resource[colonIndex+2:]
					nameParts := strings.Split(remainingPath, ":/")
					if len(nameParts) > 0 {
						fileName := nameParts[0]

						// Try to find the item by parent ID and name
						childResource := "/me/drive/items/" + parentID + "/children"
						m.mu.Lock()
						childrenResp, exists := m.RequestResponses[childResource]
						m.mu.Unlock()

						if exists {
							var children driveChildren
							if err := json.Unmarshal(childrenResp.Body, &children); err == nil {
								for _, child := range children.Children {
									if child.Name == fileName {
										itemID = child.ID
										break
									}
								}
							}
						}
					}
				}
			}
		} else {
			// Format: /me/drive/items/{id}/content
			parts := strings.Split(resource, "/")
			if len(parts) >= 4 {
				itemID = parts[len(parts)-2]
			}
		}

		// If we found an item ID, update its hash
		if itemID != "" {
			// Check if we have a mock item for this ID
			itemResource := "/me/drive/items/" + itemID
			m.mu.Lock()
			mockResponse, exists := m.RequestResponses[itemResource]
			m.mu.Unlock()

			if exists {
				// Unmarshal the item
				var item DriveItem
				if err := json.Unmarshal(mockResponse.Body, &item); err == nil {
					// If this is a file, update its hash
					if item.File != nil {
						// Calculate the QuickXorHash for the content
						contentHash := QuickXORHash(&body)

						// Update the item's hash
						item.File.Hashes.QuickXorHash = contentHash

						// Update the size from the response body
						// This is needed for tests to pass without special case code
						item.Size = uint64(len(body))

						// Marshal the updated item and update the mock response
						if updatedBody, err := json.Marshal(item); err == nil {
							m.mu.Lock()
							m.RequestResponses[itemResource] = MockResponse{
								Body:       updatedBody,
								StatusCode: http.StatusOK,
								Error:      nil,
							}
							m.mu.Unlock()
						}
					}
				}
			}
		}
	}

	// Add the original response
	m.mu.Lock()
	m.RequestResponses[resource] = MockResponse{
		Body:       body,
		StatusCode: statusCode,
		Error:      err,
	}
	m.mu.Unlock()
}

// AddMockItem adds a predefined DriveItem response for a specific resource path
func (m *MockGraphClient) AddMockItem(resource string, item *DriveItem) {
	body, _ := json.Marshal(item)
	m.AddMockResponse(resource, body, http.StatusOK, nil)
}

// AddMockItems adds a predefined list of DriveItems for a children request
func (m *MockGraphClient) AddMockItems(resource string, items []*DriveItem) {
	// Default behavior - no pagination
	m.AddMockItemsWithPagination(resource, items, 0)
}

// AddMockItemsWithPagination adds a predefined list of DriveItems with pagination support
// pageSize of 0 means no pagination
func (m *MockGraphClient) AddMockItemsWithPagination(resource string, items []*DriveItem, pageSize int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pageSize <= 0 || len(items) <= pageSize {
		// No pagination needed or requested
		response := driveChildren{
			Children: items,
		}
		body, _ := json.Marshal(response)
		m.RequestResponses[resource] = MockResponse{
			Body:       body,
			StatusCode: http.StatusOK,
			Error:      nil,
		}
		return
	}

	// Implement pagination
	for i := 0; i < len(items); i += pageSize {
		end := i + pageSize
		if end > len(items) {
			end = len(items)
		}

		pageItems := items[i:end]
		nextLink := ""
		if end < len(items) {
			nextLink = fmt.Sprintf("%s%s?skiptoken=%d", GraphURL, resource, end)
		}

		response := driveChildren{
			Children: pageItems,
			NextLink: nextLink,
		}

		body, _ := json.Marshal(response)

		// For the first page, use the original resource
		if i == 0 {
			m.RequestResponses[resource] = MockResponse{
				Body:       body,
				StatusCode: http.StatusOK,
				Error:      nil,
			}
		} else {
			// For subsequent pages, use a resource with skiptoken
			paginatedResource := fmt.Sprintf("%s?skiptoken=%d", resource, i)
			m.RequestResponses[paginatedResource] = MockResponse{
				Body:       body,
				StatusCode: http.StatusOK,
				Error:      nil,
			}
		}
	}
}

// RequestWithContext is a mock implementation of the real RequestWithContext function
func (m *MockGraphClient) RequestWithContext(ctx context.Context, resource string, method string, content io.Reader, headers ...Header) ([]byte, error) {
	// Record the call
	var contentBytes []byte
	if content != nil {
		var err error
		contentBytes, err = ioutil.ReadAll(content)
		if err != nil {
			return nil, fmt.Errorf("error reading content: %v", err)
		}
		// Create a new reader with the same content for later use
		content = strings.NewReader(string(contentBytes))
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Check if we're in operational offline mode
	if GetOperationalOffline() {
		log.Debug().Msg("Mock client in operational offline mode, returning network error")
		return nil, errors.New("operational offline mode is enabled")
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		return nil, err
	}

	// Check if we should fail the request
	if m.ShouldFailRequest {
		return nil, errors.New("mock request failure")
	}

	// Check if we have a predefined response for this resource
	m.mu.Lock()
	response, exists := m.RequestResponses[resource]

	// If not found, try with unescaped resource path
	if !exists {
		unescapedResource, err := url.PathUnescape(resource)
		if err == nil && unescapedResource != resource {
			response, exists = m.RequestResponses[unescapedResource]
		}
	}

	// If still not found, check if this is a content request and try with different path formats
	if !exists && strings.Contains(resource, "/content") {
		// Try different path formats for content requests
		alternateResources := []string{}

		// Extract the item ID and name from the resource path
		parts := strings.Split(resource, "/")

		// Format 1: /me/drive/items/{id}/content
		if len(parts) >= 4 {
			itemID := parts[len(parts)-2]
			alternateResources = append(alternateResources, "/me/drive/items/"+itemID+"/content")
		}

		// Format 2: /me/drive/items/{parentId}:/{name}:/content
		// This handles paths like /me/drive/items/parent-id:/file.txt:/content
		if strings.Contains(resource, ":/") {
			colonIndex := strings.Index(resource, ":/")
			if colonIndex > 0 {
				parentPath := resource[:colonIndex]
				remainingPath := resource[colonIndex+2:]
				// Extract parent ID
				parentParts := strings.Split(parentPath, "/")
				if len(parentParts) > 0 {
					parentID := parentParts[len(parentParts)-1]
					// Extract file name
					nameParts := strings.Split(remainingPath, ":/")
					if len(nameParts) > 0 {
						fileName := nameParts[0]
						alternateResources = append(alternateResources,
							"/me/drive/items/"+parentID+":/"+fileName+":/content")
					}
				}
			}
		}

		// Try each alternate resource path
		for _, altResource := range alternateResources {
			if mockResp, found := m.RequestResponses[altResource]; found {
				response = mockResp
				exists = true
				break
			}
		}
	}
	m.mu.Unlock()

	if exists {
		if response.Error != nil {
			return nil, response.Error
		}
		return response.Body, nil
	}

	// Default response based on the resource and method
	var result []byte
	var err error

	if strings.Contains(resource, "/children") {
		// Return empty children list by default
		result = []byte(`{"value":[]}`)
	} else if method == "GET" && strings.Contains(resource, "/content") {
		// Return empty content by default
		result = []byte{}
	} else if method == "DELETE" {
		// Return success for DELETE
		result = nil
	} else {
		// For other requests, return a generic DriveItem
		item := &DriveItem{
			ID:   "mock-id",
			Name: "mock-item",
		}
		result, _ = json.Marshal(item)
	}

	return result, err
}

// Get is a mock implementation of the real Get function
func (m *MockGraphClient) Get(resource string, headers ...Header) ([]byte, error) {
	args := []interface{}{resource}
	for _, h := range headers {
		args = append(args, h)
	}

	result, err := m.RequestWithContext(context.Background(), resource, "GET", nil, headers...)

	m.Recorder.RecordCall("Get", append(args, result)...)
	return result, err
}

// GetWithContext is a mock implementation of the real GetWithContext function
func (m *MockGraphClient) GetWithContext(ctx context.Context, resource string, headers ...Header) ([]byte, error) {
	args := []interface{}{ctx, resource}
	for _, h := range headers {
		args = append(args, h)
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		m.Recorder.RecordCallWithResult("GetWithContext", ctx.Err(), args...)
		return nil, ctx.Err()
	}

	result, err := m.RequestWithContext(ctx, resource, "GET", nil, headers...)

	m.Recorder.RecordCallWithResult("GetWithContext", result, args...)
	return result, err
}

// Patch is a mock implementation of the real Patch function
func (m *MockGraphClient) Patch(resource string, content io.Reader, headers ...Header) ([]byte, error) {
	var contentBytes []byte
	if content != nil {
		var err error
		contentBytes, err = ioutil.ReadAll(content)
		if err != nil {
			return nil, fmt.Errorf("error reading content: %v", err)
		}
		// Create a new reader with the same content
		content = strings.NewReader(string(contentBytes))
	}

	call := MockCall{
		Method:    "Patch",
		Resource:  resource,
		Args:      []interface{}{resource, contentBytes},
		Timestamp: time.Now(),
	}

	for _, h := range headers {
		call.Args = append(call.Args, h)
	}

	result, err := m.RequestWithContext(context.Background(), resource, "PATCH", content, headers...)
	call.Result = result
	m.Recorder.RecordCall(call.Method, call.Args...)
	return result, err
}

// Post is a mock implementation of the real Post function
func (m *MockGraphClient) Post(resource string, content io.Reader, headers ...Header) ([]byte, error) {
	var contentBytes []byte
	if content != nil {
		var err error
		contentBytes, err = ioutil.ReadAll(content)
		if err != nil {
			return nil, fmt.Errorf("error reading content: %v", err)
		}
		// Create a new reader with the same content
		content = strings.NewReader(string(contentBytes))
	}

	args := []interface{}{resource, contentBytes}
	for _, h := range headers {
		args = append(args, h)
	}

	result, err := m.RequestWithContext(context.Background(), resource, "POST", content, headers...)

	m.Recorder.RecordCall("Post", append(args, result)...)
	return result, err
}

// Put is a mock implementation of the real Put function
func (m *MockGraphClient) Put(resource string, content io.Reader, headers ...Header) ([]byte, error) {
	var contentBytes []byte
	if content != nil {
		var err error
		contentBytes, err = ioutil.ReadAll(content)
		if err != nil {
			return nil, fmt.Errorf("error reading content: %v", err)
		}
		// Create a new reader with the same content
		content = strings.NewReader(string(contentBytes))
	}

	call := MockCall{
		Method:    "Put",
		Resource:  resource,
		Args:      []interface{}{resource, contentBytes},
		Timestamp: time.Now(),
	}

	for _, h := range headers {
		call.Args = append(call.Args, h)
	}

	// If this is a content upload, ensure we have a mock response for it
	if strings.Contains(resource, "/content") && contentBytes != nil {
		// Check if we already have a response for this resource
		m.mu.Lock()
		_, exists := m.RequestResponses[resource]
		m.mu.Unlock()

		if !exists {
			// Extract the item ID or parent ID and name
			var itemID string
			var fileItem *DriveItem

			if strings.Contains(resource, ":/") {
				// Format: /me/drive/items/{parentId}:/{name}:/content
				colonIndex := strings.Index(resource, ":/")
				if colonIndex > 0 {
					parentPath := resource[:colonIndex]
					remainingPath := resource[colonIndex+2:]

					// Extract parent ID
					parentParts := strings.Split(parentPath, "/")
					if len(parentParts) > 0 {
						parentID := parentParts[len(parentParts)-1]

						// Extract file name
						nameParts := strings.Split(remainingPath, ":/")
						if len(nameParts) > 0 {
							fileName := nameParts[0]

							// Create a new file item
							fileItem = &DriveItem{
								ID:   "generated-id-" + fileName,
								Name: fileName,
								File: &File{
									Hashes: Hashes{
										QuickXorHash: QuickXORHash(&contentBytes),
									},
								},
								// Update the size from the content bytes
								// This is needed for tests to pass without special case code
								Size: uint64(len(contentBytes)),
							}

							// Add the item to the parent's children
							m.AddMockItem("/me/drive/items/"+parentID+":/"+fileName+":", fileItem)
						}
					}
				}
			} else {
				// Format: /me/drive/items/{id}/content
				parts := strings.Split(resource, "/")
				if len(parts) >= 4 {
					itemID = parts[len(parts)-2]

					// Get the existing item
					m.mu.Lock()
					itemResource := "/me/drive/items/" + itemID
					mockResponse, exists := m.RequestResponses[itemResource]
					m.mu.Unlock()

					if exists {
						var item DriveItem
						if err := json.Unmarshal(mockResponse.Body, &item); err == nil {
							// Update the item with new content
							if item.File == nil {
								item.File = &File{}
							}
							if item.File.Hashes.QuickXorHash == "" {
								item.File.Hashes.QuickXorHash = QuickXORHash(&contentBytes)
							}
							// Update the size from the content bytes
							// This is needed for tests to pass without special case code
							item.Size = uint64(len(contentBytes))
							fileItem = &item

							// Update the item
							m.AddMockItem(itemResource, fileItem)
						}
					}
				}
			}

			// Add a mock response for the content upload
			if fileItem != nil {
				fileItemJSON, _ := json.Marshal(fileItem)
				m.AddMockResponse(resource, fileItemJSON, http.StatusOK, nil)
			}
		}
	}

	result, err := m.RequestWithContext(context.Background(), resource, "PUT", content, headers...)
	call.Result = result
	m.Recorder.RecordCall(call.Method, call.Args...)
	return result, err
}

// Delete is a mock implementation of the real Delete function
func (m *MockGraphClient) Delete(resource string, headers ...Header) error {
	args := []interface{}{resource}
	for _, h := range headers {
		args = append(args, h)
	}

	_, err := m.RequestWithContext(context.Background(), resource, "DELETE", nil, headers...)

	m.Recorder.RecordCall("Delete", append(args, err)...)
	return err
}

// GetItemContent is a mock implementation of the real GetItemContent function
func (m *MockGraphClient) GetItemContent(id string) ([]byte, uint64, error) {
	call := MockCall{
		Method:    "GetItemContent",
		Args:      []interface{}{id},
		Timestamp: time.Now(),
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return nil, 0, err
	}

	resource := fmt.Sprintf("/me/drive/items/%s/content", id)
	m.mu.Lock()
	response, exists := m.RequestResponses[resource]
	m.mu.Unlock()

	if exists {
		if response.Error != nil {
			call.Result = response.Error
			m.Recorder.RecordCall(call.Method, call.Args...)
			return nil, 0, response.Error
		}
		call.Result = response.Body
		m.Recorder.RecordCall(call.Method, call.Args...)
		return response.Body, uint64(len(response.Body)), nil
	}

	// Default empty content
	call.Result = []byte{}
	m.Recorder.RecordCall(call.Method, call.Args...)
	return []byte{}, 0, nil
}

// GetItemContentStream is a mock implementation of the real GetItemContentStream function
func (m *MockGraphClient) GetItemContentStream(id string, output io.Writer) (uint64, error) {
	call := MockCall{
		Method:    "GetItemContentStream",
		Args:      []interface{}{id, output},
		Timestamp: time.Now(),
	}

	content, size, err := m.GetItemContent(id)
	if err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return 0, err
	}

	// Simulate bandwidth limitation if configured
	if m.NetworkConditions.Bandwidth > 0 {
		// Simple bandwidth simulation - write in chunks with delays
		chunkSize := 1024 // 1KB chunks
		for i := 0; i < len(content); i += chunkSize {
			end := i + chunkSize
			if end > len(content) {
				end = len(content)
			}

			_, err = output.Write(content[i:end])
			if err != nil {
				call.Result = err
				m.Recorder.RecordCall(call.Method, call.Args...)
				return 0, err
			}

			// Sleep based on bandwidth setting
			time.Sleep(time.Duration(1000/m.NetworkConditions.Bandwidth) * time.Millisecond)
		}
	} else {
		_, err = output.Write(content)
		if err != nil {
			call.Result = err
			m.Recorder.RecordCall(call.Method, call.Args...)
			return 0, err
		}
	}

	call.Result = size
	m.Recorder.RecordCall(call.Method, call.Args...)
	return size, nil
}

// GetItem is a mock implementation of the real GetItem function
func (m *MockGraphClient) GetItem(id string) (*DriveItem, error) {
	call := MockCall{
		Method:    "GetItem",
		Args:      []interface{}{id},
		Timestamp: time.Now(),
	}

	resource := IDPath(id)
	body, err := m.Get(resource)
	if err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return nil, err
	}

	item := &DriveItem{}
	err = json.Unmarshal(body, item)
	call.Result = item
	m.Recorder.RecordCall(call.Method, call.Args...)
	return item, err
}

// GetItemPath is a mock implementation of the real GetItemPath function
func (m *MockGraphClient) GetItemPath(path string) (*DriveItem, error) {
	call := MockCall{
		Method:    "GetItemPath",
		Args:      []interface{}{path},
		Timestamp: time.Now(),
	}

	resource := ResourcePath(path)
	body, err := m.Get(resource)
	if err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return nil, err
	}

	item := &DriveItem{}
	err = json.Unmarshal(body, item)
	call.Result = item
	m.Recorder.RecordCall(call.Method, call.Args...)
	return item, err
}

// GetItemChildren is a mock implementation of the real GetItemChildren function
func (m *MockGraphClient) GetItemChildren(id string) ([]*DriveItem, error) {
	call := MockCall{
		Method:    "GetItemChildren",
		Args:      []interface{}{id},
		Timestamp: time.Now(),
	}

	// Start with the initial resource path
	resource := childrenPathID(id)
	allChildren := make([]*DriveItem, 0)

	// Loop until we've processed all pages
	for resource != "" {
		body, err := m.Get(resource)
		if err != nil {
			call.Result = err
			m.Recorder.RecordCall(call.Method, call.Args...)
			return nil, err
		}

		var result driveChildren
		err = json.Unmarshal(body, &result)
		if err != nil {
			call.Result = err
			m.Recorder.RecordCall(call.Method, call.Args...)
			return nil, err
		}

		// Append the children from this page
		allChildren = append(allChildren, result.Children...)

		// If there's a nextLink, prepare for the next iteration
		if result.NextLink != "" {
			resource = strings.TrimPrefix(result.NextLink, GraphURL)
		} else {
			// No more pages
			resource = ""
		}
	}

	call.Result = allChildren
	m.Recorder.RecordCall(call.Method, call.Args...)
	return allChildren, nil
}

// GetItemChildrenPath is a mock implementation of the real GetItemChildrenPath function
func (m *MockGraphClient) GetItemChildrenPath(path string) ([]*DriveItem, error) {
	call := MockCall{
		Method:    "GetItemChildrenPath",
		Args:      []interface{}{path},
		Timestamp: time.Now(),
	}

	// Start with the initial resource path
	resource := childrenPath(path)
	allChildren := make([]*DriveItem, 0)

	// Loop until we've processed all pages
	for resource != "" {
		body, err := m.Get(resource)
		if err != nil {
			call.Result = err
			m.Recorder.RecordCall(call.Method, call.Args...)
			return nil, err
		}

		var result driveChildren
		err = json.Unmarshal(body, &result)
		if err != nil {
			call.Result = err
			m.Recorder.RecordCall(call.Method, call.Args...)
			return nil, err
		}

		// Append the children from this page
		allChildren = append(allChildren, result.Children...)

		// If there's a nextLink, prepare for the next iteration
		if result.NextLink != "" {
			resource = strings.TrimPrefix(result.NextLink, GraphURL)
		} else {
			// No more pages
			resource = ""
		}
	}

	call.Result = allChildren
	m.Recorder.RecordCall(call.Method, call.Args...)
	return allChildren, nil
}

// Mkdir is a mock implementation of the real Mkdir function
func (m *MockGraphClient) Mkdir(name string, parentID string) (*DriveItem, error) {
	call := MockCall{
		Method:    "Mkdir",
		Args:      []interface{}{name, parentID},
		Timestamp: time.Now(),
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return nil, err
	}

	newFolder := DriveItem{
		Name:   name,
		Folder: &Folder{},
	}
	bytePayload, _ := json.Marshal(newFolder)
	resp, err := m.Post(childrenPathID(parentID), strings.NewReader(string(bytePayload)))
	if err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return nil, err
	}

	err = json.Unmarshal(resp, &newFolder)
	call.Result = &newFolder
	m.Recorder.RecordCall(call.Method, call.Args...)
	return &newFolder, err
}

// Rename is a mock implementation of the real Rename function
func (m *MockGraphClient) Rename(itemID string, itemName string, parentID string) error {
	call := MockCall{
		Method:    "Rename",
		Args:      []interface{}{itemID, itemName, parentID},
		Timestamp: time.Now(),
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return err
	}

	patchContent := DriveItem{
		ConflictBehavior: "replace",
		Name:             itemName,
		Parent: &DriveItemParent{
			ID: parentID,
		},
	}

	jsonPatch, _ := json.Marshal(patchContent)
	_, err := m.Patch("/me/drive/items/"+itemID, strings.NewReader(string(jsonPatch)))
	call.Result = err
	m.Recorder.RecordCall(call.Method, call.Args...)
	return err
}

// Remove is a mock implementation of the real Remove function
func (m *MockGraphClient) Remove(id string) error {
	call := MockCall{
		Method:    "Remove",
		Args:      []interface{}{id},
		Timestamp: time.Now(),
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return err
	}

	err := m.Delete("/me/drive/items/" + id)
	call.Result = err
	m.Recorder.RecordCall(call.Method, call.Args...)
	return err
}

// GetUser is a mock implementation of the real GetUser function
func (m *MockGraphClient) GetUser() (User, error) {
	call := MockCall{
		Method:    "GetUser",
		Args:      []interface{}{},
		Timestamp: time.Now(),
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return User{}, err
	}

	// Check if we have a predefined response for this resource
	m.mu.Lock()
	response, exists := m.RequestResponses["/me"]
	m.mu.Unlock()

	if exists {
		if response.Error != nil {
			call.Result = response.Error
			m.Recorder.RecordCall(call.Method, call.Args...)
			return User{}, response.Error
		}

		var user User
		err := json.Unmarshal(response.Body, &user)
		call.Result = user
		m.Recorder.RecordCall(call.Method, call.Args...)
		return user, err
	}

	// Default mock user
	user := User{
		UserPrincipalName: "mock@example.com",
	}
	call.Result = user
	m.Recorder.RecordCall(call.Method, call.Args...)
	return user, nil
}

// GetUserWithContext is a mock implementation of the real GetUserWithContext function
func (m *MockGraphClient) GetUserWithContext(ctx context.Context) (User, error) {
	call := MockCall{
		Method:    "GetUserWithContext",
		Args:      []interface{}{ctx},
		Timestamp: time.Now(),
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		call.Result = ctx.Err()
		m.Recorder.RecordCall(call.Method, call.Args...)
		return User{}, ctx.Err()
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return User{}, err
	}

	// Check if we have a predefined response for this resource
	m.mu.Lock()
	response, exists := m.RequestResponses["/me"]
	m.mu.Unlock()

	if exists {
		if response.Error != nil {
			call.Result = response.Error
			m.Recorder.RecordCall(call.Method, call.Args...)
			return User{}, response.Error
		}

		var user User
		err := json.Unmarshal(response.Body, &user)
		call.Result = user
		m.Recorder.RecordCall(call.Method, call.Args...)
		return user, err
	}

	// Default mock user
	user := User{
		UserPrincipalName: "mock@example.com",
	}
	call.Result = user
	m.Recorder.RecordCall(call.Method, call.Args...)
	return user, nil
}

// GetDrive is a mock implementation of the real GetDrive function
func (m *MockGraphClient) GetDrive() (Drive, error) {
	call := MockCall{
		Method:    "GetDrive",
		Args:      []interface{}{},
		Timestamp: time.Now(),
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return Drive{}, err
	}

	// Check if we have a predefined response for this resource
	m.mu.Lock()
	response, exists := m.RequestResponses["/me/drive"]
	m.mu.Unlock()

	if exists {
		if response.Error != nil {
			call.Result = response.Error
			m.Recorder.RecordCall(call.Method, call.Args...)
			return Drive{}, response.Error
		}

		var drive Drive
		err := json.Unmarshal(response.Body, &drive)
		call.Result = drive
		m.Recorder.RecordCall(call.Method, call.Args...)
		return drive, err
	}

	// Default mock drive
	drive := Drive{
		ID:        "mock-drive-id",
		DriveType: DriveTypePersonal,
		Quota: DriveQuota{
			Total:     1024 * 1024 * 1024 * 10, // 10 GB
			Used:      1024 * 1024 * 1024 * 2,  // 2 GB
			Remaining: 1024 * 1024 * 1024 * 8,  // 8 GB
			State:     "normal",
		},
	}
	call.Result = drive
	m.Recorder.RecordCall(call.Method, call.Args...)
	return drive, nil
}

// GetItemChild is a mock implementation of the real GetItemChild function
func (m *MockGraphClient) GetItemChild(id string, name string) (*DriveItem, error) {
	call := MockCall{
		Method:    "GetItemChild",
		Args:      []interface{}{id, name},
		Timestamp: time.Now(),
	}

	// Simulate network conditions
	if err := m.simulateNetworkConditions(); err != nil {
		call.Result = err
		m.Recorder.RecordCall(call.Method, call.Args...)
		return nil, err
	}

	// Construct the resource path
	resource := fmt.Sprintf("%s:/%s", IDPath(id), url.PathEscape(name))

	// Check if we have a predefined response for this resource
	m.mu.Lock()
	response, exists := m.RequestResponses[resource]
	m.mu.Unlock()

	if exists {
		if response.Error != nil {
			call.Result = response.Error
			m.Recorder.RecordCall(call.Method, call.Args...)
			return nil, response.Error
		}

		var item DriveItem
		err := json.Unmarshal(response.Body, &item)
		call.Result = &item
		m.Recorder.RecordCall(call.Method, call.Args...)
		return &item, err
	}

	// Default mock item
	item := &DriveItem{
		ID:   "mock-child-id",
		Name: name,
	}
	call.Result = item
	m.Recorder.RecordCall(call.Method, call.Args...)
	return item, nil
}
