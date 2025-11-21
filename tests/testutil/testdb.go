package testutil

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
// Note: This requires a running PostgreSQL instance
// For CI/CD, consider using testcontainers-go or similar
func SetupTestDB(t *testing.T) *sql.DB {
	// Use environment variables for test database
	// In production CI/CD, use docker-compose or testcontainers
	connStr := "postgres://test:test@localhost:5432/backend_test?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test - database not reachable: %v", err)
		db.Close()
		return nil
	}

	// Clean up function
	t.Cleanup(func() {
		CleanupTestDB(t, db)
		db.Close()
	})

	return db
}

// CleanupTestDB removes test data
func CleanupTestDB(t *testing.T, db *sql.DB) {
	tables := []string{
		"architecture_reviews",
		"module_completions",
		"user_progress",
		"exercises",
		"generated_modules",
		"generated_courses",
		"blueprint_modules",
		"user_variables",
		"user_archetypes",
		"user_achievements",
		"recommendations",
		"activity_feed",
		"trending_courses",
		"user_relationships",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			// Table might not exist, ignore error
			t.Logf("Warning: Failed to truncate %s: %v", table, err)
		}
	}
}

// CreateTestSchema creates test database schema
// This is a simplified version - in production use migrations
func CreateTestSchema(t *testing.T, db *sql.DB) {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255),
			avatar_url VARCHAR(255),
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			last_login TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Add other tables as needed for integration tests
}
