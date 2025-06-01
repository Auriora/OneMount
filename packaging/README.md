# OneMount Packaging System

This directory contains the centralized packaging system for OneMount, designed to eliminate duplication and ensure consistency across different installation methods.

## Overview

The packaging system is built around a single source of truth: `install-manifest.json`. This file defines all files to be installed, their sources, destinations, and installation parameters for different installation types.

## Files

- **`install-manifest.json`** - Central manifest defining all installation files and their destinations
- **`rpm/onemount.spec`** - RPM package specification
- **`deb/`** - Debian package files
- **`../scripts/install-manifest.py`** - Python script that parses the manifest and generates installation commands

## Installation Types

The system supports three installation types:

1. **User Installation** (`user`) - Installs to user's home directory (`~/.local/`)
2. **System Installation** (`system`) - Installs system-wide to `/usr/local/`
3. **Package Installation** (`package`) - Used by package managers (RPM, DEB) to install to standard system locations

## Usage

### Command Line

The `install-manifest.py` script can be used directly:

```bash
# Generate Makefile install commands for user installation
python3 scripts/install-manifest.py --target makefile --type user --action install

# Generate Makefile uninstall commands for system installation
python3 scripts/install-manifest.py --target makefile --type system --action uninstall

# Generate validation commands
python3 scripts/install-manifest.py --target makefile --action validate

# Generate RPM install commands
python3 scripts/install-manifest.py --target rpm --action install

# Generate RPM files list
python3 scripts/install-manifest.py --target rpm --action files

# Generate Debian install commands
python3 scripts/install-manifest.py --target debian --action install
```

### Makefile Integration

The Makefile has been updated to use the centralized system:

```makefile
install: onemount onemount-launcher
	@python3 scripts/install-manifest.py --target makefile --type user --action install | bash

install-system: onemount onemount-launcher
	@python3 scripts/install-manifest.py --target makefile --type system --action install | bash

uninstall:
	@python3 scripts/install-manifest.py --target makefile --type user --action uninstall | bash

uninstall-system:
	@python3 scripts/install-manifest.py --target makefile --type system --action uninstall | bash

validate-packaging:
	@echo "Validating packaging requirements..."
	@python3 scripts/install-manifest.py --target makefile --action validate | bash
	@echo "All packaging requirements validated successfully"
```

### RPM Integration

The RPM spec file uses the centralized system:

```spec
%install
rm -rf $RPM_BUILD_ROOT
# Use centralized installation manifest
python3 scripts/install-manifest.py --target rpm --action install | bash

%files
# Use centralized installation manifest for files list
%(python3 scripts/install-manifest.py --target rpm --action files)
```

### Debian Integration

The Debian rules file uses the centralized system:

```makefile
override_dh_auto_install:
	# Use centralized installation manifest
	python3 scripts/install-manifest.py --target debian --action install | bash
```

## Manifest Structure

The `install-manifest.json` file is organized into sections:

- **`binaries`** - Executable files (onemount, onemount-launcher)
- **`icons`** - Icon files in various formats
- **`desktop`** - Desktop entry files (with template processing)
- **`systemd`** - Systemd service files (with template processing)
- **`documentation`** - Manual pages and documentation
- **`directories`** - Directories that need to be created
- **`post_install`** - Commands to run after installation
- **`post_uninstall`** - Commands to run after uninstallation

Each file entry specifies:
- `source` - Source file path
- `dest_user` - Destination for user installation
- `dest_system` - Destination for system installation
- `dest_package` - Destination for package installation
- `mode` - File permissions

Template files also include substitution parameters for different installation types.

## Benefits

1. **Single Source of Truth** - All installation files are defined in one place
2. **Consistency** - All packaging methods use the same file list and destinations
3. **Maintainability** - Adding or removing files only requires updating the manifest
4. **Error Prevention** - Eliminates the possibility of files being listed in some packaging methods but not others
5. **Validation** - Centralized validation ensures all required source files exist
6. **User-Friendly Output** - Colored progress messages showing what's being installed/uninstalled

## Adding New Files

To add a new file to the installation:

1. Add the file entry to the appropriate section in `install-manifest.json`
2. Specify the source path and destinations for all three installation types
3. Set the appropriate file mode
4. The file will automatically be included in all packaging methods

## Template Processing

Desktop and systemd files support template processing with different substitutions for each installation type. The system automatically handles:

- Path substitutions (`@BIN_PATH@`, `@ICON_PATH@`)
- Service configuration (`@AFTER@`, `@USER@`, `@GROUP@`, `@WANTED_BY@`)

For package installations, pre-generated system files are used instead of templates to avoid runtime dependencies on the packaging system.
