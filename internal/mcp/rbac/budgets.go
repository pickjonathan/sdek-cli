package rbac

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// BudgetManager enforces rate limits and concurrency limits for MCP tools.
type BudgetManager struct {
	rateLimiters map[string]*RateLimiter
	concLimiters map[string]*ConcurrencyLimiter
	mu           sync.RWMutex
}

// NewBudgetManager creates a new budget manager.
func NewBudgetManager() *BudgetManager {
	return &BudgetManager{
		rateLimiters: make(map[string]*RateLimiter),
		concLimiters: make(map[string]*ConcurrencyLimiter),
	}
}

// CheckBudget verifies if the tool invocation is within budget limits.
func (b *BudgetManager) CheckBudget(ctx context.Context, toolName string, budget *types.ToolBudget) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if budget.RateLimit.RequestsPerSecond > 0 {
		if _, exists := b.rateLimiters[toolName]; !exists {
			b.rateLimiters[toolName] = NewRateLimiter(int(budget.RateLimit.RequestsPerSecond))
		}

		if !b.rateLimiters[toolName].Allow() {
			return mcp.ErrRateLimited
		}
	}

	if budget.ConcurrencyLimit > 0 {
		if _, exists := b.concLimiters[toolName]; !exists {
			b.concLimiters[toolName] = NewConcurrencyLimiter(budget.ConcurrencyLimit)
		}

		if !b.concLimiters[toolName].Acquire() {
			return fmt.Errorf("concurrency limit exceeded for tool %s", toolName)
		}
	}

	return nil
}

// ReleaseConcurrency releases a concurrency slot for the tool.
func (b *BudgetManager) ReleaseConcurrency(toolName string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limiter, exists := b.concLimiters[toolName]; exists {
		limiter.Release()
	}
}

// RateLimiter implements token bucket rate limiting.
type RateLimiter struct {
	rate   float64
	tokens float64
	lastAt time.Time
	mu     sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the given requests per second.
func NewRateLimiter(requestsPerSec int) *RateLimiter {
	return &RateLimiter{
		rate:   float64(requestsPerSec),
		tokens: float64(requestsPerSec),
		lastAt: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastAt).Seconds()
	r.tokens += elapsed * r.rate

	if r.tokens > r.rate {
		r.tokens = r.rate
	}

	r.lastAt = now

	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		return true
	}

	return false
}

// ConcurrencyLimiter implements semaphore-based concurrency limiting.
type ConcurrencyLimiter struct {
	max     int
	current int
	mu      sync.Mutex
}

// NewConcurrencyLimiter creates a new concurrency limiter.
func NewConcurrencyLimiter(max int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		max:     max,
		current: 0,
	}
}

// Acquire attempts to acquire a concurrency slot.
func (c *ConcurrencyLimiter) Acquire() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current < c.max {
		c.current++
		return true
	}

	return false
}

// Release releases a concurrency slot.
func (c *ConcurrencyLimiter) Release() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current > 0 {
		c.current--
	}
}
