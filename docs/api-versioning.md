# API Versioning Strategy

## Overview

The Learnify API uses URI-based versioning to ensure backward compatibility and smooth transitions between API versions. This document outlines our versioning strategy, deprecation policy, and migration guidelines.

## Current Version

**Current Version:** v1
**Base URL:** `http://localhost:8080/api` or `https://api.learnify.example.com/api`

## Versioning Approach

### URI-Based Versioning

We use URI-based versioning where the version number is included in the API path:

```
/api/v1/courses
/api/v1/users/me
/api/v2/courses (future)
```

**Current State:** Version 1 endpoints are currently unversioned (using `/api/` instead of `/api/v1/`) for backward compatibility. When v2 is released, v1 endpoints will be explicitly versioned.

### Why URI-Based?

- **Clear and Explicit:** Version is immediately visible in the URL
- **Easy to Route:** Simple to route different versions to different handlers
- **Browser-Friendly:** Can test different versions directly in the browser
- **Cache-Friendly:** Different versions can be cached independently

## Version Lifecycle

### 1. Active Version
- Receives new features and improvements
- Gets bug fixes and security patches
- Fully supported and documented

### 2. Maintenance Version
- Receives critical bug fixes only
- Receives security patches
- No new features added
- Minimum 6 months support after deprecation announcement

### 3. Deprecated Version
- Marked as deprecated in documentation
- Returns deprecation warning headers
- No bug fixes (except critical security issues)
- Minimum 6 months notice before retirement

### 4. Retired Version
- No longer available
- Returns 410 Gone status
- Redirect headers to latest version

## Breaking vs Non-Breaking Changes

### Non-Breaking Changes (Same Version)

These changes do not require a version bump:

- **Adding new endpoints**
  ```
  POST /api/courses (new endpoint)
  ```

- **Adding optional fields to requests**
  ```json
  {
    "name": "John",
    "avatar_url": "https://..." // new optional field
  }
  ```

- **Adding new fields to responses**
  ```json
  {
    "id": "123",
    "name": "Course",
    "created_at": "2025-01-21", // new field
    "tags": ["new", "field"] // new field
  }
  ```

- **Adding new query parameters (optional)**
- **Adding new HTTP methods to existing endpoints**
- **Relaxing validation rules**
- **Bug fixes that don't change behavior**

### Breaking Changes (New Version Required)

These changes require a new API version:

- **Removing endpoints**
- **Removing fields from responses**
- **Making optional fields required**
- **Changing field types**
  ```json
  // v1
  {"progress": 50} // integer

  // v2 (breaking)
  {"progress": "50%"} // string
  ```

- **Changing authentication methods**
- **Changing error response format**
- **Renaming fields**
  ```json
  // v1
  {"user_id": "123"}

  // v2 (breaking)
  {"userId": "123"} // renamed field
  ```

- **Changing URL structures**
- **Changing status code meanings**
- **Stricter validation rules**

## Deprecation Policy

### Announcement

When deprecating an API version:

1. **Advance Notice:** Minimum 6 months notice via:
   - API documentation
   - Developer blog/newsletter
   - Email to registered developers
   - Response headers

2. **Deprecation Header:**
   ```
   Deprecation: true
   Sunset: Sat, 31 Dec 2025 23:59:59 GMT
   Link: <https://api.learnify.example.com/api/v2/>; rel="successor-version"
   ```

3. **Documentation Updates:**
   - Mark deprecated endpoints with warning banner
   - Provide migration guide
   - Show equivalent endpoints in new version

### Timeline

```
Month 0: Announce deprecation
Month 1-5: Warning headers, documentation updates
Month 6: Version moved to maintenance mode
Month 12: Version retired (returns 410 Gone)
```

## Migration Guides

### From Unversioned to v1 (Future)

When we introduce explicit versioning:

**Old URLs (still supported):**
```
/api/courses
/api/users/me
```

**New URLs (recommended):**
```
/api/v1/courses
/api/v1/users/me
```

**Migration Steps:**
1. Update all API calls to include `/v1/` in the path
2. Test in staging environment
3. Deploy updates
4. No code changes required (URLs are backward compatible)

### From v1 to v2 (Future - Example)

**Example Breaking Change:** Consolidated user endpoints

**v1 Endpoints:**
```
GET /api/v1/users/me
PATCH /api/v1/users/me
```

**v2 Endpoints (hypothetical):**
```
GET /api/v2/user/profile
PUT /api/v2/user/profile
```

**Migration Steps:**
1. Review v2 changelog and breaking changes
2. Update client code to use new endpoints
3. Update request/response models if schemas changed
4. Test against v2 endpoints in staging
5. Deploy to production
6. Continue monitoring v1 for 6 months during transition

## Version Detection

### Request Header (Optional)

Clients can optionally specify version via header:

```http
GET /api/courses
Accept: application/vnd.learnify.v1+json
```

### Response Header

All responses include version information:

```http
X-API-Version: v1
```

## Backward Compatibility Guarantees

### Within Same Major Version

- All endpoints remain functional
- Response schemas only add fields, never remove
- Required request fields never change
- HTTP status codes maintain same meanings
- Error response format remains consistent

### Across Major Versions

- No guarantees - breaking changes allowed
- Full migration guide provided
- Minimum 6 months overlap period
- Automated migration tools where possible

## Client Implementation Best Practices

### 1. Always Specify Version

```javascript
const API_BASE = 'https://api.learnify.example.com/api/v1';

async function getCourses() {
  const response = await fetch(`${API_BASE}/courses`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Accept': 'application/json'
    }
  });
  return response.json();
}
```

### 2. Handle Version Headers

```javascript
async function makeApiCall(url) {
  const response = await fetch(url);

  // Check for deprecation warning
  if (response.headers.get('Deprecation')) {
    console.warn('API version deprecated:', response.headers.get('Sunset'));
    console.log('Migrate to:', response.headers.get('Link'));
  }

  return response.json();
}
```

### 3. Graceful Handling of New Fields

```javascript
// Don't validate exact response shape
// Allow new fields to be added
const user = await getUser();
const { id, name, email } = user; // Extract only what you need
// Ignore any new fields added in non-breaking updates
```

### 4. Version-Specific Error Handling

```javascript
async function handleApiError(response) {
  const version = response.headers.get('X-API-Version');

  if (response.status === 410) {
    throw new Error(`API version ${version} has been retired. Please upgrade.`);
  }

  // Handle other errors based on version
  const error = await response.json();
  throw new Error(error.message || 'API Error');
}
```

## Versioning FAQs

### When will v2 be released?

We plan to maintain v1 for at least 12-24 months. A v2 release is not currently scheduled.

### Can I use multiple versions simultaneously?

Yes! You can use different versions for different endpoints during migration:

```javascript
const v1Client = new ApiClient('v1');
const v2Client = new ApiClient('v2');

// Use v1 for some endpoints
await v1Client.getCourses();

// Use v2 for others
await v2Client.getUserProfile();
```

### What happens if I don't specify a version?

Currently, unversioned paths (`/api/courses`) default to v1. When v2 is released, we'll maintain this behavior for 6 months, then require explicit versioning.

### How do I know when a new version is available?

- Subscribe to our developer newsletter
- Watch our GitHub repository
- Check the API changelog regularly
- Monitor response headers for deprecation warnings

## Version History

| Version | Release Date | Status | EOL Date |
|---------|--------------|--------|----------|
| v1      | 2025-01-21   | Active | TBD      |

## Resources

- **API Documentation:** https://docs.learnify.example.com
- **Changelog:** https://docs.learnify.example.com/changelog
- **Migration Guides:** https://docs.learnify.example.com/migrations
- **Support:** support@learnify.example.com

## Contact

For questions about versioning or migration assistance:
- Email: api-support@learnify.example.com
- GitHub Issues: https://github.com/learnify/api/issues
- Developer Forum: https://forum.learnify.example.com
