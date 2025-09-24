// network_simulator_test.go contains tests for the NetworkSimulator implementation
package framework

import (
	"testing"
	"time"

	"github.com/auriora/onemount/pkg/graph/mock"

	"github.com/stretchr/testify/assert"
)

func TestNetworkSimulator_SetConditions(t *testing.T) {
	simulator := NewNetworkSimulator()

	// Test valid conditions
	err := simulator.SetConditions(100*time.Millisecond, 0.1, 1000)
	assert.NoError(t, err)

	conditions := simulator.GetCurrentConditions()
	assert.Equal(t, 100*time.Millisecond, conditions.Latency)
	assert.Equal(t, 0.1, conditions.PacketLoss)
	assert.Equal(t, 1000, conditions.Bandwidth)

	// Test invalid conditions
	err = simulator.SetConditions(-100*time.Millisecond, 0.1, 1000)
	assert.Error(t, err)

	err = simulator.SetConditions(100*time.Millisecond, -0.1, 1000)
	assert.Error(t, err)

	err = simulator.SetConditions(100*time.Millisecond, 1.1, 1000)
	assert.Error(t, err)

	err = simulator.SetConditions(100*time.Millisecond, 0.1, -1000)
	assert.Error(t, err)
}

func TestNetworkSimulator_ApplyPreset(t *testing.T) {
	simulator := NewNetworkSimulator()

	// Test applying presets
	err := simulator.ApplyPreset(FastNetwork)
	assert.NoError(t, err)
	conditions := simulator.GetCurrentConditions()
	assert.Equal(t, FastNetwork.Latency, conditions.Latency)
	assert.Equal(t, FastNetwork.PacketLoss, conditions.PacketLoss)
	assert.Equal(t, FastNetwork.Bandwidth, conditions.Bandwidth)

	err = simulator.ApplyPreset(SlowNetwork)
	assert.NoError(t, err)
	conditions = simulator.GetCurrentConditions()
	assert.Equal(t, SlowNetwork.Latency, conditions.Latency)
	assert.Equal(t, SlowNetwork.PacketLoss, conditions.PacketLoss)
	assert.Equal(t, SlowNetwork.Bandwidth, conditions.Bandwidth)

	err = simulator.ApplyPreset(IntermittentConnection)
	assert.NoError(t, err)
	conditions = simulator.GetCurrentConditions()
	assert.Equal(t, IntermittentConnection.Latency, conditions.Latency)
	assert.Equal(t, IntermittentConnection.PacketLoss, conditions.PacketLoss)
	assert.Equal(t, IntermittentConnection.Bandwidth, conditions.Bandwidth)
}

func TestNetworkSimulator_DisconnectReconnect(t *testing.T) {
	simulator := NewNetworkSimulator()

	// Test initial state
	assert.True(t, simulator.IsConnected())

	// Test disconnect
	err := simulator.Disconnect()
	assert.NoError(t, err)
	assert.False(t, simulator.IsConnected())

	// Test disconnect when already disconnected
	err = simulator.Disconnect()
	assert.Error(t, err)

	// Test reconnect
	err = simulator.Reconnect()
	assert.NoError(t, err)
	assert.True(t, simulator.IsConnected())

	// Test reconnect when already connected
	err = simulator.Reconnect()
	assert.Error(t, err)
}

func TestNetworkSimulator_SimulateNetworkDelay(t *testing.T) {
	simulator := NewNetworkSimulator()

	// Set a small latency for testing
	err := simulator.SetConditions(10*time.Millisecond, 0, 1000)
	assert.NoError(t, err)

	// Measure the time it takes to simulate network delay
	start := time.Now()
	simulator.SimulateNetworkDelay()
	elapsed := time.Since(start)

	// The elapsed time should be at least the latency
	assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
}

func TestNetworkSimulator_SimulatePacketLoss(t *testing.T) {
	simulator := NewNetworkSimulator()

	// Set 100% packet loss for testing
	err := simulator.SetConditions(0, 1.0, 1000)
	assert.NoError(t, err)

	// With 100% packet loss, SimulatePacketLoss should always return true
	assert.True(t, simulator.SimulatePacketLoss())

	// Set 0% packet loss for testing
	err = simulator.SetConditions(0, 0.0, 1000)
	assert.NoError(t, err)

	// With 0% packet loss, SimulatePacketLoss should always return false
	assert.False(t, simulator.SimulatePacketLoss())
}

func TestNetworkSimulator_SimulateNetworkError(t *testing.T) {
	simulator := NewNetworkSimulator()

	// Test with connected network and no packet loss
	err := simulator.SetConditions(0, 0.0, 1000)
	assert.NoError(t, err)
	err = simulator.SimulateNetworkError()
	assert.NoError(t, err)

	// Test with disconnected network
	err = simulator.Disconnect()
	assert.NoError(t, err)
	err = simulator.SimulateNetworkError()
	assert.Error(t, err)
	assert.Equal(t, "network is disconnected", err.Error())

	// Reconnect for next test
	err = simulator.Reconnect()
	assert.NoError(t, err)

	// Test with 100% packet loss
	err = simulator.SetConditions(0, 1.0, 1000)
	assert.NoError(t, err)
	err = simulator.SimulateNetworkError()
	assert.Error(t, err)
	assert.Equal(t, "packet lost due to network conditions", err.Error())
}

func TestNetworkSimulator_RegisterProvider(t *testing.T) {
	simulator := NewNetworkSimulator()
	mockProvider := mock.NewMockGraphProvider()

	// Register the provider
	simulator.RegisterProvider(mockProvider)

	// Check that the provider was registered
	providers := simulator.GetRegisteredProviders()
	assert.Len(t, providers, 1)
	assert.Equal(t, mockProvider, providers[0])

	// Register the same provider again
	simulator.RegisterProvider(mockProvider)

	// Check that the provider was not registered twice
	providers = simulator.GetRegisteredProviders()
	assert.Len(t, providers, 1)
}

// Example of using NetworkSimulator with TestFramework
func ExampleNetworkSimulator() {
	// Create a test framework
	framework := NewTestFramework(TestConfig{}, nil)

	// Register a mock provider
	mockGraph := mock.NewMockGraphProvider()
	framework.RegisterMockProvider("graph", mockGraph)

	// Set network conditions
	framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000)

	// Apply a preset
	framework.ApplyNetworkPreset(SlowNetwork)

	// Disconnect the network
	framework.DisconnectNetwork()

	// Check if the network is connected
	if !framework.IsNetworkConnected() {
		// Handle disconnected state
	}

	// Reconnect the network
	framework.ReconnectNetwork()

	// Access the network simulator directly
	simulator := framework.GetNetworkSimulator()
	simulator.SetConditions(200*time.Millisecond, 0.2, 2000)
}

// Example of using NetworkSimulator in a test
func TestWithNetworkSimulator(t *testing.T) {
	// Create a test framework
	framework := NewTestFramework(TestConfig{}, nil)

	// Register a mock provider
	mockGraph := mock.NewMockGraphProvider()
	framework.RegisterMockProvider("graph", mockGraph)

	// Test with fast network
	framework.ApplyNetworkPreset(FastNetwork)
	// ... perform test with fast network

	// Test with slow network
	framework.ApplyNetworkPreset(SlowNetwork)
	// ... perform test with slow network

	// Test with network disconnection
	framework.DisconnectNetwork()
	// ... perform test with disconnected network

	// Test with network reconnection
	framework.ReconnectNetwork()
	// ... perform test with reconnected network
}

// MockProvider that returns errors based on network conditions
type NetworkAwareMockProvider struct {
	simulator NetworkSimulator
}

func NewNetworkAwareMockProvider(simulator NetworkSimulator) *NetworkAwareMockProvider {
	return &NetworkAwareMockProvider{
		simulator: simulator,
	}
}

func (m *NetworkAwareMockProvider) Setup() error {
	return nil
}

func (m *NetworkAwareMockProvider) Teardown() error {
	return nil
}

func (m *NetworkAwareMockProvider) DoSomething() error {
	// Simulate network delay
	m.simulator.SimulateNetworkDelay()

	// Check for network errors
	if err := m.simulator.SimulateNetworkError(); err != nil {
		return err
	}

	// Perform the actual operation
	return nil
}

// Example of creating a custom network-aware mock provider
func ExampleNetworkAwareMockProvider() {
	// Create a network simulator
	simulator := NewNetworkSimulator()

	// Create a network-aware mock provider
	provider := NewNetworkAwareMockProvider(simulator)

	// Set network conditions
	simulator.SetConditions(100*time.Millisecond, 0.1, 1000)

	// Use the provider
	err := provider.DoSomething()
	if err != nil {
		// Handle network error
	}
}
