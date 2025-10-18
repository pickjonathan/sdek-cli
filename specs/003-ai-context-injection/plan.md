
# Implementation Plan: AI Context Injection & Autonomous Evidence Collection

**Branch**: `003-ai-context-injection` | **Date**: 2025-10-17 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-ai-context-injection/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code, or `AGENTS.md` for all other agents).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary

This feature enhances AI-powered compliance analysis with two major capabilities:

**Phase 1 - Context Injection**: Automatically inject framework metadata (name, version, identifier) and exact control/section excerpts into all AI analysis prompts. This grounds AI responses in precise control language, improving accuracy and confidence. Includes mandatory PII/secret redaction, prompt/response caching, and fallback to heuristics on failures.

**Phase 2 - Autonomous Evidence Collection**: AI generates evidence collection plans specifying sources (GitHub, Jira, AWS, etc.), queries/filters, and estimated signal strength. Users approve plans (with auto-approve policy support), system executes via existing MCP connectors, normalizes data, and runs context mode analysis with collected evidence.

**Technical Approach**: Build on existing 002-ai-evidence-analysis provider abstraction. Add ContextPreamble and EvidencePlan types, extend provider interface with ProposePlan method, implement source-to-query pattern matching for auto-approval, integrate redaction pipeline pre-prompt, add digest-based caching layer, create TUI panels for context preview and plan approval, support three-tier RBAC (Admin/Analyst/Viewer).

## Technical Context
**Language/Version**: Go 1.23+ (per existing project standard)  
**Primary Dependencies**: 
- Cobra (CLI framework)
- Viper (configuration management)
- Bubble Tea + Lip Gloss + Bubbles (TUI framework)
- Existing AI provider abstraction (Anthropic/OpenAI from 002)
- Existing MCP connectors (GitHub, Jira, Slack, AWS, CI/CD, docs)
- Redaction library (TBD in research - likely regex-based with denylist)

**Storage**: 
- File-based state (existing store package)
- Cache: digest-keyed prompt/response storage (local files)
- RedactionMap: in-memory only (never persisted)
- Config: YAML via Viper (AI config section for budgets, auto-approve policies)

**Testing**: Go standard testing (`go test ./...`), golden file tests for TUI output, integration tests for E2E workflows  
**Target Platform**: macOS, Linux, Windows (cross-platform CLI)
**Project Type**: Single CLI project (extending existing sdek-cli)  
**Performance Goals**: 
- Context mode: <30s for typical evidence bundles (<100 events)
- Autonomous mode: <5min for plans with ≤10 sources
- Cache hits: <100ms response time
- Cold start: maintain existing <100ms target

**Constraints**: 
- Zero PII/secret leakage to AI providers (mandatory redaction)
- Configurable budget limits (default: 50 sources, 500 API calls, 250K tokens)
- Redaction must not remove >40% of content (warn if exceeded)
- Support up to 25 concurrent analyses (configurable)
- Deterministic Evidence Plans for same inputs

**Scale/Scope**: 
- Hundreds of framework controls (SOC2, ISO27001, PCI-DSS, etc.)
- Thousands of evidence events per analysis
- Multiple concurrent compliance managers/analysts
- Enterprise-scale MCP connector usage

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Correctness and Safety
- ✅ **Validation**: All AI config (budgets, auto-approve patterns) validated before use
- ✅ **Error Handling**: Typed errors with context (`fmt.Errorf`) for AI failures, redaction errors, MCP connector failures
- ✅ **No Panics**: Graceful degradation to heuristics mode on AI errors
- ✅ **Side Effects**: Redaction events, plan approvals, evidence collection logged via audit events

### II. Configuration Management
- ✅ **Viper Integration**: AI config (budgets, auto-approve, concurrency limits) in YAML, loaded via Viper
- ✅ **Precedence**: CLI flags (`--no-cache`, `--dry-run`, `--mode`) → env vars → config file → defaults
- ✅ **Standard Paths**: Auto-load from `$HOME/.sdek/config.yaml`

### III. Command Design (Cobra)
- ✅ **New Commands**: `sdek ai analyze`, `sdek ai plan` with descriptive help, examples
- ✅ **PreRun Hooks**: Validate framework + section excerpt exist, check AI provider availability
- ✅ **PostRun Hooks**: Log audit events, cache results

### IV. User Experience & Terminal UI (Bubble Tea)
- ✅ **New TUI Panels**: Context Preview (framework + excerpt), Autonomous Plan (approve/deny items)
- ✅ **Status Indicators**: Pills (green/yellow/red), mode badges (AI/heuristics), review-required badges
- ✅ **Non-Interactive Mode**: `--dry-run` for autonomous plan preview without execution
- ✅ **Keyboard Shortcuts**: Standard navigation, `a` approve, `d` deny, `Enter` confirm

### V. Test-Driven Development
- ✅ **Unit Tests**: ContextPreamble builder, redaction pipeline, cache key generation, auto-approve matching
- ✅ **Integration Tests**: E2E context mode, E2E autonomous mode with MCP mocks
- ✅ **Golden File Tests**: TUI panel rendering (context preview, plan approval)
- ✅ **Regression Tests**: Redaction coverage, cache hit/miss, confidence threshold flagging

### VI. Performance & Efficiency
- ✅ **Fast Start**: No new startup overhead (lazy-load AI providers)
- ✅ **Caching**: Digest-based prompt/response cache reduces redundant AI calls
- ✅ **Concurrency**: Configurable limit (default 25) prevents resource exhaustion
- ✅ **Budget Controls**: Max sources, API calls, tokens prevent runaway costs

### VII. Cross-Platform Compatibility
- ✅ **Path Safety**: All file paths use `filepath.Join`
- ✅ **Terminal Compatibility**: Bubble Tea/Lip Gloss already handle xterm, iTerm2, Windows Terminal
- ✅ **Config Paths**: Use `os.UserConfigDir()` for cross-platform config discovery

### VIII. Observability & Logging
- ✅ **Verbose Flag**: `--verbose` enables debug logs for AI prompts (redacted), cache operations, MCP calls
- ✅ **Structured Logs**: Emit to stderr with levels (debug/info/warn/error)
- ✅ **Audit Events**: JSON telemetry for plan proposals, approvals, redactions, findings

### IX. Modularity & Code Organization
- ✅ **Structure**:
  - `cmd/analyze.go`, `cmd/ai_plan.go` — New commands (thin wrappers)
  - `internal/ai/context.go` — ContextPreamble builder
  - `internal/ai/plan.go` — EvidencePlan generation and execution
  - `internal/ai/redaction.go` — PII/secret redaction pipeline
  - `internal/ai/autoapprove.go` — Auto-approve policy matching
  - `internal/analyze/confidence.go` — Confidence threshold flagging (extend existing)
  - `ui/models/plan.go` — Autonomous Plan TUI model
  - `ui/components/context_preview.go` — Context Preview component
- ✅ **Thin Commands**: Business logic in `internal/ai`, commands delegate
- ✅ **No Cycles**: AI layer depends on store, policy, providers; UI depends on AI types

### X. Extensibility & Versioning
- ✅ **Backward Compatibility**: New flags/commands don't break existing workflows
- ✅ **Semantic Versioning**: Feature addition = MINOR bump
- ✅ **Extensibility**: Auto-approve policies extensible via YAML, budget limits configurable

### XI. Documentation & Clarity
- ✅ **Auto-Generated Help**: Cobra generates markdown for new commands
- ✅ **Examples**: Context mode and autonomous mode examples in command help
- ✅ **Design Docs**: This plan, data-model.md, contracts/, quickstart.md
- ✅ **Idiomatic Go**: Standard patterns, no cleverness

**Status**: ✅ PASS (Initial Check) — No constitutional violations detected. Design aligns with all principles.

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
cmd/
├── ai_analyze.go           # NEW: sdek ai analyze command
├── ai_plan.go              # NEW: sdek ai plan command
├── analyze.go              # EXTEND: Add --mode flag support
└── root.go                 # EXTEND: Register new ai subcommands

internal/
├── ai/
│   ├── context.go          # NEW: ContextPreamble builder
│   ├── context_test.go     # NEW: Unit tests
│   ├── plan.go             # NEW: EvidencePlan generation & execution
│   ├── plan_test.go        # NEW: Unit tests
│   ├── redaction.go        # NEW: PII/secret redaction pipeline
│   ├── redaction_test.go   # NEW: Unit tests
│   ├── autoapprove.go      # NEW: Auto-approve policy matching
│   ├── autoapprove_test.go # NEW: Unit tests
│   ├── engine.go           # EXTEND: Add ProposePlan method to interface
│   ├── cache.go            # EXTEND: Add digest-based cache keys
│   └── providers/
│       ├── anthropic.go    # EXTEND: Implement ProposePlan
│       └── openai.go       # EXTEND: Implement ProposePlan
├── analyze/
│   ├── confidence.go       # EXTEND: Add <0.6 threshold flagging
│   └── confidence_test.go  # EXTEND: Test review_required logic
├── config/
│   └── loader.go           # EXTEND: Load AI config (budgets, auto-approve, concurrency)
└── store/
    └── cache.go            # EXTEND: Support prompt/response caching

pkg/
└── types/
    ├── context.go          # NEW: ContextPreamble, AutoApprovePolicy types
    ├── plan.go             # NEW: EvidencePlan, PlanItem types
    ├── finding.go          # EXTEND: Add review_required field
    └── config.go           # EXTEND: Add AI config fields

ui/
├── models/
│   ├── plan.go             # NEW: Autonomous Plan TUI model
│   └── plan_test.go        # NEW: Unit tests
└── components/
    ├── context_preview.go  # NEW: Context Preview component
    └── plan_approval.go    # NEW: Plan approval list component

tests/
├── integration/
│   ├── analyze_ai_test.go  # EXTEND: Add context mode & autonomous mode tests
│   └── workflow_test.go    # EXTEND: Add E2E scenarios
└── golden/
    └── fixtures/
        ├── context_preview_*.txt   # NEW: Golden files for TUI
        └── plan_approval_*.txt     # NEW: Golden files for TUI

testdata/
└── ai/
    ├── policies/
    │   └── autoapprove_config.yaml # NEW: Test auto-approve policies
    └── fixtures/
        ├── context_preamble_*.json  # NEW: Test context preambles
        └── evidence_plan_*.json     # NEW: Test evidence plans

config.example.yaml         # EXTEND: Add AI config section with budgets, auto-approve
```

**Structure Decision**: Single CLI project (Option 1). Extending existing sdek-cli structure with new AI subcommands, internal packages for context/plan/redaction logic, TUI components for new panels, and comprehensive test coverage. Maintains constitutional separation of concerns: thin commands in `cmd/`, business logic in `internal/`, reusable types in `pkg/`, UI components in `ui/`.

## Phase 0: Outline & Research

**Unknowns Identified from Technical Context**:
1. Redaction library choice (PII/secret detection)
2. Cache digest algorithm (prompt/response caching)
3. Auto-approve pattern matching implementation
4. Confidence threshold flagging approach
5. Concurrent analysis semaphore design

**Research Tasks**:
1. **Redaction Library for Go**
   - Evaluate: Regex-based (stdlib), github.com/ggwhite/go-masker, custom solution
   - Requirements: PII (emails, IPs, phone), secrets (API keys, tokens), performance <10ms per event
   
2. **Cache Key Digest Algorithm**
   - Evaluate: SHA256, BLAKE3, xxHash
   - Requirements: Collision-resistant, fast (<1ms for 100KB evidence), deterministic
   
3. **Pattern Matching for Auto-Approve**
   - Evaluate: Glob patterns (github.com/gobwas/glob), regex (stdlib), simple prefix/suffix
   - Requirements: Support wildcards (`auth*`, `*login*`), case-insensitive, config-driven
   
4. **Confidence Threshold Best Practices**
   - Research: Industry standards for AI confidence scoring, typical thresholds (0.5? 0.6? 0.7?)
   - Decision: 0.6 threshold (from clarifications) aligns with "more likely correct than not"
   
5. **Concurrency Patterns in Go**
   - Evaluate: Semaphore (golang.org/x/sync/semaphore), worker pool, buffered channels
   - Requirements: Configurable limit (25 default), graceful degradation, context cancellation

**Consolidation Approach**:
- Document decisions, rationale, alternatives in `research.md`
- Link to Go package docs, benchmark results, security considerations
- Provide code snippets for proof-of-concept implementations

**Output**: research.md with all decisions documented

## Phase 1: Design & Contracts
*Prerequisites: research.md complete ✅*

**Approach**:
1. Extract key entities from spec → `data-model.md` (ContextPreamble, EvidencePlan, Finding extensions, RedactionMap, AutoApprovePolicy)
2. Define internal API contracts → `contracts/` (Analyze, ProposePlan, ExecutePlan, Redact, AutoApprove)
3. Generate quickstart walkthrough → `quickstart.md` (context mode + autonomous mode examples)
4. Update agent context file → `.github/copilot-instructions.md`

**Entities Identified** (from spec):
- ContextPreamble (framework, section, excerpt, rubrics)
- EvidencePlan (items: source, query, signal strength, approval status)
- Finding (extend with review_required field)
- RedactionMap (position/hash, placeholder, type, timestamp)
- AutoApprovePolicy (source -> patterns mapping)

**Internal APIs** (not REST/GraphQL, internal Go interfaces):
- `Analyze(ctx, preamble, evidence) -> (Finding, error)` — Context mode analysis
- `ProposePlan(ctx, preamble) -> (EvidencePlan, error)` — Generate evidence plan
- `ExecutePlan(ctx, plan) -> ([]EvidenceEvent, error)` — Execute approved plan items
- `Redact(text) -> (redacted, count, types)` — PII/secret redaction
- `AutoApprove(source, query) -> bool` — Match against policy

**Test Scenarios** (from user stories):
- Context mode with SOC2 CC6.1 → finding with confidence, risk, citations
- Autonomous mode with ISO 27001 A.9.4.2 → plan proposal, approval, execution, finding
- Redaction validation → PII/secrets removed, audit log emitted

**Outputs to generate**:
1. `data-model.md` — Entity schemas with Go struct examples
2. `contracts/engine-interface.md` — Extended AI engine interface
3. `contracts/redaction-interface.md` — Redaction pipeline interface
4. `contracts/autoapprove-interface.md` — Auto-approve matcher interface
5. `quickstart.md` — Step-by-step examples for both modes
6. `.github/copilot-instructions.md` — Updated agent context

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
The `/tasks` command will:
1. Load `.specify/templates/tasks-template.md` as task breakdown base
2. Generate tasks from Phase 1 design artifacts:
   - From `data-model.md`: Model creation tasks for ContextPreamble, EvidencePlan, AutoApprovePolicy, RedactionMap (5 entities)
   - From `contracts/`: Contract test tasks for Engine, Redactor, AutoApproveMatcher interfaces (3 contracts)
   - From `quickstart.md`: Integration test tasks for 6 scenarios (context mode, autonomous mode, dry-run, low confidence, fallback, concurrent)
3. Follow TDD ordering: Tests before implementation
4. Apply dependency ordering: Types → Redaction → Cache → Context → Plan → Commands → UI
5. Mark tasks with [P] for parallel execution when files have no dependencies

**Ordering Strategy**:
- **Models First** [P]: Create types in `pkg/types/` (context.go, plan.go, finding.go extensions)
- **Core Services**: Redaction → Cache → Context builder → Plan generator → Auto-approve matcher
- **Provider Extensions**: Extend Anthropic/OpenAI with ProposePlan, context-aware Analyze
- **Commands** [P]: New `ai_analyze.go`, `ai_plan.go` commands
- **TUI Components** [P]: Context preview, plan approval panels
- **Integration Tests**: E2E workflows (context mode, autonomous mode)
- **Golden Tests**: TUI output validation

**Estimated Output**: 25-30 numbered tasks in `tasks.md`

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning described (/plan command - approach documented)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented (none required)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
