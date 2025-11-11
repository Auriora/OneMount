# Docker Test Environment Steering Rules Update

**Date**: 2025-11-11  
**Type**: Steering Rules Update  
**Component**: Testing Infrastructure  
**Priority**: High

---

## Summary

Updated `.kiro/steering/testing-conventions.md` to mandate Docker-based test execution for the OneMount project. This ensures all AI agents and developers use the validated Docker test environment consistently.

---

## Changes Made

### 1. Updated Testing Conventions Steering Rule

**File**: `.kiro/steering/testing-conventions.md`

**Key Changes**:
- Changed inclusion from `fileMatch` to `always` (applies to all contexts)
- Updated scope to include Go test files: `tests/**, internal/**/*_test.go`
- Added comprehensive Docker Test Environment section
- Documented all Docker test commands
- Added troubleshooting guide
- Included references to test setup documentation

### 2. New Docker Test Environment Section

Added mandatory Docker usage policy with:

#### Why Docker is Required
1. FUSE Dependencies - Tests require FUSE3 device access
2. Isolation - Prevents host system pollution
3. Reproducibility - Consistent environment across all systems
4. Security - Isolated test credentials
5. Dependencies - Specific Go 1.24.2 and Python 3.12 versions

#### Docker Test Commands
Documented all standard test execution commands:
- Unit tests: `docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests`
- Integration tests: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests`
- System tests: `docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests`
- All tests: `docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all`
- Interactive shell: `docker compose -f docker/compose/docker-compose.test.yml run --rm shell`

#### Test Environment Details
- Images: `onemount-base:latest` (1.49GB), `onemount-test-runner:latest` (2.21GB)
- Workspace: Mounted at `/workspace`
- Test artifacts: Output to `test-artifacts/`
- Auth tokens: `test-artifacts/.auth_tokens.json`
- FUSE device: `/dev/fuse` with SYS_ADMIN capability
- Resources: 4GB RAM / 2 CPUs (unit), 6GB RAM / 4 CPUs (system)

#### Docker Environment Modifications
Documented process for improving the Docker test environment:
1. Update Dockerfiles
2. Rebuild images
3. Update documentation
4. Test changes
5. Document rationale

#### Common Docker Test Patterns
Added examples for:
- Running tests with debugging
- Verbose output
- Coverage reports
- Environment validation

#### Troubleshooting Guide
Five-step troubleshooting process for Docker test failures:
1. Check FUSE device
2. Verify images are up-to-date
3. Check auth tokens
4. Review logs
5. Interactive debugging

#### References
Links to comprehensive documentation:
- `docs/TEST_SETUP.md` - Test environment setup
- `docs/testing/docker-test-environment.md` - Docker-specific details
- `docs/TASK_1_SUMMARY.md` - Validation results

---

## Rationale

### Problem
Without explicit steering rules, AI agents might:
- Run tests directly on host system (fails due to FUSE requirements)
- Use incorrect Go or Python versions
- Pollute host system with test artifacts
- Miss critical environment setup steps
- Inconsistent test execution across different contexts

### Solution
Mandatory Docker test environment ensures:
- ✅ All tests run in validated, isolated environment
- ✅ Consistent FUSE device access with proper capabilities
- ✅ Correct dependency versions (Go 1.24.2, Python 3.12)
- ✅ Test artifacts properly isolated and accessible
- ✅ Reproducible results across all developers and CI/CD
- ✅ Security isolation for test credentials

### Benefits
1. **Consistency**: Same environment for all developers and CI/CD
2. **Reliability**: Tests always have required dependencies
3. **Security**: Credentials isolated in containers
4. **Maintainability**: Single source of truth for test execution
5. **Documentation**: Clear commands and troubleshooting steps
6. **AI Agent Compliance**: Agents will always use correct test commands

---

## Impact

### For AI Agents
- Will always use Docker commands for test execution
- Will reference correct documentation when troubleshooting
- Will follow proper procedures for Docker environment modifications
- Will understand why Docker is required

### For Developers
- Clear, documented test execution commands
- Comprehensive troubleshooting guide
- Understanding of Docker environment architecture
- Confidence that tests run in validated environment

### For CI/CD
- Consistent test execution across all pipelines
- Same Docker images used locally and in CI
- Predictable resource requirements
- Isolated test credentials

---

## Verification

The steering rule update has been verified:

1. ✅ **Rule File Updated**: `.kiro/steering/testing-conventions.md`
2. ✅ **Inclusion Changed**: From `fileMatch` to `always` (applies globally)
3. ✅ **Priority Set**: 25 (appropriate for testing conventions)
4. ✅ **Scope Defined**: `tests/**, internal/**/*_test.go`
5. ✅ **Commands Documented**: All Docker test commands included
6. ✅ **References Added**: Links to comprehensive documentation
7. ✅ **Troubleshooting Included**: Five-step debugging process

---

## Related Documentation

### Primary References
- `.kiro/steering/testing-conventions.md` - Updated steering rule
- `docs/TEST_SETUP.md` - Comprehensive test setup guide
- `docs/testing/docker-test-environment.md` - Docker environment details
- `docs/TASK_1_SUMMARY.md` - Docker environment validation

### Supporting Documentation
- `docker/compose/docker-compose.test.yml` - Test execution configuration
- `docker/compose/docker-compose.build.yml` - Image build configuration
- `packaging/docker/Dockerfile.base` - Base image definition
- `packaging/docker/Dockerfile.test-runner` - Test runner image
- `packaging/docker/test-entrypoint.sh` - Test execution script

---

## Next Steps

1. **Verify AI Agent Compliance**: Test that agents use Docker commands
2. **Monitor Test Execution**: Ensure all tests run successfully in Docker
3. **Update CI/CD**: Align CI/CD pipelines with Docker test commands
4. **Developer Communication**: Inform team of mandatory Docker usage
5. **Documentation Review**: Ensure all test docs reference Docker commands

---

## Testing

To verify the steering rule is working:

```bash
# AI agents should now automatically use these commands:

# Run unit tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests

# Run integration tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests

# Run all tests
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner all
```

Expected behavior:
- AI agents will use Docker commands for all test execution
- AI agents will reference Docker documentation when troubleshooting
- AI agents will follow Docker environment modification procedures

---

## Conclusion

The testing conventions steering rule has been successfully updated to mandate Docker-based test execution for the OneMount project. This ensures:

- ✅ Consistent test environment across all contexts
- ✅ Proper FUSE device access and capabilities
- ✅ Correct dependency versions
- ✅ Isolated test credentials and artifacts
- ✅ Reproducible test results
- ✅ Clear documentation and troubleshooting

All AI agents will now automatically use the validated Docker test environment, improving reliability and consistency of test execution.

---

## Sign-off

**Change Type**: Steering Rules Update  
**Status**: ✅ COMPLETE  
**Date**: 2025-11-11  
**Impact**: High (affects all test execution)  
**Verification**: Steering rule active and documented  
**Approved For**: Immediate use

---

## Rules Consulted

- `operational-best-practices.md` - Tool-driven exploration and documentation consistency
- `documentation-conventions.md` - Documentation structure and placement
- `general-preferences.md` - Rule application guidelines and transparency

## Rules Applied

- `testing-conventions.md` - Updated with Docker test environment requirements
- `documentation-conventions.md` - Created update log in `docs/updates/`
- `operational-best-practices.md` - Documented rationale and verification steps
