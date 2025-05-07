# OneMount Test Framework Documentation

## Overview

This documentation provides a comprehensive guide to the OneMount test framework, which supports various types of testing for the OneMount filesystem. The framework is designed to make it easy to write effective tests that verify the behavior and performance of the system under different conditions.

## Table of Contents

1. [Test Framework Architecture](#test-framework-architecture)
2. [Core Components](#core-components)
3. [Test Types](#test-types)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)
6. [Examples](#examples)
7. [Contributing](#contributing)

## Test Framework Architecture

The OneMount test framework is built around a core `TestFramework` class that provides centralized test configuration and execution. This core framework is extended by specialized frameworks for different types of testing:

![Test Framework Architecture](../resources/test-framework-architecture.png)

The architecture is designed to be modular and extensible, allowing for easy addition of new test types and components.

For a detailed overview of the architecture, see [Testing Framework Overview](frameworks/testing-framework-overview.md).

## Core Components

### TestFramework

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

For detailed documentation, see [Testing Framework Overview](frameworks/testing-framework-overview.md).

### NetworkSimulator

The `NetworkSimulator` allows simulating different network conditions for testing, including latency, packet loss, bandwidth limitations, and network disconnections.

```go
// Set network conditions
framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000)

// Simulate network disconnection
framework.DisconnectNetwork()
```

For detailed documentation, see [Network Simulator](components/network-simulator-guide.md).

### MockProviders

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

For detailed documentation, see [Mock Providers](components/mock-providers-guide.md).

### IntegrationTestEnvironment

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

## Test Types

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

```go
// Create a performance benchmark
benchmark := testutil.NewPerformanceBenchmark(config)

// Run the benchmark
result := benchmark.Run(func(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Operation to benchmark
    }
})
```

For detailed documentation, see [Performance Testing Guide](frameworks/performance-testing-guide.md).

### Load Testing

Load testing focuses on verifying that the system can handle expected load and stress conditions. The OneMount test framework provides utilities for load testing, including concurrent user simulation, sustained load testing, and load spike testing.

```go
// Create a load test scenario
scenario := testutil.LoadTestScenario{
    Name:        "Concurrent Users",
    Description: "Tests performance with many concurrent users",
    Concurrency: 100,
    Duration:    5 * time.Minute,
    // ...
}

// Run the load test
result := testutil.RunLoadTestScenario(ctx, framework, scenario)
```

For detailed documentation, see [Load Testing Guide](frameworks/load-testing-guide.md).

### Security Testing

Security testing focuses on verifying that the system meets security requirements and is resistant to security threats. The OneMount test framework provides utilities for security testing, including vulnerability scanning, security control verification, and authentication/authorization testing.

```go
// Create a security test framework
securityFramework := testutil.NewSecurityTestFramework(config)

// Run security tests
securityScenarios := testutil.NewSecurityTestScenarios(securityFramework)
env.AddScenario(securityScenarios.AuthenticationTestScenario())
```

For detailed documentation, see [Security Testing Guide](frameworks/security-testing-guide.md).

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

If you encounter issues while using the test framework, check the [Troubleshooting Guide](troubleshooting.md) for solutions to common problems, including:

- Tests fail to start
- Mock providers not working correctly
- Network simulation issues
- Resource cleanup failures
- Integration test environment issues
- Test scenario execution issues
- Performance testing issues
- Security testing issues

## Examples

For examples of how to use the test framework for different types of testing, see the respective documentation:

- [Unit Testing Examples](frameworks/unit-testing-guide.md#examples)
- [Integration Testing Examples](frameworks/integration-testing-guide.md#examples)
- [Performance Testing Examples](frameworks/performance-testing-guide.md#examples)
- [Load Testing Examples](frameworks/load-testing-guide.md#examples)
- [Security Testing Examples](frameworks/security-testing-guide.md#examples)

## Contributing

For guidelines on contributing to the test framework, see the [Contributing Guide](../CONTRIBUTING.md).

## Related Resources

- [Test Architecture Design](../../2-architecture-and-design/test-architecture-design.md)
- [Test Guidelines](test-guidelines.md)
- [Test Sandbox Guidelines](components/test-sandbox-guide.md)