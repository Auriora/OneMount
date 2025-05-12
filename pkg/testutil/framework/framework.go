// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"context"
	"github.com/auriora/onemount/pkg/testutil"
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

// MockProvider is an interface for mock components used in testing.
// Mock providers are controlled implementations of system components that allow tests
// to run without relying on actual external services or components. They are essential for:
//
// 1. Isolation: Testing components in isolation from external dependencies
// 2. Controlled Testing: Creating predictable test environments with predefined responses
// 3. Simulation: Simulating various scenarios, including error conditions
// 4. Verification: Verifying that components interact correctly with external services
//
// All mock providers in the OneMount test framework implement this common interface,
// which ensures they can be managed consistently by the test framework.
type MockProvider interface {
	// Setup initializes the mock provider.
	// This method should be called before using the mock provider in tests.
	// It performs any necessary initialization, such as setting up internal state,
	// creating resources, or establishing connections.
	//
	// Returns an error if initialization fails.
	Setup() error

	// Teardown cleans up the mock provider.
	// This method should be called after tests are complete to clean up any
	// resources created by the mock provider. It ensures that tests leave
	// no residual state that could affect subsequent tests.
	//
	// Returns an error if cleanup fails.
	Teardown() error

	// Reset resets the mock provider to its initial state.
	// This method can be called between tests to restore the mock provider
	// to a clean state without having to tear it down and set it up again.
	// It clears any recorded calls, resets any configured responses, and
	// restores default settings.
	//
	// Returns an error if the reset operation fails.
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
// It is the central component of the OneMount test infrastructure, providing features for:
//
// - Test environment configuration
// - Resource management with automatic cleanup
// - Mock provider registration and retrieval
// - Network condition simulation
// - Test execution with timeout support
// - Context management for cancellation and timeouts
// - Structured logging
// - Signal handling for graceful test interruption
//
// The TestFramework is designed to be used in all types of tests, from unit tests
// to integration tests, system tests, performance tests, and more. It provides a
// consistent interface for test configuration and execution across all test types.
//
// Example usage:
//
//	// Create a logger
//	logger := log.With().Str("component", "test").Logger()
//
//	// Create a test configuration
//	config := framework.TestConfig{
//	    Environment:    "test",
//	    Timeout:        30,  // 30 seconds
//	    VerboseLogging: true,
//	    ArtifactsDir:   "/tmp/test-artifacts",
//	}
//
//	// Create a new TestFramework
//	tf := framework.NewTestFramework(config, &logger)
//
//	// Run a test
//	result := tf.RunTest("my-test", func(ctx context.Context) error {
//	    // Test logic here
//	    return nil
//	})
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
// This function initializes a new TestFramework instance with the specified configuration
// and logger. It sets up the basic infrastructure for running tests, including:
//
// - Initializing the resource management system
// - Creating a map for mock providers
// - Setting up a network simulator with default settings
// - Creating a base context for test execution
// - Setting up signal handling infrastructure
//
// Parameters:
//   - config: The TestConfig that defines the test environment configuration
//   - logger: The Logger to use for structured logging during tests
//
// If the ArtifactsDir in the config is not set, it will be set to a default directory.
//
// Example usage:
//
//	config := framework.TestConfig{
//	    Environment:    "test",
//	    Timeout:        30,
//	    VerboseLogging: true,
//	}
//	logger := log.With().Str("component", "test").Logger()
//	tf := framework.NewTestFramework(config, &logger)
//
// Returns a new TestFramework instance ready for use in tests.
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
// This method adds a mock provider to the TestFramework's registry of mock providers,
// making it available for use in tests. It also automatically registers the provider
// with the network simulator, allowing network conditions to affect the provider.
//
// Mock providers are used to simulate external dependencies, such as the Microsoft
// Graph API, filesystem operations, or UI interactions, without requiring actual
// external services or components.
//
// Parameters:
//   - name: A string identifier for the mock provider, used to retrieve it later
//   - provider: The MockProvider implementation to register
//
// Example usage:
//
//	// Create a mock graph provider
//	mockGraph := mock.NewMockGraphProvider()
//
//	// Register it with the framework
//	framework.RegisterMockProvider("graph", mockGraph)
//
//	// Configure the mock provider
//	mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
//	    ID:   "root",
//	    Name: "root",
//	})
func (tf *TestFramework) RegisterMockProvider(name string, provider MockProvider) {
	tf.mockProviders[name] = provider

	// Register the provider with the network simulator
	if tf.networkSimulator != nil {
		tf.networkSimulator.RegisterProvider(provider)
	}
}

// GetMockProvider returns the mock provider with the given name.
// This method retrieves a previously registered mock provider by name.
// It returns the provider and a boolean indicating whether the provider exists.
//
// Parameters:
//   - name: The string identifier of the mock provider to retrieve
//
// Example usage:
//
//	// Get a registered mock provider
//	provider, exists := framework.GetMockProvider("graph")
//	if exists {
//	    // Type assertion to get the specific mock provider type
//	    mockGraph := provider.(*mock.MockGraphProvider)
//
//	    // Use the mock provider
//	    mockGraph.AddMockResponse("/me/drive/root/children", children)
//	} else {
//	    // Handle case where provider doesn't exist
//	    log.Error("Graph provider not registered")
//	}
//
// Returns:
//   - The MockProvider implementation if found
//   - A boolean indicating whether the provider exists
func (tf *TestFramework) GetMockProvider(name string) (MockProvider, bool) {
	provider, exists := tf.mockProviders[name]
	return provider, exists
}

// SetCoverageReporter sets the coverage reporter for the test
func (tf *TestFramework) SetCoverageReporter(reporter CoverageReporter) {
	tf.coverageReporter = reporter
}

// RunTest executes a single test function with the given name.
// This method runs a test function with the specified name and returns a TestResult
// containing information about the test execution, including status, duration, failures,
// and artifacts.
//
// The test function is executed with a context that may include a timeout if configured
// in the TestConfig. If the test function returns an error, the test is considered failed,
// and the error is recorded in the TestResult.
//
// Parameters:
//   - name: A string identifier for the test, used in logs and the TestResult
//   - testFunc: The function to execute, which takes a context and returns an error
//
// The testFunc parameter should contain the actual test logic and return an error if
// the test fails. The context passed to testFunc can be used to handle timeouts and
// cancellation.
//
// Example usage:
//
//	result := framework.RunTest("my-feature-test", func(ctx context.Context) error {
//	    // Test logic here
//	    select {
//	    case <-ctx.Done():
//	        return ctx.Err()
//	    default:
//	        // Perform test operations
//	        if err := someOperation(); err != nil {
//	            return err
//	        }
//	        return nil
//	    }
//	})
//
//	// Check the test result
//	if result.Status == framework.TestStatusPassed {
//	    fmt.Println("Test passed!")
//	} else {
//	    fmt.Printf("Test failed: %v\n", result.Failures)
//	}
//
// Returns a TestResult containing information about the test execution.
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
// This method runs a collection of test functions as a suite and returns an array
// of TestResult objects, one for each test in the suite. The tests are executed
// sequentially in an undefined order (as map iteration order is not guaranteed).
//
// The method logs the start and completion of the test suite, including the total
// duration of all tests.
//
// Parameters:
//   - name: A string identifier for the test suite, used in logs
//   - tests: A map of test names to test functions
//
// Each test function in the map should take a context and return an error, just like
// the testFunc parameter in RunTest. The keys in the map are used as the test names.
//
// Example usage:
//
//	tests := map[string]func(ctx context.Context) error{
//	    "test1": func(ctx context.Context) error {
//	        // Test 1 logic
//	        return nil
//	    },
//	    "test2": func(ctx context.Context) error {
//	        // Test 2 logic
//	        return nil
//	    },
//	}
//
//	results := framework.RunTestSuite("my-test-suite", tests)
//
//	// Check the results
//	for _, result := range results {
//	    if result.Status == framework.TestStatusFailed {
//	        fmt.Printf("Test %s failed: %v\n", result.Name, result.Failures)
//	    }
//	}
//
// Returns a slice of TestResult objects, one for each test in the suite.
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

// WithTimeout returns a new context with the specified timeout and a cancel function.
func (tf *TestFramework) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(tf.ctx, timeout)
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
// This method configures the network simulator with specific latency, packet loss,
// and bandwidth values to simulate different network conditions during tests.
//
// Parameters:
//   - latency: The delay in data transmission (in time.Duration)
//   - packetLoss: The failure rate of data packets (0.0 to 1.0, where 1.0 = 100%)
//   - bandwidth: The maximum data transfer rate in Kbps
//
// Example usage:
//
//	// Set network conditions to simulate a slow connection
//	err := framework.SetNetworkConditions(200*time.Millisecond, 0.05, 1000)
//	if err != nil {
//	    // Handle error
//	}
//
// Returns an error if the parameters are invalid or if the network conditions
// could not be set.
func (tf *TestFramework) SetNetworkConditions(latency time.Duration, packetLoss float64, bandwidth int) error {
	return tf.networkSimulator.SetConditions(latency, packetLoss, bandwidth)
}

// ApplyNetworkPreset applies a predefined network condition preset.
// This method applies a predefined set of network conditions to simulate
// common network scenarios like fast networks, slow connections, mobile
// networks, etc.
//
// Parameter:
//   - preset: A predefined NetworkCondition (e.g., FastNetwork, SlowNetwork)
//
// Example usage:
//
//	// Apply the SlowNetwork preset
//	err := framework.ApplyNetworkPreset(framework.SlowNetwork)
//	if err != nil {
//	    // Handle error
//	}
//
// Available presets include:
//   - FastNetwork: Fast, reliable network (10ms latency, 0% packet loss, 100Mbps)
//   - AverageNetwork: Average home broadband (50ms latency, 1% packet loss, 20Mbps)
//   - SlowNetwork: Slow connection (200ms latency, 5% packet loss, 1Mbps)
//   - MobileNetwork: Mobile data (100ms latency, 2% packet loss, 5Mbps)
//   - IntermittentConnection: Unstable connection (300ms latency, 15% packet loss, 2Mbps)
//   - SatelliteConnection: High-latency satellite (700ms latency, 3% packet loss, 10Mbps)
//
// Returns an error if the preset could not be applied.
func (tf *TestFramework) ApplyNetworkPreset(preset NetworkCondition) error {
	return tf.networkSimulator.ApplyPreset(preset)
}

// DisconnectNetwork simulates a network disconnection.
// This method sets the network to a completely disconnected state (100% packet loss),
// which is useful for testing how the system behaves when the network is unavailable.
//
// Example usage:
//
//	// Disconnect the network
//	err := framework.DisconnectNetwork()
//	if err != nil {
//	    // Handle error
//	}
//
//	// Test behavior when network is disconnected
//	// ...
//
//	// Reconnect the network
//	err = framework.ReconnectNetwork()
//	if err != nil {
//	    // Handle error
//	}
//
// Returns an error if the network is already disconnected or if the disconnection
// could not be simulated.
func (tf *TestFramework) DisconnectNetwork() error {
	return tf.networkSimulator.Disconnect()
}

// ReconnectNetwork restores the network connection.
// This method restores the network to the conditions that were in effect before
// the DisconnectNetwork method was called. This is useful for testing how the system
// recovers when the network becomes available again after a disconnection.
//
// Example usage:
//
//	// Disconnect the network
//	framework.DisconnectNetwork()
//
//	// Test behavior when network is disconnected
//	// ...
//
//	// Reconnect the network
//	err := framework.ReconnectNetwork()
//	if err != nil {
//	    // Handle error
//	}
//
// Returns an error if the network is already connected or if the reconnection
// could not be simulated.
func (tf *TestFramework) ReconnectNetwork() error {
	return tf.networkSimulator.Reconnect()
}

// IsNetworkConnected returns whether the network is currently connected.
// This method checks if the network simulator is in a connected state.
//
// Example usage:
//
//	// Check if the network is connected
//	if !framework.IsNetworkConnected() {
//	    // Handle disconnected state
//	    fmt.Println("Network is disconnected")
//	} else {
//	    fmt.Println("Network is connected")
//	}
//
// Returns true if the network is connected, false if it is disconnected.
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
		tf.logger.Info("Signal handling already set up", "isHandling", tf.isHandling)
		return func() {
			tf.logger.Info("Signal handling already stopped by another call", "isHandling", tf.isHandling)
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
		case sig, ok := <-tf.signalChan:
			// Check if the channel is closed
			if !ok {
				tf.logger.Debug("Signal channel closed, exiting signal handling goroutine")
				return
			}

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
		tf.logger.Debug("Cleanup function called", "before_lock_isHandling", tf.isHandling)
		tf.signalMu.Lock()
		defer tf.signalMu.Unlock()
		tf.logger.Debug("Cleanup function acquired lock", "after_lock_isHandling", tf.isHandling)

		if !tf.isHandling {
			tf.logger.Info("Signal handling already stopped", "isHandling", tf.isHandling)
			return
		}

		// Stop receiving signals
		tf.logger.Debug("Stopping signal reception", "signalChan", tf.signalChan != nil)
		signal.Stop(tf.signalChan)

		// Cancel the context to stop the goroutine
		if tf.signalCancel != nil {
			tf.logger.Debug("Cancelling signal context")
			tf.signalCancel()
		}

		// Only close the channel if it's not nil
		if tf.signalChan != nil {
			tf.logger.Debug("Closing signal channel")
			close(tf.signalChan)
			tf.signalChan = nil
		}

		// Set the isHandling flag to false
		tf.logger.Debug("Setting isHandling to false", "before", tf.isHandling)
		tf.isHandling = false
		tf.logger.Debug("Set isHandling to false", "after", tf.isHandling)

		tf.logger.Info("Signal handling stopped", "isHandling", tf.isHandling)
	}
}
