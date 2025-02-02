#!/bin/sh
set -e

echo "Running database migrations..."

migrate -path /app/database/migrations -database "postgres://$(cat config.json | jq -r '.db_url')" up

echo "Starting Discord bot..."
exec ./meido
