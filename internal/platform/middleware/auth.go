package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// UserContextKey is the key for user data in context
type userContextKey struct{}

// UserIDKey is the context key for user ID (string type for tracing compatibility)
const UserIDKey ContextKey = "user_id"

// UserClaims represents JWT claims with user information
type UserClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

// Auth validates JWT tokens
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check for Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				writeError(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				writeError(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(*UserClaims)
			if !ok {
				writeError(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			// Add user context to request (both for backward compatibility and tracing)
			ctx := context.WithValue(r.Context(), userContextKey{}, claims)
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth validates JWT tokens but doesn't require them
func OptionalAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token provided, continue without user context
				next.ServeHTTP(w, r)
				return
			}

			// Check for Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Invalid format, continue without user context
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				// Invalid token, continue without user context
				next.ServeHTTP(w, r)
				return
			}

			// Extract claims and add to context if valid
			if claims, ok := token.Claims.(*UserClaims); ok {
				ctx := context.WithValue(r.Context(), userContextKey{}, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves user claims from request context
func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userContextKey{}).(*UserClaims)
	return claims, ok
}

// GetUserIDFromContext retrieves user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return "", false
	}
	return claims.UserID, true
}

// writeError writes an error response
func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
