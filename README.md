# Reliable Asset Transfer Gateway

A Go service that accepts instructions to transfer units of a tokenized asset through an unreliable external blockchain gateway. The service handles retries, concurrent requests, failure classification, and idempotency.

## Design Overview

The service implements a robust transfer system with:

- **Idempotency**: Client-provided keys prevent duplicate transfers
- **Concurrency Control**: Singleflight pattern ensures concurrent requests share one gateway call
- **Circuit Breaker**: Prevents cascading failures when the gateway is unavailable
- **State Machine**: Explicit transfer states (submitting, submitted, rejected, retryable_failure)
- **Deterministic Gateway Stub**: In-process stub for testing without real blockchain dependencies

## Architecture

The service follows **Clean Architecture** principles with clear layer separation and dependency inversion:

```
┌─────────────────────────────────────────────────────────┐
│                    Delivery Layer                        │
│                  (HTTP/gRPC/CLI)                         │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                   Use Cases Layer                        │
│              (Business Logic Orchestration)               │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                    Domain Layer                          │
│              (Entities + Repository Interfaces)          │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                Infrastructure Layer                       │
│         (Persistence, Gateway, External Services)        │
└─────────────────────────────────────────────────────────┘
```

### Directory Structure

```
internal/
├── domain/                    # Core business logic (no external dependencies)
│   ├── entities/              # Pure business entities
│   │   └── transfer.go       # Transfer, TransferRequest, Status
│   └── repositories/          # Repository interfaces (domain contracts)
│       ├── transfer.go       # TransferRepository interface
│       └── gateway.go        # GatewayRepository interface
├── usecases/                 # Business logic (orchestration)
│   └── transfer.go           # TransferUseCase
├── infrastructure/            # External implementations
│   ├── persistence/          # Storage implementations
│   │   └── memory.go        # InMemoryTransferRepository
│   └── gateway/              # Gateway implementations
│       ├── stub.go           # StubGateway
│       └── circuit_breaker.go # CircuitBreaker wrapper
├── delivery/                 # Delivery mechanisms
│   └── http/                 # HTTP handlers
│       └── handler.go        # HTTP delivery
└── errors/                   # Structured error handling
    └── errors.go            # AppError types
```

### Key Principles

- **Domain Isolation**: Pure business entities without infrastructure dependencies
- **Dependency Inversion**: High-level modules depend on abstractions (interfaces)
- **Interface Segregation**: Small, focused interfaces for each concern
- **Single Responsibility**: Each layer has one clear purpose

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation and [docs/error-handling.md](docs/error-handling.md) for error handling reference.

## Documentation

- [API Documentation](docs/api.md) - Detailed API reference with endpoints, request/response formats
- [Deployment Guide](docs/deployment.md) - How to deploy and run the service
- [Testing Guide](docs/testing.md) - How to run tests and testing strategies
- [Development Guide](docs/development.md) - Setup and development workflow

## Design Decisions and Tradeoffs

### 1. In-Memory Storage
**Decision**: Used in-memory storage for simplicity within the timebox.

**Tradeoffs**:
- **Pros**: Simple, fast, no external dependencies, sufficient for demonstration
- **Cons**: Data lost on restart, not suitable for production, no persistence

**Production Alternative**: PostgreSQL with proper transaction handling and idempotency table.

### 2. Singleflight Pattern
**Decision**: Used `golang.org/x/sync/singleflight` for concurrent request deduplication.

**Tradeoffs**:
- **Pros**: Simple, built-in, prevents race conditions effectively
- **Cons**: Singleflight keys are not garbage collected until all callers complete, potential memory leak if many unique keys

**Alternative**: Could implement custom deduplication with TTL-based cleanup.

### 3. Circuit Breaker Implementation
**Decision**: Custom circuit breaker implementation instead of using a library.

**Tradeoffs**:
- **Pros**: Full control over behavior, easy to test, no external dependency
- **Cons**: Less battle-tested than established libraries, may miss edge cases

**Alternative**: Libraries like `github.com/sony/gobreaker` provide more features.

### 4. State Machine
**Decision**: Explicit state model with four states.

**Tradeoffs**:
- **Pros**: Clear semantics, easy to reason about, prevents invalid transitions
- **Cons**: Additional complexity for simple use case

**Rationale**: Required by specification and enables proper retry logic.

### 5. JSON Field Validation
**Decision**: Used `DisallowUnknownFields()` to reject unknown JSON fields.

**Tradeoffs**:
- **Pros**: Strict validation prevents API drift, catches client errors early
- **Cons**: Less flexible for API evolution, may break clients

**Alternative**: Could log warnings for unknown fields but still accept them.

## Assignment Questions

### 1. What consistency problem appears if this service runs as two or more replicas?

**Problem**: Each replica maintains its own in-memory storage, leading to:
- **Idempotency Violation**: The same idempotency key could be used with different payloads across replicas
- **Duplicate Submissions**: Concurrent requests to different replicas could both call the gateway
- **Data Inconsistency**: GET requests to different replicas could return different states

**Solution**: Use a shared, strongly consistent data store (PostgreSQL with proper locking) and distributed locking (Redis) for idempotency enforcement across replicas.

### 2. How would you persist idempotency safely in PostgreSQL?

**Schema**:
```sql
CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    payload_hash VARCHAR(64) NOT NULL,
    transfer_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);
```

**Strategy**:
1. Use `INSERT ... ON CONFLICT` to atomically check and insert idempotency keys
2. Store SHA-256 hash of the normalized JSON payload
3. Add TTL for cleanup (e.g., 24-48 hours)
4. Use database transactions to ensure atomicity with transfer creation
5. Add unique constraint on key to prevent duplicates

**Example**:
```sql
INSERT INTO idempotency_keys (key, payload_hash, transfer_id, expires_at)
VALUES ($1, $2, $3, NOW() + INTERVAL '24 hours')
ON CONFLICT (key) DO UPDATE SET
  transfer_id = EXCLUDED.transfer_id
RETURNING (payload_hash = $2) as payload_matches;
```

### 3. What happens if the process crashes after the gateway accepts a transfer but before the local record is updated?

**Problem**: The gateway has accepted the transfer (and possibly executed it on-chain), but the local record still shows `submitting` or doesn't exist. This creates a **consistency gap** between the actual blockchain state and the service's view.

**Consequences**:
- Client receives timeout/error but transfer may have succeeded
- Retry would attempt to resubmit, potentially causing duplicate transfers
- No audit trail of what actually happened

**Mitigation Strategies**:
1. **Outbox Pattern**: Write transfer record to database first, then send to gateway in background job
2. **Idempotency at Gateway Level**: Gateway's idempotency key prevents duplicate submissions
3. **Reconciliation Process**: Periodic job to query gateway and reconcile local state
4. **Write-Ahead Log**: Log all operations before execution for recovery
5. **Saga Pattern**: Implement compensating transactions for rollback

**Current Implementation**: The gateway stub is deterministic and always returns the same result for the same idempotency key, so retries are safe. In production, this would require the real gateway to support idempotency.

### 4. What additional states or reconciliation would a real blockchain transfer need before being called final?

**Additional States**:
- `pending_confirmation`: Transaction submitted to blockchain, awaiting confirmation
- `confirmed`: Transaction has required confirmations but not final
- `final`: Transaction is irreversible (enough confirmations)
- `failed`: Transaction failed on-chain (insufficient funds, etc.)
- `reconciling`: Discrepancy detected, under investigation

**Reconciliation Requirements**:
1. **Block Confirmation Monitoring**: Listen to blockchain events for transaction confirmations
2. **Periodic State Sync**: Query blockchain for current state of all pending transfers
3. **Discrepancy Detection**: Compare local state with blockchain state
4. **Automatic Recovery**: Update local state based on blockchain truth
5. **Manual Intervention**: Flag unreconcilable discrepancies for human review
6. **Event Sourcing**: Store all state transitions for audit trail and replay

**Example Flow**:
```
submitting → pending_confirmation → confirmed → final
                 ↓                    ↓
              failed              reconciling
```

## Time Spent

Approximately 3 hours:
- Architecture and data structures: 30 minutes
- Gateway and circuit breaker implementation: 45 minutes
- Storage and service logic: 45 minutes
- HTTP handlers and routing: 30 minutes
- Test implementation: 45 minutes
- Debugging and refinement: 15 minutes

## AI-Assisted Code

This implementation was created with assistance from AI (Cascade). The AI helped with:
- Initial architecture design
- Code structure and organization
- Test case design
- Debugging and error resolution

All code was reviewed and validated by the developer before submission.

## Testing

The service includes comprehensive tests. See [Testing Guide](docs/testing.md) for details on running tests and testing strategies.

All tests pass with `go test ./...` and `go test -race ./...`.
