package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
)

// CLI represents the command-line interface
type CLI struct {
	Serve   ServeCmd   `cmd:"" help:"Start the RSVP API server" default:"1"`
	Inspect InspectCmd `cmd:"" help:"Inspect Google Sheets structure and data"`
	Sync    SyncCmd    `cmd:"" help:"Force an immediate sync between database and Google Sheets"`
}

func main() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it (this is okay in production): %v", err)
	}

	cli := &CLI{}
	ctx := kong.Parse(cli,
		kong.Name("server"),
		kong.Description("Wedding RSVP API server and management tools"),
		kong.UsageOnError(),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
