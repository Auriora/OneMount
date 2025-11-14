#!/usr/bin/env python3
"""
Tests for D-Bus service name discovery in the Nemo extension.

This test verifies the service discovery logic without requiring the full Nemo environment.
"""

import os
import tempfile
import unittest


def discover_dbus_service_name(service_name_file='/tmp/onemount-dbus-service-name'):
    """
    Standalone function to discover the D-Bus service name from the service name file.
    This is the same logic used in the Nemo extension.
    """
    try:
        with open(service_name_file, 'r') as f:
            service_name = f.read().strip()
            if service_name:
                return service_name
    except (FileNotFoundError, IOError):
        # File doesn't exist or can't be read, fall back to base name
        pass
    # Fall back to base service name (without unique suffix)
    return 'org.onemount.FileStatus'


class TestServiceDiscovery(unittest.TestCase):
    """Test D-Bus service name discovery functionality"""

    def setUp(self):
        """Set up test fixtures"""
        # Create a temporary file for testing
        self.temp_fd, self.temp_file = tempfile.mkstemp()
        
    def tearDown(self):
        """Clean up test fixtures"""
        # Remove the temporary file
        try:
            os.close(self.temp_fd)
            os.unlink(self.temp_file)
        except:
            pass

    def test_discover_service_name_from_file(self):
        """Test that service name is discovered from file"""
        # Write a service name to the temp file
        service_name = 'org.onemount.FileStatus.mnt_home-bcherrington-OneMountTest'
        with open(self.temp_file, 'w') as f:
            f.write(service_name + '\n')
        
        # Call the discovery function
        discovered_name = discover_dbus_service_name(self.temp_file)
        
        # Verify the discovered name matches what we wrote
        self.assertEqual(discovered_name, service_name)

    def test_discover_service_name_with_whitespace(self):
        """Test that service name is discovered correctly even with extra whitespace"""
        # Write a service name with extra whitespace
        service_name = 'org.onemount.FileStatus.mnt_tmp-onemount\\x20auth'
        with open(self.temp_file, 'w') as f:
            f.write('  ' + service_name + '  \n')
        
        # Call the discovery function
        discovered_name = discover_dbus_service_name(self.temp_file)
        
        # Verify the discovered name is trimmed correctly
        self.assertEqual(discovered_name, service_name)

    def test_discover_service_name_fallback_nonexistent(self):
        """Test that service name falls back to base name when file doesn't exist"""
        # Use a non-existent file path
        nonexistent_file = '/tmp/nonexistent-service-name-file-12345'
        
        # Call the discovery function
        discovered_name = discover_dbus_service_name(nonexistent_file)
        
        # Verify it falls back to the base name
        self.assertEqual(discovered_name, 'org.onemount.FileStatus')

    def test_discover_service_name_fallback_empty(self):
        """Test that service name falls back to base name when file is empty"""
        # Write an empty file
        with open(self.temp_file, 'w') as f:
            f.write('')
        
        # Call the discovery function
        discovered_name = discover_dbus_service_name(self.temp_file)
        
        # Verify it falls back to the base name
        self.assertEqual(discovered_name, 'org.onemount.FileStatus')

    def test_discover_service_name_fallback_whitespace_only(self):
        """Test that service name falls back to base name when file contains only whitespace"""
        # Write a file with only whitespace
        with open(self.temp_file, 'w') as f:
            f.write('   \n  \t  \n')
        
        # Call the discovery function
        discovered_name = discover_dbus_service_name(self.temp_file)
        
        # Verify it falls back to the base name
        self.assertEqual(discovered_name, 'org.onemount.FileStatus')


if __name__ == '__main__':
    unittest.main()
