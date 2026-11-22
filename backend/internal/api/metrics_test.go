package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRoute_KnownRoutes(t *testing.T) {
	cases := map[string]string{
		"/":                            "/",
		"/health":                      "/health",
		"/api/v1/schedule":             "/api/v1/schedule",
		"/api/v1/invite/ABC123":        "/api/v1/invite/:code",
		"/api/v1/invite/ABC123/":       "/api/v1/invite/:code/",
		"/api/v1/invite/ABC123/rsvp":   "/api/v1/invite/:code/rsvp",
		"/api/v1/invite/ABC123/travel": "/api/v1/invite/:code/travel",
	}
	for path, want := range cases {
		require.Equal(t, want, normalizeRoute(path), "path %q", path)
	}
}

func TestNormalizeRoute_UnknownRoutesAreBucketed(t *testing.T) {
	cases := []string{
		"/does-not-exist",
		"/api/v1/invite/ABC123/extra",
		"/api/v1/unknown",
		"/wp-login.php",
		"/%EF%BF%BD",
	}
	for _, path := range cases {
		require.Equal(t, "/unknown", normalizeRoute(path), "path %q", path)
	}
}

func TestNormalizeRoute_InvalidUTF8DoesNotReachPrometheus(t *testing.T) {
	// Regression test for PURPLE-SNOW-9170-3: a scanner probe for
	// "/%C0" gives r.URL.Path == "/\xc0", which is not valid UTF-8.
	// Prometheus panics on such a label value in the Logging
	// middleware, so normalizeRoute must never return it verbatim.
	for _, path := range []string{"/\xc0", "/\xff\xfe", "/api/v1/invite/\xc0/rsvp"} {
		got := normalizeRoute(path)
		require.True(t, utf8.ValidString(got), "normalizeRoute(%q) = %q, not valid UTF-8", path, got)
		require.Equal(t, "/unknown", got, "path %q", path)
	}
}

func TestLogging_InvalidUTF8PathDoesNotPanic(t *testing.T) {
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// httptest.NewRequest rejects invalid UTF-8 targets, so set the
	// decoded path the server would hand us for a "/%C0" probe.
	req.URL.Path = "/\xc0"
	rec := httptest.NewRecorder()

	require.NotPanics(t, func() {
		handler.ServeHTTP(rec, req)
	})
	require.Equal(t, http.StatusNotFound, rec.Code)
}
