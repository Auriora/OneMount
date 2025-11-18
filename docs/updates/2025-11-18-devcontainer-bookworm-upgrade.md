# Devcontainer Bookworm Upgrade

**Date**: 2025-11-18  
**Type**: Tooling  
**Component**: Devcontainer  
**Status**: Complete

## Summary

Switched the Go devcontainer base image from Debian Bullseye to Debian Bookworm so we inherit newer GLib and WebKit headers without custom backports. This ensures the MCP/Gtk stack compiles against upstream glib ≥ 2.74, eliminating the need for compatibility build tags in most environments.

## Key Changes

1. Updated `.devcontainer/Dockerfile` to reference `mcr.microsoft.com/devcontainers/go:1-1.23-bookworm`.
2. Documented the change here for future contributors who may rebuild the container or extend the image.

## Verification

- Change is limited to the base tag; container rebuild will pull the new Bookworm layer on next `Dev Containers: Rebuild`.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
