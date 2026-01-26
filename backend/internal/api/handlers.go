package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/casassg/wedding/backend/internal/sheets"
	"github.com/casassg/wedding/backend/internal/store"
	"github.com/pkg/errors"
)

// Handler holds the API dependencies
type Handler struct {
	db     *store.Store
	syncer *sheets.Syncer
}

// NewHandler creates a new API handler
func NewHandler(database *store.Store, syncer *sheets.Syncer) *Handler {
	return &Handler{db: database, syncer: syncer}
}

// GetInvite handles GET /api/v1/invite/{uuid}
func (h *Handler) GetInvite(w http.ResponseWriter, r *http.Request) {
	// Extract UUID from path
	inviteCode := r.PathValue("invite_code")
	if inviteCode == "" {
		respondError(w, "Invalid invite code", http.StatusBadRequest)
		return
	}

	// Get invite from database
	invite, err := h.db.GetInviteByInviteCode(r.Context(), inviteCode)
	if errors.Is(err, sql.ErrNoRows) {
		respondError(w, "Invite not found", http.StatusNotFound)
		return
	}
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
	inviteCode := r.PathValue("invite_code")
	if inviteCode == "" {
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
	invite, err := h.db.GetInviteByInviteCode(r.Context(), inviteCode)
	if errors.Is(err, sql.ErrNoRows) {
		respondError(w, "Invite not found", http.StatusNotFound)
		return
	}
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

	// Update RSVP in database
	dbReq := store.UpdateRSVPParams{
		InputConfirmedAdults: req.AdultCount,
		InputConfirmedKids:   req.KidCount,
		InputDietaryInfo:     req.DietaryInfo,
		InputMessage:         req.MessageForUs,
		InputSong:            req.SongRequest,
		InputInviteCode:      inviteCode,
		InputResponseCountry: GetCountry(r),
	}

	if err := h.db.UpdateRSVP(r.Context(), &dbReq); err != nil {
		respondError(w, "Failed to save RSVP", http.StatusInternalServerError)
		return
	}

	// Async update to Google Sheets
	h.syncer.TriggerSync()

	// Return success
	respondJSON(w, RSVPResponse{Success: true}, http.StatusOK)
}

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, HealthResponse{Status: "ok"}, http.StatusOK)
}

// validateRSVP checks if the RSVP request is valid
func validateRSVP(req RSVPRequest, invite *store.Invite) error {
	// If attending, adult_count is required
	if req.AdultCount < 0 || req.AdultCount > invite.MaxAdults {
		return fmt.Errorf("adult_count not valid, must be between 0 and %d", invite.MaxAdults)
	}

	if req.KidCount < 0 || req.KidCount > invite.MaxKids {
		return fmt.Errorf("kid_count not valid, must be between 0 and %d", invite.MaxKids)
	}

	return nil
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
