// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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
}

// LoadTest defines load testing parameters
type LoadTest struct {
	// Number of concurrent operations
	Concurrency int
	// Duration of the test
	Duration time.Duration
	// Ramp-up time before measurements
	RampUp time.Duration
	// Test scenario to run
	Scenario func(ctx context.Context) error
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

	return &PerformanceBenchmark{
		Name:        name,
		Description: description,
		thresholds:  thresholds,
		metrics: PerformanceMetrics{
			Latencies: make([]time.Duration, 0),
			Custom:    make(map[string]float64),
		},
		outputDir: outputDir,
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

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, pb.loadTest.Duration+pb.loadTest.RampUp)
	defer cancel()

	// Start resource monitoring
	stopMonitoring := pb.startResourceMonitoring()
	defer stopMonitoring()

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

	// Record the start time
	startTime := time.Now()

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

	// Calculate metrics
	var latencies []time.Duration
	var errors []error

	for latency := range latencyChan {
		latencies = append(latencies, latency)
	}

	for err := range errorChan {
		errors = append(errors, err)
	}

	// Record metrics
	pb.mu.Lock()
	pb.metrics.Latencies = latencies
	pb.metrics.ErrorRate = float64(len(errors)) / float64(len(latencies))
	pb.metrics.Throughput = float64(len(latencies)) / time.Since(startTime).Seconds()
	pb.mu.Unlock()

	// Generate report
	return pb.GenerateReport()
}

// startResourceMonitoring starts monitoring resource usage and returns a function to stop monitoring
func (pb *PerformanceBenchmark) startResourceMonitoring() func() {
	// Create a channel to signal the monitoring goroutine to stop
	stopChan := make(chan struct{})

	// Start a goroutine to monitor resource usage
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var maxCPU float64
		var maxMem int64

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				// Get current resource usage
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				// Convert to MB
				memMB := int64(m.Alloc) / (1024 * 1024)

				// Get CPU usage (this is a simplification)
				// In a real implementation, you would use something like github.com/shirou/gopsutil
				// to get accurate CPU usage
				cpuUsage := 0.0 // Placeholder

				// Update max values
				if cpuUsage > maxCPU {
					maxCPU = cpuUsage
				}
				if memMB > maxMem {
					maxMem = memMB
				}

				// Update metrics
				pb.mu.Lock()
				pb.metrics.ResourceUsage.CPUUsage = maxCPU
				pb.metrics.ResourceUsage.MemoryUsage = maxMem
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

	// If we have latencies, calculate percentiles
	if len(pb.metrics.Latencies) > 0 {
		// Sort latencies for percentile calculations
		sort.Slice(pb.metrics.Latencies, func(i, j int) bool {
			return pb.metrics.Latencies[i] < pb.metrics.Latencies[j]
		})

		// Calculate percentiles
		p50 := pb.calculatePercentile(50)
		p90 := pb.calculatePercentile(90)
		p95 := pb.calculatePercentile(95)
		p99 := pb.calculatePercentile(99)

		// Store as custom metrics
		pb.metrics.Custom["p50_latency_ms"] = float64(p50.Milliseconds())
		pb.metrics.Custom["p90_latency_ms"] = float64(p90.Milliseconds())
		pb.metrics.Custom["p95_latency_ms"] = float64(p95.Milliseconds())
		pb.metrics.Custom["p99_latency_ms"] = float64(p99.Milliseconds())
	}
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
        h1, h2 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .good { color: green; }
        .warning { color: orange; }
        .bad { color: red; }
        .chart-container { width: 100%; height: 400px; margin-bottom: 20px; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <h1>Performance Report - {{.Name}}</h1>
    <p>{{.Description}}</p>
    <p>Generated on: {{.Timestamp}}</p>

    <h2>Summary</h2>
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

    <h2>Latency Distribution</h2>
    <div class="chart-container">
        <canvas id="latencyChart"></canvas>
    </div>

    <h2>Custom Metrics</h2>
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

    <script>
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

	// Prepare the data for the template
	data := struct {
		Name             string
		Description      string
		Timestamp        string
		Throughput       string
		P50Latency       string
		P90Latency       string
		P95Latency       string
		P99Latency       string
		MemoryUsage      string
		CPUUsage         string
		ErrorRate        string
		Thresholds       PerformanceThresholds
		ThroughputClass  string
		LatencyClass     string
		MemoryClass      string
		CPUClass         string
		ThroughputStatus string
		LatencyStatus    string
		MemoryStatus     string
		CPUStatus        string
		CustomMetrics    []CustomMetric
	}{
		Name:             pb.Name,
		Description:      pb.Description,
		Timestamp:        time.Now().Format(time.RFC1123),
		Throughput:       fmt.Sprintf("%.2f", pb.metrics.Throughput),
		P50Latency:       fmt.Sprintf("%.2f", p50),
		P90Latency:       fmt.Sprintf("%.2f", p90),
		P95Latency:       fmt.Sprintf("%.2f", p95),
		P99Latency:       fmt.Sprintf("%.2f", p99),
		MemoryUsage:      fmt.Sprintf("%d", pb.metrics.ResourceUsage.MemoryUsage),
		CPUUsage:         fmt.Sprintf("%.2f", pb.metrics.ResourceUsage.CPUUsage),
		ErrorRate:        fmt.Sprintf("%.2f", pb.metrics.ErrorRate*100),
		Thresholds:       pb.thresholds,
		ThroughputClass:  throughputClass,
		LatencyClass:     latencyClass,
		MemoryClass:      memoryClass,
		CPUClass:         cpuClass,
		ThroughputStatus: throughputStatus,
		LatencyStatus:    latencyStatus,
		MemoryStatus:     memoryStatus,
		CPUStatus:        cpuStatus,
		CustomMetrics:    customMetrics,
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

	// Prepare the data
	type Report struct {
		Name        string                `json:"name"`
		Description string                `json:"description"`
		Timestamp   string                `json:"timestamp"`
		Metrics     PerformanceMetrics    `json:"metrics"`
		Thresholds  PerformanceThresholds `json:"thresholds"`
	}

	report := Report{
		Name:        pb.Name,
		Description: pb.Description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Metrics:     pb.metrics,
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
	}

	// Copy latencies
	metrics.Latencies = make([]time.Duration, len(pb.metrics.Latencies))
	copy(metrics.Latencies, pb.metrics.Latencies)

	// Copy custom metrics
	for k, v := range pb.metrics.Custom {
		metrics.Custom[k] = v
	}

	return metrics
}
