#!/bin/bash

# Exit on any error
set -e

# Set environment variables for test
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=kogase_test
export DB_SSL_MODE=disable
export JWT_SECRET=test_secret
export PORT=8081

# Create test database if it doesn't exist
echo "Creating test database if it doesn't exist..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -c "CREATE DATABASE $DB_NAME"

echo "Running integration tests..."
cd ./integration
go test -v ./...

echo "All tests passed!" 