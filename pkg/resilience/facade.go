package resilience

import (
	"context"
	"fmt"
)

// ResilienceFacade combines circuit breaker, retry and fallback patterns
type ResilienceFacade struct {
	name           string
	circuitBreaker *CircuitBreaker
	retry          *Retry
	fallbackFn     func(ctx context.Context, err error) (interface{}, error)
}

// ResilienceFacadeOption is a functional option for configuring the ResilienceFacade
type ResilienceFacadeOption func(*ResilienceFacade)

// WithCircuitBreaker adds a circuit breaker to the facade
func WithCircuitBreaker(cb *CircuitBreaker) ResilienceFacadeOption {
	return func(rf *ResilienceFacade) {
		rf.circuitBreaker = cb
	}
}

// WithRetry adds a retry mechanism to the facade
func WithRetry(retry *Retry) ResilienceFacadeOption {
	return func(rf *ResilienceFacade) {
		rf.retry = retry
	}
}

// WithFallback adds a fallback function to the facade
func WithFallback(fallbackFn func(ctx context.Context, err error) (interface{}, error)) ResilienceFacadeOption {
	return func(rf *ResilienceFacade) {
		rf.fallbackFn = fallbackFn
	}
}

// NewResilienceFacade creates a new resilience facade with the given options
func NewResilienceFacade(name string, opts ...ResilienceFacadeOption) *ResilienceFacade {
	rf := &ResilienceFacade{
		name: name,
		fallbackFn: func(ctx context.Context, err error) (interface{}, error) {
			// Default fallback returns the error
			return nil, err
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(rf)
	}

	return rf
}

// Execute executes the given function with the configured resilience patterns
func (rf *ResilienceFacade) Execute(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var executeWithCircuitBreaker = func() (interface{}, error) {
		if rf.retry != nil {
			var result interface{}
			var finalErr error

			// Execute with retry
			err := rf.retry.Execute(ctx, func() error {
				var err error
				result, err = fn(ctx)
				return err
			})

			if err != nil {
				finalErr = fmt.Errorf("retry failed: %w", err)
				// Use fallback if retry fails
				if rf.fallbackFn != nil {
					return rf.fallbackFn(ctx, finalErr)
				}
				return nil, finalErr
			}

			return result, nil
		}

		// No retry, just execute the function
		result, err := fn(ctx)
		if err != nil && rf.fallbackFn != nil {
			return rf.fallbackFn(ctx, err)
		}
		return result, err
	}

	// If circuit breaker is configured, wrap the execution with it
	if rf.circuitBreaker != nil {
		return rf.circuitBreaker.Execute(ctx, executeWithCircuitBreaker)
	}

	// No circuit breaker, just execute with retry and fallback
	return executeWithCircuitBreaker()
}
