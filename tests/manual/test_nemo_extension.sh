#!/bin/bash
# Manual Test Script for Nemo Extension with Real OneDrive
# Task 13.5: Test Nemo extension with manual verification
#
# This script provides guidance for manually testing the Nemo file manager
# extension with a real OneDrive mount. This test MUST be run outside Docker
# on a system with a graphical environment and Nemo file manager installed.
#
# Requirements: 8.3 (Nemo extension integration)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
MOUNT_POINT="${MOUNT_POINT:-$HOME/OneDrive}"
EXTENSION_PATH="$HOME/.local/share/nemo-python/extensions/nemo-onemount.py"
SOURCE_EXTENSION="internal/nemo/src/nemo-onemount.py"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Nemo Extension Manual Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print test step
print_step() {
    echo -e "${YELLOW}[STEP]${NC} $1"
}

# Function to print success
print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Function to print info
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Function to wait for user confirmation
wait_for_user() {
    echo -e "${YELLOW}Press Enter to continue...${NC}"
    read -r
}

# Check if running in Docker
if [ -f /.dockerenv ] || grep -q docker /proc/1/cgroup 2>/dev/null; then
    print_error "This test MUST be run outside Docker on a system with GUI"
    print_error "The Nemo extension requires:"
    print_error "  - Graphical environment (X11 or Wayland)"
    print_error "  - Nemo file manager installed"
    print_error "  - Real OneDrive mount"
    echo ""
    print_info "Please run this script on your host system:"
    print_info "  ./tests/manual/test_nemo_extension.sh"
    exit 1
fi

# Check prerequisites
print_step "Checking prerequisites..."

# Check if Nemo is installed
if ! command -v nemo &> /dev/null; then
    print_error "Nemo file manager is not installed"
    print_info "Install with:"
    print_info "  Ubuntu/Debian: sudo apt install nemo"
    print_info "  Linux Mint: Usually pre-installed"
    exit 1
fi
print_success "Nemo file manager is installed"

# Check if python-nemo or python3-nemo is installed
if ! dpkg -l | grep -q -E "python3?-nemo"; then
    print_error "python-nemo (or python3-nemo) is not installed"
    print_info "Install with one of:"
    print_info "  Ubuntu/Debian: sudo apt install python3-nemo python3-gi"
    print_info "  Linux Mint: sudo apt install python-nemo python3-gi"
    exit 1
fi
print_success "python-nemo is installed"

# Check if source extension exists
if [ ! -f "$SOURCE_EXTENSION" ]; then
    print_error "Source extension not found at: $SOURCE_EXTENSION"
    print_info "Make sure you're running from the OneMount repository root"
    exit 1
fi
print_success "Source extension found"

echo ""
print_step "Installing Nemo extension..."

# Create extensions directory if it doesn't exist
mkdir -p "$(dirname "$EXTENSION_PATH")"

# Copy extension
cp "$SOURCE_EXTENSION" "$EXTENSION_PATH"
chmod +x "$EXTENSION_PATH"
print_success "Extension installed to: $EXTENSION_PATH"

echo ""
print_step "Restarting Nemo to load extension..."
nemo -q 2>/dev/null || true
sleep 2
print_success "Nemo restarted"

echo ""
print_step "Checking if OneMount is mounted..."

if ! mount | grep -q "onemount"; then
    print_error "OneMount is not mounted"
    print_info "Please mount OneMount first:"
    print_info "  onemount $MOUNT_POINT"
    echo ""
    print_info "After mounting, run this script again"
    exit 1
fi

ACTUAL_MOUNT=$(mount | grep onemount | awk '{print $3}' | head -n 1)
print_success "OneMount is mounted at: $ACTUAL_MOUNT"
MOUNT_POINT="$ACTUAL_MOUNT"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Manual Verification Steps${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

print_info "The following steps require manual verification in Nemo"
echo ""

# Test 1: Open Nemo and navigate to mount
print_step "Test 1: Open Nemo and navigate to mounted OneDrive"
echo ""
print_info "Action: Opening Nemo at mount point..."
nemo "$MOUNT_POINT" &
NEMO_PID=$!
sleep 3
echo ""
print_info "Verify the following:"
echo "  1. Nemo window opened successfully"
echo "  2. You can see your OneDrive files"
echo "  3. Files display in the file manager"
echo ""
wait_for_user

# Test 2: Verify status icons appear
print_step "Test 2: Verify status icons appear on files"
echo ""
print_info "Look for status emblems on files:"
echo "  - Cloud icon (emblem-synchronizing-offline): File not cached"
echo "  - Check mark (emblem-default): File cached locally"
echo "  - Sync icon (emblem-synchronizing): File syncing"
echo "  - Download icon (emblem-downloads): File downloading"
echo "  - Warning icon (emblem-warning): Conflict"
echo "  - Error icon (emblem-error): Sync error"
echo ""
print_info "Do you see status icons on the files? (y/n)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    print_success "Status icons are visible"
else
    print_error "Status icons are NOT visible"
    print_info "Troubleshooting:"
    print_info "  1. Check Nemo debug output: nemo --quit && nemo --debug"
    print_info "  2. Verify extension is loaded: look for 'nemo-onemount' in debug output"
    print_info "  3. Check extension permissions: ls -l $EXTENSION_PATH"
    print_info "  4. Verify D-Bus service is running: ps aux | grep onemount"
fi
echo ""
wait_for_user

# Test 3: Trigger file operations and watch icons update
print_step "Test 3: Trigger file operations and watch icons update"
echo ""
print_info "Perform the following operations and observe icon changes:"
echo ""
echo "  a) Open a file (double-click)"
echo "     - Icon should change to 'downloading' if not cached"
echo "     - Then change to 'cached' after download completes"
echo ""
echo "  b) Create a new file"
echo "     - Right-click > Create Document > Empty File"
echo "     - Icon should show 'syncing' or 'modified'"
echo ""
echo "  c) Modify an existing file"
echo "     - Open a text file, make changes, save"
echo "     - Icon should update to show 'modified' or 'syncing'"
echo ""
echo "  d) Delete a file"
echo "     - Right-click > Move to Trash"
echo "     - File should disappear or show deletion status"
echo ""
print_info "Perform these operations now..."
wait_for_user

print_info "Did the icons update correctly during file operations? (y/n)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    print_success "Icons updated correctly during file operations"
else
    print_error "Icons did NOT update correctly"
    print_info "This may indicate:"
    print_info "  1. D-Bus signals are not being emitted"
    print_info "  2. Extension is not receiving signals"
    print_info "  3. Status update logic has issues"
fi
echo ""
wait_for_user

# Test 4: Test with different file states
print_step "Test 4: Test with different file states"
echo ""
print_info "Navigate to different folders and observe:"
echo "  1. Folders with many files (100+)"
echo "  2. Folders with large files"
echo "  3. Folders with recently modified files"
echo "  4. Folders with files that have conflicts"
echo ""
print_info "Do icons display correctly for all file states? (y/n)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    print_success "Icons display correctly for all file states"
else
    print_error "Icons do NOT display correctly for all states"
    print_info "Note which states have issues for documentation"
fi
echo ""
wait_for_user

# Test 5: Test performance with many files
print_step "Test 5: Test performance with many files"
echo ""
print_info "Navigate to a folder with many files (if available)"
print_info "Observe:"
echo "  1. Does Nemo remain responsive?"
echo "  2. Do icons load quickly?"
echo "  3. Is there any lag or freezing?"
echo ""
print_info "Is performance acceptable? (y/n)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    print_success "Performance is acceptable"
else
    print_error "Performance issues detected"
    print_info "This may indicate the extension needs optimization"
fi
echo ""
wait_for_user

# Test 6: Test D-Bus fallback
print_step "Test 6: Test D-Bus fallback (optional)"
echo ""
print_info "This test verifies the extension works without D-Bus"
print_info "Skip this test if you want to keep D-Bus running (y/n)"
read -r response
if [[ ! "$response" =~ ^[Yy]$ ]]; then
    print_info "Testing D-Bus fallback..."
    echo ""
    print_info "1. Close Nemo: nemo -q"
    nemo -q 2>/dev/null || true
    sleep 2
    
    print_info "2. Stop OneMount (this will stop D-Bus service)"
    print_info "   You'll need to manually unmount and remount"
    print_info "   Press Enter when ready to continue..."
    read -r
    
    print_info "3. Remount OneMount without D-Bus"
    print_info "   (Implementation specific - may need special flag)"
    print_info "   Press Enter when remounted..."
    read -r
    
    print_info "4. Open Nemo again"
    nemo "$MOUNT_POINT" &
    sleep 3
    
    print_info "Do icons still appear (using extended attributes)? (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_success "D-Bus fallback works correctly"
    else
        print_error "D-Bus fallback does NOT work"
    fi
fi
echo ""

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

print_info "Manual verification completed"
print_info "Please document your findings in:"
print_info "  docs/verification-tracking.md (Phase 11 section)"
echo ""

print_info "Key points to document:"
echo "  1. Whether status icons appeared correctly"
echo "  2. Whether icons updated during file operations"
echo "  3. Performance with many files"
echo "  4. Any issues or unexpected behavior"
echo "  5. D-Bus fallback behavior (if tested)"
echo ""

print_info "To uninstall the extension:"
print_info "  rm $EXTENSION_PATH"
print_info "  nemo -q && nemo"
echo ""

# Cleanup
if [ -n "$NEMO_PID" ] && ps -p $NEMO_PID > /dev/null 2>&1; then
    print_info "Nemo is still running (PID: $NEMO_PID)"
    print_info "Close it manually when done testing"
fi

echo -e "${GREEN}Test script completed${NC}"
