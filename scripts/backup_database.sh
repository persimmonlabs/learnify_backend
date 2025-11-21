#!/bin/bash
#
# PostgreSQL Automated Backup Script
# Performs full database backup with compression and S3 upload
#
# Usage: ./backup_database.sh [database_name]
#

set -e  # Exit on error
set -u  # Exit on undefined variable

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${1:-${DB_NAME:-myapp}}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/postgresql}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
S3_BUCKET="${S3_BUCKET:-}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.dump"
LOG_FILE="${BACKUP_DIR}/backup_${TIMESTAMP}.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

log_info "Starting database backup for: $DB_NAME"
log_info "Backup file: $BACKUP_FILE"

# Check if PostgreSQL is accessible
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
    log_error "PostgreSQL is not accessible at $DB_HOST:$DB_PORT"
    exit 1
fi

log_info "PostgreSQL connection verified"

# Pre-backup database statistics
DB_SIZE=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT pg_size_pretty(pg_database_size('$DB_NAME'))")
log_info "Database size: $DB_SIZE"

# Perform backup with pg_dump
log_info "Starting pg_dump..."
START_TIME=$(date +%s)

if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    --format=custom \
    --compress=9 \
    --verbose \
    --file="$BACKUP_FILE" 2>&1 | tee -a "$LOG_FILE"; then

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)

    log_info "Backup completed successfully in ${DURATION}s"
    log_info "Backup file size: $BACKUP_SIZE"
else
    log_error "Backup failed"
    exit 1
fi

# Verify backup integrity
log_info "Verifying backup integrity..."
if pg_restore --list "$BACKUP_FILE" > /dev/null 2>&1; then
    log_info "Backup integrity verified"
else
    log_error "Backup integrity check failed"
    exit 1
fi

# Calculate checksums
MD5SUM=$(md5sum "$BACKUP_FILE" | awk '{print $1}')
echo "$MD5SUM" > "${BACKUP_FILE}.md5"
log_info "MD5 checksum: $MD5SUM"

# Upload to S3 if configured
if [ -n "$S3_BUCKET" ]; then
    log_info "Uploading backup to S3: $S3_BUCKET"

    if command -v aws &> /dev/null; then
        if aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/backups/$(basename $BACKUP_FILE)" \
            --storage-class STANDARD_IA \
            --metadata "md5=$MD5SUM,database=$DB_NAME,timestamp=$TIMESTAMP"; then

            log_info "Backup uploaded to S3 successfully"

            # Upload checksum file
            aws s3 cp "${BACKUP_FILE}.md5" "s3://${S3_BUCKET}/backups/$(basename ${BACKUP_FILE}.md5)"
        else
            log_warn "Failed to upload backup to S3"
        fi
    else
        log_warn "AWS CLI not found, skipping S3 upload"
    fi
fi

# Clean up old backups
log_info "Cleaning up backups older than $RETENTION_DAYS days..."
DELETED_COUNT=$(find "$BACKUP_DIR" -name "${DB_NAME}_*.dump" -type f -mtime +$RETENTION_DAYS -delete -print | wc -l)
log_info "Deleted $DELETED_COUNT old backup(s)"

# Clean up old checksum files
find "$BACKUP_DIR" -name "${DB_NAME}_*.md5" -type f -mtime +$RETENTION_DAYS -delete

# Clean up old log files
find "$BACKUP_DIR" -name "backup_*.log" -type f -mtime +$RETENTION_DAYS -delete

# Generate backup report
cat > "${BACKUP_DIR}/latest_backup_report.txt" << EOF
Database Backup Report
======================
Date: $(date)
Database: $DB_NAME
Host: $DB_HOST:$DB_PORT
Database Size: $DB_SIZE
Backup File: $(basename $BACKUP_FILE)
Backup Size: $BACKUP_SIZE
Duration: ${DURATION}s
MD5 Checksum: $MD5SUM
S3 Upload: $([ -n "$S3_BUCKET" ] && echo "Yes" || echo "No")
Status: SUCCESS
EOF

log_info "Backup report generated: ${BACKUP_DIR}/latest_backup_report.txt"
log_info "Backup process completed successfully"

# Send notification (optional - configure your notification method)
# Examples:
# - Email: mail -s "Database Backup Success: $DB_NAME" admin@example.com < "${BACKUP_DIR}/latest_backup_report.txt"
# - Slack: curl -X POST -H 'Content-type: application/json' --data '{"text":"Backup completed"}' YOUR_SLACK_WEBHOOK
# - PagerDuty: curl -X POST https://events.pagerduty.com/v2/enqueue ...

exit 0
