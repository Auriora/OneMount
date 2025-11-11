#!/bin/bash
# Python helper script for OneMount Docker containers
# Provides guidance on how to use Python and pip in the container

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
OneMount Python Helper

This script provides guidance on how to use Python and pip in the OneMount Docker containers.

Common Python Commands:

  # Install packages from requirements file
  python3 -m pip install -r scripts/requirements-dev-cli.txt

  # Create a virtual environment
  python3 -m venv venv

  # Activate the virtual environment
  source venv/bin/activate

  # Install packages in the virtual environment
  pip install -r scripts/requirements-dev-cli.txt

  # Deactivate the virtual environment
  deactivate

Examples:

  # Install development CLI requirements
  python3 -m pip install -r scripts/requirements-dev-cli.txt

  # Create and use a virtual environment
  python3 -m venv venv
  source venv/bin/activate
  pip install -r scripts/requirements-dev-cli.txt
  # ... do your work ...
  deactivate

Notes:
  - The container includes Python 3.12 and the necessary packages for virtual environments
  - Use 'python3 -m pip' instead of just 'pip' when not in a virtual environment
  - Always use '-r' with pip install, not with python3 directly
EOF
}

# Main execution
show_help