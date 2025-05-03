# Tutorial: Performance Testing

This tutorial will guide you through the process of writing performance tests using the OneMount test framework. Performance tests help you ensure that your code meets performance requirements and identify potential bottlenecks.

> **Note**: All code examples in this tutorial are for illustration purposes only and may need to be adapted to your specific project structure and imports. The examples are not meant to be compiled directly but rather to demonstrate concepts and patterns.

## Table of Contents

1. [Introduction to Performance Testing](#introduction-to-performance-testing)
2. [The PerformanceBenchmark Component](#the-performancebenchmark-component)
3. [Setting Up Performance Tests](#setting-up-performance-tests)
4. [Measuring Performance Metrics](#measuring-performance-metrics)
5. [Benchmarking Different Operations](#benchmarking-different-operations)
6. [Load Testing](#load-testing)
7. [Best Practices](#best-practices)
8. [Complete Example](#complete-example)

## Introduction to Performance Testing

Performance testing is a critical part of ensuring that your code meets performance requirements. It helps you:

- Identify performance bottlenecks
- Ensure that your code meets performance requirements
- Compare the performance of different implementations
- Track performance changes over time
- Identify regressions in performance

The OneMount test framework provides a PerformanceBenchmark component that helps you write and run performance tests.

## The PerformanceBenchmark Component

The PerformanceBenchmark component provides tools for measuring and analyzing performance metrics. It allows you to:

- Measure execution time
- Track resource usage (CPU, memory, disk I/O, network I/O)
- Collect performance metrics
- Compare metrics against thresholds
- Generate performance reports

You can create a PerformanceBenchmark as follows:

```go
// Create a logger
logger := log.With().Str("component", "benchmark").Logger()

// Create a benchmark configuration
config := testutil.BenchmarkConfig{
    Name:             "FileOperations",
    Iterations:       100,
    WarmupIterations: 10,
    Timeout:          30 * time.Second,
    Thresholds: testutil.PerformanceThresholds{
        LatencyP50: 50 * time.Millisecond,
        LatencyP95: 100 * time.Millisecond,
        LatencyP99: 200 * time.Millisecond,
        Throughput: 100, // operations per second
    },
}

// Create a new PerformanceBenchmark
benchmark := testutil.NewPerformanceBenchmark(config, &logger)
```

## Setting Up Performance Tests

To set up a performance test, you need to:

1. Create a PerformanceBenchmark instance
2. Define the operations to benchmark
3. Run the benchmark
4. Analyze the results

Here's an example:

```go
// Create a benchmark configuration
config := testutil.BenchmarkConfig{
    Name:             "FileOperations",
    Iterations:       100,
    WarmupIterations: 10,
    Timeout:          30 * time.Second,
}

// Create a new PerformanceBenchmark
benchmark := testutil.NewPerformanceBenchmark(config, &logger)

// Define the operation to benchmark
operation := func(ctx context.Context) error {
    // Operation to benchmark
    // For example, writing a file
    data := make([]byte, 1024) // 1KB of data
    rand.Read(data)
    err := ioutil.WriteFile("/tmp/benchmark-test.txt", data, 0644)
    return err
}

// Run the benchmark
results, err := benchmark.Run(operation)
if err != nil {
    // Handle error
}

// Analyze the results
fmt.Printf("Latency (P50): %v\n", results.LatencyP50)
fmt.Printf("Latency (P95): %v\n", results.LatencyP95)
fmt.Printf("Latency (P99): %v\n", results.LatencyP99)
fmt.Printf("Throughput: %v ops/sec\n", results.Throughput)
```

## Measuring Performance Metrics

The PerformanceBenchmark component collects various performance metrics:

1. **Latency**: The time it takes to complete an operation
   - P50 (median): 50% of operations complete within this time
   - P95: 95% of operations complete within this time
   - P99: 99% of operations complete within this time

2. **Throughput**: The number of operations completed per second

3. **Resource Usage**:
   - CPU usage
   - Memory usage
   - Disk I/O
   - Network I/O

You can access these metrics from the benchmark results:

```go
// Run the benchmark
results, err := benchmark.Run(operation)
if err != nil {
    // Handle error
}

// Access latency metrics
fmt.Printf("Latency (P50): %v\n", results.LatencyP50)
fmt.Printf("Latency (P95): %v\n", results.LatencyP95)
fmt.Printf("Latency (P99): %v\n", results.LatencyP99)

// Access throughput metrics
fmt.Printf("Throughput: %v ops/sec\n", results.Throughput)

// Access resource usage metrics
fmt.Printf("CPU Usage: %v%%\n", results.ResourceMetrics.CPUUsage)
fmt.Printf("Memory Usage: %v MB\n", results.ResourceMetrics.MemoryUsage / (1024*1024))
fmt.Printf("Disk Read: %v bytes\n", results.ResourceMetrics.DiskRead)
fmt.Printf("Disk Write: %v bytes\n", results.ResourceMetrics.DiskWrite)
fmt.Printf("Network Received: %v bytes\n", results.ResourceMetrics.NetworkReceived)
fmt.Printf("Network Sent: %v bytes\n", results.ResourceMetrics.NetworkSent)
```

## Benchmarking Different Operations

You can benchmark different operations to compare their performance:

```go
// Define operations to benchmark
operations := map[string]func(ctx context.Context) error{
    "WriteFile": func(ctx context.Context) error {
        data := make([]byte, 1024) // 1KB of data
        rand.Read(data)
        return os.WriteFile("/tmp/benchmark-test.txt", data, 0644)
    },
    "ReadFile": func(ctx context.Context) error {
        _, err := os.ReadFile("/tmp/benchmark-test.txt")
        return err
    },
    "AppendFile": func(ctx context.Context) error {
        data := make([]byte, 128) // 128 bytes of data
        rand.Read(data)
        f, err := os.OpenFile("/tmp/benchmark-test.txt", os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
            return err
        }
        defer f.Close()
        _, err = f.Write(data)
        return err
    },
}

// Run benchmarks for each operation
for name, operation := range operations {
    // Update the benchmark name
    benchmark.SetName(name)

    // Run the benchmark
    results, err := benchmark.Run(operation)
    if err != nil {
        // Handle error
        continue
    }

    // Print results
    fmt.Printf("=== %s ===\n", name)
    fmt.Printf("Latency (P50): %v\n", results.LatencyP50)
    fmt.Printf("Latency (P95): %v\n", results.LatencyP95)
    fmt.Printf("Latency (P99): %v\n", results.LatencyP99)
    fmt.Printf("Throughput: %v ops/sec\n", results.Throughput)
    fmt.Println()
}
```

## Load Testing

Load testing helps you understand how your code behaves under different levels of load. The PerformanceBenchmark component supports load testing with configurable concurrency levels:

```go
// Create a load test configuration
loadConfig := testutil.LoadTestConfig{
    InitialConcurrency: 1,
    MaxConcurrency:     100,
    StepSize:           10,
    DurationPerStep:    10 * time.Second,
    RampUpTime:         1 * time.Second,
    Operation: func(ctx context.Context) error {
        // Operation to benchmark under load
        data := make([]byte, 1024) // 1KB of data
        rand.Read(data)
        return ioutil.WriteFile("/tmp/benchmark-test.txt", data, 0644)
    },
}

// Run the load test
loadResults, err := benchmark.RunLoadTest(loadConfig)
if err != nil {
    // Handle error
}

// Analyze the load test results
for concurrency, result := range loadResults {
    fmt.Printf("=== Concurrency: %d ===\n", concurrency)
    fmt.Printf("Latency (P50): %v\n", result.LatencyP50)
    fmt.Printf("Latency (P95): %v\n", result.LatencyP95)
    fmt.Printf("Latency (P99): %v\n", result.LatencyP99)
    fmt.Printf("Throughput: %v ops/sec\n", result.Throughput)
    fmt.Println()
}

// Find the optimal concurrency level
optimalConcurrency, optimalThroughput := benchmark.FindOptimalConcurrency(loadResults)
fmt.Printf("Optimal Concurrency: %d (Throughput: %v ops/sec)\n", optimalConcurrency, optimalThroughput)
```

The load test increases the concurrency level from `InitialConcurrency` to `MaxConcurrency` in steps of `StepSize`, running each step for `DurationPerStep`. This helps you identify how your code performs under different levels of load and find the optimal concurrency level.

## Best Practices

When writing performance tests, follow these best practices:

1. **Isolate the code being tested**: Ensure that your performance tests measure only the code you're interested in, not external factors.

2. **Use realistic data and operations**: Test with data and operations that reflect real-world usage.

3. **Run multiple iterations**: Run each test multiple times to get statistically significant results.

4. **Include warmup iterations**: Include warmup iterations to allow the system to stabilize before measuring performance.

5. **Set appropriate thresholds**: Set realistic performance thresholds based on your requirements.

6. **Test under different conditions**: Test performance under different conditions, such as different load levels and network conditions.

7. **Track performance over time**: Track performance metrics over time to identify trends and regressions.

8. **Test on representative hardware**: Test on hardware that is representative of your production environment.

9. **Minimize external factors**: Minimize external factors that could affect performance, such as background processes and network activity.

10. **Document performance requirements**: Document your performance requirements and ensure that your tests verify them.

## Complete Example

Here's a complete example of a performance test:

```go
package mypackage_test

import (
    "context"
    "crypto/rand"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/rs/zerolog/log"
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestFileOperationsPerformance(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "benchmark").Logger()

    // Create a benchmark configuration
    config := testutil.BenchmarkConfig{
        Name:             "FileOperations",
        Iterations:       100,
        WarmupIterations: 10,
        Timeout:          30 * time.Second,
        Thresholds: testutil.PerformanceThresholds{
            LatencyP50: 50 * time.Millisecond,
            LatencyP95: 100 * time.Millisecond,
            LatencyP99: 200 * time.Millisecond,
            Throughput: 100, // operations per second
        },
    }

    // Create a new PerformanceBenchmark
    benchmark := testutil.NewPerformanceBenchmark(config, &logger)

    // Create a temporary directory for the test
    tempDir, err := os.MkdirTemp("", "performance-test")
    require.NoError(t, err)
    defer os.RemoveAll(tempDir)

    // Define operations to benchmark
    operations := map[string]func(ctx context.Context) error{
        "WriteFile": func(ctx context.Context) error {
            data := make([]byte, 1024) // 1KB of data
            _, err := rand.Read(data)
            if err != nil {
                return err
            }
            return os.WriteFile(tempDir+"/test.txt", data, 0644)
        },
        "ReadFile": func(ctx context.Context) error {
            _, err := os.ReadFile(tempDir + "/test.txt")
            return err
        },
        "AppendFile": func(ctx context.Context) error {
            data := make([]byte, 128) // 128 bytes of data
            _, err := rand.Read(data)
            if err != nil {
                return err
            }
            f, err := os.OpenFile(tempDir+"/test.txt", os.O_APPEND|os.O_WRONLY, 0644)
            if err != nil {
                return err
            }
            defer f.Close()
            _, err = f.Write(data)
            return err
        },
    }

    // Create the test file for read and append operations
    initialData := make([]byte, 1024)
    _, err = rand.Read(initialData)
    require.NoError(t, err)
    err = os.WriteFile(tempDir+"/test.txt", initialData, 0644)
    require.NoError(t, err)

    // Run benchmarks for each operation
    for name, operation := range operations {
        t.Run(name, func(t *testing.T) {
            // Update the benchmark name
            benchmark.SetName(name)

            // Run the benchmark
            results, err := benchmark.Run(operation)
            require.NoError(t, err)

            // Print results
            t.Logf("=== %s ===", name)
            t.Logf("Latency (P50): %v", results.LatencyP50)
            t.Logf("Latency (P95): %v", results.LatencyP95)
            t.Logf("Latency (P99): %v", results.LatencyP99)
            t.Logf("Throughput: %v ops/sec", results.Throughput)

            // Verify that the results meet the thresholds
            require.LessOrEqual(t, results.LatencyP50, config.Thresholds.LatencyP50, "P50 latency exceeds threshold")
            require.LessOrEqual(t, results.LatencyP95, config.Thresholds.LatencyP95, "P95 latency exceeds threshold")
            require.LessOrEqual(t, results.LatencyP99, config.Thresholds.LatencyP99, "P99 latency exceeds threshold")
            require.GreaterOrEqual(t, results.Throughput, config.Thresholds.Throughput, "Throughput below threshold")
        })
    }

    // Run a load test for the write operation
    t.Run("LoadTest", func(t *testing.T) {
        // Create a load test configuration
        loadConfig := testutil.LoadTestConfig{
            InitialConcurrency: 1,
            MaxConcurrency:     20,
            StepSize:           5,
            DurationPerStep:    5 * time.Second,
            RampUpTime:         1 * time.Second,
            Operation: func(ctx context.Context) error {
                data := make([]byte, 1024) // 1KB of data
                _, err := rand.Read(data)
                if err != nil {
                    return err
                }
                fileName := fmt.Sprintf("%s/load-test-%d.txt", tempDir, time.Now().UnixNano())
                return os.WriteFile(fileName, data, 0644)
            },
        }

        // Run the load test
        loadResults, err := benchmark.RunLoadTest(loadConfig)
        require.NoError(t, err)

        // Analyze the load test results
        for concurrency, result := range loadResults {
            t.Logf("=== Concurrency: %d ===", concurrency)
            t.Logf("Latency (P50): %v", result.LatencyP50)
            t.Logf("Latency (P95): %v", result.LatencyP95)
            t.Logf("Latency (P99): %v", result.LatencyP99)
            t.Logf("Throughput: %v ops/sec", result.Throughput)
        }

        // Find the optimal concurrency level
        optimalConcurrency, optimalThroughput := benchmark.FindOptimalConcurrency(loadResults)
        t.Logf("Optimal Concurrency: %d (Throughput: %v ops/sec)", optimalConcurrency, optimalThroughput)
    })
}
```

This example demonstrates:
1. Setting up a performance benchmark
2. Benchmarking different file operations
3. Verifying that performance meets thresholds
4. Running a load test to find the optimal concurrency level
5. Proper cleanup of test resources

By following these patterns, you can write comprehensive performance tests that ensure your code meets performance requirements and identify potential bottlenecks.
