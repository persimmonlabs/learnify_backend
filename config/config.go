package config

import (
	"os"
	"strconv"
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
	Port string
	Host string
	Env  string
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
	Secret     string
	Expiration int
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

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Env:  getEnv("SERVER_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "learnify"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "openai"),
			APIKey:   getEnv("AI_API_KEY", ""),
			Model:    getEnv("AI_MODEL", "gpt-4"),
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			Expiration: getEnvInt("JWT_EXPIRATION", 24), // hours
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		},
	}

	// Warn about weak secrets in production
	if cfg.Server.Env == "production" {
		warnWeakSecrets(cfg)
	}

	return cfg, nil
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
