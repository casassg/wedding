package main

import (
	"context"
	"fmt"
	"log"

	"github.com/casassg/wedding/backend/internal/sheets"
	"github.com/casassg/wedding/backend/internal/store"
)

// SyncCmd forces an immediate sync
type SyncCmd struct {
	DBPath string `env:"DB_PATH" default:"wedding.db" help:"Path to SQLite database file"`
}

func (cmd *SyncCmd) Run() error {
	ctx := context.Background()

	log.Printf("Starting manual sync")
	log.Printf("Database: %s", cmd.DBPath)

	// Initialize database
	database, err := store.Open(cmd.DBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	// Initialize Google Sheets client
	sheetsClient, err := sheets.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize sheets client: %w", err)
	}

	// Create syncer and run once
	syncer := sheets.NewSyncer(database, sheetsClient)

	log.Printf("Starting sync cycle...")
	if err := syncer.SyncOnce(ctx); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	log.Printf("Sync completed successfully")
	return nil
}
