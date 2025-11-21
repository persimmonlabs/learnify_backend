# Learnify API

![Go Version](https://img.shields.io/badge/Go-1.24-blue)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue)
![Production Readiness](https://img.shields.io/badge/Production%20Ready-66.5%25-yellow)
![Architecture](https://img.shields.io/badge/Architecture-DDD-green)

Backend API for the Universal Blueprint learning platform built with Go.

## 🚀 Quick Start for Testers

**Want to test the backend quickly?** Follow these 4 steps:

```bash
# 1. Clone and navigate
git clone <repository-url> && cd backend

# 2. Copy environment file and add your OpenAI API key
cp .env.example .env
# Edit .env and add: AI_API_KEY=your-key-here

# 3. Start everything (PostgreSQL + API)
docker-compose up -d

# 4. Test it works
curl http://localhost:8080/health
```

**Detailed testing guide:** See [docs/TESTING_GUIDE.md](docs/TESTING_GUIDE.md)
**Testing checklist:** See [docs/TESTING_CHECKLIST.md](docs/TESTING_CHECKLIST.md)
**API collection:** Import `postman_collection.json` into Postman

---

## Architecture

```
backend/
├── cmd/api/              # Application entry point
├── config/               # Configuration management
├── internal/             # Private application code
│   ├── platform/         # Infrastructure (DB, HTTP, AI, Logger)
│   ├── identity/         # User domain (auth, profiles, onboarding)
│   ├── learning/         # Learning domain (courses, exercises, reviews)
│   └── social/           # Social domain (feed, recommendations, achievements)
├── migrations/           # Database migrations
└── Dockerfile            # Container configuration
```

### Domain-Driven Design

The API follows DDD principles with three main domains:

1. **Identity** (`internal/identity/`)
   - User authentication (register, login)
   - User profiles
   - Onboarding (archetype + variable selection)

2. **Learning** (`internal/learning/`)
   - Universal Blueprint templates
   - Course generation (variable injection)
   - Exercise submission & testing
   - AI Senior Review
   - Progress tracking

3. **Social** (`internal/social/`)
   - Activity feed ("Strava for Brains")
   - Follow graph
   - Netflix-style recommendations
   - Achievements (Living Resume)
   - Trending courses

Each domain has:
- `models.go` - Data structures
- `repository.go` - Data access layer
- `service.go` - Business logic
- `handler.go` - HTTP endpoints

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Docker & Docker Compose (optional)

### Option 1: Docker (Recommended for Development)

```bash
# Start PostgreSQL + API
make docker-up

# Run migrations
make migrate

# View logs
docker-compose logs -f api
```

### Option 2: Local Development

```bash
# Install dependencies
make deps

# Start PostgreSQL (if not using Docker)
docker run -d \
  --name learnify-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=learnify \
  -p 5432:5432 \
  postgres:16-alpine

# Run migrations
make migrate

# Copy environment variables
cp .env.example .env
# Edit .env with your configuration

# Run API
make run
```

## Configuration

Configuration is loaded from:
1. Environment variables (highest priority)
2. `config/config.yaml`
3. Default values (fallback)

Required environment variables:
- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_NAME`, etc.
- `JWT_SECRET` - For authentication
- `AI_API_KEY` - OpenAI or Anthropic API key

See `.env.example` for full list.

## Database Migrations

Migrations are located in `migrations/` and run sequentially:

```bash
# Run all migrations
make migrate

# Or manually with psql
psql postgresql://user:pass@localhost:5432/learnify -f migrations/001_create_identity_tables.sql
```

Migration tracking is handled by `schema_migrations` table.

## API Endpoints

### Authentication
- `POST /api/auth/register` - Create account
- `POST /api/auth/login` - Get JWT token

### Identity
- `GET /api/users/me` - Get profile
- `PATCH /api/users/me` - Update profile
- `POST /api/onboarding/complete` - Save onboarding results

### Learning
- `GET /api/courses` - List user's courses
- `GET /api/courses/:id` - Course details
- `GET /api/exercises/:id` - Exercise details
- `POST /api/exercises/:id/submit` - Submit code
- `POST /api/submissions/:id/review` - Request AI review
- `GET /api/courses/:id/progress` - Progress tracking

### Social
- `GET /api/feed` - Activity ticker
- `POST /api/users/:id/follow` - Follow user
- `GET /api/recommendations` - Netflix-style recommendations
- `GET /api/trending` - Trending courses
- `GET /api/users/:id/profile` - Living Resume
- `GET /api/users/me/achievements` - Earned badges

## Development

### Build
```bash
make build
```

### Run Tests
```bash
make test

# With coverage
make test-coverage
```

### Linting
```bash
make lint
```

### Format Code
```bash
make fmt
```

## AI Integration

The platform uses AI for:

1. **Domain Validation** - Validates user domain during onboarding
2. **Variable Extraction** - Extracts 5 universal variables (ENTITY, STATE, FLOW, LOGIC, INTERFACE)
3. **Curriculum Generation** - Creates personalized courses from blueprints
4. **Code Review** - AI Senior Review with 4-category scoring
5. **Visualization** - Generates system diagrams

Supported providers:
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude 3)

## Deployment

### Docker Production Build

```bash
docker build -t learnify-api .
docker run -p 8080:8080 --env-file .env learnify-api
```

### Environment Variables for Production

```bash
export SERVER_ENV=production
export DATABASE_SSLMODE=require
export JWT_SECRET=strong-random-secret
export AI_API_KEY=your-production-key
```

## Architecture Patterns

### Repository Pattern
Data access is abstracted in `repository.go` files for testability.

### Service Layer
Business logic lives in `service.go` files, separate from HTTP handlers.

### Middleware Chain
- CORS - Cross-origin requests
- Logging - Request/response logging
- Auth - JWT validation

### Dependency Injection
Dependencies are passed via constructors (no globals).

## Philosophy

**"Strava for Brains, not Facebook for Learning"**

- Data-driven activity only (no fluff social)
- Passive social (automated progress broadcasts)
- Netflix-style algorithmic recommendations
- Universal Blueprint (template → instance via variable injection)
- AI Senior Review (architecture over production)

## Production Readiness

**Current Status:** 66.5/100 (Not Production Ready)

See [Production Readiness Report](docs/production-readiness-report.md) for detailed assessment.

### Critical Blockers

1. Remove default JWT secret from code
2. Implement health check and metrics endpoints
3. Add comprehensive test coverage (>80%)
4. Implement rate limiting
5. Add security headers middleware
6. Add panic recovery middleware

### Key Features Implemented

- Domain-driven design architecture
- JWT authentication and authorization
- Database connection pooling
- Graceful shutdown handling
- Request logging with correlation IDs
- CORS middleware
- Environment-based configuration

### Documentation

- [Architecture Review](docs/architecture-review.md) - Comprehensive architecture analysis
- [Production Readiness Report](docs/production-readiness-report.md) - Deployment readiness assessment
- [Performance Benchmarks](docs/benchmarks.md) - Load testing and optimization guide
- [API Documentation](docs/) - API endpoints and schemas (To be completed)

## Health Checks & Monitoring

**TO BE IMPLEMENTED:**

```bash
# Health check endpoint
curl http://localhost:8080/health

# Metrics endpoint (Prometheus format)
curl http://localhost:8080/metrics

# Readiness probe
curl http://localhost:8080/readiness
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-21T12:00:00Z",
  "checks": {
    "database": "healthy",
    "ai_service": "healthy"
  }
}
```

## Performance Targets

- Throughput: 1000+ requests/second
- Latency (p95): <200ms
- Latency (p99): <500ms
- Error Rate: <0.1%
- Availability: 99.9%

See [benchmarks.md](docs/benchmarks.md) for detailed performance analysis.

## Security Features

### Implemented

- JWT token-based authentication
- Password hashing (bcrypt)
- CORS protection
- Request timeout enforcement
- Database connection encryption support

### Required Before Production

- Remove default JWT secret
- Implement rate limiting (100 req/s per IP)
- Add security headers (HSTS, CSP, X-Frame-Options)
- Input validation middleware
- Token refresh mechanism
- Token blacklist for logout

See [Architecture Review](docs/architecture-review.md) for security audit.

## Next Steps

### Phase 1: Critical Fixes (2 weeks)

- Remove default secrets
- Add health checks and metrics
- Implement rate limiting
- Add security headers
- Write comprehensive tests (>80% coverage)

### Phase 2: High Priority (2 weeks)

- Implement circuit breakers
- Add structured logging and tracing
- Generate API documentation (OpenAPI)
- Add input validation
- Implement retry logic

### Phase 3: Production Hardening (1 week)

- Load testing and optimization
- Security audit
- Create runbooks
- Set up monitoring and alerts
- Disaster recovery testing

**Estimated Timeline to Production: 5 weeks**

## Contributing

Please ensure:
- All tests pass (`make test`)
- Code is formatted (`make fmt`)
- Linting passes (`make lint`)
- Documentation is updated
- Security considerations are addressed

## Support

- Architecture Review: `docs/architecture-review.md`
- Production Readiness: `docs/production-readiness-report.md`
- Performance Benchmarks: `docs/benchmarks.md`
- Frontend: `../learnify/`
- Migrations: `migrations/README.md`

---

**Status:** In development - Production readiness: 66.5/100
**Last Assessment:** 2025-11-21
**Next Review:** After Phase 1 completion
