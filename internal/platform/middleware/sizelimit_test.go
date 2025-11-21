package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestSizeLimit(t *testing.T) {
	config := &SizeLimitConfig{
		MaxBodySize: 100, // 100 bytes
	}

	handler := RequestSizeLimit(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the body to verify it works
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))

	// Test small request (should succeed)
	t.Run("Small request", func(t *testing.T) {
		smallBody := strings.Repeat("a", 50)
		req := httptest.NewRequest("POST", "/api/test", strings.NewReader(smallBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}
	})

	// Test large request with Content-Length (should fail immediately)
	t.Run("Large request with Content-Length", func(t *testing.T) {
		largeBody := strings.Repeat("a", 200)
		req := httptest.NewRequest("POST", "/api/test", strings.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = 200
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("Expected status 413, got %d", rr.Code)
		}
	})

	// Test request at exact limit (should succeed)
	t.Run("Request at limit", func(t *testing.T) {
		exactBody := strings.Repeat("a", 100)
		req := httptest.NewRequest("POST", "/api/test", strings.NewReader(exactBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK for exact limit, got %d", rr.Code)
		}
	})
}

func TestRequestSizeLimitBytes(t *testing.T) {
	handler := RequestSizeLimitBytes(50)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(strings.Repeat("x", 100)))
	req.ContentLength = 100
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413, got %d", rr.Code)
	}
}

func TestSmallRequestLimit(t *testing.T) {
	handler := SmallRequestLimit()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create request larger than 100KB
	largeBody := strings.Repeat("a", 200*1024) // 200 KB
	req := httptest.NewRequest("POST", "/", strings.NewReader(largeBody))
	req.ContentLength = int64(len(largeBody))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413 for >100KB request, got %d", rr.Code)
	}
}

func TestLargeRequestLimit(t *testing.T) {
	handler := LargeRequestLimit()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create request smaller than 10MB
	smallBody := strings.Repeat("a", 5*1024*1024) // 5 MB
	req := httptest.NewRequest("POST", "/", strings.NewReader(smallBody))
	req.Header.Set("Content-Type", "multipart/form-data")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK for 5MB upload, got %d", rr.Code)
	}
}

func TestDefaultSizeLimitConfig(t *testing.T) {
	config := DefaultSizeLimitConfig()

	if config.MaxBodySize != DefaultMaxBodySize {
		t.Errorf("Expected default max body size %d, got %d", DefaultMaxBodySize, config.MaxBodySize)
	}
}
