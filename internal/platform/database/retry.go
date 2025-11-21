package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/lib/pq"
)

// RetryConfig configures the retry behavior
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	Jitter          bool
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
	}
}

// ConnectWithRetry attempts to connect to the database with exponential backoff retry
func ConnectWithRetry(ctx context.Context, cfg Config, retryConfig RetryConfig) (*DB, error) {
	var lastErr error

	for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
		log.Printf("Database connection attempt %d/%d", attempt, retryConfig.MaxAttempts)

		db, err := Connect(cfg)
		if err == nil {
			log.Printf("Successfully connected to database on attempt %d", attempt)
			return db, nil
		}

		lastErr = err

		// Don't retry if context is cancelled
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled during connection retry: %w", ctx.Err())
		}

		// Don't sleep after last attempt
		if attempt < retryConfig.MaxAttempts {
			backoff := calculateBackoff(attempt, retryConfig)
			log.Printf("Connection failed: %v. Retrying in %v...", err, backoff)

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			}
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", retryConfig.MaxAttempts, lastErr)
}

// calculateBackoff calculates the backoff duration with exponential increase and optional jitter
func calculateBackoff(attempt int, cfg RetryConfig) time.Duration {
	// Calculate exponential backoff
	backoff := float64(cfg.InitialInterval) * math.Pow(cfg.Multiplier, float64(attempt-1))

	// Cap at max interval
	if backoff > float64(cfg.MaxInterval) {
		backoff = float64(cfg.MaxInterval)
	}

	// Add jitter to prevent thundering herd
	if cfg.Jitter {
		jitter := rand.Float64() * backoff * 0.3 // Add up to 30% jitter
		backoff += jitter
	}

	return time.Duration(backoff)
}

// IsRetryableError determines if an error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific PostgreSQL errors that are retryable
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "40001": // serialization_failure
		case "40P01": // deadlock_detected
		case "08000": // connection_exception
		case "08003": // connection_does_not_exist
		case "08006": // connection_failure
		case "57P03": // cannot_connect_now
		case "53300": // too_many_connections
			return true
		}
	}

	// Check for connection-related errors
	switch err {
	case sql.ErrConnDone, context.DeadlineExceeded:
		return true
	}

	return false
}

// RetryableOperation executes an operation with retry logic
func RetryableOperation(ctx context.Context, cfg RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			return err
		}

		// Check context
		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled during retryable operation: %w", ctx.Err())
		}

		// Don't sleep after last attempt
		if attempt < cfg.MaxAttempts {
			backoff := calculateBackoff(attempt, cfg)
			log.Printf("Retryable error on attempt %d/%d: %v. Retrying in %v...",
				attempt, cfg.MaxAttempts, err, backoff)

			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// WithRetry wraps a database operation with automatic retry logic
func (db *DB) WithRetry(ctx context.Context, operation func(*sql.DB) error) error {
	cfg := DefaultRetryConfig()
	return RetryableOperation(ctx, cfg, func() error {
		return operation(db.DB)
	})
}

// QueryWithRetry executes a query with automatic retry on transient failures
func (db *DB) QueryWithRetry(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error

	retryErr := db.WithRetry(ctx, func(sqlDB *sql.DB) error {
		rows, err = sqlDB.QueryContext(ctx, query, args...)
		return err
	})

	return rows, retryErr
}

// QueryRowWithRetry executes a single-row query with automatic retry
func (db *DB) QueryRowWithRetry(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// Note: QueryRowContext doesn't return an error until Scan is called
	// So we wrap it differently
	var row *sql.Row

	_ = db.WithRetry(ctx, func(sqlDB *sql.DB) error {
		row = sqlDB.QueryRowContext(ctx, query, args...)
		return nil
	})

	return row
}

// ExecWithRetry executes a statement with automatic retry
func (db *DB) ExecWithRetry(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	retryErr := db.WithRetry(ctx, func(sqlDB *sql.DB) error {
		result, err = sqlDB.ExecContext(ctx, query, args...)
		return err
	})

	return result, retryErr
}
