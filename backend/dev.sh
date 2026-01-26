#!/bin/bash
# Local development server runner
# Automatically loads .env and starts the server

set -e

cd "$(dirname "$0")"

echo "Starting Wedding RSVP Backend (Local Dev)..."
echo "============================================"
echo ""
echo "Environment:"
echo "  Database: ./tmp/wedding.db"
echo "  Server:   http://localhost:8080"
echo "  Health:   http://localhost:8080/health"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Create tmp directory if it doesn't exist
mkdir -p tmp

# Run migrations using sqlite3
sqlite3 tmp/wedding.db < migrations/ddl.sql

# Run the server (godotenv will load .env automatically)
go run ./cmd/server serve
