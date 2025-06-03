#!/bin/bash
# OneMount Release Helper Script
# Automates version bumping and triggering package builds

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
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

# Check if we're in the right directory
if [ ! -f ".bumpversion.cfg" ] || [ ! -f "cmd/common/common.go" ]; then
    print_error "This script must be run from the OneMount project root directory"
    exit 1
fi

# Check if bumpversion is available
if [ ! -f ".venv/bin/bumpversion" ]; then
    print_error "bumpversion not found in .venv/bin/"
    print_status "Please install it with: .venv/bin/pip install bump2version"
    exit 1
fi

# Function to show current version
show_current_version() {
    local current_version=$(grep "current_version" .bumpversion.cfg | sed 's/current_version = //')
    print_status "Current version: $current_version"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 <bump_type> [--dry-run] [--no-push]"
    echo ""
    echo "Bump types:"
    echo "  num       - Bump release candidate number (0.1.0rc1 → 0.1.0rc2)"
    echo "  release   - Release current RC (0.1.0rc1 → 0.1.0)"
    echo "  patch     - Bump patch version (0.1.0 → 0.1.1)"
    echo "  minor     - Bump minor version (0.1.0 → 0.2.0)"
    echo "  major     - Bump major version (0.1.0 → 1.0.0)"
    echo ""
    echo "Options:"
    echo "  --dry-run  - Show what would be done without making changes"
    echo "  --no-push  - Don't push tags to GitHub (skip package building)"
    echo ""
    echo "Examples:"
    echo "  $0 num                    # Bump RC number and trigger package build"
    echo "  $0 release --dry-run      # Preview release without changes"
    echo "  $0 patch --no-push       # Bump patch but don't trigger build"
}

# Parse arguments
BUMP_TYPE=""
DRY_RUN=false
NO_PUSH=false

while [[ $# -gt 0 ]]; do
    case $1 in
        num|release|patch|minor|major)
            BUMP_TYPE="$1"
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --no-push)
            NO_PUSH=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if bump type is provided
if [ -z "$BUMP_TYPE" ]; then
    print_error "Bump type is required"
    show_usage
    exit 1
fi

# Show current version
show_current_version

# Prepare bumpversion command
BUMPVERSION_CMD=".venv/bin/bumpversion"
if [ "$DRY_RUN" = true ]; then
    BUMPVERSION_CMD="$BUMPVERSION_CMD --dry-run --verbose"
    print_warning "DRY RUN MODE - No changes will be made"
fi

# Check if working directory is dirty and add --allow-dirty if needed
if ! git diff-index --quiet HEAD --; then
    print_warning "Working directory is dirty, adding --allow-dirty flag"
    BUMPVERSION_CMD="$BUMPVERSION_CMD --allow-dirty"
fi

# Run bumpversion
print_status "Running: $BUMPVERSION_CMD $BUMP_TYPE"
$BUMPVERSION_CMD $BUMP_TYPE

if [ "$DRY_RUN" = true ]; then
    print_success "Dry run completed successfully"
    exit 0
fi

# Show new version
print_success "Version bumped successfully!"
show_current_version

# Check if we should push tags
if [ "$NO_PUSH" = true ]; then
    print_warning "Skipping tag push (--no-push specified)"
    print_status "To trigger package building later, run: git push origin --tags"
else
    # Push tags to trigger GitHub Actions
    print_status "Pushing tags to GitHub to trigger package building..."
    
    if git push origin --tags; then
        print_success "Tags pushed successfully!"
        print_status "GitHub Actions will now build packages and create a release"
        print_status "Check the progress at: https://github.com/Auriora/OneMount/actions"
    else
        print_error "Failed to push tags"
        print_status "You can push manually later with: git push origin --tags"
        exit 1
    fi
fi

print_success "Release process completed!"
