# Testing Guide for Learnify Backend

Quick setup guide for testers to get the backend running locally.

## Prerequisites

Before you start, make sure you have installed:
- **Docker Desktop** (includes Docker Compose)
  - Windows/Mac: https://www.docker.com/products/docker-desktop
  - Verify: `docker --version` and `docker-compose --version`
- **Git** (to clone the repository)
- **curl** or **Postman** (for testing API endpoints)

**Optional but helpful:**
- **Make** (for easier commands)
  - Windows: `choco install make` or use Git Bash
  - Mac: Pre-installed with Xcode Command Line Tools

## Quick Start (4 Steps)

### Step 1: Clone and Navigate

```bash
git clone <repository-url>
cd backend
```

### Step 2: Configure Environment

```bash
# Copy the example environment file
cp .env.example .env

# Open .env and add your OpenAI API key (required for AI code reviews)
# Find this line and replace with your key:
# AI_API_KEY=your-openai-api-key-here
```

**Where to get an OpenAI API key:**
1. Go to https://platform.openai.com/api-keys
2. Sign up or log in
3. Create a new API key
4. Copy and paste into `.env`

**Note:** If you don't have an API key yet, you can still test most features. Only AI code review will fail.

### Step 3: Start Services

```bash
# Start PostgreSQL database and API server
docker-compose up -d

# This will:
# 1. Download PostgreSQL 16 image (first time only)
# 2. Start database container
# 3. Build and start API container
# 4. Run database migrations automatically
```

**Wait about 30 seconds for services to fully start.**

### Step 4: Verify Everything Works

```bash
# Check services are running
docker-compose ps

# Test health endpoint
curl http://localhost:8080/health

# Expected response: {"status":"ok","timestamp":"..."}
```

## Testing the API

### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123",
    "username": "testuser"
  }'
```

**Expected response:**
```json
{
  "user": {
    "id": "...",
    "email": "test@example.com",
    "username": "testuser"
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Save the token** - you'll need it for authenticated requests!

### 2. Login (Get New Token)

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123"
  }'
```

### 3. Get Your Profile (Authenticated)

```bash
# Replace YOUR_TOKEN_HERE with the token from registration/login
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 4. View Available Courses

```bash
curl -X GET http://localhost:8080/api/courses \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 5. Check Trending Courses (Public)

```bash
curl -X GET http://localhost:8080/api/trending
```

## Load Test Data (Optional)

To see the platform with sample courses and exercises:

```bash
# Load seed data into database
docker-compose exec postgres psql -U postgres -d learnify -f /docker-entrypoint-initdb.d/seed_test_data.sql
```

This creates:
- 3 sample courses (Web Dev, Python Basics, System Design)
- 5 exercises per course
- Sample test cases

## Using Postman (Easier Testing)

1. **Import the collection:**
   - Open Postman
   - Click Import → File
   - Select `backend/postman_collection.json`

2. **Set up environment:**
   - Create a new environment
   - Add variable: `base_url` = `http://localhost:8080`
   - Add variable: `token` = (copy token after registration)

3. **Test endpoints:**
   - All requests are pre-configured
   - Update `{{token}}` after registration
   - Run requests in sequence

## Common Issues & Fixes

### Port 8080 Already in Use

**Error:** `bind: address already in use`

**Solution:**
```bash
# Find what's using port 8080
# Windows:
netstat -ano | findstr :8080

# Mac/Linux:
lsof -i :8080

# Either kill that process or change API port in docker-compose.yml
```

### Port 5432 Already in Use

**Error:** `port is already allocated`

**Solution:**
```bash
# Stop any existing PostgreSQL instance
# Or change port in docker-compose.yml:
# ports:
#   - "5433:5432"  # Map to 5433 on host
```

### Database Connection Failed

**Error:** `failed to connect to database`

**Solution:**
```bash
# Check database is running
docker-compose ps

# View database logs
docker-compose logs postgres

# Restart services
docker-compose down && docker-compose up -d
```

### AI API Key Error

**Error:** `AI provider not configured` or `401 Unauthorized`

**Solution:**
- Make sure you added your OpenAI API key to `.env`
- Restart services: `docker-compose restart api`
- Check you have API credits at https://platform.openai.com/account/usage

### Migrations Not Running

**Error:** Tables not found

**Solution:**
```bash
# Manually run migrations
docker-compose exec postgres psql -U postgres -d learnify -f /docker-entrypoint-initdb.d/001_create_identity_tables.sql

# Or reset database completely
docker-compose down -v  # WARNING: Deletes all data
docker-compose up -d
```

## Viewing Logs

```bash
# View all logs
docker-compose logs -f

# View just API logs
docker-compose logs -f api

# View just database logs
docker-compose logs -f postgres
```

## Stopping Services

```bash
# Stop but keep data
docker-compose stop

# Stop and remove containers (keeps data)
docker-compose down

# Stop and delete ALL data (fresh start)
docker-compose down -v
```

## Database Access

To inspect the database directly:

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U postgres -d learnify

# Useful queries:
\dt                           # List tables
SELECT * FROM users;          # View users
SELECT * FROM courses;        # View courses
\q                            # Exit
```

## API Documentation

**OpenAPI Spec:** `backend/openapi.yaml`

View in Swagger UI:
1. Go to https://editor.swagger.io/
2. File → Import File → Select `openapi.yaml`
3. Browse interactive API documentation

## Performance Testing

Test API performance:

```bash
# Install Apache Bench (if not installed)
# Windows: Download from Apache website
# Mac: brew install ab

# Run 1000 requests with 10 concurrent
ab -n 1000 -c 10 http://localhost:8080/health
```

## Metrics & Monitoring

View Prometheus metrics:

```bash
curl http://localhost:8080/metrics
```

This shows:
- Request counts
- Response times
- Error rates
- Database connections

## Getting Help

If you encounter issues:

1. Check logs: `docker-compose logs -f api`
2. Verify services: `docker-compose ps`
3. Review this guide's "Common Issues" section
4. Check `backend/docs/` for detailed documentation
5. Contact the developer with:
   - Error message
   - Steps to reproduce
   - Output of `docker-compose logs api`

## What to Test

See [TESTING_CHECKLIST.md](./TESTING_CHECKLIST.md) for a comprehensive list of test scenarios.

## Clean Up After Testing

```bash
# Stop everything
docker-compose down

# Remove all data (start fresh next time)
docker-compose down -v

# Remove images (free up space)
docker-compose down --rmi all
```

---

**Happy Testing!** 🚀

For detailed API documentation, see `docs/api-guide.md` and `openapi.yaml`.
