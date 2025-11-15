# Launcher Icon Window List Fix

**Date**: 2025-11-14  
**Type**: Bugfix  
**Component**: Launcher / Desktop Integration  
**Status**: Complete

## Summary

Wayland and modern GNOME shells map running windows to their desktop entries via the `GtkApplication` ID. After the app ID changed to `com.github.auriora.onemount` (May 6, 2025) the packaged `.desktop` file still shipped as `onemount-launcher.desktop`, so the shell stopped associating the running OneMount Launcher window with its icon. Window list taskbars therefore displayed the fallback icon even though the application menu entry was correct.

## Changes

1. Renamed the desktop entry artifacts to `com.github.auriora.onemount.desktop` (user and system variants) so the filename now matches the launcher’s `GtkApplication` ID.
2. Added `TryExec`, `Terminal`, `StartupWMClass` and `X-GNOME-Application-ID` metadata to both desktop entry variants for better integration with GNOME/KDE panels (kept `DBusActivatable=false` so desktop shells launch the binary directly instead of expecting D-Bus activation).
3. Updated the installation manifest to install the renamed desktop files into `share/applications` for both user and system targets.

## Testing

- Not run (desktop integration change only; requires packaging/GUI environment to verify, which is unavailable in this context).

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
