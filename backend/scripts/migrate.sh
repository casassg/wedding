#!/bin/sh
# Migration script that runs only on primary node
# Called by LiteFS before starting the application

set -e

DB_PATH="/litefs/wedding.db"

echo "Running database migrations..."

# Wait for LiteFS to be ready
until [ -d /litefs ]; do
    echo "Waiting for LiteFS to mount..."
    sleep 1
done

# Create tables if they don't exist
sqlite3 "$DB_PATH" <<'EOF'
CREATE TABLE IF NOT EXISTS invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    max_adults INTEGER NOT NULL DEFAULT 1,
    max_kids INTEGER NOT NULL DEFAULT 0,
    attending INTEGER,
    adult_count INTEGER,
    kid_count INTEGER,
    dietary_info TEXT,
    transport_needs TEXT,
    response_at DATETIME,
    response_country TEXT,
    sheet_row INTEGER,
    synced_at DATETIME,
    deleted_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invites_uuid ON invites(uuid);
CREATE INDEX IF NOT EXISTS idx_invites_deleted_at ON invites(deleted_at);
CREATE INDEX IF NOT EXISTS idx_invites_response_at ON invites(response_at);
EOF

echo "Migrations completed successfully"
