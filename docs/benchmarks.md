# Performance Benchmarks - Learnify Backend API

**Last Updated:** 2025-11-21
**Version:** Initial Assessment
**Environment:** Development/Local

---

## Executive Summary

This document outlines performance benchmarks, load testing results, and performance optimization recommendations for the Learnify Backend API.

**Current Status:** Baseline benchmarks pending implementation
**Target Performance:** 1000+ requests/second with <200ms p95 latency

---

## 1. Benchmark Methodology

### 1.1 Testing Tools

**Load Testing:**
- **Primary:** Apache Bench (ab), k6, or wrk
- **Advanced:** Grafana k6 for complex scenarios
- **Infrastructure:** Prometheus + Grafana for metrics

**Database Profiling:**
- PostgreSQL `pg_stat_statements`
- `EXPLAIN ANALYZE` for query optimization

**Application Profiling:**
- Go pprof (CPU, memory, goroutines)
- trace tool for goroutine analysis

### 1.2 Test Environment

**Hardware Specifications (Target):**
- CPU: 4 vCPUs
- RAM: 4GB
- Network: 1Gbps
- Database: PostgreSQL 16 (separate instance)

**Software:**
- Go 1.24.0
- PostgreSQL 16
- Operating System: Linux (Ubuntu 22.04 or equivalent)

---

## 2. Expected Performance Targets

### 2.1 API Response Times

| Endpoint Category | p50 | p95 | p99 | Max Acceptable |
|-------------------|-----|-----|-----|----------------|
| **Authentication** | 50ms | 100ms | 150ms | 200ms |
| Public routes (no auth) | 20ms | 50ms | 100ms | 150ms |
| Simple queries (GET user) | 30ms | 50ms | 80ms | 120ms |
| Complex queries (GET feed) | 100ms | 200ms | 400ms | 800ms |
| AI-powered operations | 1000ms | 2000ms | 3000ms | 5000ms |
| Database writes | 40ms | 80ms | 120ms | 200ms |

### 2.2 Throughput Targets

**Per Instance:**
- **Target:** 1000 requests/second sustained
- **Peak:** 1500 requests/second for 1 minute
- **Concurrent Users:** 5000+

**Database:**
- **Queries:** 500 queries/second
- **Connections:** 25 concurrent connections
- **Transaction Rate:** 300 transactions/second

### 2.3 Resource Utilization

**Under Normal Load (500 req/s):**
- CPU: < 40%
- Memory: < 1GB
- Goroutines: < 500
- Database Connections: < 15 active

**Under Peak Load (1000 req/s):**
- CPU: < 70%
- Memory: < 1.5GB
- Goroutines: < 1000
- Database Connections: < 20 active

---

## 3. Baseline Benchmarks (To Be Measured)

### 3.1 Endpoint Performance

**Authentication Endpoints:**

**POST /api/auth/register**
```bash
# Test Command
ab -n 1000 -c 10 -p register.json -T application/json \
   http://localhost:8080/api/auth/register

# Expected Results (To Be Measured)
# Requests per second:    ??? req/s
# Time per request:       ??? ms (mean)
# Time per request:       ??? ms (mean, across all concurrent requests)
# Transfer rate:          ??? KB/s
# Success rate:           100%
# Error rate:             0%
```

**POST /api/auth/login**
```bash
# Test Command
ab -n 1000 -c 10 -p login.json -T application/json \
   http://localhost:8080/api/auth/login

# Expected Results (To Be Measured)
# Target: 100-150 ms average
# Target: >200 requests/second
```

**Protected Endpoints:**

**GET /api/users/me**
```bash
# Test Command
ab -n 10000 -c 50 -H "Authorization: Bearer $TOKEN" \
   http://localhost:8080/api/users/me

# Expected Results (To Be Measured)
# Target: 20-50 ms average
# Target: >500 requests/second
```

**GET /api/courses**
```bash
# Test Command
ab -n 5000 -c 25 -H "Authorization: Bearer $TOKEN" \
   http://localhost:8080/api/courses

# Expected Results (To Be Measured)
# Target: 50-100 ms average
# Target: >300 requests/second
```

**GET /api/feed**
```bash
# Test Command
ab -n 1000 -c 10 -H "Authorization: Bearer $TOKEN" \
   http://localhost:8080/api/feed

# Expected Results (To Be Measured)
# Target: 100-200 ms average
# Target: >100 requests/second
# Note: Complex join queries, may require optimization
```

### 3.2 Database Performance

**Connection Pool Metrics:**
```sql
-- Monitor active connections
SELECT count(*) as active_connections
FROM pg_stat_activity
WHERE datname = 'learnify';

-- Expected: < 20 under normal load
```

**Query Performance (Top 5 Slowest):**
```sql
-- Enable pg_stat_statements extension
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Find slowest queries
SELECT
    query,
    calls,
    total_exec_time / 1000 as total_time_seconds,
    mean_exec_time as avg_time_ms,
    max_exec_time as max_time_ms
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 5;

-- Target: All queries < 100ms average
```

**Database Write Performance:**
```bash
# User registration (INSERT into users table)
# Expected: < 50ms per insert
# With indexes: < 80ms per insert

# Course enrollment (INSERT into user_courses)
# Expected: < 30ms per insert
```

### 3.3 Middleware Overhead

**Measurement:**
```go
// Middleware chain latency breakdown
// (To be measured with instrumentation)

LoggingSimple: ??? ms
CORS:          ??? ms
Auth:          ??? ms (JWT parsing + validation)
Handler:       ??? ms (business logic)
Total:         ??? ms

// Target total middleware overhead: < 10ms
```

---

## 4. Load Testing Scenarios

### 4.1 Baseline Load Test

**Objective:** Establish baseline performance

**Parameters:**
- Concurrent Users: 100
- Duration: 10 minutes
- Request Mix: 70% reads, 30% writes
- Ramp-up: 2 minutes

**Success Criteria:**
- [ ] Error rate < 0.1%
- [ ] p95 latency < 200ms
- [ ] p99 latency < 500ms
- [ ] CPU < 60%
- [ ] Memory < 1.5GB
- [ ] No goroutine leaks
- [ ] Database connections < 20

**K6 Test Script:**
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '10m', target: 100 }, // Stay at 100
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = 'http://localhost:8080/api';
let authToken = '';

export function setup() {
  // Login to get auth token
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: 'test@example.com',
    password: 'password123'
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  return { token: loginRes.json('token') };
}

export default function(data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  // 70% GET requests
  if (Math.random() < 0.7) {
    const endpoints = [
      '/users/me',
      '/courses',
      '/feed',
      '/trending',
    ];
    const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
    const res = http.get(`${BASE_URL}${endpoint}`, { headers });
    check(res, {
      'status is 200': (r) => r.status === 200,
      'response time < 200ms': (r) => r.timings.duration < 200,
    });
  }
  // 30% POST/PATCH requests
  else {
    const res = http.patch(`${BASE_URL}/users/me`, JSON.stringify({
      bio: 'Updated bio',
    }), { headers });
    check(res, {
      'status is 200': (r) => r.status === 200,
    });
  }

  sleep(1); // 1 request per second per user
}
```

### 4.2 Stress Test

**Objective:** Find breaking point and failure modes

**Parameters:**
- Start: 100 concurrent users
- Increment: 100 users every 2 minutes
- End: Until failure or 5000 users
- Duration: Until system degrades

**Success Criteria:**
- [ ] Graceful degradation (no crashes)
- [ ] Clear breaking point identified
- [ ] Resource exhaustion point identified
- [ ] Auto-scaling triggers (if configured)

**Expected Breaking Points:**
1. Database connection pool exhaustion (~500-1000 users)
2. CPU saturation (~1500-2000 users)
3. Memory exhaustion (~3000-4000 users)

**Monitoring During Test:**
```bash
# CPU and Memory
top -p $(pgrep -f 'learnify-api')

# Goroutines
curl localhost:6060/debug/pprof/goroutine?debug=1

# Database connections
psql -c "SELECT count(*) FROM pg_stat_activity WHERE datname='learnify';"

# HTTP metrics (if implemented)
curl localhost:8080/metrics | grep http_requests
```

### 4.3 Endurance Test (Soak Test)

**Objective:** Identify memory leaks and stability issues

**Parameters:**
- Concurrent Users: 500 (moderate load)
- Duration: 24 hours
- Request Rate: Steady (no spikes)

**Success Criteria:**
- [ ] No memory leaks (stable memory usage)
- [ ] No goroutine leaks
- [ ] Stable response times (no degradation)
- [ ] Stable error rate
- [ ] No database connection leaks

**Monitoring Points:**
- Every 15 minutes: Memory usage, goroutine count
- Every 1 hour: Full metrics snapshot
- Every 6 hours: Database connection pool stats

**K6 Endurance Test:**
```javascript
export let options = {
  stages: [
    { duration: '5m', target: 500 },   // Ramp up
    { duration: '24h', target: 500 },  // Stay at 500 for 24 hours
    { duration: '5m', target: 0 },     // Ramp down
  ],
};
```

### 4.4 Spike Test

**Objective:** Test handling of sudden traffic increases

**Parameters:**
- Baseline: 100 concurrent users
- Spike: Instant jump to 2000 users
- Spike Duration: 5 minutes
- Recovery: Return to 100 users

**Success Criteria:**
- [ ] System remains available during spike
- [ ] Error rate < 5% during spike
- [ ] Auto-scaling triggers within 2 minutes
- [ ] Full recovery within 5 minutes after spike ends
- [ ] No lingering performance degradation

**K6 Spike Test:**
```javascript
export let options = {
  stages: [
    { duration: '5m', target: 100 },    // Normal load
    { duration: '0s', target: 2000 },   // Instant spike
    { duration: '5m', target: 2000 },   // Maintain spike
    { duration: '0s', target: 100 },    // Instant drop
    { duration: '5m', target: 100 },    // Recovery
  ],
};
```

---

## 5. Database Optimization

### 5.1 Query Optimization Checklist

**Before Optimization:**
- [ ] Enable `pg_stat_statements`
- [ ] Identify top 10 slowest queries
- [ ] Run EXPLAIN ANALYZE on slow queries

**Optimization Techniques:**

**1. Index Optimization:**
```sql
-- User lookup by email (authentication)
CREATE INDEX idx_users_email ON users(email);

-- Course queries by user
CREATE INDEX idx_user_courses_user_id ON user_courses(user_id);
CREATE INDEX idx_user_courses_course_id ON user_courses(course_id);

-- Feed queries (activity feed)
CREATE INDEX idx_activities_user_id_created_at
  ON activities(user_id, created_at DESC);

-- Follow graph queries
CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_following_id ON follows(following_id);
```

**2. Query Optimization Examples:**

**Before:**
```sql
-- Slow: Multiple sequential queries
SELECT * FROM users WHERE id = $1;
SELECT * FROM user_courses WHERE user_id = $1;
SELECT * FROM achievements WHERE user_id = $1;
```

**After:**
```sql
-- Fast: Single JOIN query
SELECT
  u.*,
  json_agg(DISTINCT uc.*) as courses,
  json_agg(DISTINCT a.*) as achievements
FROM users u
LEFT JOIN user_courses uc ON uc.user_id = u.id
LEFT JOIN achievements a ON a.user_id = u.id
WHERE u.id = $1
GROUP BY u.id;
```

**3. Connection Pool Tuning:**
```go
// Current configuration
MaxOpenConns:    25
MaxIdleConns:    5
ConnMaxLifetime: 5 * time.Minute
ConnMaxIdleTime: 10 * time.Minute

// Recommended for high load
MaxOpenConns:    50  // Increase for more concurrency
MaxIdleConns:    10  // More idle connections for bursts
ConnMaxLifetime: 5 * time.Minute
ConnMaxIdleTime: 10 * time.Minute
```

### 5.2 Database Performance Monitoring

**Key Metrics:**
```sql
-- Connection stats
SELECT
  count(*) as total_connections,
  count(*) FILTER (WHERE state = 'active') as active,
  count(*) FILTER (WHERE state = 'idle') as idle
FROM pg_stat_activity;

-- Cache hit ratio (should be > 99%)
SELECT
  sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as cache_hit_ratio
FROM pg_statio_user_tables;

-- Index usage
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Table bloat check
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## 6. Application Performance Optimization

### 6.1 Go Profiling

**CPU Profiling:**
```bash
# Enable pprof in main.go
import _ "net/http/pprof"
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# Capture CPU profile (30 seconds)
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze profile
go tool pprof cpu.prof
> top10
> list <function_name>
> web
```

**Memory Profiling:**
```bash
# Capture heap profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Analyze memory
go tool pprof heap.prof
> top10
> list <function_name>
```

**Goroutine Leak Detection:**
```bash
# Check goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# Analyze goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 6.2 Optimization Opportunities

**1. Response Caching:**
```go
// Example: Cache course list for 5 minutes
type CachedResponse struct {
    Data      interface{}
    ExpiresAt time.Time
}

var courseListCache sync.Map

func GetCourses(ctx context.Context) ([]Course, error) {
    // Check cache
    if cached, ok := courseListCache.Load("all_courses"); ok {
        entry := cached.(CachedResponse)
        if time.Now().Before(entry.ExpiresAt) {
            return entry.Data.([]Course), nil
        }
    }

    // Cache miss - fetch from database
    courses, err := repo.FindAll(ctx)
    if err != nil {
        return nil, err
    }

    // Store in cache
    courseListCache.Store("all_courses", CachedResponse{
        Data:      courses,
        ExpiresAt: time.Now().Add(5 * time.Minute),
    })

    return courses, nil
}
```

**2. Database Query Batching:**
```go
// Before: N+1 query problem
for _, courseID := range courseIDs {
    course, err := repo.FindByID(ctx, courseID)
    // Process course
}

// After: Batch query
courses, err := repo.FindByIDs(ctx, courseIDs)
// Process all courses at once
```

**3. Concurrent Processing:**
```go
// Example: Parallel AI review processing
func ProcessReviews(ctx context.Context, submissions []Submission) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(submissions))

    for _, sub := range submissions {
        wg.Add(1)
        go func(s Submission) {
            defer wg.Done()
            if err := processReview(ctx, s); err != nil {
                errChan <- err
            }
        }(sub)
    }

    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 6.3 Middleware Optimization

**JWT Caching:**
```go
// Cache parsed JWT tokens to avoid repeated parsing
type JWTCache struct {
    cache *sync.Map
}

func (c *JWTCache) Get(tokenString string) (*UserClaims, bool) {
    if claims, ok := c.cache.Load(tokenString); ok {
        return claims.(*UserClaims), true
    }
    return nil, false
}

func (c *JWTCache) Set(tokenString string, claims *UserClaims, ttl time.Duration) {
    c.cache.Store(tokenString, claims)
    time.AfterFunc(ttl, func() {
        c.cache.Delete(tokenString)
    })
}
```

---

## 7. Performance Testing Results

### 7.1 Results Template

**Test Date:** [TO BE FILLED]
**Environment:** [Development/Staging/Production]
**Version:** [Git commit SHA]

**Hardware:**
- CPU:
- RAM:
- Database:

**Results:**

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Requests/second | 1000 | ??? | ❓ |
| p50 latency | 50ms | ??? | ❓ |
| p95 latency | 200ms | ??? | ❓ |
| p99 latency | 500ms | ??? | ❓ |
| Error rate | <0.1% | ??? | ❓ |
| CPU usage (avg) | <60% | ??? | ❓ |
| Memory usage (avg) | <1.5GB | ??? | ❓ |
| DB connections (avg) | <20 | ??? | ❓ |

**Bottlenecks Identified:**
1. [To be filled]
2. [To be filled]

**Optimizations Applied:**
1. [To be filled]
2. [To be filled]

**Next Steps:**
- [ ] [Action item 1]
- [ ] [Action item 2]

---

## 8. Continuous Performance Monitoring

### 8.1 Real-Time Metrics

**Application Metrics (Prometheus):**
```prometheus
# Request rate
rate(http_requests_total[5m])

# Average latency
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# Database connection pool
db_connections_open
db_connections_idle
db_connections_wait_duration_seconds
```

**Grafana Dashboards:**
- Application Overview (request rate, latency, errors)
- Database Performance (connections, query times)
- System Resources (CPU, memory, goroutines)
- Business Metrics (registrations, enrollments)

### 8.2 Performance Alerts

**Critical Alerts:**
- p99 latency > 2000ms for 5 minutes
- Error rate > 5% for 5 minutes
- Database connection pool > 90% for 5 minutes

**Warning Alerts:**
- p95 latency > 500ms for 10 minutes
- Error rate > 1% for 10 minutes
- CPU > 80% for 10 minutes
- Memory > 85% for 10 minutes

---

## 9. Optimization Roadmap

### 9.1 Short-Term (1-2 weeks)

- [ ] Implement response caching for static data
- [ ] Add database query result caching (Redis)
- [ ] Optimize top 5 slowest database queries
- [ ] Add missing database indexes
- [ ] Implement connection pooling for AI client

### 9.2 Medium-Term (1-2 months)

- [ ] Implement HTTP/2 support
- [ ] Add response compression (gzip)
- [ ] Implement request coalescing
- [ ] Database read replica for scaling
- [ ] CDN integration for static assets

### 9.3 Long-Term (3-6 months)

- [ ] Implement microservices architecture (if needed)
- [ ] Add message queue for async processing (RabbitMQ/Kafka)
- [ ] Implement edge caching
- [ ] Database sharding strategy
- [ ] Multi-region deployment

---

## 10. Conclusion

**Current Status:** Baseline performance benchmarks pending

**Next Steps:**
1. Implement health check and metrics endpoints
2. Run baseline load tests
3. Identify and fix performance bottlenecks
4. Re-test and validate improvements
5. Establish continuous performance monitoring

**Performance Goals:**
- Target: 1000+ req/s sustained throughput
- Latency: p95 < 200ms, p99 < 500ms
- Availability: 99.9% uptime

**Estimated Timeline:** 2-3 weeks for initial optimization and testing

---

**Document Owner:** Platform Team
**Last Updated:** 2025-11-21
**Next Review:** After load testing completion
