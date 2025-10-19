# Contract: AI Engine Interface (Extended)

**Feature**: 003-ai-context-injection  
**Purpose**: Define extended AI engine interface for context injection and autonomous evidence planning

---

## Interface Definition

```go
package ai

import (
    "context"
    "github.com/pickjonathan/sdek-cli/pkg/types"
)

// Engine provides AI-powered compliance analysis with context injection
// and autonomous evidence planning capabilities.
type Engine interface {
    // Analyze performs AI analysis with explicit context grounding.
    // Phase 1: Context Mode
    //
    // Parameters:
    //   ctx: Cancellation context
    //   preamble: Framework metadata + control excerpt for grounding
    //   evidence: Normalized evidence events to analyze
    //
    // Returns:
    //   Finding with confidence score, residual risk, citations, and review flag
    //   Error if AI provider fails (caller should fallback to heuristics)
    //
    // Behavior:
    //   - MUST inject preamble (framework, section, excerpt) into prompt
    //   - MUST redact PII/secrets from evidence before sending to provider
    //   - MUST cache prompt/response using digest-based key
    //   - MUST respect --no-cache flag
    //   - MUST set Finding.ReviewRequired = true if confidence < 0.6
    //   - MUST set Finding.Mode = "ai" on success
    //   - MUST NOT leak redaction map to provider
    //
    // Performance:
    //   Target: <30s for typical evidence bundles (<100 events)
    //   Cache hits: <100ms
    Analyze(ctx context.Context, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error)
    
    // ProposePlan generates an evidence collection plan for a given control.
    // Phase 2: Autonomous Planning
    //
    // Parameters:
    //   ctx: Cancellation context
    //   preamble: Framework metadata + control excerpt (no evidence provided)
    //
    // Returns:
    //   EvidencePlan with sources, queries, signal strengths, and approval status
    //   Error if AI provider fails or plan generation fails
    //
    // Behavior:
    //   - MUST inject preamble into prompt
    //   - MUST generate diverse sources (GitHub, Jira, AWS, etc.)
    //   - MUST include queries/filters for each source
    //   - MUST estimate signal strength (0.0-1.0) per item
    //   - MUST apply auto-approve policy to mark matching items
    //   - MUST sort plan items deterministically (source, then query)
    //   - MUST enforce budget limits (max sources, calls, tokens)
    //   - MUST be deterministic for same preamble + policy config
    //
    // Performance:
    //   Target: <10s for typical control
    ProposePlan(ctx context.Context, preamble types.ContextPreamble) (*types.EvidencePlan, error)
    
    // ExecutePlan executes an approved evidence plan via MCP connectors.
    // Phase 2: Autonomous Execution
    //
    // Parameters:
    //   ctx: Cancellation context
    //   plan: Evidence plan with approved items (ApprovalStatus = approved|auto_approved)
    //
    // Returns:
    //   EvidenceBundle with normalized events from all approved sources
    //   Error if MCP connector orchestration fails (partial results allowed)
    //
    // Behavior:
    //   - MUST skip items with ApprovalStatus = denied|pending
    //   - MUST execute approved items in parallel (respecting connector rate limits)
    //   - MUST normalize MCP connector outputs to EvidenceEvent schema
    //   - MUST handle partial failures (log, continue with remaining sources)
    //   - MUST update PlanItem.ExecutionStatus and EventsCollected
    //   - MUST respect budget limits (max API calls)
    //   - MUST emit audit events for each source execution
    //
    // Performance:
    //   Target: <5min for plans with ≤10 sources
    ExecutePlan(ctx context.Context, plan *types.EvidencePlan) (types.EvidenceBundle, error)
}
```

---

## Method Contracts

### Analyze()

**Preconditions**:
- `preamble.Framework`, `preamble.Section`, `preamble.Excerpt` MUST NOT be empty
- `preamble.Excerpt` length MUST be 50-10,000 characters
- `evidence` MAY be empty (will result in low-confidence finding)
- AI provider MUST be configured and accessible

**Postconditions**:
- Returns `Finding` with all required fields populated:
  - `finding_summary` (string)
  - `mapped_controls` ([]string)
  - `confidence_score` (float64, 0.0-1.0)
  - `residual_risk` (string: "low"|"medium"|"high")
  - `justification` (string)
  - `citations` ([]string)
  - `review_required` (bool, true if confidence < 0.6)
  - `mode` (string: "ai"|"heuristics")
- Redaction map MUST NOT be included in finding
- On error, returns nil and error (caller should fallback to heuristics)

**Side Effects**:
- Caches prompt + response (unless `--no-cache` flag set)
- Emits audit event with:
  - Framework, section, confidence, redaction count, cache hit/miss
  - MUST NOT include prompt text, redacted content, or PII
- Logs redaction statistics (count, types) to audit log

**Error Conditions**:
- `ErrProviderUnavailable`: AI provider API is down or unreachable
- `ErrInvalidPreamble`: Preamble validation failed
- `ErrRedactionExceeded`: Redaction removed >40% of evidence (warning, not error)
- `ErrPromptTooLarge`: Prompt exceeds provider token limit

---

### ProposePlan()

**Preconditions**:
- `preamble.Framework`, `preamble.Section`, `preamble.Excerpt` MUST NOT be empty
- AI provider MUST be configured and accessible
- AutoApprovePolicy MUST be loaded from config

**Postconditions**:
- Returns `EvidencePlan` with:
  - At least 1 plan item (up to budget limit)
  - Each item has: source, query, signal_strength, approval_status
  - Items sorted deterministically (source asc, query asc)
  - Items matching auto-approve policy have `auto_approved = true`
  - Budget estimates populated (sources, calls, tokens)
  - Status = "pending" (requires user approval)
- Plan MUST be deterministic for same inputs
- On error, returns nil and error

**Side Effects**:
- Emits audit event with:
  - Framework, section, item count, auto-approved count, budget estimates
- NO caching (plans should be fresh)

**Error Conditions**:
- `ErrProviderUnavailable`: AI provider API is down
- `ErrInvalidPreamble`: Preamble validation failed
- `ErrBudgetExceeded`: Generated plan exceeds budget limits
- `ErrNoPlanItems`: AI failed to generate any valid plan items

---

### ExecutePlan()

**Preconditions**:
- `plan.Status` MUST be "approved"
- At least one `PlanItem` MUST have `ApprovalStatus = approved|auto_approved`
- MCP connectors for approved sources MUST be configured

**Postconditions**:
- Returns `EvidenceBundle` with normalized events from all approved sources
- `PlanItem.ExecutionStatus` updated for each item (complete|failed)
- `PlanItem.EventsCollected` populated for successful items
- `PlanItem.Error` populated for failed items
- Partial success allowed (some sources may fail)
- On catastrophic error, returns nil and error

**Side Effects**:
- Emits audit event for each source execution:
  - Source, query, events collected, duration, error (if any)
- Updates plan in-memory (caller responsible for persistence)
- May trigger MCP connector rate limiting (exponential backoff applied)

**Error Conditions**:
- `ErrPlanNotApproved`: Plan status is not "approved"
- `ErrNoApprovedItems`: No items have approval_status = approved|auto_approved
- `ErrMCPConnectorFailed`: All MCP connectors failed (partial failure = not error)

---

## Testing Strategy

### Unit Tests
- Mock AI provider responses
- Test context injection (preamble in prompt)
- Test redaction pipeline (PII/secrets removed)
- Test cache hit/miss logic
- Test auto-approve matching
- Test confidence threshold flagging
- Test budget limit enforcement

### Integration Tests
- E2E context mode: spec → preamble → analyze → finding
- E2E autonomous mode: spec → preamble → propose → approve → execute → analyze → finding
- Cache persistence across invocations
- Redaction audit log verification
- MCP connector failures (partial success)

### Contract Tests
- Assert `Analyze()` returns valid Finding schema
- Assert `ProposePlan()` returns valid EvidencePlan schema
- Assert `ExecutePlan()` returns valid EvidenceBundle schema
- Assert errors match expected error types

---

## Example Usage

### Context Mode (Phase 1)
```go
preamble := types.ContextPreamble{
    Framework: "SOC2",
    Version:   "2017",
    Section:   "CC6.1",
    Excerpt:   "The entity implements logical access security measures...",
    Rubrics: types.AnalysisRubrics{
        ConfidenceThreshold: 0.6,
        RiskLevels:          []string{"low", "medium", "high"},
        RequiredCitations:   3,
    },
}

evidence := loadEvidenceFromFiles("./evidence/*.json")

finding, err := engine.Analyze(ctx, preamble, evidence)
if err != nil {
    // Fallback to heuristics mode
    finding = heuristicAnalysis(preamble, evidence)
    finding.Mode = types.ModeHeuristics
}

if finding.ReviewRequired {
    fmt.Println("⚠ Low confidence - manual review required")
}
```

### Autonomous Mode (Phase 2)
```go
preamble := types.ContextPreamble{
    Framework: "ISO27001",
    Version:   "2013",
    Section:   "A.9.4.2",
    Excerpt:   "Where required by the access control policy, access to systems...",
}

// 1. Generate plan
plan, err := engine.ProposePlan(ctx, preamble)
if err != nil {
    log.Fatalf("Plan generation failed: %v", err)
}

// 2. User reviews and approves (or auto-approve applies)
for i, item := range plan.Items {
    if item.AutoApproved {
        plan.Items[i].ApprovalStatus = types.ApprovalAutoApproved
    } else {
        // Show TUI for manual approval
        plan.Items[i].ApprovalStatus = getUserApproval(item)
    }
}
plan.Status = types.PlanApproved

// 3. Execute plan
evidence, err := engine.ExecutePlan(ctx, plan)
if err != nil {
    log.Fatalf("Plan execution failed: %v", err)
}

// 4. Analyze with collected evidence
finding, err := engine.Analyze(ctx, preamble, evidence)
if err != nil {
    finding = heuristicAnalysis(preamble, evidence)
    finding.Mode = types.ModeHeuristics
}

// Finding includes provenance showing which sources contributed
for _, p := range finding.Provenance {
    fmt.Printf("Source: %s, Events: %d, Contribution: %.2f\n", 
        p.Source, p.EventsUsed, p.Contribution)
}
```

---

## Backward Compatibility

This interface extends the existing `Engine` interface from 002-ai-evidence-analysis:
- `Analyze()` signature changes: adds `preamble` parameter (breaking change)
- Mitigation: Old call sites can pass empty preamble (will work but won't ground AI)
- `ProposePlan()` and `ExecutePlan()` are new methods (additive, no breaking change)

**Migration Path**:
1. Update all `Analyze()` call sites to construct `ContextPreamble`
2. Extract framework/section from existing code or config
3. Load excerpts from policy files (already exist)
4. Test with seeded demo data to verify behavior

---

## Summary

This contract defines three core methods:
1. **Analyze()**: Context-grounded AI analysis with redaction and caching
2. **ProposePlan()**: Generate evidence collection plans with auto-approval
3. **ExecutePlan()**: Orchestrate MCP connectors to collect evidence

All methods have clear preconditions, postconditions, side effects, and error conditions. Ready for implementation and testing.
