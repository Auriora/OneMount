# Consolidate Build Scripts

**Date**: 2026-02-13  
**Type**: Refactoring  
**Status**: Completed

## Problem

There were multiple build scripts for Debian packages with inconsistent functionality:

1. **Root script**: `build-deb-package.sh` - Simple, working shell script with webkit fix
2. **Python CLI**: `scripts/dev.py build deb --docker` - Complex Python implementation via `scripts/utils/docker_build.py`
3. **Inconsistency**: The root script had the webkit build tag fix, but the Python version didn't

This duplication created maintenance burden and confusion about which script to use.

## Solution

Consolidate to a single, canonical build script:

1. Move the working `build-deb-package.sh` from root to `scripts/build-deb-package.sh`
2. Update Python CLI to call the shell script instead of reimplementing the logic
3. Keep the Python implementation as a fallback but mark it as deprecated
4. Update documentation to reflect the canonical script location

## Changes Made

### 1. Moved Build Script

**From**: `build-deb-package.sh` (project root)  
**To**: `scripts/build-deb-package.sh`

This script is now the canonical way to build Debian packages. It includes:
- Proper webkit build tag detection via `detect-go-build-tags.sh`
- Docker image management
- Clean build process with proper artifact handling
- Clear status output

### 2. Updated Python CLI

Modified `scripts/utils/docker_build.py` to call the shell script instead of reimplementing the logic. This:
- Reduces code duplication
- Ensures consistency between CLI and direct script usage
- Maintains the Python CLI interface users expect
- Simplifies maintenance (one place to fix bugs)

### 3. Documentation Updates

- Updated `scripts/README.md` to reference the canonical script
- Removed outdated build documentation from project root
- Added this consolidation document

## Usage

### Direct Script Usage (Recommended)

```bash
# Build Debian package using Docker
./scripts/build-deb-package.sh
```

### Python CLI Usage

```bash
# Build via Python CLI (calls the shell script internally)
./scripts/dev.py build deb --docker
```

Both methods now use the same underlying implementation.

## Benefits

1. **Single Source of Truth**: One script to maintain and fix
2. **Consistency**: Same behavior whether called directly or via Python CLI
3. **Simplicity**: Shell script is easier to understand and debug than Python Docker API
4. **Maintainability**: Fixes only need to be applied once
5. **Reliability**: The working script with webkit fix is now the canonical version

## Technical Details

### Build Process

The canonical script (`scripts/build-deb-package.sh`):

1. Extracts version from `packaging/rpm/onemount.spec`
2. Creates build directory structure
3. Ensures Docker image `onemount-deb-builder:latest` exists
4. Runs build in Docker container with:
   - Source tarball creation
   - Go vendor directory setup
   - Debian source package build
   - Debian binary package build
5. Outputs `.deb` file to `build/packages/deb/`

### WebKit Build Tag Support

The script properly handles webkit version detection:
- Uses `scripts/detect-go-build-tags.sh` in the Debian rules
- Automatically selects webkit40 or webkit41 based on available libraries
- Works with both Ubuntu 22.04 (webkit 4.0) and 24.04 (webkit 4.1)

## Migration Notes

### For Users

No changes required. Both methods continue to work:
- `./scripts/build-deb-package.sh` (direct)
- `./scripts/dev.py build deb --docker` (CLI)

### For Developers

When modifying the build process:
1. Edit `scripts/build-deb-package.sh` (canonical script)
2. Test both direct and CLI usage
3. Do NOT modify `scripts/utils/docker_build.py` build logic (it's a wrapper)

## Future Considerations

### Potential Simplifications

1. **Remove Python Docker Implementation**: The complex Python Docker API code in `scripts/utils/docker_build.py` could be simplified to just a wrapper that calls the shell script

2. **Consolidate Native Build**: The native build (`scripts/utils/native_build.py`) could also be converted to a shell script for consistency

3. **Unified Build Interface**: Consider a single `scripts/build.sh` that handles both Docker and native builds with flags

### Keeping Python CLI

The Python CLI (`scripts/dev.py`) should remain as it provides:
- Rich terminal output and progress indicators
- Consistent interface across all dev operations
- Integration with other Python-based dev tools
- Better error handling and user feedback

## Related Files

- `scripts/build-deb-package.sh` - Canonical build script (moved from root)
- `scripts/utils/docker_build.py` - Python wrapper (simplified)
- `scripts/commands/build_commands.py` - CLI command definitions
- `packaging/ubuntu/rules` - Debian build rules (includes webkit fix)
- `scripts/detect-go-build-tags.sh` - Build tag detection
- `docker/images/deb-builder/Dockerfile` - Docker build image

## Rules Applied

- **general-preferences.md** (Priority 50): Avoided duplication, single source of truth
- **operational-best-practices.md** (Priority 40): Simplified maintenance, clear documentation
- **coding-standards.md** (Priority 100): DRY principle, maintainable code structure
