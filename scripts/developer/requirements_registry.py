#!/usr/bin/env python3
"""
Requirements and Use Cases Registry for OneMount

This script manages a centralized registry of requirements and use cases in the OneMount project. It scans
the requirements and use cases documentation and creates JSON registries, similar to
the test_id_registry.json file.

Functionality:
    - Scans requirements documentation for requirement IDs and details
    - Scans use cases documentation for use case IDs and details
    - Creates JSON registries of all requirements and use cases
    - Provides commands to list, search, and filter requirements and use cases

Usage:
    ./requirements_registry.py [command] [args...]
    ./requirements_registry.py update [output_dir]
    ./requirements_registry.py list [--type TYPE] [--category CATEGORY] [--registry REGISTRY]
    ./requirements_registry.py search <search_term> [--registry REGISTRY]
    ./requirements_registry.py filter <field> <value> [--registry REGISTRY]

Arguments:
    command - Command to execute (update, list, search, filter)
    output_dir - Optional directory to save the registry files (default: data/)
    type - Requirement type (FR, NFR) or use case type (UC)
    category - Requirement or use case category (FS, AUTH, PERF, etc.)
    registry - Registry to operate on (requirements, usecases, or both)
    search_term - Term to search for in requirement or use case descriptions
    field - Field to filter on (id, type, category, priority, name, etc.)
    value - Value to filter for

Output:
    - JSON files with the registry data (requirements_registry.json, usecases_registry.json)
    - Console output with the results of the command

Installation:
    The Requirements and Use Cases Registry tool is included in the OneMount repository. No additional installation is required.

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

Use Case ID Structure:
    The use case ID structure follows this pattern:
    <TYPE>-<CATEGORY>-<NUMBER>

    Where:
    - <TYPE> is always UC (Use Case)
    - <CATEGORY> is the category of the use case:
      - FS - Filesystem Operations
      - AUTH - Authentication
      - OFF - Offline Functionality
      - UI - User Interface
      - STAT - Statistics and Analysis
      - INT - Integration with External Systems
    - <NUMBER> is a 3-digit number uniquely identifying the use case

Examples:
    Updating the Registries:
        ./requirements_registry.py update
        This will scan the documentation and update both requirements and use cases registries.

    Listing All Requirements:
        ./requirements_registry.py list --registry requirements
        This will list all requirements in the registry.

    Listing All Use Cases:
        ./requirements_registry.py list --registry usecases
        This will list all use cases in the registry.

    Listing All Items (Both Requirements and Use Cases):
        ./requirements_registry.py list
        This will list all requirements and use cases in both registries.

    Listing Functional Requirements:
        ./requirements_registry.py list --type FR --registry requirements
        This will list all functional requirements in the registry.

    Listing Filesystem Use Cases:
        ./requirements_registry.py list --category FS --registry usecases
        This will list all filesystem use cases in the registry.

    Searching for Requirements:
        ./requirements_registry.py search "authentication" --registry requirements
        This will search for requirements containing "authentication" in their description.

    Searching for Use Cases:
        ./requirements_registry.py search "offline" --registry usecases
        This will search for use cases containing "offline" in their description.

    Filtering Requirements:
        ./requirements_registry.py filter priority "Must-have" --registry requirements
        This will filter requirements with priority "Must-have".

    Filtering Use Cases:
        ./requirements_registry.py filter category "FS" --registry usecases
        This will filter use cases with category "FS".

Registry Files:
    The requirements registry file is stored in `data/requirements_registry.json`. It contains a JSON object with the following structure:

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

    The use cases registry file is stored in `data/usecases_registry.json`. It contains a JSON object with the following structure:

    {
      "usecases": [
        {
          "uc_id": "UC-FS-001",
          "uc_type": "UC",
          "category": "FS",
          "number": "001",
          "name": "Mount OneDrive Filesystem",
          "description": "This use case describes how a user mounts OneDrive as a filesystem.",
          "primary_actors": "User",
          "file_path": "docs/1-requirements/srs/4-use-cases.md",
          "line_number": 9
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

Use Cases Format:
    Use cases in the OneMount project follow a specific format:

    - ID: A unique identifier for the use case (e.g., UC-FS-001)
    - Name: The name of the use case
    - Primary Actors: The primary actors involved in the use case
    - Stakeholders & Interests: The stakeholders and their interests
    - Pre-conditions: The conditions that must be true before the use case can start
    - Post-conditions: The conditions that must be true after the use case completes
    - Main Flow: The main sequence of steps in the use case
    - Alternative Flows: Alternative sequences of steps
    - Special Requirements: Special requirements for the use case
    - Related FRs: Related functional requirements
    - Related NFRs: Related non-functional requirements

Best Practices:
    1. Always update the registries before using them: Run `./requirements_registry.py update` to ensure 
       the registries are up to date with the latest documentation.

    2. Use the list, search, and filter commands to find specific requirements or use cases: These commands 
       provide a convenient way to find items based on various criteria.

    3. Use the registries for traceability: The registries can be used to trace requirements and use cases 
       to test cases, design elements, and implementation components.

    4. Specify the registry when using commands: Use the `--registry` parameter to specify whether you want 
       to work with requirements, use cases, or both.

Troubleshooting:
    If you encounter issues with the Requirements and Use Cases Registry tool:

    1. Make sure the documentation follows the expected format.
    2. Check that the `docs/1-requirements/srs` directory exists and contains the requirements and use cases documentation.
    3. Ensure you have the necessary permissions to read the documentation files and write to the `data` directory.
    4. If use cases are not being found, check that the use case documentation follows the expected format with the correct table structure.

See Also:
    - Software Requirements Specification (docs/1-requirements/srs/3-specific-requirements.md)
    - Use Cases Documentation (docs/1-requirements/srs/4-use-cases.md)
    - Requirements Traceability Matrix (docs/2-architecture-and-design/sas-requirements-traceability-matrix.md)
    - Test Cases Traceability Matrix (docs/4-testing/test-cases-traceability-matrix.md)

Note:
    The registry files are saved to the `data/` directory by default. This directory is created 
    automatically if it doesn't exist. You can specify a different output directory as a command-line argument.
    The requirements registry is saved as `requirements_registry.json` and the use cases registry is saved as `usecases_registry.json`.

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

# Registry file names
REQUIREMENTS_REGISTRY_FILE_NAME = "requirements_registry.json"
USECASES_REGISTRY_FILE_NAME = "usecases_registry.json"

# Regular expression to match requirement IDs in markdown files
REQUIREMENT_PATTERN = re.compile(r'<a id="([^"]+)">(?:\*\*)(FR|NFR)-([A-Z]+)-(\d+)(?:\*\*)</a>\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)')

# Regular expression to match use case IDs in markdown files
USECASE_PATTERN = re.compile(r'\|\s*\*\*Use Case ID\*\*\s*\|\s*(UC-[A-Z]+-\d+)\s*\|')

# Regular expression to match use case names in markdown files
USECASE_NAME_PATTERN = re.compile(r'\|\s*\*\*Name\*\*\s*\|\s*([^|]+)\s*\|')

# Regular expression to match use case primary actors in markdown files
USECASE_ACTORS_PATTERN = re.compile(r'\|\s*\*\*Primary Actors\*\*\s*\|\s*([^|]+)\s*\|')

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


class UseCase:
    def __init__(self, uc_id, name, primary_actors, file_path, line_number):
        self.uc_id = uc_id
        # Extract type, category, and number from uc_id (e.g., UC-FS-001)
        parts = uc_id.split('-')
        self.uc_type = parts[0]  # UC
        self.category = parts[1]  # FS, AUTH, etc.
        self.number = parts[2]    # 001, 002, etc.
        self.name = name.strip()
        self.primary_actors = primary_actors.strip()
        self.file_path = file_path
        self.line_number = line_number

    def to_dict(self):
        """Convert UseCase to dictionary for JSON serialization."""
        return {
            "uc_id": self.uc_id,
            "uc_type": self.uc_type,
            "category": self.category,
            "number": self.number,
            "name": self.name,
            "primary_actors": self.primary_actors,
            "file_path": self.file_path,
            "line_number": self.line_number
        }

    @classmethod
    def from_dict(cls, data):
        """Create UseCase from dictionary."""
        return cls(
            data["uc_id"],
            data["name"],
            data["primary_actors"],
            data["file_path"],
            data["line_number"]
        )

    def __str__(self):
        return f"{self.uc_id}: {self.name} (Actors: {self.primary_actors}) - {self.file_path}:{self.line_number}"

def scan_file_for_requirements(file_path):
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
        print(f"Error scanning file {file_path} for requirements: {e}")

    return requirements

def scan_file_for_usecases(file_path):
    """Scan a file for use cases and return a list of UseCase objects."""
    usecases = []

    try:
        with open(file_path, 'r') as f:
            lines = f.readlines()

        # We need to find use case IDs, names, and primary actors
        # These are in different lines in the file, so we need to track the current use case
        current_uc_id = None
        current_uc_id_line = 0
        current_uc_name = None
        current_uc_actors = None

        for i, line in enumerate(lines):
            # Look for use case ID
            id_match = USECASE_PATTERN.search(line)
            if id_match:
                # If we found a new use case ID, save the previous one if it exists
                if current_uc_id and current_uc_name and current_uc_actors:
                    usecase = UseCase(current_uc_id, current_uc_name, current_uc_actors, file_path, current_uc_id_line)
                    usecases.append(usecase)

                # Start tracking a new use case
                current_uc_id = id_match.group(1)
                current_uc_id_line = i + 1
                current_uc_name = None
                current_uc_actors = None
                continue

            # If we're tracking a use case, look for its name
            if current_uc_id and not current_uc_name:
                name_match = USECASE_NAME_PATTERN.search(line)
                if name_match:
                    current_uc_name = name_match.group(1)
                    continue

            # If we have a use case ID and name, look for primary actors
            if current_uc_id and current_uc_name and not current_uc_actors:
                actors_match = USECASE_ACTORS_PATTERN.search(line)
                if actors_match:
                    current_uc_actors = actors_match.group(1)
                    continue

        # Don't forget to save the last use case if it exists
        if current_uc_id and current_uc_name and current_uc_actors:
            usecase = UseCase(current_uc_id, current_uc_name, current_uc_actors, file_path, current_uc_id_line)
            usecases.append(usecase)

    except Exception as e:
        print(f"Error scanning file {file_path} for use cases: {e}")

    return usecases

def scan_file(file_path):
    """Scan a file for requirements and use cases and return a tuple of lists."""
    requirements = scan_file_for_requirements(file_path)
    usecases = scan_file_for_usecases(file_path)
    return requirements, usecases

def scan_directory(directory):
    """Recursively scan a directory for markdown files and return a tuple of lists (requirements, usecases)."""
    all_requirements = []
    all_usecases = []

    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith('.md'):
                file_path = os.path.join(root, file)
                requirements, usecases = scan_file(file_path)
                all_requirements.extend(requirements)
                all_usecases.extend(usecases)

    return all_requirements, all_usecases

def scan_all_directories():
    """Scan all directories for requirements and use cases."""
    all_requirements = []
    all_usecases = []
    for directory in DIRECTORIES:
        requirements, usecases = scan_directory(directory)
        all_requirements.extend(requirements)
        all_usecases.extend(usecases)
    return all_requirements, all_usecases

def load_requirements_registry(output_dir=None):
    """Load the requirements registry from file."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    registry_file = os.path.join(output_dir, REQUIREMENTS_REGISTRY_FILE_NAME)

    if not os.path.exists(registry_file):
        return {"requirements": [], "last_updated": ""}

    try:
        with open(registry_file, 'r') as f:
            registry = json.load(f)

        # Convert dictionaries back to Requirement objects
        registry["requirements"] = [Requirement.from_dict(req) for req in registry["requirements"]]
        return registry
    except Exception as e:
        print(f"Error loading requirements registry: {e}")
        return {"requirements": [], "last_updated": ""}

def load_usecases_registry(output_dir=None):
    """Load the use cases registry from file."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    registry_file = os.path.join(output_dir, USECASES_REGISTRY_FILE_NAME)

    if not os.path.exists(registry_file):
        return {"usecases": [], "last_updated": ""}

    try:
        with open(registry_file, 'r') as f:
            registry = json.load(f)

        # Convert dictionaries back to UseCase objects
        registry["usecases"] = [UseCase.from_dict(uc) for uc in registry["usecases"]]
        return registry
    except Exception as e:
        print(f"Error loading use cases registry: {e}")
        return {"usecases": [], "last_updated": ""}

def save_requirements_registry(registry, output_dir=None):
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

    registry_file = os.path.join(output_dir, REQUIREMENTS_REGISTRY_FILE_NAME)

    try:
        with open(registry_file, 'w') as f:
            json.dump(registry_copy, f, indent=2)
        print(f"Requirements registry saved to {registry_file}")
    except Exception as e:
        print(f"Error saving requirements registry: {e}")

def save_usecases_registry(registry, output_dir=None):
    """Save the use cases registry to file."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    # Convert UseCase objects to dictionaries
    registry_copy = {
        "usecases": [uc.to_dict() for uc in registry["usecases"]],
        "last_updated": datetime.now().isoformat()
    }

    registry_file = os.path.join(output_dir, USECASES_REGISTRY_FILE_NAME)

    try:
        with open(registry_file, 'w') as f:
            json.dump(registry_copy, f, indent=2)
        print(f"Use cases registry saved to {registry_file}")
    except Exception as e:
        print(f"Error saving use cases registry: {e}")

# For backward compatibility
def load_registry(output_dir=None):
    """Load the requirements registry from file (for backward compatibility)."""
    return load_requirements_registry(output_dir)

def save_registry(registry, output_dir=None):
    """Save the requirements registry to file (for backward compatibility)."""
    save_requirements_registry(registry, output_dir)

def update_registry(output_dir=None):
    """Update the requirements and use cases registries with the latest data."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    # Scan all directories for requirements and use cases
    all_requirements, all_usecases = scan_all_directories()

    # Create a new requirements registry
    requirements_registry = {
        "requirements": all_requirements,
        "last_updated": datetime.now().isoformat()
    }

    # Create a new use cases registry
    usecases_registry = {
        "usecases": all_usecases,
        "last_updated": datetime.now().isoformat()
    }

    # Save the registries
    save_requirements_registry(requirements_registry, output_dir)
    save_usecases_registry(usecases_registry, output_dir)

    print(f"Requirements registry updated with {len(all_requirements)} requirements.")
    print(f"Use cases registry updated with {len(all_usecases)} use cases.")

    return requirements_registry, usecases_registry

def list_requirements(req_type=None, category=None, output_dir=None):
    """List all requirements in the registry, optionally filtered by type and category."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_requirements_registry(output_dir)

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

def list_usecases(category=None, output_dir=None):
    """List all use cases in the registry, optionally filtered by category."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_usecases_registry(output_dir)

    # Filter use cases
    filtered_ucs = registry["usecases"]
    if category:
        filtered_ucs = [uc for uc in filtered_ucs if uc.category == category]

    # Sort use cases by ID
    filtered_ucs.sort(key=lambda uc: uc.uc_id)

    # Print use cases
    print(f"Found {len(filtered_ucs)} use cases:")
    for uc in filtered_ucs:
        print(f"  {uc.uc_id}: {uc.name} (Actors: {uc.primary_actors}) - {uc.file_path}:{uc.line_number}")

    return filtered_ucs

def search_requirements(search_term, output_dir=None):
    """Search for requirements containing the search term in their description."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_requirements_registry(output_dir)

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

def search_usecases(search_term, output_dir=None):
    """Search for use cases containing the search term in their name."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_usecases_registry(output_dir)

    # Filter use cases
    filtered_ucs = [uc for uc in registry["usecases"] 
                   if search_term.lower() in uc.name.lower()]

    # Sort use cases by ID
    filtered_ucs.sort(key=lambda uc: uc.uc_id)

    # Print use cases
    print(f"Found {len(filtered_ucs)} use cases matching '{search_term}':")
    for uc in filtered_ucs:
        print(f"  {uc.uc_id}: {uc.name} (Actors: {uc.primary_actors}) - {uc.file_path}:{uc.line_number}")

    return filtered_ucs

def filter_requirements(field, value, output_dir=None):
    """Filter requirements by a specific field and value."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_requirements_registry(output_dir)

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

def filter_usecases(field, value, output_dir=None):
    """Filter use cases by a specific field and value."""
    if output_dir is None:
        output_dir = OUTPUT_DIR

    registry = load_usecases_registry(output_dir)

    # Filter use cases
    filtered_ucs = []
    for uc in registry["usecases"]:
        if hasattr(uc, field) and getattr(uc, field).lower() == value.lower():
            filtered_ucs.append(uc)

    # Sort use cases by ID
    filtered_ucs.sort(key=lambda uc: uc.uc_id)

    # Print use cases
    print(f"Found {len(filtered_ucs)} use cases with {field}='{value}':")
    for uc in filtered_ucs:
        print(f"  {uc.uc_id}: {uc.name} (Actors: {uc.primary_actors}) - {uc.file_path}:{uc.line_number}")

    return filtered_ucs

def main():
    """Main function."""
    parser = argparse.ArgumentParser(description="Requirements and Use Cases Registry")
    subparsers = parser.add_subparsers(dest="command", help="Command to execute")

    # Update registry command
    update_parser = subparsers.add_parser("update", help="Update the requirements and use cases registries")
    update_parser.add_argument("output_dir", nargs="?", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # List command
    list_parser = subparsers.add_parser("list", help="List all requirements or use cases")
    list_parser.add_argument("--type", help="Filter by requirement type (FR, NFR)")
    list_parser.add_argument("--category", help="Filter by category (FS, AUTH, PERF, etc.)")
    list_parser.add_argument("--registry", choices=["requirements", "usecases", "both"], default="both", help="Registry to operate on (default: both)")
    list_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # Search command
    search_parser = subparsers.add_parser("search", help="Search for requirements or use cases")
    search_parser.add_argument("search_term", help="Term to search for in descriptions or names")
    search_parser.add_argument("--registry", choices=["requirements", "usecases", "both"], default="both", help="Registry to operate on (default: both)")
    search_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    # Filter command
    filter_parser = subparsers.add_parser("filter", help="Filter requirements or use cases by field")
    filter_parser.add_argument("field", help="Field to filter on (req_type, category, priority, name, etc.)")
    filter_parser.add_argument("value", help="Value to filter for")
    filter_parser.add_argument("--registry", choices=["requirements", "usecases", "both"], default="both", help="Registry to operate on (default: both)")
    filter_parser.add_argument("--output-dir", default=OUTPUT_DIR, help=f"Output directory (default: {OUTPUT_DIR})")

    args = parser.parse_args()

    # Create output directory if it doesn't exist
    output_dir = getattr(args, "output_dir", OUTPUT_DIR)
    os.makedirs(output_dir, exist_ok=True)

    if args.command == "update":
        update_registry(args.output_dir)
    elif args.command == "list":
        if args.registry in ["requirements", "both"]:
            list_requirements(args.type, args.category, args.output_dir)
        if args.registry in ["usecases", "both"]:
            list_usecases(args.category, args.output_dir)
    elif args.command == "search":
        if args.registry in ["requirements", "both"]:
            search_requirements(args.search_term, args.output_dir)
        if args.registry in ["usecases", "both"]:
            search_usecases(args.search_term, args.output_dir)
    elif args.command == "filter":
        if args.registry in ["requirements", "both"]:
            filter_requirements(args.field, args.value, args.output_dir)
        if args.registry in ["usecases", "both"]:
            filter_usecases(args.field, args.value, args.output_dir)
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
