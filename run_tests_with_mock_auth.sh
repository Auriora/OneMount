#!/bin/bash

# Set environment variable to use mock authentication
export ONEDRIVER_MOCK_AUTH=1

# Run the tests
go test ./fs/...

# Print success message
echo "Tests completed with mock authentication"