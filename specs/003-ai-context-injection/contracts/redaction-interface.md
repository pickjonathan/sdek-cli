# Contract: Redaction Pipeline

**Feature**: 003-ai-context-injection  
**Purpose**: Define redaction interface for PII/secret removal from evidence

---

## Interface Definition

```go
package ai

import "github.com/pickjonathan/sdek-cli/pkg/types"

// Redactor removes PII and secrets from text before sending to AI providers.
type Redactor interface {
    // Redact removes PII/secrets and returns redacted text + metadata.
    //
    // Parameters:
    //   text: Original text (evidence, excerpt, etc.)
    //
    // Returns:
    //   redacted: Text with PII/secrets replaced by placeholders
    //   redactionMap: Metadata about what was redacted (never persisted)
    //   error: If regex compilation or processing fails
    //
    // Behavior:
    //   - MUST apply patterns in order: denylist → emails → IPs → API keys
    //   - MUST use placeholders: [REDACTED:PII:EMAIL], [REDACTED:SECRET]
    //   - MUST be idempotent (redacting twice = same result)
    //   - MUST NOT store redaction map to disk
    //   - MUST warn if redaction removes >40% of text
    //
    // Performance: <10ms per event (<100 events = <1s total)
    Redact(text string) (redacted string, redactionMap *types.RedactionMap, error error)
}
```

---

## Patterns

**PII Patterns**:
- Email: `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b` → `[REDACTED:PII:EMAIL]`
- IPv4: `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b` → `[REDACTED:PII:IP]`
- IPv6: `\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b` → `[REDACTED:PII:IP]`
- Phone: `\b(?:\+?1[-.]?)?\(?([0-9]{3})\)?[-.]?([0-9]{3})[-.]?([0-9]{4})\b` → `[REDACTED:PII:PHONE]`

**Secret Patterns**:
- AWS Key: `AKIA[0-9A-Z]{16}` → `[REDACTED:SECRET]`
- Generic API Key: `\b[A-Za-z0-9]{32,64}\b` (heuristic) → `[REDACTED:SECRET]`

**Denylist** (exact matches, case-insensitive):
- Configurable via `redaction.denylist` in config.yaml
- Example: ["password123", "secret-token", "api-key-xyz"]

---

## Testing

**Unit Tests**:
- Email redaction: `user@example.com` → `[REDACTED:PII:EMAIL]`
- IP redaction: `192.168.1.1` → `[REDACTED:PII:IP]`
- API key redaction: `AKIAIOSFODNN7EXAMPLE` → `[REDACTED:SECRET]`
- Denylist: `secret-token` → `[REDACTED:SECRET]`
- >40% redaction warning

**Golden Tests**:
- Input: `events_with_pii.json` → Output: `events_redacted.json`
- Verify redaction map counts match expected

**Performance Tests**:
- 100 events, each 1KB → total redaction time <1s
