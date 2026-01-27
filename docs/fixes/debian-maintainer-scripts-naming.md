# Debian Maintainer Scripts Naming Convention Fix

**Date**: 2026-01-27  
**Issue**: Maintainer scripts not being included in .deb package  
**Status**: Fixed

## Problem

The Debian package was building successfully, but maintainer scripts (postinst, prerm, postrm) were not being included in the final .deb package. This caused two issues:

1. Desktop menu items weren't being refreshed after install/uninstall
2. Systemd daemon wasn't being reloaded after service file installation

The scripts existed in `packaging/ubuntu/` with the correct `#DEBHELPER#` marker, but debhelper wasn't finding them.

## Root Cause

Debhelper uses specific file naming conventions to identify maintainer scripts:

- **Single binary package**: `debian/postinst`, `debian/prerm`, `debian/postrm`
- **Named package**: `debian/PACKAGENAME.postinst`, `debian/PACKAGENAME.prerm`, etc.

Our scripts were named generically (`postinst`, `prerm`, `postrm`) instead of with the package name prefix (`onemount.postinst`, `onemount.prerm`, `onemount.postrm`).

## Solution

Renamed the maintainer scripts to include the package name:

```bash
packaging/ubuntu/postinst    → packaging/ubuntu/onemount.postinst
packaging/ubuntu/prerm       → packaging/ubuntu/onemount.prerm
packaging/ubuntu/postrm      → packaging/ubuntu/onemount.postrm
```

## How Debhelper Works

1. During `dpkg-buildpackage`, the `packaging/ubuntu/` directory is moved to `debian/`
2. Debhelper scans `debian/` for files matching patterns like:
   - `debian/PACKAGENAME.postinst`
   - `debian/PACKAGENAME.prerm`
   - `debian/PACKAGENAME.postrm`
3. For each match, debhelper:
   - Processes the `#DEBHELPER#` marker
   - Inserts auto-generated code (systemd handling, etc.)
   - Includes the final script in the .deb package

## Verification

After rebuilding the package, verify the scripts are included:

```bash
# Build the package
./build-deb-package.sh

# Extract and inspect
dpkg-deb -e build/packages/deb/onemount_*.deb /tmp/deb-control
ls -la /tmp/deb-control/
cat /tmp/deb-control/postinst
```

You should see:
- `postinst`, `prerm`, `postrm` files in the control directory
- Auto-generated debhelper code inserted at the `#DEBHELPER#` marker

## References

- [Debian Policy Manual - Maintainer Scripts](https://www.debian.org/doc/debian-policy/ch-maintainerscripts.html)
- [Debhelper Man Page](https://manpages.debian.org/testing/debhelper/debhelper.7.en.html)
- [Debian Maintainer Scripts Guide](https://pmhahn.github.io/debian-102-maintainer-scripts/)

## Related Files

- `packaging/ubuntu/onemount.postinst` - Post-installation script
- `packaging/ubuntu/onemount.prerm` - Pre-removal script
- `packaging/ubuntu/onemount.postrm` - Post-removal script
- `packaging/ubuntu/rules` - Build rules (calls debhelper)
- `build-deb-package.sh` - Package build script
