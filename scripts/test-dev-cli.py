#!/usr/bin/env python3
"""
Test script for OneMount Development CLI Tool

This script performs basic validation of the CLI tool to ensure
it's working correctly and all commands are accessible.
"""

import subprocess
import sys
from pathlib import Path

def run_command(cmd):
    """Run a command and return success status."""
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.TimeoutExpired:
        return False, "", "Command timed out"
    except Exception as e:
        return False, "", str(e)

def test_cli_help():
    """Test that the CLI tool shows help correctly."""
    print("Testing CLI help...")
    
    cli_path = Path(__file__).parent / "onemount-dev.py"
    
    success, stdout, stderr = run_command([sys.executable, str(cli_path), "--help"])
    
    if success and "OneMount Development CLI Tool" in stdout:
        print("✓ CLI help works correctly")
        return True
    else:
        print(f"✗ CLI help failed: {stderr}")
        return False

def test_command_groups():
    """Test that all command groups are accessible."""
    print("Testing command groups...")
    
    cli_path = Path(__file__).parent / "onemount-dev.py"
    groups = ["build", "test", "release", "github", "analyze", "deploy"]
    
    all_passed = True
    
    for group in groups:
        success, stdout, stderr = run_command([sys.executable, str(cli_path), group, "--help"])
        
        if success:
            print(f"✓ {group} command group works")
        else:
            print(f"✗ {group} command group failed: {stderr}")
            all_passed = False
    
    return all_passed

def test_status_command():
    """Test the status command."""
    print("Testing status command...")
    
    cli_path = Path(__file__).parent / "onemount-dev.py"
    
    success, stdout, stderr = run_command([sys.executable, str(cli_path), "status"])
    
    if success and "OneMount Development Environment Status" in stdout:
        print("✓ Status command works correctly")
        return True
    else:
        print(f"✗ Status command failed: {stderr}")
        return False

def test_imports():
    """Test that all required imports are available."""
    print("Testing imports...")
    
    try:
        import click
        print("✓ click imported successfully")
    except ImportError:
        print("✗ click not available - install with: pip install click")
        return False
    
    try:
        import rich
        print("✓ rich imported successfully")
    except ImportError:
        print("✗ rich not available - install with: pip install rich")
        return False
    
    return True

def main():
    """Run all tests."""
    print("OneMount Development CLI Tool - Test Suite")
    print("=" * 50)
    
    tests = [
        test_imports,
        test_cli_help,
        test_command_groups,
        test_status_command,
    ]
    
    passed = 0
    total = len(tests)
    
    for test in tests:
        if test():
            passed += 1
        print()
    
    print("=" * 50)
    print(f"Test Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 All tests passed! The CLI tool is ready to use.")
        return 0
    else:
        print("❌ Some tests failed. Please check the output above.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
