# Observability Documentation

## Overview

This document describes the observability stack implemented in the Learnify backend, including health checks, metrics, logging, and distributed tracing.

## Health Check Endpoints

### Liveness Probe: `/health`

Simple health check that returns 200 OK if the server is running. Used by orchestrators (Kubernetes, Docker) to determine if the container should be restarted.

**Response Example:**
```json
{
  "status": "UP",
  "version": "1.0.0",
  "uptime": "2h34m12s",
  "timestamp": "2025-11-21T08:00:00Z"
}
```

### Readiness Probe: `/health/ready`

Deep health check that verifies all dependencies are available. Returns 503 Service Unavailable if any critical dependency is down.

**Response Example (Healthy):**
```json
{
  "status": "UP",
  "version": "1.0.0",
  "uptime": "2h34m12s",
  "timestamp": "2025-11-21T08:00:00Z",
  "checks": [
    {
      "name": "database",
      "status": "UP"
    },
    {
      "name": "memory",
      "status": "UP"
    }
  ]
}
```

**Response Example (Unhealthy):**
```json
{
  "status": "DOWN",
  "version": "1.0.0",
  "uptime": "2h34m12s",
  "timestamp": "2025-11-21T08:00:00Z",
  "checks": [
    {
      "name": "database",
      "status": "DOWN",
      "error": "connection refused"
    },
    {
      "name": "memory",
      "status": "UP"
    }
  ]
}
```

## Metrics Endpoint

### Prometheus Metrics: `/metrics`

Exposes metrics in Prometheus format for scraping. Metrics are automatically collected and updated.

## Metrics Catalog

### HTTP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | method, endpoint, status | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | method, endpoint, status | Request duration |
| `http_request_size_bytes` | Histogram | method, endpoint | Request body size |
| `http_response_size_bytes` | Histogram | method, endpoint | Response body size |

### Database Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `db_connections_open` | Gauge | - | Open database connections |
| `db_connections_in_use` | Gauge | - | Connections currently in use |
| `db_connections_idle` | Gauge | - | Idle connections |
| `db_query_duration_seconds` | Histogram | query_type | Query execution time |

### Authentication Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `jwt_validation_total` | Counter | status | JWT validation attempts |

### Business Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `user_registrations_total` | Counter | - | Total user registrations |
| `user_logins_total` | Counter | - | Total user logins |
| `exercise_submissions_total` | Counter | status | Exercise submissions |
| `ai_requests_total` | Counter | provider, status | AI API requests |
| `ai_request_duration_seconds` | Histogram | provider | AI request duration |

### Performance Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `memory_alloc_bytes` | Gauge | - | Currently allocated memory |
| `memory_total_alloc_bytes` | Counter | - | Cumulative allocated memory |
| `memory_sys_bytes` | Gauge | - | Memory from system |
| `goroutines_count` | Gauge | - | Number of goroutines |
| `gc_pause_duration_seconds` | Histogram | - | GC pause duration |
| `gc_runs_total` | Counter | - | Total GC runs |

### SLI/SLO Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `sli_latency_seconds` | Histogram | endpoint, method | Request latency for SLI |
| `sli_availability_total` | Counter | status_class | Availability tracking |

## Structured Logging

### Log Levels

- **DEBUG**: Detailed diagnostic information
- **INFO**: General informational messages
- **WARN**: Warning messages
- **ERROR**: Error messages

### Log Fields

All logs include structured fields for filtering and analysis:

- `request_id`: Unique request identifier (correlation ID)
- `user_id`: Authenticated user ID (when available)
- `method`: HTTP method
- `path`: Request path
- `status_code`: HTTP status code
- `duration_ms`: Request duration in milliseconds
- `error`: Error message (for error logs)

### Example Log Output (Development)
```
time=2025-11-21T08:00:00.000Z level=INFO msg="http_request" request_id=550e8400-e29b-41d4-a716-446655440000 method=POST path=/api/auth/login status_code=200 duration_ms=45
```

### Example Log Output (Production - JSON)
```json
{
  "time": "2025-11-21T08:00:00.000Z",
  "level": "INFO",
  "msg": "http_request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/api/auth/login",
  "status_code": 200,
  "duration_ms": 45
}
```

## Distributed Tracing

### Correlation IDs

Every request is assigned a unique `request_id` (correlation ID) that:
- Is included in all log entries for that request
- Is propagated through the application context
- Is returned in the `X-Request-ID` response header
- Can be used to trace requests across services

### Request Flow

1. Request arrives → Tracing middleware generates `request_id`
2. `request_id` added to request context
3. `X-Request-ID` header added to response
4. All logs for this request include the `request_id`
5. Client can use `X-Request-ID` for support/debugging

## Prometheus Configuration

### Scrape Configuration Example

```yaml
scrape_configs:
  - job_name: 'learnify-backend'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Docker Compose Example

```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

volumes:
  grafana-storage:
```

## Grafana Dashboard Examples

### Request Rate Panel (PromQL)
```promql
# Requests per second by endpoint
rate(http_requests_total[5m])

# Error rate (4xx and 5xx)
sum(rate(http_requests_total{status=~"4..|5.."}[5m]))
```

### Latency Panel (PromQL)
```promql
# 95th percentile latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Average latency by endpoint
avg(rate(http_request_duration_seconds_sum[5m])) by (endpoint)
```

### Database Connections Panel (PromQL)
```promql
# Open connections
db_connections_open

# Connection pool utilization
(db_connections_in_use / db_connections_open) * 100
```

### Memory Usage Panel (PromQL)
```promql
# Allocated memory in MB
memory_alloc_bytes / 1024 / 1024

# Memory growth rate
rate(memory_total_alloc_bytes[5m])
```

## Alerting Recommendations

### Critical Alerts

1. **Service Down**
   ```promql
   up{job="learnify-backend"} == 0
   ```

2. **High Error Rate**
   ```promql
   rate(http_requests_total{status=~"5.."}[5m]) > 0.05
   ```

3. **Database Connection Pool Exhausted**
   ```promql
   db_connections_idle == 0
   ```

### Warning Alerts

1. **High Latency**
   ```promql
   histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1.0
   ```

2. **Memory Growth**
   ```promql
   rate(memory_alloc_bytes[5m]) > 10485760  # 10MB/s
   ```

3. **High Goroutine Count**
   ```promql
   goroutines_count > 1000
   ```

## SLI/SLO Examples

### Availability SLO

**Target**: 99.9% availability (< 0.1% 5xx errors)

```promql
# Calculate availability over 30 days
(
  sum(rate(sli_availability_total{status_class="2xx"}[30d]))
  /
  sum(rate(sli_availability_total[30d]))
) * 100
```

### Latency SLO

**Target**: 95% of requests complete in < 500ms

```promql
# Calculate percentage of requests under 500ms
(
  sum(rate(sli_latency_seconds_bucket{le="0.5"}[30d]))
  /
  sum(rate(sli_latency_seconds_count[30d]))
) * 100
```

## Best Practices

1. **Monitor both RED and USE metrics**
   - RED: Rate, Errors, Duration (for requests)
   - USE: Utilization, Saturation, Errors (for resources)

2. **Set up alerts for SLO violations**
   - Track error budgets
   - Alert when approaching budget exhaustion

3. **Use structured logging**
   - Always include correlation IDs
   - Use JSON format in production
   - Index logs for fast searching

4. **Dashboard organization**
   - Overview dashboard (key metrics)
   - Service-specific dashboards
   - Infrastructure dashboards

5. **Regular review**
   - Review alerts monthly
   - Update SLOs based on actual performance
   - Retire unused metrics

## Troubleshooting

### High Memory Usage

1. Check `memory_alloc_bytes` metric
2. Review `gc_pause_duration_seconds` for GC pressure
3. Check for memory leaks in application logs
4. Use pprof for detailed memory profiling

### Slow Requests

1. Check `http_request_duration_seconds` histogram
2. Review database query metrics
3. Check AI request latency
4. Look for correlation IDs in logs to trace specific requests

### Database Connection Issues

1. Check `db_connections_open` and `db_connections_in_use`
2. Review database health check status
3. Check connection pool configuration
4. Monitor query duration for slow queries

## Integration with APM Tools

This observability stack is compatible with:
- Prometheus + Grafana
- Datadog (via Prometheus integration)
- New Relic (OpenMetrics)
- AWS CloudWatch (EMF format)
- Google Cloud Monitoring

## Related Documentation

- [Deployment Guide](./deployment.md)
- [Configuration Reference](./configuration.md)
- [Monitoring Runbook](./runbook.md)
