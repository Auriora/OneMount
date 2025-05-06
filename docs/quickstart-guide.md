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
sudo dnf copr enable auriora/onemount sudo dnf install onemount
```S

#### Ubuntu/Debian
**Coming soon**

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

| Operation | Description | Action |
|-----------|-------------|---------|
| View Files | Browse mounted directory | Navigate to ~/OneDrive |
| Open Files | Access content | Double-click file |
| Edit Files | Modify content | Edit normally - auto-sync |
| Create Files | Add new content | Create in mount point |
| Delete Files | Remove content | Delete normally - auto-sync |

### File Status

View sync status through:
- File manager: Right-click > Properties
- Command line: `onemount --stats /mount/path`

## Command Reference

| Command | Purpose |
|---------|----------|
| `onemount-launcher` | Start OneMount |
| `onemount --stats` | Check sync status |
| `onemount --help`  | View all options |

## Advanced Topics

- [Complete Installation Guide](installation-guide.md)
- [Offline Usage](https://github.com/auriora/OneMount/wiki/Offline-Usage)
- [Command-Line Options](https://github.com/auriora/OneMount/wiki/Command-Line-Options)
- [Auto-mount Configuration](https://github.com/auriora/OneMount/wiki/Auto-Mount)
