# Kiro Specs and Steering Reference Added

**Date**: 2026-01-28  
**Time**: 16:00:50  
**Type**: Documentation  
**Status**: ✅ Complete

## Summary

Created a developer reference document that consolidates Kiro spec structure, OneMount’s `.kiro/specs/` layout, and the local `.kiro/steering/` rules. The guide also records external Kiro docs used to align terminology and file intent.

## Changes Made

1. **New reference guide**
   - Added `docs/guides/developer/kiro-specs-steering-reference.md` with:
     - Kiro spec file definitions (requirements/design/tasks)
     - OneMount spec structure and lifecycle notes
     - Steering inclusion modes and OneMount steering file inventory
     - Local and external reference links

2. **Developer guide index update**
   - Linked the new reference in `docs/guides/developer/README.md`

3. **Updates index entry**
   - Added this update to `docs/updates/index.md`

## References

- Local: `.kiro/specs/README.md`, `.kiro/steering/*.md`, `docs/guides/ai-agent/`
- External: Kiro docs for specs and steering (see reference guide)

## Rules Applied

**Rules consulted**:
- `.kiro/steering/general-preferences.md` (Priority 50) — rule discovery and documentation expectations
- `.kiro/steering/operational-best-practices.md` (Priority 40) — documentation placement and process transparency
- `.kiro/steering/documentation-conventions.md` (Priority 20) — docs structure and update log requirements

**Rules applied**:
- Stored documentation under `docs/guides/developer/`
- Logged work in `docs/updates/` and updated `docs/updates/index.md`

**Overrides**: None
