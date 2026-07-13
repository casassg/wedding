package store

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

type Store struct {
	*Queries
	DB *sql.DB
}

// New creates a new database connection
func Open(dbPath string) (*Store, error) {
	sqlDB, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_txlock=immediate")
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

// Begin starts a transaction, clearing a transaction left open by a failed commit.
func (s *Store) Begin() (*sql.Tx, error) {
	tx, err := s.DB.Begin()
	if err != nil && strings.Contains(err.Error(), "within a transaction") {
		_, _ = s.DB.Exec("ROLLBACK")
		tx, err = s.DB.Begin()
	}
	return tx, err
}
