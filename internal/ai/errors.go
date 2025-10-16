package ai

import "errors"

// Request validation errors
var (
	// ErrInvalidRequest indicates the analysis request failed validation
	ErrInvalidRequest = errors.New("ai: invalid analysis request")

	// ErrZeroEvents indicates no events were provided for analysis
	ErrZeroEvents = errors.New("ai: no events to analyze")
)

// Provider errors (retryable with backoff)
var (
	// ErrProviderTimeout indicates the provider request exceeded the timeout
	ErrProviderTimeout = errors.New("ai: provider request timeout")

	// ErrProviderRateLimit indicates the provider rate limit was exceeded
	ErrProviderRateLimit = errors.New("ai: provider rate limit exceeded")

	// ErrProviderUnavailable indicates the provider returned a 5xx error
	ErrProviderUnavailable = errors.New("ai: provider unavailable (5xx)")
)

// Provider errors (non-retryable, fail fast)
var (
	// ErrProviderAuth indicates authentication with the provider failed
	ErrProviderAuth = errors.New("ai: provider authentication failed")

	// ErrInvalidJSON indicates the provider returned invalid JSON
	ErrInvalidJSON = errors.New("ai: provider returned invalid JSON")

	// ErrProviderQuotaExceeded indicates the provider quota was exhausted
	ErrProviderQuotaExceeded = errors.New("ai: provider quota exhausted")
)

// IsRetryable returns true if the error should be retried with backoff
func IsRetryable(err error) bool {
	return errors.Is(err, ErrProviderTimeout) ||
		errors.Is(err, ErrProviderRateLimit) ||
		errors.Is(err, ErrProviderUnavailable)
}

// IsFatalError returns true if the error should not be retried
func IsFatalError(err error) bool {
	return errors.Is(err, ErrProviderAuth) ||
		errors.Is(err, ErrInvalidJSON) ||
		errors.Is(err, ErrProviderQuotaExceeded) ||
		errors.Is(err, ErrInvalidRequest) ||
		errors.Is(err, ErrZeroEvents)
}
