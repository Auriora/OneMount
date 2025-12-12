# Launcher Icon Packaging Fix

**Date**: 2025-11-23  
**Type**: Tooling  
**Component**: Packaging / Launcher  
**Status**: Complete

## Summary

Ensure the launcher installs and prefers the square OneMount icon set (SVG + 16–512 px PNGs) so window icons and About dialog use the correct logo instead of the wide banner PNG.

## Key Changes

1. Updated `packaging/install-manifest.json` to ship both the square icon series (`onemount-icon.svg` + 16–512 px PNGs) and the rectangular banner PNGs (`onemount.png`, `onemount-128.png`) for in-app use.
2. Launcher now prefers the square icons for window/taskbar, while the About dialog uses the rectangular banner PNG to match in-app branding.

## Verification

- Pending: rebuild artifacts and run `make install` (or `scripts/dev build manifest --target makefile --type user --action install`) then start `onemount-launcher`; the window and About dialog should display the square logo without warnings.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
