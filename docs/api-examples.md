# API Request/Response Examples

Complete curl examples for all Learnify API endpoints with sample data.

## Table of Contents

- [Authentication](#authentication)
- [User Management](#user-management)
- [Courses](#courses)
- [Exercises](#exercises)
- [Social Features](#social-features)
- [Achievements](#achievements)

## Authentication

### Register New User

**Request:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123!",
    "name": "Alice Johnson"
  }'
```

**Success Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzZTQ1NjctZTg5Yi0xMmQzLWE0NTYtNDI2NjE0MTc0MDAwIiwiZXhwIjoxNzM3NTU0NDAwfQ.dGVzdF90b2tlbl9zaWduYXR1cmVfaGVyZQ",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "alice@example.com",
    "name": "Alice Johnson",
    "avatar_url": "",
    "created_at": "2025-01-21T10:30:00Z",
    "updated_at": "2025-01-21T10:30:00Z",
    "last_login": "2025-01-21T10:30:00Z"
  }
}
```

**Error Response - Invalid Email (400 Bad Request):**
```json
{
  "error": "invalid email format"
}
```

**Error Response - Password Too Short (400 Bad Request):**
```json
{
  "error": "password must be at least 8 characters"
}
```

**Error Response - Email Already Exists (409 Conflict):**
```json
{
  "error": "email already registered"
}
```

### Login

**Request:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }'
```

**Success Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzZTQ1NjctZTg5Yi0xMmQzLWE0NTYtNDI2NjE0MTc0MDAwIiwiZXhwIjoxNzM3NTU0NDAwfQ.dGVzdF90b2tlbl9zaWduYXR1cmVfaGVyZQ",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "alice@example.com",
    "name": "Alice Johnson",
    "avatar_url": "",
    "created_at": "2025-01-21T10:30:00Z",
    "updated_at": "2025-01-21T10:30:00Z",
    "last_login": "2025-01-21T14:45:00Z"
  }
}
```

**Error Response - Invalid Credentials (401 Unauthorized):**
```json
{
  "error": "invalid email or password"
}
```

## User Management

### Get Current User Profile

**Request:**
```bash
curl -X GET http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "alice@example.com",
  "name": "Alice Johnson",
  "avatar_url": "https://example.com/avatars/alice.jpg",
  "privacy_settings": {
    "profile_visibility": "public",
    "activity_visibility": "friends",
    "progress_visibility": "public",
    "allow_followers": true,
    "show_in_leaderboards": true,
    "show_completed_courses": true
  },
  "created_at": "2025-01-21T10:30:00Z",
  "updated_at": "2025-01-21T10:30:00Z",
  "last_login": "2025-01-21T14:45:00Z"
}
```

**Error Response - Unauthorized (401):**
```json
{
  "error": "unauthorized"
}
```

**Error Response - User Not Found (404):**
```json
{
  "error": "user not found"
}
```

### Update User Profile

**Request:**
```bash
curl -X PATCH http://localhost:8080/api/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice J. Updated",
    "avatar_url": "https://example.com/avatars/alice-new.jpg"
  }'
```

**Success Response (200 OK):**
```json
{
  "message": "profile updated successfully"
}
```

**Error Response - Invalid Body (400):**
```json
{
  "error": "invalid request body"
}
```

### Complete Onboarding

**Request:**
```bash
curl -X POST http://localhost:8080/api/onboarding/complete \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "meta_category": "programming",
    "domain": "web-development",
    "skill_level": "beginner",
    "variables": {
      "preferred_language": "javascript",
      "learning_goal": "build-web-apps",
      "time_commitment": "10-hours-week"
    }
  }'
```

**Success Response (200 OK):**
```json
{
  "message": "onboarding completed successfully"
}
```

**Error Response - Missing Required Fields (400):**
```json
{
  "error": "meta_category, domain, and skill_level are required"
}
```

## Courses

### Get User's Courses

**Request:**
```bash
curl -X GET http://localhost:8080/api/courses \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "456e7890-e89b-12d3-a456-426614174001",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "archetype_id": "789abc12-e89b-12d3-a456-426614174002",
      "title": "Modern Web Development with React",
      "description": "Learn to build modern, responsive web applications using React, Node.js, and PostgreSQL. Master component design, state management, API integration, and deployment.",
      "meta_category": "programming",
      "status": "active",
      "created_at": "2025-01-21T11:00:00Z",
      "updated_at": "2025-01-21T11:00:00Z"
    },
    {
      "id": "567fgh90-e89b-12d3-a456-426614174003",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "archetype_id": "789abc12-e89b-12d3-a456-426614174002",
      "title": "Advanced JavaScript Patterns",
      "description": "Deep dive into advanced JavaScript concepts including closures, prototypes, async patterns, and functional programming.",
      "meta_category": "programming",
      "status": "active",
      "created_at": "2025-01-21T11:30:00Z",
      "updated_at": "2025-01-21T11:30:00Z"
    }
  ]
}
```

**Error Response - Unauthorized (401):**
```json
{
  "error": "Unauthorized",
  "message": "Unauthorized"
}
```

### Get Course Details

**Request:**
```bash
curl -X GET http://localhost:8080/api/courses/456e7890-e89b-12d3-a456-426614174001 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "course": {
      "id": "456e7890-e89b-12d3-a456-426614174001",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "archetype_id": "789abc12-e89b-12d3-a456-426614174002",
      "title": "Modern Web Development with React",
      "description": "Learn to build modern, responsive web applications",
      "meta_category": "programming",
      "status": "active",
      "created_at": "2025-01-21T11:00:00Z",
      "updated_at": "2025-01-21T11:00:00Z"
    },
    "modules": [
      {
        "id": "mod-001-uuid",
        "course_id": "456e7890-e89b-12d3-a456-426614174001",
        "blueprint_module_id": "blueprint-001",
        "module_number": 1,
        "title": "Introduction to React Components",
        "description": "Learn the fundamentals of React components, JSX, and the component lifecycle",
        "status": "unlocked",
        "unlocked_at": "2025-01-21T11:00:00Z",
        "created_at": "2025-01-21T11:00:00Z"
      },
      {
        "id": "mod-002-uuid",
        "course_id": "456e7890-e89b-12d3-a456-426614174001",
        "blueprint_module_id": "blueprint-002",
        "module_number": 2,
        "title": "State Management with Hooks",
        "description": "Master React hooks including useState, useEffect, and custom hooks",
        "status": "locked",
        "unlocked_at": null,
        "created_at": "2025-01-21T11:00:00Z"
      }
    ]
  }
}
```

**Error Response - Course Not Found (404):**
```json
{
  "error": "Not Found",
  "message": "course not found"
}
```

### Get Course Progress

**Request:**
```bash
curl -X GET http://localhost:8080/api/courses/456e7890-e89b-12d3-a456-426614174001/progress \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "progress-uuid-123",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "course_id": "456e7890-e89b-12d3-a456-426614174001",
    "current_module_id": "mod-001-uuid",
    "progress_percentage": 45,
    "time_spent_minutes": 320,
    "last_activity": "2025-01-21T15:30:00Z",
    "started_at": "2025-01-21T11:00:00Z",
    "completed_at": null
  }
}
```

## Exercises

### Get Exercise Details

**Request:**
```bash
curl -X GET http://localhost:8080/api/exercises/exercise-uuid-789 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "exercise-uuid-789",
    "module_id": "mod-001-uuid",
    "exercise_number": 1,
    "title": "Build a Counter Component",
    "description": "Create a React component that implements a counter with increment and decrement functionality. Use the useState hook to manage state.",
    "language": "javascript",
    "starter_code": "import React from 'react';\n\nexport default function Counter() {\n  // Your code here\n  return (\n    <div>\n      <p>Count: 0</p>\n      <button>+</button>\n      <button>-</button>\n    </div>\n  );\n}",
    "difficulty": "easy",
    "points": 100,
    "hints": [
      "Use the useState hook to create a count state variable",
      "Create two button elements for increment and decrement",
      "Use onClick handlers to update the state"
    ],
    "created_at": "2025-01-21T11:00:00Z"
  }
}
```

**Error Response - Exercise Not Found (404):**
```json
{
  "error": "Not Found",
  "message": "exercise not found"
}
```

### Submit Exercise Solution

**Request:**
```bash
curl -X POST http://localhost:8080/api/exercises/exercise-uuid-789/submit \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "code": "import React, { useState } from '\''react'\'';\n\nexport default function Counter() {\n  const [count, setCount] = useState(0);\n  \n  return (\n    <div>\n      <p>Count: {count}</p>\n      <button onClick={() => setCount(count + 1)}>+</button>\n      <button onClick={() => setCount(count - 1)}>-</button>\n    </div>\n  );\n}",
    "language": "javascript"
  }'
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "submission-uuid-abc",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "module_id": "mod-001-uuid",
    "exercise_id": "exercise-uuid-789",
    "submitted_code": "import React, { useState } from 'react';\n\nexport default function Counter() {\n  const [count, setCount] = useState(0);\n  \n  return (\n    <div>\n      <p>Count: {count}</p>\n      <button onClick={() => setCount(count + 1)}>+</button>\n      <button onClick={() => setCount(count - 1)}>-</button>\n    </div>\n  );\n}",
    "language": "javascript",
    "passed": true,
    "score": 95,
    "attempts": 1,
    "hints_used": 0,
    "time_spent_minutes": 15,
    "submitted_at": "2025-01-21T15:30:00Z"
  }
}
```

**Error Response - Missing Code (400):**
```json
{
  "error": "Bad Request",
  "message": "Code is required"
}
```

**Error Response - Missing Language (400):**
```json
{
  "error": "Bad Request",
  "message": "Language is required"
}
```

### Request AI Code Review

**Request:**
```bash
curl -X POST http://localhost:8080/api/submissions/submission-uuid-abc/review \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "review-uuid-def",
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "module_id": "mod-001-uuid",
    "submission_id": "submission-uuid-abc",
    "overall_score": 88,
    "code_sense_score": 90,
    "efficiency_score": 85,
    "edge_cases_score": 92,
    "taste_score": 86,
    "feedback": {
      "strengths": [
        "Clean and readable component structure",
        "Proper use of React hooks (useState)",
        "Correct event handling with arrow functions",
        "Good separation of concerns"
      ],
      "improvements": [
        "Consider adding prop validation with PropTypes or TypeScript",
        "Extract button logic into a separate reusable Button component",
        "Add accessibility attributes (aria-label) to buttons",
        "Consider adding min/max boundaries for the counter"
      ],
      "suggestions": "Excellent work on this exercise! You've demonstrated a solid understanding of React state management. As next steps, consider learning about useReducer for more complex state logic, and explore React.memo for performance optimization. Keep up the great work!"
    },
    "reviewed_at": "2025-01-21T15:31:00Z"
  }
}
```

**Error Response - Submission Not Found (400):**
```json
{
  "error": "Internal Server Error",
  "message": "submission not found"
}
```

## Social Features

### Follow User

**Request:**
```bash
curl -X POST http://localhost:8080/api/users/user-uuid-999/follow \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (201 Created):**
```json
{
  "message": "Successfully followed user"
}
```

**Error Response - Missing User ID (400):**
```json
"User ID is required"
```

### Unfollow User

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/users/user-uuid-999/follow \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "message": "Successfully unfollowed user"
}
```

### Get Activity Feed

**Request:**
```bash
curl -X GET "http://localhost:8080/api/feed?limit=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "activities": [
    {
      "id": "activity-uuid-001",
      "user_id": "user-uuid-999",
      "activity_type": "module_completed",
      "reference_type": "module",
      "reference_id": "mod-001-uuid",
      "metadata": {
        "module_title": "Introduction to React Components",
        "course_title": "Modern Web Development with React",
        "score": 95
      },
      "visibility": "public",
      "created_at": "2025-01-21T14:30:00Z"
    },
    {
      "id": "activity-uuid-002",
      "user_id": "user-uuid-888",
      "activity_type": "achievement_unlocked",
      "reference_type": "achievement",
      "reference_id": "achievement-uuid-123",
      "metadata": {
        "achievement_name": "Fast Learner",
        "achievement_description": "Complete 5 modules in one day"
      },
      "visibility": "public",
      "created_at": "2025-01-21T13:15:00Z"
    }
  ],
  "count": 2
}
```

### Get Recommendations

**Request:**
```bash
curl -X GET http://localhost:8080/api/recommendations \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "recommendations": [
    {
      "id": "rec-uuid-001",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "course_id": "course-uuid-typescript",
      "recommendation_type": "skill_adjacency",
      "match_score": 87,
      "reason": "Based on your JavaScript skills, you'll love TypeScript",
      "created_at": "2025-01-21T12:00:00Z"
    },
    {
      "id": "rec-uuid-002",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "course_id": "course-uuid-nodejs",
      "recommendation_type": "collaborative_filtering",
      "match_score": 92,
      "reason": "Users who completed React also enjoyed Node.js backend development",
      "created_at": "2025-01-21T12:00:00Z"
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

### Get Trending Courses

**Request (No authentication required):**
```bash
curl -X GET http://localhost:8080/api/trending
```

**Success Response (200 OK):**
```json
{
  "trending": [
    {
      "id": "trending-uuid-001",
      "course_id": "course-uuid-react-adv",
      "velocity": 2.5,
      "signups_24h": 156,
      "signups_previous_24h": 62,
      "rank": 1,
      "meta_category": "programming",
      "calculated_at": "2025-01-21T16:00:00Z"
    },
    {
      "id": "trending-uuid-002",
      "course_id": "course-uuid-python-ml",
      "velocity": 2.1,
      "signups_24h": 142,
      "signups_previous_24h": 68,
      "rank": 2,
      "meta_category": "data-science",
      "calculated_at": "2025-01-21T16:00:00Z"
    }
  ],
  "count": 2
}
```

### Get User Profile (Living Resume)

**Request:**
```bash
curl -X GET http://localhost:8080/api/users/user-uuid-999/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "user": {
    "id": "user-uuid-999",
    "name": "Bob Developer",
    "avatar_url": "https://example.com/avatars/bob.jpg",
    "email": "bob@example.com"
  },
  "stats": {
    "total_courses": 5,
    "completed_courses": 2,
    "total_exercises": 45,
    "achievements_count": 8
  },
  "courses": [
    {
      "id": "course-uuid-001",
      "title": "Modern Web Development",
      "status": "completed",
      "completed_at": "2025-01-15T10:00:00Z"
    }
  ],
  "achievements": [
    {
      "id": "achievement-uuid-123",
      "name": "First Steps",
      "unlocked_at": "2025-01-10T12:00:00Z"
    }
  ],
  "recent_activity": [
    {
      "activity_type": "module_completed",
      "created_at": "2025-01-21T14:30:00Z"
    }
  ]
}
```

## Achievements

### Get User Achievements

**Request:**
```bash
curl -X GET http://localhost:8080/api/users/me/achievements \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response (200 OK):**
```json
{
  "achievements": [
    {
      "id": "achievement-uuid-001",
      "name": "First Steps",
      "description": "Complete your first exercise",
      "badge_icon": "https://example.com/badges/first-steps.png",
      "rarity": "common",
      "unlocked_at": "2025-01-21T12:00:00Z"
    },
    {
      "id": "achievement-uuid-002",
      "name": "Fast Learner",
      "description": "Complete 5 modules in one day",
      "badge_icon": "https://example.com/badges/fast-learner.png",
      "rarity": "rare",
      "unlocked_at": "2025-01-21T15:00:00Z"
    },
    {
      "id": "achievement-uuid-003",
      "name": "Perfect Score",
      "description": "Get 100% on an exercise on the first try",
      "badge_icon": "https://example.com/badges/perfect-score.png",
      "rarity": "epic",
      "unlocked_at": "2025-01-21T14:30:00Z"
    }
  ],
  "count": 3
}
```

## Authentication Header Examples

All protected endpoints require the JWT token in the Authorization header:

**Format:**
```
Authorization: Bearer <token>
```

**Example:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzZTQ1NjctZTg5Yi0xMmQzLWE0NTYtNDI2NjE0MTc0MDAwIiwiZXhwIjoxNzM3NTU0NDAwfQ.dGVzdF90b2tlbl9zaWduYXR1cmVfaGVyZQ
```

**Full curl example:**
```bash
curl -X GET http://localhost:8080/api/courses \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```
