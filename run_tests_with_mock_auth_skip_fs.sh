#!/bin/bash

# Set environment variable to use mock authentication
export ONEDRIVER_MOCK_AUTH=1

# Run the tests for fs/graph and fs/offline packages
echo "Running tests for fs/graph package..."
go test ./fs/graph/...

echo "Running tests for fs/offline package..."
go test ./fs/offline/...

# Print success message
echo "Tests completed with mock authentication (skipping fs package tests)"