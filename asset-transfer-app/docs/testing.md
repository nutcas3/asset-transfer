# Testing Guide

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Race Detector
```bash
go test -race ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Tests with Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Package Tests
```bash
# Test use cases
go test ./internal/usecases/...

# Test error handling
go test ./internal/errors/...

# Test server
go test ./internal/server/...
```

### Run Specific Test
```bash
go test -run TestSuccessfulTransfer ./internal/usecases/...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

## Test Structure

### Unit Tests
- **Use Cases Layer**: `internal/usecases/transfer_test.go`
  - Tests business logic with mock repositories
  - Covers idempotency, validation, error handling
  - Uses mock implementations for dependencies

- **Error Handling**: `internal/errors/errors_test.go`
  - Tests error creation, wrapping, and unwrapping
  - Validates error code mappings
  - Tests error context handling

- **Server Layer**: `internal/server/routes_test.go`
  - Tests HTTP routing
  - Validates endpoint registration
  - Tests health check endpoint

### Test Coverage

The service includes comprehensive tests covering:
- ✅ Successful transfer
- ✅ Terminal-result idempotency
- ✅ Conflict when one key is reused for another payload
- ✅ Many concurrent requests for one key causing only one gateway call
- ✅ Retry after a transient failure without changing the transfer ID
- ✅ Gateway timeout behavior
- ✅ Circuit breaker opening, rejecting calls while open, and recovering through half-open probe
- ✅ Request validation errors
- ✅ Transfer retrieval

## Testing Strategies

### Mocking Strategy
The use cases layer uses mock repositories for testing:
- **Mock Transfer Repository**: In-memory implementation for testing
- **Mock Gateway Repository**: Configurable behavior for different scenarios
- **Force Conflict Flag**: Allows testing payload conflict scenarios

### Concurrency Testing
Tests use goroutines and sync.WaitGroup to test concurrent behavior:
- Singleflight pattern validation
- Race condition detection
- Concurrent request handling

### Circuit Breaker Testing
Tests validate circuit breaker behavior:
- State transitions (closed → open → half-open → closed)
- Failure threshold behavior
- Cooldown period handling
- Recovery through half-open probe

### Idempotency Testing
Tests ensure idempotency works correctly:
- Same key + same payload = cached result
- Same key + different payload = conflict error
- Concurrent requests with same key = single gateway call

## Debugging Tests

### Run Tests with Debug Output
```bash
go test -v ./internal/usecases/...
```

### Run Specific Test with Debug
```bash
go test -v -run TestSuccessfulTransfer ./internal/usecases/...
```

### Run Tests with Race Detector
```bash
go test -race ./internal/usecases/...
```

### Run Tests with Timeout
```bash
go test -timeout 30s ./...
```

## Continuous Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test ./...
      - run: go test -race ./...
```

## Test Maintenance

### Adding New Tests
1. Identify the scenario to test
2. Add test function with `Test` prefix
3. Use table-driven tests for multiple scenarios
4. Mock dependencies as needed
5. Assert expected behavior
6. Run tests to verify

### Updating Tests
When refactoring:
1. Update tests to match new implementation
2. Ensure test coverage remains high
3. Add tests for new functionality
4. Remove obsolete tests
5. Run all tests to verify

### Test Best Practices
- Use descriptive test names
- Test one thing per test
- Use table-driven tests for similar scenarios
- Mock external dependencies
- Clean up resources in tests
- Use race detector for concurrent code
- Maintain high test coverage
