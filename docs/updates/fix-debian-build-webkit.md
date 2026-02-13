# Fix Debian Package Build - WebKit Build Tags

**Date**: 2026-02-13  
**Type**: Bug Fix  
**Status**: Completed

## Problem

The Debian package build was failing with the error:
```
Package webkit2gtk-4.0 was not found in the pkg-config search path
```

The build was using the old `scripts/cgo-helper.sh` approach which modified source files in place to switch between webkit versions. This approach was problematic and didn't work with the newer build tag system.

## Root Cause

1. The project has two webkit implementation files with build tags:
   - `internal/graph/oauth2_gtk_webkit40.go` (build tag: `webkit40`)
   - `internal/graph/oauth2_gtk_webkit41.go` (build tag: `webkit41`)

2. The Docker base image (`docker/images/builder/Dockerfile`) only has `libwebkit2gtk-4.1-dev` installed

3. The Debian rules file (`packaging/ubuntu/rules`) was using the old `scripts/cgo-helper.sh` script instead of the proper build tag detection script

4. The build wasn't passing the appropriate `-tags` flag to `go build`, so it was trying to compile both webkit versions

## Solution

Updated `packaging/ubuntu/rules` to:
1. Use `scripts/detect-go-build-tags.sh` to detect available webkit version
2. Pass the detected build tags to `go build` via `-tags` flag
3. Remove the call to the obsolete `scripts/cgo-helper.sh`

### Changes Made

**File**: `packaging/ubuntu/rules`

Changed from:
```makefile
override_dh_auto_build:
	bash scripts/cgo-helper.sh
	# Create build directory for binaries
	mkdir -p build
	GOCACHE=/tmp/go-cache go build -v -mod=vendor \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell cat .commit)" \
		-o build/onemount \
		./cmd/onemount
```

Changed to:
```makefile
override_dh_auto_build:
	# Detect appropriate build tags for this environment
	$(eval BUILD_TAGS := $(shell bash scripts/detect-go-build-tags.sh))
	# Create build directory for binaries
	mkdir -p build
	GOCACHE=/tmp/go-cache go build -v -mod=vendor \
		-tags="$(BUILD_TAGS)" \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell cat .commit)" \
		-o build/onemount \
		./cmd/onemount
```

## Verification

Build completed successfully:
```bash
$ ./scripts/build-deb-package.sh
[INFO] Building OneMount v0.1.0rc1-1%{?dist} Debian package
[INFO] Starting Docker build...
...
[SUCCESS] Build completed!
-rw-r--r-- 1 bcherrington bcherrington 7.7M Feb 13 13:41 build/packages/deb/onemount_0.1.0rc1_amd64.deb
```

The build correctly detected and used `webkit41` build tag, compiling against `libwebkit2gtk-4.1-dev` which is available in the Docker image.

## Technical Details

### Build Tag Detection

The `scripts/detect-go-build-tags.sh` script:
1. Calls `scripts/detect-webkit-version.sh` to check which webkit version is available via pkg-config
2. Returns `webkit41` if `webkit2gtk-4.1` is found
3. Returns `webkit40` if `webkit2gtk-4.0` is found
4. Also detects GLib version tags for compatibility

### WebKit Version Support

The project supports both webkit versions:
- **webkit2gtk-4.0**: Older version, used on older Ubuntu/Debian systems
- **webkit2gtk-4.1**: Newer version, used on Ubuntu 24.04 and newer

The Docker base image uses Ubuntu 24.04 which has webkit2gtk-4.1, so the build correctly uses that version.

## Related Files

- `packaging/ubuntu/rules` - Updated to use build tags
- `scripts/detect-go-build-tags.sh` - Detects available build tags
- `scripts/detect-webkit-version.sh` - Detects webkit version
- `internal/graph/oauth2_gtk_webkit40.go` - WebKit 4.0 implementation
- `internal/graph/oauth2_gtk_webkit41.go` - WebKit 4.1 implementation
- `docker/images/builder/Dockerfile` - Base image with webkit2gtk-4.1

## Future Considerations

The old `scripts/cgo-helper.sh` script could potentially be removed if it's no longer used elsewhere in the build system. However, it should be verified that no other build processes depend on it before removal.

## Rules Applied

- **operational-best-practices.md** (Priority 40): Used tool-driven exploration to identify the issue
- **coding-standards.md** (Priority 100): Maintained clean, documented build process
- **general-preferences.md** (Priority 50): Preferred direct implementation over extensive analysis
