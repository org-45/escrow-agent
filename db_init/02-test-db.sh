#!/bin/bash
set -e

echo "Checking if test database exists..."
DB_EXIST=$(psql -U "$POSTGRES_USER" -tAc "SELECT 1 FROM pg_database WHERE datname = '${TEST_DB_NAME}'")

if [ "$DB_EXIST" != "1" ]; then
    echo "Creating test database: ${TEST_DB_NAME}..."
    psql -U "$POSTGRES_USER" -c "CREATE DATABASE ${TEST_DB_NAME};"
    
    echo "Copying schema from ${POSTGRES_DB} to ${TEST_DB_NAME}..."
    pg_dump --schema-only --no-owner --no-privileges -U "$POSTGRES_USER" "$POSTGRES_DB" | psql -U "$POSTGRES_USER" -d "$TEST_DB_NAME"
    
    echo "Test database ${TEST_DB_NAME} initialized successfully."
else
    echo "Test database ${TEST_DB_NAME} already exists. Skipping creation."
fi
