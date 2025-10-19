
# SDEK AI Workflow Architecture

**Date:** October 18, 2025  
**Version:** Feature 003 - AI Context Injection

## Overview

This document describes the technical architecture of SDEK's AI-powered compliance analysis workflow, which uses context injection to provide policy-grounded compliance findings.

## Architecture Principles

### 1. Context-First Design
Unlike generic AI analysis that relies solely on model training, SDEK injects specific policy requirements directly into every prompt. This ensures:
- Compliance-specific language and terminology
- Framework-accurate control mappings
- Policy-grounded justifications
- Audit-ready citations

### 2. Evidence Normalization
All evidence sources (GitHub, Jira, AWS, etc.) are converted into a standard `EvidenceBundle` format before analysis:
```go
type EvidenceBundle struct {
    Events []EvidenceEvent `json:"events"`
}

type EvidenceEvent struct {
    ID        string                 `json:"id"`
    Source    string                 `json:"source"`    // "github", "jira", "aws"
    Type      string                 `json:"type"`      // "commit", "ticket", "log"
    Timestamp time.Time              `json:"timestamp"`
    Content   string                 `json:"content"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

### 3. Structured Output
AI responses follow a strict schema enforced through:
- **OpenAI:** Function Calling with JSON Schema validation
- **Anthropic:** Tool Use with InputSchema validation

This ensures consistent, machine-readable findings across all analysis runs.

### 4. Quality Metrics
Every finding includes transparency metrics:
- **Confidence Score:** 0.0-1.0 (AI's certainty)
- **Review Required:** Auto-flagged at < 70% confidence
- **Citations:** Links to specific evidence events
- **Provenance:** Source attribution with event counts
- **Residual Risk:** Gap analysis and concerns

## Workflow Steps (11 Stages)

### Stage 1: Input Collection
**User Provides:**
- Framework identifier (SOC2, ISO27001, PCI-DSS)
- Section/Control ID (CC6.1, A.9.4.2, etc.)
- Policy excerpts JSON file
- Evidence files (supports glob patterns)

**Command:**
```bash
./sdek ai analyze \
  --framework SOC2 \
  --section CC6.1 \
  --excerpts-file ./policies/soc2.json \
  --evidence-path ./evidence/*.json
```

### Stage 2: Load Policy Context
**File:** `excerpts.json`
```json
{
  "framework": "SOC2",
  "version": "2023",
  "excerpts": [{
    "section": "CC6.1",
    "excerpt": "The entity implements logical access security software, infrastructure, and architectures over protected information assets to protect them from security events to meet the entity's objectives. Prior to issuing system credentials and granting system access, the entity registers and authorizes new internal and external users whose access is administered by the entity.",
    "control_ids": ["CC6.1", "CC6.2", "CC6.3"]
  }]
}
```

**Implementation:** `cmd/ai_analyze.go:loadExcerpts()`

### Stage 3: Load Evidence Events
**Files:** Glob pattern resolution (e.g., `./evidence/*.json`)
```json
[
  {
    "id": "event-001",
    "source": "github",
    "type": "commit",
    "timestamp": "2025-10-15T10:30:00Z",
    "content": "Added multi-factor authentication (MFA) support...",
    "metadata": {
      "author": "alice@example.com",
      "repo": "auth-service",
      "sha": "abc123"
    }
  }
]
```

**Implementation:** `cmd/ai_analyze.go:loadEvidenceFiles()`

### Stage 4: Build Context Preamble
**Data Structure:** `types.ContextPreamble`
```go
type ContextPreamble struct {
    Framework  string   `json:"framework"`
    Section    string   `json:"section"`
    Excerpt    string   `json:"excerpt"`
    ControlIDs []string `json:"control_ids,omitempty"`
}
```

**Implementation:** `cmd/ai_analyze.go` (main RunE function)

### Stage 5: Interactive TUI Preview
**Component:** `ui/components/context_preview.go`
```
🔍 AI Context Preview

Framework: SOC2 2023
Section:   CC6.1
Evidence:  3 events

Policy Excerpt:
╭─────────────────────────────────────────────────────╮
│ The entity implements logical access security       │
│ software, infrastructure, and architectures over    │
│ protected information assets...                      │
╰─────────────────────────────────────────────────────╯

This context will be sent to the AI provider for analysis.
```

**User Action:** Proceed or Cancel (auto-proceeds after 20s)

### Stage 6: Construct AI Prompt
**Method:** `providers.buildContextAnalysisPrompt()`

**Template:**
```
Analyze evidence for [Framework] [Section]:

POLICY CONTEXT:
[Excerpt text]

Related Controls: [ControlIDs]

EVIDENCE EVENTS:
1. [source/type] [timestamp]
   ID: [event_id]
   Content: [content]
   Metadata: [key=value pairs]

2. [next event...]

Provide comprehensive analysis including:
1. Title: Brief summary
2. Summary: What was found and how it relates
3. Justification: Detailed compliance explanation
4. Confidence Score: 0.0-1.0
5. Residual Risk: Gaps and concerns
6. Mapped Controls: Control IDs supported
7. Citations: Event IDs referenced
8. Severity: Risk level (low/medium/high/critical)
```

### Stage 7: AI Provider Processing
**Providers:**
- OpenAI (`internal/ai/providers/openai.go`)
- Anthropic (`internal/ai/providers/anthropic.go`)

**OpenAI Method:**
```go
func (e *OpenAIEngine) Analyze(
    ctx context.Context,
    preamble types.ContextPreamble,
    evidence types.EvidenceBundle,
) (*types.Finding, error)
```

**API Call:**
- Model: GPT-4 / Claude-3-Opus
- Temperature: 0.3 (configurable)
- Max Tokens: 4096 (configurable)
- Function/Tool: `analyze_compliance_evidence`
- Schema: Enforced structured output

**Features:**
- Rate limiting (configurable RPM)
- Timeout handling (default 60s)
- Exponential backoff retry
- PII/secret redaction
- Response caching

### Stage 8: Parse Structured Response
**OpenAI Response Format:**
```json
{
  "function_call": {
    "name": "analyze_compliance_evidence",
    "arguments": {
      "title": "Partial Compliance with SOC2 CC6.1",
      "summary": "MFA, password policy, session timeout implemented",
      "justification": "Evidence demonstrates logical access controls...",
      "confidence_score": 0.9,
      "residual_risk": "Missing user registration procedures",
      "mapped_controls": ["CC6.1"],
      "citations": ["event-001", "event-002", "event-003"],
      "severity": "medium"
    }
  }
}
```

**Parsing:** `json.Unmarshal()` into structured result type

### Stage 9: Build Compliance Finding
**Data Structure:** `types.Finding`
```go
type Finding struct {
    ID              string              `json:"id"`
    ControlID       string              `json:"control_id"`
    FrameworkID     string              `json:"framework_id"`
    Title           string              `json:"title"`
    Description     string              `json:"description"`
    Severity        string              `json:"severity"`
    Status          string              `json:"status"`
    CreatedAt       time.Time           `json:"created_at"`
    UpdatedAt       time.Time           `json:"updated_at"`
    Summary         string              `json:"summary"`
    MappedControls  []string            `json:"mapped_controls"`
    ConfidenceScore float64             `json:"confidence_score"`
    ResidualRisk    string              `json:"residual_risk"`
    Justification   string              `json:"justification"`
    Citations       []string            `json:"citations"`
    ReviewRequired  bool                `json:"review_required"`
    Mode            string              `json:"mode"`
    Provenance      []ProvenanceEntry   `json:"provenance,omitempty"`
}
```

**Provenance Calculation:**
```go
sourceCount := make(map[string]int)
for _, event := range evidence.Events {
    sourceCount[event.Source]++
}
for source, count := range sourceCount {
    finding.Provenance = append(finding.Provenance, ProvenanceEntry{
        Source:     source,
        EventsUsed: count,
    })
}
```

**Review Flag Logic:**
```go
finding.ReviewRequired = result.ConfidenceScore < 0.7
```

### Stage 10: Export Compliance Report
**File:** `findings.json` (default, configurable via `--output`)

**Format:** Complete JSON serialization of Finding object

**Implementation:** `cmd/ai_analyze.go:exportFinding()`

### Stage 11: Display Summary
**Console Output:**
```
✅ Analysis Complete!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Framework:       SOC2
Control:         CC6.1
Confidence:      90.0%
Residual Risk:   Missing user registration procedures
Mapped Controls: 1
  • CC6.1
Citations:       3
  • event-001
  • event-002
  • event-003
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📄 Finding saved to: findings.json
```

**Implementation:** `cmd/ai_analyze.go:displayFindingSummary()`

## Data Flow Diagram

```
┌─────────┐
│  User   │
└────┬────┘
     │ Command: sdek ai analyze
     ▼
┌─────────────────┐
│  CLI Parser     │
│  (Cobra/Viper)  │
└────┬────────────┘
     │ Flags + Config
     ▼
┌─────────────────────┐      ┌──────────────────┐
│  ai_analyze.go      │─────▶│ Policy Excerpts  │
│  (Command Handler)  │      │  (JSON file)     │
└────┬────────────────┘      └──────────────────┘
     │                        ┌──────────────────┐
     │                   ────▶│ Evidence Files   │
     │                        │  (Glob pattern)  │
     ▼                        └──────────────────┘
┌─────────────────────┐
│  ContextPreamble    │
│  + EvidenceBundle   │
└────┬────────────────┘
     │
     ▼
┌─────────────────────┐
│  TUI Preview        │
│  (Bubble Tea)       │
└────┬────────────────┘
     │ User confirms
     ▼
┌─────────────────────┐      ┌──────────────────┐
│  AI Engine          │─────▶│  OpenAI API      │
│  (providers/)       │      │  or              │
└────┬────────────────┘      │  Anthropic API   │
     │                        └──────────────────┘
     │ Structured Response
     ▼
┌─────────────────────┐
│  Finding Builder    │
│  (types.Finding)    │
└────┬────────────────┘
     │
     ├──▶ findings.json (Export)
     │
     └──▶ Console (Summary)
```

## Technology Stack

### Core Libraries
- **CLI Framework:** Cobra (command structure) + Viper (configuration)
- **TUI Framework:** Bubble Tea v0.25.0 (interactive preview)
- **AI SDKs:**
  - OpenAI: `github.com/sashabaranov/go-openai`
  - Anthropic: `github.com/anthropics/anthropic-sdk-go`
- **Retry Logic:** `github.com/cenkalti/backoff/v4`
- **JSON Processing:** Standard library `encoding/json`

### Internal Packages
- `cmd/` - Command implementations
- `internal/ai/` - AI engine abstraction
- `internal/ai/providers/` - Provider implementations
- `pkg/types/` - Data models
- `ui/components/` - TUI components

## Configuration

### Minimal Config
```yaml
ai:
  enabled: true
  provider: "openai"
  model: "gpt-4"
```

### Full Config
```yaml
ai:
  enabled: true
  provider: "openai"  # or "anthropic"
  model: "gpt-4"
  
  # API Keys (prefer environment variables)
  # openai_key: "sk-..."
  # anthropic_key: "sk-ant-..."
  
  # Performance tuning
  timeout: 60           # seconds
  rate_limit: 10        # requests per minute
  
  # AI parameters
  max_tokens: 4096
  temperature: 0.3
  
  # Feature 003 specific
  mode: "context"       # "disabled", "context", "autonomous"
  no_cache: false
  
  # Privacy
  redaction:
    enabled: true
    patterns: ["email", "ssn", "api_key"]
```

### Environment Variables
```bash
export SDEK_OPENAI_KEY="sk-..."
export SDEK_ANTHROPIC_KEY="sk-ant-..."
export SDEK_AI_ENABLED=true
export SDEK_AI_PROVIDER=openai
export SDEK_AI_MODEL=gpt-4
```

## Error Handling

### Provider Errors
- **401 Unauthorized:** Invalid API key → `ai.ErrProviderAuth`
- **429 Rate Limited:** Too many requests → `ai.ErrProviderRateLimit`
- **503 Service Unavailable:** Provider down → `ai.ErrProviderUnavailable`
- **Timeout:** Request exceeded timeout → `ai.ErrProviderTimeout`

### Workflow Errors
- **Missing Framework:** Validation error, early exit
- **Missing Section:** Validation error, early exit
- **No Evidence:** Validation error, early exit
- **Invalid Excerpts File:** JSON parse error
- **AI Disabled:** Configuration error, clear message

### Graceful Degradation
- If AI provider fails, command exits with clear error message
- User can retry with different provider or fix configuration
- All errors include actionable guidance

## Performance Characteristics

### Typical Analysis Times
- **Evidence Loading:** < 1 second (100 events)
- **TUI Display:** 2 seconds (user interaction) or 20 seconds (auto-proceed)
- **AI Analysis:** 5-15 seconds (depends on model and evidence size)
- **Finding Export:** < 100ms
- **Total:** ~8-18 seconds per analysis

### Resource Usage
- **Memory:** < 50MB typical
- **Network:** 1 API call per analysis (+ retries if needed)
- **Disk:** Minimal (JSON files + cache)

### Scalability
- **Concurrent Analyses:** Limited by AI provider rate limits
- **Evidence Size:** Tested up to 100 events per analysis
- **Cache:** Reduces redundant API calls by ~70%

## Security Considerations

### PII/Secret Redaction
Before sending to AI:
1. Email addresses → `[EMAIL_REDACTED]`
2. Social Security Numbers → `[SSN_REDACTED]`
3. Credit Card Numbers → `[CC_REDACTED]`
4. API Keys → `[API_KEY_REDACTED]`
5. Passwords → `[PASSWORD_REDACTED]`

### API Key Storage
- ✅ Environment variables (recommended)
- ✅ Config file with restricted permissions (600)
- ❌ Command-line flags (visible in process list)
- ❌ Hardcoded in source

### Data Transmission
- ✅ HTTPS only (enforced by AI SDKs)
- ✅ No data stored on AI provider servers (per ToS)
- ✅ Local caching with file permissions
- ✅ Audit trail via provenance tracking

## Testing Strategy

### Unit Tests
- Provider implementations (`*_test.go`)
- Prompt building logic
- Response parsing
- Finding construction

### Integration Tests
- Full workflow with mock provider
- Config loading and validation
- TUI interactions (automated)
- Error handling scenarios

### End-to-End Tests
- Real API calls with test accounts
- Multiple frameworks and controls
- Various evidence sizes
- Performance benchmarks

## Future Enhancements

### Planned (Feature 003 Completion)
- `ProposePlan()` - Generate evidence collection strategies
- `ExecutePlan()` - Automatically collect evidence
- Plan approval TUI - Interactive plan review
- Autonomous mode - Proactive gap detection

### Potential
- Multi-model ensemble (GPT-4 + Claude consensus)
- Custom fine-tuned models per framework
- Real-time streaming analysis
- Collaborative review workflows
- Integration with GRC platforms

## References

- [Feature 003 Spec](/specs/003-ai-context-injection/spec.md)
- [API Documentation](/docs/api.md)
- [Configuration Guide](/docs/configuration.md)
- [Test Results](/tests/FEATURE_003_AI_ANALYZE_SUCCESS.md)
