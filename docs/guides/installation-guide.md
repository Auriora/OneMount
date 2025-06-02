# OneMount Installation Guide

## Introduction

This guide provides detailed instructions for installing, configuring, and running OneMount on various Linux distributions. OneMount is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing.

## Installation types

This guide explains the steps and instructions required to install OneMount on supported Linux distributions. It also explains how to configure, start, and uninstall OneMount.

| **Type** | **Description** | **More information** |
| --------- | ----------------- | -------------------  |
| Package Manager Installation | Install OneMount using your distribution's package manager | [Package Manager Installation](#package-manager-installation) |
| Building from Source | Build and install OneMount from source code | [Building from Source](#building-from-source) |

## Overview

The installation process involves:

1. Installing OneMount using your distribution's package manager or building from source
2. Configuring OneMount to access your Microsoft OneDrive account
3. Setting up OneMount to run automatically on system startup (optional)
4. Verifying the installation

## System requirements

Before installing OneMount, ensure your system meets the following requirements:

### Minimum Requirements
* **Operating System**: Linux with FUSE support (kernel 2.6.14 or later)
* **Architecture**: x86_64 (64-bit)
* **Memory**: 512 MB RAM minimum, 1 GB recommended
* **Storage**: 100 MB free disk space for installation, additional space for file cache
* **Network**: Internet connection (for initial setup and downloading files)
* **Account**: Microsoft OneDrive account

### Supported Distributions
* **Fedora**: 35 and later
* **CentOS/RHEL**: 8 and later
* **Ubuntu**: 20.04 LTS and later
* **Debian**: 11 (Bullseye) and later
* **Arch Linux**: Current rolling release
* **Other**: Most modern Linux distributions with FUSE support

## Before you begin

Before installing OneMount, ensure you have:

* Administrative privileges (sudo access) for system-wide installation
* FUSE filesystem support enabled on your system

### For Building from Source

The following dependencies are required when building OneMount from source:

#### Runtime Dependencies
* **Go**: Version 1.24.2 or later
* **GCC**: C compiler for CGO compilation
* **pkg-config**: For library configuration

#### Development Libraries
* **webkit2gtk-4.0**: For GUI authentication components
* **json-glib**: For JSON processing in GUI components

#### Verification Commands
```bash
# Check Go version
go version

# Check GCC availability
gcc --version

# Check pkg-config
pkg-config --version

# Verify FUSE support
modinfo fuse
```

## Installation steps

### Package Manager Installation

#### Fedora/CentOS/RHEL

Users on Fedora/CentOS/RHEL systems are recommended to install OneMount from [COPR](https://copr.fedorainfracloud.org/coprs/auriora/OneMount/).

1. Enable the COPR repository:
   ```bash
   sudo dnf copr enable auriora/onemount
   ```

2. Install OneMount:
   ```bash
   sudo dnf install onemount
   ```

#### Ubuntu/Pop!\_OS/Debian

**Package installation for Ubuntu/Debian is currently in development.**

For now, please use the [Building from Source](#building-from-source) method below.

If you previously installed onemount via PPA, you can remove the old PPA from your system:

```bash
# Remove old PPA if previously added
sudo add-apt-repository --remove ppa:auriora/onemount
sudo apt update

# Or manually remove PPA files
sudo rm -f /etc/apt/sources.list.d/auriora-ubuntu-onemount-*.list
sudo apt update
```

**Note**: Official Ubuntu/Debian packages will be available in a future release.

### Building from Source

In addition to the traditional [Go tooling](https://golang.org/dl/), you will need a C compiler and development headers for `webkit2gtk-4.0` and `json-glib`.

#### On Fedora:

```bash
sudo dnf install golang gcc pkg-config webkit2gtk4.0-devel json-glib-devel
```

#### On Ubuntu/Debian:

```bash
sudo apt install golang gcc pkg-config libwebkit2gtk-4.0-dev libjson-glib-dev
```

#### On Arch:

```bash
sudo pacman -S go gcc pkg-config webkit2gtk json-glib
```

#### Building and Installing:

1. Clone the repository:
   ```bash
   git clone https://github.com/auriora/OneMount
   cd onemount
   ```

2. Build the project:
   ```bash
   make
   ```

3. Install system-wide:
   ```bash
   sudo make install
   ```

## Configuration

After installing OneMount, you can configure it in two ways:

### Using the GUI (Recommended)

Launch the `onemount-launcher` desktop app from your application menu. This provides a graphical interface to:
- Add OneDrive accounts
- Configure mount points
- Set up automatic startup
- Manage existing mounts

### Using the Command Line

To configure OneMount to start automatically on login via systemd:

```bash
# create the mountpoint and determine the service name
mkdir -p $MOUNTPOINT
export SERVICE_NAME=$(systemd-escape --template onemount@.service --path $MOUNTPOINT)

# mount onedrive and set it to automatically mount on login
systemctl --user daemon-reload
systemctl --user enable --now $SERVICE_NAME

# check onemount's logs for the current day
journalctl --user -u $SERVICE_NAME --since today
```

## Running OneMount

You can run OneMount in several ways:

### Using the GUI Launcher

The simplest way is to use the `onemount-launcher` desktop app, which will handle mounting and authentication for you.

### Using the Command Line

```bash
# Mount OneDrive at a specific location
onemount /path/to/mount/onedrive/at

# View statistics about your OneDrive cache without mounting
onemount --stats /path/to/mount/onedrive/at
```

### Using Systemd (for automatic startup)

If you've configured onemount with systemd as described in the configuration section, it will start automatically on login.

## Verify installation

To verify that onemount is installed and running correctly:

1. Check if the onemount process is running:
   ```bash
   ps aux | grep onemount
   ```

2. Check if the filesystem is mounted:
   ```bash
   mount | grep onemount
   ```

3. Try accessing your OneDrive files through the mount point:
   ```bash
   ls /path/to/mount/onedrive/at
   ```

## Troubleshooting

If you encounter issues with onemount, here are some common problems and solutions:

| Issue | Solution |
| ----- | -------- |
| Filesystem appears to hang or "freeze" | The filesystem may have crashed. Restart by unmounting and remounting: `fusermount3 -uz $MOUNTPOINT` and then remount. |
| "Read-only filesystem" error | Your computer is likely offline. onemount automatically switches to read-only mode when offline. It will restore write access when you reconnect. |
| Need to reset onemount completely | Delete all cached local data by running `onemount -w` or removing mounts through the GUI. |

For more detailed troubleshooting:

1. Check the logs:
   ```bash
   journalctl --user -u $SERVICE_NAME --since today
   ```

2. Enable debug logging:
   ```bash
   ONEMOUNT_DEBUG=1 onemount /path/to/mount
   ```

3. If you encounter a bug, please report it on the [GitHub Issues page](https://github.com/auriora/OneMount/issues) with:
   - Log output
   - Steps to reproduce the issue
   - Your Linux distribution and version

## Uninstallation

To uninstall OneMount:

1. Stop any running onemount instances:
   ```bash
   fusermount3 -uz /path/to/mount/onedrive/at
   ```

2. Disable the systemd service if enabled:
   ```bash
   export SERVICE_NAME=$(systemd-escape --template onemount@.service --path $MOUNTPOINT)
   systemctl --user disable $SERVICE_NAME
   ```

3. Uninstall the package using your distribution's package manager:

   For Fedora/CentOS/RHEL:
   ```bash
   sudo dnf remove onemount
   ```

   For Ubuntu/Debian:
   ```bash
   sudo apt remove onemount
   ```

   For Arch:
   ```bash
   sudo pacman -R onemount
   ```

## Next steps

After successfully installing onemount, you can:

1. Start using your OneDrive files directly from your Linux filesystem
2. Explore the advanced features of OneMount
3. Configure automatic mounting on system startup

For more information, refer to the [onemount GitHub repository](https://github.com/auriora/OneMount).