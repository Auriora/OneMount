#!/usr/bin/env python3
"""
Requirements Registry for OneMount

This script manages a centralized registry of requirements in the OneMount project. It scans
the requirements documentation and creates a JSON registry of all requirements, similar to
the test_id_registry.json file.

Functionality:
    - Scans requirements documentation for requirement IDs and details
    - Creates a JSON registry of all requirements
    - Provides commands to list, search, and filter requirements

Usage:
    ./requirements_registry.py [command] [args...]
    ./requirements_registry.py update [output_dir]
    ./requirements_registry.py list [--type TYPE] [--category CATEGORY]
    ./requirements_registry.py search <search_term>
    ./requirements_registry.py filter <field> <value>

Arguments:
    command - Command to execute (update, list, search, filter)
    output_dir - Optional directory to save the registry file (default: data/)
    type - Requirement type (FR, NFR)
    category - Requirement category (FS, AUTH, PERF, etc.)
    search_term - Term to search for in requirement descriptions
    field - Field to filter on (id, type, category, priority, etc.)
    value - Value to filter for

Output:
    - A JSON file with the registry data (requirements_registry.json)
    - Console output with the results of the command

Installation:
    The Requirements Registry tool is included in the OneMount repository. No additional installation is required.

Requirement ID Structure:
    The requirement ID structure follows this pattern:
    <TYPE>-<CATEGORY>-<NUMBER>

    Where:
    - <TYPE> is the requirement type:
      - FR - Functional Requirement
      - NFR - Non-Functional Requirement
    - <CATEGORY> is the category of the requirement:
      - FS - Filesystem Operations
      - AUTH - Authentication
      - OFF - Offline Functionality
      - UI - User Interface
      - STAT - Statistics and Analysis
      - INT - Integration with External Systems
      - DEV - Developer Tools
      - PERF - Performance
      - SEC - Security
      - USE - Usability
      - REL - Reliability
      - MNT - Maintainability
    - <NUMBER> is a 3-digit number uniquely identifying the requirement

Examples:
    Updating the Registry:
        ./requirements_registry.py update
        This will scan the requirements documentation and update the registry.

    Listing All Requirements:
        ./requirements_registry.py list
        This will list all requirements in the registry.

    Listing Functional Requirements:
        ./requirements_registry.py list --type FR
        This will list all functional requirements in the registry.

    Listing Filesystem Requirements:
        ./requirements_registry.py list --category FS
        This will list all filesystem requirements in the registry.

    Searching for Requirements:
        ./requirements_registry.py search "authentication"
        This will search for requirements containing "authentication" in their description.

    Filtering Requirements:
        ./requirements_registry.py filter priority "Must-have"
        This will filter requirements with priority "Must-have".

Registry File:
    The registry file is stored in `data/requirements_registry.json`. It contains a JSON object with the following structure:

    {
      "requirements": [
        {
          "req_id": "FR-FS-001",
          "req_type": "FR",
          "category": "FS",
          "number": "001",
          "description": "The system shall mount OneDrive as a native Linux filesystem using FUSE.",
          "priority": "Must-have",
          "rationale": "Essential for providing filesystem access to OneDrive content.",
          "file_path": "docs/1-requirements/srs/3-specific-requirements.md",
          "line_number": 11
        }
      ],
      "last_updated": "2023-05-07T12:34:56.789012"
    }

Requirements Format:
    Requirements in the OneMount project follow a specific format:

    - ID: A unique identifier for the requirement (e.g., FR-FS-001, NFR-PERF-001)
    - Type: The type of requirement (FR for Functional Requirement, NFR for Non-Functional Requirement)
    - Category: The category of the requirement (FS, AUTH, PERF, etc.)
    - Description: A description of the requirement
    - Priority: The priority of the requirement (Must-have, Should-have, Could-have)
    - Rationale: The reason for the requirement

Best Practices:
    1. Always update the registry before using it: Run `./requirements_registry.py update` to ensure 
       the registry is up to date with the latest requirements documentation.

    2. Use the list, search, and filter commands to find specific requirements: These commands 
       provide a convenient way to find requirements based on various criteria.

    3. Use the registry for requirements traceability: The registry can be used to trace requirements 
       to test cases, design elements, and implementation components.

Troubleshooting:
    If you encounter issues with the Requirements Registry tool:

    1. Make sure the requirements documentation follows the expected format.
    2. Check that the `docs/1-requirements/srs` directory exists and contains the requirements documentation.
    3. Ensure you have the necessary permissions to read the documentation files and write to the `data` directory.

See Also:
    - Software Requirements Specification (docs/1-requirements/srs/3-specific-requirements.md)
    - Requirements Traceability Matrix (docs/2-architecture-and-design/sas-requirements-traceability-matrix.md)
    - Test Cases Traceability Matrix (docs/4-testing/test-cases-traceability-matrix.md)

Note:
    The registry file is saved to the `data/` directory by default. This directory is created 
    automatically if it doesn't exist. You can specify a different output directory as a command-line argument.

Author: OneMount Team
"""

import os
import re
import sys
import json
import argparse
from datetime import datetime

# Define the directories to scan
DIRECTORIES = [
    "docs/1-requirements/srs"
]

# Define the output directory
OUTPUT_DIR = "data"

# Registry file name
REGISTRY_FILE_NAME = "requirements_registry.json"

# Regular expression to match requirement IDs in markdown files
REQUIREMENT_PATTERN = re.compile(r'<a id="([^"]+)">(?:\*\*)(FR|NFR)-([A-Z]+)-(\d+)(?:\*\*)</a>\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)')

class Requirement:
    def __init__(self, req_id, req_type, category, number, description, priority, rationale, file_path, line_number):
        self.req_id = req_id
        self.req_type = req_type
        self.category = category
        self.number = number
        self.description = description.strip()
        self.priority = priority.strip()
        self.rationale = rationale.strip()
        self.file_path = file_path
        self.line_number = line_number

    def to_dict(self):
        """Convert Requirement to dictionary for JSON serialization."""
        return {
            "req_id": self.req_id,
            "req_type": self.req_type,
            "category": self.category,
            "number": self.number,
            "description": self.description,
            "priority": self.priority,
            "rationale": self.rationale,
            "file_path": self.file_path,
            "line_number": self.line_number
        }

    @classmethod
    def from_dict(cls, data):
        """Create Requirement from dictionary."""
        return cls(
            data["req_id"],
            data["req_type"],
            data["category"],
            data["number"],
            data["description"],
            data["priority"],
            data["rationale"],
            data["file_path"],
            data["line_number"]
        )

    def __str__(self):
        return f"{self.req_id}: {self.description} ({self.priority}) - {self.file_path}:{self.line_number}"

def scan_file(file_path):
    """Scan a file for requirements and return a list of Requirement objects."""
    requirements = []

    try:
        with open(file_path, 'r') as f:
            lines = f.readlines()

        for i, line in enumerate(lines):
            match = REQUIREMENT_PATTERN.search(line)
            if match:
                anchor_id = match.group(1)
                req_type = match.group(2)
                category = match.group(3)
                number = match.group(4)
                description = match.group(5)
                priority = match.group(6)
                rationale = match.group(7)

                req_id = f"{req_type}-{category}-{number}"
                requirement = Requirement(req_id, req_type, category, number, description, priority, rationale, file_path, i + 1)
                requirements.append(requirement)
    except Exception as e:
        print(f"Error scanning file {file_path}: {e}")

    return requirements

def scan_directory(directory):
    """Recursively scan a directory for markdown files and return a list of Requirement objects."""
    requirements = []

    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith('.md'):
                file_path = os.path.join(root, file)
                requirements.extend(scan_file(file_path))

    return requirements

def scan_all_directories():
    """Scan all directories for requirements."""
    all_requirements = []
    for directory in DIRECTORIES:
        all_requirements.extend(scan_directory(directory))
    return all_requirements

def load_registry(output_dir=None):
    """Load the requirements registry from file."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    registry_file = os.path.join(output_dir, REGISTRY_FILE_NAME)

    if not os.path.exists(registry_file):
        return {"requirements": [], "last_updated": ""}

    try:
        with open(registry_file, 'r') as f:
            registry = json.load(f)

        # Convert dictionaries back to Requirement objects
        registry["requirements"] = [Requirement.from_dict(req) for req in registry["requirements"]]
        return registry
    except Exception as e:
        print(f"Error loading registry: {e}")
        return {"requirements": [], "last_updated": ""}

def save_registry(registry, output_dir=None):
    """Save the requirements registry to file."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    # Convert Requirement objects to dictionaries
    registry_copy = {
        "requirements": [req.to_dict() for req in registry["requirements"]],
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
    """Update the requirements registry with the latest requirements."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Scan all directories for requirements
    all_requirements = scan_all_directories()

    # Create a new registry
    registry = {
        "requirements": all_requirements,
        "last_updated": datetime.now().isoformat()
    }

    # Save the registry
    save_registry(registry, output_dir)

    print(f"Registry updated with {len(all_requirements)} requirements.")
    return registry

def list_requirements(req_type=None, category=None, output_dir=None):
    """List all requirements in the registry, optionally filtered by type and category."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_registry(output_dir)

    # Filter requirements
    filtered_reqs = registry["requirements"]
    if req_type:
        filtered_reqs = [req for req in filtered_reqs if req.req_type == req_type]
    if category:
        filtered_reqs = [req for req in filtered_reqs if req.category == category]

    # Sort requirements by ID
    filtered_reqs.sort(key=lambda req: req.req_id)

    # Print requirements
    print(f"Found {len(filtered_reqs)} requirements:")
    for req in filtered_reqs:
        print(f"  {req.req_id}: {req.description} ({req.priority}) - {req.file_path}:{req.line_number}")

    return filtered_reqs

def search_requirements(search_term, output_dir=None):
    """Search for requirements containing the search term in their description."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_registry(output_dir)

    # Filter requirements
    filtered_reqs = [req for req in registry["requirements"] 
                    if search_term.lower() in req.description.lower()]

    # Sort requirements by ID
    filtered_reqs.sort(key=lambda req: req.req_id)

    # Print requirements
    print(f"Found {len(filtered_reqs)} requirements matching '{search_term}':")
    for req in filtered_reqs:
        print(f"  {req.req_id}: {req.description} ({req.priority}) - {req.file_path}:{req.line_number}")

    return filtered_reqs

def filter_requirements(field, value, output_dir=None):
    """Filter requirements by a specific field and value."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_registry(output_dir)

    # Filter requirements
    filtered_reqs = []
    for req in registry["requirements"]:
        if hasattr(req, field) and getattr(req, field).lower() == value.lower():
            filtered_reqs.append(req)

    # Sort requirements by ID
    filtered_reqs.sort(key=lambda req: req.req_id)

    # Print requirements
    print(f"Found {len(filtered_reqs)} requirements with {field}='{value}':")
    for req in filtered_reqs:
        print(f"  {req.req_id}: {req.description} ({req.priority}) - {req.file_path}:{req.line_number}")

    return filtered_reqs

def main():
    """Main function."""
    parser = argparse.ArgumentParser(description="Requirements Registry")
    subparsers = parser.add_subparsers(dest="command", help="Command to execute")

    # Update registry command
    update_parser = subparsers.add_parser("update", help="Update the requirements registry")
    update_parser.add_argument("output_dir", nargs="?", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # List requirements command
    list_parser = subparsers.add_parser("list", help="List all requirements")
    list_parser.add_argument("--type", help="Filter by requirement type (FR, NFR)")
    list_parser.add_argument("--category", help="Filter by category (FS, AUTH, PERF, etc.)")
    list_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # Search requirements command
    search_parser = subparsers.add_parser("search", help="Search for requirements")
    search_parser.add_argument("search_term", help="Term to search for in requirement descriptions")
    search_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # Filter requirements command
    filter_parser = subparsers.add_parser("filter", help="Filter requirements by field")
    filter_parser.add_argument("field", help="Field to filter on (req_type, category, priority, etc.)")
    filter_parser.add_argument("value", help="Value to filter for")
    filter_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    args = parser.parse_args()

    # Create output directory if it doesn't exist
    output_dir = getattr(args, "output_dir", OUTPUT_DIR)
    os.makedirs(output_dir, exist_ok=True)

    if args.command == "update":
        update_registry(args.output_dir)
    elif args.command == "list":
        list_requirements(args.type, args.category, args.output_dir)
    elif args.command == "search":
        search_requirements(args.search_term, args.output_dir)
    elif args.command == "filter":
        filter_requirements(args.field, args.value, args.output_dir)
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
