# Devcontainer X11 Bootstrap Helper

**Date**: 2025-11-16  
**Type**: Tooling  
**Component**: Devcontainer / GUI Tooling  
**Status**: Complete

## Summary

Added an `initializeCommand` hook plus a host-side helper script so the devcontainer automatically mirrors the `xcalc` DISPLAY export recipe. The helper grants `xhost` permissions before the container starts, ensuring OneMount GUI binaries and the `host-open` portal launcher can reach the host X server without manual steps.

## Key Changes

1. Introduced `.devcontainer/bin/ensure-host-x11`, which authorizes the current host user (or `HOST_X11_USER`) against `${DISPLAY}`/`HOST_DISPLAY` using `xhost`, and degrades gracefully when `xhost` or `DISPLAY` are unavailable.
2. Wired the helper into `devcontainer.json` via `initializeCommand` so the authorization runs once per workspace launch before Docker spins up the container.
3. Dropped the `type=tmpfs,target=/tmp` mount from `devcontainer.json` so JetBrains' Remote Dev worker can execute binaries from `/tmp`; the tmpfs mount forced `noexec`, causing IJent to bail out with exit code 126.
4. Documented the change here so future contributors know why the extra lifecycle hook exists and how it keeps the `xcalc`-style flow intact.

## Verification

- Ran `.devcontainer/bin/ensure-host-x11` locally to confirm it authorizes the active user (`xhost +SI:localuser:$USER`) and prints a status line.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
