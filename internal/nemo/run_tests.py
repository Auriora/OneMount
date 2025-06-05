#!/usr/bin/env python3
"""
Simple test runner for OneMount Nemo Extension tests.

This is a lightweight wrapper that provides the same functionality as the
integrated CLI but can be used independently.

For the full-featured CLI, use: scripts/dev.py test nemo
"""

import os
import sys
import subprocess
import argparse
from pathlib import Path


def run_tests(
    test_type="all",
    verbose=False,
    coverage=False,
    test_file=None,
    test_function=None,
    pytest_args=None,
):
    """Run Nemo extension tests."""
    # Build pytest command
    cmd = ["python3", "-m", "pytest"]
    
    # Add test directory
    cmd.append("tests/")
    
    # Add verbosity
    if verbose:
        cmd.append("-v")
    
    # Add specific test markers
    if test_type == "unit":
        cmd.extend(["-m", "unit"])
    elif test_type == "integration":
        cmd.extend(["-m", "integration"])
    elif test_type == "dbus":
        cmd.extend(["-m", "dbus"])
    elif test_type == "mock":
        cmd.extend(["-m", "mock"])
    # "all" runs everything by default
    
    # Add coverage if requested
    if coverage:
        cmd.extend([
            "--cov=../src",
            "--cov-report=html",
            "--cov-report=term-missing",
            "--cov-report=xml"
        ])
    
    # Add specific test file if provided
    if test_file:
        cmd.append(f"tests/{test_file}")
    
    # Add specific test function if provided
    if test_function:
        if test_file:
            cmd[-1] += f"::{test_function}"
        else:
            print("Error: --test-function requires --test-file")
            return False
    
    # Add extra pytest arguments
    if pytest_args:
        cmd.extend(pytest_args.split())
    
    print(f"Running command: {' '.join(cmd)}")
    
    # Run the tests
    try:
        result = subprocess.run(cmd, cwd=Path(__file__).parent)
        return result.returncode == 0
    except KeyboardInterrupt:
        print("\nTests interrupted by user")
        return False
    except Exception as e:
        print(f"Error running tests: {e}")
        return False


def check_dependencies():
    """Check if required dependencies are installed."""
    required_packages = ["pytest"]
    missing_packages = []
    
    for package in required_packages:
        try:
            __import__(package)
        except ImportError:
            missing_packages.append(package)
    
    if missing_packages:
        print("Missing required packages:")
        for package in missing_packages:
            print(f"  - {package}")
        print("\nInstall them with:")
        print(f"  pip install {' '.join(missing_packages)}")
        return False
    
    return True


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Run OneMount Nemo Extension tests",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s                          # Run all tests
  %(prog)s --unit-only              # Run only unit tests
  %(prog)s --integration-only       # Run only integration tests
  %(prog)s --dbus-only              # Run only D-Bus tests
  %(prog)s --mock-only              # Run only mock tests
  %(prog)s --coverage               # Run with coverage reporting
  %(prog)s --test-file test_simple.py  # Run specific test file

Note: For full CLI features, use: scripts/dev.py test nemo
        """
    )
    
    parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Verbose output"
    )
    
    parser.add_argument(
        "--unit-only",
        action="store_true",
        help="Run only unit tests"
    )
    
    parser.add_argument(
        "--integration-only",
        action="store_true",
        help="Run only integration tests"
    )
    
    parser.add_argument(
        "--dbus-only",
        action="store_true",
        help="Run only D-Bus tests"
    )
    
    parser.add_argument(
        "--mock-only",
        action="store_true",
        help="Run only mock tests"
    )
    
    parser.add_argument(
        "--coverage",
        action="store_true",
        help="Generate coverage report"
    )
    
    parser.add_argument(
        "--test-file",
        help="Run specific test file (e.g., test_simple.py)"
    )
    
    parser.add_argument(
        "--test-function",
        help="Run specific test function (requires --test-file)"
    )
    
    parser.add_argument(
        "--pytest-args",
        help="Additional arguments to pass to pytest"
    )
    
    parser.add_argument(
        "--check-deps",
        action="store_true",
        help="Check if required dependencies are installed"
    )
    
    args = parser.parse_args()
    
    if args.check_deps:
        if check_dependencies():
            print("All required dependencies are installed.")
            return 0
        else:
            return 1
    
    # Check dependencies before running tests
    if not check_dependencies():
        return 1
    
    # Determine test type
    test_type = "all"
    if args.unit_only:
        test_type = "unit"
    elif args.integration_only:
        test_type = "integration"
    elif args.dbus_only:
        test_type = "dbus"
    elif args.mock_only:
        test_type = "mock"
    
    success = run_tests(
        test_type=test_type,
        verbose=args.verbose,
        coverage=args.coverage,
        test_file=args.test_file,
        test_function=args.test_function,
        pytest_args=args.pytest_args
    )
    
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())
