# Task Summary: Build Scripts and Scripts Directory Cleanup

**Date**: 2026-02-13  
**Branch**: `chore/scripts-cleanup`  
**Status**: Completed

## Overview

Completed two major tasks:
1. Scripts directory cleanup and consolidation
2. Debian package build fixes and script consolidation

## Task 1: Scripts Directory Cleanup

### Objective
Review and clean up the scripts/ directory to remove duplicates, redundancies, and obsolete one-off scripts.

### Changes Made

1. **Removed 7 obsolete migration scripts**:
   - `cleanup-old-auth-scripts.sh`
   - `cleanup-old-runners.sh`
   - `label-unlabeled-tests.sh`
   - `label-remaining-tests.sh`
   - `label-final-tests.sh`
   - `update-auth-token-paths.sh`
   - `host-sync-codex-config.sh`

2. **Consolidated mount timeout scripts**:
   - Merged 3 separate scripts into `scripts/mount-timeout-tools.sh`
   - New tool provides subcommands: `diagnose`, `fix`, `test`

3. **Removed redundant deployment script**:
   - Deleted `deploy-optimized-remote.sh` (functionality in `deploy-docker-remote.sh`)

4. **Reorganized test scripts**:
   - Moved 4 system test scripts from `scripts/` to `tests/system/`:
     - `cache-management.sh`
     - `task-5.4-filesystem-operations.sh`
     - `task-5.5-unmounting-cleanup.sh`
     - `task-5.6-signal-handling.sh`

### Results
- **Before**: 40 scripts total
- **After**: 30 scripts (26 in scripts/, 4 in tests/system/)
- **Reduction**: 25%

### Documentation
- `docs/updates/scripts-cleanup-analysis.md` - Detailed analysis
- `docs/updates/scripts-cleanup-summary.md` - Implementation summary

### Commits
- `94d036a` - Move system test scripts to tests/system/
- `9ea2be1` - Update scripts cleanup analysis with results
- `84a0a2e` - Add scripts cleanup implementation summary

## Task 2: Debian Package Build Fixes

### Objective
Fix Debian package build failures and consolidate build scripts.

### Problem 1: WebKit Build Tag Issue

**Issue**: Build was failing with:
```
Package webkit2gtk-4.0 was not found in the pkg-config search path
```

**Root Cause**: 
- Docker image only has `libwebkit2gtk-4.1-dev` installed
- Debian rules file was using obsolete `scripts/cgo-helper.sh`
- Build wasn't using proper build tags to select webkit version

**Solution**:
Updated `packaging/ubuntu/rules` to:
- Use `scripts/detect-go-build-tags.sh` for webkit version detection
- Pass detected build tags to `go build` via `-tags` flag
- Remove obsolete `cgo-helper.sh` call

**Result**: Build now correctly detects and uses webkit41 build tag.

### Problem 2: Duplicate Build Scripts

**Issue**: Multiple build scripts with inconsistent functionality:
- Root `build-deb-package.sh` - Simple, working shell script
- Python CLI via `scripts/utils/docker_build.py` - Complex reimplementation
- Root script had webkit fix, Python version didn't

**Solution**:
1. Moved `build-deb-package.sh` from root to `scripts/`
2. Updated Python CLI to call the canonical shell script
3. Marked Python `DockerPackageBuilder` class as deprecated
4. Updated all documentation references

**Benefits**:
- Single source of truth for build logic
- Consistency between CLI and direct usage
- Easier maintenance
- Both interfaces now have webkit fix

### Documentation
- `docs/updates/fix-debian-build-webkit.md` - WebKit fix details
- `docs/updates/consolidate-build-scripts.md` - Script consolidation
- Updated references in 3 other docs

### Commits
- `9ebda59` - Fix webkit build tags in Debian package build
- `b80b776` - Consolidate build scripts

## Usage

### Building Debian Packages

Both methods now use the same underlying script:

```bash
# Direct usage (recommended)
./scripts/build-deb-package.sh

# Via Python CLI
./scripts/dev.py build deb --docker
```

### Mount Timeout Tools

```bash
# Diagnose mount timeout issues
./scripts/mount-timeout-tools.sh diagnose

# Apply timeout fix
./scripts/mount-timeout-tools.sh fix

# Test mount timeout behavior
./scripts/mount-timeout-tools.sh test
```

## Technical Details

### WebKit Build Tag System

The project supports both webkit versions via build tags:
- `internal/graph/oauth2_gtk_webkit40.go` - Build tag: `webkit40`
- `internal/graph/oauth2_gtk_webkit41.go` - Build tag: `webkit41`

Detection script (`scripts/detect-go-build-tags.sh`):
1. Checks for `webkit2gtk-4.1` via pkg-config
2. Falls back to `webkit2gtk-4.0` if not found
3. Returns appropriate build tag

### Build Process

The canonical build script (`scripts/build-deb-package.sh`):
1. Extracts version from `packaging/rpm/onemount.spec`
2. Creates build directory structure
3. Ensures Docker image exists
4. Runs build in container:
   - Creates source tarball
   - Sets up Go vendor directory
   - Builds source and binary packages
5. Outputs `.deb` to `build/packages/deb/`

## Files Modified

### Scripts
- `scripts/build-deb-package.sh` - Moved from root, canonical build script
- `scripts/mount-timeout-tools.sh` - New consolidated tool
- `scripts/utils/docker_build.py` - Simplified to call shell script

### Packaging
- `packaging/ubuntu/rules` - Updated to use build tags

### Documentation
- `docs/updates/scripts-cleanup-analysis.md` - New
- `docs/updates/scripts-cleanup-summary.md` - New
- `docs/updates/fix-debian-build-webkit.md` - New
- `docs/updates/consolidate-build-scripts.md` - New
- `docs/fixes/debian-maintainer-scripts-naming.md` - Updated paths
- `docs/updates/2026-01-27-debian-packaging-fixes-complete.md` - Updated paths
- `docs/updates/fix-debian-build-webkit.md` - Updated paths

### Tests
- Moved 4 scripts from `scripts/` to `tests/system/`

## Verification

### Build Verification
```bash
$ ./scripts/build-deb-package.sh
[INFO] Building OneMount v0.1.0rc1-1%{?dist} Debian package
...
[SUCCESS] Build completed!
-rw-r--r-- 1 user user 7.7M Feb 13 17:56 build/packages/deb/onemount_0.1.0rc1_amd64.deb
```

Build correctly uses `webkit41` build tag and produces working package.

### Script Count Verification
```bash
$ ls scripts/*.sh | wc -l
26

$ ls tests/system/*.sh | wc -l
4

Total: 30 scripts (down from 40)
```

## Rules Applied

- **general-preferences.md** (Priority 50): DRY principle, single source of truth
- **operational-best-practices.md** (Priority 40): Tool-driven exploration, clear documentation
- **coding-standards.md** (Priority 100): Maintainable code structure, proper documentation
- **git-conventions.md** (Priority 15): Logical commit grouping, descriptive messages

## Next Steps

### Potential Future Improvements

1. **Remove Python Docker Implementation**: The `DockerPackageBuilder` class could be fully removed since it's now just a wrapper

2. **Consolidate Native Build**: Convert `scripts/utils/native_build.py` to shell script for consistency

3. **Remove Obsolete cgo-helper.sh**: Verify no other build processes use it, then remove

4. **Further Script Consolidation**: Review remaining scripts for additional consolidation opportunities

### Maintenance Notes

When modifying the build process:
1. Edit `scripts/build-deb-package.sh` (canonical script)
2. Test both direct and CLI usage
3. Do NOT modify `scripts/utils/docker_build.py` build logic (it's a wrapper)

## Summary

Successfully cleaned up scripts directory (25% reduction) and fixed Debian package build issues. The build now properly handles webkit version detection and uses a single, canonical build script accessible via both direct execution and Python CLI. All changes are documented and tested.
