package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sqlcdb "github.com/casassg/wedding/backend/internal/db/sqlc"
)

// RSVPRequest represents an RSVP submission
type RSVPRequest struct {
	Attending    bool
	AdultCount   *int
	KidCount     *int
	DietaryInfo  string
	MessageForUs string
	SongRequest  string
}

// GetInviteByUUID retrieves an invite by its UUID
func (d *DB) GetInviteByUUID(uuid string) (*sqlcdb.Invite, error) {
	ctx := context.Background()
	invite, err := d.queries.GetInviteByUUID(ctx, uuid)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}

	return &invite, nil
}

// UpdateRSVP updates an invite with RSVP response data
func (d *DB) UpdateRSVP(uuid string, req RSVPRequest, country string) error {
	ctx := context.Background()
	var adultCount, kidCount sql.NullInt64
	if req.AdultCount != nil {
		adultCount = sql.NullInt64{Int64: int64(*req.AdultCount), Valid: true}
	}
	if req.KidCount != nil {
		kidCount = sql.NullInt64{Int64: int64(*req.KidCount), Valid: true}
	}

	now := time.Now()
	err := d.queries.UpdateRSVP(ctx, sqlcdb.UpdateRSVPParams{
		Attending:       sql.NullBool{Bool: req.Attending, Valid: true},
		AdultCount:      adultCount,
		KidCount:        kidCount,
		DietaryInfo:     sql.NullString{String: req.DietaryInfo, Valid: req.DietaryInfo != ""},
		MessageForUs:    sql.NullString{String: req.MessageForUs, Valid: req.MessageForUs != ""},
		SongRequest:     sql.NullString{String: req.SongRequest, Valid: req.SongRequest != ""},
		ResponseAt:      sql.NullTime{Time: now, Valid: true},
		ResponseCountry: sql.NullString{String: country, Valid: country != ""},
		UpdatedAt:       sql.NullTime{Time: now, Valid: true},
		Uuid:            uuid,
	})

	if err != nil {
		return fmt.Errorf("failed to update RSVP: %w", err)
	}

	return nil
}

// UpsertInvite inserts or updates an invite from Google Sheets sync
func (d *DB) UpsertInvite(uuid, name string, maxAdults, maxKids int, sheetRow sql.NullInt64) error {
	ctx := context.Background()
	now := time.Now()

	err := d.queries.UpsertInvite(ctx, sqlcdb.UpsertInviteParams{
		Uuid:      uuid,
		Name:      name,
		MaxAdults: int64(maxAdults),
		MaxKids:   int64(maxKids),
		SheetRow:  sheetRow,
		SyncedAt:  sql.NullTime{Time: now, Valid: true},
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: sql.NullTime{Time: now, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to upsert invite: %w", err)
	}

	return nil
}

// SoftDeleteInvite marks an invite as deleted
func (d *DB) SoftDeleteInvite(uuid string) error {
	ctx := context.Background()
	now := time.Now()
	err := d.queries.SoftDeleteInvite(ctx, sqlcdb.SoftDeleteInviteParams{
		DeletedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: sql.NullTime{Time: now, Valid: true},
		Uuid:      uuid,
	})

	if err != nil {
		return fmt.Errorf("failed to soft delete invite: %w", err)
	}

	return nil
}

// GetPendingSyncInvites returns invites that need to be synced to Google Sheets
func (d *DB) GetPendingSyncInvites() ([]sqlcdb.Invite, error) {
	ctx := context.Background()
	rows, err := d.queries.GetPendingSyncInvites(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sync invites: %w", err)
	}

	return rows, nil
}

// MarkInviteSynced updates the synced_at timestamp
func (d *DB) MarkInviteSynced(uuid string) error {
	ctx := context.Background()
	now := time.Now()
	err := d.queries.MarkInviteSynced(ctx, sqlcdb.MarkInviteSyncedParams{
		SyncedAt:  sql.NullTime{Time: now, Valid: true},
		UpdatedAt: sql.NullTime{Time: now, Valid: true},
		Uuid:      uuid,
	})

	if err != nil {
		return fmt.Errorf("failed to mark invite as synced: %w", err)
	}

	return nil
}
