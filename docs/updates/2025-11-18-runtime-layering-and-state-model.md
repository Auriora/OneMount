# Runtime Layering & State Model Alignment (2025-11-18)

**Type**: Documentation Update  
**Status**: Complete  
**Components**: `.kiro/specs/system-verification-and-fix/requirements.md`, `.kiro/specs/system-verification-and-fix/design.md`

## Summary

- Clarified Requirement 2 so every FUSE callback is explicitly local-first and virtual filesystem entries live inside the metadata DB with overlay policies instead of a separate wrapper.
- Added hydration/eviction state expectations to Requirement 3 and introduced a new Requirement 21 that formalizes the `GHOST → HYDRATING → HYDRATED → DIRTY_LOCAL/…` state machine, ensuring cache, sync, and conflict logic share a vocabulary.
- Updated Requirements 5, 17, and 20 to lock the change-notification layer to Socket.IO (webhook transport removed entirely) and removed the duplicated Requirement 9 block for consistency.
- Updated the design document with runtime-layering diagrams, the item state table, virtual-item/overlay guidance, and a ChangeNotifier architecture section plus component description so the specs now match the revised requirements.

## Testing

- Not required (documentation-only changes)

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
