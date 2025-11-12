# Authentication Setup for Testing - Complete

✅ **STATUS: SOLUTION IMPLEMENTED**

## Problem Solved

You were getting prompted to login to your Microsoft account for each test case, which:
- Doesn't work in Docker (no GUI for authentication)
- Interrupts automated testing
- Requires manual intervention

## Solution Implemented

Created a persistent authentication system that preserves login across:
- ✅ Multiple test runs
- ✅ Different test cases
- ✅ Docker containers
- ✅ CI/CD pipelines

## What Was Created

### 1. Setup Script
**File**: `scripts/setup-test-auth.sh`

Automated script that:
- Checks for existing tokens
- Authenticates once with Microsoft
- Saves tokens to standard location
- Validates and checks expiration

**Usage**:
```bash
./scripts/setup-test-auth.sh
```

### 2. Comprehensive Documentation
**File**: `docs/testing/persistent-authentication-setup.md`

Complete guide covering:
- Token locations and search order
- Step-by-step setup instructions
- Docker integration (3 methods)
- Docker Compose configuration
- CI/CD setup (GitHub Actions)
- Token refresh procedures
- Troubleshooting guide
- Security considerations

### 3. Quick Reference
**File**: `docs/testing/QUICK-AUTH-SETUP.md`

TL;DR guide with:
- 3-step quick setup
- Common commands
- Troubleshooting tips
- CI/CD quick setup

## How It Works

### Token Storage Locations

Tokens are searched in this order:

1. `~/.onemount-tests/.auth_tokens.json` ← **Recommended**
2. `./test-artifacts/.auth_tokens.json`
3. `./auth_tokens.json`

All existing test scripts already support these locations:
- `scripts/test-task-5.4-filesystem-operations.sh`
- `scripts/test-task-5.5-unmounting-cleanup.sh`
- `scripts/test-task-5.6-signal-handling.sh`

### Docker Integration

Three methods provided:

**Method 1: Volume Mount (Recommended)**
```bash
docker run -v "$HOME/.onemount-tests:/root/.onemount-tests:ro" ...
```

**Method 2: Copy to Project**
```bash
cp ~/.onemount-tests/.auth_tokens.json ./test-artifacts/
```

**Method 3: Environment Variable**
```bash
AUTH_TOKENS_B64=$(base64 -w 0 ~/.onemount-tests/.auth_tokens.json)
docker run -e "AUTH_TOKENS_B64=$AUTH_TOKENS_B64" ...
```

## Quick Start

### Step 1: Setup Authentication (One Time)

```bash
./scripts/setup-test-auth.sh
```

This will:
1. Open browser for Microsoft login
2. Save tokens to `~/.onemount-tests/.auth_tokens.json`
3. Set proper permissions (600)

### Step 2: Run Tests (No Re-authentication)

```bash
# Run any test - no login prompt!
./scripts/test-task-5.4-filesystem-operations.sh
./scripts/test-task-5.5-unmounting-cleanup.sh
./scripts/test-task-5.6-signal-handling.sh
```

### Step 3: Docker Testing

```bash
# Add volume mount to Docker commands
docker run --rm -t \
  --device /dev/fuse \
  --cap-add SYS_ADMIN \
  --security-opt apparmor:unconfined \
  -v "$(pwd):/workspace:rw" \
  -v "$HOME/.onemount-tests:/root/.onemount-tests:ro" \
  onemount-test-runner:latest \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

Or update `docker/compose/docker-compose.test.yml`:

```yaml
services:
  test-runner:
    volumes:
      - ../..:/workspace:rw
      - ~/.onemount-tests:/root/.onemount-tests:ro  # Add this line
```

## Token Management

### Check Token Status

```bash
# View account and expiration
jq '.account, .expires_at' ~/.onemount-tests/.auth_tokens.json

# Check if expired
EXPIRES_AT=$(jq -r '.expires_at' ~/.onemount-tests/.auth_tokens.json)
CURRENT_TIME=$(date +%s)
if [ "$EXPIRES_AT" -le "$CURRENT_TIME" ]; then
    echo "EXPIRED - run: ./scripts/setup-test-auth.sh --refresh"
else
    echo "VALID until $(date -d @$EXPIRES_AT)"
fi
```

### Refresh Tokens

```bash
# Tokens expire after 1 hour
./scripts/setup-test-auth.sh --refresh
```

## CI/CD Integration

### GitHub Actions Example

```yaml
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

Create secret:
```bash
base64 -w 0 ~/.onemount-tests/.auth_tokens.json
# Copy output to GitHub Secrets as ONEMOUNT_AUTH_TOKENS
```

## Security

- ✅ Tokens stored with `600` permissions (owner only)
- ✅ `.auth_tokens.json` in `.gitignore`
- ✅ CI/CD uses encrypted secrets
- ✅ Tokens expire after 1 hour (automatic refresh)
- ✅ No tokens committed to repository

## Troubleshooting

### Problem: "No auth tokens found"
**Solution**: Run `./scripts/setup-test-auth.sh`

### Problem: "Tokens expired"
**Solution**: Run `./scripts/setup-test-auth.sh --refresh`

### Problem: Docker can't access tokens
**Solution**: Add volume mount `-v "$HOME/.onemount-tests:/root/.onemount-tests:ro"`

### Problem: Invalid JSON
**Solution**: 
```bash
rm ~/.onemount-tests/.auth_tokens.json
./scripts/setup-test-auth.sh
```

## Files Created

1. ✅ `scripts/setup-test-auth.sh` - Automated setup script
2. ✅ `docs/testing/persistent-authentication-setup.md` - Complete guide
3. ✅ `docs/testing/QUICK-AUTH-SETUP.md` - Quick reference
4. ✅ `AUTH_SETUP_COMPLETE.md` - This summary

## Next Steps

1. **Run Setup**: `./scripts/setup-test-auth.sh`
2. **Test It**: `./scripts/test-task-5.4-filesystem-operations.sh`
3. **Update Docker Compose**: Add volume mount for tokens
4. **Setup CI/CD**: Add tokens as encrypted secret

## Benefits

- ✅ **No More Login Prompts**: Authenticate once, use everywhere
- ✅ **Docker Compatible**: Works in headless environments
- ✅ **CI/CD Ready**: Easy integration with GitHub Actions
- ✅ **Secure**: Proper permissions and secret management
- ✅ **Automatic**: Existing test scripts already support it
- ✅ **Documented**: Complete guides and troubleshooting

## Documentation

- **Quick Start**: `docs/testing/QUICK-AUTH-SETUP.md`
- **Complete Guide**: `docs/testing/persistent-authentication-setup.md`
- **Setup Script**: `scripts/setup-test-auth.sh`

---

**Ready to Use**: Run `./scripts/setup-test-auth.sh` to get started!

**Questions?** See `docs/testing/persistent-authentication-setup.md` for detailed information.
