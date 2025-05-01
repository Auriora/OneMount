// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"context"
	"fmt"
	"time"
)

// LoadTestScenario defines a load test scenario with a specific pattern and configuration
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

// RunLoadTestScenario runs a load test scenario and returns the benchmark result
func RunLoadTestScenario(ctx context.Context, framework *TestFramework, scenario LoadTestScenario) (*PerformanceBenchmark, error) {
	// Create a benchmark
	benchmark := NewPerformanceBenchmark(
		scenario.Name,
		scenario.Description,
		scenario.Thresholds,
		framework.Config.ArtifactsDir,
	)

	// Set up the benchmark
	benchmark.SetSetupFunc(func() error {
		// Get the mock graph provider if needed
		mockGraph, exists := framework.GetMockProvider("graph")
		if !exists {
			return fmt.Errorf("mock graph provider not found")
		}

		// Configure the mock as needed
		_ = mockGraph

		return nil
	})

	// Set up the load test
	loadTest := &LoadTest{
		Concurrency:       scenario.Pattern.BaseConcurrency,
		MaxConcurrency:    scenario.Pattern.PeakConcurrency,
		Duration:          scenario.Pattern.Duration,
		RampUp:            10 * time.Second,
		Scenario:          scenario.Scenario,
		Pattern:           &scenario.Pattern,
		AdditionalMetrics: scenario.AdditionalMetrics,
	}

	// Set the load test
	benchmark.SetLoadTest(loadTest)

	// Run the load test
	if err := benchmark.RunLoadTest(ctx); err != nil {
		return nil, err
	}

	return benchmark, nil
}

// ConcurrentUsersScenario creates a load test scenario with many concurrent users
func ConcurrentUsersScenario(concurrency int, duration time.Duration, scenario func(ctx context.Context) error) LoadTestScenario {
	return LoadTestScenario{
		Name:        "ConcurrentUsersScenario",
		Description: fmt.Sprintf("Load test with %d concurrent users for %s", concurrency, duration),
		Pattern: LoadPattern{
			Type:            ConstantLoad,
			BaseConcurrency: concurrency,
			Duration:        duration,
		},
		Scenario: scenario,
		Thresholds: PerformanceThresholds{
			MaxLatency:     1000, // 1 second
			MinThroughput:  10,   // 10 ops/sec
			MaxMemoryUsage: 1024, // 1 GB
			MaxCPUUsage:    80,   // 80%
		},
	}
}

// SustainedHighLoadScenario creates a load test scenario with sustained high load
func SustainedHighLoadScenario(concurrency int, duration time.Duration, scenario func(ctx context.Context) error) LoadTestScenario {
	return LoadTestScenario{
		Name:        "SustainedHighLoadScenario",
		Description: fmt.Sprintf("Sustained high load test with %d concurrent users for %s", concurrency, duration),
		Pattern: LoadPattern{
			Type:            ConstantLoad,
			BaseConcurrency: concurrency,
			Duration:        duration,
		},
		Scenario: scenario,
		Thresholds: PerformanceThresholds{
			MaxLatency:     2000, // 2 seconds
			MinThroughput:  5,    // 5 ops/sec
			MaxMemoryUsage: 2048, // 2 GB
			MaxCPUUsage:    90,   // 90%
		},
	}
}

// LoadSpikeScenario creates a load test scenario with a sudden spike in load
func LoadSpikeScenario(baseConcurrency, peakConcurrency int, duration time.Duration, scenario func(ctx context.Context) error) LoadTestScenario {
	return LoadTestScenario{
		Name:        "LoadSpikeScenario",
		Description: fmt.Sprintf("Load spike test from %d to %d concurrent users for %s", baseConcurrency, peakConcurrency, duration),
		Pattern: LoadPattern{
			Type:            SpikeLoad,
			BaseConcurrency: baseConcurrency,
			PeakConcurrency: peakConcurrency,
			Duration:        duration,
			Params: map[string]interface{}{
				"spikeStart":    duration / 3,
				"spikeDuration": duration / 6,
			},
		},
		Scenario: scenario,
		Thresholds: PerformanceThresholds{
			MaxLatency:     5000, // 5 seconds
			MinThroughput:  1,    // 1 ops/sec
			MaxMemoryUsage: 4096, // 4 GB
			MaxCPUUsage:    95,   // 95%
		},
	}
}

// RampUpLoadScenario creates a load test scenario with gradually increasing load
func RampUpLoadScenario(startConcurrency, endConcurrency int, duration time.Duration, scenario func(ctx context.Context) error) LoadTestScenario {
	return LoadTestScenario{
		Name:        "RampUpLoadScenario",
		Description: fmt.Sprintf("Ramp-up load test from %d to %d concurrent users over %s", startConcurrency, endConcurrency, duration),
		Pattern: LoadPattern{
			Type:            RampUpLoad,
			BaseConcurrency: startConcurrency,
			PeakConcurrency: endConcurrency,
			Duration:        duration,
		},
		Scenario: scenario,
		Thresholds: PerformanceThresholds{
			MaxLatency:     3000, // 3 seconds
			MinThroughput:  2,    // 2 ops/sec
			MaxMemoryUsage: 3072, // 3 GB
			MaxCPUUsage:    90,   // 90%
		},
	}
}

// WaveLoadScenario creates a load test scenario with a sinusoidal wave pattern of load
func WaveLoadScenario(minConcurrency, maxConcurrency int, duration time.Duration, frequency float64, scenario func(ctx context.Context) error) LoadTestScenario {
	return LoadTestScenario{
		Name:        "WaveLoadScenario",
		Description: fmt.Sprintf("Wave load test oscillating between %d and %d concurrent users with frequency %.1f over %s", minConcurrency, maxConcurrency, frequency, duration),
		Pattern: LoadPattern{
			Type:            WaveLoad,
			BaseConcurrency: minConcurrency,
			PeakConcurrency: maxConcurrency,
			Duration:        duration,
			Params: map[string]interface{}{
				"frequency": frequency,
			},
		},
		Scenario: scenario,
		Thresholds: PerformanceThresholds{
			MaxLatency:     2500, // 2.5 seconds
			MinThroughput:  3,    // 3 ops/sec
			MaxMemoryUsage: 2048, // 2 GB
			MaxCPUUsage:    85,   // 85%
		},
	}
}

// StepLoadScenario creates a load test scenario with step-wise increasing load
func StepLoadScenario(startConcurrency, endConcurrency, steps int, duration time.Duration, scenario func(ctx context.Context) error) LoadTestScenario {
	return LoadTestScenario{
		Name:        "StepLoadScenario",
		Description: fmt.Sprintf("Step load test from %d to %d concurrent users in %d steps over %s", startConcurrency, endConcurrency, steps, duration),
		Pattern: LoadPattern{
			Type:            StepLoad,
			BaseConcurrency: startConcurrency,
			PeakConcurrency: endConcurrency,
			Duration:        duration,
			Params: map[string]interface{}{
				"steps": steps,
			},
		},
		Scenario: scenario,
		Thresholds: PerformanceThresholds{
			MaxLatency:     4000, // 4 seconds
			MinThroughput:  2,    // 2 ops/sec
			MaxMemoryUsage: 3072, // 3 GB
			MaxCPUUsage:    90,   // 90%
		},
	}
}

// RunAllLoadTestScenarios runs all load test scenarios with the given operation scenario
func RunAllLoadTestScenarios(ctx context.Context, framework *TestFramework, operationScenario func(ctx context.Context) error) ([]*PerformanceBenchmark, error) {
	// Define the scenarios
	scenarios := []LoadTestScenario{
		ConcurrentUsersScenario(50, 1*time.Minute, operationScenario),
		SustainedHighLoadScenario(100, 2*time.Minute, operationScenario),
		LoadSpikeScenario(20, 200, 3*time.Minute, operationScenario),
		RampUpLoadScenario(10, 100, 2*time.Minute, operationScenario),
		WaveLoadScenario(20, 80, 3*time.Minute, 3.0, operationScenario),
		StepLoadScenario(10, 100, 5, 2*time.Minute, operationScenario),
	}

	// Run all scenarios
	results := make([]*PerformanceBenchmark, 0, len(scenarios))
	for _, scenario := range scenarios {
		fmt.Printf("Running load test scenario: %s\n", scenario.Name)
		benchmark, err := RunLoadTestScenario(ctx, framework, scenario)
		if err != nil {
			return results, fmt.Errorf("failed to run scenario %s: %v", scenario.Name, err)
		}
		results = append(results, benchmark)
	}

	return results, nil
}

// FileDownloadLoadTest runs load tests for file download operations
func FileDownloadLoadTest(ctx context.Context, framework *TestFramework, fileSize int64) ([]*PerformanceBenchmark, error) {
	// Create a scenario function for file download
	downloadScenario := func(ctx context.Context) error {
		return simulateFileDownload(fileSize)
	}

	// Run all load test scenarios with the download scenario
	return RunAllLoadTestScenarios(ctx, framework, downloadScenario)
}

// FileUploadLoadTest runs load tests for file upload operations
func FileUploadLoadTest(ctx context.Context, framework *TestFramework, fileSize int64) ([]*PerformanceBenchmark, error) {
	// Create a scenario function for file upload
	uploadScenario := func(ctx context.Context) error {
		return simulateFileUpload(fileSize)
	}

	// Run all load test scenarios with the upload scenario
	return RunAllLoadTestScenarios(ctx, framework, uploadScenario)
}

// MetadataOperationsLoadTest runs load tests for metadata operations
func MetadataOperationsLoadTest(ctx context.Context, framework *TestFramework, numItems int) ([]*PerformanceBenchmark, error) {
	// Create a scenario function for metadata operations
	metadataScenario := func(ctx context.Context) error {
		return simulateMetadataOperations(numItems)
	}

	// Run all load test scenarios with the metadata scenario
	return RunAllLoadTestScenarios(ctx, framework, metadataScenario)
}

// MixedOperationsLoadTest runs load tests for a mix of operations
func MixedOperationsLoadTest(ctx context.Context, framework *TestFramework) ([]*PerformanceBenchmark, error) {
	// Create a scenario function for mixed operations
	mixedScenario := func(ctx context.Context) error {
		// Randomly choose an operation
		switch randomInt(3) {
		case 0:
			return simulateFileDownload(1024 * 1024) // 1MB
		case 1:
			return simulateFileUpload(1024 * 1024) // 1MB
		case 2:
			return simulateMetadataOperations(10)
		default:
			return nil
		}
	}

	// Run all load test scenarios with the mixed scenario
	return RunAllLoadTestScenarios(ctx, framework, mixedScenario)
}

// Helper function to generate a random integer between 0 and max-1
func randomInt(max int) int {
	return int(time.Now().UnixNano() % int64(max))
}
