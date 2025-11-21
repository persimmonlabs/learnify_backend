package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// Admin middleware ensures only admin users can access the endpoint
func Admin(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract claims from context (set by Auth middleware)
			// For now, we'll extract from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized: no authorization header", http.StatusUnauthorized)
				return
			}

			// Remove "Bearer " prefix
			tokenString := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			}

			// Parse token
			token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(*UserClaims)
			if !ok || !token.Valid {
				http.Error(w, "Unauthorized: invalid token claims", http.StatusUnauthorized)
				return
			}

			// Check if user is admin
			if !claims.IsAdmin {
				http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
				return
			}

			// Add admin context
			ctx := context.WithValue(r.Context(), "is_admin", true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin middleware enforces strict admin-only access
// Uses JWT claims to verify admin status
func RequireAdmin(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context (set by Auth middleware)
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized: authentication required", http.StatusUnauthorized)
				return
			}

			// Enforce strict admin check
			if !claims.IsAdmin {
				http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
				return
			}

			// Add admin context flag
			ctx := context.WithValue(r.Context(), "is_admin", true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
