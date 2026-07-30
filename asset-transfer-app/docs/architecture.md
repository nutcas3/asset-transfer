# Architecture Documentation

## Clean Architecture

The asset transfer service follows Clean Architecture principles with clear layer separation and dependency inversion.

### Architecture Layers

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

### Layer Responsibilities

#### Domain Layer (`internal/domain/`)

**Purpose:** Core business logic and entities, completely independent of infrastructure.

**Components:**
- **Entities:** Pure business objects (Transfer, TransferRequest, Status)
- **Repository Interfaces:** Contracts for external dependencies

**Key Characteristics:**
- No external dependencies
- Pure Go types and interfaces
- Contains business rules and invariants
- Infrastructure-agnostic

**Example:**
```go
// Domain entity - pure business logic
type Transfer struct {
    ID             string
    IdempotencyKey string
    Request        TransferRequest
    Status         Status
    GatewayRef     string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// Domain interface - contract for infrastructure
type TransferRepository interface {
    Store(ctx context.Context, transfer *Transfer) error
    Get(ctx context.Context, id string) (*Transfer, bool)
    GetByIdempotencyKey(ctx context.Context, key string) (*Transfer, bool)
    CheckPayloadConflict(ctx context.Context, key string, request TransferRequest) (bool, error)
}
```

#### Use Cases Layer (`internal/usecases/`)

**Purpose:** Business logic orchestration, depends only on domain interfaces.

**Components:**
- **Use Cases:** Business logic workflows (CreateTransfer, GetTransfer)

**Key Characteristics:**
- Depends on domain interfaces, not implementations
- Contains business rules and validation
- Orchestrates repository calls
- No HTTP or infrastructure concerns

**Example:**
```go
type TransferUseCase struct {
    transferRepo repositories.TransferRepository  // Interface, not implementation
    gatewayRepo  repositories.GatewayRepository   // Interface, not implementation
    sf           singleflight.Group
}

func (uc *TransferUseCase) CreateTransfer(ctx context.Context, idempotencyKey string, request entities.TransferRequest) (*entities.Transfer, error) {
    // Business logic using only interfaces
    if err := uc.validateRequest(idempotencyKey, request); err != nil {
        return nil, err
    }
    // ... orchestration logic
}
```

#### Infrastructure Layer (`internal/infrastructure/`)

**Purpose:** Concrete implementations of domain interfaces.

**Components:**
- **Persistence:** Storage implementations (InMemory, PostgreSQL, Redis)
- **Gateway:** External service implementations (Stub, Real blockchain)

**Key Characteristics:**
- Implements domain interfaces
- Contains infrastructure-specific logic
- Can be swapped without affecting business logic
- Framework and library dependencies live here

**Example:**
```go
// Infrastructure implementation
type InMemoryTransferRepository struct {
    mu          sync.RWMutex
    transfers   map[string]*entities.Transfer
    idempotency map[string]string
    payloadHash map[string]string
}

func (r *InMemoryTransferRepository) Store(ctx context.Context, transfer *entities.Transfer) error {
    // Implementation details
}
```

#### Delivery Layer (`internal/delivery/`)

**Purpose:** External interface to the application (HTTP, gRPC, CLI).

**Components:**
- **HTTP:** HTTP handlers and routing
- **gRPC:** gRPC service implementations (future)

**Key Characteristics:**
- Depends on use cases, not repositories
- Handles delivery-specific concerns (HTTP status codes, JSON serialization)
- Thin layer - delegates to use cases

**Example:**
```go
type Handler struct {
    transferUseCase *usecases.TransferUseCase  // Depends on use case, not repositories
}

func (h *Handler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
    // HTTP-specific logic
    var req entities.TransferRequest
    // ... parse request
    
    // Delegate to use case
    transfer, err := h.transferUseCase.CreateTransfer(r.Context(), idempotencyKey, req)
    // ... format response
}
```

### Dependency Flow

**Correct Dependencies (Inward):**
```
Delivery → Use Cases → Domain ← Infrastructure
```

**Incorrect Dependencies (Avoid):**
```
❌ Domain → Infrastructure
❌ Use Cases → Infrastructure
❌ Delivery → Infrastructure
```

### Key Principles

- **Domain Isolation**: Pure business entities without infrastructure dependencies
- **Dependency Inversion**: High-level modules depend on abstractions (interfaces)
- **Interface Segregation**: Small, focused interfaces for each concern
- **Single Responsibility**: Each layer has one clear purpose

See [docs/error-handling.md](error-handling.md) for detailed error handling documentation.
