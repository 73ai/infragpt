#!/bin/bash

# Script to run database migrations
# Usage: ./run_migration.sh [migration_file]

set -e

# Default migration file
MIGRATION_FILE=${1:-"001_identity_tables.sql"}

# Database connection details from config.yaml
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="infragpt"
DB_USER="infragpt"
DB_PASSWORD="infragptisgreatsre"

echo "Running migration: $MIGRATION_FILE"
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
echo ""

# Check if migration file exists
if [ ! -f "$MIGRATION_FILE" ]; then
    echo "Error: Migration file '$MIGRATION_FILE' not found"
    exit 1
fi

# Run the migration
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"

echo ""
echo "Migration completed successfully!"