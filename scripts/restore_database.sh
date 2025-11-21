#!/bin/bash
#
# PostgreSQL Database Restore Script
# Restores database from backup file with safety checks
#
# Usage: ./restore_database.sh <backup_file> [target_database]
#

set -e  # Exit on error
set -u  # Exit on undefined variable

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
BACKUP_FILE="${1:-}"
TARGET_DB="${2:-${DB_NAME:-myapp}}"
PARALLEL_JOBS="${PARALLEL_JOBS:-4}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_prompt() {
    echo -e "${BLUE}[PROMPT]${NC} $1"
}

# Usage information
usage() {
    cat << EOF
Usage: $0 <backup_file> [target_database]

Arguments:
  backup_file       Path to the backup file (.dump)
  target_database   Name of target database (default: $TARGET_DB)

Environment Variables:
  DB_HOST          Database host (default: localhost)
  DB_PORT          Database port (default: 5432)
  DB_USER          Database user (default: postgres)
  PARALLEL_JOBS    Number of parallel restore jobs (default: 4)

Examples:
  $0 /backups/myapp_20250121.dump
  $0 /backups/myapp_20250121.dump myapp_restore
  DB_HOST=db.example.com $0 /backups/myapp_20250121.dump

EOF
    exit 1
}

# Validate arguments
if [ -z "$BACKUP_FILE" ]; then
    log_error "Backup file not specified"
    usage
fi

if [ ! -f "$BACKUP_FILE" ]; then
    log_error "Backup file not found: $BACKUP_FILE"
    exit 1
fi

log_info "Database Restore Utility"
log_info "========================"
log_info "Backup file: $BACKUP_FILE"
log_info "Target database: $TARGET_DB"
log_info "Host: $DB_HOST:$DB_PORT"

# Check if PostgreSQL is accessible
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
    log_error "PostgreSQL is not accessible at $DB_HOST:$DB_PORT"
    exit 1
fi

log_info "PostgreSQL connection verified"

# Verify backup integrity
log_info "Verifying backup integrity..."
if ! pg_restore --list "$BACKUP_FILE" > /dev/null 2>&1; then
    log_error "Backup file is corrupted or invalid"
    exit 1
fi

log_info "Backup integrity verified"

# Check MD5 checksum if available
if [ -f "${BACKUP_FILE}.md5" ]; then
    log_info "Checking MD5 checksum..."
    EXPECTED_MD5=$(cat "${BACKUP_FILE}.md5")
    ACTUAL_MD5=$(md5sum "$BACKUP_FILE" | awk '{print $1}')

    if [ "$EXPECTED_MD5" == "$ACTUAL_MD5" ]; then
        log_info "MD5 checksum verified"
    else
        log_error "MD5 checksum mismatch!"
        log_error "Expected: $EXPECTED_MD5"
        log_error "Actual: $ACTUAL_MD5"
        exit 1
    fi
fi

# Get backup information
BACKUP_INFO=$(pg_restore --list "$BACKUP_FILE" | head -20)
log_info "Backup contains:"
echo "$BACKUP_INFO" | grep "DATABASE\|TABLE" | head -10

# Check if target database exists
DB_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -t -c \
    "SELECT 1 FROM pg_database WHERE datname='$TARGET_DB'" | tr -d '[:space:]')

if [ "$DB_EXISTS" == "1" ]; then
    log_warn "Database '$TARGET_DB' already exists!"
    log_prompt "This operation will DROP the existing database. Are you sure? (yes/no): "
    read -r CONFIRM

    if [ "$CONFIRM" != "yes" ]; then
        log_info "Restore cancelled by user"
        exit 0
    fi

    # Check for active connections
    ACTIVE_CONNS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -t -c \
        "SELECT COUNT(*) FROM pg_stat_activity WHERE datname='$TARGET_DB' AND pid <> pg_backend_pid()")

    if [ "$ACTIVE_CONNS" -gt 0 ]; then
        log_warn "Found $ACTIVE_CONNS active connection(s) to database"
        log_prompt "Terminate active connections? (yes/no): "
        read -r TERMINATE

        if [ "$TERMINATE" == "yes" ]; then
            log_info "Terminating active connections..."
            psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c \
                "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='$TARGET_DB' AND pid <> pg_backend_pid()"
            sleep 2
        else
            log_error "Cannot proceed with active connections"
            exit 1
        fi
    fi

    # Drop existing database
    log_info "Dropping existing database: $TARGET_DB"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "DROP DATABASE \"$TARGET_DB\""
fi

# Create new database
log_info "Creating database: $TARGET_DB"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE \"$TARGET_DB\" OWNER $DB_USER"

# Restore database
log_info "Starting database restore with $PARALLEL_JOBS parallel jobs..."
START_TIME=$(date +%s)

if pg_restore -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$TARGET_DB" \
    --verbose \
    --jobs="$PARALLEL_JOBS" \
    --no-owner \
    --no-privileges \
    "$BACKUP_FILE" 2>&1 | tee restore_output.log; then

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    log_info "Restore completed in ${DURATION}s"
else
    log_warn "Restore completed with warnings (check restore_output.log)"
fi

# Verify restore
log_info "Verifying restore..."

# Check table counts
TABLE_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$TARGET_DB" -t -c \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'")

log_info "Restored tables: $TABLE_COUNT"

# Display table statistics
log_info "Table statistics:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$TARGET_DB" -c \
    "SELECT schemaname, tablename,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
     FROM pg_tables
     WHERE schemaname='public'
     ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
     LIMIT 10"

# Analyze database to update statistics
log_info "Analyzing database to update query planner statistics..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$TARGET_DB" -c "ANALYZE"

# Generate restore report
cat > restore_report.txt << EOF
Database Restore Report
=======================
Date: $(date)
Backup File: $BACKUP_FILE
Target Database: $TARGET_DB
Host: $DB_HOST:$DB_PORT
Duration: ${DURATION}s
Tables Restored: $TABLE_COUNT
Parallel Jobs: $PARALLEL_JOBS
Status: SUCCESS

Verification:
$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$TARGET_DB" -t -c \
  "SELECT COUNT(*) || ' total tables' FROM information_schema.tables WHERE table_schema='public'")

Next Steps:
1. Verify critical data integrity
2. Update application configuration to point to restored database
3. Run application smoke tests
4. Monitor database performance
EOF

cat restore_report.txt

log_info "Restore report saved: restore_report.txt"
log_info "Restore completed successfully!"

# Clean up
rm -f restore_output.log

exit 0
