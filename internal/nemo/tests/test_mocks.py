#!/usr/bin/env python3
"""
Mock tests for the OneMount Nemo Extension.

This module contains tests using mocked dependencies to verify behavior
in offline scenarios, service unavailability, and edge cases.
"""

import os
import sys
import pytest
import importlib.util
from unittest.mock import Mock, patch, MagicMock, PropertyMock
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


class TestMockDBusService:
    """Test extension behavior with mocked D-Bus service."""
    
    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_dbus_service_responses(self):
        """Test extension with various mocked D-Bus service responses."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus:
            
            mock_bus = Mock()
            mock_proxy = Mock()
            mock_session_bus.return_value = mock_bus
            mock_bus.get_object.return_value = mock_proxy
            
            # Test different status responses
            test_cases = [
                ("Cloud", "emblem-synchronizing-offline"),
                ("Local", "emblem-default"),
                ("Syncing", "emblem-synchronizing"),
                ("Error", "emblem-error"),
                ("Unknown", "emblem-question")
            ]
            
            extension = nemo_onemount.OneMountExtension()
            
            for status, expected_emblem in test_cases:
                mock_method = Mock(return_value=status)
                mock_proxy.get_dbus_method.return_value = mock_method
                
                result_status = extension._get_file_status(f"/test/path/{status.lower()}.txt")
                assert result_status == status

    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_dbus_service_errors(self):
        """Test extension handling of various D-Bus service errors."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus, \
             patch('os.getxattr', return_value=b"Local"):
            
            mock_bus = Mock()
            mock_proxy = Mock()
            mock_session_bus.return_value = mock_bus
            mock_bus.get_object.return_value = mock_proxy
            
            extension = nemo_onemount.OneMountExtension()
            
            # Test different error scenarios
            error_scenarios = [
                Exception("Connection timeout"),
                Exception("Service not available"),
                Exception("Method not found"),
                Exception("Invalid arguments")
            ]
            
            for error in error_scenarios:
                mock_method = Mock(side_effect=error)
                mock_proxy.get_dbus_method.return_value = mock_method
                
                with patch.object(extension, 'connect_to_dbus'):
                    status = extension._get_file_status("/test/path/error_test.txt")
                    # Should fall back to xattr
                    assert status == "Local"

    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_dbus_signal_simulation(self):
        """Test extension with simulated D-Bus signals."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus, \
             patch('gi.repository.Gio.File.new_for_path') as mock_new_for_path, \
             patch('gi.repository.Nemo.FileInfo.invalidate_extension_info') as mock_invalidate:
            
            mock_bus = Mock()
            mock_session_bus.return_value = mock_bus
            mock_bus.get_object.return_value = Mock()
            
            extension = nemo_onemount.OneMountExtension()
            
            # Simulate various signal scenarios
            signal_scenarios = [
                ("/test/path/file1.txt", "Syncing"),
                ("/test/path/file2.txt", "Local"),
                ("/test/path/file3.txt", "Error"),
                ("/test/path/file4.txt", "Conflict")
            ]
            
            for path, status in signal_scenarios:
                extension._on_file_status_changed(path, status)
                
                # Verify cache update
                assert extension.file_status_cache[path] == status
            
            # Verify Nemo refresh was called for each signal
            assert mock_new_for_path.call_count == len(signal_scenarios)
            assert mock_invalidate.call_count == len(signal_scenarios)


class TestMockFilesystem:
    """Test extension behavior with mocked filesystem operations."""
    
    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_proc_mounts_various_scenarios(self):
        """Test mount detection with various /proc/mounts scenarios."""
        test_scenarios = [
            # No OneMount mounts
            """
/dev/sda1 / ext4 rw,relatime 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev 0 0
""",
            # Single OneMount mount
            """
/dev/sda1 / ext4 rw,relatime 0 0
onemount /home/user/OneDrive fuse.onemount rw,nosuid,nodev,relatime 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev 0 0
""",
            # Multiple OneMount mounts
            """
/dev/sda1 / ext4 rw,relatime 0 0
onemount /home/user/OneDrive fuse.onemount rw,nosuid,nodev,relatime 0 0
onemount /mnt/shared fuse.onemount rw,nosuid,nodev,relatime 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev 0 0
""",
            # Malformed entries
            """
/dev/sda1 / ext4 rw,relatime 0 0
onemount
tmpfs /tmp tmpfs rw,nosuid,nodev 0 0
"""
        ]
        
        expected_counts = [0, 1, 2, 0]  # Expected number of mounts for each scenario
        
        for scenario, expected_count in zip(test_scenarios, expected_counts):
            with patch('dbus.mainloop.glib.DBusGMainLoop'), \
                 patch('dbus.SessionBus'), \
                 patch('builtins.open', create=True) as mock_open:
                
                mock_open.return_value.__enter__.return_value.__iter__ = lambda self: iter(scenario.strip().split('\n'))
                
                extension = nemo_onemount.OneMountExtension()
                mounts = extension._get_onemount_mounts()
                
                assert len(mounts) == expected_count

    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_xattr_operations(self):
        """Test extension with various extended attribute scenarios."""
        xattr_scenarios = [
            # Normal status values
            (b"Local", "Local"),
            (b"Cloud", "Cloud"),
            (b"Syncing", "Syncing"),
            (b"Error", "Error"),
            # Edge cases
            (b"", ""),
            (b"InvalidStatus", "InvalidStatus"),
            # Unicode content
            (b"Local\xc3\xa9", "Localé"),
        ]
        
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'):
            
            extension = nemo_onemount.OneMountExtension()
            extension.dbus_proxy = None  # Force xattr path
            
            for xattr_value, expected_status in xattr_scenarios:
                with patch('os.getxattr', return_value=xattr_value):
                    status = extension._get_file_status("/test/path/file.txt")
                    assert status == expected_status

    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_xattr_errors(self):
        """Test extension handling of extended attribute errors."""
        xattr_error_scenarios = [
            # Filesystem doesn't support xattrs
            OSError(95, "Operation not supported"),
            # File doesn't exist
            OSError(2, "No such file or directory"),
            # Permission denied
            OSError(13, "Permission denied"),
            # Other errors
            OSError(5, "Input/output error"),
        ]
        
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'):
            
            extension = nemo_onemount.OneMountExtension()
            extension.dbus_proxy = None  # Force xattr path
            
            for error in xattr_error_scenarios:
                with patch('os.getxattr', side_effect=error):
                    status = extension._get_file_status("/test/path/file.txt")
                    assert status == "Unknown"


class TestMockNemoIntegration:
    """Test extension behavior with mocked Nemo components."""
    
    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_nemo_file_objects(self):
        """Test extension with various mocked Nemo file objects."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Test different file object scenarios
            file_scenarios = [
                # Normal file
                ("/home/user/OneDrive/document.pdf", True),
                # File with no path
                (None, False),
                # File outside OneMount
                ("/home/user/Documents/local.txt", False),
                # File with special characters
                ("/home/user/OneDrive/файл.txt", True),
            ]
            
            for path, should_process in file_scenarios:
                mock_file = Mock()
                mock_info = Mock()
                
                if path:
                    mock_file.get_location().get_path.return_value = path
                else:
                    mock_file.get_location().get_path.return_value = None
                
                # Mock mount detection
                with patch.object(extension, '_get_onemount_mounts', return_value=["/home/user/OneDrive"]), \
                     patch.object(extension, '_get_file_status', return_value="Local"):
                    
                    result = extension.update_file_info(mock_file, mock_info)
                    
                    assert result == "complete"
                    if should_process and path and path.startswith("/home/user/OneDrive"):
                        mock_info.add_emblem.assert_called()
                    else:
                        mock_info.add_emblem.assert_not_called()
                
                # Reset mocks
                mock_info.reset_mock()

    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_nemo_callback_handling(self):
        """Test extension with mocked Nemo callback functions."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'):
            
            extension = nemo_onemount.OneMountExtension()
            
            mock_file = Mock()
            mock_info = Mock()
            mock_callback = Mock()
            
            mock_file.get_location().get_path.return_value = "/test/path/file.txt"
            
            # Test with callback
            result = extension.update_file_info(mock_file, mock_info, mock_callback)
            
            assert result == "complete"
            mock_callback.assert_called_once()
            
            # Test without callback
            result = extension.update_file_info(mock_file, mock_info)
            
            assert result == "complete"

    @pytest.mark.mock
    @pytest.mark.unit
    def test_mock_nemo_refresh_errors(self):
        """Test extension handling of Nemo refresh errors."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'), \
             patch('gi.repository.Gio.File.new_for_path', side_effect=Exception("Gio error")), \
             patch('gi.repository.Nemo.FileInfo.invalidate_extension_info', side_effect=Exception("Nemo error")):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Should handle errors gracefully
            extension._on_file_status_changed("/test/path/file.txt", "Local")
            
            # Cache should still be updated despite errors
            assert extension.file_status_cache["/test/path/file.txt"] == "Local"


class TestMockOfflineScenarios:
    """Test extension behavior in offline/disconnected scenarios."""
    
    @pytest.mark.mock
    @pytest.mark.unit
    def test_complete_offline_scenario(self):
        """Test extension behavior when completely offline."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus', side_effect=Exception("No D-Bus session")), \
             patch('os.getxattr', return_value=b"Local"):
            
            # Should initialize without D-Bus
            extension = nemo_onemount.OneMountExtension()
            assert extension.dbus_proxy is None
            
            # Should still work with xattrs
            status = extension._get_file_status("/test/path/file.txt")
            assert status == "Local"

    @pytest.mark.mock
    @pytest.mark.unit
    def test_service_unavailable_scenario(self):
        """Test extension when OneMount service is not running."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus, \
             patch('os.getxattr', return_value=b"Unknown"):
            
            mock_bus = Mock()
            mock_session_bus.return_value = mock_bus
            mock_bus.get_object.side_effect = Exception("Service not available")
            
            extension = nemo_onemount.OneMountExtension()
            assert extension.dbus_proxy is None
            
            # Should fall back to xattrs
            status = extension._get_file_status("/test/path/file.txt")
            assert status == "Unknown"

    @pytest.mark.mock
    @pytest.mark.unit
    def test_filesystem_limitations_scenario(self):
        """Test extension on filesystems that don't support extended attributes."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus', side_effect=Exception("No D-Bus")), \
             patch('os.getxattr', side_effect=OSError(95, "Operation not supported")):
            
            extension = nemo_onemount.OneMountExtension()
            
            # Should handle gracefully
            status = extension._get_file_status("/test/path/file.txt")
            assert status == "Unknown"
