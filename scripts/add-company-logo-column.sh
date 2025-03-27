#!/bin/bash

# This script adds the company_logo column to the jobs table in the database
# It should be run once to update the schema

# Load configuration from .env file
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
  echo "Loaded configuration from .env file"
else
  echo "Warning: No .env file found"
fi

# Determine which connection string to use
if [ "$MODE" = "production" ]; then
  CONNECTION_STRING="$POSTGRES_CONNECTION_PROD"
  echo "Using production database"
else
  CONNECTION_STRING="$POSTGRES_CONNECTION_LOCAL"
  echo "Using development database"
fi

if [ -z "$CONNECTION_STRING" ]; then
  echo "Error: Database connection string not found in environment"
  exit 1
fi

# Add the company_logo column to the jobs table
echo "Adding company_logo column to jobs table..."

# Create a temporary SQL file
SQL_FILE=$(mktemp)
cat > "$SQL_FILE" << EOF
-- Check if column exists and add if it doesn't
DO \$\$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'jobs' AND column_name = 'company_logo'
    ) THEN
        ALTER TABLE jobs ADD COLUMN company_logo TEXT;
        RAISE NOTICE 'Added company_logo column to jobs table';
    ELSE
        RAISE NOTICE 'company_logo column already exists in jobs table';
    END IF;
END \$\$;
EOF

# Execute the SQL using psql
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE"

# Remove the temporary SQL file
rm "$SQL_FILE"

echo "Migration completed. Now you can run the database regeneration script." 