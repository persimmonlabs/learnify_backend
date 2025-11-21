package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sony/gobreaker"
)

// ExampleUsage demonstrates how to use the database resilience patterns
// This file is for documentation purposes and shows best practices

// Example 1: Initialize database with circuit breaker and health monitoring
func ExampleInitializeProduction() {
	ctx := context.Background()

	// Configure database connection
	config := Config{
		Host:            "localhost",
		Port:            5432,
		User:            "myapp",
		Password:        "secret",
		DBName:          "myapp_db",
		SSLMode:         "require",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	// Connect with automatic retry logic
	retryConfig := DefaultRetryConfig()
	db, err := ConnectWithRetry(ctx, config, retryConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Wrap with circuit breaker for production resilience
	cbConfig := DefaultCircuitBreakerConfig()
	cbConfig.OnStateChange = func(name string, from, to gobreaker.State) {
		log.Printf("Circuit breaker state changed: %s -> %s", from, to)
		// Send alert to monitoring system
	}
	cbDB := NewCircuitBreakerDB(db, cbConfig)

	// Start health monitoring
	healthConfig := DefaultHealthThresholds()
	monitor := NewHealthMonitor(db, 30*time.Second, healthConfig)
	monitor.RegisterAlertCallback(func(alert HealthAlert) {
		log.Printf("[%s] Health Alert: %s", alert.Severity, alert.Message)
		// Send to monitoring/alerting system
	})
	monitor.Start()
	defer monitor.Stop()

	// Use cbDB for all database operations
	_ = cbDB
}

// Example 2: Using transactions with automatic rollback
func ExampleTransactionUsage(db *DB) error {
	ctx := context.Background()

	// Simple transaction
	err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Insert user
		_, err := tx.ExecContext(ctx,
			"INSERT INTO users (email, username) VALUES ($1, $2)",
			"user@example.com", "john_doe")
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		// Insert profile
		_, err = tx.ExecContext(ctx,
			"INSERT INTO profiles (user_id, bio) VALUES ($1, $2)",
			1, "Software Engineer")
		if err != nil {
			return fmt.Errorf("failed to insert profile: %w", err)
		}

		// Transaction will automatically commit if no error is returned
		// or rollback if an error is returned
		return nil
	})

	return err
}

// Example 3: Using nested transactions with savepoints
func ExampleNestedTransaction(db *DB) error {
	ctx := context.Background()

	return db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Outer transaction: Create order
		var orderID int
		err := tx.QueryRowContext(ctx,
			"INSERT INTO orders (user_id, total) VALUES ($1, $2) RETURNING id",
			1, 100.00).Scan(&orderID)
		if err != nil {
			return err
		}

		// Nested transaction: Add order items
		err = WithNestedTransaction(ctx, tx, "order_items", func(tx *sql.Tx) error {
			items := []struct{ productID, quantity int }{
				{101, 2},
				{102, 1},
			}

			for _, item := range items {
				_, err := tx.ExecContext(ctx,
					"INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)",
					orderID, item.productID, item.quantity)
				if err != nil {
					return err // This will rollback to savepoint, not entire transaction
				}
			}

			return nil
		})

		// If nested transaction fails, we can still continue or handle differently
		if err != nil {
			log.Printf("Failed to add order items: %v", err)
			// Could rollback entire transaction or handle gracefully
			return err
		}

		return nil
	})
}

// Example 4: Using circuit breaker for queries
func ExampleCircuitBreakerQuery(cbDB *CircuitBreakerDB) error {
	ctx := context.Background()

	// Query with circuit breaker protection
	rows, err := cbDB.Query(ctx, "SELECT id, email FROM users WHERE active = $1", true)
	if err != nil {
		// Circuit breaker may be open, check state
		if cbDB.GetState() == gobreaker.StateOpen {
			log.Println("Circuit breaker is open, database is unhealthy")
			// Return cached data or error
		}
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var email string
		if err := rows.Scan(&id, &email); err != nil {
			return err
		}
		fmt.Printf("User: %d - %s\n", id, email)
	}

	return rows.Err()
}

// Example 5: Using circuit breaker with fallback
func ExampleCircuitBreakerWithFallback(cbDB *CircuitBreakerDB) ([]User, error) {
	ctx := context.Background()

	// Define fallback operation
	fallback := DegradedOperation{
		FallbackFunc: func() (interface{}, error) {
			// Return cached data or default data
			log.Println("Using fallback data due to circuit breaker")
			return []User{
				{ID: 1, Email: "cached@example.com"},
			}, nil
		},
	}

	// Execute with fallback
	result, err := cbDB.ExecuteWithFallback(func(db *DB) (interface{}, error) {
		rows, err := db.QueryContext(ctx, "SELECT id, email FROM users LIMIT 10")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Email); err != nil {
				return nil, err
			}
			users = append(users, u)
		}

		return users, rows.Err()
	}, fallback)

	if err != nil {
		return nil, err
	}

	return result.([]User), nil
}

// Example 6: Running database migrations
func ExampleMigrations(db *DB) error {
	ctx := context.Background()

	// Create migration manager
	mm := NewMigrationManager(db)

	// Initialize migration tables
	if err := mm.Initialize(ctx); err != nil {
		return err
	}

	// Register migrations
	mm.Register(Migration{
		Version:     1,
		Description: "Create users table",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				CREATE TABLE users (
					id SERIAL PRIMARY KEY,
					email VARCHAR(255) UNIQUE NOT NULL,
					username VARCHAR(100) NOT NULL,
					created_at TIMESTAMP DEFAULT NOW()
				)
			`)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE users")
			return err
		},
	})

	mm.Register(Migration{
		Version:     2,
		Description: "Create profiles table",
		Up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`
				CREATE TABLE profiles (
					id SERIAL PRIMARY KEY,
					user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
					bio TEXT,
					created_at TIMESTAMP DEFAULT NOW()
				)
			`)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE profiles")
			return err
		},
	})

	// Check migration status
	if err := mm.Status(ctx); err != nil {
		return err
	}

	// Run migrations (dry-run first to verify)
	if err := mm.MigrateTo(ctx, -1, true); err != nil {
		return err
	}

	// Apply migrations
	if err := mm.Migrate(ctx); err != nil {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

// Example 7: Health monitoring and alerts
func ExampleHealthMonitoring(db *DB) {
	// Create health monitor with custom thresholds
	thresholds := HealthThresholds{
		MaxIdleConnPct:        70.0,
		MinOpenConns:          2,
		MaxConnectionWaitTime: 500 * time.Millisecond,
		PingTimeout:           2 * time.Second,
		QueryTimeout:          3 * time.Second,
	}

	monitor := NewHealthMonitor(db, 15*time.Second, thresholds)

	// Register alert callbacks
	monitor.RegisterAlertCallback(func(alert HealthAlert) {
		switch alert.Severity {
		case "critical":
			// Send to PagerDuty
			log.Printf("CRITICAL: %s", alert.Message)
		case "warning":
			// Send to Slack
			log.Printf("WARNING: %s", alert.Message)
		}
	})

	// Start monitoring
	monitor.Start()

	// Get current metrics
	metrics := monitor.GetMetrics()
	fmt.Printf("Database Health: %v\n", metrics.Healthy)
	fmt.Printf("Open Connections: %d\n", metrics.OpenConnections)
	fmt.Printf("Ping Latency: %v\n", metrics.PingLatency)

	// Print status report
	fmt.Println(monitor.GetConnectionPoolStatus())

	// Recycle stale connections if needed
	ctx := context.Background()
	if metrics.Idle > 10 {
		if err := monitor.RecycleStaleConnections(ctx); err != nil {
			log.Printf("Failed to recycle connections: %v", err)
		}
	}
}

// Example 8: Retry logic for transient errors
func ExampleRetryLogic(db *DB) error {
	ctx := context.Background()

	// Execute query with automatic retry
	rows, err := db.QueryWithRetry(ctx,
		"SELECT id, email FROM users WHERE active = $1",
		true)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Process results...
	return nil
}

// Example 9: Read-only transaction for reporting
func ExampleReadOnlyTransaction(db *DB) error {
	ctx := context.Background()

	return db.ReadOnlyTransaction(ctx, func(tx *sql.Tx) error {
		// This transaction cannot modify data
		rows, err := tx.QueryContext(ctx, `
			SELECT DATE(created_at), COUNT(*)
			FROM users
			GROUP BY DATE(created_at)
			ORDER BY DATE(created_at) DESC
			LIMIT 30
		`)
		if err != nil {
			return err
		}
		defer rows.Close()

		// Process reporting data...
		return nil
	})
}

// Example 10: Serializable transaction for critical operations
func ExampleSerializableTransaction(db *DB) error {
	ctx := context.Background()

	// Use serializable isolation for operations that require consistency
	return db.SerializableTransaction(ctx, func(tx *sql.Tx) error {
		// Check inventory
		var inventory int
		err := tx.QueryRowContext(ctx,
			"SELECT inventory FROM products WHERE id = $1 FOR UPDATE",
			101).Scan(&inventory)
		if err != nil {
			return err
		}

		if inventory < 1 {
			return fmt.Errorf("insufficient inventory")
		}

		// Decrease inventory
		_, err = tx.ExecContext(ctx,
			"UPDATE products SET inventory = inventory - 1 WHERE id = $1",
			101)
		if err != nil {
			return err
		}

		// Create order
		_, err = tx.ExecContext(ctx,
			"INSERT INTO orders (product_id, quantity) VALUES ($1, $2)",
			101, 1)

		return err
	})
}

// Helper types for examples
type User struct {
	ID    int
	Email string
}
