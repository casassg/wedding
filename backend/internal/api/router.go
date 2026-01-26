package api

import (
	"net/http"

	"github.com/casassg/wedding/backend/internal/sheets"
	"github.com/casassg/wedding/backend/internal/store"
)

// NewRouter creates the HTTP router with all routes and middleware
func NewRouter(database *store.Store, syncer *sheets.Syncer, allowedOrigins []string) http.Handler {
	handler := NewHandler(database, syncer)

	// Create rate limiter (10 requests per minute)
	rateLimiter := NewRateLimiter(10)

	// Create mux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("GET /api/v1/invite/{invite_code}/", handler.GetInvite)
	mux.HandleFunc("POST /api/v1/invite/{invite_code}/rsvp", handler.PostRSVP)

	// Apply middleware chain
	return Chain(
		mux,
		Logging,
		CORS(allowedOrigins),
		rateLimiter.Middleware,
	)
}
