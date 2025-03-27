#!/bin/bash

# Script to rebuild the database from cached API responses
# This is useful when you've made schema changes or want to reprocess 
# all jobs through the filtering logic

echo "Building database regeneration tool..."
go build -o db-regenerate ./cmd/db-regenerate

if [ $? -ne 0 ]; then
  echo "Build failed. Exiting."
  exit 1
fi

echo "Running database regeneration..."
./db-regenerate

if [ $? -ne 0 ]; then
  echo "Database regeneration failed. Check the logs for details."
  exit 1
fi

echo "Database regeneration completed successfully!"
echo "The jobs database has been cleared and repopulated with filtered data from cached API responses."

# Clean up the binary
rm db-regenerate 