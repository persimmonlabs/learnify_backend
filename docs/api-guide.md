# Learnify API Usage Guide

## Table of Contents

- [Getting Started](#getting-started)
- [Authentication](#authentication)
- [Common Workflows](#common-workflows)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [Pagination](#pagination)
- [CORS Configuration](#cors-configuration)
- [Best Practices](#best-practices)

## Getting Started

### Base URL

**Development:**
```
http://localhost:8080
```

**Production:**
```
https://api.learnify.example.com
```

### Quick Start

1. **Register a new user**
2. **Receive JWT token**
3. **Complete onboarding**
4. **Start learning!**

## Authentication

### Overview

The Learnify API uses JWT (JSON Web Tokens) for authentication. Once you register or login, you'll receive a token that must be included in all subsequent requests to protected endpoints.

### Registration Flow

**Step 1: Register a new account**

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "name": "John Doe"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-01-21T10:30:00Z"
  }
}
```

### Login Flow

**Login with existing credentials**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "name": "John Doe",
    "last_login": "2025-01-21T14:45:00Z"
  }
}
```

### Using JWT Token

Include the token in the `Authorization` header for all protected endpoints:

```bash
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Token Expiration

- **JWT tokens expire after 24 hours**
- **When a token expires, you'll receive a 401 Unauthorized response**
- **Solution: Login again to get a new token**

```json
{
  "error": "token expired"
}
```

### Password Requirements

- Minimum 8 characters
- Recommended: Mix of uppercase, lowercase, numbers, and symbols

## Common Workflows

### 1. Complete User Onboarding

After registration, complete onboarding to generate personalized courses:

```bash
curl -X POST http://localhost:8080/api/onboarding/complete \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "meta_category": "programming",
    "domain": "web-development",
    "skill_level": "beginner",
    "variables": {
      "preferred_language": "javascript",
      "learning_goal": "build-web-apps"
    }
  }'
```

**Response:**
```json
{
  "message": "onboarding completed successfully"
}
```

### 2. Browse Your Courses

Get all courses assigned to you:

```bash
curl -X GET http://localhost:8080/api/courses \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "course-uuid-123",
      "title": "Modern Web Development with React",
      "description": "Learn to build modern web applications",
      "status": "active",
      "meta_category": "programming"
    }
  ]
}
```

### 3. Get Course Details

View modules and content for a specific course:

```bash
curl -X GET http://localhost:8080/api/courses/course-uuid-123 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "course": {
      "id": "course-uuid-123",
      "title": "Modern Web Development with React",
      "description": "Learn to build modern web applications"
    },
    "modules": [
      {
        "id": "module-uuid-456",
        "module_number": 1,
        "title": "Introduction to React",
        "status": "unlocked"
      }
    ]
  }
}
```

### 4. Complete an Exercise

**Step 1: Get exercise details**

```bash
curl -X GET http://localhost:8080/api/exercises/exercise-uuid-789 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Step 2: Submit your solution**

```bash
curl -X POST http://localhost:8080/api/exercises/exercise-uuid-789/submit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "import React, { useState } from \"react\";\n\nexport default function Counter() {\n  const [count, setCount] = useState(0);\n  return (\n    <div>\n      <p>Count: {count}</p>\n      <button onClick={() => setCount(count + 1)}>+</button>\n      <button onClick={() => setCount(count - 1)}>-</button>\n    </div>\n  );\n}",
    "language": "javascript"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "submission-uuid",
    "passed": true,
    "score": 95,
    "attempts": 1,
    "submitted_at": "2025-01-21T15:30:00Z"
  }
}
```

### 5. Request AI Code Review

Get detailed feedback on your submission:

```bash
curl -X POST http://localhost:8080/api/submissions/submission-uuid/review \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "overall_score": 88,
    "code_sense_score": 90,
    "efficiency_score": 85,
    "edge_cases_score": 92,
    "taste_score": 86,
    "feedback": {
      "strengths": [
        "Clean component structure",
        "Proper use of React hooks"
      ],
      "improvements": [
        "Consider adding prop validation",
        "Extract button logic to separate component"
      ],
      "suggestions": "Great job! Consider learning about useReducer for more complex state management."
    }
  }
}
```

### 6. Track Your Progress

Check progress for a specific course:

```bash
curl -X GET http://localhost:8080/api/courses/course-uuid-123/progress \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "course_id": "course-uuid-123",
    "progress_percentage": 45,
    "time_spent_minutes": 320,
    "current_module_id": "module-uuid-456",
    "started_at": "2025-01-15T10:00:00Z"
  }
}
```

### 7. Social Features

**Follow a user:**

```bash
curl -X POST http://localhost:8080/api/users/user-uuid-999/follow \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Get activity feed:**

```bash
curl -X GET "http://localhost:8080/api/feed?limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Get recommendations:**

```bash
curl -X GET http://localhost:8080/api/recommendations \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "recommendations": [
    {
      "course_id": "rec-course-uuid",
      "recommendation_type": "skill_adjacency",
      "match_score": 87,
      "reason": "Based on your JavaScript skills, you'll love TypeScript"
    }
  ],
  "sections": {
    "collaborative_filtering": "Because You Completed",
    "skill_adjacency": "Next Level Skills",
    "social_signal": "Friends Are Learning",
    "trending": "Trending Now"
  }
}
```

### 8. View Trending Courses

Public endpoint - no authentication required:

```bash
curl -X GET http://localhost:8080/api/trending
```

## Rate Limiting

### Current Limits

**Development Environment:**
- No rate limiting currently enforced

**Production Environment (Planned):**
- **Authenticated requests:** 1000 requests per hour per user
- **Unauthenticated requests:** 100 requests per hour per IP
- **Burst limit:** 20 requests per second

### Rate Limit Headers

Future implementation will include these headers:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1737468000
```

### Rate Limit Exceeded Response

```json
{
  "error": "rate limit exceeded",
  "retry_after": 3600
}
```

**HTTP Status:** 429 Too Many Requests

## Error Handling

### Error Response Format

All errors follow a consistent format:

```json
{
  "error": "error message here"
}
```

Or for learning endpoints:

```json
{
  "error": "Bad Request",
  "message": "specific error details"
}
```

### Common HTTP Status Codes

| Status | Meaning | When It Occurs |
|--------|---------|----------------|
| 200    | OK      | Successful GET, DELETE |
| 201    | Created | Successful POST (resource created) |
| 400    | Bad Request | Invalid request body or parameters |
| 401    | Unauthorized | Missing or invalid JWT token |
| 403    | Forbidden | Valid token but insufficient permissions |
| 404    | Not Found | Resource doesn't exist |
| 409    | Conflict | Resource already exists (e.g., email taken) |
| 500    | Internal Server Error | Server-side error |

### Error Handling Examples

**Invalid Request Body:**
```json
{
  "error": "invalid request body"
}
```

**Validation Error:**
```json
{
  "error": "password must be at least 8 characters"
}
```

**Resource Not Found:**
```json
{
  "error": "course not found"
}
```

**Unauthorized:**
```json
{
  "error": "unauthorized"
}
```

### Best Practices for Error Handling

```javascript
async function makeApiCall(url, options) {
  try {
    const response = await fetch(url, options);

    // Check if response is OK
    if (!response.ok) {
      const error = await response.json();

      // Handle specific status codes
      switch (response.status) {
        case 401:
          // Token expired - redirect to login
          window.location.href = '/login';
          break;
        case 404:
          throw new Error('Resource not found');
        case 500:
          throw new Error('Server error - please try again later');
        default:
          throw new Error(error.error || error.message || 'API Error');
      }
    }

    return await response.json();
  } catch (error) {
    console.error('API Error:', error);
    throw error;
  }
}
```

## Pagination

### Current Implementation

**Pagination is not currently implemented** for most endpoints. All results are returned in a single response.

### Feed Endpoint Limiting

The activity feed endpoint supports a `limit` query parameter:

```bash
curl -X GET "http://localhost:8080/api/feed?limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Parameters:**
- `limit`: Maximum number of activities to return (default: 50, max: 100)

### Future Pagination (Planned)

Future versions will support cursor-based pagination:

```json
{
  "data": [...],
  "pagination": {
    "next_cursor": "cursor-token-abc123",
    "has_more": true
  }
}
```

## CORS Configuration

### Allowed Origins

The API includes CORS middleware that allows cross-origin requests.

**Current Configuration:**
- **Development:** All origins allowed (`*`)
- **Production:** Specific origins whitelisted

### CORS Headers

```http
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

### Preflight Requests

The API automatically handles OPTIONS preflight requests:

```bash
curl -X OPTIONS http://localhost:8080/api/courses \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET"
```

## Best Practices

### 1. Always Use HTTPS in Production

```javascript
const API_BASE = process.env.NODE_ENV === 'production'
  ? 'https://api.learnify.example.com'
  : 'http://localhost:8080';
```

### 2. Store Tokens Securely

**Good:**
```javascript
// Use httpOnly cookies (server-side)
// Or secure storage like IndexedDB with encryption
```

**Bad:**
```javascript
// Don't store in localStorage (XSS vulnerable)
localStorage.setItem('token', token); // ❌
```

### 3. Handle Token Refresh

```javascript
async function apiCall(url, options) {
  let response = await fetch(url, options);

  if (response.status === 401) {
    // Token expired - refresh or re-login
    await refreshToken();
    response = await fetch(url, options); // Retry
  }

  return response;
}
```

### 4. Implement Retry Logic

```javascript
async function fetchWithRetry(url, options, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      const response = await fetch(url, options);
      if (response.ok) return response;

      if (response.status >= 500 && i < retries - 1) {
        await sleep(1000 * Math.pow(2, i)); // Exponential backoff
        continue;
      }

      return response;
    } catch (error) {
      if (i === retries - 1) throw error;
      await sleep(1000 * Math.pow(2, i));
    }
  }
}
```

### 5. Validate Requests Client-Side

```javascript
function validateExerciseSubmission(code, language) {
  if (!code || code.trim().length === 0) {
    throw new Error('Code cannot be empty');
  }

  if (!language) {
    throw new Error('Language is required');
  }

  const allowedLanguages = ['javascript', 'python', 'java', 'go'];
  if (!allowedLanguages.includes(language)) {
    throw new Error('Unsupported language');
  }
}
```

### 6. Use Environment Variables

```javascript
// .env
REACT_APP_API_BASE_URL=http://localhost:8080
REACT_APP_API_TIMEOUT=30000

// code
const api = axios.create({
  baseURL: process.env.REACT_APP_API_BASE_URL,
  timeout: process.env.REACT_APP_API_TIMEOUT
});
```

### 7. Monitor API Usage

```javascript
// Log all API calls for debugging
const originalFetch = window.fetch;
window.fetch = function(...args) {
  console.log('API Call:', args[0]);
  return originalFetch.apply(this, args);
};
```

## Support

### Resources

- **OpenAPI Specification:** `/backend/openapi.yaml`
- **API Examples:** `/backend/docs/api-examples.md`
- **Schema Documentation:** `/backend/docs/api-schemas.md`

### Contact

- **Email:** api-support@learnify.example.com
- **GitHub:** https://github.com/learnify/api
- **Discord:** https://discord.gg/learnify

### Reporting Issues

When reporting API issues, include:
1. HTTP method and endpoint
2. Request headers and body
3. Response status and body
4. Timestamp of the request
5. Your user ID (if applicable)

Example:
```
POST /api/exercises/abc-123/submit
Headers: Authorization: Bearer xxx...
Body: {"code": "...", "language": "javascript"}
Response: 500 Internal Server Error
Time: 2025-01-21T15:30:45Z
User: user-uuid-789
```
