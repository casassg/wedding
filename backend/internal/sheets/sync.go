package sheets

import (
	"context"
	"log"
	"time"

	"github.com/casassg/wedding/backend/internal/store"
	"github.com/pkg/errors"
)

// Default wedding year for schedule parsing
const defaultWeddingYear = 2026

// Syncer handles bidirectional sync between Google Sheets and the database
type Syncer struct {
	store        *store.Store
	sheetsClient *Client
	listener     chan struct{}
}

// NewSyncer creates a new syncer
func NewSyncer(s *store.Store, client *Client) *Syncer {
	return &Syncer{
		store:        s,
		sheetsClient: client,
		listener:     make(chan struct{}),
	}
}

// Start begins the background sync loop
func (s *Syncer) Start(ctx context.Context, interval time.Duration) {
	if !s.sheetsClient.IsConfigured() {
		log.Println("Google Sheets sync disabled (credentials not configured)")
		return
	}

	log.Printf("Starting Google Sheets sync every %s", interval)

	// Start ticker
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.SyncOnce(ctx); err != nil {
				log.Printf("Error during sync: %v", err)
			}
		case <-s.listener:
			log.Println("Received manual sync request")
			if err := s.SyncOnce(ctx); err != nil {
				log.Printf("Error during manual sync: %v", err)
			}
		case <-ctx.Done():
			log.Println("Stopping Google Sheets sync")
			return
		}
	}
}

// TriggerSync signals the syncer to perform an immediate sync
func (s *Syncer) TriggerSync() {
	s.listener <- struct{}{}
}

// SyncOnce performs a single sync cycle (used for manual sync command)
func (s *Syncer) SyncOnce(ctx context.Context) error {
	if !s.sheetsClient.IsConfigured() {
		return errors.New("Google Sheets credentials not configured")
	}

	// Sync invites from sheet to DB (master data)
	if err := s.SyncFromSheet(ctx); err != nil {
		return errors.Wrap(err, "sync from sheet failed")
	}

	// Sync RSVPs from DB to sheet (responses)
	if err := s.SyncToSheet(ctx); err != nil {
		return errors.Wrap(err, "sync to sheet failed")
	}

	// Sync schedule from sheet to DB (one-way, sheet is source of truth)
	if err := s.SyncScheduleFromSheet(ctx); err != nil {
		return errors.Wrap(err, "sync schedule from sheet failed")
	}

	log.Println("Sync cycle completed")
	return nil
}

// SyncFromSheet reads the sheet and updates the database
func (s *Syncer) SyncFromSheet(ctx context.Context) error {
	rows, err := s.sheetsClient.ReadSheet(ctx)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		log.Println("No invites found in sheet, skipping...")
		return nil
	}

	// Start transaction
	tx, err := s.store.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	q := s.store.WithTx(tx)

	// Upsert each row into the database
	for _, row := range rows {
		if err := q.UpsertInvite(ctx, row); err != nil {
			log.Printf("Failed to upsert invite %s: %v", row.InviteCode, err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	log.Printf("Synced %d invites from sheet to database", len(rows))

	return nil
}

// SyncToSheet writes pending RSVP responses back to the sheet
func (s *Syncer) SyncToSheet(ctx context.Context) error {
	// Get invites that need syncing
	invites, err := s.store.GetPendingSyncInvites(ctx)
	if err != nil {
		return err
	}

	if len(invites) == 0 {
		log.Println("No pending RSVPs to sync to sheet")
		return nil
	}

	log.Printf("Syncing %d RSVP responses to sheet", len(invites))

	tx, err := s.store.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	q := s.store.WithTx(tx)

	// Write each invite to the sheet
	for _, invite := range invites {
		if invite.SheetRow == nil {
			log.Printf("Skipping invite %s: no sheet row number", invite.InviteCode)
			continue
		}

		if err := s.sheetsClient.WriteRSVP(ctx, invite); err != nil {
			log.Printf("Failed to write RSVP for invite %s: %v", invite.InviteCode, err)
			continue
		}

		// Mark as synced in database
		if err := q.MarkInviteSynced(ctx, invite.InviteCode); err != nil {
			log.Printf("Failed to mark invite %s as synced: %v", invite.InviteCode, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	log.Printf("Successfully synced %d RSVPs to sheet", len(invites))
	return nil
}

// SyncScheduleFromSheet reads the schedule sheet and replaces all events in DB
// This is a one-way sync: Google Sheets is the source of truth for schedule
// Only public events are returned from ReadScheduleSheet, so we store everything we receive.
func (s *Syncer) SyncScheduleFromSheet(ctx context.Context) error {
	events, err := s.sheetsClient.ReadScheduleSheet(ctx, defaultWeddingYear)
	if err != nil {
		return err
	}

	if events == nil {
		log.Println("Schedule sync skipped (client not configured)")
		return nil
	}

	// Start transaction
	tx, err := s.store.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	q := s.store.WithTx(tx)

	// Delete all existing schedule events (full replace strategy)
	if err := q.DeleteAllScheduleEvents(ctx); err != nil {
		return errors.Wrap(err, "failed to delete existing schedule events")
	}

	// Insert all events from sheet
	for _, event := range events {
		params := &store.InsertScheduleEventParams{
			StartTime:     event.StartTime,
			EndTime:       event.EndTime,
			EventNameEs:   event.EventNameES,
			EventNameEn:   event.EventNameEN,
			EventNameCa:   event.EventNameCA,
			Location:      event.Location,
			DescriptionEs: event.DescriptionES,
			DescriptionEn: event.DescriptionEN,
			DescriptionCa: event.DescriptionCA,
		}

		if err := q.InsertScheduleEvent(ctx, params); err != nil {
			log.Printf("Failed to insert schedule event '%s': %v", event.EventNameES, err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	log.Printf("Synced %d schedule events from sheet to database", len(events))
	return nil
}

// toNullString converts a string to sql.NullString equivalent (empty string for NULL)
func toNullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
