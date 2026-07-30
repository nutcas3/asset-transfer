package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := New(CodeInvalidRequest, "test message", http.StatusBadRequest)

	if err.Code != CodeInvalidRequest {
		t.Errorf("expected code %s, got %s", CodeInvalidRequest, err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("expected message 'test message', got '%s'", err.Message)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}

	if err.Layer != "application" {
		t.Errorf("expected layer 'application', got '%s'", err.Layer)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, CodeStorageError, "wrapped message", http.StatusInternalServerError, "service")

	if wrapped.Code != CodeStorageError {
		t.Errorf("expected code %s, got %s", CodeStorageError, wrapped.Code)
	}

	if wrapped.Message != "wrapped message" {
		t.Errorf("expected message 'wrapped message', got '%s'", wrapped.Message)
	}

	if wrapped.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, wrapped.HTTPStatus)
	}

	if wrapped.Layer != "service" {
		t.Errorf("expected layer 'service', got '%s'", wrapped.Layer)
	}

	if wrapped.Err != originalErr {
		t.Errorf("expected original error, got %v", wrapped.Err)
	}
}

func TestErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, CodeStorageError, "wrapped message", http.StatusInternalServerError, "service")

	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != originalErr {
		t.Errorf("expected original error after unwrap, got %v", unwrapped)
	}
}

func TestErrorIs(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, CodeStorageError, "wrapped message", http.StatusInternalServerError, "service")

	if !errors.Is(wrapped, originalErr) {
		t.Errorf("errors.Is should return true for wrapped error")
	}
}

func TestErrorAs(t *testing.T) {
	err := New(CodeInvalidRequest, "test message", http.StatusBadRequest)

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Errorf("errors.As should return true for AppError")
	}

	if appErr.Code != CodeInvalidRequest {
		t.Errorf("expected code %s, got %s", CodeInvalidRequest, appErr.Code)
	}
}

func TestErrorWithContext(t *testing.T) {
	err := New(CodeInvalidRequest, "test message", http.StatusBadRequest)
	err = err.WithContext("field", "quantity_units")
	err = err.WithContext("value", -1)

	if err.Context == nil {
		t.Errorf("expected context to be initialized")
	}

	if err.Context["field"] != "quantity_units" {
		t.Errorf("expected context field 'quantity_units', got '%s'", err.Context["field"])
	}

	if err.Context["value"] != -1 {
		t.Errorf("expected context value -1, got %v", err.Context["value"])
	}
}

func TestGetHTTPStatus(t *testing.T) {
	err := New(CodeInvalidRequest, "test message", http.StatusBadRequest)
	if err.GetHTTPStatus() != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.GetHTTPStatus())
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		wantCode   ErrorCode
		wantStatus int
	}{
		{
			name:       "ErrInvalidRequest",
			err:        ErrInvalidRequest,
			wantCode:   CodeInvalidRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrIdempotencyKeyMissing",
			err:        ErrIdempotencyKeyMissing,
			wantCode:   CodeIdempotencyKeyMissing,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrPayloadConflict",
			err:        ErrPayloadConflict,
			wantCode:   CodePayloadConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "ErrTransferNotFound",
			err:        ErrTransferNotFound,
			wantCode:   CodeTransferNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrGatewayTimeout",
			err:        ErrGatewayTimeout,
			wantCode:   CodeGatewayTimeout,
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "ErrGatewayRejected",
			err:        ErrGatewayRejected,
			wantCode:   CodeGatewayRejected,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("expected code %s, got %s", tt.wantCode, tt.err.Code)
			}
			if tt.err.HTTPStatus != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, tt.err.HTTPStatus)
			}
		})
	}
}

func TestErrorString(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, CodeStorageError, "wrapped message", http.StatusInternalServerError, "service")

	errStr := wrapped.Error()
	expected := "wrapped message: original error"
	if errStr != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, errStr)
	}

	// Test error without underlying error
	simple := New(CodeInvalidRequest, "simple message", http.StatusBadRequest)
	errStr = simple.Error()
	expected = "simple message"
	if errStr != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, errStr)
	}
}

func TestChainedWrapping(t *testing.T) {
	// Simulate error chain through layers
	dbErr := errors.New("connection failed")
	storageErr := Wrap(dbErr, CodeStorageError, "failed to query user", http.StatusInternalServerError, "storage")
	serviceErr := Wrap(storageErr, CodeInvalidRequest, "failed to create user", http.StatusBadRequest, "service")

	// Test that we can unwrap to the original error
	if !errors.Is(serviceErr, dbErr) {
		t.Errorf("should be able to unwrap to original database error")
	}

	// Test that we can extract each layer
	var storageAppErr *AppError
	if errors.As(serviceErr, &storageAppErr) {
		if storageAppErr.Layer != "service" {
			t.Errorf("expected service layer, got %s", storageAppErr.Layer)
		}
	}

	// Unwrap once to get storage layer
	unwrapped := errors.Unwrap(serviceErr)
	var storageErr2 *AppError
	if errors.As(unwrapped, &storageErr2) {
		if storageErr2.Layer != "storage" {
			t.Errorf("expected storage layer, got %s", storageErr2.Layer)
		}
	}
}
