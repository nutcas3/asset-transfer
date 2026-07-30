package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a machine-readable error code
type ErrorCode string

const (
	// Service layer error codes
	CodeInvalidRequest        ErrorCode = "INVALID_REQUEST"
	CodeIdempotencyKeyMissing ErrorCode = "IDEMPOTENCY_KEY_MISSING"
	CodePayloadConflict       ErrorCode = "PAYLOAD_CONFLICT"
	CodeTransferNotFound      ErrorCode = "TRANSFER_NOT_FOUND"
	CodeInternalError         ErrorCode = "INTERNAL_ERROR"

	// Gateway layer error codes
	CodeGatewayTimeout     ErrorCode = "GATEWAY_TIMEOUT"
	CodeGatewayUnavailable ErrorCode = "GATEWAY_UNAVAILABLE"
	CodeGatewayRejected    ErrorCode = "GATEWAY_REJECTED"
	CodeCircuitOpen        ErrorCode = "CIRCUIT_OPEN"

	// Storage layer error codes
	CodeStorageError ErrorCode = "STORAGE_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode      `json:"code"`
	Message    string         `json:"message"`
	HTTPStatus int            `json:"-"`
	Err        error          `json:"-"`
	Layer      string         `json:"-"`
	Context    map[string]any `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Layer:      "application",
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string, httpStatus int, layer string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
		Layer:      layer,
	}
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// HTTPStatus returns the HTTP status code
func (e *AppError) GetHTTPStatus() int {
	return e.HTTPStatus
}

// Service layer errors
var (
	ErrInvalidRequest        = New(CodeInvalidRequest, "Invalid request", http.StatusBadRequest)
	ErrIdempotencyKeyMissing = New(CodeIdempotencyKeyMissing, "Idempotency key is required", http.StatusBadRequest)
	ErrPayloadConflict       = New(CodePayloadConflict, "Idempotency key already used with different payload", http.StatusConflict)
	ErrTransferNotFound      = New(CodeTransferNotFound, "Transfer not found", http.StatusNotFound)
	ErrInternalError         = New(CodeInternalError, "Internal server error", http.StatusInternalServerError)
)

// Gateway layer errors
var (
	ErrGatewayTimeout     = New(CodeGatewayTimeout, "Gateway timeout", http.StatusServiceUnavailable)
	ErrGatewayUnavailable = New(CodeGatewayUnavailable, "Gateway unavailable", http.StatusServiceUnavailable)
	ErrGatewayRejected    = New(CodeGatewayRejected, "Gateway rejected transfer", http.StatusUnprocessableEntity)
	ErrCircuitOpen        = New(CodeCircuitOpen, "Circuit breaker is open", http.StatusServiceUnavailable)
)

// Storage layer errors
var (
	ErrStorageError = New(CodeStorageError, "Storage error", http.StatusInternalServerError)
)
