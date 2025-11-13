# Testing Documentation

This directory contains all testing-related documentation for OneMount.

## Quick Start

- [TEST_SETUP.md](TEST_SETUP.md) - Initial test environment setup
- [Getting Started Guide](guides/getting-started.md) - Introduction to the test framework
- [Test Guidelines](guides/test-guidelines.md) - Best practices for writing tests

## Test Execution

### Docker-Based Testing

All OneMount tests MUST be run inside Docker containers. See [Docker Test Environment](docker/docker-test-environment.md) for details.

**Quick Commands:**
```bash
# Unit tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# System tests
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# All tests with coverage
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all
```

See [Testing Conventions](../guides/ai-agent/AGENT-RULE-Testing-Conventions.md) for complete Docker testing guidelines.

## Documentation Structure

### [docker/](docker/)
Docker-based test environment documentation:
- Docker test environment setup
- Authentication configuration
- System tests guide
- CI/CD integration

### [guides/](guides/)
Test framework guides and best practices:
- Testing framework overview
- Unit, integration, and performance testing guides
- Test guidelines and troubleshooting
- Coverage and CI integration

### [training/](training/)
Training materials for the test framework:
- Getting started tutorials
- Step-by-step examples
- Practice exercises
- Advanced topics

## Test Types

- **Unit Tests** - Fast, isolated component tests (no FUSE required)
- **Integration Tests** - Component interaction tests (requires FUSE)
- **System Tests** - End-to-end tests with real OneDrive (requires auth tokens)

## Test Plans & Checklists

- [test-plan.md](test-plan.md) - Comprehensive test plan
- [RETEST_CHECKLIST.md](RETEST_CHECKLIST.md) - Regression testing checklist
- [test-cases-traceability-matrix.md](test-cases-traceability-matrix.md) - Requirements traceability

## Manual Testing

- [manual-crash-recovery-testing.md](manual-crash-recovery-testing.md)
- [manual-network-error-testing.md](manual-network-error-testing.md)
- [manual-rate-limit-testing.md](manual-rate-limit-testing.md)

## Test Results & Status

- [test-results-summary.md](test-results-summary.md) - Latest test results
- [tests_that_should_be_passing.md](tests_that_should_be_passing.md) - Expected passing tests

## Related Documentation

- [Test Architecture Design](../2-architecture/test-architecture-design.md) - Test system architecture
- [Developer Guides](../guides/developer/) - Development guidelines
- [Implementation Details](../3-implementation/) - Code implementation
