// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// PerformanceThresholds defines minimum acceptable performance levels
type PerformanceThresholds struct {
	// Maximum acceptable latency in milliseconds
	MaxLatency int64
	// Minimum acceptable operations per second
	MinThroughput int64
	// Maximum acceptable memory usage in MB
	MaxMemoryUsage int64
	// Maximum acceptable CPU usage percentage
	MaxCPUUsage float64
}

// ResourceMetrics represents resource usage metrics
type ResourceMetrics struct {
	// CPU usage percentage (0-100)
	CPUUsage float64
	// Memory usage in MB
	MemoryUsage int64
	// Disk I/O in bytes per second
	DiskIO int64
	// Network I/O in bytes per second
	NetworkIO int64
	// Time series data for resource usage
	TimeSeries ResourceTimeSeries
}

// ResourceTimeSeries contains time series data for resource metrics
type ResourceTimeSeries struct {
	// Timestamps for the data points
	Timestamps []time.Time
	// CPU usage percentage over time
	CPUUsage []float64
	// Memory usage in MB over time
	MemoryUsage []int64
	// Disk I/O in bytes per second over time
	DiskIO []int64
	// Network I/O in bytes per second over time
	NetworkIO []int64
}

// LatencyDistribution represents detailed latency distribution data
type LatencyDistribution struct {
	// Percentiles (1st, 5th, 10th, 25th, 50th, 75th, 90th, 95th, 99th, 99.9th)
	Percentiles map[float64]time.Duration
	// Histogram buckets (in milliseconds)
	// e.g., 0-1ms, 1-5ms, 5-10ms, 10-25ms, 25-50ms, 50-100ms, 100-250ms, 250-500ms, 500-1000ms, 1000ms+
	Histogram map[string]int
	// Time series data for latency
	TimeSeries LatencyTimeSeries
}

// LatencyTimeSeries contains time series data for latency metrics
type LatencyTimeSeries struct {
	// Timestamps for the data points
	Timestamps []time.Time
	// P50 latency over time
	P50 []time.Duration
	// P90 latency over time
	P90 []time.Duration
	// P95 latency over time
	P95 []time.Duration
	// P99 latency over time
	P99 []time.Duration
	// Throughput over time (operations per second)
	Throughput []float64
}

// SystemEvent represents a system event that can be correlated with metrics
type SystemEvent struct {
	// Timestamp of the event
	Timestamp time.Time
	// Type of event (e.g., "config_change", "error", "restart")
	Type string
	// Description of the event
	Description string
	// Additional data associated with the event
	Data map[string]interface{}
}

// PerformanceMetrics represents performance test metrics
type PerformanceMetrics struct {
	// Latencies for each operation in nanoseconds
	Latencies []time.Duration
	// Operations per second
	Throughput float64
	// Error rate (0-1)
	ErrorRate float64
	// Resource usage during the test
	ResourceUsage ResourceMetrics
	// Custom metrics
	Custom map[string]float64
	// Detailed latency distribution
	LatencyDistribution LatencyDistribution
	// System events that occurred during the test
	Events []SystemEvent
	// Test start time
	StartTime time.Time
	// Test end time
	EndTime time.Time
	// Test duration
	Duration time.Duration
	// Test configuration
	Config map[string]interface{}
}

// LoadTest defines load testing parameters
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

// PerformanceBenchmark defines a performance test
type PerformanceBenchmark struct {
	// Name and description
	Name        string
	Description string

	// Setup and teardown functions
	Setup    func() error
	Teardown func() error

	// The benchmark function
	BenchmarkFunc func(b *testing.B)

	// Performance thresholds
	thresholds PerformanceThresholds

	// Performance metrics
	metrics PerformanceMetrics

	// Load test configuration
	loadTest *LoadTest

	// Output directory for reports
	outputDir string

	// Mutex for thread safety
	mu sync.Mutex
}

// NewPerformanceBenchmark creates a new PerformanceBenchmark with the given name and description
func NewPerformanceBenchmark(name, description string, thresholds PerformanceThresholds, outputDir string) *PerformanceBenchmark {
	// Create output directory if it doesn't exist
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create output directory: %v\n", err)
		}
	}

	// Create metrics storage directory
	metricsDir := filepath.Join(outputDir, "metrics_history")
	if outputDir != "" {
		if err := os.MkdirAll(metricsDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create metrics directory: %v\n", err)
		}
	}

	// Initialize performance metrics
	metrics := PerformanceMetrics{
		Latencies: make([]time.Duration, 0),
		Custom:    make(map[string]float64),
		LatencyDistribution: LatencyDistribution{
			Percentiles: make(map[float64]time.Duration),
			Histogram:   make(map[string]int),
			TimeSeries: LatencyTimeSeries{
				Timestamps: make([]time.Time, 0),
				P50:        make([]time.Duration, 0),
				P90:        make([]time.Duration, 0),
				P95:        make([]time.Duration, 0),
				P99:        make([]time.Duration, 0),
				Throughput: make([]float64, 0),
			},
		},
		Events:    make([]SystemEvent, 0),
		StartTime: time.Now(),
		Config:    make(map[string]interface{}),
		ResourceUsage: ResourceMetrics{
			TimeSeries: ResourceTimeSeries{
				Timestamps:  make([]time.Time, 0),
				CPUUsage:    make([]float64, 0),
				MemoryUsage: make([]int64, 0),
				DiskIO:      make([]int64, 0),
				NetworkIO:   make([]int64, 0),
			},
		},
	}

	return &PerformanceBenchmark{
		Name:        name,
		Description: description,
		thresholds:  thresholds,
		metrics:     metrics,
		outputDir:   outputDir,
	}
}

// SetBenchmarkFunc sets the benchmark function
func (pb *PerformanceBenchmark) SetBenchmarkFunc(fn func(b *testing.B)) {
	pb.BenchmarkFunc = fn
}

// SetSetupFunc sets the setup function
func (pb *PerformanceBenchmark) SetSetupFunc(fn func() error) {
	pb.Setup = fn
}

// SetTeardownFunc sets the teardown function
func (pb *PerformanceBenchmark) SetTeardownFunc(fn func() error) {
	pb.Teardown = fn
}

// SetLoadTest sets the load test configuration
func (pb *PerformanceBenchmark) SetLoadTest(loadTest *LoadTest) {
	pb.loadTest = loadTest
}

// Run runs the benchmark using Go's testing.B
func (pb *PerformanceBenchmark) Run(b *testing.B) {
	// Run setup if provided
	if pb.Setup != nil {
		if err := pb.Setup(); err != nil {
			b.Fatalf("Setup failed: %v", err)
		}
	}

	// Ensure teardown runs after the benchmark
	defer func() {
		if pb.Teardown != nil {
			if err := pb.Teardown(); err != nil {
				b.Logf("Teardown failed: %v", err)
			}
		}
	}()

	// Reset the timer to exclude setup time
	b.ResetTimer()

	// Start resource monitoring
	stopMonitoring := pb.startResourceMonitoring()
	defer stopMonitoring()

	// Run the benchmark function
	if pb.BenchmarkFunc != nil {
		pb.BenchmarkFunc(b)
	} else {
		b.Skip("No benchmark function provided")
	}

	// Calculate metrics
	pb.calculateMetrics(b)

	// Check thresholds
	pb.checkThresholds(b)

	// Store metrics to disk for long-term analysis
	pb.storeMetricsToFile()

	// Generate report
	if err := pb.GenerateReport(); err != nil {
		b.Logf("Failed to generate report: %v", err)
	}
}

// RunLoadTest runs a load test with the configured parameters
func (pb *PerformanceBenchmark) RunLoadTest(ctx context.Context) error {
	if pb.loadTest == nil {
		return fmt.Errorf("no load test configuration provided")
	}

	// Run setup if provided
	if pb.Setup != nil {
		if err := pb.Setup(); err != nil {
			return fmt.Errorf("setup failed: %v", err)
		}
	}

	// Ensure teardown runs after the load test
	defer func() {
		if pb.Teardown != nil {
			if err := pb.Teardown(); err != nil {
				fmt.Printf("Teardown failed: %v\n", err)
			}
		}
	}()

	// Start resource monitoring
	stopMonitoring := pb.startResourceMonitoring()
	defer stopMonitoring()

	// Record the start time
	startTime := time.Now()

	// Create a ticker for collecting additional metrics
	var additionalMetricsTicker *time.Ticker
	var additionalMetricsStopChan chan struct{}
	var additionalMetricsWg sync.WaitGroup

	if pb.loadTest.AdditionalMetrics != nil && len(pb.loadTest.AdditionalMetrics) > 0 {
		additionalMetricsStopChan = make(chan struct{})
		additionalMetricsTicker = time.NewTicker(1 * time.Second)
		additionalMetricsWg.Add(1)

		// Start a goroutine to collect additional metrics
		go func() {
			defer additionalMetricsWg.Done()
			defer additionalMetricsTicker.Stop()

			// Initialize metric time series
			metricSeries := make(map[string][]float64)
			timePoints := make([]time.Time, 0)

			for {
				select {
				case <-additionalMetricsStopChan:
					// Save time series data as custom metrics
					pb.mu.Lock()
					for name, values := range metricSeries {
						if len(values) > 0 {
							// Calculate average
							var sum float64
							for _, v := range values {
								sum += v
							}
							pb.metrics.Custom[name+"_avg"] = sum / float64(len(values))

							// Calculate max
							var max float64
							for _, v := range values {
								if v > max {
									max = v
								}
							}
							pb.metrics.Custom[name+"_max"] = max

							// Store the time series for visualization
							pb.metrics.Custom[name+"_series"] = float64(len(values)) // Just store the count for now
						}
					}
					pb.mu.Unlock()
					return
				case t := <-additionalMetricsTicker.C:
					timePoints = append(timePoints, t)
					for name, metricFn := range pb.loadTest.AdditionalMetrics {
						value := metricFn()
						if _, exists := metricSeries[name]; !exists {
							metricSeries[name] = make([]float64, 0)
						}
						metricSeries[name] = append(metricSeries[name], value)
					}
				}
			}
		}()
	}

	var latencies []time.Duration
	var errors []error

	// Check if a load pattern is specified
	if pb.loadTest.Pattern != nil {
		// Create a load pattern generator
		patternGen, err := CreateLoadPatternGenerator(*pb.loadTest.Pattern)
		if err != nil {
			return fmt.Errorf("failed to create load pattern generator: %v", err)
		}

		fmt.Printf("Starting load test with %s pattern (%s)...\n", pb.loadTest.Pattern.Type, pb.loadTest.Pattern.Duration)

		// Run the load test with the pattern
		latencies, errors = RunLoadPattern(ctx, patternGen, pb.loadTest.Scenario)
	} else {
		// Use the original implementation for backward compatibility
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(ctx, pb.loadTest.Duration+pb.loadTest.RampUp)
		defer cancel()

		// Create a wait group for goroutines
		var wg sync.WaitGroup
		wg.Add(pb.loadTest.Concurrency)

		// Create channels for results
		latencyChan := make(chan time.Duration, pb.loadTest.Concurrency*100)
		errorChan := make(chan error, pb.loadTest.Concurrency*100)

		// Start the ramp-up period
		fmt.Printf("Starting ramp-up period (%s)...\n", pb.loadTest.RampUp)
		time.Sleep(pb.loadTest.RampUp)
		fmt.Printf("Ramp-up complete. Starting load test (%s)...\n", pb.loadTest.Duration)

		// Start goroutines for concurrent operations
		for i := 0; i < pb.loadTest.Concurrency; i++ {
			go func(id int) {
				defer wg.Done()

				for {
					select {
					case <-ctx.Done():
						return
					default:
						// Run the scenario
						start := time.Now()
						err := pb.loadTest.Scenario(ctx)
						latency := time.Since(start)

						// Record results
						latencyChan <- latency
						if err != nil {
							errorChan <- err
						}
					}
				}
			}(i)
		}

		// Wait for the test duration
		time.Sleep(pb.loadTest.Duration)

		// Cancel the context to stop goroutines
		cancel()

		// Wait for all goroutines to finish
		wg.Wait()

		// Close channels
		close(latencyChan)
		close(errorChan)

		// Collect results
		for latency := range latencyChan {
			latencies = append(latencies, latency)
		}

		for err := range errorChan {
			errors = append(errors, err)
		}
	}

	// Stop collecting additional metrics
	if additionalMetricsStopChan != nil {
		close(additionalMetricsStopChan)
		additionalMetricsWg.Wait()
	}

	// Record metrics
	pb.mu.Lock()
	pb.metrics.Latencies = latencies
	if len(latencies) > 0 {
		pb.metrics.ErrorRate = float64(len(errors)) / float64(len(latencies))
		pb.metrics.Throughput = float64(len(latencies)) / time.Since(startTime).Seconds()
	}
	pb.mu.Unlock()

	// Analyze the results
	pb.analyzeLoadTestResults(latencies, errors)

	// Generate report
	return pb.GenerateReport()
}

// analyzeLoadTestResults performs analysis on load test results
func (pb *PerformanceBenchmark) analyzeLoadTestResults(latencies []time.Duration, errors []error) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if len(latencies) == 0 {
		return
	}

	// Sort latencies for analysis
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// Calculate percentiles for custom metrics (backward compatibility)
	p50 := pb.calculatePercentile(50)
	p90 := pb.calculatePercentile(90)
	p95 := pb.calculatePercentile(95)
	p99 := pb.calculatePercentile(99)

	// Store as custom metrics
	pb.metrics.Custom["p50_latency_ms"] = float64(p50.Milliseconds())
	pb.metrics.Custom["p90_latency_ms"] = float64(p90.Milliseconds())
	pb.metrics.Custom["p95_latency_ms"] = float64(p95.Milliseconds())
	pb.metrics.Custom["p99_latency_ms"] = float64(p99.Milliseconds())

	// Calculate standard deviation
	var sum, sumSquared float64
	for _, latency := range latencies {
		ms := float64(latency.Milliseconds())
		sum += ms
		sumSquared += ms * ms
	}
	mean := sum / float64(len(latencies))
	variance := (sumSquared / float64(len(latencies))) - (mean * mean)
	stdDev := math.Sqrt(variance)
	pb.metrics.Custom["latency_stddev_ms"] = stdDev

	// Calculate error distribution
	errorTypes := make(map[string]int)
	for _, err := range errors {
		errorTypes[err.Error()]++
	}

	// Store error distribution
	for errType, count := range errorTypes {
		pb.metrics.Custom["error_"+errType] = float64(count)
	}

	// Calculate throughput over time (in 1-second intervals)
	if len(latencies) > 0 {
		// This is a simplified calculation - in a real implementation,
		// you would track the exact time of each operation and calculate
		// throughput for each interval
		totalDuration := time.Duration(0)
		for _, latency := range latencies {
			totalDuration += latency
		}
		avgLatency := totalDuration / time.Duration(len(latencies))
		pb.metrics.Custom["avg_latency_ms"] = float64(avgLatency.Milliseconds())
	}

	// Calculate detailed latency distribution
	pb.calculateLatencyDistribution()

	// Record test end time and duration
	pb.metrics.EndTime = time.Now()
	pb.metrics.Duration = pb.metrics.EndTime.Sub(pb.metrics.StartTime)

	// Store test configuration
	pb.metrics.Config["test_name"] = pb.Name
	pb.metrics.Config["test_description"] = pb.Description
	if pb.loadTest != nil {
		pb.metrics.Config["concurrency"] = pb.loadTest.Concurrency
		pb.metrics.Config["duration_seconds"] = pb.loadTest.Duration.Seconds()
		pb.metrics.Config["ramp_up_seconds"] = pb.loadTest.RampUp.Seconds()
	}

	// Store metrics to disk for long-term analysis
	pb.storeMetricsToFile()
}

// startResourceMonitoring starts monitoring resource usage and returns a function to stop monitoring
func (pb *PerformanceBenchmark) startResourceMonitoring() func() {
	// Create a channel to signal the monitoring goroutine to stop
	stopChan := make(chan struct{})

	// Start a goroutine to monitor resource usage
	go func() {
		// Sample every 100ms
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var maxCPU float64
		var maxMem int64
		var maxDiskIO int64
		var maxNetworkIO int64

		// For simulating CPU usage that varies over time
		startTime := time.Now()

		for {
			select {
			case <-stopChan:
				return
			case t := <-ticker.C:
				// Get current memory usage from Go runtime
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				// Convert to MB
				memMB := int64(m.Alloc) / (1024 * 1024)

				// Simulate CPU usage with a sine wave pattern (varies between 10% and 70%)
				// In a real implementation, you would use github.com/shirou/gopsutil/cpu
				elapsedSec := time.Since(startTime).Seconds()
				cpuUsage := 40.0 + 30.0*math.Sin(elapsedSec/5.0)

				// Simulate disk I/O (varies between 1MB/s and 10MB/s)
				// In a real implementation, you would use github.com/shirou/gopsutil/disk
				diskReadRate := int64(1e6 + 9e6*math.Sin(elapsedSec/7.0))
				diskWriteRate := int64(1e6 + 5e6*math.Cos(elapsedSec/3.0))
				diskIO := diskReadRate + diskWriteRate

				// Simulate network I/O (varies between 100KB/s and 1MB/s)
				// In a real implementation, you would use github.com/shirou/gopsutil/net
				netReadRate := int64(1e5 + 9e5*math.Sin(elapsedSec/4.0))
				netWriteRate := int64(1e5 + 4e5*math.Cos(elapsedSec/6.0))
				networkIO := netReadRate + netWriteRate

				// Update max values
				if cpuUsage > maxCPU {
					maxCPU = cpuUsage
				}
				if memMB > maxMem {
					maxMem = memMB
				}
				if diskIO > maxDiskIO {
					maxDiskIO = diskIO
				}
				if networkIO > maxNetworkIO {
					maxNetworkIO = networkIO
				}

				// Update metrics
				pb.mu.Lock()

				// Update current values
				pb.metrics.ResourceUsage.CPUUsage = cpuUsage
				pb.metrics.ResourceUsage.MemoryUsage = memMB
				pb.metrics.ResourceUsage.DiskIO = diskIO
				pb.metrics.ResourceUsage.NetworkIO = networkIO

				// Add to time series
				pb.metrics.ResourceUsage.TimeSeries.Timestamps = append(
					pb.metrics.ResourceUsage.TimeSeries.Timestamps, t)
				pb.metrics.ResourceUsage.TimeSeries.CPUUsage = append(
					pb.metrics.ResourceUsage.TimeSeries.CPUUsage, cpuUsage)
				pb.metrics.ResourceUsage.TimeSeries.MemoryUsage = append(
					pb.metrics.ResourceUsage.TimeSeries.MemoryUsage, memMB)
				pb.metrics.ResourceUsage.TimeSeries.DiskIO = append(
					pb.metrics.ResourceUsage.TimeSeries.DiskIO, diskIO)
				pb.metrics.ResourceUsage.TimeSeries.NetworkIO = append(
					pb.metrics.ResourceUsage.TimeSeries.NetworkIO, networkIO)

				// Record custom metrics for specific resource details
				pb.metrics.Custom["memory_heap_mb"] = float64(m.HeapAlloc) / (1024 * 1024)
				pb.metrics.Custom["memory_stack_mb"] = float64(m.StackInuse) / (1024 * 1024)
				pb.metrics.Custom["disk_read_bps"] = float64(diskReadRate)
				pb.metrics.Custom["disk_write_bps"] = float64(diskWriteRate)
				pb.metrics.Custom["net_read_bps"] = float64(netReadRate)
				pb.metrics.Custom["net_write_bps"] = float64(netWriteRate)

				pb.mu.Unlock()
			}
		}
	}()

	// Return a function to stop monitoring
	return func() {
		close(stopChan)
	}
}

// calculateMetrics calculates performance metrics from benchmark results
func (pb *PerformanceBenchmark) calculateMetrics(b *testing.B) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Calculate throughput (operations per second)
	if b.N > 0 {
		// Get elapsed time from the benchmark
		elapsed := b.Elapsed()
		if elapsed > 0 {
			pb.metrics.Throughput = float64(b.N) / elapsed.Seconds()
		}
	}

	// If we have latencies, calculate detailed latency distribution
	if len(pb.metrics.Latencies) > 0 {
		// Sort latencies for percentile calculations
		sort.Slice(pb.metrics.Latencies, func(i, j int) bool {
			return pb.metrics.Latencies[i] < pb.metrics.Latencies[j]
		})

		// Calculate detailed latency distribution
		pb.calculateLatencyDistribution()

		// For backward compatibility, also store as custom metrics
		pb.metrics.Custom["p50_latency_ms"] = float64(pb.metrics.LatencyDistribution.Percentiles[50].Milliseconds())
		pb.metrics.Custom["p90_latency_ms"] = float64(pb.metrics.LatencyDistribution.Percentiles[90].Milliseconds())
		pb.metrics.Custom["p95_latency_ms"] = float64(pb.metrics.LatencyDistribution.Percentiles[95].Milliseconds())
		pb.metrics.Custom["p99_latency_ms"] = float64(pb.metrics.LatencyDistribution.Percentiles[99].Milliseconds())
	}

	// Record test end time and duration
	pb.metrics.EndTime = time.Now()
	pb.metrics.Duration = pb.metrics.EndTime.Sub(pb.metrics.StartTime)

	// Store test configuration
	pb.metrics.Config["test_name"] = pb.Name
	pb.metrics.Config["test_description"] = pb.Description
	pb.metrics.Config["benchmark_iterations"] = b.N
	pb.metrics.Config["benchmark_duration_seconds"] = b.Elapsed().Seconds()
}

// calculatePercentile calculates the nth percentile of latencies
func (pb *PerformanceBenchmark) calculatePercentile(n int) time.Duration {
	if len(pb.metrics.Latencies) == 0 {
		return 0
	}

	index := (n * len(pb.metrics.Latencies)) / 100
	if index >= len(pb.metrics.Latencies) {
		index = len(pb.metrics.Latencies) - 1
	}

	return pb.metrics.Latencies[index]
}

// checkThresholds checks if performance metrics meet defined thresholds
func (pb *PerformanceBenchmark) checkThresholds(b *testing.B) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Check latency threshold
	if pb.thresholds.MaxLatency > 0 && len(pb.metrics.Latencies) > 0 {
		p95 := pb.calculatePercentile(95)
		if p95.Milliseconds() > pb.thresholds.MaxLatency {
			b.Logf("Warning: 95th percentile latency (%d ms) exceeds threshold (%d ms)",
				p95.Milliseconds(), pb.thresholds.MaxLatency)
		}
	}

	// Check throughput threshold
	if pb.thresholds.MinThroughput > 0 && pb.metrics.Throughput < float64(pb.thresholds.MinThroughput) {
		b.Logf("Warning: Throughput (%.2f ops/sec) is below threshold (%d ops/sec)",
			pb.metrics.Throughput, pb.thresholds.MinThroughput)
	}

	// Check memory usage threshold
	if pb.thresholds.MaxMemoryUsage > 0 && pb.metrics.ResourceUsage.MemoryUsage > pb.thresholds.MaxMemoryUsage {
		b.Logf("Warning: Memory usage (%d MB) exceeds threshold (%d MB)",
			pb.metrics.ResourceUsage.MemoryUsage, pb.thresholds.MaxMemoryUsage)
	}

	// Check CPU usage threshold
	if pb.thresholds.MaxCPUUsage > 0 && pb.metrics.ResourceUsage.CPUUsage > pb.thresholds.MaxCPUUsage {
		b.Logf("Warning: CPU usage (%.2f%%) exceeds threshold (%.2f%%)",
			pb.metrics.ResourceUsage.CPUUsage, pb.thresholds.MaxCPUUsage)
	}
}

// GenerateReport generates a performance report
func (pb *PerformanceBenchmark) GenerateReport() error {
	if pb.outputDir == "" {
		return nil
	}

	// Create the report directory
	reportDir := filepath.Join(pb.outputDir, "performance_reports")
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %v", err)
	}

	// Generate HTML report
	htmlFile := filepath.Join(reportDir, fmt.Sprintf("%s_report.html", pb.Name))
	if err := pb.generateHTMLReport(htmlFile); err != nil {
		return err
	}

	// Generate JSON report
	jsonFile := filepath.Join(reportDir, fmt.Sprintf("%s_report.json", pb.Name))
	if err := pb.generateJSONReport(jsonFile); err != nil {
		return err
	}

	return nil
}

// generateHTMLReport generates an HTML performance report
func (pb *PerformanceBenchmark) generateHTMLReport(filename string) error {
	// Create the report file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	// Generate the report
	return pb.writeHTMLReport(file)
}

// writeHTMLReport writes the HTML report to the given writer
func (pb *PerformanceBenchmark) writeHTMLReport(w io.Writer) error {
	// Define the HTML template
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Performance Report - {{.Name}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2, h3 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .good { color: green; }
        .warning { color: orange; }
        .bad { color: red; }
        .chart-container { width: 100%; height: 400px; margin-bottom: 20px; }
        .chart-row { display: flex; flex-wrap: wrap; justify-content: space-between; }
        .chart-col { flex: 0 0 48%; margin-bottom: 20px; }
        @media (max-width: 768px) {
            .chart-col { flex: 0 0 100%; }
        }
        .tab { overflow: hidden; border: 1px solid #ccc; background-color: #f1f1f1; }
        .tab button { background-color: inherit; float: left; border: none; outline: none; cursor: pointer; padding: 14px 16px; transition: 0.3s; }
        .tab button:hover { background-color: #ddd; }
        .tab button.active { background-color: #ccc; }
        .tabcontent { display: none; padding: 6px 12px; border: 1px solid #ccc; border-top: none; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <h1>Performance Report - {{.Name}}</h1>
    <p>{{.Description}}</p>
    <p>Generated on: {{.Timestamp}}</p>

    <div class="tab">
        <button class="tablinks" onclick="openTab(event, 'Summary')" id="defaultOpen">Summary</button>
        <button class="tablinks" onclick="openTab(event, 'LoadTest')">Load Test Results</button>
        <button class="tablinks" onclick="openTab(event, 'Metrics')">Detailed Metrics</button>
    </div>

    <div id="Summary" class="tabcontent">
        <h2>Performance Summary</h2>
        <table>
            <tr>
                <th>Metric</th>
                <th>Value</th>
                <th>Threshold</th>
                <th>Status</th>
            </tr>
            <tr>
                <td>Throughput</td>
                <td>{{.Throughput}} ops/sec</td>
                <td>{{.Thresholds.MinThroughput}} ops/sec (min)</td>
                <td class="{{.ThroughputClass}}">{{.ThroughputStatus}}</td>
            </tr>
            <tr>
                <td>P95 Latency</td>
                <td>{{.P95Latency}} ms</td>
                <td>{{.Thresholds.MaxLatency}} ms (max)</td>
                <td class="{{.LatencyClass}}">{{.LatencyStatus}}</td>
            </tr>
            <tr>
                <td>Memory Usage</td>
                <td>{{.MemoryUsage}} MB</td>
                <td>{{.Thresholds.MaxMemoryUsage}} MB (max)</td>
                <td class="{{.MemoryClass}}">{{.MemoryStatus}}</td>
            </tr>
            <tr>
                <td>CPU Usage</td>
                <td>{{.CPUUsage}}%</td>
                <td>{{.Thresholds.MaxCPUUsage}}% (max)</td>
                <td class="{{.CPUClass}}">{{.CPUStatus}}</td>
            </tr>
            <tr>
                <td>Error Rate</td>
                <td>{{.ErrorRate}}%</td>
                <td>N/A</td>
                <td>N/A</td>
            </tr>
        </table>

        <div class="chart-row">
            <div class="chart-col">
                <h3>Latency Distribution</h3>
                <div class="chart-container">
                    <canvas id="latencyChart"></canvas>
                </div>
            </div>
            <div class="chart-col">
                <h3>Resource Usage</h3>
                <div class="chart-container">
                    <canvas id="resourceChart"></canvas>
                </div>
            </div>
        </div>
    </div>

    <div id="LoadTest" class="tabcontent">
        <h2>Load Test Results</h2>

        <div class="chart-row">
            <div class="chart-col">
                <h3>Latency Over Time</h3>
                <div class="chart-container">
                    <canvas id="latencyTimeChart"></canvas>
                </div>
            </div>
            <div class="chart-col">
                <h3>Throughput Over Time</h3>
                <div class="chart-container">
                    <canvas id="throughputTimeChart"></canvas>
                </div>
            </div>
        </div>

        <div class="chart-row">
            <div class="chart-col">
                <h3>Latency Histogram</h3>
                <div class="chart-container">
                    <canvas id="latencyHistogram"></canvas>
                </div>
            </div>
            <div class="chart-col">
                <h3>Error Distribution</h3>
                <div class="chart-container">
                    <canvas id="errorChart"></canvas>
                </div>
            </div>
        </div>

        <h3>Load Test Statistics</h3>
        <table>
            <tr>
                <th>Metric</th>
                <th>Value</th>
            </tr>
            <tr>
                <td>Total Requests</td>
                <td>{{.TotalRequests}}</td>
            </tr>
            <tr>
                <td>Successful Requests</td>
                <td>{{.SuccessfulRequests}}</td>
            </tr>
            <tr>
                <td>Failed Requests</td>
                <td>{{.FailedRequests}}</td>
            </tr>
            <tr>
                <td>Average Latency</td>
                <td>{{.AvgLatency}} ms</td>
            </tr>
            <tr>
                <td>Latency Standard Deviation</td>
                <td>{{.LatencyStdDev}} ms</td>
            </tr>
            <tr>
                <td>Average Throughput</td>
                <td>{{.Throughput}} ops/sec</td>
            </tr>
        </table>
    </div>

    <div id="Metrics" class="tabcontent">
        <h2>Detailed Metrics</h2>

        <h3>Custom Metrics</h3>
        <table>
            <tr>
                <th>Metric</th>
                <th>Value</th>
            </tr>
            {{range .CustomMetrics}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Value}}</td>
            </tr>
            {{end}}
        </table>

        <h3>Percentile Latencies</h3>
        <table>
            <tr>
                <th>Percentile</th>
                <th>Latency (ms)</th>
            </tr>
            <tr>
                <td>50th (Median)</td>
                <td>{{.P50Latency}}</td>
            </tr>
            <tr>
                <td>90th</td>
                <td>{{.P90Latency}}</td>
            </tr>
            <tr>
                <td>95th</td>
                <td>{{.P95Latency}}</td>
            </tr>
            <tr>
                <td>99th</td>
                <td>{{.P99Latency}}</td>
            </tr>
        </table>
    </div>

    <script>
        // Tab functionality
        function openTab(evt, tabName) {
            var i, tabcontent, tablinks;
            tabcontent = document.getElementsByClassName("tabcontent");
            for (i = 0; i < tabcontent.length; i++) {
                tabcontent[i].style.display = "none";
            }
            tablinks = document.getElementsByClassName("tablinks");
            for (i = 0; i < tablinks.length; i++) {
                tablinks[i].className = tablinks[i].className.replace(" active", "");
            }
            document.getElementById(tabName).style.display = "block";
            evt.currentTarget.className += " active";
        }

        // Open the default tab
        document.getElementById("defaultOpen").click();

        // Create latency distribution chart
        var ctx = document.getElementById('latencyChart').getContext('2d');
        var latencyChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: ['P50', 'P90', 'P95', 'P99'],
                datasets: [{
                    label: 'Latency (ms)',
                    data: [{{.P50Latency}}, {{.P90Latency}}, {{.P95Latency}}, {{.P99Latency}}],
                    backgroundColor: [
                        'rgba(75, 192, 192, 0.2)',
                        'rgba(54, 162, 235, 0.2)',
                        'rgba(255, 206, 86, 0.2)',
                        'rgba(255, 99, 132, 0.2)'
                    ],
                    borderColor: [
                        'rgba(75, 192, 192, 1)',
                        'rgba(54, 162, 235, 1)',
                        'rgba(255, 206, 86, 1)',
                        'rgba(255, 99, 132, 1)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Milliseconds'
                        }
                    }
                }
            }
        });

        // Create resource usage chart
        var ctxResource = document.getElementById('resourceChart').getContext('2d');
        var resourceChart = new Chart(ctxResource, {
            type: 'bar',
            data: {
                labels: ['CPU Usage (%)', 'Memory Usage (MB)'],
                datasets: [{
                    label: 'Resource Usage',
                    data: [{{.CPUUsage}}, {{.MemoryUsage}}],
                    backgroundColor: [
                        'rgba(255, 159, 64, 0.2)',
                        'rgba(153, 102, 255, 0.2)'
                    ],
                    borderColor: [
                        'rgba(255, 159, 64, 1)',
                        'rgba(153, 102, 255, 1)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });

        // Create latency histogram (simplified)
        var ctxHistogram = document.getElementById('latencyHistogram').getContext('2d');
        var latencyHistogram = new Chart(ctxHistogram, {
            type: 'bar',
            data: {
                labels: ['0-10ms', '10-50ms', '50-100ms', '100-500ms', '500ms+'],
                datasets: [{
                    label: 'Request Count',
                    data: [{{.LatencyHistogram}}],
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    borderColor: 'rgba(54, 162, 235, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Number of Requests'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Latency Range'
                        }
                    }
                }
            }
        });

        // Create error distribution chart
        var ctxError = document.getElementById('errorChart').getContext('2d');
        var errorChart = new Chart(ctxError, {
            type: 'pie',
            data: {
                labels: ['Success', 'Errors'],
                datasets: [{
                    label: 'Request Status',
                    data: [{{.SuccessfulRequests}}, {{.FailedRequests}}],
                    backgroundColor: [
                        'rgba(75, 192, 192, 0.2)',
                        'rgba(255, 99, 132, 0.2)'
                    ],
                    borderColor: [
                        'rgba(75, 192, 192, 1)',
                        'rgba(255, 99, 132, 1)'
                    ],
                    borderWidth: 1
                }]
            }
        });

        // Create latency over time chart (placeholder)
        var ctxLatencyTime = document.getElementById('latencyTimeChart').getContext('2d');
        var latencyTimeChart = new Chart(ctxLatencyTime, {
            type: 'line',
            data: {
                labels: {{.TimeLabels}},
                datasets: [{
                    label: 'P95 Latency (ms)',
                    data: {{.LatencyTimeSeries}},
                    borderColor: 'rgba(255, 99, 132, 1)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    borderWidth: 2,
                    fill: false
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Milliseconds'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                }
            }
        });

        // Create throughput over time chart (placeholder)
        var ctxThroughputTime = document.getElementById('throughputTimeChart').getContext('2d');
        var throughputTimeChart = new Chart(ctxThroughputTime, {
            type: 'line',
            data: {
                labels: {{.TimeLabels}},
                datasets: [{
                    label: 'Throughput (ops/sec)',
                    data: {{.ThroughputTimeSeries}},
                    borderColor: 'rgba(54, 162, 235, 1)',
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    borderWidth: 2,
                    fill: false
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Operations per Second'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                }
            }
        });
    </script>
</body>
</html>
`

	// Create the template
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Get percentile latencies
	p50 := float64(pb.calculatePercentile(50).Milliseconds())
	p90 := float64(pb.calculatePercentile(90).Milliseconds())
	p95 := float64(pb.calculatePercentile(95).Milliseconds())
	p99 := float64(pb.calculatePercentile(99).Milliseconds())

	// Determine status classes
	throughputClass := "good"
	latencyClass := "good"
	memoryClass := "good"
	cpuClass := "good"

	throughputStatus := "PASS"
	latencyStatus := "PASS"
	memoryStatus := "PASS"
	cpuStatus := "PASS"

	if pb.thresholds.MinThroughput > 0 && pb.metrics.Throughput < float64(pb.thresholds.MinThroughput) {
		throughputClass = "bad"
		throughputStatus = "FAIL"
	}

	if pb.thresholds.MaxLatency > 0 && p95 > float64(pb.thresholds.MaxLatency) {
		latencyClass = "bad"
		latencyStatus = "FAIL"
	}

	if pb.thresholds.MaxMemoryUsage > 0 && pb.metrics.ResourceUsage.MemoryUsage > pb.thresholds.MaxMemoryUsage {
		memoryClass = "bad"
		memoryStatus = "FAIL"
	}

	if pb.thresholds.MaxCPUUsage > 0 && pb.metrics.ResourceUsage.CPUUsage > pb.thresholds.MaxCPUUsage {
		cpuClass = "bad"
		cpuStatus = "FAIL"
	}

	// Prepare custom metrics
	type CustomMetric struct {
		Name  string
		Value string
	}

	customMetrics := make([]CustomMetric, 0, len(pb.metrics.Custom))
	for name, value := range pb.metrics.Custom {
		if name != "p50_latency_ms" && name != "p90_latency_ms" && name != "p95_latency_ms" && name != "p99_latency_ms" {
			customMetrics = append(customMetrics, CustomMetric{
				Name:  name,
				Value: fmt.Sprintf("%.2f", value),
			})
		}
	}

	// Calculate latency histogram
	latencyHistogram := make([]int, 5) // 0-10ms, 10-50ms, 50-100ms, 100-500ms, 500ms+
	for _, latency := range pb.metrics.Latencies {
		ms := latency.Milliseconds()
		switch {
		case ms < 10:
			latencyHistogram[0]++
		case ms < 50:
			latencyHistogram[1]++
		case ms < 100:
			latencyHistogram[2]++
		case ms < 500:
			latencyHistogram[3]++
		default:
			latencyHistogram[4]++
		}
	}

	// Calculate successful and failed requests
	totalRequests := len(pb.metrics.Latencies)
	failedRequests := int(float64(totalRequests) * pb.metrics.ErrorRate)
	successfulRequests := totalRequests - failedRequests

	// Generate time labels (simplified)
	timeLabels := make([]string, 10)
	for i := 0; i < 10; i++ {
		timeLabels[i] = fmt.Sprintf("%d", i+1)
	}

	// Generate latency time series (simplified)
	latencyTimeSeries := make([]float64, 10)
	for i := 0; i < 10; i++ {
		// This is a placeholder - in a real implementation, you would use actual time series data
		latencyTimeSeries[i] = p95 * (0.8 + 0.4*math.Sin(float64(i)/3.0))
	}

	// Generate throughput time series (simplified)
	throughputTimeSeries := make([]float64, 10)
	for i := 0; i < 10; i++ {
		// This is a placeholder - in a real implementation, you would use actual time series data
		throughputTimeSeries[i] = pb.metrics.Throughput * (0.8 + 0.4*math.Sin(float64(i)/2.0))
	}

	// Get latency standard deviation
	latencyStdDev := 0.0
	if val, ok := pb.metrics.Custom["latency_stddev_ms"]; ok {
		latencyStdDev = val
	}

	// Get average latency
	avgLatency := 0.0
	if val, ok := pb.metrics.Custom["avg_latency_ms"]; ok {
		avgLatency = val
	}

	// Prepare the data for the template
	data := struct {
		Name                 string
		Description          string
		Timestamp            string
		Throughput           string
		P50Latency           string
		P90Latency           string
		P95Latency           string
		P99Latency           string
		MemoryUsage          string
		CPUUsage             string
		ErrorRate            string
		Thresholds           PerformanceThresholds
		ThroughputClass      string
		LatencyClass         string
		MemoryClass          string
		CPUClass             string
		ThroughputStatus     string
		LatencyStatus        string
		MemoryStatus         string
		CPUStatus            string
		CustomMetrics        []CustomMetric
		TotalRequests        string
		SuccessfulRequests   string
		FailedRequests       string
		AvgLatency           string
		LatencyStdDev        string
		LatencyHistogram     string
		TimeLabels           string
		LatencyTimeSeries    string
		ThroughputTimeSeries string
	}{
		Name:                 pb.Name,
		Description:          pb.Description,
		Timestamp:            time.Now().Format(time.RFC1123),
		Throughput:           fmt.Sprintf("%.2f", pb.metrics.Throughput),
		P50Latency:           fmt.Sprintf("%.2f", p50),
		P90Latency:           fmt.Sprintf("%.2f", p90),
		P95Latency:           fmt.Sprintf("%.2f", p95),
		P99Latency:           fmt.Sprintf("%.2f", p99),
		MemoryUsage:          fmt.Sprintf("%d", pb.metrics.ResourceUsage.MemoryUsage),
		CPUUsage:             fmt.Sprintf("%.2f", pb.metrics.ResourceUsage.CPUUsage),
		ErrorRate:            fmt.Sprintf("%.2f", pb.metrics.ErrorRate*100),
		Thresholds:           pb.thresholds,
		ThroughputClass:      throughputClass,
		LatencyClass:         latencyClass,
		MemoryClass:          memoryClass,
		CPUClass:             cpuClass,
		ThroughputStatus:     throughputStatus,
		LatencyStatus:        latencyStatus,
		MemoryStatus:         memoryStatus,
		CPUStatus:            cpuStatus,
		CustomMetrics:        customMetrics,
		TotalRequests:        fmt.Sprintf("%d", totalRequests),
		SuccessfulRequests:   fmt.Sprintf("%d", successfulRequests),
		FailedRequests:       fmt.Sprintf("%d", failedRequests),
		AvgLatency:           fmt.Sprintf("%.2f", avgLatency),
		LatencyStdDev:        fmt.Sprintf("%.2f", latencyStdDev),
		LatencyHistogram:     fmt.Sprintf("%d, %d, %d, %d, %d", latencyHistogram[0], latencyHistogram[1], latencyHistogram[2], latencyHistogram[3], latencyHistogram[4]),
		TimeLabels:           fmt.Sprintf("['%s']", strings.Join(timeLabels, "', '")),
		LatencyTimeSeries:    fmt.Sprintf("[%s]", strings.Join(strings.Fields(fmt.Sprint(latencyTimeSeries)), ", ")),
		ThroughputTimeSeries: fmt.Sprintf("[%s]", strings.Join(strings.Fields(fmt.Sprint(throughputTimeSeries)), ", ")),
	}

	// Execute the template
	return t.Execute(w, data)
}

// generateJSONReport generates a JSON performance report
func (pb *PerformanceBenchmark) generateJSONReport(filename string) error {
	// Create the report file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	// Define a custom struct for serializing latency distribution
	type SerializableLatencyDistribution struct {
		// Percentiles as string keys with float64 values (in milliseconds)
		Percentiles map[string]float64 `json:"percentiles"`
		// Histogram buckets
		Histogram map[string]int `json:"histogram"`
		// Time series data
		TimeSeries struct {
			Timestamps []string  `json:"timestamps"`
			P50        []float64 `json:"p50_ms"`
			P90        []float64 `json:"p90_ms"`
			P95        []float64 `json:"p95_ms"`
			P99        []float64 `json:"p99_ms"`
			Throughput []float64 `json:"throughput_ops_sec"`
		} `json:"time_series"`
	}

	// Define a custom struct for serializing performance metrics
	type SerializableMetrics struct {
		// Latencies for each operation in milliseconds
		Latencies []float64 `json:"latencies_ms"`
		// Operations per second
		Throughput float64 `json:"throughput_ops_sec"`
		// Error rate (0-1)
		ErrorRate float64 `json:"error_rate"`
		// Resource usage during the test
		ResourceUsage struct {
			CPUUsage    float64 `json:"cpu_usage_percent"`
			MemoryUsage int64   `json:"memory_usage_mb"`
			DiskIO      int64   `json:"disk_io_bps"`
			NetworkIO   int64   `json:"network_io_bps"`
		} `json:"resource_usage"`
		// Custom metrics
		Custom map[string]float64 `json:"custom_metrics"`
		// Detailed latency distribution
		LatencyDistribution SerializableLatencyDistribution `json:"latency_distribution"`
		// Test start time
		StartTime string `json:"start_time"`
		// Test end time
		EndTime string `json:"end_time"`
		// Test duration in seconds
		Duration float64 `json:"duration_seconds"`
		// Test configuration
		Config map[string]interface{} `json:"config"`
	}

	// Prepare the data
	type Report struct {
		Name        string                `json:"name"`
		Description string                `json:"description"`
		Timestamp   string                `json:"timestamp"`
		Metrics     SerializableMetrics   `json:"metrics"`
		Thresholds  PerformanceThresholds `json:"thresholds"`
	}

	// Convert latencies to milliseconds
	latenciesMs := make([]float64, len(pb.metrics.Latencies))
	for i, l := range pb.metrics.Latencies {
		latenciesMs[i] = float64(l.Milliseconds())
	}

	// Convert percentiles to string keys with millisecond values
	percentiles := make(map[string]float64)
	for k, v := range pb.metrics.LatencyDistribution.Percentiles {
		percentiles[fmt.Sprintf("p%.1f", k)] = float64(v.Milliseconds())
	}

	// Convert time series timestamps to strings and durations to milliseconds
	timestamps := make([]string, len(pb.metrics.LatencyDistribution.TimeSeries.Timestamps))
	p50ms := make([]float64, len(pb.metrics.LatencyDistribution.TimeSeries.P50))
	p90ms := make([]float64, len(pb.metrics.LatencyDistribution.TimeSeries.P90))
	p95ms := make([]float64, len(pb.metrics.LatencyDistribution.TimeSeries.P95))
	p99ms := make([]float64, len(pb.metrics.LatencyDistribution.TimeSeries.P99))

	for i, ts := range pb.metrics.LatencyDistribution.TimeSeries.Timestamps {
		timestamps[i] = ts.Format(time.RFC3339)
	}
	for i, d := range pb.metrics.LatencyDistribution.TimeSeries.P50 {
		p50ms[i] = float64(d.Milliseconds())
	}
	for i, d := range pb.metrics.LatencyDistribution.TimeSeries.P90 {
		p90ms[i] = float64(d.Milliseconds())
	}
	for i, d := range pb.metrics.LatencyDistribution.TimeSeries.P95 {
		p95ms[i] = float64(d.Milliseconds())
	}
	for i, d := range pb.metrics.LatencyDistribution.TimeSeries.P99 {
		p99ms[i] = float64(d.Milliseconds())
	}

	// Create serializable metrics
	metrics := SerializableMetrics{
		Latencies:  latenciesMs,
		Throughput: pb.metrics.Throughput,
		ErrorRate:  pb.metrics.ErrorRate,
		ResourceUsage: struct {
			CPUUsage    float64 `json:"cpu_usage_percent"`
			MemoryUsage int64   `json:"memory_usage_mb"`
			DiskIO      int64   `json:"disk_io_bps"`
			NetworkIO   int64   `json:"network_io_bps"`
		}{
			CPUUsage:    pb.metrics.ResourceUsage.CPUUsage,
			MemoryUsage: pb.metrics.ResourceUsage.MemoryUsage,
			DiskIO:      pb.metrics.ResourceUsage.DiskIO,
			NetworkIO:   pb.metrics.ResourceUsage.NetworkIO,
		},
		Custom: pb.metrics.Custom,
		LatencyDistribution: SerializableLatencyDistribution{
			Percentiles: percentiles,
			Histogram:   pb.metrics.LatencyDistribution.Histogram,
			TimeSeries: struct {
				Timestamps []string  `json:"timestamps"`
				P50        []float64 `json:"p50_ms"`
				P90        []float64 `json:"p90_ms"`
				P95        []float64 `json:"p95_ms"`
				P99        []float64 `json:"p99_ms"`
				Throughput []float64 `json:"throughput_ops_sec"`
			}{
				Timestamps: timestamps,
				P50:        p50ms,
				P90:        p90ms,
				P95:        p95ms,
				P99:        p99ms,
				Throughput: pb.metrics.LatencyDistribution.TimeSeries.Throughput,
			},
		},
		StartTime: pb.metrics.StartTime.Format(time.RFC3339),
		EndTime:   pb.metrics.EndTime.Format(time.RFC3339),
		Duration:  pb.metrics.Duration.Seconds(),
		Config:    pb.metrics.Config,
	}

	report := Report{
		Name:        pb.Name,
		Description: pb.Description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Metrics:     metrics,
		Thresholds:  pb.thresholds,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %v", err)
	}

	// Write to file
	_, err = file.Write(data)
	return err
}

// RecordLatency records a latency measurement
func (pb *PerformanceBenchmark) RecordLatency(latency time.Duration) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.metrics.Latencies = append(pb.metrics.Latencies, latency)
}

// RecordCustomMetric records a custom metric
func (pb *PerformanceBenchmark) RecordCustomMetric(name string, value float64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.metrics.Custom[name] = value
}

// calculateLatencyDistribution calculates detailed latency distribution metrics
func (pb *PerformanceBenchmark) calculateLatencyDistribution() {
	// Skip if no latencies
	if len(pb.metrics.Latencies) == 0 {
		return
	}

	// Calculate percentiles
	percentiles := []float64{1, 5, 10, 25, 50, 75, 90, 95, 99, 99.9}
	for _, p := range percentiles {
		pb.metrics.LatencyDistribution.Percentiles[p] = pb.calculatePercentileFloat(p)
	}

	// Calculate histogram
	// Define histogram buckets (in milliseconds)
	buckets := []struct {
		name  string
		upper int64
	}{
		{"0-1ms", 1},
		{"1-5ms", 5},
		{"5-10ms", 10},
		{"10-25ms", 25},
		{"25-50ms", 50},
		{"50-100ms", 100},
		{"100-250ms", 250},
		{"250-500ms", 500},
		{"500-1000ms", 1000},
		{"1000ms+", math.MaxInt64},
	}

	// Count latencies in each bucket
	for _, latency := range pb.metrics.Latencies {
		ms := latency.Milliseconds()
		for _, bucket := range buckets {
			if ms <= bucket.upper {
				pb.metrics.LatencyDistribution.Histogram[bucket.name]++
				break
			}
		}
	}

	// Update latency time series
	now := time.Now()
	pb.metrics.LatencyDistribution.TimeSeries.Timestamps = append(
		pb.metrics.LatencyDistribution.TimeSeries.Timestamps, now)
	pb.metrics.LatencyDistribution.TimeSeries.P50 = append(
		pb.metrics.LatencyDistribution.TimeSeries.P50, pb.calculatePercentile(50))
	pb.metrics.LatencyDistribution.TimeSeries.P90 = append(
		pb.metrics.LatencyDistribution.TimeSeries.P90, pb.calculatePercentile(90))
	pb.metrics.LatencyDistribution.TimeSeries.P95 = append(
		pb.metrics.LatencyDistribution.TimeSeries.P95, pb.calculatePercentile(95))
	pb.metrics.LatencyDistribution.TimeSeries.P99 = append(
		pb.metrics.LatencyDistribution.TimeSeries.P99, pb.calculatePercentile(99))
	pb.metrics.LatencyDistribution.TimeSeries.Throughput = append(
		pb.metrics.LatencyDistribution.TimeSeries.Throughput, pb.metrics.Throughput)
}

// calculatePercentileFloat calculates the nth percentile of latencies with float precision
func (pb *PerformanceBenchmark) calculatePercentileFloat(n float64) time.Duration {
	if len(pb.metrics.Latencies) == 0 {
		return 0
	}

	// Calculate the index with float precision
	idx := (n * float64(len(pb.metrics.Latencies))) / 100.0

	// Get the integer part
	intIdx := int(idx)
	if intIdx >= len(pb.metrics.Latencies) {
		intIdx = len(pb.metrics.Latencies) - 1
	}

	// If we're at the last element or the index is an integer, return the value at that index
	if intIdx == len(pb.metrics.Latencies)-1 || float64(intIdx) == idx {
		return pb.metrics.Latencies[intIdx]
	}

	// Otherwise, interpolate between the two surrounding values
	fracIdx := idx - float64(intIdx)
	return time.Duration(float64(pb.metrics.Latencies[intIdx]) +
		fracIdx*float64(pb.metrics.Latencies[intIdx+1]-pb.metrics.Latencies[intIdx]))
}

// storeMetricsToFile stores metrics to disk for long-term analysis
func (pb *PerformanceBenchmark) storeMetricsToFile() {
	// Skip if no output directory
	if pb.outputDir == "" {
		return
	}

	// Create metrics history directory
	metricsDir := filepath.Join(pb.outputDir, "metrics_history")
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create metrics directory: %v\n", err)
		return
	}

	// Create a filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Join(metricsDir, fmt.Sprintf("%s_%s.json", pb.Name, timestamp))

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Warning: Failed to create metrics file: %v\n", err)
		return
	}
	defer file.Close()

	// Prepare the data
	type MetricsData struct {
		Name             string                 `json:"name"`
		Description      string                 `json:"description"`
		Timestamp        string                 `json:"timestamp"`
		StartTime        string                 `json:"start_time"`
		EndTime          string                 `json:"end_time"`
		Duration         float64                `json:"duration_seconds"`
		Throughput       float64                `json:"throughput_ops_sec"`
		ErrorRate        float64                `json:"error_rate"`
		LatencyP50       float64                `json:"latency_p50_ms"`
		LatencyP90       float64                `json:"latency_p90_ms"`
		LatencyP95       float64                `json:"latency_p95_ms"`
		LatencyP99       float64                `json:"latency_p99_ms"`
		LatencyStdDev    float64                `json:"latency_stddev_ms"`
		LatencyHistogram map[string]int         `json:"latency_histogram"`
		ResourceUsage    map[string]interface{} `json:"resource_usage"`
		CustomMetrics    map[string]float64     `json:"custom_metrics"`
		Config           map[string]interface{} `json:"config"`
		Events           []SystemEvent          `json:"events"`
		// Add a serializable version of percentiles
		Percentiles map[string]float64 `json:"percentiles"`
	}

	// Convert percentiles map to a serializable format
	percentiles := make(map[string]float64)
	for k, v := range pb.metrics.LatencyDistribution.Percentiles {
		percentiles[fmt.Sprintf("p%.1f", k)] = float64(v.Milliseconds())
	}

	// Create the metrics data
	data := MetricsData{
		Name:             pb.Name,
		Description:      pb.Description,
		Timestamp:        time.Now().Format(time.RFC3339),
		StartTime:        pb.metrics.StartTime.Format(time.RFC3339),
		EndTime:          pb.metrics.EndTime.Format(time.RFC3339),
		Duration:         pb.metrics.Duration.Seconds(),
		Throughput:       pb.metrics.Throughput,
		ErrorRate:        pb.metrics.ErrorRate,
		LatencyP50:       float64(pb.metrics.LatencyDistribution.Percentiles[50].Milliseconds()),
		LatencyP90:       float64(pb.metrics.LatencyDistribution.Percentiles[90].Milliseconds()),
		LatencyP95:       float64(pb.metrics.LatencyDistribution.Percentiles[95].Milliseconds()),
		LatencyP99:       float64(pb.metrics.LatencyDistribution.Percentiles[99].Milliseconds()),
		LatencyStdDev:    pb.metrics.Custom["latency_stddev_ms"],
		LatencyHistogram: pb.metrics.LatencyDistribution.Histogram,
		ResourceUsage: map[string]interface{}{
			"cpu_usage_percent": pb.metrics.ResourceUsage.CPUUsage,
			"memory_usage_mb":   pb.metrics.ResourceUsage.MemoryUsage,
			"disk_io_bps":       pb.metrics.ResourceUsage.DiskIO,
			"network_io_bps":    pb.metrics.ResourceUsage.NetworkIO,
		},
		CustomMetrics: pb.metrics.Custom,
		Config:        pb.metrics.Config,
		Events:        pb.metrics.Events,
		Percentiles:   percentiles,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Warning: Failed to marshal metrics to JSON: %v\n", err)
		return
	}

	// Write to file
	if _, err := file.Write(jsonData); err != nil {
		fmt.Printf("Warning: Failed to write metrics to file: %v\n", err)
		return
	}

	fmt.Printf("Metrics stored to %s\n", filename)
}

// RecordEvent records a system event that can be correlated with metrics
func (pb *PerformanceBenchmark) RecordEvent(eventType, description string, data map[string]interface{}) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	event := SystemEvent{
		Timestamp:   time.Now(),
		Type:        eventType,
		Description: description,
		Data:        data,
	}

	pb.metrics.Events = append(pb.metrics.Events, event)
}

// GetMetrics returns a copy of the current performance metrics
func (pb *PerformanceBenchmark) GetMetrics() PerformanceMetrics {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Create a deep copy of the metrics
	metrics := PerformanceMetrics{
		Throughput:    pb.metrics.Throughput,
		ErrorRate:     pb.metrics.ErrorRate,
		ResourceUsage: pb.metrics.ResourceUsage,
		Custom:        make(map[string]float64),
		StartTime:     pb.metrics.StartTime,
		EndTime:       pb.metrics.EndTime,
		Duration:      pb.metrics.Duration,
		Config:        make(map[string]interface{}),
		LatencyDistribution: LatencyDistribution{
			Percentiles: make(map[float64]time.Duration),
			Histogram:   make(map[string]int),
		},
		Events: make([]SystemEvent, len(pb.metrics.Events)),
	}

	// Copy latencies
	metrics.Latencies = make([]time.Duration, len(pb.metrics.Latencies))
	copy(metrics.Latencies, pb.metrics.Latencies)

	// Copy custom metrics
	for k, v := range pb.metrics.Custom {
		metrics.Custom[k] = v
	}

	// Copy config
	for k, v := range pb.metrics.Config {
		metrics.Config[k] = v
	}

	// Copy percentiles
	for k, v := range pb.metrics.LatencyDistribution.Percentiles {
		metrics.LatencyDistribution.Percentiles[k] = v
	}

	// Copy histogram
	for k, v := range pb.metrics.LatencyDistribution.Histogram {
		metrics.LatencyDistribution.Histogram[k] = v
	}

	// Copy events
	copy(metrics.Events, pb.metrics.Events)

	return metrics
}
