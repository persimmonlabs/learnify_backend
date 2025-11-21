package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery is middleware that recovers from panics and returns a 500 error
// This prevents a single panic from crashing the entire server
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get request ID from context if available
					requestID := GetRequestIDFromContext(r.Context())

					// Capture stack trace
					stackTrace := string(debug.Stack())

					// Log the panic with full context
					slog.Error("panic_recovered",
						"request_id", requestID,
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
						"remote_addr", r.RemoteAddr,
						"stack_trace", stackTrace,
					)

					// Return 500 Internal Server Error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					// Include request ID in error response for tracking
					response := fmt.Sprintf(`{"error":"internal server error","request_id":"%s"}`, requestID)
					w.Write([]byte(response))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
