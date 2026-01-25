package sheets

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/casassg/wedding/backend/internal/db"
)

// Syncer handles bidirectional sync between Google Sheets and the database
type Syncer struct {
	db            *db.DB
	client        *Client
	currentRegion string
	primaryRegion string
}

// NewSyncer creates a new syncer
func NewSyncer(database *db.DB, client *Client, currentRegion, primaryRegion string) *Syncer {
	return &Syncer{
		db:            database,
		client:        client,
		currentRegion: currentRegion,
		primaryRegion: primaryRegion,
	}
}

// Start begins the background sync loop
func (s *Syncer) Start(ctx context.Context, interval time.Duration) {
	// Only run sync on primary region to avoid write conflicts
	if s.currentRegion != s.primaryRegion {
		log.Printf("Google Sheets sync disabled (not primary region: %s != %s)", s.currentRegion, s.primaryRegion)
		return
	}

	if !s.client.IsConfigured() {
		log.Println("Google Sheets sync disabled (credentials not configured)")
		return
	}

	log.Printf("Starting Google Sheets sync every %s (primary region: %s)", interval, s.primaryRegion)

	// Initial sync
	s.runSync(ctx)

	// Start ticker
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runSync(ctx)
		case <-ctx.Done():
			log.Println("Stopping Google Sheets sync")
			return
		}
	}
}

// SyncOnce performs a single sync cycle (used for manual sync command)
func (s *Syncer) SyncOnce(ctx context.Context) error {
	// Check if running in primary region
	if s.currentRegion != s.primaryRegion {
		log.Printf("Skipping sync: not primary region (%s != %s)", s.currentRegion, s.primaryRegion)
		return nil
	}

	if !s.client.IsConfigured() {
		return fmt.Errorf("Google Sheets credentials not configured")
	}

	s.runSync(ctx)
	return nil
}

// runSync performs one complete sync cycle
func (s *Syncer) runSync(ctx context.Context) {
	log.Println("Starting sync cycle...")

	// Sync from sheet to DB (master data)
	if err := s.syncFromSheet(ctx); err != nil {
		log.Printf("Error syncing from sheet: %v", err)
	}

	// Sync from DB to sheet (RSVP responses)
	if err := s.syncToSheet(ctx); err != nil {
		log.Printf("Error syncing to sheet: %v", err)
	}

	log.Println("Sync cycle completed")
}

// syncFromSheet reads the sheet and updates the database
func (s *Syncer) syncFromSheet(ctx context.Context) error {
	rows, err := s.client.ReadSheet(ctx)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		log.Println("No invites found in sheet")
		return nil
	}

	// Track which UUIDs exist in the sheet
	sheetUUIDs := make(map[string]bool)

	// Upsert each row into the database
	for _, row := range rows {
		sheetUUIDs[row.InviteCode] = true

		// Convert Parella to max_adults
		maxAdults := 1
		if row.Parella == "Si" || row.Parella == "si" || row.Parella == "SI" {
			maxAdults = 2
		}

		sheetRow := sql.NullInt64{Int64: int64(row.RowNumber), Valid: true}
		if err := s.db.UpsertInvite(row.InviteCode, row.Name, maxAdults, row.Fills, sheetRow); err != nil {
			log.Printf("Failed to upsert invite %s: %v", row.InviteCode, err)
			continue
		}
	}

	log.Printf("Synced %d invites from sheet to database", len(rows))

	// TODO: Implement soft-delete for invites not in sheet
	// This requires getting all UUIDs from DB and comparing
	// For now, we skip this to keep it simple

	return nil
}

// syncToSheet writes pending RSVP responses back to the sheet
func (s *Syncer) syncToSheet(ctx context.Context) error {
	// Get invites that need syncing
	invites, err := s.db.GetPendingSyncInvites()
	if err != nil {
		return err
	}

	if len(invites) == 0 {
		log.Println("No pending RSVPs to sync to sheet")
		return nil
	}

	log.Printf("Syncing %d RSVP responses to sheet", len(invites))

	// Write each invite to the sheet
	for _, invite := range invites {
		if !invite.SheetRow.Valid {
			log.Printf("Skipping invite %s: no sheet row number", invite.Uuid)
			continue
		}

		rowData := SheetRow{
			RowNumber:   int(invite.SheetRow.Int64),
			Attending:   nullBoolToPtr(invite.Attending),
			Adults:      nullIntToPtr(invite.AdultCount),
			Kids:        nullIntToPtr(invite.KidCount),
			Dietary:     nullStringToString(invite.DietaryInfo),
			Transport:   nullStringToString(invite.TransportNeeds),
			RespondedAt: nullTimeToString(invite.ResponseAt),
			Country:     nullStringToString(invite.ResponseCountry),
		}

		if err := s.client.WriteRSVP(ctx, rowData.RowNumber, rowData); err != nil {
			log.Printf("Failed to write RSVP for invite %s: %v", invite.Uuid, err)
			continue
		}

		// Mark as synced in database
		if err := s.db.MarkInviteSynced(invite.Uuid); err != nil {
			log.Printf("Failed to mark invite %s as synced: %v", invite.Uuid, err)
		}
	}

	log.Printf("Successfully synced %d RSVPs to sheet", len(invites))
	return nil
}

// Helper functions for converting sql.Null* types

func nullBoolToPtr(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	b := nb.Bool
	return &b
}

func nullIntToPtr(ni sql.NullInt64) *int {
	if !ni.Valid {
		return nil
	}
	i := int(ni.Int64)
	return &i
}

func nullStringToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func nullTimeToString(nt sql.NullTime) string {
	if !nt.Valid {
		return ""
	}
	return nt.Time.Format(time.RFC3339)
}
