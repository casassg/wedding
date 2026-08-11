package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func openTestDB(t *testing.T) *Store {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "wedding-test-*.db")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	s, err := Open(f.Name())
	require.NoError(t, err)
	t.Cleanup(func() { s.DB.Close() })

	ddl, err := os.ReadFile(filepath.Join("..", "..", "migrations", "ddl.sql"))
	require.NoError(t, err)
	_, err = s.DB.Exec(string(ddl))
	require.NoError(t, err)
	return s
}

func seedInvite(t *testing.T, s *Store, code, location string) {
	t.Helper()
	ctx := context.Background()
	row := int64(1)
	err := s.UpsertInvite(ctx, &UpsertInviteParams{
		InviteCode:      code,
		Name:            "Test Guest",
		MaxAdults:       2,
		MaxKids:         0,
		ConfirmedAdults: 0,
		SheetRow:        &row,
		Location:        location,
	})
	require.NoError(t, err)
}

func TestUpdateTravelInfoRoundTrip(t *testing.T) {
	s := openTestDB(t)
	ctx := context.Background()
	seedInvite(t, s, "TESTCODE", "")

	err := s.UpdateTravelInfo(ctx, &UpdateTravelInfoParams{
		InputBusTo:         "thursday",
		InputPickup:        "sap",
		InputArrivalFlight: "AV 620",
		InputBusReturn:     "sunday_morning_sap",
		InputHotel:         "marina_copan",
		InputNotes:         "window seat please",
		InputReturnDetail:  "UA1422 · Sun, Dec 20 12:30",
		InputInviteCode:    "TESTCODE",
	})
	require.NoError(t, err)

	invite, err := s.GetInviteByInviteCode(ctx, "TESTCODE")
	require.NoError(t, err)
	require.Equal(t, "thursday", invite.TravelBusTo)
	require.Equal(t, "sap", invite.TravelPickup)
	require.Equal(t, "AV 620", invite.TravelArrivalFlight)
	require.Equal(t, "sunday_morning_sap", invite.TravelBusReturn)
	require.Equal(t, "marina_copan", invite.TravelHotel)
	require.Equal(t, "window seat please", invite.TravelNotes)
	require.Equal(t, "UA1422 · Sun, Dec 20 12:30", invite.TravelReturnDetail)
	require.NotNil(t, invite.TravelUpdatedAt)
	require.NotNil(t, invite.ResponseAt)
}

func TestUpsertInvitePreservesLocation(t *testing.T) {
	s := openTestDB(t)
	ctx := context.Background()
	seedInvite(t, s, "HN001", "HONDURAS")

	invite, err := s.GetInviteByInviteCode(ctx, "HN001")
	require.NoError(t, err)
	require.Equal(t, "HONDURAS", invite.Location)
}

func TestUpsertInviteDoesNotTouchTravelColumns(t *testing.T) {
	s := openTestDB(t)
	ctx := context.Background()
	seedInvite(t, s, "TRVL1", "")

	// SQLite's datetime('now') has second precision. Sleep past the second
	// boundary so response_at (set below) is strictly after updated_at (set
	// by seedInvite above) — otherwise the pending-changes guard below can't
	// tell the two events apart and the re-upsert would wrongly proceed.
	time.Sleep(1100 * time.Millisecond)

	// Write travel data first
	err := s.UpdateTravelInfo(ctx, &UpdateTravelInfoParams{
		InputBusTo:      "friday",
		InputPickup:     "welchez",
		InputBusReturn:  "sunday_afternoon_san_pedro",
		InputHotel:      "some_hotel",
		InputNotes:      "notes",
		InputInviteCode: "TRVL1",
	})
	require.NoError(t, err)

	// Re-upsert (simulates sheet sync) — travel columns must be preserved
	row := int64(1)
	err = s.UpsertInvite(ctx, &UpsertInviteParams{
		InviteCode:      "TRVL1",
		Name:            "Test Guest Updated",
		MaxAdults:       2,
		MaxKids:         0,
		ConfirmedAdults: 0,
		SheetRow:        &row,
		Location:        "SOMEWHERE",
	})
	require.NoError(t, err)

	invite, err := s.GetInviteByInviteCode(ctx, "TRVL1")
	require.NoError(t, err)
	require.Equal(t, "friday", invite.TravelBusTo)
}

func TestUpsertInviteRestoresTravelOnFreshBoot(t *testing.T) {
	s := openTestDB(t)
	ctx := context.Background()

	// Simulates a fresh deployment's empty DB being repopulated from a sheet
	// that already has travel answers for this guest (response_at is NULL,
	// so the pending-changes guard doesn't block the update).
	row := int64(1)
	responseAt := time.Now().UTC().Add(-24 * time.Hour)
	travelUpdatedAt := time.Now().UTC().Add(-12 * time.Hour)
	err := s.UpsertInvite(ctx, &UpsertInviteParams{
		InviteCode:          "TRVL2",
		Name:                "Test Guest",
		MaxAdults:           2,
		MaxKids:             0,
		ConfirmedAdults:     2,
		ResponseAt:          &responseAt,
		SheetRow:            &row,
		Location:            "HONDURAS",
		TravelBusTo:         "thursday",
		TravelPickup:        "sap",
		TravelArrivalFlight: "AV 620",
		TravelBusReturn:     "sunday_morning_sap",
		TravelHotel:         "marina_copan",
		TravelNotes:         "window seat please",
		TravelCocktail:      "yes",
		TravelBrunch:        "no",
		TravelReturnDetail:  "UA1422 · Sun, Dec 20 12:30",
		TravelUpdatedAt:     &travelUpdatedAt,
	})
	require.NoError(t, err)

	invite, err := s.GetInviteByInviteCode(ctx, "TRVL2")
	require.NoError(t, err)
	require.Equal(t, "thursday", invite.TravelBusTo)
	require.Equal(t, "sap", invite.TravelPickup)
	require.Equal(t, "AV 620", invite.TravelArrivalFlight)
	require.Equal(t, "sunday_morning_sap", invite.TravelBusReturn)
	require.Equal(t, "marina_copan", invite.TravelHotel)
	require.Equal(t, "window seat please", invite.TravelNotes)
	require.Equal(t, "yes", invite.TravelCocktail)
	require.Equal(t, "no", invite.TravelBrunch)
	require.Equal(t, "UA1422 · Sun, Dec 20 12:30", invite.TravelReturnDetail)
	require.NotNil(t, invite.TravelUpdatedAt)
}

func TestUpsertInviteTimestampNormalization(t *testing.T) {
	s := openTestDB(t)
	ctx := context.Background()

	// Simulate initial sheet sync with a response already recorded.
	responseAt := time.Date(2026, time.April, 8, 21, 8, 46, 0, time.UTC)
	travelUpdatedAt := time.Date(2026, time.April, 8, 20, 0, 0, 0, time.UTC)
	row := int64(5)
	err := s.UpsertInvite(ctx, &UpsertInviteParams{
		InviteCode:      "ts-test-1",
		Name:            "Timestamp Test",
		MaxAdults:       2,
		ConfirmedAdults: 1,
		ResponseAt:      &responseAt,
		SheetRow:        &row,
		TravelBusTo:     "friday",
		TravelUpdatedAt: &travelUpdatedAt,
	})
	require.NoError(t, err)

	// The invite must NOT show up as pending sync — the response_at and
	// updated_at must compare correctly after datetime() normalization.
	pending, err := s.GetPendingSyncInvites(ctx)
	require.NoError(t, err)
	require.Empty(t, pending, "sheet-sourced invite should not be pending sync")

	// A second upsert with a new sheet_row must update sheet_row (master data
	// is unconditional now).
	newRow := int64(10)
	err = s.UpsertInvite(ctx, &UpsertInviteParams{
		InviteCode:      "ts-test-1",
		Name:            "Timestamp Test Renamed",
		MaxAdults:       3,
		ConfirmedAdults: 1,
		ResponseAt:      &responseAt,
		SheetRow:        &newRow,
		TravelBusTo:     "friday",
		TravelUpdatedAt: &travelUpdatedAt,
	})
	require.NoError(t, err)

	invite, err := s.GetInviteByInviteCode(ctx, "ts-test-1")
	require.NoError(t, err)
	require.Equal(t, int64(10), *invite.SheetRow, "sheet_row should be updated")
	require.Equal(t, "Timestamp Test Renamed", invite.Name, "name should be updated")
	require.Equal(t, int64(3), invite.MaxAdults, "max_adults should be updated")
}

func TestUpsertGuardProtectsLocalChanges(t *testing.T) {
	s := openTestDB(t)
	ctx := context.Background()

	// Seed invite from sheet.
	seedInvite(t, s, "guard-1", "SPAIN")

	// Simulate 1s delay so response_at > updated_at.
	time.Sleep(1100 * time.Millisecond)

	// Guest submits RSVP locally.
	err := s.UpdateRSVP(ctx, &UpdateRSVPParams{
		InputConfirmedAdults: 2,
		InputConfirmedKids:   0,
		InputDietaryInfo:     "vegan",
		InputMessage:         "so excited!",
		InputSong:            "Bohemian Rhapsody",
		InputInviteCode:      "guard-1",
	})
	require.NoError(t, err)

	// Sheet sync arrives with different RSVP values AND a new sheet_row.
	newRow := int64(99)
	responseAt := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	err = s.UpsertInvite(ctx, &UpsertInviteParams{
		InviteCode:      "guard-1",
		Name:            "Updated Name",
		MaxAdults:       2,
		ConfirmedAdults: 0, // sheet has old values
		ConfirmedKids:   0,
		DietaryInfo:     "",
		MessageForUs:    "",
		SongRequest:     "",
		ResponseAt:      &responseAt,
		SheetRow:        &newRow,
		Location:        "CATALONIA",
	})
	require.NoError(t, err)

	invite, err := s.GetInviteByInviteCode(ctx, "guard-1")
	require.NoError(t, err)
	// Master data follows sheet.
	require.Equal(t, "Updated Name", invite.Name)
	require.Equal(t, int64(99), *invite.SheetRow)
	require.Equal(t, "CATALONIA", invite.Location)
	// Guest-entered fields preserved.
	require.Equal(t, int64(2), invite.ConfirmedAdults, "local RSVP adults should be preserved")
	require.Equal(t, "vegan", invite.DietaryInfo, "local dietary info should be preserved")
	require.Equal(t, "so excited!", invite.MessageForUs, "local message should be preserved")
	require.Equal(t, "Bohemian Rhapsody", invite.SongRequest, "local song should be preserved")

	// Mark as synced, then re-upsert — now everything should apply.
	err = s.MarkInviteSynced(ctx, "guard-1")
	require.NoError(t, err)

	err = s.UpsertInvite(ctx, &UpsertInviteParams{
		InviteCode:      "guard-1",
		Name:            "Final Name",
		MaxAdults:       2,
		ConfirmedAdults: 1,
		DietaryInfo:     "none",
		SheetRow:        &newRow,
		Location:        "CATALONIA",
	})
	require.NoError(t, err)

	invite, err = s.GetInviteByInviteCode(ctx, "guard-1")
	require.NoError(t, err)
	require.Equal(t, "Final Name", invite.Name)
	require.Equal(t, int64(1), invite.ConfirmedAdults, "after sync, upsert should apply sheet values")
	require.Equal(t, "none", invite.DietaryInfo, "after sync, upsert should apply sheet values")
}
