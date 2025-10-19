package ai

import (
	"regexp"
	"sync"
	"time"
)

// AIConfig represents configuration for AI provider selection and behavior
type AIConfig struct {
	Provider    string  // "openai" | "anthropic" | "none"
	Enabled     bool    // Master switch for AI analysis
	Model       string  // Model identifier (e.g., "gpt-4", "claude-3-opus")
	MaxTokens   int     // Token limit for requests (default: 4096)
	Temperature float32 // Randomness (0.0-1.0, default: 0.3)
	Timeout     int     // Request timeout in seconds (default: 60)
	RateLimit   int     // Max requests per minute (default: 10)

	// API credentials (from env vars or config)
	OpenAIKey    string `mapstructure:"openai_key"`
	AnthropicKey string `mapstructure:"anthropic_key"`
}

// AnalysisRequest represents input to AI provider for a specific control analysis
type AnalysisRequest struct {
	RequestID   string // Unique request identifier (UUID)
	ControlID   string // Compliance control identifier (e.g., "SOC2-CC1.1")
	ControlName string // Human-readable control name
	Framework   string // "SOC2" | "ISO27001" | "PCI-DSS"

	// Policy context
	PolicyExcerpt string // Relevant policy text (200-500 words)

	// Events to analyze (normalized, redacted)
	Events []AnalysisEvent

	// Metadata
	Timestamp time.Time // Request creation time
	CacheKey  string    // SHA256 hash for cache lookup
}

// AnalysisEvent represents a single event to be analyzed
type AnalysisEvent struct {
	EventID     string    // Original event UUID
	EventType   string    // "commit" | "build" | "ticket" | "message" | "doc"
	Source      string    // "git" | "cicd" | "jira" | "slack" | "docs"
	Description string    // Brief summary (max 200 chars)
	Content     string    // Redacted event content (max 1000 chars)
	Timestamp   time.Time // Event occurrence time
}

// AnalysisResponse represents structured output from AI provider
type AnalysisResponse struct {
	RequestID string // Matches AnalysisRequest.RequestID

	// AI-generated fields (from JSON schema)
	EvidenceLinks []string // Event IDs that support the control
	Justification string   // Explanation of relevance (50-500 chars)
	Confidence    int      // 0-100 confidence score
	ResidualRisk  string   // Optional notes on gaps (0-500 chars)

	// Metadata
	Provider   string    // "openai" | "anthropic"
	Model      string    // Actual model used
	TokensUsed int       // Total tokens consumed
	Latency    int       // Response time in milliseconds
	Timestamp  time.Time // Response received time
	CacheHit   bool      // True if served from cache
}

// CachedResult represents persisted AI response for cache reuse
type CachedResult struct {
	CacheKey string           // SHA256 hash of request inputs
	Response AnalysisResponse // Stored AI response

	// Cache metadata
	CachedAt     time.Time // Cache entry creation time
	EventIDs     []string  // Event IDs for invalidation tracking
	ControlID    string    // Control ID for invalidation tracking
	Provider     string    // AI provider used
	ModelVersion string    // Model version for compatibility
}

// PrivacyFilter handles PII and secret detection/redaction before AI transmission
type PrivacyFilter struct {
	// Patterns (compiled regexes)
	EmailPattern      *regexp.Regexp
	PhonePattern      *regexp.Regexp
	APIKeyPattern     *regexp.Regexp
	CreditCardPattern *regexp.Regexp
	SSNPattern        *regexp.Regexp

	// Custom patterns (user-configurable)
	CustomPatterns []*regexp.Regexp

	// Allowlist (fields safe to send)
	AllowedFields []string // e.g., ["timestamp", "log_level", "status_code"]

	// Statistics (thread-safe)
	redactionCount sync.Map // Pattern name -> count of redactions
}

// RedactionResult represents the result of redacting text
type RedactionResult struct {
	Original   string          // Original text
	Redacted   string          // Text with redactions
	Redactions []RedactionInfo // Details of what was redacted
}

// RedactionInfo contains details about a single redaction
type RedactionInfo struct {
	PatternName string // "email" | "api_key" | "phone" | etc.
	Position    int    // Character offset in original text
	Length      int    // Length of redacted text
	Replacement string // Placeholder used (e.g., "<EMAIL_REDACTED>")
}

// GetStatistics returns the current redaction statistics
func (pf *PrivacyFilter) GetStatistics() map[string]int {
	stats := make(map[string]int)
	pf.redactionCount.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if v, ok := value.(int); ok {
				stats[k] = v
			}
		}
		return true
	})
	return stats
}

// IncrementRedactionCount increments the count for a specific pattern
func (pf *PrivacyFilter) IncrementRedactionCount(patternName string) {
	val, _ := pf.redactionCount.LoadOrStore(patternName, 0)
	count := val.(int)
	pf.redactionCount.Store(patternName, count+1)
}
