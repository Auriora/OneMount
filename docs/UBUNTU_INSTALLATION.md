# Ubuntu and Linux Mint Installation Guide

OneMount is optimized for Ubuntu 24.04 LTS and Linux Mint 22, providing the best compatibility and performance on these distributions.

## Supported Distributions

### Primary Support (Recommended)
- **Ubuntu 24.04 LTS (Noble Numbat)** - Full support with Go 1.22
- **Linux Mint 22 (Wilma)** - Full support (based on Ubuntu 24.04)

### Secondary Support
- **Ubuntu 22.04 LTS (Jammy Jellyfish)** - Limited support with Go 1.18
- **Linux Mint 21 (Vanessa/Vera/Victoria/Virginia)** - Limited support (based on Ubuntu 22.04)

## Installation Methods

### Method 1: Download Pre-built Package (Recommended)

1. **Download the latest .deb package** from the [GitHub Releases](https://github.com/auriora/onemount/releases) page
2. **Install the package:**
   ```bash
   sudo apt update
   sudo apt install ./onemount_*.deb
   ```

### Method 2: Build from Source

#### Prerequisites
```bash
# Ubuntu 24.04 / Linux Mint 22
sudo apt update
sudo apt install -y golang-go build-essential pkg-config libwebkit2gtk-4.1-dev git fuse3

# Ubuntu 22.04 / Linux Mint 21 (requires additional setup)
sudo apt update
sudo apt install -y golang-go build-essential pkg-config libwebkit2gtk-4.0-dev git fuse3
```

#### Build and Install
```bash
# Clone the repository
git clone https://github.com/auriora/onemount.git
cd onemount

# Build the application
make all

# Install for current user
make install

# OR install system-wide
sudo make install-system
```

### Method 3: Docker-based Build

If you want to build packages in a clean environment:

```bash
# Clone the repository
git clone https://github.com/auriora/onemount.git
cd onemount

# Build the Docker image
make ubuntu-docker-image

# Build Ubuntu packages
make ubuntu-docker
```

## Post-Installation Setup

### 1. Enable FUSE for Non-root Users
```bash
# Add your user to the fuse group
sudo usermod -a -G fuse $USER

# Log out and log back in, or run:
newgrp fuse
```

### 2. Verify Installation
```bash
# Check if OneMount is installed
onemount --version

# Check if the launcher is available
onemount-launcher --help
```

### 3. First Run
```bash
# Launch OneMount GUI
onemount-launcher

# OR mount from command line
mkdir ~/OneDrive
onemount ~/OneDrive
```

## Desktop Integration

OneMount integrates seamlessly with Ubuntu and Linux Mint desktop environments:

### File Manager Integration
- **Nemo** (Linux Mint default) - Full support
- **Nautilus** (Ubuntu default) - Full support
- **Dolphin** (KDE) - Basic support
- **Thunar** (XFCE) - Basic support

### System Tray
OneMount appears in the system tray when running, providing quick access to:
- Mount/unmount operations
- Sync status
- Settings and preferences

## Troubleshooting

### Common Issues

#### 1. "Permission denied" when mounting
```bash
# Ensure you're in the fuse group
groups | grep fuse

# If not, add yourself and restart your session
sudo usermod -a -G fuse $USER
```

#### 2. WebKit dependency issues
```bash
# Ubuntu 24.04 / Linux Mint 22
sudo apt install libwebkit2gtk-4.1-0

# Ubuntu 22.04 / Linux Mint 21
sudo apt install libwebkit2gtk-4.0-37
```

#### 3. Go version compatibility (Ubuntu 22.04/Mint 21)
```bash
# Check Go version
go version

# If Go 1.18, you may need to install a newer version manually
# or use the pre-built packages instead of building from source
```

### Getting Help

- **GitHub Issues**: [Report bugs and request features](https://github.com/auriora/onemount/issues)
- **Documentation**: [Full documentation](https://github.com/auriora/onemount/tree/main/docs)
- **System Tests**: Run `make system-test` to verify your installation

## Uninstallation

### If installed via package:
```bash
sudo apt remove onemount
```

### If installed from source:
```bash
# User installation
make uninstall

# System installation
sudo make uninstall-system
```

## Advanced Configuration

### Systemd Service (Optional)
To run OneMount as a system service:

```bash
# Enable for current user
systemctl --user enable onemount
systemctl --user start onemount

# Check status
systemctl --user status onemount
```

### Custom Mount Options
```bash
# Mount with custom options
onemount ~/OneDrive --cache-size 1GB --sync-interval 30s
```

For more advanced configuration options, see the [Configuration Guide](CONFIGURATION.md).
