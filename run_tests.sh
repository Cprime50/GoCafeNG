#!/bin/bash

# This script temporarily replaces the .env file with .env.test for running tests,
# then restores the original .env file when done

# Backup the test environment file
cp .env.test .env.test.bak

# Backup the original .env file and replace it with the test one
mv .env .env.orig
cp .env.test .env

echo "Running tests with test environment..."

# Run the specified tests or all tests if none specified
if [ $# -eq 0 ]; then
  go test -v ./internal/...
else
  go test -v "$@"
fi

TEST_EXIT_CODE=$?

# Restore the original .env file
mv .env.orig .env
cp .env.test.bak .env.test
rm -f .env.test.bak

echo "Restored original environment."

# Exit with the same exit code as the tests
exit $TEST_EXIT_CODE 