#!/bin/bash
set -euo pipefail

# Health check script for deployed application
# Usage: ./health-check.sh <environment>

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

    # shellcheck disable=SC1090
    source "$env_file"
}

# Check if service is responding
check_health_endpoint() {
    log_info "Checking health endpoint..."

    local response
    local status_code

    response=$(curl -s -w "\n%{http_code}" "${APP_URL}/health" || echo "000")
    status_code=$(echo "$response" | tail -n1)

    if [ "$status_code" -eq 200 ]; then
        log_info "Health endpoint: OK (200)"
        return 0
    else
        log_error "Health endpoint: FAILED (${status_code})"
        return 1
    fi
}

# Check database connectivity
check_database() {
    log_info "Checking database connectivity..."

    local response
    local status_code

    response=$(curl -s -w "\n%{http_code}" "${APP_URL}/health/db" || echo "000")
    status_code=$(echo "$response" | tail -n1)

    if [ "$status_code" -eq 200 ]; then
        log_info "Database: OK"
        return 0
    else
        log_error "Database: FAILED"
        return 1
    fi
}

# Check Redis connectivity
check_redis() {
    log_info "Checking Redis connectivity..."

    local response
    local status_code

    response=$(curl -s -w "\n%{http_code}" "${APP_URL}/health/redis" || echo "000")
    status_code=$(echo "$response" | tail -n1)

    if [ "$status_code" -eq 200 ]; then
        log_info "Redis: OK"
        return 0
    else
        log_warn "Redis: FAILED (non-critical)"
        return 0
    fi
}

# Check API endpoints
check_api_endpoints() {
    log_info "Checking critical API endpoints..."

    local endpoints=(
        "/api/v1/users"
        "/api/v1/products"
    )

    local failed=0

    for endpoint in "${endpoints[@]}"; do
        local status_code
        status_code=$(curl -s -o /dev/null -w "%{http_code}" "${APP_URL}${endpoint}" || echo "000")

        if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 500 ]; then
            log_info "Endpoint ${endpoint}: OK (${status_code})"
        else
            log_error "Endpoint ${endpoint}: FAILED (${status_code})"
            failed=$((failed + 1))
        fi
    done

    if [ $failed -gt 0 ]; then
        return 1
    fi

    return 0
}

# Check response time
check_response_time() {
    log_info "Checking response time..."

    local start_time
    local end_time
    local duration

    start_time=$(date +%s%N)
    curl -s -o /dev/null "${APP_URL}/health"
    end_time=$(date +%s%N)

    duration=$(( (end_time - start_time) / 1000000 ))

    log_info "Response time: ${duration}ms"

    if [ "$duration" -gt 1000 ]; then
        log_warn "Response time is high (>${duration}ms)"
        return 1
    fi

    return 0
}

# Check container health
check_container_health() {
    log_info "Checking container health..."

    if ! command -v docker &> /dev/null; then
        log_warn "Docker not available, skipping container health check"
        return 0
    fi

    local container_name="${APP_NAME:-backend}"

    if ! docker ps --filter "name=${container_name}" --format "{{.Names}}" | grep -q "${container_name}"; then
        log_error "Container ${container_name} is not running"
        return 1
    fi

    local health_status
    health_status=$(docker inspect --format='{{.State.Health.Status}}' "${container_name}" 2>/dev/null || echo "unknown")

    if [ "$health_status" = "healthy" ]; then
        log_info "Container health: OK"
        return 0
    else
        log_error "Container health: ${health_status}"
        return 1
    fi
}

# Check metrics endpoint
check_metrics() {
    log_info "Checking metrics endpoint..."

    local response
    local status_code

    response=$(curl -s -w "\n%{http_code}" "${APP_URL}/metrics" || echo "000")
    status_code=$(echo "$response" | tail -n1)

    if [ "$status_code" -eq 200 ]; then
        log_info "Metrics: OK"

        # Parse some basic metrics
        local body
        body=$(echo "$response" | head -n -1)

        if echo "$body" | grep -q "go_goroutines"; then
            local goroutines
            goroutines=$(echo "$body" | grep "go_goroutines" | awk '{print $2}')
            log_info "Active goroutines: $goroutines"
        fi

        return 0
    else
        log_warn "Metrics: Not available"
        return 0
    fi
}

# Comprehensive health check
run_health_checks() {
    log_info "Running comprehensive health checks for ${ENVIRONMENT}..."

    local failed_checks=0
    local total_checks=0

    # Run all checks
    checks=(
        "check_health_endpoint"
        "check_database"
        "check_redis"
        "check_api_endpoints"
        "check_response_time"
        "check_container_health"
        "check_metrics"
    )

    for check in "${checks[@]}"; do
        total_checks=$((total_checks + 1))

        if ! $check; then
            failed_checks=$((failed_checks + 1))
        fi

        echo ""
    done

    # Summary
    log_info "Health Check Summary:"
    log_info "Total checks: $total_checks"
    log_info "Passed: $((total_checks - failed_checks))"
    log_info "Failed: $failed_checks"

    if [ $failed_checks -eq 0 ]; then
        log_info "All health checks passed!"
        return 0
    else
        log_error "Some health checks failed"
        return 1
    fi
}

# Main function
main() {
    log_info "Health Check Tool"
    log_info "Environment: ${ENVIRONMENT}"

    # Load configuration
    load_config

    # Run health checks
    if run_health_checks; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main
