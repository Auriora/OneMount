# Systemd User Service Fix

**Date**: 2026-01-27  
**Issue**: Launcher unable to start systemd services from .deb package installation  
**Status**: Fixed

## Problem

After installing from the .deb package, the launcher failed with:
```
Failed to start unit. error="Unit onemount@home-bcherrington-onmount\\x2dauriora.service not found."
```

## Root Cause

The service was installed as a **system service** (`/usr/lib/systemd/system/`) but the launcher was trying to start it via the **user session bus**. System services require root/polkit authorization to start, which is inappropriate for a user-facing GUI application.

## Solution

Changed the package to install the service as a **user service** instead:

1. **Install location**: Changed from `/usr/lib/systemd/system/` to `/usr/lib/systemd/user/`
2. **Service configuration**: Removed `User=%i` and `Group=%i` directives (not needed for user services)
3. **Post-install script**: Updated to reload user systemd daemons for all logged-in users
4. **Launcher code**: Simplified to always use session bus (appropriate for user services)

## Changes Made

### packaging/install-manifest.json
- Changed `dest_package` from `usr/lib/systemd/system/` to `usr/lib/systemd/user/`
- Updated substitutions to remove User/Group directives
- Changed `@BIN_PATH@` to `/usr/bin` for package installations
- Updated directories list

### packaging/ubuntu/onemount.postinst
- Changed from `systemctl daemon-reload` to user daemon reload
- Iterates through logged-in users and reloads their systemd user instances

### internal/ui/systemd/systemd.go
- Simplified `getSystemdConnection()` to always use session bus
- Removed system bus detection logic (not needed for user services)

## User vs System Services

### User Services (`/usr/lib/systemd/user/`)
- ✅ Can be started by any user without root
- ✅ Run in user's session context
- ✅ Appropriate for user-facing applications
- ✅ Automatically cleaned up when user logs out
- ❌ Not available at boot (only after user login)

### System Services (`/usr/lib/systemd/system/`)
- ✅ Available at boot time
- ✅ Can run as any user via User= directive
- ❌ Require root/polkit to start/stop
- ❌ Inappropriate for GUI applications
- ❌ Complex permission management

## Verification

After installing the updated package:

```bash
# Check service is installed
ls -la /usr/lib/systemd/user/onemount@.service

# Reload user daemon
systemctl --user daemon-reload

# Start a mount
onemount-launcher
# Click "Create mountpoint" - should work without errors

# Verify service is running
systemctl --user list-units 'onemount@*.service'
```

## References

- [systemd User Services](https://www.freedesktop.org/software/systemd/man/systemd.unit.html)
- [D-Bus Session vs System Bus](https://dbus.freedesktop.org/doc/dbus-daemon.1.html)
- [Debian systemd Integration](https://wiki.debian.org/systemd)
