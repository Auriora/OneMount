// Package testutil provides testing utilities for the OneMount project.
package mock

import (
	"context"
	"io"
	"time"

	"github.com/auriora/onemount/internal/fs/graph"
)

// MockCall represents a record of a method call on a mock
type MockCall struct {
	Method string
	Args   []interface{}
	Result interface{}
	Error  error
	Time   time.Time
}

// NetworkConditions simulates different network scenarios
type NetworkConditions struct {
	Latency    time.Duration
	PacketLoss float64
	Bandwidth  int
}

// MockRecorder records and verifies mock interactions
type MockRecorder interface {
	// RecordCall records a method call
	RecordCall(method string, args ...interface{})
	// RecordCallWithResult records a method call with a result and error
	RecordCallWithResult(method string, result interface{}, err error, args ...interface{})
	// GetCalls returns all recorded calls
	GetCalls() []MockCall
	// VerifyCall verifies a method was called a specific number of times
	VerifyCall(method string, times int) bool
}

// MockConfig defines configuration for mock behavior
type MockConfig struct {
	Latency        time.Duration
	ErrorRate      float64
	ResponseDelay  time.Duration
	CustomBehavior map[string]interface{}
}

// BasicMockRecorder is a simple implementation of the MockRecorder interface
type BasicMockRecorder struct {
	calls []MockCall
}

// NewBasicMockRecorder creates a new BasicMockRecorder
func NewBasicMockRecorder() *BasicMockRecorder {
	return &BasicMockRecorder{
		calls: make([]MockCall, 0),
	}
}

// RecordCall records a method call
func (r *BasicMockRecorder) RecordCall(method string, args ...interface{}) {
	r.calls = append(r.calls, MockCall{
		Method: method,
		Args:   args,
		Time:   time.Now(),
	})
}

// RecordCallWithResult records a method call with a result and error
func (r *BasicMockRecorder) RecordCallWithResult(method string, result interface{}, err error, args ...interface{}) {
	r.calls = append(r.calls, MockCall{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
		Time:   time.Now(),
	})
}

// GetCalls returns all recorded calls
func (r *BasicMockRecorder) GetCalls() []MockCall {
	return r.calls
}

// VerifyCall verifies a method was called a specific number of times
func (r *BasicMockRecorder) VerifyCall(method string, times int) bool {
	count := 0
	for _, call := range r.calls {
		if call.Method == method {
			count++
		}
	}
	return count == times
}

// MockGraphProvider implements the MockProvider interface for simulating Microsoft Graph API responses
type MockGraphProvider struct {
	// Underlying mock graph client
	Client *graph.MockGraphClient

	// Record of calls made to the mock
	recorder MockRecorder

	// Simulated network conditions
	networkConditions NetworkConditions

	// Configuration for mock behavior
	config MockConfig
}

// NewMockGraphProvider creates a new MockGraphProvider
func NewMockGraphProvider() *MockGraphProvider {
	return &MockGraphProvider{
		Client:   graph.NewMockGraphClient(),
		recorder: NewBasicMockRecorder(),
		config:   MockConfig{},
	}
}

// Setup initializes the mock provider
func (m *MockGraphProvider) Setup() error {
	// Nothing to do for basic setup
	return nil
}

// Teardown cleans up the mock provider
func (m *MockGraphProvider) Teardown() error {
	// Nothing to do for basic teardown
	return nil
}

// Reset resets the mock provider to its initial state
func (m *MockGraphProvider) Reset() error {
	m.Client = graph.NewMockGraphClient()
	m.recorder = NewBasicMockRecorder()
	m.config = MockConfig{}
	m.networkConditions = NetworkConditions{}
	return nil
}

// SetNetworkConditions sets the simulated network conditions
func (m *MockGraphProvider) SetNetworkConditions(latency time.Duration, packetLoss float64, bandwidth int) {
	m.networkConditions = NetworkConditions{
		Latency:    latency,
		PacketLoss: packetLoss,
		Bandwidth:  bandwidth,
	}
}

// SetConfig sets the mock configuration
func (m *MockGraphProvider) SetConfig(config MockConfig) {
	m.config = config
}

// GetRecorder returns the mock recorder
func (m *MockGraphProvider) GetRecorder() MockRecorder {
	return m.recorder
}

// AddMockResponse adds a predefined response for a specific resource path
func (m *MockGraphProvider) AddMockResponse(resource string, body []byte, statusCode int, err error) {
	m.Client.AddMockResponse(resource, body, statusCode, err)
	m.recorder.RecordCall("AddMockResponse", resource, body, statusCode, err)
}

// AddMockItem adds a predefined DriveItem response for a specific resource path
func (m *MockGraphProvider) AddMockItem(resource string, item *graph.DriveItem) {
	m.Client.AddMockItem(resource, item)
	m.recorder.RecordCall("AddMockItem", resource, item)
}

// AddMockItems adds a predefined list of DriveItems for a children request
func (m *MockGraphProvider) AddMockItems(resource string, items []*graph.DriveItem) {
	m.Client.AddMockItems(resource, items)
	m.recorder.RecordCall("AddMockItems", resource, items)
}

// SimulateNetworkDelay simulates network delay based on current network conditions
func (m *MockGraphProvider) SimulateNetworkDelay() {
	if m.networkConditions.Latency > 0 {
		time.Sleep(m.networkConditions.Latency)
	}
}

// SimulateNetworkError returns an error based on the error rate in the config
func (m *MockGraphProvider) SimulateNetworkError() error {
	if m.config.ErrorRate > 0 {
		// Generate a random number between 0 and 1
		// If it's less than the error rate, return an error
		if float64(time.Now().UnixNano()%100)/100 < m.config.ErrorRate {
			return io.ErrUnexpectedEOF
		}
	}
	return nil
}

// RequestWithContext wraps the client's RequestWithContext method with recording and network simulation
func (m *MockGraphProvider) RequestWithContext(ctx context.Context, resource string, method string, content io.Reader, headers ...graph.Header) ([]byte, error) {
	m.recorder.RecordCall("RequestWithContext", resource, method, content, headers)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.RequestWithContext(ctx, resource, method, content, headers...)

	// Record the result
	m.recorder.RecordCallWithResult("RequestWithContext", result, err, resource, method, content, headers)

	return result, err
}

// Get wraps the client's Get method with recording and network simulation
func (m *MockGraphProvider) Get(resource string, headers ...graph.Header) ([]byte, error) {
	m.recorder.RecordCall("Get", resource, headers)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.Get(resource, headers...)

	// Record the result
	m.recorder.RecordCallWithResult("Get", result, err, resource, headers)

	return result, err
}

// GetWithContext wraps the client's GetWithContext method with recording and network simulation
func (m *MockGraphProvider) GetWithContext(ctx context.Context, resource string, headers ...graph.Header) ([]byte, error) {
	m.recorder.RecordCall("GetWithContext", ctx, resource, headers)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.GetWithContext(ctx, resource, headers...)

	// Record the result
	m.recorder.RecordCallWithResult("GetWithContext", result, err, ctx, resource, headers)

	return result, err
}

// GetItem wraps the client's GetItem method with recording and network simulation
func (m *MockGraphProvider) GetItem(id string) (*graph.DriveItem, error) {
	m.recorder.RecordCall("GetItem", id)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.GetItem(id)

	// Record the result
	m.recorder.RecordCallWithResult("GetItem", result, err, id)

	return result, err
}

// GetItemChildren wraps the client's GetItemChildren method with recording and network simulation
func (m *MockGraphProvider) GetItemChildren(id string) ([]*graph.DriveItem, error) {
	m.recorder.RecordCall("GetItemChildren", id)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.GetItemChildren(id)

	// Record the result
	m.recorder.RecordCallWithResult("GetItemChildren", result, err, id)

	return result, err
}
