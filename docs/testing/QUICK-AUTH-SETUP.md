# Quick Authentication Setup for Testing

**TL;DR**: Run `./scripts/setup-test-auth.sh` once, then all tests work without re-authentication.

## Problem

Tests keep prompting for Microsoft login, which:
- ❌ Doesn't work in Docker (no GUI)
- ❌ Interrupts automated testing
- ❌ Requires manual intervention each time

## Solution

Set up persistent authentication tokens once, use everywhere.

## Quick Setup (3 Steps)

### 1. Run Setup Script

```bash
./scripts/setup-test-auth.sh
```

This will:
- Check for existing tokens
- Authenticate with Microsoft (browser opens)
- Save tokens to `~/.onemount-tests/.auth_tokens.json`

### 2. Run Tests

```bash
# All tests now work without re-authentication
./scripts/test-task-5.4-filesystem-operations.sh
./scripts/test-task-5.5-unmounting-cleanup.sh
./scripts/test-task-5.6-signal-handling.sh
```

### 3. Docker Setup

Add volume mount to Docker commands:

```bash
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

## Token Locations

Tokens are searched in this order:

1. `~/.onemount-tests/.auth_tokens.json` ← **Recommended**
2. `./test-artifacts/.auth_tokens.json`
3. `./auth_tokens.json`

## Token Expiration

Tokens expire after 1 hour. To refresh:

```bash
./scripts/setup-test-auth.sh --refresh
```

Or check manually:

```bash
# Check expiration
jq '.expires_at' ~/.onemount-tests/.auth_tokens.json

# Current time
date +%s

# If expires_at < current time, tokens are expired
```

## Troubleshooting

### "No auth tokens found"

```bash
# Run setup script
./scripts/setup-test-auth.sh
```

### "Tokens expired"

```bash
# Refresh tokens
./scripts/setup-test-auth.sh --refresh
```

### Docker can't find tokens

```bash
# Check volume mount
docker run --rm -t \
  -v "$HOME/.onemount-tests:/root/.onemount-tests:ro" \
  onemount-test-runner:latest \
  ls -la /root/.onemount-tests/

# Should show .auth_tokens.json with 600 permissions
```

### Invalid JSON

```bash
# Validate
jq empty ~/.onemount-tests/.auth_tokens.json

# If invalid, re-run setup
rm ~/.onemount-tests/.auth_tokens.json
./scripts/setup-test-auth.sh
```

## CI/CD Setup

### GitHub Actions

1. Create secret `ONEMOUNT_AUTH_TOKENS`:
   ```bash
   base64 -w 0 ~/.onemount-tests/.auth_tokens.json
   ```

2. Add to workflow:
   ```yaml
   - name: Setup auth
     run: |
       mkdir -p ~/.onemount-tests
       echo "${{ secrets.ONEMOUNT_AUTH_TOKENS }}" | base64 -d > ~/.onemount-tests/.auth_tokens.json
       chmod 600 ~/.onemount-tests/.auth_tokens.json
   ```

## Security

- ✅ Tokens are stored with `600` permissions (owner read/write only)
- ✅ `.auth_tokens.json` is in `.gitignore`
- ✅ Use encrypted secrets in CI/CD
- ⚠️ Tokens expire after 1 hour (refresh as needed)

## More Information

See full documentation: `docs/testing/persistent-authentication-setup.md`

---

**Quick Commands**:

```bash
# Setup
./scripts/setup-test-auth.sh

# Check status
jq '.account, .expires_at' ~/.onemount-tests/.auth_tokens.json

# Refresh
./scripts/setup-test-auth.sh --refresh

# Run tests
./scripts/test-task-5.4-filesystem-operations.sh
```
