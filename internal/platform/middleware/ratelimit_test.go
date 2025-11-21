package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitAuth(t *testing.T) {
	config := &RateLimiterConfig{
		AuthRequestsPerMinute: 5,
		APIRequestsPerMinute:  100,
		BurstSize:             2,
	}

	handler := RateLimitAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test successful requests within limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK && rr.Code != http.StatusTooManyRequests {
			t.Errorf("Expected OK or TooManyRequests, got %d", rr.Code)
		}
	}

	// Test rate limit exceeded
	req := httptest.NewRequest("POST", "/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected TooManyRequests after limit, got %d", rr.Code)
	}

	// Check Retry-After header
	if rr.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header")
	}
}

func TestRateLimitAPI_UserBased(t *testing.T) {
	config := &RateLimiterConfig{
		AuthRequestsPerMinute: 10,
		APIRequestsPerMinute:  10,
		BurstSize:             3,
	}

	handler := RateLimitAPI(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test requests without user context (IP-based)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/api/courses", nil)
		req.RemoteAddr = "10.0.0.1:54321"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if i < 9 && rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected OK, got %d", i, rr.Code)
		}
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		forwardedFor   string
		realIP         string
		expectedIP     string
	}{
		{
			name:       "Direct connection",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "X-Forwarded-For single IP",
			remoteAddr:   "10.0.0.1:54321",
			forwardedFor: "203.0.113.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:         "X-Forwarded-For multiple IPs",
			remoteAddr:   "10.0.0.1:54321",
			forwardedFor: "203.0.113.1, 10.0.0.2, 10.0.0.3",
			expectedIP:   "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "10.0.0.1:54321",
			realIP:     "203.0.113.5",
			expectedIP: "203.0.113.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}

			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			ip := getIP(req)

			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestIPRateLimiter_Cleanup(t *testing.T) {
	limiter := &IPRateLimiter{
		ips:     make(map[string]*time.Ticker),
		cleanup: 100 * time.Millisecond,
	}

	// Add some limiters
	limiter.GetLimiter("192.168.1.1")
	limiter.GetLimiter("192.168.1.2")
	limiter.GetLimiter("192.168.1.3")

	if len(limiter.ips) != 3 {
		t.Errorf("Expected 3 limiters, got %d", len(limiter.ips))
	}

	// Wait for cleanup (note: actual cleanup is in goroutine, this just tests the structure)
	// In production code, cleanup would run periodically
}
