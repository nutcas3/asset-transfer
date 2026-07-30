package repositories

import (
	"context"

	"asset-transfer-app/internal/domain/entities"
)

// GatewayRepository defines the interface for external blockchain gateway
// This is a domain interface - infrastructure layer will implement this
type GatewayRepository interface {
	// Submit submits a transfer to the blockchain gateway
	Submit(ctx context.Context, idempotencyKey string, request entities.TransferRequest) (string, error)
}

// Gateway errors
var (
	ErrGatewayRejected   = NewGatewayError("gateway rejected transfer")
	ErrGatewayTimeout    = NewGatewayError("gateway timeout")
	ErrGatewayUnavailable = NewGatewayError("gateway unavailable")
	ErrCircuitOpen       = NewGatewayError("circuit breaker is open")
)

// GatewayError represents a gateway-specific error
type GatewayError struct {
	Message string
}

func NewGatewayError(msg string) error {
	return &GatewayError{Message: msg}
}

func (e *GatewayError) Error() string {
	return e.Message
}
