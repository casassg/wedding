package store

import (
	"path/filepath"
	"testing"
)

func TestBeginRecoversDanglingTransaction(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "wedding.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.DB.Close(); err != nil {
			t.Error(err)
		}
	})

	var journalMode string
	if err := s.DB.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatal(err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal mode = %q, want wal", journalMode)
	}

	_, err = s.DB.Exec("BEGIN")
	if err != nil {
		t.Fatal(err)
	}

	tx, err := s.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
}
