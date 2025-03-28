# Go9jaJobs API Testing

## Running Tests

### Run All Tests
```bash
go test ./internal/...
```

### Run Tests with Coverage
```bash
go test -cover ./internal/...
```

### Run Tests with Verbose Output
```bash
go test -v ./internal/...
```

### Run Tests for a Specific Package
```bash
go test ./internal/db
go test ./internal/api
go test ./internal/fetcher
```

## GitHub Actions Workflow

A GitHub Actions workflow is set up to automatically run tests on:
- Push to the main branch
- Any pull request targeting the main branch

The workflow:
1. Sets up a PostgreSQL database for integration tests
2. Creates a test environment with mock API keys
3. Runs all unit and integration tests
4. Checks code coverage
5. Verifies code formatting (gofmt)
6. Runs go vet to catch common errors

## Test Dependencies

- `github.com/stretchr/testify/assert` - For test assertions
- `github.com/DATA-DOG/go-sqlmock` - For mocking database connections
- `net/http/httptest` - For testing HTTP handlers and middleware

## Coverage Goals

We aim to maintain at least 80% test coverage across the codebase, with critical paths (error handling, API responses) at 90%+ coverage.

## Adding New Tests

When adding new functionality:
1. Add tests in the same package as the code
2. Cover both success and error cases
3. Mock external dependencies
4. Test edge cases and boundary conditions

For API endpoints:
1. Test with various inputs including invalid ones
2. Verify response status codes and bodies
3. Test authorization and authentication

For database operations:
1. Use SQLite for unit tests
2. Test transactions and rollbacks
3. Test error handling for database operations 