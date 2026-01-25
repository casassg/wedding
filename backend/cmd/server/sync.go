package main

import (
	"context"
	"fmt"
	"log"

	"github.com/casassg/wedding/backend/internal/db"
	"github.com/casassg/wedding/backend/internal/sheets"
)

// SyncCmd forces an immediate sync
type SyncCmd struct {
	DBPath        string `env:"DB_PATH" default:"/litefs/wedding.db" help:"Path to SQLite database file"`
	PrimaryRegion string `env:"PRIMARY_REGION" default:"iad" help:"Primary region for writes"`
	CurrentRegion string `env:"FLY_REGION" default:"iad" help:"Current region (auto-set by Fly.io)"`
}

func (cmd *SyncCmd) Run() error {
	ctx := context.Background()

	log.Printf("Starting manual sync")
	log.Printf("Database: %s", cmd.DBPath)
	log.Printf("Region: %s (primary: %s)", cmd.CurrentRegion, cmd.PrimaryRegion)

	// Initialize database
	database, err := db.New(cmd.DBPath)
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
	syncer := sheets.NewSyncer(database, sheetsClient, cmd.CurrentRegion, cmd.PrimaryRegion)

	if cmd.CurrentRegion != cmd.PrimaryRegion {
		log.Printf("WARNING: Not running in primary region (%s). Sync may be skipped.", cmd.PrimaryRegion)
	}

	log.Printf("Starting sync cycle...")
	if err := syncer.SyncOnce(ctx); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	log.Printf("Sync completed successfully")
	return nil
}
