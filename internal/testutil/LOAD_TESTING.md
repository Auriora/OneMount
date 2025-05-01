# Load Testing Framework for OneMount

This document describes the load testing framework for the OneMount project, which is part of the test utilities package.

## Overview

The load testing framework allows you to test the performance of the OneMount system under various load conditions, such as many concurrent users, sustained high load, and load spikes. It provides tools for generating and applying load patterns, collecting metrics under different load conditions, analyzing system behavior under load, and reporting load test results with visualizations.

## Key Components

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

## Predefined Load Test Scenarios

The framework provides several predefined load test scenarios:

1. **ConcurrentUsersScenario**: A scenario with many concurrent users
2. **SustainedHighLoadScenario**: A scenario with sustained high load
3. **LoadSpikeScenario**: A scenario with a sudden spike in load
4. **RampUpLoadScenario**: A scenario with gradually increasing load
5. **WaveLoadScenario**: A scenario with a sinusoidal wave pattern of load
6. **StepLoadScenario**: A scenario with step-wise increasing load

## Predefined Load Test Operations

The framework provides several predefined load test operations:

1. **FileDownloadLoadTest**: Runs load tests for file download operations
2. **FileUploadLoadTest**: Runs load tests for file upload operations
3. **MetadataOperationsLoadTest**: Runs load tests for metadata operations
4. **MixedOperationsLoadTest**: Runs load tests for a mix of operations

## Usage Examples

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

## Load Test Reports

The load testing framework generates HTML reports with visualizations of the load test results. The reports include:

1. **Summary**: A summary of the load test results, including throughput, latency, and resource usage
2. **Load Test Results**: Detailed results of the load test, including latency over time, throughput over time, latency histogram, and error distribution
3. **Detailed Metrics**: Detailed metrics collected during the load test, including custom metrics and percentile latencies

The reports are saved in the artifacts directory specified in the test framework configuration.

## Extending the Framework

### Adding a New Load Pattern

To add a new load pattern, you need to:

1. Add a new constant to the `LoadPatternType` enum
2. Create a new load pattern generator that implements the `LoadPatternGenerator` interface
3. Update the `CreateLoadPatternGenerator` function to handle the new pattern type

### Adding a New Load Test Scenario

To add a new load test scenario, you can create a new function that returns a `LoadTestScenario` struct with the desired configuration.

### Adding a New Load Test Operation

To add a new load test operation, you can create a new function that takes a test framework and returns a function that can be used as a scenario in a load test.

## Conclusion

The load testing framework provides a flexible and extensible way to test the performance of the OneMount system under various load conditions. It allows you to define custom load patterns and scenarios, collect and analyze metrics, and generate reports with visualizations.