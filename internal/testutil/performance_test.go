// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPerformanceBenchmark(t *testing.T) {
	// Create a temporary directory for test artifacts
	tempDir, err := os.MkdirTemp("", "performance-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a benchmark
	thresholds := PerformanceThresholds{
		MaxLatency:     100,  // 100ms
		MinThroughput:  10,   // 10 ops/sec
		MaxMemoryUsage: 100,  // 100MB
		MaxCPUUsage:    50.0, // 50%
	}

	benchmark := NewPerformanceBenchmark(
		"TestBenchmark",
		"Test benchmark for unit testing",
		thresholds,
		tempDir,
	)

	// Set a simple benchmark function
	benchmark.SetBenchmarkFunc(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
		}
	})

	// Run the benchmark with a testing.B
	b := &testing.B{}
	benchmark.Run(b)

	// Check that metrics were collected
	metrics := benchmark.GetMetrics()
	assert.True(t, metrics.Throughput > 0, "Throughput should be greater than 0")
	assert.True(t, metrics.ResourceUsage.MemoryUsage > 0, "Memory usage should be greater than 0")

	// Check that reports were generated
	htmlReport := filepath.Join(tempDir, "performance_reports", "TestBenchmark_report.html")
	jsonReport := filepath.Join(tempDir, "performance_reports", "TestBenchmark_report.json")

	_, err = os.Stat(htmlReport)
	assert.NoError(t, err, "HTML report should exist")

	_, err = os.Stat(jsonReport)
	assert.NoError(t, err, "JSON report should exist")
}

func TestLoadTest(t *testing.T) {
	// Create a temporary directory for test artifacts
	tempDir, err := os.MkdirTemp("", "load-test-")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a benchmark
	thresholds := PerformanceThresholds{
		MaxLatency:     100,  // 100ms
		MinThroughput:  10,   // 10 ops/sec
		MaxMemoryUsage: 100,  // 100MB
		MaxCPUUsage:    50.0, // 50%
	}

	benchmark := NewPerformanceBenchmark(
		"TestLoadTest",
		"Test load test for unit testing",
		thresholds,
		tempDir,
	)

	// Set up a load test
	loadTest := &LoadTest{
		Concurrency: 5,
		Duration:    1 * time.Second,
		RampUp:      100 * time.Millisecond,
		Scenario: func(ctx context.Context) error {
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	}

	benchmark.SetLoadTest(loadTest)

	// Run the load test
	ctx := context.Background()
	err = benchmark.RunLoadTest(ctx)
	assert.NoError(t, err, "Load test should not return an error")

	// Check that metrics were collected
	metrics := benchmark.GetMetrics()
	assert.True(t, metrics.Throughput > 0, "Throughput should be greater than 0")
	assert.True(t, len(metrics.Latencies) > 0, "Latencies should be recorded")
	assert.Equal(t, 0.0, metrics.ErrorRate, "Error rate should be 0")

	// Check that reports were generated
	htmlReport := filepath.Join(tempDir, "performance_reports", "TestLoadTest_report.html")
	jsonReport := filepath.Join(tempDir, "performance_reports", "TestLoadTest_report.json")

	_, err = os.Stat(htmlReport)
	assert.NoError(t, err, "HTML report should exist")

	_, err = os.Stat(jsonReport)
	assert.NoError(t, err, "JSON report should exist")
}

func TestBenchmarkScenarios(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping benchmark scenarios test in short mode")
	}

	// Create a test logger
	logger := &testLogger{t: t}

	// Create a test framework
	framework := NewTestFramework(TestConfig{
		Environment:    "test",
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   os.TempDir(),
	}, logger)

	// Register a mock graph provider
	mockGraph := &mockGraphProvider{}
	framework.RegisterMockProvider("graph", mockGraph)

	// Define thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:     1000, // 1000ms
		MinThroughput:  1,    // 1 ops/sec
		MaxMemoryUsage: 200,  // 200MB
		MaxCPUUsage:    80.0, // 80%
	}

	// Run file download benchmark
	t.Run("FileDownloadBenchmark", func(t *testing.T) {
		b := &testing.B{}
		b.N = 5                                                    // Run only 5 iterations for testing
		FileDownloadBenchmark(b, framework, 1024*1024, thresholds) // 1MB file
	})

	// Run file upload benchmark
	t.Run("FileUploadBenchmark", func(t *testing.T) {
		b := &testing.B{}
		b.N = 5                                                  // Run only 5 iterations for testing
		FileUploadBenchmark(b, framework, 1024*1024, thresholds) // 1MB file
	})

	// Run metadata operations benchmark
	t.Run("MetadataOperationsBenchmark", func(t *testing.T) {
		b := &testing.B{}
		b.N = 5                                                   // Run only 5 iterations for testing
		MetadataOperationsBenchmark(b, framework, 10, thresholds) // 10 items
	})

	// Run concurrent operations benchmark
	t.Run("ConcurrentOperationsBenchmark", func(t *testing.T) {
		b := &testing.B{}
		b.N = 5                                                    // Run only 5 iterations for testing
		ConcurrentOperationsBenchmark(b, framework, 5, thresholds) // 5 concurrent operations
	})

	// Run load tests
	ctx := context.Background()

	t.Run("LoadTestFileDownload", func(t *testing.T) {
		err := LoadTestFileDownload(ctx, framework, 1024*1024, 2, 1*time.Second, thresholds)
		assert.NoError(t, err, "Load test should not return an error")
	})

	t.Run("LoadTestFileUpload", func(t *testing.T) {
		err := LoadTestFileUpload(ctx, framework, 1024*1024, 2, 1*time.Second, thresholds)
		assert.NoError(t, err, "Load test should not return an error")
	})

	t.Run("LoadTestMetadataOperations", func(t *testing.T) {
		err := LoadTestMetadataOperations(ctx, framework, 10, 2, 1*time.Second, thresholds)
		assert.NoError(t, err, "Load test should not return an error")
	})
}

// Mock implementation of the MockProvider interface for testing
type mockGraphProvider struct{}

func (m *mockGraphProvider) Setup() error {
	return nil
}

func (m *mockGraphProvider) Teardown() error {
	return nil
}

func (m *mockGraphProvider) Reset() error {
	return nil
}

// testLogger is a simple implementation of the Logger interface for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, args ...interface{}) {
	l.t.Logf("DEBUG: "+msg, args...)
}

func (l *testLogger) Info(msg string, args ...interface{}) {
	l.t.Logf("INFO: "+msg, args...)
}

func (l *testLogger) Warn(msg string, args ...interface{}) {
	l.t.Logf("WARN: "+msg, args...)
}

func (l *testLogger) Error(msg string, args ...interface{}) {
	l.t.Logf("ERROR: "+msg, args...)
}
