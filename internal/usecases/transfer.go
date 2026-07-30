package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"asset-transfer-app/internal/domain/entities"
	"asset-transfer-app/internal/domain/repositories"
	apperrors "asset-transfer-app/internal/errors"

	"golang.org/x/sync/singleflight"
)

// TransferUseCase handles transfer business logic
type TransferUseCase struct {
	transferRepo repositories.TransferRepository
	gatewayRepo  repositories.GatewayRepository
	sf           singleflight.Group
}

// NewTransferUseCase creates a new transfer use case
func NewTransferUseCase(
	transferRepo repositories.TransferRepository,
	gatewayRepo repositories.GatewayRepository,
) *TransferUseCase {
	return &TransferUseCase{
		transferRepo: transferRepo,
		gatewayRepo:  gatewayRepo,
	}
}

// CreateTransfer creates a new transfer with idempotency handling
func (uc *TransferUseCase) CreateTransfer(ctx context.Context, idempotencyKey string, request entities.TransferRequest) (*entities.Transfer, error) {
	// Validate request
	if err := uc.validateRequest(idempotencyKey, request); err != nil {
		return nil, err
	}

	// Check for payload conflict
	conflict, err := uc.transferRepo.CheckPayloadConflict(ctx, idempotencyKey, request)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeStorageError, "failed to check payload conflict", 500, "usecase")
	}
	if conflict {
		return nil, apperrors.ErrPayloadConflict
	}

	// Use singleflight to ensure concurrent requests with same key share one submission
	result, err, _ := uc.sf.Do(idempotencyKey, func() (any, error) {
		return uc.processTransfer(ctx, idempotencyKey, request)
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.Transfer), nil
}

// GetTransfer retrieves a transfer by ID
func (uc *TransferUseCase) GetTransfer(ctx context.Context, id string) (*entities.Transfer, error) {
	transfer, exists := uc.transferRepo.Get(ctx, id)
	if !exists {
		return nil, apperrors.ErrTransferNotFound
	}
	return transfer, nil
}

// processTransfer processes a transfer (called within singleflight)
func (uc *TransferUseCase) processTransfer(ctx context.Context, idempotencyKey string, request entities.TransferRequest) (*entities.Transfer, error) {
	// Check if transfer already exists
	if existing, exists := uc.transferRepo.GetByIdempotencyKey(ctx, idempotencyKey); exists {
		log.Printf("[INFO] processTransfer - existing transfer found: id=%s, status=%s", existing.ID, existing.Status)

		// If terminal, return cached result
		if existing.IsTerminal() {
			log.Printf("[INFO] processTransfer - returning cached terminal result: id=%s, status=%s", existing.ID, existing.Status)
			return existing, nil
		}

		// If retryable failure, retry
		if existing.Status == entities.StatusRetryableFailure {
			log.Printf("[INFO] processTransfer - retrying transfer: id=%s", existing.ID)
			return uc.retryTransfer(ctx, existing)
		}

		// If still submitting, return current state
		log.Printf("[INFO] processTransfer - transfer still submitting: id=%s", existing.ID)
		return existing, nil
	}

	// Create new transfer record
	transfer := &entities.Transfer{
		ID:             generateID(),
		IdempotencyKey: idempotencyKey,
		Request:        request,
		Status:         entities.StatusSubmitting,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	log.Printf("[INFO] processTransfer - creating new transfer: id=%s, idempotency_key=%s", transfer.ID, idempotencyKey)

	// Store initial record
	if err := uc.transferRepo.Store(ctx, transfer); err != nil {
		log.Printf("[ERROR] processTransfer - failed to store transfer: id=%s, error=%v", transfer.ID, err)
		return nil, apperrors.Wrap(err, apperrors.CodeStorageError, "failed to store transfer", 500, "usecase")
	}

	// Submit to gateway
	log.Printf("[INFO] processTransfer - submitting to gateway: id=%s", transfer.ID)
	gatewayRef, err := uc.gatewayRepo.Submit(ctx, idempotencyKey, request)

	// Update transfer based on result
	if err != nil {
		if isGatewayRejected(err) {
			transfer.Status = entities.StatusRejected
			log.Printf("[WARN] processTransfer - gateway rejected transfer: id=%s", transfer.ID)
		} else {
			transfer.Status = entities.StatusRetryableFailure
			log.Printf("[WARN] processTransfer - gateway retryable failure: id=%s, error=%v", transfer.ID, err)
		}
	} else {
		transfer.Status = entities.StatusSubmitted
		transfer.GatewayRef = gatewayRef
		log.Printf("[INFO] processTransfer - gateway accepted transfer: id=%s, gateway_ref=%s", transfer.ID, gatewayRef)
	}

	transfer.UpdatedAt = time.Now()

	// Store updated record
	if err := uc.transferRepo.Store(ctx, transfer); err != nil {
		log.Printf("[ERROR] processTransfer - failed to update transfer: id=%s, error=%v", transfer.ID, err)
		return nil, apperrors.Wrap(err, apperrors.CodeStorageError, "failed to update transfer", 500, "usecase")
	}

	log.Printf("[INFO] processTransfer - transfer completed: id=%s, status=%s", transfer.ID, transfer.Status)
	return transfer, nil
}

// retryTransfer retries a failed transfer
func (uc *TransferUseCase) retryTransfer(ctx context.Context, transfer *entities.Transfer) (*entities.Transfer, error) {
	log.Printf("[INFO] retryTransfer - retrying transfer: id=%s", transfer.ID)

	// Update status to submitting
	transfer.Status = entities.StatusSubmitting
	transfer.UpdatedAt = time.Now()
	if err := uc.transferRepo.Store(ctx, transfer); err != nil {
		log.Printf("[ERROR] retryTransfer - failed to update transfer: id=%s, error=%v", transfer.ID, err)
		return nil, apperrors.Wrap(err, apperrors.CodeStorageError, "failed to update transfer", 500, "usecase")
	}

	// Submit to gateway
	log.Printf("[INFO] retryTransfer - submitting to gateway: id=%s", transfer.ID)
	gatewayRef, err := uc.gatewayRepo.Submit(ctx, transfer.IdempotencyKey, transfer.Request)

	// Update transfer based on result
	if err != nil {
		if isGatewayRejected(err) {
			transfer.Status = entities.StatusRejected
			log.Printf("[WARN] retryTransfer - gateway rejected transfer: id=%s", transfer.ID)
		} else {
			transfer.Status = entities.StatusRetryableFailure
			log.Printf("[WARN] retryTransfer - gateway retryable failure: id=%s, error=%v", transfer.ID, err)
		}
	} else {
		transfer.Status = entities.StatusSubmitted
		transfer.GatewayRef = gatewayRef
		log.Printf("[INFO] retryTransfer - gateway accepted transfer: id=%s, gateway_ref=%s", transfer.ID, gatewayRef)
	}

	transfer.UpdatedAt = time.Now()

	// Store updated record
	if err := uc.transferRepo.Store(ctx, transfer); err != nil {
		log.Printf("[ERROR] retryTransfer - failed to update transfer: id=%s, error=%v", transfer.ID, err)
		return nil, apperrors.Wrap(err, apperrors.CodeStorageError, "failed to update transfer", 500, "usecase")
	}

	log.Printf("[INFO] retryTransfer - retry completed: id=%s, status=%s", transfer.ID, transfer.Status)
	return transfer, nil
}

// validateRequest validates the transfer request
func (uc *TransferUseCase) validateRequest(idempotencyKey string, request entities.TransferRequest) error {
	if idempotencyKey == "" {
		return apperrors.ErrIdempotencyKeyMissing
	}

	if request.FromAccount == "" {
		return apperrors.Wrap(apperrors.ErrInvalidRequest, apperrors.CodeInvalidRequest, "from_account is required", 400, "usecase")
	}

	if request.ToAccount == "" {
		return apperrors.Wrap(apperrors.ErrInvalidRequest, apperrors.CodeInvalidRequest, "to_account is required", 400, "usecase")
	}

	if request.AssetID == "" {
		return apperrors.Wrap(apperrors.ErrInvalidRequest, apperrors.CodeInvalidRequest, "asset_id is required", 400, "usecase")
	}

	if request.FromAccount == request.ToAccount {
		return apperrors.Wrap(apperrors.ErrInvalidRequest, apperrors.CodeInvalidRequest, "source and destination accounts must differ", 400, "usecase")
	}

	if request.QuantityUnits <= 0 {
		return apperrors.Wrap(apperrors.ErrInvalidRequest, apperrors.CodeInvalidRequest, "quantity_units must be a positive integer", 400, "usecase")
	}

	return nil
}

// generateID generates a unique transfer ID
func generateID() string {
	return fmt.Sprintf("transfer-%d", time.Now().UnixNano())
}

// isGatewayRejected checks if the error is a gateway rejection
func isGatewayRejected(err error) bool {
	if gwErr, ok := err.(*repositories.GatewayError); ok {
		return gwErr == repositories.ErrGatewayRejected
	}
	return false
}
