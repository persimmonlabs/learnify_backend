package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS handles Cross-Origin Resource Sharing with default configuration
func CORS() func(http.Handler) http.Handler {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig handles CORS with custom configuration
func CORSWithConfig(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowedOrigin := ""
			if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
				allowedOrigin = "*"
			} else {
				for _, allowedOrigin := range config.AllowedOrigins {
					if allowedOrigin == origin {
						allowedOrigin = origin
						break
					}
				}
			}

			// Set CORS headers if origin is allowed
			if allowedOrigin != "" {
				// Set Access-Control-Allow-Origin
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)

				// Set Access-Control-Allow-Methods
				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				}

				// Set Access-Control-Allow-Headers
				if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				}

				// Set Access-Control-Expose-Headers
				if len(config.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
				}

				// Set Access-Control-Allow-Credentials
				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Set Access-Control-Max-Age for preflight requests
				if r.Method == http.MethodOptions && config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
				}
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// CORSPermissive returns a CORS middleware with permissive settings for development
func CORSPermissive() func(http.Handler) http.Handler {
	config := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false, // Can't use credentials with wildcard origin
		MaxAge:           86400,
	}
	return CORSWithConfig(config)
}

// CORSStrict returns a CORS middleware with strict settings for production
func CORSStrict(allowedOrigins []string) func(http.Handler) http.Handler {
	config := CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
	return CORSWithConfig(config)
}
