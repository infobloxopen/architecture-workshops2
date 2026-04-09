package cases

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal operation — requests pass through
	StateOpen                         // Failing fast — requests rejected immediately
	StateHalfOpen                     // Probing — one request allowed to test recovery
)

func (s CircuitState) String() string {
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

// Breaker is a simple circuit breaker.
// LAB: STEP6 TODO — This breaker is disabled by default (threshold=0).
// Without it, every request attempts a DB call even when the DB is down,
// causing connection timeouts that block all concurrency slots.
// Fix: set Threshold to 5 and Timeout to 5*time.Second.
type Breaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	Threshold    int           // Consecutive failures before opening. 0 = disabled.
	Timeout      time.Duration // How long to stay open before half-opening.
	lastFailedAt time.Time
}

// Allow checks if a request should be allowed through.
// Returns true if the request can proceed, false if the breaker is open.
func (b *Breaker) Allow() bool {
	if b.Threshold <= 0 {
		return true // Breaker disabled
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastFailedAt) > b.Timeout {
			b.state = StateHalfOpen
			return true // Allow one probe request
		}
		return false
	case StateHalfOpen:
		return false // Already probing, reject others
	}
	return true
}

// RecordSuccess records a successful request. Resets the breaker to closed.
func (b *Breaker) RecordSuccess() {
	if b.Threshold <= 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures = 0
	b.state = StateClosed
}

// RecordFailure records a failed request. Opens the breaker after threshold.
func (b *Breaker) RecordFailure() {
	if b.Threshold <= 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures++
	b.lastFailedAt = time.Now()
	if b.failures >= b.Threshold {
		b.state = StateOpen
	}
}

// State returns the current breaker state.
func (b *Breaker) State() CircuitState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Failures returns the current consecutive failure count.
func (b *Breaker) Failures() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.failures
}
