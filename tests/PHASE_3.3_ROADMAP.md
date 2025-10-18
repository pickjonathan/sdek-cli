# Phase 3.3 Implementation Roadmap

**Feature**: 003-ai-context-injection  
**Prerequisites**: Phase 3.1 ✅ Complete, Phase 3.2 ✅ Complete  
**Goal**: Make all 102 unit tests pass

---

## Quick Start

### Current State
```bash
$ go test ./tests/unit -v
# FAIL: All tests fail to compile (expected)
# Error: undefined types and interfaces
```

### Target State
```bash
$ go test ./tests/unit -v
# PASS: All 102 tests pass
# Coverage: Types validated, interfaces implemented
```

---

## Implementation Order (T017-T026)

### Stage 1: Type Definitions (Foundation)

#### T017: ContextPreamble & AnalysisRubrics
**File**: `pkg/types/context.go`
```go
type ContextPreamble struct {
    Framework   string
    Version     string
    Section     string
    Excerpt     string
    ControlIDs  []string
    Rubrics     AnalysisRubrics
    CreatedAt   time.Time
}

type AnalysisRubrics struct {
    ConfidenceThreshold float64
    RiskLevels          []string
    RequiredCitations   int
}

// Constructors
func NewContextPreamble(...) (*ContextPreamble, error)
func NewContextPreambleWithRubrics(...) (*ContextPreamble, error)

// Validation
func (cp *ContextPreamble) Validate() error
```

**Validation**: `go test ./tests/unit -run TestNewContextPreamble`

---

#### T018: EvidencePlan & PlanItem
**File**: `pkg/types/plan.go`
```go
type EvidencePlan struct {
    ID               string
    Framework        string
    Section          string
    Items            []PlanItem
    EstimatedSources int
    EstimatedCalls   int
    EstimatedTokens  int
    Status           PlanStatus
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

type PlanItem struct {
    Source          string
    Query           string
    Filters         []string
    SignalStrength  float64
    Rationale       string
    ApprovalStatus  ApprovalStatus
    AutoApproved    bool
    ExecutionStatus ExecStatus
    EventsCollected int
    Error           string
}

// Enums
type PlanStatus string
const (
    PlanPending PlanStatus = "pending"
    PlanApproved PlanStatus = "approved"
    // ...
)
```

**Validation**: `go test ./tests/unit -run TestProposePlan`

---

#### T019: RedactionMap & RedactionEntry
**File**: `pkg/types/redaction.go`
```go
type RedactionMap struct {
    entries         map[string]RedactionEntry // Not exported
    TotalRedactions int
    RedactionTypes  []RedactionType
}

type RedactionEntry struct {
    OriginalHash string         `json:"-"` // Never exported
    Placeholder  string
    Type         RedactionType
    Position     int
    Timestamp    time.Time
}

type RedactionType string
const (
    RedactionPII    RedactionType = "pii"
    RedactionSecret RedactionType = "secret"
)
```

**Validation**: `go test ./tests/unit -run TestRedact`

---

#### T020: EvidenceBundle & EvidenceEvent
**File**: `pkg/types/bundle.go`
```go
type EvidenceBundle struct {
    Events []EvidenceEvent
}

type EvidenceEvent struct {
    ID        string
    Source    string
    Type      string
    Timestamp time.Time
    Content   string
    Metadata  map[string]interface{}
}
```

**Validation**: `go test ./tests/unit -run TestAnalyze -run TestExecutePlan`

---

#### T021: Extend Finding Type
**File**: `pkg/types/finding.go` (extend existing)
```go
type Finding struct {
    // ... existing fields ...
    
    // NEW: Feature 003 additions
    Summary        string         `json:"summary"`
    MappedControls []string       `json:"mapped_controls"`
    ConfidenceScore float64       `json:"confidence_score"`
    ResidualRisk   string         `json:"residual_risk"`
    Justification  string         `json:"justification"`
    Citations      []string       `json:"citations"`
    ReviewRequired bool           `json:"review_required"`
    Mode           string         `json:"mode"`
    Provenance     []ProvenanceEntry `json:"provenance,omitempty"`
}

type ProvenanceEntry struct {
    Source     string
    Query      string
    EventsUsed int
}
```

**Validation**: `go test ./tests/unit -run TestAnalyze`

---

### Stage 2: Core Interfaces

#### T022: Redactor Implementation
**File**: `internal/ai/redactor.go`
```go
type Redactor interface {
    Redact(text string) (redacted string, redactionMap *types.RedactionMap, error error)
}

type redactor struct {
    config *types.Config
    patterns map[string]*regexp.Regexp
}

func NewRedactor(cfg *types.Config) Redactor {
    return &redactor{
        config: cfg,
        patterns: compilePatterns(),
    }
}

func (r *redactor) Redact(text string) (string, *types.RedactionMap, error) {
    if !r.config.AI.Redaction.Enabled {
        return text, &types.RedactionMap{}, nil
    }
    
    // Apply denylist → emails → IPs → keys
    // Track redactions in map
    // Warn if >40% redacted
    // Return redacted text + map
}
```

**Patterns**:
- Email: `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`
- IPv4: `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
- AWS Key: `AKIA[0-9A-Z]{16}`

**Validation**: `go test ./tests/unit -run TestRedact` (21 tests should pass)

---

#### T023: AutoApproveMatcher Implementation
**File**: `internal/ai/autoapprove.go`
```go
type AutoApproveMatcher interface {
    Matches(source, query string) bool
}

type autoApproveMatcher struct {
    policy types.AutoApproveConfig
    enabled bool
    globs map[string][]glob.Glob // Compiled patterns
}

func NewAutoApproveMatcher(cfg *types.Config) AutoApproveMatcher {
    matcher := &autoApproveMatcher{
        policy: cfg.AI.Autonomous.AutoApprove,
        enabled: cfg.AI.Autonomous.Enabled,
        globs: make(map[string][]glob.Glob),
    }
    matcher.compilePatterns()
    return matcher
}

func (m *autoApproveMatcher) Matches(source, query string) bool {
    if !m.enabled {
        return false
    }
    
    // Case-insensitive source lookup
    // Try each glob pattern for source
    // Return true on first match
}
```

**Dependencies**: `github.com/gobwas/glob` (already added)

**Validation**: `go test ./tests/unit -run TestAutoApprove` (20 tests should pass)

---

### Stage 3: Engine Extensions

#### T024: Extend Engine.Analyze()
**File**: `internal/ai/engine.go` (extend existing)
```go
func (e *engine) Analyze(ctx context.Context, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error) {
    // 1. Validate preamble
    if err := preamble.Validate(); err != nil {
        return nil, fmt.Errorf("invalid preamble: %w", err)
    }
    
    // 2. Redact PII from evidence
    redactor := NewRedactor(e.config)
    redactedEvidence := redactBundle(evidence, redactor)
    
    // 3. Build prompt with preamble injection
    prompt := buildPromptWithContext(preamble, redactedEvidence)
    
    // 4. Check cache
    digest := hashPrompt(prompt)
    if !e.config.AI.NoCache {
        if cached := e.cache.Get(digest); cached != nil {
            return cached.(*types.Finding), nil
        }
    }
    
    // 5. Call AI provider
    response, err := e.provider.Complete(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("provider error: %w", err)
    }
    
    // 6. Parse response to Finding
    finding := parseResponseToFinding(response)
    finding.Mode = "ai"
    
    // 7. Set ReviewRequired flag
    if finding.ConfidenceScore < 0.6 {
        finding.ReviewRequired = true
    }
    
    // 8. Cache result
    if !e.config.AI.NoCache {
        e.cache.Set(digest, finding)
    }
    
    return finding, nil
}
```

**Validation**: `go test ./tests/unit -run TestAnalyze` (14 tests should pass)

---

#### T025: Implement Engine.ProposePlan()
**File**: `internal/ai/engine.go`
```go
func (e *engine) ProposePlan(ctx context.Context, preamble types.ContextPreamble) (*types.EvidencePlan, error) {
    // 1. Validate preamble
    if err := preamble.Validate(); err != nil {
        return nil, fmt.Errorf("invalid preamble: %w", err)
    }
    
    // 2. Build planning prompt with preamble
    prompt := buildPlanningPrompt(preamble)
    
    // 3. Call AI provider (NO CACHING)
    response, err := e.provider.Complete(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    // 4. Parse response to plan items
    items := parseResponseToPlanItems(response)
    if len(items) == 0 {
        return nil, ErrNoPlanItems
    }
    
    // 5. Apply auto-approve policy
    matcher := NewAutoApproveMatcher(e.config)
    for i := range items {
        if matcher.Matches(items[i].Source, items[i].Query) {
            items[i].AutoApproved = true
            items[i].ApprovalStatus = types.ApprovalAutoApproved
        } else {
            items[i].ApprovalStatus = types.ApprovalPending
        }
    }
    
    // 6. Sort items deterministically (source asc, query asc)
    sort.Slice(items, func(i, j int) bool {
        if items[i].Source == items[j].Source {
            return items[i].Query < items[j].Query
        }
        return items[i].Source < items[j].Source
    })
    
    // 7. Enforce budget limits
    items = enforceBudgets(items, e.config.AI.Budgets)
    
    // 8. Build plan
    plan := &types.EvidencePlan{
        ID:        uuid.New().String(),
        Framework: preamble.Framework,
        Section:   preamble.Section,
        Items:     items,
        Status:    types.PlanPending,
        CreatedAt: time.Now(),
    }
    plan.EstimatedSources = len(items)
    plan.EstimatedCalls = estimateAPICalls(items)
    plan.EstimatedTokens = estimateTokens(items)
    
    return plan, nil
}
```

**Validation**: `go test ./tests/unit -run TestProposePlan` (16 tests should pass)

---

#### T026: Implement Engine.ExecutePlan()
**File**: `internal/ai/engine.go`
```go
func (e *engine) ExecutePlan(ctx context.Context, plan *types.EvidencePlan) (types.EvidenceBundle, error) {
    // 1. Validate plan status
    if plan.Status != types.PlanApproved {
        return types.EvidenceBundle{}, ErrPlanNotApproved
    }
    
    // 2. Filter approved items
    approvedItems := filterApprovedItems(plan.Items)
    if len(approvedItems) == 0 {
        return types.EvidenceBundle{}, ErrNoApprovedItems
    }
    
    // 3. Execute items in parallel
    var wg sync.WaitGroup
    results := make(chan executionResult, len(approvedItems))
    
    for i := range approvedItems {
        wg.Add(1)
        go func(item *types.PlanItem) {
            defer wg.Done()
            events, err := e.connector.Collect(ctx, item.Source, item.Query)
            results <- executionResult{item: item, events: events, err: err}
        }(&plan.Items[i])
    }
    
    wg.Wait()
    close(results)
    
    // 4. Collect results and update plan
    var allEvents []types.EvidenceEvent
    failures := 0
    
    for result := range results {
        if result.err != nil {
            result.item.ExecutionStatus = types.ExecFailed
            result.item.Error = result.err.Error()
            failures++
        } else {
            result.item.ExecutionStatus = types.ExecComplete
            result.item.EventsCollected = len(result.events)
            allEvents = append(allEvents, result.events...)
        }
    }
    
    // 5. Check if all failed
    if failures == len(approvedItems) {
        return types.EvidenceBundle{}, ErrMCPConnectorFailed
    }
    
    return types.EvidenceBundle{Events: allEvents}, nil
}
```

**Validation**: `go test ./tests/unit -run TestExecutePlan` (15 tests should pass)

---

## Progress Tracking

### Checklist
- [ ] T017: ContextPreamble types (16 tests)
- [ ] T018: EvidencePlan types (16 tests)
- [ ] T019: RedactionMap types (21 tests)
- [ ] T020: EvidenceBundle types (14 tests)
- [ ] T021: Extend Finding type (14 tests)
- [ ] T022: Implement Redactor (21 tests)
- [ ] T023: Implement AutoApproveMatcher (20 tests)
- [ ] T024: Extend Engine.Analyze() (14 tests)
- [ ] T025: Implement Engine.ProposePlan() (16 tests)
- [ ] T026: Implement Engine.ExecutePlan() (15 tests)

### Validation Commands
```bash
# After each task
go test ./tests/unit -v -run <TestPattern>

# Final validation
go test ./tests/unit -v
go test ./tests/unit -cover
go build ./...
```

---

## Mock Implementations Needed

For tests to compile and run, you'll need mock implementations:

### MockProvider
**File**: `internal/ai/providers/mock.go` (already exists, may need extension)
```go
type MockProvider struct {
    responses      []string
    errors         []error
    callCount      int
    lastPrompt     string
    planItems      []types.PlanItem
    confidenceScore float64
}

func (m *MockProvider) Complete(ctx context.Context, prompt string) (string, error)
func (m *MockProvider) SetError(err error)
func (m *MockProvider) SetPlanItems(items []types.PlanItem)
func (m *MockProvider) SetConfidenceScore(score float64)
func (m *MockProvider) GetLastPrompt() string
func (m *MockProvider) GetCallCount() int
```

### MockMCPConnector
**File**: `internal/ai/mcp_mock.go` (new)
```go
type MockMCPConnector struct {
    events map[string][]types.EvidenceEvent
    errors map[string]error
    delay  time.Duration
}

func NewMockMCPConnector() *MockMCPConnector
func (m *MockMCPConnector) Collect(ctx context.Context, source, query string) ([]types.EvidenceEvent, error)
func (m *MockMCPConnector) SetEvents(source string, events []types.EvidenceEvent)
func (m *MockMCPConnector) SetError(source string, err error)
func (m *MockMCPConnector) SetDelay(d time.Duration)
```

---

## Common Pitfalls

1. **Validation Order**: Validate inputs before processing
2. **Error Wrapping**: Use `fmt.Errorf("context: %w", err)` for error chains
3. **Case Sensitivity**: AutoApprove matching is case-insensitive
4. **Cache Keys**: Use consistent hashing for cache digests
5. **Redaction Order**: Apply patterns in order: denylist → emails → IPs → keys
6. **Parallel Safety**: Use sync.WaitGroup and channels for ExecutePlan
7. **Budget Enforcement**: Apply limits AFTER sorting, BEFORE execution
8. **Determinism**: Sort plan items consistently (source asc, query asc)

---

## Success Criteria

✅ All 102 tests pass  
✅ No compile errors  
✅ Performance targets met  
✅ Validation rules enforced  
✅ Error handling complete  

**Estimated time**: 4-6 hours for full implementation
