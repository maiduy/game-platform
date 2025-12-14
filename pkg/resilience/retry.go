package resilience

import (
	"context"
	"time"

	"github.com/eapache/go-resiliency/retrier"
)

// RetryConfig holds the configuration for the retry mechanism
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// DefaultRetryConfig returns a default configuration for the retry mechanism
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
	}
}

// Retry wraps the eapache/go-resiliency retrier implementation
type Retry struct {
	retrier *retrier.Retrier
	config  RetryConfig
}

// NewRetry creates a new retry mechanism with the given configuration
func NewRetry(config RetryConfig) *Retry {
	// Create a backoff strategy with the config values
	backoffTimes := retrier.ExponentialBackoff(config.MaxRetries, config.InitialBackoff)

	return &Retry{
		retrier: retrier.New(backoffTimes, nil), // nil uses the default classifier
		config:  config,
	}
}

// Execute executes the given function with retry logic
func (r *Retry) Execute(ctx context.Context, fn func() error) error {
	var err error
	err = r.retrier.Run(func() error {
		// Check if context is done before executing
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fn()
	})
	return err
}

// ExecuteWithValue executes the given function with retry logic and returns a value
func (r *Retry) ExecuteWithValue(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	var err error

	err = r.retrier.Run(func() error {
		// Check if context is done before executing
		if ctx.Err() != nil {
			return ctx.Err()
		}

		result, err = fn()
		return err
	})

	return result, err
}
