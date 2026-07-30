package httpdelivery

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"asset-transfer-app/internal/domain/entities"
	apperrors "asset-transfer-app/internal/errors"
	"asset-transfer-app/internal/usecases"
)

// Handler handles HTTP requests
type Handler struct {
	transferUseCase *usecases.TransferUseCase
}

// NewHandler creates a new HTTP handler
func NewHandler(transferUseCase *usecases.TransferUseCase) *Handler {
	return &Handler{
		transferUseCase: transferUseCase,
	}
}

// CreateTransfer handles POST /v1/transfers
func (h *Handler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get idempotency key
	idempotencyKey := r.Header.Get("Idempotency-Key")
	log.Printf("[INFO] POST /v1/transfers - idempotency_key=%s", idempotencyKey)

	if idempotencyKey == "" {
		log.Printf("[WARN] POST /v1/transfers - missing idempotency key")
		h.writeAppError(w, apperrors.ErrIdempotencyKeyMissing)
		return
	}

	// Parse request body with strict field checking
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var req entities.TransferRequest
	if err := decoder.Decode(&req); err != nil {
		log.Printf("[WARN] POST /v1/transfers - invalid JSON body: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid JSON body or unknown fields")
		return
	}

	log.Printf("[INFO] POST /v1/transfers - request: from=%s, to=%s, asset=%s, quantity=%d",
		req.FromAccount, req.ToAccount, req.AssetID, req.QuantityUnits)

	// Create transfer
	t, err := h.transferUseCase.CreateTransfer(r.Context(), idempotencyKey, req)
	if err != nil {
		log.Printf("[ERROR] POST /v1/transfers - failed to create transfer: %v", err)
		h.handleError(w, err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("[INFO] POST /v1/transfers - success: id=%s, status=%s, duration=%s",
		t.ID, t.Status, duration)

	// Return response based on status
	switch t.Status {
	case entities.StatusSubmitted:
		h.writeJSON(w, http.StatusCreated, t)
	case entities.StatusRejected:
		h.writeJSON(w, http.StatusUnprocessableEntity, t)
	case entities.StatusRetryableFailure:
		h.writeJSON(w, http.StatusServiceUnavailable, t)
	default:
		h.writeJSON(w, http.StatusCreated, t)
	}
}

// GetTransfer handles GET /v1/transfers/{id}
func (h *Handler) GetTransfer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/v1/transfers/"):]
	log.Printf("[INFO] GET /v1/transfers/%s", id)

	if id == "" {
		log.Printf("[WARN] GET /v1/transfers - missing transfer ID")
		h.writeError(w, http.StatusNotFound, "Transfer ID is required")
		return
	}

	t, err := h.transferUseCase.GetTransfer(r.Context(), id)
	if err != nil {
		log.Printf("[ERROR] GET /v1/transfers/%s - failed to get transfer: %v", id, err)
		h.handleError(w, err)
		return
	}

	log.Printf("[INFO] GET /v1/transfers/%s - success: status=%s", id, t.Status)
	h.writeJSON(w, http.StatusOK, t)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a simple error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// writeAppError writes a structured AppError response
func (h *Handler) writeAppError(w http.ResponseWriter, appErr *apperrors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.GetHTTPStatus())
	json.NewEncoder(w).Encode(appErr)
}

// handleError interprets errors and writes appropriate responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	// Check if it's an AppError
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		h.writeAppError(w, appErr)
		return
	}

	// Fallback for unexpected errors
	h.writeError(w, http.StatusInternalServerError, "Internal server error")
}
