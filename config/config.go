package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AI       AIConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string
	Host            string
	Env             string
	ShutdownTimeout time.Duration // Graceful shutdown timeout
	RequestTimeout  time.Duration // HTTP request timeout
}

// DatabaseConfig holds PostgreSQL connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AIConfig holds AI service configuration (OpenAI, Anthropic, etc.)
type AIConfig struct {
	Provider string
	APIKey   string
	Model    string
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret           string
	ExpirationSeconds int           // JWT expiration in seconds
	ExpirationDuration time.Duration // JWT expiration as duration (derived from ExpirationSeconds)
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string // Comma-separated list of allowed origins
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// JWT_SECRET is required - no default value for security
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, &ConfigError{
			Field:   "JWT_SECRET",
			Message: "JWT_SECRET environment variable is required and cannot be empty",
		}
	}

	// Validate JWT secret strength
	if err := validateJWTSecret(jwtSecret); err != nil {
		return nil, err
	}

	// JWT expiration in seconds (default 24 hours = 86400 seconds)
	jwtExpirationSeconds := getEnvInt("JWT_EXPIRATION_SECONDS", 86400)

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Env:             getEnv("SERVER_ENV", "development"),
			ShutdownTimeout: getEnvDuration("GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second),
			RequestTimeout:  getEnvDuration("REQUEST_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", getEnv("DB_HOST", "localhost")),
			Port:     getEnv("DATABASE_PORT", getEnv("DB_PORT", "5432")),
			User:     getEnv("DATABASE_USER", getEnv("DB_USER", "postgres")),
			Password: getEnv("DATABASE_PASSWORD", getEnv("DB_PASSWORD", "postgres")),
			DBName:   getEnv("DATABASE_NAME", getEnv("DB_NAME", "learnify")),
			SSLMode:  getEnv("DATABASE_SSL_MODE", getEnv("DB_SSL_MODE", "disable")),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "openai"),
			APIKey:   getEnv("AI_API_KEY", ""),
			Model:    getEnv("AI_MODEL", "gpt-4"),
		},
		JWT: JWTConfig{
			Secret:             jwtSecret,
			ExpirationSeconds:  jwtExpirationSeconds,
			ExpirationDuration: time.Duration(jwtExpirationSeconds) * time.Second,
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		},
	}

	// Validate and warn about configuration issues
	if cfg.Server.Env == "production" {
		if err := validateProductionConfig(cfg); err != nil {
			return nil, err
		}
		warnWeakSecrets(cfg)
	}

	return cfg, nil
}

// validateProductionConfig ensures production environment has secure configuration
func validateProductionConfig(cfg *Config) error {
	// Require strong database password in production
	if cfg.Database.Password == "" || cfg.Database.Password == "postgres" {
		return &ConfigError{
			Field:   "DATABASE_PASSWORD",
			Message: "Production environment requires a strong database password (cannot be empty or default)",
		}
	}

	// Require AI API key in production
	if cfg.AI.APIKey == "" {
		return &ConfigError{
			Field:   "AI_API_KEY",
			Message: "Production environment requires AI API key to be configured",
		}
	}

	// Require SSL mode for database in production
	if cfg.Database.SSLMode == "disable" {
		logWarning("Database SSL is disabled in production - this is insecure for production environments")
	}

	// Require specific CORS origins in production (not wildcard)
	if cfg.CORS.AllowedOrigins == "*" {
		logWarning("CORS is configured with wildcard (*) in production - consider restricting to specific origins")
	}

	return nil
}

// validateJWTSecret ensures JWT secret meets minimum security requirements
func validateJWTSecret(secret string) error {
	if len(secret) < 32 {
		return ErrWeakJWTSecret
	}
	return nil
}

// warnWeakSecrets logs warnings for insecure configuration in production
func warnWeakSecrets(cfg *Config) {
	if cfg.Database.Password == "postgres" {
		logWarning("Database password appears to be default - consider using a stronger password")
	}
}

// logWarning prints a warning message (in production, this would use proper logger)
func logWarning(message string) {
	println("WARNING:", message)
}

// ErrWeakJWTSecret is returned when JWT secret is too weak
var ErrWeakJWTSecret = &ConfigError{
	Field:   "JWT_SECRET",
	Message: "JWT secret must be at least 32 characters long for security",
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool retrieves an environment variable as a boolean or returns a default value
// Accepts: "true", "1", "yes", "on" as true (case-insensitive)
// Accepts: "false", "0", "no", "off" as false (case-insensitive)
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// getEnvDuration retrieves an environment variable as a time.Duration or returns a default value
// Supports duration strings like "30s", "5m", "1h", "24h"
// Also accepts raw integers (interpreted as seconds for backward compatibility)
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		// Try parsing as duration string first (e.g., "30s", "1h")
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}

		// Try parsing as integer (seconds) for backward compatibility
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}
