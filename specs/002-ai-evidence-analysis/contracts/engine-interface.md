# API Contract: AI Engine Interface

**Date**: 2025-10-11  
**Package**: `internal/ai`

## Interface Definition

```go
package ai

import (
    "context"
    "time"
)

// Engine is the core abstraction for AI provider integrations.
// Implementations must support OpenAI and Anthropic initially.
type Engine interface {
    // Analyze sends an analysis request to the AI provider and returns
    // the structured response. Returns error if provider fails, times out,
    // or returns invalid JSON.
    //
    // Context cancellation triggers immediate abort (no retry).
    // Timeout specified in ctx or falls back to AIConfig.Timeout.
    Analyze(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error)
    
    // Provider returns the provider identifier ("openai" | "anthropic" | "mock").
    Provider() string
    
    // Health checks if the provider is reachable and configured correctly.
    // Returns error if API key invalid, quota exceeded, or network unreachable.
    Health(ctx context.Context) error
}
```

---

## Contract: `Analyze` Method

### Input: `AnalysisRequest`

**Required Fields**:
```go
{
    "request_id": "550e8400-e29b-41d4-a716-446655440000",  // UUID v4
    "control_id": "SOC2-CC1.1",                            // Must exist in frameworks
    "control_name": "Access Control Policy",
    "framework": "SOC2",                                   // "SOC2" | "ISO27001" | "PCI-DSS"
    "policy_excerpt": "The organization has...",           // 200-500 words
    "events": [                                            // Length >= 1
        {
            "event_id": "evt-123",
            "event_type": "commit",
            "source": "git",
            "description": "Updated access control config",
            "content": "User <EMAIL_REDACTED> modified auth.yml", // Redacted
            "timestamp": "2025-10-10T14:30:00Z"
        }
    ],
    "timestamp": "2025-10-11T10:00:00Z",
    "cache_key": "a3f5d9..." // SHA256 hash
}
```

**Validation Rules**:
- `request_id` MUST be valid UUID
- `control_id` MUST match existing control in frameworks
- `events` MUST have length > 0
- `policy_excerpt` MUST be non-empty
- `events[].content` MUST be sanitized (no PII/secrets)
- `cache_key` MUST be 64-character hex SHA256

**Error Conditions**:
- Returns `ErrInvalidRequest` if validation fails
- Returns `ErrZeroEvents` if events slice is empty

### Output: `AnalysisResponse`

**Success Response**:
```go
{
    "request_id": "550e8400-e29b-41d4-a716-446655440000",  // Matches request
    "evidence_links": ["evt-123"],                         // Event IDs from request
    "justification": "Event shows access control update",  // 50-500 chars
    "confidence": 85,                                      // 0-100 integer
    "residual_risk": "No periodic review process",         // Optional, 0-500 chars
    "provider": "openai",                                  // "openai" | "anthropic"
    "model": "gpt-4",                                      // Actual model used
    "tokens_used": 1234,                                   // Total tokens consumed
    "latency": 2500,                                       // Milliseconds
    "timestamp": "2025-10-11T10:00:03Z",                   // Response time
    "cache_hit": false                                     // True if from cache
}
```

**Validation Rules**:
- `request_id` MUST match input
- `evidence_links` SHOULD reference valid event IDs (log warning if not)
- `justification` MUST be non-empty
- `confidence` MUST be 0 <= confidence <= 100
- `residual_risk` MAY be empty string
- `provider` MUST match engine's Provider() method
- `latency` MUST be > 0
- `cache_hit` MUST be false for Analyze() call (true only from cache)

**Error Conditions**:
- Returns `ErrProviderTimeout` if request exceeds 60 seconds
- Returns `ErrProviderRateLimit` if quota exceeded
- Returns `ErrProviderAuth` if API key invalid or missing
- Returns `ErrInvalidJSON` if response not parseable
- Returns `ErrProviderUnavailable` if provider returns 5xx errors

### Context Handling

**Timeout Behavior**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

resp, err := engine.Analyze(ctx, req)
if errors.Is(err, context.DeadlineExceeded) {
    // Fallback to heuristics
}
```

**Cancellation Behavior**:
```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    <-interrupt // User pressed Ctrl+C
    cancel()
}()

resp, err := engine.Analyze(ctx, req)
if errors.Is(err, context.Canceled) {
    // User aborted, stop processing
}
```

---

## Contract: `Provider` Method

### Output

**Returns**: String identifier for the provider

**Valid Values**:
- `"openai"` - OpenAI GPT models
- `"anthropic"` - Anthropic Claude models
- `"mock"` - Test mock provider

**Example**:
```go
engine := ai.NewOpenAIEngine(config)
fmt.Println(engine.Provider()) // Output: "openai"
```

---

## Contract: `Health` Method

### Purpose

Check if AI provider is reachable and correctly configured before starting analysis batch.

### Input

**Context**: Standard context with optional timeout

### Output

**Success**: Returns `nil` if provider is healthy

**Error Conditions**:
- Returns `ErrProviderAuth` if API key invalid or missing
- Returns `ErrProviderUnavailable` if provider unreachable (network issue, 5xx)
- Returns `ErrProviderQuotaExceeded` if quota exhausted
- Returns `context.DeadlineExceeded` if health check times out

**Example**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := engine.Health(ctx); err != nil {
    log.Warn("AI provider unhealthy, using heuristics", "error", err)
    // Fall back to heuristic-only analysis
}
```

---

## Error Types

```go
package ai

import "errors"

var (
    // Request validation errors
    ErrInvalidRequest  = errors.New("ai: invalid analysis request")
    ErrZeroEvents      = errors.New("ai: no events to analyze")
    
    // Provider errors (retryable with backoff)
    ErrProviderTimeout      = errors.New("ai: provider request timeout")
    ErrProviderRateLimit    = errors.New("ai: provider rate limit exceeded")
    ErrProviderUnavailable  = errors.New("ai: provider unavailable (5xx)")
    
    // Provider errors (non-retryable, fail fast)
    ErrProviderAuth         = errors.New("ai: provider authentication failed")
    ErrInvalidJSON          = errors.New("ai: provider returned invalid JSON")
    ErrProviderQuotaExceeded = errors.New("ai: provider quota exhausted")
)
```

---

## Retry Policy (Implementation Detail)

**Retryable Errors**: Use exponential backoff with jitter
- `ErrProviderTimeout`
- `ErrProviderRateLimit`
- `ErrProviderUnavailable`

**Max Retries**: 3 attempts
**Initial Delay**: 1 second
**Max Delay**: 30 seconds
**Multiplier**: 2x per retry

**Non-Retryable Errors**: Fail immediately, fallback to heuristics
- `ErrProviderAuth`
- `ErrInvalidJSON`
- `ErrProviderQuotaExceeded`
- `ErrInvalidRequest`

---

## Test Contract

All implementations of `Engine` must pass these contract tests:

```go
func TestEngine_AnalyzeSuccess(t *testing.T) {
    // Given valid request with events
    // When Analyze() called
    // Then returns AnalysisResponse with all required fields
    // And confidence is 0-100
    // And justification is non-empty
}

func TestEngine_AnalyzeTimeout(t *testing.T) {
    // Given context with 1ms timeout
    // When Analyze() called
    // Then returns ErrProviderTimeout or context.DeadlineExceeded
}

func TestEngine_AnalyzeInvalidRequest(t *testing.T) {
    // Given request with empty events slice
    // When Analyze() called
    // Then returns ErrZeroEvents or ErrInvalidRequest
}

func TestEngine_AnalyzeAuthFailure(t *testing.T) {
    // Given engine with invalid API key
    // When Analyze() called
    // Then returns ErrProviderAuth
}

func TestEngine_ProviderReturnsName(t *testing.T) {
    // When Provider() called
    // Then returns non-empty string matching provider type
}

func TestEngine_HealthSuccess(t *testing.T) {
    // Given healthy provider with valid config
    // When Health() called
    // Then returns nil
}

func TestEngine_HealthAuthFailure(t *testing.T) {
    // Given invalid API key
    // When Health() called
    // Then returns ErrProviderAuth
}
```

---

## Thread Safety

**Requirement**: All `Engine` implementations MUST be safe for concurrent use.

**Rationale**: Analysis command may process multiple controls in parallel (future optimization).

**Example**:
```go
var wg sync.WaitGroup
for _, control := range controls {
    wg.Add(1)
    go func(c Control) {
        defer wg.Done()
        resp, err := engine.Analyze(ctx, buildRequest(c))
        // Handle response
    }(control)
}
wg.Wait()
```

---

## Provider-Specific Notes

### OpenAI Implementation

- Uses function calling for structured JSON output
- Model default: `gpt-4-turbo-preview` (configurable via `ai.model`)
- Temperature: 0.3 (low for deterministic analysis)
- Max tokens: 4096 (configurable via `ai.max_tokens`)

### Anthropic Implementation

- Uses tool use for structured JSON output
- Model default: `claude-3-opus-20240229` (configurable via `ai.model`)
- Temperature: 0.3 (same as OpenAI for consistency)
- Max tokens: 4096 (configurable via `ai.max_tokens`)

### Mock Implementation (for tests)

- Returns hardcoded responses from golden files
- No network calls, instant responses
- Configurable success/failure modes for error path testing
