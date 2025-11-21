# Database Disaster Recovery Plan

## Overview

This document outlines the disaster recovery procedures for the PostgreSQL database, including backup strategies, restore procedures, and failover mechanisms.

## Recovery Objectives

- **RTO (Recovery Time Objective):** 1 hour
- **RPO (Recovery Point Objective):** 15 minutes
- **Availability Target:** 99.9% uptime

## Backup Strategy

### 1. Full Database Backups (pg_dump)

**Frequency:** Daily at 2:00 AM UTC
**Retention:** 30 days
**Storage:** S3-compatible object storage with versioning

```bash
# Automated daily backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME \
  --format=custom \
  --compress=9 \
  --file=/backups/$(date +%Y%m%d)_full_backup.dump

# Verify backup integrity
pg_restore --list /backups/$(date +%Y%m%d)_full_backup.dump
```

### 2. WAL (Write-Ahead Log) Archiving

**Purpose:** Point-in-time recovery (PITR)
**Frequency:** Continuous (every WAL segment)
**Retention:** 7 days

**PostgreSQL Configuration (`postgresql.conf`):**
```ini
# Enable WAL archiving
wal_level = replica
archive_mode = on
archive_command = 'cp %p /var/lib/postgresql/wal_archive/%f'
archive_timeout = 300  # Force archive every 5 minutes

# WAL settings
max_wal_size = 2GB
min_wal_size = 1GB
wal_keep_size = 512MB
```

**Archive Script:**
```bash
#!/bin/bash
# scripts/archive_wal.sh
WAL_FILE=$1
ARCHIVE_DIR="/var/lib/postgresql/wal_archive"
S3_BUCKET="s3://your-backup-bucket/wal"

# Copy to local archive
cp "$WAL_FILE" "$ARCHIVE_DIR/"

# Upload to S3
aws s3 cp "$WAL_FILE" "$S3_BUCKET/$(basename $WAL_FILE)"

# Clean old WAL files (older than 7 days)
find "$ARCHIVE_DIR" -name "*.wal" -mtime +7 -delete
```

### 3. Incremental Backups (pg_basebackup)

**Frequency:** Every 6 hours
**Retention:** 48 hours

```bash
# Incremental backup
pg_basebackup -h $DB_HOST -U $DB_USER \
  --pgdata=/backups/basebackup_$(date +%Y%m%d_%H%M) \
  --format=tar \
  --gzip \
  --checkpoint=fast \
  --progress
```

### 4. Replication Slots

**Purpose:** Ensure WAL files aren't deleted before standby receives them

```sql
-- Create replication slot
SELECT pg_create_physical_replication_slot('standby_slot');

-- Monitor replication lag
SELECT slot_name, active,
       pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) AS lag_bytes
FROM pg_replication_slots;
```

## Restore Procedures

### Full Database Restore (from pg_dump)

**Use Case:** Complete database loss or corruption
**Estimated Time:** 30-60 minutes (depends on database size)

```bash
# 1. Stop application servers
systemctl stop backend-api

# 2. Drop existing database (if exists)
psql -h $DB_HOST -U postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"

# 3. Create fresh database
psql -h $DB_HOST -U postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"

# 4. Restore from backup
pg_restore -h $DB_HOST -U $DB_USER -d $DB_NAME \
  --verbose \
  --jobs=4 \
  /backups/20250121_full_backup.dump

# 5. Verify restore
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt"
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM users;"

# 6. Restart application
systemctl start backend-api
```

### Point-in-Time Recovery (PITR)

**Use Case:** Recover to specific timestamp before data corruption
**Estimated Time:** 45-90 minutes

```bash
# 1. Stop PostgreSQL
systemctl stop postgresql

# 2. Backup current data directory
mv /var/lib/postgresql/14/main /var/lib/postgresql/14/main.old

# 3. Restore base backup
mkdir /var/lib/postgresql/14/main
tar -xzf /backups/basebackup_20250121_0600.tar.gz -C /var/lib/postgresql/14/main

# 4. Create recovery configuration
cat > /var/lib/postgresql/14/main/recovery.conf << EOF
restore_command = 'cp /var/lib/postgresql/wal_archive/%f %p'
recovery_target_time = '2025-01-21 14:30:00'
recovery_target_action = 'promote'
EOF

# 5. Start PostgreSQL (will enter recovery mode)
systemctl start postgresql

# 6. Monitor recovery progress
tail -f /var/log/postgresql/postgresql-14-main.log

# 7. Verify recovery
psql -U postgres -c "SELECT pg_is_in_recovery();"  # Should return false when complete
```

### Table-Level Restore

**Use Case:** Single table corruption or accidental deletion
**Estimated Time:** 5-15 minutes

```bash
# 1. Extract specific table from backup
pg_restore -h $DB_HOST -U $DB_USER -d $DB_NAME \
  --table=users \
  --clean \
  --if-exists \
  /backups/20250121_full_backup.dump

# 2. Verify table data
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM users;"
```

## Failover Procedures

### Primary Database Failure

**Automatic Failover (with Patroni/repmgr):**

1. Health check detects primary failure
2. Standby automatically promoted to primary
3. Application connection strings updated via DNS/load balancer
4. Old primary rejoins as standby when recovered

**Manual Failover:**

```bash
# 1. Promote standby to primary
pg_ctl promote -D /var/lib/postgresql/14/main

# Or using SQL
SELECT pg_promote();

# 2. Update application configuration
# Point DATABASE_URL to new primary IP

# 3. Verify new primary
psql -h <new-primary-ip> -U postgres -c "SELECT pg_is_in_recovery();"
# Should return false

# 4. Configure old primary as standby (when available)
# Create standby.signal file
touch /var/lib/postgresql/14/main/standby.signal

# Update postgresql.conf
primary_conninfo = 'host=<new-primary-ip> port=5432 user=replicator'

# Restart old primary
systemctl restart postgresql
```

### Network Partition (Split Brain Prevention)

**Detection:**
```bash
# Check replication status on all nodes
psql -c "SELECT * FROM pg_stat_replication;"

# Check for multiple primaries (should only be one)
for host in primary-1 primary-2 standby-1; do
  echo "Checking $host..."
  psql -h $host -c "SELECT pg_is_in_recovery();"
done
```

**Resolution:**
1. Identify true primary (most recent LSN)
2. Fence old primary (block network access)
3. Rebuild old primary as standby

### Replication Lag Recovery

**Monitoring:**
```sql
SELECT client_addr,
       state,
       pg_wal_lsn_diff(pg_current_wal_lsn(), sent_lsn) AS send_lag,
       pg_wal_lsn_diff(sent_lsn, flush_lsn) AS flush_lag,
       pg_wal_lsn_diff(flush_lsn, replay_lsn) AS replay_lag
FROM pg_stat_replication;
```

**Remediation:**
```bash
# If lag > 100MB, consider:
# 1. Increase wal_keep_size
ALTER SYSTEM SET wal_keep_size = '1GB';

# 2. Adjust network bandwidth
# 3. Check standby resource utilization

# If lag is excessive (> 1GB), rebuild standby
pg_basebackup -h primary -D /var/lib/postgresql/14/main -P -R
```

## Backup Verification

**Weekly Backup Tests (Every Sunday):**

```bash
#!/bin/bash
# scripts/verify_backup.sh

BACKUP_FILE="/backups/latest_full_backup.dump"
TEST_DB="backup_test_$(date +%s)"

# Create test database
psql -U postgres -c "CREATE DATABASE $TEST_DB;"

# Restore backup
pg_restore -U postgres -d $TEST_DB $BACKUP_FILE

# Run verification queries
psql -U postgres -d $TEST_DB << EOF
-- Check table counts
SELECT schemaname, tablename,
       (xpath('/row/cnt/text()',
              query_to_xml('SELECT COUNT(*) as cnt FROM '||schemaname||'.'||tablename, false, true, ''))
       )[1]::text::int AS row_count
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;

-- Verify critical data
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM posts;
EOF

# Cleanup
psql -U postgres -c "DROP DATABASE $TEST_DB;"

echo "Backup verification complete: $(date)"
```

## Monitoring and Alerts

**Critical Alerts:**
- Backup failure (notify within 15 minutes)
- Replication lag > 100MB (notify immediately)
- Standby disconnected > 5 minutes (notify immediately)
- WAL archive failure (notify within 5 minutes)

**Monitoring Queries:**
```sql
-- Backup age check
SELECT NOW() - MAX(applied_at) AS last_backup_age
FROM schema_migrations;

-- Replication status
SELECT * FROM pg_stat_replication;

-- WAL disk usage
SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0') / (1024*1024) AS wal_mb;
```

## Testing Schedule

1. **Daily:** Automated backup verification
2. **Weekly:** Table-level restore test
3. **Monthly:** Full database restore test
4. **Quarterly:** Complete disaster recovery drill (with failover)

## Emergency Contacts

- **Database Team Lead:** [Contact Info]
- **DevOps On-Call:** [PagerDuty/On-Call System]
- **Cloud Provider Support:** [AWS/GCP Support Number]

## Runbook Quick Reference

| Scenario | Action | Command |
|----------|--------|---------|
| Full DB restore | Restore latest backup | `pg_restore -d $DB_NAME /backups/latest.dump` |
| PITR | Restore to timestamp | Configure `recovery_target_time` |
| Promote standby | Manual failover | `SELECT pg_promote();` |
| Check replication | Monitor lag | `SELECT * FROM pg_stat_replication;` |
| Rebuild standby | Resync from primary | `pg_basebackup -h primary -D /data -R` |

## Change Log

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-01-21 | 1.0 | Initial disaster recovery plan | Database Team |
