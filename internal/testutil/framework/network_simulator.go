// Package framework provides testing utilities for the OneMount project.
//
// This file implements the NetworkSimulator interface for simulating different network
// conditions during testing. The NetworkSimulator is a key component of the OneMount
// test framework that allows simulating different network conditions for testing.
// This is particularly important for testing a filesystem that interacts with a cloud
// service like OneDrive, as network conditions can significantly impact the behavior
// and performance of the system.
package framework

import (
	"errors"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/graph/mock"
)

// NetworkCondition represents a specific network condition configuration.
// It encapsulates various network characteristics that can be simulated during testing.
//
// Fields:
//   - Name: A short identifier for the network condition
//   - Latency: The delay in data transmission over the network
//   - PacketLoss: The failure rate of data packets (0.0 to 1.0, where 1.0 = 100%)
//   - Bandwidth: The maximum data transfer rate in Kbps
//   - Description: A human-readable description of the network condition
type NetworkCondition struct {
	Name        string
	Latency     time.Duration
	PacketLoss  float64
	Bandwidth   int // in Kbps
	Description string
}

// Common network condition presets
var (
	// FastNetwork represents a fast, reliable network connection
	FastNetwork = NetworkCondition{
		Name:        "Fast",
		Latency:     10 * time.Millisecond,
		PacketLoss:  0.0,
		Bandwidth:   100000, // 100 Mbps
		Description: "Fast, reliable network connection",
	}

	// AverageNetwork represents an average home broadband connection
	AverageNetwork = NetworkCondition{
		Name:        "Average",
		Latency:     50 * time.Millisecond,
		PacketLoss:  0.01,  // 1% packet loss
		Bandwidth:   20000, // 20 Mbps
		Description: "Average home broadband connection",
	}

	// SlowNetwork represents a slow network connection
	SlowNetwork = NetworkCondition{
		Name:        "Slow",
		Latency:     200 * time.Millisecond,
		PacketLoss:  0.05, // 5% packet loss
		Bandwidth:   1000, // 1 Mbps
		Description: "Slow network connection",
	}

	// MobileNetwork represents a mobile data connection
	MobileNetwork = NetworkCondition{
		Name:        "Mobile",
		Latency:     100 * time.Millisecond,
		PacketLoss:  0.02, // 2% packet loss
		Bandwidth:   5000, // 5 Mbps
		Description: "Mobile data connection",
	}

	// IntermittentConnection represents an unstable connection
	IntermittentConnection = NetworkCondition{
		Name:        "Intermittent",
		Latency:     300 * time.Millisecond,
		PacketLoss:  0.15, // 15% packet loss
		Bandwidth:   2000, // 2 Mbps
		Description: "Unstable connection with high packet loss",
	}

	// SatelliteConnection represents a high-latency satellite connection
	SatelliteConnection = NetworkCondition{
		Name:        "Satellite",
		Latency:     700 * time.Millisecond,
		PacketLoss:  0.03,  // 3% packet loss
		Bandwidth:   10000, // 10 Mbps
		Description: "High-latency satellite connection",
	}

	// Disconnected represents a network disconnection
	Disconnected = NetworkCondition{
		Name:        "Disconnected",
		Latency:     0,
		PacketLoss:  1.0, // 100% packet loss
		Bandwidth:   0,
		Description: "Network disconnection",
	}
)

// NetworkSimulator simulates different network conditions for testing.
// This interface defines methods for controlling and simulating various network conditions
// during tests, allowing developers to test how their code behaves under different
// network scenarios such as high latency, packet loss, limited bandwidth, or complete
// disconnection.
type NetworkSimulator interface {
	// SetConditions sets the network conditions with the specified parameters.
	//
	// Parameters:
	//   - latency: The delay in data transmission (in time.Duration)
	//   - packetLoss: The failure rate of data packets (0.0 to 1.0, where 1.0 = 100%)
	//   - bandwidth: The maximum data transfer rate in Kbps
	//
	// Returns an error if the parameters are invalid.
	SetConditions(latency time.Duration, packetLoss float64, bandwidth int) error

	// ApplyPreset applies a predefined network condition preset.
	// This is a convenience method for applying common network scenarios.
	//
	// Parameter:
	//   - preset: A predefined NetworkCondition (e.g., FastNetwork, SlowNetwork)
	//
	// Returns an error if the preset could not be applied.
	ApplyPreset(preset NetworkCondition) error

	// GetCurrentConditions returns the current network conditions.
	// This can be used to check what network conditions are currently being simulated.
	GetCurrentConditions() NetworkCondition

	// Disconnect simulates a network disconnection.
	// This sets the network to a completely disconnected state (100% packet loss).
	//
	// Returns an error if the network is already disconnected.
	Disconnect() error

	// Reconnect restores the network connection.
	// This restores the network to the conditions that were in effect before disconnection.
	//
	// Returns an error if the network is already connected.
	Reconnect() error

	// IsConnected returns whether the network is currently connected.
	// Returns true if the network is connected, false if it is disconnected.
	IsConnected() bool

	// SimulateNetworkDelay simulates network delay based on current conditions.
	// This method blocks for the duration specified by the current latency setting.
	SimulateNetworkDelay()

	// SimulatePacketLoss returns true if a packet should be dropped based on current conditions.
	// This can be used to simulate packet loss in network operations.
	//
	// Returns true if the packet should be dropped, false otherwise.
	SimulatePacketLoss() bool

	// SimulateNetworkError returns an error if network conditions warrant it.
	// This can be used to simulate network errors in operations.
	//
	// Returns an error if the network is disconnected or if a packet is lost,
	// nil otherwise.
	SimulateNetworkError() error

	// RegisterProvider registers a mock provider to apply network conditions to.
	// This allows the network simulator to affect the behavior of mock providers.
	//
	// Parameter:
	//   - provider: The mock provider to register
	RegisterProvider(provider MockProvider)

	// GetRegisteredProviders returns all registered mock providers.
	// This can be used to check which providers are currently affected by the network simulator.
	//
	// Returns a slice of all registered mock providers.
	GetRegisteredProviders() []MockProvider

	// SimulateIntermittentConnection simulates an unstable connection by alternating
	// between connected and disconnected states.
	//
	// Parameters:
	//   - disconnectDuration: How long the network should remain disconnected
	//   - connectDuration: How long the network should remain connected
	//
	// Returns an error if the simulation could not be started.
	SimulateIntermittentConnection(disconnectDuration, connectDuration time.Duration) error
}

// DefaultNetworkSimulator is the default implementation of NetworkSimulator.
// It provides a concrete implementation of all the methods defined in the NetworkSimulator
// interface, allowing for simulation of various network conditions during testing.
//
// The DefaultNetworkSimulator maintains the current network conditions, connection state,
// and a list of registered mock providers that are affected by the network conditions.
// It uses a mutex to ensure thread safety when accessing or modifying its state.
type DefaultNetworkSimulator struct {
	// Current network conditions
	currentCondition NetworkCondition

	// Previous network conditions (for reconnection)
	previousCondition NetworkCondition

	// Whether the network is currently connected
	connected bool

	// Registered mock providers
	providers []MockProvider

	// Mutex for thread safety
	mu sync.Mutex
}

// NewNetworkSimulator creates a new NetworkSimulator with default settings.
// The simulator is initialized with FastNetwork conditions and a connected state.
//
// Example usage:
//
//	simulator := NewNetworkSimulator()
//	err := simulator.ApplyPreset(SlowNetwork)
//	if err != nil {
//	    // Handle error
//	}
//	simulator.SimulateNetworkDelay() // Blocks for the duration of the latency
//
// Returns a new DefaultNetworkSimulator instance ready for use in tests.
func NewNetworkSimulator() *DefaultNetworkSimulator {
	return &DefaultNetworkSimulator{
		currentCondition:  FastNetwork,
		previousCondition: FastNetwork,
		connected:         true,
		providers:         make([]MockProvider, 0),
	}
}

// SetConditions sets the network conditions with the specified parameters.
// This method allows for fine-grained control over the network simulation by
// setting specific values for latency, packet loss, and bandwidth.
//
// Parameters:
//   - latency: The delay in data transmission (in time.Duration)
//   - packetLoss: The failure rate of data packets (0.0 to 1.0, where 1.0 = 100%)
//   - bandwidth: The maximum data transfer rate in Kbps
//
// Example usage:
//
//	// Set a high-latency, low-bandwidth connection with moderate packet loss
//	err := simulator.SetConditions(500*time.Millisecond, 0.05, 1000)
//	if err != nil {
//	    // Handle error
//	}
//
// Returns an error if any of the parameters are invalid:
//   - latency must be non-negative
//   - packetLoss must be between 0 and 1
//   - bandwidth must be non-negative
//
// The method also updates all registered mock providers with the new network conditions.
func (s *DefaultNetworkSimulator) SetConditions(latency time.Duration, packetLoss float64, bandwidth int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input parameters
	if latency < 0 {
		return errors.New("latency must be non-negative")
	}
	if packetLoss < 0 || packetLoss > 1 {
		return errors.New("packet loss must be between 0 and 1")
	}
	if bandwidth < 0 {
		return errors.New("bandwidth must be non-negative")
	}

	// Update current conditions
	s.currentCondition = NetworkCondition{
		Name:        "Custom",
		Latency:     latency,
		PacketLoss:  packetLoss,
		Bandwidth:   bandwidth,
		Description: "Custom network condition",
	}

	// Update all registered providers
	for _, provider := range s.providers {
		if graphProvider, ok := provider.(*mock.MockGraphProvider); ok {
			graphProvider.SetNetworkConditions(latency, packetLoss, bandwidth)
		}
	}

	return nil
}

// ApplyPreset applies a predefined network condition preset.
// This is a convenience method for applying common network scenarios without
// having to specify individual latency, packet loss, and bandwidth values.
//
// Parameter:
//   - preset: A predefined NetworkCondition (e.g., FastNetwork, SlowNetwork, MobileNetwork)
//
// Example usage:
//
//	// Apply the SlowNetwork preset
//	err := simulator.ApplyPreset(SlowNetwork)
//	if err != nil {
//	    // Handle error
//	}
//
//	// Apply the MobileNetwork preset
//	err = simulator.ApplyPreset(MobileNetwork)
//	if err != nil {
//	    // Handle error
//	}
//
// Available presets include:
//   - FastNetwork: Fast, reliable network (10ms latency, 0% packet loss, 100Mbps)
//   - AverageNetwork: Average home broadband (50ms latency, 1% packet loss, 20Mbps)
//   - SlowNetwork: Slow connection (200ms latency, 5% packet loss, 1Mbps)
//   - MobileNetwork: Mobile data (100ms latency, 2% packet loss, 5Mbps)
//   - IntermittentConnection: Unstable connection (300ms latency, 15% packet loss, 2Mbps)
//   - SatelliteConnection: High-latency satellite (700ms latency, 3% packet loss, 10Mbps)
//
// Returns an error if the preset could not be applied (delegates to SetConditions).
func (s *DefaultNetworkSimulator) ApplyPreset(preset NetworkCondition) error {
	return s.SetConditions(preset.Latency, preset.PacketLoss, preset.Bandwidth)
}

// GetCurrentConditions returns the current network conditions
func (s *DefaultNetworkSimulator) GetCurrentConditions() NetworkCondition {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentCondition
}

// Disconnect simulates a network disconnection.
// This method sets the network to a completely disconnected state (100% packet loss),
// which is useful for testing how the system behaves when the network is unavailable.
//
// The current network conditions are saved so they can be restored when Reconnect is called.
// All registered mock providers are updated to reflect the disconnected state.
//
// Example usage:
//
//	// Disconnect the network
//	err := simulator.Disconnect()
//	if err != nil {
//	    // Handle error
//	}
//
//	// Test behavior when network is disconnected
//	// ...
//
//	// Reconnect the network
//	err = simulator.Reconnect()
//	if err != nil {
//	    // Handle error
//	}
//
// Returns an error if the network is already disconnected.
func (s *DefaultNetworkSimulator) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return errors.New("network is already disconnected")
	}

	// Save current conditions for later reconnection
	s.previousCondition = s.currentCondition

	// Set disconnected state
	s.connected = false
	s.currentCondition = Disconnected

	// Update all registered providers
	for _, provider := range s.providers {
		if graphProvider, ok := provider.(*mock.MockGraphProvider); ok {
			graphProvider.SetNetworkConditions(0, 1.0, 0)
		}
	}

	return nil
}

// Reconnect restores the network connection.
// This method restores the network to the conditions that were in effect before
// the Disconnect method was called. This is useful for testing how the system
// recovers when the network becomes available again after a disconnection.
//
// All registered mock providers are updated to reflect the restored network conditions.
//
// Example usage:
//
//	// Disconnect the network
//	err := simulator.Disconnect()
//	if err != nil {
//	    // Handle error
//	}
//
//	// Test behavior when network is disconnected
//	// ...
//
//	// Reconnect the network
//	err = simulator.Reconnect()
//	if err != nil {
//	    // Handle error
//	}
//
//	// Test behavior when network is reconnected
//	// ...
//
// Returns an error if the network is already connected.
func (s *DefaultNetworkSimulator) Reconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected {
		return errors.New("network is already connected")
	}

	// Restore previous conditions
	s.connected = true
	s.currentCondition = s.previousCondition

	// Update all registered providers
	for _, provider := range s.providers {
		if graphProvider, ok := provider.(*mock.MockGraphProvider); ok {
			graphProvider.SetNetworkConditions(
				s.currentCondition.Latency,
				s.currentCondition.PacketLoss,
				s.currentCondition.Bandwidth,
			)
		}
	}

	return nil
}

// IsConnected returns whether the network is currently connected
func (s *DefaultNetworkSimulator) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connected
}

// SimulateNetworkDelay simulates network delay based on current conditions
func (s *DefaultNetworkSimulator) SimulateNetworkDelay() {
	s.mu.Lock()
	condition := s.currentCondition
	s.mu.Unlock()

	if condition.Latency > 0 {
		time.Sleep(condition.Latency)
	}
}

// SimulatePacketLoss returns true if a packet should be dropped based on current conditions
func (s *DefaultNetworkSimulator) SimulatePacketLoss() bool {
	s.mu.Lock()
	packetLoss := s.currentCondition.PacketLoss
	s.mu.Unlock()

	if packetLoss <= 0 {
		return false
	}

	// Generate a random number between 0 and 1
	// If it's less than the packet loss rate, drop the packet
	return float64(time.Now().UnixNano()%100)/100 < packetLoss
}

// SimulateNetworkError returns an error if network conditions warrant it
func (s *DefaultNetworkSimulator) SimulateNetworkError() error {
	s.mu.Lock()
	connected := s.connected
	packetLoss := s.currentCondition.PacketLoss
	s.mu.Unlock()

	if !connected {
		return errors.New("network is disconnected")
	}

	// Simulate packet loss without holding the lock
	if packetLoss > 0 && float64(time.Now().UnixNano()%100)/100 < packetLoss {
		return errors.New("packet lost due to network conditions")
	}

	return nil
}

// RegisterProvider registers a mock provider to apply network conditions to.
// This method allows the network simulator to affect the behavior of mock providers,
// such as the MockGraphProvider, by applying the current network conditions to them.
// This is particularly useful for testing how the system interacts with external
// services under different network conditions.
//
// The method checks if the provider is already registered to avoid duplicates.
// If the provider is a MockGraphProvider, the current network conditions are
// immediately applied to it.
//
// Parameter:
//   - provider: The mock provider to register with the network simulator
//
// Example usage:
//
//	// Create a mock graph provider
//	graphProvider := mock.NewMockGraphProvider()
//
//	// Register it with the network simulator
//	simulator.RegisterProvider(graphProvider)
//
//	// Now when network conditions change, the graph provider will be affected
//	simulator.ApplyPreset(SlowNetwork)
//
//	// Use the graph provider in tests
//	// ...
func (s *DefaultNetworkSimulator) RegisterProvider(provider MockProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if provider is already registered
	for _, p := range s.providers {
		if p == provider {
			return
		}
	}

	s.providers = append(s.providers, provider)

	// Apply current network conditions to the provider
	if graphProvider, ok := provider.(*mock.MockGraphProvider); ok {
		graphProvider.SetNetworkConditions(
			s.currentCondition.Latency,
			s.currentCondition.PacketLoss,
			s.currentCondition.Bandwidth,
		)
	}
}

// GetRegisteredProviders returns all registered mock providers
func (s *DefaultNetworkSimulator) GetRegisteredProviders() []MockProvider {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy of the providers slice to prevent modification
	result := make([]MockProvider, len(s.providers))
	copy(result, s.providers)
	return result
}

// SimulateIntermittentConnection simulates an unstable connection by alternating
// between connected and disconnected states at the specified intervals.
// This is useful for testing how the system behaves under unstable network conditions,
// such as mobile networks with spotty coverage or connections that periodically drop.
//
// The method will start a background goroutine that alternates between connected and
// disconnected states according to the specified durations. This allows tests to
// continue running while the network conditions change in the background.
//
// Parameters:
//   - disconnectDuration: how long the network should remain disconnected in each cycle
//   - connectDuration: how long the network should remain connected in each cycle
//
// Example usage:
//
//	// Simulate a connection that drops for 2 seconds every 5 seconds
//	err := simulator.SimulateIntermittentConnection(2*time.Second, 5*time.Second)
//	if err != nil {
//	    // Handle error
//	}
//
//	// Run tests with the intermittent connection
//	// ...
//
// Returns an error if the simulation could not be started, such as if the
// network is already in an intermittent state or if the durations are invalid.
//
// Note: This is a stub implementation that will be completed in a future update.
func (s *DefaultNetworkSimulator) SimulateIntermittentConnection(disconnectDuration, connectDuration time.Duration) error {
	// This is a stub implementation that will be completed later
	return errors.New("SimulateIntermittentConnection not implemented yet")
}
