# OneMount CLI Troubleshooting Enhancements

**Date**: 2026-01-22  
**Status**: Proposal  
**Priority**: High  
**Related**: Task 45.4 - Authentication token path consistency audit

## Executive Summary

This proposal outlines enhancements to the OneMount CLI to improve troubleshooting capabilities and error messages. The goal is to make OneMount easier to debug and more user-friendly when things go wrong.

## Current Problems

### 1. Generic Error Messages
**Problem**: Errors don't provide actionable guidance
```
❌ Current: "Authentication failed"
✅ Better: "Authentication failed: token file not found at ~/.cache/onemount/home-user-OneDrive/auth_tokens.json
          
          To fix this:
          1. Run: onemount --auth-only /path/to/mountpoint
          2. Or copy existing tokens to the expected location"
```

### 2. No Built-in Diagnostics
**Problem**: Users can't easily check system health
- No way to verify FUSE is available
- No way to check authentication status
- No way to validate configuration
- No way to test network connectivity

### 3. Unclear Mount Failures
**Problem**: Mount failures don't explain root cause
```
❌ Current: "Mount failed"
✅ Better: "Mount failed: FUSE device not accessible
          
          Common causes:
          1. FUSE not installed: sudo apt install fuse3
          2. User not in fuse group: sudo usermod -a -G fuse $USER
          3. /dev/fuse permissions: ls -l /dev/fuse"
```

### 4. No Troubleshooting Mode
**Problem**: No verbose diagnostic output option
- Can't see what OneMount is trying to do
- Can't see which checks are passing/failing
- Can't see detailed error context

## Proposed Enhancements

### Enhancement 1: Add `--doctor` Command

A comprehensive system health check command:

```bash
onemount --doctor [mountpoint]
```

**Output Example**:
```
🏥 OneMount System Diagnostics
================================

✅ FUSE Support
   ✓ FUSE3 installed: /usr/bin/fusermount3
   ✓ FUSE device accessible: /dev/fuse
   ✓ User in fuse group: yes
   ✓ user_allow_other enabled: yes

✅ Authentication
   ✓ Token file exists: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json
   ✓ Token file permissions: 0600 (secure)
   ✓ Token format valid: yes
   ✓ Token expiration: 2026-01-23 14:30:00 (23 hours remaining)
   ✓ Account: user@example.com

✅ Network Connectivity
   ✓ DNS resolution: graph.microsoft.com
   ✓ HTTPS connectivity: https://graph.microsoft.com/v1.0/
   ✓ API authentication: valid access token
   ✓ Drive accessible: yes

✅ Cache Directory
   ✓ Cache directory exists: ~/.cache/onemount
   ✓ Cache directory writable: yes
   ✓ Available space: 45.2 GB
   ✓ Metadata database: 2.3 MB (healthy)

✅ Mount Point
   ✓ Mount point exists: /home/user/OneDrive
   ✓ Mount point is directory: yes
   ✓ Mount point is empty: yes
   ✓ Mount point writable: yes
   ✓ Not currently mounted: yes

📊 Summary: All checks passed! System is ready to mount.

To mount: onemount /home/user/OneDrive
```

**With Failures**:
```
🏥 OneMount System Diagnostics
================================

✅ FUSE Support
   ✓ FUSE3 installed
   ✓ FUSE device accessible

❌ Authentication
   ✗ Token file not found: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json
   
   To fix:
   1. Run authentication: onemount --auth-only /home/user/OneDrive
   2. Or set ONEMOUNT_AUTH_PATH: export ONEMOUNT_AUTH_PATH=/path/to/tokens

⚠️  Network Connectivity
   ✓ DNS resolution: graph.microsoft.com
   ✗ HTTPS connectivity: connection timeout
   
   Possible causes:
   1. No internet connection
   2. Firewall blocking HTTPS
   3. Proxy configuration needed

✅ Cache Directory
   ✓ All checks passed

❌ Mount Point
   ✗ Mount point not empty: /home/user/OneDrive (3 files)
   
   To fix:
   1. Choose an empty directory
   2. Or clear the directory: rm -rf /home/user/OneDrive/*

📊 Summary: 2 errors, 1 warning. Fix the issues above before mounting.
```

### Enhancement 2: Add `--verify` Command

Quick verification of existing mount:

```bash
onemount --verify /path/to/mountpoint
```

**Output Example**:
```
🔍 Verifying OneMount at /home/user/OneDrive
=============================================

✅ Mount Status
   ✓ Filesystem is mounted
   ✓ FUSE connection healthy
   ✓ Mount point accessible

✅ Authentication
   ✓ Token valid (expires in 23 hours)
   ✓ API access working

✅ Synchronization
   ✓ Delta sync active (last run: 2 minutes ago)
   ✓ Realtime notifications: connected
   ✓ Upload queue: 0 pending
   ✓ Download queue: 0 pending

✅ Cache Health
   ✓ Metadata cache: 1,234 items
   ✓ Content cache: 45 files (234 MB)
   ✓ No errors detected

📊 Summary: Mount is healthy and fully operational.
```

### Enhancement 3: Enhanced Error Messages

#### 3.1 Authentication Errors

**Current**:
```go
logging.Error().Err(err).Msg("Authentication failed")
```

**Enhanced**:
```go
func enhanceAuthError(err error, authPath string) error {
    if os.IsNotExist(err) {
        return fmt.Errorf(`authentication failed: token file not found

Token file: %s

This usually means:
1. You haven't authenticated yet
   Fix: onemount --auth-only %s

2. The token file was deleted
   Fix: Re-authenticate with the command above

3. Wrong cache directory
   Fix: Check --cache-dir flag or config file

For more help: onemount --doctor`, authPath, mountpoint)
    }
    
    if os.IsPermission(err) {
        return fmt.Errorf(`authentication failed: permission denied

Token file: %s
Current permissions: %s

Fix: chmod 600 %s

The token file must be readable only by you (0600 permissions).`, 
            authPath, getFilePerms(authPath), authPath)
    }
    
    // ... more specific error cases
}
```

#### 3.2 Mount Errors

**Current**:
```go
logging.Error().Msg("Mount failed")
```

**Enhanced**:
```go
func enhanceMountError(err error, mountpoint string) error {
    errStr := err.Error()
    
    if strings.Contains(errStr, "permission denied") {
        return fmt.Errorf(`mount failed: permission denied

Mount point: %s

Common causes:
1. FUSE not installed
   Fix: sudo apt install fuse3

2. User not in fuse group
   Fix: sudo usermod -a -G fuse $USER
   Then: Log out and log back in

3. /dev/fuse permissions
   Check: ls -l /dev/fuse
   Should be: crw-rw-rw- 1 root root

For more help: onemount --doctor %s`, mountpoint, mountpoint)
    }
    
    if strings.Contains(errStr, "already mounted") {
        return fmt.Errorf(`mount failed: mountpoint already in use

Mount point: %s

To fix:
1. Check what's mounted: findmnt %s
2. Unmount existing: fusermount3 -uz %s
3. Try mounting again

If unmount fails:
1. Check for processes: lsof %s
2. Kill processes if needed
3. Force unmount: sudo umount -l %s`, 
            mountpoint, mountpoint, mountpoint, mountpoint, mountpoint)
    }
    
    // ... more specific error cases
}
```

#### 3.3 Network Errors

**Current**:
```go
logging.Error().Err(err).Msg("Network error")
```

**Enhanced**:
```go
func enhanceNetworkError(err error) error {
    if isTimeoutError(err) {
        return fmt.Errorf(`network error: connection timeout

This usually means:
1. No internet connection
   Check: ping 8.8.8.8

2. Firewall blocking HTTPS
   Check: curl -I https://graph.microsoft.com

3. Proxy configuration needed
   Set: export HTTPS_PROXY=http://proxy:port

4. Microsoft Graph API is down
   Check: https://status.cloud.microsoft/

OneMount will continue in offline mode.
Cached files remain accessible.`)
    }
    
    if isDNSError(err) {
        return fmt.Errorf(`network error: DNS resolution failed

Cannot resolve: graph.microsoft.com

This usually means:
1. DNS server not configured
   Check: cat /etc/resolv.conf

2. Network not connected
   Check: ip addr show

3. VPN interfering with DNS
   Try: Disconnect VPN temporarily

OneMount will continue in offline mode.`)
    }
    
    // ... more specific error cases
}
```

### Enhancement 4: Add `--troubleshoot` Flag

Verbose diagnostic mode for debugging:

```bash
onemount --troubleshoot /path/to/mountpoint
```

**Output Example**:
```
🔧 OneMount Troubleshooting Mode
=================================

[00:00.001] Parsing command-line arguments...
[00:00.002] ✓ Mountpoint: /home/user/OneDrive
[00:00.003] ✓ Cache directory: ~/.cache/onemount

[00:00.010] Loading configuration...
[00:00.011] ✓ Config file: ~/.config/onemount/config.yml
[00:00.012] ✓ Log level: debug
[00:00.013] ✓ Delta interval: 300 seconds

[00:00.020] Checking FUSE availability...
[00:00.021] ✓ FUSE3 binary: /usr/bin/fusermount3
[00:00.022] ✓ FUSE device: /dev/fuse (accessible)
[00:00.023] ✓ User in fuse group: yes

[00:00.030] Loading authentication...
[00:00.031] ✓ Token file: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json
[00:00.032] ✓ Token format: valid JSON
[00:00.033] ✓ Account: user@example.com
[00:00.034] ✓ Expires: 2026-01-23 14:30:00 (23h remaining)

[00:00.040] Testing network connectivity...
[00:00.041] ✓ DNS: graph.microsoft.com → 20.190.151.7
[00:00.150] ✓ HTTPS: https://graph.microsoft.com/v1.0/ → 200 OK
[00:00.250] ✓ API auth: Bearer token accepted

[00:00.260] Initializing filesystem...
[00:00.261] ✓ Cache directory created
[00:00.262] ✓ Metadata database opened
[00:00.263] ✓ Content cache initialized

[00:00.270] Mounting filesystem...
[00:00.271] ✓ FUSE options: name=onemount, fsname=onemount
[00:00.272] ✓ Mount point validated
[00:00.350] ✓ FUSE server started

[00:00.360] Starting background services...
[00:00.361] ✓ Delta sync loop started
[00:00.362] ✓ Cache cleanup scheduled
[00:00.363] ✓ Upload manager started
[00:00.364] ✓ Download manager started

[00:00.370] ✅ Mount successful!

Filesystem mounted at: /home/user/OneDrive
Press Ctrl+C to unmount
```

### Enhancement 5: Add `--check-auth` Command

Quick authentication status check:

```bash
onemount --check-auth [mountpoint]
```

**Output Example**:
```
🔐 Authentication Status
========================

Token File: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json
Account: user@example.com
Status: ✅ Valid

Token Details:
  Created: 2026-01-22 14:30:00
  Expires: 2026-01-23 14:30:00
  Time Remaining: 23 hours 45 minutes
  
Access Token: ✅ Valid (tested against Microsoft Graph API)
Refresh Token: ✅ Present

Permissions:
  ✅ user.read
  ✅ files.readwrite.all
  ✅ offline_access

Next Steps:
  - Token will auto-refresh when it expires
  - No action needed

To re-authenticate: onemount --auth-only /home/user/OneDrive
```

### Enhancement 6: Add `--explain-error` Command

Explain common error codes:

```bash
onemount --explain-error <error-code>
```

**Examples**:
```bash
$ onemount --explain-error EACCES
Error Code: EACCES (Permission Denied)

Common Causes in OneMount:
1. FUSE device not accessible
   - User not in fuse group
   - /dev/fuse has wrong permissions

2. Mount point not writable
   - Directory owned by another user
   - Parent directory not writable

3. Token file not readable
   - Wrong file permissions (should be 0600)
   - File owned by another user

How to Fix:
1. Check FUSE access: ls -l /dev/fuse
2. Check mount point: ls -ld /path/to/mountpoint
3. Check token file: ls -l ~/.cache/onemount/*/auth_tokens.json

For more help: onemount --doctor
```

## Implementation Plan

### Phase 1: Core Diagnostics (Week 1)
- [ ] Implement `--doctor` command
- [ ] Add FUSE availability checks
- [ ] Add authentication validation
- [ ] Add network connectivity tests
- [ ] Add cache directory checks
- [ ] Add mount point validation

### Phase 2: Enhanced Errors (Week 1-2)
- [ ] Create error enhancement framework
- [ ] Enhance authentication errors
- [ ] Enhance mount errors
- [ ] Enhance network errors
- [ ] Enhance cache errors
- [ ] Add contextual help to all errors

### Phase 3: Additional Commands (Week 2)
- [ ] Implement `--verify` command
- [ ] Implement `--check-auth` command
- [ ] Implement `--troubleshoot` flag
- [ ] Implement `--explain-error` command

### Phase 4: Documentation (Week 2-3)
- [ ] Update man page with new commands
- [ ] Create troubleshooting guide
- [ ] Add error code reference
- [ ] Update README with diagnostic commands

## Code Structure

### New Files
```
cmd/onemount/
├── diagnostics.go      # --doctor implementation
├── verify.go           # --verify implementation
├── errors.go           # Enhanced error messages
└── troubleshoot.go     # --troubleshoot implementation

internal/diagnostics/
├── fuse.go            # FUSE checks
├── auth.go            # Auth validation
├── network.go         # Network tests
├── cache.go           # Cache checks
└── mount.go           # Mount validation
```

### Error Enhancement Pattern
```go
// internal/errors/enhanced.go
type EnhancedError struct {
    Original    error
    Context     string
    Causes      []string
    Fixes       []string
    MoreHelp    string
}

func (e *EnhancedError) Error() string {
    var b strings.Builder
    
    // Original error
    fmt.Fprintf(&b, "%s: %v\n\n", e.Context, e.Original)
    
    // Common causes
    if len(e.Causes) > 0 {
        b.WriteString("Common causes:\n")
        for i, cause := range e.Causes {
            fmt.Fprintf(&b, "%d. %s\n", i+1, cause)
        }
        b.WriteString("\n")
    }
    
    // How to fix
    if len(e.Fixes) > 0 {
        b.WriteString("To fix:\n")
        for i, fix := range e.Fixes {
            fmt.Fprintf(&b, "%d. %s\n", i+1, fix)
        }
        b.WriteString("\n")
    }
    
    // More help
    if e.MoreHelp != "" {
        fmt.Fprintf(&b, "For more help: %s\n", e.MoreHelp)
    }
    
    return b.String()
}
```

## Benefits

1. **Faster Troubleshooting**: Users can diagnose issues themselves
2. **Better Error Messages**: Clear guidance on how to fix problems
3. **Reduced Support Load**: Self-service diagnostics
4. **Improved UX**: More user-friendly CLI
5. **Better Debugging**: Verbose mode for developers

## Success Metrics

- ✅ All common errors have enhanced messages
- ✅ `--doctor` command catches 90%+ of configuration issues
- ✅ Error messages include actionable fixes
- ✅ Users can self-diagnose without documentation
- ✅ Support requests decrease by 50%+

## Examples of Enhanced Error Messages

### Before vs After

#### Authentication Error
**Before**:
```
ERROR: Authentication failed
```

**After**:
```
❌ Authentication failed: token file not found

Token file: ~/.cache/onemount/home-user-OneDrive/auth_tokens.json

This usually means:
1. You haven't authenticated yet
   Fix: onemount --auth-only /home/user/OneDrive

2. The token file was deleted
   Fix: Re-authenticate with the command above

3. Wrong cache directory
   Fix: Check --cache-dir flag or config file

For more help: onemount --doctor /home/user/OneDrive
```

#### Mount Error
**Before**:
```
ERROR: Mount failed
```

**After**:
```
❌ Mount failed: FUSE device not accessible

Mount point: /home/user/OneDrive

Common causes:
1. FUSE not installed
   Fix: sudo apt install fuse3

2. User not in fuse group
   Fix: sudo usermod -a -G fuse $USER
   Then: Log out and log back in

3. /dev/fuse permissions incorrect
   Check: ls -l /dev/fuse
   Should be: crw-rw-rw- 1 root root

For more help: onemount --doctor /home/user/OneDrive
```

## Conclusion

These enhancements will make OneMount significantly easier to troubleshoot and use. The `--doctor` command provides comprehensive diagnostics, while enhanced error messages guide users to solutions. Together, these improvements will reduce support burden and improve user satisfaction.

**Key Principle**: *Every error message should tell the user exactly how to fix the problem.*
