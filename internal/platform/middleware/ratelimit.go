package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	AuthRequestsPerMinute int
	APIRequestsPerMinute  int
	BurstSize             int
}

// DefaultRateLimiterConfig returns default rate limiter settings
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		AuthRequestsPerMinute: getEnvInt("RATE_LIMIT_AUTH", 10),
		APIRequestsPerMinute:  getEnvInt("RATE_LIMIT_API", 100),
		BurstSize:             getEnvInt("RATE_LIMIT_BURST", 5),
	}
}

// IPRateLimiter manages rate limiters for IP addresses
type IPRateLimiter struct {
	ips     map[string]*rate.Limiter
	mu      sync.RWMutex
	limit   rate.Limit
	burst   int
	cleanup time.Duration
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(requestsPerMinute int, burst int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips:     make(map[string]*rate.Limiter),
		limit:   rate.Limit(float64(requestsPerMinute) / 60.0), // Convert to per-second
		burst:   burst,
		cleanup: 10 * time.Minute,
	}

	// Start cleanup goroutine
	go limiter.cleanupRoutine()

	return limiter
}

// GetLimiter returns the rate limiter for the given IP
func (l *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(l.limit, l.burst)
		l.ips[ip] = limiter
	}

	return limiter
}

// cleanupRoutine periodically removes old limiters
func (l *IPRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		// Create new map to avoid memory leaks
		l.ips = make(map[string]*rate.Limiter)
		l.mu.Unlock()
	}
}

// UserRateLimiter manages rate limiters for authenticated users
type UserRateLimiter struct {
	users   map[string]*rate.Limiter
	mu      sync.RWMutex
	limit   rate.Limit
	burst   int
	cleanup time.Duration
}

// NewUserRateLimiter creates a new user-based rate limiter
func NewUserRateLimiter(requestsPerMinute int, burst int) *UserRateLimiter {
	limiter := &UserRateLimiter{
		users:   make(map[string]*rate.Limiter),
		limit:   rate.Limit(float64(requestsPerMinute) / 60.0), // Convert to per-second
		burst:   burst,
		cleanup: 10 * time.Minute,
	}

	// Start cleanup goroutine
	go limiter.cleanupRoutine()

	return limiter
}

// GetLimiter returns the rate limiter for the given user
func (l *UserRateLimiter) GetLimiter(userID string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.users[userID]
	if !exists {
		limiter = rate.NewLimiter(l.limit, l.burst)
		l.users[userID] = limiter
	}

	return limiter
}

// cleanupRoutine periodically removes old limiters
func (l *UserRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		// Create new map to avoid memory leaks
		l.users = make(map[string]*rate.Limiter)
		l.mu.Unlock()
	}
}

// RateLimitAuth creates a rate limiter for authentication endpoints
func RateLimitAuth(config *RateLimiterConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	limiter := NewIPRateLimiter(config.AuthRequestsPerMinute, config.BurstSize)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)
			if ip == "" {
				writeRateLimitError(w, "unable to determine IP address", http.StatusBadRequest)
				return
			}

			if !limiter.GetLimiter(ip).Allow() {
				w.Header().Set("Retry-After", "60")
				writeRateLimitError(w, fmt.Sprintf("rate limit exceeded: max %d requests per minute", config.AuthRequestsPerMinute), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitAPI creates a rate limiter for API endpoints (user-based)
func RateLimitAPI(config *RateLimiterConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	userLimiter := NewUserRateLimiter(config.APIRequestsPerMinute, config.BurstSize)
	ipLimiter := NewIPRateLimiter(config.APIRequestsPerMinute, config.BurstSize)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get user ID from context (for authenticated requests)
			userID, hasUser := GetUserIDFromContext(r.Context())

			var allowed bool
			if hasUser && userID != "" {
				// Use user-based rate limiting for authenticated requests
				allowed = userLimiter.GetLimiter(userID).Allow()
			} else {
				// Fall back to IP-based rate limiting for unauthenticated requests
				ip := getIP(r)
				if ip == "" {
					writeRateLimitError(w, "unable to determine IP address", http.StatusBadRequest)
					return
				}
				allowed = ipLimiter.GetLimiter(ip).Allow()
			}

			if !allowed {
				w.Header().Set("Retry-After", "60")
				writeRateLimitError(w, fmt.Sprintf("rate limit exceeded: max %d requests per minute", config.APIRequestsPerMinute), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts the IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (used by proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		ips := parseForwardedFor(forwarded)
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// parseForwardedFor parses X-Forwarded-For header
func parseForwardedFor(header string) []string {
	var ips []string
	for _, ip := range splitAndTrim(header, ",") {
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// splitAndTrim splits a string and trims whitespace
func splitAndTrim(s string, sep string) []string {
	var result []string
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitString is a simple string split helper
func splitString(s string, sep string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && isSpace(s[start]) {
		start++
	}

	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isSpace checks if a byte is whitespace
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// writeRateLimitError writes a rate limit error response
func writeRateLimitError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}

// getEnvInt retrieves an environment variable as an integer with a default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
