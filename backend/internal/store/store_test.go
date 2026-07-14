package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
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

func TestUpsertInviteImportsRSVP(t *testing.T) {
	s := openTestDB(t)

	responseAt := time.Date(2026, time.April, 8, 21, 8, 46, 0, time.UTC)
	sheetRow := int64(2)
	err := s.UpsertInvite(context.Background(), &UpsertInviteParams{
		InviteCode:      "active-moon-6338",
		Name:            "Arnau",
		MaxAdults:       1,
		ConfirmedAdults: 1,
		ConfirmedKids:   0,
		DietaryInfo:     "Pregunta: la mare i la Silvia necesiten invitacio extra per les llagostes i els cocos o no conten com a acompanyants?",
		MessageForUs:    "",
		SongRequest:     "WHERE IS MY HUSBAND! - RAYE",
		ResponseAt:      &responseAt,
		SheetRow:        &sheetRow,
	})
	if err != nil {
		t.Fatal(err)
	}

	invite, err := s.GetInviteByInviteCode(context.Background(), "active-moon-6338")
	if err != nil {
		t.Fatal(err)
	}
	if invite.ResponseAt == nil {
		t.Fatal("response timestamp was not imported")
	}
	if invite.ConfirmedAdults != 1 || invite.ConfirmedKids != 0 {
		t.Fatalf("confirmed guests = %d adults, %d kids", invite.ConfirmedAdults, invite.ConfirmedKids)
	}
	if invite.DietaryInfo != "Pregunta: la mare i la Silvia necesiten invitacio extra per les llagostes i els cocos o no conten com a acompanyants?" {
		t.Fatalf("dietary info = %q", invite.DietaryInfo)
	}
	if invite.SongRequest != "WHERE IS MY HUSBAND! - RAYE" {
		t.Fatalf("song request = %q", invite.SongRequest)
	}
}
