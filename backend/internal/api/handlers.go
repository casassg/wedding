package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

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

// GetInvite handles GET /api/v1/invite/{invite_code}
func (h *Handler) GetInvite(w http.ResponseWriter, r *http.Request) {
	// Extract UUID from path
	inviteCode := r.PathValue("invite_code")
	if inviteCode == "" {
		respondError(w, "Invalid invite code", http.StatusBadRequest)
		return
	}

	// Get invite from database
	invite, err := h.db.GetInviteByInviteCode(r.Context(), inviteCode)
	if errors.Is(err, sql.ErrNoRows) || (invite == nil) {
		log.Printf("Invite not found for code %s, triggering sync", inviteCode)
		if err := h.syncer.SyncOnce(r.Context()); err != nil {
			inviteLookups.WithLabelValues("not_found").Inc()
			respondError(w, "Invite not found", http.StatusNotFound)
			return
		}
		// Retry fetching invite after sync
		invite, err = h.db.GetInviteByInviteCode(r.Context(), inviteCode)
		if invite == nil || errors.Is(err, sql.ErrNoRows) {
			ip := getIP(r)
			log.Printf("Invite not found for code %s from IP=%s UA=%s", inviteCode, ip, r.UserAgent())
			inviteLookups.WithLabelValues("not_found").Inc()
			respondError(w, "Invite not found", http.StatusNotFound)
			return
		}
	}
	if err != nil {
		log.Printf("Error fetching invite: %v", err)
		inviteLookups.WithLabelValues("error").Inc()
		respondError(w, "Invite not found", http.StatusNotFound)
		return
	}

	ip := getIP(r)
	log.Printf("Invite viewed: %s (%s) from IP=%s UA=%s", invite.Name, invite.InviteCode, ip, r.UserAgent())
	inviteLookups.WithLabelValues("found").Inc()
	inviteViews.WithLabelValues(invite.InviteCode).Inc()

	// Return public response
	respondJSON(w, ToInviteResponse(invite), http.StatusOK)
}

// PostRSVP handles POST /api/v1/invite/{invite_code}/rsvp
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
	}

	if err := h.db.UpdateRSVP(r.Context(), &dbReq); err != nil {
		rsvpSubmissions.WithLabelValues("error").Inc()
		respondError(w, "Failed to save RSVP", http.StatusInternalServerError)
		return
	}

	// Async update to Google Sheets
	h.syncer.TriggerSync()

	// Return success
	rsvpSubmissions.WithLabelValues("success").Inc()
	respondJSON(w, RSVPResponse{Success: true}, http.StatusOK)
}

// PostTravel handles POST /api/v1/invite/{invite_code}/travel
func (h *Handler) PostTravel(w http.ResponseWriter, r *http.Request) {
	inviteCode := r.PathValue("invite_code")
	if inviteCode == "" {
		respondError(w, "Invalid invite code", http.StatusBadRequest)
		return
	}

	var req TravelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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
	if invite.ConfirmedAdults == 0 {
		respondError(w, "Travel info is only available after confirming attendance", http.StatusBadRequest)
		return
	}

	if err := validateTravel(req); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Normalize dependent fields for guests from Honduras (no bus/flight questions).
	inHonduras := strings.EqualFold(invite.Location, "HONDURAS")

	busTo := req.BusTo
	pickup := req.Pickup
	arrivalFlight := req.ArrivalFlight
	busReturn := req.BusReturn
	hotel := req.Hotel
	notes := req.Notes

	if inHonduras {
		busTo = ""
		pickup = ""
		arrivalFlight = ""
		busReturn = ""
	}

	// Normalize dependent fields: pickup and flight only relevant when taking bus.
	if busTo == "" || busTo == "none" {
		pickup = ""
		arrivalFlight = ""
	}
	// Arrival flight only relevant when pickup is sap.
	if pickup != "sap" {
		arrivalFlight = ""
	}
	// Bus return sub-options: no cleanup needed (san_pedro vs sap are both valid).

	dbReq := store.UpdateTravelInfoParams{
		InputBusTo:         busTo,
		InputPickup:        pickup,
		InputArrivalFlight: arrivalFlight,
		InputBusReturn:     busReturn,
		InputHotel:         hotel,
		InputNotes:         notes,
		InputInviteCode:    inviteCode,
	}

	if err := h.db.UpdateTravelInfo(r.Context(), &dbReq); err != nil {
		log.Printf("Error updating travel info for %s: %v", inviteCode, err)
		travelSubmissions.WithLabelValues("error").Inc()
		respondError(w, "Failed to save travel info", http.StatusInternalServerError)
		return
	}

	h.syncer.TriggerSync()

	travelSubmissions.WithLabelValues("success").Inc()
	respondJSON(w, RSVPResponse{Success: true}, http.StatusOK)
}

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, HealthResponse{Status: "ok"}, http.StatusOK)
}

// GetSchedule handles GET /api/v1/schedule
// Returns all public schedule events with timezone info
func (h *Handler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	events, err := h.db.GetScheduleEvents(r.Context())
	if err != nil {
		log.Printf("Error fetching schedule events: %v", err)
		respondError(w, "Failed to fetch schedule", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	eventResponses := make([]ScheduleEventResponse, 0, len(events))
	for _, event := range events {
		eventResponses = append(eventResponses, ToScheduleEventResponse(event))
	}

	// Return response with timezone info (Copan is UTC-6, no DST)
	response := ScheduleResponse{
		Timezone:       "America/Tegucigalpa",
		TimezoneOffset: "-06:00",
		Events:         eventResponses,
	}

	respondJSON(w, response, http.StatusOK)
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

const maxTextLen = 500

// validateTravel validates travel request enums and text field lengths.
func validateTravel(req TravelRequest) error {
	validBusTo := map[string]bool{"": true, "thursday": true, "friday": true, "none": true}
	if !validBusTo[req.BusTo] {
		return fmt.Errorf("invalid bus_to value %q", req.BusTo)
	}

	validPickup := map[string]bool{"": true, "sap": true, "welchez": true}
	if !validPickup[req.Pickup] {
		return fmt.Errorf("invalid pickup value %q", req.Pickup)
	}

	validBusReturn := map[string]bool{
		"": true, "sunday_san_pedro": true, "sunday_sap": true,
		"monday_san_pedro": true, "monday_sap": true, "none": true,
	}
	if !validBusReturn[req.BusReturn] {
		return fmt.Errorf("invalid bus_return value %q", req.BusReturn)
	}

	if utf8.RuneCountInString(req.ArrivalFlight) > maxTextLen {
		return fmt.Errorf("arrival_flight exceeds %d characters", maxTextLen)
	}
	if utf8.RuneCountInString(req.Hotel) > maxTextLen {
		return fmt.Errorf("hotel exceeds %d characters", maxTextLen)
	}
	if utf8.RuneCountInString(req.Notes) > maxTextLen {
		return fmt.Errorf("notes exceeds %d characters", maxTextLen)
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
