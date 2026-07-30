package gateway

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"asset-transfer-app/internal/domain/entities"
	"asset-transfer-app/internal/domain/repositories"
)

// CircuitBreaker implements a circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    int
	lastFailureTime time.Time
	halfOpenSuccess bool

	// Configurable values
	failureThreshold int
	cooldownDuration time.Duration
	timeout          time.Duration

	// The underlying gateway
	gateway repositories.GatewayRepository
}

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// NewCircuitBreaker creates a new circuit breaker with default values
func NewCircuitBreaker(gateway repositories.GatewayRepository) repositories.GatewayRepository {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: 3,
		cooldownDuration: 5 * time.Second,
		timeout:          10 * time.Second,
		gateway:          gateway,
	}
}

// WithFailureThreshold sets the failure threshold
func (cb *CircuitBreaker) WithFailureThreshold(threshold int) *CircuitBreaker {
	cb.failureThreshold = threshold
	return cb
}

// WithCooldownDuration sets the cooldown duration
func (cb *CircuitBreaker) WithCooldownDuration(duration time.Duration) *CircuitBreaker {
	cb.cooldownDuration = duration
	return cb
}

// WithTimeout sets the gateway timeout
func (cb *CircuitBreaker) WithTimeout(timeout time.Duration) *CircuitBreaker {
	cb.timeout = timeout
	return cb
}

// Submit implements GatewayRepository interface with circuit breaker protection
func (cb *CircuitBreaker) Submit(ctx context.Context, idempotencyKey string, request entities.TransferRequest) (string, error) {
	log.Printf("[INFO] CircuitBreaker.Submit - idempotency_key=%s, state=%s", idempotencyKey, cb.GetState())

	if !cb.allowRequest() {
		log.Printf("[WARN] CircuitBreaker.Submit - circuit open, rejecting request: idempotency_key=%s", idempotencyKey)
		return "", repositories.ErrCircuitOpen
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, cb.timeout)
	defer cancel()

	log.Printf("[INFO] CircuitBreaker.Submit - calling gateway: idempotency_key=%s", idempotencyKey)
	result, err := cb.gateway.Submit(ctx, idempotencyKey, request)
	cb.recordResult(err)

	if err != nil {
		log.Printf("[WARN] CircuitBreaker.Submit - gateway call failed: idempotency_key=%s, error=%v, state=%s", idempotencyKey, err, cb.GetState())
	} else {
		log.Printf("[INFO] CircuitBreaker.Submit - gateway call succeeded: idempotency_key=%s, gateway_ref=%s, state=%s", idempotencyKey, result, cb.GetState())
	}

	return result, err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if cooldown period has elapsed
		if time.Since(cb.lastFailureTime) > cb.cooldownDuration {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.halfOpenSuccess = false
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		// Allow exactly one probe call
		if !cb.halfOpenSuccess {
			return true
		}
		return false
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Deterministic rejections don't count as failures
	if errors.Is(err, repositories.ErrGatewayRejected) {
		log.Printf("[INFO] CircuitBreaker.recordResult - deterministic rejection, not counting as failure")
		return
	}

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
			log.Printf("[WARN] CircuitBreaker.recordResult - circuit opened: failure_count=%d, threshold=%d", cb.failureCount, cb.failureThreshold)
		} else {
			log.Printf("[WARN] CircuitBreaker.recordResult - failure recorded: failure_count=%d, threshold=%d", cb.failureCount, cb.failureThreshold)
		}
	} else {
		// Success
		if cb.state == StateHalfOpen {
			cb.halfOpenSuccess = true
			cb.state = StateClosed
			log.Printf("[INFO] CircuitBreaker.recordResult - circuit closed after successful probe")
		}
		cb.failureCount = 0
		log.Printf("[INFO] CircuitBreaker.recordResult - success recorded, failure count reset")
	}
}

// GetState returns the current circuit breaker state (for testing)
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state (for testing)
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.halfOpenSuccess = false
}
