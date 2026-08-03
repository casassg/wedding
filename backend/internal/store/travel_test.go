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
		InputBusReturn:     "sunday_sap",
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
	require.Equal(t, "sunday_sap", invite.TravelBusReturn)
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
		InputBusReturn:  "sunday_san_pedro",
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
		TravelBusReturn:     "sunday_sap",
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
	require.Equal(t, "sunday_sap", invite.TravelBusReturn)
	require.Equal(t, "marina_copan", invite.TravelHotel)
	require.Equal(t, "window seat please", invite.TravelNotes)
	require.Equal(t, "yes", invite.TravelCocktail)
	require.Equal(t, "no", invite.TravelBrunch)
	require.Equal(t, "UA1422 · Sun, Dec 20 12:30", invite.TravelReturnDetail)
	require.NotNil(t, invite.TravelUpdatedAt)
}
