package identity

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@example.co.uk", true},
		{"user+tag@example.com", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"user@", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := emailRegex.MatchString(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestJWTTokenGeneration(t *testing.T) {
	service := &Service{
		jwtSecret: "test-secret-key",
	}

	userID := "user-123"
	email := "test@example.com"

	token, err := service.generateToken(userID, email)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)

	// Verify expiration is set
	assert.NotNil(t, claims.ExpiresAt)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestJWTClaimsStructure(t *testing.T) {
	claims := &Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid password", "password123", true},
		{"too short", "short", false},
		{"exactly 8 chars", "12345678", true},
		{"long password", "this is a very long password with many characters", true},
		{"empty password", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.password) >= 8
			assert.Equal(t, tt.valid, isValid)
		})
	}
}
