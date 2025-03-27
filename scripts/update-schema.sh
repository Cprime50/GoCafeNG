#!/bin/bash

# This script runs the update-schema tool to add the company_logo column to the jobs table

echo "Building schema update tool..."
go build -o update-schema ./cmd/update-schema

if [ $? -ne 0 ]; then
  echo "Build failed. Exiting."
  exit 1
fi

echo "Running schema update..."
./update-schema

if [ $? -ne 0 ]; then
  echo "Schema update failed. Check the logs for details."
  exit 1
fi

echo "Schema update completed successfully!"
echo "Now you can run the database regeneration script."

# Clean up the binary
rm update-schema 