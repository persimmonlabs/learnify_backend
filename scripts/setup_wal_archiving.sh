#!/bin/bash
#
# PostgreSQL WAL Archiving Setup Script
# Configures continuous WAL archiving for point-in-time recovery
#
# Usage: sudo ./setup_wal_archiving.sh
#

set -e

# Configuration
PG_VERSION="${PG_VERSION:-14}"
PG_DATA_DIR="${PG_DATA_DIR:-/var/lib/postgresql/$PG_VERSION/main}"
WAL_ARCHIVE_DIR="${WAL_ARCHIVE_DIR:-/var/lib/postgresql/wal_archive}"
S3_BUCKET="${S3_BUCKET:-}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (use sudo)"
    exit 1
fi

log_info "PostgreSQL WAL Archiving Setup"
log_info "=============================="

# Create WAL archive directory
log_info "Creating WAL archive directory: $WAL_ARCHIVE_DIR"
mkdir -p "$WAL_ARCHIVE_DIR"
chown postgres:postgres "$WAL_ARCHIVE_DIR"
chmod 700 "$WAL_ARCHIVE_DIR"

# Create archive script
ARCHIVE_SCRIPT="/usr/local/bin/archive_wal.sh"
log_info "Creating WAL archive script: $ARCHIVE_SCRIPT"

cat > "$ARCHIVE_SCRIPT" << 'EOF'
#!/bin/bash
# WAL Archive Script
# Called by PostgreSQL for each WAL segment

WAL_PATH="$1"
WAL_FILE="$2"
ARCHIVE_DIR="/var/lib/postgresql/wal_archive"
S3_BUCKET="__S3_BUCKET__"

# Copy to local archive
cp "$WAL_PATH" "$ARCHIVE_DIR/$WAL_FILE"

# Upload to S3 if configured
if [ -n "$S3_BUCKET" ] && command -v aws &> /dev/null; then
    aws s3 cp "$WAL_PATH" "s3://$S3_BUCKET/wal/$WAL_FILE" \
        --storage-class STANDARD_IA
fi

# Exit with success
exit 0
EOF

# Replace placeholder with actual S3 bucket
sed -i "s|__S3_BUCKET__|$S3_BUCKET|g" "$ARCHIVE_SCRIPT"

chmod +x "$ARCHIVE_SCRIPT"
chown postgres:postgres "$ARCHIVE_SCRIPT"

# Create WAL cleanup script
CLEANUP_SCRIPT="/usr/local/bin/cleanup_wal_archive.sh"
log_info "Creating WAL cleanup script: $CLEANUP_SCRIPT"

cat > "$CLEANUP_SCRIPT" << EOF
#!/bin/bash
# WAL Archive Cleanup Script
# Removes WAL files older than retention period

ARCHIVE_DIR="$WAL_ARCHIVE_DIR"
RETENTION_DAYS=$RETENTION_DAYS

# Delete old WAL files
find "\$ARCHIVE_DIR" -name "*.wal" -type f -mtime +\$RETENTION_DAYS -delete

# Log cleanup
echo "\$(date): Cleaned WAL archives older than \$RETENTION_DAYS days"
EOF

chmod +x "$CLEANUP_SCRIPT"
chown postgres:postgres "$CLEANUP_SCRIPT"

# Configure PostgreSQL
PG_CONF="$PG_DATA_DIR/postgresql.conf"
log_info "Configuring PostgreSQL: $PG_CONF"

# Backup original config
cp "$PG_CONF" "$PG_CONF.backup.$(date +%Y%m%d_%H%M%S)"

# Add or update WAL archiving settings
log_info "Updating PostgreSQL configuration..."

# Remove existing WAL settings if present
sed -i '/^wal_level/d' "$PG_CONF"
sed -i '/^archive_mode/d' "$PG_CONF"
sed -i '/^archive_command/d' "$PG_CONF"
sed -i '/^archive_timeout/d' "$PG_CONF"
sed -i '/^max_wal_size/d' "$PG_CONF"
sed -i '/^min_wal_size/d' "$PG_CONF"
sed -i '/^wal_keep_size/d' "$PG_CONF"

# Add new settings
cat >> "$PG_CONF" << EOF

# WAL Archiving Configuration
# Added by setup_wal_archiving.sh on $(date)
wal_level = replica
archive_mode = on
archive_command = '$ARCHIVE_SCRIPT %p %f'
archive_timeout = 300

# WAL Settings
max_wal_size = 2GB
min_wal_size = 1GB
wal_keep_size = 512MB
EOF

log_info "PostgreSQL configuration updated"

# Setup cron job for cleanup
log_info "Setting up daily WAL cleanup cron job..."
CRON_JOB="0 2 * * * $CLEANUP_SCRIPT >> /var/log/postgresql/wal_cleanup.log 2>&1"

# Add to postgres user crontab
(crontab -u postgres -l 2>/dev/null | grep -v "$CLEANUP_SCRIPT"; echo "$CRON_JOB") | crontab -u postgres -

log_info "Cron job configured: Daily cleanup at 2:00 AM"

# Create restore script
RESTORE_SCRIPT="/usr/local/bin/restore_wal.sh"
log_info "Creating WAL restore script: $RESTORE_SCRIPT"

cat > "$RESTORE_SCRIPT" << EOF
#!/bin/bash
# WAL Restore Script
# Used during PITR recovery

WAL_FILE=\$1
WAL_DEST=\$2
ARCHIVE_DIR="$WAL_ARCHIVE_DIR"
S3_BUCKET="$S3_BUCKET"

# Try local archive first
if [ -f "\$ARCHIVE_DIR/\$WAL_FILE" ]; then
    cp "\$ARCHIVE_DIR/\$WAL_FILE" "\$WAL_DEST"
    exit 0
fi

# Try S3 if configured
if [ -n "\$S3_BUCKET" ] && command -v aws &> /dev/null; then
    aws s3 cp "s3://\$S3_BUCKET/wal/\$WAL_FILE" "\$WAL_DEST"
    exit \$?
fi

# WAL file not found
exit 1
EOF

chmod +x "$RESTORE_SCRIPT"
chown postgres:postgres "$RESTORE_SCRIPT"

# Test archive command
log_info "Testing archive configuration..."
sudo -u postgres psql -c "SELECT pg_switch_wal()" > /dev/null 2>&1 || true
sleep 2

# Check if WAL files are being archived
WAL_COUNT=$(ls -1 "$WAL_ARCHIVE_DIR" 2>/dev/null | wc -l)
if [ "$WAL_COUNT" -gt 0 ]; then
    log_info "WAL archiving is working ($WAL_COUNT files archived)"
else
    log_warn "No WAL files found in archive yet. Check logs after PostgreSQL restart."
fi

# Display configuration summary
cat << EOF

WAL Archiving Setup Complete!
==============================

Configuration:
  Archive Directory: $WAL_ARCHIVE_DIR
  Archive Script: $ARCHIVE_SCRIPT
  Cleanup Script: $CLEANUP_SCRIPT
  Restore Script: $RESTORE_SCRIPT
  Retention Period: $RETENTION_DAYS days
  S3 Bucket: ${S3_BUCKET:-Not configured}

Next Steps:
  1. Restart PostgreSQL:
     sudo systemctl restart postgresql

  2. Verify archiving is working:
     sudo -u postgres psql -c "SELECT pg_switch_wal();"
     ls -lh $WAL_ARCHIVE_DIR

  3. Monitor WAL archiving:
     tail -f /var/log/postgresql/postgresql-$PG_VERSION-main.log | grep archive

  4. Test restore command:
     $RESTORE_SCRIPT <wal_filename> /tmp/test_restore

For Point-in-Time Recovery, see:
  /path/to/docs/database-disaster-recovery.md

EOF

log_warn "PostgreSQL restart required for changes to take effect"
log_info "Run: sudo systemctl restart postgresql"

exit 0
