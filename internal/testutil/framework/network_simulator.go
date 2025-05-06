// network_simulator.go implements the NetworkSimulator interface for simulating network conditions
package framework

import (
	"errors"
	"github.com/auriora/onemount/internal/testutil/mock"
	"sync"
	"time"
)

// NetworkCondition represents a specific network condition configuration
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

// NetworkSimulator simulates different network conditions for testing
type NetworkSimulator interface {
	// SetConditions sets the network conditions
	SetConditions(latency time.Duration, packetLoss float64, bandwidth int) error

	// ApplyPreset applies a predefined network condition preset
	ApplyPreset(preset NetworkCondition) error

	// GetCurrentConditions returns the current network conditions
	GetCurrentConditions() NetworkCondition

	// Disconnect simulates a network disconnection
	Disconnect() error

	// Reconnect restores the network connection
	Reconnect() error

	// IsConnected returns whether the network is currently connected
	IsConnected() bool

	// SimulateNetworkDelay simulates network delay based on current conditions
	SimulateNetworkDelay()

	// SimulatePacketLoss returns true if a packet should be dropped based on current conditions
	SimulatePacketLoss() bool

	// SimulateNetworkError returns an error if network conditions warrant it
	SimulateNetworkError() error

	// RegisterProvider registers a mock provider to apply network conditions to
	RegisterProvider(provider MockProvider)

	// GetRegisteredProviders returns all registered mock providers
	GetRegisteredProviders() []MockProvider
}

// DefaultNetworkSimulator is the default implementation of NetworkSimulator
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

// NewNetworkSimulator creates a new NetworkSimulator with default settings
func NewNetworkSimulator() *DefaultNetworkSimulator {
	return &DefaultNetworkSimulator{
		currentCondition:  FastNetwork,
		previousCondition: FastNetwork,
		connected:         true,
		providers:         make([]MockProvider, 0),
	}
}

// SetConditions sets the network conditions
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

// ApplyPreset applies a predefined network condition preset
func (s *DefaultNetworkSimulator) ApplyPreset(preset NetworkCondition) error {
	return s.SetConditions(preset.Latency, preset.PacketLoss, preset.Bandwidth)
}

// GetCurrentConditions returns the current network conditions
func (s *DefaultNetworkSimulator) GetCurrentConditions() NetworkCondition {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentCondition
}

// Disconnect simulates a network disconnection
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

// Reconnect restores the network connection
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

// RegisterProvider registers a mock provider to apply network conditions to
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
