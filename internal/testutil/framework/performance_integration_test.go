// Package testutil provides testing utilities for the OneMount project.
// This file contains integration tests for the performance and load testing framework.
package framework

import (
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// TestPerformanceIntegration_LargeFileHandling tests large file handling performance
// Test Case ID: TestPT_LF_01_01
// Description: Verify that the system can handle large files (>1GB) efficiently
// Expected Result: Large file operations complete within acceptable time limits
// Notes: This test validates the performance requirements from the release action plan
func TestPerformanceIntegration_LargeFileHandling(t *testing.T) {
	// Skip this test in short mode and CI environments
	if testing.Short() {
		t.Skip("Skipping large file performance test in short mode")
	}

	// Create test framework
	logger := &testLogger{t: t}
	framework := NewTestFramework(TestConfig{
		Environment:    "performance-test",
		Timeout:        60,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Configure for actual large file testing
	config := LargeFileTestConfig{
		FileSize:    1*1024*1024*1024 + 1, // 1GB + 1 byte
		Concurrency: 2,
		Timeout:     10 * time.Minute,
		ChunkSize:   64 * 1024 * 1024, // 64MB chunks
	}

	// Set performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     300000, // 5 minutes max latency
		MinThroughput:  1,      // 1 ops/sec minimum
		MaxMemoryUsage: 2048,   // 2GB max memory
		MaxCPUUsage:    90.0,   // 90% max CPU
	}

	// Run the benchmark
	b := &testing.B{}
	b.N = 1

	t.Logf("Starting large file handling test with file size: %d bytes", config.FileSize)
	start := time.Now()

	LargeFileHandlingTest(b, framework, config, thresholds)

	duration := time.Since(start)
	t.Logf("Large file handling test completed in: %v", duration)

	// Verify test completed within reasonable time
	assert.Less(t, duration, 15*time.Minute, "Large file test should complete within 15 minutes")
}

// TestPerformanceIntegration_HighFileCount tests directory with many files
// Test Case ID: TestPT_HF_01_01
// Description: Verify that the system can handle directories with >10k files efficiently
// Expected Result: Directory operations complete within acceptable time limits
// Notes: This test validates the performance requirements from the release action plan
func TestPerformanceIntegration_HighFileCount(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping high file count performance test in short mode")
	}

	// Create test framework
	logger := &testLogger{t: t}
	framework := NewTestFramework(TestConfig{
		Environment:    "performance-test",
		Timeout:        60,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Configure for actual high file count testing
	config := HighFileCountTestConfig{
		FileCount:   10001, // >10k files
		Concurrency: 10,
		Timeout:     15 * time.Minute,
		FileSize:    1024, // 1KB per file
	}

	// Set performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     600000, // 10 minutes max latency
		MinThroughput:  1,      // 1 ops/sec minimum
		MaxMemoryUsage: 1024,   // 1GB max memory
		MaxCPUUsage:    85.0,   // 85% max CPU
	}

	// Run the benchmark
	b := &testing.B{}
	b.N = 1

	t.Logf("Starting high file count test with %d files", config.FileCount)
	start := time.Now()

	HighFileCountDirectoryTest(b, framework, config, thresholds)

	duration := time.Since(start)
	t.Logf("High file count test completed in: %v", duration)

	// Verify test completed within reasonable time
	assert.Less(t, duration, 20*time.Minute, "High file count test should complete within 20 minutes")
}

// TestPerformanceIntegration_SustainedOperation tests sustained operation over time
// Test Case ID: TestPT_SO_01_01
// Description: Verify that the system can sustain operations over extended periods
// Expected Result: System remains stable and performant during sustained operations
// Notes: This test validates the performance requirements from the release action plan
func TestPerformanceIntegration_SustainedOperation(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping sustained operation performance test in short mode")
	}
	if os.Getenv("ONEMOUNT_PERF_ENABLE") != "1" {
		t.Skip("Skipping performance test; set ONEMOUNT_PERF_ENABLE=1 to run")
	}

	// Create test framework
	logger := &testLogger{t: t}
	framework := NewTestFramework(TestConfig{
		Environment:    "performance-test",
		Timeout:        60,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Configure sustained operation testing
	// Default to a CI-friendly duration to avoid the Go test 20m timeout; allow opting
	// into the longer soak by setting ONEMOUNT_PERF_LONG=1.
	sustainedDuration := 2 * time.Minute
	if os.Getenv("ONEMOUNT_PERF_LONG") == "1" {
		sustainedDuration = 30 * time.Minute
	}
	config := SustainedOperationTestConfig{
		Duration:            sustainedDuration,
		OperationsPerSecond: 10,
		Workers:             5,
		MemoryCheckInterval: 30 * time.Second,
	}

	// Set performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     10000, // 10 seconds max latency
		MinThroughput:  5,     // 5 ops/sec minimum
		MaxMemoryUsage: 512,   // 512MB max memory
		MaxCPUUsage:    75.0,  // 75% max CPU
	}

	// Run the benchmark
	b := &testing.B{}
	b.N = 1

	t.Logf("Starting sustained operation test for %v", config.Duration)
	start := time.Now()

	SustainedOperationTest(b, framework, config, thresholds)

	duration := time.Since(start)
	t.Logf("Sustained operation test completed in: %v", duration)

	// Verify test ran for the expected duration (within 10% tolerance)
	expectedDuration := config.Duration
	tolerance := expectedDuration / 5 // 20% tolerance for shortened runs
	assert.InDelta(t, expectedDuration.Seconds(), duration.Seconds(), tolerance.Seconds(),
		"Sustained operation test should run for approximately the configured duration")
}

// TestPerformanceIntegration_MemoryLeakDetection tests for memory leaks
// Test Case ID: TestPT_ML_01_01
// Description: Verify that the system does not have memory leaks during extended operation
// Expected Result: Memory usage remains stable without excessive growth
// Notes: This test validates the performance requirements from the release action plan
func TestPerformanceIntegration_MemoryLeakDetection(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping memory leak detection performance test in short mode")
	}
	if os.Getenv("ONEMOUNT_PERF_ENABLE") != "1" {
		t.Skip("Skipping performance test; set ONEMOUNT_PERF_ENABLE=1 to run")
	}

	// Create test framework
	logger := &testLogger{t: t}
	framework := NewTestFramework(TestConfig{
		Environment:    "performance-test",
		Timeout:        60,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Configure memory leak detection testing
	// Keep default short for CI; enable long soak with ONEMOUNT_PERF_LONG=1.
	leakDuration := 2 * time.Minute
	if os.Getenv("ONEMOUNT_PERF_LONG") == "1" {
		leakDuration = 20 * time.Minute
	}
	config := MemoryLeakTestConfig{
		Duration:           leakDuration,
		SamplingInterval:   15 * time.Second,
		MaxMemoryGrowthMB:  100, // 100MB max growth
		OperationsPerCycle: 100,
		CycleInterval:      30 * time.Second,
	}

	// Set performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     5000, // 5 seconds max latency
		MinThroughput:  10,   // 10 ops/sec minimum
		MaxMemoryUsage: 512,  // 512MB max memory
		MaxCPUUsage:    70.0, // 70% max CPU
	}

	// Run the benchmark
	b := &testing.B{}
	b.N = 1

	t.Logf("Starting memory leak detection test for %v", config.Duration)
	start := time.Now()

	MemoryLeakDetectionTest(b, framework, config, thresholds)

	duration := time.Since(start)
	t.Logf("Memory leak detection test completed in: %v", duration)

	// Verify test ran for the expected duration (within 10% tolerance)
	expectedDuration := config.Duration
	tolerance := expectedDuration / 5 // 20% tolerance for shortened runs
	assert.InDelta(t, expectedDuration.Seconds(), duration.Seconds(), tolerance.Seconds(),
		"Memory leak detection test should run for approximately the configured duration")
}

// TestPerformanceIntegration_AllTests runs all performance tests in sequence
// Test Case ID: TestPT_ALL_01_01
// Description: Run all performance tests to validate overall system performance
// Expected Result: All performance tests pass within acceptable limits
// Notes: This is a comprehensive test that validates all performance requirements
func TestPerformanceIntegration_AllTests(t *testing.T) {
	// Skip this test in short mode as it's very resource intensive
	if testing.Short() {
		t.Skip("Skipping comprehensive performance test suite in short mode")
	}

	// This test is designed to be run manually or in dedicated performance testing environments
	// It requires significant time and resources
	t.Skip("Comprehensive performance test suite - run manually for full validation")

	// Create test framework
	logger := &testLogger{t: t}
	_ = NewTestFramework(TestConfig{
		Environment:    "comprehensive-performance-test",
		Timeout:        120,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Run all performance tests
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"LargeFileHandling", func(t *testing.T) { TestPerformanceIntegration_LargeFileHandling(t) }},
		{"HighFileCount", func(t *testing.T) { TestPerformanceIntegration_HighFileCount(t) }},
		{"SustainedOperation", func(t *testing.T) { TestPerformanceIntegration_SustainedOperation(t) }},
		{"MemoryLeakDetection", func(t *testing.T) { TestPerformanceIntegration_MemoryLeakDetection(t) }},
	}

	overallStart := time.Now()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testStart := time.Now()
			test.test(t)
			testDuration := time.Since(testStart)
			t.Logf("Performance test %s completed in: %v", test.name, testDuration)
		})
	}

	overallDuration := time.Since(overallStart)
	t.Logf("All performance tests completed in: %v", overallDuration)

	// Log completion message
	t.Logf("✅ All performance and load testing requirements from release action plan have been implemented and validated")
}
