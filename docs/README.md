# Learnify API Documentation

Complete API documentation for the Learnify backend service.

## Quick Links

### Production Readiness & Architecture
- **Executive Summary:** [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md) - Production readiness overview
- **Architecture Review:** [`architecture-review.md`](architecture-review.md) - Comprehensive technical analysis
- **Production Readiness Report:** [`production-readiness-report.md`](production-readiness-report.md) - Deployment assessment
- **Performance Benchmarks:** [`benchmarks.md`](benchmarks.md) - Load testing and optimization
- **Security Audit:** [`security-audit.md`](security-audit.md) - Security assessment
- **Observability:** [`observability.md`](observability.md) - Monitoring and logging setup
- **Database Operations:** [`database-operations.md`](database-operations.md) - Database management
- **CI/CD:** [`cicd.md`](cicd.md) - Deployment automation

### API Documentation
- **OpenAPI Specification:** [`../openapi.yaml`](../openapi.yaml)
- **Postman Collection:** [`../postman_collection.json`](../postman_collection.json)
- **API Usage Guide:** [`api-guide.md`](api-guide.md)
- **API Examples:** [`api-examples.md`](api-examples.md)
- **Schema Documentation:** [`api-schemas.md`](api-schemas.md)
- **Versioning Strategy:** [`api-versioning.md`](api-versioning.md)

## Documentation Overview

### 📄 OpenAPI Specification (`openapi.yaml`)

Complete OpenAPI 3.0 specification with:
- All 20+ API endpoints
- Request/response schemas
- Authentication (JWT Bearer)
- Error responses
- Example values
- Tags for organization

**View in Swagger UI:**
1. Visit [Swagger Editor](https://editor.swagger.io/)
2. File > Import File > Select `backend/openapi.yaml`
3. Explore interactive documentation

**Or use Swagger UI locally:**
```bash
npx swagger-ui-watcher openapi.yaml
```

### 📮 Postman Collection (`postman_collection.json`)

Ready-to-use Postman collection featuring:
- All API endpoints organized by category
- Pre-configured environment variables
- Automatic token extraction from auth responses
- Test scripts for response validation
- Example requests with realistic data

**Import to Postman:**
1. Open Postman
2. Import > Upload Files > Select `backend/postman_collection.json`
3. Configure environment variables (see below)

**Environment Variables:**
- `base_url` - API base URL (default: `http://localhost:8080`)
- `auth_token` - JWT token (auto-populated after login)
- `user_id` - Current user ID (auto-populated)
- `course_id` - Course ID for testing
- `exercise_id` - Exercise ID for testing
- `submission_id` - Submission ID for testing

### 📖 API Usage Guide (`api-guide.md`)

Comprehensive guide covering:
- Getting started with the API
- Authentication flows (registration & login)
- Common workflows (onboarding, courses, exercises)
- Rate limiting (planned)
- Error handling strategies
- CORS configuration
- Best practices for API clients

### 💡 API Examples (`api-examples.md`)

Complete curl examples for every endpoint:
- Authentication (register, login)
- User management (profile, onboarding)
- Courses (list, details, progress)
- Exercises (get, submit, review)
- Social features (follow, feed, recommendations)
- Achievements

Each example includes:
- Full curl command with headers
- Request body (if applicable)
- Success response (200/201)
- Error responses (400/401/404/500)

### 🗂️ Schema Documentation (`api-schemas.md`)

Detailed documentation of all data models:
- Field names and types
- Required vs optional fields
- Validation rules
- Allowed values (enums)
- Data ranges
- Example JSON

Organized by domain:
- User & Identity
- Learning Domain
- Social Domain
- Common Types

### 🔄 Versioning Strategy (`api-versioning.md`)

API versioning policy:
- Current version: v1 (unversioned paths)
- URI-based versioning approach
- Breaking vs non-breaking changes
- Deprecation policy (6 months notice)
- Migration guides
- Version lifecycle

## API Endpoint Summary

### Authentication (Public)
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login

### User Management (Protected)
- `GET /api/users/me` - Get current user profile
- `PATCH /api/users/me` - Update user profile
- `POST /api/onboarding/complete` - Complete onboarding

### Courses (Protected)
- `GET /api/courses` - Get user's courses
- `GET /api/courses/{id}` - Get course details
- `GET /api/courses/{id}/progress` - Get course progress

### Exercises (Protected)
- `GET /api/exercises/{id}` - Get exercise details
- `POST /api/exercises/{id}/submit` - Submit exercise
- `POST /api/submissions/{id}/review` - Request AI review

### Social (Protected)
- `GET /api/feed` - Get activity feed
- `POST /api/users/{id}/follow` - Follow user
- `DELETE /api/users/{id}/follow` - Unfollow user
- `GET /api/recommendations` - Get course recommendations
- `GET /api/users/{id}/profile` - Get user profile
- `GET /api/users/me/achievements` - Get achievements

### Social (Public)
- `GET /api/trending` - Get trending courses

### Health (Public)
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness check
- `GET /metrics` - Prometheus metrics

## Authentication

All protected endpoints require JWT Bearer token:

```bash
Authorization: Bearer <your-jwt-token>
```

**Get a token:**
1. Register: `POST /api/auth/register`
2. Or Login: `POST /api/auth/login`
3. Extract `token` from response
4. Use in all subsequent requests

**Token Expiration:** 24 hours

## Quick Start

### 1. Start the API Server

```bash
cd backend
go run cmd/api/main.go
```

Server runs at: `http://localhost:8080`

### 2. Register a User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "name": "Test User"
  }'
```

### 3. Save the Token

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { ... }
}
```

### 4. Make Authenticated Requests

```bash
export TOKEN="your-token-here"

curl -X GET http://localhost:8080/api/courses \
  -H "Authorization: Bearer $TOKEN"
```

## Testing with Swagger UI

### Option 1: Online Swagger Editor

1. Go to https://editor.swagger.io/
2. File > Import File
3. Select `backend/openapi.yaml`
4. Click "Authorize" button
5. Enter your JWT token
6. Try out endpoints interactively

### Option 2: Local Swagger UI

```bash
# Install swagger-ui-watcher
npm install -g swagger-ui-watcher

# Start Swagger UI
cd backend
npx swagger-ui-watcher openapi.yaml
```

Opens at: `http://localhost:3001`

## Testing with Postman

1. **Import Collection:**
   - Open Postman
   - Import > `backend/postman_collection.json`

2. **Set Environment:**
   - Create new environment "Learnify Local"
   - Add variable: `base_url` = `http://localhost:8080`

3. **Run Authentication Flow:**
   - Run "Register New User" (auto-saves token)
   - Or "Login" (auto-saves token)

4. **Test Protected Endpoints:**
   - Token is automatically included
   - Run any protected endpoint

5. **Run Tests:**
   - Each request has validation tests
   - View results in Test Results tab

## Common Workflows

### Complete Onboarding Flow

1. Register → 2. Complete Onboarding → 3. Get Courses

```bash
# 1. Register
TOKEN=$(curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"pass123","name":"User"}' \
  | jq -r '.token')

# 2. Complete Onboarding
curl -X POST http://localhost:8080/api/onboarding/complete \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "meta_category": "programming",
    "domain": "web-development",
    "skill_level": "beginner",
    "variables": {"preferred_language": "javascript"}
  }'

# 3. Get Courses
curl -X GET http://localhost:8080/api/courses \
  -H "Authorization: Bearer $TOKEN"
```

### Complete Exercise Flow

1. Get Courses → 2. Get Course Details → 3. Get Exercise → 4. Submit → 5. Review

```bash
# 1-2. Get course and extract first exercise ID
EXERCISE_ID=$(curl -X GET http://localhost:8080/api/courses/COURSE_ID \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data.modules[0].exercises[0].id')

# 3. Get Exercise
curl -X GET http://localhost:8080/api/exercises/$EXERCISE_ID \
  -H "Authorization: Bearer $TOKEN"

# 4. Submit Solution
SUBMISSION_ID=$(curl -X POST http://localhost:8080/api/exercises/$EXERCISE_ID/submit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code":"...","language":"javascript"}' \
  | jq -r '.data.id')

# 5. Request Review
curl -X POST http://localhost:8080/api/submissions/$SUBMISSION_ID/review \
  -H "Authorization: Bearer $TOKEN"
```

## Error Handling

### HTTP Status Codes

| Code | Meaning | Action |
|------|---------|--------|
| 200 | Success | Continue |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Fix request body/parameters |
| 401 | Unauthorized | Re-authenticate (login again) |
| 404 | Not Found | Check resource ID |
| 409 | Conflict | Resource already exists |
| 500 | Server Error | Retry or contact support |

### Error Response Format

```json
{
  "error": "error message"
}
```

Or (learning endpoints):

```json
{
  "error": "Bad Request",
  "message": "specific error details"
}
```

## Data Model Overview

### Key Entities

- **User** - User account with profile and settings
- **Course** - Personalized learning course
- **Module** - Section within a course
- **Exercise** - Coding challenge
- **Submission** - User's exercise solution
- **Review** - AI-generated code review
- **Achievement** - Unlockable badge
- **Activity** - Social feed item
- **Recommendation** - Personalized course suggestion

### Relationships

```
User
  ├── Courses (1:many)
  │   └── Modules (1:many)
  │       └── Exercises (1:many)
  │           └── Submissions (1:many)
  │               └── Reviews (1:1)
  ├── Achievements (many:many)
  ├── Activities (1:many)
  └── Recommendations (1:many)
```

## Support Resources

### Documentation Files

- OpenAPI Spec: Complete API reference
- Postman Collection: Interactive testing
- Usage Guide: Getting started and workflows
- Examples: Copy-paste curl commands
- Schemas: Data model reference
- Versioning: Migration guides

### External Resources

- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.0)
- [JWT.io](https://jwt.io/) - Decode JWT tokens
- [Swagger Editor](https://editor.swagger.io/)
- [Postman Learning Center](https://learning.postman.com/)

### Getting Help

For issues or questions:
1. Check the relevant documentation file
2. Review API examples for your use case
3. Validate requests against OpenAPI spec
4. Test with Postman collection
5. Contact: api-support@learnify.example.com

## Contributing

When making API changes:

1. Update `openapi.yaml` specification
2. Add examples to `api-examples.md`
3. Update schema docs in `api-schemas.md`
4. Add/update Postman requests
5. Document breaking changes in `api-versioning.md`
6. Update this README if needed

## Validation

### Validate OpenAPI Spec

```bash
# Using Swagger CLI
npm install -g @apidevtools/swagger-cli
swagger-cli validate openapi.yaml

# Using Spectral
npm install -g @stoplight/spectral-cli
spectral lint openapi.yaml
```

### Test All Endpoints

Use the Postman collection with Newman:

```bash
npm install -g newman
newman run postman_collection.json \
  --environment postman_environment.json
```

## License

MIT License - See LICENSE file for details
