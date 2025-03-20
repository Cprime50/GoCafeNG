#!/bin/bash
# run.sh - Simple script to run Go9jaJobs

# Default port
PORT=${PORT:-8080}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Check for required environment variables
if [ -z "$POSTGRES_CONNECTION" ]; then
    echo "Warning: POSTGRES_CONNECTION environment variable is not set."
    echo "You can set it by running: export POSTGRES_CONNECTION='your-connection-string'"
    # Default to a local PostgreSQL connection for development
    export POSTGRES_CONNECTION="postgres://postgres:postgres@localhost:5432/go9jajobs?sslmode=disable"
    echo "Using default connection string: $POSTGRES_CONNECTION"
fi

echo "Starting Go9jaJobs application..."
echo "================================================================"
echo "  URL: http://localhost:$PORT"
echo "  To change the port: export PORT=<your-port> before running"
echo "  To disable auto browser opening: export OPEN_BROWSER=false"
echo "================================================================"

# Start the application
go run cmd/server/main.go 