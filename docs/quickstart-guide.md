# OneMount Quickstart Guide

## Overview

This quickstart guides you through:

* [Part 1: Installing OneMount](#part-1-installing-onemount)
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

## Part 1: Installing OneMount

This section guides you through installing OneMount on your Linux system.

### Step 1: Install OneMount using your distribution's package manager

#### For Fedora/CentOS/RHEL:
```bash
sudo dnf copr enable bcherrington/onemount
sudo dnf install onemount
```

#### For Ubuntu/Debian:
TODO complete these instructions

## Part 2: Setting up your first OneDrive mount

This section guides you through setting up your first OneDrive mount.

### Step 1: Launch OneMount

You can launch OneMount using either the GUI or command line:

#### Using the GUI (Recommended):
1. Open your application menu
2. Find and click on "OneMount"
3. The OneMount launcher will open

#### Using the command line:
```bash
onemount-launcher
```

### Step 2: Add your OneDrive account

1. In the OneMount launcher, click the "+" button to add a new account
2. A browser window will open for Microsoft authentication
3. Sign in with your Microsoft account credentials
4. Grant permission for onemount to access your OneDrive files
5. The browser will redirect back to OneMount

### Step 3: Configure your mount point

1. After authentication, onemount will ask where to mount your OneDrive
2. Choose a location (default is ~/OneDrive)
3. Click "Mount" to create the mount

## Part 3: Using your OneDrive files

This section guides you through basic usage of onemount.

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
2. Look for the onemount status information (available in file managers with onemount integration)

Alternatively, check the status using the command line:
```bash
onemount --stats /path/to/mount/onedrive/at
```

## Next steps

Now that you've completed this quickstart, try these to learn more about OneMount:

* Read the [complete installation guide](installation-guide.md) for advanced configuration options
* Learn about [offline usage](https://github.com/bcherrington/OneMount/wiki/Offline-Usage) for working with files when disconnected
* Explore [command-line options](https://github.com/bcherrington/OneMount/wiki/Command-Line-Options) for advanced usage
* Set up [automatic mounting](installation-guide.md#configuration) on system startup
