# Tutorial: Network Simulation

This tutorial will guide you through the process of using network simulation in your tests. Network simulation allows you to test how your code behaves under different network conditions, such as latency, packet loss, and bandwidth limitations.

> **Note**: All code examples in this tutorial are for illustration purposes only and may need to be adapted to your specific project structure and imports. The examples are not meant to be compiled directly but rather to demonstrate concepts and patterns.

## Table of Contents

1. [Introduction to Network Simulation](#introduction-to-network-simulation)
2. [The NetworkSimulator Component](#the-networksimulator-component)
3. [Setting Up Network Conditions](#setting-up-network-conditions)
4. [Using Network Presets](#using-network-presets)
5. [Simulating Network Disconnection](#simulating-network-disconnection)
6. [Testing with Different Network Conditions](#testing-with-different-network-conditions)
7. [Best Practices](#best-practices)
8. [Complete Example](#complete-example)

## Introduction to Network Simulation

Network simulation is a powerful technique for testing how your code behaves under different network conditions. It allows you to:

- Test how your code handles slow or unreliable networks
- Verify that your code gracefully handles network disconnections
- Ensure that your code performs well under various network conditions
- Identify and fix issues that only occur under specific network conditions

The OneMount test framework provides a NetworkSimulator component that allows you to simulate different network conditions in your tests.

## The NetworkSimulator Component

The NetworkSimulator component is part of the TestFramework and provides methods for:

- Setting network conditions (latency, packet loss, bandwidth)
- Applying predefined network condition presets
- Simulating network disconnection and reconnection
- Checking the current network status

You can access the NetworkSimulator through the TestFramework:

```go
// Get the network simulator from the framework
simulator := framework.GetNetworkSimulator()

// Or use the framework's convenience methods
framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000) // 100ms latency, 10% packet loss, 1Mbps bandwidth
```

> Note: The code examples in this tutorial are for illustration purposes and may need to be adapted to your specific project structure and imports.

## Setting Up Network Conditions

You can set up network conditions using the `SetNetworkConditions` method:

```go
// Set network conditions (latency, packet loss, bandwidth)
framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000) // 100ms latency, 10% packet loss, 1Mbps bandwidth
```

> Note: The code examples in this tutorial are for illustration purposes and may need to be adapted to your specific project structure and imports.

The parameters are:

1. **Latency**: The delay before data is sent (time.Duration)
2. **Packet Loss**: The probability of a packet being lost (float64, 0.0-1.0)
3. **Bandwidth**: The maximum data transfer rate in kilobits per second (int)

These conditions affect all mock providers registered with the framework, simulating how they would behave under the specified network conditions.

## Using Network Presets

The TestFramework provides predefined network condition presets that you can apply:

```go
// Apply a predefined network condition preset
framework.ApplyNetworkPreset(testutil.SlowNetwork)
```

Available presets include:

- **FastNetwork**: Fast, reliable network connection (10ms latency, 0% packet loss, 100Mbps)
- **AverageNetwork**: Average home broadband (50ms latency, 1% packet loss, 20Mbps)
- **SlowNetwork**: Slow connection (200ms latency, 5% packet loss, 1Mbps)
- **MobileNetwork**: Mobile data connection (100ms latency, 2% packet loss, 5Mbps)
- **IntermittentConnection**: Unstable connection (300ms latency, 15% packet loss, 2Mbps)
- **SatelliteConnection**: High-latency satellite (700ms latency, 3% packet loss, 10Mbps)

These presets provide a convenient way to test your code under common network conditions.

## Simulating Network Disconnection

You can simulate network disconnection and reconnection using the `DisconnectNetwork` and `ReconnectNetwork` methods:

```go
// Simulate network disconnection
framework.DisconnectNetwork()

// Check if the network is connected
if !framework.IsNetworkConnected() {
    // Handle disconnected state
}

// Restore network connection
framework.ReconnectNetwork()
```

When the network is disconnected, all mock providers will simulate being offline, returning appropriate errors for network operations.

## Testing with Different Network Conditions

To test how your code behaves under different network conditions, you can run the same test multiple times with different network conditions:

```go
// Test with normal network conditions
framework.ApplyNetworkPreset(testutil.FastNetwork)
result := framework.RunTest("my-feature-test-fast-network", func(ctx context.Context) error {
    // Test logic using the context with normal network conditions
    return nil
})

// Test with slow network
framework.ApplyNetworkPreset(testutil.SlowNetwork)
result = framework.RunTest("my-feature-test-slow-network", func(ctx context.Context) error {
    // Test logic using the context with slow network
    return nil
})

// Test with network disconnection
framework.DisconnectNetwork()
result = framework.RunTest("my-feature-test-disconnected", func(ctx context.Context) error {
    // Test logic with network disconnection
    return nil
})

// Reconnect for cleanup
framework.ReconnectNetwork()
```

This approach allows you to verify that your code handles different network conditions correctly.

## Best Practices

When using network simulation in your tests, follow these best practices:

1. **Test with realistic conditions**: Use network conditions that reflect real-world scenarios.
2. **Test both normal and error cases**: Verify that your code works correctly under good network conditions and handles poor network conditions gracefully.
3. **Use network presets for consistency**: Use the predefined network presets to ensure consistent test conditions.
4. **Always reconnect after disconnection tests**: Ensure that the network is reconnected after testing with disconnection.
5. **Use dynamic waiting**: When testing with slow networks, use dynamic waiting instead of fixed timeouts.
6. **Test timeout handling**: Verify that your code handles timeouts correctly under slow network conditions.
7. **Test retry logic**: If your code includes retry logic, test it under poor network conditions.
8. **Document network conditions**: Comment your tests to explain what network conditions are being simulated.

## Complete Example

Here's a complete example of using network simulation in a test:

```go
package mypackage_test

import (
    "context"
    "testing"
    "time"

    "github.com/rs/zerolog/log"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestNetworkConditions(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "test").Logger()

    // Create a test configuration
    config := testutil.TestConfig{
        Environment:    "test",
        Timeout:        30,
        VerboseLogging: true,
    }

    // Create a new TestFramework
    framework := testutil.NewTestFramework(config, &logger)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        framework.CleanupResources()
    })

    // Create and register mock providers
    mockGraph := testutil.NewMockGraphProvider()
    framework.RegisterMockProvider("graph", mockGraph)

    // Configure mock responses
    resource := "/me/drive/root:/test.txt"
    mockItem := &graph.DriveItem{
        ID:   "item123",
        Name: "test.txt",
        Size: 1024,
        File: &graph.File{
            MimeType: "text/plain",
        },
    }
    mockGraph.AddMockItem(resource, mockItem)
    mockGraph.AddMockContent(resource, []byte("Hello, World!"))

    // Test with fast network
    framework.ApplyNetworkPreset(testutil.FastNetwork)
    startTime := time.Now()
    result := framework.RunTest("file-download-fast-network", func(ctx context.Context) error {
        // Get the file from the Graph API
        item, err := mockGraph.GetItemPath("/test.txt")
        if err != nil {
            return err
        }

        // Download the file content
        content, err := mockGraph.GetItemContent(item.ID)
        if err != nil {
            return err
        }

        // Verify the content
        if string(content) != "Hello, World!" {
            return fmt.Errorf("file content mismatch: got %q, want %q", string(content), "Hello, World!")
        }

        return nil
    })
    fastNetworkDuration := time.Since(startTime)

    // Check the result
    require.Equal(t, testutil.TestStatusPassed, result.Status, "Test failed with fast network: %v", result.Failures)

    // Test with slow network
    framework.ApplyNetworkPreset(testutil.SlowNetwork)
    startTime = time.Now()
    result = framework.RunTest("file-download-slow-network", func(ctx context.Context) error {
        // Get the file from the Graph API
        item, err := mockGraph.GetItemPath("/test.txt")
        if err != nil {
            return err
        }

        // Download the file content
        content, err := mockGraph.GetItemContent(item.ID)
        if err != nil {
            return err
        }

        // Verify the content
        if string(content) != "Hello, World!" {
            return fmt.Errorf("file content mismatch: got %q, want %q", string(content), "Hello, World!")
        }

        return nil
    })
    slowNetworkDuration := time.Since(startTime)

    // Check the result
    require.Equal(t, testutil.TestStatusPassed, result.Status, "Test failed with slow network: %v", result.Failures)

    // Verify that the slow network test took longer
    assert.Greater(t, slowNetworkDuration, fastNetworkDuration, "Slow network test should take longer than fast network test")

    // Test with network disconnection
    framework.DisconnectNetwork()
    result = framework.RunTest("file-download-disconnected", func(ctx context.Context) error {
        // Get the file from the Graph API
        _, err := mockGraph.GetItemPath("/test.txt")

        // Verify that an error is returned
        if err == nil {
            return fmt.Errorf("expected error when network is disconnected, got nil")
        }

        // Verify that the error is a network error
        var netErr *graph.NetworkError
        if !errors.As(err, &netErr) {
            return fmt.Errorf("expected NetworkError, got %T: %v", err, err)
        }

        return nil
    })

    // Check the result
    require.Equal(t, testutil.TestStatusPassed, result.Status, "Test failed with disconnected network: %v", result.Failures)

    // Reconnect for cleanup
    framework.ReconnectNetwork()
}
```

This example demonstrates:
1. Testing with different network conditions (fast, slow, disconnected)
2. Verifying that the code behaves correctly under each condition
3. Measuring the performance impact of different network conditions
4. Testing error handling when the network is disconnected
5. Properly reconnecting the network after testing

By using network simulation in your tests, you can ensure that your code handles different network conditions correctly and gracefully.
