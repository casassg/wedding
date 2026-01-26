package api

import (
	"github.com/casassg/wedding/backend/internal/store"
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

// ToInviteResponse converts sqlc Invite to API InviteResponse
func ToInviteResponse(invite *store.Invite) InviteResponse {
	return InviteResponse{
		Name:         invite.Name,
		MaxAdults:    int(invite.MaxAdults),
		MaxKids:      int(invite.MaxKids),
		HasResponded: invite.ConfirmedAdults > 0,
	}
}
