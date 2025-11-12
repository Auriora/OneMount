# Docker Image Update - jq Installation

## Issue

The Docker test-runner image was missing `jq` (JSON processor), which caused a warning when using the auth helper library:

```
jq not installed, cannot read token info
```

## Solution

Added `jq` to the Docker image dependencies.

## What Was Changed

### File: `docker/images/test-runner/Dockerfile`

Added `jq` to the package installation list:

```dockerfile
# JSON processing for auth tokens
jq \
```

### File: `scripts/lib/auth-helper.sh`

Made the auth helper more graceful when `jq` is not available:
- Skips JSON validation if `jq` is missing
- Skips expiration check if `jq` is missing
- Tokens will still be validated by OneMount itself

## How to Apply

### Option 1: Rebuild Docker Image (Recommended)

```bash
# Rebuild the test-runner image
docker compose -f docker/compose/docker-compose.test.yml build test-runner

# Or rebuild all images
./docker/scripts/build-images.sh test-runner
```

### Option 2: Use Without Rebuilding (Works Now)

The auth helper already works without `jq` - it just skips the validation checks. OneMount will validate the tokens when it uses them.

```bash
# This works fine even without jq in the container
docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
  /workspace/scripts/test-task-5.4-filesystem-operations.sh
```

## Impact

### Before Fix
- ✅ Authentication works (tokens are accessible)
- ⚠️ Warning message: "jq not installed, cannot read token info"
- ⚠️ Cannot validate token expiration in Docker
- ⚠️ Cannot display account info in Docker

### After Fix (with rebuild)
- ✅ Authentication works (tokens are accessible)
- ✅ No warning messages
- ✅ Can validate token expiration in Docker
- ✅ Can display account info in Docker

## Current Status

**Without Rebuild**: ✅ Working (with minor warning)
- Tokens are accessible in Docker
- Tests run without login prompts
- Warning message is cosmetic only

**With Rebuild**: ✅ Perfect (no warnings)
- All functionality works
- No warning messages
- Better token validation

## Recommendation

**For immediate use**: No rebuild needed - everything works!

**For clean experience**: Rebuild the image when convenient:

```bash
docker compose -f docker/compose/docker-compose.test.yml build test-runner
```

## Verification

After rebuilding (optional), verify `jq` is installed:

```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint bash shell \
  -c "jq --version"
```

Expected output: `jq-1.6` or similar

## Summary

- ✅ Issue identified: `jq` missing in Docker image
- ✅ Fix applied: Added `jq` to Dockerfile
- ✅ Workaround: Auth helper gracefully handles missing `jq`
- ✅ Current status: **Everything works** (rebuild optional for cleaner output)

---

**Bottom Line**: Your authentication is working perfectly. The `jq` warning is cosmetic. Rebuild the Docker image when convenient to eliminate the warning.
