# Test Utilities

This package provides testing utilities for the OneMount project, including the TestFramework and IntegrationTestEnvironment.

## TestFramework

The `TestFramework` provides a centralized test configuration and execution environment for the OneMount project. It helps manage test resources, mock providers, test execution, and context management.

### Features

- Test environment configuration
- Resource management with automatic cleanup
- Mock provider registration and retrieval
- Network condition simulation
- Test execution with timeout support
- Context management for cancellation and timeouts
- Structured logging

### Usage

#### Creating a TestFramework

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

#### Managing Resources

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

#### Using Mock Providers

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

#### Running Tests

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

#### Network Simulation

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

#### Context Management

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

### Example: Complete Test

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

### Best Practices

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

## IntegrationTestEnvironment

The `IntegrationTestEnvironment` provides a controlled environment for integration tests. It allows configuring which components are real and which are mocked, integrates with the NetworkSimulator, implements the TestDataManager interface for managing test data, and supports component isolation via the IsolationConfig.

### Features

- Component configuration (real vs mocked)
- Network simulation integration
- Test data management
- Component isolation via network rules
- Scenario-based testing
- Setup and teardown methods

### Usage

#### Creating an IntegrationTestEnvironment

```go
import (
    "context"
    "github.com/yourusername/onemount/internal/testutil"
)

// Create a logger
logger := &testutil.TestLogger{}

// Create a context
ctx := context.Background()

// Create a new IntegrationTestEnvironment
env := testutil.NewIntegrationTestEnvironment(ctx, logger)
```

#### Configuring Components

```go
// Configure which components should be mocked
env.SetIsolationConfig(testutil.IsolationConfig{
    MockedServices: []string{"graph", "filesystem", "ui"},
    NetworkRules:   []testutil.NetworkRule{},
    DataIsolation:  true,
})

// Set up the environment
err := env.SetupEnvironment()
if err != nil {
    // Handle error
}

// Get a component
graphComponent, err := env.GetComponent("graph")
if err != nil {
    // Handle error
}

// Use the component
mockGraph := graphComponent.(*testutil.MockGraphProvider)
mockGraph.AddMockItem("/drive/root", &graph.DriveItem{
    Name: "root",
    // ...
})
```

#### Managing Test Data

```go
// Get the test data manager
testDataManager := env.GetTestDataManager()

// Load test data
err := testDataManager.LoadTestData("test-data-set")
if err != nil {
    // Handle error
}

// Get test data
data := testDataManager.GetTestData("test-file.txt")

// Clean up test data
err = testDataManager.CleanupTestData()
if err != nil {
    // Handle error
}
```

#### Network Simulation

```go
// Get the network simulator
networkSimulator := env.GetNetworkSimulator()

// Disconnect the network
err := networkSimulator.Disconnect()
if err != nil {
    // Handle error
}

// Check if the network is connected
if !networkSimulator.IsConnected() {
    // Handle disconnected state
}

// Reconnect the network
err = networkSimulator.Reconnect()
if err != nil {
    // Handle error
}

// Set network conditions
err = networkSimulator.SetConditions(100*time.Millisecond, 0.1, 1000)
if err != nil {
    // Handle error
}
```

#### Component Isolation

```go
// Configure component isolation
env.SetIsolationConfig(testutil.IsolationConfig{
    MockedServices: []string{"graph", "filesystem"},
    NetworkRules: []testutil.NetworkRule{
        {
            Source:      "graph",
            Destination: "filesystem",
            Allow:       false,
        },
    },
    DataIsolation: true,
})

// Set up the environment with the new isolation config
err := env.SetupEnvironment()
if err != nil {
    // Handle error
}
```

#### Scenario-Based Testing

```go
// Create a test scenario
scenario := testutil.TestScenario{
    Name:        "File Operations",
    Description: "Tests file creation, modification, and deletion",
    Steps: []testutil.TestStep{
        {
            Name: "Create file",
            Action: func(ctx context.Context) error {
                // Implementation would create a file
                return nil
            },
            Validation: func(ctx context.Context) error {
                // Implementation would verify file was created
                return nil
            },
        },
        {
            Name: "Modify file",
            Action: func(ctx context.Context) error {
                // Implementation would modify the file
                return nil
            },
            Validation: func(ctx context.Context) error {
                // Implementation would verify file was modified
                return nil
            },
        },
        // More steps...
    },
    Assertions: []testutil.TestAssertion{
        {
            Name: "File operations completed successfully",
            Condition: func(ctx context.Context) bool {
                // Implementation would check if all operations were successful
                return true
            },
            Message: "File operations did not complete successfully",
        },
    },
    Cleanup: []testutil.CleanupStep{
        {
            Name: "Clean up test files",
            Action: func(ctx context.Context) error {
                // Implementation would clean up any remaining test files
                return nil
            },
            AlwaysRun: true,
        },
    },
}

// Add the scenario to the environment
env.AddScenario(scenario)

// Run the scenario
err := env.RunScenario("File Operations")
if err != nil {
    // Handle error
}

// Run all scenarios
errors := env.RunAllScenarios()
if len(errors) > 0 {
    // Handle errors
}
```

#### Teardown

```go
// Tear down the environment
err := env.TeardownEnvironment()
if err != nil {
    // Handle error
}
```

### Example: Complete Integration Test

```go
package mypackage_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestIntegration(t *testing.T) {
    // Create a logger
    logger := &testutil.TestLogger{}

    // Create a context
    ctx := context.Background()

    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(ctx, logger)
    require.NotNil(t, env)

    // Set up isolation config to mock all components
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph", "filesystem", "ui"},
        NetworkRules:   []testutil.NetworkRule{},
        DataIsolation:  true,
    })

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "Authentication Flow",
        Description: "Tests the complete authentication process",
        Steps: []testutil.TestStep{
            {
                Name: "Initialize authentication",
                Action: func(ctx context.Context) error {
                    // Implementation would initialize the authentication process
                    return nil
                },
            },
            {
                Name: "Request authorization",
                Action: func(ctx context.Context) error {
                    // Implementation would request authorization
                    return nil
                },
            },
            // More steps...
        },
        Assertions: []testutil.TestAssertion{
            {
                Name: "Token is valid",
                Condition: func(ctx context.Context) bool {
                    // Implementation would check if token is valid
                    return true
                },
                Message: "Authentication token is not valid",
            },
        },
        Cleanup: []testutil.CleanupStep{
            {
                Name: "Clear authentication tokens",
                Action: func(ctx context.Context) error {
                    // Implementation would clear authentication tokens
                    return nil
                },
                AlwaysRun: true,
            },
        },
    }

    // Add the scenario to the environment
    env.AddScenario(scenario)

    // Run the scenario
    err = env.RunScenario("Authentication Flow")
    require.NoError(t, err)

    // Test with network disconnection
    networkSimulator := env.GetNetworkSimulator()
    err = networkSimulator.Disconnect()
    require.NoError(t, err)

    // Create an offline scenario
    offlineScenario := testutil.TestScenario{
        Name:        "Offline Mode",
        Description: "Tests operation in offline mode",
        Steps: []testutil.TestStep{
            {
                Name: "Verify offline status",
                Action: func(ctx context.Context) error {
                    // Implementation would verify offline status
                    return nil
                },
            },
            // More steps...
        },
    }

    // Add and run the offline scenario
    env.AddScenario(offlineScenario)
    err = env.RunScenario("Offline Mode")
    require.NoError(t, err)

    // Reconnect for cleanup
    err = networkSimulator.Reconnect()
    require.NoError(t, err)
}
```

### Best Practices

1. Always use `t.Cleanup()` to ensure the environment is torn down, even if tests panic
2. Configure component isolation to match the test requirements
3. Use scenario-based testing for complex integration tests
4. Test with different network conditions to ensure robustness
5. Always reconnect the network after disconnection tests
6. Clean up test data after each test
7. Use descriptive names for scenarios, steps, and assertions
8. Include validation steps to verify the results of actions
9. Use cleanup steps to ensure the environment is left in a clean state
10. Test both normal operation and error handling under various conditions