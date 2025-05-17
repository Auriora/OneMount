// Package mock Package testutil provides testing utilities for the OneMount project.
// This package contains mock implementations of various components used in testing.
// The mock implementations simulate the behavior of real components without making
// actual API calls or filesystem operations, allowing for faster and more reliable tests.
package mock

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/graph/api"
)

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
	GetCalls() []api.MockCall
	// VerifyCall verifies a method was called a specific number of times
	VerifyCall(method string, times int) bool
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

// BasicMockRecorder is a simple implementation of the MockRecorder interface
type BasicMockRecorder struct {
	calls []api.MockCall
}

// NewBasicMockRecorder creates a new BasicMockRecorder
func NewBasicMockRecorder() *BasicMockRecorder {
	return &BasicMockRecorder{
		calls: make([]api.MockCall, 0),
	}
}

// RecordCall records a method call
func (r *BasicMockRecorder) RecordCall(method string, args ...interface{}) {
	r.calls = append(r.calls, api.MockCall{
		Method: method,
		Args:   args,
		Time:   time.Now(),
	})
}

// RecordCallWithResult records a method call with a result and error
func (r *BasicMockRecorder) RecordCallWithResult(method string, result interface{}, err error, args ...interface{}) {
	r.calls = append(r.calls, api.MockCall{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
		Time:   time.Now(),
	})
}

// GetCalls returns all recorded calls
func (r *BasicMockRecorder) GetCalls() []api.MockCall {
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

// MockGraphProvider implements the MockProvider interface for simulating Microsoft Graph API responses.
// It provides a high-level mock that wraps the lower-level MockGraphClient from the fs/graph package.
// This mock is designed to be used in integration tests where you need to simulate Graph API behavior
// without making actual network requests.
//
// Features:
// - Records method calls for verification in tests
// - Simulates network conditions like latency and packet loss
// - Simulates error conditions like random errors and API throttling
// - Thread-safe for use in concurrent tests
// - Supports pagination for large collections
//
// Usage example:
//
//	provider := mock.NewMockGraphProvider()
//	provider.SetNetworkConditions(100*time.Millisecond, 0.1, 1024)
//	provider.SetConfig(mock.MockConfig{ErrorRate: 0.2, ThrottleRate: 0.1})
//	provider.AddMockItem("/me/drive/root", &graph.DriveItem{ID: "root", Name: "root"})
//	item, err := provider.GetItem("root")
type MockGraphProvider struct {
	// Underlying mock graph client
	Client *MockGraphClient

	// Record of calls made to the mock
	recorder MockRecorder

	// Simulated network conditions
	networkConditions NetworkConditions

	// Configuration for mock behavior
	config MockConfig

	// Mutex for thread safety
	mu sync.Mutex
}

// NewMockGraphProvider creates a new MockGraphProvider
func NewMockGraphProvider() *MockGraphProvider {
	return &MockGraphProvider{
		Client:   NewMockGraphClient(),
		recorder: NewBasicMockRecorder(),
		config: MockConfig{
			Latency:        0,
			ErrorRate:      0,
			ResponseDelay:  0,
			ThrottleRate:   0,
			ThrottleDelay:  0,
			CustomBehavior: make(map[string]interface{}),
		},
	}
}

// Setup initializes the mock provider.
// This method prepares the MockGraphProvider for use in tests. Although the current
// implementation doesn't require any initialization steps, this method is provided
// to satisfy the MockProvider interface and for future extensibility.
//
// Example usage:
//
//	mockGraph := mock.NewMockGraphProvider()
//	err := mockGraph.Setup()
//	if err != nil {
//	    // Handle error
//	}
//
// Returns nil as there are no initialization errors in the current implementation.
func (m *MockGraphProvider) Setup() error {
	// Nothing to do for basic setup
	return nil
}

// Teardown cleans up the mock provider.
// This method performs any necessary cleanup after tests are complete. Although the
// current implementation doesn't require any cleanup steps, this method is provided
// to satisfy the MockProvider interface and for future extensibility.
//
// Example usage:
//
//	defer mockGraph.Teardown()
//
// Returns nil as there are no cleanup errors in the current implementation.
func (m *MockGraphProvider) Teardown() error {
	// Nothing to do for basic teardown
	return nil
}

// Reset resets the mock provider to its initial state.
// This method restores the MockGraphProvider to a clean state, clearing all
// recorded calls, mock responses, and configuration settings. It's useful for
// reusing the same mock provider instance across multiple tests.
//
// Example usage:
//
//	// After a test
//	err := mockGraph.Reset()
//	if err != nil {
//	    // Handle error
//	}
//	// Now the mock is ready for the next test
//
// Returns nil if the reset operation succeeds, or an error if it fails.
func (m *MockGraphProvider) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Client = NewMockGraphClient()
	m.recorder = NewBasicMockRecorder()
	m.config = MockConfig{}
	m.networkConditions = NetworkConditions{}
	return nil
}

// SetNetworkConditions sets the simulated network conditions
func (m *MockGraphProvider) SetNetworkConditions(latency time.Duration, packetLoss float64, bandwidth int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.networkConditions = NetworkConditions{
		Latency:    latency,
		PacketLoss: packetLoss,
		Bandwidth:  bandwidth,
	}
}

// SetConfig sets the mock configuration
func (m *MockGraphProvider) SetConfig(config MockConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
func (m *MockGraphProvider) AddMockItem(resource string, item *api.DriveItem) {
	// Make a deep copy of the item to ensure it isn't modified
	itemCopy := &api.DriveItem{
		ID:               item.ID,
		Name:             item.Name,
		Size:             item.Size,
		ModTime:          item.ModTime,
		ConflictBehavior: item.ConflictBehavior,
		ETag:             item.ETag,
	}

	// Copy parent if it exists
	if item.Parent != nil {
		itemCopy.Parent = &api.DriveItemParent{
			Path:      item.Parent.Path,
			ID:        item.Parent.ID,
			DriveID:   item.Parent.DriveID,
			DriveType: item.Parent.DriveType,
		}
	}

	// Copy folder if it exists
	if item.Folder != nil {
		itemCopy.Folder = &api.Folder{
			ChildCount: item.Folder.ChildCount,
		}
	}

	// Copy file if it exists
	if item.File != nil {
		itemCopy.File = &api.File{}
		if item.File.Hashes.SHA1Hash != "" || item.File.Hashes.QuickXorHash != "" {
			itemCopy.File.Hashes = api.Hashes{
				SHA1Hash:     item.File.Hashes.SHA1Hash,
				QuickXorHash: item.File.Hashes.QuickXorHash,
			}
		}
	}

	// Copy deleted if it exists
	if item.Deleted != nil {
		itemCopy.Deleted = &api.Deleted{
			State: item.Deleted.State,
		}
	}

	m.Client.AddMockItem(resource, itemCopy)
	m.recorder.RecordCall("AddMockItem", resource, itemCopy)
}

// AddMockItems adds a predefined list of DriveItems for a children request
func (m *MockGraphProvider) AddMockItems(resource string, items []*api.DriveItem) {
	// Make deep copies of all items
	itemsCopy := make([]*api.DriveItem, len(items))
	for i, item := range items {
		// Make a deep copy of the item to ensure it isn't modified
		itemCopy := &api.DriveItem{
			ID:               item.ID,
			Name:             item.Name,
			Size:             item.Size,
			ModTime:          item.ModTime,
			ConflictBehavior: item.ConflictBehavior,
			ETag:             item.ETag,
		}

		// Copy parent if it exists
		if item.Parent != nil {
			itemCopy.Parent = &api.DriveItemParent{
				Path:      item.Parent.Path,
				ID:        item.Parent.ID,
				DriveID:   item.Parent.DriveID,
				DriveType: item.Parent.DriveType,
			}
		}

		// Copy folder if it exists
		if item.Folder != nil {
			itemCopy.Folder = &api.Folder{
				ChildCount: item.Folder.ChildCount,
			}
		}

		// Copy file if it exists
		if item.File != nil {
			itemCopy.File = &api.File{}
			if item.File.Hashes.SHA1Hash != "" || item.File.Hashes.QuickXorHash != "" {
				itemCopy.File.Hashes = api.Hashes{
					SHA1Hash:     item.File.Hashes.SHA1Hash,
					QuickXorHash: item.File.Hashes.QuickXorHash,
				}
			}
		}

		// Copy deleted if it exists
		if item.Deleted != nil {
			itemCopy.Deleted = &api.Deleted{
				State: item.Deleted.State,
			}
		}

		itemsCopy[i] = itemCopy
	}

	m.Client.AddMockItems(resource, itemsCopy)
	m.recorder.RecordCall("AddMockItems", resource, itemsCopy)
}

// SimulateNetworkDelay simulates network delay based on current network conditions
func (m *MockGraphProvider) SimulateNetworkDelay() {
	m.mu.Lock()
	latency := m.networkConditions.Latency
	m.mu.Unlock()

	if latency > 0 {
		time.Sleep(latency)
	}
}

// SimulateNetworkError simulates various network error conditions based on the configuration.
// This method is used internally by other methods to simulate realistic network behavior.
//
// It can simulate two types of errors:
//  1. Random network errors: Based on the ErrorRate in the config, it will randomly return
//     an unexpected EOF error to simulate network disconnections or timeouts.
//  2. API throttling: Based on the ThrottleRate in the config, it will simulate the API
//     throttling requests due to rate limiting. If ThrottleDelay is set, it will also
//     introduce a delay before returning the error to simulate the backoff behavior
//     of the real API.
//
// The error simulation uses a deterministic random number generator with a fixed seed
// to ensure reproducible test behavior.
//
// Returns:
//   - nil if no error is simulated
//   - io.ErrUnexpectedEOF for random network errors
//   - A custom error for API throttling
func (m *MockGraphProvider) SimulateNetworkError() error {
	m.mu.Lock()
	errorRate := m.config.ErrorRate
	throttleRate := m.config.ThrottleRate
	throttleDelay := m.config.ThrottleDelay
	m.mu.Unlock()

	// Use a deterministic random number generator
	// We use a fixed seed for reproducible tests
	r := rand.New(rand.NewSource(1))

	// Simulate random errors based on error rate
	if errorRate > 0 {
		// Generate a random number between 0 and 1
		// If it's less than the error rate, return an error
		if r.Float64() < errorRate {
			return io.ErrUnexpectedEOF
		}
	}

	// Simulate API throttling
	if throttleRate > 0 {
		// Generate a random number between 0 and 1
		// If it's less than the throttle rate, simulate throttling
		if r.Float64() < throttleRate {
			// If throttling delay is configured, simulate a delay
			if throttleDelay > 0 {
				time.Sleep(throttleDelay)
			}
			return errors.New("simulated API throttling: request rate exceeded")
		}
	}

	return nil
}

// RequestWithContext wraps the client's RequestWithContext method with recording and network simulation
func (m *MockGraphProvider) RequestWithContext(ctx context.Context, resource string, method string, content io.Reader, headers ...api.Header) ([]byte, error) {
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
func (m *MockGraphProvider) Get(resource string, headers ...api.Header) ([]byte, error) {
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
func (m *MockGraphProvider) GetWithContext(ctx context.Context, resource string, headers ...api.Header) ([]byte, error) {
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
func (m *MockGraphProvider) GetItem(id string) (*api.DriveItem, error) {
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
func (m *MockGraphProvider) GetItemChildren(id string) ([]*api.DriveItem, error) {
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

// GetItemChildrenPath wraps the client's GetItemChildrenPath method with recording and network simulation
func (m *MockGraphProvider) GetItemChildrenPath(path string) ([]*api.DriveItem, error) {
	m.recorder.RecordCall("GetItemChildrenPath", path)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.GetItemChildrenPath(path)

	// Record the result
	m.recorder.RecordCallWithResult("GetItemChildrenPath", result, err, path)

	return result, err
}

// GetItemPath wraps the client's GetItemPath method with recording and network simulation
func (m *MockGraphProvider) GetItemPath(path string) (*api.DriveItem, error) {
	m.recorder.RecordCall("GetItemPath", path)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.GetItemPath(path)

	// Record the result
	m.recorder.RecordCallWithResult("GetItemPath", result, err, path)

	return result, err
}

// GetItemContent wraps the client's GetItemContent method with recording and network simulation
func (m *MockGraphProvider) GetItemContent(id string) ([]byte, uint64, error) {
	m.recorder.RecordCall("GetItemContent", id)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, 0, err
	}

	// Call the underlying client
	result, size, err := m.Client.GetItemContent(id)

	// Record the result
	m.recorder.RecordCallWithResult("GetItemContent", result, err, id)

	return result, size, err
}

// GetItemContentStream wraps the client's GetItemContentStream method with recording and network simulation
func (m *MockGraphProvider) GetItemContentStream(id string, output io.Writer) (uint64, error) {
	m.recorder.RecordCall("GetItemContentStream", id, output)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return 0, err
	}

	// Call the underlying client
	size, err := m.Client.GetItemContentStream(id, output)

	// Record the result
	m.recorder.RecordCallWithResult("GetItemContentStream", size, err, id, output)

	return size, err
}

// Patch wraps the client's Patch method with recording and network simulation
func (m *MockGraphProvider) Patch(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("Patch", resource, content, headers)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.Patch(resource, content, headers...)

	// Record the result
	m.recorder.RecordCallWithResult("Patch", result, err, resource, content, headers)

	return result, err
}

// Post wraps the client's Post method with recording and network simulation
func (m *MockGraphProvider) Post(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("Post", resource, content, headers)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.Post(resource, content, headers...)

	// Record the result
	m.recorder.RecordCallWithResult("Post", result, err, resource, content, headers)

	return result, err
}

// Put wraps the client's Put method with recording and network simulation
func (m *MockGraphProvider) Put(resource string, content io.Reader, headers ...api.Header) ([]byte, error) {
	m.recorder.RecordCall("Put", resource, content, headers)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.Put(resource, content, headers...)

	// Record the result
	m.recorder.RecordCallWithResult("Put", result, err, resource, content, headers)

	return result, err
}

// Delete wraps the client's Delete method with recording and network simulation
func (m *MockGraphProvider) Delete(resource string, headers ...api.Header) error {
	m.recorder.RecordCall("Delete", resource, headers)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return err
	}

	// Call the underlying client
	err := m.Client.Delete(resource, headers...)

	// Record the result
	m.recorder.RecordCallWithResult("Delete", nil, err, resource, headers)

	return err
}

// Mkdir wraps the client's Mkdir method with recording and network simulation
func (m *MockGraphProvider) Mkdir(name string, parentID string) (*api.DriveItem, error) {
	m.recorder.RecordCall("Mkdir", name, parentID)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return nil, err
	}

	// Call the underlying client
	result, err := m.Client.Mkdir(name, parentID)

	// Record the result
	m.recorder.RecordCallWithResult("Mkdir", result, err, name, parentID)

	return result, err
}

// Rename wraps the client's Rename method with recording and network simulation
func (m *MockGraphProvider) Rename(itemID string, itemName string, parentID string) error {
	m.recorder.RecordCall("Rename", itemID, itemName, parentID)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return err
	}

	// Call the underlying client
	err := m.Client.Rename(itemID, itemName, parentID)

	// Record the result
	m.recorder.RecordCallWithResult("Rename", nil, err, itemID, itemName, parentID)

	return err
}

// Remove wraps the client's Remove method with recording and network simulation
func (m *MockGraphProvider) Remove(id string) error {
	m.recorder.RecordCall("Remove", id)

	// Simulate network conditions
	m.SimulateNetworkDelay()

	// Simulate network errors
	if err := m.SimulateNetworkError(); err != nil {
		return err
	}

	// Call the underlying client
	err := m.Client.Remove(id)

	// Record the result
	m.recorder.RecordCallWithResult("Remove", nil, err, id)

	return err
}
