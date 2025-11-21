package middleware

import (
	"net/http"
	"os"
)

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	ContentTypeNosniff           bool
	FrameOptions                 string
	XSSProtection                string
	StrictTransportSecurity      string
	ContentSecurityPolicy        string
	ReferrerPolicy               string
	PermissionsPolicy            string
	CrossOriginEmbedderPolicy    string
	CrossOriginOpenerPolicy      string
	CrossOriginResourcePolicy    string
}

// DefaultSecurityHeadersConfig returns default security headers
func DefaultSecurityHeadersConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		ContentTypeNosniff:        true,
		FrameOptions:              getEnv("SECURITY_FRAME_OPTIONS", "DENY"),
		XSSProtection:             getEnv("SECURITY_XSS_PROTECTION", "1; mode=block"),
		StrictTransportSecurity:   getEnv("SECURITY_HSTS", "max-age=31536000; includeSubDomains"),
		ContentSecurityPolicy:     getEnv("SECURITY_CSP", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'"),
		ReferrerPolicy:            getEnv("SECURITY_REFERRER_POLICY", "strict-origin-when-cross-origin"),
		PermissionsPolicy:         getEnv("SECURITY_PERMISSIONS_POLICY", "geolocation=(), microphone=(), camera=()"),
		CrossOriginEmbedderPolicy: getEnv("SECURITY_COEP", "require-corp"),
		CrossOriginOpenerPolicy:   getEnv("SECURITY_COOP", "same-origin"),
		CrossOriginResourcePolicy: getEnv("SECURITY_CORP", "same-origin"),
	}
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders(config *SecurityHeadersConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityHeadersConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// X-Content-Type-Options: Prevents MIME type sniffing
			if config.ContentTypeNosniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			// X-Frame-Options: Prevents clickjacking attacks
			if config.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// X-XSS-Protection: Enables browser XSS filter
			if config.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", config.XSSProtection)
			}

			// Strict-Transport-Security: Enforces HTTPS
			if config.StrictTransportSecurity != "" && r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
			}

			// Content-Security-Policy: Prevents XSS and data injection attacks
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Referrer-Policy: Controls referrer information
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			// Permissions-Policy: Controls browser features
			if config.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
			}

			// Cross-Origin-Embedder-Policy: Isolates resources
			if config.CrossOriginEmbedderPolicy != "" {
				w.Header().Set("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedderPolicy)
			}

			// Cross-Origin-Opener-Policy: Isolates browsing contexts
			if config.CrossOriginOpenerPolicy != "" {
				w.Header().Set("Cross-Origin-Opener-Policy", config.CrossOriginOpenerPolicy)
			}

			// Cross-Origin-Resource-Policy: Controls resource loading
			if config.CrossOriginResourcePolicy != "" {
				w.Header().Set("Cross-Origin-Resource-Policy", config.CrossOriginResourcePolicy)
			}

			// Remove potentially dangerous headers
			w.Header().Del("X-Powered-By")
			w.Header().Del("Server")

			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeaders is a convenience function using default security headers
func SecureHeaders() func(http.Handler) http.Handler {
	return SecurityHeaders(DefaultSecurityHeadersConfig())
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
