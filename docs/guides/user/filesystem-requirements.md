# OneMount Filesystem Requirements

## Overview

This document describes the filesystem requirements for mounting OneMount and explains how file status tracking works.

## Filesystem Requirements

### Minimum Requirements

OneMount has **minimal filesystem requirements** and works on virtually any Linux filesystem:

✅ **Supported Filesystems**:
- ext4, ext3, ext2
- XFS, Btrfs, F2FS
- tmpfs, ramfs
- NFS, CIFS/SMB
- FUSE-based filesystems
- Any POSIX-compliant filesystem

✅ **No Special Features Required**:
- ❌ Extended attributes (xattr) support NOT required
- ❌ ACL support NOT required
- ❌ Special permissions NOT required
- ❌ Specific mount options NOT required

### Why No Xattr Support Required?

OneMount stores file status information **in-memory only**, not on the underlying filesystem. This design choice means:

1. **Universal Compatibility**: Works on all filesystem types
2. **No Configuration**: No special mount options needed
3. **No Permissions**: No special filesystem permissions required
4. **Consistent Behavior**: Same behavior on all filesystems

## File Status Tracking

### How It Works

OneMount tracks file status (synced, downloading, modified, etc.) using two mechanisms:

1. **In-Memory Extended Attributes**
   - Stored in RAM, not on disk
   - Accessible via FUSE xattr operations
   - Lost on unmount (by design)
   - No filesystem xattr support required

2. **D-Bus Signals** (optional)
   - Real-time status updates
   - Used by file manager extensions (Nemo, Nautilus)
   - Falls back to xattr-only if unavailable

### Status Information Available

Within the mounted filesystem, you can query file status using standard tools:

```bash
# Get file status
getfattr -n user.onemount.status /path/to/mounted/file

# Get error message (if any)
getfattr -n user.onemount.error /path/to/mounted/file

# List all OneMount attributes
getfattr -d -m "user.onemount.*" /path/to/mounted/file
```

**Status Values**:
- `cloud`: File exists in OneDrive only (not downloaded)
- `local`: File is cached locally
- `synced`: File is synchronized with OneDrive
- `downloading`: File is being downloaded
- `syncing`: File is being uploaded
- `modified`: File has local changes pending upload
- `outofSync`: File needs update from cloud
- `conflict`: File has conflicting local and remote changes
- `error`: File operation failed (check error attribute)

### File Manager Integration

OneMount integrates with file managers to display status icons:

**Nemo** (Linux Mint):
- Install: `python3-nemo` extension
- Location: `~/.local/share/nemo-python/extensions/`
- Status: Automatic (uses D-Bus signals)

**Nautilus** (GNOME):
- Install: `python3-nautilus` extension
- Location: `~/.local/share/nautilus-python/extensions/`
- Status: Automatic (uses D-Bus signals)

**Other File Managers**:
- May not show status icons
- File status still available via xattr queries
- All file operations work normally

## Behavior on Different Filesystems

### tmpfs / ramfs

**Status**: ✅ Fully Supported

```bash
# Mount on tmpfs
mkdir -p /tmp/onemount
onemount /tmp/onemount
```

**Notes**:
- Fast performance (RAM-based)
- Lost on reboot
- Good for temporary access
- No persistence required

### Network Filesystems (NFS, CIFS)

**Status**: ✅ Fully Supported

```bash
# Mount on NFS share
mkdir -p /mnt/nfs/onemount
onemount /mnt/nfs/onemount
```

**Notes**:
- Works without network filesystem xattr support
- Status information in-memory only
- May have higher latency
- Network filesystem errors handled gracefully

### Encrypted Filesystems

**Status**: ✅ Fully Supported

```bash
# Mount on encrypted filesystem
mkdir -p /home/user/encrypted/onemount
onemount /home/user/encrypted/onemount
```

**Notes**:
- Works with LUKS, eCryptfs, etc.
- No special configuration needed
- Encryption handled by underlying filesystem

### Read-Only Filesystems

**Status**: ⚠️ Not Supported

OneMount requires write access to the mount point for:
- Creating FUSE mount
- Managing cache
- Tracking file changes

## Troubleshooting

### Issue: "Cannot query file status"

**Symptom**: `getfattr` returns "No such attribute"

**Cause**: File status not yet initialized

**Solution**: Access the file first to initialize status:
```bash
ls -l /path/to/mounted/file
getfattr -n user.onemount.status /path/to/mounted/file
```

### Issue: "File manager doesn't show status icons"

**Symptom**: No sync status icons in file manager

**Possible Causes**:
1. File manager extension not installed
2. D-Bus service not running
3. File manager doesn't support extensions

**Solutions**:
```bash
# Check D-Bus service
busctl --user list | grep onemount

# Check file manager extension
ls ~/.local/share/nemo-python/extensions/
ls ~/.local/share/nautilus-python/extensions/

# Restart file manager
killall nemo  # or nautilus
nemo &        # or nautilus &
```

**Workaround**: Query status manually using `getfattr` (see above)

### Issue: "Status information lost after unmount"

**Symptom**: File status not preserved across unmounts

**Cause**: This is expected behavior (by design)

**Explanation**:
- Status information is in-memory only
- Lost on unmount (intentional)
- Regenerated on next mount
- No filesystem persistence required

**Impact**: None - status is recalculated on mount

### Issue: "Filesystem doesn't support xattr"

**Symptom**: Warning about xattr support

**Cause**: Misunderstanding of OneMount's design

**Explanation**:
- OneMount does NOT require filesystem xattr support
- Xattrs are in-memory only
- Works on all filesystems
- Warning can be ignored

## Performance Considerations

### Mount Point Location

**Best Performance**:
- Local SSD/NVMe: Fastest
- Local HDD: Good
- tmpfs/ramfs: Fastest (but not persistent)

**Acceptable Performance**:
- Network filesystems: Slower but functional
- Encrypted filesystems: Slight overhead

**Not Recommended**:
- Very slow storage (USB 1.0, old HDDs)
- High-latency network filesystems

### Cache Location

OneMount stores cache in `~/.cache/onemount/` by default.

**Recommendations**:
- Use fast local storage for cache
- Avoid network filesystems for cache
- Consider SSD for best performance

**Configuration**:
```bash
# Use custom cache location
onemount --cache-dir /path/to/fast/storage /mount/point
```

## Advanced Configuration

### Xattr Status in Statistics

Check xattr support status:

```bash
# Get filesystem statistics
onemount --stats /mount/point

# Look for:
# XAttrSupported: true (always true for in-memory xattrs)
```

### D-Bus Configuration

Configure D-Bus service name:

```bash
# Use custom D-Bus service name
onemount --dbus-name com.example.onemount /mount/point

# Disable D-Bus (xattr-only mode)
onemount --no-dbus /mount/point
```

## Summary

### Key Points

1. ✅ **No special filesystem requirements**
   - Works on all POSIX filesystems
   - No xattr support needed
   - No special permissions required

2. ✅ **File status always available**
   - In-memory xattr storage
   - Accessible via FUSE operations
   - D-Bus signals for real-time updates

3. ✅ **Universal compatibility**
   - tmpfs, network filesystems, encrypted filesystems
   - Consistent behavior everywhere
   - No configuration needed

4. ⚠️ **Status not persisted**
   - Lost on unmount (by design)
   - Regenerated on mount
   - No impact on functionality

### Getting Help

If you encounter issues:

1. Check this troubleshooting guide
2. Review logs: `journalctl --user -u onemount`
3. Enable debug logging: `onemount --log-level debug`
4. Report issues: https://github.com/auriora/onemount/issues

## See Also

- [Installation Guide](installation.md)
- [Configuration Guide](configuration.md)
- [Troubleshooting Guide](troubleshooting.md)
- [File Manager Integration](file-manager-integration.md)
