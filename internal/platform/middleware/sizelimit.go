package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	// DefaultMaxBodySize is the default maximum request body size (1MB)
	DefaultMaxBodySize int64 = 1 << 20 // 1 MB

	// MaxBodySize10MB for file upload endpoints
	MaxBodySize10MB int64 = 10 << 20 // 10 MB

	// MaxBodySize100KB for standard API endpoints
	MaxBodySize100KB int64 = 100 << 10 // 100 KB
)

// SizeLimitConfig holds request size limit configuration
type SizeLimitConfig struct {
	MaxBodySize int64
}

// DefaultSizeLimitConfig returns default size limit configuration
func DefaultSizeLimitConfig() *SizeLimitConfig {
	maxSize := DefaultMaxBodySize
	if envSize := getEnvInt64("MAX_REQUEST_SIZE", 0); envSize > 0 {
		maxSize = envSize
	}
	return &SizeLimitConfig{
		MaxBodySize: maxSize,
	}
}

// RequestSizeLimit limits the size of request bodies
func RequestSizeLimit(config *SizeLimitConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSizeLimitConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check Content-Length header first (optimization)
			if r.ContentLength > config.MaxBodySize {
				writeSizeLimitError(w, fmt.Sprintf("request body too large: %d bytes (max: %d bytes)", r.ContentLength, config.MaxBodySize), http.StatusRequestEntityTooLarge)
				return
			}

			// Wrap the request body with a limited reader
			r.Body = http.MaxBytesReader(w, r.Body, config.MaxBodySize)

			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeLimitBytes creates a size limiter with a specific byte limit
func RequestSizeLimitBytes(maxBytes int64) func(http.Handler) http.Handler {
	return RequestSizeLimit(&SizeLimitConfig{MaxBodySize: maxBytes})
}

// SmallRequestLimit limits requests to 100KB (for most API endpoints)
func SmallRequestLimit() func(http.Handler) http.Handler {
	return RequestSizeLimitBytes(MaxBodySize100KB)
}

// LargeRequestLimit limits requests to 10MB (for file upload endpoints)
func LargeRequestLimit() func(http.Handler) http.Handler {
	return RequestSizeLimitBytes(MaxBodySize10MB)
}

// limitedReader wraps an io.Reader to enforce size limits
type limitedReader struct {
	reader    io.Reader
	remaining int64
}

// Read reads data while enforcing the size limit
func (lr *limitedReader) Read(p []byte) (n int, err error) {
	if lr.remaining <= 0 {
		return 0, fmt.Errorf("request body too large")
	}

	if int64(len(p)) > lr.remaining {
		p = p[:lr.remaining]
	}

	n, err = lr.reader.Read(p)
	lr.remaining -= int64(n)

	return n, err
}

// Close closes the underlying reader if it implements io.Closer
func (lr *limitedReader) Close() error {
	if closer, ok := lr.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// writeSizeLimitError writes a size limit error response
func writeSizeLimitError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}

// getEnvInt64 retrieves an environment variable as int64 with a default
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := getEnv(key, ""); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
