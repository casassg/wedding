package db

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpsertInvitePreservesSyncedAt(t *testing.T) {
	// Create temporary database
	tmpDB, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	// Initialize database
	database, err := New(tmpDB.Name())
	require.NoError(t, err)
	defer database.Close()

	testUUID := "test-invite-123"
	sheetRow := sql.NullInt64{Int64: 42, Valid: true}

	// Step 1: Initial upsert from Google Sheets (first sync)
	err = database.UpsertInvite(testUUID, "John Doe", 2, 0, sheetRow)
	require.NoError(t, err)

	// Verify initial state - should have synced_at set
	invite1, err := database.GetInviteByUUID(testUUID)
	require.NoError(t, err)
	require.NotNil(t, invite1)
	require.True(t, invite1.SyncedAt.Valid, "synced_at should be set after initial upsert")

	// Wait a bit to ensure timestamps are different
	time.Sleep(10 * time.Millisecond)

	// Step 2: User submits RSVP
	rsvpReq := RSVPRequest{
		Attending:      true,
		AdultCount:     intPtr(2),
		KidCount:       intPtr(0),
		DietaryInfo:    "Vegetarian",
		TransportNeeds: "Need shuttle",
	}
	err = database.UpdateRSVP(testUUID, rsvpReq, "US")
	require.NoError(t, err)

	// Verify RSVP was recorded and synced_at was reset to NULL
	invite2, err := database.GetInviteByUUID(testUUID)
	require.NoError(t, err)
	require.NotNil(t, invite2)
	require.True(t, invite2.ResponseAt.Valid, "response_at should be set after RSVP")
	require.False(t, invite2.SyncedAt.Valid, "synced_at should be NULL after RSVP to trigger sync")

	// Step 3: Verify invite appears in pending sync list
	pendingInvites, err := database.GetPendingSyncInvites()
	require.NoError(t, err)
	require.Len(t, pendingInvites, 1, "should have 1 pending invite to sync")
	require.Equal(t, testUUID, pendingInvites[0].Uuid)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Step 4: Simulate sync FROM sheet (this is where the bug was)
	// This should NOT update synced_at since the record already exists
	err = database.UpsertInvite(testUUID, "John Doe", 2, 0, sheetRow)
	require.NoError(t, err)

	// Step 5: Verify synced_at is STILL NULL (the bug fix)
	invite3, err := database.GetInviteByUUID(testUUID)
	require.NoError(t, err)
	require.NotNil(t, invite3)
	require.False(t, invite3.SyncedAt.Valid, "synced_at should STILL be NULL after upsert (bug fix)")
	require.True(t, invite3.ResponseAt.Valid, "response_at should still be set")

	// Step 6: Verify invite STILL appears in pending sync list
	pendingInvites2, err := database.GetPendingSyncInvites()
	require.NoError(t, err)
	require.Len(t, pendingInvites2, 1, "should STILL have 1 pending invite after sheet upsert")
	require.Equal(t, testUUID, pendingInvites2[0].Uuid)

	// Step 7: Mark as synced (simulating successful sync TO sheet)
	err = database.MarkInviteSynced(testUUID)
	require.NoError(t, err)

	// Step 8: Verify synced_at is now set and invite no longer pending
	invite4, err := database.GetInviteByUUID(testUUID)
	require.NoError(t, err)
	require.NotNil(t, invite4)
	require.True(t, invite4.SyncedAt.Valid, "synced_at should be set after marking synced")
	require.True(t, invite4.SyncedAt.Time.After(invite4.ResponseAt.Time), "synced_at should be after response_at")

	pendingInvites3, err := database.GetPendingSyncInvites()
	require.NoError(t, err)
	require.Len(t, pendingInvites3, 0, "should have 0 pending invites after marking synced")

	// Step 9: One more upsert from sheet should preserve the synced_at
	finalSyncedAt := invite4.SyncedAt.Time
	time.Sleep(10 * time.Millisecond)

	err = database.UpsertInvite(testUUID, "John Doe Updated", 2, 0, sheetRow)
	require.NoError(t, err)

	invite5, err := database.GetInviteByUUID(testUUID)
	require.NoError(t, err)
	require.NotNil(t, invite5)
	require.True(t, invite5.SyncedAt.Valid, "synced_at should still be valid")
	require.Equal(t, finalSyncedAt.Unix(), invite5.SyncedAt.Time.Unix(), "synced_at should be preserved, not updated")
	require.Equal(t, "John Doe Updated", invite5.Name, "name should be updated from sheet")
}

func TestNewInviteSetsSyncedAt(t *testing.T) {
	// Create temporary database
	tmpDB, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	// Initialize database
	database, err := New(tmpDB.Name())
	require.NoError(t, err)
	defer database.Close()

	testUUID := "new-invite-456"
	sheetRow := sql.NullInt64{Int64: 99, Valid: true}

	// Insert new invite
	err = database.UpsertInvite(testUUID, "Jane Smith", 1, 2, sheetRow)
	require.NoError(t, err)

	// Verify synced_at is set for new records
	invite, err := database.GetInviteByUUID(testUUID)
	require.NoError(t, err)
	require.NotNil(t, invite)
	require.True(t, invite.SyncedAt.Valid, "synced_at should be set for new invites")
	require.False(t, invite.ResponseAt.Valid, "response_at should be NULL for new invites")
}

// Helper function
func intPtr(i int) *int {
	return &i
}
