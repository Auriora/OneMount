# GLib Build Tag Auto-Detection

**Date**: 2025-11-16  
**Type**: Bugfix  
**Component**: Build / GTK Tooling  
**Status**: Complete

## Summary

Added automatic GLib version detection so the Go build always passes the correct `gotk3` build tag when the devcontainer ships an older (≤2.66) GLib. This prevents `C.g_binding_dup_source/target` lookup failures during `make build` on Debian-based images that lag behind GLib 2.68.

## Key Changes

1. Introduced `scripts/detect-glib-build-tag.sh`, which inspects `pkg-config --modversion glib-2.0` and emits `glib_2_xx` when the installed GLib is older than 2.68.
2. Plumbed the detected tag through new `GO_TAGS_FLAG` Makefile plumbing so every `go build` invocation automatically passes `-tags "glib_2_66"` (or nothing when unnecessary).
3. Left the detection overridable via `GOTK3_GLIB_TAG` so developers with newer GLib releases or custom setups can force a specific tag if needed.

## Verification

- `make onemount` (after detection) — succeeds and emits the `build/onemount` binary without `C.g_binding_dup_*` errors.
- `make onemount-launcher` — succeeds with the same tags, producing `build/onemount-launcher`.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
