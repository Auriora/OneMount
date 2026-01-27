# Systemd Service Authentication Flow Fix

**Date**: 2026-01-27  
**Status**: In Progress  
**Related**: `docs/fixes/systemd-user-service-fix.md`, `docs/fixes/mount-account-registry-integration.md`

## Problem

When starting `onemount` as a systemd service without existing authentication tokens, the OAuth flow fails with:
```
AADSTS70000: The provided value for the 'code' parameter is not valid
```

This happens because:
1. Systemd service has no controlling terminal
2. OAuth browser flow opens but code exchange fails
3. Timing issues or code reuse problems

## Current Behavior

1. User clicks "Create Mount" in launcher
2. Launcher calls `systemctl --user start onemount@<mount>.service`
3. Service starts `onemount <mountpoint>`
4. `onemount` detects no tokens, tries to authenticate
5. Opens browser for OAuth
6. Code exchange fails

## Proposed Solution

The launcher should detect when authentication is needed and run `onemount --auth-only` in a terminal BEFORE starting the service.

### Flow

1. User clicks "Create Mount"
2. Launcher checks registry for account
3. If no account registered:
   - Run `onemount --auth-only <mountpoint>` in terminal
   - Wait for authentication to complete
   - Tokens saved, mount registered
4. Start systemd service
5. Service mounts filesystem using existing tokens

### Implementation

Update `cmd/onemount-launcher/main.go` to:
- Check if mount is in registry
- If not, spawn terminal with `onemount --auth-only`
- Wait for completion
- Then start service

## Alternative: Pre-Authentication Command

Users can manually authenticate before using the launcher:
```bash
onemount --auth-only /path/to/mountpoint
```

This creates tokens and registers the mount, then the launcher can start the service normally.

## Status

- ✅ Registry integration complete
- ✅ Systemd service fixed (User/Group removed)
- ⏳ Launcher pre-authentication flow (TODO)
- ⏳ Documentation for manual pre-auth (TODO)
