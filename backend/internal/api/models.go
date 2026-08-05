package api

import (
	"strings"

	"github.com/casassg/wedding/backend/internal/store"
)

// InviteResponse is the public API response for GET /invite/{uuid}
type InviteResponse struct {
	Name         string `json:"name"`
	MaxAdults    int    `json:"max_adults"`
	MaxKids      int    `json:"max_kids"`
	HasResponded bool   `json:"has_responded"`
	IsAttending  bool   `json:"is_attending"`

	// Travel fields
	HasTravelInfo       bool   `json:"has_travel_info"`
	InHonduras          bool   `json:"in_honduras"`
	TravelBusTo         string `json:"travel_bus_to"`
	TravelPickup        string `json:"travel_pickup"`
	TravelArrivalFlight string `json:"travel_arrival_flight"`
	TravelBusReturn     string `json:"travel_bus_return"`
	TravelHotel         string `json:"travel_hotel"`
	TravelNotes         string `json:"travel_notes"`
	TravelCocktail      string `json:"travel_cocktail"`
	TravelBrunch        string `json:"travel_brunch"`
	TravelReturnDetail  string `json:"travel_return_detail"`
}

// TravelRequest is the request payload for POST /invite/{code}/travel
type TravelRequest struct {
	BusTo         string `json:"bus_to"`
	Pickup        string `json:"pickup"`
	ArrivalFlight string `json:"arrival_flight"`
	BusReturn     string `json:"bus_return"`
	Hotel         string `json:"hotel"`
	Notes         string `json:"notes"`
	Cocktail      string `json:"cocktail"`
	Brunch        string `json:"brunch"`
	ReturnDetail  string `json:"return_detail"`
}

// RSVPRequest is the request payload for POST /invite/{uuid}/rsvp
type RSVPRequest struct {
	AdultCount   int64  `json:"adult_count,omitempty"`
	KidCount     int64  `json:"kid_count,omitempty"`
	DietaryInfo  string `json:"dietary_info,omitempty"`
	MessageForUs string `json:"message_for_us,omitempty"`
	SongRequest  string `json:"song_request,omitempty"`
}

// RSVPResponse is the success response for POST /invite/{uuid}/rsvp
type RSVPResponse struct {
	Success bool `json:"success"`
}

// ErrorResponse is returned for API errors
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse is returned by /health
type HealthResponse struct {
	Status string `json:"status"`
}

// ScheduleEventResponse is a single event in the schedule
// Returns all language variants so the frontend can pick the right one
type ScheduleEventResponse struct {
	StartTime   string                `json:"start_time"`  // ISO8601 format: "2026-12-19T16:00:00-06:00"
	EndTime     string                `json:"end_time"`    // ISO8601 format or empty
	Name        ScheduleEventI18nText `json:"name"`        // Event name in all languages
	Location    string                `json:"location"`    // Event location
	Description ScheduleEventI18nText `json:"description"` // Event description in all languages
}

// ScheduleEventI18nText holds text in all supported languages
type ScheduleEventI18nText struct {
	ES string `json:"es"` // Spanish (default)
	EN string `json:"en"` // English
	CA string `json:"ca"` // Catalan
}

// ScheduleResponse is returned by GET /api/v1/schedule
type ScheduleResponse struct {
	Timezone       string                  `json:"timezone"`        // IANA timezone: "America/Tegucigalpa"
	TimezoneOffset string                  `json:"timezone_offset"` // UTC offset: "-06:00"
	Events         []ScheduleEventResponse `json:"events"`
}

// ToScheduleEventResponse converts a store.ScheduleEvent to API response
func ToScheduleEventResponse(event *store.ScheduleEvent) ScheduleEventResponse {
	endTime := ""
	if event.EndTime != nil {
		endTime = *event.EndTime
	}

	return ScheduleEventResponse{
		StartTime: event.StartTime,
		EndTime:   endTime,
		Name: ScheduleEventI18nText{
			ES: event.EventNameEs,
			EN: event.EventNameEn,
			CA: event.EventNameCa,
		},
		Location: event.Location,
		Description: ScheduleEventI18nText{
			ES: event.DescriptionEs,
			EN: event.DescriptionEn,
			CA: event.DescriptionCa,
		},
	}
}

// ToInviteResponse converts sqlc Invite to API InviteResponse
func ToInviteResponse(invite *store.Invite) InviteResponse {
	return InviteResponse{
		Name:                invite.Name,
		MaxAdults:           int(invite.MaxAdults),
		MaxKids:             int(invite.MaxKids),
		HasResponded:        invite.ResponseAt != nil,
		IsAttending:         invite.ConfirmedAdults > 0,
		HasTravelInfo:       invite.TravelUpdatedAt != nil,
		InHonduras:          strings.EqualFold(invite.Location, "HONDURAS"),
		TravelBusTo:         invite.TravelBusTo,
		TravelPickup:        invite.TravelPickup,
		TravelArrivalFlight: invite.TravelArrivalFlight,
		TravelBusReturn:     invite.TravelBusReturn,
		TravelHotel:         invite.TravelHotel,
		TravelNotes:         invite.TravelNotes,
		TravelCocktail:      invite.TravelCocktail,
		TravelBrunch:        invite.TravelBrunch,
		TravelReturnDetail:  invite.TravelReturnDetail,
	}
}
