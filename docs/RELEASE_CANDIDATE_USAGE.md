# OneMount Release Candidate Version Management

OneMount is now configured to support release candidates using bump2version. The current version is **0.1.0rc1**.

## Current Configuration

The project now supports both standard releases and release candidates:
- **Standard versions**: `0.1.0`, `0.2.0`, `1.0.0`
- **Release candidates**: `0.1.0alpha1`, `0.1.0beta2`, `0.1.0rc3`

## Installation

Make sure bump2version is installed in the virtual environment:

```bash
python3 -m venv .venv
.venv/bin/pip install bump2version
```

## Usage Examples

### Release Candidate Progression

```bash
# Current version: 0.1.0rc1

# Bump RC number: 0.1.0rc1 → 0.1.0rc2
.venv/bin/bumpversion num

# Progress to final release: 0.1.0rc1 → 0.1.0
.venv/bin/bumpversion release
```

### Standard Version Bumps

```bash
# From stable version 0.1.0:
.venv/bin/bumpversion patch    # 0.1.0 → 0.1.1
.venv/bin/bumpversion minor    # 0.1.0 → 0.2.0  
.venv/bin/bumpversion major    # 0.1.0 → 1.0.0
```

### Creating New Release Candidates

```bash
# Jump to specific RC version
.venv/bin/bumpversion --new-version 0.2.0alpha1 minor  # Start new alpha
.venv/bin/bumpversion --new-version 0.2.0beta1 minor   # Start new beta
.venv/bin/bumpversion --new-version 0.2.0rc1 minor     # Start new RC

# Progress through pre-release stages
# alpha1 → alpha2 → beta1 → beta2 → rc1 → rc2 → final
```

### Release Stage Progression

The release stages progress in this order:
1. **alpha** - Early development versions
2. **beta** - Feature-complete, testing versions  
3. **rc** - Release candidates, final testing
4. **release** - Final stable release (optional value, becomes standard version)

## Files Managed

Bumpversion automatically updates version strings in:

1. **`cmd/common/common.go`** - Go version constant
2. **`docs/man/onemount.1`** - Man page version header
3. **`packaging/rpm/onemount.spec`** - RPM spec file version
4. **`packaging/deb/changelog`** - Debian changelog version
5. **`packaging/ubuntu/changelog`** - Ubuntu changelog version
6. **`.bumpversion.cfg`** - Configuration file current version

## Dry Run Testing

Always test changes with `--dry-run` first:

```bash
.venv/bin/bumpversion --dry-run --verbose num
.venv/bin/bumpversion --dry-run --verbose release
```

## Git Integration

The configuration is now set to:
- **commit = True** - Automatic commits when bumping versions
- **tag = True** - Automatic git tags when bumping versions
- **tag_name = v{new_version}** - Tags follow the format `v0.1.0rc1`, `v0.1.0`, etc.

When you bump a version, bumpversion will:
1. Update all configured files with the new version
2. Create a git commit with the message "Bump version: {old} → {new}"
3. Create a git tag with the new version (e.g., `v0.1.0rc2`)

## Automated Package Building

The project now includes automated package building triggered by version tags:

### How it works:
1. **Bump the version** using bumpversion (creates a git tag)
2. **Push the tag** to GitHub: `git push origin --tags`
3. **GitHub Actions automatically builds** Ubuntu packages
4. **Creates a GitHub Release** with the built packages attached

### Example workflow:
```bash
# Bump to next release candidate
.venv/bin/bumpversion num  # 0.1.0rc1 → 0.1.0rc2

# Push the new tag to trigger package building
git push origin --tags

# GitHub will automatically:
# - Build Ubuntu packages
# - Run package validation
# - Create a GitHub release
# - Upload packages as release assets
```

## GitHub Workflows

The project now has separate workflows for different purposes:

### 1. Continuous Integration (`.github/workflows/ci.yml`)
- **Triggers**: Every push to main (except tags), pull requests
- **Purpose**: Run tests, linting, and basic builds
- **Fast feedback** for development

### 2. Package Building (`.github/workflows/build-packages.yml`)
- **Triggers**: Version tags (e.g., `v0.1.0rc1`, `v0.1.0`)
- **Purpose**: Build and release Ubuntu packages
- **Creates GitHub releases** with package assets

### 3. Coverage Analysis (`.github/workflows/coverage.yml`)
- **Triggers**: Pushes to main
- **Purpose**: Generate code coverage reports

### 4. System Tests (`.github/workflows/system-tests*.yml`)
- **Triggers**: Manual or scheduled
- **Purpose**: End-to-end testing with real OneDrive accounts

## Current Project Status

- **Current Version**: 0.1.0rc1
- **All project files updated** to use the new version
- **Release candidate support** fully configured and tested
- **Backward compatible** with standard version format

The project is now ready for release candidate development workflow!
