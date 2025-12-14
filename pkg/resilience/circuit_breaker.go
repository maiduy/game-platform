package resilience

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerConfig holds the configuration for the circuit breaker
type CircuitBreakerConfig struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   func(counts gobreaker.Counts) bool
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// DefaultCircuitBreakerConfig returns a default configuration for the circuit breaker
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 100,
		Interval:    5 * time.Minute,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker %s state change from %s to %s\n", name, from, to)
		},
	}
}

// CircuitBreaker wraps the Sony gobreaker implementation
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	cbSettings := gobreaker.Settings{
		Name:          config.Name,
		MaxRequests:   config.MaxRequests,
		Interval:      config.Interval,
		Timeout:       config.Timeout,
		ReadyToTrip:   config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	}

	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

// Execute executes the given function if the circuit breaker is closed or half-open
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return cb.cb.Execute(func() (interface{}, error) {
		// Check if context is done before executing
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return fn()
	})
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() gobreaker.State {
	return cb.cb.State()
}

// Counts returns the current counts of the circuit breaker
func (cb *CircuitBreaker) Counts() gobreaker.Counts {
	return cb.cb.Counts()
}
