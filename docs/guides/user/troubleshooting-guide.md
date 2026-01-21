# OneMount Troubleshooting Guide

## Table of Contents
1. [Common Issues](#common-issues)
2. [Installation Problems](#installation-problems)
3. [Authentication Issues](#authentication-issues)
4. [Network and Connectivity](#network-and-connectivity)
5. [File Operation Problems](#file-operation-problems)
   - [Filesystem Requirements and Extended Attributes](#filesystem-requirements-and-extended-attributes)
6. [Performance Issues](#performance-issues)
7. [Socket.IO Realtime Notification Issues](#socketio-realtime-notification-issues)
8. [Cache Management Issues](#cache-management-issues)
9. [Upload and Download Issues](#upload-and-download-issues)
10. [Offline Mode Issues](#offline-mode-issues)
11. [Debugging and Logging](#debugging-and-logging)
12. [Getting Help](#getting-help)

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

### Socket.IO Realtime Issues

**Symptoms:**
- Changes in OneDrive don't appear immediately locally
- High polling frequency despite realtime being enabled
- "Socket.IO connection failed" messages in logs

**Causes:**
- WebSocket connections blocked by firewall/proxy
- Network doesn't support WebSocket protocols
- Corporate network restrictions

**Solutions:**
1. **Check Socket.IO status:**
   ```bash
   # View realtime notification status
   onemount --stats /path/to/mount/point | grep -A 10 "Realtime"
   ```

2. **Test WebSocket connectivity:**
   ```bash
   # Test if WebSocket connections work
   curl -I https://graph.microsoft.com/v1.0/subscriptions/socketIo
   ```

3. **Force polling-only mode:**
   ```bash
   # Disable Socket.IO and use polling only
   onemount --polling-only /path/to/mount/point
   ```

4. **Configure polling-only in config file:**
   ```yaml
   # ~/.config/onemount/config.yml
   realtime:
     enabled: true
     pollingOnly: true
     fallbackIntervalSeconds: 300  # 5 minutes
   ```

5. **Check firewall/proxy settings:**
   - Ensure WebSocket (WSS) traffic is allowed to `*.graph.microsoft.com`
   - Configure proxy settings if needed
   - Contact network administrator about WebSocket support

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

### Filesystem Requirements and Extended Attributes

**Symptoms:**
- File status icons don't appear in file manager
- Warnings about extended attributes in logs
- Some features not working as expected

**Background:**
OneMount uses extended attributes (xattrs) to store file status information that can be displayed by file managers like Nemo or Nautilus. Not all filesystems support extended attributes.

**Supported Filesystems:**
- ✅ **ext4** (recommended) - Full support for extended attributes
- ✅ **ext3** - Full support for extended attributes
- ✅ **XFS** - Full support for extended attributes
- ✅ **Btrfs** - Full support for extended attributes
- ⚠️ **tmpfs** - Limited or no extended attribute support
- ⚠️ **NFS** - May not support extended attributes depending on configuration
- ⚠️ **FAT32/exFAT** - No extended attribute support

**Check Extended Attribute Support:**
```bash
# Check if your filesystem supports extended attributes
getfattr --version

# Test extended attributes on your mount point
touch /path/to/mount/point/test_file
setfattr -n user.test -v "test_value" /path/to/mount/point/test_file
getfattr -n user.test /path/to/mount/point/test_file
rm /path/to/mount/point/test_file
```

**Check OneMount's Extended Attribute Status:**
```bash
# View statistics including xattr support status
onemount --stats /path/to/mount/point | grep -i xattr
```

**Solutions:**
1. **If extended attributes are not supported:**
   - OneMount will continue to work normally
   - File status information will still be available via D-Bus
   - File manager extensions may not show status icons
   - Consider using a filesystem that supports extended attributes (e.g., ext4)

2. **If you need file status icons:**
   - Ensure your mount point is on a filesystem that supports extended attributes
   - Verify the Nemo/Nautilus extension is installed
   - Check that D-Bus is available and working

3. **Verify D-Bus is working:**
   ```bash
   # Check if D-Bus session is available
   echo $DBUS_SESSION_BUS_ADDRESS
   
   # Monitor D-Bus signals from OneMount
   dbus-monitor --session "interface='com.github.jstaf.onedriver.FileStatus'"
   ```

**Note:** Extended attributes are optional. OneMount will work without them, but some visual features (like file status icons in file managers) may not be available.

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

### Slow Directory Listings (Large Directories)

**Symptoms:**
- Listing directories with many files (>1000) takes more than 2 seconds
- File manager becomes unresponsive when browsing large directories
- High memory usage during directory operations

**Causes:**
- Large number of files in a single directory
- Metadata cache not optimized for large datasets
- Statistics collection overhead

**Solutions:**
1. **Check directory size:**
   ```bash
   # Count files in directory
   ls -1 /path/to/mount/point/large_directory | wc -l
   ```

2. **Monitor performance:**
   ```bash
   # Time directory listing
   time ls -la /path/to/mount/point/large_directory
   ```

3. **Optimize cache:**
   ```bash
   # Ensure cache is properly configured
   onemount --stats /path/to/mount/point | grep -A 5 "Cache"
   ```

4. **Workaround for very large directories:**
   - Organize files into subdirectories (recommended: <500 files per directory)
   - Use command-line tools instead of GUI file managers for large directories
   - Consider using `ls` with specific patterns instead of listing all files

### High Memory Usage

**Symptoms:**
- OneMount uses more than 200 MB of memory during active sync
- Memory usage grows over time
- System becomes slow or unresponsive

**Causes:**
- Large file uploads/downloads in progress
- Many concurrent operations
- Cache size not limited
- Memory leaks (rare)

**Solutions:**
1. **Check current memory usage:**
   ```bash
   # Monitor OneMount memory usage
   ps aux | grep onemount | awk '{print $6/1024 " MB"}'
   ```

2. **Configure cache size limits:**
   ```yaml
   # ~/.config/onemount/config.yml
   cache:
     maxSizeMB: 5000  # Limit cache to 5 GB
     expirationDays: 30
   ```

3. **Restart OneMount if memory usage is excessive:**
   ```bash
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

4. **Monitor for memory leaks:**
   ```bash
   # Watch memory usage over time
   watch -n 5 'ps aux | grep onemount'
   ```

### Slow Statistics Collection

**Symptoms:**
- `onemount --stats` command takes a long time to complete
- High CPU usage when viewing statistics
- File manager status queries are slow

**Causes:**
- Large number of files in filesystem (>100,000)
- Statistics calculated on-demand without caching
- Database queries not optimized

**Solutions:**
1. **Use basic stats only:**
   ```bash
   # Get quick overview without detailed statistics
   onemount --stats /path/to/mount/point | head -20
   ```

2. **Avoid frequent stats queries:**
   - Don't poll statistics continuously
   - Use longer intervals between checks (>60 seconds)

3. **Optimize database:**
   ```bash
   # Unmount and remount to rebuild indexes
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

## Socket.IO Realtime Notification Issues

### Understanding Realtime Modes

OneMount supports three realtime notification modes:

1. **Socket.IO Mode** (default when enabled): Uses WebSocket connections for instant notifications
2. **Polling-Only Mode**: Regular polling without WebSocket connections  
3. **Disabled Mode**: Minimal polling for basic functionality

### Checking Realtime Status

```bash
# View current realtime configuration and status
onemount --stats /path/to/mount/point

# Look for the "Realtime Notifications" section:
# Mode: socketio | polling | disabled
# Status: healthy | degraded | failed | unknown
# Last heartbeat: timestamp of last Socket.IO activity
```

### Common Realtime Issues

**Issue: Changes don't appear immediately**
- Check if realtime is enabled: `realtime.enabled: true` in config
- Verify Socket.IO status is "healthy" in stats output
- If status is "failed", consider using polling-only mode

**Issue: High CPU/network usage**
- Check polling interval: should be 1800s (30 min) with healthy Socket.IO
- If polling every few minutes, Socket.IO may be failing
- Enable polling-only mode to reduce connection attempts

**Issue: Corporate network blocks WebSocket**
- Use polling-only mode: `pollingOnly: true` in config
- Set reasonable polling interval: 300-900 seconds (5-15 minutes)
- Contact IT about allowing WebSocket traffic to Microsoft Graph

### Configuration Examples

**Optimal for home networks:**
```yaml
realtime:
  enabled: true
  pollingOnly: false
  fallbackIntervalSeconds: 1800  # 30 minutes
```

**Optimal for corporate networks:**
```yaml
realtime:
  enabled: true
  pollingOnly: true
  fallbackIntervalSeconds: 600   # 10 minutes
```

**Minimal resource usage:**
```yaml
realtime:
  enabled: false
```

### Socket.IO Connection Diagnostics

**Check Socket.IO health:**
```bash
# View detailed Socket.IO status
onemount --stats /path/to/mount/point | grep -A 20 "Realtime"

# Expected output for healthy connection:
# Mode: socketio
# Status: healthy
# Last heartbeat: <recent timestamp>
# Reconnect attempts: 0
```

**Common Socket.IO error patterns:**

1. **Connection refused:**
   ```
   Error: dial tcp: connection refused
   ```
   - **Cause**: Firewall blocking WebSocket connections
   - **Solution**: Enable polling-only mode or configure firewall

2. **Handshake timeout:**
   ```
   Error: Engine.IO handshake timeout
   ```
   - **Cause**: Network latency or proxy issues
   - **Solution**: Check network connectivity, configure proxy settings

3. **Heartbeat missed:**
   ```
   Warning: 2 consecutive heartbeats missed
   ```
   - **Cause**: Network instability or high latency
   - **Solution**: System will automatically fall back to polling

4. **Authentication failed:**
   ```
   Error: 401 Unauthorized
   ```
   - **Cause**: Token expired or invalid
   - **Solution**: Re-authenticate with `onemount --auth-only`

**Enable Socket.IO debug logging:**
```bash
# Set log level to debug for detailed Socket.IO logs
ONEMOUNT_LOG_LEVEL=debug onemount /path/to/mount/point

# Look for Socket.IO-specific log entries:
# - "Engine.IO handshake completed"
# - "Socket.IO ping/pong timing"
# - "Reconnection attempt with backoff"
```

**Test WebSocket connectivity manually:**
```bash
# Test if WebSocket connections work to Microsoft Graph
curl -I https://graph.microsoft.com/v1.0/subscriptions/socketIo

# Expected: HTTP 200 or 401 (authentication required)
# If connection fails, WebSocket may be blocked
```

## Cache Management Issues

### Cache Not Invalidating on Remote Changes

**Symptoms:**
- Old file content served even after remote changes
- File modifications in OneDrive web interface not reflected locally
- Stale data persists after delta sync

**Causes:**
- Cache invalidation not triggered on ETag changes
- Delta sync not detecting remote modifications
- Cache cleanup not running

**Solutions:**
1. **Force cache invalidation:**
   ```bash
   # Unmount and remount to trigger full sync
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

2. **Check delta sync status:**
   ```bash
   # View last sync time and delta link
   onemount --stats /path/to/mount/point | grep -A 5 "Delta Sync"
   ```

3. **Clear specific file from cache:**
   ```bash
   # Remove file to force re-download
   rm /path/to/mount/point/file.txt
   # Access file again to re-download
   cat /path/to/mount/point/file.txt
   ```

4. **Enable explicit cache invalidation:**
   - This is handled automatically by delta sync
   - If issues persist, check logs for delta sync errors

### Cache Growing Too Large

**Symptoms:**
- Cache directory consuming excessive disk space
- Disk space warnings
- Cache size exceeds configured limits

**Causes:**
- No cache size limit configured
- Cache cleanup not running
- Many large files downloaded

**Solutions:**
1. **Check current cache size:**
   ```bash
   # View cache statistics
   onemount --stats /path/to/mount/point | grep -A 10 "Cache"
   
   # Check cache directory size
   du -sh ~/.cache/onemount/
   ```

2. **Configure cache size limit:**
   ```yaml
   # ~/.config/onemount/config.yml
   cache:
     maxSizeMB: 5000  # 5 GB limit
     expirationDays: 30
     cleanupIntervalHours: 24
   ```

3. **Manually trigger cache cleanup:**
   ```bash
   # Unmount and remount to trigger cleanup
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

4. **Clear entire cache (last resort):**
   ```bash
   # WARNING: This will delete all cached files
   rm -rf ~/.cache/onemount/*
   ```

### Cache Cleanup Not Running

**Symptoms:**
- Old files not being removed from cache
- Cache size continues to grow
- Expired files still present

**Causes:**
- Cache cleanup interval too long (default: 24 hours)
- Cleanup disabled (expirationDays <= 0)
- OneMount not running long enough for cleanup to trigger

**Solutions:**
1. **Check cleanup configuration:**
   ```yaml
   # ~/.config/onemount/config.yml
   cache:
     expirationDays: 30  # Must be > 0
     cleanupIntervalHours: 24  # Adjust as needed
   ```

2. **Verify cleanup is enabled:**
   ```bash
   # Check logs for cleanup activity
   journalctl --user -u onemount@* | grep -i "cache cleanup"
   ```

3. **Force immediate cleanup:**
   ```bash
   # Unmount and remount to trigger cleanup
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

4. **Configure more frequent cleanup:**
   ```yaml
   # ~/.config/onemount/config.yml
   cache:
     cleanupIntervalHours: 6  # Run every 6 hours
   ```

## Upload and Download Issues

### Large File Upload Failures

**Symptoms:**
- Files larger than 250 MB fail to upload
- Upload progress stops or hangs
- "Upload failed" errors in logs

**Causes:**
- Network interruptions during chunked upload
- Upload session timeout
- Retry logic not working correctly

**Solutions:**
1. **Check upload status:**
   ```bash
   # View upload queue and status
   onemount --stats /path/to/mount/point | grep -A 10 "Upload"
   ```

2. **Monitor upload progress:**
   ```bash
   # Enable debug logging to see upload details
   ONEMOUNT_DEBUG=1 onemount /path/to/mount/point
   ```

3. **Verify network stability:**
   ```bash
   # Test sustained connection to Microsoft Graph
   ping -c 100 graph.microsoft.com
   ```

4. **Retry failed upload:**
   ```bash
   # Unmount and remount to retry uploads
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

5. **Check upload retry configuration:**
   ```yaml
   # ~/.config/onemount/config.yml
   upload:
     maxRetries: 5  # Increase if needed
     retryDelaySeconds: 2
   ```

### Upload Max Retries Exceeded

**Symptoms:**
- "Max retries exceeded" errors in logs
- Files stuck in upload queue
- Upload status shows "Error"

**Causes:**
- Persistent network issues
- Server-side errors (500, 503)
- File conflicts or permission issues

**Solutions:**
1. **Check error details:**
   ```bash
   # View detailed error messages
   journalctl --user -u onemount@* | grep -i "upload.*error"
   ```

2. **Verify file permissions:**
   - Check if file can be uploaded via OneDrive web interface
   - Ensure account has write permissions

3. **Clear upload queue:**
   ```bash
   # Remove file and re-add to reset upload state
   mv /path/to/mount/point/file.txt /tmp/file.txt
   # Wait a moment
   mv /tmp/file.txt /path/to/mount/point/file.txt
   ```

4. **Increase retry limit:**
   ```yaml
   # ~/.config/onemount/config.yml
   upload:
     maxRetries: 10  # Increase for unreliable networks
   ```

### Download Manager Memory Usage

**Symptoms:**
- High memory usage during downloads
- System slowdown when downloading large files
- Out of memory errors

**Causes:**
- Large files loaded entirely into memory
- Multiple concurrent downloads
- No streaming for large files

**Solutions:**
1. **Limit concurrent downloads:**
   ```yaml
   # ~/.config/onemount/config.yml
   download:
     workerPoolSize: 2  # Reduce from default 3
     queueSize: 100     # Reduce from default 500
   ```

2. **Monitor memory during downloads:**
   ```bash
   # Watch memory usage
   watch -n 2 'ps aux | grep onemount | awk "{print \$6/1024 \" MB\"}"'
   ```

3. **Download large files one at a time:**
   - Avoid opening multiple large files simultaneously
   - Wait for one download to complete before starting another

4. **Restart OneMount if memory usage is excessive:**
   ```bash
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

## Offline Mode Issues

### Offline Detection False Positives

**Symptoms:**
- OneMount switches to offline mode when network is available
- "Network disconnected" notifications when online
- Frequent offline/online transitions

**Causes:**
- Overly conservative offline detection
- Authentication errors misinterpreted as network errors
- Temporary network glitches

**Solutions:**
1. **Check actual network connectivity:**
   ```bash
   # Test Microsoft Graph API connectivity
   curl -I https://graph.microsoft.com/v1.0/me
   ```

2. **Review offline detection logs:**
   ```bash
   # Check what triggered offline mode
   journalctl --user -u onemount@* | grep -i "offline"
   ```

3. **Verify authentication is valid:**
   ```bash
   # Re-authenticate if needed
   onemount --auth-only /path/to/mount/point
   ```

4. **Check for permission errors:**
   ```bash
   # Look for 401/403 errors that shouldn't trigger offline mode
   journalctl --user -u onemount@* | grep -E "(401|403|permission)"
   ```

### Offline Mode Not Detected

**Symptoms:**
- Network disconnected but OneMount doesn't switch to offline mode
- Operations hang instead of failing gracefully
- No offline notifications

**Causes:**
- Offline detection not working
- Network errors not matching known patterns
- Connectivity check disabled

**Solutions:**
1. **Enable connectivity checks:**
   ```yaml
   # ~/.config/onemount/config.yml
   offline:
     connectivityCheckInterval: 15  # seconds
     connectivityTimeout: 10        # seconds
   ```

2. **Test offline detection:**
   ```bash
   # Disconnect network and trigger an operation
   # Should see offline mode activation in logs
   journalctl --user -u onemount@* -f
   ```

3. **Check network error patterns:**
   ```bash
   # View errors that should trigger offline mode
   journalctl --user -u onemount@* | grep -E "(no such host|network unreachable|connection refused)"
   ```

### Offline Changes Not Syncing

**Symptoms:**
- Changes made while offline don't upload when back online
- Files modified offline show old content
- Offline change queue not processing

**Causes:**
- Change tracking not working
- Upload queue not processing
- Conflicts preventing sync

**Solutions:**
1. **Check offline change queue:**
   ```bash
   # View pending offline changes
   onemount --stats /path/to/mount/point | grep -A 10 "Offline"
   ```

2. **Force sync after coming online:**
   ```bash
   # Unmount and remount to trigger sync
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

3. **Check for conflicts:**
   ```bash
   # Look for conflict files
   find /path/to/mount/point -name "*conflict*" -type f
   ```

4. **Review sync errors:**
   ```bash
   # Check for upload errors
   journalctl --user -u onemount@* | grep -i "sync.*error"
   ```

### Offline Change Queue Limit Reached

**Symptoms:**
- "Maximum pending changes limit reached" errors
- Cannot make more changes while offline
- Some offline changes not tracked

**Causes:**
- Too many changes made while offline
- Queue limit too low (default: 1000)
- Changes not being processed

**Solutions:**
1. **Check queue status:**
   ```bash
   # View current queue size
   onemount --stats /path/to/mount/point | grep -i "pending changes"
   ```

2. **Increase queue limit:**
   ```yaml
   # ~/.config/onemount/config.yml
   offline:
     maxPendingChanges: 5000  # Increase from default 1000
   ```

3. **Process changes immediately when online:**
   ```bash
   # Ensure network is available and trigger sync
   fusermount3 -uz /path/to/mount/point
   onemount /path/to/mount/point
   ```

4. **Prioritize important changes:**
   - Sync smaller batches of changes
   - Avoid making thousands of changes while offline

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

# Cache and statistics
onemount --stats /path/to/mount/point

# Recent logs
journalctl --user -u onemount@* --since "1 hour ago" --no-pager

# Check for errors
journalctl --user -u onemount@* | grep -i error | tail -20

# Check Socket.IO status
journalctl --user -u onemount@* | grep -i "socket.io\|engine.io" | tail -20

# Check offline mode transitions
journalctl --user -u onemount@* | grep -i "offline\|online" | tail -20
```

### Common Diagnostic Commands

**Check file status:**
```bash
# View extended attributes (file status)
getfattr -d /path/to/mount/point/file.txt

# Check if file is cached
ls -la ~/.cache/onemount/content/
```

**Monitor real-time activity:**
```bash
# Watch OneMount logs in real-time
journalctl --user -u onemount@* -f

# Monitor D-Bus signals
dbus-monitor --session "interface='com.github.jstaf.onedriver.FileStatus'"

# Watch network activity
sudo tcpdump -i any host graph.microsoft.com
```

**Check database state:**
```bash
# View metadata database location
ls -lh ~/.config/onemount/*.db

# Check database size
du -sh ~/.config/onemount/
```

**Test specific operations:**
```bash
# Test file read
time cat /path/to/mount/point/test.txt > /dev/null

# Test file write
echo "test" > /path/to/mount/point/test.txt

# Test directory listing
time ls -la /path/to/mount/point/

# Test file metadata
stat /path/to/mount/point/test.txt
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
