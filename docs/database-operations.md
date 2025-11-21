# Database Operations Guide

## Overview

This guide provides operational procedures, best practices, and troubleshooting steps for maintaining the PostgreSQL database in production.

## Connection Pool Tuning

### Recommended Settings

```go
// For API servers (high concurrency, short transactions)
MaxOpenConns:    25  // Maximum connections per instance
MaxIdleConns:    5   // Idle connections to keep ready
ConnMaxLifetime: 5 * time.Minute   // Recycle connections
ConnMaxIdleTime: 10 * time.Minute  // Close idle connections

// For batch/background jobs (low concurrency, long transactions)
MaxOpenConns:    10
MaxIdleConns:    2
ConnMaxLifetime: 15 * time.Minute
ConnMaxIdleTime: 30 * time.Minute
```

### Calculating Pool Size

**Formula:**
```
connections_needed = (core_count * 2) + effective_spindle_count
max_pool_size = connections_needed / number_of_app_instances
```

**Example:**
- Server: 4 cores, SSD storage (spindle = 4)
- Formula: (4 * 2) + 4 = 12 connections needed
- App instances: 3
- Pool size per instance: 12 / 3 = 4 active connections
- Set MaxOpenConns: 10 (buffer for spikes)
- Set MaxIdleConns: 4

### Monitoring Pool Health

```go
stats := db.Stats()

// Critical metrics
fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
fmt.Printf("In Use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
fmt.Printf("Wait Count: %d\n", stats.WaitCount)
fmt.Printf("Wait Duration: %v\n", stats.WaitDuration)
fmt.Printf("Max Idle Closed: %d\n", stats.MaxIdleClosed)
fmt.Printf("Max Lifetime Closed: %d\n", stats.MaxLifetimeClosed)

// Alert conditions
if stats.WaitCount > 100 {
    log.Warn("High connection wait count, consider increasing pool size")
}

if float64(stats.Idle)/float64(stats.OpenConnections) > 0.8 {
    log.Warn("High idle connection ratio, consider decreasing MaxIdleConns")
}
```

## Query Optimization

### Identifying Slow Queries

```sql
-- Enable query logging in postgresql.conf
log_min_duration_statement = 1000  -- Log queries slower than 1s

-- View slow query log
tail -f /var/log/postgresql/postgresql-14-main.log | grep "duration:"

-- Top 10 slowest queries (requires pg_stat_statements)
SELECT
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    max_exec_time,
    stddev_exec_time,
    rows
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

### Query Analysis Checklist

1. **Explain Analyze** - Always test queries
```sql
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT * FROM users WHERE email = 'user@example.com';
```

2. **Look for:**
   - Sequential scans on large tables (add index)
   - High buffer reads (inefficient query)
   - Nested loops with large row counts (consider JOIN order)
   - Bitmap heap scans (may need better index)

3. **Common Fixes:**
```sql
-- Missing index
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- Partial index for common queries
CREATE INDEX idx_active_users ON users(status) WHERE status = 'active';

-- Composite index for multi-column queries
CREATE INDEX idx_users_status_created ON users(status, created_at);

-- Covering index to avoid table lookups
CREATE INDEX idx_users_lookup ON users(email) INCLUDE (id, username);
```

## Index Management

### Index Monitoring

```sql
-- Find missing indexes (requires pg_stat_statements)
SELECT schemaname, tablename,
       seq_scan, seq_tup_read,
       idx_scan, idx_tup_fetch,
       seq_tup_read / seq_scan AS avg_seq_tup
FROM pg_stat_user_tables
WHERE seq_scan > 0
  AND seq_scan > idx_scan
ORDER BY seq_tup_read DESC
LIMIT 10;

-- Find unused indexes
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%pkey'
ORDER BY pg_relation_size(indexrelid) DESC;

-- Index bloat detection
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
       pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) AS index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Index Maintenance

```sql
-- Rebuild bloated indexes (online, no downtime)
REINDEX INDEX CONCURRENTLY idx_users_email;

-- Rebuild all indexes on a table
REINDEX TABLE CONCURRENTLY users;

-- Remove duplicate indexes
-- First, identify duplicates manually, then:
DROP INDEX CONCURRENTLY idx_duplicate_name;
```

## Vacuum and Maintenance

### Autovacuum Configuration

**postgresql.conf settings:**
```ini
# Enable autovacuum
autovacuum = on
autovacuum_max_workers = 3
autovacuum_naptime = 1min

# Aggressive for high-write tables
autovacuum_vacuum_scale_factor = 0.05     # Vacuum after 5% of table changes
autovacuum_analyze_scale_factor = 0.02    # Analyze after 2% of table changes
autovacuum_vacuum_cost_limit = 2000       # Higher = faster vacuum

# For large tables
autovacuum_vacuum_cost_delay = 10ms
```

### Manual Vacuum

```sql
-- Standard vacuum (doesn't block reads/writes)
VACUUM ANALYZE users;

-- Aggressive vacuum (reclaims more space)
VACUUM FULL ANALYZE users;  -- WARNING: Locks table!

-- Vacuum with verbose output
VACUUM (VERBOSE, ANALYZE) users;
```

### Monitoring Vacuum Progress

```sql
-- Check last vacuum times
SELECT schemaname, relname,
       last_vacuum, last_autovacuum,
       last_analyze, last_autoanalyze,
       n_dead_tup, n_live_tup,
       n_dead_tup::float / NULLIF(n_live_tup, 0) AS dead_ratio
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;

-- View active vacuum operations
SELECT pid, age(clock_timestamp(), query_start), usename, query
FROM pg_stat_activity
WHERE query LIKE '%VACUUM%' AND query NOT LIKE '%pg_stat_activity%';
```

### Maintenance Schedule

| Task | Frequency | Command | Downtime Required |
|------|-----------|---------|-------------------|
| Analyze statistics | Daily | `ANALYZE` | No |
| Vacuum dead tuples | Daily (autovacuum) | `VACUUM` | No |
| Reindex bloated indexes | Weekly | `REINDEX CONCURRENTLY` | No |
| VACUUM FULL | Quarterly | `VACUUM FULL` | Yes (off-hours) |
| Update planner statistics | After bulk changes | `ANALYZE` | No |

## Common Troubleshooting Scenarios

### 1. Connection Pool Exhausted

**Symptoms:**
```
Error: pq: sorry, too many clients already
Error: connection timeout
```

**Diagnosis:**
```sql
-- Check active connections
SELECT count(*), state, wait_event_type
FROM pg_stat_activity
WHERE datname = 'your_database'
GROUP BY state, wait_event_type;

-- Find long-running queries
SELECT pid, now() - pg_stat_activity.query_start AS duration, query
FROM pg_stat_activity
WHERE state = 'active'
  AND now() - pg_stat_activity.query_start > interval '5 minutes';
```

**Solution:**
```sql
-- Kill long-running queries
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE pid = <problematic_pid>;

-- Increase max_connections (requires restart)
ALTER SYSTEM SET max_connections = 200;
-- Then: sudo systemctl restart postgresql
```

### 2. Slow Queries After Deployment

**Symptoms:**
- Previously fast queries now slow
- Different execution plans

**Diagnosis:**
```sql
-- Check if statistics are stale
SELECT schemaname, tablename, last_analyze, last_autoanalyze
FROM pg_stat_user_tables
WHERE last_analyze < NOW() - interval '1 week';

-- Compare query plans
EXPLAIN (ANALYZE, BUFFERS) <your_query>;
```

**Solution:**
```sql
-- Update statistics
ANALYZE;

-- Reset query planner statistics
SELECT pg_stat_reset();

-- If still slow, check for missing indexes
```

### 3. Disk Space Full

**Symptoms:**
```
ERROR: could not extend file: No space left on device
```

**Diagnosis:**
```sql
-- Find largest tables
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 10;

-- Find largest databases
SELECT datname, pg_size_pretty(pg_database_size(datname))
FROM pg_database
ORDER BY pg_database_size(datname) DESC;

-- Check WAL disk usage
SELECT pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0'));
```

**Solution:**
```bash
# Clean old WAL files
find /var/lib/postgresql/14/main/pg_wal -name "0*" -mtime +7 -delete

# Archive old logs
find /var/log/postgresql -name "*.log" -mtime +30 -exec gzip {} \;

# VACUUM FULL to reclaim space (requires downtime)
VACUUM FULL;

# Archive old data (application-specific)
DELETE FROM audit_logs WHERE created_at < NOW() - interval '90 days';
```

### 4. Replication Lag

**Symptoms:**
- Stale data on read replicas
- High lag in monitoring

**Diagnosis:**
```sql
-- On primary
SELECT client_addr, state,
       pg_wal_lsn_diff(pg_current_wal_lsn(), sent_lsn) AS send_lag_bytes,
       pg_wal_lsn_diff(sent_lsn, flush_lsn) AS flush_lag_bytes
FROM pg_stat_replication;

-- On replica
SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;
```

**Solution:**
```sql
-- On primary: increase WAL retention
ALTER SYSTEM SET wal_keep_size = '2GB';

-- Check network bandwidth between primary and replica
-- Check replica resource utilization (CPU, disk I/O)

-- If lag persists, rebuild replica
pg_basebackup -h primary-host -D /var/lib/postgresql/14/main -R -P
```

### 5. Deadlocks

**Symptoms:**
```
ERROR: deadlock detected
DETAIL: Process 1234 waits for ShareLock on transaction 5678
```

**Diagnosis:**
```sql
-- Enable deadlock logging
log_lock_waits = on
deadlock_timeout = 1s

-- View recent deadlocks in logs
grep "deadlock detected" /var/log/postgresql/postgresql-14-main.log

-- Monitor locks
SELECT locktype, relation::regclass, mode, granted, pid
FROM pg_locks
WHERE NOT granted;
```

**Solution:**
```go
// Always acquire locks in consistent order
// Example: always lock users before posts

// Use appropriate transaction isolation
db.SerializableTransaction(ctx, func(tx *sql.Tx) error {
    // Critical operations
    return nil
})

// Keep transactions short
// Avoid user input inside transactions
```

## Performance Monitoring Queries

```sql
-- Database size growth
SELECT datname,
       pg_size_pretty(pg_database_size(datname)) AS size,
       (SELECT pg_size_pretty(pg_database_size(datname))
        FROM pg_database
        WHERE datname = 'template0') AS baseline
FROM pg_database
ORDER BY pg_database_size(datname) DESC;

-- Cache hit ratio (should be > 99%)
SELECT
    sum(blks_hit) * 100.0 / sum(blks_hit + blks_read) AS cache_hit_ratio
FROM pg_stat_database;

-- Index hit ratio (should be > 99%)
SELECT
    sum(idx_blks_hit) * 100.0 / sum(idx_blks_hit + idx_blks_read) AS index_hit_ratio
FROM pg_statio_user_indexes;

-- Transaction throughput
SELECT datname,
       xact_commit,
       xact_rollback,
       xact_commit::float / (xact_commit + xact_rollback) AS commit_ratio
FROM pg_stat_database
WHERE datname = current_database();
```

## Emergency Procedures

### Read-Only Mode (Maintenance Window)

```sql
-- Enable read-only mode
ALTER DATABASE your_database SET default_transaction_read_only = on;

-- Disconnect active sessions
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'your_database' AND pid <> pg_backend_pid();

-- Perform maintenance...

-- Disable read-only mode
ALTER DATABASE your_database SET default_transaction_read_only = off;
```

### Kill All Connections

```sql
-- Terminate all connections except current
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE datname = 'your_database'
  AND pid <> pg_backend_pid();
```

## Best Practices Summary

1. **Connection Management**
   - Use connection pooling
   - Set appropriate timeouts
   - Monitor pool utilization

2. **Query Performance**
   - Always use EXPLAIN ANALYZE
   - Add indexes for common queries
   - Avoid SELECT *

3. **Maintenance**
   - Let autovacuum run
   - Monitor table bloat
   - Rebuild indexes regularly

4. **Monitoring**
   - Track slow queries
   - Monitor replication lag
   - Alert on connection pool exhaustion

5. **Resilience**
   - Use transactions appropriately
   - Implement retry logic
   - Configure circuit breakers
