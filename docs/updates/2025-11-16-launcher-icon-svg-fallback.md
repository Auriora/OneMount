# Launcher Icon SVG Fallback

**Date**: 2025-11-16  
**Type**: Bugfix  
**Component**: Launcher / Desktop Integration  
**Status**: Complete

## Summary

- `onemount-launcher` tried to load `onemount.svg` as the window icon, which fails on systems without an SVG loader (e.g., missing `librsvg`).
- Added a PNG-based fallback sequence so the launcher can always attach an icon without emitting warnings.

## Changes

1. Introduced `setWindowIcon`, which iterates over candidate files until a logo loads successfully.
2. Updated the launcher activation path to try `onemount.svg`, `onemount.png`, and `onemount-128.png` in order, logging success/failure per candidate.

## Testing

- Not run (requires a graphical session; change limited to icon selection logic).

## Rules Consulted

- AGENT-GUIDE-Coding-Standards (priority 100)
- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
