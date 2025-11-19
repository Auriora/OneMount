# Realtime-Only Change Notification (2025-11-18)

**Type**: Feature Removal  
**Status**: Complete  
**Components**: `cmd/common/config.go`, `cmd/onemount/main.go`, `internal/fs/*`, `.kiro/specs/system-verification-and-fix/*`

## Summary

- Removed the legacy HTTPS webhook listener (code, config surface, requirements, and design references) so the realtime path is exclusively Socket.IO per the latest requirements.
- Simplified configuration: `realtime` block now only exposes `enabled`, `pollingOnly`, `resource`, `clientState`, and `fallbackInterval`; CLI help/man page text now references Socket.IO only.
- Deleted the unused Graph webhook client code and subscription manager, tightened delta-loop logging to report “realtime” instead of “webhook,” and kept the Socket.IO manager + health reporting intact.
- Updated requirements/design/plan documents to reflect the Socket.IO-only architecture and documented review findings for the upcoming runtime-layering/state-machine work.

## Testing

- `go test ./cmd/common` *(fails: missing system dependency `webkit2gtk-4.0` in the current environment; Socket.IO logic unchanged).* 

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
