package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	// emailRegex validates email addresses
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// commonPasswords is a small list of commonly used passwords to reject
	commonPasswords = map[string]bool{
		"password":     true,
		"12345678":     true,
		"123456789":    true,
		"qwerty":       true,
		"abc123":       true,
		"password1":    true,
		"password123":  true,
		"admin":        true,
		"letmein":      true,
		"welcome":      true,
		"monkey":       true,
		"dragon":       true,
		"master":       true,
		"sunshine":     true,
		"princess":     true,
		"football":     true,
		"iloveyou":     true,
		"admin123":     true,
		"welcome123":   true,
	}
)

// Validator wraps go-playground validator with custom validations
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator with custom rules
func New() *Validator {
	v := validator.New()

	// Register custom validations
	v.RegisterValidation("strong_password", validateStrongPassword)
	v.RegisterValidation("safe_password", validateSafePassword)

	return &Validator{
		validate: v,
	}
}

// Struct validates a struct
func (v *Validator) Struct(s interface{}) error {
	return v.validate.Struct(s)
}

// Var validates a single variable
func (v *Validator) Var(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// ValidateEmail validates an email address format
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	// Additional checks
	if len(email) > 254 {
		return fmt.Errorf("email is too long")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format")
	}

	localPart := parts[0]
	if len(localPart) > 64 {
		return fmt.Errorf("email local part is too long")
	}

	return nil
}

// PasswordComplexity holds password complexity requirements
type PasswordComplexity struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// DefaultPasswordComplexity returns standard password requirements
func DefaultPasswordComplexity() PasswordComplexity {
	return PasswordComplexity{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

// ValidatePassword validates a password against complexity requirements
func ValidatePassword(password string) error {
	return ValidatePasswordWithRules(password, DefaultPasswordComplexity())
}

// ValidatePasswordWithRules validates a password with custom rules
func ValidatePasswordWithRules(password string, rules PasswordComplexity) error {
	if len(password) < rules.MinLength {
		return fmt.Errorf("password must be at least %d characters long", rules.MinLength)
	}

	if len(password) > 128 {
		return fmt.Errorf("password is too long (max 128 characters)")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if rules.RequireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if rules.RequireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if rules.RequireNumber && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if rules.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check against common passwords (case-insensitive)
	lowerPassword := strings.ToLower(password)
	for commonPwd := range commonPasswords {
		if lowerPassword == commonPwd || password == commonPwd {
			return fmt.Errorf("password is too common, please choose a stronger password")
		}
	}

	// Check for repeated characters (e.g., "aaaaaaa")
	if hasRepeatedChars(password, 4) {
		return fmt.Errorf("password contains too many repeated characters")
	}

	// Check for sequential characters (e.g., "12345", "abcde")
	if hasSequentialChars(password, 4) {
		return fmt.Errorf("password contains sequential characters")
	}

	return nil
}

// hasRepeatedChars checks if password has n or more repeated characters
func hasRepeatedChars(s string, n int) bool {
	if len(s) < n {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= n {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}

// hasSequentialChars checks if password has n or more sequential characters
func hasSequentialChars(s string, n int) bool {
	if len(s) < n {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1]+1 || s[i] == s[i-1]-1 {
			count++
			if count >= n {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}

// validateStrongPassword is a custom validator for strong passwords
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return ValidatePassword(password) == nil
}

// validateSafePassword is a custom validator for safe passwords (less strict)
func validateSafePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	rules := PasswordComplexity{
		MinLength:      8,
		RequireUpper:   false,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: false,
	}
	return ValidatePasswordWithRules(password, rules) == nil
}

// SanitizeString removes potentially dangerous characters from user input
func SanitizeString(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeHTML escapes HTML special characters
// Order matters: ampersand must be escaped first to avoid double-escaping
func SanitizeHTML(input string) string {
	result := input

	// Escape ampersand first
	result = strings.ReplaceAll(result, "&", "&amp;")

	// Then escape other special characters
	result = strings.ReplaceAll(result, "<", "&lt;")
	result = strings.ReplaceAll(result, ">", "&gt;")
	result = strings.ReplaceAll(result, "\"", "&quot;")
	result = strings.ReplaceAll(result, "'", "&#39;")

	return result
}

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 32 {
		return fmt.Errorf("username is too long (max 32 characters)")
	}

	// Username should only contain alphanumeric characters, underscores, and hyphens
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
	}

	// Username should not start or end with special characters
	if strings.HasPrefix(username, "_") || strings.HasPrefix(username, "-") ||
		strings.HasSuffix(username, "_") || strings.HasSuffix(username, "-") {
		return fmt.Errorf("username cannot start or end with special characters")
	}

	return nil
}

// ValidateURL validates a URL format
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL is required")
	}

	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]+$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("invalid URL format")
	}

	return nil
}
