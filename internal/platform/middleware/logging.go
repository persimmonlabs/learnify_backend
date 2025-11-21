package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logging logs HTTP requests with detailed information
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID if not present
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response headers
			w.Header().Set("X-Request-ID", requestID)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // default status
			}

			// Add request ID to context
			ctx := r.Context()
			ctx = contextWithRequestID(ctx, requestID)
			r = r.WithContext(ctx)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Log request details
			logger.Info("http_request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"status_code", wrapped.statusCode,
				"bytes_written", wrapped.written,
				"duration_ms", duration.Milliseconds(),
				"duration", duration.String(),
			)

			// Log errors for non-2xx status codes
			if wrapped.statusCode >= 400 {
				logger.Warn("http_request_error",
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"status_code", wrapped.statusCode,
					"duration_ms", duration.Milliseconds(),
				)
			}
		})
	}
}

// LoggingSimple provides simple request logging without dependencies
func LoggingSimple() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			w.Header().Set("X-Request-ID", requestID)

			// Wrap response writer
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Add request ID to context
			ctx := contextWithRequestID(r.Context(), requestID)
			r = r.WithContext(ctx)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Simple logging to stdout
			slog.Info("http_request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

// requestIDKey is the key for request ID in context
type requestIDKey struct{}

// contextWithRequestID adds request ID to context
func contextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// GetRequestIDFromContext retrieves request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}
