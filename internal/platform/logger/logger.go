package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

// Logger wraps structured logging
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Env    string
	Level  slog.Level
	Output io.Writer
}

// New creates a new structured logger
func New(env string) *Logger {
	return NewWithConfig(Config{
		Env:    env,
		Level:  slog.LevelInfo,
		Output: os.Stdout,
	})
}

// NewWithConfig creates a new logger with custom configuration
func NewWithConfig(cfg Config) *Logger {
	var handler slog.Handler

	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
		AddSource: cfg.Env == "development",
	}

	// Use JSON handler for production, text handler for development
	if cfg.Env == "production" {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithContext adds context fields to logger
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	args := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		args = append(args, key, value)
	}

	return &Logger{
		Logger: l.With(args...),
	}
}

// WithField adds a single field to logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger: l.With(key, value),
	}
}

// WithRequestID adds request ID to logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithField("request_id", requestID)
}

// WithUserID adds user ID to logger
func (l *Logger) WithUserID(userID string) *Logger {
	return l.WithField("user_id", userID)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*Logger); ok {
		return logger
	}
	return New("production") // fallback
}

// ToContext adds logger to context
func (l *Logger) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

type loggerKey struct{}

// ContextKey type for logger context keys
type ContextKey string

const (
	// CorrelationIDKey is the context key for correlation/request ID
	CorrelationIDKey ContextKey = "correlation_id"
)

// GenerateCorrelationID generates a new correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// WithCorrelationID adds correlation ID to logger from context
func (l *Logger) WithCorrelationID(ctx context.Context) *Logger {
	if corrID, ok := ctx.Value(CorrelationIDKey).(string); ok && corrID != "" {
		return l.WithField("correlation_id", corrID)
	}
	// Also check for request_id key (from middleware)
	type requestIDKey string
	if reqID, ok := ctx.Value(requestIDKey("request_id")).(string); ok && reqID != "" {
		return l.WithField("correlation_id", reqID)
	}
	return l
}

// Helper methods for common log patterns

// LogError logs an error with context
func (l *Logger) LogError(msg string, err error, fields map[string]interface{}) {
	args := []any{"error", err}
	for key, value := range fields {
		args = append(args, key, value)
	}
	l.Error(msg, args...)
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(method, path string, statusCode int, duration int64, fields map[string]interface{}) {
	args := []any{
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", duration,
	}
	for key, value := range fields {
		args = append(args, key, value)
	}
	l.Info("http_request", args...)
}

// LogDatabaseQuery logs a database query
func (l *Logger) LogDatabaseQuery(query string, duration int64, err error) {
	args := []any{
		"query", query,
		"duration_ms", duration,
	}
	if err != nil {
		args = append(args, "error", err)
		l.Error("database_query_failed", args...)
	} else {
		l.Debug("database_query", args...)
	}
}
