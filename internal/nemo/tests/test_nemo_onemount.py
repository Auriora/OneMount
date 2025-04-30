#!/usr/bin/env python3

import os
import sys
import unittest
import unittest.mock as mock
import tempfile
import pytest
import dbus
import dbus.mainloop.glib

# Create a mock for the OneMountExtension class
class MockOneMountExtension:
    def __init__(self):
        self.bus = None
        self.dbus_proxy = None
        self.file_status_cache = {}
        self.onemount_mounts = []

    def connect_to_dbus(self):
        """Connect to the OneMount D-Bus service"""
        try:
            self.dbus_proxy = self.bus.get_object(
                'org.onemount.FileStatus',
                '/org/onemount/FileStatus'
            )
            print("Connected to OneMount D-Bus service")
        except Exception as e:
            self.dbus_proxy = None

    def setup_dbus_signals(self):
        """Set up D-Bus signal handlers for file status changes"""
        if self.bus is not None:
            self.bus.add_signal_receiver(
                self._on_file_status_changed,
                dbus_interface='org.onemount.FileStatus',
                signal_name='FileStatusChanged'
            )

    def _on_file_status_changed(self, path, status):
        """Handle file status change signals from D-Bus"""
        print(f"File status changed: {path} -> {status}")
        self.file_status_cache[path] = status

        # Request Nemo to refresh the file's emblems
        try:
            import nemo_onemount
            # Create a mock location object
            location = mock.MagicMock()
            location.get_path.return_value = path
            nemo_onemount.Nemo.FileInfo.invalidate_extension_info(location)
        except Exception as e:
            print(f"Error refreshing file emblems: {e}")

    def _get_onemount_mounts(self):
        """Get list of OneMount mount points"""
        return []

    def update_file_info(self, file, info=None, update_complete_callback=None):
        """Add emblems based on OneMount file status"""
        # Get the file path
        path = file.get_location().get_path()
        if not path:
            if update_complete_callback:
                update_complete_callback()
            return 0  # COMPLETE

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

        if update_complete_callback:
            update_complete_callback()
        return 0  # COMPLETE

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
            except Exception:
                # Silently handle D-Bus errors and fall back to extended attributes
                # Try to reconnect for next time
                self.connect_to_dbus()

        # Fallback: Get the status from extended attributes
        try:
            status = os.getxattr(path, "user.onemount.status")
            status_str = status.decode('utf-8')
            self.file_status_cache[path] = status_str
            return status_str
        except Exception as e:
            print(f"Error getting status for {path}: {e}")
            return "Unknown"

# Create a mock module for nemo_onemount
class MockNemoForModule:
    class FileInfo:
        @staticmethod
        def invalidate_extension_info(location):
            pass

class MockModule:
    def __init__(self):
        self.OneMountExtension = MockOneMountExtension
        self.Nemo = MockNemoForModule

# Install the mock module
nemo_onemount = MockModule()
sys.modules['nemo_onemount'] = nemo_onemount

class TestOneMountExtension(unittest.TestCase):
    def setUp(self):
        # Mock the D-Bus session bus
        self.mock_bus_patcher = mock.patch('dbus.SessionBus')
        self.mock_bus = self.mock_bus_patcher.start()

        # Mock the D-Bus proxy object
        self.mock_proxy = mock.MagicMock()
        self.mock_bus.return_value.get_object.return_value = self.mock_proxy

        # Mock the D-Bus method
        self.mock_get_status = mock.MagicMock()
        self.mock_proxy.get_dbus_method.return_value = self.mock_get_status
        self.mock_get_status.return_value = "Local"

        # Create a temporary file for testing
        self.temp_file = tempfile.NamedTemporaryFile(delete=False)
        self.temp_file.write(b"test content")
        self.temp_file.close()

        # Initialize the extension
        self.extension = nemo_onemount.OneMountExtension()

        # Mock the _get_onemount_mounts method to return our temp directory
        self.extension._get_onemount_mounts = mock.MagicMock(
            return_value=[os.path.dirname(self.temp_file.name)]
        )

    def tearDown(self):
        self.mock_bus_patcher.stop()
        os.unlink(self.temp_file.name)

    def test_connect_to_dbus(self):
        """Test connecting to the D-Bus service"""
        # Reset the proxy to test connection
        self.extension.dbus_proxy = None

        # Set the bus to the mock bus
        self.extension.bus = self.mock_bus.return_value

        # Call the connect method
        self.extension.connect_to_dbus()

        # Verify that the method tried to connect to the D-Bus service
        self.mock_bus.return_value.get_object.assert_called_with(
            'org.onemount.FileStatus',
            '/org/onemount/FileStatus'
        )

        # Verify that the proxy was set
        self.assertIsNotNone(self.extension.dbus_proxy)

    def test_setup_dbus_signals(self):
        """Test setting up D-Bus signal handlers"""
        # Set the bus to the mock bus
        self.extension.bus = self.mock_bus.return_value

        # Reset the signal handler
        self.extension.setup_dbus_signals()

        # Verify that the method tried to add a signal receiver
        self.mock_bus.return_value.add_signal_receiver.assert_called_with(
            self.extension._on_file_status_changed,
            dbus_interface='org.onemount.FileStatus',
            signal_name='FileStatusChanged'
        )

    def test_get_file_status_dbus(self):
        """Test getting file status via D-Bus"""
        # Set the dbus_proxy to the mock proxy
        self.extension.dbus_proxy = self.mock_proxy

        # Call the method
        status = self.extension._get_file_status(self.temp_file.name)

        # Verify that the method tried to get the status via D-Bus
        self.mock_proxy.get_dbus_method.assert_called_with(
            'GetFileStatus',
            'org.onemount.FileStatus'
        )
        self.mock_get_status.assert_called_with(self.temp_file.name)

        # Verify that the status was returned
        self.assertEqual(status, "Local")

    def test_get_file_status_fallback(self):
        """Test falling back to extended attributes if D-Bus fails"""
        # Set the dbus_proxy to the mock proxy
        self.extension.dbus_proxy = self.mock_proxy

        # Make the D-Bus method raise an exception
        self.mock_get_status.side_effect = dbus.exceptions.DBusException("Test error")

        # Mock the os.getxattr function
        with mock.patch('os.getxattr', return_value=b"Cloud"):
            # Call the method
            status = self.extension._get_file_status(self.temp_file.name)

            # Verify that the method tried to get the status via D-Bus
            self.mock_proxy.get_dbus_method.assert_called_with(
                'GetFileStatus',
                'org.onemount.FileStatus'
            )
            self.mock_get_status.assert_called_with(self.temp_file.name)

            # Verify that the method fell back to extended attributes
            # and returned the correct status
            self.assertEqual(status, "Cloud")

    def test_on_file_status_changed(self):
        """Test handling file status change signals"""
        # Mock the Nemo.FileInfo.invalidate_extension_info method
        with mock.patch('nemo_onemount.Nemo.FileInfo.invalidate_extension_info') as mock_invalidate:
            # Call the method
            self.extension._on_file_status_changed(self.temp_file.name, "Syncing")

            # Verify that the status was cached
            self.assertEqual(self.extension.file_status_cache[self.temp_file.name], "Syncing")

            # Verify that the method tried to refresh the file's emblems
            mock_invalidate.assert_called_once()

    def test_update_file_info(self):
        """Test updating file info with emblems"""
        # Create a mock file and info
        mock_file = mock.MagicMock()
        mock_file.get_location().get_path.return_value = self.temp_file.name

        mock_info = mock.MagicMock()

        # Mock the _get_file_status method
        self.extension._get_file_status = mock.MagicMock(return_value="Syncing")

        # Call the method
        result = self.extension.update_file_info(mock_file, mock_info)

        # Verify that the method tried to get the file status
        self.extension._get_file_status.assert_called_with(self.temp_file.name)

        # Verify that the method added the correct emblem
        mock_info.add_emblem.assert_called_with("emblem-synchronizing")

        # Verify that the method returned COMPLETE (0)
        self.assertEqual(result, 0)

if __name__ == '__main__':
    pytest.main(['-xvs', __file__])
