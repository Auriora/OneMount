# Testing Guidelines Improvements

**Date**: 2026-01-21  
**Author**: AI Agent  
**Context**: Task 30.8 - Virtual File State Handling Test Implementation

## Issue

During implementation of task 30.8, the AI agent initially attempted to run tests directly on the host system instead of using Docker containers and the timeout wrapper script, despite clear guidelines in `testing-conventions.md`.

## Root Cause

While the guidelines are comprehensive, they may not be prominent enough for AI agents to consistently follow, especially when:
1. The critical requirements are buried in longer documentation
2. There are no explicit "DO NOT" examples showing wrong approaches
3. The consequences of not following the guidelines aren't immediately clear

## Recommendations

### 1. Add Critical Warning Banner

Add a prominent warning at the very top of `testing-conventions.md`:

```markdown
# Testing Conventions

**🚨 CRITICAL - READ THIS FIRST 🚨**

**ALL TESTS MUST BE RUN IN DOCKER CONTAINERS - NO EXCEPTIONS**

**NEVER run tests directly on the host system with `go test`**

**ALWAYS use the timeout wrapper script for integration tests**

If you are an AI agent and you see yourself about to run `go test` directly, STOP and use Docker instead.
```

### 2. Add AI Agent Checklist

Add a pre-flight checklist section:

```markdown
## AI Agent Test Execution Checklist

Before running ANY test, verify ALL of these:
- [ ] Am I using Docker? (docker compose -f ...)
- [ ] Am I using the timeout wrapper for integration tests? (./scripts/timeout-test-wrapper.sh)
- [ ] Am I in the correct working directory? (workspace root, not a subdirectory)
- [ ] Have I included the auth override for integration/system tests? (-f docker/compose/docker-compose.auth.yml)
- [ ] Am I NOT using the `cd` command? (it's forbidden)

If ANY checkbox is unchecked, you are doing it wrong!
```

### 3. Add Explicit Wrong vs Right Examples

Add a section showing common mistakes:

```markdown
## ❌ WRONG - DO NOT DO THIS:

```bash
# WRONG: Running tests directly on host
go test -v -run TestIT_FS ./internal/fs

# WRONG: Using cd command
cd OneMount && go test ...

# WRONG: Not using timeout wrapper for integration tests
docker compose ... run test-runner go test -run TestIT_FS_30_08 ./internal/fs
```

## ✅ CORRECT - DO THIS:

```bash
# CORRECT: Using timeout wrapper for integration tests
./scripts/timeout-test-wrapper.sh "TestIT_FS_30_08" 60

# CORRECT: Using Docker directly with auth override
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run TestIT_FS_30_08 ./internal/fs

# CORRECT: Unit tests in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```
```

### 4. Add Consequence Section

Add a section explaining why these rules exist:

```markdown
## Why These Rules Exist

### Docker Requirement
- **FUSE Dependencies**: Tests require FUSE3 device access with specific capabilities
- **Isolation**: Prevents test artifacts from polluting the host system
- **Reproducibility**: Ensures consistent environment across all developers and CI/CD
- **Security**: Test credentials and OneDrive access are isolated in containers

### Timeout Wrapper Requirement
- **Prevents Hanging**: Some FUSE filesystem tests may hang indefinitely
- **Resource Cleanup**: Ensures containers are properly cleaned up even on timeout
- **Debugging**: Provides detailed logs in test-artifacts/debug/

### No `cd` Command
- **Shell Context**: The `cd` command doesn't work as expected in tool execution
- **Working Directory**: Use the `cwd` parameter instead for bash commands
```

### 5. Update System Prompt

Consider adding to the system prompt or implicit rules:

```markdown
## Testing Protocol (MANDATORY)

When executing tests in the OneMount project:

1. **ALWAYS use Docker** - Never run `go test` directly on host
2. **ALWAYS use timeout wrapper** - For integration/system tests: `./scripts/timeout-test-wrapper.sh "TestPattern" 60`
3. **NEVER use `cd` command** - The workspace root is already correct
4. **ALWAYS include auth override** - For tests requiring authentication: `-f docker/compose/docker-compose.auth.yml`

Violation of these rules will result in test failures and environment corruption.
```

## Implementation Status

### Task 30.8 Results

✅ **Test Created**: `TestIT_FS_30_08_VirtualFile_StateHandling_CorrectlyManaged`
✅ **Test Passed**: All assertions passed (0.03s execution time)
✅ **Coverage**: 
- Virtual entries have item_state=HYDRATED ✓
- Virtual entries have remote_id=NULL and is_virtual=TRUE ✓
- Virtual entries bypass sync/upload logic ✓
- Virtual entries participate in directory listings ✓
- Virtual entries cannot transition to invalid states ✓

### Test Output Summary

```
=== RUN   TestIT_FS_30_08_VirtualFile_StateHandling_CorrectlyManaged
    ✓ Created virtual file: .xdg-volume-info (ID: local-.xdg-volume-info)
    ✓ Created virtual file: virtual-config.txt (ID: local-virtual-config.txt)
    ✓ Created virtual file: virtual-readme.md (ID: local-virtual-readme.md)
    ✓ All virtual files have state: HYDRATED
    ✓ All virtual files have RemoteID="" and Virtual=true
    ✓ All virtual files bypass sync/upload logic
    ✓ All virtual files found in directory listings
    ✓ All virtual files correctly prevent invalid state transitions
--- PASS: TestIT_FS_30_08_VirtualFile_StateHandling_CorrectlyManaged (0.03s)
```

## Next Steps

1. Update `testing-conventions.md` with the recommended improvements
2. Consider adding these checks to a pre-test validation script
3. Update AI agent system prompts to emphasize these requirements
4. Add automated checks in CI/CD to catch violations

## References

- Task: `.kiro/specs/system-verification-and-fix/tasks.md` - Task 30.8
- Requirement: `21.10` - Virtual file state handling
- Test File: `internal/fs/fs_integration_test.go` - TestIT_FS_30_08
- Guidelines: `.kiro/steering/testing-conventions.md`
