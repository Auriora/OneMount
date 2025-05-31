# OneMount Quickstart Guide

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Initial Setup](#initial-setup)
4. [Daily Usage](#daily-usage)
5. [Command Reference](#command-reference)
6. [Advanced Topics](#advanced-topics)

## Prerequisites

| Requirement | Description |
|------------|-------------|
| System     | Linux with FUSE support |
| Account    | Microsoft OneDrive account |
| Network    | Internet connection for setup and sync |
| Privileges | Administrative (sudo) access |

## Installation

### Package Manager Installation

#### Fedora/CentOS/RHEL
```bash
sudo dnf copr enable auriora/onemount
sudo dnf install onemount
```

#### Ubuntu/Debian
**Building from source required** - See [Installation Guide](installation-guide.md)

## Initial Setup

### GUI Setup (Recommended)
1. Launch OneMount from your application menu
2. Click "+" to add a new account
3. Complete Microsoft authentication
4. Configure mount point (default: ~/OneDrive)

### Command-Line Setup
```bash 
onemount-launcher
``` 

## Daily Usage

### File Operations

| Operation | Description | Action | Offline Support |
|-----------|-------------|---------|-----------------|
| View Files | Browse mounted directory | Navigate to ~/OneDrive | ✅ Cached files |
| Open Files | Access content | Double-click file | ✅ Cached files |
| Edit Files | Modify content | Edit normally - auto-sync | ✅ Changes cached |
| Create Files | Add new content | Create in mount point | ✅ Changes cached |
| Delete Files | Remove content | Delete normally - auto-sync | ✅ Changes cached |
| Move/Rename | Reorganize files | Standard file operations | ✅ Changes cached |

### File Status and Sync

OneMount provides several ways to check file status and sync information:

#### Command Line Status
```bash
# View overall sync statistics
onemount --stats /mount/path

# Check if filesystem is online/offline
mount | grep onemount
```

#### File Manager Integration
- **File Properties**: Right-click files to see sync status
- **Mount Status**: Check if mount point is accessible
- **Network Indicator**: Filesystem automatically switches to read-only when offline

#### Offline Functionality
- **Automatic Detection**: Network connectivity is monitored automatically
- **Cached Access**: Previously accessed files remain available offline
- **Change Tracking**: Modifications made offline are synchronized when reconnected
- **Conflict Resolution**: Automatic handling of conflicts when changes occur both locally and remotely

## Command Reference

| Command | Purpose |
|---------|----------|
| `onemount-launcher` | Start OneMount |
| `onemount --stats` | Check sync status |
| `onemount --help`  | View all options |

## Advanced Topics

- [Complete Installation Guide](installation-guide.md) - Detailed installation and configuration instructions
- [Troubleshooting Guide](troubleshooting-guide.md) - Solutions for common issues and problems
- [Offline Functionality](offline-functionality.md) - Complete guide to offline features and synchronization
- [Development Guidelines](DEVELOPMENT.md) - Information for contributors and developers
