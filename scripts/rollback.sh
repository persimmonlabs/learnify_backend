#!/bin/bash
set -euo pipefail

# Rollback script for failed deployments
# Usage: ./rollback.sh <environment>

ENVIRONMENT="${1:-staging}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

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
}

# Get previous deployment version
get_previous_version() {
    log_info "Finding previous deployment..."

    local versions
    versions=$(docker images "${DOCKER_REGISTRY}/${DOCKER_IMAGE}" \
        --format "{{.Tag}}" | \
        grep -v "latest" | \
        sort -r)

    # Get second version (first is current, second is previous)
    local previous_version
    previous_version=$(echo "$versions" | sed -n '2p')

    if [ -z "$previous_version" ]; then
        log_error "No previous version found"
        exit 1
    fi

    echo "$previous_version"
}

# Confirm rollback
confirm_rollback() {
    local version="$1"

    log_warn "You are about to rollback ${ENVIRONMENT} to version ${version}"

    if [ "${ENVIRONMENT}" = "production" ]; then
        log_error "PRODUCTION ROLLBACK - This is a critical operation!"
        read -p "Type 'ROLLBACK' to confirm: " -r
        if [ "$REPLY" != "ROLLBACK" ]; then
            log_info "Rollback cancelled"
            exit 0
        fi
    else
        read -p "Continue? (yes/no): " -r
        if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
            log_info "Rollback cancelled"
            exit 0
        fi
    fi
}

# Create snapshot before rollback
create_snapshot() {
    log_info "Creating snapshot of current state..."

    local snapshot_dir="${PROJECT_ROOT}/snapshots/${ENVIRONMENT}"
    mkdir -p "$snapshot_dir"

    local timestamp
    timestamp=$(date +%Y%m%d_%H%M%S)
    local snapshot_file="${snapshot_dir}/snapshot_${timestamp}.txt"

    {
        echo "Rollback snapshot - ${timestamp}"
        echo "Environment: ${ENVIRONMENT}"
        echo ""
        echo "Current container:"
        docker ps --filter "name=${APP_NAME}" --format "{{.ID}}\t{{.Image}}\t{{.Status}}"
        echo ""
        echo "Container logs (last 100 lines):"
        docker logs "${APP_NAME}" --tail 100 2>&1 || true
    } > "$snapshot_file"

    log_info "Snapshot saved to: $snapshot_file"
}

# Backup database
backup_database() {
    log_info "Creating database backup before rollback..."

    local backup_dir="${PROJECT_ROOT}/backups"
    mkdir -p "$backup_dir"

    local backup_file="rollback_backup_${ENVIRONMENT}_$(date +%Y%m%d_%H%M%S).sql"

    if command -v pg_dump &> /dev/null; then
        pg_dump "$DATABASE_URL" > "${backup_dir}/${backup_file}" || {
            log_warn "Database backup failed, but continuing with rollback..."
        }
        log_info "Database backup: ${backup_dir}/${backup_file}"
    else
        log_warn "pg_dump not found, skipping database backup"
    fi
}

# Stop current deployment
stop_current_deployment() {
    log_info "Stopping current deployment..."

    if docker ps -q --filter "name=${APP_NAME}" | grep -q .; then
        docker stop "${APP_NAME}" || {
            log_error "Failed to stop current deployment"
            exit 1
        }
        docker rm "${APP_NAME}" || true
        log_info "Current deployment stopped"
    else
        log_warn "No running container found"
    fi
}

# Start previous version
start_previous_version() {
    local version="$1"
    local image="${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${version}"

    log_info "Starting previous version: ${version}"

    # Pull image if not available
    if ! docker images | grep -q "$image"; then
        log_info "Pulling image: $image"
        docker pull "$image" || {
            log_error "Failed to pull previous version"
            exit 1
        }
    fi

    # Start container with previous version
    docker run -d \
        --name "${APP_NAME}" \
        --restart unless-stopped \
        -p "${APP_PORT}:8080" \
        -e "DATABASE_URL=${DATABASE_URL}" \
        -e "REDIS_URL=${REDIS_URL}" \
        -e "ENV=${ENVIRONMENT}" \
        -e "LOG_LEVEL=${LOG_LEVEL:-info}" \
        --network "${DOCKER_NETWORK:-bridge}" \
        --health-cmd="curl -f http://localhost:8080/health || exit 1" \
        --health-interval=30s \
        --health-timeout=10s \
        --health-retries=3 \
        "$image" || {
            log_error "Failed to start previous version"
            exit 1
        }

    log_info "Previous version started"
}

# Wait for rollback to be healthy
wait_for_health() {
    log_info "Waiting for rolled-back application to be healthy..."

    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if docker inspect --format='{{.State.Health.Status}}' "${APP_NAME}" 2>/dev/null | grep -q "healthy"; then
            log_info "Rolled-back application is healthy"
            return 0
        fi

        attempt=$((attempt + 1))
        log_info "Health check attempt $attempt/$max_attempts..."
        sleep 2
    done

    log_error "Rolled-back application failed to become healthy"
    docker logs "${APP_NAME}" --tail 50
    return 1
}

# Rollback database migrations
rollback_migrations() {
    log_warn "Database migrations may need manual rollback"
    log_info "Check migration history and rollback if necessary"

    # Optionally run automatic rollback
    # "${SCRIPT_DIR}/migrate.sh" "$ENVIRONMENT" down 1
}

# Verify rollback
verify_rollback() {
    log_info "Verifying rollback..."

    # Run health checks
    if [ -f "${SCRIPT_DIR}/health-check.sh" ]; then
        "${SCRIPT_DIR}/health-check.sh" "$ENVIRONMENT" || {
            log_warn "Health checks failed after rollback"
            return 1
        }
    fi

    # Check container status
    local status
    status=$(docker inspect --format='{{.State.Status}}' "${APP_NAME}")

    if [ "$status" = "running" ]; then
        log_info "Container is running"
    else
        log_error "Container is not running (status: $status)"
        return 1
    fi

    return 0
}

# Send notification
send_notification() {
    local status="$1"
    local version="$2"

    if [ -z "${SLACK_WEBHOOK_URL:-}" ]; then
        return
    fi

    local color="warning"
    if [ "$status" = "success" ]; then
        color="good"
    elif [ "$status" = "failure" ]; then
        color="danger"
    fi

    curl -X POST "$SLACK_WEBHOOK_URL" \
        -H 'Content-Type: application/json' \
        -d "{
            \"attachments\": [{
                \"color\": \"$color\",
                \"title\": \"⚠️ Rollback Executed\",
                \"text\": \"Rollback to version ${version} completed\",
                \"fields\": [
                    {\"title\": \"Environment\", \"value\": \"${ENVIRONMENT}\", \"short\": true},
                    {\"title\": \"Version\", \"value\": \"${version}\", \"short\": true},
                    {\"title\": \"Status\", \"value\": \"${status}\", \"short\": true}
                ]
            }]
        }" || true
}

# Main rollback function
main() {
    log_info "Emergency Rollback Tool"
    log_info "Environment: ${ENVIRONMENT}"

    # Load configuration
    load_config

    # Get previous version
    local previous_version
    previous_version=$(get_previous_version)
    log_info "Previous version: ${previous_version}"

    # Confirm rollback
    confirm_rollback "$previous_version"

    # Create snapshot
    create_snapshot

    # Backup database
    backup_database

    # Perform rollback
    stop_current_deployment
    start_previous_version "$previous_version"

    # Wait for health
    if wait_for_health; then
        log_info "Rollback health check passed"
    else
        log_error "Rollback health check failed"
        send_notification "failure" "$previous_version"
        exit 1
    fi

    # Verify rollback
    if verify_rollback; then
        log_info "Rollback verification passed"
    else
        log_warn "Rollback verification had warnings"
    fi

    # Database migrations
    rollback_migrations

    # Send notification
    send_notification "success" "$previous_version"

    log_info "Rollback completed successfully to version ${previous_version}"
    log_warn "Please investigate the root cause of the deployment failure"
}

# Run main function
main
