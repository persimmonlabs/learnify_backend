package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	config := DefaultSecurityHeadersConfig()

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	tests := []struct {
		header   string
		expected string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Cross-Origin-Embedder-Policy", "require-corp"},
		{"Cross-Origin-Opener-Policy", "same-origin"},
		{"Cross-Origin-Resource-Policy", "same-origin"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := rr.Header().Get(tt.header)
			if got != tt.expected {
				t.Errorf("Expected %s: %s, got: %s", tt.header, tt.expected, got)
			}
		})
	}

	// Test that dangerous headers are removed
	if rr.Header().Get("X-Powered-By") != "" {
		t.Error("X-Powered-By header should be removed")
	}

	if rr.Header().Get("Server") != "" {
		t.Error("Server header should be removed")
	}
}

func TestSecurityHeaders_HSTS_NoTLS(t *testing.T) {
	config := DefaultSecurityHeadersConfig()

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	// Note: req.TLS is nil (no TLS connection)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// HSTS header should NOT be set for non-TLS connections
	if rr.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS header should not be set for non-TLS connections")
	}
}

func TestSecureHeaders_DefaultConfig(t *testing.T) {
	handler := SecureHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Verify essential headers are set with defaults
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options should be set to nosniff")
	}

	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("X-Frame-Options should be set to DENY")
	}
}

func TestSecurityHeaders_CustomConfig(t *testing.T) {
	config := &SecurityHeadersConfig{
		ContentTypeNosniff: true,
		FrameOptions:       "SAMEORIGIN",
		XSSProtection:      "0",
	}

	handler := SecurityHeaders(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("Expected X-Frame-Options: SAMEORIGIN, got: %s", rr.Header().Get("X-Frame-Options"))
	}

	if rr.Header().Get("X-XSS-Protection") != "0" {
		t.Errorf("Expected X-XSS-Protection: 0, got: %s", rr.Header().Get("X-XSS-Protection"))
	}
}
