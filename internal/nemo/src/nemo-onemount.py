#!/usr/bin/python3

import gi
gi.require_version('Nemo', '3.0')
from gi.repository import Nemo, GObject, Gio, GLib
import os
import dbus
import dbus.mainloop.glib

class OneMountExtension(GObject.GObject, Nemo.InfoProvider):
    def __init__(self):
        # Initialize D-Bus main loop
        dbus.mainloop.glib.DBusGMainLoop(set_as_default=True)

        # Connect to D-Bus
        self.bus = dbus.SessionBus()
        self.dbus_proxy = None
        self.connect_to_dbus()

        # Set up signal handlers for file status changes
        self.file_status_cache = {}
        self.setup_dbus_signals()

        # Get list of OneMount mount points
        self.onemount_mounts = self._get_onemount_mounts()

    def connect_to_dbus(self):
        """Connect to the OneMount D-Bus service"""
        try:
            self.dbus_proxy = self.bus.get_object(
                'org.onemount.FileStatus',
                '/org/onemount/FileStatus'
            )
            print("Connected to OneMount D-Bus service")
        except dbus.exceptions.DBusException as e:
            # Silently handle the case when the D-Bus service is not available
            # This is expected when onemount is not running or D-Bus service is not registered
            self.dbus_proxy = None

    def setup_dbus_signals(self):
        """Set up D-Bus signal handlers for file status changes"""
        if self.dbus_proxy is None:
            return

        try:
            self.bus.add_signal_receiver(
                self._on_file_status_changed,
                dbus_interface='org.onemount.FileStatus',
                signal_name='FileStatusChanged'
            )
            print("Set up D-Bus signal handler for file status changes")
        except dbus.exceptions.DBusException as e:
            # Silently handle the case when the D-Bus signal handler cannot be set up
            # This is expected when onemount is not running or D-Bus service is not registered
            pass

    def _on_file_status_changed(self, path, status):
        """Handle file status change signals from D-Bus"""
        print(f"File status changed: {path} -> {status}")
        self.file_status_cache[path] = status

        # Request Nemo to refresh the file's emblems
        try:
            location = Gio.File.new_for_path(path)
            Nemo.FileInfo.invalidate_extension_info(location)
        except Exception as e:
            print(f"Error refreshing file emblems: {e}")

    def _get_onemount_mounts(self):
        """Get list of OneMount mount points"""
        mounts = []
        try:
            # Check if there are any mounted onemount filesystems
            with open('/proc/mounts', 'r') as f:
                for line in f:
                    if 'fuse.onemount' in line:
                        mount_point = line.split()[1]
                        mounts.append(mount_point)
        except Exception as e:
            print(f"Error getting OneMount mounts: {e}")
        return mounts

    def update_file_info(self, file, info=None, update_complete_callback=None):
        """Add emblems based on OneMount file status"""
        # Refresh the list of OneMount mounts
        self.onemount_mounts = self._get_onemount_mounts()

        # Check if the file is within a OneMount mount
        path = file.get_location().get_path()
        if not path:
            if update_complete_callback:
                update_complete_callback()
            return Nemo.OperationResult.COMPLETE

        for mount in self.onemount_mounts:
            if path.startswith(mount):
                # Query OneMount status for this file
                status = self._get_file_status(path)

                if info is not None:
                    if status == "Cloud":
                        info.add_emblem("emblem-synchronizing-offline")
                    elif status == "Local":
                        info.add_emblem("emblem-default")
                    elif status == "LocalModified":
                        info.add_emblem("emblem-synchronizing-locally-modified")
                    elif status == "Syncing":
                        info.add_emblem("emblem-synchronizing")
                    elif status == "Downloading":
                        info.add_emblem("emblem-downloads")
                    elif status == "OutofSync":
                        info.add_emblem("emblem-important")
                    elif status == "Error":
                        info.add_emblem("emblem-error")
                    elif status == "Conflict":
                        info.add_emblem("emblem-warning")
                    elif status == "Unknown":
                        info.add_emblem("emblem-question")
                    else:
                        # Default emblem for any unrecognized status
                        print(f"Unrecognized status: {status}")
                        info.add_emblem("emblem-question")

                break

        if update_complete_callback:
            update_complete_callback()
        return Nemo.OperationResult.COMPLETE

    def _get_file_status(self, path):
        """Get the OneMount status via D-Bus or extended attributes as fallback"""
        # First check if we have a cached status
        if path in self.file_status_cache:
            return self.file_status_cache[path]

        # Try to get status via D-Bus
        if self.dbus_proxy is not None:
            try:
                get_status = self.dbus_proxy.get_dbus_method(
                    'GetFileStatus',
                    'org.onemount.FileStatus'
                )
                status = get_status(path)
                self.file_status_cache[path] = status
                return status
            except dbus.exceptions.DBusException:
                # Silently handle D-Bus errors and fall back to extended attributes
                # This is expected when onemount is not running or D-Bus service is not registered
                # Try to reconnect for next time
                self.connect_to_dbus()
                # Fall back to extended attributes

        # Fallback: Get the status from extended attributes
        try:
            status = os.getxattr(path, "user.onemount.status")
            status_str = status.decode('utf-8')
            self.file_status_cache[path] = status_str
            return status_str
        except OSError as e:
            # Check if this is a filesystem limitation error (ENOTSUP, EOPNOTSUPP, ENOENT)
            # Errno 95 is ENOTSUP/EOPNOTSUPP (Operation not supported)
            # Errno 2 is ENOENT (No such file or directory)
            if e.errno in (95, 2):
                # Silently handle filesystem limitation errors
                return "Unknown"
            else:
                # Log other types of errors
                print(f"Error getting status for {path}: {e}")
                return "Unknown"
        except Exception as e:
            print(f"Error getting status for {path}: {e}")
            return "Unknown"

# Register the extension with Nemo
def module_init():
    return OneMountExtension()
