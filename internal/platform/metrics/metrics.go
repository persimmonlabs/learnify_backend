package metrics

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP Metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	// Database Metrics
	dbConnectionsOpen = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_open",
			Help: "Number of open database connections",
		},
	)

	dbConnectionsInUse = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Number of database connections currently in use",
		},
	)

	dbConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query_type"},
	)

	// JWT/Authentication Metrics
	jwtValidationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jwt_validation_total",
			Help: "Total number of JWT validation attempts",
		},
		[]string{"status"},
	)

	// Business Metrics
	userRegistrationsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	userLoginsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins",
		},
	)

	exerciseSubmissionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "exercise_submissions_total",
			Help: "Total number of exercise submissions",
		},
		[]string{"status"},
	)

	aiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_requests_total",
			Help: "Total number of AI requests",
		},
		[]string{"provider", "status"},
	)

	aiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_request_duration_seconds",
			Help:    "AI request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)
)

func init() {
	// Register all metrics with Prometheus
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestSize,
		httpResponseSize,
		dbConnectionsOpen,
		dbConnectionsInUse,
		dbConnectionsIdle,
		dbQueryDuration,
		jwtValidationTotal,
		userRegistrationsTotal,
		userLoginsTotal,
		exerciseSubmissionsTotal,
		aiRequestsTotal,
		aiRequestDuration,
	)
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, endpoint string, status int, duration time.Duration, reqSize, respSize int64) {
	statusStr := strconv.Itoa(status)

	httpRequestsTotal.WithLabelValues(method, endpoint, statusStr).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint, statusStr).Observe(duration.Seconds())

	if reqSize > 0 {
		httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(reqSize))
	}
	if respSize > 0 {
		httpResponseSize.WithLabelValues(method, endpoint).Observe(float64(respSize))
	}
}

// RecordDatabaseQuery records database query metrics
func RecordDatabaseQuery(queryType string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

// UpdateDatabaseMetrics updates database connection pool metrics
func UpdateDatabaseMetrics(db *sql.DB) {
	if db == nil {
		return
	}

	stats := db.Stats()
	dbConnectionsOpen.Set(float64(stats.OpenConnections))
	dbConnectionsInUse.Set(float64(stats.InUse))
	dbConnectionsIdle.Set(float64(stats.Idle))
}

// RecordJWTValidation records JWT validation metrics
func RecordJWTValidation(success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	jwtValidationTotal.WithLabelValues(status).Inc()
}

// RecordUserRegistration records a user registration
func RecordUserRegistration() {
	userRegistrationsTotal.Inc()
}

// RecordUserLogin records a user login
func RecordUserLogin() {
	userLoginsTotal.Inc()
}

// RecordExerciseSubmission records an exercise submission
func RecordExerciseSubmission(success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	exerciseSubmissionsTotal.WithLabelValues(status).Inc()
}

// RecordAIRequest records AI request metrics
func RecordAIRequest(provider string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	aiRequestsTotal.WithLabelValues(provider, status).Inc()
	aiRequestDuration.WithLabelValues(provider).Observe(duration.Seconds())
}

// Handler returns the Prometheus HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartDatabaseMetricsCollector starts a background goroutine to collect database metrics
func StartDatabaseMetricsCollector(db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			UpdateDatabaseMetrics(db)
		}
	}()
}
