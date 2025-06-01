# JetBrains Run Configurations for OneMount

This directory contains JetBrains IDE run configurations for running different types of tests in the OneMount project. These configurations are stored in the `.run/` folder, which is the modern standard for JetBrains IDEs and allows sharing run configurations across team members.

## Available Configurations

### Go Test Configurations

#### 1. Unit Tests (`Unit_Tests.xml`)
- **Purpose**: Run unit tests only (fast, isolated tests)
- **Command**: `go test -short ./...`
- **Equivalent Make target**: `make unit-test`
- **Description**: Runs all tests marked with the `-short` flag, focusing on unit tests that don't require external dependencies

#### 2. Integration Tests (`Integration_Tests.xml`)
- **Purpose**: Run integration tests that test component interactions
- **Command**: `go test -timeout 5m ./pkg/testutil/integration_test_env_test.go`
- **Equivalent Make target**: `make integration-test`
- **Description**: Tests the integration test environment and framework

#### 3. System Tests (`System_Tests.xml`)
- **Purpose**: Run basic system tests
- **Command**: `go test -timeout 5m ./pkg/testutil/system_test_env_test.go`
- **Equivalent Make target**: `make system-test`
- **Description**: Tests the system test environment and framework

#### 4. System Tests (Real OneDrive) (`System_Tests_Real_OneDrive.xml`)
- **Purpose**: Run comprehensive system tests with real OneDrive account
- **Command**: `go test -timeout 30m -run TestSystemST_.* ./tests/system`
- **Equivalent Make target**: `make system-test-go`
- **Description**: Runs all system tests that use a real OneDrive account for end-to-end testing

#### 5. System Tests (Performance) (`System_Tests_Performance.xml`)
- **Purpose**: Run performance-focused system tests
- **Command**: `go test -timeout 30m -run TestSystemST_PERFORMANCE_.* ./tests/system`
- **Equivalent Make target**: `make system-test-performance`
- **Description**: Runs system tests focused on performance measurements

#### 6. System Tests (Reliability) (`System_Tests_Reliability.xml`)
- **Purpose**: Run reliability-focused system tests
- **Command**: `go test -timeout 30m -run TestSystemST_RELIABILITY_.* ./tests/system`
- **Equivalent Make target**: `make system-test-reliability`
- **Description**: Runs system tests focused on reliability and error recovery

#### 7. All Tests (`All_Tests.xml`)
- **Purpose**: Run all tests in the project
- **Command**: `go test ./...`
- **Equivalent Make target**: `make test`
- **Description**: Runs all tests including unit, integration, and system tests

#### 8. Tests with Coverage (`Tests_with_Coverage.xml`)
- **Purpose**: Run all tests and generate coverage report
- **Command**: `go test -coverprofile=coverage/coverage.out ./...`
- **Equivalent Make target**: `make coverage`
- **Description**: Runs all tests and generates a coverage profile

### Make Configurations

#### 9. Make: System Tests (Real) (`Make_System_Tests_Real.xml`)
- **Purpose**: Run comprehensive system tests using the Make script
- **Command**: `make system-test-real`
- **Description**: Uses the shell script for more comprehensive output and logging

#### 10. Make: System Tests (All Categories) (`Make_System_Tests_All.xml`)
- **Purpose**: Run all system test categories
- **Command**: `make system-test-all`
- **Description**: Runs comprehensive, performance, reliability, integration, and stress tests

#### 11. Make: Coverage Report (`Make_Coverage_Report.xml`)
- **Purpose**: Generate comprehensive coverage report with analysis
- **Command**: `make coverage-report`
- **Description**: Generates detailed coverage reports with gap analysis and trends

## Usage Instructions

### In JetBrains IDEs (GoLand, IntelliJ IDEA with Go plugin)

1. **Access Run Configurations**:
   - Go to `Run` → `Edit Configurations...`
   - Or click the run configuration dropdown in the toolbar

2. **Select a Configuration**:
   - Choose from the available configurations listed above
   - Each configuration is pre-configured with appropriate parameters

3. **Run Tests**:
   - Click the green play button, or
   - Use `Ctrl+Shift+F10` (Windows/Linux) or `Cmd+Shift+R` (Mac)
   - Or right-click on the configuration and select "Run"

4. **Debug Tests**:
   - Click the debug button (bug icon), or
   - Use `Ctrl+Shift+F9` (Windows/Linux) or `Cmd+Shift+D` (Mac)

### Recommended Workflow

1. **During Development**: Use `Unit Tests` for quick feedback
2. **Before Committing**: Run `All Tests` to ensure nothing is broken
3. **Integration Testing**: Use `Integration Tests` when working on component interactions
4. **End-to-End Testing**: Use `System Tests (Real OneDrive)` for comprehensive validation
5. **Performance Analysis**: Use `System Tests (Performance)` when optimizing
6. **Coverage Analysis**: Use `Tests with Coverage` or `Make: Coverage Report`

## Prerequisites

### For System Tests with Real OneDrive
- OneDrive authentication tokens in `~/.onemount-tests/.auth_tokens.json`
- Network connectivity to OneDrive
- Sufficient storage space on OneDrive account

### For All Tests
- Go 1.19+ installed
- Required system dependencies (see main README.md)
- For some tests: `sudo` access (for network simulation)

## Troubleshooting

### Common Issues

1. **Authentication Errors**: Ensure OneDrive tokens are valid and accessible
2. **Timeout Errors**: Increase timeout values in configuration if needed
3. **Permission Errors**: Some tests may require elevated privileges
4. **Network Errors**: Check internet connectivity for system tests

### Logs and Output

- Test output appears in the IDE's run window
- System test logs are written to `~/.onemount-tests/logs/`
- Coverage reports are generated in `./coverage/`

## Customization

You can modify these configurations by:
1. Opening `Run` → `Edit Configurations...`
2. Selecting the configuration to modify
3. Adjusting parameters, timeouts, or test patterns as needed
4. Saving the changes

The configurations are stored as XML files in the `.run/` directory and can be version controlled with your project. This location is the modern standard for JetBrains IDEs and allows team members to share run configurations.
