# Testing Checklist

Use this checklist to systematically test all features of the Learnify backend.

## Pre-Testing Setup

- [ ] Docker Desktop is running
- [ ] `.env` file created with OpenAI API key
- [ ] Services started: `docker-compose up -d`
- [ ] Health check passes: `curl http://localhost:8080/health`
- [ ] Postman collection imported (optional but recommended)

---

## 1. Authentication Flow

### Registration
- [ ] **Register new user** (POST `/api/auth/register`)
  - [ ] With valid email, password, username
  - [ ] Save the returned JWT token
  - [ ] Verify returns user object and token
  - [ ] Try registering same email again (should fail)
  - [ ] Try weak password (should fail)
  - [ ] Try invalid email format (should fail)

### Login
- [ ] **Login with created user** (POST `/api/auth/login`)
  - [ ] With correct email and password
  - [ ] Verify returns JWT token
  - [ ] Save token for authenticated requests
  - [ ] Try wrong password (should fail with 401)
  - [ ] Try non-existent email (should fail with 401)

---

## 2. User Profile Management

### Get Profile
- [ ] **Fetch current user profile** (GET `/api/users/me`)
  - [ ] With valid JWT token in Authorization header
  - [ ] Verify returns user details (id, email, username)
  - [ ] Try without token (should fail with 401)
  - [ ] Try with invalid token (should fail with 401)

### Update Profile
- [ ] **Update profile** (PATCH `/api/users/me`)
  - [ ] Update username
  - [ ] Update bio
  - [ ] Update archetype
  - [ ] Verify changes persist
  - [ ] Try with invalid data (should validate)

---

## 3. Onboarding Flow

### Complete Onboarding
- [ ] **Submit onboarding data** (POST `/api/onboarding/complete`)
  - [ ] With domain and archetype
  - [ ] With 5 universal variables (ENTITY, STATE, FLOW, LOGIC, INTERFACE)
  - [ ] Verify data saved to user profile
  - [ ] Check variable extraction works
  - [ ] Verify courses generated based on selections

**Example payload:**
```json
{
  "domain": "e-commerce",
  "archetype": "code-craftsperson",
  "variables": {
    "ENTITY": "Product",
    "STATE": "Inventory",
    "FLOW": "Checkout",
    "LOGIC": "PricingEngine",
    "INTERFACE": "ProductCatalog"
  }
}
```

---

## 4. Course Management

### List Courses
- [ ] **Get user's courses** (GET `/api/courses`)
  - [ ] With authentication
  - [ ] Verify returns array of courses
  - [ ] Check courses match user's onboarding selections
  - [ ] Verify pagination works (if implemented)

### Get Course Details
- [ ] **Fetch single course** (GET `/api/courses/{id}`)
  - [ ] With valid course ID
  - [ ] Verify returns full course details
  - [ ] Check exercises are included
  - [ ] Verify metadata (difficulty, duration, etc.)
  - [ ] Try with invalid ID (should 404)

### View Trending Courses
- [ ] **Public trending endpoint** (GET `/api/trending`)
  - [ ] Without authentication (public)
  - [ ] Verify returns popular courses
  - [ ] Check sorted by popularity metrics
  - [ ] Verify includes enrollment counts

---

## 5. Exercise Submission & Testing

### Get Exercise Details
- [ ] **Fetch exercise** (GET `/api/exercises/{id}`)
  - [ ] With valid exercise ID
  - [ ] Verify returns exercise description
  - [ ] Check test cases included
  - [ ] Verify starter code present (if applicable)

### Submit Code Solution
- [ ] **Submit solution** (POST `/api/exercises/{id}/submit`)
  - [ ] With valid code
  - [ ] Verify runs test cases
  - [ ] Check returns test results (pass/fail)
  - [ ] Verify execution output captured
  - [ ] Verify submission saved to database
  - [ ] Try with syntax errors (should handle gracefully)

**Example payload:**
```json
{
  "code": "function add(a, b) { return a + b; }",
  "language": "javascript"
}
```

---

## 6. AI Senior Review

### Request Code Review
- [ ] **Request AI review** (POST `/api/submissions/{id}/review`)
  - [ ] With valid submission ID
  - [ ] Verify AI review generated
  - [ ] Check 4-category scoring:
    - [ ] Correctness
    - [ ] Architecture
    - [ ] Performance
    - [ ] Maintainability
  - [ ] Verify includes improvement suggestions
  - [ ] Check response time reasonable (<30s)
  - [ ] Verify review saved and linked to submission

**Note:** Requires valid OpenAI API key in `.env`

---

## 7. Progress Tracking

### View Course Progress
- [ ] **Get progress** (GET `/api/courses/{id}/progress`)
  - [ ] With valid course ID
  - [ ] Verify shows completed exercises
  - [ ] Check completion percentage accurate
  - [ ] Verify timestamps for submissions
  - [ ] Check streak tracking (if implemented)

---

## 8. Social Features

### Activity Feed
- [ ] **View feed** (GET `/api/feed`)
  - [ ] With authentication
  - [ ] Verify shows recent activities
  - [ ] Check includes followed users' activities
  - [ ] Verify pagination works
  - [ ] Check sorted by timestamp (newest first)

### Follow/Unfollow Users
- [ ] **Follow a user** (POST `/api/users/{id}/follow`)
  - [ ] With valid user ID
  - [ ] Verify follow relationship created
  - [ ] Check feed updates to show their activity
  - [ ] Try following already-followed user (should handle)
  - [ ] Try following self (should fail)

- [ ] **Unfollow a user** (DELETE `/api/users/{id}/follow`)
  - [ ] With valid user ID
  - [ ] Verify relationship removed
  - [ ] Check feed no longer shows their activity

### Recommendations
- [ ] **Get recommendations** (GET `/api/recommendations`)
  - [ ] With authentication
  - [ ] Verify returns personalized courses
  - [ ] Check based on user's profile/history
  - [ ] Verify diversity of recommendations
  - [ ] Check not recommending already-enrolled courses

### User Profile (Public)
- [ ] **View other user's profile** (GET `/api/users/{id}/profile`)
  - [ ] With valid user ID
  - [ ] Verify shows public profile info
  - [ ] Check "Living Resume" data:
    - [ ] Completed courses
    - [ ] Achievements
    - [ ] Skills/expertise
    - [ ] Activity summary
  - [ ] Try with invalid ID (should 404)

### Achievements
- [ ] **Get user achievements** (GET `/api/users/me/achievements`)
  - [ ] With authentication
  - [ ] Verify returns earned badges/achievements
  - [ ] Check includes timestamps
  - [ ] Verify achievement criteria displayed
  - [ ] Check locked achievements shown

---

## 9. Infrastructure & Monitoring

### Health Checks
- [ ] **Liveness check** (GET `/health`)
  - [ ] No authentication required
  - [ ] Returns 200 OK when healthy
  - [ ] Returns service status

- [ ] **Readiness check** (GET `/health/ready`)
  - [ ] Checks database connection
  - [ ] Checks AI service connectivity
  - [ ] Returns detailed status of dependencies

### Metrics
- [ ] **Prometheus metrics** (GET `/metrics`)
  - [ ] No authentication required
  - [ ] Returns metrics in Prometheus format
  - [ ] Check includes:
    - [ ] HTTP request counts
    - [ ] Response times
    - [ ] Error rates
    - [ ] Database connection pool stats

---

## 10. Error Handling & Edge Cases

### Invalid Requests
- [ ] Missing required fields (should return 400)
- [ ] Invalid JSON format (should return 400)
- [ ] Invalid data types (should return 400)
- [ ] SQL injection attempts (should be sanitized)
- [ ] XSS attempts (should be sanitized)

### Authentication Errors
- [ ] Missing Authorization header (should return 401)
- [ ] Invalid JWT token (should return 401)
- [ ] Expired JWT token (should return 401)
- [ ] Malformed Bearer token (should return 401)

### Rate Limiting
- [ ] Make rapid repeated requests
- [ ] Verify rate limiting kicks in (should return 429)
- [ ] Verify rate limit resets after time window

### Large Payloads
- [ ] Submit very large code (>1MB)
- [ ] Verify request size limit enforced
- [ ] Should return 413 Payload Too Large

---

## 11. Performance & Load Testing

### Response Times
- [ ] Health check: <50ms
- [ ] Registration: <200ms
- [ ] Login: <200ms
- [ ] Get courses: <300ms
- [ ] Submit code: <2000ms (depends on tests)
- [ ] AI review: <30000ms (depends on AI provider)

### Concurrent Users
- [ ] 10 concurrent registrations
- [ ] 50 concurrent course fetches
- [ ] 20 concurrent submissions
- [ ] Verify no errors or timeouts

**Tool:** Use Apache Bench or Postman Collection Runner

---

## 12. Database & Data Integrity

### Data Persistence
- [ ] Register user → Stop services → Restart → Login
  - [ ] User data persists
- [ ] Submit code → Restart → Check submission history
  - [ ] Submissions persist
- [ ] Earn achievement → Restart → Check achievements
  - [ ] Achievements persist

### Transactions
- [ ] Submit code with DB temporarily unavailable
  - [ ] Should handle gracefully
  - [ ] Should not corrupt data
- [ ] Simultaneous submissions to same exercise
  - [ ] All should succeed
  - [ ] No race conditions

---

## 13. Security Testing

### Authentication
- [ ] Access protected endpoint without token (should 401)
- [ ] Use another user's token to access resources (should 403)
- [ ] Token stolen from another user (should fail CSRF protection)

### Input Validation
- [ ] SQL injection in email field
- [ ] Script tags in username
- [ ] Path traversal in file paths
- [ ] Command injection in code submission

**All should be properly sanitized/rejected**

### CORS
- [ ] Request from allowed origin (should succeed)
- [ ] Request from different origin (should be blocked)
- [ ] Verify CORS headers present

---

## Test Summary

After completing all tests, summarize:

- **Total Tests Run:** ___
- **Passed:** ___
- **Failed:** ___
- **Blocked:** ___ (due to missing setup/dependencies)

### Critical Issues Found:

1.
2.
3.

### Minor Issues Found:

1.
2.
3.

### Suggested Improvements:

1.
2.
3.

---

## Notes for Developers

Document any unexpected behavior here:

-
-
-

---

**Testing completed by:** ________________
**Date:** ________________
**Environment:** Development / Staging / Production
**API Version:** ________________
