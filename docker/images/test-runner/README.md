# Test Runner Image

Docker image for running OneMount tests.

**Base Image**: `onemount-builder` (from `packaging/docker/Dockerfile.builder`)

## Building

```bash
# Build test runner image
./docker/scripts/build-images.sh test-runner

# Build with no cache
./docker/scripts/build-images.sh test-runner --no-cache
```

## Usage

### Run Unit Tests

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

### Run Integration Tests

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

### Run System Tests

```bash
# Requires auth tokens in test-artifacts/.auth_tokens.json
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests
```

### Interactive Shell

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

## Features

- Go 1.24.2
- Python 3.12 with pytest
- FUSE3 support for filesystem testing
- Development tools (vim, less)
- Pre-built OneMount binaries for faster test execution

## Quick Reference

```bash
# Build image
./docker/scripts/build-images.sh test-runner

# Run unit tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Run all tests
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all

# Interactive shell
docker compose -f docker/compose/docker-compose.test.yml run --rm shell
```

## See Also

- Test compose file: `docker/compose/docker-compose.test.yml`
- Test entrypoint: `docker/scripts/test-entrypoint.sh`
- Test documentation: `docs/TEST_SETUP.md`
- Build script: `docker/scripts/build-images.sh`
