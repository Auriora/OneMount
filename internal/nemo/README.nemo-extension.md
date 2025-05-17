# Nemo Extension for OneMount

This directory contains the Nemo file manager extension and related integration code for OneMount. It enables enhanced file manager support and D-Bus integration for OneDrive mounts.

# OneMount Nemo Extension

This extension adds file status indicators (emblems) to files and folders managed by OneMount in the Nemo file manager.

## Status Indicators

The extension shows the following status indicators:

- **Cloud** (emblem-synchronizing-offline): File exists in the cloud but not in local cache
- **Local** (emblem-default): File exists in the local cache
- **LocalModified** (emblem-synchronizing-locally-modified): File has been modified locally but not synced
- **Syncing** (emblem-synchronizing): File is currently being synchronized
- **Downloading** (emblem-downloads): File is currently being downloaded
- **OutOfSync** (emblem-important): File needs to be updated from OneDrive cloud
- **Error** (emblem-error): There was an error synchronizing the file
- **Conflict** (emblem-warning): There is a conflict between local and remote versions

## Installation

1. Make sure you have the required dependencies:
   ```
   sudo apt install python3-nemo python3-gi
   ```

2. Copy the extension to the Nemo extensions directory:
   ```
   mkdir -p ~/.local/share/nemo-python/extensions/
   cp nemo-onemount.py ~/.local/share/nemo-python/extensions/
   chmod +x ~/.local/share/nemo-python/extensions/nemo-onemount.py
   ```

3. Restart Nemo:
   ```
   nemo -q
   nemo
   ```

## Troubleshooting

If the extension doesn't work:

1. Check if the extension is loaded:
   ```
   nemo --quit
   nemo --debug
   ```
   Look for messages related to "nemo-onemount" in the output.

2. Verify that the extension has execute permissions:
   ```
   chmod +x ~/.local/share/nemo-python/extensions/nemo-onemount.py
   ```

3. Make sure you have the required dependencies installed:
   ```
   sudo apt install python3-nemo python3-gi
   ```

4. Check if OneMount is properly mounted:
   ```
   mount | grep onemount
   ```

## Uninstallation

To remove the extension:
```
rm ~/.local/share/nemo-python/extensions/nemo-onemount.py
nemo -q
nemo
```