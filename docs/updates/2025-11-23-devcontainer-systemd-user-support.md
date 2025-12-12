# Devcontainer Systemd User Support

**Date**: 2025-11-23  
**Type**: Tooling  
**Component**: Devcontainer  
**Status**: Complete

## Summary

Enable the devcontainer to run a real `systemd` PID 1 so `scripts/with-user-systemd.sh` and `onemount-launcher` can talk to a functional user systemd instance during local development.

## Key Changes

1. Ensure the container now launches as `root` to let `/sbin/init` (systemd) run as PID 1, while the editor still logs in as `vscode` via `remoteUser`.
2. Explicitly install `systemd`/`systemd-sysv`, set `container=docker`, and mask tty/logind units so systemd can start cleanly inside the devcontainer.
3. Keep privileged/cgroup/`tmpfs` run arguments plus writable `/sys/fs/cgroup` to allow the user instance to create delegated cgroups; retain FUSE/AppArmor relaxations for launcher tests.

## Verification

- Pending: rebuild/reopen the devcontainer (`Dev Containers: Rebuild and Reopen in Container`) so PID 1 becomes systemd, then rerun `scripts/with-user-systemd.sh onemount-launcher`; user systemd should stay up and the launcher should no longer log `Process org.freedesktop.systemd1 exited with status 1`.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)  
- AGENT-GUIDE-Operational-Best-Practices (priority 40)  
- AGENT-RULE-Documentation-Conventions (priority 20)
