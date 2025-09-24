// Package mock provides mock implementations for testing.
package mock

import (
	"github.com/auriora/onemount/internal/graph/api"
	"github.com/auriora/onemount/internal/graph/mock"
	graphmock "github.com/auriora/onemount/internal/graph/mock"
)

// NewMockGraphProvider creates a new APIGraphProvider that implements the api.GraphProvider interface.
// This function is provided for backward compatibility with existing code.
func NewMockGraphProvider() *mock.APIGraphProvider {
	return mock.NewAPIGraphProvider()
}

// NetworkConditions is provided for backward compatibility.
type NetworkConditions struct {
	Latency    interface{} // Not used, kept for backward compatibility
	PacketLoss interface{} // Not used, kept for backward compatibility
	Bandwidth  interface{} // Not used, kept for backward compatibility
}

// MockConfig is provided for backward compatibility.
type MockConfig struct {
	Latency        interface{}            // Not used, kept for backward compatibility
	ErrorRate      interface{}            // Not used, kept for backward compatibility
	ResponseDelay  interface{}            // Not used, kept for backward compatibility
	ThrottleRate   interface{}            // Not used, kept for backward compatibility
	ThrottleDelay  interface{}            // Not used, kept for backward compatibility
	CustomBehavior map[string]interface{} // Not used, kept for backward compatibility
}

// Re-export NewBasicMockRecorder for backward compatibility
var NewBasicMockRecorder = mock.NewBasicMockRecorder

// MockGraphProvider is provided for backward compatibility.
// It wraps the MockGraphProvider from pkg/graph/mock.
type MockGraphProvider struct {
	provider *graphmock.MockGraphProvider
}

// Setup initializes the provider.
func (m *MockGraphProvider) Setup() error {
	return m.provider.Setup()
}

// Teardown cleans up the provider.
func (m *MockGraphProvider) Teardown() error {
	return m.provider.Teardown()
}

// Reset resets the provider to its initial state.
func (m *MockGraphProvider) Reset() error {
	return m.provider.Reset()
}

// GetRecorder returns the mock recorder.
func (m *MockGraphProvider) GetRecorder() api.MockRecorder {
	return m.provider.GetRecorder()
}

// SetNetworkConditions sets the simulated network conditions.
// This method is provided for backward compatibility but doesn't do anything.
func (m *MockGraphProvider) SetNetworkConditions(latency interface{}, packetLoss interface{}, bandwidth interface{}) {
	// No-op for backward compatibility
}

// SetConfig sets the mock configuration.
// This method is provided for backward compatibility but doesn't do anything.
func (m *MockGraphProvider) SetConfig(config MockConfig) {
	// No-op for backward compatibility
}

// AddMockResponse adds a predefined response for a specific resource path.
func (m *MockGraphProvider) AddMockResponse(resource string, body []byte, statusCode int, err error) {
	m.provider.AddMockResponse(resource, body, statusCode, err)
}

// AddMockItem adds a predefined DriveItem response for a specific resource path.
func (m *MockGraphProvider) AddMockItem(resource string, item *api.DriveItem) {
	m.provider.AddMockItem(resource, item)
}

// AddMockItems adds a predefined collection of DriveItems for a specific resource path.
func (m *MockGraphProvider) AddMockItems(resource string, items []*api.DriveItem) {
	m.provider.AddMockItems(resource, items)
}
