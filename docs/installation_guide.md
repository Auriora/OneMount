# onedriver Installation Guide

## Introduction

This guide provides detailed instructions for installing, configuring, and running onedriver on various Linux distributions. onedriver is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing.

## Installation types

This guide explains the steps and instructions required to install onedriver on supported Linux distributions. It also explains how to configure, start, and uninstall onedriver.

| **Type** | **Description** | **More information** |
| --------- | ----------------- | -------------------  |
| Package Manager Installation | Install onedriver using your distribution's package manager | [Package Manager Installation](#package-manager-installation) |
| Building from Source | Build and install onedriver from source code | [Building from Source](#building-from-source) |

## Overview

The installation process involves:

1. Installing onedriver using your distribution's package manager or building from source
2. Configuring onedriver to access your Microsoft OneDrive account
3. Setting up onedriver to run automatically on system startup (optional)
4. Verifying the installation

## System requirements

Before installing onedriver, ensure your system meets the following requirements:

* A Linux system with FUSE support
* A Microsoft OneDrive account
* Internet connection (for initial setup and downloading files)

## Before you begin

Before installing onedriver, ensure you have:

* Administrative privileges (sudo access) for system-wide installation
* FUSE filesystem support enabled on your system
* For building from source:
  * Go programming language
  * GCC compiler
  * webkit2gtk-4.0 and json-glib development headers

## Installation steps

### Package Manager Installation

#### Fedora/CentOS/RHEL

Users on Fedora/CentOS/RHEL systems are recommended to install onedriver from [COPR](https://copr.fedorainfracloud.org/coprs/jstaf/onedriver/).

1. Enable the COPR repository:
   ```bash
   sudo dnf copr enable jstaf/onedriver
   ```

2. Install onedriver:
   ```bash
   sudo dnf install onedriver
   ```

#### OpenSUSE

OpenSUSE users need to add the COPR repo either for Leap or Tumbleweed:

1. For Leap 15.4:
   ```bash
   sudo zypper addrepo -g -r https://copr.fedorainfracloud.org/coprs/jstaf/onedriver/repo/opensuse-leap-15.4/jstaf-onedriver-opensuse-leap-15.4.repo onedriver
   sudo zypper --gpg-auto-import-keys refresh
   sudo zypper install onedriver
   ```

2. For Tumbleweed:
   ```bash
   sudo zypper addrepo -g -r https://copr.fedorainfracloud.org/coprs/jstaf/onedriver/repo/opensuse-tumbleweed/jstaf-onedriver-opensuse-tumbleweed.repo onedriver
   sudo zypper --gpg-auto-import-keys refresh
   sudo zypper install onedriver
   ```

#### Ubuntu/Pop!\_OS/Debian

Ubuntu/Pop!\_OS/Debian users can install onedriver from the [OpenSUSE Build Service](https://software.opensuse.org/download.html?project=home%3Ajstaf&package=onedriver).

If you previously installed onedriver via PPA, you can purge the old PPA from your system via:
```bash
sudo add-apt-repository --remove ppa:jstaf/onedriver
```

#### Arch/Manjaro/EndeavourOS

Arch/Manjaro/EndeavourOS users can install onedriver from the [AUR](https://aur.archlinux.org/packages/onedriver/).

#### Gentoo

Gentoo users can install onedriver from [this ebuild overlay](https://github.com/foopsss/gentoo-overlay) provided by a user. If you don't want to add user-hosted overlays to your system you may copy the ebuild for the latest version to a local overlay, which can be created by following the instructions available in the [Gentoo Wiki](https://wiki.gentoo.org/wiki/Creating_an_ebuild_repository).

Make sure to carefully review the ebuild for the package before installing it.

#### NixOS/NixPkgs

NixOS and Nix users can install onedriver from [the unstable channel](https://search.nixos.org/packages?channel=unstable&query=onedriver) either by adding the package to their system's configuration (if they are using NixOS) or by installing it manually via `nix-env -iA unstable.onedriver`.

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
   git clone https://github.com/jstaf/onedriver
   cd onedriver
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

After installing onedriver, you can configure it in two ways:

### Using the GUI (Recommended)

Launch the `onedriver-launcher` desktop app from your application menu. This provides a graphical interface to:
- Add OneDrive accounts
- Configure mount points
- Set up automatic startup
- Manage existing mounts

### Using the Command Line

To configure onedriver to start automatically on login via systemd:

```bash
# create the mountpoint and determine the service name
mkdir -p $MOUNTPOINT
export SERVICE_NAME=$(systemd-escape --template onedriver@.service --path $MOUNTPOINT)

# mount onedrive and set it to automatically mount on login
systemctl --user daemon-reload
systemctl --user enable --now $SERVICE_NAME

# check onedriver's logs for the current day
journalctl --user -u $SERVICE_NAME --since today
```

## Running onedriver

You can run onedriver in several ways:

### Using the GUI Launcher

The simplest way is to use the `onedriver-launcher` desktop app, which will handle mounting and authentication for you.

### Using the Command Line

```bash
# Mount OneDrive at a specific location
onedriver /path/to/mount/onedrive/at

# View statistics about your OneDrive cache without mounting
onedriver --stats /path/to/mount/onedrive/at
```

### Using Systemd (for automatic startup)

If you've configured onedriver with systemd as described in the configuration section, it will start automatically on login.

## Verify installation

To verify that onedriver is installed and running correctly:

1. Check if the onedriver process is running:
   ```bash
   ps aux | grep onedriver
   ```

2. Check if the filesystem is mounted:
   ```bash
   mount | grep onedriver
   ```

3. Try accessing your OneDrive files through the mount point:
   ```bash
   ls /path/to/mount/onedrive/at
   ```

## Troubleshooting

If you encounter issues with onedriver, here are some common problems and solutions:

| Issue | Solution |
| ----- | -------- |
| Filesystem appears to hang or "freeze" | The filesystem may have crashed. Restart by unmounting and remounting: `fusermount3 -uz $MOUNTPOINT` and then remount. |
| "Read-only filesystem" error | Your computer is likely offline. onedriver automatically switches to read-only mode when offline. It will restore write access when you reconnect. |
| Need to reset onedriver completely | Delete all cached local data by running `onedriver -w` or removing mounts through the GUI. |

For more detailed troubleshooting:

1. Check the logs:
   ```bash
   journalctl --user -u $SERVICE_NAME --since today
   ```

2. Enable debug logging:
   ```bash
   ONEDRIVER_DEBUG=1 onedriver /path/to/mount
   ```

3. If you encounter a bug, please report it on the [GitHub Issues page](https://github.com/jstaf/onedriver/issues) with:
   - Log output
   - Steps to reproduce the issue
   - Your Linux distribution and version

## Uninstallation

To uninstall onedriver:

1. Stop any running onedriver instances:
   ```bash
   fusermount3 -uz /path/to/mount/onedrive/at
   ```

2. Disable the systemd service if enabled:
   ```bash
   export SERVICE_NAME=$(systemd-escape --template onedriver@.service --path $MOUNTPOINT)
   systemctl --user disable $SERVICE_NAME
   ```

3. Uninstall the package using your distribution's package manager:

   For Fedora/CentOS/RHEL:
   ```bash
   sudo dnf remove onedriver
   ```

   For Ubuntu/Debian:
   ```bash
   sudo apt remove onedriver
   ```

   For Arch:
   ```bash
   sudo pacman -R onedriver
   ```

## Next steps

After successfully installing onedriver, you can:

1. Start using your OneDrive files directly from your Linux filesystem
2. Explore the advanced features of onedriver
3. Configure automatic mounting on system startup

For more information, refer to the [onedriver GitHub repository](https://github.com/jstaf/onedriver).