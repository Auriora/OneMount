#!/usr/bin/env python3
"""
Unit tests for the OneMount Nemo Extension.

This module contains comprehensive unit tests for the OneMountExtension class,
covering initialization, D-Bus integration, file status handling, emblem assignment,
and error scenarios.
"""

import os
import sys
import pytest
import importlib.util
from unittest.mock import Mock, patch, MagicMock, call
from pathlib import Path

# Import the extension module
try:
    import nemo_onemount
except ImportError:
    # Add src directory to path and try again
    src_dir = Path(__file__).parent.parent / "src"
    sys.path.insert(0, str(src_dir))
    # Import with the actual filename
    import importlib.util
    spec = importlib.util.spec_from_file_location("nemo_onemount", src_dir / "nemo-onemount.py")
    nemo_onemount = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(nemo_onemount)


class TestOneMountExtensionInitialization:
    """Test OneMountExtension initialization and setup."""
    
    @pytest.mark.unit
    def test_extension_initialization_success(self, mock_dbus, mock_proc_mounts):
        """Test successful extension initialization."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            assert extension is not None
            assert hasattr(extension, 'bus')
            assert hasattr(extension, 'dbus_proxy')
            assert hasattr(extension, 'file_status_cache')
            assert hasattr(extension, 'onemount_mounts')
            assert isinstance(extension.file_status_cache, dict)
            assert isinstance(extension.onemount_mounts, list)

    @pytest.mark.unit
    def test_extension_initialization_no_dbus(self, mock_proc_mounts):
        """Test extension initialization when D-Bus is not available."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus', side_effect=Exception("D-Bus not available")):
            
            extension = nemo_onemount.OneMountExtension()
            
            assert extension is not None
            assert extension.dbus_proxy is None

    @pytest.mark.unit
    def test_dbus_connection_success(self, mock_dbus):
        """Test successful D-Bus connection."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Verify D-Bus connection was attempted
            mock_dbus['session_bus'].assert_called_once()
            mock_dbus['bus'].get_object.assert_called_once_with(
                'org.onemount.FileStatus',
                '/org/onemount/FileStatus'
            )

    @pytest.mark.unit
    def test_dbus_connection_failure(self, mock_proc_mounts):
        """Test D-Bus connection failure handling."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus') as mock_session_bus:
            
            mock_bus = Mock()
            mock_session_bus.return_value = mock_bus
            mock_bus.get_object.side_effect = Exception("Service not available")
            
            extension = nemo_onemount.OneMountExtension()
            
            assert extension.dbus_proxy is None


class TestMountPointDetection:
    """Test OneMount mount point detection functionality."""
    
    @pytest.mark.unit
    def test_get_onemount_mounts_success(self, mock_proc_mounts):
        """Test successful detection of OneMount mounts."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'):
            
            extension = nemo_onemount.OneMountExtension()
            mounts = extension._get_onemount_mounts()
            
            assert isinstance(mounts, list)
            assert len(mounts) > 0
            assert mock_proc_mounts in mounts

    @pytest.mark.unit
    def test_get_onemount_mounts_no_mounts(self):
        """Test mount detection when no OneMount filesystems are mounted."""
        mount_content = """
/dev/sda1 / ext4 rw,relatime 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev 0 0
"""
        
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'), \
             patch('builtins.open', create=True) as mock_open:
            
            mock_open.return_value.__enter__.return_value.__iter__ = lambda self: iter(mount_content.strip().split('\n'))
            
            extension = nemo_onemount.OneMountExtension()
            mounts = extension._get_onemount_mounts()
            
            assert isinstance(mounts, list)
            assert len(mounts) == 0

    @pytest.mark.unit
    def test_get_onemount_mounts_file_error(self):
        """Test mount detection when /proc/mounts cannot be read."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'), \
             patch('builtins.open', side_effect=IOError("Permission denied")):
            
            extension = nemo_onemount.OneMountExtension()
            mounts = extension._get_onemount_mounts()
            
            assert isinstance(mounts, list)
            assert len(mounts) == 0


class TestFileStatusRetrieval:
    """Test file status retrieval functionality."""
    
    @pytest.mark.unit
    def test_get_file_status_dbus_success(self, mock_dbus):
        """Test successful file status retrieval via D-Bus."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Mock the D-Bus method call
            mock_method = Mock(return_value="Syncing")
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            status = extension._get_file_status("/test/path/file.txt")
            
            assert status == "Syncing"
            mock_dbus['proxy'].get_dbus_method.assert_called_with(
                'GetFileStatus',
                'org.onemount.FileStatus'
            )
            mock_method.assert_called_with("/test/path/file.txt")

    @pytest.mark.unit
    def test_get_file_status_cached(self, mock_dbus):
        """Test file status retrieval from cache."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Pre-populate cache
            test_path = "/test/path/file.txt"
            extension.file_status_cache[test_path] = "Local"
            
            status = extension._get_file_status(test_path)
            
            assert status == "Local"
            # Verify D-Bus was not called
            mock_dbus['proxy'].get_dbus_method.assert_not_called()

    @pytest.mark.unit
    def test_get_file_status_dbus_fallback_to_xattr(self, mock_dbus, mock_xattr):
        """Test fallback to extended attributes when D-Bus fails."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Mock D-Bus failure
            mock_method = Mock(side_effect=Exception("D-Bus error"))
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            status = extension._get_file_status("/test/path/file.txt")
            
            assert status == "Local"  # From mock_xattr fixture
            mock_xattr.assert_called_with("/test/path/file.txt", "user.onemount.status")

    @pytest.mark.unit
    def test_get_file_status_xattr_not_supported(self, mock_dbus):
        """Test handling when extended attributes are not supported."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('os.getxattr', side_effect=OSError(95, "Operation not supported")):
            
            extension = nemo_onemount.OneMountExtension()
            extension.dbus_proxy = None  # Force xattr path
            
            status = extension._get_file_status("/test/path/file.txt")
            
            assert status == "Unknown"

    @pytest.mark.unit
    def test_get_file_status_file_not_found(self, mock_dbus):
        """Test handling when file does not exist."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('os.getxattr', side_effect=OSError(2, "No such file or directory")):
            
            extension = nemo_onemount.OneMountExtension()
            extension.dbus_proxy = None  # Force xattr path
            
            status = extension._get_file_status("/test/path/nonexistent.txt")
            
            assert status == "Unknown"


class TestEmblemAssignment:
    """Test emblem assignment functionality."""
    
    @pytest.mark.unit
    def test_emblem_assignment_all_statuses(self, mock_dbus, mock_proc_mounts, 
                                          mock_file_object, mock_file_info, 
                                          sample_file_statuses):
        """Test emblem assignment for all known file statuses."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Set up the file to be within a OneMount mount
            mock_file_object.get_location().get_path.return_value = f"{mock_proc_mounts}/test/file.txt"
            
            for status, expected_emblem in sample_file_statuses.items():
                # Reset the mock
                mock_file_info.reset_mock()
                
                # Mock the status retrieval
                with patch.object(extension, '_get_file_status', return_value=status):
                    result = extension.update_file_info(mock_file_object, mock_file_info)
                    
                    assert result == "complete"  # Nemo.OperationResult.COMPLETE
                    mock_file_info.add_emblem.assert_called_once_with(expected_emblem)

    @pytest.mark.unit
    def test_emblem_assignment_unrecognized_status(self, mock_dbus, mock_proc_mounts,
                                                  mock_file_object, mock_file_info):
        """Test emblem assignment for unrecognized status."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Set up the file to be within a OneMount mount
            mock_file_object.get_location().get_path.return_value = f"{mock_proc_mounts}/test/file.txt"
            
            # Mock an unrecognized status
            with patch.object(extension, '_get_file_status', return_value="UnknownStatus"):
                result = extension.update_file_info(mock_file_object, mock_file_info)
                
                assert result == "complete"
                mock_file_info.add_emblem.assert_called_once_with("emblem-question")

    @pytest.mark.unit
    def test_no_emblem_for_non_onemount_files(self, mock_dbus, mock_file_object, mock_file_info):
        """Test that no emblem is assigned for files outside OneMount mounts."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Set up the file to be outside OneMount mounts
            mock_file_object.get_location().get_path.return_value = "/home/user/regular/file.txt"
            
            result = extension.update_file_info(mock_file_object, mock_file_info)
            
            assert result == "complete"
            mock_file_info.add_emblem.assert_not_called()

    @pytest.mark.unit
    def test_update_file_info_no_path(self, mock_dbus, mock_file_info):
        """Test handling when file has no path."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            mock_file = Mock()
            mock_file.get_location().get_path.return_value = None
            
            result = extension.update_file_info(mock_file, mock_file_info)
            
            assert result == "complete"
            mock_file_info.add_emblem.assert_not_called()


class TestSignalHandling:
    """Test D-Bus signal handling functionality."""
    
    @pytest.mark.unit
    def test_file_status_changed_signal(self, mock_dbus):
        """Test handling of file status change signals."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('gi.repository.Gio.File.new_for_path') as mock_new_for_path, \
             patch('gi.repository.Nemo.FileInfo.invalidate_extension_info') as mock_invalidate:
            
            extension = nemo_onemount.OneMountExtension()
            
            # Simulate a file status change signal
            test_path = "/test/path/file.txt"
            test_status = "Syncing"
            
            extension._on_file_status_changed(test_path, test_status)
            
            # Verify cache was updated
            assert extension.file_status_cache[test_path] == test_status
            
            # Verify Nemo was asked to refresh the file
            mock_new_for_path.assert_called_once_with(test_path)
            mock_invalidate.assert_called_once()

    @pytest.mark.unit
    def test_file_status_changed_signal_error(self, mock_dbus):
        """Test handling of errors during signal processing."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('gi.repository.Gio.File.new_for_path', side_effect=Exception("Gio error")):
            
            extension = nemo_onemount.OneMountExtension()
            
            # This should not raise an exception
            extension._on_file_status_changed("/test/path/file.txt", "Error")
            
            # Cache should still be updated
            assert extension.file_status_cache["/test/path/file.txt"] == "Error"


class TestErrorHandling:
    """Test error handling scenarios."""
    
    @pytest.mark.unit
    def test_dbus_reconnection_on_error(self, mock_dbus):
        """Test D-Bus reconnection when method calls fail."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'):
            extension = nemo_onemount.OneMountExtension()
            
            # Mock D-Bus method failure
            mock_method = Mock(side_effect=Exception("D-Bus connection lost"))
            mock_dbus['proxy'].get_dbus_method.return_value = mock_method
            
            with patch.object(extension, 'connect_to_dbus') as mock_reconnect, \
                 patch('os.getxattr', return_value=b"Local"):
                
                status = extension._get_file_status("/test/path/file.txt")
                
                # Should attempt reconnection
                mock_reconnect.assert_called_once()
                # Should fall back to xattr
                assert status == "Local"

    @pytest.mark.unit
    def test_module_init_function(self):
        """Test the module_init function."""
        with patch('dbus.mainloop.glib.DBusGMainLoop'), \
             patch('dbus.SessionBus'):
            
            extension = nemo_onemount.module_init()
            
            assert isinstance(extension, nemo_onemount.OneMountExtension)
