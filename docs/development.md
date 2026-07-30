# Development Guide

## Setup

### Prerequisites
- Go 1.21 or higher
- Git
- Make (optional)

### Clone Repository
```bash
git clone <repository-url>
cd asset-transfer-app
```

### Install Dependencies
```bash
go mod download
```

### Verify Installation
```bash
go version
go test ./...
```

## Project Structure

```
asset-transfer-app/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── domain/              # Core business logic
│   │   ├── entities/        # Business entities
│   │   └── repositories/    # Repository interfaces
│   ├── usecases/            # Business logic orchestration
│   ├── infrastructure/      # External implementations
│   │   ├── persistence/     # Storage implementations
│   │   └── gateway/         # Gateway implementations
│   ├── delivery/            # Delivery mechanisms
│   │   └── http/            # HTTP handlers
│   ├── errors/              # Structured error handling
│   └── server/              # Server setup
├── docs/                    # Documentation
├── go.mod                   # Go module file
├── go.sum                   # Go module checksums
└── README.md                # Project README
```

## Development Workflow

### Make Changes
1. Create a new branch for your feature
2. Make your changes following the Clean Architecture principles
3. Write tests for your changes
4. Run tests to ensure everything works
5. Commit your changes

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -cover ./...
```

### Build and Run
```bash
# Build the application
go build -o asset-transfer-app cmd/api/main.go

# Run the application
./asset-transfer-app

# Or run directly
go run cmd/api/main.go
```

## Coding Standards

### Clean Architecture Principles
- **Domain Layer**: Pure business logic, no external dependencies
- **Use Cases Layer**: Business logic orchestration, depends on interfaces
- **Infrastructure Layer**: External implementations, implements interfaces
- **Delivery Layer**: HTTP handlers, depends on use cases

### Error Handling
- Use structured errors from `internal/errors`
- Wrap errors with appropriate context
- Include layer attribution for debugging
- Use machine-readable error codes

### Code Style
- Follow Go standard formatting (`go fmt`)
- Use meaningful variable names
- Add comments for complex logic
- Keep functions focused and small
- Use interfaces for dependencies

### Testing
- Write unit tests for business logic
- Use mocks for external dependencies
- Test error conditions
- Test concurrent behavior with race detector
- Maintain high test coverage

## Adding New Features

### Adding a New Use Case
1. Define the use case in `internal/usecases/`
2. Add repository interfaces to `internal/domain/repositories/`
3. Implement the use case logic
4. Add tests for the use case
5. Update delivery layer to expose the use case

### Adding a New Repository Implementation
1. Implement the repository interface in `internal/infrastructure/persistence/`
2. Add tests for the implementation
3. Update dependency injection in `internal/server/routes.go`

### Adding a New HTTP Endpoint
1. Add handler method in `internal/delivery/http/handler.go`
2. Register the route in `internal/server/routes.go`
3. Add error handling as needed
4. Update API documentation

## Debugging

### Enable Debug Logging
```bash
# Set log level (if implemented)
LOG_LEVEL=debug go run cmd/api/main.go
```

### Use Delve Debugger
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug cmd/api/main.go
```

### Add Debug Points
```go
// Add temporary debug logging
fmt.Printf("Debug: transfer = %+v\n", transfer)
```

## Common Tasks

### Adding a New Error Code
1. Add error code to `internal/errors/errors.go`
2. Add predefined error variable
3. Update error handling in use cases
4. Update error handling in delivery layer
5. Update documentation

### Updating Dependencies
```bash
# Update all dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Running Linters
```bash
# Run go vet
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

## Git Workflow

### Branch Naming
- `feature/feature-name` for new features
- `bugfix/bug-description` for bug fixes
- `hotfix/critical-fix` for urgent fixes

### Commit Messages
- Use clear, descriptive commit messages
- Start with a verb (e.g., "Add", "Fix", "Update")
- Keep messages concise but informative
- Reference related issues if applicable

### Pull Request Process
1. Create a pull request with clear description
2. Ensure all tests pass
3. Request code review
4. Address review comments
5. Merge after approval

## Performance Considerations

### Memory Usage
- Monitor in-memory storage size
- Implement storage limits if needed
- Consider adding TTL for old records

### Concurrency
- Use singleflight for request deduplication
- Test with race detector
- Use proper synchronization primitives

### Response Time
- Monitor gateway call latency
- Implement timeouts for external calls
- Use circuit breaker to prevent cascading failures

## Documentation

### Update Documentation When
- Adding new API endpoints
- Changing error codes
- Modifying architecture
- Adding new features
- Changing deployment process

### Documentation Files
- `README.md` - Project overview and quick start
- `docs/architecture.md` - Architecture details
- `docs/error-handling.md` - Error handling reference
- `docs/api.md` - API documentation
- `docs/deployment.md` - Deployment guide
- `docs/testing.md` - Testing guide
- `docs/development.md` - This file

## Getting Help

### Internal Resources
- Check existing documentation
- Review test files for examples
- Look at similar implementations

### External Resources
- Go documentation: https://go.dev/doc/
- Clean Architecture: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
- Go error handling: https://go.dev/blog/error-values-and-apis
