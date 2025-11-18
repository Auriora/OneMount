# Devcontainer Python Toolchain

**Date**: 2025-11-18  
**Type**: Tooling  
**Component**: Devcontainer  
**Status**: Complete

## Summary

Added a first-party Python toolchain to the Bookworm-based devcontainer so contributors can run the Nemo extension tests, Python-based CLI helpers, and packaging scripts without resorting to host-level installs or ad-hoc containers.

## Key Changes

1. Extended `.devcontainer/Dockerfile` so the base apt layer installs `python3`, `python3-dev`, `python3-pip`, `python3-venv`, and `python-is-python3` alongside the existing Go dependencies.
2. Documented this tooling addition here to make the devcontainer dependency footprint explicit for future rebuilds.

## Verification

- Not run: requires rebuilding the devcontainer image (`Dev Containers: Rebuild Container`) to pull the new apt layer and verify `python3 --version` inside the environment.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
