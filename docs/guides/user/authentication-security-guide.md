# Authentication and Security Best Practices

**Last Updated**: 2026-02-13  
**Audience**: End Users  
**Related**: [Troubleshooting Guide](troubleshooting-guide.md), [Installation Guide](installation-guide.md)

## Overview

This guide provides security best practices for managing authentication tokens and setting up multiple OneDrive accounts with OneMount. Following these guidelines will help keep your OneDrive data secure.

## Table of Contents

1. [Token Management](#token-management)
2. [Multi-Account Setup](#multi-account-setup)
3. [Security Best Practices](#security-best-practices)
4. [Troubleshooting Authentication](#troubleshooting-authentication)

## Token Management

### What Are Authentication Tokens?

Authentication tokens are credentials that OneMount uses to access your OneDrive account. They consist of:

- **Access Token**: Short-lived token (expires after ~1 hour) used for API requests
- **Refresh Token**: Long-lived token (typically 90 days) used to get new access tokens

### Where Are Tokens Stored?

OneMount stores tokens securely in your home directory:

```
~/.cache/onemount/accounts/{account-hash}/auth_tokens.json
```

The `account-hash` is a unique identifier derived from your account email address.

### Token Security Features

1. **Encryption**: Tokens are encrypted at rest using AES-256 encryption
2. **File Permissions**: Token files have `0600` permissions (only you can read/write)
3. **Secure Location**: Tokens stored in your user cache directory
4. **No Logging**: Tokens are never written to log files

### Token Lifecycle

1. **Initial Authentication**: You authenticate once with Microsoft
2. **Automatic Refresh**: OneMount automatically refreshes expired access tokens
3. **Re-authentication**: If refresh fails, you'll be prompted to authenticate again
4. **Token Expiration**: Refresh tokens expire after ~90 days of inactivity

### Managing Tokens

#### View Token Status

```bash
# Check if tokens exist for a mount point
ls -la ~/.cache/onemount/accounts/*/auth_tokens.json

# View token expiration (requires jq)
cat ~/.cache/onemount/accounts/*/auth_tokens.json | jq '.expires_at, .account'
```

#### Authenticate Without Mounting

```bash
# Authenticate and save tokens without mounting
onemount --auth-only ~/OneDrive
```

#### Remove Tokens

```bash
# Remove all cached data including tokens
onemount --wipe-cache --cache-dir ~/.cache/onemount

# Or manually remove token files
rm -rf ~/.cache/onemount/accounts/
```

#### Refresh Tokens Manually

Tokens are automatically refreshed, but you can force re-authentication:

```bash
# Remove old tokens and re-authenticate
rm ~/.cache/onemount/accounts/*/auth_tokens.json
onemount --auth-only ~/OneDrive
```

## Multi-Account Setup

### Supported Account Types

OneMount supports multiple account types simultaneously:

- **Personal OneDrive**: Microsoft personal accounts (outlook.com, hotmail.com, etc.)
- **OneDrive for Business**: Work or school accounts (Microsoft 365)
- **Shared Drives**: Drives shared with you by other users
- **Shared Items**: Individual files/folders shared with you

### Setting Up Multiple Accounts

#### Step 1: Create Mount Points

```bash
# Create directories for each account
mkdir -p ~/OneDrive-Personal
mkdir -p ~/OneDrive-Work
mkdir -p ~/OneDrive-Shared
```

#### Step 2: Authenticate Each Account

```bash
# Authenticate personal account
onemount --auth-only ~/OneDrive-Personal

# Authenticate work account
onemount --auth-only ~/OneDrive-Work

# Authenticate shared drive (requires drive ID)
onemount --auth-only --drive-id abc123 ~/OneDrive-Shared
```

#### Step 3: Mount Accounts

```bash
# Mount personal account
onemount ~/OneDrive-Personal

# Mount work account
onemount ~/OneDrive-Work

# Mount shared drive
onemount --drive-id abc123 ~/OneDrive-Shared
```

### Account Isolation

Each account has:

- **Separate tokens**: Stored in account-specific directories
- **Separate cache**: Metadata and file content cached independently
- **Separate sync**: Changes tracked independently per account
- **No cross-contamination**: Accounts don't interfere with each other

### Managing Multiple Accounts

#### List Active Mounts

```bash
# View all mounted OneDrive accounts
mount | grep onemount

# Or use the stats command
onemount --stats ~/OneDrive-Personal
onemount --stats ~/OneDrive-Work
```

#### Unmount Accounts

```bash
# Unmount specific account
fusermount -u ~/OneDrive-Personal

# Or use onemount unmount
onemount --unmount ~/OneDrive-Personal
```

#### Switch Between Accounts

You can mount and unmount accounts as needed:

```bash
# Unmount personal account
fusermount -u ~/OneDrive-Personal

# Mount work account instead
onemount ~/OneDrive-Work
```

### Using the Launcher

The OneMount launcher provides a GUI for managing multiple accounts:

```bash
# Start the launcher
onemount-launcher
```

Features:
- View all configured mount points
- See which account is associated with each mount
- Mount/unmount accounts with one click
- Add custom labels to identify accounts

## Security Best Practices

### 1. Protect Your Tokens

- **Never share token files** with others
- **Don't commit tokens to git** (they're in `.gitignore` by default)
- **Use secure backups** if backing up your home directory
- **Rotate tokens regularly** by re-authenticating

### 2. Use Strong Microsoft Account Security

- **Enable two-factor authentication** on your Microsoft account
- **Use a strong password** for your Microsoft account
- **Review authorized apps** regularly in Microsoft account settings
- **Revoke access** if you stop using OneMount

### 3. Secure Your System

- **Use full disk encryption** to protect cached data
- **Lock your screen** when away from your computer
- **Keep your system updated** with security patches
- **Use a firewall** to protect network connections

### 4. Monitor Access

- **Check Microsoft account activity** regularly
- **Review OneMount logs** for suspicious activity
- **Monitor file access** in OneDrive web interface
- **Set up alerts** for unusual activity

### 5. Handle Shared Computers

If using OneMount on a shared computer:

- **Unmount before leaving** your session
- **Clear cache** when done: `onemount --wipe-cache`
- **Use separate user accounts** on the system
- **Consider using headless mode** to avoid browser history

### 6. Network Security

- **Use trusted networks** for authentication
- **Avoid public Wi-Fi** when authenticating
- **Use VPN** on untrusted networks
- **Verify HTTPS** in browser during authentication

## Troubleshooting Authentication

### "Authentication failed" Error

**Symptoms**: OneMount can't authenticate or mount fails

**Causes**:
- Expired or missing tokens
- Network connectivity issues
- Microsoft account issues
- Incorrect permissions on token files

**Solutions**:

1. **Check token files exist**:
   ```bash
   ls -la ~/.cache/onemount/accounts/*/auth_tokens.json
   ```

2. **Verify file permissions**:
   ```bash
   # Should show -rw------- (0600)
   ls -l ~/.cache/onemount/accounts/*/auth_tokens.json
   ```

3. **Check token expiration**:
   ```bash
   cat ~/.cache/onemount/accounts/*/auth_tokens.json | jq '.expires_at'
   # Compare with current time: date +%s
   ```

4. **Re-authenticate**:
   ```bash
   rm ~/.cache/onemount/accounts/*/auth_tokens.json
   onemount --auth-only ~/OneDrive
   ```

5. **Check network connectivity**:
   ```bash
   ping graph.microsoft.com
   curl -I https://graph.microsoft.com
   ```

### "Token refresh failed" Error

**Symptoms**: OneMount can't refresh expired access token

**Causes**:
- Refresh token expired (>90 days inactive)
- Microsoft account password changed
- App authorization revoked
- Network issues

**Solutions**:

1. **Re-authenticate**:
   ```bash
   onemount --auth-only ~/OneDrive
   ```

2. **Check Microsoft account**:
   - Verify password hasn't changed
   - Check if OneMount is still authorized
   - Visit https://account.microsoft.com/privacy/app-access

3. **Clear old tokens**:
   ```bash
   rm -rf ~/.cache/onemount/accounts/
   onemount --auth-only ~/OneDrive
   ```

### "Permission denied" on Token Files

**Symptoms**: Can't read or write token files

**Causes**:
- Incorrect file permissions
- File owned by different user
- Filesystem issues

**Solutions**:

1. **Fix permissions**:
   ```bash
   chmod 600 ~/.cache/onemount/accounts/*/auth_tokens.json
   ```

2. **Fix ownership**:
   ```bash
   chown $USER:$USER ~/.cache/onemount/accounts/*/auth_tokens.json
   ```

3. **Recreate tokens**:
   ```bash
   rm -rf ~/.cache/onemount/accounts/
   onemount --auth-only ~/OneDrive
   ```

### Multiple Accounts Not Working

**Symptoms**: Can't mount multiple accounts simultaneously

**Causes**:
- Using same mount point for different accounts
- Token conflicts
- Registry issues

**Solutions**:

1. **Use different mount points**:
   ```bash
   # Wrong - same mount point
   onemount ~/OneDrive  # Personal
   onemount ~/OneDrive  # Work (conflicts!)
   
   # Right - different mount points
   onemount ~/OneDrive-Personal
   onemount ~/OneDrive-Work
   ```

2. **Check mount registry**:
   ```bash
   cat ~/.config/onemount/mounts.json
   ```

3. **Clear registry and re-mount**:
   ```bash
   rm ~/.config/onemount/mounts.json
   onemount --auth-only ~/OneDrive-Personal
   onemount --auth-only ~/OneDrive-Work
   ```

### Headless Authentication Issues

**Symptoms**: Can't authenticate on headless system

**Causes**:
- No browser available
- Display not set
- GTK not available

**Solutions**:

1. **Use headless mode**:
   ```bash
   onemount --auth-only --no-browser ~/OneDrive
   ```

2. **Follow device code flow**:
   - OneMount displays a URL and code
   - Open URL in browser on another device
   - Enter the code
   - Copy the redirect URL back to OneMount

3. **Use SSH with X11 forwarding** (if available):
   ```bash
   ssh -X user@server
   onemount --auth-only ~/OneDrive
   ```

### Docker/Container Authentication

**Symptoms**: Authentication fails in Docker containers

**Causes**:
- Token files not mounted into container
- Incorrect paths in container
- Permission issues

**Solutions**:

See [Docker Test Environment](../../4-testing/docker/persistent-authentication-setup.md) for detailed Docker authentication setup.

## Getting Help

If you continue to have authentication issues:

1. **Check logs**:
   ```bash
   journalctl -u onemount@$(systemd-escape ~/OneDrive).service
   ```

2. **Enable debug logging**:
   ```bash
   onemount --debug ~/OneDrive
   ```

3. **Review documentation**:
   - [Troubleshooting Guide](troubleshooting-guide.md)
   - [Installation Guide](installation-guide.md)
   - [Architecture Documentation](../../2-architecture/authentication.md)

4. **Report issues**:
   - GitHub Issues: https://github.com/jstaf/onedriver/issues
   - Include logs (with tokens redacted)
   - Describe steps to reproduce

## References

- [Authentication Architecture](../../2-architecture/authentication.md)
- [Authentication Token Paths](../developer/authentication-token-paths.md)
- [Security Testing Guide](../../4-testing/guides/frameworks/security-testing-guide.md)
- [Troubleshooting Guide](troubleshooting-guide.md)
