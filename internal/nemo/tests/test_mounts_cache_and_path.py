#!/usr/bin/env python3
"""
Tests for OneMount mount caching and path normalization logic.
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
def test_is_in_onemount_handles_trailing_slashes():
    with patch('dbus.mainloop.glib.DBusGMainLoop'), patch('dbus.SessionBus'):
        ext = nemo_onemount.OneMountExtension()
        # Simulate a mount with trailing slash
        ext.onemount_mounts = ["/mnt/om/"]
        assert ext._is_in_onemount("/mnt/om") is True
        assert ext._is_in_onemount("/mnt/om/sub/file.txt") is True
        assert ext._is_in_onemount("/mnt/om2/file.txt") is False


@pytest.mark.unit
def test_mounts_list_is_cached_with_ttl(monkeypatch):
    with patch('dbus.mainloop.glib.DBusGMainLoop'), patch('dbus.SessionBus'):
        ext = nemo_onemount.OneMountExtension()
        calls = {"n": 0}

        def fake_get_mounts():
            calls["n"] += 1
            return ["/mnt/om"]

        # Monkeypatch the underlying getter to count calls
        ext._get_onemount_mounts = fake_get_mounts

        # First call should populate the cache
        file1 = Mock()
        loc1 = Mock()
        loc1.get_path.return_value = "/mnt/om/file1.txt"
        file1.get_location.return_value = loc1
        info = Mock()
        info.add_emblem = Mock()
        ext.update_file_info(file1, info=info)

        # Second call within TTL should not re-read mounts
        file2 = Mock()
        loc2 = Mock()
        loc2.get_path.return_value = "/mnt/om/file2.txt"
        file2.get_location.return_value = loc2
        ext.update_file_info(file2, info=info)

        assert calls["n"] == 1

