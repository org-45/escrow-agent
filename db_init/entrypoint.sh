#!/bin/bash
set -e

echo "Starting PostgreSQL..."
# Start the official PostgreSQL entrypoint (which initializes the main DB if needed)
# in the background.
exec /usr/local/bin/docker-entrypoint.sh postgres &

# Wait for PostgreSQL to become available
echo "Waiting for PostgreSQL to become available..."
until pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"; do
    sleep 1
done

echo "PostgreSQL is ready!"

# Run the test database initialization script
echo "Running test database initialization..."
/docker-entrypoint-initdb.d/02-test-db.sh

# Wait indefinitely (forward signals to the background process)
wait
