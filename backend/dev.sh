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
echo "  Server:   http://localhost:8081"
echo "  Health:   http://localhost:8081/health"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Create tmp directory if it doesn't exist
mkdir -p tmp

# Run the server (godotenv will load .env automatically)
go run cmd/server/main.go
