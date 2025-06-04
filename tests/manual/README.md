# Manual Testing Utilities

This directory contains standalone test utilities for manual testing and debugging of OneMount functionality.

## Available Tests

### test_signal_handling.go

A standalone test utility for manually testing the upload manager's signal handling capabilities.

**Purpose:**
- Tests upload manager initialization with signal handling
- Verifies upload session creation and progress tracking
- Tests JSON marshaling without infinite recursion
- Validates graceful shutdown on signal reception (SIGTERM/SIGINT)
- Verifies upload progress persistence capability

**Usage:**
```bash
cd tests/manual
go run test_signal_handling.go
```

The test will:
1. Create a mock upload manager with signal handling
2. Simulate upload activity
3. Wait for either Ctrl+C or automatically send SIGTERM after 10 seconds
4. Test graceful shutdown and cleanup

**Expected Output:**
- Upload manager initialization messages
- Signal handling setup confirmation
- Upload session creation details
- Signal reception and graceful shutdown process
- Verification of key features

## Adding New Manual Tests

When adding new manual test utilities to this directory:

1. Use descriptive filenames (e.g., `test_<feature>_<scenario>.go`)
2. Include a `main()` function for standalone execution
3. Add comprehensive logging to show test progress
4. Include cleanup of any temporary resources
5. Update this README with usage instructions

## Notes

- These tests are for development and debugging purposes
- They are not part of the automated test suite
- Run them manually when investigating specific functionality
- Ensure proper cleanup of any resources created during testing
