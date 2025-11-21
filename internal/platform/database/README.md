# Database Resilience Package

This package provides production-ready database resilience patterns for PostgreSQL, including transactions, retries, circuit breakers, and health monitoring.

## Features

- **Transaction Management** - Automatic commit/rollback with panic recovery
- **Connection Retry** - Exponential backoff with jitter for transient failures
- **Circuit Breaker** - Automatic fault detection and recovery
- **Health Monitoring** - Continuous connection pool and query performance monitoring
- **Migration Management** - Version-controlled schema migrations with rollback support
- **Nested Transactions** - Savepoint-based nested transaction support

## Quick Start

### 1. Initialize Database with Resilience

```go
import (
    "context"
    "log"
    "time"
    "backend/internal/platform/database"
)

func main() {
    ctx := context.Background()

    // Configure database
    config := database.Config{
        Host:            "localhost",
        Port:            5432,
        User:            "myapp",
        Password:        "secret",
        DBName:          "myapp_db",
        SSLMode:         "require",
        MaxOpenConns:    25,
        MaxIdleConns:    5,
        ConnMaxLifetime: 5 * time.Minute,
    }

    // Connect with retry
    db, err := database.ConnectWithRetry(ctx, config, database.DefaultRetryConfig())
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Wrap with circuit breaker
    cbDB := database.NewCircuitBreakerDB(db, database.DefaultCircuitBreakerConfig())

    // Start health monitoring
    monitor := database.NewHealthMonitor(db, 30*time.Second, database.DefaultHealthThresholds())
    monitor.Start()
    defer monitor.Stop()

    // Use cbDB for all operations
}
```

### 2. Using Transactions

```go
// Simple transaction with automatic rollback on error
err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
    _, err := tx.ExecContext(ctx, "INSERT INTO users (email) VALUES ($1)", "user@example.com")
    if err != nil {
        return err // Automatic rollback
    }
    return nil // Automatic commit
})
```

### 3. Circuit Breaker Protection

```go
// Query with circuit breaker
rows, err := cbDB.Query(ctx, "SELECT id, email FROM users")
if err != nil {
    if cbDB.GetState() == gobreaker.StateOpen {
        log.Println("Circuit breaker open - database unhealthy")
    }
    return err
}
defer rows.Close()
```

### 4. Database Migrations

```go
mm := database.NewMigrationManager(db)
mm.Initialize(ctx)

mm.Register(database.Migration{
    Version:     1,
    Description: "Create users table",
    Up: func(tx *sql.Tx) error {
        _, err := tx.Exec("CREATE TABLE users (...)")
        return err
    },
    Down: func(tx *sql.Tx) error {
        _, err := tx.Exec("DROP TABLE users")
        return err
    },
})

// Run migrations
mm.Migrate(ctx)
```

## Architecture Components

### 1. Transaction Management (`transaction.go`)

**Features:**
- Automatic commit on success
- Automatic rollback on error or panic
- Nested transactions via savepoints
- Read-only transaction support
- Serializable isolation for critical operations

**Usage:**
```go
db.WithTransaction(ctx, func(tx *sql.Tx) error { ... })
db.ReadOnlyTransaction(ctx, func(tx *sql.Tx) error { ... })
db.SerializableTransaction(ctx, func(tx *sql.Tx) error { ... })
```

### 2. Retry Logic (`retry.go`)

**Features:**
- Exponential backoff with configurable multiplier
- Jitter to prevent thundering herd
- Context-aware cancellation
- Retryable error detection
- Automatic query/exec retry wrappers

**Configuration:**
```go
retryConfig := database.RetryConfig{
    MaxAttempts:     5,
    InitialInterval: 100 * time.Millisecond,
    MaxInterval:     10 * time.Second,
    Multiplier:      2.0,
    Jitter:          true,
}
```

### 3. Circuit Breaker (`circuitbreaker.go`)

**Features:**
- Configurable failure threshold
- Half-open state for testing recovery
- State change callbacks for monitoring
- Graceful degradation with fallback operations
- Metrics and monitoring

**Settings:**
```go
cbConfig := database.CircuitBreakerConfig{
    MaxRequests: 2,              // Requests in half-open
    Interval:    60 * time.Second,  // Clear interval
    Timeout:     30 * time.Second,  // Open timeout
    MaxFailures: 5,              // Failures to trigger
}
```

**States:**
- **Closed** - Normal operation, all requests pass through
- **Open** - Circuit tripped, requests fail fast
- **Half-Open** - Testing recovery, limited requests allowed

### 4. Health Monitoring (`health.go`)

**Features:**
- Periodic connection health checks
- Pool statistics monitoring
- Ping and query latency tracking
- Alert callbacks for threshold violations
- Connection recycling

**Monitored Metrics:**
- Open connections
- Idle connections
- Connection wait time
- Ping latency
- Query latency

**Thresholds:**
```go
thresholds := database.HealthThresholds{
    MaxIdleConnPct:        80.0,
    MinOpenConns:          1,
    MaxConnectionWaitTime: 1 * time.Second,
    PingTimeout:           2 * time.Second,
    QueryTimeout:          3 * time.Second,
}
```

### 5. Migration Management (`migrations.go`)

**Features:**
- Version-controlled migrations
- Automatic locking to prevent concurrent migrations
- Rollback support
- Dry-run mode
- Migration status tracking

**Workflow:**
```go
mm := database.NewMigrationManager(db)
mm.Initialize(ctx)
mm.Register(migration)
mm.Status(ctx)           // Check migration status
mm.Migrate(ctx)          // Run pending migrations
mm.Rollback(ctx, 1)      // Rollback last migration
```

## Production Configuration

### Connection Pool Tuning

**High concurrency (API servers):**
```go
MaxOpenConns:    25
MaxIdleConns:    5
ConnMaxLifetime: 5 * time.Minute
ConnMaxIdleTime: 10 * time.Minute
```

**Low concurrency (background jobs):**
```go
MaxOpenConns:    10
MaxIdleConns:    2
ConnMaxLifetime: 15 * time.Minute
ConnMaxIdleTime: 30 * time.Minute
```

### Circuit Breaker Settings

**Conservative (tolerate more failures):**
```go
MaxFailures: 10
Timeout:     60 * time.Second
```

**Aggressive (fail fast):**
```go
MaxFailures: 3
Timeout:     15 * time.Second
```

### Health Monitoring Intervals

- **Production:** 30-60 seconds
- **Development:** 10-15 seconds
- **Critical systems:** 5-10 seconds

## Disaster Recovery

### Backup Scripts

Located in `scripts/`:
- `backup_database.sh` - Full database backup with pg_dump
- `restore_database.sh` - Database restore with verification
- `setup_wal_archiving.sh` - Configure WAL archiving for PITR

### Documentation

Located in `docs/`:
- `database-disaster-recovery.md` - Complete DR procedures
- `database-operations.md` - Operational guide and troubleshooting

## Monitoring Integration

### Metrics to Track

```go
// Circuit breaker metrics
metrics := cbDB.GetMetrics()
// - State (closed/open/half-open)
// - Total requests
// - Success/failure counts
// - Consecutive failures

// Health metrics
healthMetrics := monitor.GetMetrics()
// - Connection pool stats
// - Ping latency
// - Query latency
// - Active connections

// Database stats
stats := db.Stats()
// - OpenConnections
// - InUse
// - Idle
// - WaitCount
// - WaitDuration
```

### Alert Conditions

**Critical:**
- Circuit breaker open for > 1 minute
- Health check failure
- Connection pool exhausted
- Ping latency > 1 second

**Warning:**
- High idle connection percentage (> 80%)
- Connection wait time > 500ms
- Ping latency > 500ms

## Error Handling

### Retryable Errors

Automatically retried:
- Connection failures
- Deadlocks
- Serialization failures
- Temporary network issues

### Non-Retryable Errors

Fail immediately:
- Constraint violations
- Permission errors
- Syntax errors
- Business logic errors

## Testing

### Unit Tests

```bash
go test ./internal/platform/database/...
```

### Integration Tests

```bash
# Requires PostgreSQL running
docker-compose up -d postgres
go test -tags=integration ./internal/platform/database/...
```

### Load Testing

See `database-operations.md` for load testing procedures.

## Best Practices

1. **Always use transactions** for multi-step operations
2. **Wrap with circuit breaker** in production
3. **Monitor health metrics** continuously
4. **Configure retry logic** for startup connections
5. **Keep transactions short** to minimize lock contention
6. **Use appropriate isolation levels** for consistency requirements
7. **Test disaster recovery** procedures quarterly
8. **Monitor circuit breaker state** for early fault detection

## Troubleshooting

See `docs/database-operations.md` for detailed troubleshooting scenarios including:
- Connection pool exhausted
- Slow queries
- Disk space issues
- Replication lag
- Deadlocks

## Recovery Objectives

- **RTO (Recovery Time Objective):** 1 hour
- **RPO (Recovery Point Objective):** 15 minutes
- **Availability Target:** 99.9% uptime

## Dependencies

```go
require (
    github.com/lib/pq v1.10.9
    github.com/sony/gobreaker v1.0.0
)
```

## License

This package is part of the backend application and follows the project license.
