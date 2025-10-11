# Data Model: AI Evidence Analysis

**Phase 1 Output** | **Date**: 2025-10-11

## Core Entities

### 1. AI Provider Configuration

**Purpose**: Configuration for AI provider selection and behavior

**Fields**:
```go
type AIConfig struct {
    Provider    string  // "openai" | "anthropic" | "none"
    Enabled     bool    // Master switch for AI analysis
    Model       string  // Model identifier (e.g., "gpt-4", "claude-3-opus")
    MaxTokens   int     // Token limit for requests (default: 4096)
    Temperature float32 // Randomness (0.0-1.0, default: 0.3)
    Timeout     int     // Request timeout in seconds (default: 60)
    RateLimit   int     // Max requests per minute (default: 10)
    
    // API credentials (from env vars or config)
    OpenAIKey    string `viper:"ai.openai_key"`
    AnthropicKey string `viper:"ai.anthropic_key"`
}
```

**Validation Rules**:
- Provider must be one of: "openai", "anthropic", "none"
- MaxTokens must be > 0 and <= 32768
- Temperature must be 0.0 <= temp <= 1.0
- Timeout must be > 0 and <= 300 seconds
- RateLimit must be > 0
- API key required if Enabled && Provider != "none"

**Relationships**:
- Used by `ai.Engine` to initialize provider adapters
- Loaded via Viper from config file + env vars + CLI flags

---

### 2. AI Analysis Request

**Purpose**: Input to AI provider for a specific control analysis

**Fields**:
```go
type AnalysisRequest struct {
    RequestID   string    // Unique request identifier (UUID)
    ControlID   string    // Compliance control identifier (e.g., "SOC2-CC1.1")
    ControlName string    // Human-readable control name
    Framework   string    // "SOC2" | "ISO27001" | "PCI-DSS"
    
    // Policy context
    PolicyExcerpt string  // Relevant policy text (200-500 words)
    
    // Events to analyze (normalized, redacted)
    Events []AnalysisEvent
    
    // Metadata
    Timestamp   time.Time // Request creation time
    CacheKey    string    // SHA256 hash for cache lookup
}

type AnalysisEvent struct {
    EventID     string    // Original event UUID
    EventType   string    // "commit" | "build" | "ticket" | "message" | "doc"
    Source      string    // "git" | "cicd" | "jira" | "slack" | "docs"
    Description string    // Brief summary (max 200 chars)
    Content     string    // Redacted event content (max 1000 chars)
    Timestamp   time.Time // Event occurrence time
}
```

**Validation Rules**:
- RequestID must be valid UUID
- ControlID must match existing control in frameworks
- Events slice must have length > 0 (zero events skip AI)
- PolicyExcerpt must be non-empty
- Event content must be sanitized (no PII/secrets)

**Relationships**:
- Created by `internal/analyze/mapper.go` from state data
- Consumed by `ai.Engine` interface implementations
- Cached using CacheKey for lookup

---

### 3. AI Analysis Response

**Purpose**: Structured output from AI provider

**Fields**:
```go
type AnalysisResponse struct {
    RequestID   string    // Matches AnalysisRequest.RequestID
    
    // AI-generated fields (from JSON schema)
    EvidenceLinks   []string // Event IDs that support the control
    Justification   string   // Explanation of relevance (50-500 chars)
    Confidence      int      // 0-100 confidence score
    ResidualRisk    string   // Optional notes on gaps (0-500 chars)
    
    // Metadata
    Provider        string    // "openai" | "anthropic"
    Model           string    // Actual model used
    TokensUsed      int       // Total tokens consumed
    Latency         int       // Response time in milliseconds
    Timestamp       time.Time // Response received time
    CacheHit        bool      // True if served from cache
}
```

**Validation Rules**:
- EvidenceLinks must reference valid event IDs from request
- Justification must be non-empty string
- Confidence must be 0 <= confidence <= 100
- ResidualRisk can be empty string
- Provider must match request provider
- Minimal type validation (trust AI output per clarification Q4)

**Relationships**:
- Returned by `ai.Engine` interface implementations
- Cached in `ai.Cache` with request CacheKey
- Consumed by `internal/analyze/mapper.go` to enhance evidence

---

### 4. Cached AI Result

**Purpose**: Persisted AI response for cache reuse

**Fields**:
```go
type CachedResult struct {
    CacheKey    string           // SHA256 hash of request inputs
    Response    AnalysisResponse // Stored AI response
    
    // Cache metadata
    CachedAt    time.Time        // Cache entry creation time
    EventIDs    []string         // Event IDs for invalidation tracking
    ControlID   string           // Control ID for invalidation tracking
    Provider    string           // AI provider used
    ModelVersion string          // Model version for compatibility
}
```

**Validation Rules**:
- CacheKey must be valid SHA256 hash (64 hex chars)
- EventIDs must be non-empty
- CachedAt must be <= current time

**Relationships**:
- Stored in filesystem: `~/.cache/sdek/ai-cache/{cache_key}.json`
- Invalidated by `ai.Cache` when events change
- Loaded by `ai.Cache` before AI API call

**Lifecycle**:
1. **Create**: After successful AI response
2. **Read**: Before AI API call (cache hit)
3. **Invalidate**: When any referenced event is added/modified/deleted
4. **Delete**: Manual cleanup or LRU eviction (future)

---

### 5. Enhanced Evidence (Extended Type)

**Purpose**: Original evidence enhanced with AI metadata

**Fields**:
```go
// Extends existing types.Evidence struct
type Evidence struct {
    // ... existing fields (ID, SourceID, EventID, ControlID, etc.)
    
    // NEW: AI analysis metadata
    AIAnalyzed      bool      // True if AI was used for this evidence
    AIJustification string    // AI-generated explanation
    AIConfidence    int       // AI confidence score (0-100)
    AIResidualRisk  string    // AI risk notes
    
    // NEW: Hybrid confidence scoring
    HeuristicConfidence int   // Original keyword-based score (0-100)
    CombinedConfidence  int   // Weighted average (70% AI + 30% heuristic)
    
    // NEW: Analysis method indicator
    AnalysisMethod  string    // "ai+heuristic" | "heuristic-only" | "no-ai"
}
```

**Validation Rules**:
- If AIAnalyzed = true, AIJustification must be non-empty
- AIConfidence, HeuristicConfidence must be 0-100 or -1 (not applicable)
- CombinedConfidence = (0.7 * AIConfidence) + (0.3 * HeuristicConfidence)
- AnalysisMethod must match AIAnalyzed state

**Relationships**:
- Created by `internal/analyze/mapper.go` after AI analysis
- Persisted in state files (JSON)
- Consumed by `internal/report/exporter.go` for report generation

---

### 6. Privacy Filter

**Purpose**: PII and secret detection/redaction before AI transmission

**Fields**:
```go
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
    
    // Statistics
    RedactionCount map[string]int // Pattern name -> count of redactions
}

type RedactionResult struct {
    Original    string            // Original text
    Redacted    string            // Text with redactions
    Redactions  []RedactionInfo   // Details of what was redacted
}

type RedactionInfo struct {
    PatternName string // "email" | "api_key" | "phone" | etc.
    Position    int    // Character offset in original text
    Length      int    // Length of redacted text
    Replacement string // Placeholder used (e.g., "<EMAIL_REDACTED>")
}
```

**Validation Rules**:
- All patterns must be valid regex (compile at init)
- AllowedFields must not contain sensitive field names
- RedactionCount must be thread-safe (use sync.Map)

**Relationships**:
- Used by `internal/ai/privacy.go` to sanitize events
- Applied before creating `AnalysisRequest`
- Statistics logged at end of analysis run

**Behavior**:
- Scan event content with all patterns
- Replace matches with placeholders
- Track redaction statistics
- Preserve text structure (e.g., "User <EMAIL> created ticket")

---

## Entity Relationships Diagram

```
┌─────────────────┐
│   AIConfig      │
│  (from Viper)   │
└────────┬────────┘
         │ configures
         ▼
┌─────────────────┐        creates        ┌──────────────────┐
│   ai.Engine     │───────────────────────▶│ AnalysisRequest  │
│   (interface)   │                        │  + Events        │
└────────┬────────┘                        └────────┬─────────┘
         │                                          │
         │ implements                               │ sanitized by
         │                                          │
    ┌────┴─────┐                            ┌──────▼──────────┐
    │          │                            │ PrivacyFilter   │
┌───▼──┐   ┌──▼────┐                        └─────────────────┘
│OpenAI│   │Anthrop│                                │
│Adaptr│   │icAdapt│                                │ redacts
└───┬──┘   └──┬────┘                                │
    │         │                                     │
    └────┬────┘                                     │
         │ returns                                  │
         ▼                                          │
┌─────────────────┐        cached in        ┌──────▼──────────┐
│ AnalysisResponse│───────────────────────▶│  CachedResult   │
│  (AI output)    │                        │ (filesystem)    │
└────────┬────────┘                        └─────────────────┘
         │                                          │
         │ enhances                                 │ invalidated by
         ▼                                          │
┌─────────────────┐                         ┌──────▼──────────┐
│ Enhanced        │                         │  Event Changes  │
│ Evidence        │                         │  (add/mod/del)  │
│ (with AI fields)│                         └─────────────────┘
└────────┬────────┘
         │ exported by
         ▼
┌─────────────────┐
│  Report (JSON)  │
│  + AI metadata  │
└─────────────────┘
```

---

## State Transitions

### Cache Lifecycle

```
[No Cache Entry]
      │
      │ AI request succeeds
      ▼
[Cached Result] ─────────────────┐
      │                          │
      │ Cache hit                │ Event added/modified/deleted
      │ (reuse)                  │
      ▼                          │
[Analysis Response]              │
                                 │
                                 ▼
                          [Invalidated]
                                 │
                                 │ Cleanup
                                 ▼
                          [Deleted from disk]
```

### AI Analysis Flow

```
[Control + Events]
      │
      │ Check cache
      ▼
┌─────────────┐
│ Cache Hit?  │
└─────┬───┬───┘
      │   │
  Yes │   │ No
      │   │
      │   │ Check AI config
      │   ▼
      │ ┌─────────────┐
      │ │ AI Enabled? │
      │ └─────┬───┬───┘
      │       │   │
      │   Yes │   │ No
      │       │   │
      │       │   │ Fallback
      │       │   ▼
      │       │ [Heuristic-Only Evidence]
      │       │
      │       │ Redact PII
      │       ▼
      │     [AI Request]
      │       │
      │       │ 60s timeout
      │       ▼
      │     ┌─────────────┐
      │     │  Success?   │
      │     └─────┬───┬───┘
      │           │   │
      │       Yes │   │ No (error/timeout)
      │           │   │
      │           │   │ Fallback
      │           │   ▼
      │           │ [Heuristic-Only Evidence]
      │           │
      │           │ Cache response
      │           ▼
      │         [AI Response]
      │           │
      └───────────┴────────┐
                           │ Combine scores (70% AI + 30% heuristic)
                           ▼
                    [Enhanced Evidence]
```

---

## Validation Strategy

### Input Validation (Before AI Call)
1. ControlID exists in frameworks ✓
2. PolicyExcerpt non-empty ✓
3. Events list non-empty ✓
4. All event content redacted (PII/secrets removed) ✓
5. Total token count < MaxTokens ✓

### Output Validation (After AI Call)
1. Response is valid JSON ✓
2. Required fields present (evidence_links, justification, confidence, residual_risk) ✓
3. Confidence is integer 0-100 ✓ (lenient: coerce if float)
4. EvidenceLinks reference valid event IDs ⚠️ (log warning, don't fail)
5. Justification non-empty ✓

### Cache Validation (On Load)
1. CacheKey is valid SHA256 ✓
2. CachedAt <= current time ✓
3. EventIDs all exist in current state ⚠️ (invalidate if mismatch)
4. Provider matches current config ⚠️ (invalidate if mismatch)

---

## Next Steps (Phase 2)

1. Generate contract tests for `ai.Engine` interface
2. Create mock implementations for testing
3. Define API contracts in OpenAPI format (optional, for documentation)
4. Extract integration test scenarios from feature spec
5. Update quickstart.md with AI-enabled analysis examples
