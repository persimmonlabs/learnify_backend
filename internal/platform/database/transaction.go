package database

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
)

// TxFunc is a function that operates within a transaction
type TxFunc func(*sql.Tx) error

// TxOptions extends sql.TxOptions with additional retry capabilities
type TxOptions struct {
	sql.TxOptions
	MaxRetries int
}

// WithTransaction executes a function within a database transaction
// It automatically handles commit on success and rollback on error or panic
func (db *DB) WithTransaction(ctx context.Context, fn TxFunc) error {
	return db.WithTransactionOptions(ctx, nil, fn)
}

// WithTransactionOptions executes a function within a transaction with custom options
func (db *DB) WithTransactionOptions(ctx context.Context, opts *TxOptions, fn TxFunc) error {
	tx, err := db.beginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is properly closed
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback()
			debug.PrintStack()
			panic(p) // Re-throw panic after rollback
		}
	}()

	// Execute the transaction function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// beginTx starts a transaction with the given options
func (db *DB) beginTx(ctx context.Context, opts *TxOptions) (*sql.Tx, error) {
	if opts == nil {
		return db.BeginTx(ctx, nil)
	}
	return db.BeginTx(ctx, &opts.TxOptions)
}

// Savepoint creates a savepoint within a transaction for nested transaction support
type Savepoint struct {
	tx   *sql.Tx
	name string
}

// CreateSavepoint creates a named savepoint in the transaction
func CreateSavepoint(ctx context.Context, tx *sql.Tx, name string) (*Savepoint, error) {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to create savepoint %s: %w", name, err)
	}
	return &Savepoint{tx: tx, name: name}, nil
}

// Rollback rolls back to this savepoint
func (s *Savepoint) Rollback(ctx context.Context) error {
	_, err := s.tx.ExecContext(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", s.name))
	if err != nil {
		return fmt.Errorf("failed to rollback to savepoint %s: %w", s.name, err)
	}
	return nil
}

// Release releases the savepoint (commits the nested transaction)
func (s *Savepoint) Release(ctx context.Context) error {
	_, err := s.tx.ExecContext(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", s.name))
	if err != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", s.name, err)
	}
	return nil
}

// WithNestedTransaction executes a function with savepoint support
func WithNestedTransaction(ctx context.Context, tx *sql.Tx, name string, fn TxFunc) error {
	sp, err := CreateSavepoint(ctx, tx, name)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = sp.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := sp.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("nested transaction error: %w, savepoint rollback error: %v", err, rbErr)
		}
		return err
	}

	return sp.Release(ctx)
}

// ReadOnlyTransaction executes a read-only transaction
func (db *DB) ReadOnlyTransaction(ctx context.Context, fn TxFunc) error {
	opts := &TxOptions{
		TxOptions: sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  true,
		},
	}
	return db.WithTransactionOptions(ctx, opts, fn)
}

// SerializableTransaction executes a serializable transaction for critical operations
func (db *DB) SerializableTransaction(ctx context.Context, fn TxFunc) error {
	opts := &TxOptions{
		TxOptions: sql.TxOptions{
			Isolation: sql.LevelSerializable,
			ReadOnly:  false,
		},
	}
	return db.WithTransactionOptions(ctx, opts, fn)
}
