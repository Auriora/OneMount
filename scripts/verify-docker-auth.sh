#!/bin/bash
# Verify Docker authentication integration
# This script tests that Docker containers can access authentication tokens

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

print_info "Docker Authentication Integration Verification"
print_info "=============================================="
echo ""

# Check 1: Host tokens exist
print_info "Check 1: Verifying tokens on host..."
if [ -f "$HOME/.onemount-tests/.auth_tokens.json" ]; then
    print_success "Tokens found on host: $HOME/.onemount-tests/.auth_tokens.json"
    
    # Check if valid JSON
    if command -v jq > /dev/null 2>&1; then
        if jq empty "$HOME/.onemount-tests/.auth_tokens.json" 2>/dev/null; then
            ACCOUNT=$(jq -r '.account' "$HOME/.onemount-tests/.auth_tokens.json")
            EXPIRES_AT=$(jq -r '.expires_at' "$HOME/.onemount-tests/.auth_tokens.json")
            CURRENT_TIME=$(date +%s)
            
            print_info "  Account: $ACCOUNT"
            
            if [ "$EXPIRES_AT" -gt "$CURRENT_TIME" ]; then
                EXPIRES_DATE=$(date -d "@$EXPIRES_AT" 2>/dev/null || date -r "$EXPIRES_AT" 2>/dev/null)
                print_success "  Tokens valid until: $EXPIRES_DATE"
            else
                print_error "  Tokens are EXPIRED"
                print_info "  Run: ./scripts/setup-test-auth.sh --refresh"
                exit 1
            fi
        else
            print_error "Tokens are not valid JSON"
            exit 1
        fi
    fi
else
    print_error "No tokens found on host"
    print_info "Run: ./scripts/setup-test-auth.sh"
    exit 1
fi
echo ""

# Check 2: Docker Compose file exists
print_info "Check 2: Verifying Docker Compose configuration..."
if [ -f "docker/compose/docker-compose.test.yml" ]; then
    print_success "Docker Compose file found"
    
    # Check if volume mount is configured
    if grep -q "/.onemount-tests:/tmp/home-tester/.onemount-tests-auth:ro" docker/compose/docker-compose.test.yml; then
        print_success "Auth token volume mount configured"
    else
        print_warning "Auth token volume mount not found in docker-compose.test.yml"
        print_info "Expected line: - \${HOME}/.onemount-tests:/tmp/home-tester/.onemount-tests-auth:ro"
    fi
else
    print_error "Docker Compose file not found"
    exit 1
fi
echo ""

# Check 3: Docker is available
print_info "Check 3: Verifying Docker availability..."
if command -v docker > /dev/null 2>&1; then
    print_success "Docker is installed"
    DOCKER_VERSION=$(docker --version)
    print_info "  $DOCKER_VERSION"
else
    print_error "Docker is not installed"
    exit 1
fi

if docker compose version > /dev/null 2>&1; then
    print_success "Docker Compose is available"
    COMPOSE_VERSION=$(docker compose version)
    print_info "  $COMPOSE_VERSION"
else
    print_error "Docker Compose is not available"
    exit 1
fi
echo ""

# Check 4: Test token mounting in Docker
print_info "Check 4: Testing token mounting in Docker..."
print_info "Starting Docker container to verify token access..."

if docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint bash shell \
    -c "ls -la /tmp/home-tester/.onemount-tests-auth/.auth_tokens.json" > /dev/null 2>&1; then
    print_success "Tokens are accessible in Docker container"
    
    # Get account info from Docker
    DOCKER_ACCOUNT=$(docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint bash shell \
        -c "cat /tmp/home-tester/.onemount-tests-auth/.auth_tokens.json 2>/dev/null | jq -r '.account' 2>/dev/null" || echo "unknown")
    
    if [ "$DOCKER_ACCOUNT" != "unknown" ] && [ -n "$DOCKER_ACCOUNT" ]; then
        print_success "  Docker can read tokens"
        print_info "  Account in Docker: $DOCKER_ACCOUNT"
    else
        print_warning "  Could not read account from Docker (but file is accessible)"
    fi
else
    print_error "Tokens are NOT accessible in Docker container"
    print_info "Possible issues:"
    print_info "  1. Docker Compose file not updated with volume mount"
    print_info "  2. Docker doesn't have permission to access $HOME/.onemount-tests/"
    print_info "  3. Docker Compose version too old"
    exit 1
fi
echo ""

# Check 5: Test auth helper library
print_info "Check 5: Verifying auth helper library..."
if [ -f "scripts/lib/auth-helper.sh" ]; then
    print_success "Auth helper library found"
    
    # Test in Docker
    if docker compose -f docker/compose/docker-compose.test.yml run --rm --entrypoint bash shell \
        -c "source /workspace/scripts/lib/auth-helper.sh && setup_auth_tokens" > /dev/null 2>&1; then
        print_success "Auth helper works in Docker"
    else
        print_warning "Auth helper may have issues in Docker"
    fi
else
    print_warning "Auth helper library not found at scripts/lib/auth-helper.sh"
fi
echo ""

# Summary
print_success "=============================================="
print_success "Docker Authentication Integration: VERIFIED"
print_success "=============================================="
echo ""
print_info "Your Docker containers can now access authentication tokens!"
echo ""
print_info "Next steps:"
print_info "  1. Run tests: docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner /workspace/scripts/test-task-5.4-filesystem-operations.sh"
print_info "  2. Run system tests: docker compose -f docker/compose/docker-compose.test.yml run --rm system-tests"
print_info "  3. Interactive shell: docker compose -f docker/compose/docker-compose.test.yml run --rm shell"
echo ""
