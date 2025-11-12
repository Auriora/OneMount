# Persistent Authentication Setup for Testing

**Date**: 2025-11-12  
**Purpose**: Configure persistent authentication tokens for test runs and Docker environments

## Overview

This guide explains how to set up persistent authentication tokens that work across:
- Multiple test runs
- Different test cases
- Docker containers (without GUI)
- CI/CD pipelines

## Problem

By default, OneMount prompts for Microsoft account login for each test case, which:
- Interrupts automated testing
- Doesn't work in Docker (no GUI for authentication)
- Slows down test execution
- Requires manual intervention

## Solution

Use pre-authenticated tokens stored in a standard location that all tests can access.

## Authentication Token Locations

OneMount searches for authentication tokens in the following order:

1. `$HOME/.onemount-tests/.auth_tokens.json` (recommended for testing)
2. `./test-artifacts/.auth_tokens.json` (project-specific)
3. `./auth_tokens.json` (workspace root)
4. Custom path via `--auth-path` flag

## Setup Instructions

### Step 1: Authenticate Once (Interactive)

Run OneMount once interactively to generate authentication tokens:

```bash
# Create directory for test tokens
mkdir -p ~/.onemount-tests

# Run OneMount with authentication
./build/onemount --cache-dir=/tmp/onemount-cache-auth /tmp/onemount-mount-auth
```

This will:
1. Open a browser for Microsoft account login
2. Save tokens to `~/.cache/onemount/tmp-onemount-auth/auth_tokens.json`
3. Mount your OneDrive

### Step 2: Copy Tokens to Test Location

```bash
# Copy tokens to standard test location
cp ~/.cache/onemount/*/auth_tokens.json ~/.onemount-tests/.auth_tokens.json

# Set proper permissions
chmod 600 ~/.onemount-tests/.auth_tokens.json

# Verify tokens are valid JSON
jq . ~/.onemount-tests/.auth_tokens.json
```

### Step 3: Verify Token Expiration

```bash
# Check token expiration
jq '.expires_at' ~/.onemount-tests/.auth_tokens.json

# Compare with current time
date +%s

# If expires_at < current time, tokens are expired and need refresh
```

### Step 4: Test Token Usage

```bash
# Run a test script to verify tokens work
./scripts/test-task-5.4-filesystem-operations.sh
```

## Docker Setup

### Option 1: Mount Tokens as Volume (Recommended)

```bash
# Run Docker with token volume
docker run --rm -t \
  --user root \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -v "$(pwd):/workspace:rw" \
  -v "$HOME/.onemount-tests:/root/.onemount-tests:ro" \
  onemount-test-runner:latest \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

### Option 2: Copy Tokens to Project

```bash
# Copy tokens to project directory
cp ~/.onemount-tests/.auth_tokens.json ./test-artifacts/.auth_tokens.json

# Docker will find them automatically
docker run --rm -t \
  --user root \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -v "$(pwd):/workspace:rw" \
  onemount-test-runner:latest \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

### Option 3: Environment Variable (Base64 Encoded)

```bash
# Encode tokens as base64
AUTH_TOKENS_B64=$(base64 -w 0 ~/.onemount-tests/.auth_tokens.json)

# Pass to Docker
docker run --rm -t \
  --user root \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -v "$(pwd):/workspace:rw" \
  -e "AUTH_TOKENS_B64=$AUTH_TOKENS_B64" \
  onemount-test-runner:latest \
  bash -c "echo \$AUTH_TOKENS_B64 | base64 -d > /root/.onemount-tests/.auth_tokens.json && chmod 600 /root/.onemount-tests/.auth_tokens.json && /workspace/scripts/test-task-5.4-filesystem-operations.sh"
```

## Docker Compose Setup

Update `docker/compose/docker-compose.test.yml`:

```yaml
services:
  test-runner:
    image: onemount-test-runner:latest
    volumes:
      - ../..:/workspace:rw
      - ~/.onemount-tests:/root/.onemount-tests:ro  # Add this line
    devices:
      - /dev/fuse
    cap_add:
      - SYS_ADMIN
    security_opt:
      - apparmor:unconfined
```

Then run tests:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

## CI/CD Setup (GitHub Actions)

### Step 1: Create GitHub Secret

1. Go to repository Settings → Secrets and variables → Actions
2. Create new secret: `ONEMOUNT_AUTH_TOKENS`
3. Value: Base64-encoded auth tokens

```bash
# Generate base64 value for secret
base64 -w 0 ~/.onemount-tests/.auth_tokens.json
```

### Step 2: Update Workflow

```yaml
name: Test OneMount

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup auth tokens
        run: |
          mkdir -p ~/.onemount-tests
          echo "${{ secrets.ONEMOUNT_AUTH_TOKENS }}" | base64 -d > ~/.onemount-tests/.auth_tokens.json
          chmod 600 ~/.onemount-tests/.auth_tokens.json
      
      - name: Run tests
        run: |
          docker compose -f docker/compose/docker-compose.test.yml run --rm \
            -v ~/.onemount-tests:/root/.onemount-tests:ro \
            test-runner /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

## Token Refresh

Tokens expire after 1 hour by default. To refresh:

### Manual Refresh

```bash
# Run OneMount to trigger automatic refresh
./build/onemount --cache-dir=/tmp/cache-refresh /tmp/mount-refresh

# Copy refreshed tokens
cp ~/.cache/onemount/*/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
```

### Automatic Refresh Script

Create `scripts/refresh-auth-tokens.sh`:

```bash
#!/bin/bash
set -e

AUTH_FILE="$HOME/.onemount-tests/.auth_tokens.json"

if [ ! -f "$AUTH_FILE" ]; then
    echo "No auth tokens found at $AUTH_FILE"
    exit 1
fi

# Check expiration
EXPIRES_AT=$(jq -r '.expires_at' "$AUTH_FILE")
CURRENT_TIME=$(date +%s)

if [ "$EXPIRES_AT" -le "$CURRENT_TIME" ]; then
    echo "Tokens expired, refreshing..."
    
    # Mount to trigger refresh
    TEMP_MOUNT=$(mktemp -d)
    TEMP_CACHE=$(mktemp -d)
    
    timeout 30 ./build/onemount --cache-dir="$TEMP_CACHE" "$TEMP_MOUNT" &
    MOUNT_PID=$!
    
    sleep 10
    
    # Kill mount
    kill $MOUNT_PID 2>/dev/null || true
    fusermount3 -uz "$TEMP_MOUNT" 2>/dev/null || true
    
    # Copy refreshed tokens
    find "$TEMP_CACHE" -name "auth_tokens.json" -exec cp {} "$AUTH_FILE" \;
    
    # Cleanup
    rm -rf "$TEMP_MOUNT" "$TEMP_CACHE"
    
    echo "Tokens refreshed successfully"
else
    echo "Tokens still valid (expires: $(date -d @$EXPIRES_AT))"
fi
```

## Troubleshooting

### Tokens Not Found

```bash
# Check all possible locations
for loc in "$HOME/.onemount-tests/.auth_tokens.json" \
           "./test-artifacts/.auth_tokens.json" \
           "./auth_tokens.json"; do
    if [ -f "$loc" ]; then
        echo "Found: $loc"
        jq '.account, .expires_at' "$loc"
    fi
done
```

### Tokens Expired

```bash
# Check expiration
EXPIRES_AT=$(jq -r '.expires_at' ~/.onemount-tests/.auth_tokens.json)
CURRENT_TIME=$(date +%s)

if [ "$EXPIRES_AT" -le "$CURRENT_TIME" ]; then
    echo "EXPIRED: $(date -d @$EXPIRES_AT)"
    echo "Current: $(date -d @$CURRENT_TIME)"
    echo "Run: ./scripts/refresh-auth-tokens.sh"
fi
```

### Invalid JSON

```bash
# Validate JSON
jq empty ~/.onemount-tests/.auth_tokens.json

# If invalid, re-authenticate
rm ~/.onemount-tests/.auth_tokens.json
./build/onemount --cache-dir=/tmp/cache /tmp/mount
```

### Docker Can't Access Tokens

```bash
# Check volume mount
docker run --rm -t \
  -v "$HOME/.onemount-tests:/root/.onemount-tests:ro" \
  onemount-test-runner:latest \
  ls -la /root/.onemount-tests/

# Check permissions
ls -la ~/.onemount-tests/.auth_tokens.json
# Should be: -rw------- (600)
```

## Security Considerations

1. **File Permissions**: Always use `chmod 600` for token files
2. **Git Ignore**: Ensure `.auth_tokens.json` is in `.gitignore`
3. **CI/CD Secrets**: Use encrypted secrets, never commit tokens
4. **Token Rotation**: Refresh tokens regularly (they expire after 1 hour)
5. **Access Control**: Limit who can access token files

## Best Practices

1. **Use Standard Location**: `~/.onemount-tests/.auth_tokens.json` for consistency
2. **Check Expiration**: Verify tokens before running tests
3. **Automate Refresh**: Use refresh script in CI/CD pipelines
4. **Document Setup**: Include auth setup in test documentation
5. **Test Isolation**: Each test should work with same tokens

## Integration with Existing Tests

All test scripts already support this setup:

- `scripts/test-task-5.4-filesystem-operations.sh`
- `scripts/test-task-5.5-unmounting-cleanup.sh`
- `scripts/test-task-5.6-signal-handling.sh`

They automatically search for tokens in standard locations.

## Quick Start

```bash
# 1. Authenticate once
./build/onemount --cache-dir=/tmp/cache /tmp/mount

# 2. Copy tokens
mkdir -p ~/.onemount-tests
cp ~/.cache/onemount/*/auth_tokens.json ~/.onemount-tests/.auth_tokens.json
chmod 600 ~/.onemount-tests/.auth_tokens.json

# 3. Run tests
./scripts/test-task-5.4-filesystem-operations.sh

# 4. For Docker
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  -v ~/.onemount-tests:/root/.onemount-tests:ro \
  test-runner /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

## References

- Authentication Code: `internal/graph/oauth2.go`
- Token Storage: `internal/graph/authenticator.go`
- Test Scripts: `scripts/test-task-*.sh`
- Docker Compose: `docker/compose/docker-compose.test.yml`

---

**Last Updated**: 2025-11-12  
**Maintained By**: OneMount Development Team
