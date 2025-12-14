package resilience

import (
	"context"
)

// Fallback provides a fallback mechanism for resilience
type Fallback struct {
	primaryFn   func(ctx context.Context) (interface{}, error)
	fallbackFn  func(ctx context.Context, err error) (interface{}, error)
	shouldRetry func(err error) bool
}

// FallbackOption is a functional option for configuring the Fallback
type FallbackOption func(*Fallback)

// WithRetryCondition sets the condition for when to retry
func WithRetryCondition(shouldRetry func(err error) bool) FallbackOption {
	return func(f *Fallback) {
		f.shouldRetry = shouldRetry
	}
}

// NewFallback creates a new fallback mechanism
func NewFallback(
	primaryFn func(ctx context.Context) (interface{}, error),
	fallbackFn func(ctx context.Context, err error) (interface{}, error),
	opts ...FallbackOption,
) *Fallback {
	f := &Fallback{
		primaryFn:  primaryFn,
		fallbackFn: fallbackFn,
		shouldRetry: func(err error) bool {
			return err != nil
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Execute executes the primary function and falls back to the fallback function if needed
func (f *Fallback) Execute(ctx context.Context) (interface{}, error) {
	result, err := f.primaryFn(ctx)
	if err == nil || !f.shouldRetry(err) {
		return result, err
	}

	// Primary function failed, use fallback
	return f.fallbackFn(ctx, err)
}
