package store

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type Store struct {
	*Queries
	DB *sql.DB
}

// New creates a new database connection
func Open(dbPath string) (*Store, error) {
	sqlDB, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single writer
	sqlDB.SetMaxIdleConns(1)

	log.Println("Database connection established")

	return &Store{
		Queries: New(sqlDB),
		DB:      sqlDB,
	}, nil
}
