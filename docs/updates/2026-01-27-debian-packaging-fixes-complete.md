# Debian Package Issues - Complete Resolution

**Date**: 2026-01-27  
**Status**: ✅ Complete  
**Commits**: bd27351, f740438, aeffc57, 004d622

## Summary

Successfully resolved all Debian package issues related to systemd service configuration and maintainer scripts. The package now builds correctly and includes all necessary lifecycle scripts.

## Issues Resolved

### 1. Systemd Service Configuration (Tasks 50-51)

**Problem**: Service used `Type=dbus` with invalid `BusName=org.onemount.FileStatus.mnt_%i` causing "Unit not found" errors. Systemd's `%i` escaping creates invalid D-Bus names.

**Solution**:
- Changed service to `Type=exec` (removed `BusName`)
- Added D-Bus name sanitization in `internal/nemo/dbus.go`
- Created pre-processed service file: `deployments/systemd/onemount@.service`
- Maintained template for future use: `deployments/systemd/onemount@.service.template`
- Updated install manifest to reference correct service file

**Files Modified**:
- `deployments/systemd/onemount@.service` (new)
- `deployments/systemd/onemount@.service.template` (updated)
- `internal/nemo/dbus.go` (sanitization logic)
- `internal/nemo/dbus_test.go` (updated tests)
- `packaging/install-manifest.json` (updated references)

### 2. Maintainer Scripts Not Included (Task 52)

**Problem**: Maintainer scripts existed but weren't being included in .deb package, preventing:
- Desktop menu database refresh
- Icon cache updates
- MIME database refresh
- Systemd daemon reload

**Root Cause**: Scripts were named generically (`postinst`, `prerm`, `postrm`) instead of with package name prefix required by debhelper.

**Solution**: Renamed scripts to follow Debian conventions:
```
packaging/ubuntu/postinst  → packaging/ubuntu/onemount.postinst
packaging/ubuntu/prerm     → packaging/ubuntu/onemount.prerm
packaging/ubuntu/postrm    → packaging/ubuntu/onemount.postrm
```

**Verification**: Extracted control files from built package confirmed:
- Scripts are now included
- Debhelper processed `#DEBHELPER#` markers
- Auto-generated code properly inserted

## How Debhelper Works

Debhelper scans `debian/` directory for files matching patterns:
- `debian/PACKAGENAME.postinst`
- `debian/PACKAGENAME.prerm`
- `debian/PACKAGENAME.postrm`

For each match, it:
1. Processes the `#DEBHELPER#` marker
2. Inserts auto-generated code (systemd handling, icon cache updates, etc.)
3. Includes the final script in the .deb package

## Maintainer Scripts Functionality

### postinst (Post-Installation)
- Reloads systemd daemon to pick up new service files
- Updates desktop database for menu entries
- Updates icon cache
- Updates MIME database

### prerm (Pre-Removal)
- Stops all running onemount instances before removal/upgrade
- Uses systemd to enumerate and stop `onemount@*.service` units

### postrm (Post-Removal)
- Updates desktop database after removal
- Updates icon cache after removal
- Updates MIME database after removal
- Reloads systemd daemon after service file removal

## Testing

Package builds successfully:
```bash
./build-deb-package.sh
```

Verification:
```bash
# Extract control files
dpkg-deb -e build/packages/deb/onemount_*.deb /tmp/deb-control

# Verify scripts are included
ls -la /tmp/deb-control/
# Output shows: postinst, prerm, postrm

# Verify debhelper processing
cat /tmp/deb-control/postinst
# Shows custom code + auto-generated debhelper code
```

## Documentation

Created comprehensive documentation:
- `docs/fixes/debian-maintainer-scripts-naming.md` - Detailed explanation of the fix
- This summary document

## References

- [Debian Policy Manual - Maintainer Scripts](https://www.debian.org/doc/debian-policy/ch-maintainerscripts.html)
- [Debhelper Man Page](https://manpages.debian.org/testing/debhelper/debhelper.7.en.html)
- [Debian Maintainer Scripts Guide](https://pmhahn.github.io/debian-102-maintainer-scripts/)

## Rules Applied

- **git-conventions.md** (Priority 15): Proper commit message format
- **general-preferences.md** (Priority 50): Direct implementation, documentation in `docs/`
- **operational-best-practices.md** (Priority 40): Tool-driven exploration, documentation consistency

## Next Steps

The Debian package is now fully functional. Future improvements could include:
1. Add lintian checks to CI/CD pipeline
2. Test package installation on various Debian/Ubuntu versions
3. Consider adding autopkgtest for automated testing
4. Document package installation and removal procedures for users
