package retry

import (
	"fmt"
	"math"
	"time"
)

// Config contains retry configuration
type Config struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// DefaultConfig returns default retry configuration
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Do executes a function with exponential backoff retry
func Do(config *Config, fn func() error) error {
	if config == nil {
		config = DefaultConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Check if error is retryable
		if !isRetryable(err, config.RetryableErrors) {
			return err
		}

		// Wait before retry
		time.Sleep(delay)

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.BackoffFactor)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// DoWithCallback executes a function with retry and callback
func DoWithCallback(config *Config, fn func() error, onRetry func(attempt int, err error)) error {
	if config == nil {
		config = DefaultConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt == config.MaxAttempts {
			break
		}

		if !isRetryable(err, config.RetryableErrors) {
			return err
		}

		if onRetry != nil {
			onRetry(attempt, err)
		}

		time.Sleep(delay)
		delay = time.Duration(math.Min(float64(delay)*config.BackoffFactor, float64(config.MaxDelay)))
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// isRetryable checks if an error is retryable
func isRetryable(err error, retryableErrors []string) bool {
	if len(retryableErrors) == 0 {
		// By default, retry on all errors
		return true
	}

	errStr := err.Error()
	for _, retryableErr := range retryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)*2 && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)/2+len(substr)%2] == substr))
}
