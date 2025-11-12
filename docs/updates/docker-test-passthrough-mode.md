# Docker Test Runner Pass-Through Mode

**Date**: 2025-11-12  
**Type**: Enhancement  
**Component**: Docker Test Infrastructure

## Problem

The Docker test runner entrypoint script only accepted specific helper commands (`unit`, `integration`, `system`, etc.), which caused confusion when trying to run custom test commands. Users and AI agents had to use workarounds like `--entrypoint /bin/bash` to run specific test patterns.

Example of what didn't work:
```bash
# This failed with "Unknown command: go"
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  integration-tests go test -v -run TestIT_FS_ETag ./internal/fs
```

## Solution

Modified `docker/scripts/test-entrypoint.sh` to support **pass-through mode**:

- **Helper commands** (unit, integration, system, all, coverage, shell, build, help) are caught and executed as before
- **Any other command** is passed through and executed directly after environment setup and binary building

This makes the test runner much more intuitive and flexible.

## Usage

### Helper Commands (unchanged)
```bash
# Run all integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
```

### Pass-Through Mode (new)
```bash
# IMPORTANT: Use test-runner service for pass-through, not specialized services

# Run specific test pattern
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -run TestIT_FS_ETag ./internal/fs

# Run with custom flags
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner go test -v -timeout 10m -race ./internal/...
```

## Benefits

1. **Intuitive**: Commands work as users expect
2. **Flexible**: Any go test command can be run without workarounds
3. **Consistent**: Environment setup and binary building happen automatically
4. **Backward Compatible**: All existing helper commands work exactly as before

## Files Modified

- `docker/scripts/test-entrypoint.sh` - Added pass-through mode logic
- `.kiro/steering/testing-conventions.md` - Updated documentation with examples

## Testing

The entrypoint script passes shell diagnostics and maintains backward compatibility with all existing helper commands while adding the new pass-through capability.

**Important**: After modifying the entrypoint script, the Docker image must be rebuilt:
```bash
docker compose -f docker/compose/docker-compose.test.yml build test-runner
```

Verified working with:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner go version
# Output: go version go1.24.2 linux/amd64
```

## Rules Consulted

- `testing-conventions.md` (Priority 25) - Docker test environment requirements
- `operational-best-practices.md` (Priority 40) - Tool usage and command patterns
- `general-preferences.md` (Priority 50) - Code quality and documentation standards

## Rules Applied

- Maintained backward compatibility (no breaking changes)
- Added comprehensive documentation
- Followed DRY principle (reused existing setup functions)
- Updated help text to reflect new capability
