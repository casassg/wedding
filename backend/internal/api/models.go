package api

import (
	"database/sql"

	sqlcdb "github.com/casassg/wedding/backend/internal/db/sqlc"
)

// InviteResponse is the public API response for GET /invite/{uuid}
type InviteResponse struct {
	Name         string `json:"name"`
	MaxAdults    int    `json:"max_adults"`
	MaxKids      int    `json:"max_kids"`
	HasResponded bool   `json:"has_responded"`
}

// RSVPRequest is the request payload for POST /invite/{uuid}/rsvp
type RSVPRequest struct {
	Attending    bool   `json:"attending"`
	AdultCount   *int   `json:"adult_count,omitempty"`
	KidCount     *int   `json:"kid_count,omitempty"`
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

// ToInviteResponse converts sqlc Invite to API InviteResponse
func ToInviteResponse(invite *sqlcdb.Invite) InviteResponse {
	return InviteResponse{
		Name:         invite.Name,
		MaxAdults:    int(invite.MaxAdults),
		MaxKids:      int(invite.MaxKids),
		HasResponded: invite.ResponseAt.Valid,
	}
}

// Helper functions for converting sql.Null* types

func NullBoolToPtr(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	b := nb.Bool
	return &b
}

func NullIntToPtr(ni sql.NullInt64) *int {
	if !ni.Valid {
		return nil
	}
	i := int(ni.Int64)
	return &i
}

func NullStringToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
