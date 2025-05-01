# Security Testing Guide

## Overview

This guide provides instructions for using the security testing framework to test the security aspects of the OneMount system. The framework includes tools for security scanning, simulating security attacks, verifying security controls, testing authentication and authorization, and testing data protection mechanisms.

## Getting Started

### Setting Up the Security Testing Framework

To use the security testing framework, you need to create a `SecurityTestFramework` instance and configure it with a `SecurityTestConfig`:

```go
// Create a logger
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

// Create a security test config
securityConfig := testutil.SecurityTestConfig{
    Timeout:        30,
    VerboseLogging: true,
    ArtifactsDir:   "/tmp/security-test-artifacts",
    CustomOptions:  map[string]interface{}{"option1": "value1"},
}

// Create a security test framework
securityFramework := testutil.NewSecurityTestFramework(securityConfig, logger)
```

### Creating Security Test Scenarios

The security testing framework provides a `SecurityTestScenarios` struct that includes common security test scenarios:

```go
// Create security test scenarios
securityScenarios := testutil.NewSecurityTestScenarios(securityFramework)
```

The following security test scenarios are available:

1. **Vulnerability Scan Scenario**: Scans the system for security vulnerabilities
2. **Security Attack Simulation Scenario**: Simulates various security attacks against the system
3. **Security Control Verification Scenario**: Verifies that security controls are properly implemented and effective
4. **Authentication Test Scenario**: Tests the security of authentication mechanisms
5. **Authorization Test Scenario**: Tests the security of authorization mechanisms
6. **Data Protection Test Scenario**: Tests the security of data protection mechanisms

## Running Security Tests

### Using the Integration Test Environment

You can run security tests using the integration test environment:

```go
// Create an integration test environment
ctx := context.Background()
env := testutil.NewIntegrationTestEnvironment(ctx, logger)

// Setup the environment
if err := env.SetupEnvironment(); err != nil {
    logger.Error("Failed to setup integration test environment", "error", err)
    return
}
defer env.TeardownEnvironment()

// Add security test scenarios to the integration test environment
env.AddScenario(securityScenarios.VulnerabilityScanScenario())
env.AddScenario(securityScenarios.SecurityAttackSimulationScenario())
env.AddScenario(securityScenarios.SecurityControlVerificationScenario())
env.AddScenario(securityScenarios.AuthenticationTestScenario())
env.AddScenario(securityScenarios.AuthorizationTestScenario())
env.AddScenario(securityScenarios.DataProtectionTestScenario())

// Run all security test scenarios
errors := env.RunAllScenarios()
if len(errors) > 0 {
    logger.Error("Some security test scenarios failed", "errors", len(errors))
    for i, err := range errors {
        logger.Error(fmt.Sprintf("Error %d", i+1), "error", err)
    }
} else {
    logger.Info("All security test scenarios passed")
}
```

### Using the System Test Environment

You can also run security tests using the system test environment:

```go
// Create a system test environment
ctx := context.Background()
env := testutil.NewSystemTestEnvironment(ctx, logger)

// Setup the environment
if err := env.SetupEnvironment(); err != nil {
    logger.Error("Failed to setup system test environment", "error", err)
    return
}
defer env.TeardownEnvironment()

// Add security test scenarios to the system test environment
env.AddScenario(securityScenarios.VulnerabilityScanScenario())
env.AddScenario(securityScenarios.SecurityAttackSimulationScenario())
env.AddScenario(securityScenarios.SecurityControlVerificationScenario())
env.AddScenario(securityScenarios.AuthenticationTestScenario())
env.AddScenario(securityScenarios.AuthorizationTestScenario())
env.AddScenario(securityScenarios.DataProtectionTestScenario())

// Run all security test scenarios
errors := env.RunAllScenarios()
if len(errors) > 0 {
    logger.Error("Some security test scenarios failed", "errors", len(errors))
    for i, err := range errors {
        logger.Error(fmt.Sprintf("Error %d", i+1), "error", err)
    }
} else {
    logger.Info("All security test scenarios passed")
}
```

### Using the Test Framework Directly

You can also use the security testing framework directly with the main test framework:

```go
// Create a test framework
testConfig := testutil.TestConfig{
    Timeout:        30,
    VerboseLogging: true,
    ArtifactsDir:   "/tmp/test-artifacts",
    CustomOptions:  map[string]interface{}{"option1": "value1"},
}
testFramework := testutil.NewTestFramework(testConfig, logger)

// Create a test that uses the security framework
testFunc := func(ctx context.Context) error {
    // Use the security framework to test authentication
    authScenario := securityScenarios.AuthenticationTestScenario()
    runner := testutil.NewScenarioRunner(testFramework)
    result := runner.RunScenario(authScenario)
    
    if result.Status != testutil.TestStatusPassed {
        return fmt.Errorf("authentication test failed")
    }

    // Use the security framework to test authorization
    authzScenario := securityScenarios.AuthorizationTestScenario()
    result = runner.RunScenario(authzScenario)
    
    if result.Status != testutil.TestStatusPassed {
        return fmt.Errorf("authorization test failed")
    }

    return nil
}

// Run the test
result := testFramework.RunTest("Security Test", testFunc)
if result.Status == testutil.TestStatusFailed {
    errorMsg := "unknown error"
    if len(result.Failures) > 0 {
        errorMsg = result.Failures[0].Message
    }
    logger.Error("Security test failed", "error", errorMsg)
} else {
    logger.Info("Security test passed")
}
```

## Testing with Network Simulation

The security testing framework can be used with network simulation to test how security mechanisms behave under different network conditions:

```go
// Get the network simulator
networkSimulator := env.GetNetworkSimulator()

// Create a test framework for running scenarios
testConfig := testutil.TestConfig{
    Timeout:        30,
    VerboseLogging: true,
    ArtifactsDir:   "/tmp/test-artifacts",
}
testFramework := testutil.NewTestFramework(testConfig, logger)

// Create a custom test scenario that tests security under different network conditions
securityUnderNetworkConditionsScenario := testutil.TestScenario{
    Name:        "Security Under Network Conditions",
    Description: "Tests security mechanisms under different network conditions",
    Tags:        []string{"security", "network"},
    Steps: []testutil.TestStep{
        {
            Name: "Test Security with High Latency",
            Action: func(ctx context.Context) error {
                // Set high latency
                if err := networkSimulator.SetConditions(500*time.Millisecond, 0, 1000); err != nil {
                    return err
                }

                // Run authentication test using the security scenarios
                authScenario := securityScenarios.AuthenticationTestScenario()
                runner := testutil.NewScenarioRunner(testFramework)
                result := runner.RunScenario(authScenario)
                
                logger.Info("Authentication under high latency", "status", result.Status)
                return nil
            },
        },
        {
            Name: "Test Security with Packet Loss",
            Action: func(ctx context.Context) error {
                // Set packet loss
                if err := networkSimulator.SetConditions(100*time.Millisecond, 0.2, 1000); err != nil {
                    return err
                }

                // Run authentication test using the security scenarios
                authScenario := securityScenarios.AuthenticationTestScenario()
                runner := testutil.NewScenarioRunner(testFramework)
                result := runner.RunScenario(authScenario)
                
                logger.Info("Authentication under packet loss", "status", result.Status)
                return nil
            },
        },
        {
            Name: "Test Security with Network Disconnection",
            Action: func(ctx context.Context) error {
                // Disconnect network
                if err := networkSimulator.Disconnect(); err != nil {
                    return err
                }

                // Run authentication test (should fail gracefully)
                authScenario := securityScenarios.AuthenticationTestScenario()
                runner := testutil.NewScenarioRunner(testFramework)
                result := runner.RunScenario(authScenario)
                
                // We expect a failure here, so we'll log it but not return an error
                logger.Info("Authentication under network disconnection", "status", result.Status)

                // Reconnect network
                if err := networkSimulator.Reconnect(); err != nil {
                    return err
                }

                return nil
            },
        },
    },
    Assertions: []testutil.TestAssertion{
        {
            Name: "Verify Security Resilience to Network Conditions",
            Condition: func(ctx context.Context) bool {
                // This is a placeholder. In a real scenario, you would check the test results
                // stored in a shared state or database.
                return true
            },
            Message: "Security mechanisms are not resilient to network conditions",
        },
    },
    Cleanup: []testutil.CleanupStep{
        {
            Name: "Reset Network Conditions",
            Action: func(ctx context.Context) error {
                // Reset network conditions to normal
                return networkSimulator.SetConditions(0, 0, 0)
            },
            AlwaysRun: true,
        },
    },
}

// Add the custom scenario to the environment
env.AddScenario(securityUnderNetworkConditionsScenario)

// Run the custom scenario
if err := env.RunScenario("Security Under Network Conditions"); err != nil {
    logger.Error("Security under network conditions scenario failed", "error", err)
} else {
    logger.Info("Security under network conditions scenario passed")
}
```

## Creating Custom Security Test Scenarios

You can create custom security test scenarios by defining a `TestScenario` with security-related steps, assertions, and cleanup actions:

```go
customSecurityScenario := testutil.TestScenario{
    Name:        "Custom Security Test",
    Description: "A custom security test scenario",
    Tags:        []string{"security", "custom"},
    Steps: []testutil.TestStep{
        {
            Name: "Test Custom Security Feature",
            Action: func(ctx context.Context) error {
                // Implement your custom security test here
                return nil
            },
        },
    },
    Assertions: []testutil.TestAssertion{
        {
            Name: "Verify Custom Security Feature",
            Condition: func(ctx context.Context) bool {
                // Implement your custom security assertion here
                return true
            },
            Message: "Custom security feature is not working correctly",
        },
    },
    Cleanup: []testutil.CleanupStep{
        {
            Name: "Cleanup Custom Security Test",
            Action: func(ctx context.Context) error {
                // Implement your custom cleanup here
                return nil
            },
            AlwaysRun: true,
        },
    },
}
```

## Best Practices

### 1. Security Scanning

- Regularly scan the codebase for vulnerabilities
- Include security scanning in the CI/CD pipeline
- Prioritize vulnerabilities based on severity
- Track remediation of vulnerabilities over time

### 2. Security Attack Simulation

- Simulate a variety of attack types
- Include both common and targeted attacks
- Test attack vectors specific to the application
- Regularly update attack simulations as new threats emerge

### 3. Security Control Verification

- Verify all security controls are properly implemented
- Test controls under various conditions
- Ensure controls are effective against real threats
- Regularly review and update security controls

### 4. Authentication Testing

- Test all authentication mechanisms
- Include both positive and negative test cases
- Test edge cases such as expired credentials
- Verify proper handling of authentication failures

### 5. Authorization Testing

- Test access control for all resources
- Verify proper role-based access control
- Test with different user roles and permissions
- Ensure proper handling of unauthorized access attempts

### 6. Data Protection Testing

- Test encryption of sensitive data
- Verify data integrity mechanisms
- Test key management procedures
- Ensure proper handling of cryptographic operations

## Conclusion

The security testing framework provides a comprehensive set of tools for testing the security aspects of the OneMount system. By integrating security testing into the overall testing strategy, we can ensure that the system meets its security requirements and is resilient against security threats.