# Docker Test Environment

This directory contains documentation for the Docker-based test environment.

## Overview

All OneMount tests MUST be run inside Docker containers to ensure:
- FUSE3 device access with proper capabilities
- Isolation from host system
- Reproducible environment across developers and CI/CD
- Secure handling of test credentials

## Key Documentation

### Setup & Configuration

- [docker-test-environment.md](docker-test-environment.md) - Complete Docker test environment guide
- [QUICK-AUTH-SETUP.md](QUICK-AUTH-SETUP.md) - Quick authentication setup
- [DOCKER-AUTH-INTEGRATION.md](DOCKER-AUTH-INTEGRATION.md) - Authentication integration details
- [persistent-authentication-setup.md](persistent-authentication-setup.md) - Persistent auth configuration

### Test Execution

- [system-tests-guide.md](system-tests-guide.md) - Running system tests
- [end-to-end-tests.md](end-to-end-tests.md) - End-to-end testing
- [large-file-system-tests.md](large-file-system-tests.md) - Large file testing
- [test-log-redirection.md](test-log-redirection.md) - Log management

### CI/CD Integration

- [ci-system-tests-setup.md](ci-system-tests-setup.md) - CI system test setup
- [personal-onedrive-ci-setup.md](personal-onedrive-ci-setup.md) - OneDrive CI configuration

### Implementation Notes

- [IMPLEMENTATION_NOTES.md](IMPLEMENTATION_NOTES.md) - Technical implementation details

## Quick Start

```bash
# Build test images
./docker/scripts/build-images.sh test-runner

# Run unit tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Run integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# Run system tests (requires auth tokens)
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests

# Interactive shell for debugging
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

## Test Services

- **unit-tests** - Fast unit tests (no FUSE required)
- **integration-tests** - Integration tests (requires FUSE)
- **system-tests** - System tests (requires OneDrive auth)
- **test-runner** - Generic test runner for custom commands
- **shell** - Interactive debugging shell

## Authentication

Test authentication tokens are stored in `test-artifacts/.auth_tokens.json` (gitignored).

See [QUICK-AUTH-SETUP.md](QUICK-AUTH-SETUP.md) for setup instructions.

## Related Documentation

- [Testing Conventions](../../guides/ai-agent/AGENT-RULE-Testing-Conventions.md) - Complete Docker testing rules
- [Test Setup](../TEST_SETUP.md) - General test setup
- [Test Guidelines](../guides/test-guidelines.md) - Testing best practices
