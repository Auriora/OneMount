# TestFramework Core Documentation

## Overview

The `TestFramework` is the central component of the OneMount test infrastructure. It provides a centralized test configuration and execution environment, with features for resource management, mock provider registration, test execution, and context management.

## Features

- Test environment configuration
- Resource management with automatic cleanup
- Mock provider registration and retrieval
- Network condition simulation
- Test execution with timeout support
- Context management for cancellation and timeouts
- Structured logging

## API Reference

### Types

#### TestConfig

```go
type TestConfig struct {
    // Test environment configuration
    Environment string
    // Timeout for tests in seconds
    Timeout int
    // Whether to enable verbose logging
    VerboseLogging bool
    // Directory for test artifacts
    ArtifactsDir string
    // Custom configuration options
    CustomOptions map[string]interface{}
}
```

The `TestConfig` struct defines configuration options for the test environment.

#### TestResource

```go
type TestResource interface {
    // Cleanup performs necessary cleanup operations for the resource.
    Cleanup() error
}
```

The `TestResource` interface represents a resource that needs cleanup after tests.

#### MockProvider

```go
type MockProvider interface {
    // Setup initializes the mock provider.
    Setup() error
    // Teardown cleans up the mock provider.
    Teardown() error
    // Reset resets the mock provider to its initial state.
    Reset() error
}
```

The `MockProvider` interface is implemented by mock components.

#### TestStatus

```go
type TestStatus string

const (
    // TestStatusPassed indicates the test passed.
    TestStatusPassed TestStatus = "PASSED"
    // TestStatusFailed indicates the test failed.
    TestStatusFailed TestStatus = "FAILED"
    // TestStatusSkipped indicates the test was skipped.
    TestStatusSkipped TestStatus = "SKIPPED"
)
```

The `TestStatus` type represents the status of a test.

#### TestFailure

```go
type TestFailure struct {
    // Message describes the failure.
    Message string
    // Location is where the failure occurred.
    Location string
    // Expected is what was expected.
    Expected interface{}
    // Actual is what was actually received.
    Actual interface{}
}
```

The `TestFailure` struct represents a test failure.

#### TestArtifact

```go
type TestArtifact struct {
    // Name of the artifact.
    Name string
    // Type of the artifact.
    Type string
    // Location where the artifact is stored.
    Location string
}
```

The `TestArtifact` struct represents a test artifact.

#### TestResult

```go
type TestResult struct {
    // Name of the test.
    Name string
    // Duration of the test.
    Duration time.Duration
    // Status of the test.
    Status TestStatus
    // Failures that occurred during the test.
    Failures []TestFailure
    // Artifacts generated during the test.
    Artifacts []TestArtifact
}
```

The `TestResult` struct represents the result of a test.

#### TestLifecycle

```go
type TestLifecycle interface {
    // BeforeTest is called before a test is executed.
    BeforeTest(ctx context.Context) error
    // AfterTest is called after a test is executed.
    AfterTest(ctx context.Context) error
    // OnFailure is called when a test fails.
    OnFailure(ctx context.Context, failure TestFailure) error
}
```

The `TestLifecycle` interface defines hooks for test lifecycle events.

#### TestFramework

```go
type TestFramework struct {
    // Configuration for the test environment.
    Config TestConfig

    // Test resources that need cleanup.
    resources []TestResource

    // Mock providers.
    mockProviders map[string]MockProvider

    // Coverage reporting.
    coverageReporter CoverageReporter

    // Network simulation.
    networkSimulator NetworkSimulator

    // Context for timeout/cancellation.
    ctx context.Context

    // Structured logging.
    logger Logger
}
```

The `TestFramework` struct provides centralized test configuration and execution.

### Functions

#### NewTestFramework

```go
func NewTestFramework(config TestConfig, logger Logger) *TestFramework
```

Creates a new `TestFramework` with the given configuration.

### Methods

#### AddResource

```go
func (tf *TestFramework) AddResource(resource TestResource)
```

Adds a resource to be cleaned up after tests.

#### CleanupResources

```go
func (tf *TestFramework) CleanupResources() error
```

Cleans up all registered resources.

#### RegisterMockProvider

```go
func (tf *TestFramework) RegisterMockProvider(name string, provider MockProvider)
```

Registers a mock provider with the given name.

#### GetMockProvider

```go
func (tf *TestFramework) GetMockProvider(name string) (MockProvider, bool)
```

Returns the mock provider with the given name.

#### SetCoverageReporter

```go
func (tf *TestFramework) SetCoverageReporter(reporter CoverageReporter)
```

Sets the coverage reporter for the test framework.

#### RunTest

```go
func (tf *TestFramework) RunTest(name string, testFunc func(ctx context.Context) error) TestResult
```

Executes a single test function with the given name.

#### RunTestSuite

```go
func (tf *TestFramework) RunTestSuite(name string, tests map[string]func(ctx context.Context) error) []TestResult
```

Executes a suite of tests and returns the results.

#### WithTimeout

```go
func (tf *TestFramework) WithTimeout(timeout time.Duration) context.Context
```

Returns a new context with the specified timeout.

#### WithCancel

```go
func (tf *TestFramework) WithCancel() (context.Context, context.CancelFunc)
```

Returns a new context with a cancel function.

#### SetContext

```go
func (tf *TestFramework) SetContext(ctx context.Context)
```

Sets the base context for the test framework.

#### GetNetworkSimulator

```go
func (tf *TestFramework) GetNetworkSimulator() NetworkSimulator
```

Returns the network simulator.

#### SetNetworkSimulator

```go
func (tf *TestFramework) SetNetworkSimulator(simulator NetworkSimulator)
```

Sets the network simulator.

#### SetNetworkConditions

```go
func (tf *TestFramework) SetNetworkConditions(latency time.Duration, packetLoss float64, bandwidth int) error
```

Sets the network conditions.

#### ApplyNetworkPreset

```go
func (tf *TestFramework) ApplyNetworkPreset(preset NetworkCondition) error
```

Applies a predefined network condition preset.

#### DisconnectNetwork

```go
func (tf *TestFramework) DisconnectNetwork() error
```

Simulates a network disconnection.

#### ReconnectNetwork

```go
func (tf *TestFramework) ReconnectNetwork() error
```

Restores the network connection.

#### IsNetworkConnected

```go
func (tf *TestFramework) IsNetworkConnected() bool
```

Returns whether the network is currently connected.

## Usage Examples

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

## Complete Example

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

## Related Components

- [NetworkSimulator](network-simulator.md): Network condition simulation
- [MockProviders](mock-providers.md): Mock implementations of system components
- [CoverageReporter](coverage-reporter.md): Test coverage collection and reporting
- [IntegrationTestEnvironment](integration-test-environment.md): Integration testing environment