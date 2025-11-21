package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func generateTestToken(secret, userID, email string) string {
	claims := &UserClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret"

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		checkContext   bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + generateTestToken(secret, "user-123", "test@example.com"),
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name:           "invalid header format - no Bearer",
			authHeader:     generateTestToken(secret, "user-123", "test@example.com"),
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name:           "wrong secret",
			authHeader:     "Bearer " + generateTestToken("wrong-secret", "user-123", "test@example.com"),
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler that checks context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					claims, ok := GetUserFromContext(r.Context())
					assert.True(t, ok)
					assert.NotNil(t, claims)
					assert.Equal(t, "user-123", claims.UserID)
					assert.Equal(t, "test@example.com", claims.Email)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with auth middleware
			authMiddleware := Auth(secret)
			handler := authMiddleware(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			// Execute
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	secret := "test-secret"

	tests := []struct {
		name         string
		authHeader   string
		checkContext bool
	}{
		{
			name:         "valid token",
			authHeader:   "Bearer " + generateTestToken(secret, "user-123", "test@example.com"),
			checkContext: true,
		},
		{
			name:         "no token - should proceed",
			authHeader:   "",
			checkContext: false,
		},
		{
			name:         "invalid token - should proceed without context",
			authHeader:   "Bearer invalid.token",
			checkContext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, ok := GetUserFromContext(r.Context())
				if tt.checkContext {
					assert.True(t, ok)
					assert.NotNil(t, claims)
					assert.Equal(t, "user-123", claims.UserID)
				} else {
					assert.False(t, ok)
					assert.Nil(t, claims)
				}
				w.WriteHeader(http.StatusOK)
			})

			optionalAuth := OptionalAuth(secret)
			handler := optionalAuth(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// All requests should succeed with optional auth
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func(*http.Request) *http.Request
		expectFound bool
	}{
		{
			name: "user exists in context",
			setupCtx: func(r *http.Request) *http.Request {
				claims := &UserClaims{
					UserID: "user-123",
					Email:  "test@example.com",
				}
				ctx := context.WithValue(r.Context(), userContextKey{}, claims)
				return r.WithContext(ctx)
			},
			expectFound: true,
		},
		{
			name: "user not in context",
			setupCtx: func(r *http.Request) *http.Request {
				return r
			},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setupCtx(req)

			claims, ok := GetUserFromContext(req.Context())

			if tt.expectFound {
				assert.True(t, ok)
				assert.NotNil(t, claims)
				assert.Equal(t, "user-123", claims.UserID)
			} else {
				assert.False(t, ok)
				assert.Nil(t, claims)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Test with no user in context
	userID, ok := GetUserIDFromContext(req.Context())
	assert.False(t, ok)
	assert.Empty(t, userID)

	// Test with user in context
	claims := &UserClaims{
		UserID: "user-123",
		Email:  "test@example.com",
	}
	ctx := context.WithValue(req.Context(), userContextKey{}, claims)
	req = req.WithContext(ctx)

	userID, ok = GetUserIDFromContext(req.Context())
	assert.True(t, ok)
	assert.Equal(t, "user-123", userID)
}

