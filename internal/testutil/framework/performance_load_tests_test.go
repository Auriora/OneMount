// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"testing"
	"time"

	"github.com/auriora/onemount/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// TestLargeFileHandling tests the large file handling performance test
func TestUT_Framework_PerformanceLoad_LargeFileHandling(t *testing.T) {
	// Skip this test in short mode as it's resource intensive
	if testing.Short() {
		t.Skip("Skipping large file handling test in short mode")
	}

	// Create a test logger
	logger := &testLogger{t: t}

	// Create a test framework
	framework := NewTestFramework(TestConfig{
		Environment:    "test",
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Use smaller file size for testing (1MB instead of 1GB)
	config := LargeFileTestConfig{
		FileSize:    1 * 1024 * 1024, // 1MB for testing
		Concurrency: 1,
		Timeout:     2 * time.Minute,
		ChunkSize:   64 * 1024, // 64KB chunks
	}

	// Define performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     30000, // 30 seconds
		MinThroughput:  1,     // 1 ops/sec
		MaxMemoryUsage: 500,   // 500MB
		MaxCPUUsage:    80.0,  // 80%
	}

	// Run the test
	t.Run("LargeFileHandlingTest", func(t *testing.T) {
		// Create a benchmark test
		b := &testing.B{}
		b.N = 1 // Run once for testing

		// This should not panic or fail
		assert.NotPanics(t, func() {
			LargeFileHandlingTest(b, framework, config, thresholds)
		})
	})
}

// TestHighFileCountDirectory tests the high file count directory performance test
func TestUT_Framework_PerformanceLoad_HighFileCountDirectory(t *testing.T) {
	// Skip this test in short mode as it's resource intensive
	if testing.Short() {
		t.Skip("Skipping high file count directory test in short mode")
	}

	// Create a test logger
	logger := &testLogger{t: t}

	// Create a test framework
	framework := NewTestFramework(TestConfig{
		Environment:    "test",
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Use smaller file count for testing (100 instead of 10k)
	config := HighFileCountTestConfig{
		FileCount:   100, // 100 files for testing
		Concurrency: 5,
		Timeout:     2 * time.Minute,
		FileSize:    1024, // 1KB per file
	}

	// Define performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     30000, // 30 seconds
		MinThroughput:  1,     // 1 ops/sec
		MaxMemoryUsage: 500,   // 500MB
		MaxCPUUsage:    80.0,  // 80%
	}

	// Run the test
	t.Run("HighFileCountDirectoryTest", func(t *testing.T) {
		// Create a benchmark test
		b := &testing.B{}
		b.N = 1 // Run once for testing

		// This should not panic or fail
		assert.NotPanics(t, func() {
			HighFileCountDirectoryTest(b, framework, config, thresholds)
		})
	})
}

// TestSustainedOperation tests the sustained operation performance test
func TestUT_Framework_PerformanceLoad_SustainedOperation(t *testing.T) {
	// Skip this test in short mode as it's time intensive
	if testing.Short() {
		t.Skip("Skipping sustained operation test in short mode")
	}

	// Create a test logger
	logger := &testLogger{t: t}

	// Create a test framework
	framework := NewTestFramework(TestConfig{
		Environment:    "test",
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Use shorter duration for testing (30 seconds instead of 30 minutes)
	config := SustainedOperationTestConfig{
		Duration:            30 * time.Second, // 30 seconds for testing
		OperationsPerSecond: 5,
		Workers:             2,
		MemoryCheckInterval: 5 * time.Second,
	}

	// Define performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     5000, // 5 seconds
		MinThroughput:  1,    // 1 ops/sec
		MaxMemoryUsage: 200,  // 200MB
		MaxCPUUsage:    80.0, // 80%
	}

	// Run the test
	t.Run("SustainedOperationTest", func(t *testing.T) {
		// Create a benchmark test
		b := &testing.B{}
		b.N = 1 // Run once for testing

		// This should not panic or fail
		assert.NotPanics(t, func() {
			SustainedOperationTest(b, framework, config, thresholds)
		})
	})
}

// TestMemoryLeakDetection tests the memory leak detection performance test
func TestUT_Framework_PerformanceLoad_MemoryLeakDetection(t *testing.T) {
	// Skip this test in short mode as it's time intensive
	if testing.Short() {
		t.Skip("Skipping memory leak detection test in short mode")
	}

	// Create a test logger
	logger := &testLogger{t: t}

	// Create a test framework
	framework := NewTestFramework(TestConfig{
		Environment:    "test",
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   testutil.TestSandboxTmpDir,
	}, logger)

	// Use shorter duration for testing (1 minute instead of 20 minutes)
	config := MemoryLeakTestConfig{
		Duration:           1 * time.Minute, // 1 minute for testing
		SamplingInterval:   5 * time.Second,
		MaxMemoryGrowthMB:  50, // 50MB max growth
		OperationsPerCycle: 10,
		CycleInterval:      10 * time.Second,
	}

	// Define performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     5000, // 5 seconds
		MinThroughput:  1,    // 1 ops/sec
		MaxMemoryUsage: 200,  // 200MB
		MaxCPUUsage:    80.0, // 80%
	}

	// Run the test
	t.Run("MemoryLeakDetectionTest", func(t *testing.T) {
		// Create a benchmark test
		b := &testing.B{}
		b.N = 1 // Run once for testing

		// This should not panic or fail
		assert.NotPanics(t, func() {
			MemoryLeakDetectionTest(b, framework, config, thresholds)
		})
	})
}

// TestMemoryTracker tests the memory tracker functionality
func TestUT_Framework_PerformanceLoad_MemoryTracker(t *testing.T) {
	tracker := NewMemoryTracker()

	// Take initial sample
	tracker.Sample()
	time.Sleep(100 * time.Millisecond)

	// Take another sample
	tracker.Sample()

	// Get samples
	samples := tracker.GetSamples()
	assert.Len(t, samples, 2, "Should have 2 samples")

	// Check that timestamps are different
	assert.True(t, samples[1].Timestamp.After(samples[0].Timestamp), "Second sample should be after first")

	// Get memory growth
	growth, err := tracker.GetMemoryGrowth()
	assert.NoError(t, err, "Should not error when calculating growth")
	assert.GreaterOrEqual(t, growth, int64(-10), "Memory growth should be reasonable (allowing for small decreases)")
}

// TestDefaultConfigurations tests that default configurations are reasonable
func TestUT_Framework_PerformanceLoad_DefaultConfigurations(t *testing.T) {
	t.Run("LargeFileTestConfig", func(t *testing.T) {
		config := DefaultLargeFileTestConfig()
		assert.Greater(t, config.FileSize, int64(1024*1024*1024), "Should be > 1GB")
		assert.Greater(t, config.Concurrency, 0, "Should have positive concurrency")
		assert.Greater(t, config.Timeout, time.Duration(0), "Should have positive timeout")
		assert.Greater(t, config.ChunkSize, int64(0), "Should have positive chunk size")
	})

	t.Run("HighFileCountTestConfig", func(t *testing.T) {
		config := DefaultHighFileCountTestConfig()
		assert.Greater(t, config.FileCount, 10000, "Should be > 10k files")
		assert.Greater(t, config.Concurrency, 0, "Should have positive concurrency")
		assert.Greater(t, config.Timeout, time.Duration(0), "Should have positive timeout")
		assert.Greater(t, config.FileSize, int64(0), "Should have positive file size")
	})

	t.Run("SustainedOperationTestConfig", func(t *testing.T) {
		config := DefaultSustainedOperationTestConfig()
		assert.Greater(t, config.Duration, time.Duration(0), "Should have positive duration")
		assert.Greater(t, config.OperationsPerSecond, 0, "Should have positive ops/sec")
		assert.Greater(t, config.Workers, 0, "Should have positive workers")
		assert.Greater(t, config.MemoryCheckInterval, time.Duration(0), "Should have positive memory check interval")
	})

	t.Run("MemoryLeakTestConfig", func(t *testing.T) {
		config := DefaultMemoryLeakTestConfig()
		assert.Greater(t, config.Duration, time.Duration(0), "Should have positive duration")
		assert.Greater(t, config.SamplingInterval, time.Duration(0), "Should have positive sampling interval")
		assert.Greater(t, config.MaxMemoryGrowthMB, int64(0), "Should have positive max memory growth")
		assert.Greater(t, config.OperationsPerCycle, 0, "Should have positive operations per cycle")
		assert.Greater(t, config.CycleInterval, time.Duration(0), "Should have positive cycle interval")
	})
}
