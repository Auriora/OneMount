# Troubleshooting Extended Attribute (Xattr) Issues

## Overview

This guide helps troubleshoot issues related to extended attributes (xattrs) in OneMount.

## Important: OneMount's Xattr Design

**Key Fact**: OneMount stores xattrs **in-memory only**, not on the underlying filesystem.

This means:
- ✅ No filesystem xattr support required
- ✅ Works on all filesystem types
- ✅ No xattr-related errors from filesystem
- ⚠️ Xattrs lost on unmount (by design)

## Common Issues

### 1. "No such attribute" Error

**Symptom**:
```bash
$ getfattr -n user.onemount.status /mnt/onedrive/file.txt
/mnt/onedrive/file.txt: user.onemount.status: No such attribute
```

**Cause**: File status not yet initialized

**Solution**:
```bash
# Access the file first to initialize status
ls -l /mnt/onedrive/file.txt

# Then query status
getfattr -n user.onemount.status /mnt/onedrive/file.txt
```

**Why**: Status is computed on-demand when file is accessed

---

### 2. "Operation not supported" Error

**Symptom**:
```bash
$ setfattr -n user.test -v "value" /mnt/onedrive/file.txt
setfattr: /mnt/onedrive/file.txt: Operation not supported
```

**Cause**: Trying to set custom xattrs (not supported)

**Explanation**:
- OneMount only supports `user.onemount.*` xattrs
- Custom xattrs are not supported
- This is intentional (status tracking only)

**Solution**: Use OneMount's status xattrs only:
```bash
# Supported xattrs
getfattr -n user.onemount.status /mnt/onedrive/file.txt
getfattr -n user.onemount.error /mnt/onedrive/file.txt
```

---

### 3. Xattrs Lost After Unmount

**Symptom**: File status information disappears after unmounting

**Cause**: This is expected behavior (by design)

**Explanation**:
- Xattrs are in-memory only
- Not persisted to filesystem
- Regenerated on next mount

**Impact**: None - status is recalculated automatically

**Example**:
```bash
# Before unmount
$ getfattr -n user.onemount.status /mnt/onedrive/file.txt
user.onemount.status="local"

# Unmount
$ fusermount3 -u /mnt/onedrive

# Remount
$ onemount /mnt/onedrive

# Status regenerated
$ getfattr -n user.onemount.status /mnt/onedrive/file.txt
user.onemount.status="local"
```

---

### 4. File Manager Not Showing Status Icons

**Symptom**: No sync status icons in Nemo/Nautilus

**Possible Causes**:
1. File manager extension not installed
2. D-Bus service not running
3. File manager doesn't support Python extensions

**Diagnosis**:
```bash
# Check D-Bus service
busctl --user list | grep onemount

# Check extension installation
ls ~/.local/share/nemo-python/extensions/onemount*
ls ~/.local/share/nautilus-python/extensions/onemount*

# Check if file manager supports extensions
nemo --version  # Should show Python support
```

**Solutions**:

**For Nemo**:
```bash
# Install Python support
sudo apt install python3-nemo

# Install extension
cp extensions/nemo/onemount-nemo.py ~/.local/share/nemo-python/extensions/

# Restart Nemo
killall nemo
nemo &
```

**For Nautilus**:
```bash
# Install Python support
sudo apt install python3-nautilus

# Install extension
cp extensions/nautilus/onemount-nautilus.py ~/.local/share/nautilus-python/extensions/

# Restart Nautilus
killall nautilus
nautilus &
```

**Workaround**: Query status manually:
```bash
getfattr -n user.onemount.status /mnt/onedrive/file.txt
```

---

### 5. "Filesystem doesn't support xattr" Warning

**Symptom**: Warning message about xattr support

**Cause**: Misunderstanding of OneMount's design

**Explanation**:
- OneMount does NOT require filesystem xattr support
- Xattrs are in-memory only
- Works on all filesystems (tmpfs, NFS, etc.)
- Warning can be safely ignored

**Example**:
```bash
# Works fine on tmpfs (no xattr support)
mkdir /tmp/onemount
onemount /tmp/onemount

# Xattrs still work (in-memory)
getfattr -n user.onemount.status /tmp/onemount/file.txt
```

---

### 6. Xattrs Not Visible Outside Mount Point

**Symptom**: Cannot query xattrs on underlying filesystem

**Example**:
```bash
# Inside mount point - works
$ getfattr -n user.onemount.status /mnt/onedrive/file.txt
user.onemount.status="local"

# Unmount
$ fusermount3 -u /mnt/onedrive

# Outside mount point - doesn't work
$ getfattr -n user.onemount.status /mnt/onedrive/file.txt
/mnt/onedrive/file.txt: user.onemount.status: No such attribute
```

**Cause**: Xattrs are in-memory only (FUSE layer)

**Explanation**:
- Xattrs only exist while mounted
- Not written to underlying filesystem
- This is intentional design

**Impact**: None - this is expected behavior

---

## Debugging Xattr Issues

### Enable Debug Logging

```bash
# Start OneMount with debug logging
onemount --log-level debug /mnt/onedrive

# Watch logs
journalctl --user -u onemount -f
```

**Look for**:
- "Initialized in-memory xattrs map"
- "Updated file status xattrs"
- "In-memory extended attributes initialized"

### Check Xattr Support Status

```bash
# Get filesystem statistics
onemount --stats /mnt/onedrive

# Look for:
# XAttrSupported: true (should always be true)
```

### Test Xattr Operations

```bash
# List all xattrs
getfattr -d /mnt/onedrive/file.txt

# Get specific xattr
getfattr -n user.onemount.status /mnt/onedrive/file.txt

# Get all OneMount xattrs
getfattr -d -m "user.onemount.*" /mnt/onedrive/file.txt
```

### Verify D-Bus Integration

```bash
# Check D-Bus service
busctl --user list | grep onemount

# Monitor D-Bus signals
dbus-monitor --session "type='signal',interface='com.github.auriora.onemount.FileStatus'"

# Test D-Bus method call
busctl --user call com.github.auriora.onemount \
  /com/github/auriora/onemount \
  com.github.auriora.onemount.FileStatus \
  GetFileStatus s "/mnt/onedrive/file.txt"
```

---

## Advanced Troubleshooting

### Xattr Performance Issues

**Symptom**: Slow file status queries

**Diagnosis**:
```bash
# Time xattr query
time getfattr -n user.onemount.status /mnt/onedrive/file.txt

# Should be < 10ms
```

**Possible Causes**:
1. Status determination is expensive (hash calculation)
2. Database queries are slow
3. Cache is not working

**Solutions**:
```bash
# Check cache status
onemount --stats /mnt/onedrive | grep -i cache

# Increase cache TTL (if needed)
# Edit configuration file

# Enable status cache
# (should be enabled by default)
```

### Memory Usage

**Symptom**: High memory usage

**Diagnosis**:
```bash
# Check process memory
ps aux | grep onemount

# Check xattr map size
# (requires debug build)
```

**Explanation**:
- Each inode has an xattrs map
- Typically 2-3 xattrs per file
- Small memory footprint (< 100 bytes per file)

**Expected Memory**:
- 1,000 files: ~100 KB
- 10,000 files: ~1 MB
- 100,000 files: ~10 MB

---

## Best Practices

### 1. Query Status Efficiently

**Good**:
```bash
# Query status once
status=$(getfattr --only-values -n user.onemount.status file.txt)
echo "Status: $status"
```

**Bad**:
```bash
# Multiple queries (slower)
getfattr -n user.onemount.status file.txt
getfattr -n user.onemount.error file.txt
getfattr -n user.onemount.status file.txt  # duplicate
```

### 2. Use D-Bus for Real-Time Updates

**Good**:
```bash
# Monitor D-Bus signals for real-time updates
dbus-monitor --session "type='signal',interface='com.github.auriora.onemount.FileStatus'"
```

**Bad**:
```bash
# Polling xattrs (inefficient)
while true; do
  getfattr -n user.onemount.status file.txt
  sleep 1
done
```

### 3. Batch Status Queries

**Good**:
```bash
# Query multiple files efficiently
for file in /mnt/onedrive/*; do
  getfattr -n user.onemount.status "$file"
done
```

**Better**:
```bash
# Use D-Bus batch query (if available)
# (future enhancement)
```

---

## When to Report Issues

Report an issue if:
- ✅ Xattr queries hang or timeout
- ✅ Status information is incorrect
- ✅ Memory usage is excessive
- ✅ D-Bus integration fails
- ✅ File manager extension crashes

Do NOT report if:
- ❌ Xattrs lost on unmount (expected)
- ❌ Cannot set custom xattrs (not supported)
- ❌ Xattrs not visible outside mount (expected)
- ❌ "Filesystem doesn't support xattr" warning (can be ignored)

---

## Getting Help

If you need help:

1. **Check this guide** for common issues
2. **Enable debug logging** and review logs
3. **Test xattr operations** manually
4. **Check D-Bus integration** if using file manager
5. **Report issue** with logs and reproduction steps

**Report Issues**:
- GitHub: https://github.com/auriora/onemount/issues
- Include: logs, reproduction steps, system info

---

## See Also

- [Filesystem Requirements](../user/filesystem-requirements.md)
- [File Manager Integration](../user/file-manager-integration.md)
- [Configuration Guide](../user/configuration.md)
- [Troubleshooting Guide](troubleshooting.md)
