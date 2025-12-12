# Devcontainer Systemd User Support

**Date**: 2025-11-23  
**Type**: Tooling  
**Component**: Devcontainer  
**Status**: Complete

## Summary

Enable the devcontainer to run a real `systemd` PID 1 so `scripts/with-user-systemd.sh` and `onemount-launcher` can talk to a functional user systemd instance during local development.

## Key Changes

1. Switched the devcontainer to start with `/sbin/init` (systemd) as PID 1 by setting `overrideCommand` and disabling the default `init` shim.
2. Added privileged/cgroup/`tmpfs` run arguments and a writable `/sys/fs/cgroup` bind to allow nested user systemd to create its delegated hierarchy.
3. Retained existing FUSE device exposure and AppArmor relaxation so launcher/systemd integration tests keep working.

## Verification

- Pending: rebuild/reopen the devcontainer (`Dev Containers: Rebuild and Reopen in Container`) and re-run `scripts/with-user-systemd.sh onemount-launcher`; user systemd should stay up and the launcher should no longer log `Process org.freedesktop.systemd1 exited with status 1`.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-GUIDE-Operational-Best-Practices (priority 40)  
- AGENT-RULE-Documentation-Conventions (priority 20)
