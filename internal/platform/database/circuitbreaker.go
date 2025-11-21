package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerDB wraps a database connection with circuit breaker protection
type CircuitBreakerDB struct {
	db *DB
	cb *gobreaker.CircuitBreaker
}

// CircuitBreakerConfig defines circuit breaker settings
type CircuitBreakerConfig struct {
	Name          string
	MaxRequests   uint32        // Max requests allowed in half-open state
	Interval      time.Duration // Interval to clear internal counts
	Timeout       time.Duration // Timeout in open state before half-open
	MaxFailures   uint32        // Consecutive failures to trigger open state
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// DefaultCircuitBreakerConfig returns recommended circuit breaker settings
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:        "database",
		MaxRequests: 2,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		MaxFailures: 5,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit breaker '%s' state changed from %s to %s", name, from, to)
		},
	}
}

// NewCircuitBreakerDB creates a new database wrapper with circuit breaker
func NewCircuitBreakerDB(db *DB, config CircuitBreakerConfig) *CircuitBreakerDB {
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip to open state after consecutive failures
			return counts.ConsecutiveFailures >= config.MaxFailures
		},
		OnStateChange: config.OnStateChange,
		IsSuccessful: func(err error) bool {
			// Only count database errors as failures, not business logic errors
			return !IsRetryableError(err)
		},
	}

	cb := gobreaker.NewCircuitBreaker(settings)

	return &CircuitBreakerDB{
		db: db,
		cb: cb,
	}
}

// GetState returns the current circuit breaker state
func (cbdb *CircuitBreakerDB) GetState() gobreaker.State {
	return cbdb.cb.State()
}

// GetCounts returns circuit breaker counts
func (cbdb *CircuitBreakerDB) GetCounts() gobreaker.Counts {
	return cbdb.cb.Counts()
}

// Execute wraps any database operation with circuit breaker protection
func (cbdb *CircuitBreakerDB) Execute(operation func(*DB) (interface{}, error)) (interface{}, error) {
	return cbdb.cb.Execute(func() (interface{}, error) {
		return operation(cbdb.db)
	})
}

// Query executes a query with circuit breaker protection
func (cbdb *CircuitBreakerDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	result, err := cbdb.Execute(func(db *DB) (interface{}, error) {
		return db.QueryContext(ctx, query, args...)
	})

	if err != nil {
		return nil, err
	}

	return result.(*sql.Rows), nil
}

// QueryRow executes a single-row query with circuit breaker protection
func (cbdb *CircuitBreakerDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	result, err := cbdb.Execute(func(db *DB) (interface{}, error) {
		row := db.QueryRowContext(ctx, query, args...)
		return row, nil
	})

	if err != nil {
		// Return a row with the error that will be returned on Scan
		log.Printf("Circuit breaker blocked query: %v", err)
		return &sql.Row{} // This will return an error on Scan
	}

	return result.(*sql.Row)
}

// Exec executes a statement with circuit breaker protection
func (cbdb *CircuitBreakerDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := cbdb.Execute(func(db *DB) (interface{}, error) {
		return db.ExecContext(ctx, query, args...)
	})

	if err != nil {
		return nil, err
	}

	return result.(sql.Result), nil
}

// WithTransaction executes a transaction with circuit breaker protection
func (cbdb *CircuitBreakerDB) WithTransaction(ctx context.Context, fn TxFunc) error {
	_, err := cbdb.Execute(func(db *DB) (interface{}, error) {
		return nil, db.WithTransaction(ctx, fn)
	})
	return err
}

// Ping checks database connectivity with circuit breaker protection
func (cbdb *CircuitBreakerDB) Ping(ctx context.Context) error {
	_, err := cbdb.Execute(func(db *DB) (interface{}, error) {
		return nil, db.PingContext(ctx)
	})
	return err
}

// HealthCheck performs health check with circuit breaker protection
func (cbdb *CircuitBreakerDB) HealthCheck() error {
	_, err := cbdb.Execute(func(db *DB) (interface{}, error) {
		return nil, db.HealthCheck()
	})
	return err
}

// Stats returns database statistics
func (cbdb *CircuitBreakerDB) Stats() sql.DBStats {
	return cbdb.db.Stats()
}

// Close closes the underlying database connection
func (cbdb *CircuitBreakerDB) Close() error {
	return cbdb.db.Close()
}

// GetDB returns the underlying database for operations that need direct access
func (cbdb *CircuitBreakerDB) GetDB() *DB {
	return cbdb.db
}

// CircuitBreakerMetrics holds metrics for monitoring
type CircuitBreakerMetrics struct {
	State               string    `json:"state"`
	TotalRequests       uint32    `json:"total_requests"`
	TotalSuccesses      uint32    `json:"total_successes"`
	TotalFailures       uint32    `json:"total_failures"`
	ConsecutiveSuccesses uint32    `json:"consecutive_successes"`
	ConsecutiveFailures uint32    `json:"consecutive_failures"`
	Timestamp           time.Time `json:"timestamp"`
}

// GetMetrics returns current circuit breaker metrics
func (cbdb *CircuitBreakerDB) GetMetrics() CircuitBreakerMetrics {
	counts := cbdb.cb.Counts()
	return CircuitBreakerMetrics{
		State:                cbdb.cb.State().String(),
		TotalRequests:        counts.Requests,
		TotalSuccesses:       counts.TotalSuccesses,
		TotalFailures:        counts.TotalFailures,
		ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		ConsecutiveFailures:  counts.ConsecutiveFailures,
		Timestamp:            time.Now(),
	}
}

// DegradedOperation provides fallback behavior when circuit is open
type DegradedOperation struct {
	CacheEnabled bool
	CacheTTL     time.Duration
	FallbackFunc func() (interface{}, error)
}

// ExecuteWithFallback attempts operation, falls back to cache or alternative on circuit open
func (cbdb *CircuitBreakerDB) ExecuteWithFallback(operation func(*DB) (interface{}, error), fallback DegradedOperation) (interface{}, error) {
	result, err := cbdb.Execute(operation)

	if err != nil {
		// Circuit is open or operation failed
		if fallback.FallbackFunc != nil {
			log.Printf("Circuit breaker triggered, using fallback operation")
			return fallback.FallbackFunc()
		}
		return nil, fmt.Errorf("operation failed and no fallback available: %w", err)
	}

	return result, nil
}
