package api

import (
	"regexp"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
)

var inviteCodeRe = regexp.MustCompile(`/api/v1/invite/[^/]+`)

// normalizeRoute maps a request path to a bounded set of metric label
// values. It replaces invite codes with a placeholder and buckets every
// other path that is not a known route as "/unknown", so metrics don't
// explode in cardinality from brute-force guesses or scanner probes.
//
// Unknown paths must never be used as a label value verbatim: a probe
// such as "/%C0" decodes to invalid UTF-8 ("/\xc0"), and Prometheus
// panics on label values that are not valid UTF-8. That panic in the
// Logging middleware is what Sentry reported as PURPLE-SNOW-9170-3.
func normalizeRoute(path string) string {
	if !utf8.ValidString(path) {
		return "/unknown"
	}
	path = inviteCodeRe.ReplaceAllString(path, "/api/v1/invite/:code")
	switch path {
	case "/",
		"/health",
		"/api/v1/schedule",
		"/api/v1/invite/:code",
		"/api/v1/invite/:code/",
		"/api/v1/invite/:code/rsvp",
		"/api/v1/invite/:code/travel":
		return path
	default:
		return "/unknown"
	}
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wedding_http_requests_total",
		Help: "Total HTTP requests.",
	}, []string{"method", "route", "status"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "wedding_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})

	inviteLookups = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wedding_invite_lookups_total",
		Help: "Invite lookups by result.",
	}, []string{"result"})

	inviteViews = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wedding_invite_views_total",
		Help: "Successful invite views by code.",
	}, []string{"invite_code"})

	rsvpSubmissions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wedding_rsvp_submissions_total",
		Help: "RSVP submissions by result.",
	}, []string{"result"})

	travelSubmissions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "wedding_travel_submissions_total",
		Help: "Travel form submissions by result.",
	}, []string{"result"})

	rateLimitRejections = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wedding_rate_limit_rejections_total",
		Help: "Requests rejected by rate limiter.",
	})
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		inviteLookups,
		inviteViews,
		rsvpSubmissions,
		travelSubmissions,
		rateLimitRejections,
	)
}
