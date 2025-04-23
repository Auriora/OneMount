#!/bin/bash

# Build onedriver if needed
if [ ! -f ./onedriver ] || [ "$(find cmd -type f -name "*.go" -newer onedriver | wc -l)" -gt 0 ]; then
    echo "Building onedriver..."
    make
fi

# Test with the problematic command
echo "Testing with 'f' as mountpoint (should show helpful error):"
./onedriver f .config/onedriver/config.yml /home/bcherrington/Projects/Goland/onedriver/mount 2>&1 | grep -i "mountpoint"

# Test with correct flag usage
echo -e "\nTesting with correct flag usage:"
mkdir -p test_mount
./onedriver -f .config/onedriver/config.yml test_mount

# Clean up
echo -e "\nCleaning up..."
rmdir test_mount 2>/dev/null
