# onedriver Quickstart Guide

## Overview

This quickstart guides you through:

* [Part 1: Installing onedriver](#part-1-installing-onedriver)
* [Part 2: Setting up your first OneDrive mount](#part-2-setting-up-your-first-onedrive-mount)
* [Part 3: Using your OneDrive files](#part-3-using-your-onedrive-files)

It is intended for Linux users who want to access their Microsoft OneDrive files directly from their Linux filesystem. It assumes that you have basic knowledge of:

* Using a terminal or command line interface
* Basic Linux file operations
* Microsoft OneDrive account management

## Before you start

Before running this quickstart, complete the following prerequisites:

* A Linux system with FUSE support
* A Microsoft OneDrive account
* Internet connection (for initial setup and downloading files)
* Administrative privileges (sudo access) for system-wide installation

## Part 1: Installing onedriver

This section guides you through installing onedriver on your Linux system.

### Step 1: Install onedriver using your distribution's package manager

#### For Fedora/CentOS/RHEL:
```bash
sudo dnf copr enable bcherrington/onedriver
sudo dnf install onedriver
```

#### For Ubuntu/Debian:
TODO complete these instructions

## Part 2: Setting up your first OneDrive mount

This section guides you through setting up your first OneDrive mount.

### Step 1: Launch onedriver

You can launch onedriver using either the GUI or command line:

#### Using the GUI (Recommended):
1. Open your application menu
2. Find and click on "onedriver"
3. The onedriver launcher will open

#### Using the command line:
```bash
onedriver-launcher
```

### Step 2: Add your OneDrive account

1. In the onedriver launcher, click the "+" button to add a new account
2. A browser window will open for Microsoft authentication
3. Sign in with your Microsoft account credentials
4. Grant permission for onedriver to access your OneDrive files
5. The browser will redirect back to onedriver

### Step 3: Configure your mount point

1. After authentication, onedriver will ask where to mount your OneDrive
2. Choose a location (default is ~/OneDrive)
3. Click "Mount" to create the mount

## Part 3: Using your OneDrive files

This section guides you through basic usage of onedriver.

### Step 1: Access your files

1. Open your file manager
2. Navigate to your mount point (e.g., ~/OneDrive)
3. You should see your OneDrive files and folders

Files are downloaded on-demand when you access them, so there might be a slight delay when opening a file for the first time.

### Step 2: Work with your files

* **Opening files**: Double-click on any file to open it with your default application
* **Creating files**: Create new files directly in the mount point
* **Editing files**: Edit files as you normally would; changes are automatically uploaded
* **Deleting files**: Delete files as you normally would; deletions are synchronized to OneDrive

### Step 3: Check file status

To see the status of your files (synced, uploading, etc.):

1. Right-click on a file in your file manager
2. Look for the onedriver status information (available in file managers with onedriver integration)

Alternatively, check the status using the command line:
```bash
onedriver --stats /path/to/mount/onedrive/at
```

## Next steps

Now that you've completed this quickstart, try these to learn more about onedriver:

* Read the [complete installation guide](installation_guide.md) for advanced configuration options
* Learn about [offline usage](https://github.com/bcherrington/onedriver/wiki/Offline-Usage) for working with files when disconnected
* Explore [command-line options](https://github.com/bcherrington/onedriver/wiki/Command-Line-Options) for advanced usage
* Set up [automatic mounting](installation_guide.md#configuration) on system startup