package db

import (
	"database/sql"
	"fmt"
	"log"

	sqlcdb "github.com/casassg/wedding/backend/internal/db/sqlc"
	_ "modernc.org/sqlite"
)

const migrationSQL = `
-- Initial schema for wedding RSVP system
CREATE TABLE IF NOT EXISTS invites (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid            TEXT UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    max_adults      INTEGER NOT NULL DEFAULT 1,
    max_kids        INTEGER NOT NULL DEFAULT 0,
    
    -- RSVP response fields
    attending       BOOLEAN,
    adult_count     INTEGER,
    kid_count       INTEGER,
    dietary_info    TEXT,
    transport_needs TEXT,
    
    -- Metadata
    response_at     DATETIME,
    response_country TEXT,
    
    -- Sync tracking
    sheet_row       INTEGER,
    synced_at       DATETIME,
    deleted_at      DATETIME,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invites_uuid ON invites(uuid);
CREATE INDEX IF NOT EXISTS idx_invites_deleted ON invites(deleted_at);
CREATE INDEX IF NOT EXISTS idx_invites_synced ON invites(synced_at);
`

// DB wraps the database connection and sqlc queries
type DB struct {
	db      *sql.DB
	queries *sqlcdb.Queries
}

// New creates a new database connection and runs migrations
func New(dbPath string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single writer
	sqlDB.SetMaxIdleConns(1)

	db := &DB{
		db:      sqlDB,
		queries: sqlcdb.New(sqlDB),
	}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.db.Close()
}

// Queries returns the sqlc-generated queries
func (d *DB) Queries() *sqlcdb.Queries {
	return d.queries
}

// runMigrations executes all migration files
func (d *DB) runMigrations() error {
	// Execute migration
	_, err := d.db.Exec(migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
