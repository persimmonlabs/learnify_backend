package middleware

import (
	"net/http"
	"time"

	"backend/internal/platform/metrics"
)

// metricsResponseWriter wraps http.ResponseWriter to capture metrics (distinct from responseWriter in logging.go)
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// Metrics middleware records HTTP metrics for Prometheus
func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture metrics
			rw := &metricsResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // default status
				bytesWritten:   0,
			}

			// Get request size
			reqSize := r.ContentLength
			if reqSize < 0 {
				reqSize = 0
			}

			// Process request
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start)

			// Record metrics
			metrics.RecordHTTPRequest(
				r.Method,
				r.URL.Path,
				rw.statusCode,
				duration,
				reqSize,
				rw.bytesWritten,
			)
		})
	}
}
