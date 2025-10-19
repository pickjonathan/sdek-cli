# Data Model: AI Context Injection & Autonomous Evidence Collection

**Feature**: 003-ai-context-injection  
**Date**: 2025-10-17  
**Purpose**: Define data structures, validation rules, and relationships for context injection and autonomous evidence collection

---

## Core Entities

### 1. ContextPreamble

Represents framework metadata and control excerpt used to ground AI prompts.

**Go Struct**:
```go
package types

type ContextPreamble struct {
    // Framework metadata
    Framework   string `json:"framework"`   // e.g., "SOC2", "ISO27001"
    Version     string `json:"version"`     // e.g., "2017", "2013"
    Section     string `json:"section"`     // e.g., "CC6.1", "A.9.4.2"
    
    // Control context
    Excerpt     string `json:"excerpt"`     // Full control text
    ControlIDs  []string `json:"control_ids"` // Related control identifiers
    
    // Analysis configuration
    Rubrics     AnalysisRubrics `json:"rubrics"` // Confidence/risk criteria
    
    // Timestamps
    CreatedAt   time.Time `json:"created_at"`
}

type AnalysisRubrics struct {
    ConfidenceThreshold float64 `json:"confidence_threshold"` // Default: 0.6
    RiskLevels          []string `json:"risk_levels"`          // ["low", "medium", "high"]
    RequiredCitations   int     `json:"required_citations"`    // Min citations for high confidence
}
```

**Validation Rules**:
- `Framework`: MUST NOT be empty, valid values from policy loader
- `Version`: MUST NOT be empty
- `Section`: MUST NOT be empty
- `Excerpt`: MUST NOT be empty, MIN 50 characters, MAX 10,000 characters
- `ControlIDs`: OPTIONAL, each ID MUST match pattern `^[A-Z0-9.-]+$`
- `Rubrics.ConfidenceThreshold`: MUST be in range [0.0, 1.0]

**Relationships**:
- Used as input to `Analyze()` method
- Used as input to `ProposePlan()` method
- Referenced in Finding via framework/section fields

---

### 2. EvidencePlan

Represents a proposed evidence collection plan with approval workflow.

**Go Struct**:
```go
package types

type EvidencePlan struct {
    // Plan metadata
    ID          string    `json:"id"`          // Unique plan ID
    Framework   string    `json:"framework"`   // From preamble
    Section     string    `json:"section"`     // From preamble
    
    // Plan items
    Items       []PlanItem `json:"items"`       // Evidence sources to collect
    
    // Budget tracking
    EstimatedSources int `json:"estimated_sources"` // Total sources
    EstimatedCalls   int `json:"estimated_calls"`   // Total API calls
    EstimatedTokens  int `json:"estimated_tokens"`  // Total AI tokens
    
    // Status
    Status      PlanStatus `json:"status"`      // pending|approved|rejected|executing|complete
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

type PlanItem struct {
    // Source configuration
    Source          string   `json:"source"`           // "github", "jira", "aws", etc.
    Query           string   `json:"query"`            // Search query or filter
    Filters         []string `json:"filters"`          // Additional filters
    
    // Metadata
    SignalStrength  float64  `json:"signal_strength"`  // 0.0-1.0 estimated relevance
    Rationale       string   `json:"rationale"`        // Why this source/query
    
    // Approval
    ApprovalStatus  ApprovalStatus `json:"approval_status"` // pending|approved|denied|auto_approved
    AutoApproved    bool           `json:"auto_approved"`   // Matched auto-approve policy
    
    // Execution
    ExecutionStatus ExecStatus `json:"execution_status,omitempty"` // pending|running|complete|failed
    EventsCollected int        `json:"events_collected,omitempty"` // Count after execution
    Error           string     `json:"error,omitempty"`            // Error if failed
}

type PlanStatus string
const (
    PlanPending   PlanStatus = "pending"
    PlanApproved  PlanStatus = "approved"
    PlanRejected  PlanStatus = "rejected"
    PlanExecuting PlanStatus = "executing"
    PlanComplete  PlanStatus = "complete"
)

type ApprovalStatus string
const (
    ApprovalPending      ApprovalStatus = "pending"
    ApprovalApproved     ApprovalStatus = "approved"
    ApprovalDenied       ApprovalStatus = "denied"
    ApprovalAutoApproved ApprovalStatus = "auto_approved"
)

type ExecStatus string
const (
    ExecPending  ExecStatus = "pending"
    ExecRunning  ExecStatus = "running"
    ExecComplete ExecStatus = "complete"
    ExecFailed   ExecStatus = "failed"
)
```

**Validation Rules**:
- `ID`: MUST be unique, UUID v4 format
- `Framework`, `Section`: MUST NOT be empty
- `Items`: MUST have at least 1 item, MAX determined by budget limits
- `PlanItem.Source`: MUST be one of configured MCP connectors
- `PlanItem.Query`: MUST NOT be empty
- `PlanItem.SignalStrength`: MUST be in range [0.0, 1.0]
- Budget fields: MUST NOT exceed configured limits (default: 50/500/250K)

**State Transitions**:
```
pending → approved (user/auto-approve)
pending → rejected (user)
approved → executing (plan execution starts)
executing → complete (all items processed)
```

**Relationships**:
- Created by `ProposePlan()` method
- Input to `ExecutePlan()` method
- Results in EvidenceBundle for `Analyze()` method

---

### 3. Finding (Extended)

Extends existing Finding type with review_required field and provenance.

**Go Struct** (additions only):
```go
package types

type Finding struct {
    // ... existing fields (finding_summary, mapped_controls, confidence_score, etc.) ...
    
    // NEW: Review flagging
    ReviewRequired bool `json:"review_required"` // true if confidence < 0.6
    
    // NEW: Provenance (for autonomous mode)
    Provenance []ProvenanceEntry `json:"provenance,omitempty"` // Sources that contributed
    
    // NEW: Analysis mode
    Mode AnalysisMode `json:"mode"` // "ai" or "heuristics"
}

type ProvenanceEntry struct {
    Source        string `json:"source"`         // "github", "aws", etc.
    Query         string `json:"query"`          // Query used
    EventsUsed    int    `json:"events_used"`    // Count of events from this source
    Contribution  float64 `json:"contribution"`  // 0.0-1.0 estimated contribution to finding
}

type AnalysisMode string
const (
    ModeAI         AnalysisMode = "ai"
    ModeHeuristics AnalysisMode = "heuristics"
)
```

**Validation Rules**:
- `ReviewRequired`: MUST be true if `ConfidenceScore < 0.6`
- `Provenance`: OPTIONAL, only present in autonomous mode
- `ProvenanceEntry.Contribution`: MUST be in range [0.0, 1.0], all contributions SHOULD sum to ~1.0
- `Mode`: MUST be "ai" or "heuristics"

**Relationships**:
- Output of `Analyze()` method
- Referenced in exports (JSON/HTML)
- Displayed in TUI with review badge if `ReviewRequired == true`

---

### 4. RedactionMap

Represents redacted content mapping. **Stored in-memory only, never persisted.**

**Go Struct**:
```go
package types

type RedactionMap struct {
    // Map of original hash -> redaction metadata
    entries map[string]RedactionEntry
    
    // Statistics
    TotalRedactions int    `json:"total_redactions"`
    RedactionTypes  []RedactionType `json:"redaction_types"`
}

type RedactionEntry struct {
    OriginalHash string         `json:"-"` // SHA256 of original text (not exported)
    Placeholder  string         `json:"placeholder"` // e.g., "[REDACTED:PII:EMAIL]"
    Type         RedactionType  `json:"type"`
    Position     int            `json:"position"` // Character offset in original text
    Timestamp    time.Time      `json:"timestamp"`
}

type RedactionType string
const (
    RedactionPII    RedactionType = "pii"      // Email, phone, IP
    RedactionSecret RedactionType = "secret"   // API keys, tokens
)
```

**Validation Rules**:
- `entries`: MUST NOT be serialized to JSON/disk
- `OriginalHash`: SHA256 hex string (64 characters)
- `Placeholder`: MUST match pattern `^\[REDACTED:(PII|SECRET)(:[A-Z]+)?\]$`
- `Type`: MUST be "pii" or "secret"

**Lifecycle**:
- Created per analysis request
- Populated during redaction phase
- Used for audit logging (count/types only, not content)
- Discarded after analysis completes

**Security**:
- MUST NOT be persisted to disk
- MUST NOT be sent to AI providers
- MUST NOT be included in exports
- Audit logs MAY include counts and types, MUST NOT include hashes or original content

---

### 5. AutoApprovePolicy

Configuration for automatically approving evidence plan items.

**Go Struct**:
```go
package types

type AutoApprovePolicy struct {
    // Source -> patterns mapping
    Rules map[string][]string `yaml:"rules" json:"rules"` // e.g., {"github": ["auth*", "*login*"]}
    
    // Global settings
    Enabled bool `yaml:"enabled" json:"enabled"` // Default: false
}
```

**YAML Config Format**:
```yaml
ai:
  autonomous:
    autoApprove:
      enabled: false # Explicit opt-in
      rules:
        github: ["auth*", "*login*", "mfa*", "2fa*"]
        aws: ["iam*", "security*", "kms*"]
        jira: ["INFOSEC-*", "SEC-*"]
        slack: ["#security*"]
```

**Validation Rules**:
- `Rules`: MUST be map[string][]string
- Source keys: MUST match configured MCP connector names
- Patterns: MUST be valid glob patterns (validated via gobwas/glob)
- `Enabled`: MUST be explicitly set to true for auto-approval to work

**Matching Logic**:
```go
func (p *AutoApprovePolicy) Matches(source, query string) bool {
    if !p.Enabled {
        return false
    }
    
    patterns, ok := p.Rules[source]
    if !ok {
        return false // Source not whitelisted
    }
    
    queryLower := strings.ToLower(query)
    for _, pattern := range patterns {
        g := glob.MustCompile(strings.ToLower(pattern))
        if g.Match(queryLower) {
            return true
        }
    }
    
    return false
}
```

**Relationships**:
- Loaded from config at startup
- Used during `ProposePlan()` to mark items as auto-approved
- Affects `PlanItem.AutoApproved` and `PlanItem.ApprovalStatus`

---

## Entity Relationships Diagram

```
┌─────────────────┐
│ ContextPreamble │
└────────┬────────┘
         │
         ├─── input to ───→ ProposePlan() ───→ ┌──────────────┐
         │                                       │ EvidencePlan │
         │                                       └──────┬───────┘
         │                                              │
         │                               input to ──────┤
         │                                              ↓
         │                                    ExecutePlan()
         │                                              ↓
         │                                  ┌────────────────────┐
         │                                  │  EvidenceBundle    │
         │                                  └─────────┬──────────┘
         │                                            │
         └─── input to ───→ Analyze() ←──────────────┘
                                ↓
                         ┌──────────┐
                         │ Finding  │ ──→ (includes Provenance, ReviewRequired)
                         └──────────┘

┌──────────────┐
│ RedactionMap │ ──→ (used during Analyze, never persisted)
└──────────────┘

┌───────────────────┐
│ AutoApprovePolicy │ ──→ (used during ProposePlan, affects EvidencePlan items)
└───────────────────┘
```

---

## Validation Summary

| Entity | Critical Validations |
|--------|---------------------|
| ContextPreamble | Framework/Section/Excerpt non-empty, Excerpt 50-10K chars, ConfidenceThreshold 0.0-1.0 |
| EvidencePlan | ID unique UUID, Items non-empty, Budget limits enforced, Status transitions valid |
| Finding | ReviewRequired = (confidence < 0.6), Provenance contributions sum to ~1.0, Mode valid |
| RedactionMap | Never serialized, OriginalHash SHA256, Placeholder matches pattern |
| AutoApprovePolicy | Enabled explicit opt-in, Rules map[string][]string, Patterns valid globs |

---

## State Transition Diagrams

### EvidencePlan Status
```
┌─────────┐
│ pending │
└────┬────┘
     │
     ├─── user/auto-approve ───→ ┌──────────┐
     │                            │ approved │
     │                            └────┬─────┘
     │                                 │
     │                           start execution
     │                                 │
     │                                 ↓
     │                          ┌────────────┐
     │                          │ executing  │
     │                          └─────┬──────┘
     │                                │
     │                           all items done
     │                                │
     │                                ↓
     │                          ┌──────────┐
     │                          │ complete │
     │                          └──────────┘
     │
     └─── user reject ───→ ┌──────────┐
                           │ rejected │
                           └──────────┘
```

### PlanItem ApprovalStatus
```
┌─────────┐
│ pending │
└────┬────┘
     │
     ├─── manual approval ───→ ┌──────────┐
     │                          │ approved │
     │                          └──────────┘
     │
     ├─── auto-approve match ───→ ┌────────────────┐
     │                             │ auto_approved  │
     │                             └────────────────┘
     │
     └─── user deny ───→ ┌────────┐
                         │ denied │
                         └────────┘
```

---

## Summary

This data model defines 5 core entities and their relationships:
1. **ContextPreamble**: Input for grounding AI prompts
2. **EvidencePlan**: Proposed collection plan with approval workflow
3. **Finding** (extended): Output with review flagging and provenance
4. **RedactionMap**: In-memory-only redaction tracking
5. **AutoApprovePolicy**: Configuration-driven auto-approval

All entities have clear validation rules, state transitions (where applicable), and security constraints (RedactionMap never persisted). Ready for contract definition and implementation.
