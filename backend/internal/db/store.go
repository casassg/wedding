package db

import (
	"database/sql"
	"fmt"
	"log"

	sqlcdb "github.com/casassg/wedding/backend/internal/db/sqlc"
	_ "modernc.org/sqlite"
)

// DB wraps the database connection and sqlc queries
type DB struct {
	db      *sql.DB
	queries *sqlcdb.Queries
}

// New creates a new database connection
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

	log.Println("Database connection established")
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
