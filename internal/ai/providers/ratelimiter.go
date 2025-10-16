package providers

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter wraps rate.Limiter for AI provider rate limiting
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a new rate limiter
// rateLimit is requests per minute (0 = unlimited)
func NewRateLimiter(rateLimit int) *RateLimiter {
	if rateLimit <= 0 {
		// Unlimited rate
		return &RateLimiter{
			limiter: rate.NewLimiter(rate.Inf, 1),
		}
	}

	// Convert requests per minute to requests per second
	rps := float64(rateLimit) / 60.0
	burst := 1
	if rateLimit > 60 {
		burst = rateLimit / 60
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Wait blocks until the rate limiter allows an action
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}
