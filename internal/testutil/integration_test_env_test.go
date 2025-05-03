// integration_test_env_test.go contains tests for the IntegrationTestEnvironment
package testutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationTestEnvironment_Setup(t *testing.T) {
	// Create a test logger
	logger := NewZerologLogger("integration-test")

	// Create a context
	ctx := context.Background()

	// Create a test environment
	env := NewIntegrationTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up isolation config to mock all components
	env.SetIsolationConfig(IsolationConfig{
		MockedServices: []string{"graph", "filesystem", "ui"},
		NetworkRules:   []NetworkRule{},
		DataIsolation:  true,
	})

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Verify that components were created
	graphComponent, err := env.GetComponent("graph")
	require.NoError(t, err)
	require.NotNil(t, graphComponent)

	fsComponent, err := env.GetComponent("filesystem")
	require.NoError(t, err)
	require.NotNil(t, fsComponent)

	uiComponent, err := env.GetComponent("ui")
	require.NoError(t, err)
	require.NotNil(t, uiComponent)

	// Verify that network simulator was created
	networkSimulator := env.GetNetworkSimulator()
	require.NotNil(t, networkSimulator)

	// Verify that test data manager was created
	testDataManager := env.GetTestDataManager()
	require.NotNil(t, testDataManager)

	// Tear down the environment
	err = env.TeardownEnvironment()
	require.NoError(t, err)
}

func TestIntegrationTestEnvironment_TestDataManager(t *testing.T) {
	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp(TestSandboxTmpDir, "test-data-manager")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test data set
	dataSetDir := filepath.Join(tempDir, "test-data-set")
	err = os.Mkdir(dataSetDir, 0755)
	require.NoError(t, err)

	// Create a test file in the data set
	testFile := filepath.Join(dataSetDir, "test-file.txt")
	err = os.WriteFile(testFile, []byte("test data"), 0644)
	require.NoError(t, err)

	// Create a test data manager
	testDataManager := NewTestDataManager(tempDir)
	require.NotNil(t, testDataManager)

	// Load the test data set
	err = testDataManager.LoadTestData("test-data-set")
	require.NoError(t, err)

	// Get the test data
	data := testDataManager.GetTestData("test-file.txt")
	require.NotNil(t, data)
	assert.Equal(t, []byte("test data"), data)

	// Clean up the test data
	err = testDataManager.CleanupTestData()
	require.NoError(t, err)

	// Verify that the data was cleaned up
	data = testDataManager.GetTestData("test-file.txt")
	assert.Nil(t, data)
}

func TestIntegrationTestEnvironment_RunScenario(t *testing.T) {
	// Create a test logger
	logger := NewZerologLogger("integration-test")

	// Create a context
	ctx := context.Background()

	// Create a test environment
	env := NewIntegrationTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up isolation config to mock all components
	env.SetIsolationConfig(IsolationConfig{
		MockedServices: []string{"graph", "filesystem", "ui"},
		NetworkRules:   []NetworkRule{},
		DataIsolation:  true,
	})

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)
	defer env.TeardownEnvironment()

	// Create a test scenario
	scenario := TestScenario{
		Name:        "Test Scenario",
		Description: "A test scenario",
		Steps: []TestStep{
			{
				Name: "Step 1",
				Action: func(ctx context.Context) error {
					return nil
				},
			},
			{
				Name: "Step 2",
				Action: func(ctx context.Context) error {
					return nil
				},
				Validation: func(ctx context.Context) error {
					return nil
				},
			},
		},
		Assertions: []TestAssertion{
			{
				Name: "Assertion 1",
				Condition: func(ctx context.Context) bool {
					return true
				},
				Message: "Assertion 1 failed",
			},
		},
		Cleanup: []CleanupStep{
			{
				Name: "Cleanup 1",
				Action: func(ctx context.Context) error {
					return nil
				},
				AlwaysRun: true,
			},
		},
	}

	// Add the scenario to the environment
	env.AddScenario(scenario)

	// Run the scenario
	err = env.RunScenario("Test Scenario")
	require.NoError(t, err)
}

func TestIntegrationTestEnvironment_NetworkSimulation(t *testing.T) {
	// Create a test logger
	logger := NewZerologLogger("integration-test")

	// Create a context
	ctx := context.Background()

	// Create a test environment
	env := NewIntegrationTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up isolation config to mock all components
	env.SetIsolationConfig(IsolationConfig{
		MockedServices: []string{"graph"},
		NetworkRules:   []NetworkRule{},
		DataIsolation:  true,
	})

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)
	defer env.TeardownEnvironment()

	// Get the network simulator
	networkSimulator := env.GetNetworkSimulator()
	require.NotNil(t, networkSimulator)

	// Verify that the network is connected
	assert.True(t, networkSimulator.IsConnected())

	// Disconnect the network
	err = networkSimulator.Disconnect()
	require.NoError(t, err)

	// Verify that the network is disconnected
	assert.False(t, networkSimulator.IsConnected())

	// Reconnect the network
	err = networkSimulator.Reconnect()
	require.NoError(t, err)

	// Verify that the network is connected again
	assert.True(t, networkSimulator.IsConnected())

	// Set network conditions
	err = networkSimulator.SetConditions(100*time.Millisecond, 0.1, 1000)
	require.NoError(t, err)

	// Get current conditions
	conditions := networkSimulator.GetCurrentConditions()
	assert.Equal(t, 100*time.Millisecond, conditions.Latency)
	assert.Equal(t, 0.1, conditions.PacketLoss)
	assert.Equal(t, 1000, conditions.Bandwidth)
}

func TestIntegrationTestEnvironment_ComponentIsolation(t *testing.T) {
	// Create a test logger
	logger := NewZerologLogger("integration-test")

	// Create a context
	ctx := context.Background()

	// Create a test environment
	env := NewIntegrationTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up isolation config with network rules
	env.SetIsolationConfig(IsolationConfig{
		MockedServices: []string{"graph", "filesystem"},
		NetworkRules: []NetworkRule{
			{
				Source:      "graph",
				Destination: "filesystem",
				Allow:       false,
			},
		},
		DataIsolation: true,
	})

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)
	defer env.TeardownEnvironment()

	// Verify that components were created
	graphComponent, err := env.GetComponent("graph")
	require.NoError(t, err)
	require.NotNil(t, graphComponent)

	fsComponent, err := env.GetComponent("filesystem")
	require.NoError(t, err)
	require.NotNil(t, fsComponent)

	// Get the isolation config
	isolationConfig := env.GetIsolationConfig()
	require.NotNil(t, isolationConfig)

	// Verify the network rules
	require.Len(t, isolationConfig.NetworkRules, 1)
	assert.Equal(t, "graph", isolationConfig.NetworkRules[0].Source)
	assert.Equal(t, "filesystem", isolationConfig.NetworkRules[0].Destination)
	assert.False(t, isolationConfig.NetworkRules[0].Allow)
}
