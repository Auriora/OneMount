# Integration Testing Framework

This document provides an overview of the integration testing framework for the OneMount project.

## Overview

The integration testing framework extends the functionality of the test framework with additional utilities specifically designed for testing component interactions. It provides:

1. Utilities for setting up integrated components
2. Helpers for configuring component interactions
3. Utilities for verifying interface contracts
4. Support for defining and executing integration test scenarios
5. Utilities for testing component interactions under various conditions

## Key Components

### IntegrationFramework

The `IntegrationFramework` is the main entry point for integration testing. It provides methods for:

- Setting up integrated components
- Configuring component interactions
- Validating interface contracts
- Creating and running integration test scenarios
- Testing component interactions under various conditions

### ComponentInteractionConfig

The `ComponentInteractionConfig` defines how components interact with each other. It includes:

- Source component
- Target component
- Interaction type (e.g., "sync", "async", "event-driven")
- Custom configuration options

### InterfaceContractValidator

The `InterfaceContractValidator` validates that a component implements an interface correctly. It includes:

- Interface type to validate against
- Validation functions for each method
- Custom validation options

### InteractionCondition

The `InteractionCondition` represents a condition under which component interactions are tested. It includes:

- Setup function to establish the condition
- Teardown function to clean up after testing
- Description of the condition

## Usage Examples

### Setting Up Integrated Components

Example:

    // Create a new integration framework
    ctx := context.Background()
    logger := &TestLogger{}
    framework := NewIntegrationFramework(ctx, logger)

    // Set up integrated components
    err := framework.SetupIntegratedComponents("graph", "filesystem", "ui")
    if err != nil {
        // Handle error
    }

### Configuring Component Interactions

Example:

    // Configure how components interact with each other
    config := &ComponentInteractionConfig{
        Name:            "fs-graph-interaction",
        SourceComponent: "filesystem",
        TargetComponent: "graph",
        InteractionType: "sync",
        Options: map[string]interface{}{
            "retryCount": 3,
            "timeout":    5 * time.Second,
        },
    }

    err := framework.ConfigureComponentInteraction(config)
    if err != nil {
        // Handle error
    }

### Validating Interface Contracts

Example:

    // Create a validator for an interface
    validator := &InterfaceContractValidator{
        Name:         "graph-client-validator",
        InterfaceType: reflect.TypeOf((*graph.Client)(nil)).Elem(),
        MethodValidators: map[string]MethodValidator{
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
    if err != nil {
        // Handle error
    }

### Creating and Running Integration Test Scenarios

Example:

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
            // Implement interaction logic
            return nil
        },
        func(ctx context.Context, source, target interface{}) error {
            // Implement validation logic
            return nil
        },
    )

    // Run the integration test
    err := framework.RunIntegrationTest(scenario)
    if err != nil {
        // Handle error
    }

### Testing Component Interactions Under Various Conditions

Example:

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
    err := framework.TestComponentInteractionUnderCondition(
        ctx,
        config,
        slowNetworkCondition,
        func(ctx context.Context, source, target interface{}) error {
            // Test interaction under slow network condition
            return nil
        },
    )
    if err != nil {
        // Handle error
    }

## Best Practices

1. **Component Isolation**: Use mocks for external dependencies to isolate the components being tested.
2. **Test Data Management**: Use the test data manager to load and clean up test data.
3. **Network Simulation**: Use network conditions to test component interactions under various network conditions.
4. **Error Handling**: Test how components handle errors from other components.
5. **Cleanup**: Always clean up resources after tests, even if they fail.

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
