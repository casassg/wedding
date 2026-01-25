package api

import (
	"net/http"
	"strings"

	"github.com/casassg/wedding/backend/internal/db"
)

// NewRouter creates the HTTP router with all routes and middleware
func NewRouter(database *db.DB, allowedOrigins []string) http.Handler {
	handler := NewHandler(database)

	// Create rate limiter (10 requests per minute)
	rateLimiter := NewRateLimiter(10)

	// Create mux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/api/v1/invite/", func(w http.ResponseWriter, r *http.Request) {
		// Route to appropriate handler based on path
		if strings.HasSuffix(r.URL.Path, "/rsvp") {
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			handler.PostRSVP(w, r)
		} else {
			if r.Method != http.MethodGet {
				http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			handler.GetInvite(w, r)
		}
	})

	// Apply middleware chain
	return Chain(
		mux,
		Logging,
		CORS(allowedOrigins),
		rateLimiter.Middleware,
	)
}
