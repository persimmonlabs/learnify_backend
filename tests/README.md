# Backend Test Suite Documentation

## Overview

This test suite provides comprehensive coverage for the Go backend application across three domains: identity (authentication), learning (courses), and social (feed/relationships).

## Test Structure

```
backend/
├── internal/
│   ├── identity/
│   │   ├── service_test.go       # Identity service unit tests
│   │   ├── repository_test.go    # Identity repository unit tests
│   │   └── handler_test.go       # Identity HTTP handler tests
│   ├── learning/
│   │   ├── service_test.go       # Learning service unit tests
│   │   ├── repository_test.go    # Learning repository unit tests (TODO)
│   │   └── handler_test.go       # Learning HTTP handler tests (TODO)
│   ├── social/
│   │   ├── service_test.go       # Social service unit tests
│   │   ├── repository_test.go    # Social repository unit tests (TODO)
│   │   └── handler_test.go       # Social HTTP handler tests (TODO)
│   └── platform/
│       └── middleware/
│           ├── auth_test.go      # JWT authentication middleware tests
│           ├── cors_test.go      # CORS middleware tests (TODO)
│           └── logging_test.go   # Logging middleware tests (TODO)
└── tests/
    ├── integration/
    │   ├── api_test.go           # End-to-end API flow tests
    │   ├── database_test.go      # Database integration tests (TODO)
    │   └── auth_flow_test.go     # Complete auth workflow tests (TODO)
    ├── testutil/
    │   ├── fixtures.go           # Test data fixtures
    │   ├── mocks.go              # Mock implementations
    │   └── testdb.go             # Test database setup/teardown
    └── README.md                 # This file
```

## Running Tests

### Run All Tests
```bash
cd backend
go test -v ./...
```

### Run Tests with Coverage
```bash
go test -v -cover ./...
```

### Run Tests with Coverage Report
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Package Tests
```bash
# Identity tests
go test -v ./internal/identity/...

# Learning tests
go test -v ./internal/learning/...

# Social tests
go test -v ./internal/social/...

# Middleware tests
go test -v ./internal/platform/middleware/...

# Integration tests
go test -v ./tests/integration/...
```

### Run Tests with Race Detector
```bash
go test -v -race ./...
```

## Test Types

### 1. Unit Tests

**Purpose**: Test individual functions and methods in isolation

**Location**: Alongside source code in `internal/` directories

**Characteristics**:
- Fast execution
- Isolated from external dependencies
- Use mocks for database and external services
- Focus on business logic

**Example**:
```go
func TestServiceRegister(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo, "test-secret")

    // Test implementation...
}
```

### 2. Integration Tests

**Purpose**: Test component interactions and end-to-end workflows

**Location**: `tests/integration/`

**Characteristics**:
- Test multiple components together
- May use test database
- Verify complete workflows
- Slower than unit tests

**Example**:
```go
func TestAuthFlowIntegration(t *testing.T) {
    db := testutil.SetupTestDB(t)
    // Test registration -> login -> profile access
}
```

### 3. HTTP Handler Tests

**Purpose**: Test HTTP request/response handling

**Characteristics**:
- Use `httptest` for request/response recording
- Test status codes, headers, body content
- Test both success and error cases

**Example**:
```go
func TestHandlerRegister(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/register", body)
    w := httptest.NewRecorder()
    handler.Register(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

## Test Utilities

### Fixtures (`tests/testutil/fixtures.go`)

Helper functions to create test data:
- `CreateTestUser()` - Create user with default values
- `CreateTestArchetype()` - Create user archetype
- `CreateTestGeneratedCourse()` - Create course instance
- `CreateTestExercise()` - Create exercise/challenge
- And more...

### Mocks (`tests/testutil/mocks.go`)

Mock implementations for external dependencies:
- `MockAIClient` - Mock AI service responses
- `MockCourseGenerator` - Mock course generation
- Repository mocks (using testify/mock)

### Test Database (`tests/testutil/testdb.go`)

Database setup and teardown:
- `SetupTestDB()` - Create test database connection
- `CleanupTestDB()` - Remove test data
- `CreateTestSchema()` - Create tables for testing

## Testing Strategy

### Table-Driven Tests

Use table-driven tests for multiple test cases:

```go
tests := []struct {
    name        string
    input       string
    expectError bool
}{
    {"valid input", "test@example.com", false},
    {"invalid input", "invalid", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
}
```

### Test Coverage Requirements

- **Minimum**: 60% overall code coverage
- **Target**: 80% coverage for business logic
- **Critical paths**: 100% coverage for authentication and authorization

### Testing Best Practices

1. **Arrange-Act-Assert**: Structure tests with clear setup, execution, and verification
2. **Test Naming**: Use descriptive names that explain what is being tested
3. **Independent Tests**: Each test should run independently without side effects
4. **Mock External Dependencies**: Use mocks for AI services, external APIs, etc.
5. **Test Both Success and Failure**: Cover happy paths and error cases
6. **Use Test Helpers**: Leverage fixtures and utilities for cleaner tests

## Database Testing

### Requirements

Tests that require a database will skip if no database is available:

```go
db := testutil.SetupTestDB(t)
if db == nil {
    return // Skip test
}
```

### Local Testing

For local testing with PostgreSQL:

```bash
# Start PostgreSQL with Docker
docker run --name backend-test-db \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=backend_test \
  -p 5432:5432 \
  -d postgres:15

# Run tests
go test -v ./...
```

### CI/CD Testing

For CI/CD pipelines, use Docker Compose:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: backend_test
    ports:
      - "5432:5432"
```

## Mocking Strategy

### Repository Mocking

Use `testify/mock` for repository interfaces:

```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) CreateUser(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}
```

### AI Service Mocking

Use custom mocks for AI client:

```go
mockAI := &testutil.MockAIClient{
    ShouldFail: false,
    CurriculumResult: &ai.Curriculum{...},
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: backend_test
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run tests
        run: |
          cd backend
          go test -v -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
```

## Test Data Management

### Test Isolation

Each test should:
- Create its own test data
- Clean up after itself
- Not depend on other tests

### Fixtures vs Factories

- **Fixtures**: Pre-defined test data (use for common scenarios)
- **Factories**: Generate test data dynamically (use for specific cases)

## Debugging Tests

### Verbose Output
```bash
go test -v ./internal/identity/...
```

### Run Single Test
```bash
go test -v -run TestServiceRegister ./internal/identity/
```

### Enable Debug Logging
```bash
go test -v -args -debug ./...
```

## Common Test Patterns

### Testing Error Cases
```go
t.Run("error case", func(t *testing.T) {
    err := service.DoSomething("invalid")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected error message")
})
```

### Testing with Context
```go
ctx := context.Background()
ctx = context.WithValue(ctx, userIDKey, "user-123")
result, err := service.MethodWithContext(ctx)
```

### Testing HTTP Middleware
```go
middleware := Auth("secret")
handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Handler logic
}))
handler.ServeHTTP(w, req)
```

## Coverage Goals

### Current Coverage

Run to check current coverage:
```bash
go test -cover ./...
```

### Coverage by Package

| Package | Target | Current |
|---------|--------|---------|
| identity | 80% | TBD |
| learning | 70% | TBD |
| social | 70% | TBD |
| middleware | 85% | TBD |
| **Overall** | **60%+** | **TBD** |

## Troubleshooting

### Tests Skipping

If tests are skipping, ensure:
- PostgreSQL is running (for database tests)
- Connection string is correct in `testutil/testdb.go`
- Required dependencies are installed

### Flaky Tests

If tests fail intermittently:
- Check for race conditions (run with `-race`)
- Ensure proper cleanup between tests
- Use deterministic test data (avoid random values)

### Slow Tests

If tests are slow:
- Use mocks instead of real database where possible
- Parallelize independent tests with `t.Parallel()`
- Reduce test data size

## Future Improvements

- [ ] Add learning repository tests
- [ ] Add learning handler tests
- [ ] Add social repository tests
- [ ] Add social handler tests
- [ ] Add CORS middleware tests
- [ ] Add logging middleware tests
- [ ] Add benchmark tests
- [ ] Add performance tests
- [ ] Implement testcontainers for database
- [ ] Add E2E tests with full server

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
