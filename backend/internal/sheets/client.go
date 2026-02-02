package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/casassg/wedding/backend/internal/store"
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

// ReadSheet reads all invite data from the sheet
func (c *Client) ReadSheet(ctx context.Context) ([]*store.UpsertInviteParams, error) {
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

	var rows []*store.UpsertInviteParams
	for i, row := range resp.Values {
		rowNum := int64(i + 2) // Sheet rows start at 1, and we skip header row

		// Parse row data
		sheetRow := store.UpsertInviteParams{SheetRow: &rowNum}

		// Column A: Name
		if len(row) > 0 {
			sheetRow.Name = toString(row[0])
		}

		// Column B: Parella (Si/No)
		// Convert Parella to max_adults
		maxAdults := int64(1)
		if len(row) > 1 {
			if strings.ToLower(toString(row[1])) == "si" {
				maxAdults = 2
			}
		}
		sheetRow.MaxAdults = maxAdults

		// Column C: Fills (kids)
		if len(row) > 2 {
			sheetRow.MaxKids = toInt(row[2])
		}

		// Column H: Invite Code (index 7)
		if len(row) > 7 {
			sheetRow.InviteCode = toString(row[7])
		}

		// Column I: Adults confirmed (index 8)
		if len(row) > 8 {
			sheetRow.ConfirmedAdults = toInt(row[8])
		}

		// Skip rows without invite code or name
		if sheetRow.InviteCode == "" || sheetRow.Name == "" {
			continue
		}

		rows = append(rows, &sheetRow)
	}

	log.Printf("Read %d invites from Google Sheet '%s'", len(rows), c.sheetName)
	return rows, nil
}

// WriteRSVP writes RSVP response data back to the sheet
func (c *Client) WriteRSVP(ctx context.Context, data *store.Invite) error {
	if !c.IsConfigured() {
		return nil // No-op when not configured
	}

	if data.SheetRow == nil {
		return fmt.Errorf("no sheet row number for invite %s", data.InviteCode)
	}

	rowNum := *data.SheetRow

	responseAt := time.Now().UTC()
	if data.ResponseAt != nil {
		responseAt = *data.ResponseAt
	}

	// Prepare values for columns I-N (Adults confirmed, Kids confirmed, Dietary, Message for us, Song request, Updated At)
	values := []interface{}{
		data.ConfirmedAdults, // Column I: Adults confirmed
		data.ConfirmedKids,   // Column J: Kids confirmed
		data.DietaryInfo,     // Column K: Dietary
		data.MessageForUs,    // Column L: Message for us
		data.SongRequest,     // Column M: Song request
		responseAt,           // Column O: Response At
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

func toInt(v interface{}) int64 {
	if v == nil {
		return 0
	}
	if s, ok := v.(string); ok {
		var i int64
		fmt.Sscanf(s, "%d", &i)
		return i
	}
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return 0
}

// ScheduleEventRow represents a schedule event parsed from Google Sheets
// Supports multilingual event names and descriptions (ES=default, EN, CA)
type ScheduleEventRow struct {
	StartTime     string  // ISO8601 format: "2026-12-19T16:00:00-06:00"
	EndTime       *string // ISO8601 format (nullable)
	EventNameES   string  // Spanish (from "Evento" column D)
	EventNameEN   string  // English (from column H)
	EventNameCA   string  // Catalan (from column I)
	Location      string  // Location (column F)
	DescriptionES string  // Spanish (from "Description" column G)
	DescriptionEN string  // English (from column J)
	DescriptionCA string  // Catalan (from column K)
}

// ReadScheduleSheet reads schedule events from the "Schedule" sheet
// Only returns public events (filtered here before returning).
// Column mapping (based on user's sheet):
// A: Start Time, B: End Time, C: Public (checkbox), D: Evento (Spanish name)
// E: Team/Person, F: Location, G: Description (Spanish)
// H: Event name (English), I: Nombre catalan, J: Descripcion English, K: Descripcion Catalan
func (c *Client) ReadScheduleSheet(ctx context.Context, weddingYear int) ([]*ScheduleEventRow, error) {
	if !c.IsConfigured() {
		return nil, nil // Return empty when not configured
	}

	// Read data from 'Schedule' sheet (rows 2+, columns A-K)
	readRange := "'Schedule'!A2:K"
	resp, err := c.service.Spreadsheets.Values.Get(c.sheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read schedule sheet: %w", err)
	}

	var events []*ScheduleEventRow
	var currentDate string // Track current date from day header rows (ISO format)

	// Regex to match day header like "Friday Dec 18" or "Saturday Dec 19"
	dayHeaderRegex := regexp.MustCompile(`(?i)^(monday|tuesday|wednesday|thursday|friday|saturday|sunday)\s+(\w+)\s+(\d+)$`)

	// Month name to number mapping
	monthMap := map[string]int{
		"jan": 1, "january": 1,
		"feb": 2, "february": 2,
		"mar": 3, "march": 3,
		"apr": 4, "april": 4,
		"may": 5,
		"jun": 6, "june": 6,
		"jul": 7, "july": 7,
		"aug": 8, "august": 8,
		"sep": 9, "september": 9,
		"oct": 10, "october": 10,
		"nov": 11, "november": 11,
		"dec": 12, "december": 12,
	}

	// Copan timezone (UTC-6)
	copanLoc := time.FixedZone("America/Tegucigalpa", -6*60*60)

	for _, row := range resp.Values {
		// Column A: Start Time
		startTimeRaw := ""
		if len(row) > 0 {
			startTimeRaw = strings.TrimSpace(toString(row[0]))
		}

		// Column B: End Time
		endTimeRaw := ""
		if len(row) > 1 {
			endTimeRaw = strings.TrimSpace(toString(row[1]))
		}

		// Column C: Public (TRUE/FALSE checkbox)
		isPublic := false
		if len(row) > 2 {
			publicVal := strings.ToUpper(strings.TrimSpace(toString(row[2])))
			isPublic = publicVal == "TRUE"
		}

		// Column D: Evento (Spanish event name - default)
		eventNameES := ""
		if len(row) > 3 {
			eventNameES = strings.TrimSpace(toString(row[3]))
		}

		// Column E: Team/Person (skip - internal use only)

		// Column F: Location
		location := ""
		if len(row) > 5 {
			location = strings.TrimSpace(toString(row[5]))
		}

		// Column G: Description (Spanish - default)
		descriptionES := ""
		if len(row) > 6 {
			descriptionES = strings.TrimSpace(toString(row[6]))
		}

		// Column H: Event name (English)
		eventNameEN := ""
		if len(row) > 7 {
			eventNameEN = strings.TrimSpace(toString(row[7]))
		}

		// Column I: Nombre catalan
		eventNameCA := ""
		if len(row) > 8 {
			eventNameCA = strings.TrimSpace(toString(row[8]))
		}

		// Column J: Descripcion English
		descriptionEN := ""
		if len(row) > 9 {
			descriptionEN = strings.TrimSpace(toString(row[9]))
		}

		// Column K: Descripcion Catalan
		descriptionCA := ""
		if len(row) > 10 {
			descriptionCA = strings.TrimSpace(toString(row[10]))
		}

		// Check if this is a day header row (empty times, event name matches day pattern)
		if startTimeRaw == "" && endTimeRaw == "" && eventNameES != "" {
			matches := dayHeaderRegex.FindStringSubmatch(eventNameES)
			if matches != nil {
				// Parse month and day from "Friday Dec 18"
				monthStr := strings.ToLower(matches[2])
				dayStr := matches[3]

				if monthNum, ok := monthMap[monthStr]; ok {
					day, err := strconv.Atoi(dayStr)
					if err == nil {
						// Format as ISO date with the wedding year
						currentDate = fmt.Sprintf("%d-%02d-%02d", weddingYear, monthNum, day)
						log.Printf("Schedule: Found day header '%s' -> date %s", eventNameES, currentDate)
					}
				}
				continue // Skip day header rows, don't add as events
			}
		}

		// Skip rows without event name, without a current date context, or non-public
		if eventNameES == "" || currentDate == "" || !isPublic {
			continue
		}

		// Skip rows without start time (can't create a valid datetime)
		if startTimeRaw == "" {
			continue
		}

		// Parse times from "8:00 PM" format to 24h "20:00" format
		startTime24 := parseTimeTo24h(startTimeRaw)
		endTime24 := parseTimeTo24h(endTimeRaw)

		// Build full datetime from date + time in Copan timezone
		startDateTime, err := parseDateTime(currentDate, startTime24, copanLoc)
		if err != nil {
			log.Printf("Schedule: Skipping event '%s' - invalid start time: %v", eventNameES, err)
			continue
		}

		// Format as ISO8601 string
		startTimeISO := startDateTime.Format(time.RFC3339)

		var endTimeISO *string
		if endTime24 != "" {
			endDT, err := parseDateTime(currentDate, endTime24, copanLoc)
			if err == nil {
				s := endDT.Format(time.RFC3339)
				endTimeISO = &s
			}
		}

		event := &ScheduleEventRow{
			StartTime:     startTimeISO,
			EndTime:       endTimeISO,
			EventNameES:   eventNameES,
			EventNameEN:   eventNameEN,
			EventNameCA:   eventNameCA,
			Location:      location,
			DescriptionES: descriptionES,
			DescriptionEN: descriptionEN,
			DescriptionCA: descriptionCA,
		}

		events = append(events, event)
	}

	log.Printf("Read %d public schedule events from Google Sheet 'Schedule'", len(events))
	return events, nil
}

// parseDateTime combines a date string and time string into a time.Time in the given location
func parseDateTime(dateStr, timeStr string, loc *time.Location) (time.Time, error) {
	// dateStr is "2026-12-19", timeStr is "16:00"
	combined := dateStr + "T" + timeStr + ":00"
	return time.ParseInLocation("2006-01-02T15:04:05", combined, loc)
}

// parseTimeTo24h converts "8:00 PM" or "4:00 PM" to "20:00" or "16:00"
func parseTimeTo24h(timeStr string) string {
	if timeStr == "" {
		return ""
	}

	timeStr = strings.TrimSpace(strings.ToUpper(timeStr))

	// Match patterns like "8:00 PM", "4:00 AM", "11:30 PM"
	timeRegex := regexp.MustCompile(`^(\d{1,2}):(\d{2})\s*(AM|PM)?$`)
	matches := timeRegex.FindStringSubmatch(timeStr)
	if matches == nil {
		// Already in 24h format or unrecognized
		return timeStr
	}

	hour, _ := strconv.Atoi(matches[1])
	minute := matches[2]
	period := matches[3]

	if period == "PM" && hour != 12 {
		hour += 12
	} else if period == "AM" && hour == 12 {
		hour = 0
	}

	return fmt.Sprintf("%02d:%s", hour, minute)
}
