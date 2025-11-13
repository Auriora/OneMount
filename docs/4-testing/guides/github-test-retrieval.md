# GitHub Test Results Retrieval Guide

This guide explains how to retrieve test results, coverage reports, and other artifacts from GitHub Actions workflows using OneMount's development CLI.

## Overview

OneMount's GitHub integration allows you to programmatically download test results and reports from GitHub Actions workflows. This is useful for:

- **Local Analysis**: Download test reports for offline analysis
- **CI/CD Integration**: Retrieve results from previous runs in pipelines
- **Debugging**: Access detailed logs and artifacts from failed runs
- **Reporting**: Generate custom reports from historical test data
- **Monitoring**: Track test trends over time

## Prerequisites

1. **GitHub Token**: Set up a GitHub personal access token
2. **Dependencies**: Ensure required Python packages are installed

### Setup GitHub Token

```bash
# Create a GitHub personal access token with 'repo' and 'actions:read' permissions
# Then set it as an environment variable
export GITHUB_TOKEN="your_token_here"

# Or pass it directly to commands
./scripts/dev github get-test-results --token "your_token_here"
```

### Install Dependencies

```bash
# Install CLI dependencies (includes requests for GitHub API)
pip install -r scripts/requirements-dev-cli.txt
```

## Usage

### Basic Commands

```bash
# Download latest test results from all workflows
./scripts/dev github get-test-results

# Download to specific directory
./scripts/dev github get-test-results --output ./my-test-results

# Download from specific workflow run
./scripts/dev github get-test-results --run-id 15522867228

# Filter by workflow name
./scripts/dev github get-test-results --workflow "Coverage"
```

### Available Artifacts

The command can download various types of test artifacts:

- **`test-results`**: JUnit XML and JSON from CI workflow
- **`coverage-reports`**: HTML, JSON, Cobertura XML coverage reports  
- **`system-test-results`**: System test JUnit XML and JSON
- **`system-test-logs`**: Detailed system test execution logs
- **`coverage-gap-analysis`**: Coverage gap analysis reports

### Output Structure

Downloaded artifacts are organized by type:

```
test-results-download/
├── test-results/
│   ├── junit.xml
│   └── go-test-report.json
├── coverage-reports/
│   ├── coverage.html
│   ├── coverage.json
│   ├── cobertura.xml
│   └── junit.xml
├── system-test-results/
│   ├── junit.xml
│   └── system-tests.json
└── system-test-logs/
    └── system_tests.log
```

## Advanced Usage

### Programmatic Access

Use the Python API directly for custom integrations:

```python
from scripts.utils.github_test_retriever import GitHubTestRetriever

# Initialize retriever
retriever = GitHubTestRetriever(token="your_token")

# Get recent workflow runs
runs = retriever.get_workflow_runs(workflow_name="CI", limit=10)

# Download artifacts from specific run
artifacts = retriever.get_run_artifacts(run_id=12345)
for artifact in artifacts:
    extract_dir = retriever.download_artifact(
        artifact['id'], 
        Path("./downloads") / artifact['name']
    )
    print(f"Downloaded: {artifact['name']} to {extract_dir}")
```

### GitHub CLI Integration

You can also use GitHub CLI for simpler access:

```bash
# List recent workflow runs
gh run list --limit 10

# Download artifacts from specific run
gh run download <run-id>

# View run details
gh run view <run-id>
```

### API Access

Direct GitHub API access for custom scripts:

```bash
# List workflow runs
curl -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/repos/Auriora/OneMount/actions/runs"

# Get artifacts for a run
curl -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/repos/Auriora/OneMount/actions/runs/{run_id}/artifacts"
```

## Use Cases

### 1. Local Test Analysis

```bash
# Download latest test results
./scripts/dev github get-test-results

# Open coverage report in browser
open test-results-download/coverage-reports/coverage.html

# Analyze JUnit XML in your IDE
code test-results-download/test-results/junit.xml
```

### 2. CI/CD Integration

```yaml
# In GitHub Actions workflow
- name: Download previous test results
  run: |
    ./scripts/dev github get-test-results --run-id ${{ github.run_id }}
    
- name: Compare test results
  run: |
    python scripts/compare_test_results.py \
      --current ./current-results \
      --previous ./test-results-download
```

### 3. Test Monitoring

```bash
# Download results from multiple recent runs
for run_id in $(gh run list --json databaseId --jq '.[].databaseId' | head -5); do
  ./scripts/dev github get-test-results --run-id $run_id --output "./history/$run_id"
done

# Analyze trends
python scripts/analyze_test_trends.py ./history/
```

## Troubleshooting

### Common Issues

**Authentication Error:**
```bash
# Ensure token has correct permissions
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"

# Test token access
curl -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/user"
```

**No Artifacts Found:**
```bash
# Check if workflows have completed
./scripts/dev github workflows

# Verify artifacts exist for specific run
gh run view <run-id>
```

**Missing Dependencies:**
```bash
# Install required packages
pip install requests rich typer

# Or install full dev requirements
pip install -r scripts/requirements-dev-cli.txt
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
./scripts/dev --verbose github get-test-results
```

## Best Practices

1. **Token Security**: Store GitHub tokens securely, never commit them
2. **Rate Limiting**: Be mindful of GitHub API rate limits
3. **Artifact Retention**: GitHub artifacts expire after 30-90 days
4. **Storage**: Clean up downloaded artifacts regularly
5. **Permissions**: Use tokens with minimal required permissions

## Integration Examples

### Makefile Integration

```makefile
download-test-results:
	./scripts/dev github get-test-results --output ./test-analysis

analyze-coverage:
	./scripts/dev github get-test-results --workflow "Coverage"
	open test-results-download/coverage-reports/coverage.html
```

### Shell Script Integration

```bash
#!/bin/bash
# download-latest-results.sh

set -e

echo "Downloading latest test results..."
./scripts/dev github get-test-results

echo "Generating summary report..."
python scripts/generate_test_summary.py test-results-download/

echo "Results available in: test-results-download/"
```

This integration provides a seamless way to access GitHub Actions test results directly from your OneMount development workflow!
