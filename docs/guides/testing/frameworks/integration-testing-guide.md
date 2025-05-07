# Integration Testing Guide

## Overview

Integration testing is a critical testing approach that focuses on verifying that different components of the system work together correctly. In the OneMount project, integration tests ensure that components interact properly with each other and with external dependencies.

The OneMount test framework provides specialized utilities for integration testing, making it easier to set up controlled test environments, configure component interactions, and verify interface contracts.

## Key Concepts

### Integration Testing Principles

1. **Component Interaction**: Test how components interact with each other
2. **Interface Contracts**: Verify that components implement their interfaces correctly
3. **Controlled Environment**: Test in an environment that simulates production but with controlled conditions
4. **Realistic Scenarios**: Test realistic usage scenarios that involve multiple components
5. **Isolation**: Isolate the components being tested from external dependencies when necessary

### Integration Test Structure

Integration tests in the OneMount project follow a standard structure:

1. **Environment Setup**: Set up the integration test environment with the necessary components
2. **Component Configuration**: Configure how components interact with each other
3. **Scenario Execution**: Execute test scenarios that involve multiple components
4. **Verification**: Verify that components interact correctly and produce the expected results
5. **Environment Teardown**: Clean up the test environment

## Integration Testing Framework

The integration testing framework extends the functionality of the test framework with additional utilities specifically designed for testing component interactions. It provides:

1. Utilities for setting up integrated components
2. Helpers for configuring component interactions
3. Utilities for verifying interface contracts
4. Support for defining and executing integration test scenarios
5. Utilities for testing component interactions under various conditions

### IntegrationTestEnvironment

The `IntegrationTestEnvironment` provides a controlled environment for integration tests:

```go
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
}
```

### Creating an IntegrationTestEnvironment

```go
// Create a new integration test environment
ctx := context.Background()
logger := &testutil.TestLogger{}
env := testutil.NewIntegrationTestEnvironment(ctx, logger)

// Set up the environment
err := env.SetupEnvironment()
if err != nil {
    // Handle error
}

// Add cleanup to ensure the environment is torn down
t.Cleanup(func() {
    env.TeardownEnvironment()
})
```

### IsolationConfig

The `IsolationConfig` defines component isolation for tests:

```go
type IsolationConfig struct {
    // List of services that should be mocked
    MockedServices []string
    // Network rules for component isolation
    NetworkRules []NetworkRule
    // Whether to isolate test data
    DataIsolation bool
}

type NetworkRule struct {
    Source      string
    Destination string
    Allow       bool
}
```

### Configuring Component Isolation

```go
// Configure which components should be mocked
env.SetIsolationConfig(testutil.IsolationConfig{
    MockedServices: []string{"graph", "filesystem", "ui"},
    NetworkRules: []testutil.NetworkRule{
        {
            Source:      "filesystem",
            Destination: "graph",
            Allow:       true,
        },
        {
            Source:      "ui",
            Destination: "filesystem",
            Allow:       true,
        },
    },
    DataIsolation: true,
})
```

### TestDataManager

The `TestDataManager` manages test data for integration tests:

```go
type TestDataManager interface {
    // LoadTestData loads test data from a specified data set
    LoadTestData(dataSet string) error
    // CleanupTestData cleans up test data
    CleanupTestData() error
    // GetTestData retrieves test data by key
    GetTestData(key string) interface{}
}
```

### Managing Test Data

```go
// Get the test data manager
testDataManager := env.GetTestDataManager()

// Load test data
err := testDataManager.LoadTestData("test-data-set")
if err != nil {
    // Handle error
}

// Get test data
data := testDataManager.GetTestData("test-file.txt")

// Clean up test data
err = testDataManager.CleanupTestData()
if err != nil {
    // Handle error
}
```

### TestScenario

The `TestScenario` represents a scenario-based test:

```go
type TestScenario struct {
    // Name of the scenario
    Name string
    // Description of the scenario
    Description string
    // Tags for categorizing scenarios
    Tags []string
    // Steps to execute in the scenario
    Steps []TestStep
    // Assertions to verify after steps are executed
    Assertions []TestAssertion
    // Cleanup steps to run after the scenario
    Cleanup []CleanupStep
}

type TestStep struct {
    // Name of the step
    Name string
    // Action to perform in the step
    Action func(ctx context.Context) error
    // Validation to perform after the action
    Validation func(ctx context.Context) error
}

type TestAssertion struct {
    // Name of the assertion
    Name string
    // Condition to check
    Condition func(ctx context.Context) bool
    // Message to display if the condition is not met
    Message string
}

type CleanupStep struct {
    // Name of the cleanup step
    Name string
    // Action to perform in the cleanup step
    Action func(ctx context.Context) error
    // Whether to always run the cleanup step, even if the scenario fails
    AlwaysRun bool
}
```

### Creating and Running Test Scenarios

```go
// Create a test scenario
scenario := testutil.TestScenario{
    Name:        "File Operations",
    Description: "Tests file creation, modification, and deletion",
    Steps: []testutil.TestStep{
        {
            Name: "Create file",
            Action: func(ctx context.Context) error {
                // Implementation would create a file
                return nil
            },
            Validation: func(ctx context.Context) error {
                // Implementation would verify file was created
                return nil
            },
        },
        {
            Name: "Modify file",
            Action: func(ctx context.Context) error {
                // Implementation would modify the file
                return nil
            },
            Validation: func(ctx context.Context) error {
                // Implementation would verify file was modified
                return nil
            },
        },
        // More steps...
    },
    Assertions: []testutil.TestAssertion{
        {
            Name: "File operations completed successfully",
            Condition: func(ctx context.Context) bool {
                // Implementation would check if all operations were successful
                return true
            },
            Message: "File operations did not complete successfully",
        },
    },
    Cleanup: []testutil.CleanupStep{
        {
            Name: "Clean up test files",
            Action: func(ctx context.Context) error {
                // Implementation would clean up any remaining test files
                return nil
            },
            AlwaysRun: true,
        },
    },
}

// Add the scenario to the environment
env.AddScenario(scenario)

// Run the scenario
err := env.RunScenario("File Operations")
if err != nil {
    // Handle error
}

// Run all scenarios
errors := env.RunAllScenarios()
if len(errors) > 0 {
    // Handle errors
}
```

### IntegrationFramework

The `IntegrationFramework` is the main entry point for integration testing. It provides methods for:

- Setting up integrated components
- Configuring component interactions
- Validating interface contracts
- Creating and running integration test scenarios
- Testing component interactions under various conditions

```go
type IntegrationFramework struct {
    // Embedded TestFramework for core functionality
    *TestFramework

    // Integration test environment
    Environment *IntegrationTestEnvironment

    // Component interaction configuration
    InteractionConfigs []*ComponentInteractionConfig

    // Interface contract validators
    ContractValidators []*InterfaceContractValidator
}
```

### ComponentInteractionConfig

The `ComponentInteractionConfig` defines how components interact with each other:

```go
type ComponentInteractionConfig struct {
    // Name of the interaction configuration
    Name string
    // Source component
    SourceComponent string
    // Target component
    TargetComponent string
    // Interaction type (e.g., "sync", "async", "event-driven")
    InteractionType string
    // Custom configuration options
    Options map[string]interface{}
}
```

### InterfaceContractValidator

The `InterfaceContractValidator` validates that a component implements an interface correctly:

```go
type InterfaceContractValidator struct {
    // Name of the validator
    Name string
    // Interface type to validate against
    InterfaceType reflect.Type
    // Validation functions for each method
    MethodValidators map[string]MethodValidator
    // Custom validation options
    Options map[string]interface{}
}

type MethodValidator func(component interface{}, args []interface{}) error
```

### InteractionCondition

The `InteractionCondition` represents a condition under which component interactions are tested. It includes:

- Setup function to establish the condition
- Teardown function to clean up after testing
- Description of the condition

## Writing Integration Tests

### Basic Integration Test

```go
func TestIntegration(t *testing.T) {
    // Create a logger
    logger := &testutil.TestLogger{}

    // Create a context
    ctx := context.Background()

    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(ctx, logger)
    require.NotNil(t, env)

    // Set up isolation config to mock all components
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph", "filesystem", "ui"},
        NetworkRules:   []testutil.NetworkRule{},
        DataIsolation:  true,
    })

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Get components
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)

    fsComponent, err := env.GetComponent("filesystem")
    require.NoError(t, err)
    mockFS := fsComponent.(*testutil.MockFileSystemProvider)

    // Configure mock behavior
    mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
        ID:   "root",
        Name: "root",
    })

    mockFS.AddMockFile("/path/to/file.txt", []byte("file content"), nil)

    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "File Synchronization",
        Description: "Tests synchronization between local filesystem and OneDrive",
        Steps: []testutil.TestStep{
            {
                Name: "Initialize synchronization",
                Action: func(ctx context.Context) error {
                    // Implementation would initialize synchronization
                    return nil
                },
                Validation: func(ctx context.Context) error {
                    // Implementation would verify initialization
                    return nil
                },
            },
            {
                Name: "Synchronize file",
                Action: func(ctx context.Context) error {
                    // Implementation would synchronize a file
                    return nil
                },
                Validation: func(ctx context.Context) error {
                    // Implementation would verify synchronization
                    return nil
                },
            },
            // More steps...
        },
        Assertions: []testutil.TestAssertion{
            {
                Name: "Synchronization completed successfully",
                Condition: func(ctx context.Context) bool {
                    // Implementation would check if synchronization was successful
                    return true
                },
                Message: "Synchronization did not complete successfully",
            },
        },
        Cleanup: []testutil.CleanupStep{
            {
                Name: "Clean up synchronized files",
                Action: func(ctx context.Context) error {
                    // Implementation would clean up synchronized files
                    return nil
                },
                AlwaysRun: true,
            },
        },
    }

    // Add the scenario to the environment
    env.AddScenario(scenario)

    // Run the scenario
    err = env.RunScenario("File Synchronization")
    require.NoError(t, err)

    // Verify interactions with components
    mockGraph.VerifyCalled("/me/drive/root")
    mockFS.VerifyOperation("/path/to/file.txt", testutil.FSOperationRead)
}
```

### Testing with Network Conditions

```go
func TestWithNetworkConditions(t *testing.T) {
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), &testutil.TestLogger{})
    require.NotNil(t, env)

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Get the network simulator
    networkSimulator := env.GetNetworkSimulator()
    require.NotNil(t, networkSimulator)

    // Test with normal network conditions
    err = env.RunScenario("Normal Network Scenario")
    require.NoError(t, err)

    // Test with slow network
    err = networkSimulator.SetConditions(500*time.Millisecond, 0.1, 1000)
    require.NoError(t, err)

    err = env.RunScenario("Slow Network Scenario")
    require.NoError(t, err)

    // Test with packet loss
    err = networkSimulator.SetConditions(100*time.Millisecond, 0.2, 1000)
    require.NoError(t, err)

    err = env.RunScenario("Packet Loss Scenario")
    require.NoError(t, err)

    // Test with network disconnection
    err = networkSimulator.Disconnect()
    require.NoError(t, err)

    err = env.RunScenario("Offline Scenario")
    require.NoError(t, err)

    // Reconnect for cleanup
    err = networkSimulator.Reconnect()
    require.NoError(t, err)
}
```

### Testing Component Interactions

```go
func TestComponentInteractions(t *testing.T) {
    // Create an integration framework
    framework := testutil.NewIntegrationFramework(t)

    // Set up integrated components
    err := framework.SetupIntegratedComponents("graph", "filesystem", "ui")
    require.NoError(t, err)

    // Configure how components interact with each other
    config := &testutil.ComponentInteractionConfig{
        Name:            "fs-graph-interaction",
        SourceComponent: "filesystem",
        TargetComponent: "graph",
        InteractionType: "sync",
        Options: map[string]interface{}{
            "retryCount": 3,
            "timeout":    5 * time.Second,
        },
    }

    err = framework.ConfigureComponentInteraction(config)
    require.NoError(t, err)

    // Create an integration test scenario
    scenario := framework.CreateIntegrationTestScenario(
        "file-upload-scenario",
        "Tests file upload from filesystem to OneDrive",
    )

    // Add a step to test component interaction
    framework.AddComponentInteractionStep(
        scenario,
        "upload-file",
        "filesystem",
        "graph",
        func(ctx context.Context, source, target interface{}) error {
            // Implementation would upload a file
            return nil
        },
        func(ctx context.Context, source, target interface{}) error {
            // Implementation would verify the upload
            return nil
        },
    )

    // Run the integration test
    err = framework.RunIntegrationTest(scenario)
    require.NoError(t, err)
}
```

### Testing Interface Contracts

```go
func TestInterfaceContracts(t *testing.T) {
    // Create an integration framework
    framework := testutil.NewIntegrationFramework(t)

    // Create a validator for an interface
    validator := &testutil.InterfaceContractValidator{
        Name:         "graph-client-validator",
        InterfaceType: reflect.TypeOf((*graph.Client)(nil)).Elem(),
        MethodValidators: map[string]testutil.MethodValidator{
            "GetItem": func(component interface{}, args []interface{}) error {
                // Validate GetItem method
                return nil
            },
            "GetItemChildren": func(component interface{}, args []interface{}) error {
                // Validate GetItemChildren method
                return nil
            },
        },
    }

    // Register the validator
    framework.RegisterInterfaceContractValidator(validator)

    // Validate that a component implements the interface correctly
    err := framework.ValidateInterfaceContract("graph", "graph-client-validator")
    require.NoError(t, err)
}
```

### Testing Component Interactions Under Various Conditions

```go
func TestComponentInteractionsUnderConditions(t *testing.T) {
    // Create an integration framework
    framework := testutil.NewIntegrationFramework(t)

    // Set up integrated components
    err := framework.SetupIntegratedComponents("graph", "filesystem")
    require.NoError(t, err)

    // Configure component interaction
    config := &testutil.ComponentInteractionConfig{
        Name:            "fs-graph-interaction",
        SourceComponent: "filesystem",
        TargetComponent: "graph",
        InteractionType: "sync",
    }

    // Create a network condition
    slowNetworkCondition := framework.CreateNetworkCondition(
        "slow-network",
        500*time.Millisecond,
        0.1,
        1000,
    )

    // Create a disconnected condition
    disconnectedCondition := framework.CreateDisconnectedCondition()

    // Create an error condition
    errorCondition := framework.CreateErrorCondition(
        "graph-error",
        "graph",
        func(component interface{}) error {
            // Set up error condition
            return nil
        },
        func(component interface{}) error {
            // Clean up error condition
            return nil
        },
    )

    // Test component interaction under a specific condition
    ctx := context.Background()
    err = framework.TestComponentInteractionUnderCondition(
        ctx,
        config,
        slowNetworkCondition,
        func(ctx context.Context, source, target interface{}) error {
            // Test interaction under slow network condition
            return nil
        },
    )
    require.NoError(t, err)
}
```

## Best Practices

### Do's

1. **Isolate External Dependencies**: Use mocks for external dependencies to create a controlled test environment.

2. **Test Realistic Scenarios**: Design integration tests that reflect real-world usage scenarios.

3. **Test Component Interactions**: Focus on testing how components interact with each other, not just their individual behavior.

4. **Use Scenario-Based Testing**: Organize integration tests as scenarios with clear steps, validations, and assertions.

5. **Test with Different Network Conditions**: Verify that component interactions work correctly under various network conditions.

6. **Manage Test Data Carefully**: Use the test data manager to load and clean up test data.

7. **Clean Up After Tests**: Always clean up resources after tests, even if they fail.

8. **Use Descriptive Scenario Names**: Give scenarios descriptive names that clearly indicate what is being tested.

9. **Include Validation Steps**: Add validation steps to verify the results of actions.

10. **Test Error Handling**: Verify that components handle errors from other components correctly.

### Don'ts

1. **Don't Test Too Many Components at Once**: Keep integration tests focused on specific component interactions.

2. **Don't Rely on External Services**: Use mocks for external services to create a controlled test environment.

3. **Don't Ignore Network Conditions**: Test component interactions under various network conditions.

4. **Don't Skip Cleanup**: Always clean up resources after tests to avoid affecting other tests.

5. **Don't Hardcode Test Data**: Use the test data manager to load and manage test data.

6. **Don't Write Brittle Tests**: Avoid tests that break when implementation details change but behavior remains the same.

7. **Don't Ignore Test Failures**: Address integration test failures promptly to maintain the reliability of the test suite.

8. **Don't Test Implementation Details**: Focus on testing the behavior and interactions, not implementation details.

9. **Don't Write Flaky Tests**: Avoid tests that sometimes pass and sometimes fail without changes to the code.

10. **Don't Skip Error Handling Tests**: Test how components handle errors from other components.

## Integration with Existing Test Framework

The integration testing framework builds on top of the existing test framework and uses its components:

- TestFramework for test configuration, setup, and execution
- MockProviders for simulating external dependencies
- NetworkSimulator for simulating network conditions
- TestScenario for defining test scenarios

## Extending the Framework

The integration testing framework is designed to be extensible. You can:

1. Add new types of conditions for testing component interactions
2. Create custom validators for specific interfaces
3. Define reusable test scenarios for common integration patterns
4. Add new utilities for specific types of component interactions

## Examples

### Testing File Synchronization

```go
func TestFileSynchronization(t *testing.T) {
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), &testutil.TestLogger{})
    require.NotNil(t, env)

    // Set up isolation config
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph"},
        NetworkRules:   []testutil.NetworkRule{},
        DataIsolation:  true,
    })

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Get components
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)

    // Configure mock behavior
    mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
        ID:   "root",
        Name: "root",
    })

    mockGraph.AddMockResponse("/me/drive/root:/path/to/file.txt", &graph.DriveItem{
        ID:   "123",
        Name: "file.txt",
        File: &graph.File{
            MimeType: "text/plain",
        },
    })

    // Create a temporary file for testing
    tempDir := t.TempDir()
    filePath := filepath.Join(tempDir, "file.txt")
    err = os.WriteFile(filePath, []byte("test content"), 0644)
    require.NoError(t, err)

    // Create a synchronization service
    syncService := fs.NewSyncService(mockGraph)

    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "File Synchronization",
        Description: "Tests synchronization of a file to OneDrive",
        Steps: []testutil.TestStep{
            {
                Name: "Synchronize file",
                Action: func(ctx context.Context) error {
                    return syncService.SyncFile(ctx, filePath, "/path/to/file.txt")
                },
                Validation: func(ctx context.Context) error {
                    // Verify the file was synchronized
                    return nil
                },
            },
        },
        Assertions: []testutil.TestAssertion{
            {
                Name: "Graph API was called correctly",
                Condition: func(ctx context.Context) bool {
                    return mockGraph.VerifyCalled("/me/drive/root:/path/to/file.txt")
                },
                Message: "Graph API was not called correctly",
            },
        },
    }

    // Add the scenario to the environment
    env.AddScenario(scenario)

    // Run the scenario
    err = env.RunScenario("File Synchronization")
    require.NoError(t, err)
}
```

### Testing Authentication Flow

```go
func TestAuthenticationFlow(t *testing.T) {
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), &testutil.TestLogger{})
    require.NotNil(t, env)

    // Set up isolation config
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph", "ui"},
        NetworkRules:   []testutil.NetworkRule{},
        DataIsolation:  true,
    })

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Get components
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)

    uiComponent, err := env.GetComponent("ui")
    require.NoError(t, err)
    mockUI := uiComponent.(*testutil.MockUIProvider)

    // Configure mock behavior
    mockGraph.SetAuthState(testutil.AuthStateNone)

    mockUI.AddMockResponse("login-button", testutil.UIOperationClick, map[string]interface{}{
        "success": true,
    })

    // Create an authentication service
    authService := auth.NewAuthService(mockGraph, mockUI)

    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "Authentication Flow",
        Description: "Tests the complete authentication process",
        Steps: []testutil.TestStep{
            {
                Name: "Initialize authentication",
                Action: func(ctx context.Context) error {
                    return authService.Initialize(ctx)
                },
                Validation: func(ctx context.Context) error {
                    // Verify initialization
                    return nil
                },
            },
            {
                Name: "Authenticate user",
                Action: func(ctx context.Context) error {
                    return authService.Authenticate(ctx)
                },
                Validation: func(ctx context.Context) error {
                    // Verify authentication
                    return mockGraph.GetAuthState() == testutil.AuthStateValid
                },
            },
        },
        Assertions: []testutil.TestAssertion{
            {
                Name: "UI interaction occurred",
                Condition: func(ctx context.Context) bool {
                    return mockUI.VerifyOperation("login-button", testutil.UIOperationClick)
                },
                Message: "UI interaction did not occur",
            },
            {
                Name: "Authentication state is valid",
                Condition: func(ctx context.Context) bool {
                    return mockGraph.GetAuthState() == testutil.AuthStateValid
                },
                Message: "Authentication state is not valid",
            },
        },
    }

    // Add the scenario to the environment
    env.AddScenario(scenario)

    // Run the scenario
    err = env.RunScenario("Authentication Flow")
    require.NoError(t, err)
}
```

### Testing Offline Mode

```go
func TestOfflineMode(t *testing.T) {
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), &testutil.TestLogger{})
    require.NotNil(t, env)

    // Set up isolation config
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph"},
        NetworkRules:   []testutil.NetworkRule{},
        DataIsolation:  true,
    })

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Get components
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)

    // Get the network simulator
    networkSimulator := env.GetNetworkSimulator()
    require.NotNil(t, networkSimulator)

    // Configure mock behavior
    mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
        ID:   "root",
        Name: "root",
    })

    // Create an offline service
    offlineService := fs.NewOfflineService(mockGraph)

    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "Offline Mode",
        Description: "Tests operation in offline mode",
        Steps: []testutil.TestStep{
            {
                Name: "Prepare for offline mode",
                Action: func(ctx context.Context) error {
                    return offlineService.PrepareForOffline(ctx)
                },
                Validation: func(ctx context.Context) error {
                    // Verify preparation
                    return nil
                },
            },
            {
                Name: "Disconnect network",
                Action: func(ctx context.Context) error {
                    return networkSimulator.Disconnect()
                },
                Validation: func(ctx context.Context) error {
                    // Verify network is disconnected
                    if networkSimulator.IsConnected() {
                        return errors.New("network is still connected")
                    }
                    return nil
                },
            },
            {
                Name: "Access cached data",
                Action: func(ctx context.Context) error {
                    return offlineService.AccessCachedData(ctx)
                },
                Validation: func(ctx context.Context) error {
                    // Verify cached data access
                    return nil
                },
            },
        },
        Assertions: []testutil.TestAssertion{
            {
                Name: "Offline mode is active",
                Condition: func(ctx context.Context) bool {
                    return offlineService.IsOfflineMode()
                },
                Message: "Offline mode is not active",
            },
        },
        Cleanup: []testutil.CleanupStep{
            {
                Name: "Reconnect network",
                Action: func(ctx context.Context) error {
                    return networkSimulator.Reconnect()
                },
                AlwaysRun: true,
            },
        },
    }

    // Add the scenario to the environment
    env.AddScenario(scenario)

    // Run the scenario
    err = env.RunScenario("Offline Mode")
    require.NoError(t, err)
}
```

## Related Documentation

- [Testing Framework Guide](testing-framework-guide.md)
- [Network Simulator](network-simulator.md)
- [Mock Providers](mock-providers.md)
- [Unit Testing](../unit-testing.md)
- [System Testing](system-testing.md)
- [Troubleshooting](testing-troubleshooting.md)
