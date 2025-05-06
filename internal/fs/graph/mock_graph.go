// Package graph provides mocks for testing without actual API calls
package graph

import (
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
	CustomBehavior map[string]interface{} // Custom behavior configuration
}

// MockResponse represents a predefined response for a specific request
type MockResponse struct {
	Body       []byte
	StatusCode int
	Error      error
}

// MockGraphClient is a mock implementation for testing Graph API interactions
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
}

// NewMockGraphClient creates a new MockGraphClient with default values
func NewMockGraphClient() *MockGraphClient {
	return &MockGraphClient{
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
			CustomBehavior: make(map[string]interface{}),
		},
	}
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

// simulateNetworkConditions applies the configured network conditions to a request
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
	m.RequestResponses[resource] = MockResponse{
		Body:       body,
		StatusCode: statusCode,
		Error:      err,
	}
}

// AddMockItem adds a predefined DriveItem response for a specific resource path
func (m *MockGraphClient) AddMockItem(resource string, item *DriveItem) {
	body, _ := json.Marshal(item)
	m.AddMockResponse(resource, body, http.StatusOK, nil)
}

// AddMockItems adds a predefined list of DriveItems for a children request
func (m *MockGraphClient) AddMockItems(resource string, items []*DriveItem) {
	response := driveChildren{
		Children: items,
	}
	body, _ := json.Marshal(response)
	m.AddMockResponse(resource, body, http.StatusOK, nil)
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

	resource := childrenPathID(id)
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

	call.Result = result.Children
	m.Recorder.RecordCall(call.Method, call.Args...)
	return result.Children, nil
}

// GetItemChildrenPath is a mock implementation of the real GetItemChildrenPath function
func (m *MockGraphClient) GetItemChildrenPath(path string) ([]*DriveItem, error) {
	call := MockCall{
		Method:    "GetItemChildrenPath",
		Args:      []interface{}{path},
		Timestamp: time.Now(),
	}

	resource := childrenPath(path)
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

	call.Result = result.Children
	m.Recorder.RecordCall(call.Method, call.Args...)
	return result.Children, nil
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
