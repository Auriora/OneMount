#!/bin/bash

# OneMount Workflow Optimization Test Script
# This script tests the optimization implementations

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

print_info "OneMount Workflow Optimization Test Suite"
echo "=========================================="

# Test 1: Check if optimization files exist
print_info "Test 1: Checking optimization files..."

required_files=(
    "scripts/setup-optimized-runners.sh"
    "scripts/monitor-workflow-performance.py"
    "docs/WORKFLOW_OPTIMIZATION_GUIDE.md"
    ".github/workflow-optimization.yml"
)

for file in "${required_files[@]}"; do
    if [[ -f "$PROJECT_ROOT/$file" ]]; then
        print_success "Found: $file"
    else
        print_error "Missing: $file"
        exit 1
    fi
done

# Test 2: Check if scripts are executable
print_info "Test 2: Checking script permissions..."

executable_scripts=(
    "scripts/setup-optimized-runners.sh"
    "scripts/monitor-workflow-performance.py"
)

for script in "${executable_scripts[@]}"; do
    if [[ -x "$PROJECT_ROOT/$script" ]]; then
        print_success "Executable: $script"
    else
        print_error "Not executable: $script"
        exit 1
    fi
done

# Test 3: Validate workflow syntax
print_info "Test 3: Validating workflow syntax..."

workflow_files=(
    ".github/workflows/ci.yml"
    ".github/workflows/coverage.yml"
    ".github/workflows/build-packages.yml"
)

for workflow in "${workflow_files[@]}"; do
    if command -v yamllint &> /dev/null; then
        if yamllint "$PROJECT_ROOT/$workflow" &> /dev/null; then
            print_success "Valid YAML: $workflow"
        else
            print_error "Invalid YAML: $workflow"
            yamllint "$PROJECT_ROOT/$workflow"
            exit 1
        fi
    else
        print_warning "yamllint not available, skipping YAML validation"
        break
    fi
done

# Test 4: Check Docker optimization
print_info "Test 4: Checking Docker optimization..."

dockerfile="packaging/docker/Dockerfile.deb-builder"
if grep -q "FROM.*AS base" "$PROJECT_ROOT/$dockerfile"; then
    print_success "Multi-stage Dockerfile detected"
else
    print_error "Multi-stage Dockerfile not found"
    exit 1
fi

if grep -q "mount=type=cache" "$PROJECT_ROOT/$dockerfile"; then
    print_success "BuildKit cache mounts detected"
else
    print_error "BuildKit cache mounts not found"
    exit 1
fi

# Test 5: Check workflow runner configuration
print_info "Test 5: Checking workflow runner configuration..."

ci_workflow=".github/workflows/ci.yml"
if grep -q "self-hosted.*onemount-testing" "$PROJECT_ROOT/$ci_workflow"; then
    print_success "Self-hosted runner configuration found in CI workflow"
else
    print_error "Self-hosted runner configuration not found in CI workflow"
    exit 1
fi

# Test 6: Check caching improvements
print_info "Test 6: Checking caching improvements..."

if grep -q "actions/cache@v4" "$PROJECT_ROOT/$ci_workflow"; then
    print_success "Updated cache action version found"
else
    print_error "Cache action not updated to v4"
    exit 1
fi

if grep -q "go-build" "$PROJECT_ROOT/$ci_workflow"; then
    print_success "Go build cache configuration found"
else
    print_error "Go build cache configuration not found"
    exit 1
fi

# Test 7: Test setup script help
print_info "Test 7: Testing setup script help..."

if "$PROJECT_ROOT/scripts/setup-optimized-runners.sh" --help &> /dev/null; then
    print_success "Setup script help works"
else
    print_error "Setup script help failed"
    exit 1
fi

# Test 8: Check BuildKit optimization in build workflow
print_info "Test 8: Checking BuildKit optimization..."

build_workflow=".github/workflows/build-packages.yml"
if grep -q "docker buildx build" "$PROJECT_ROOT/$build_workflow"; then
    print_success "BuildKit optimization found in build workflow"
else
    print_error "BuildKit optimization not found in build workflow"
    exit 1
fi

if grep -q "cache-from type=gha" "$PROJECT_ROOT/$build_workflow"; then
    print_success "GitHub Actions cache integration found"
else
    print_error "GitHub Actions cache integration not found"
    exit 1
fi

# Test 9: Validate optimization config
print_info "Test 9: Validating optimization configuration..."

config_file=".github/workflow-optimization.yml"
if command -v python3 &> /dev/null; then
    if python3 -c "import yaml; yaml.safe_load(open('$PROJECT_ROOT/$config_file'))" &> /dev/null; then
        print_success "Optimization config is valid YAML"
    else
        print_error "Optimization config has invalid YAML"
        exit 1
    fi
else
    print_warning "Python3 not available, skipping config validation"
fi

# Test 10: Check documentation completeness
print_info "Test 10: Checking documentation completeness..."

guide_file="docs/WORKFLOW_OPTIMIZATION_GUIDE.md"
required_sections=(
    "Quick Start"
    "Performance Monitoring"
    "Architecture Overview"
    "Troubleshooting"
)

for section in "${required_sections[@]}"; do
    if grep -q "$section" "$PROJECT_ROOT/$guide_file"; then
        print_success "Documentation section found: $section"
    else
        print_error "Documentation section missing: $section"
        exit 1
    fi
done

# Summary
print_info "Test Summary"
echo "============"
print_success "All optimization tests passed!"
echo ""
print_info "Next steps:"
echo "1. Set up self-hosted runners: ./scripts/setup-optimized-runners.sh setup-all"
echo "2. Monitor performance: python3 scripts/monitor-workflow-performance.py"
echo "3. Read the guide: docs/WORKFLOW_OPTIMIZATION_GUIDE.md"
echo ""
print_success "Optimization implementation is ready for use!"
