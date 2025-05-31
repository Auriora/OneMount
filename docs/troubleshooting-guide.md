# OneMount Troubleshooting Guide

## Table of Contents
1. [Common Issues](#common-issues)
2. [Installation Problems](#installation-problems)
3. [Authentication Issues](#authentication-issues)
4. [Network and Connectivity](#network-and-connectivity)
5. [File Operation Problems](#file-operation-problems)
6. [Performance Issues](#performance-issues)
7. [Debugging and Logging](#debugging-and-logging)
8. [Getting Help](#getting-help)

## Common Issues

### Filesystem Appears to Hang or "Freeze"

**Symptoms:**
- File operations become unresponsive
- Directory listings don't complete
- Applications hang when accessing files

**Causes:**
- Network connectivity issues
- OneMount process has crashed
- FUSE filesystem is in an inconsistent state

**Solutions:**
1. **Check if OneMount is running:**
   ```bash
   ps aux | grep onemount
   ```

2. **Unmount and remount the filesystem:**
   ```bash
   # Force unmount
   fusermount3 -uz /path/to/mount/point
   
   # Wait a moment, then remount
   onemount /path/to/mount/point
   ```

3. **Check system logs for errors:**
   ```bash
   journalctl --user -u onemount@* --since "1 hour ago"
   ```

### "Read-only filesystem" Error

**Symptoms:**
- Cannot create, modify, or delete files
- Error messages about read-only filesystem
- File operations fail with permission errors

**Causes:**
- Computer is offline (OneMount automatically switches to read-only mode)
- Network connectivity issues
- Authentication token has expired

**Solutions:**
1. **Check network connectivity:**
   ```bash
   ping -c 3 graph.microsoft.com
   ```

2. **Verify OneMount can reach Microsoft Graph API:**
   ```bash
   # Enable debug logging to see network requests
   ONEMOUNT_DEBUG=1 onemount /path/to/mount/point
   ```

3. **Check authentication status:**
   ```bash
   # Try re-authenticating
   onemount --auth-only /path/to/mount/point
   ```

### Files Not Syncing

**Symptoms:**
- Changes made locally don't appear in OneDrive web interface
- Changes made in OneDrive don't appear locally
- Sync appears stuck or incomplete

**Causes:**
- Network connectivity issues
- Large files taking time to upload
- Conflict resolution in progress
- Authentication issues

**Solutions:**
1. **Check sync status:**
   ```bash
   onemount --stats /path/to/mount/point
   ```

2. **Force synchronization:**
   ```bash
   # Unmount and remount to trigger sync
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

3. **Check for conflicts:**
   ```bash
   # Look for conflict files (files with "conflict" in the name)
   find /path/to/mount/point -name "*conflict*" -type f
   ```

## Installation Problems

### Package Installation Fails

**Fedora/CentOS/RHEL:**
```bash
# Ensure COPR repository is properly enabled
sudo dnf copr enable auriora/onemount
sudo dnf clean all
sudo dnf install onemount
```

**Ubuntu/Debian:**
```bash
# Currently requires building from source
# See installation guide for complete instructions
sudo apt update
sudo apt install golang gcc pkg-config libwebkit2gtk-4.0-dev libjson-glib-dev
```

### Build from Source Fails

**Missing Dependencies:**
```bash
# Fedora/CentOS/RHEL
sudo dnf install golang gcc pkg-config webkit2gtk4.0-devel json-glib-devel

# Ubuntu/Debian
sudo apt install golang gcc pkg-config libwebkit2gtk-4.0-dev libjson-glib-dev

# Arch Linux
sudo pacman -S go gcc pkg-config webkit2gtk json-glib
```

**Go Version Issues:**
```bash
# Check Go version (requires 1.24.2 or later)
go version

# Update Go if needed
# Download from https://golang.org/dl/
```

## Authentication Issues

### Authentication Fails

**Symptoms:**
- Browser doesn't open for authentication
- Authentication window shows errors
- "Authentication failed" messages

**Solutions:**
1. **Clear existing authentication:**
   ```bash
   # Remove cached authentication tokens
   rm -f ~/.config/onemount/.auth_tokens.json
   ```

2. **Try authentication-only mode:**
   ```bash
   onemount --auth-only /path/to/mount/point
   ```

3. **Check browser availability:**
   ```bash
   # Ensure a web browser is available
   which firefox || which chromium || which google-chrome
   ```

### Token Refresh Fails

**Symptoms:**
- Periodic authentication failures
- "Token expired" errors
- Automatic re-authentication doesn't work

**Solutions:**
1. **Manual re-authentication:**
   ```bash
   onemount --auth-only /path/to/mount/point
   ```

2. **Check system time:**
   ```bash
   # Ensure system time is correct
   timedatectl status
   ```

## Network and Connectivity

### Network Detection Issues

**Symptoms:**
- OneMount doesn't detect when network is restored
- Stays in offline mode when online
- Frequent online/offline transitions

**Solutions:**
1. **Check network connectivity:**
   ```bash
   # Test Microsoft Graph API connectivity
   curl -I https://graph.microsoft.com/v1.0/me
   ```

2. **Force network state check:**
   ```bash
   # Restart OneMount to re-detect network state
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

### Proxy and Firewall Issues

**Symptoms:**
- Cannot connect to Microsoft Graph API
- Authentication fails
- Network timeouts

**Solutions:**
1. **Configure proxy settings:**
   ```bash
   export HTTP_PROXY=http://proxy.example.com:8080
   export HTTPS_PROXY=http://proxy.example.com:8080
   onemount /path/to/mount/point
   ```

2. **Check firewall rules:**
   ```bash
   # Ensure access to Microsoft Graph API endpoints
   # Required domains: graph.microsoft.com, login.microsoftonline.com
   ```

## File Operation Problems

### Large File Upload/Download Issues

**Symptoms:**
- Large files fail to upload or download
- Transfers are interrupted
- Progress appears stuck

**Solutions:**
1. **Check available disk space:**
   ```bash
   df -h /path/to/mount/point
   ```

2. **Monitor transfer progress:**
   ```bash
   # Enable debug logging to see transfer details
   ONEMOUNT_DEBUG=1 onemount /path/to/mount/point
   ```

3. **Resume interrupted transfers:**
   ```bash
   # OneMount automatically resumes interrupted transfers
   # If stuck, unmount and remount to retry
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

### Permission Errors

**Symptoms:**
- "Permission denied" errors
- Cannot access certain files or directories
- Operations fail with access errors

**Solutions:**
1. **Check mount point permissions:**
   ```bash
   ls -la /path/to/mount/point
   ```

2. **Verify user permissions:**
   ```bash
   # Ensure user has access to mount point
   whoami
   groups
   ```

3. **Check OneDrive permissions:**
   - Verify file permissions in OneDrive web interface
   - Ensure account has necessary access rights

## Performance Issues

### Slow File Operations

**Symptoms:**
- File operations take a long time
- Directory listings are slow
- High CPU or memory usage

**Solutions:**
1. **Check system resources:**
   ```bash
   # Monitor OneMount resource usage
   top -p $(pgrep onemount)
   ```

2. **Clear cache if needed:**
   ```bash
   # Reset OneMount completely (will re-download files)
   onemount -w /path/to/mount/point
   ```

3. **Optimize cache settings:**
   ```bash
   # Check cache usage
   onemount --stats /path/to/mount/point
   ```

## Debugging and Logging

### Enable Debug Logging

```bash
# Enable detailed debug output
ONEMOUNT_DEBUG=1 onemount /path/to/mount/point

# Or set environment variable permanently
echo 'export ONEMOUNT_DEBUG=1' >> ~/.bashrc
```

### Check System Logs

```bash
# View OneMount logs
journalctl --user -u onemount@* --since today

# View system FUSE logs
dmesg | grep -i fuse

# View authentication logs
journalctl --user -u onemount@* | grep -i auth
```

### Collect Diagnostic Information

```bash
# System information
uname -a
lsb_release -a

# OneMount version
onemount --version

# Mount information
mount | grep onemount
cat /proc/mounts | grep onemount

# Network connectivity
ping -c 3 graph.microsoft.com
curl -I https://graph.microsoft.com/v1.0/me
```

## Getting Help

### Before Reporting Issues

1. **Check this troubleshooting guide**
2. **Search existing issues**: [GitHub Issues](https://github.com/auriora/OneMount/issues)
3. **Collect diagnostic information** (see above)
4. **Try basic troubleshooting steps**

### Reporting Bugs

When reporting issues, please include:

1. **System Information:**
   - Linux distribution and version
   - OneMount version
   - Go version (if building from source)

2. **Problem Description:**
   - What you were trying to do
   - What happened instead
   - Steps to reproduce the issue

3. **Logs and Output:**
   - Debug output (`ONEMOUNT_DEBUG=1`)
   - System logs (`journalctl` output)
   - Error messages

4. **Configuration:**
   - Mount command used
   - Any custom configuration
   - Network environment (proxy, firewall, etc.)

### Community Support

- **GitHub Issues**: [https://github.com/auriora/OneMount/issues](https://github.com/auriora/OneMount/issues)
- **Documentation**: [https://github.com/auriora/OneMount/tree/main/docs](https://github.com/auriora/OneMount/tree/main/docs)

### Emergency Recovery

If OneMount is completely broken and you need to recover:

```bash
# Force unmount all OneMount filesystems
sudo umount -f /path/to/mount/point
fusermount3 -uz /path/to/mount/point

# Kill any stuck OneMount processes
pkill -f onemount

# Clear all cached data (last resort)
rm -rf ~/.config/onemount/
rm -rf ~/.cache/onemount/

# Reinstall OneMount
# Follow installation guide for your distribution
```

**Warning**: Clearing cached data will require re-downloading all files and re-authentication.
