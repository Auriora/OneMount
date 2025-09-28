[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go 1.23+](https://img.shields.io/badge/Go-1.23+-blue.svg?logo=go&logoColor=white)](https://golang.org/dl/)
[![Status: Development](https://img.shields.io/badge/Status-Development-lightgrey.svg)]()
[![GitHub release](https://img.shields.io/github/v/release/auriora/OneMount?include_prereleases)](https://github.com/auriora/OneMount/releases)
[![Ubuntu 24.04](https://img.shields.io/badge/Ubuntu-24.04%20LTS-E95420?logo=ubuntu&logoColor=white)](https://ubuntu.com/)
[![Linux Mint 22](https://img.shields.io/badge/Linux%20Mint-22-87CF3E?logo=linux-mint&logoColor=white)](https://linuxmint.com/)

[//]: # ([![GitHub stars]&#40;https://img.shields.io/github/stars/auriora/OneMount?style=social&#41;]&#40;https://github.com/auriora/OneMount/stargazers&#41;)
[![GitHub issues](https://img.shields.io/github/issues/auriora/OneMount)](https://github.com/auriora/OneMount/issues)
[![GitHub last commit](https://img.shields.io/github/last-commit/auriora/OneMount)](https://github.com/auriora/OneMount/commits/main)
[![Build Ubuntu Packages](https://github.com/auriora/OneMount/workflows/Build%20Ubuntu%20Packages/badge.svg)](https://github.com/auriora/OneMount/actions?query=workflow%3A%22Build+Ubuntu+Packages%22)
[![Run tests](https://github.com/auriora/OneMount/workflows/Run%20tests/badge.svg)](https://github.com/auriora/OneMount/actions?query=workflow%3A%22Run+tests%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/auriora/OneMount)](https://goreportcard.com/report/github.com/auriora/OneMount)
[![Platform: Linux](https://img.shields.io/badge/Platform-Linux-blue.svg?logo=linux&logoColor=white)](https://www.linux.org/)

![OneMount](assets/icons/onemount-64.png)

# OneMount

Mount your Microsoft OneDrive account as a native filesystem on Linux.

---

This repository was forked from [Jeff Stafford's one-driver](https://github.com/jstaf/one-driver) repository. Extensive changes have been made, leading to the decision to rename the project.

---

## Table of contents

1. [Project description](#project-description)
2. [Who this project is for](#who-this-project-is-for)
3. [Project dependencies](#project-dependencies)
4. [Instructions for using OneMount](#instructions-for-using-OneMount)
   - [Quick Installation Guide](#quick-installation-guide)
5. [Contributing guidelines](#contributing-guidelines)
6. [Additional documentation](#additional-documentation)
7. [Terms of use](#terms-of-use)

## Project description

OneMount is a network filesystem that gives your computer direct access to your
files on Microsoft OneDrive. This is not a sync client. Instead of syncing
files, OneMount performs an on-demand download of files when your computer
attempts to use them. OneMount allows you to use files on OneDrive as if they
were files on your local computer.

OneMount is extremely straightforward to use:

- Install OneMount using your favorite installation method.
- Click the "+" button in the app to setup one or more OneDrive accounts.
  (There's a command-line workflow for those who prefer doing things that way
  too!)
- Just start using your files on OneDrive as if they were normal files.

**Microsoft OneDrive works on Linux.**

Getting started with your files on OneDrive is as easy as running:
`OneMount /path/to/mount/onedrive/at` (there's also a helpful GUI!). To get a
list of all the arguments OneMount can be run with you can read the manual page
by typing `man OneMount` or get a quick summary with `OneMount --help`.

You can also view statistics about your OneDrive cache without mounting by using
the `--stats` flag: `OneMount --stats /path/to/mount/onedrive/at`. This will
display information about the metadata cache, content cache, upload queue,
file statuses, and the embedded bbolt database used for persistent storage.
The stats command now includes detailed metadata analysis such as file type distribution,
directory depth statistics, file size distribution, and file age information derived
from the bbolt database.

### Key features

OneMount has several nice features that make it significantly more useful than
other OneDrive clients:

- **Files are only downloaded when you use them.** OneMount will only download
  a file if you (or a program on your computer) uses that file. You don't need
  to wait hours for a sync client to sync your entire OneDrive account to your
  local computer or try to guess which files and folders you might need later
  while setting up a "selective sync". OneMount gives you instant access to
  _all_ of your files and only downloads the ones you use.

- **Bidirectional sync.** Although OneMount doesn't actually "sync" any files,
  any changes that occur on OneDrive will be automatically reflected on your
  local machine. OneMount will only redownload a file when you access a file
  that has been changed remotely on OneDrive. If you somehow simultaneously
  modify a file both locally on your computer and also remotely on OneDrive,
  your local copy will always take priority (to avoid you losing any local
  work).

- **Robust offline functionality.** Files you've opened previously will be available even
  if your computer has no access to the internet. OneMount now supports full read-write
  operations while offline, with comprehensive conflict resolution when you reconnect.
  Changes made offline are automatically synchronized with intelligent conflict detection
  and multiple resolution strategies (last-writer-wins, keep-both, user choice).

- **Fast and resilient.** Great care has been taken to ensure that OneMount never makes a
  network request unless it actually needs to. OneMount caches both filesystem
  metadata and file contents both in memory and on-disk. The system includes comprehensive
  error handling, retry mechanisms with exponential backoff, and automatic network recovery.
  Accessing your OneDrive files will be fast and snappy even if you're engaged in a fight
  to the death for the last power outlet at a coffeeshop with bad wifi. (This has definitely
  never happened to me before, why do you ask?)

- **Has a user interface.** You can add and remove your OneDrive accounts
  without ever using the command-line. Once you've added your OneDrive accounts,
  there's no special interface beyond your normal file browser.

- **Free and open-source.** They're your files. Why should you have to pay to
  access them? OneMount is licensed under the GPLv3, which means you will
  _always_ have access to use OneMount to access your files on OneDrive.

## Who this project is for

This project is intended for Linux users who want to:
- Access their Microsoft OneDrive files directly from their Linux filesystem
- Avoid syncing their entire OneDrive account to their local computer
- Have a seamless experience working with OneDrive files on Linux
- Easily switch between working on files locally and in Microsoft 365 online apps
- Migrate from Windows to Linux while keeping their files accessible

OneMount is particularly useful for:
- Linux desktop users who need to access OneDrive files
- Users with limited disk space who can't sync their entire OneDrive
- Users who work across multiple platforms (Windows, Mac, Linux)
- Users who want to view and edit OneDrive photos and documents on Linux

## Project dependencies

Before using OneMount, ensure you have:

* A Linux system with FUSE support
* A Microsoft OneDrive account
* Internet connection (for initial setup and downloading files)

For building from source, you'll need:
* Go programming language
* GCC compiler
* webkit2gtk-4.0 and json-glib development headers

## Instructions for using OneMount

Get started with OneMount by installing it using your distribution's package manager.

### Quick Installation Guide

#### Ubuntu 24.04 LTS / Linux Mint 22 (Recommended)

1. **Download and install the .deb package:**

   ```bash
   # Download the latest release
   wget https://github.com/auriora/OneMount/releases/latest/download/onemount_*.deb

   # Install the package
   sudo apt update
   sudo apt install ./onemount_*.deb
   ```

2. **Or build from source:**

   ```bash
   # Install dependencies
   sudo apt update
   sudo apt install golang-go build-essential pkg-config libwebkit2gtk-4.1-dev git fuse3

   # Clone and build
   git clone https://github.com/auriora/OneMount.git
   cd OneMount
   make all
   make install
   ```

#### Other Distributions

   ```bash
   # Arch/Manjaro
   # Install from AUR: https://aur.archlinux.org/packages/OneMount/

   # Fedora/RHEL/CentOS
   # RPM packages available in releases
   ```

For detailed installation instructions including Ubuntu 22.04/Linux Mint 21 support, see our [Ubuntu Installation Guide](docs/UBUNTU_INSTALLATION.md).

#### Getting Started

1. **Launch the application** using the GUI launcher or command line:

   ```bash
   # Using GUI
   OneMount-launcher

   # Using command line
   OneMount /path/to/mount/onedrive/at
   ```

2. **Authenticate** with your Microsoft account when prompted.

For detailed installation and configuration instructions, troubleshooting, and advanced usage, please refer to:
- [Ubuntu Installation Guide](docs/UBUNTU_INSTALLATION.md) - Ubuntu/Linux Mint specific instructions
- [Complete Installation Guide](docs/guides/installation-guide.md) - All distributions
- [Quickstart Guide](docs/guides/quickstart-guide.md) - Step-by-step getting started

## Contributing guidelines

If you're interested in contributing to OneMount or understanding its internals, please refer to our [Development Guidelines](docs/DEVELOPMENT.md) document. It provides information about:

* Project structure
* Tech stack
* Building from source
* Running tests
* Coding standards and best practices

### Release Management

OneMount uses automated release management with version-controlled package building:

#### For Maintainers

**Automated Release Process (using CLI tool - recommended):**
```bash
# Bump version and trigger automated package building
./scripts/dev release bump num          # 0.1.0rc1 → 0.1.0rc2 (release candidate)
./scripts/dev release bump release      # 0.1.0rc1 → 0.1.0 (stable release)
./scripts/dev release bump patch        # 0.1.0 → 0.1.1 (patch release)

# Preview changes without committing
./scripts/dev release bump num --dry-run

# Bump version without triggering package build
./scripts/dev release bump patch --no-push
```



**What happens automatically:**
1. **Version bump** - Updates all project files with new version
2. **Git commit and tag** - Creates commit and version tag (e.g., `v0.1.0rc2`)
3. **GitHub Actions trigger** - Tag push triggers package building workflow
4. **Package building** - Builds Ubuntu packages in clean Docker environment
5. **GitHub Release** - Creates release with packages as downloadable assets

**Manual version management:**
```bash
# Using bumpversion directly
.venv/bin/bumpversion num         # Bump RC number
.venv/bin/bumpversion release     # Release current RC
.venv/bin/bumpversion patch       # Bump patch version

# Push tags to trigger package building
git push origin --tags
```

#### For Contributors

The project uses separate workflows for different purposes:
- **CI workflow** (`.github/workflows/ci.yml`) - Runs on every push for testing
- **Package building** (`.github/workflows/build-packages.yml`) - Runs only on version tags
- **Coverage analysis** - Tracks code coverage trends
- **System tests** - End-to-end testing with real OneDrive accounts

See [RELEASE_CANDIDATE_USAGE.md](docs/RELEASE_CANDIDATE_USAGE.md) for detailed version management documentation.

### Building from source

In addition to the traditional [Go tooling](https://golang.org/dl/), you will
need a C compiler and development headers for `webkit2gtk-4.0` and `json-glib`.

On Fedora:
```bash
dnf install golang gcc pkg-config webkit2gtk3-devel json-glib-devel
```

On Ubuntu:
```bash
apt install golang gcc pkg-config libwebkit2gtk-4.0-dev libjson-glib-dev
```

Basic build and run:
```bash
# to build and run the binary
make
mkdir mount
./OneMount mount/

# in new window, check out the mounted filesystem
ls -l mount

# unmount the filesystem
fusermount3 -uz mount
# you can also just "ctrl-c" OneMount to unmount it
```

### Development CLI Tool

OneMount includes a unified development CLI tool that consolidates all development, build, testing, and deployment operations:

```bash
# Install CLI dependencies (first time only)
pip install -r scripts/requirements-dev-cli.txt

# Check development environment status
./scripts/dev info

# Build packages
./scripts/dev build deb --docker

# Run tests with coverage
./scripts/dev test coverage --threshold-line 85
./scripts/dev test system --category comprehensive

# Analyze code quality
./scripts/dev analyze test-suite --mode resolve

# Release management
./scripts/dev release bump patch --dry-run

# Cleanup operations
./scripts/dev clean all

# Get help for any command
./scripts/dev --help
./scripts/dev build --help
```

The CLI tool provides:
- **Unified interface** for all development operations
- **Rich terminal output** with colors and progress indicators
- **Built-in help** with examples for every command
- **Error handling** with prerequisite checking
- **Organized commands** in logical groups (build, test, release, github, analyze, deploy)

For detailed CLI documentation, see [Development CLI Guide](scripts/README.md).

### Running the tests

The tests will write and delete files/folders on your onedrive account at the
path `/onemount_tests`. Note that the offline test suite requires `sudo` to
remove network access to simulate being offline.

```bash
# setup test tooling for first time run
make test-init

# actually run tests
make test

# run only the Python tests for nemo-OneMount.py
make test-python

# run only the Go tests for the D-Bus interface
go test -v ./fs -run TestDBus

# run comprehensive system tests with real OneDrive account
make system-test-real

# run all system test categories
make system-test-all

# Using the new CLI tool (recommended)
./scripts/dev test coverage --threshold-line 80
./scripts/dev test system --category comprehensive
./scripts/dev test docker all --verbose
```

The test suite includes:
- Go tests for the filesystem functionality
- Go tests for the D-Bus interface that provides file status updates
- Python pytest tests for the nemo-OneMount.py extension that uses the D-Bus interface
- Offline tests that simulate network disconnection
- **System tests with real OneDrive account** for comprehensive end-to-end testing

### Installation from source

OneMount has multiple installation methods depending on your needs.

```bash
# install directly from source
make
sudo make install

# create an RPM for system-wide installation on RHEL/CentOS/Fedora using mock
sudo dnf install golang gcc webkit2gtk3-devel json-glib-devel pkg-config git \
    rsync rpmdevtools rpm-build mock
sudo usermod -aG mock $USER
newgrp mock
make rpm

# create a .deb for system-wide installation on Ubuntu/Debian using pbuilder
sudo apt update
sudo apt install golang gcc libwebkit2gtk-4.0-dev libjson-glib-dev pkg-config git \
    rsync devscripts debhelper build-essential pbuilder
sudo pbuilder create  # may need to add "--distribution focal" on ubuntu
make deb
```

## Additional documentation

For more information about OneMount:

### User Documentation
* [Quickstart Guide](docs/guides/quickstart-guide.md) - Step-by-step guide to get started quickly
* [Installation Guide](docs/guides/installation-guide.md) - Detailed installation and configuration instructions
* [Troubleshooting Guide](docs/guides/troubleshooting-guide.md) - Solutions for common issues and problems
* [Offline Functionality](docs/offline-functionality.md) - Complete guide to offline features and synchronization

### Developer Documentation
* [Development Guidelines](docs/DEVELOPMENT.md) - Information about the project structure, tech stack, and best practices
* [Debugging Guide](docs/guides/debugging.md) - Developer-focused procedures for logs, tracing, and quick diagnostics

* [OneMount Consolidated Action Plan](docs/OneMount-Consolidated-Action-Plan.md) - **Current project status, priorities, and AI implementation prompts**
* [Solo Developer AI Process](docs/Solo-Developer-AI-Process.md) - Project-agnostic development methodology
* [Documentation Consolidation Summary](docs/0-project-management/DOCUMENTATION_CONSOLIDATION_SUMMARY.md) - Overview of recent documentation updates

### Project Resources
* [GitHub Issues](https://github.com/auriora/OneMount/issues) - Report bugs or request features
* [GitHub Releases](https://github.com/auriora/OneMount/releases) - Download the latest releases

### Known limitations

* **File browser thumbnails**: Many file browsers (like [GNOME's Nautilus](https://gitlab.gnome.org/GNOME/nautilus/-/issues/1209)) will attempt to automatically download all files within a directory to create thumbnail images. This only needs to happen once - thumbnails will persist between filesystem restarts.

* **Symbolic links**: Microsoft does not support symbolic links on OneDrive. Attempting to create symbolic links returns ENOSYS (function not implemented).

* **OneDrive Recycle Bin**: Microsoft does not expose the OneDrive Recycle Bin APIs. To empty or restore the OneDrive Recycle Bin, you must use the OneDrive web UI. OneMount uses the native system trash/restore functionality independently.

* **Large files**: OneMount loads files into memory when you access them. This makes things fast but doesn't work well with very large files. Use a sync client like [rclone](https://rclone.org/) for multi-gigabyte files.

* **Backups**: OneDrive is not recommended for backups. Use tools like [restic](https://restic.net/) or [borg](https://www.borgbackup.org/) for reliable encrypted backups.

## How to get help

If you encounter issues with OneMount:

1. **Check the troubleshooting guide**: [Troubleshooting Guide](docs/guides/troubleshooting-guide.md) - comprehensive solutions for common issues
2. **Review installation documentation**: [Installation Guide](docs/guides/installation-guide.md) - detailed setup and configuration instructions
3. **Search existing issues**: [GitHub Issues](https://github.com/auriora/OneMount/issues) to see if your problem has been reported
4. **Report new issues** with:
   - System information (Linux distribution, OneMount version)
   - Debug output (`ONEMOUNT_DEBUG=1 onemount /mount/path`)
   - Log output (`journalctl --user -u onemount@* --since today`)
   - Steps to reproduce the issue
   - Expected vs. actual behavior

## Terms of use

OneMount is licensed under the [GNU General Public License v3.0 (GPLv3)](https://github.com/auriora/OneMount/blob/master/LICENSE).

This project is provided AS IS with no warranties or guarantees. It is in active development.

## Project structure

- `cmd/` — Main application entry points (CLI, GUI, etc.)
- `internal/` — Internal Go packages
- `pkg/` — (deprecated) previously reusable public packages; code migrated under `internal/`
- `scripts/` — General-purpose shell and Python scripts for development, testing, and tooling
- `packaging/` — Files for building distribution packages (deb, rpm, etc.)
- `build/` — Build artifacts (binaries, release zips/tars, etc.)
- `assets/`, `configs/`, `docs/`, etc. — Supporting files and documentation
