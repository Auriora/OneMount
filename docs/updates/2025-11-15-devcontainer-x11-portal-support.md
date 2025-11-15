# Devcontainer X11 + Portal Support

**Date**: 2025-11-15  
**Type**: Tooling  
**Component**: Devcontainer / GUI Tooling  
**Status**: Complete

## Summary

The `.devcontainer` image now mirrors the proven `xcalc` sample so OneMount GUI tooling (e.g., `onemount-launcher`) can run inside the container while rendering on the host X server. Portal-aware helpers like `host-open` are preinstalled so `gio open`, `xdg-open`, and MIME handlers call back into the host desktop for URIs/files.

## Key Changes

1. Added `x11-apps`, `xdg-utils`, and `libglib2.0-bin` to the devcontainer image so X11 diagnostics (`xclock`, `xdpyinfo`) are available for smoke tests.  
2. Vendored the `host-open` helper (from the `xcalc` demo) under `.devcontainer/bin/` and baked it into `/usr/local/bin`.  
3. Extended `devcontainer.json` with DISPLAY/Xauthority/session-bus env passthrough, bind-mounts for `/tmp/.X11-unix`, `${XDG_RUNTIME_DIR}`, host XDG config/data, and system MIME/desktop dirs, plus portal env defaults.  
4. Preserved the existing FUSE/capability settings while keeping AppArmor relaxed (`apparmor=unconfined`) so DBus portals are reachable from the container.

## Host Preparation / Usage

1. On the host, authorize the container user once per boot: `xhost +SI:localuser:$(whoami)`.
2. Ensure `XDG_RUNTIME_DIR`, `DBUS_SESSION_BUS_ADDRESS`, and portal services (`xdg-desktop-portal`, `xdg-desktop-portal-gtk`, `xdg-document-portal`) are running; the bind mount uses `propagation=rshared` so `/run/user/<uid>/doc` stays visible.
3. Launch the devcontainer normally; rebuild if Docker cached the older image (`Dev Containers: Rebuild and Reopen in Container`).
4. From inside the container, validate the display: `xclock & disown` (or any other X11 tool) should appear on the host.
5. Launch the GUI entrypoint via `onemount-launcher`—windows should appear on the host display with host themes/fonts.
6. Test portal routing with `host-open https://example.com` or `gio open /etc/hosts`; these commands should hand off to the host browser/file handler.

## Verification

- `xclock` renders on host display from inside the devcontainer.  
- `host-open https://example.com` opens the host browser via the portal backend.  
- `onemount-launcher` runs without `DISPLAY`/`Xauthority` errors and uses host MIME handlers.

## Follow-Ups

- None at this time; revisit if non-Linux hosts require alternative transports (e.g., XQuartz on macOS).

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-GUIDE-Coding-Standards (priority 100)  
- AGENT-RULE-Documentation-Conventions (priority 20)
