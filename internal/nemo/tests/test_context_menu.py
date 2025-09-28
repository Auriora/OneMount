#!/usr/bin/env python3
"""
Tests for Nemo context menu items provided by OneMountExtension.
"""

import sys
from unittest.mock import Mock, patch
from pathlib import Path

import pytest

# Import the extension module
try:
    import nemo_onemount
except ImportError:
    src_dir = Path(__file__).parent.parent / "src"
    sys.path.insert(0, str(src_dir))
    import importlib.util
    spec = importlib.util.spec_from_file_location("nemo_onemount", src_dir / "nemo-onemount.py")
    nemo_onemount = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(nemo_onemount)


@pytest.mark.unit
def test_get_file_items_inside_mount_returns_item(mock_proc_mounts):
    with patch('dbus.mainloop.glib.DBusGMainLoop'), patch('dbus.SessionBus'):
        ext = nemo_onemount.OneMountExtension()
        # Ensure mount list contains our mocked mount
        ext.onemount_mounts = [mock_proc_mounts]

        # Build a mock Nemo file within the mount
        mock_file = Mock()
        mock_loc = Mock()
        mock_loc.get_path.return_value = f"{mock_proc_mounts}/file.txt"
        mock_file.get_location.return_value = mock_loc

        items = ext.get_file_items(None, [mock_file])
        assert isinstance(items, list)
        assert len(items) >= 1
        first = items[0]
        # Our MockNemo.MenuItem stores label
        assert hasattr(first, 'label')
        assert 'Refresh' in first.label


@pytest.mark.unit
def test_get_file_items_outside_mount_returns_none(mock_proc_mounts):
    with patch('dbus.mainloop.glib.DBusGMainLoop'), patch('dbus.SessionBus'):
        ext = nemo_onemount.OneMountExtension()
        ext.onemount_mounts = [mock_proc_mounts]

        mock_file = Mock()
        mock_loc = Mock()
        mock_loc.get_path.return_value = "/home/user/Documents/file.txt"
        mock_file.get_location.return_value = mock_loc

        items = ext.get_file_items(None, [mock_file])
        assert items is None


@pytest.mark.unit
def test_activate_refresh_emblems_triggers_invalidate(mock_proc_mounts):
    with patch('dbus.mainloop.glib.DBusGMainLoop'), patch('dbus.SessionBus'), \
         patch('gi.repository.Gio.File.new_for_path') as mock_new_for_path, \
         patch('gi.repository.Nemo.FileInfo.invalidate_extension_info') as mock_invalidate:
        ext = nemo_onemount.OneMountExtension()
        ext.onemount_mounts = [mock_proc_mounts]

        mock_file = Mock()
        mock_loc = Mock()
        path = f"{mock_proc_mounts}/file.txt"
        mock_loc.get_path.return_value = path
        mock_file.get_location.return_value = mock_loc

        items = ext.get_file_items(None, [mock_file])
        assert items and hasattr(items[0], 'activate_for_test')
        items[0].activate_for_test()

        mock_new_for_path.assert_called_once_with(path)
        mock_invalidate.assert_called_once()

