#!/bin/bash

# OneMount Personal OneDrive CI Setup Script
# This script helps you set up CI system tests with your personal OneDrive account

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

# Function to display usage
usage() {
    echo "OneMount Personal OneDrive CI Setup"
    echo ""
    echo "This script helps you set up CI system tests using your personal OneDrive account."
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  --check-auth            Check if authentication is available"
    echo "  --generate-secret       Generate the GitHub secret value"
    echo "  --verify-setup          Verify the CI setup is working"
    echo ""
    echo "Examples:"
    echo "  $0 --check-auth         # Check if OneMount authentication exists"
    echo "  $0 --generate-secret    # Generate the secret for GitHub"
    echo "  $0 --verify-setup       # Test the complete setup"
}

# Global variable to store the selected auth file path
SELECTED_AUTH_FILE=""

# Function to check if OneMount authentication exists
check_auth() {
    print_status "Checking OneMount authentication..."

    AUTH_FILE="$HOME/.cache/onemount/auth_tokens.json"
    TEST_AUTH_FILE="$HOME/.onemount-tests/.auth_tokens.json"

    # Check both locations
    if [[ -f "$AUTH_FILE" ]]; then
        SELECTED_AUTH_FILE="$AUTH_FILE"
        print_status "Found authentication at: $AUTH_FILE"
    elif [[ -f "$TEST_AUTH_FILE" ]]; then
        SELECTED_AUTH_FILE="$TEST_AUTH_FILE"
        print_status "Found authentication at: $TEST_AUTH_FILE"
    else
        print_error "OneMount authentication not found"
        print_error "Checked locations:"
        print_error "  - $AUTH_FILE"
        print_error "  - $TEST_AUTH_FILE"
        print_error ""
        print_error "Please authenticate with OneMount first:"
        print_error "  make onemount"
        print_error "  ./build/onemount --auth-only"
        print_error ""
        print_error "Follow the authentication prompts to sign in to your OneDrive account."
        return 1
    fi
    
    # Check if the file is valid JSON
    if ! python3 -c "import json; json.load(open('$SELECTED_AUTH_FILE'))" 2>/dev/null; then
        print_error "Authentication file exists but is not valid JSON"
        print_error "Please re-authenticate with OneMount"
        return 1
    fi

    # Check token expiration
    EXPIRES_AT=$(python3 -c "import json; data=json.load(open('$SELECTED_AUTH_FILE')); print(data.get('expires_at', 0))")
    CURRENT_TIME=$(date +%s)
    
    if [[ "$EXPIRES_AT" -le "$CURRENT_TIME" ]]; then
        print_warning "Authentication tokens appear to be expired"
        print_warning "Please re-authenticate with OneMount:"
        print_warning "  ./build/onemount --auth-only"
        return 1
    fi
    
    # Get account info if available
    ACCOUNT=$(python3 -c "import json; data=json.load(open('$SELECTED_AUTH_FILE')); print(data.get('account', 'Unknown'))")
    TIME_LEFT=$((EXPIRES_AT - CURRENT_TIME))
    HOURS_LEFT=$((TIME_LEFT / 3600))
    
    print_success "OneMount authentication found and valid"
    print_status "Account: $ACCOUNT"
    print_status "Token expires in: ${HOURS_LEFT} hours"
    
    return 0
}

# Function to generate the GitHub secret value
generate_secret() {
    print_status "Generating GitHub secret value..."

    if ! check_auth; then
        return 1
    fi

    # Generate base64-encoded secret using the selected auth file
    SECRET_VALUE=$(base64 -w 0 "$SELECTED_AUTH_FILE")
    
    print_success "GitHub secret value generated successfully!"
    print_status ""
    print_status "Copy the following value and add it as a GitHub secret:"
    print_status ""
    print_status "Secret Name: ONEDRIVE_PERSONAL_TOKENS"
    print_status "Secret Value:"
    echo "$SECRET_VALUE"
    print_status ""
    print_status "To add this secret to your GitHub repository:"
    print_status "1. Go to your repository on GitHub"
    print_status "2. Click Settings → Secrets and variables → Actions"
    print_status "3. Click 'New repository secret'"
    print_status "4. Name: ONEDRIVE_PERSONAL_TOKENS"
    print_status "5. Value: Paste the value above"
    print_status "6. Click 'Add secret'"
    print_status ""
    print_status "After adding the secret, the CI workflow will automatically run system tests!"
}

# Function to verify the setup
verify_setup() {
    print_status "Verifying CI setup..."
    
    # Check if workflow file exists
    WORKFLOW_FILE=".github/workflows/system-tests-personal.yml"
    if [[ ! -f "$WORKFLOW_FILE" ]]; then
        print_error "CI workflow file not found: $WORKFLOW_FILE"
        print_error "Please ensure the workflow file is committed to your repository"
        return 1
    fi
    
    print_success "CI workflow file found: $WORKFLOW_FILE"
    
    # Check authentication
    if ! check_auth; then
        return 1
    fi
    
    # Test OneDrive access
    print_status "Testing OneDrive access..."

    # Use the selected auth file from check_auth function
    if [[ -z "$SELECTED_AUTH_FILE" ]]; then
        # Fallback to checking auth again if SELECTED_AUTH_FILE is not set
        AUTH_FILE="$HOME/.cache/onemount/auth_tokens.json"
        TEST_AUTH_FILE="$HOME/.onemount-tests/.auth_tokens.json"

        if [[ -f "$AUTH_FILE" ]]; then
            SELECTED_AUTH_FILE="$AUTH_FILE"
        elif [[ -f "$TEST_AUTH_FILE" ]]; then
            SELECTED_AUTH_FILE="$TEST_AUTH_FILE"
        else
            print_error "Authentication file not found"
            return 1
        fi
    fi

    # Check if jq is available, if not use python as fallback
    if command -v jq >/dev/null 2>&1; then
        ACCESS_TOKEN=$(jq -r '.access_token' "$SELECTED_AUTH_FILE")
    else
        ACCESS_TOKEN=$(python3 -c "import json; data=json.load(open('$SELECTED_AUTH_FILE')); print(data.get('access_token', ''))")
    fi

    if [[ -z "$ACCESS_TOKEN" || "$ACCESS_TOKEN" == "null" ]]; then
        print_error "Could not extract access token from authentication file"
        return 1
    fi

    RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
        "https://graph.microsoft.com/v1.0/me/drive/root" 2>/dev/null || echo "")

    # Check if response contains an ID field (indicates success)
    if command -v jq >/dev/null 2>&1; then
        if echo "$RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
            DRIVE_NAME=$(echo "$RESPONSE" | jq -r '.name // "Unknown"')
            print_success "OneDrive access verified successfully"
            print_status "Drive Name: $DRIVE_NAME"
        else
            print_error "Failed to access OneDrive"
            print_error "Please check your internet connection and re-authenticate"
            return 1
        fi
    else
        # Use python as fallback for JSON parsing
        if echo "$RESPONSE" | python3 -c "import json, sys; data=json.load(sys.stdin); print(data.get('id', ''))" 2>/dev/null | grep -q .; then
            DRIVE_NAME=$(echo "$RESPONSE" | python3 -c "import json, sys; data=json.load(sys.stdin); print(data.get('name', 'Unknown'))" 2>/dev/null)
            print_success "OneDrive access verified successfully"
            print_status "Drive Name: $DRIVE_NAME"
        else
            print_error "Failed to access OneDrive"
            print_error "Please check your internet connection and re-authenticate"
            return 1
        fi
    fi
    
    # Check if we can run tests locally
    print_status "Checking if system tests can run locally..."
    
    if [[ ! -f "scripts/run-system-tests.sh" ]]; then
        print_error "System test script not found: scripts/run-system-tests.sh"
        return 1
    fi
    
    # Copy auth tokens to test location (if not already there)
    mkdir -p ~/.onemount-tests
    COPIED_AUTH_FILE=false
    if [[ "$SELECTED_AUTH_FILE" != "$HOME/.onemount-tests/.auth_tokens.json" ]]; then
        cp "$SELECTED_AUTH_FILE" ~/.onemount-tests/.auth_tokens.json
        chmod 600 ~/.onemount-tests/.auth_tokens.json
        COPIED_AUTH_FILE=true
    else
        print_status "Auth tokens already in test location"
    fi

    # Test the script (dry run)
    if ./scripts/run-system-tests.sh --help > /dev/null 2>&1; then
        print_success "System test script is executable and ready"
    else
        print_error "System test script has issues"
        return 1
    fi

    # Clean up test auth file only if we copied it
    if [[ "$COPIED_AUTH_FILE" == "true" ]]; then
        rm -f ~/.onemount-tests/.auth_tokens.json
        print_status "Cleaned up temporary auth file"
    fi
    
    print_success "✅ CI setup verification completed successfully!"
    print_status ""
    print_status "Your setup is ready! To complete the CI configuration:"
    print_status "1. Run: $0 --generate-secret"
    print_status "2. Add the generated secret to GitHub"
    print_status "3. Push your code to trigger the CI tests"
    print_status ""
    print_status "You can also manually trigger tests in GitHub Actions:"
    print_status "- Go to Actions tab → System Tests (Personal OneDrive) → Run workflow"
}

# Main execution
main() {
    case "${1:-}" in
        -h|--help)
            usage
            exit 0
            ;;
        --check-auth)
            check_auth
            ;;
        --generate-secret)
            generate_secret
            ;;
        --verify-setup)
            verify_setup
            ;;
        "")
            print_status "OneMount Personal OneDrive CI Setup"
            print_status ""
            print_status "This script will help you set up CI system tests with your personal OneDrive."
            print_status ""
            print_status "Step 1: Check authentication"
            if check_auth; then
                print_status ""
                print_status "Step 2: Generate GitHub secret"
                generate_secret
                print_status ""
                print_status "Step 3: Verify complete setup"
                verify_setup
            fi
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
