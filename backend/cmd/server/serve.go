package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/casassg/wedding/backend/internal/api"
	"github.com/casassg/wedding/backend/internal/db"
	"github.com/casassg/wedding/backend/internal/sheets"
)

const shutdownTimeout = 5 * time.Second

// ServeCmd runs the HTTP server
type ServeCmd struct {
	DBPath         string `env:"DB_PATH" default:"/litefs/wedding.db" help:"Path to SQLite database file"`
	Port           string `env:"PORT" default:"8081" help:"Port to listen on"`
	AllowedOrigins string `env:"ALLOWED_ORIGINS" default:"https://lauraygerard.wedding,https://www.lauraygerard.wedding" help:"Comma-separated list of allowed CORS origins"`
	SyncInterval   string `env:"SHEETS_SYNC_INTERVAL" default:"5m" help:"Interval between Google Sheets syncs"`
	PrimaryRegion  string `env:"PRIMARY_REGION" default:"iad" help:"Primary region for writes"`
	CurrentRegion  string `env:"FLY_REGION" default:"iad" help:"Current region (auto-set by Fly.io)"`
}

func (cmd *ServeCmd) Run() error {
	// Create context that listens for SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Parse allowed origins
	allowedOrigins := strings.Split(cmd.AllowedOrigins, ",")
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}

	// Parse sync interval
	interval, err := time.ParseDuration(cmd.SyncInterval)
	if err != nil {
		return fmt.Errorf("invalid SHEETS_SYNC_INTERVAL: %w", err)
	}

	log.Printf("Starting Wedding RSVP API")
	log.Printf("Database: %s", cmd.DBPath)
	log.Printf("Port: %s", cmd.Port)
	log.Printf("Region: %s (primary: %s)", cmd.CurrentRegion, cmd.PrimaryRegion)
	log.Printf("Allowed origins: %v", allowedOrigins)
	log.Printf("Sync interval: %s", interval)

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

	// Start background sync (only on primary region)
	syncer := sheets.NewSyncer(database, sheetsClient, cmd.CurrentRegion, cmd.PrimaryRegion)
	go syncer.Start(ctx, interval)

	// Create HTTP router
	router := api.NewRouter(database, allowedOrigins)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cmd.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server listening on %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case <-ctx.Done():
		log.Printf("Received shutdown signal, starting graceful shutdown")

		// Give outstanding requests time to complete
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		log.Println("Server stopped gracefully")
	}

	return nil
}
