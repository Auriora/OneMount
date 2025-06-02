#!/usr/bin/env python3
"""
Script to fix deferred time.Since() calls in Go code.

This script finds patterns like:
    defer logging.LogMethodExit(methodName, time.Since(startTime), result)

And converts them to:
    defer func() {
        logging.LogMethodExit(methodName, time.Since(startTime), result)
    }()
"""

import re
import sys
import os

def fix_deferred_time_since(file_path):
    """Fix deferred time.Since calls in a Go file."""
    with open(file_path, 'r') as f:
        lines = f.readlines()

    original_lines = lines[:]
    modified = False

    i = 0
    while i < len(lines):
        line = lines[i].strip()

        # Look for defer statements with time.Since
        if line.startswith('defer ') and 'time.Since(' in line:
            # Check if this is a logging method call
            if 'logging.LogMethodExit(' in line or 'logging.LogMethodExitWithContext(' in line:
                # Find the complete statement (might span multiple lines)
                statement_lines = [lines[i]]
                j = i + 1

                # Count parentheses to find the end of the statement
                open_parens = line.count('(') - line.count(')')
                while j < len(lines) and open_parens > 0:
                    statement_lines.append(lines[j])
                    open_parens += lines[j].count('(') - lines[j].count(')')
                    j += 1

                # Extract the method call (everything after 'defer ')
                full_statement = ''.join(statement_lines)
                defer_match = re.match(r'(\s*)defer\s+(.*)', full_statement, re.DOTALL)

                if defer_match:
                    indent = defer_match.group(1)
                    method_call = defer_match.group(2).strip()

                    # Create the new defer func() block
                    new_lines = [
                        f'{indent}defer func() {{\n',
                        f'{indent}\t{method_call}\n',
                        f'{indent}}}()\n'
                    ]

                    # Replace the original lines
                    lines[i:j] = new_lines
                    modified = True

                    # Skip the lines we just processed
                    i += len(new_lines)
                    continue

        i += 1

    if modified:
        with open(file_path, 'w') as f:
            f.writelines(lines)
        print(f"Fixed deferred time.Since calls in {file_path}")
        return True
    else:
        print(f"No changes needed in {file_path}")
        return False

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 fix-deferred-time-since.py <file1> [file2] ...")
        sys.exit(1)

    files_changed = 0
    for file_path in sys.argv[1:]:
        if os.path.exists(file_path) and file_path.endswith('.go'):
            if fix_deferred_time_since(file_path):
                files_changed += 1
        else:
            print(f"Skipping {file_path} (not a Go file or doesn't exist)")

    print(f"\nFixed {files_changed} files")

if __name__ == "__main__":
    main()
