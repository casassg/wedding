package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/casassg/wedding/backend/internal/db"
	sqlcdb "github.com/casassg/wedding/backend/internal/db/sqlc"
)

// Handler holds the API dependencies
type Handler struct {
	db *db.DB
}

// NewHandler creates a new API handler
func NewHandler(database *db.DB) *Handler {
	return &Handler{db: database}
}

// GetInvite handles GET /api/v1/invite/{uuid}
func (h *Handler) GetInvite(w http.ResponseWriter, r *http.Request) {
	// Extract UUID from path
	uuid := extractUUID(r.URL.Path)
	if uuid == "" {
		respondError(w, "Invalid invite code", http.StatusBadRequest)
		return
	}

	// Get invite from database
	invite, err := h.db.GetInviteByUUID(uuid)
	if err != nil {
		respondError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if invite == nil {
		respondError(w, "Invite not found", http.StatusNotFound)
		return
	}

	// Return public response
	respondJSON(w, ToInviteResponse(invite), http.StatusOK)
}

// PostRSVP handles POST /api/v1/invite/{uuid}/rsvp
func (h *Handler) PostRSVP(w http.ResponseWriter, r *http.Request) {
	// Extract UUID from path
	uuid := extractUUID(r.URL.Path)
	if uuid == "" {
		respondError(w, "Invalid invite code", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req RSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get invite to validate against
	invite, err := h.db.GetInviteByUUID(uuid)
	if err != nil {
		respondError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if invite == nil {
		respondError(w, "Invite not found", http.StatusNotFound)
		return
	}

	// Validate request
	if err := validateRSVP(req, invite); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get country from Fly.io header
	country := GetCountry(r)

	// Update RSVP in database
	dbReq := db.RSVPRequest{
		Attending:      req.Attending,
		AdultCount:     req.AdultCount,
		KidCount:       req.KidCount,
		DietaryInfo:    req.DietaryInfo,
		TransportNeeds: req.TransportNeeds,
	}
	if err := h.db.UpdateRSVP(uuid, dbReq, country); err != nil {
		respondError(w, "Failed to save RSVP", http.StatusInternalServerError)
		return
	}

	// Return success
	respondJSON(w, RSVPResponse{Success: true}, http.StatusOK)
}

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, HealthResponse{Status: "ok"}, http.StatusOK)
}

// validateRSVP checks if the RSVP request is valid
func validateRSVP(req RSVPRequest, invite *sqlcdb.Invite) error {
	// If not attending, we don't validate counts
	if !req.Attending {
		return nil
	}

	// If attending, adult_count is required
	if req.AdultCount == nil {
		return fmt.Errorf("adult_count is required when attending")
	}

	// Validate adult count
	if *req.AdultCount < 1 {
		return fmt.Errorf("adult_count must be at least 1")
	}

	if *req.AdultCount > int(invite.MaxAdults) {
		return fmt.Errorf("adult_count exceeds maximum allowed (%d)", invite.MaxAdults)
	}

	// If kids are allowed, validate kid count
	if invite.MaxKids > 0 {
		if req.KidCount == nil {
			return fmt.Errorf("kid_count is required when max_kids > 0")
		}

		if *req.KidCount < 0 {
			return fmt.Errorf("kid_count cannot be negative")
		}

		if *req.KidCount > int(invite.MaxKids) {
			return fmt.Errorf("kid_count exceeds maximum allowed (%d)", invite.MaxKids)
		}
	}

	return nil
}

// extractUUID extracts the UUID from the URL path
func extractUUID(path string) string {
	// Handle both /api/v1/invite/{uuid} and /api/v1/invite/{uuid}/rsvp
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "invite" {
		return parts[3]
	}
	return ""
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends a JSON error response
func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, ErrorResponse{Error: message}, status)
}
