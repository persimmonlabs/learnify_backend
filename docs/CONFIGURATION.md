# Configuration Guide

## Overview

The Learnify API is configured exclusively via **environment variables**. This design ensures:
- ✅ 12-Factor App compliance
- ✅ Container/Kubernetes compatibility
- ✅ Clear separation between code and configuration
- ✅ No accidental secret commits

## Configuration Loading Priority

The application loads configuration from environment variables with the following fallback patterns:

1. Primary environment variable (new naming convention)
2. Legacy environment variable (for backward compatibility)
3. Hardcoded default (development-friendly, production validation enforced)

## Environment Variables

### Server Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `SERVER_PORT` | string | `"8080"` | HTTP server port |
| `SERVER_HOST` | string | `"0.0.0.0"` | HTTP server bind address |
| `SERVER_ENV` | string | `"development"` | Environment (`development`, `staging`, `production`) |
| `GRACEFUL_SHUTDOWN_TIMEOUT` | duration | `30s` | Graceful shutdown timeout (e.g., `30s`, `1m`, `60`) |
| `REQUEST_TIMEOUT` | duration | `30s` | HTTP request timeout (e.g., `30s`, `5m`) |

### Database Configuration

Supports both `DATABASE_*` (primary) and `DB_*` (legacy) prefixes:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DATABASE_HOST` or `DB_HOST` | string | `"localhost"` | PostgreSQL host |
| `DATABASE_PORT` or `DB_PORT` | string | `"5432"` | PostgreSQL port |
| `DATABASE_USER` or `DB_USER` | string | `"postgres"` | Database username |
| `DATABASE_PASSWORD` or `DB_PASSWORD` | string | `"postgres"` | Database password ⚠️ |
| `DATABASE_NAME` or `DB_NAME` | string | `"learnify"` | Database name |
| `DATABASE_SSL_MODE` or `DB_SSL_MODE` | string | `"disable"` | SSL mode (`disable`, `require`, `verify-full`) |

### JWT Authentication

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `JWT_SECRET` | string | **REQUIRED** | JWT signing secret (min 32 characters) |
| `JWT_EXPIRATION_SECONDS` | int | `86400` | Token expiration in seconds (86400 = 24 hours) |

### AI Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `AI_PROVIDER` | string | `"openai"` | AI provider (`openai`, `anthropic`, etc.) |
| `AI_API_KEY` | string | `""` | AI API key (required in production) |
| `AI_MODEL` | string | `"gpt-4"` | AI model identifier |

### CORS Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | string | `"*"` | Comma-separated allowed origins (use `*` for development only) |

## Environment-Specific Configuration

### Development (`.env`)

```bash
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_ENV=development
GRACEFUL_SHUTDOWN_TIMEOUT=30s

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=learnify_dev
DATABASE_SSL_MODE=disable

# JWT
JWT_SECRET=your-development-secret-min-32-chars
JWT_EXPIRATION_SECONDS=86400

# AI
AI_PROVIDER=openai
AI_API_KEY=sk-your-dev-key
AI_MODEL=gpt-4

# CORS
CORS_ALLOWED_ORIGINS=*
```

### Staging (`staging.env`)

```bash
# Server
SERVER_PORT=8080
SERVER_ENV=staging
GRACEFUL_SHUTDOWN_TIMEOUT=45s

# Database
DATABASE_HOST=staging-db.example.com
DATABASE_PORT=5432
DATABASE_USER=learnify_staging
DATABASE_PASSWORD=strong_staging_password
DATABASE_NAME=learnify_staging
DATABASE_SSL_MODE=require

# JWT
JWT_SECRET=your-strong-staging-jwt-secret-min-32-chars
JWT_EXPIRATION_SECONDS=3600  # 1 hour for staging

# AI
AI_PROVIDER=openai
AI_API_KEY=sk-your-staging-key
AI_MODEL=gpt-4

# CORS
CORS_ALLOWED_ORIGINS=https://staging.example.com
```

### Production (`production.env.example`)

```bash
# Server
SERVER_PORT=8080
SERVER_ENV=production
GRACEFUL_SHUTDOWN_TIMEOUT=60s
REQUEST_TIMEOUT=30s

# Database (SSL required in production)
DATABASE_HOST=production-db.example.com
DATABASE_PORT=5432
DATABASE_USER=learnify_prod
DATABASE_PASSWORD=${STRONG_DATABASE_PASSWORD}
DATABASE_NAME=learnify_production
DATABASE_SSL_MODE=require

# JWT (strong secret required)
JWT_SECRET=${STRONG_JWT_SECRET}  # Min 32 chars, randomly generated
JWT_EXPIRATION_SECONDS=3600      # 1 hour recommended for production

# AI
AI_PROVIDER=openai
AI_API_KEY=${AI_API_KEY}         # Required in production
AI_MODEL=gpt-4

# CORS (specific origins required)
CORS_ALLOWED_ORIGINS=https://www.example.com,https://app.example.com
```

## Data Types & Parsing

### Duration Strings

Duration variables support Go duration strings or integers (interpreted as seconds):

```bash
# All equivalent to 30 seconds:
GRACEFUL_SHUTDOWN_TIMEOUT=30s
GRACEFUL_SHUTDOWN_TIMEOUT=30
GRACEFUL_SHUTDOWN_TIMEOUT=0.5m

# Valid duration units: ns, us, ms, s, m, h
GRACEFUL_SHUTDOWN_TIMEOUT=1m30s  # 1 minute 30 seconds
REQUEST_TIMEOUT=5m               # 5 minutes
```

### Boolean Values

Boolean variables accept multiple formats (case-insensitive):

```bash
# True values:
SECURE_COOKIES=true
HELMET_ENABLED=1
CSRF_ENABLED=yes
HSTS_ENABLED=on

# False values:
SECURE_COOKIES=false
HELMET_ENABLED=0
CSRF_ENABLED=no
HSTS_ENABLED=off
```

### Integer Values

Integer variables are parsed with `strconv.Atoi`:

```bash
JWT_EXPIRATION_SECONDS=86400     # 24 hours
DATABASE_MAX_CONNECTIONS=100
RATE_LIMIT_REQUESTS=100
```

## Production Validation

When `SERVER_ENV=production`, the following validations are enforced:

### ✅ Required Validations

1. **JWT Secret Strength**
   - Minimum 32 characters
   - Cannot be empty
   - Error: Application fails to start

2. **Database Password**
   - Cannot be empty
   - Cannot be default value (`"postgres"`)
   - Error: Application fails to start

3. **AI API Key**
   - Cannot be empty in production
   - Error: Application fails to start

### ⚠️ Warning Validations

1. **Database SSL Mode**
   - Warning if `DATABASE_SSL_MODE=disable`
   - Recommendation: Use `require` or `verify-full`

2. **CORS Configuration**
   - Warning if `CORS_ALLOWED_ORIGINS=*`
   - Recommendation: Specify exact origins

## Railway-Specific Configuration

Railway provides database variables in a different format:

```bash
# Railway provides:
DATABASE_HOST=containers-us-west-123.railway.app
DATABASE_NAME=railway
DATABASE_USER=postgres
DATABASE_PASSWORD=generated_password

# Application automatically uses these via fallback logic
```

The application supports both `DATABASE_*` and `DB_*` prefixes for maximum compatibility.

## Configuration Best Practices

### ✅ DO

- ✅ Use environment variables for ALL configuration
- ✅ Use strong, randomly generated secrets in production
- ✅ Set `SERVER_ENV=production` in production
- ✅ Use `DATABASE_SSL_MODE=require` in production
- ✅ Specify exact CORS origins in production
- ✅ Store secrets in secret managers (AWS Secrets Manager, Vault, etc.)
- ✅ Use different JWT secrets per environment
- ✅ Keep JWT expiration short in production (1-2 hours)

### ❌ DON'T

- ❌ Commit `.env` files with real secrets to version control
- ❌ Use default passwords in production
- ❌ Use `CORS_ALLOWED_ORIGINS=*` in production
- ❌ Disable database SSL in production
- ❌ Use weak JWT secrets (<32 characters)
- ❌ Share secrets between environments
- ❌ Hardcode configuration in application code

## Migrating from YAML Configuration

**The `config.yaml.example` file is deprecated and NOT used by the application.**

If migrating from YAML configuration:

1. Convert YAML values to environment variables using the table above
2. Use `.env.example` as a template
3. Delete or ignore `config.yaml` files

Example migration:

```yaml
# Old YAML (NOT LOADED):
jwt:
  expiration: 86400  # seconds

# New Environment Variables:
JWT_EXPIRATION_SECONDS=86400
```

## Troubleshooting

### Configuration Not Loading

**Problem:** Environment variables not being read

**Solution:**
1. Verify `.env` file exists in project root (for development)
2. Check environment variable names (case-sensitive)
3. Restart application after changing environment variables
4. Check application logs for configuration warnings

### Production Validation Failures

**Problem:** Application fails to start with configuration error

**Solution:**
1. Check error message for specific validation failure
2. Verify all required production variables are set
3. Ensure JWT secret is at least 32 characters
4. Ensure database password is not empty or default

### JWT Expiration Issues

**Problem:** JWT tokens expiring too quickly/slowly

**Solution:**
1. Check `JWT_EXPIRATION_SECONDS` value
2. Verify it's in seconds (not hours)
3. Common values:
   - Development: `86400` (24 hours)
   - Production: `3600` (1 hour)

### Database Connection Failures

**Problem:** Cannot connect to database

**Solution:**
1. Verify `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_NAME`
2. Check database credentials (`DATABASE_USER`, `DATABASE_PASSWORD`)
3. For Railway: Use `DATABASE_*` variables (automatically set)
4. Check database SSL mode matches server requirements

## Reference

### Default Values Summary

| Variable | Development Default | Production Requirement |
|----------|-------------------|----------------------|
| `JWT_SECRET` | **REQUIRED** | **REQUIRED** (min 32 chars) |
| `DATABASE_PASSWORD` | `"postgres"` | **STRONG PASSWORD REQUIRED** |
| `AI_API_KEY` | `""` (empty) | **REQUIRED** |
| `DATABASE_SSL_MODE` | `"disable"` | `"require"` recommended |
| `CORS_ALLOWED_ORIGINS` | `"*"` | Specific origins required |
| `GRACEFUL_SHUTDOWN_TIMEOUT` | `30s` | `60s` recommended |
| `JWT_EXPIRATION_SECONDS` | `86400` (24h) | `3600` (1h) recommended |

### Configuration Utilities

The `config` package provides utility functions for parsing:

```go
// String with default
value := getEnv("KEY", "default")

// Integer with default
count := getEnvInt("COUNT", 100)

// Boolean with default
enabled := getEnvBool("ENABLED", false)

// Duration with default
timeout := getEnvDuration("TIMEOUT", 30*time.Second)
```

## See Also

- [Production Readiness](./PRODUCTION_READINESS.md)
- [Deployment Guide](./DEPLOYMENT.md)
- [Security Guide](./SECURITY.md)
