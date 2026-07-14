package api

import (
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
)

var inviteCodeRe = regexp.MustCompile(`/api/v1/invite/[^/]+`)

// normalizeRoute replaces invite codes in URL paths with a placeholder
// so metrics don't explode in cardinality from brute-force guesses.
func normalizeRoute(path string) string {
	return inviteCodeRe.ReplaceAllString(path, "/api/v1/invite/:code")
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
