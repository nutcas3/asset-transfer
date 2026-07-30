# Error Handling Reference

The service uses structured error handling with machine-readable codes and layer-specific interpretation.

## Structured Error Type

```go
type AppError struct {
    Code       ErrorCode `json:"code"`
    Message    string    `json:"message"`
    HTTPStatus int       `json:"-"`
    Err        error     `json:"-"`
    Layer      string    `json:"-"`
    Context    map[string]interface{} `json:"context,omitempty"`
}
```

**Key Features:**
- **Machine-readable codes** (e.g., `IDEMPOTENCY_KEY_MISSING`, `GATEWAY_TIMEOUT`)
- **Pre-defined HTTP status mappings** for consistent API responses
- **Layer attribution** for debugging (service, gateway, storage)
- **Optional context** for additional debugging information
- **Error wrapping** with `Unwrap()` for `errors.Is()` compatibility

## Layer-Specific Error Definitions

**Service Layer:**
```go
var (
    ErrInvalidRequest        = New(CodeInvalidRequest, "Invalid request", http.StatusBadRequest)
    ErrIdempotencyKeyMissing = New(CodeIdempotencyKeyMissing, "Idempotency key is required", http.StatusBadRequest)
    ErrPayloadConflict       = New(CodePayloadConflict, "Idempotency key already used with different payload", http.StatusConflict)
    ErrTransferNotFound      = New(CodeTransferNotFound, "Transfer not found", http.StatusNotFound)
)
```

**Gateway Layer:**
```go
var (
    ErrGatewayTimeout     = New(CodeGatewayTimeout, "Gateway timeout", http.StatusServiceUnavailable)
    ErrGatewayUnavailable = New(CodeGatewayUnavailable, "Gateway unavailable", http.StatusServiceUnavailable)
    ErrGatewayRejected    = New(CodeGatewayRejected, "Gateway rejected transfer", http.StatusUnprocessableEntity)
    ErrCircuitOpen       = New(CodeCircuitOpen, "Circuit breaker is open", http.StatusServiceUnavailable)
)
```

**Storage Layer:**
```go
var (
    ErrStorageError = New(CodeStorageError, "Storage error", http.StatusInternalServerError)
)
```

## Layer Interpretation Pattern

**Use Cases Layer** wraps infrastructure errors with business context:
```go
if err := s.storage.Store(t); err != nil {
    return nil, apperrors.Wrap(err, apperrors.CodeStorageError, 
        "failed to store transfer", http.StatusInternalServerError, "service")
}
```

**Delivery Layer** interprets all errors uniformly:
```go
func (h *Handler) handleError(w http.ResponseWriter, err error) {
    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        h.writeAppError(w, appErr)
        return
    }
    // Fallback for unexpected errors
    h.writeError(w, http.StatusInternalServerError, "Internal server error")
}
```

## Error Response Format

**Structured Response:**
```json
{
  "code": "STORAGE_ERROR",
  "message": "failed to store transfer",
  "layer": "service",
  "context": {
    "transfer_id": "transfer-123"
  }
}
```

## Error Code Hierarchy

```
Application Errors
├── Service Layer
│   ├── INVALID_REQUEST
│   ├── IDEMPOTENCY_KEY_MISSING
│   ├── PAYLOAD_CONFLICT
│   └── TRANSFER_NOT_FOUND
├── Gateway Layer
│   ├── GATEWAY_TIMEOUT
│   ├── GATEWAY_UNAVAILABLE
│   ├── GATEWAY_REJECTED
│   └── CIRCUIT_OPEN
└── Storage Layer
    └── STORAGE_ERROR
```

## Usage Examples

**Use Cases Layer:**
```go
// Simple error
if idempotencyKey == "" {
    return apperrors.ErrIdempotencyKeyMissing
}

// Wrapped error with context
if err := s.storage.Store(t); err != nil {
    return nil, apperrors.Wrap(err, apperrors.CodeStorageError, 
        "failed to store transfer", http.StatusInternalServerError, "service")
}

// Error with additional context
if err != nil {
    return apperrors.ErrInvalidRequest.WithContext("field", "quantity_units")
}
```

**Delivery Layer:**
```go
// Automatic error interpretation
t, err := h.transferUseCase.CreateTransfer(r.Context(), idempotencyKey, req)
if err != nil {
    h.handleError(w, err)  // Handles all error types uniformly
    return
}

// Direct AppError response
if idempotencyKey == "" {
    h.writeAppError(w, apperrors.ErrIdempotencyKeyMissing)
    return
}
```

## Error Handling Benefits

- **Frontend-Friendly**: Machine-readable codes enable conditional UI logic
- **Clean Layer Boundaries**: Handlers don't need to know internal error types
- **Prevents Verbosity**: No error message stacking, clean API responses
- **Better Debugging**: Layer attribution and structured context
- **Type-Safe**: Works with `errors.Is()` and `errors.As()`

## Integration with Clean Architecture

The error handling integrates with Clean Architecture layers:

1. **Domain Layer** - Defines error types and codes
2. **Use Cases Layer** - Wraps infrastructure errors with business context
3. **Infrastructure Layer** - Returns low-level errors
4. **Delivery Layer** - Interprets errors and maps to HTTP responses

This combination provides:
- Clean architectural boundaries
- Consistent error handling across layers
- Machine-readable error responses
- Easy debugging and monitoring
