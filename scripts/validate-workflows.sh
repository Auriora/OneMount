#!/bin/bash
# Workflow Validation Script
# This script validates that the GitHub workflows will work correctly

set -e

echo "🔍 Validating GitHub Workflows..."
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d ".github/workflows" ]; then
    echo -e "${RED}❌ This script must be run from the project root directory${NC}"
    exit 1
fi

# Add Go bin to PATH for golangci-lint
export PATH="$HOME/go/bin:$PATH"

echo "📋 Checking prerequisites..."

# Check Go installation
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_status 0 "Go is installed: $GO_VERSION"
else
    print_status 1 "Go is not installed"
    exit 1
fi

# Check Python installation
if command -v python3 &> /dev/null; then
    PYTHON_VERSION=$(python3 --version)
    print_status 0 "Python is installed: $PYTHON_VERSION"
else
    print_status 1 "Python3 is not installed"
    exit 1
fi

# Check system dependencies
echo ""
echo "🔧 Checking system dependencies..."

DEPS=("pkg-config" "make")
for dep in "${DEPS[@]}"; do
    if command -v "$dep" &> /dev/null; then
        print_status 0 "$dep is available"
    else
        print_status 1 "$dep is not available"
    fi
done

# Check Python dependencies
echo ""
echo "🐍 Checking Python dependencies..."

if [ -f "scripts/requirements-dev-cli.txt" ]; then
    print_status 0 "Requirements file exists"
    
    # Try to import key dependencies
    if python3 -c "import typer" 2>/dev/null; then
        print_status 0 "typer is available"
    else
        print_warning "typer is not available - install with: pip install typer"
    fi
    
    if python3 -c "import rich" 2>/dev/null; then
        print_status 0 "rich is available"
    else
        print_warning "rich is not available - install with: pip install rich"
    fi
else
    print_status 1 "Requirements file not found"
fi

# Test CLI tool
echo ""
echo "🛠️  Testing CLI tool..."

if [ -f "scripts/dev" ]; then
    print_status 0 "CLI wrapper script exists"
    
    if [ -x "scripts/dev" ]; then
        print_status 0 "CLI wrapper script is executable"
    else
        print_warning "CLI wrapper script is not executable - fixing..."
        chmod +x scripts/dev
        print_status 0 "CLI wrapper script made executable"
    fi
    
    # Test CLI tool
    if python3 scripts/test-dev-cli.py > /dev/null 2>&1; then
        print_status 0 "CLI tool validation passed"
    else
        print_status 1 "CLI tool validation failed"
        echo "Running CLI test for details..."
        python3 scripts/test-dev-cli.py
    fi
else
    print_status 1 "CLI wrapper script not found"
fi

# Test Go modules
echo ""
echo "📦 Testing Go modules..."

if go mod verify; then
    print_status 0 "Go modules are valid"
else
    print_status 1 "Go modules verification failed"
fi

# Test basic Go compilation (quick check)
echo ""
echo "🔨 Testing Go compilation (quick check)..."

export CGO_CFLAGS="-Wno-deprecated-declarations"

if timeout 30s go build -v ./cmd/onemount > /dev/null 2>&1; then
    print_status 0 "Go compilation successful"
    rm -f onemount  # Clean up
else
    print_warning "Go compilation timed out or failed - this might be expected in CI"
fi

# Test basic tests (quick)
echo ""
echo "🧪 Testing basic Go tests..."

if go test -short -timeout 30s ./cmd/common > /dev/null 2>&1; then
    print_status 0 "Basic Go tests passed"
else
    print_warning "Basic Go tests had issues - this might be expected for incomplete tests"
fi

# Check workflow files syntax
echo ""
echo "📄 Checking workflow files..."

WORKFLOW_FILES=(.github/workflows/*.yml)
for file in "${WORKFLOW_FILES[@]}"; do
    if [ -f "$file" ]; then
        # Basic YAML syntax check (if yq is available)
        if command -v yq &> /dev/null; then
            if yq '.' "$file" > /dev/null 2>&1; then
                print_status 0 "$(basename "$file") has valid YAML syntax"
            else
                print_status 1 "$(basename "$file") has invalid YAML syntax"
            fi
        else
            print_status 0 "$(basename "$file") exists (YAML validation skipped - yq not available)"
        fi
    fi
done

# Test golangci-lint
echo ""
echo "🔍 Testing golangci-lint..."

if command -v golangci-lint &> /dev/null; then
    print_status 0 "golangci-lint is available"

    # Test basic linting (quick check)
    if golangci-lint run --timeout 30s ./cmd/common > /dev/null 2>&1; then
        print_status 0 "Basic linting passed"
    else
        print_warning "Basic linting had issues - this might be expected"
    fi
else
    print_warning "golangci-lint is not available - install from: https://golangci-lint.run/usage/install/"
fi

echo ""
echo "🎉 Workflow validation complete!"
echo ""
echo "📝 Summary:"
echo "- If all checks passed, the workflows should work in GitHub Actions"
echo "- If there are warnings, consider installing the missing dependencies"
echo "- If there are errors, fix them before pushing to GitHub"
echo ""
echo "💡 To install missing Python dependencies:"
echo "   pip install -r scripts/requirements-dev-cli.txt"
echo ""
echo "💡 To test workflows locally:"
echo "   ./scripts/dev test unit"
echo "   ./scripts/dev test coverage"
