# NetworkSimulator Documentation

## Overview

The `NetworkSimulator` is a key component of the OneMount test framework that allows simulating different network conditions for testing. This is particularly important for testing a filesystem that interacts with a cloud service like OneDrive, as network conditions can significantly impact the behavior and performance of the system.

## Features

- Simulate network latency
- Simulate packet loss
- Simulate bandwidth limitations
- Simulate network disconnection and reconnection
- Apply predefined network condition presets
- Register mock providers for network condition simulation
- Simulate network errors

## API Reference

### Types

#### NetworkCondition

```go
type NetworkCondition struct {
    // Name of the network condition
    Name string
    // Latency in milliseconds
    Latency time.Duration
    // Packet loss as a percentage (0.0 to 1.0)
    PacketLoss float64
    // Bandwidth in Kbps
    Bandwidth int
}
```

The `NetworkCondition` struct represents a specific network condition.

#### NetworkSimulator

```go
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
    // RegisterProvider registers a mock provider for network condition simulation
    RegisterProvider(provider MockProvider) error
    // SimulateNetworkDelay simulates network delay based on current conditions
    SimulateNetworkDelay()
    // SimulatePacketLoss simulates packet loss based on current conditions
    SimulatePacketLoss() bool
    // SimulateNetworkError simulates a network error
    SimulateNetworkError() error
}
```

The `NetworkSimulator` interface defines methods for simulating different network conditions.

#### DefaultNetworkSimulator

```go
type DefaultNetworkSimulator struct {
    // Current network conditions
    currentConditions NetworkCondition
    // Whether the network is connected
    connected bool
    // Registered mock providers
    providers []MockProvider
    // Mutex for thread safety
    mu sync.Mutex
}
```

The `DefaultNetworkSimulator` is the default implementation of the `NetworkSimulator` interface.

### Predefined Network Conditions

```go
var (
    // FastNetwork represents a fast, reliable network connection
    FastNetwork = NetworkCondition{
        Name:       "Fast Network",
        Latency:    10 * time.Millisecond,
        PacketLoss: 0.0,
        Bandwidth:  100000, // 100 Mbps
    }

    // AverageNetwork represents an average home broadband connection
    AverageNetwork = NetworkCondition{
        Name:       "Average Network",
        Latency:    50 * time.Millisecond,
        PacketLoss: 0.01, // 1%
        Bandwidth:  20000, // 20 Mbps
    }

    // SlowNetwork represents a slow connection
    SlowNetwork = NetworkCondition{
        Name:       "Slow Network",
        Latency:    200 * time.Millisecond,
        PacketLoss: 0.05, // 5%
        Bandwidth:  1000,  // 1 Mbps
    }

    // MobileNetwork represents a mobile data connection
    MobileNetwork = NetworkCondition{
        Name:       "Mobile Network",
        Latency:    100 * time.Millisecond,
        PacketLoss: 0.02, // 2%
        Bandwidth:  5000,  // 5 Mbps
    }

    // IntermittentConnection represents an unstable connection
    IntermittentConnection = NetworkCondition{
        Name:       "Intermittent Connection",
        Latency:    300 * time.Millisecond,
        PacketLoss: 0.15, // 15%
        Bandwidth:  2000,  // 2 Mbps
    }

    // SatelliteConnection represents a high-latency satellite connection
    SatelliteConnection = NetworkCondition{
        Name:       "Satellite Connection",
        Latency:    700 * time.Millisecond,
        PacketLoss: 0.03, // 3%
        Bandwidth:  10000, // 10 Mbps
    }
)
```

### Functions

#### NewNetworkSimulator

```go
func NewNetworkSimulator() *DefaultNetworkSimulator
```

Creates a new `DefaultNetworkSimulator` with default settings.

### Methods

#### SetConditions

```go
func (s *DefaultNetworkSimulator) SetConditions(latency time.Duration, packetLoss float64, bandwidth int) error
```

Sets the network conditions.

#### ApplyPreset

```go
func (s *DefaultNetworkSimulator) ApplyPreset(preset NetworkCondition) error
```

Applies a predefined network condition preset.

#### GetCurrentConditions

```go
func (s *DefaultNetworkSimulator) GetCurrentConditions() NetworkCondition
```

Returns the current network conditions.

#### Disconnect

```go
func (s *DefaultNetworkSimulator) Disconnect() error
```

Simulates a network disconnection.

#### Reconnect

```go
func (s *DefaultNetworkSimulator) Reconnect() error
```

Restores the network connection.

#### IsConnected

```go
func (s *DefaultNetworkSimulator) IsConnected() bool
```

Returns whether the network is currently connected.

#### RegisterProvider

```go
func (s *DefaultNetworkSimulator) RegisterProvider(provider MockProvider) error
```

Registers a mock provider for network condition simulation.

#### SimulateNetworkDelay

```go
func (s *DefaultNetworkSimulator) SimulateNetworkDelay()
```

Simulates network delay based on current conditions.

#### SimulatePacketLoss

```go
func (s *DefaultNetworkSimulator) SimulatePacketLoss() bool
```

Simulates packet loss based on current conditions.

#### SimulateNetworkError

```go
func (s *DefaultNetworkSimulator) SimulateNetworkError() error
```

Simulates a network error.

## Usage Examples

### Setting Network Conditions

```go
// Create a network simulator
simulator := testutil.NewNetworkSimulator()

// Set network conditions (latency, packet loss, bandwidth)
err := simulator.SetConditions(100*time.Millisecond, 0.1, 1000) // 100ms latency, 10% packet loss, 1Mbps bandwidth
if err != nil {
    // Handle error
}

// Get current conditions
conditions := simulator.GetCurrentConditions()
fmt.Printf("Latency: %v, Packet Loss: %.2f%%, Bandwidth: %d Kbps\n",
    conditions.Latency, conditions.PacketLoss*100, conditions.Bandwidth)
```

### Using Network Presets

```go
// Apply a predefined network condition preset
err := simulator.ApplyPreset(testutil.SlowNetwork)
if err != nil {
    // Handle error
}

// Available presets:
// - FastNetwork: Fast, reliable network connection (10ms latency, 0% packet loss, 100Mbps)
// - AverageNetwork: Average home broadband (50ms latency, 1% packet loss, 20Mbps)
// - SlowNetwork: Slow connection (200ms latency, 5% packet loss, 1Mbps)
// - MobileNetwork: Mobile data connection (100ms latency, 2% packet loss, 5Mbps)
// - IntermittentConnection: Unstable connection (300ms latency, 15% packet loss, 2Mbps)
// - SatelliteConnection: High-latency satellite (700ms latency, 3% packet loss, 10Mbps)
```

### Simulating Network Disconnection

```go
// Simulate network disconnection
err := simulator.Disconnect()
if err != nil {
    // Handle error
}

// Check if the network is connected
if !simulator.IsConnected() {
    fmt.Println("Network is disconnected")
}

// Restore network connection
err = simulator.Reconnect()
if err != nil {
    // Handle error
}

// Check if the network is connected
if simulator.IsConnected() {
    fmt.Println("Network is connected")
}
```

### Simulating Network Delay and Packet Loss

```go
// Simulate network delay
simulator.SimulateNetworkDelay()

// Simulate packet loss
if simulator.SimulatePacketLoss() {
    fmt.Println("Packet lost")
} else {
    fmt.Println("Packet delivered")
}

// Simulate network error
err := simulator.SimulateNetworkError()
if err != nil {
    fmt.Printf("Network error: %v\n", err)
}
```

### Registering Mock Providers

```go
// Create a mock provider
mockGraph := NewMockGraphProvider()

// Register the mock provider with the network simulator
err := simulator.RegisterProvider(mockGraph)
if err != nil {
    // Handle error
}
```

## Integration with TestFramework

The `NetworkSimulator` is integrated with the `TestFramework` to provide network condition simulation for tests:

```go
// Create a test framework
framework := testutil.NewTestFramework(config, &logger)

// Set network conditions
framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000)

// Apply a network preset
framework.ApplyNetworkPreset(testutil.SlowNetwork)

// Disconnect the network
framework.DisconnectNetwork()

// Reconnect the network
framework.ReconnectNetwork()

// Check if the network is connected
if !framework.IsNetworkConnected() {
    // Handle disconnected state
}

// Get the network simulator for advanced usage
simulator := framework.GetNetworkSimulator()
```

## Best Practices

1. **Test with Different Network Conditions**: Test your application under various network conditions to ensure it behaves correctly in all scenarios.

2. **Use Presets for Consistency**: Use the predefined network condition presets for consistent testing across different test runs.

3. **Simulate Real-World Scenarios**: Simulate real-world network scenarios, such as mobile networks, satellite connections, and intermittent connections.

4. **Test Disconnection Handling**: Test how your application handles network disconnections and reconnections.

5. **Combine with Mock Providers**: Use the network simulator in conjunction with mock providers to simulate network-related behavior of external dependencies.

6. **Clean Up After Tests**: Always reconnect the network after disconnection tests to ensure a clean state for subsequent tests.

7. **Use in CI/CD Pipelines**: Include network condition testing in your CI/CD pipelines to catch network-related issues early.

8. **Test Error Handling**: Test how your application handles network errors, such as timeouts and connection failures.

9. **Simulate Gradual Degradation**: Test how your application behaves as network conditions gradually degrade.

10. **Monitor Performance Metrics**: Monitor performance metrics under different network conditions to identify potential bottlenecks.

## Related Components

- [Testing Framework Guide](testing-framework-guide.md): Core test configuration and execution
- [MockProviders](mock-providers.md): Mock implementations of system components
- [IntegrationTestEnvironment](integration-test-environment.md): Integration testing environment
