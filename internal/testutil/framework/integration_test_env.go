// integration_test_env.go implements the IntegrationTestEnvironment for testing
package framework

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/graph/mock"
	"github.com/auriora/onemount/internal/testutil"
)

// NetworkRule represents a network isolation rule
type NetworkRule struct {
	Source      string
	Destination string
	Allow       bool
}

// IsolationConfig defines component isolation for tests
type IsolationConfig struct {
	// List of services that should be mocked
	MockedServices []string
	// Network rules for component isolation
	NetworkRules []NetworkRule
	// Whether to isolate test data
	DataIsolation bool
}

// TestDataManager manages test data for integration tests
type TestDataManager interface {
	// LoadTestData loads test data from a specified data set
	LoadTestData(dataSet string) error
	// CleanupTestData cleans up test data
	CleanupTestData() error
	// GetTestData retrieves test data by key
	GetTestData(key string) interface{}
}

// DefaultTestDataManager is the default implementation of TestDataManager
type DefaultTestDataManager struct {
	// Base directory for test data
	baseDir string
	// Current data set
	currentDataSet string
	// Loaded test data
	data map[string]interface{}
	// Mutex for thread safety
	mu sync.Mutex
}

// NewTestDataManager creates a new DefaultTestDataManager
func NewTestDataManager(baseDir string) *DefaultTestDataManager {
	return &DefaultTestDataManager{
		baseDir: baseDir,
		data:    make(map[string]interface{}),
	}
}

// LoadTestData loads test data from a specified data set
func (m *DefaultTestDataManager) LoadTestData(dataSet string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clean up any existing data
	m.data = make(map[string]interface{})

	// Set the current data set
	m.currentDataSet = dataSet

	// Create the full path to the data set
	dataSetPath := filepath.Join(m.baseDir, dataSet)

	// Check if the data set exists
	if _, err := os.Stat(dataSetPath); os.IsNotExist(err) {
		return fmt.Errorf("data set %s does not exist", dataSet)
	}

	// Load data from the data set
	// This is a simple implementation that just loads files from the data set directory
	// In a real implementation, this would parse the files and load them into the data map
	files, err := os.ReadDir(dataSetPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dataSetPath, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// Store the data with the file name as the key
		m.data[file.Name()] = data
	}

	return nil
}

// CleanupTestData cleans up test data
func (m *DefaultTestDataManager) CleanupTestData() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear the data map
	m.data = make(map[string]interface{})
	m.currentDataSet = ""

	return nil
}

// GetTestData retrieves test data by key
func (m *DefaultTestDataManager) GetTestData(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.data[key]
}

// Using existing TestStep, TestAssertion, CleanupStep, and TestScenario types from scenario.go

// IntegrationTestEnvironment provides a controlled environment for integration tests
type IntegrationTestEnvironment struct {
	// Real or mock components based on configuration
	components map[string]interface{}

	// Network simulation
	networkSimulator NetworkSimulator

	// Test data management
	testData TestDataManager

	// Test scenarios
	scenarios []TestScenario

	// Component isolation
	isolation IsolationConfig

	// Logger for test output
	logger Logger

	// Context for test execution
	ctx context.Context

	// Mutex for thread safety
	mu sync.Mutex
}

// NewIntegrationTestEnvironment creates a new IntegrationTestEnvironment
func NewIntegrationTestEnvironment(ctx context.Context, logger Logger) *IntegrationTestEnvironment {
	return &IntegrationTestEnvironment{
		components:       make(map[string]interface{}),
		networkSimulator: NewNetworkSimulator(),
		testData:         NewTestDataManager(filepath.Join(testutil.TestSandboxTmpDir, "onemount-test-data")),
		scenarios:        make([]TestScenario, 0),
		isolation: IsolationConfig{
			MockedServices: make([]string, 0),
			NetworkRules:   make([]NetworkRule, 0),
			DataIsolation:  true,
		},
		logger: logger,
		ctx:    ctx,
	}
}

// SetupEnvironment sets up the test environment
func (e *IntegrationTestEnvironment) SetupEnvironment() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.Info("Setting up integration test environment")

	// Initialize components based on isolation configuration
	if err := e.setupComponents(); err != nil {
		return err
	}

	// Apply network rules
	if err := e.applyNetworkRules(); err != nil {
		return err
	}

	return nil
}

// TeardownEnvironment tears down the test environment
func (e *IntegrationTestEnvironment) TeardownEnvironment() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.Info("Tearing down integration test environment")

	// Clean up test data
	if err := e.testData.CleanupTestData(); err != nil {
		e.logger.Error("Failed to clean up test data", "error", err)
		return err
	}

	// Reset network simulator
	if e.networkSimulator != nil {
		if !e.networkSimulator.IsConnected() {
			if err := e.networkSimulator.Reconnect(); err != nil {
				e.logger.Error("Failed to reconnect network", "error", err)
				return err
			}
		}
	}

	// Clean up components
	for name, component := range e.components {
		if provider, ok := component.(MockProvider); ok {
			if err := provider.Teardown(); err != nil {
				e.logger.Error("Failed to tear down component", "component", name, "error", err)
				return err
			}
		}
	}

	// Clear components
	e.components = make(map[string]interface{})

	return nil
}

// SetupComponents configures which components are real and which are mocked
func (e *IntegrationTestEnvironment) setupComponents() error {
	// Register mock providers based on isolation configuration
	for _, service := range e.isolation.MockedServices {
		var provider MockProvider
		switch service {
		case "graph":
			provider = mock.NewMockGraphProvider()
		case "filesystem":
			provider = mock.NewMockFileSystemProvider()
		case "ui":
			provider = mock.NewMockUIProvider()
		default:
			return fmt.Errorf("unknown service: %s", service)
		}

		// Setup the provider
		if err := provider.Setup(); err != nil {
			return err
		}

		// Register the provider
		e.components[service] = provider

		// Register with network simulator
		e.networkSimulator.RegisterProvider(provider)
	}

	return nil
}

// applyNetworkRules applies network isolation rules
func (e *IntegrationTestEnvironment) applyNetworkRules() error {
	// This is a simplified implementation
	// In a real implementation, this would configure network isolation between components
	// For now, we just log the rules
	for _, rule := range e.isolation.NetworkRules {
		action := "block"
		if rule.Allow {
			action = "allow"
		}
		e.logger.Info("Applying network rule", "source", rule.Source, "destination", rule.Destination, "action", action)
	}

	return nil
}

// GetComponent returns a component by name
func (e *IntegrationTestEnvironment) GetComponent(name string) (interface{}, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	component, exists := e.components[name]
	if !exists {
		return nil, fmt.Errorf("component not found: %s", name)
	}

	return component, nil
}

// SetComponent sets a component by name
func (e *IntegrationTestEnvironment) SetComponent(name string, component interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.components[name] = component
}

// GetNetworkSimulator returns the network simulator
func (e *IntegrationTestEnvironment) GetNetworkSimulator() NetworkSimulator {
	return e.networkSimulator
}

// SetNetworkSimulator sets the network simulator
func (e *IntegrationTestEnvironment) SetNetworkSimulator(simulator NetworkSimulator) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.networkSimulator = simulator
}

// GetTestDataManager returns the test data manager
func (e *IntegrationTestEnvironment) GetTestDataManager() TestDataManager {
	return e.testData
}

// SetTestDataManager sets the test data manager
func (e *IntegrationTestEnvironment) SetTestDataManager(manager TestDataManager) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.testData = manager
}

// GetIsolationConfig returns the isolation configuration
func (e *IntegrationTestEnvironment) GetIsolationConfig() IsolationConfig {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.isolation
}

// SetIsolationConfig sets the isolation configuration
func (e *IntegrationTestEnvironment) SetIsolationConfig(config IsolationConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.isolation = config
}

// AddScenario adds a test scenario
func (e *IntegrationTestEnvironment) AddScenario(scenario TestScenario) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.scenarios = append(e.scenarios, scenario)
}

// RunScenario runs a test scenario
func (e *IntegrationTestEnvironment) RunScenario(scenarioName string) error {
	e.mu.Lock()
	var scenario *TestScenario
	for i, s := range e.scenarios {
		if s.Name == scenarioName {
			scenario = &e.scenarios[i]
			break
		}
	}
	e.mu.Unlock()

	if scenario == nil {
		return fmt.Errorf("scenario not found: %s", scenarioName)
	}

	e.logger.Info("Running scenario", "name", scenarioName)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(e.ctx, 5*time.Minute)
	defer cancel()

	// Run each step in the scenario
	for _, step := range scenario.Steps {
		e.logger.Info("Running step", "scenario", scenarioName, "step", step.Name)

		// Execute the action
		if err := step.Action(ctx); err != nil {
			e.logger.Error("Step action failed", "scenario", scenarioName, "step", step.Name, "error", err)
			return err
		}

		// Execute the validation
		if step.Validation != nil {
			if err := step.Validation(ctx); err != nil {
				e.logger.Error("Step validation failed", "scenario", scenarioName, "step", step.Name, "error", err)
				return err
			}
		}
	}

	// Check all assertions
	for _, assertion := range scenario.Assertions {
		e.logger.Info("Checking assertion", "scenario", scenarioName, "assertion", assertion.Name)

		if !assertion.Condition(ctx) {
			err := errors.New(assertion.Message)
			e.logger.Error("Assertion failed", "scenario", scenarioName, "assertion", assertion.Name, "error", err)
			return err
		}
	}

	// Run cleanup steps
	for _, cleanup := range scenario.Cleanup {
		e.logger.Info("Running cleanup", "scenario", scenarioName, "cleanup", cleanup.Name)

		if err := cleanup.Action(ctx); err != nil {
			e.logger.Error("Cleanup failed", "scenario", scenarioName, "cleanup", cleanup.Name, "error", err)
			// Continue with other cleanup steps even if one fails
		}
	}

	e.logger.Info("Scenario completed successfully", "name", scenarioName)
	return nil
}

// RunAllScenarios runs all test scenarios
func (e *IntegrationTestEnvironment) RunAllScenarios() []error {
	e.mu.Lock()
	scenarios := make([]TestScenario, len(e.scenarios))
	copy(scenarios, e.scenarios)
	e.mu.Unlock()

	errors := make([]error, 0)

	for _, scenario := range scenarios {
		if err := e.RunScenario(scenario.Name); err != nil {
			errors = append(errors, fmt.Errorf("scenario %s failed: %w", scenario.Name, err))
		}
	}

	return errors
}
