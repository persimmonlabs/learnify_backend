#!/bin/bash
set -euo pipefail

# Deployment script for Go backend
# Usage: ./deploy.sh <environment> <version>

ENVIRONMENT="${1:-staging}"
VERSION="${2:-latest}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment-specific configuration
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

# Pull Docker image
pull_image() {
    local image="${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}"
    log_info "Pulling Docker image: $image"

    docker pull "$image" || {
        log_error "Failed to pull Docker image"
        exit 1
    }
}

# Stop existing containers
stop_containers() {
    log_info "Stopping existing containers..."

    if docker ps -q --filter "name=${APP_NAME}" | grep -q .; then
        docker stop "${APP_NAME}" || true
        docker rm "${APP_NAME}" || true
    fi
}

# Start new container
start_container() {
    local image="${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}"

    log_info "Starting new container..."

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
            log_error "Failed to start container"
            exit 1
        }

    log_info "Container started successfully"
}

# Wait for application to be healthy
wait_for_health() {
    log_info "Waiting for application to be healthy..."
    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if docker inspect --format='{{.State.Health.Status}}' "${APP_NAME}" | grep -q "healthy"; then
            log_info "Application is healthy"
            return 0
        fi

        attempt=$((attempt + 1))
        log_info "Health check attempt $attempt/$max_attempts..."
        sleep 2
    done

    log_error "Application failed to become healthy"
    docker logs "${APP_NAME}" --tail 50
    return 1
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."

    docker run --rm \
        -e "DATABASE_URL=${DATABASE_URL}" \
        "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}" \
        migrate -database "$DATABASE_URL" -path /app/migrations up || {
            log_error "Migration failed"
            return 1
        }

    log_info "Migrations completed successfully"
}

# Blue-green deployment
blue_green_deploy() {
    log_info "Performing blue-green deployment..."

    # Start new container with temporary name
    local new_container="${APP_NAME}-green"
    local old_container="${APP_NAME}-blue"

    # Rename current container if exists
    if docker ps -q --filter "name=${APP_NAME}$" | grep -q .; then
        docker rename "${APP_NAME}" "$old_container"
    fi

    # Start new container
    docker run -d \
        --name "$new_container" \
        --restart unless-stopped \
        -p "${APP_PORT}:8080" \
        -e "DATABASE_URL=${DATABASE_URL}" \
        -e "REDIS_URL=${REDIS_URL}" \
        -e "ENV=${ENVIRONMENT}" \
        "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}"

    # Wait for health check
    sleep 5

    if docker inspect --format='{{.State.Health.Status}}' "$new_container" | grep -q "healthy"; then
        log_info "New container is healthy, switching traffic..."

        # Rename containers
        docker rename "$new_container" "${APP_NAME}"

        # Stop and remove old container
        if docker ps -aq --filter "name=$old_container" | grep -q .; then
            docker stop "$old_container"
            docker rm "$old_container"
        fi

        log_info "Deployment completed successfully"
    else
        log_error "New container is not healthy, rolling back..."
        docker stop "$new_container"
        docker rm "$new_container"

        # Restore old container name
        if docker ps -aq --filter "name=$old_container" | grep -q .; then
            docker rename "$old_container" "${APP_NAME}"
        fi

        exit 1
    fi
}

# Cleanup old images
cleanup() {
    log_info "Cleaning up old images..."

    docker images "${DOCKER_REGISTRY}/${DOCKER_IMAGE}" --format "{{.Tag}}" | \
        grep -v "${VERSION}" | \
        grep -v "latest" | \
        head -n -3 | \
        xargs -r -I {} docker rmi "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:{}" || true
}

# Main deployment flow
main() {
    log_info "Starting deployment to ${ENVIRONMENT} with version ${VERSION}"

    # Load configuration
    load_config

    # Pull new image
    pull_image

    # Run migrations
    if [ "${RUN_MIGRATIONS:-true}" = "true" ]; then
        run_migrations || {
            log_error "Migrations failed, aborting deployment"
            exit 1
        }
    fi

    # Perform deployment
    if [ "${DEPLOYMENT_STRATEGY:-rolling}" = "blue-green" ]; then
        blue_green_deploy
    else
        stop_containers
        start_container
        wait_for_health || {
            log_error "Deployment failed"
            exit 1
        }
    fi

    # Cleanup
    cleanup

    log_info "Deployment completed successfully!"
}

# Run main function
main
