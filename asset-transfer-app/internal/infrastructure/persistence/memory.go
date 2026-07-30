package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"asset-transfer-app/internal/domain/entities"
	"asset-transfer-app/internal/domain/repositories"
)

// InMemoryTransferRepository implements TransferRepository using in-memory storage
type InMemoryTransferRepository struct {
	mu          sync.RWMutex
	transfers   map[string]*entities.Transfer
	idempotency map[string]string // idempotency key -> transfer ID
	payloadHash map[string]string // idempotency key -> payload hash
}

// NewInMemoryTransferRepository creates a new in-memory transfer repository
func NewInMemoryTransferRepository() repositories.TransferRepository {
	return &InMemoryTransferRepository{
		transfers:   make(map[string]*entities.Transfer),
		idempotency: make(map[string]string),
		payloadHash: make(map[string]string),
	}
}

// Store saves a transfer record
func (r *InMemoryTransferRepository) Store(ctx context.Context, transfer *entities.Transfer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	transfer.UpdatedAt = time.Now()

	// If it's a new transfer, set created time
	if existing, exists := r.transfers[transfer.ID]; !exists {
		transfer.CreatedAt = time.Now()
	} else {
		transfer.CreatedAt = existing.CreatedAt
	}

	r.transfers[transfer.ID] = transfer
	r.idempotency[transfer.IdempotencyKey] = transfer.ID

	// Store payload hash for idempotency checking
	hash, err := hashPayload(transfer.Request)
	if err == nil {
		r.payloadHash[transfer.IdempotencyKey] = hash
	}

	return nil
}

// Get retrieves a transfer by ID
func (r *InMemoryTransferRepository) Get(ctx context.Context, id string) (*entities.Transfer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.transfers[id]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	return &entities.Transfer{
		ID:             t.ID,
		IdempotencyKey: t.IdempotencyKey,
		Request:        t.Request,
		Status:         t.Status,
		GatewayRef:     t.GatewayRef,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}, true
}

// GetByIdempotencyKey retrieves a transfer by idempotency key
func (r *InMemoryTransferRepository) GetByIdempotencyKey(ctx context.Context, key string) (*entities.Transfer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transferID, exists := r.idempotency[key]
	if !exists {
		return nil, false
	}

	t, exists := r.transfers[transferID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	return &entities.Transfer{
		ID:             t.ID,
		IdempotencyKey: t.IdempotencyKey,
		Request:        t.Request,
		Status:         t.Status,
		GatewayRef:     t.GatewayRef,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}, true
}

// CheckPayloadConflict checks if the idempotency key is already bound to a different payload
func (r *InMemoryTransferRepository) CheckPayloadConflict(ctx context.Context, key string, request entities.TransferRequest) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	existingHash, exists := r.payloadHash[key]
	if !exists {
		return false, nil
	}

	newHash, err := hashPayload(request)
	if err != nil {
		return false, err
	}

	return existingHash != newHash, nil
}

// hashPayload creates a deterministic hash of the request payload
func hashPayload(request entities.TransferRequest) (string, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}
