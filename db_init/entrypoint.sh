#!/bin/bash
set -e

# start PostgreSQL using the default entrypoint in the background.
echo "Starting PostgreSQL using the default entrypoint..."
/usr/local/bin/docker-entrypoint.sh postgres &


echo "Waiting for PostgreSQL to become available..."

sleep 10

echo "PostgreSQL is ready!"


echo "Running test-db.sh to create test database..."
/docker-entrypoint-initdb.d/02-test-db.sh


wait
