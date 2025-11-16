# Devcontainer Codex Config Snapshot

**Date**: 2025-11-16  
**Type**: Tooling  
**Component**: Devcontainer / Codex CLI  
**Status**: Complete

## Summary

Mirrored the host Codex CLI configuration (with its MCP server definitions) into the repository's `.devcontainer` folder so the devcontainer image can source identical settings during future automation work. Auth secrets stay only on the host; contributors should provide their own `auth.json` or environment variables when building the container.

## Key Changes

1. Created `.devcontainer/codex/config.toml` as a direct copy of `~/.codex/config.toml`, preserving the `mcp_servers` block requested for MCP parity.
2. Noted that authentication material (e.g., `auth.json`, `.mcp-auth/`) remains host-only to avoid accidentally committing API keys; follow-up work can introduce a secrets mount or template if needed.
3. Documented this action here so other contributors know the config now lives in-repo and should be kept in sync when MCP server settings change.

## Verification

- Ran `ls .devcontainer/codex` to confirm `config.toml` is present and ready for devcontainer build context.

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
