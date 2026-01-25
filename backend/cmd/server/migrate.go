package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrateCmd struct {
	DBPath        string `help:"Path to the SQLite database" default:"/litefs/wedding.db" env:"DB_PATH"`
	MigrationsDir string `help:"Path to migrations directory" default:"file:///app/migrations" env:"MIGRATIONS_DIR"`
}

func (c *MigrateCmd) Run() error {
	log.Printf("Running migrations on %s from %s", c.DBPath, c.MigrationsDir)

	m, err := migrate.New(
		c.MigrationsDir,
		fmt.Sprintf("sqlite3://%s", c.DBPath),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No migration needed")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
