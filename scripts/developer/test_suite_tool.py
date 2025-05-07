#!/usr/bin/env python3
"""
Test Suite Tool for OneMount

This script provides tools for analyzing and managing the OneMount test suite.
It combines the functionality of the former test_analyzer.py and test_id_resolver.py scripts.

Functionality:
    Analysis Mode:
        - Identifies duplicate test IDs
        - Finds tests with functional overlap
        - Detects tests with ambiguous boundaries
        - Identifies gaps in sequential test case numbering
    Resolution Mode:
        - Identifies duplicate test IDs
        - Suggests new test IDs for tests with duplicate IDs
        - Generates a shell script to rename the tests

Usage:
    ./test_suite_tool.py [--analyze | --resolve] [output_dir]

Arguments:
    --analyze  - Run in analysis mode to identify various issues in the test suite (default)
    --resolve  - Run in resolution mode to resolve duplicate test IDs
    output_dir - Optional directory to save the output files (default: tmp/)

Output:
    Analysis mode:
        - A report with the analysis results (test_analysis_report.md)
        - Console output with a summary of the findings

    Resolution mode:
        - A report with the analysis results and suggestions (test_id_resolution_report.md)
        - A shell script to rename the tests (rename_tests.sh)
        - Console output with a summary of the findings

Examples:
    Analyzing the Test Suite:
        ./test_suite_tool.py --analyze
        This will analyze the test suite and generate a report in tmp/test_analysis_report.md.

    Resolving Duplicate Test IDs:
        ./test_suite_tool.py --resolve
        This will identify duplicate test IDs, suggest new test IDs, and generate a rename script in tmp/rename_tests.sh.

Note:
    All scripts save their output to the `tmp/` directory by default. This directory is created 
    automatically if it doesn't exist. You can specify a different output directory as a command-line argument.

Author: OneMount Team
"""

import os
import re
import sys
import argparse
from collections import defaultdict

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

# Regular expression to match test function declarations
TEST_FUNC_PATTERN = re.compile(r'func\s+(Test\w+)\s*\(\s*t\s+\*testing\.T\s*\)\s*{')

# Regular expression to extract test ID components
TEST_ID_PATTERN = re.compile(r'Test(UT|IT)_(\w+)_(\d+)_(\d+)_([^(]+)')

class TestCase:
    def __init__(self, func_name, file_path, line_number):
        self.func_name = func_name
        self.file_path = file_path
        self.line_number = line_number

        # Extract test ID components
        match = TEST_ID_PATTERN.match(func_name)
        if match:
            self.test_type = match.group(1)  # UT or IT
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

    def __str__(self):
        return f"{self.test_id}: {self.description} ({self.expected_behavior}) - {self.file_path}:{self.line_number}"

def scan_file(file_path):
    """Scan a file for test functions and return a list of TestCase objects."""
    test_cases = []

    with open(file_path, 'r') as f:
        lines = f.readlines()

    for i, line in enumerate(lines):
        match = TEST_FUNC_PATTERN.search(line)
        if match:
            func_name = match.group(1)
            test_case = TestCase(func_name, file_path, i + 1)
            test_cases.append(test_case)

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

def find_duplicate_test_ids(test_cases):
    """Find duplicate test IDs in the list of test cases."""
    test_id_map = defaultdict(list)

    for test_case in test_cases:
        test_id_map[test_case.test_id].append(test_case)

    # Filter out duplicates that are actually the same test (same function, file, and line)
    filtered_duplicates = {}
    for test_id, cases in test_id_map.items():
        if len(cases) > 1:
            # Create a set of unique tests based on function name, file path, and line number
            unique_cases = set()
            filtered_cases = []

            for case in cases:
                case_identity = (case.func_name, case.file_path, case.line_number)
                if case_identity not in unique_cases:
                    unique_cases.add(case_identity)
                    filtered_cases.append(case)

            # Only include in duplicates if there are still multiple cases after filtering
            if len(filtered_cases) > 1:
                filtered_duplicates[test_id] = filtered_cases

    return filtered_duplicates

# Analysis functions from test_analyzer.py
def find_similar_tests(test_cases):
    """Find tests with similar descriptions that might be testing the same functionality."""
    # Group tests by module and feature number
    module_feature_map = defaultdict(list)

    for test_case in test_cases:
        if test_case.module and test_case.feature_num:
            key = f"{test_case.module}_{test_case.feature_num}"
            module_feature_map[key].append(test_case)

    # Find similar tests within each module/feature group
    similar_tests = []

    for key, cases in module_feature_map.items():
        # Group by description keywords
        description_map = defaultdict(list)

        for case in cases:
            # Extract keywords from description
            keywords = set(word.lower() for word in re.findall(r'\w+', case.description))

            # Add to all matching keyword groups
            for keyword in keywords:
                if len(keyword) > 3:  # Only consider meaningful keywords
                    description_map[keyword].append(case)

        # Find groups with multiple tests
        for keyword, keyword_cases in description_map.items():
            if len(keyword_cases) > 1:
                similar_tests.append((keyword, keyword_cases))

    return similar_tests

def find_ambiguous_boundaries(test_cases):
    """Find tests with ambiguous or unclear boundaries between them."""
    # Group tests by module
    module_map = defaultdict(list)

    for test_case in test_cases:
        if test_case.module:
            module_map[test_case.module].append(test_case)

    # Find sequential feature numbers with potential boundary issues
    ambiguous_boundaries = []

    for module, cases in module_map.items():
        # Sort by feature number and test number
        cases.sort(key=lambda x: (int(x.feature_num), int(x.test_num)))

        # Check for potential boundary issues between adjacent feature numbers
        for i in range(len(cases) - 1):
            current = cases[i]
            next_case = cases[i + 1]

            # If they have the same feature number or adjacent feature numbers
            if current.feature_num == next_case.feature_num or int(current.feature_num) + 1 == int(next_case.feature_num):
                # Check for similar descriptions or expected behaviors
                current_words = set(word.lower() for word in re.findall(r'\w+', current.description + ' ' + current.expected_behavior))
                next_words = set(word.lower() for word in re.findall(r'\w+', next_case.description + ' ' + next_case.expected_behavior))

                # If they share significant words, they might have ambiguous boundaries
                common_words = current_words.intersection(next_words)
                if len(common_words) >= 2:
                    ambiguous_boundaries.append((current, next_case))

    return ambiguous_boundaries

def find_sequential_gaps(test_cases):
    """Find gaps in sequential test case numbering IDs."""
    # Group tests by type, module, and feature number
    group_map = defaultdict(list)

    for test_case in test_cases:
        if test_case.test_type and test_case.module and test_case.feature_num:
            key = f"{test_case.test_type}_{test_case.module}_{test_case.feature_num}"
            group_map[key].append(test_case)

    # Find gaps in test numbers within each group
    gaps = {}

    for key, cases in group_map.items():
        # Extract test numbers and convert to integers
        test_nums = [int(case.test_num) for case in cases]
        test_nums.sort()

        # Find gaps in the sequence
        expected_nums = list(range(1, max(test_nums) + 1))
        missing_nums = [num for num in expected_nums if num not in test_nums]

        if missing_nums:
            # Get the type, module, and feature from the key
            parts = key.split('_')
            test_type = parts[0]
            module = parts[1]
            feature_num = parts[2]

            gaps[key] = {
                'test_type': test_type,
                'module': module,
                'feature_num': feature_num,
                'missing_nums': missing_nums,
                'existing_nums': test_nums
            }

    return gaps

def generate_analysis_report(test_cases, duplicates, similar_tests, ambiguous_boundaries, sequential_gaps):
    """Generate a report with the analysis results."""
    report = []

    # Header
    report.append("# OneMount Test Suite Analysis Report")
    report.append("")

    # Summary
    report.append("## Summary")
    report.append("")
    report.append(f"- Total test cases analyzed: {len(test_cases)}")
    report.append(f"- Duplicate test IDs found: {len(duplicates)}")
    report.append(f"- Groups of similar tests found: {len(similar_tests)}")
    report.append(f"- Tests with ambiguous boundaries: {len(ambiguous_boundaries)}")
    report.append(f"- Groups with sequential test ID gaps: {len(sequential_gaps)}")
    report.append("")

    # Duplicate Test IDs
    report.append("## Duplicate Test IDs")
    report.append("")

    if duplicates:
        for test_id, cases in duplicates.items():
            report.append(f"### Test ID: {test_id}")
            report.append("")
            for case in cases:
                report.append(f"- {case.func_name} - {case.file_path}:{case.line_number}")
            report.append("")
    else:
        report.append("No duplicate test IDs found.")
        report.append("")

    # Similar Tests
    report.append("## Tests with Functional Overlap")
    report.append("")

    if similar_tests:
        for i, (keyword, cases) in enumerate(similar_tests):
            report.append(f"### Group {i+1}: Tests related to '{keyword}'")
            report.append("")
            for case in cases:
                report.append(f"- {case.func_name} - {case.file_path}:{case.line_number}")
                report.append(f"  Description: {case.description} ({case.expected_behavior})")
            report.append("")
    else:
        report.append("No significant functional overlap found between tests.")
        report.append("")

    # Ambiguous Boundaries
    report.append("## Tests with Ambiguous Boundaries")
    report.append("")

    if ambiguous_boundaries:
        for i, (case1, case2) in enumerate(ambiguous_boundaries):
            report.append(f"### Pair {i+1}: Potential boundary issue")
            report.append("")
            report.append(f"- {case1.func_name} - {case1.file_path}:{case1.line_number}")
            report.append(f"  Description: {case1.description} ({case1.expected_behavior})")
            report.append("")
            report.append(f"- {case2.func_name} - {case2.file_path}:{case2.line_number}")
            report.append(f"  Description: {case2.description} ({case2.expected_behavior})")
            report.append("")
    else:
        report.append("No significant boundary issues found between tests.")
        report.append("")

    # Recommendations
    report.append("## Recommendations")
    report.append("")

    # Recommendations for duplicate test IDs
    if duplicates:
        report.append("### Resolving Duplicate Test IDs")
        report.append("")
        report.append("The following test IDs have duplicates and should be renamed:")
        report.append("")
        for test_id, cases in duplicates.items():
            report.append(f"- {test_id}: Rename tests to have unique test numbers")
            for i, case in enumerate(cases):
                suggested_id = f"{case.test_type}_{case.module}_{case.feature_num}_{int(case.test_num) + i:02d}"
                report.append(f"  - {case.func_name} -> Test{suggested_id}_{case.description}_{case.expected_behavior}")
        report.append("")

    # Recommendations for similar tests
    if similar_tests:
        report.append("### Consolidating Similar Tests")
        report.append("")
        report.append("Consider consolidating these similar tests to reduce redundancy:")
        report.append("")
        for i, (keyword, cases) in enumerate(similar_tests):
            report.append(f"- Group {i+1} ('{keyword}'): Consider merging into a table-driven test")
            for case in cases:
                report.append(f"  - {case.func_name}")
        report.append("")

    # Recommendations for ambiguous boundaries
    if ambiguous_boundaries:
        report.append("### Clarifying Test Boundaries")
        report.append("")
        report.append("The following test pairs have ambiguous boundaries and should be clarified:")
        report.append("")
        for i, (case1, case2) in enumerate(ambiguous_boundaries):
            report.append(f"- Pair {i+1}:")
            report.append(f"  - {case1.func_name} and {case2.func_name}")
            report.append(f"  - Recommendation: Clarify the distinction between these tests in their descriptions")
            report.append(f"    or consider merging them if they test the same functionality.")
        report.append("")

    # Sequential Test ID Gaps
    report.append("## Sequential Test ID Gaps")
    report.append("")

    if sequential_gaps:
        report.append("The following groups have gaps in their sequential test case numbering:")
        report.append("")

        for key, gap_info in sequential_gaps.items():
            test_type = gap_info['test_type']
            module = gap_info['module']
            feature_num = gap_info['feature_num']
            missing_nums = gap_info['missing_nums']
            existing_nums = gap_info['existing_nums']

            report.append(f"### {test_type}_{module}_{feature_num}")
            report.append("")
            report.append(f"- Existing test numbers: {', '.join(str(num) for num in sorted(existing_nums))}")
            report.append(f"- Missing test numbers: {', '.join(str(num) for num in sorted(missing_nums))}")
            report.append("")
            report.append("Recommendations:")
            report.append("")

            for missing_num in sorted(missing_nums):
                report.append(f"- Add test case with ID: {test_type}_{module}_{feature_num}_{missing_num:02d}")

            report.append("")
    else:
        report.append("No gaps found in sequential test case numbering.")
        report.append("")

    # General recommendations
    report.append("### General Recommendations")
    report.append("")
    report.append("1. **Standardize Test ID Format**: Ensure all tests follow the `TestXX_YY_ZZ_NN_Description_ExpectedBehavior` format")
    report.append("2. **Use Table-Driven Tests**: Convert similar tests to table-driven tests to reduce code duplication")
    report.append("3. **Improve Test Descriptions**: Make test descriptions more specific to clearly indicate what's being tested")
    report.append("4. **Document Test Boundaries**: Add comments to clarify the boundaries between related tests")
    report.append("5. **Regular Test Review**: Periodically review the test suite for overlaps and redundancies")
    report.append("6. **Fill Sequential Gaps**: Add tests to fill gaps in sequential test case numbering")
    report.append("7. **Verify Test Coverage**: Ensure that all functionality is adequately tested, especially where gaps exist")
    report.append("")

    return "\n".join(report)

# Resolution functions from test_id_resolver.py
def get_next_available_test_number(test_cases, test_type, module, feature_num):
    """Find the next available test number for a given test type, module, and feature number."""
    max_test_num = 0

    for test_case in test_cases:
        if (test_case.test_type == test_type and 
            test_case.module == module and 
            test_case.feature_num == feature_num):
            try:
                test_num = int(test_case.test_num)
                max_test_num = max(max_test_num, test_num)
            except ValueError:
                pass

    return max_test_num + 1

def suggest_new_test_ids(test_cases, duplicates):
    """Suggest new test IDs for tests with duplicate IDs."""
    suggestions = {}

    for test_id, cases in duplicates.items():
        # Sort cases by file path to ensure consistent suggestions
        sorted_cases = sorted(cases, key=lambda x: x.file_path)

        # Keep the first occurrence as is
        suggestions[sorted_cases[0].func_name] = sorted_cases[0].func_name

        # Suggest new IDs for the duplicates
        for i, case in enumerate(sorted_cases[1:], 1):
            # Get the next available test number
            next_test_num = get_next_available_test_number(
                test_cases, case.test_type, case.module, case.feature_num)

            # Create new test ID
            new_test_id = f"{case.test_type}_{case.module}_{case.feature_num}_{next_test_num:02d}"

            # Create new function name
            if case.expected_behavior:
                new_func_name = f"Test{new_test_id}_{case.description}_{case.expected_behavior}"
            else:
                new_func_name = f"Test{new_test_id}_{case.description}"

            suggestions[case.func_name] = new_func_name

            # Update the test case with the new ID to avoid conflicts in future suggestions
            case.test_num = f"{next_test_num:02d}"
            case.test_id = new_test_id

    return suggestions

def generate_rename_script(suggestions):
    """Generate a shell script to rename test functions."""
    script_lines = [
        "#!/bin/bash",
        "",
        "# This script renames test functions to resolve duplicate test IDs",
        "# Generated by test_suite_tool.py",
        "",
        "set -e",
        ""
    ]

    for old_name, new_name in suggestions.items():
        if old_name != new_name:
            script_lines.append(f"# Rename {old_name} to {new_name}")
            script_lines.append(f"find . -name '*_test.go' -exec sed -i 's/{old_name}/{new_name}/g' {{}} \\;")
            script_lines.append("")

    script_lines.append("echo 'Test renaming complete.'")

    return "\n".join(script_lines)

def generate_resolution_report(test_cases, duplicates, suggestions):
    """Generate a report with the analysis results and suggestions."""
    report = []

    # Header
    report.append("# OneMount Test ID Resolution Report")
    report.append("")

    # Summary
    report.append("## Summary")
    report.append("")
    report.append(f"- Total test cases analyzed: {len(test_cases)}")
    report.append(f"- Duplicate test IDs found: {len(duplicates)}")
    report.append(f"- Test functions to rename: {sum(1 for old, new in suggestions.items() if old != new)}")
    report.append("")

    # Duplicate Test IDs
    report.append("## Duplicate Test IDs")
    report.append("")

    if duplicates:
        for test_id, cases in duplicates.items():
            report.append(f"### Test ID: {test_id}")
            report.append("")
            for case in cases:
                report.append(f"- {case.func_name} - {case.file_path}:{case.line_number}")
            report.append("")
    else:
        report.append("No duplicate test IDs found.")
        report.append("")

    # Suggested Renames
    report.append("## Suggested Renames")
    report.append("")

    if suggestions:
        report.append("| Current Function Name | Suggested Function Name |")
        report.append("|------------------------|--------------------------|")

        for old_name, new_name in suggestions.items():
            if old_name != new_name:
                report.append(f"| {old_name} | {new_name} |")

        report.append("")
    else:
        report.append("No renames suggested.")
        report.append("")

    # Implementation Instructions
    report.append("## Implementation Instructions")
    report.append("")
    report.append("To implement these changes, you can:")
    report.append("")
    report.append("1. **Manual Approach**: Rename each function individually using your IDE or text editor")
    report.append("2. **Automated Approach**: Run the generated `rename_tests.sh` script")
    report.append("")
    report.append("### Using the Rename Script")
    report.append("")
    report.append("```bash")
    report.append("chmod +x rename_tests.sh")
    report.append("./rename_tests.sh")
    report.append("```")
    report.append("")
    report.append("**Note**: The script uses `sed` to perform the renames. Make sure to review the changes after running the script.")
    report.append("")
    report.append("### Verifying the Changes")
    report.append("")
    report.append("After renaming the tests, run the test analyzer again to verify that there are no more duplicate test IDs:")
    report.append("")
    report.append("```bash")
    report.append("./test_suite_tool.py --analyze")
    report.append("```")
    report.append("")

    return "\n".join(report)

def analyze_mode(test_cases, output_dir):
    """Run in analysis mode to identify various issues in the test suite."""
    # Find duplicate test IDs
    duplicates = find_duplicate_test_ids(test_cases)

    # Find similar tests
    similar_tests = find_similar_tests(test_cases)

    # Find tests with ambiguous boundaries
    ambiguous_boundaries = find_ambiguous_boundaries(test_cases)

    # Find gaps in sequential test case numbering
    sequential_gaps = find_sequential_gaps(test_cases)

    # Generate report
    report = generate_analysis_report(test_cases, duplicates, similar_tests, ambiguous_boundaries, sequential_gaps)

    # Write report to file
    output_file = os.path.join(output_dir, "test_analysis_report.md")
    with open(output_file, "w") as f:
        f.write(report)

    print(f"Analysis complete. Found {len(test_cases)} test cases.")
    print(f"- Duplicate test IDs: {len(duplicates)}")
    print(f"- Similar test groups: {len(similar_tests)}")
    print(f"- Ambiguous boundaries: {len(ambiguous_boundaries)}")
    print(f"- Groups with sequential test ID gaps: {len(sequential_gaps)}")
    print(f"Report written to {output_file}")

def resolve_mode(test_cases, output_dir):
    """Run in resolution mode to resolve duplicate test IDs."""
    # Find duplicate test IDs
    duplicates = find_duplicate_test_ids(test_cases)

    # Suggest new test IDs
    suggestions = suggest_new_test_ids(test_cases, duplicates)

    # Generate rename script
    rename_script = generate_rename_script(suggestions)

    # Write rename script to file
    script_file = os.path.join(output_dir, "rename_tests.sh")
    with open(script_file, "w") as f:
        f.write(rename_script)

    # Make the script executable
    os.chmod(script_file, 0o755)

    # Generate report
    report = generate_resolution_report(test_cases, duplicates, suggestions)

    # Write report to file
    report_file = os.path.join(output_dir, "test_id_resolution_report.md")
    with open(report_file, "w") as f:
        f.write(report)

    print(f"Analysis complete. Found {len(test_cases)} test cases.")
    print(f"- Duplicate test IDs: {len(duplicates)}")
    print(f"- Test functions to rename: {sum(1 for old, new in suggestions.items() if old != new)}")
    print(f"Report written to {report_file}")
    print(f"Rename script written to {script_file}")

def main():
    # Parse command line arguments
    parser = argparse.ArgumentParser(description="Test Suite Tool for OneMount")
    group = parser.add_mutually_exclusive_group()
    group.add_argument("--analyze", action="store_true", help="Run in analysis mode (default)")
    group.add_argument("--resolve", action="store_true", help="Run in resolution mode")
    parser.add_argument("output_dir", nargs="?", default=OUTPUT_DIR, help="Output directory (default: tmp/)")
    args = parser.parse_args()

    # Create output directory if it doesn't exist
    os.makedirs(args.output_dir, exist_ok=True)

    # Scan all directories
    all_test_cases = []
    for directory in DIRECTORIES:
        all_test_cases.extend(scan_directory(directory))

    # Run in the selected mode
    if args.resolve:
        resolve_mode(all_test_cases, args.output_dir)
    else:
        # Default to analyze mode
        analyze_mode(all_test_cases, args.output_dir)

if __name__ == "__main__":
    main()
