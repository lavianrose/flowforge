#!/bin/sh
set -e

echo "Starting FlowForge Backend..."

# Run migrations
echo "Running database migrations..."
./migrate up

# Run seed (idempotent, safe to run multiple times)
echo "Seeding database..."
./seed

# Start the API
echo "Starting API server..."
exec ./api
