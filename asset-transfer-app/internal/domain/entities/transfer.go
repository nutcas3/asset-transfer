package entities

import "time"

// Status represents the state of a transfer
type Status string

const (
	StatusSubmitting      Status = "submitting"
	StatusSubmitted       Status = "submitted"
	StatusRejected        Status = "rejected"
	StatusRetryableFailure Status = "retryable_failure"
)

// TransferRequest represents a client transfer request
type TransferRequest struct {
	FromAccount   string `json:"from_account"`
	ToAccount     string `json:"to_account"`
	AssetID       string `json:"asset_id"`
	QuantityUnits int    `json:"quantity_units"`
}

// Transfer represents a transfer record
type Transfer struct {
	ID             string          `json:"id"`
	IdempotencyKey string          `json:"idempotency_key"`
	Request        TransferRequest `json:"request"`
	Status         Status          `json:"status"`
	GatewayRef     string          `json:"gateway_ref,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// IsTerminal returns true if the status is terminal
func (t *Transfer) IsTerminal() bool {
	return t.Status == StatusSubmitted || t.Status == StatusRejected
}
