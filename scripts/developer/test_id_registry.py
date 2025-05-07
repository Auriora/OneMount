#!/usr/bin/env python3
"""
Test ID Registry for OneMount

This script manages a centralized registry of test IDs in the OneMount project. It helps prevent
duplicate test IDs by providing a way to check if a test ID is already in use, get the next
available test number, and register new test IDs.

Functionality:
    - Checks if a test ID is already in use
    - Gets the next available test number for a given test type, module, and feature number
    - Registers new test IDs
    - Lists all test IDs in the registry

Usage:
    ./test_id_registry.py [command] [args...]
    ./test_id_registry.py update [output_dir]
    ./test_id_registry.py check <test_id>
    ./test_id_registry.py next <test_type> <module> <feature_num>
    ./test_id_registry.py register <test_type> <module> <feature_num> <description> <expected_behavior>
    ./test_id_registry.py list [--test-type TYPE] [--module MODULE] [--feature-num NUM]

Arguments:
    command - Command to execute (update, check, next, register, list)
    output_dir - Optional directory to save the registry file (default: tmp/)
    test_id - Test ID to check
    test_type - Test type (UT, IT, etc.)
    module - Module code (FS, GR, etc.)
    feature_num - Feature number
    description - Test description
    expected_behavior - Expected behavior

Output:
    - A JSON file with the registry data (test_id_registry.json)
    - Console output with the results of the command

Test ID Structure:
    The test ID structure follows this pattern:
    <TYPE>_<COMPONENT>_<TESTNUMBER>_<SUBTESTNUMER>

    Where:
    - <TYPE> is the test type (2 letters):
      - UT - Unit Test
      - IT - Integration Test
      - ST - System Test
      - PT - Performance Test
      - LT - Load Test
      - SC - Scenario Test
      - UA - User Acceptance Test
      - etc.
    - <COMPONENT> is the component being tested (2/3 letters):
      - FS - File System
      - GR - Graph
      - UI - User Interface
      - CMD - Command
      - etc.
    - <TESTNUMBER> is a 2-digit number uniquely identifying the test
    - <SUBTESTNUMER> is a 2-digit number uniquely identifying the sub-test or test variant

Test Function Naming Convention:
    Test function names follow this pattern:
    Test<TYPE>_<COMPONENT>_<TESTNUMBER>_<SUBTESTNUMER>_<UNIT-OF-WORK>_<STATE-UNDER-TEST>_<EXPECTED-BEHAVIOR>

    Where:
    - <TYPE>, <COMPONENT>, <TESTNUMBER>, and <SUBTESTNUMER> are the same as in the test ID structure
    - <UNIT-OF-WORK> represents a single method, a class, or multiple classes
    - <STATE-UNDER-TEST> represents the inputs or conditions being tested
    - <EXPECTED-BEHAVIOR> represents the output or result

Examples:
    Checking if a Test ID is Already in Use:
        ./test_id_registry.py check UT_FS_01_01
        This will check if the test ID `UT_FS_01_01` is already in use.

    Getting the Next Available Test Number:
        ./test_id_registry.py next UT FS 01
        This will get the next available test number for unit tests in the file system module with feature number 01.

    Registering a New Test ID:
        ./test_id_registry.py register UT FS 01 FileOperations_BasicReadWrite SuccessfullyPreservesContent
        This will register a new test ID for a unit test in the file system module with feature number 01.

Best Practices:
    1. Always update the registry before using it: Run `./test_id_registry.py update` before checking or 
       registering test IDs to ensure the registry is up to date.

    2. Register test IDs before implementing tests: Register your test ID before implementing the test 
       to ensure it's reserved for your use.

    3. Use descriptive unit-of-work, state-under-test, and expected-behavior: These parts of the test 
       function name should clearly describe what the test is testing and what the expected outcome is.

    4. Follow the naming convention: Always follow the test function naming convention to ensure 
       consistency across the project.

    5. Check for existing tests with similar functionality: Before creating a new test, check if there 
       are existing tests with similar functionality that you can reuse or extend.

Troubleshooting:
    Registry Out of Sync:
        If the registry seems out of sync with the actual test IDs in the codebase, try updating the registry:
        ./test_id_registry.py update

    Test ID Already in Use:
        If you try to register a test ID that's already in use, the registry will tell you which test is using it. 
        You can either:
        1. Choose a different test ID
        2. Update the existing test to cover your use case
        3. If the existing test is obsolete, remove it and then register your test ID

    Missing Test IDs:
        If the registry doesn't contain a test ID that you know exists in the codebase, it might be because:
        1. The test file is not in one of the directories scanned by the registry
        2. The test function name doesn't follow the naming convention
        3. The registry hasn't been updated since the test was added
        Try updating the registry and check if the test file is in one of the scanned directories.

Note:
    All scripts save their output to the `tmp/` directory by default. This directory is created 
    automatically if it doesn't exist. You can specify a different output directory as a command-line argument.

Author: OneMount Team
"""

import os
import re
import sys
import json
import argparse
from collections import defaultdict
from datetime import datetime

# Define the directories to scan
DIRECTORIES = [
    "cmd/common",
    "internal/fs",
    "internal/fs/graph",
    "internal/fs/offline",
    "internal/ui"
]

# Define the output directory
OUTPUT_DIR = "tmp"

# Registry file name
REGISTRY_FILE_NAME = "test_id_registry.json"

# Regular expression to match test function declarations
TEST_FUNC_PATTERN = re.compile(r'func\s+(Test\w+)\s*\(\s*t\s+\*testing\.T\s*\)\s*{')

# Regular expression to extract test ID components
TEST_ID_PATTERN = re.compile(r'Test(UT|IT|ST|PT|LT|SC|UA)_(\w+)_(\d+)_(\d+)_([^(]+)')

class TestCase:
    def __init__(self, func_name, file_path, line_number):
        self.func_name = func_name
        self.file_path = file_path
        self.line_number = line_number

        # Extract test ID components
        match = TEST_ID_PATTERN.match(func_name)
        if match:
            self.test_type = match.group(1)  # UT, IT, etc.
            self.module = match.group(2)     # Module code (FS, GR, etc.)
            self.feature_num = match.group(3)  # Feature number
            self.test_num = match.group(4)     # Test case number

            # Extract description and expected behavior
            description_parts = match.group(5).split('_')
            if len(description_parts) >= 2:
                self.description = '_'.join(description_parts[:-1])
                self.expected_behavior = description_parts[-1]
            else:
                self.description = match.group(5)
                self.expected_behavior = ""

            # Create test ID
            self.test_id = f"{self.test_type}_{self.module}_{self.feature_num}_{self.test_num}"
        else:
            self.test_type = ""
            self.module = ""
            self.feature_num = ""
            self.test_num = ""
            self.description = func_name
            self.expected_behavior = ""
            self.test_id = func_name

    def to_dict(self):
        """Convert TestCase to dictionary for JSON serialization."""
        return {
            "test_id": self.test_id,
            "func_name": self.func_name,
            "file_path": self.file_path,
            "line_number": self.line_number,
            "test_type": self.test_type,
            "module": self.module,
            "feature_num": self.feature_num,
            "test_num": self.test_num,
            "description": self.description,
            "expected_behavior": self.expected_behavior
        }

    @classmethod
    def from_dict(cls, data):
        """Create TestCase from dictionary."""
        test_case = cls(data["func_name"], data["file_path"], data["line_number"])
        test_case.test_id = data["test_id"]
        test_case.test_type = data["test_type"]
        test_case.module = data["module"]
        test_case.feature_num = data["feature_num"]
        test_case.test_num = data["test_num"]
        test_case.description = data["description"]
        test_case.expected_behavior = data["expected_behavior"]
        return test_case

    def __str__(self):
        return f"{self.test_id}: {self.description} ({self.expected_behavior}) - {self.file_path}:{self.line_number}"

def scan_file(file_path):
    """Scan a file for test functions and return a list of TestCase objects."""
    test_cases = []

    try:
        with open(file_path, 'r') as f:
            lines = f.readlines()

        for i, line in enumerate(lines):
            match = TEST_FUNC_PATTERN.search(line)
            if match:
                func_name = match.group(1)
                test_case = TestCase(func_name, file_path, i + 1)
                test_cases.append(test_case)
    except Exception as e:
        print(f"Error scanning file {file_path}: {e}")

    return test_cases

def scan_directory(directory):
    """Recursively scan a directory for Go test files and return a list of TestCase objects."""
    test_cases = []

    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith('_test.go'):
                file_path = os.path.join(root, file)
                test_cases.extend(scan_file(file_path))

    return test_cases

def scan_all_directories():
    """Scan all directories for test cases."""
    all_test_cases = []
    for directory in DIRECTORIES:
        all_test_cases.extend(scan_directory(directory))
    return all_test_cases

def load_registry(output_dir=None):
    """Load the test ID registry from file."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    registry_file = os.path.join(output_dir, REGISTRY_FILE_NAME)

    if not os.path.exists(registry_file):
        return {"test_cases": [], "last_updated": ""}

    try:
        with open(registry_file, 'r') as f:
            registry = json.load(f)

        # Convert dictionaries back to TestCase objects
        registry["test_cases"] = [TestCase.from_dict(tc) for tc in registry["test_cases"]]
        return registry
    except Exception as e:
        print(f"Error loading registry: {e}")
        return {"test_cases": [], "last_updated": ""}

def save_registry(registry, output_dir=None):
    """Save the test ID registry to file."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    # Convert TestCase objects to dictionaries
    registry_copy = {
        "test_cases": [tc.to_dict() for tc in registry["test_cases"]],
        "last_updated": datetime.now().isoformat()
    }

    registry_file = os.path.join(output_dir, REGISTRY_FILE_NAME)

    try:
        with open(registry_file, 'w') as f:
            json.dump(registry_copy, f, indent=2)
        print(f"Registry saved to {registry_file}")
    except Exception as e:
        print(f"Error saving registry: {e}")

def update_registry(output_dir=None):
    """Update the test ID registry with the latest test cases."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Scan all directories for test cases
    all_test_cases = scan_all_directories()

    # Create a new registry
    registry = {
        "test_cases": all_test_cases,
        "last_updated": datetime.now().isoformat()
    }

    # Save the registry
    save_registry(registry, output_dir)

    print(f"Registry updated with {len(all_test_cases)} test cases.")
    return registry

def check_test_id(test_id, output_dir=None):
    """Check if a test ID is already in use."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_registry(output_dir)

    # Find all test cases with the given test ID
    matching_cases = [tc for tc in registry["test_cases"] if tc.test_id == test_id]

    if matching_cases:
        print(f"Test ID '{test_id}' is already in use by the following test cases:")
        for tc in matching_cases:
            print(f"  {tc.func_name} - {tc.file_path}:{tc.line_number}")
        return True
    else:
        print(f"Test ID '{test_id}' is available.")
        return False

def get_next_available_test_number(test_type, module, feature_num, output_dir=None):
    """Get the next available test number for a given test type, module, and feature number."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_registry(output_dir)

    # Find all test cases with the given test type, module, and feature number
    matching_cases = [tc for tc in registry["test_cases"] 
                     if tc.test_type == test_type and tc.module == module and tc.feature_num == feature_num]

    # Get the highest test number
    max_test_num = 0
    for tc in matching_cases:
        try:
            test_num = int(tc.test_num)
            max_test_num = max(max_test_num, test_num)
        except ValueError:
            pass

    # Return the next available test number
    return max_test_num + 1

def register_test_id(test_type, module, feature_num, description, expected_behavior, output_dir=None):
    """Register a new test ID."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Get the next available test number
    test_num = get_next_available_test_number(test_type, module, feature_num, output_dir)

    # Create the test ID
    test_id = f"{test_type}_{module}_{feature_num}_{test_num:02d}"

    # Check if the test ID is already in use
    if check_test_id(test_id, output_dir):
        print(f"Error: Test ID '{test_id}' is already in use.")
        return None

    # Create the function name
    func_name = f"Test{test_id}_{description}_{expected_behavior}"

    # Create a new test case
    test_case = TestCase(func_name, "", 0)
    test_case.test_id = test_id
    test_case.test_type = test_type
    test_case.module = module
    test_case.feature_num = feature_num
    test_case.test_num = f"{test_num:02d}"
    test_case.description = description
    test_case.expected_behavior = expected_behavior

    # Load the registry
    registry = load_registry(output_dir)

    # Add the new test case
    registry["test_cases"].append(test_case)

    # Save the registry
    save_registry(registry, output_dir)

    print(f"Test ID '{test_id}' registered successfully.")
    print(f"Function name: {func_name}")

    return test_case

def list_test_ids(test_type=None, module=None, feature_num=None, output_dir=None):
    """List all test IDs in the registry, optionally filtered by test type, module, and feature number."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_registry(output_dir)

    # Filter test cases
    filtered_cases = registry["test_cases"]
    if test_type:
        filtered_cases = [tc for tc in filtered_cases if tc.test_type == test_type]
    if module:
        filtered_cases = [tc for tc in filtered_cases if tc.module == module]
    if feature_num:
        filtered_cases = [tc for tc in filtered_cases if tc.feature_num == feature_num]

    # Sort test cases by test ID
    filtered_cases.sort(key=lambda tc: tc.test_id)

    # Print test cases
    print(f"Found {len(filtered_cases)} test cases:")
    for tc in filtered_cases:
        print(f"  {tc.test_id}: {tc.func_name} - {tc.file_path}:{tc.line_number}")

    return filtered_cases

def main():
    """Main function."""
    parser = argparse.ArgumentParser(description="Test ID Registry")
    subparsers = parser.add_subparsers(dest="command", help="Command to execute")

    # Update registry command
    update_parser = subparsers.add_parser("update", help="Update the test ID registry")
    update_parser.add_argument("output_dir", nargs="?", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # Check test ID command
    check_parser = subparsers.add_parser("check", help="Check if a test ID is already in use")
    check_parser.add_argument("test_id", help="Test ID to check")
    check_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # Get next available test number command
    next_parser = subparsers.add_parser("next", help="Get the next available test number")
    next_parser.add_argument("test_type", help="Test type (UT, IT, etc.)")
    next_parser.add_argument("module", help="Module code (FS, GR, etc.)")
    next_parser.add_argument("feature_num", help="Feature number")
    next_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # Register test ID command
    register_parser = subparsers.add_parser("register", help="Register a new test ID")
    register_parser.add_argument("test_type", help="Test type (UT, IT, etc.)")
    register_parser.add_argument("module", help="Module code (FS, GR, etc.)")
    register_parser.add_argument("feature_num", help="Feature number")
    register_parser.add_argument("description", help="Test description")
    register_parser.add_argument("expected_behavior", help="Expected behavior")
    register_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # List test IDs command
    list_parser = subparsers.add_parser("list", help="List all test IDs")
    list_parser.add_argument("--test-type", help="Filter by test type (UT, IT, etc.)")
    list_parser.add_argument("--module", help="Filter by module code (FS, GR, etc.)")
    list_parser.add_argument("--feature-num", help="Filter by feature number")
    list_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    args = parser.parse_args()

    # Create output directory if it doesn't exist
    output_dir = getattr(args, "output_dir", OUTPUT_DIR)
    os.makedirs(output_dir, exist_ok=True)

    if args.command == "update":
        update_registry(args.output_dir)
    elif args.command == "check":
        check_test_id(args.test_id, args.output_dir)
    elif args.command == "next":
        next_num = get_next_available_test_number(args.test_type, args.module, args.feature_num, args.output_dir)
        print(f"Next available test number: {next_num:02d}")
    elif args.command == "register":
        register_test_id(args.test_type, args.module, args.feature_num, args.description, args.expected_behavior, args.output_dir)
    elif args.command == "list":
        list_test_ids(args.test_type, args.module, args.feature_num, args.output_dir)
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
