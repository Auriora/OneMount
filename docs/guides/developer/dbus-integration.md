# OneMount D-Bus Interface

This document describes the D-Bus interface for OneMount file status updates.

## Overview

OneMount now provides a D-Bus interface for file status updates. This allows other applications, such as the Nemo file manager extension, to receive real-time updates about file status changes without having to poll the filesystem or read extended attributes.

## D-Bus Interface Specification

- **Service Name**: `org.onemount.FileStatus`
- **Object Path**: `/org/onemount/FileStatus`
- **Interface**: `org.onemount.FileStatus`

### Methods

- **GetFileStatus(path: string) -> status: string**
  - Gets the status of a file at the specified path
  - Parameters:
    - `path`: The full path to the file
  - Returns:
    - `status`: The status of the file (e.g., "Cloud", "Local", "Syncing", etc.)

### Signals

- **FileStatusChanged(path: string, status: string)**
  - Emitted when the status of a file changes
  - Parameters:
    - `path`: The full path to the file
    - `status`: The new status of the file

## Implementation Details

### Server Side (OneMount)

The D-Bus server is implemented in Go using the `github.com/godbus/dbus/v5` package. The server is started when OneMount is mounted and provides methods for getting file status and signals for file status changes.

The server is implemented in the `fs/dbus.go` file and is integrated with the existing file status tracking system in OneMount.

### Client Side (Nemo Extension)

The Nemo extension (`nemo-onemount.py`) has been updated to use the D-Bus interface for file status updates. It connects to the D-Bus service when initialized and listens for file status change signals.

The extension still falls back to reading extended attributes if the D-Bus service is not available, ensuring backward compatibility.

## Benefits

- **Real-time updates**: File status changes are immediately reflected in the file manager without polling
- **Reduced overhead**: No need to read extended attributes for every file, which can be expensive
- **Better integration**: Provides a standard interface for other applications to integrate with OneMount

## Example Usage

### Using D-Bus from the Command Line

You can query the file status using the `dbus-send` command:

```bash
dbus-send --session --print-reply --dest=org.onemount.FileStatus \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/path/to/your/file"
```

### Using D-Bus in Python

```python
import dbus

# Connect to the D-Bus session bus
bus = dbus.SessionBus()

# Get the OneMount D-Bus object
onemount = bus.get_object('org.onemount.FileStatus', '/org/onemount/FileStatus')

# Get the file status method
get_status = onemount.get_dbus_method('GetFileStatus', 'org.onemount.FileStatus')

# Get the status of a file
status = get_status('/path/to/your/file')
print(f"File status: {status}")

# Set up a signal handler for file status changes
def on_file_status_changed(path, status):
    print(f"File status changed: {path} -> {status}")

bus.add_signal_receiver(
    on_file_status_changed,
    dbus_interface='org.onemount.FileStatus',
    signal_name='FileStatusChanged'
)

# Start the main loop to receive signals
from dbus.mainloop.glib import DBusGMainLoop
from gi.repository import GLib

DBusGMainLoop(set_as_default=True)
loop = GLib.MainLoop()
loop.run()
```