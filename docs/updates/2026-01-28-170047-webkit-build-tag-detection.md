# WebKit Build Tag Detection for Go Builds and Tests

**Date**: 2026-01-28  
**Time**: 17:00:47  
**Type**: Build/Test  
**Status**: ✅ Complete

## Summary

Added automatic Go build tag detection for WebKit2GTK (4.0 vs 4.1) and GLib compatibility so Docker build/test entrypoints and the packaging Docker build choose the correct tags based on installed libraries.

## Changes Made

1. **New tag detection script**
   - Added `scripts/detect-go-build-tags.sh` to emit comma-separated Go build tags for the current environment.

2. **Docker entrypoints now apply tags**
   - `docker/scripts/test-entrypoint.sh` now sets `GOFLAGS` with detected tags before running tests/builds.
   - `docker/scripts/build-entrypoint.sh` now sets `GOFLAGS` with detected tags before building.

3. **Packaging Docker build uses tags**
   - `packaging/docker/Dockerfile` now computes build tags and passes them to `go build`.
   - Added required detection scripts to the Docker build context.

## Files Touched

- `scripts/detect-go-build-tags.sh`
- `docker/scripts/test-entrypoint.sh`
- `docker/scripts/build-entrypoint.sh`
- `packaging/docker/Dockerfile`

## Rules Applied

**Rules consulted**:
- `.kiro/steering/general-preferences.md` (Priority 50) — rule discovery and documentation expectations
- `.kiro/steering/operational-best-practices.md` (Priority 40) — tool-driven exploration and process transparency
- `.kiro/steering/documentation-conventions.md` (Priority 20) — update log requirements

**Rules applied**:
- Logged work in `docs/updates/` and updated `docs/updates/index.md`

**Overrides**: None
