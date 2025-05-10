// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"context"
	"github.com/auriora/onemount/pkg/testutil/mock"
	"testing"
	"time"
)

func TestLoadTestScenarios(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping load test scenarios in short mode")
	}

	// Create a test framework
	framework := NewTestFramework(TestConfig{
		// ArtifactsDir will be set to the default value by NewTestFramework
	}, nil) // Using nil logger for simplicity

	// Add a mock graph provider
	mockGraph := mock.NewMockGraphProvider()
	framework.RegisterMockProvider("graph", mockGraph)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test ConcurrentUsersScenario
	t.Run("ConcurrentUsersScenario", func(t *testing.T) {
		scenario := ConcurrentUsersScenario(10, 5*time.Second, func(ctx context.Context) error {
			// Simulate a simple operation
			time.Sleep(10 * time.Millisecond)
			return nil
		})

		benchmark, err := RunLoadTestScenario(ctx, framework, scenario)
		if err != nil {
			t.Fatalf("Failed to run ConcurrentUsersScenario: %v", err)
		}

		// Verify the benchmark results
		if benchmark.metrics.Throughput <= 0 {
			t.Errorf("Expected positive throughput, got %.2f", benchmark.metrics.Throughput)
		}
		if len(benchmark.metrics.Latencies) == 0 {
			t.Errorf("Expected latencies to be recorded")
		}
	})

	// Test LoadSpikeScenario
	t.Run("LoadSpikeScenario", func(t *testing.T) {
		scenario := LoadSpikeScenario(5, 20, 5*time.Second, func(ctx context.Context) error {
			// Simulate a simple operation
			time.Sleep(10 * time.Millisecond)
			return nil
		})

		benchmark, err := RunLoadTestScenario(ctx, framework, scenario)
		if err != nil {
			t.Fatalf("Failed to run LoadSpikeScenario: %v", err)
		}

		// Verify the benchmark results
		if benchmark.metrics.Throughput <= 0 {
			t.Errorf("Expected positive throughput, got %.2f", benchmark.metrics.Throughput)
		}
		if len(benchmark.metrics.Latencies) == 0 {
			t.Errorf("Expected latencies to be recorded")
		}
	})

	// Test RampUpLoadScenario
	t.Run("RampUpLoadScenario", func(t *testing.T) {
		scenario := RampUpLoadScenario(5, 15, 5*time.Second, func(ctx context.Context) error {
			// Simulate a simple operation
			time.Sleep(10 * time.Millisecond)
			return nil
		})

		benchmark, err := RunLoadTestScenario(ctx, framework, scenario)
		if err != nil {
			t.Fatalf("Failed to run RampUpLoadScenario: %v", err)
		}

		// Verify the benchmark results
		if benchmark.metrics.Throughput <= 0 {
			t.Errorf("Expected positive throughput, got %.2f", benchmark.metrics.Throughput)
		}
		if len(benchmark.metrics.Latencies) == 0 {
			t.Errorf("Expected latencies to be recorded")
		}
	})

	// Test FileDownloadLoadTest with a small file size and short duration
	t.Run("FileDownloadLoadTest", func(t *testing.T) {
		// Use a very small file size and only test one scenario for unit testing
		fileSize := int64(1024) // 1KB

		// Create a custom scenario for testing
		scenario := ConcurrentUsersScenario(5, 2*time.Second, func(ctx context.Context) error {
			return simulateFileDownload(fileSize)
		})

		benchmark, err := RunLoadTestScenario(ctx, framework, scenario)
		if err != nil {
			t.Fatalf("Failed to run FileDownloadLoadTest: %v", err)
		}

		// Verify the benchmark results
		if benchmark.metrics.Throughput <= 0 {
			t.Errorf("Expected positive throughput, got %.2f", benchmark.metrics.Throughput)
		}
		if len(benchmark.metrics.Latencies) == 0 {
			t.Errorf("Expected latencies to be recorded")
		}
	})
}
