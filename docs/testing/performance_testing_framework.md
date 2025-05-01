# Performance Testing Framework

## Overview

The Performance Testing Framework provides utilities for comprehensive performance testing of the OneMount system. It includes tools for benchmarking, load testing, measuring performance metrics, defining performance thresholds, and generating performance reports.

## Components

The Performance Testing Framework consists of the following components:

### 1. Performance Benchmarks

Performance benchmarks measure the performance of specific operations under controlled conditions.

```go
// PerformanceBenchmark defines a performance test
type PerformanceBenchmark struct {
    // Name and description
    Name string
    Description string

    // Setup and teardown functions
    Setup func() error
    Teardown func() error

    // The benchmark function
    BenchmarkFunc func(b *testing.B)

    // Performance thresholds
    thresholds PerformanceThresholds

    // Performance metrics
    metrics PerformanceMetrics

    // Load test configuration
    loadTest *LoadTest
}
```

### 2. Performance Thresholds

Performance thresholds define minimum acceptable performance levels.

```go
// PerformanceThresholds defines minimum acceptable performance levels
type PerformanceThresholds struct {
    MaxLatency      int64   // Maximum acceptable latency in milliseconds
    MinThroughput   int64   // Minimum acceptable operations per second
    MaxMemoryUsage  int64   // Maximum acceptable memory usage in MB
    MaxCPUUsage     float64 // Maximum acceptable CPU usage percentage
}
```

### 3. Performance Metrics

Performance metrics represent performance test metrics.

```go
// PerformanceMetrics represents performance test metrics
type PerformanceMetrics struct {
    Latencies    []time.Duration
    Throughput   float64
    ErrorRate    float64
    ResourceUsage ResourceMetrics
    Custom       map[string]float64
}
```

### 4. Load Tests

Load tests measure performance under various load conditions. See the [Load Testing Framework](load-testing-framework.md) for more details.

## Predefined Performance Benchmarks

The framework provides several predefined performance benchmarks:

1. **FileDownloadBenchmark**: Benchmarks file download performance
2. **FileUploadBenchmark**: Benchmarks file upload performance
3. **MetadataOperationsBenchmark**: Benchmarks metadata operations performance
4. **ConcurrentOperationsBenchmark**: Benchmarks concurrent operations performance

## Usage

### Creating a Performance Benchmark

```go
// Create performance thresholds
thresholds := PerformanceThresholds{
    MaxLatency:     1000, // 1 second
    MinThroughput:  10,   // 10 ops/sec
    MaxMemoryUsage: 1024, // 1 GB
    MaxCPUUsage:    80,   // 80%
}

// Create a new performance benchmark
benchmark := NewPerformanceBenchmark(
    "TestBenchmark",
    "A test benchmark",
    thresholds,
    "/tmp/benchmark-reports",
)

// Set the benchmark function
benchmark.SetBenchmarkFunc(func(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Perform the operation being benchmarked
        _ = someOperation()
    }
})

// Set the setup function
benchmark.SetSetupFunc(func() error {
    // Perform setup operations
    return nil
})

// Set the teardown function
benchmark.SetTeardownFunc(func() error {
    // Perform teardown operations
    return nil
})
```

### Running a Performance Benchmark

```go
// Create a testing.B instance
b := &testing.B{}

// Run the benchmark
benchmark.Run(b)

// Generate a report
err := benchmark.GenerateReport()
if err != nil {
    log.Fatalf("Failed to generate report: %v", err)
}

// Check if the benchmark meets the thresholds
if benchmark.GetMetrics().Throughput < float64(benchmark.GetThresholds().MinThroughput) {
    log.Fatalf("Benchmark failed: throughput %f is below threshold %d", 
        benchmark.GetMetrics().Throughput, benchmark.GetThresholds().MinThroughput)
}
```

### Using Predefined Benchmarks

```go
// Create a test framework
framework := NewTestFramework(TestConfig{
    ArtifactsDir: "test-artifacts",
}, nil)

// Add a mock graph provider
mockGraph := NewMockGraphProvider()
framework.RegisterMockProvider("graph", mockGraph)

// Create a testing.B instance
b := &testing.B{}

// Define performance thresholds
thresholds := PerformanceThresholds{
    MaxLatency:     1000, // 1 second
    MinThroughput:  10,   // 10 ops/sec
    MaxMemoryUsage: 1024, // 1 GB
    MaxCPUUsage:    80,   // 80%
}

// Run a file download benchmark
FileDownloadBenchmark(b, framework, 1024*1024, thresholds, "/tmp/benchmark-reports") // 1MB file

// Run a file upload benchmark
FileUploadBenchmark(b, framework, 1024*1024, thresholds, "/tmp/benchmark-reports") // 1MB file

// Run a metadata operations benchmark
MetadataOperationsBenchmark(b, framework, 100, thresholds, "/tmp/benchmark-reports") // 100 items

// Run a concurrent operations benchmark
ConcurrentOperationsBenchmark(b, framework, 10, thresholds, "/tmp/benchmark-reports") // 10 concurrent operations
```

### Running Load Tests

Load tests are a type of performance test that measure performance under various load conditions. See the [Load Testing Framework](load-testing-framework.md) for more details on load testing.

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

// Define performance thresholds
thresholds := PerformanceThresholds{
    MaxLatency:     1000, // 1 second
    MinThroughput:  10,   // 10 ops/sec
    MaxMemoryUsage: 1024, // 1 GB
    MaxCPUUsage:    80,   // 80%
}

// Create a load test
loadTest := &LoadTest{
    Concurrency: 10,
    Duration:    5 * time.Minute,
    RampUp:      30 * time.Second,
    Scenario: func(ctx context.Context) error {
        // Perform the operation being tested
        return nil
    },
}

// Create a performance benchmark with the load test
benchmark := NewPerformanceBenchmark(
    "LoadTest",
    "A load test",
    thresholds,
    "/tmp/benchmark-reports",
)
benchmark.SetLoadTest(loadTest)

// Run the load test
err := benchmark.RunLoadTest(ctx)
if err != nil {
    log.Fatalf("Failed to run load test: %v", err)
}

// Generate a report
err = benchmark.GenerateReport()
if err != nil {
    log.Fatalf("Failed to generate report: %v", err)
}
```

## Performance Reports

The performance testing framework generates HTML reports with visualizations of the performance test results. The reports include:

1. **Summary**: A summary of the performance test results, including throughput, latency, and resource usage
2. **Performance Results**: Detailed results of the performance test, including latency over time, throughput over time, latency histogram, and error distribution
3. **Detailed Metrics**: Detailed metrics collected during the performance test, including custom metrics and percentile latencies

The reports are saved in the output directory specified when creating the performance benchmark.

## Best Practices

### 1. Benchmark Design

- Design benchmarks to measure specific operations
- Include setup and teardown functions to ensure a clean environment
- Use realistic data sizes and operations
- Benchmark both simple and complex operations

### 2. Performance Thresholds

- Set realistic performance thresholds based on requirements
- Include thresholds for latency, throughput, and resource usage
- Adjust thresholds based on the environment (development, testing, production)
- Regularly review and update thresholds as the system evolves

### 3. Performance Metrics

- Collect a variety of performance metrics
- Include latency, throughput, error rate, and resource usage
- Calculate percentile latencies (p50, p95, p99) for a better understanding of performance distribution
- Track performance metrics over time to identify trends

### 4. Load Testing

- Test with various load patterns (constant, ramp-up, spike, wave, step)
- Include both normal and peak load conditions
- Test with realistic user behavior
- Monitor resource usage during load tests

### 5. Performance Reports

- Generate detailed performance reports
- Include visualizations for better understanding
- Compare results against thresholds
- Track performance changes over time

## Integration with Test Framework

The Performance Testing Framework integrates with the main TestFramework to provide comprehensive performance testing capabilities. It can be used in conjunction with other testing components to create end-to-end test scenarios that include performance testing.

```go
// Create a test framework
testFramework := NewTestFramework(TestConfig{}, nil)

// Create a performance benchmark
benchmark := NewPerformanceBenchmark(
    "IntegratedTest",
    "An integrated performance test",
    PerformanceThresholds{
        MaxLatency:     1000,
        MinThroughput:  10,
        MaxMemoryUsage: 1024,
        MaxCPUUsage:    80,
    },
    "/tmp/benchmark-reports",
)

// Set the benchmark function
benchmark.SetBenchmarkFunc(func(b *testing.B) {
    // Run a test with the framework
    result := testFramework.RunTest("performance-test", func(ctx context.Context) error {
        // Perform the operation being benchmarked
        return nil
    })

    // Verify the result
    if result.Status != TestStatusPassed {
        b.Fatalf("Test failed: %v", result.Failures)
    }
})

// Run the benchmark
b := &testing.B{}
benchmark.Run(b)
```

## Conclusion

The Performance Testing Framework provides a comprehensive set of tools for testing the performance aspects of the OneMount system. By integrating performance testing into the overall testing strategy, we can ensure that the system meets its performance requirements and provides a good user experience.