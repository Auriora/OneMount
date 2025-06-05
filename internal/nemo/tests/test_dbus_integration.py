#!/usr/bin/env python3
"""
D-Bus integration tests for the OneMount Nemo Extension.

This module contains integration tests that verify the D-Bus communication
between the OneMount Go service and the Python Nemo extension.
"""

import os
import sys
import pytest
import time
import threading
import importlib.util
from unittest.mock import Mock, patch, MagicMock
from pathlib import Path

# Import the extension module
try:
    import nemo_onemount
except ImportError:
    # Add src directory to path and try again
    src_dir = Path(__file__).parent.parent / "src"
    sys.path.insert(0, str(src_dir))
    # Import with the actual filename
    spec = importlib.util.spec_from_file_location("nemo_onemount", src_dir / "nemo-onemount.py")
    nemo_onemount = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(nemo_onemount)


class TestDBusServiceConnection:
    """Test D-Bus service connection and basic functionality."""
    
    @pytest.mark.integration
    @pytest.mark.dbus
    def test_dbus_service_connection_success(self, mock_dbus):
        """Test successful connection to D-Bus service."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Verify connection was established
            assert extension.dbus_proxy is not None
            mock_dbus['session_bus'].assert_called_once()
            mock_dbus['bus'].get_object.assert_called_once_with(
                'org.onemount.FileStatus',
                '/org/onemount/FileStatus'
            )

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_dbus_service_connection_retry(self):
        """Test D-Bus service connection retry mechanism."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus:
            
            mock_bus = Mock()
            mock_session_bus.return_value = mock_bus
            
            # First call fails, second succeeds
            mock_proxy = Mock()
            mock_bus.get_object.side_effect = [Exception("Service not ready"), mock_proxy]
            
            # First attempt - should fail gracefully
            extension = nemo_onemount.OneMountExtension()
            assert extension.dbus_proxy is None
            
            # Retry connection - should succeed
            extension.connect_to_dbus()
            assert extension.dbus_proxy is not None

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_dbus_signal_setup_success(self, mock_dbus):
        """Test successful D-Bus signal handler setup."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Verify signal handler was set up
            mock_dbus['bus'].add_signal_receiver.assert_called_once_with(
                extension._on_file_status_changed,
                dbus_interface='org.onemount.FileStatus',
                signal_name='FileStatusChanged'
            )

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_dbus_signal_setup_failure(self):
        """Test D-Bus signal handler setup failure handling."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus:
            
            mock_bus = Mock()
            mock_session_bus.return_value = mock_bus
            mock_bus.get_object.return_value = Mock()
            mock_bus.add_signal_receiver.side_effect = Exception("Signal setup failed")
            
            # Should not raise exception
            extension = nemo_onemount.OneMountExtension()
            assert extension.dbus_proxy is not None


class TestDBusMethodCalls:
    """Test D-Bus method calls and responses."""
    
    @pytest.mark.integration
    @pytest.mark.dbus
    def test_get_file_status_method_call(self, mock_dbus):
        """Test GetFileStatus D-Bus method call."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Configure mock method
            mock_method = Mock(return_value="Downloading")
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            test_path = "/test/mount/path/file.txt"
            status = extension._get_file_status(test_path)
            
            # Verify method was called correctly
            mock_dbus['proxy'].get_dbus_method.assert_called_with(
                'GetFileStatus',
                'org.onemount.FileStatus'
            )
            mock_method.assert_called_with(test_path)
            assert status == "Downloading"
            
            # Verify status was cached
            assert extension.file_status_cache[test_path] == "Downloading"

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_get_file_status_method_timeout(self, mock_dbus):
        """Test GetFileStatus method call timeout handling."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('os.getxattr', return_value=b"Local"):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Configure mock method to timeout
            import dbus.exceptions
            mock_method = Mock(side_effect=dbus.exceptions.DBusException("Timeout"))
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            test_path = "/test/mount/path/file.txt"
            status = extension._get_file_status(test_path)
            
            # Should fall back to xattr
            assert status == "Local"

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_get_file_status_invalid_response(self, mock_dbus):
        """Test handling of invalid D-Bus method responses."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('os.getxattr', return_value=b"Unknown"):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Configure mock method to return invalid data
            mock_method = Mock(return_value=None)
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            test_path = "/test/mount/path/file.txt"
            status = extension._get_file_status(test_path)
            
            # Should handle gracefully and cache the result
            assert status is None or status == "Unknown"


class TestDBusSignalHandling:
    """Test D-Bus signal reception and handling."""
    
    @pytest.mark.integration
    @pytest.mark.dbus
    def test_file_status_changed_signal_reception(self, mock_dbus):
        """Test reception and processing of FileStatusChanged signals."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('gi.repository.Gio.File.new_for_path') as mock_new_for_path, \
             patch('gi.repository.Nemo.FileInfo.invalidate_extension_info') as mock_invalidate:
            
            extension = nemo_onemount.OneMountExtension()
            
            # Simulate signal reception
            test_path = "/test/mount/path/updated_file.txt"
            test_status = "LocalModified"
            
            extension._on_file_status_changed(test_path, test_status)
            
            # Verify cache was updated
            assert extension.file_status_cache[test_path] == test_status
            
            # Verify Nemo refresh was triggered
            mock_new_for_path.assert_called_once_with(test_path)
            mock_invalidate.assert_called_once()

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_multiple_signal_reception(self, mock_dbus):
        """Test handling of multiple rapid signal updates."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('gi.repository.Gio.File.new_for_path'), \
             patch('gi.repository.Nemo.FileInfo.invalidate_extension_info'):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Simulate multiple rapid updates
            test_paths = [
                "/test/mount/path/file1.txt",
                "/test/mount/path/file2.txt", 
                "/test/mount/path/file3.txt"
            ]
            test_statuses = ["Syncing", "Local", "Error"]
            
            for path, status in zip(test_paths, test_statuses):
                extension._on_file_status_changed(path, status)
            
            # Verify all updates were cached
            for path, status in zip(test_paths, test_statuses):
                assert extension.file_status_cache[path] == status

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_signal_handling_with_nemo_error(self, mock_dbus):
        """Test signal handling when Nemo refresh fails."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('gi.repository.Gio.File.new_for_path', side_effect=Exception("Nemo error")):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Should not raise exception even if Nemo refresh fails
            test_path = "/test/mount/path/file.txt"
            test_status = "Conflict"
            
            extension._on_file_status_changed(test_path, test_status)
            
            # Cache should still be updated
            assert extension.file_status_cache[test_path] == test_status


class TestDBusServiceAvailability:
    """Test handling of D-Bus service availability changes."""
    
    @pytest.mark.integration
    @pytest.mark.dbus
    def test_service_unavailable_at_startup(self):
        """Test extension behavior when D-Bus service is unavailable at startup."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus:
            
            mock_bus = Mock()
            mock_session_bus.return_value = mock_bus
            mock_bus.get_object.side_effect = Exception("Service not available")
            
            extension = nemo_onemount.OneMountExtension()
            
            # Should handle gracefully
            assert extension.dbus_proxy is None
            assert extension.bus is not None

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_service_becomes_unavailable(self, mock_dbus):
        """Test handling when D-Bus service becomes unavailable during operation."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('os.getxattr', return_value=b"Local"):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Initially working
            mock_method = Mock(return_value="Syncing")
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            status1 = extension._get_file_status("/test/path/file1.txt")
            assert status1 == "Syncing"
            
            # Service becomes unavailable
            import dbus.exceptions
            mock_method.side_effect = dbus.exceptions.DBusException("Service unavailable")
            
            with patch.object(extension, 'connect_to_dbus') as mock_reconnect:
                status2 = extension._get_file_status("/test/path/file2.txt")
                
                # Should attempt reconnection
                mock_reconnect.assert_called_once()
                # Should fall back to xattr
                assert status2 == "Local"

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_service_reconnection_success(self, mock_dbus):
        """Test successful reconnection to D-Bus service."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Simulate service becoming unavailable
            extension.dbus_proxy = None
            
            # Reconnect
            extension.connect_to_dbus()
            
            # Should be connected again
            assert extension.dbus_proxy is not None


class TestDBusIntegrationEndToEnd:
    """End-to-end D-Bus integration tests."""
    
    @pytest.mark.integration
    @pytest.mark.dbus
    @pytest.mark.slow
    def test_full_dbus_workflow(self, mock_dbus, mock_proc_mounts, mock_file_object, mock_file_info):
        """Test complete D-Bus workflow from file info request to emblem assignment."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Set up file within OneMount mount
            test_path = f"{mock_proc_mounts}/test/document.pdf"
            mock_file_object.get_location().get_path.return_value = test_path
            
            # Configure D-Bus response
            mock_method = Mock(return_value="Downloading")
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            # Process file info update
            result = extension.update_file_info(mock_file_object, mock_file_info)
            
            # Verify complete workflow
            assert result == "complete"
            mock_dbus['proxy'].get_dbus_method.assert_called_with(
                'GetFileStatus',
                'org.onemount.FileStatus'
            )
            mock_method.assert_called_with(test_path)
            mock_file_info.add_emblem.assert_called_once_with("emblem-downloads")
            
            # Verify caching
            assert extension.file_status_cache[test_path] == "Downloading"

    @pytest.mark.integration
    @pytest.mark.dbus
    def test_dbus_to_xattr_fallback_workflow(self, mock_dbus, mock_proc_mounts, 
                                           mock_file_object, mock_file_info):
        """Test workflow when D-Bus fails and falls back to extended attributes."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('os.getxattr', return_value=b"Local"):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Set up file within OneMount mount
            test_path = f"{mock_proc_mounts}/test/document.pdf"
            mock_file_object.get_location().get_path.return_value = test_path
            
            # Configure D-Bus to fail
            import dbus.exceptions
            mock_method = Mock(side_effect=dbus.exceptions.DBusException("Service error"))
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            with patch.object(extension, 'connect_to_dbus') as mock_reconnect:
                # Process file info update
                result = extension.update_file_info(mock_file_object, mock_file_info)
                
                # Verify fallback workflow
                assert result == "complete"
                mock_reconnect.assert_called_once()
                mock_file_info.add_emblem.assert_called_once_with("emblem-default")
                
                # Verify caching
                assert extension.file_status_cache[test_path] == "Local"
