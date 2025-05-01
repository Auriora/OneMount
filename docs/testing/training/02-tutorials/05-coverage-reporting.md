# Tutorial: Test Coverage Reporting

This tutorial will guide you through the process of using test coverage reporting in the OneMount test framework. Test coverage reporting helps you identify areas of your code that need more testing and track your progress toward coverage goals.

> **Note**: All code examples in this tutorial are for illustration purposes only and may need to be adapted to your specific project structure and imports. The examples are not meant to be compiled directly but rather to demonstrate concepts and patterns.

## Table of Contents

1. [Introduction to Test Coverage](#introduction-to-test-coverage)
2. [The CoverageReporter Component](#the-coveragereporter-component)
3. [Running Tests with Coverage](#running-tests-with-coverage)
4. [Analyzing Coverage Reports](#analyzing-coverage-reports)
5. [Setting Coverage Goals](#setting-coverage-goals)
6. [Tracking Coverage Trends](#tracking-coverage-trends)
7. [Best Practices](#best-practices)
8. [Complete Example](#complete-example)

## Introduction to Test Coverage

Test coverage is a measure of how much of your code is executed during your tests. It helps you:

- Identify areas of your code that need more testing
- Track your progress toward coverage goals
- Identify dead code that is never executed
- Ensure that critical code paths are tested
- Provide metrics for code quality

The OneMount test framework provides a CoverageReporter component that helps you collect, analyze, and report test coverage metrics.

## The CoverageReporter Component

The CoverageReporter component provides tools for collecting, analyzing, and reporting test coverage metrics. It allows you to:

- Collect coverage data from Go's built-in coverage tools
- Report coverage metrics (line, function, branch coverage)
- Check coverage against thresholds
- Track coverage trends over time
- Generate HTML reports for visualization

You can create a CoverageReporter as follows:

```go
// Create a logger
logger := log.With().Str("component", "coverage").Logger()

// Create a coverage configuration
config := testutil.CoverageConfig{
    OutputDir:    "./coverage",
    HistoryFile:  "./coverage/history.json",
    HTMLReport:   true,
    Thresholds: testutil.CoverageThresholds{
        LineCoverage:   80.0, // 80% line coverage
        FuncCoverage:   90.0, // 90% function coverage
        BranchCoverage: 70.0, // 70% branch coverage
    },
}

// Create a new CoverageReporter
reporter := testutil.NewCoverageReporter(config, &logger)
```

## Running Tests with Coverage

To run tests with coverage, you need to:

1. Run your tests with the `-coverprofile` flag
2. Process the coverage profile with the CoverageReporter

Here's an example:

```go
// Run tests with coverage
cmd := exec.Command("go", "test", "./...", "-coverprofile=coverage.out")
output, err := cmd.CombinedOutput()
if err != nil {
    // Handle error
    fmt.Printf("Error running tests: %v\n", err)
    fmt.Println(string(output))
    return
}

// Create a coverage reporter
reporter := testutil.NewCoverageReporter(config, &logger)

// Process the coverage profile
err = reporter.ProcessCoverageProfile("coverage.out")
if err != nil {
    // Handle error
    fmt.Printf("Error processing coverage profile: %v\n", err)
    return
}

// Get coverage metrics
metrics, err := reporter.GetCoverageMetrics()
if err != nil {
    // Handle error
    fmt.Printf("Error getting coverage metrics: %v\n", err)
    return
}

// Print coverage metrics
fmt.Printf("Line Coverage: %.2f%%\n", metrics.LineCoverage)
fmt.Printf("Function Coverage: %.2f%%\n", metrics.FuncCoverage)
fmt.Printf("Branch Coverage: %.2f%%\n", metrics.BranchCoverage)

// Check if coverage meets thresholds
if !reporter.CheckThresholds() {
    fmt.Println("Coverage does not meet thresholds")
    // Handle threshold failure
}

// Generate HTML report
err = reporter.GenerateHTMLReport()
if err != nil {
    // Handle error
    fmt.Printf("Error generating HTML report: %v\n", err)
    return
}
```

## Analyzing Coverage Reports

The CoverageReporter generates HTML reports that help you visualize your coverage data. These reports show:

1. Overall coverage metrics
2. Coverage by package
3. Coverage by file
4. Line-by-line coverage highlighting

To analyze these reports:

1. Open the HTML report in a web browser
2. Look for files with low coverage
3. Identify uncovered lines and branches
4. Focus on critical code paths that need more testing

You can also analyze coverage data programmatically:

```go
// Get coverage metrics by package
packageMetrics, err := reporter.GetPackageMetrics()
if err != nil {
    // Handle error
    return
}

// Find packages with low coverage
for pkg, metrics := range packageMetrics {
    if metrics.LineCoverage < 70.0 {
        fmt.Printf("Package %s has low line coverage: %.2f%%\n", pkg, metrics.LineCoverage)
    }
}

// Get coverage metrics for a specific package
metrics, err := reporter.GetPackageMetrics("github.com/yourusername/onemount/internal/fs")
if err != nil {
    // Handle error
    return
}
fmt.Printf("Line Coverage for fs package: %.2f%%\n", metrics.LineCoverage)

// Get uncovered functions
uncoveredFuncs, err := reporter.GetUncoveredFunctions()
if err != nil {
    // Handle error
    return
}
fmt.Println("Uncovered functions:")
for _, fn := range uncoveredFuncs {
    fmt.Printf("- %s\n", fn)
}
```

## Setting Coverage Goals

The CoverageReporter allows you to set coverage goals for your project and track your progress toward them. You can set:

1. Overall coverage thresholds
2. Package-specific coverage goals
3. Incremental coverage goals

Here's an example of setting package-specific coverage goals:

```go
// Create coverage goals
goals := map[string]testutil.CoverageGoal{
    "github.com/yourusername/onemount/internal/fs": {
        LineCoverage:   90.0, // 90% line coverage
        FuncCoverage:   95.0, // 95% function coverage
        BranchCoverage: 80.0, // 80% branch coverage
        Priority:       testutil.HighPriority,
    },
    "github.com/yourusername/onemount/internal/ui": {
        LineCoverage:   70.0, // 70% line coverage
        FuncCoverage:   80.0, // 80% function coverage
        BranchCoverage: 60.0, // 60% branch coverage
        Priority:       testutil.MediumPriority,
    },
}

// Set coverage goals
reporter.SetCoverageGoals(goals)

// Check if coverage meets goals
if !reporter.CheckGoals() {
    fmt.Println("Coverage does not meet goals")
    // Get unmet goals
    unmetGoals, err := reporter.GetUnmetGoals()
    if err != nil {
        // Handle error
        return
    }
    fmt.Println("Unmet goals:")
    for pkg, goal := range unmetGoals {
        fmt.Printf("- %s: Line Coverage %.2f%% (goal: %.2f%%)\n", pkg, goal.CurrentLineCoverage, goal.LineCoverage)
    }
}
```

## Tracking Coverage Trends

The CoverageReporter can track coverage trends over time, helping you identify regressions and track progress. It stores historical coverage data in a JSON file and provides methods for analyzing trends:

```go
// Save current coverage data to history
err = reporter.SaveToHistory()
if err != nil {
    // Handle error
    return
}

// Get coverage trend
trend, err := reporter.GetCoverageTrend(30) // Last 30 days
if err != nil {
    // Handle error
    return
}

// Analyze trend
fmt.Printf("Line Coverage Trend: %.2f%%\n", trend.LineCoverageTrend)
fmt.Printf("Function Coverage Trend: %.2f%%\n", trend.FuncCoverageTrend)
fmt.Printf("Branch Coverage Trend: %.2f%%\n", trend.BranchCoverageTrend)

// Check for regressions
if trend.LineCoverageTrend < 0 {
    fmt.Println("Line coverage is decreasing")
}

// Get coverage history
history, err := reporter.GetCoverageHistory()
if err != nil {
    // Handle error
    return
}

// Print coverage history
for date, metrics := range history {
    fmt.Printf("%s: Line Coverage: %.2f%%\n", date, metrics.LineCoverage)
}
```

## Best Practices

When using test coverage reporting, follow these best practices:

1. **Set realistic coverage goals**: Not all code needs the same level of coverage. Set higher goals for critical code and lower goals for less critical code.

2. **Focus on critical code paths**: Prioritize testing critical code paths, such as error handling, security-related code, and core business logic.

3. **Don't chase 100% coverage**: Achieving 100% coverage is often impractical and may not be worth the effort. Focus on meaningful tests rather than just increasing coverage.

4. **Use coverage as a guide, not a goal**: Coverage is a tool to help you identify areas that need more testing, not an end in itself.

5. **Track coverage trends**: Track coverage trends over time to identify regressions and ensure that coverage is improving.

6. **Include coverage in CI/CD**: Include coverage reporting in your CI/CD pipeline to catch coverage regressions early.

7. **Review uncovered code**: Regularly review uncovered code to determine if it needs testing or if it's dead code that can be removed.

8. **Test edge cases**: Focus on testing edge cases and error conditions, not just the happy path.

9. **Balance coverage with test quality**: High coverage doesn't necessarily mean good tests. Focus on writing meaningful tests that verify the correct behavior of your code.

10. **Document coverage requirements**: Document your coverage requirements and ensure that your tests verify them.

## Complete Example

Here's a complete example of using the CoverageReporter in a test suite:

```go
package main

import (
    "fmt"
    "os"
    "os/exec"
    "testing"
    "time"

    "github.com/rs/zerolog/log"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestMain(m *testing.M) {
    // Create a logger
    logger := log.With().Str("component", "coverage").Logger()

    // Create a coverage configuration
    config := testutil.CoverageConfig{
        OutputDir:    "./coverage",
        HistoryFile:  "./coverage/history.json",
        HTMLReport:   true,
        Thresholds: testutil.CoverageThresholds{
            LineCoverage:   80.0, // 80% line coverage
            FuncCoverage:   90.0, // 90% function coverage
            BranchCoverage: 70.0, // 70% branch coverage
        },
    }

    // Create output directory if it doesn't exist
    if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
        fmt.Printf("Error creating output directory: %v\n", err)
        os.Exit(1)
    }

    // Run tests and exit if they fail
    result := m.Run()
    if result != 0 {
        os.Exit(result)
    }

    // Create a coverage reporter
    reporter := testutil.NewCoverageReporter(config, &logger)

    // Process the coverage profile
    err := reporter.ProcessCoverageProfile("coverage.out")
    if err != nil {
        fmt.Printf("Error processing coverage profile: %v\n", err)
        os.Exit(1)
    }

    // Get coverage metrics
    metrics, err := reporter.GetCoverageMetrics()
    if err != nil {
        fmt.Printf("Error getting coverage metrics: %v\n", err)
        os.Exit(1)
    }

    // Print coverage metrics
    fmt.Printf("Line Coverage: %.2f%%\n", metrics.LineCoverage)
    fmt.Printf("Function Coverage: %.2f%%\n", metrics.FuncCoverage)
    fmt.Printf("Branch Coverage: %.2f%%\n", metrics.BranchCoverage)

    // Check if coverage meets thresholds
    if !reporter.CheckThresholds() {
        fmt.Println("Coverage does not meet thresholds")
        os.Exit(1)
    }

    // Generate HTML report
    err = reporter.GenerateHTMLReport()
    if err != nil {
        fmt.Printf("Error generating HTML report: %v\n", err)
        os.Exit(1)
    }

    // Save current coverage data to history
    err = reporter.SaveToHistory()
    if err != nil {
        fmt.Printf("Error saving to history: %v\n", err)
        os.Exit(1)
    }

    // Get coverage trend
    trend, err := reporter.GetCoverageTrend(30) // Last 30 days
    if err != nil {
        fmt.Printf("Error getting coverage trend: %v\n", err)
        os.Exit(1)
    }

    // Print trend
    fmt.Printf("Line Coverage Trend: %.2f%%\n", trend.LineCoverageTrend)
    fmt.Printf("Function Coverage Trend: %.2f%%\n", trend.FuncCoverageTrend)
    fmt.Printf("Branch Coverage Trend: %.2f%%\n", trend.BranchCoverageTrend)

    // Exit with success
    os.Exit(0)
}

func TestCoverageWithGoTest(t *testing.T) {
    // Skip this test when running with -cover flag
    if testing.CoverMode() != "" {
        t.Skip("Skipping when running with coverage")
    }

    // Run tests with coverage
    cmd := exec.Command("go", "test", "./...", "-coverprofile=coverage.out")
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("Error running tests: %v\n%s", err, output)
    }

    // Verify that the coverage file was created
    _, err = os.Stat("coverage.out")
    if os.IsNotExist(err) {
        t.Fatal("Coverage file was not created")
    }
}

func TestPackageCoverage(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "coverage").Logger()

    // Create a coverage configuration
    config := testutil.CoverageConfig{
        OutputDir:   "./coverage",
        HistoryFile: "./coverage/history.json",
        HTMLReport:  true,
    }

    // Create a coverage reporter
    reporter := testutil.NewCoverageReporter(config, &logger)

    // Process the coverage profile
    err := reporter.ProcessCoverageProfile("coverage.out")
    if err != nil {
        t.Fatalf("Error processing coverage profile: %v", err)
    }

    // Define packages to check
    packagesToCheck := []string{
        "github.com/yourusername/onemount/internal/fs",
        "github.com/yourusername/onemount/internal/ui",
        "github.com/yourusername/onemount/internal/testutil",
    }

    // Check coverage for each package
    for _, pkg := range packagesToCheck {
        metrics, err := reporter.GetPackageMetrics(pkg)
        if err != nil {
            t.Fatalf("Error getting metrics for package %s: %v", pkg, err)
        }

        t.Logf("Package %s:", pkg)
        t.Logf("  Line Coverage: %.2f%%", metrics.LineCoverage)
        t.Logf("  Function Coverage: %.2f%%", metrics.FuncCoverage)
        t.Logf("  Branch Coverage: %.2f%%", metrics.BranchCoverage)

        // Verify minimum coverage for critical packages
        if pkg == "github.com/yourusername/onemount/internal/fs" {
            if metrics.LineCoverage < 80.0 {
                t.Errorf("Line coverage for %s is below 80%%: %.2f%%", pkg, metrics.LineCoverage)
            }
            if metrics.FuncCoverage < 90.0 {
                t.Errorf("Function coverage for %s is below 90%%: %.2f%%", pkg, metrics.FuncCoverage)
            }
        }
    }
}
```

This example demonstrates:
1. Setting up a coverage reporter
2. Running tests with coverage
3. Processing coverage data
4. Generating coverage reports
5. Checking coverage against thresholds
6. Tracking coverage trends
7. Analyzing package-specific coverage

By following these patterns, you can effectively use test coverage reporting to improve the quality of your tests and ensure that your code is well-tested.