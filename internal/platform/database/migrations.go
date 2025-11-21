package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          func(*sql.Tx) error
	Down        func(*sql.Tx) error
}

// MigrationManager handles database migrations
type MigrationManager struct {
	db         *DB
	migrations []Migration
	lockTable  string
	versionTable string
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	Version     int
	Description string
	AppliedAt   time.Time
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *DB) *MigrationManager {
	return &MigrationManager{
		db:           db,
		migrations:   make([]Migration, 0),
		lockTable:    "schema_migrations_lock",
		versionTable: "schema_migrations",
	}
}

// Register adds a migration to the manager
func (mm *MigrationManager) Register(m Migration) {
	mm.migrations = append(mm.migrations, m)
}

// Initialize creates the migrations table and lock table
func (mm *MigrationManager) Initialize(ctx context.Context) error {
	// Create version table
	createVersionTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`, mm.versionTable)

	if _, err := mm.db.ExecContext(ctx, createVersionTable); err != nil {
		return fmt.Errorf("failed to create version table: %w", err)
	}

	// Create lock table to prevent concurrent migrations
	createLockTable := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY DEFAULT 1,
			locked BOOLEAN NOT NULL DEFAULT FALSE,
			locked_at TIMESTAMP,
			locked_by TEXT,
			CONSTRAINT single_row CHECK (id = 1)
		)
	`, mm.lockTable)

	if _, err := mm.db.ExecContext(ctx, createLockTable); err != nil {
		return fmt.Errorf("failed to create lock table: %w", err)
	}

	// Insert initial lock row if it doesn't exist
	insertLock := fmt.Sprintf(`
		INSERT INTO %s (id, locked)
		VALUES (1, FALSE)
		ON CONFLICT (id) DO NOTHING
	`, mm.lockTable)

	if _, err := mm.db.ExecContext(ctx, insertLock); err != nil {
		return fmt.Errorf("failed to initialize lock: %w", err)
	}

	log.Println("Migration tables initialized")
	return nil
}

// AcquireLock acquires the migration lock
func (mm *MigrationManager) AcquireLock(ctx context.Context, identifier string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET locked = TRUE, locked_at = NOW(), locked_by = $1
		WHERE id = 1 AND locked = FALSE
	`, mm.lockTable)

	result, err := mm.db.ExecContext(ctx, query, identifier)
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check lock acquisition: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("migration lock is already held by another process")
	}

	log.Printf("Migration lock acquired by %s", identifier)
	return nil
}

// ReleaseLock releases the migration lock
func (mm *MigrationManager) ReleaseLock(ctx context.Context) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET locked = FALSE, locked_at = NULL, locked_by = NULL
		WHERE id = 1
	`, mm.lockTable)

	if _, err := mm.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to release migration lock: %w", err)
	}

	log.Println("Migration lock released")
	return nil
}

// GetAppliedMigrations returns all applied migrations
func (mm *MigrationManager) GetAppliedMigrations(ctx context.Context) ([]MigrationRecord, error) {
	query := fmt.Sprintf(`
		SELECT version, description, applied_at
		FROM %s
		ORDER BY version ASC
	`, mm.versionTable)

	rows, err := mm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var r MigrationRecord
		if err := rows.Scan(&r.Version, &r.Description, &r.AppliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		records = append(records, r)
	}

	return records, rows.Err()
}

// GetCurrentVersion returns the current migration version
func (mm *MigrationManager) GetCurrentVersion(ctx context.Context) (int, error) {
	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(version), 0)
		FROM %s
	`, mm.versionTable)

	var version int
	if err := mm.db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		return 0, fmt.Errorf("failed to get current version: %w", err)
	}

	return version, nil
}

// Migrate runs all pending migrations
func (mm *MigrationManager) Migrate(ctx context.Context) error {
	return mm.MigrateTo(ctx, -1, false)
}

// MigrateTo migrates to a specific version (use -1 for latest)
func (mm *MigrationManager) MigrateTo(ctx context.Context, targetVersion int, dryRun bool) error {
	// Sort migrations by version
	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version < mm.migrations[j].Version
	})

	// Acquire lock
	lockID := fmt.Sprintf("migration-%d", time.Now().Unix())
	if err := mm.AcquireLock(ctx, lockID); err != nil {
		return err
	}
	defer mm.ReleaseLock(ctx)

	// Get current version
	currentVersion, err := mm.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}

	log.Printf("Current migration version: %d", currentVersion)

	// Determine target version
	if targetVersion == -1 {
		if len(mm.migrations) > 0 {
			targetVersion = mm.migrations[len(mm.migrations)-1].Version
		} else {
			targetVersion = 0
		}
	}

	if dryRun {
		log.Println("DRY RUN MODE - No changes will be applied")
	}

	// Apply pending migrations
	applied := 0
	for _, migration := range mm.migrations {
		if migration.Version <= currentVersion {
			continue
		}
		if migration.Version > targetVersion {
			break
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Description)

		if dryRun {
			log.Printf("  [DRY RUN] Would apply migration %d", migration.Version)
			continue
		}

		if err := mm.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		applied++
	}

	if dryRun {
		log.Printf("DRY RUN complete. Would apply %d migrations", applied)
	} else {
		log.Printf("Successfully applied %d migrations", applied)
	}

	return nil
}

// applyMigration applies a single migration
func (mm *MigrationManager) applyMigration(ctx context.Context, m Migration) error {
	return mm.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Execute migration
		if err := m.Up(tx); err != nil {
			return fmt.Errorf("migration up failed: %w", err)
		}

		// Record migration
		query := fmt.Sprintf(`
			INSERT INTO %s (version, description, applied_at)
			VALUES ($1, $2, NOW())
		`, mm.versionTable)

		if _, err := tx.ExecContext(ctx, query, m.Version, m.Description); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		return nil
	})
}

// Rollback rolls back the last N migrations
func (mm *MigrationManager) Rollback(ctx context.Context, steps int) error {
	// Sort migrations by version descending for rollback
	sort.Slice(mm.migrations, func(i, j int) bool {
		return mm.migrations[i].Version > mm.migrations[j].Version
	})

	// Acquire lock
	lockID := fmt.Sprintf("rollback-%d", time.Now().Unix())
	if err := mm.AcquireLock(ctx, lockID); err != nil {
		return err
	}
	defer mm.ReleaseLock(ctx)

	// Get applied migrations
	applied, err := mm.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Determine how many to rollback
	toRollback := steps
	if toRollback > len(applied) || toRollback < 0 {
		toRollback = len(applied)
	}

	// Rollback migrations
	rolledBack := 0
	for i := len(applied) - 1; i >= len(applied)-toRollback; i-- {
		version := applied[i].Version

		// Find migration definition
		var migration *Migration
		for _, m := range mm.migrations {
			if m.Version == version {
				migration = &m
				break
			}
		}

		if migration == nil {
			return fmt.Errorf("migration definition not found for version %d", version)
		}

		log.Printf("Rolling back migration %d: %s", version, migration.Description)

		if err := mm.rollbackMigration(ctx, *migration); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", version, err)
		}

		rolledBack++
	}

	log.Printf("Successfully rolled back %d migrations", rolledBack)
	return nil
}

// rollbackMigration rolls back a single migration
func (mm *MigrationManager) rollbackMigration(ctx context.Context, m Migration) error {
	return mm.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Execute rollback
		if err := m.Down(tx); err != nil {
			return fmt.Errorf("migration down failed: %w", err)
		}

		// Remove migration record
		query := fmt.Sprintf(`
			DELETE FROM %s WHERE version = $1
		`, mm.versionTable)

		if _, err := tx.ExecContext(ctx, query, m.Version); err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}

		return nil
	})
}

// Status prints the migration status
func (mm *MigrationManager) Status(ctx context.Context) error {
	applied, err := mm.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	appliedMap := make(map[int]MigrationRecord)
	for _, record := range applied {
		appliedMap[record.Version] = record
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("================")

	for _, m := range mm.migrations {
		if record, ok := appliedMap[m.Version]; ok {
			fmt.Printf("✓ %d: %s (applied at %s)\n", m.Version, m.Description, record.AppliedAt.Format(time.RFC3339))
		} else {
			fmt.Printf("✗ %d: %s (pending)\n", m.Version, m.Description)
		}
	}

	return nil
}
