# API Schema Documentation

Complete documentation of all data models, field descriptions, validation rules, and data types.

## Table of Contents

- [User & Identity](#user--identity)
- [Learning Domain](#learning-domain)
- [Social Domain](#social-domain)
- [Common Types](#common-types)

## User & Identity

### User

Represents a user account in the system.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `id` | UUID | Yes | Unique user identifier | Auto-generated |
| `email` | String | Yes | User's email address | Valid email format, unique |
| `password_hash` | String | Internal | Hashed password | Min 8 chars (before hashing) |
| `name` | String | Yes | User's display name | 1-100 characters |
| `avatar_url` | String | No | Profile picture URL | Valid URL or empty |
| `privacy_settings` | Object | No | Privacy preferences | See PrivacySettings schema |
| `created_at` | Timestamp | Yes | Account creation time | ISO 8601 format |
| `updated_at` | Timestamp | Yes | Last update time | ISO 8601 format |
| `last_login` | Timestamp | Yes | Last login time | ISO 8601 format |

**Example:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "name": "John Doe",
  "avatar_url": "https://example.com/avatars/user123.jpg",
  "privacy_settings": { ... },
  "created_at": "2025-01-21T10:30:00Z",
  "updated_at": "2025-01-21T14:45:00Z",
  "last_login": "2025-01-21T14:45:00Z"
}
```

### PrivacySettings

User privacy preferences.

| Field | Type | Required | Description | Allowed Values |
|-------|------|----------|-------------|----------------|
| `profile_visibility` | String | Yes | Who can view profile | `public`, `friends`, `private` |
| `activity_visibility` | String | Yes | Who can see activity | `public`, `friends`, `private` |
| `progress_visibility` | String | Yes | Who can see progress | `public`, `friends`, `private` |
| `allow_followers` | Boolean | Yes | Allow others to follow | `true`, `false` |
| `show_in_leaderboards` | Boolean | Yes | Appear in leaderboards | `true`, `false` |
| `show_completed_courses` | Boolean | Yes | Display completed courses | `true`, `false` |

**Default Values:**
```json
{
  "profile_visibility": "public",
  "activity_visibility": "friends",
  "progress_visibility": "public",
  "allow_followers": true,
  "show_in_leaderboards": true,
  "show_completed_courses": true
}
```

### RegisterRequest

Payload for user registration.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `email` | String | Yes | Email address | Valid email format |
| `password` | String | Yes | User password | Min 8 characters |
| `name` | String | Yes | Display name | 1-100 characters |

**Example:**
```json
{
  "email": "newuser@example.com",
  "password": "SecurePass123!",
  "name": "Jane Smith"
}
```

### LoginRequest

Payload for user login.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `email` | String | Yes | Email address | Valid email format |
| `password` | String | Yes | User password | Non-empty |

**Example:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

### AuthResponse

Response after successful authentication.

| Field | Type | Description |
|-------|------|-------------|
| `token` | String | JWT authentication token |
| `user` | User | User object (without password) |

**Token Format:** JWT with 24-hour expiration

**Example:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { ... }
}
```

### UpdateProfileRequest

Payload for updating user profile.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `name` | String | No | New display name | 1-100 characters |
| `avatar_url` | String | No | New avatar URL | Valid URL |

**Note:** All fields are optional. Only provided fields will be updated.

**Example:**
```json
{
  "name": "John Updated",
  "avatar_url": "https://example.com/avatars/new.jpg"
}
```

### OnboardingRequest

Payload for completing user onboarding.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `meta_category` | String | Yes | Learning category | Non-empty string |
| `domain` | String | Yes | Specific domain | Non-empty string |
| `skill_level` | String | Yes | User's skill level | `beginner`, `intermediate`, `advanced` |
| `variables` | Object | No | Custom variables | Key-value pairs |

**Example:**
```json
{
  "meta_category": "programming",
  "domain": "web-development",
  "skill_level": "beginner",
  "variables": {
    "preferred_language": "javascript",
    "learning_goal": "build-web-apps",
    "time_commitment": "10-hours-week"
  }
}
```

## Learning Domain

### Course

Represents a personalized course generated for a user.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Unique course identifier |
| `user_id` | UUID | Yes | Owner user ID |
| `archetype_id` | UUID | Yes | Associated archetype |
| `title` | String | Yes | Course title |
| `description` | String | Yes | Course description |
| `meta_category` | String | Yes | High-level category |
| `injected_variables` | JSON | No | Personalization variables |
| `status` | String | Yes | Course status |
| `created_at` | Timestamp | Yes | Creation time |
| `updated_at` | Timestamp | Yes | Last update time |

**Status Values:** `active`, `completed`, `archived`

**Example:**
```json
{
  "id": "456e7890-e89b-12d3-a456-426614174001",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "archetype_id": "789abc12-e89b-12d3-a456-426614174002",
  "title": "Modern Web Development with React",
  "description": "Learn to build modern web applications",
  "meta_category": "programming",
  "status": "active",
  "created_at": "2025-01-21T11:00:00Z",
  "updated_at": "2025-01-21T11:00:00Z"
}
```

### Module

Represents a module within a course.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Unique module identifier |
| `course_id` | UUID | Yes | Parent course ID |
| `blueprint_module_id` | UUID | Yes | Blueprint template ID |
| `module_number` | Integer | Yes | Sequential number (1-based) |
| `title` | String | Yes | Module title |
| `description` | String | Yes | Module description |
| `content` | JSON | No | Module content data |
| `status` | String | Yes | Module status |
| `unlocked_at` | Timestamp | No | When module was unlocked |
| `created_at` | Timestamp | Yes | Creation time |

**Status Values:** `locked`, `unlocked`, `completed`

**Module Number:** Sequential, starts at 1

**Example:**
```json
{
  "id": "mod-001-uuid",
  "course_id": "456e7890-e89b-12d3-a456-426614174001",
  "module_number": 1,
  "title": "Introduction to React Components",
  "description": "Learn React fundamentals",
  "status": "unlocked",
  "unlocked_at": "2025-01-21T11:00:00Z",
  "created_at": "2025-01-21T11:00:00Z"
}
```

### Exercise

Represents a coding exercise within a module.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `id` | UUID | Yes | Unique exercise identifier | Auto-generated |
| `module_id` | UUID | Yes | Parent module ID | Valid UUID |
| `exercise_number` | Integer | Yes | Sequential number | Positive integer |
| `title` | String | Yes | Exercise title | Non-empty |
| `description` | String | Yes | Exercise description | Non-empty |
| `language` | String | Yes | Programming language | Lowercase (e.g., `javascript`) |
| `starter_code` | String | No | Initial code template | Valid code |
| `solution_code` | String | Internal | Reference solution | Not exposed to users |
| `test_cases` | JSON | No | Test case definitions | Array of test objects |
| `difficulty` | String | Yes | Difficulty level | `easy`, `medium`, `hard` |
| `points` | Integer | Yes | Points awarded | 0-1000 |
| `hints` | JSON | No | Available hints | Array of strings |
| `created_at` | Timestamp | Yes | Creation time | ISO 8601 |

**Example:**
```json
{
  "id": "exercise-uuid-789",
  "module_id": "mod-001-uuid",
  "exercise_number": 1,
  "title": "Build a Counter Component",
  "description": "Create a React counter with increment/decrement",
  "language": "javascript",
  "starter_code": "import React from 'react';\n\nexport default function Counter() {\n  // Your code here\n}",
  "difficulty": "easy",
  "points": 100,
  "hints": [
    "Use the useState hook",
    "Create two button elements"
  ],
  "created_at": "2025-01-21T11:00:00Z"
}
```

### SubmitExerciseRequest

Payload for submitting an exercise solution.

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `code` | String | Yes | Submitted code | Non-empty |
| `language` | String | Yes | Programming language | Non-empty |

**Example:**
```json
{
  "code": "import React, { useState } from 'react';\n\nexport default function Counter() {\n  const [count, setCount] = useState(0);\n  return (\n    <div>\n      <p>Count: {count}</p>\n      <button onClick={() => setCount(count + 1)}>+</button>\n      <button onClick={() => setCount(count - 1)}>-</button>\n    </div>\n  );\n}",
  "language": "javascript"
}
```

### ModuleCompletion

Represents a submitted exercise solution.

| Field | Type | Required | Description | Range |
|-------|------|----------|-------------|-------|
| `id` | UUID | Yes | Unique submission ID | - |
| `user_id` | UUID | Yes | Submitter user ID | - |
| `module_id` | UUID | Yes | Parent module ID | - |
| `exercise_id` | UUID | Yes | Parent exercise ID | - |
| `submitted_code` | String | Yes | User's code | - |
| `language` | String | Yes | Programming language | - |
| `test_results` | JSON | No | Test execution results | - |
| `passed` | Boolean | Yes | Whether tests passed | `true`, `false` |
| `score` | Integer | Yes | Score earned | 0-100 |
| `attempts` | Integer | Yes | Number of attempts | >= 1 |
| `hints_used` | Integer | Yes | Hints used count | >= 0 |
| `time_spent_minutes` | Integer | Yes | Time spent | >= 0 |
| `submitted_at` | Timestamp | Yes | Submission time | ISO 8601 |

**Example:**
```json
{
  "id": "submission-uuid-abc",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "module_id": "mod-001-uuid",
  "exercise_id": "exercise-uuid-789",
  "submitted_code": "...",
  "language": "javascript",
  "passed": true,
  "score": 95,
  "attempts": 1,
  "hints_used": 0,
  "time_spent_minutes": 15,
  "submitted_at": "2025-01-21T15:30:00Z"
}
```

### ArchitectureReview

AI-powered code review with detailed feedback.

| Field | Type | Required | Description | Range |
|-------|------|----------|-------------|-------|
| `id` | UUID | Yes | Unique review ID | - |
| `user_id` | UUID | Yes | User ID | - |
| `module_id` | UUID | Yes | Module ID | - |
| `submission_id` | UUID | Yes | Reviewed submission | - |
| `overall_score` | Integer | Yes | Overall score | 0-100 |
| `code_sense_score` | Integer | Yes | Code clarity score | 0-100 |
| `efficiency_score` | Integer | Yes | Performance score | 0-100 |
| `edge_cases_score` | Integer | Yes | Edge case handling | 0-100 |
| `taste_score` | Integer | Yes | Code style score | 0-100 |
| `feedback` | JSON | Yes | Detailed feedback | Object with arrays |
| `reviewed_at` | Timestamp | Yes | Review timestamp | ISO 8601 |

**Feedback Structure:**
```json
{
  "strengths": ["Array of positive points"],
  "improvements": ["Array of areas to improve"],
  "suggestions": "Overall suggestions text"
}
```

**Example:**
```json
{
  "id": "review-uuid-def",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "submission_id": "submission-uuid-abc",
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
      "Add prop validation",
      "Extract button logic"
    ],
    "suggestions": "Great job! Consider learning useReducer next."
  },
  "reviewed_at": "2025-01-21T15:31:00Z"
}
```

### UserProgress

Tracks user's progress through a course.

| Field | Type | Required | Description | Range |
|-------|------|----------|-------------|-------|
| `id` | UUID | Yes | Unique progress ID | - |
| `user_id` | UUID | Yes | User ID | - |
| `course_id` | UUID | Yes | Course ID | - |
| `current_module_id` | UUID | Yes | Current module | - |
| `progress_percentage` | Integer | Yes | Completion % | 0-100 |
| `time_spent_minutes` | Integer | Yes | Total time spent | >= 0 |
| `last_activity` | Timestamp | Yes | Last activity time | ISO 8601 |
| `started_at` | Timestamp | Yes | Start time | ISO 8601 |
| `completed_at` | Timestamp | No | Completion time | ISO 8601 or null |

**Example:**
```json
{
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
```

## Social Domain

### UserRelationship

Represents a follow relationship between users.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Unique relationship ID |
| `follower_id` | UUID | Yes | User who follows |
| `following_id` | UUID | Yes | User being followed |
| `created_at` | Timestamp | Yes | When relationship created |

**Example:**
```json
{
  "id": "relationship-uuid",
  "follower_id": "user-uuid-123",
  "following_id": "user-uuid-456",
  "created_at": "2025-01-21T10:00:00Z"
}
```

### ActivityFeed

Represents an activity in the user's feed.

| Field | Type | Required | Description | Allowed Values |
|-------|------|----------|-------------|----------------|
| `id` | UUID | Yes | Unique activity ID | - |
| `user_id` | UUID | Yes | User who performed activity | - |
| `activity_type` | String | Yes | Type of activity | See below |
| `reference_type` | String | Yes | Type of referenced entity | `course`, `module`, `exercise`, `achievement` |
| `reference_id` | UUID | Yes | Referenced entity ID | - |
| `metadata` | JSON | No | Additional context | Object |
| `visibility` | String | Yes | Who can see this | `public`, `friends`, `private` |
| `created_at` | Timestamp | Yes | Activity timestamp | ISO 8601 |

**Activity Types:**
- `course_started`
- `module_completed`
- `exercise_submitted`
- `achievement_unlocked`
- `user_followed`

**Example:**
```json
{
  "id": "activity-uuid-001",
  "user_id": "user-uuid-999",
  "activity_type": "module_completed",
  "reference_type": "module",
  "reference_id": "mod-001-uuid",
  "metadata": {
    "module_title": "Introduction to React",
    "course_title": "Modern Web Development",
    "score": 95
  },
  "visibility": "public",
  "created_at": "2025-01-21T14:30:00Z"
}
```

### Achievement

Represents an achievement definition.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Unique achievement ID |
| `name` | String | Yes | Achievement name |
| `description` | String | Yes | Achievement description |
| `badge_icon` | String | Yes | Badge icon URL |
| `criteria` | JSON | No | Unlock criteria |
| `rarity` | String | Yes | Rarity level |
| `created_at` | Timestamp | Yes | Creation time |

**Rarity Levels:** `common`, `rare`, `epic`, `legendary`

**Example:**
```json
{
  "id": "achievement-uuid-001",
  "name": "First Steps",
  "description": "Complete your first exercise",
  "badge_icon": "https://example.com/badges/first-steps.png",
  "rarity": "common",
  "created_at": "2025-01-15T10:00:00Z"
}
```

### UserAchievement

Links a user to an unlocked achievement.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Unique record ID |
| `user_id` | UUID | Yes | User ID |
| `achievement_id` | UUID | Yes | Achievement ID |
| `unlocked_at` | Timestamp | Yes | Unlock timestamp |

### Recommendation

Represents a course recommendation for a user.

| Field | Type | Required | Description | Range |
|-------|------|----------|-------------|-------|
| `id` | UUID | Yes | Unique recommendation ID | - |
| `user_id` | UUID | Yes | User ID | - |
| `course_id` | UUID | Yes | Recommended course | - |
| `recommendation_type` | String | Yes | Recommendation algorithm | See below |
| `match_score` | Integer | Yes | Match score | 0-100 |
| `reason` | String | Yes | Why recommended | - |
| `metadata` | JSON | No | Additional data | - |
| `created_at` | Timestamp | Yes | Recommendation time | ISO 8601 |
| `expires_at` | Timestamp | No | Expiration time | ISO 8601 or null |

**Recommendation Types:**
- `collaborative_filtering` - Based on similar users
- `skill_adjacency` - Based on skill progression
- `social_signal` - Based on friends' activity
- `trending` - Based on current trends

**Example:**
```json
{
  "id": "rec-uuid-001",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "course_id": "course-uuid-typescript",
  "recommendation_type": "skill_adjacency",
  "match_score": 87,
  "reason": "Based on your JavaScript skills, you'll love TypeScript",
  "created_at": "2025-01-21T12:00:00Z",
  "expires_at": null
}
```

### TrendingCourse

Represents trending course data.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Unique trending record ID |
| `course_id` | UUID | Yes | Course ID |
| `velocity` | Float | Yes | Growth rate |
| `signups_24h` | Integer | Yes | Signups in last 24 hours |
| `signups_previous_24h` | Integer | Yes | Signups in previous 24 hours |
| `rank` | Integer | Yes | Trending rank (1-based) |
| `meta_category` | String | Yes | Course category |
| `calculated_at` | Timestamp | Yes | When calculated |

**Velocity Calculation:** `signups_24h / signups_previous_24h`

**Example:**
```json
{
  "id": "trending-uuid-001",
  "course_id": "course-uuid-react",
  "velocity": 2.5,
  "signups_24h": 156,
  "signups_previous_24h": 62,
  "rank": 1,
  "meta_category": "programming",
  "calculated_at": "2025-01-21T16:00:00Z"
}
```

## Common Types

### UUID Format

All UUIDs follow the standard format:
```
123e4567-e89b-12d3-a456-426614174000
```

### Timestamp Format

All timestamps use ISO 8601 format with UTC timezone:
```
2025-01-21T15:30:00Z
```

### Response Wrappers

**Success Response (Learning Endpoints):**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error Response (Identity Endpoints):**
```json
{
  "error": "error message"
}
```

**Error Response (Learning Endpoints):**
```json
{
  "error": "HTTP Status Text",
  "message": "Detailed error message"
}
```

### Enum Values Reference

**User Privacy:**
- `public` - Visible to everyone
- `friends` - Visible to followers/following
- `private` - Visible only to user

**Course Status:**
- `active` - Currently in progress
- `completed` - Finished successfully
- `archived` - No longer active

**Module Status:**
- `locked` - Not yet accessible
- `unlocked` - Accessible to user
- `completed` - User finished

**Difficulty Levels:**
- `easy` - Beginner-friendly
- `medium` - Intermediate level
- `hard` - Advanced challenge

**Skill Levels:**
- `beginner` - New to the topic
- `intermediate` - Some experience
- `advanced` - Expert level

**Achievement Rarity:**
- `common` - Easy to unlock
- `rare` - Moderate difficulty
- `epic` - Hard to achieve
- `legendary` - Extremely rare
