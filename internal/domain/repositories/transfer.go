package repositories

import (
	"context"

	"asset-transfer-app/internal/domain/entities"
)

// TransferRepository defines the interface for transfer persistence
// This is a domain interface - infrastructure layer will implement this
type TransferRepository interface {
	// Store saves a transfer record
	Store(ctx context.Context, transfer *entities.Transfer) error
	
	// Get retrieves a transfer by ID
	Get(ctx context.Context, id string) (*entities.Transfer, bool)
	
	// GetByIdempotencyKey retrieves a transfer by idempotency key
	GetByIdempotencyKey(ctx context.Context, key string) (*entities.Transfer, bool)
	
	// CheckPayloadConflict checks if the idempotency key is used with a different payload
	CheckPayloadConflict(ctx context.Context, key string, request entities.TransferRequest) (bool, error)
}
