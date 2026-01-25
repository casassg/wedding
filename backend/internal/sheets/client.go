package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Client wraps the Google Sheets API client
type Client struct {
	service   *sheets.Service
	sheetID   string
	sheetName string
}

// NewClient creates a new Google Sheets client
func NewClient(ctx context.Context) (*Client, error) {
	sheetID := os.Getenv("GOOGLE_SHEET_ID")
	if sheetID == "" {
		log.Println("Warning: GOOGLE_SHEET_ID not set, sync disabled")
		return &Client{}, nil // Return empty client when not configured
	}

	// Try to get credentials - support both env var formats
	var service *sheets.Service
	var err error

	// Option 1: GOOGLE_APPLICATION_CREDENTIALS (path to file)
	credsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsFile != "" {
		log.Printf("Using Google credentials from file: %s", credsFile)
		service, err = sheets.NewService(ctx, option.WithCredentialsFile(credsFile))
		if err != nil {
			return nil, fmt.Errorf("failed to create sheets service from file: %w", err)
		}
	} else {
		// Option 2: GOOGLE_SHEETS_CREDENTIALS (JSON string)
		credsJSON := os.Getenv("GOOGLE_SHEETS_CREDENTIALS")
		if credsJSON == "" {
			log.Println("Warning: No credentials configured (GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_SHEETS_CREDENTIALS), sync disabled")
			return &Client{}, nil // Return empty client when not configured
		}

		// Parse credentials to validate JSON
		var creds map[string]interface{}
		if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
			return nil, fmt.Errorf("failed to parse credentials JSON: %w", err)
		}

		log.Println("Using Google credentials from GOOGLE_SHEETS_CREDENTIALS env var")
		service, err = sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credsJSON)))
		if err != nil {
			return nil, fmt.Errorf("failed to create sheets service from JSON: %w", err)
		}
	}

	// Get the sheet name (default to "Guests" if not specified)
	sheetName := os.Getenv("GOOGLE_SHEET_NAME")
	if sheetName == "" {
		sheetName = "Guests"
	}

	log.Printf("Google Sheets client initialized for sheet: %s (name: %s)", sheetID, sheetName)

	return &Client{
		service:   service,
		sheetID:   sheetID,
		sheetName: sheetName,
	}, nil
}

// IsConfigured returns whether the client is configured
func (c *Client) IsConfigured() bool {
	return c.service != nil
}

// SheetRow represents a row in the Google Sheet
type SheetRow struct {
	RowNumber   int
	Name        string
	Parella     string // "Si" or "No"
	Fills       int    // Number of kids allowed
	InviteCode  string
	Attending   *bool
	Adults      *int
	Kids        *int
	Dietary     string
	Message     string
	Song        string
	RespondedAt string
	Country     string
}

// ReadSheet reads all invite data from the sheet
func (c *Client) ReadSheet(ctx context.Context) ([]SheetRow, error) {
	if !c.IsConfigured() {
		return nil, nil // Return empty when not configured
	}

	// Read data from 'Guests' sheet (rows 2+, columns A-N)
	// Column mapping:
	// A: Name, B: Parella, C: Fills, D: Location, E: State, F: Total, G: No Hijos
	// H: Invite Code, I: Adults confirmed, J: Kids confirmed, K: Dietary, L: Message for us, M: Song request, N: Updated At
	readRange := fmt.Sprintf("'%s'!A2:N", c.sheetName)
	resp, err := c.service.Spreadsheets.Values.Get(c.sheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}

	var rows []SheetRow
	for i, row := range resp.Values {
		rowNum := i + 2 // Sheet rows start at 1, and we skip header row

		// Parse row data
		sheetRow := SheetRow{RowNumber: rowNum}

		// Column A: Name
		if len(row) > 0 {
			sheetRow.Name = toString(row[0])
		}

		// Column B: Parella (Si/No)
		if len(row) > 1 {
			sheetRow.Parella = toString(row[1])
		}

		// Column C: Fills (kids)
		if len(row) > 2 {
			sheetRow.Fills = toInt(row[2])
		}

		// Column H: Invite Code (index 7)
		if len(row) > 7 {
			sheetRow.InviteCode = toString(row[7])
		}

		// Skip rows without invite code or name
		if sheetRow.InviteCode == "" || sheetRow.Name == "" {
			continue
		}

		rows = append(rows, sheetRow)
	}

	log.Printf("Read %d invites from Google Sheet '%s'", len(rows), c.sheetName)
	return rows, nil
}

// WriteRSVP writes RSVP response data back to the sheet
func (c *Client) WriteRSVP(ctx context.Context, rowNum int, data SheetRow) error {
	if !c.IsConfigured() {
		return nil // No-op when not configured
	}

	// Prepare values for columns I-N (Adults confirmed, Kids confirmed, Dietary, Message for us, Song request, Updated At)
	values := []interface{}{
		intToString(data.Adults), // Column I: Adults confirmed
		intToString(data.Kids),   // Column J: Kids confirmed
		data.Dietary,             // Column K: Dietary
		data.Message,             // Column L: Message for us
		data.Song,                // Column M: Song request
		data.RespondedAt,         // Column N: Updated At
	}

	// Write to sheet
	writeRange := fmt.Sprintf("'%s'!I%d:N%d", c.sheetName, rowNum, rowNum)
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := c.service.Spreadsheets.Values.Update(c.sheetID, writeRange, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("failed to write to sheet: %w", err)
	}

	return nil
}

// Helper functions for type conversion

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	if s, ok := v.(string); ok {
		var i int
		fmt.Sscanf(s, "%d", &i)
		return i
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

func boolToString(b *bool) string {
	if b == nil {
		return ""
	}
	if *b {
		return "TRUE"
	}
	return "FALSE"
}

func intToString(i *int) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}
