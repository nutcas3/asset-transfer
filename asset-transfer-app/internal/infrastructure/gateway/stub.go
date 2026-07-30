package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"asset-transfer-app/internal/domain/entities"
	"asset-transfer-app/internal/domain/repositories"
)

// StubGateway is a deterministic in-process stub for testing
// It implements the GatewayRepository interface from the domain layer
type StubGateway struct {
	mu              sync.Mutex
	requestCount    int
	rejectCount     int
	timeoutCount    int
	unavailableCount int
	
	// Configurable behavior
	RejectAfter      int
	TimeoutAfter     int
	UnavailableAfter int
}

// NewStubGateway creates a new stub gateway
func NewStubGateway() repositories.GatewayRepository {
	return &StubGateway{
		RejectAfter:      9999, // Never reject by default
		TimeoutAfter:     9999, // Never timeout by default
		UnavailableAfter: 9999, // Never be unavailable by default
	}
}

// Submit simulates a gateway submission with deterministic behavior
func (sg *StubGateway) Submit(ctx context.Context, idempotencyKey string, request entities.TransferRequest) (string, error) {
	sg.mu.Lock()
	sg.requestCount++
	count := sg.requestCount
	sg.mu.Unlock()

	// Simulate timeout
	if count >= sg.TimeoutAfter {
		sg.mu.Lock()
		sg.timeoutCount++
		sg.mu.Unlock()
		select {
		case <-ctx.Done():
			return "", repositories.ErrGatewayTimeout
		case <-time.After(100 * time.Millisecond):
			return "", repositories.ErrGatewayTimeout
		}
	}

	// Simulate unavailability
	if count >= sg.UnavailableAfter {
		sg.mu.Lock()
		sg.unavailableCount++
		sg.mu.Unlock()
		return "", repositories.ErrGatewayUnavailable
	}

	// Simulate deterministic rejection
	if count >= sg.RejectAfter {
		sg.mu.Lock()
		sg.rejectCount++
		sg.mu.Unlock()
		return "", repositories.ErrGatewayRejected
	}

	// Success - return a deterministic reference
	ref := fmt.Sprintf("gw-ref-%d", count)
	return ref, nil
}

// GetStats returns statistics for testing
func (sg *StubGateway) GetStats() (request, reject, timeout, unavailable int) {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	return sg.requestCount, sg.rejectCount, sg.timeoutCount, sg.unavailableCount
}

// Reset resets the stub state (for testing)
func (sg *StubGateway) Reset() {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.requestCount = 0
	sg.rejectCount = 0
	sg.timeoutCount = 0
	sg.unavailableCount = 0
}
