// Package testutil provides testing utilities for the OneMount project.
package mock

import (
	"time"
)

// UIState represents the current state of a UI component
type UIState struct {
	Visible    bool
	Enabled    bool
	Text       string
	Properties map[string]interface{}
}

// UIEvent represents a user interaction with a UI component
type UIEvent struct {
	Type      string // e.g., "click", "input", "focus", "blur"
	Target    string
	Timestamp time.Time
	Data      map[string]interface{}
}

// UIResponse represents a response to a UI interaction
type UIResponse struct {
	Success bool
	Data    interface{}
	Error   error
}

// MockUIProvider implements the MockProvider interface for simulating UI interactions
type MockUIProvider struct {
	// Simulated UI state
	state map[string]UIState

	// Record of UI events
	events []UIEvent

	// Configured responses to UI interactions
	responses map[string]UIResponse

	// Mock recorder for verification
	recorder MockRecorder

	// Configuration for mock behavior
	config MockConfig
}

// NewMockUIProvider creates a new MockUIProvider
func NewMockUIProvider() *MockUIProvider {
	return &MockUIProvider{
		state:     make(map[string]UIState),
		events:    make([]UIEvent, 0),
		responses: make(map[string]UIResponse),
		recorder:  NewBasicMockRecorder(),
		config:    MockConfig{},
	}
}

// Setup initializes the mock provider
func (m *MockUIProvider) Setup() error {
	// Nothing to do for basic setup
	return nil
}

// Teardown cleans up the mock provider
func (m *MockUIProvider) Teardown() error {
	// Nothing to do for basic teardown
	return nil
}

// Reset resets the mock provider to its initial state
func (m *MockUIProvider) Reset() error {
	m.state = make(map[string]UIState)
	m.events = make([]UIEvent, 0)
	m.responses = make(map[string]UIResponse)
	m.recorder = NewBasicMockRecorder()
	m.config = MockConfig{}
	return nil
}

// SetConfig sets the mock configuration
func (m *MockUIProvider) SetConfig(config MockConfig) {
	m.config = config
}

// GetRecorder returns the mock recorder
func (m *MockUIProvider) GetRecorder() MockRecorder {
	return m.recorder
}

// SetComponentState sets the state of a UI component
func (m *MockUIProvider) SetComponentState(componentID string, state UIState) {
	m.recorder.RecordCall("SetComponentState", componentID, state)
	m.state[componentID] = state
}

// GetComponentState gets the state of a UI component
func (m *MockUIProvider) GetComponentState(componentID string) (UIState, bool) {
	m.recorder.RecordCall("GetComponentState", componentID)
	state, exists := m.state[componentID]
	m.recorder.RecordCallWithResult("GetComponentState", state, nil, componentID)
	return state, exists
}

// AddUIResponse adds a predefined response for a specific UI interaction
func (m *MockUIProvider) AddUIResponse(eventType string, target string, response UIResponse) {
	m.recorder.RecordCall("AddUIResponse", eventType, target, response)
	key := eventType + ":" + target
	m.responses[key] = response
}

// SimulateEvent simulates a UI event and returns the response
func (m *MockUIProvider) SimulateEvent(eventType string, target string, data map[string]interface{}) UIResponse {
	m.recorder.RecordCall("SimulateEvent", eventType, target, data)

	// Create the event
	event := UIEvent{
		Type:      eventType,
		Target:    target,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Record the event
	m.events = append(m.events, event)

	// Check if we have a predefined response
	key := eventType + ":" + target
	if response, exists := m.responses[key]; exists {
		m.recorder.RecordCallWithResult("SimulateEvent", response, nil, eventType, target, data)
		return response
	}

	// Default response
	defaultResponse := UIResponse{
		Success: true,
		Data:    nil,
		Error:   nil,
	}

	m.recorder.RecordCallWithResult("SimulateEvent", defaultResponse, nil, eventType, target, data)
	return defaultResponse
}

// GetEvents returns all recorded UI events
func (m *MockUIProvider) GetEvents() []UIEvent {
	return m.events
}

// GetEventsByType returns all recorded UI events of a specific type
func (m *MockUIProvider) GetEventsByType(eventType string) []UIEvent {
	var filteredEvents []UIEvent
	for _, event := range m.events {
		if event.Type == eventType {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return filteredEvents
}

// GetEventsByTarget returns all recorded UI events for a specific target
func (m *MockUIProvider) GetEventsByTarget(target string) []UIEvent {
	var filteredEvents []UIEvent
	for _, event := range m.events {
		if event.Target == target {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return filteredEvents
}

// VerifyEventOccurred verifies that a specific event occurred
func (m *MockUIProvider) VerifyEventOccurred(eventType string, target string) bool {
	for _, event := range m.events {
		if event.Type == eventType && event.Target == target {
			return true
		}
	}
	return false
}

// VerifyEventOccurredTimes verifies that a specific event occurred a specific number of times
func (m *MockUIProvider) VerifyEventOccurredTimes(eventType string, target string, times int) bool {
	count := 0
	for _, event := range m.events {
		if event.Type == eventType && event.Target == target {
			count++
		}
	}
	return count == times
}

// SimulateClick simulates a click event on a UI component
func (m *MockUIProvider) SimulateClick(target string) UIResponse {
	return m.SimulateEvent("click", target, nil)
}

// SimulateInput simulates an input event on a UI component
func (m *MockUIProvider) SimulateInput(target string, value string) UIResponse {
	data := map[string]interface{}{
		"value": value,
	}
	return m.SimulateEvent("input", target, data)
}

// SimulateFocus simulates a focus event on a UI component
func (m *MockUIProvider) SimulateFocus(target string) UIResponse {
	return m.SimulateEvent("focus", target, nil)
}

// SimulateBlur simulates a blur event on a UI component
func (m *MockUIProvider) SimulateBlur(target string) UIResponse {
	return m.SimulateEvent("blur", target, nil)
}

// SimulateKeyPress simulates a keypress event on a UI component
func (m *MockUIProvider) SimulateKeyPress(target string, key string) UIResponse {
	data := map[string]interface{}{
		"key": key,
	}
	return m.SimulateEvent("keypress", target, data)
}

// SimulateDragAndDrop simulates a drag and drop event
func (m *MockUIProvider) SimulateDragAndDrop(source string, target string) UIResponse {
	data := map[string]interface{}{
		"source": source,
		"target": target,
	}
	return m.SimulateEvent("dragdrop", target, data)
}
