# TestFramework

The `TestFramework` provides a centralized test configuration and execution environment for the OneMount project. It helps manage test resources, mock providers, test execution, and context management.

## Features

- Test environment configuration
- Resource management with automatic cleanup
- Mock provider registration and retrieval
- Network condition simulation
- Test execution with timeout support
- Context management for cancellation and timeouts
- Structured logging

## Usage

### Creating a TestFramework

```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    "github.com/yourusername/onemount/internal/testutil"
)

// Create a logger
logger := log.With().Str("component", "test").Logger()

// Create a test configuration
config := testutil.TestConfig{
    Environment:    "test",
    Timeout:        30,  // 30 seconds
    VerboseLogging: true,
    ArtifactsDir:   "/tmp/test-artifacts",
}

// Create a new TestFramework
framework := testutil.NewTestFramework(config, &logger)
```

### Managing Resources

```go
// Add a resource to be cleaned up after tests
resource := NewSomeResource()
framework.AddResource(resource)

// Clean up all resources
err := framework.CleanupResources()
if err != nil {
    // Handle error
}
```

### Using Mock Providers

```go
// Register a mock provider
mockGraph := NewMockGraphProvider()
framework.RegisterMockProvider("graph", mockGraph)

// Get a registered mock provider
provider, exists := framework.GetMockProvider("graph")
if exists {
    // Use the provider
}
```

### Running Tests

```go
// Run a single test
result := framework.RunTest("test-name", func(ctx context.Context) error {
    // Test logic here
    return nil
})

// Check the test result
if result.Status == testutil.TestStatusPassed {
    // Test passed
} else {
    // Test failed
    for _, failure := range result.Failures {
        fmt.Printf("Failure: %s at %s\n", failure.Message, failure.Location)
    }
}

// Run a test suite
tests := map[string]func(ctx context.Context) error{
    "test1": func(ctx context.Context) error {
        // Test 1 logic
        return nil
    },
    "test2": func(ctx context.Context) error {
        // Test 2 logic
        return nil
    },
}

results := framework.RunTestSuite("suite-name", tests)
```

### Network Simulation

```go
// Set network conditions (latency, packet loss, bandwidth)
framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000) // 100ms latency, 10% packet loss, 1Mbps bandwidth

// Apply a predefined network condition preset
framework.ApplyNetworkPreset(testutil.SlowNetwork)

// Available presets:
// - FastNetwork: Fast, reliable network connection (10ms latency, 0% packet loss, 100Mbps)
// - AverageNetwork: Average home broadband (50ms latency, 1% packet loss, 20Mbps)
// - SlowNetwork: Slow connection (200ms latency, 5% packet loss, 1Mbps)
// - MobileNetwork: Mobile data connection (100ms latency, 2% packet loss, 5Mbps)
// - IntermittentConnection: Unstable connection (300ms latency, 15% packet loss, 2Mbps)
// - SatelliteConnection: High-latency satellite (700ms latency, 3% packet loss, 10Mbps)

// Simulate network disconnection
framework.DisconnectNetwork()

// Check if the network is connected
if !framework.IsNetworkConnected() {
    // Handle disconnected state
}

// Restore network connection
framework.ReconnectNetwork()

// Access the network simulator directly for advanced usage
simulator := framework.GetNetworkSimulator()
```

### Context Management

```go
// Create a context with timeout
ctx := framework.WithTimeout(5 * time.Second)

// Create a context with cancel function
ctx, cancel := framework.WithCancel()
defer cancel()

// Set a custom context
customCtx := context.WithValue(context.Background(), "key", "value")
framework.SetContext(customCtx)
```

## Example: Complete Test

```go
package mypackage_test

import (
    "context"
    "testing"
    "time"

    "github.com/rs/zerolog/log"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestMyFeature(t *testing.T) {
    // Create logger
    logger := log.With().Str("component", "test").Logger()

    // Create test framework
    framework := testutil.NewTestFramework(testutil.TestConfig{
        Environment:    "test",
        Timeout:        30,
        VerboseLogging: true,
    }, &logger)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        framework.CleanupResources()
    })

    // Register mock providers
    mockGraph := NewMockGraphProvider()
    framework.RegisterMockProvider("graph", mockGraph)

    // Set up test resources
    tempDir := createTempDir()
    framework.AddResource(tempDir)

    // Configure network conditions for the test
    framework.ApplyNetworkPreset(testutil.AverageNetwork)

    // Run the test with normal network conditions
    result := framework.RunTest("my-feature-test-normal-network", func(ctx context.Context) error {
        // Test logic using the context with normal network conditions
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Perform test operations
            return nil
        }
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
        // Should handle offline mode or return appropriate errors
        return nil
    })

    // Reconnect for cleanup
    framework.ReconnectNetwork()

    // Check the result
    if result.Status != testutil.TestStatusPassed {
        t.Errorf("Test failed: %v", result.Failures)
    }
}
```

## Best Practices

1. Always use `t.Cleanup()` to ensure resources are cleaned up, even if tests panic
2. Use context timeouts for tests that might hang
3. Register mock providers with descriptive names
4. Use structured logging for better test diagnostics
5. Add resources in the order they should be cleaned up (cleanup happens in reverse order)
6. Test with different network conditions to ensure robustness
7. Always reconnect the network after disconnection tests
8. Use network presets for consistent test conditions
9. Test both normal operation and error handling under poor network conditions
10. Consider using network simulation in CI/CD pipelines to catch network-related issues early
