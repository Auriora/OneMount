#!/usr/bin/env python3
"""
Pytest configuration and fixtures for OneMount Nemo Extension tests.

This module provides common fixtures and configuration for testing the
OneMount Nemo file manager extension.
"""

import os
import sys
import tempfile
import shutil
from unittest.mock import Mock, MagicMock, patch
from pathlib import Path
import pytest

# Add the src directory to Python path for imports
src_dir = Path(__file__).parent.parent / "src"
sys.path.insert(0, str(src_dir))

# Mock GI imports before any real imports using real module objects
import types

# Create gi module with require_version
gi_mod = types.ModuleType('gi')
setattr(gi_mod, 'require_version', lambda name, version: None)

# Create gi.repository package module
repo_mod = types.ModuleType('gi.repository')

# Nemo module
nemo_mod = types.ModuleType('Nemo')
class InfoProvider:
    def __init__(self, *args, **kwargs):
        pass
class MenuProvider:
    def __init__(self, *args, **kwargs):
        pass
class MenuItem:
    def __init__(self, name=None, label=None, tip=None, icon=None):
        self.name = name
        self.label = label
        self.tip = tip
        self.icon = icon
        self._callbacks = []
    def connect(self, signal, callback):
        self._callbacks.append(callback)
        return len(self._callbacks)
    def activate_for_test(self):
        for cb in self._callbacks:
            cb(self)
class FileInfo:
    @staticmethod
    def invalidate_extension_info(location):
        pass
class OperationResult:
    COMPLETE = "complete"
setattr(nemo_mod, 'InfoProvider', InfoProvider)
setattr(nemo_mod, 'MenuProvider', MenuProvider)
setattr(nemo_mod, 'MenuItem', MenuItem)
setattr(nemo_mod, 'FileInfo', FileInfo)
setattr(nemo_mod, 'OperationResult', OperationResult)

# GObject module
gobject_mod = types.ModuleType('GObject')
class _GObject:
    def __init__(self, *args, **kwargs):
        pass
setattr(gobject_mod, 'GObject', _GObject)

# Gio module
gio_mod = types.ModuleType('Gio')
class _File:
    @staticmethod
    def new_for_path(path):
        mock_file = Mock()
        mock_file.get_location.return_value.get_path.return_value = path
        return mock_file
setattr(gio_mod, 'File', _File)

# GLib module
glib_mod = types.ModuleType('GLib')

# Install into sys.modules
sys.modules['gi'] = gi_mod
sys.modules['gi.repository'] = repo_mod
sys.modules['gi.repository.Nemo'] = nemo_mod
sys.modules['gi.repository.GObject'] = gobject_mod
sys.modules['gi.repository.Gio'] = gio_mod
sys.modules['gi.repository.GLib'] = glib_mod

# Mock dbus modules
# Mock dbus modules with proper Exception types
exceptions_mod = types.ModuleType('dbus.exceptions')
class DBusException(Exception):
    pass
setattr(exceptions_mod, 'DBusException', DBusException)

mainloop_mod = types.ModuleType('dbus.mainloop')
glib_mainloop_mod = types.ModuleType('dbus.mainloop.glib')
# Link glib submodule under dbus.mainloop for attribute access
setattr(mainloop_mod, 'glib', glib_mainloop_mod)

class DBusGMainLoop:
    def __init__(self, *args, **kwargs):
        pass
setattr(glib_mainloop_mod, 'DBusGMainLoop', DBusGMainLoop)

dbus_mod = types.ModuleType('dbus')
# SessionBus will be patched in tests; define placeholder
class _SessionBus:
    def __init__(self, *args, **kwargs):
        raise Exception("No D-Bus session")
setattr(dbus_mod, 'SessionBus', _SessionBus)
# Link submodules on parent dbus module for attribute access
setattr(dbus_mod, 'mainloop', mainloop_mod)
setattr(dbus_mod, 'exceptions', exceptions_mod)


# Install dbus submodules
sys.modules['dbus'] = dbus_mod
sys.modules['dbus.mainloop'] = mainloop_mod
sys.modules['dbus.mainloop.glib'] = glib_mainloop_mod
sys.modules['dbus.exceptions'] = exceptions_mod


@pytest.fixture
def mock_dbus():
    """Provide a mock D-Bus interface."""
    with patch('dbus.SessionBus') as mock_session_bus:
        mock_bus = Mock()
        mock_proxy = Mock()

        # Configure the mock bus
        mock_session_bus.return_value = mock_bus
        mock_bus.get_object.return_value = mock_proxy
        mock_bus.add_signal_receiver = Mock()

        # Configure the mock proxy
        mock_proxy.get_dbus_method.return_value = Mock(return_value="Local")

        yield {
            'bus': mock_bus,
            'proxy': mock_proxy,
            'session_bus': mock_session_bus
        }

@pytest.fixture
def temp_mount_point():
    """Create a temporary directory to simulate a mount point."""
    temp_dir = tempfile.mkdtemp(prefix="onemount_test_")
    yield temp_dir
    shutil.rmtree(temp_dir, ignore_errors=True)

@pytest.fixture
def mock_proc_mounts(temp_mount_point):
    """Mock /proc/mounts with OneMount filesystem entries."""
    mount_content = f"""
/dev/sda1 / ext4 rw,relatime 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev 0 0
onemount {temp_mount_point} fuse.onemount rw,nosuid,nodev,relatime,user_id=1000,group_id=1000 0 0
"""

    with patch('builtins.open', create=True) as mock_open:
        mock_open.return_value.__enter__.return_value.read.return_value = mount_content.strip()
        mock_open.return_value.__enter__.return_value.__iter__ = lambda self: iter(mount_content.strip().split('\n'))
        yield temp_mount_point

@pytest.fixture
def mock_file_info():
    """Provide a mock Nemo FileInfo object."""
    mock_info = Mock()
    mock_info.add_emblem = Mock()
    return mock_info

@pytest.fixture
def mock_file_object():
    """Provide a mock Nemo File object."""
    mock_file = Mock()
    mock_location = Mock()
    mock_location.get_path.return_value = "/test/path/file.txt"
    mock_file.get_location.return_value = mock_location
    return mock_file

@pytest.fixture
def sample_file_statuses():
    """Provide sample file status mappings for testing."""
    return {
        "Cloud": "emblem-synchronizing-offline",
        "Local": "emblem-default",
        "LocalModified": "emblem-synchronizing-locally-modified",
        "Syncing": "emblem-synchronizing",
        "Downloading": "emblem-downloads",
        "OutofSync": "emblem-important",
        "Error": "emblem-error",
        "Conflict": "emblem-warning",
        "Unknown": "emblem-question"
    }

@pytest.fixture
def mock_xattr():
    """Mock extended attributes functionality."""
    with patch('os.getxattr') as mock_getxattr:
        mock_getxattr.return_value = b"Local"
        yield mock_getxattr

@pytest.fixture(autouse=True)
def setup_test_environment():
    """Set up the test environment before each test."""
    # Ensure clean state
    if 'nemo_onemount' in sys.modules:
        del sys.modules['nemo_onemount']

    yield

    # Cleanup after test
    if 'nemo_onemount' in sys.modules:
        del sys.modules['nemo_onemount']
