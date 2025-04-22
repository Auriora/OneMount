#!/usr/bin/python3

import gi
gi.require_version('Nemo', '3.0')
from gi.repository import Nemo, GObject, Gio
import os

class OneDriverExtension(GObject.GObject, Nemo.InfoProvider):
    def __init__(self):
        self.onedriver_mounts = self._get_onedriver_mounts()

    def _get_onedriver_mounts(self):
        """Get list of OneDriver mount points"""
        mounts = []
        try:
            # Check if there are any mounted onedriver filesystems
            with open('/proc/mounts', 'r') as f:
                for line in f:
                    if 'fuse.onedriver' in line:
                        mount_point = line.split()[1]
                        mounts.append(mount_point)
        except Exception as e:
            print(f"Error getting OneDriver mounts: {e}")
        return mounts

    def update_file_info(self, file, info, update_complete_callback):
        """Add emblems based on OneDriver file status"""
        # Check if the file is within a OneDriver mount
        path = file.get_location().get_path()
        if not path:
            update_complete_callback()
            return Nemo.OperationResult.COMPLETE

        for mount in self.onedriver_mounts:
            if path.startswith(mount):
                # Query OneDriver status for this file
                status = self._get_file_status(path)

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

                break

        update_complete_callback()
        return Nemo.OperationResult.COMPLETE

    def _get_file_status(self, path):
        """Get the OneDriver status from extended attributes"""
        try:
            # Get the status from extended attributes
            status = os.getxattr(path, "user.onedriver.status")
            return status.decode('utf-8')
        except:
            return "unknown"

# Register the extension with Nemo
def module_init():
    return OneDriverExtension()
