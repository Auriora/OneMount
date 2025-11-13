# Docker Authentication Integration - Complete

✅ **STATUS: INTEGRATED**

## What Was Done

Docker Compose and test scripts have been updated to support persistent authentication tokens in Docker containers.

## Changes Made

### 1. Docker Compose Configuration

**File**: `docker/compose/docker-compose.test.yml`

Added authentication token volume mounts to all test services:

```yaml
volumes:
  - ${HOME}/.onemount-tests:/tmp/home-tester/.onemount-tests-auth:ro
```

**Services Updated**:
- ✅ `test-runner` (base service)
- ✅ `system-tests`
- ✅ `coverage`
- ✅ `shell` (inherits from test-runner)
- ✅ `unit-tests` (inherits from test-runner)
- ✅ `integration-tests` (inherits from test-runner)

### 2. Authentication Helper Library

**File**: `scripts/lib/auth-helper.sh`

Created reusable authentication helper functions:
- `setup_auth_tokens()` - Finds and copies tokens to expected location
- `is_docker()` - Detects if running in Docker
- `get_auth_info()` - Displays token information

**Token Search Order**:
1. `$HOME/.onemount-tests/.auth_tokens.json` (standard)
2. `/tmp/home-tester/.onemount-tests-auth/.auth_tokens.json` (Docker mounted)
3. `./test-artifacts/.auth_tokens.json` (project)
4. `./auth_tokens.json` (workspace root)

## How to Use

### Step 1: Setup Authentication (One Time)

```bash
# Run on your host machine
./scripts/setup-test-auth.sh
```

This creates `~/.onemount-tests/.auth_tokens.json` on your host.

### Step 2: Run Docker Tests

The tokens are automatically mounted and available in Docker:

```bash
# Using docker compose (recommended)
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh

# Or specific test services
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests
docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests
docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests
```

### Step 3: Verify Token Mounting

```bash
# Check if tokens are accessible in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  ls -la /tmp/home-tester/.onemount-tests-auth/

# Should show .auth_tokens.json with 600 permissions
```

## Token Flow in Docker

```
Host Machine                          Docker Container
─────────────                         ────────────────
~/.onemount-tests/                    /tmp/home-tester/.onemount-tests-auth/
  └── .auth_tokens.json  ──mount──>    └── .auth_tokens.json (read-only)
                                              │
                                              │ (copied by helper)
                                              ↓
                                        $HOME/.onemount-tests/
                                          └── .auth_tokens.json (working copy)
```

## Environment Variables

The Docker Compose file uses `${HOME}` to find your home directory:

```yaml
- ${HOME}/.onemount-tests:/tmp/home-tester/.onemount-tests-auth:ro
```

This works automatically on Linux/macOS. For custom paths:

```bash
# Set custom home directory
export HOME=/custom/path
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner ...
```

## Manual Docker Run

If not using Docker Compose:

```bash
docker run --rm -t \
  --user root \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -v "$(pwd):/workspace:rw" \
  -v "$HOME/.onemount-tests:/tmp/home-tester/.onemount-tests-auth:ro" \
  onemount-test-runner:latest \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

## Updating Existing Test Scripts

To use the auth helper in your test scripts:

```bash
#!/bin/bash
set -e

# Source the auth helper
source "$(dirname "$0")/lib/auth-helper.sh"

# Setup authentication
if ! setup_auth_tokens; then
    echo "Failed to setup authentication"
    exit 1
fi

# Your test code here
# Tokens are now available at $HOME/.onemount-tests/.auth_tokens.json
```

## Troubleshooting

### Problem: "No auth tokens found"

**Check 1**: Verify tokens exist on host
```bash
ls -la ~/.onemount-tests/.auth_tokens.json
```

**Check 2**: Verify Docker mount
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  ls -la /tmp/home-tester/.onemount-tests-auth/
```

**Solution**: Run setup script
```bash
./scripts/setup-test-auth.sh
```

### Problem: "Permission denied"

**Check**: File permissions
```bash
ls -la ~/.onemount-tests/.auth_tokens.json
# Should be: -rw------- (600)
```

**Solution**: Fix permissions
```bash
chmod 600 ~/.onemount-tests/.auth_tokens.json
```

### Problem: "Tokens expired"

**Check**: Expiration
```bash
jq '.expires_at' ~/.onemount-tests/.auth_tokens.json
date +%s
```

**Solution**: Refresh tokens
```bash
./scripts/setup-test-auth.sh --refresh
```

### Problem: Docker can't find $HOME

**Solution**: Set explicitly
```bash
export HOME=/home/yourusername
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner ...
```

### Problem: Volume mount not working

**Check**: Docker Compose version
```bash
docker compose version
# Should be v2.x or higher
```

**Solution**: Update Docker Compose or use absolute path
```bash
# Edit docker-compose.test.yml
volumes:
  - /home/yourusername/.onemount-tests:/tmp/home-tester/.onemount-tests-auth:ro
```

## Testing the Integration

### Test 1: Verify Token Mounting

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  bash -c "ls -la /tmp/home-tester/.onemount-tests-auth/ && cat /tmp/home-tester/.onemount-tests-auth/.auth_tokens.json | jq '.account'"
```

Expected output: Your Microsoft account email

### Test 2: Run Filesystem Operations Test

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

Expected: No login prompt, test runs successfully

### Test 3: Run System Tests

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests
```

Expected: All system tests run without authentication prompts

## CI/CD Integration

For GitHub Actions or other CI/CD:

```yaml
- name: Setup auth tokens
  run: |
    mkdir -p ~/.onemount-tests
    echo "${{ secrets.ONEMOUNT_AUTH_TOKENS }}" | base64 -d > ~/.onemount-tests/.auth_tokens.json
    chmod 600 ~/.onemount-tests/.auth_tokens.json

- name: Run Docker tests
  run: |
    docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
      /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

## Security Notes

- ✅ Tokens mounted as **read-only** (`:ro`) in Docker
- ✅ Tokens have **600 permissions** (owner only)
- ✅ Tokens stored in **user home directory** (not in project)
- ✅ Tokens **not committed** to git (in .gitignore)
- ✅ Tokens **expire after 1 hour** (automatic refresh)

## Benefits

- ✅ **No Login Prompts**: Authenticate once, use everywhere
- ✅ **Docker Compatible**: Works in headless containers
- ✅ **Automatic**: Test scripts automatically find tokens
- ✅ **Secure**: Read-only mount, proper permissions
- ✅ **Consistent**: Same tokens across all test runs
- ✅ **CI/CD Ready**: Easy integration with pipelines

## Quick Commands

```bash
# Setup (one time)
./scripts/setup-test-auth.sh

# Run tests in Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh

# Check token status
docker compose -f docker/compose/docker-compose.test.yml run --rm shell \
  bash -c "source /workspace/scripts/lib/auth-helper.sh && get_auth_info"

# Refresh tokens
./scripts/setup-test-auth.sh --refresh
```

## Files Modified

1. ✅ `docker/compose/docker-compose.test.yml` - Added token volume mounts
2. ✅ `scripts/lib/auth-helper.sh` - Created auth helper library

## Files Created

1. ✅ `docs/testing/DOCKER-AUTH-INTEGRATION.md` - This guide

## Next Steps

1. ✅ Setup complete - tokens automatically mounted in Docker
2. ✅ Run `./scripts/setup-test-auth.sh` if you haven't already
3. ✅ Test with: `docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner /workspace/scripts/test-task-5.4-filesystem-operations.sh`

---

**Integration Complete**: Docker containers now have automatic access to authentication tokens!
