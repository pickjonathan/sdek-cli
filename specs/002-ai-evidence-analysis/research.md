# Research: AI Evidence Analysis

**Phase 0 Output** | **Date**: 2025-10-11

## Research Tasks

### 1. AI SDK Selection

**Task**: Research OpenAI and Anthropic Go SDKs for compliance use case

**Decision**: 
- **OpenAI**: Use `github.com/sashabaranov/go-openai` (community SDK, 10k+ stars, active maintenance)
- **Anthropic**: Use `github.com/anthropics/anthropic-sdk-go` (official SDK, released 2024)

**Rationale**:
- **go-openai** is the de facto standard for OpenAI in Go with excellent API coverage and test suite
- **anthropic-sdk-go** is the official SDK with first-class support and future compatibility
- Both support structured JSON outputs via function calling (OpenAI) and tool use (Anthropic)
- Both support streaming (not required initially but available for future enhancements)
- Both handle retries internally (can be supplemented with custom backoff strategies)

**Alternatives Considered**:
- **Direct HTTP calls**: Rejected - SDKs handle auth, retries, versioning, reducing boilerplate
- **go-gpt3** (older OpenAI SDK): Rejected - less maintained, superseded by go-openai
- **LangChain Go**: Rejected - over-engineered for this use case, adds unnecessary abstraction layers

---

### 2. Structured JSON Output Best Practices

**Task**: Research reliable methods for getting structured JSON from LLMs

**Decision**: 
- **OpenAI**: Use function calling with `Functions` parameter to define JSON schema
- **Anthropic**: Use tool use with `tools` parameter and JSON schema specification
- **Fallback**: Include JSON schema in system prompt with explicit formatting instructions

**Rationale**:
- Function calling/tool use produces more reliable JSON than prompt engineering alone
- Schemas define required fields (evidence_links, justification, confidence, residual_risk)
- Native validation by AI model before returning response (reduces parsing errors)
- Both providers support nested objects and arrays for complex evidence structures

**Best Practices**:
1. Define strict JSON schema with required fields and type constraints
2. Include examples in system prompt for format reinforcement
3. Set `temperature=0.3` (low) for more deterministic outputs
4. Use `max_tokens` to prevent truncation of JSON structure
5. Parse with lenient JSON decoder (accept trailing commas, comments if present)

**Alternatives Considered**:
- **Prompt engineering only**: Rejected - unreliable, high failure rate (~10-20%)
- **YAML output**: Rejected - JSON is standard for API contracts, easier to parse
- **XML output**: Rejected - verbose, not LLM-native format

---

### 3. PII/Secret Redaction Patterns

**Task**: Research effective PII detection and redaction strategies

**Decision**: Use **regex-based detection** with configurable patterns + **hash-based anonymization**

**Patterns to Detect**:
- **Email addresses**: `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`
- **Phone numbers**: `\b(\+?1?[-.\s]?)?(\(?\d{3}\)?[-.\s]?)?\d{3}[-.\s]?\d{4}\b`
- **API keys**: `\b[A-Za-z0-9_-]{32,}\b` (common for 256-bit keys in base64/hex)
- **AWS keys**: `AKIA[0-9A-Z]{16}`, secret: `[0-9a-zA-Z/+=]{40}`
- **GitHub tokens**: `ghp_[A-Za-z0-9]{36}`, `gho_[A-Za-z0-9]{36}`, `ghs_[A-Za-z0-9]{36}`
- **Private keys**: `-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----`
- **Passwords in URLs**: `://[^:]+:([^@]+)@` (extract credential portion)
- **Credit card numbers**: `\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b` (with Luhn validation)
- **SSNs**: `\b\d{3}-\d{2}-\d{4}\b`

**Redaction Strategy**:
- **Replace with placeholders**: `<EMAIL_REDACTED>`, `<API_KEY_REDACTED>`, `<PHONE_REDACTED>`
- **Preserve structure**: Keep domain for emails (`user@<REDACTED>`), last 4 digits for cards
- **Hash sensitive IDs**: Use SHA256 with salt for user IDs, ticket IDs (preserves uniqueness)
- **Configurable allowlist**: Some fields safe to send (timestamps, status codes, log levels)

**Rationale**:
- Regex-based detection has high recall for known patterns (catches 95%+ of common PII)
- Placeholder replacement maintains text structure for AI context (e.g., "User <EMAIL> created ticket")
- Hashing preserves uniqueness for correlation (same user ID → same hash)
- Allowlist reduces false positives (prevents redacting technical terms like "email service")

**Alternatives Considered**:
- **ML-based NER**: Rejected - overkill, requires model hosting, slower inference
- **Full text stripping**: Rejected - removes too much context, degrades AI analysis quality
- **No redaction**: Rejected - violates privacy requirements (FR-010, FR-011, NFR-005)

---

### 4. Caching Strategy

**Task**: Research cache invalidation strategies for event-driven systems

**Decision**: **Event-driven cache invalidation** with content-based keys

**Cache Key Design**:
```
cache_key = SHA256(
    control_id + 
    sorted_event_ids + 
    event_content_hash + 
    ai_provider + 
    ai_model_version
)
```

**Invalidation Triggers**:
1. **Event added**: Invalidate all cache entries referencing any control
2. **Event modified**: Invalidate entries with that event's ID in the key
3. **Event deleted**: Same as modified
4. **Control definition changed**: Invalidate entries for that control
5. **AI provider switch**: New cache namespace (provider name in key)

**Storage**:
- **Location**: `~/.cache/sdek/ai-cache/` (respects `XDG_CACHE_HOME` on Linux)
- **Format**: JSON files, one per cache entry: `{cache_key}.json`
- **Metadata**: Include `cached_at`, `ttl` (optional future), `event_ids`, `control_id`

**Rationale**:
- Content-based keys ensure cache correctness (same input → same output)
- Event IDs in key enable surgical invalidation (O(1) lookup)
- Filesystem cache is simple, portable, and user-inspectable
- Cache miss is safe (falls back to AI call or heuristics)

**Best Practices**:
1. Write cache atomically (`ioutil.WriteFile` with temp + rename)
2. Gracefully handle cache read failures (treat as miss, log warning)
3. Periodic cleanup of orphaned entries (future: LRU eviction policy)
4. Cache statistics logged at end of analysis (hits, misses, invalidations)

**Alternatives Considered**:
- **In-memory cache**: Rejected - lost on process exit, not useful for CLI runs
- **Database cache (SQLite)**: Rejected - overkill, adds dependency, harder to debug
- **TTL-based expiration**: Rejected - clarification resolved to event-driven invalidation
- **Redis/Memcached**: Rejected - external dependency, not suitable for CLI tool

---

### 5. Retry and Rate Limiting Strategies

**Task**: Research resilient API call patterns for AI providers

**Decision**: **Exponential backoff with jitter** + **circuit breaker** + **rate limiter**

**Exponential Backoff**:
- Library: `github.com/cenkalti/backoff/v4`
- Initial delay: 1 second
- Max delay: 30 seconds
- Multiplier: 2x per retry
- Jitter: ±25% randomization to avoid thundering herd
- Max retries: 3 attempts before fallback

**Circuit Breaker**:
- Pattern: Half-open after 5 consecutive failures
- Reset: After 60 seconds or 1 successful call
- State transitions: Closed → Open → Half-Open → Closed
- Behavior: Open state immediately returns error (fallback to heuristics)

**Rate Limiter**:
- Library: `golang.org/x/time/rate`
- Default: 10 requests per minute (configurable via `ai.rate_limit`)
- Per-provider limits (OpenAI vs Anthropic may differ)
- Burst: Allow 3 rapid requests, then throttle

**Rationale**:
- Exponential backoff handles transient failures (network glitches, temporary quota issues)
- Circuit breaker prevents cascading failures during provider outages
- Rate limiter respects provider quotas, prevents self-inflicted rate limit errors
- Jitter prevents synchronized retries across multiple controls

**Retryable Errors**:
- HTTP 429 (rate limit): Retry with backoff
- HTTP 500, 502, 503, 504 (server errors): Retry with backoff
- Network timeouts: Retry with backoff
- Connection refused: Retry once, then circuit break

**Non-Retryable Errors**:
- HTTP 400 (bad request): Validation error, fallback immediately
- HTTP 401, 403 (auth errors): Config error, fail fast with clear message
- HTTP 404 (not found): Model/endpoint error, fail fast
- Invalid JSON response: Parse error, fallback immediately

**Alternatives Considered**:
- **Fixed delay retry**: Rejected - inefficient, increases latency on transient failures
- **Unlimited retries**: Rejected - can cause indefinite hangs, violates 60s timeout
- **No rate limiting**: Rejected - risks provider account suspension
- **Token bucket**: Considered - rate.Limiter uses token bucket internally, sufficient

---

### 6. Prompt Engineering for Compliance Context

**Task**: Research effective prompt structures for compliance evidence analysis

**Decision**: **System + User prompt pattern** with JSON schema enforcement

**Prompt Template Structure**:
```
System Prompt:
You are a compliance analysis assistant. Your task is to analyze security 
and operational events to determine their relevance to specific compliance 
control requirements. Respond ONLY with valid JSON matching this schema:

{
  "evidence_links": ["event_id_1", "event_id_2"],
  "justification": "Brief explanation of relevance",
  "confidence": 85,
  "residual_risk": "Optional notes on gaps or concerns"
}

User Prompt:
Control: [CONTROL_ID] - [CONTROL_TITLE]
Framework: [SOC2 / ISO 27001 / PCI DSS]
Description: [CONTROL_DESCRIPTION]

Policy Excerpt:
[RELEVANT_POLICY_TEXT]

Events:
1. [EVENT_ID]: [EVENT_TYPE] - [EVENT_DESCRIPTION]
   Source: [GIT/CICD/JIRA/SLACK/DOCS]
   Timestamp: [TIMESTAMP]
   Content: [REDACTED_EVENT_CONTENT]

2. [EVENT_ID]: ...

Analyze which events provide evidence for this control and assign a 
confidence score (0-100).
```

**Best Practices**:
1. **Concise policy excerpts**: Include only relevant sections (200-500 words)
2. **Event summarization**: Truncate long events to 200 chars (preserves context)
3. **Schema in system prompt**: Reinforces JSON structure expectations
4. **Examples for few-shot**: Include 2-3 examples in system prompt (optional, future)
5. **Temperature tuning**: Use 0.3 for structured analysis (deterministic), 0.7 for residual risk notes (creative)

**Rationale**:
- System prompt sets role and output format consistently
- User prompt provides structured context (control + events)
- Schema enforcement reduces parsing errors
- Truncation/summarization respects token limits (4K-8K tokens)

**Alternatives Considered**:
- **Single-turn prompt**: Rejected - less control over output format
- **Few-shot examples**: Deferred - adds token overhead, test if needed
- **Chain-of-thought**: Rejected - unnecessary for this task, increases latency

---

## Summary of Research Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **OpenAI SDK** | `github.com/sashabaranov/go-openai` | Community standard, active maintenance |
| **Anthropic SDK** | `github.com/anthropics/anthropic-sdk-go` | Official SDK, first-class support |
| **Structured Output** | Function calling / tool use | More reliable than prompt engineering |
| **PII Redaction** | Regex-based detection + placeholders | High recall, preserves context |
| **Caching** | Event-driven invalidation, filesystem | Surgical invalidation, portable |
| **Retry Strategy** | Exponential backoff + circuit breaker | Resilient to transient failures |
| **Rate Limiting** | `golang.org/x/time/rate` (10 req/min) | Prevents quota exhaustion |
| **Prompt Design** | System + User with JSON schema | Structured context, enforced format |

---

## Next Steps (Phase 1)

1. Define data model for AI requests/responses in `data-model.md`
2. Generate API contracts for `ai.Engine` interface in `contracts/`
3. Create contract tests for provider adapters
4. Extract integration test scenarios from feature spec
5. Update `.github/copilot-instructions.md` with AI feature context
