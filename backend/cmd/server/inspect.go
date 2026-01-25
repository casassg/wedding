package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	googsheets "google.golang.org/api/sheets/v4"
)

// InspectCmd inspects Google Sheets structure
type InspectCmd struct{}

func (cmd *InspectCmd) Run() error {
	ctx := context.Background()

	// Get credentials
	credsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	credsJSON := os.Getenv("GOOGLE_SHEETS_CREDENTIALS")

	if credsFile == "" && credsJSON == "" {
		return fmt.Errorf("either GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_SHEETS_CREDENTIALS must be set")
	}

	sheetID := os.Getenv("GOOGLE_SHEET_ID")
	if sheetID == "" {
		return fmt.Errorf("GOOGLE_SHEET_ID not set")
	}

	// Create service
	var service *googsheets.Service
	var err error

	if credsJSON != "" {
		log.Printf("Using credentials from GOOGLE_SHEETS_CREDENTIALS env var")
		service, err = googsheets.NewService(ctx, option.WithCredentialsJSON([]byte(credsJSON)))
	} else {
		log.Printf("Using credentials from file: %s", credsFile)
		service, err = googsheets.NewService(ctx, option.WithCredentialsFile(credsFile))
	}

	if err != nil {
		return fmt.Errorf("failed to create sheets service: %w", err)
	}

	log.Printf("Connected to Google Sheets API")
	log.Printf("Sheet ID: %s", sheetID)

	// Get spreadsheet metadata to list all sheets
	spreadsheet, err := service.Spreadsheets.Get(sheetID).Do()
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	fmt.Println("\n=== Spreadsheet Info ===")
	fmt.Printf("Title: %s\n", spreadsheet.Properties.Title)
	fmt.Printf("\nAvailable Sheets:\n")
	for i, sheet := range spreadsheet.Sheets {
		fmt.Printf("%d. %s (gid: %d)\n", i+1, sheet.Properties.Title, sheet.Properties.SheetId)
	}

	// Try to find the sheet with gid 1815613445
	var targetSheet *googsheets.Sheet
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.SheetId == 1815613445 {
			targetSheet = sheet
			break
		}
	}

	if targetSheet == nil {
		return fmt.Errorf("could not find sheet with gid 1815613445")
	}

	sheetName := targetSheet.Properties.Title
	fmt.Printf("\n=== Target Sheet: %s ===\n", sheetName)

	// Read first 5 rows to see structure
	readRange := fmt.Sprintf("'%s'!A1:M5", sheetName)
	resp, err := service.Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}

	fmt.Printf("\nFirst 5 rows from '%s':\n\n", sheetName)
	for i, row := range resp.Values {
		rowNum := i + 1
		fmt.Printf("Row %d: ", rowNum)
		rowJSON, _ := json.MarshalIndent(row, "", "  ")
		fmt.Printf("%s\n\n", rowJSON)
	}

	// Count total rows
	countRange := fmt.Sprintf("'%s'!A:A", sheetName)
	countResp, err := service.Spreadsheets.Values.Get(sheetID, countRange).Do()
	if err != nil {
		return fmt.Errorf("failed to count rows: %w", err)
	}

	fmt.Printf("Total rows in sheet: %d\n", len(countResp.Values))

	// Read all data to analyze
	dataRange := fmt.Sprintf("'%s'!A2:M", sheetName)
	dataResp, err := service.Spreadsheets.Values.Get(sheetID, dataRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	fmt.Printf("\n=== Data Analysis ===\n")
	fmt.Printf("Total data rows (excluding header): %d\n", len(dataResp.Values))

	// Count rows with invite codes
	inviteCodeCount := 0
	for _, row := range dataResp.Values {
		// Column H (index 7) should have invite code
		if len(row) > 7 && row[7] != nil && row[7] != "" {
			inviteCodeCount++
		}
	}
	fmt.Printf("Rows with invite codes (column H): %d\n", inviteCodeCount)

	// Show a few sample invite codes
	fmt.Printf("\nSample invite codes:\n")
	count := 0
	for i, row := range dataResp.Values {
		if len(row) > 7 && row[7] != nil && row[7] != "" {
			fmt.Printf("  Row %d: %v (Name: %v)\n", i+2, row[7], row[0])
			count++
			if count >= 5 {
				break
			}
		}
	}

	return nil
}
