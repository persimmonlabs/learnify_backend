#!/bin/bash
set -euo pipefail

# Database migration script
# Usage: ./migrate.sh <environment> <action> [steps]

ENVIRONMENT="${1:-development}"
ACTION="${2:-up}"
STEPS="${3:-1}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="${PROJECT_ROOT}/db/migrations"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Load environment configuration
load_config() {
    local env_file="${PROJECT_ROOT}/config/${ENVIRONMENT}.env"

    if [ ! -f "$env_file" ]; then
        log_error "Configuration file not found: $env_file"
        exit 1
    fi

    log_info "Loading configuration for ${ENVIRONMENT}"
    # shellcheck disable=SC1090
    source "$env_file"

    if [ -z "${DATABASE_URL:-}" ]; then
        log_error "DATABASE_URL not set in configuration"
        exit 1
    fi
}

# Check if migrate tool is installed
check_migrate_tool() {
    if ! command -v migrate &> /dev/null; then
        log_warn "migrate tool not found, installing..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    fi
}

# Create backup before migration
create_backup() {
    if [ "${ENVIRONMENT}" = "production" ]; then
        log_info "Creating database backup..."
        local backup_file="backup_${ENVIRONMENT}_$(date +%Y%m%d_%H%M%S).sql"
        local backup_dir="${PROJECT_ROOT}/backups"

        mkdir -p "$backup_dir"

        # Extract database name from URL
        local db_name=$(echo "$DATABASE_URL" | sed -n 's/.*\/\([^?]*\).*/\1/p')

        pg_dump "$DATABASE_URL" > "${backup_dir}/${backup_file}" || {
            log_warn "Backup failed, but continuing..."
        }

        log_info "Backup created: ${backup_dir}/${backup_file}"
    fi
}

# Get current migration version
get_current_version() {
    local version
    version=$(migrate -database "$DATABASE_URL" -path "$MIGRATIONS_DIR" version 2>&1 || echo "none")
    echo "$version"
}

# Validate migration files
validate_migrations() {
    log_info "Validating migration files..."

    if [ ! -d "$MIGRATIONS_DIR" ]; then
        log_error "Migrations directory not found: $MIGRATIONS_DIR"
        exit 1
    fi

    local file_count
    file_count=$(find "$MIGRATIONS_DIR" -name "*.sql" | wc -l)

    if [ "$file_count" -eq 0 ]; then
        log_warn "No migration files found"
        return
    fi

    # Check naming convention
    for file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$file" ]; then
            local filename
            filename=$(basename "$file")

            if ! [[ "$filename" =~ ^[0-9]{14}_[a-z_]+\.(up|down)\.sql$ ]]; then
                log_error "Invalid migration filename: $filename"
                log_error "Expected format: YYYYMMDDHHMMSS_description.(up|down).sql"
                exit 1
            fi
        fi
    done

    log_info "Migration files validated successfully"
}

# Run migration up
migrate_up() {
    log_info "Running migrations up..."
    local current_version
    current_version=$(get_current_version)
    log_info "Current version: $current_version"

    migrate -database "$DATABASE_URL" -path "$MIGRATIONS_DIR" up || {
        log_error "Migration up failed"
        exit 1
    }

    local new_version
    new_version=$(get_current_version)
    log_info "New version: $new_version"
    log_info "Migrations completed successfully"
}

# Run migration down
migrate_down() {
    log_info "Rolling back $STEPS migration(s)..."
    local current_version
    current_version=$(get_current_version)
    log_info "Current version: $current_version"

    # Confirm if production
    if [ "${ENVIRONMENT}" = "production" ]; then
        log_warn "You are about to rollback migrations in PRODUCTION"
        read -p "Are you sure? (yes/no): " -r
        if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
            log_info "Migration rollback cancelled"
            exit 0
        fi
    fi

    migrate -database "$DATABASE_URL" -path "$MIGRATIONS_DIR" down "$STEPS" || {
        log_error "Migration down failed"
        exit 1
    }

    local new_version
    new_version=$(get_current_version)
    log_info "New version: $new_version"
    log_info "Rollback completed successfully"
}

# Show migration status
show_status() {
    log_info "Migration status for ${ENVIRONMENT}:"
    local version
    version=$(get_current_version)

    echo "Current version: $version"
    echo ""
    echo "Available migrations:"

    if [ -d "$MIGRATIONS_DIR" ]; then
        ls -1 "$MIGRATIONS_DIR" | grep ".up.sql$" | sed 's/.up.sql$//'
    fi
}

# Create new migration
create_migration() {
    local name="$1"

    if [ -z "$name" ]; then
        log_error "Migration name required"
        echo "Usage: $0 <environment> create <migration_name>"
        exit 1
    fi

    log_info "Creating migration: $name"

    local timestamp
    timestamp=$(date +%Y%m%d%H%M%S)
    local up_file="${MIGRATIONS_DIR}/${timestamp}_${name}.up.sql"
    local down_file="${MIGRATIONS_DIR}/${timestamp}_${name}.down.sql"

    mkdir -p "$MIGRATIONS_DIR"

    cat > "$up_file" <<EOF
-- Migration: $name
-- Created at: $(date)

-- Add your SQL statements here

EOF

    cat > "$down_file" <<EOF
-- Rollback migration: $name
-- Created at: $(date)

-- Add your rollback SQL statements here

EOF

    log_info "Migration files created:"
    log_info "  Up:   $up_file"
    log_info "  Down: $down_file"
}

# Verify database connection
verify_connection() {
    log_info "Verifying database connection..."

    if command -v psql &> /dev/null; then
        psql "$DATABASE_URL" -c "SELECT 1;" &> /dev/null || {
            log_error "Failed to connect to database"
            exit 1
        }
        log_info "Database connection successful"
    else
        log_warn "psql not found, skipping connection verification"
    fi
}

# Main function
main() {
    log_info "Database migration tool"
    log_info "Environment: ${ENVIRONMENT}"
    log_info "Action: ${ACTION}"

    # Load configuration
    load_config

    # Check tools
    check_migrate_tool

    # Verify connection
    verify_connection

    # Validate migrations
    validate_migrations

    # Execute action
    case "$ACTION" in
        up)
            create_backup
            migrate_up
            ;;
        down)
            create_backup
            migrate_down
            ;;
        status)
            show_status
            ;;
        create)
            create_migration "$STEPS"
            ;;
        *)
            log_error "Unknown action: $ACTION"
            echo "Available actions: up, down, status, create"
            exit 1
            ;;
    esac

    log_info "Operation completed successfully"
}

# Run main function
main
