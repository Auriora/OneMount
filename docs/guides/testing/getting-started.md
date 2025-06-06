# OneMount Test Framework Documentation

## Overview

This documentation provides a comprehensive guide to the OneMount test framework, which supports various types of testing for the OneMount filesystem. The framework is designed to make it easy to write effective tests that verify the behavior and performance of the system under different conditions.

## Key Concepts

The OneMount test framework supports various types of testing:

### Unit Testing

Unit testing focuses on testing individual components or units of code in isolation. The OneMount test framework provides specialized utilities for unit testing, making it easier to write comprehensive and effective unit tests.

```go
// Create a unit test framework
framework := testutil.NewUnitTestFramework(t)

// Use assertions
framework.Assertions.Equal(expected, actual, "values should be equal")
```

For detailed documentation, see [Unit Testing Guide](frameworks/unit-testing-guide.md).

### Integration Testing

Integration testing focuses on verifying that different components of the system work together correctly. The OneMount test framework provides specialized utilities for integration testing, making it easier to set up controlled test environments, configure component interactions, and verify interface contracts.

```go
// Create a test scenario
scenario := testutil.TestScenario{
    Name:        "File Operations",
    Description: "Tests file creation, modification, and deletion",
    Steps: []testutil.TestStep{
        // Steps...
    },
    Assertions: []testutil.TestAssertion{
        // Assertions...
    },
    Cleanup: []testutil.CleanupStep{
        // Cleanup steps...
    },
}

// Run the scenario
err := env.RunScenario("File Operations")
```

For detailed documentation, see [Integration Testing Guide](frameworks/integration-testing-guide.md).

### Performance Testing

Performance testing focuses on verifying that the system meets performance requirements under various conditions. The OneMount test framework provides utilities for performance testing, including metrics collection, threshold checking, and result visualization.

For detailed documentation, see [Performance Testing Guide](frameworks/performance-testing-guide.md).

### Load Testing

Load testing focuses on verifying that the system can handle expected load and stress conditions. The OneMount test framework provides utilities for load testing, including concurrent user simulation, sustained load testing, and load spike testing.

For detailed documentation, see [Load Testing Guide](frameworks/load-testing-guide.md).

### Security Testing

Security testing focuses on verifying that the system meets security requirements and is resistant to security threats. The OneMount test framework provides utilities for security testing, including vulnerability scanning, security control verification, and authentication/authorization testing.

For detailed documentation, see [Security Testing Guide](frameworks/security-testing-guide.md).

## Architecture

The OneMount test framework is built around a core `TestFramework` class that provides centralized test configuration and execution. This core framework is extended by specialized frameworks for different types of testing:

![Test Framework Architecture](../resources/test-framework-architecture.png)

The architecture is designed to be modular and extensible, allowing for easy addition of new test types and components.

For a detailed overview of the architecture, see [Testing Framework Overview](testing-framework-overview.md).

### Components

The OneMount test framework consists of several key components:

#### TestFramework

The `TestFramework` is the central component of the test infrastructure, providing features for resource management, mock provider registration, test execution, and context management.

```go
// Create a test framework
framework := testutil.NewTestFramework(config, &logger)

// Run a test
result := framework.RunTest("test-name", func(ctx context.Context) error {
    // Test logic here
    return nil
})
```

For detailed documentation, see [Testing Framework Overview](testing-framework-overview.md).

#### NetworkSimulator

The `NetworkSimulator` allows simulating different network conditions for testing, including latency, packet loss, bandwidth limitations, and network disconnections.

```go
// Set network conditions
framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000)

// Simulate network disconnection
framework.DisconnectNetwork()
```

For detailed documentation, see the code comments in `pkg/testutil/framework/network_simulator.go`.

#### MockProviders

Mock providers are controlled implementations of system components that allow tests to run without relying on actual external services or components.

```go
// Register a mock provider
mockGraph := testutil.NewMockGraphProvider()
framework.RegisterMockProvider("graph", mockGraph)

// Configure mock behavior
mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
    ID:   "root",
    Name: "root",
})
```

For detailed documentation, see the code comments in `pkg/testutil/mock/mock_graph.go`, `pkg/testutil/mock/mock_filesystem.go`, and `pkg/testutil/mock/mock_ui.go`.

#### IntegrationTestEnvironment

The `IntegrationTestEnvironment` provides a controlled environment for integration tests, with support for component isolation, network simulation, and test data management.

```go
// Create a test environment
env := testutil.NewIntegrationTestEnvironment(ctx, logger)

// Set up isolation config
env.SetIsolationConfig(testutil.IsolationConfig{
    MockedServices: []string{"graph", "filesystem", "ui"},
    NetworkRules:   []testutil.NetworkRule{},
    DataIsolation:  true,
})
```

For detailed documentation, see [Integration Testing Guide](frameworks/integration-testing-guide.md).

## Getting Started

To start using the OneMount test framework, follow these steps:

1. **Import the test framework package**:
   ```go
   import "github.com/onemount/testutil"
   ```

2. **Create a test framework instance**:
   ```go
   framework := testutil.NewTestFramework(config, &logger)
   ```

3. **Configure the test environment**:
   ```go
   framework.SetupTestEnvironment(testutil.EnvironmentConfig{
       DataDir:     "/tmp/test-data",
       NetworkMode: testutil.NetworkModeSimulated,
   })
   ```

4. **Write and run tests**:
   ```go
   result := framework.RunTest("my-test", func(ctx context.Context) error {
       // Test logic here
       return nil
   })
   ```

5. **Clean up resources**:
   ```go
   framework.Cleanup()
   ```

For more detailed examples, refer to the specific testing guides for each test type.

## Best Practices

### General Best Practices

1. **Use the Right Test Type**: Choose the appropriate test type for what you're testing.
2. **Isolate Tests**: Make sure tests don't depend on each other or on external state.
3. **Clean Up Resources**: Always clean up resources after tests, even if they fail.
4. **Test Edge Cases**: Include tests for edge cases, such as empty inputs, boundary values, and error conditions.
5. **Use Descriptive Names**: Give tests descriptive names that clearly indicate what is being tested.

For detailed best practices for each test type, see the respective documentation:

- [Unit Testing Best Practices](frameworks/unit-testing-guide.md#best-practices)
- [Integration Testing Best Practices](frameworks/integration-testing-guide.md#best-practices)
- [Performance Testing Best Practices](frameworks/performance-testing-guide.md#best-practices)
- [Load Testing Best Practices](frameworks/load-testing-guide.md#best-practices)
- [Security Testing Best Practices](frameworks/security-testing-guide.md#best-practices)

## Troubleshooting

If you encounter issues while using the test framework, check the [Troubleshooting Guide](testing-troubleshooting.md) for solutions to common problems, including:

- Tests fail to start
- Mock providers not working correctly
- Network simulation issues
- Resource cleanup failures
- Integration test environment issues
- Test scenario execution issues
- Performance testing issues
- Security testing issues

## Related Resources

- [Test Architecture Design](../../2-architecture-and-design/test-architecture-design.md)
- [Test Guidelines](test-guidelines.md)
- [Test Sandbox Guidelines](components/test-sandbox-guide.md)
- [Unit Testing Guide](frameworks/unit-testing-guide.md)
- [Integration Testing Guide](frameworks/integration-testing-guide.md)
- [Performance Testing Guide](frameworks/performance-testing-guide.md)
- [Load Testing Guide](frameworks/load-testing-guide.md)
- [Security Testing Guide](frameworks/security-testing-guide.md)
- Network Simulator (see code comments in `pkg/testutil/framework/network_simulator.go`)
- Mock Providers (see code comments in `pkg/testutil/mock/` directory)
- [System Testing Guide](frameworks/system-testing-guide.md)
