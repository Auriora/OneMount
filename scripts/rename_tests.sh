#!/bin/bash
#
# Rename Tests Script for OneMount
#
# This script renames test functions in the OneMount project based on a file input listing name changes.
# It can be used to resolve duplicate test IDs or to rename tests for other reasons.
#
# Functionality:
#     - Reads a file containing old and new test function names
#     - Renames the test functions in all test files
#     - Logs the rename operations
#
# Usage:
#     ./rename_tests.sh <input_file>
#
# Arguments:
#     input_file - File containing the name changes, one per line, in the format "old_name:new_name"
#
# Example input file:
#     TestUT_FS_01_01_FileOperations_BasicReadWrite_SuccessfullyPreservesContent:TestUT_FS_01_02_FileOperations_BasicReadWrite_SuccessfullyPreservesContent
#     TestUT_GR_01_01_ResourcePath_ValidPath_ReturnsCorrectURL:TestUT_GR_01_02_ResourcePath_ValidPath_ReturnsCorrectURL
#
# Output:
#     - Console output with the results of the rename operations
#     - A log file with the results of the rename operations (tmp/rename_tests.log)
#
# Note:
#     All scripts save their output to the `tmp/` directory by default. This directory is created 
#     automatically if it doesn't exist.
#
# Example:
#     ./rename_tests.sh rename_list.txt
#     This will rename the test functions listed in rename_list.txt.
#
# Author: OneMount Team
#

set -e

# Check if input file is provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <input_file>"
    echo "  input_file - File containing the name changes, one per line, in the format \"old_name:new_name\""
    exit 1
fi

INPUT_FILE="$1"

# Check if input file exists
if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: Input file '$INPUT_FILE' does not exist."
    exit 1
fi

# Create tmp directory if it doesn't exist
mkdir -p tmp

# Log file
LOG_FILE="tmp/rename_tests.log"
echo "Rename Tests Log - $(date)" > "$LOG_FILE"
echo "Input file: $INPUT_FILE" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# Process each line in the input file
while IFS=: read -r old_name new_name; do
    # Skip empty lines and comments
    if [[ -z "$old_name" || "$old_name" == \#* ]]; then
        continue
    fi

    # Skip if old_name and new_name are the same
    if [ "$old_name" = "$new_name" ]; then
        echo "Skipping $old_name (no change needed)"
        echo "Skipping $old_name (no change needed)" >> "$LOG_FILE"
        continue
    fi

    echo "Renaming $old_name to $new_name"
    echo "Renaming $old_name to $new_name" >> "$LOG_FILE"

    # Find all test files and rename the function
    find . -name '*_test.go' -exec sed -i "s/$old_name/$new_name/g" {} \;

    # Check if the rename was successful
    if [ $? -eq 0 ]; then
        echo "  Success"
        echo "  Success" >> "$LOG_FILE"
    else
        echo "  Failed"
        echo "  Failed" >> "$LOG_FILE"
    fi
done < "$INPUT_FILE"

echo "Rename operations complete. Log written to $LOG_FILE"
