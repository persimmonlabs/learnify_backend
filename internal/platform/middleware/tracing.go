package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ContextKey type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

// RequestID middleware generates a unique request ID and adds it to context and response headers
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID already exists in headers (from upstream proxy/load balancer)
			requestID := r.Header.Get("X-Request-ID")

			// Generate new request ID if not present
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// Add request ID to response headers
			w.Header().Set("X-Request-ID", requestID)

			// Continue with request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetUserID extracts the user ID from context (set by Auth middleware)
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
