# Load Testing Framework for OneMount

## Overview

The load testing framework allows you to test the performance of the OneMount system under various load conditions, such as many concurrent users, sustained high load, and load spikes. It provides tools for generating and applying load patterns, collecting metrics under different load conditions, analyzing system behavior under load, and reporting load test results with visualizations.

## Key Concepts

Load testing is a type of performance testing that simulates real-world load on a system to evaluate its behavior under expected and peak load conditions. Key concepts in load testing include:

1. **Concurrency**: The number of simultaneous operations or users interacting with the system.
2. **Load Pattern**: The way load is applied over time (constant, ramping up, spikes, etc.).
3. **Throughput**: The rate at which the system processes operations, typically measured in operations per second.
4. **Latency**: The time it takes for the system to respond to a request, typically measured in milliseconds.
5. **Resource Utilization**: The amount of system resources (CPU, memory, network, etc.) used during the test.
6. **Performance Thresholds**: The minimum acceptable performance levels for the system.

## Architecture

The load testing framework consists of several key components:

### LoadTest

The `LoadTest` struct defines the parameters for a load test:

```go
type LoadTest struct {
    // Number of concurrent operations (used for constant load or as base concurrency for patterns)
    Concurrency int
    // Maximum number of concurrent operations (used for non-constant load patterns)
    MaxConcurrency int
    // Duration of the test
    Duration time.Duration
    // Ramp-up time before measurements
    RampUp time.Duration
    // Test scenario to run
    Scenario func(ctx context.Context) error
    // Load pattern to apply (if nil, constant load is used)
    Pattern *LoadPattern
    // Additional metrics to collect during the test
    AdditionalMetrics map[string]func() float64
}
```

### LoadPattern

The `LoadPattern` struct defines a pattern for applying load during a test:

```go
type LoadPattern struct {
    // Type of load pattern
    Type LoadPatternType
    // Base concurrency level
    BaseConcurrency int
    // Peak concurrency level (for non-constant patterns)
    PeakConcurrency int
    // Duration of the pattern
    Duration time.Duration
    // Additional parameters for specific patterns
    Params map[string]interface{}
}
```

### LoadPatternType

The `LoadPatternType` defines the type of load pattern to apply:

```go
type LoadPatternType string

const (
    // ConstantLoad applies a constant number of concurrent operations
    ConstantLoad LoadPatternType = "constant"
    // RampUpLoad gradually increases the number of concurrent operations
    RampUpLoad LoadPatternType = "ramp-up"
    // SpikeLoad applies a sudden spike in concurrent operations
    SpikeLoad LoadPatternType = "spike"
    // WaveLoad applies a sinusoidal pattern of concurrent operations
    WaveLoad LoadPatternType = "wave"
    // StepLoad increases the number of concurrent operations in steps
    StepLoad LoadPatternType = "step"
)
```

### LoadTestScenario

The `LoadTestScenario` struct defines a load test scenario with a specific pattern and configuration:

```go
type LoadTestScenario struct {
    // Name of the scenario
    Name string
    // Description of the scenario
    Description string
    // Load pattern to apply
    Pattern LoadPattern
    // Test scenario to run
    Scenario func(ctx context.Context) error
    // Additional metrics to collect
    AdditionalMetrics map[string]func() float64
    // Performance thresholds
    Thresholds PerformanceThresholds
}
```

### Predefined Load Test Scenarios

The framework provides several predefined load test scenarios:

1. **ConcurrentUsersScenario**: A scenario with many concurrent users
2. **SustainedHighLoadScenario**: A scenario with sustained high load
3. **LoadSpikeScenario**: A scenario with a sudden spike in load
4. **RampUpLoadScenario**: A scenario with gradually increasing load
5. **WaveLoadScenario**: A scenario with a sinusoidal wave pattern of load
6. **StepLoadScenario**: A scenario with step-wise increasing load

### Predefined Load Test Operations

The framework provides several predefined load test operations:

1. **FileDownloadLoadTest**: Runs load tests for file download operations
2. **FileUploadLoadTest**: Runs load tests for file upload operations
3. **MetadataOperationsLoadTest**: Runs load tests for metadata operations
4. **MixedOperationsLoadTest**: Runs load tests for a mix of operations

## Getting Started

### Running a Simple Load Test

```go
// Create a test framework
framework := NewTestFramework(TestConfig{
    ArtifactsDir: "test-artifacts",
}, nil)

// Add a mock graph provider
mockGraph := NewMockGraphProvider()
framework.RegisterMockProvider("graph", mockGraph)

// Create a context
ctx := context.Background()

// Create a load test scenario
scenario := ConcurrentUsersScenario(50, 1*time.Minute, func(ctx context.Context) error {
    // Simulate a file download
    return simulateFileDownload(1024 * 1024) // 1MB
})

// Run the load test
benchmark, err := RunLoadTestScenario(ctx, framework, scenario)
if err != nil {
    log.Fatalf("Failed to run load test: %v", err)
}

// Print the results
fmt.Printf("Throughput: %.2f ops/sec\n", benchmark.GetMetrics().Throughput)
fmt.Printf("P95 Latency: %.2f ms\n", benchmark.GetMetrics().Custom["p95_latency_ms"])
```

### Running All Load Test Scenarios for File Downloads

```go
// Create a test framework
framework := NewTestFramework(TestConfig{
    ArtifactsDir: "test-artifacts",
}, nil)

// Add a mock graph provider
mockGraph := NewMockGraphProvider()
framework.RegisterMockProvider("graph", mockGraph)

// Create a context
ctx := context.Background()

// Run all load test scenarios for file downloads
benchmarks, err := FileDownloadLoadTest(ctx, framework, 1024*1024) // 1MB
if err != nil {
    log.Fatalf("Failed to run load tests: %v", err)
}

// Print the results
for _, benchmark := range benchmarks {
    fmt.Printf("Scenario: %s\n", benchmark.Name)
    fmt.Printf("Throughput: %.2f ops/sec\n", benchmark.GetMetrics().Throughput)
    fmt.Printf("P95 Latency: %.2f ms\n", benchmark.GetMetrics().Custom["p95_latency_ms"])
    fmt.Println()
}
```

### Creating a Custom Load Test Scenario

```go
// Create a custom load test scenario
scenario := LoadTestScenario{
    Name:        "CustomScenario",
    Description: "A custom load test scenario",
    Pattern: LoadPattern{
        Type:            WaveLoad,
        BaseConcurrency: 10,
        PeakConcurrency: 50,
        Duration:        5 * time.Minute,
        Params: map[string]interface{}{
            "frequency": 2.0,
        },
    },
    Scenario: func(ctx context.Context) error {
        // Custom operation
        return nil
    },
    Thresholds: PerformanceThresholds{
        MaxLatency:     1000, // 1 second
        MinThroughput:  10,   // 10 ops/sec
        MaxMemoryUsage: 1024, // 1 GB
        MaxCPUUsage:    80,   // 80%
    },
}

// Run the custom load test scenario
benchmark, err := RunLoadTestScenario(ctx, framework, scenario)
if err != nil {
    log.Fatalf("Failed to run load test: %v", err)
}
```

### Load Test Reports

The load testing framework generates HTML reports with visualizations of the load test results. The reports include:

1. **Summary**: A summary of the load test results, including throughput, latency, and resource usage
2. **Load Test Results**: Detailed results of the load test, including latency over time, throughput over time, latency histogram, and error distribution
3. **Detailed Metrics**: Detailed metrics collected during the load test, including custom metrics and percentile latencies

The reports are saved in the artifacts directory specified in the test framework configuration.

### Extending the Framework

#### Adding a New Load Pattern

To add a new load pattern, you need to:

1. Add a new constant to the `LoadPatternType` enum
2. Create a new load pattern generator that implements the `LoadPatternGenerator` interface
3. Update the `CreateLoadPatternGenerator` function to handle the new pattern type

#### Adding a New Load Test Scenario

To add a new load test scenario, you can create a new function that returns a `LoadTestScenario` struct with the desired configuration.

#### Adding a New Load Test Operation

To add a new load test operation, you can create a new function that takes a test framework and returns a function that can be used as a scenario in a load test.

## Best Practices

When using the load testing framework, follow these best practices to ensure accurate and meaningful results:

1. **Define Clear Test Objectives**: Clearly define what you want to measure and what success criteria are before starting load testing.

2. **Use Realistic Scenarios**: Create load test scenarios that reflect real-world usage patterns of the OneMount system.

3. **Start with Baseline Tests**: Establish baseline performance metrics before testing with high load or complex patterns.

4. **Isolate Test Environment**: Ensure the test environment is isolated from other activities that might affect performance measurements.

5. **Monitor System Resources**: Always monitor CPU, memory, disk I/O, and network usage during load tests to identify bottlenecks.

6. **Gradually Increase Load**: Start with low load and gradually increase it to identify at what point performance degrades.

7. **Test Different Load Patterns**: Use various load patterns (constant, ramp-up, spike, etc.) to understand system behavior under different conditions.

8. **Analyze Trends Over Time**: Look for performance trends over time, not just peak or average values.

9. **Automate Load Tests**: Incorporate load tests into your CI/CD pipeline to catch performance regressions early.

10. **Document Test Results**: Keep detailed records of test configurations, results, and system behavior for future reference.

## Troubleshooting

When working with the load testing framework, you might encounter these common issues:

### Test Setup Issues

- **Framework initialization fails**: Ensure the test framework is properly configured with valid parameters.
- **Mock providers not working**: Verify that mock providers are registered correctly and configured with appropriate responses.
- **Invalid load pattern configuration**: Check that load pattern parameters are within valid ranges and compatible with each other.

### Test Execution Issues

- **Tests fail to start**: Ensure the test environment has sufficient resources to run the specified load.
- **Unexpected errors during test execution**: Check logs for specific error messages and ensure all dependencies are available.
- **Premature test termination**: Verify that timeouts are set appropriately for the expected test duration.

### Results Analysis Issues

- **Missing or incomplete metrics**: Ensure that metrics collection is properly configured and that the test ran to completion.
- **Unexpected performance results**: Verify that the test environment was isolated and that no external factors affected the results.
- **Report generation fails**: Check that the artifacts directory is writable and has sufficient space.

For more detailed troubleshooting information, see the [Testing Troubleshooting Guide](../testing-troubleshooting.md).

## Related Resources

- [Performance Testing Framework](performance-testing-guide.md)
- [Integration Testing Guide](integration-testing-guide.md)
- [System Testing Framework](system-testing-guide.md)
- Network Simulator (see code comments in `pkg/testutil/framework/network_simulator.go`)
- Mock Providers (see code comments in `pkg/testutil/mock/` directory)
- Testing Framework (see code comments in `pkg/testutil/framework/framework.go`)
- [Test Guidelines](../test-guidelines.md)
