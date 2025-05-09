// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"context"
	"github.com/auriora/onemount/internal/testutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// TestConfig defines configuration options for the test environment.
type TestConfig struct {
	// Test environment configuration
	Environment string
	// Timeout for tests in seconds
	Timeout int
	// Whether to enable verbose logging
	VerboseLogging bool
	// Directory for test artifacts
	ArtifactsDir string
	// Custom configuration options
	CustomOptions map[string]interface{}
}

// TestResource represents a resource that needs cleanup after tests.
type TestResource interface {
	// Cleanup performs necessary cleanup operations for the resource.
	Cleanup() error
}

// MockProvider is an interface for mock components.
type MockProvider interface {
	// Setup initializes the mock provider.
	Setup() error
	// Teardown cleans up the mock provider.
	Teardown() error
	// Reset resets the mock provider to its initial state.
	Reset() error
}

// CoverageReporter collects and reports test coverage.
type CoverageReporter interface {
	// CollectCoverage collects coverage data.
	CollectCoverage() error
	// ReportCoverage generates a coverage report.
	ReportCoverage() error
	// CheckThresholds checks if coverage meets defined thresholds.
	CheckThresholds() (bool, error)
}

// Logger provides structured logging for tests.
type Logger interface {
	// Debug logs a debug message.
	Debug(msg string, args ...interface{})
	// Info logs an informational message.
	Info(msg string, args ...interface{})
	// Warn logs a warning message.
	Warn(msg string, args ...interface{})
	// Error logs an error message.
	Error(msg string, args ...interface{})
}

// TestStatus represents the status of a test.
type TestStatus string

const (
	// TestStatusPassed indicates the test passed.
	TestStatusPassed TestStatus = "PASSED"
	// TestStatusFailed indicates the test failed.
	TestStatusFailed TestStatus = "FAILED"
	// TestStatusSkipped indicates the test was skipped.
	TestStatusSkipped TestStatus = "SKIPPED"
)

// TestFailure represents a test failure.
type TestFailure struct {
	// Message describes the failure.
	Message string
	// Location is where the failure occurred.
	Location string
	// Expected is what was expected.
	Expected interface{}
	// Actual is what was actually received.
	Actual interface{}
}

// TestArtifact represents a test artifact.
type TestArtifact struct {
	// Name of the artifact.
	Name string
	// Type of the artifact.
	Type string
	// Location where the artifact is stored.
	Location string
}

// TestResult represents the result of a test.
type TestResult struct {
	// Name of the test.
	Name string
	// Duration of the test.
	Duration time.Duration
	// Status of the test.
	Status TestStatus
	// Failures that occurred during the test.
	Failures []TestFailure
	// Artifacts generated during the test.
	Artifacts []TestArtifact
}

// TestLifecycle defines hooks for test lifecycle events.
type TestLifecycle interface {
	// BeforeTest is called before a test is executed.
	BeforeTest(ctx context.Context) error
	// AfterTest is called after a test is executed.
	AfterTest(ctx context.Context) error
	// OnFailure is called when a test fails.
	OnFailure(ctx context.Context, failure TestFailure) error
}

// TestFramework provides centralized test configuration and execution.
type TestFramework struct {
	// Configuration for the test environment.
	Config TestConfig

	// Test resources that need cleanup.
	resources []TestResource

	// Mock providers.
	mockProviders map[string]MockProvider

	// Coverage reporting.
	coverageReporter CoverageReporter

	// Network simulation.
	networkSimulator NetworkSimulator

	// Context for timeout/cancellation.
	ctx context.Context

	// Structured logging.
	logger Logger

	// Signal handling
	signalChan   chan os.Signal
	signalMu     sync.Mutex
	isHandling   bool
	signalCtx    context.Context
	signalCancel context.CancelFunc
}

// NewTestFramework creates a new TestFramework with the given configuration.
func NewTestFramework(config TestConfig, logger Logger) *TestFramework {
	// Set default ArtifactsDir if not provided
	if config.ArtifactsDir == "" {
		config.ArtifactsDir = testutil.GetDefaultArtifactsDir()
	}

	// Create a context with cancellation for signal handling
	signalCtx, signalCancel := context.WithCancel(context.Background())

	return &TestFramework{
		Config:           config,
		resources:        make([]TestResource, 0),
		mockProviders:    make(map[string]MockProvider),
		networkSimulator: NewNetworkSimulator(),
		ctx:              context.Background(),
		logger:           logger,
		signalChan:       nil,
		isHandling:       false,
		signalCtx:        signalCtx,
		signalCancel:     signalCancel,
	}
}

// AddResource adds a resource to be cleaned up after tests.
func (tf *TestFramework) AddResource(resource TestResource) {
	tf.resources = append(tf.resources, resource)
}

// CleanupResources cleans up all registered resources.
func (tf *TestFramework) CleanupResources() error {
	var lastErr error
	// Clean up resources in reverse order (LIFO)
	for i := len(tf.resources) - 1; i >= 0; i-- {
		resource := tf.resources[i]
		if err := resource.Cleanup(); err != nil {
			tf.logger.Error("Failed to clean up resource", "error", err)
			lastErr = err
		}
	}
	// Clear the resources slice
	tf.resources = make([]TestResource, 0)
	return lastErr
}

// RegisterMockProvider registers a mock provider with the given name.
func (tf *TestFramework) RegisterMockProvider(name string, provider MockProvider) {
	tf.mockProviders[name] = provider

	// Register the provider with the network simulator
	if tf.networkSimulator != nil {
		tf.networkSimulator.RegisterProvider(provider)
	}
}

// GetMockProvider returns the mock provider with the given name.
func (tf *TestFramework) GetMockProvider(name string) (MockProvider, bool) {
	provider, exists := tf.mockProviders[name]
	return provider, exists
}

// SetCoverageReporter sets the coverage reporter for the test
func (tf *TestFramework) SetCoverageReporter(reporter CoverageReporter) {
	tf.coverageReporter = reporter
}

// RunTest executes a single test function with the given name.
func (tf *TestFramework) RunTest(name string, testFunc func(ctx context.Context) error) TestResult {
	startTime := time.Now()
	result := TestResult{
		Name:      name,
		Status:    TestStatusPassed,
		Failures:  make([]TestFailure, 0),
		Artifacts: make([]TestArtifact, 0),
	}

	// Create a context with timeout if configured
	ctx := tf.ctx
	if tf.Config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(tf.Config.Timeout)*time.Second)
		defer cancel()
	}

	// Log test start
	tf.logger.Info("Starting test", "name", name)

	// Execute the test
	err := testFunc(ctx)

	// Check for errors
	if err != nil {
		result.Status = TestStatusFailed
		result.Failures = append(result.Failures, TestFailure{
			Message:  err.Error(),
			Location: "test execution",
		})
		tf.logger.Error("Test failed", "name", name, "error", err)
	} else {
		tf.logger.Info("Test passed", "name", name)
	}

	// Calculate duration
	result.Duration = time.Since(startTime)

	return result
}

// RunTestSuite executes a suite of tests and returns the results.
func (tf *TestFramework) RunTestSuite(name string, tests map[string]func(ctx context.Context) error) []TestResult {
	results := make([]TestResult, 0, len(tests))

	tf.logger.Info("Starting test suite", "name", name, "tests", len(tests))
	startTime := time.Now()

	// Run each test in the suite
	for testName, testFunc := range tests {
		result := tf.RunTest(testName, testFunc)
		results = append(results, result)
	}

	// Log suite completion
	duration := time.Since(startTime)
	tf.logger.Info("Test suite completed", "name", name, "duration", duration)

	return results
}

// WithTimeout returns a new context with the specified timeout.
func (tf *TestFramework) WithTimeout(timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(tf.ctx, timeout)
	return ctx
}

// WithCancel returns a new context with a cancel function.
func (tf *TestFramework) WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(tf.ctx)
}

// SetContext sets the base context for the test
func (tf *TestFramework) SetContext(ctx context.Context) {
	tf.ctx = ctx
}

// GetNetworkSimulator returns the network simulator.
func (tf *TestFramework) GetNetworkSimulator() NetworkSimulator {
	return tf.networkSimulator
}

// SetNetworkSimulator sets the network simulator.
func (tf *TestFramework) SetNetworkSimulator(simulator NetworkSimulator) {
	tf.networkSimulator = simulator
}

// SetNetworkConditions sets the network conditions.
func (tf *TestFramework) SetNetworkConditions(latency time.Duration, packetLoss float64, bandwidth int) error {
	return tf.networkSimulator.SetConditions(latency, packetLoss, bandwidth)
}

// ApplyNetworkPreset applies a predefined network condition preset.
func (tf *TestFramework) ApplyNetworkPreset(preset NetworkCondition) error {
	return tf.networkSimulator.ApplyPreset(preset)
}

// DisconnectNetwork simulates a network disconnection.
func (tf *TestFramework) DisconnectNetwork() error {
	return tf.networkSimulator.Disconnect()
}

// ReconnectNetwork restores the network connection.
func (tf *TestFramework) ReconnectNetwork() error {
	return tf.networkSimulator.Reconnect()
}

// IsNetworkConnected returns whether the network is currently connected.
func (tf *TestFramework) IsNetworkConnected() bool {
	return tf.networkSimulator.IsConnected()
}

// SetupSignalHandling registers signal handlers for SIGINT and SIGTERM to ensure
// proper cleanup when tests are interrupted. It returns a function that can be
// called to stop signal handling.
func (tf *TestFramework) SetupSignalHandling() func() {
	tf.signalMu.Lock()
	defer tf.signalMu.Unlock()

	// If signal handling is already set up, return a no-op cleanup function
	if tf.isHandling {
		tf.logger.Info("Signal handling already set up")
		return func() {
			tf.logger.Info("Signal handling already stopped by another call")
		}
	}

	// If there's an existing signal channel, stop and close it first
	if tf.signalChan != nil {
		signal.Stop(tf.signalChan)
		close(tf.signalChan)
		tf.signalChan = nil
	}

	// Cancel any existing signal handling goroutine
	if tf.signalCancel != nil {
		tf.signalCancel()
	}

	// Create a new context with cancellation for this signal handling session
	tf.signalCtx, tf.signalCancel = context.WithCancel(context.Background())

	// Create a channel to receive OS signals
	tf.signalChan = make(chan os.Signal, 1)

	// Register signal handlers for SIGINT and SIGTERM
	signal.Notify(tf.signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Set the isHandling flag to true
	tf.isHandling = true

	tf.logger.Info("Signal handling set up for SIGINT and SIGTERM")

	// Start a goroutine to handle signals
	go func(ctx context.Context) {
		select {
		case sig := <-tf.signalChan:
			tf.logger.Info("Received signal", "signal", sig)

			// Clean up resources
			tf.logger.Info("Cleaning up resources due to signal")
			if err := tf.CleanupResources(); err != nil {
				tf.logger.Error("Error cleaning up resources", "error", err)
			}

			// Exit with a non-zero status code
			tf.logger.Info("Exiting due to signal")
			os.Exit(1)
		case <-ctx.Done():
			// Context was cancelled, just exit the goroutine
			tf.logger.Debug("Signal handling goroutine exiting due to context cancellation")
			return
		}
	}(tf.signalCtx)

	// Return a function that can be called to stop signal handling
	return func() {
		tf.signalMu.Lock()
		defer tf.signalMu.Unlock()

		if !tf.isHandling {
			tf.logger.Info("Signal handling already stopped")
			return
		}

		// Stop receiving signals
		signal.Stop(tf.signalChan)

		// Only close the channel if it's not nil
		if tf.signalChan != nil {
			close(tf.signalChan)
			tf.signalChan = nil
		}

		// Cancel the context to stop the goroutine
		if tf.signalCancel != nil {
			tf.signalCancel()
		}

		// Set the isHandling flag to false
		tf.isHandling = false

		tf.logger.Info("Signal handling stopped")
	}
}
