# Performance and Load Testing Guide

## Overview

This guide covers the performance and load testing framework for OneMount, specifically focusing on the four key performance testing requirements from the release action plan:

1. **Large File Handling Tests** - Tests for files >1GB
2. **High File Count Directory Tests** - Tests for directories with >10k files  
3. **Sustained Operation Tests** - Long-running tests to verify stability
4. **Memory Leak Detection Tests** - Tests to detect memory leaks over time

## Quick Start

### Running Performance Tests

```bash
# Run all performance tests (requires significant time and resources)
go test -v ./internal/testutil/framework -run "TestPerformanceIntegration"

# Run specific performance tests
go test -v ./internal/testutil/framework -run "TestPerformanceIntegration_LargeFileHandling"
go test -v ./internal/testutil/framework -run "TestPerformanceIntegration_HighFileCount"
go test -v ./internal/testutil/framework -run "TestPerformanceIntegration_SustainedOperation"
go test -v ./internal/testutil/framework -run "TestPerformanceIntegration_MemoryLeakDetection"

# Run unit tests for the performance framework
go test -v ./internal/testutil/framework -run "TestLargeFileHandling|TestHighFileCountDirectory" -short
```

### Skipping Resource-Intensive Tests

Performance tests are automatically skipped in short mode to prevent resource exhaustion during regular testing:

```bash
# This will skip the actual performance tests
go test -short ./internal/testutil/framework
```

## Test Types

### 1. Large File Handling Tests

Tests the system's ability to handle files larger than 1GB efficiently.

**Configuration:**
```go
config := LargeFileTestConfig{
    FileSize:    1*1024*1024*1024 + 1, // 1GB + 1 byte
    Concurrency: 2,
    Timeout:     10 * time.Minute,
    ChunkSize:   64 * 1024 * 1024, // 64MB chunks
}
```

**What it tests:**
- Large file upload performance
- Large file download performance
- Chunk-based operations
- Memory usage during large file operations

**Usage:**
```go
func BenchmarkLargeFileHandling(b *testing.B) {
    framework := NewTestFramework(config, logger)
    config := DefaultLargeFileTestConfig()
    thresholds := PerformanceThresholds{
        MaxLatency:     300000, // 5 minutes
        MinThroughput:  1,      // 1 ops/sec
        MaxMemoryUsage: 2048,   // 2GB
        MaxCPUUsage:    90.0,   // 90%
    }
    
    LargeFileHandlingTest(b, framework, config, thresholds)
}
```

### 2. High File Count Directory Tests

Tests the system's ability to handle directories with more than 10,000 files.

**Configuration:**
```go
config := HighFileCountTestConfig{
    FileCount:   10001, // >10k files
    Concurrency: 10,
    Timeout:     15 * time.Minute,
    FileSize:    1024, // 1KB per file
}
```

**What it tests:**
- Directory creation with many files
- Directory listing performance
- File cleanup operations
- Concurrent file operations

**Usage:**
```go
func BenchmarkHighFileCount(b *testing.B) {
    framework := NewTestFramework(config, logger)
    config := DefaultHighFileCountTestConfig()
    thresholds := PerformanceThresholds{
        MaxLatency:     600000, // 10 minutes
        MinThroughput:  1,      // 1 ops/sec
        MaxMemoryUsage: 1024,   // 1GB
        MaxCPUUsage:    85.0,   // 85%
    }
    
    HighFileCountDirectoryTest(b, framework, config, thresholds)
}
```

### 3. Sustained Operation Tests

Tests the system's stability and performance during extended operation periods.

**Configuration:**
```go
config := SustainedOperationTestConfig{
    Duration:            30 * time.Minute,
    OperationsPerSecond: 10,
    Workers:             5,
    MemoryCheckInterval: 1 * time.Minute,
}
```

**What it tests:**
- System stability over time
- Performance consistency
- Memory usage patterns
- Resource utilization

**Usage:**
```go
func BenchmarkSustainedOperation(b *testing.B) {
    framework := NewTestFramework(config, logger)
    config := DefaultSustainedOperationTestConfig()
    thresholds := PerformanceThresholds{
        MaxLatency:     10000, // 10 seconds
        MinThroughput:  5,     // 5 ops/sec
        MaxMemoryUsage: 512,   // 512MB
        MaxCPUUsage:    75.0,  // 75%
    }
    
    SustainedOperationTest(b, framework, config, thresholds)
}
```

### 4. Memory Leak Detection Tests

Tests for memory leaks during extended operation cycles.

**Configuration:**
```go
config := MemoryLeakTestConfig{
    Duration:           20 * time.Minute,
    SamplingInterval:   30 * time.Second,
    MaxMemoryGrowthMB:  100, // 100MB max growth
    OperationsPerCycle: 100,
    CycleInterval:      1 * time.Minute,
}
```

**What it tests:**
- Memory growth over time
- Garbage collection effectiveness
- Memory leak detection
- Resource cleanup

**Usage:**
```go
func BenchmarkMemoryLeakDetection(b *testing.B) {
    framework := NewTestFramework(config, logger)
    config := DefaultMemoryLeakTestConfig()
    thresholds := PerformanceThresholds{
        MaxLatency:     5000, // 5 seconds
        MinThroughput:  10,   // 10 ops/sec
        MaxMemoryUsage: 512,  // 512MB
        MaxCPUUsage:    70.0, // 70%
    }
    
    MemoryLeakDetectionTest(b, framework, config, thresholds)
}
```

## Memory Tracking

The framework includes a comprehensive memory tracking system:

```go
// Create a memory tracker
tracker := NewMemoryTracker()

// Take memory samples
tracker.Sample()

// Get all samples
samples := tracker.GetSamples()

// Calculate memory growth
growth, err := tracker.GetMemoryGrowth()
```

**Memory Sample Structure:**
```go
type MemorySample struct {
    Timestamp time.Time
    AllocMB   int64  // Allocated memory in MB
    SysMB     int64  // System memory in MB
    HeapMB    int64  // Heap memory in MB
    StackMB   int64  // Stack memory in MB
}
```

## Performance Thresholds

Configure performance thresholds to define acceptable limits:

```go
thresholds := PerformanceThresholds{
    MaxLatency:     30000, // Maximum latency in milliseconds
    MinThroughput:  1,     // Minimum operations per second
    MaxMemoryUsage: 2048,  // Maximum memory usage in MB
    MaxCPUUsage:    90.0,  // Maximum CPU usage percentage
}
```

## Test Reports

Performance tests automatically generate detailed reports:

- **HTML Reports**: Visual performance dashboards
- **JSON Reports**: Machine-readable performance data
- **Metrics History**: Time-series performance data

Reports are stored in:
- `~/.onemount-tests/tmp/performance_reports/`
- `~/.onemount-tests/tmp/metrics_history/`

## Best Practices

### 1. Resource Management
- Run performance tests on dedicated hardware when possible
- Ensure sufficient disk space for large file tests
- Monitor system resources during test execution

### 2. Test Configuration
- Adjust timeouts based on system capabilities
- Use appropriate concurrency levels for your hardware
- Set realistic performance thresholds

### 3. Test Environment
- Run tests in isolated environments
- Avoid running other resource-intensive applications
- Use consistent hardware configurations for comparison

### 4. Continuous Integration
- Skip performance tests in regular CI runs using `-short` flag
- Run performance tests in dedicated CI pipelines
- Set up performance regression detection

## Troubleshooting

### Common Issues

**Test Timeouts:**
```bash
# Increase timeout in test configuration
config.Timeout = 20 * time.Minute
```

**Memory Issues:**
```bash
# Reduce file sizes or counts for testing
config.FileSize = 100 * 1024 * 1024  // 100MB instead of 1GB
config.FileCount = 1000               // 1k instead of 10k
```

**Disk Space:**
```bash
# Check available disk space before running tests
df -h /tmp
```

### Performance Debugging

Enable verbose logging for detailed performance insights:

```go
framework := NewTestFramework(TestConfig{
    VerboseLogging: true,
    // ... other config
}, logger)
```

## Integration with Release Action Plan

These performance tests directly address the requirements from the OneMount release action plan:

- ✅ **Large file handling (>1GB)** - `LargeFileHandlingTest`
- ✅ **Directory with many files (>10k files)** - `HighFileCountDirectoryTest`  
- ✅ **Sustained operation over time** - `SustainedOperationTest`
- ✅ **Memory leak detection tests** - `MemoryLeakDetectionTest`

All tests include comprehensive reporting, threshold validation, and integration with the existing OneMount test framework.
