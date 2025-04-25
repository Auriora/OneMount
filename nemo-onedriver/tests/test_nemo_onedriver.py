#!/usr/bin/env python3

import os
import sys
import unittest
import unittest.mock as mock
import tempfile
import pytest
import dbus
import dbus.mainloop.glib

# Mock GLib since it might not be available in the test environment
class MockGLib:
    MainLoop = mock.MagicMock

    @staticmethod
    def markup_escape_text(text):
        # Simple mock that returns the input text unchanged
        return text

    @staticmethod
    def get_application_name():
        # Return a dummy application name
        return "nemo-onedriver-test"

    @staticmethod
    def set_application_name(name):
        # Mock implementation that does nothing
        pass

    @staticmethod
    def get_prgname():
        # Return a dummy program name
        return "nemo-onedriver-test"

# Mock the gi.repository.GLib module
sys.modules['gi.repository.GLib'] = MockGLib

# Add the current directory to the path so we can import nemo-onedriver.py
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

# Mock the Nemo module since it's not available in the test environment
class MockNemo:
    class InfoProvider:
        pass

    class OperationResult:
        COMPLETE = 0

    class FileInfo:
        @staticmethod
        def invalidate_extension_info(location):
            pass

# Mock the Gio module
class MockGio:
    class File:
        @staticmethod
        def new_for_path(path):
            return MockFile(path)

class MockFile:
    def __init__(self, path):
        self.path = path

    def get_path(self):
        return self.path

# Create mocks for the modules
sys.modules['gi.repository.Nemo'] = MockNemo
sys.modules['gi.repository.Gio'] = MockGio

# Now we can import the nemo-onedriver module
# Python module names can't have hyphens, so we need to use importlib
import importlib.util

# Mock dbus.mainloop.glib.DBusGMainLoop to prevent D-Bus initialization during import
original_dbus_mainloop = dbus.mainloop.glib.DBusGMainLoop
dbus.mainloop.glib.DBusGMainLoop = lambda set_as_default=False: None

# Import the module
spec = importlib.util.spec_from_file_location("nemo_onedriver", "../src/nemo-onedriver.py")
nemo_onedriver = importlib.util.module_from_spec(spec)
spec.loader.exec_module(nemo_onedriver)

# Restore the original DBusGMainLoop
dbus.mainloop.glib.DBusGMainLoop = original_dbus_mainloop

# Add the module to sys.modules
sys.modules['nemo_onedriver'] = nemo_onedriver

# Mock the OneDriverExtension.__init__ method to prevent D-Bus connection and mount point detection
original_init = nemo_onedriver.OneDriverExtension.__init__
def mock_init(self):
    # Skip D-Bus initialization and mount point detection
    self.bus = None
    self.dbus_proxy = None
    self.file_status_cache = {}
    self.onedriver_mounts = []
nemo_onedriver.OneDriverExtension.__init__ = mock_init

class TestOneDriverExtension(unittest.TestCase):
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
        self.extension = nemo_onedriver.OneDriverExtension()

        # Mock the _get_onedriver_mounts method to return our temp directory
        self.extension._get_onedriver_mounts = mock.MagicMock(
            return_value=[os.path.dirname(self.temp_file.name)]
        )

    def tearDown(self):
        self.mock_bus_patcher.stop()
        os.unlink(self.temp_file.name)

    def test_connect_to_dbus(self):
        """Test connecting to the D-Bus service"""
        # Reset the proxy to test connection
        self.extension.dbus_proxy = None

        # Call the connect method
        self.extension.connect_to_dbus()

        # Verify that the method tried to connect to the D-Bus service
        self.mock_bus.return_value.get_object.assert_called_with(
            'org.onedriver.FileStatus',
            '/org/onedriver/FileStatus'
        )

        # Verify that the proxy was set
        self.assertIsNotNone(self.extension.dbus_proxy)

    def test_setup_dbus_signals(self):
        """Test setting up D-Bus signal handlers"""
        # Reset the signal handler
        self.extension.setup_dbus_signals()

        # Verify that the method tried to add a signal receiver
        self.mock_bus.return_value.add_signal_receiver.assert_called_with(
            self.extension._on_file_status_changed,
            dbus_interface='org.onedriver.FileStatus',
            signal_name='FileStatusChanged'
        )

    def test_get_file_status_dbus(self):
        """Test getting file status via D-Bus"""
        # Call the method
        status = self.extension._get_file_status(self.temp_file.name)

        # Verify that the method tried to get the status via D-Bus
        self.mock_proxy.get_dbus_method.assert_called_with(
            'GetFileStatus',
            'org.onedriver.FileStatus'
        )
        self.mock_get_status.assert_called_with(self.temp_file.name)

        # Verify that the status was returned
        self.assertEqual(status, "Local")

    def test_get_file_status_fallback(self):
        """Test falling back to extended attributes if D-Bus fails"""
        # Make the D-Bus method raise an exception
        self.mock_get_status.side_effect = dbus.exceptions.DBusException("Test error")

        # Mock the os.getxattr function
        with mock.patch('os.getxattr', return_value=b"Cloud"):
            # Call the method
            status = self.extension._get_file_status(self.temp_file.name)

            # Verify that the method tried to get the status via D-Bus
            self.mock_proxy.get_dbus_method.assert_called_with(
                'GetFileStatus',
                'org.onedriver.FileStatus'
            )
            self.mock_get_status.assert_called_with(self.temp_file.name)

            # Verify that the method fell back to extended attributes
            # and returned the correct status
            self.assertEqual(status, "Cloud")

    def test_on_file_status_changed(self):
        """Test handling file status change signals"""
        # Mock the Nemo.FileInfo.invalidate_extension_info method
        with mock.patch('nemo_onedriver.Nemo.FileInfo.invalidate_extension_info') as mock_invalidate:
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

        # Verify that the method returned COMPLETE
        self.assertEqual(result, MockNemo.OperationResult.COMPLETE)

if __name__ == '__main__':
    pytest.main(['-xvs', __file__])
