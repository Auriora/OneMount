# Coverage CI/CD Integration Guide

This guide explains how OneMount integrates test coverage analysis into its CI/CD pipeline to ensure code quality and track coverage trends over time.

## Overview

OneMount's coverage integration provides:

- **Automated Coverage Reporting** - Generate detailed coverage reports on every push and PR
- **Coverage Threshold Enforcement** - Fail builds that don't meet minimum coverage requirements
- **Trend Analysis** - Track coverage changes over time and detect regressions
- **Gap Analysis** - Identify files and packages that need more test coverage
- **Multiple Report Formats** - HTML, JSON, Cobertura XML, and JUnit XML reports

## CI/CD Workflow

### GitHub Actions Workflow

The coverage analysis runs automatically via GitHub Actions in `.github/workflows/coverage.yml`:

```yaml
name: Coverage Analysis and Reporting
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
```

### Workflow Steps

1. **Environment Setup** - Install Go, Python, and system dependencies
2. **Test Execution** - Run all tests with coverage profiling
3. **Report Generation** - Create multiple coverage report formats
4. **Threshold Checking** - Verify coverage meets minimum requirements
5. **Trend Analysis** - Update coverage history and analyze trends
6. **Artifact Upload** - Store reports for download and review
7. **PR Comments** - Add coverage summary to pull request comments

## Coverage Thresholds

OneMount enforces the following coverage thresholds:

| Metric | Threshold | Description |
|--------|-----------|-------------|
| Line Coverage | 80% | Minimum percentage of code lines executed |
| Function Coverage | 90% | Minimum percentage of functions called |
| Branch Coverage | 70% | Minimum percentage of code branches taken |

### Configuring Thresholds

Thresholds can be adjusted in the coverage scripts:

```bash
# In scripts/coverage-report.sh
THRESHOLD_LINE=80
THRESHOLD_FUNC=90
THRESHOLD_BRANCH=70

# Or via command line
./scripts/coverage-report.sh --threshold-line 85 --threshold-func 95
```

## Local Coverage Analysis

### Basic Coverage Report

Generate a basic coverage report locally:

```bash
# Run tests with coverage
make coverage

# Generate comprehensive report
make coverage-report

# Run CI-style analysis
make coverage-ci
```

### Advanced Analysis

For detailed trend analysis:

```bash
# Generate trend analysis (requires Python dependencies)
make coverage-trend

# Or run directly
python3 scripts/coverage-trend-analysis.py \
  --input coverage/coverage_history.json \
  --output coverage/trends.html \
  --plot
```

## Report Formats

### HTML Reports

- **Standard Report** (`coverage/coverage.html`) - Go's built-in HTML coverage report
- **Detailed Report** (`coverage/coverage-detailed.html`) - Enhanced report with package analysis
- **Trend Report** (`coverage/trends.html`) - Historical trend analysis with visualizations

### Machine-Readable Reports

- **JSON Report** (`coverage/coverage.json`) - Structured data for programmatic access
- **Cobertura XML** (`coverage/cobertura.xml`) - For CI/CD integration
- **JUnit XML** (`coverage/junit.xml`) - Test result format for CI systems

### Text Reports

- **Function Coverage** (`coverage/coverage-func.txt`) - Function-level coverage details
- **Package Analysis** (`coverage/package-analysis.txt`) - Package-level aggregated coverage
- **Coverage Gaps** (`coverage/coverage-gaps.txt`) - Files below threshold
- **Summary** (`coverage/summary.txt`) - High-level coverage summary

## Coverage History and Trends

### Historical Tracking

Coverage data is automatically tracked in `coverage/coverage_history.json`:

```json
[
  {
    "timestamp": 1703123456,
    "total_coverage": 82.5,
    "date": "2023-12-21T10:30:56Z"
  }
]
```

### Trend Analysis

The trend analysis tool provides:

- **Trend Direction** - Improving, declining, or stable coverage
- **Regression Detection** - Identify significant coverage drops
- **Visual Charts** - Coverage trends over time (requires matplotlib)
- **Change Analysis** - Per-commit coverage changes

## Integration with Development Workflow

### Pull Request Integration

Coverage analysis automatically:

1. **Comments on PRs** with coverage summary
2. **Uploads artifacts** with detailed reports
3. **Fails checks** if coverage drops below thresholds
4. **Highlights gaps** in coverage for review

### Branch Protection

Configure branch protection rules to require coverage checks:

```yaml
# In GitHub repository settings
required_status_checks:
  - "Coverage Analysis"
```

### IDE Integration

Coverage reports are compatible with:

- **JetBrains GoLand** - Automatic coverage.out detection
- **VS Code** - Go extension coverage support
- **Vim/Neovim** - Coverage highlighting plugins

## Troubleshooting

### Common Issues

**Coverage file not found:**
```bash
# Ensure tests run with coverage
go test -coverprofile=coverage.out ./...
```

**Python dependencies missing:**
```bash
# Install required packages
pip install matplotlib pandas numpy jinja2
```

**Threshold failures:**
```bash
# Check which files need more coverage
cat coverage/coverage-gaps.txt
```

### Debug Mode

Run coverage analysis in verbose mode:

```bash
# Enable debug output
bash -x scripts/coverage-report.sh --ci
```

## Best Practices

### Writing Coverage-Friendly Tests

1. **Test Public APIs** - Focus on exported functions and methods
2. **Cover Error Paths** - Test error conditions and edge cases
3. **Use Table Tests** - Efficiently test multiple scenarios
4. **Mock External Dependencies** - Isolate units under test

### Maintaining Coverage

1. **Monitor Trends** - Review daily coverage reports
2. **Address Gaps** - Prioritize low-coverage files
3. **Set Package Goals** - Define coverage targets per package
4. **Review PRs** - Ensure new code includes tests

### Performance Considerations

1. **Parallel Testing** - Use `go test -parallel` for faster execution
2. **Selective Coverage** - Focus on critical packages
3. **Cache Dependencies** - Use CI caching for faster builds
4. **Optimize Test Data** - Minimize test setup overhead

## Advanced Configuration

### Custom Coverage Goals

Define package-specific coverage goals:

```go
// In test files
var coverageGoals = map[string]CoverageGoal{
    "internal/fs": {
        LineCoverage:   90.0,
        FuncCoverage:   95.0,
        BranchCoverage: 85.0,
        Priority:       HighPriority,
    },
}
```

### Integration with External Tools

- **SonarQube** - Import Cobertura XML reports
- **Codecov** - Automatic upload via GitHub Actions
- **Coveralls** - Alternative coverage tracking service
- **Code Climate** - Code quality and coverage analysis

## Monitoring and Alerts

### Coverage Alerts

Set up alerts for coverage regressions:

```yaml
# GitHub Actions notification
- name: Notify on coverage drop
  if: failure()
  uses: actions/github-script@v6
  with:
    script: |
      github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: '⚠️ Coverage check failed! Please review the coverage report.'
      });
```

### Dashboard Integration

Integrate coverage metrics into project dashboards:

- **Grafana** - Visualize coverage trends
- **Prometheus** - Collect coverage metrics
- **Custom Dashboards** - Use JSON reports for custom visualizations

This comprehensive coverage integration ensures OneMount maintains high code quality while providing developers with the tools they need to write effective tests and monitor coverage trends.
