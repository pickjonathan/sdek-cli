
# Implementation Plan: AI Evidence Analysis

**Branch**: `002-ai-evidence-analysis` | **Date**: 2025-10-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/Users/pickjonathan/WorkSpacePrivate/sdek-cli/specs/002-ai-evidence-analysis/spec.md`

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

This feature adds an AI-powered evidence analysis layer to enhance compliance control mapping. The system will use OpenAI or Anthropic via a unified abstraction to analyze events from five sources (Git, CI/CD, Jira, Slack, Docs) against compliance frameworks (SOC2, ISO 27001, PCI DSS), generating structured outputs with evidence references, justifications, confidence scores (0-100), and residual risk notes.

**Key Capabilities**:
- **Multi-provider AI abstraction**: Configurable OpenAI/Anthropic support with graceful fallback to deterministic heuristics
- **Privacy-first design**: Automatic PII/secret redaction, local-only processing, no external data persistence
- **Hybrid confidence scoring**: Weighted average (70% AI + 30% heuristic) for balanced accuracy
- **Event-driven caching**: Cache invalidation on event changes, 60-second AI timeout with fallback
- **Enhanced reporting**: AI metadata (justifications, confidence, residual risk) integrated into compliance reports

## Technical Context

**Language/Version**: Go 1.23+ (latest stable, per existing project)  
**Primary Dependencies**: 
- **AI SDKs**: OpenAI Go SDK (`github.com/sashabaranov/go-openai`), Anthropic SDK (`github.com/anthropics/anthropic-sdk-go`)
- **Existing**: Cobra v1.10.1 (commands), Viper v1.21.0 (config), UUID v1.6.0 (IDs)
- **Retry/Rate Limiting**: `github.com/cenkalti/backoff/v4` (exponential backoff), `golang.org/x/time/rate` (rate limiter)
- **Caching**: Built-in map + filesystem (JSON), invalidated on event changes

**Storage**: 
- **AI cache**: JSON files in user cache directory (`os.UserCacheDir()/sdek/ai-cache/`)
- **State integration**: Extends existing JSON state files with AI metadata fields
- **Local only**: No external persistence; all AI responses cached locally

**Testing**: 
- **Unit tests**: Go standard library `testing` package
- **Golden file tests**: Recorded AI response fixtures for deterministic testing
- **Mock providers**: Interface-based mocks for AI provider abstraction
- **Integration tests**: End-to-end analyze command with cached responses

**Target Platform**: Cross-platform CLI (Linux, macOS, Windows) - existing constraint  

**Project Type**: Single project (existing sdek-cli structure)  

**Performance Goals**: 
- **AI request timeout**: 60 seconds per control analysis
- **Overall analysis**: <60 seconds for 95% of typical workloads (100 events per control)
- **Cache hit rate**: >70% for repeated analyses on stable data
- **Cold start**: Maintain existing <100ms target (AI not invoked at startup)

**Constraints**: 
- **Privacy**: No PII/secrets transmitted to AI providers (GDPR/CCPA compliance)
- **Offline mode**: Deterministic heuristics must work without AI (test/CI environments)
- **Token limits**: AI inputs truncated/summarized to fit 4K-8K token limits
- **Reliability**: 100% uptime via fallback; AI failures must not block analysis
- **Auditability**: All AI requests/responses logged locally for compliance review

**Scale/Scope**: 
- **Controls**: 124 compliance controls (45 SOC2 + 64 ISO 27001 + 15 PCI DSS) - existing
- **Events**: Typical workload 100-500 events per analysis run across 5 sources
- **AI calls**: Batched per control (1 AI request per control with matching events)
- **Providers**: 2 initially (OpenAI, Anthropic), extensible for future providers
- **Cache size**: Unbounded initially; user-managed cleanup (future: LRU eviction)

**Additional Context from User**:
- Package layout: `internal/ai/engine.go` (interface + orchestration), `internal/ai/providers/{openai,anthropic}.go` (adapters)
- Analysis integration: `internal/analyze/mapper.go` (batching, fallback), `internal/policy/loader.go` (SOC2/ISO/PCI excerpts)
- Report extension: `internal/report/` enhanced with AI metadata fields (justification, confidence scores, residual risk)
- CLI flags: `--ai-provider=openai|anthropic|none`, config keys: `ai.provider`, `ai.model`, `ai.max_tokens`, `ai.temperature`, `ai.enabled`
- Security: Configurable field allowlist, exponential backoff with circuit breaker on repeated failures

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Pre-Research)

**I. Correctness and Safety**: ✅ PASS
- AI provider errors handled via fallback to heuristics (FR-004, FR-019)
- 60-second timeout prevents hanging (NFR-004)
- Typed errors for AI failures, cache errors, validation failures
- No panics; all errors logged with context

**II. Configuration Management**: ✅ PASS
- Viper remains single source of truth for `ai.*` config keys
- Precedence: CLI flags (`--ai-provider`) → env vars → config file → defaults
- Config keys: `ai.provider`, `ai.model`, `ai.max_tokens`, `ai.temperature`, `ai.enabled`
- Backward compatible: AI disabled by default; existing behavior unchanged

**III. Command Design (Cobra)**: ✅ PASS
- Extends existing `analyze` command with AI support (no new command)
- Flags: `--ai-provider` (openai|anthropic|none), `--no-ai` (disable)
- PreRun: Validate AI config, load policy excerpts, initialize provider
- PostRun: Log cache statistics, AI usage metrics

**IV. User Experience & Terminal UI (Bubble Tea)**: ✅ PASS
- Non-interactive mode for CI/CD: `--ai-provider=none` or `ai.enabled=false`
- AI analysis progress shown via Bubble Tea spinner (existing pattern)
- Fallback to heuristics silent (logged only); no user interruption
- Errors displayed clearly: "AI analysis failed, using heuristics" (yellow warning)

**V. Test-Driven Development**: ✅ PASS
- Unit tests for AI engine interface, provider adapters, privacy filters
- Integration tests for `analyze` command with AI enabled/disabled
- Golden file tests with recorded AI response fixtures
- Regression tests for fallback behavior, cache invalidation

**VI. Performance & Efficiency**: ✅ PASS
- Cold start unaffected: AI not invoked until `analyze` command runs
- Caching reduces redundant API calls (target >70% hit rate)
- 60-second timeout prevents indefinite blocking
- Exponential backoff for transient failures, circuit breaker for sustained failures

**VII. Cross-Platform Compatibility**: ✅ PASS
- AI SDKs support Linux, macOS, Windows
- Cache directory uses `os.UserCacheDir()` (cross-platform)
- Path operations use `filepath.Join` (existing pattern)

**VIII. Observability & Logging**: ✅ PASS
- Structured logs (`log/slog`) for AI requests, responses, failures, cache hits/misses
- Telemetry events: AI provider used, latency, token count, cache efficiency
- Audit logs: All AI requests/responses logged locally for compliance (NFR-006)
- `--verbose` enables debug logs for prompt templates, redaction stats

**IX. Modularity & Code Organization**: ✅ PASS
- New packages: `internal/ai/` (engine + providers), `internal/policy/` (loader)
- Existing packages extended: `internal/analyze/mapper.go`, `internal/report/`
- No cyclic dependencies: AI → types, analyze → AI, report → AI metadata
- Interface-based design: `ai.Engine` for provider abstraction, mockable for tests

**X. Extensibility & Versioning**: ✅ PASS
- AI provider abstraction supports future providers (AWS Bedrock, Azure OpenAI)
- Backward compatible: AI disabled by default, existing heuristics unchanged
- Semantic versioning: Minor bump (new feature, backward compatible)
- AI response schema versioned in cache (future-proof for format changes)

**XI. Documentation & Clarity**: ✅ PASS
- Auto-generated help for new flags (`--ai-provider`, `--no-ai`)
- Examples in README: enabling AI, switching providers, disabling for CI
- Design rationale documented: privacy-first, fallback strategy, weighted confidence
- Code comments: prompt templates, redaction patterns, cache invalidation triggers

### Violations Requiring Justification

**None identified**. This feature integrates cleanly within existing architecture:
- Extends existing `analyze` command (no new complexity)
- Uses established patterns (Viper config, interface-based design, golden file tests)
- Maintains simplicity via fallback (AI optional, heuristics sufficient)

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
internal/
├── ai/
│   ├── engine.go              # AI provider interface + orchestration
│   ├── cache.go               # Local filesystem cache with event-driven invalidation
│   ├── privacy.go             # PII/secret redaction filters
│   ├── prompt.go              # Templated prompt generation
│   ├── types.go               # Request/Response/Config types
│   └── providers/
│       ├── openai.go          # OpenAI adapter (implements Engine interface)
│       ├── anthropic.go       # Anthropic adapter (implements Engine interface)
│       └── mock.go            # Mock provider for tests
│
├── analyze/
│   ├── mapper.go              # EXTENDED: Batching, AI integration, fallback orchestration
│   ├── evidence.go            # EXTENDED: AI metadata fields (justification, scores, risk)
│   └── mapper_test.go         # EXTENDED: AI-enhanced mapping tests
│
├── policy/
│   ├── loader.go              # NEW: Load SOC2/ISO/PCI policy excerpts
│   ├── excerpts.go            # NEW: Policy text by control ID
│   └── loader_test.go         # NEW: Policy loading tests
│
├── report/
│   ├── exporter.go            # EXTENDED: Include AI metadata in reports
│   ├── formatter.go           # EXTENDED: Format AI justifications, confidence scores
│   └── exporter_test.go       # EXTENDED: AI metadata in report tests
│
├── types/
│   ├── evidence.go            # EXTENDED: Add AI-related fields
│   └── ai.go                  # NEW: AI-specific types (if not in internal/ai/types.go)
│
└── config/
    └── config.go              # EXTENDED: AI configuration keys

cmd/
└── analyze.go                 # EXTENDED: AI provider flags, AI-enabled analysis

testdata/
├── ai/
│   ├── fixtures/              # Golden file tests: recorded AI responses
│   │   ├── openai_response_soc2_cc1.1.json
│   │   ├── anthropic_response_iso_a5.1.json
│   │   └── ...
│   └── policies/              # Policy excerpt test data
│       ├── soc2_excerpts.json
│       └── iso27001_excerpts.json
│
└── analyze/
    └── ai_scenarios/          # Integration test scenarios with AI

tests/
├── ai_test.go                 # AI engine interface tests
├── privacy_test.go            # PII/secret redaction tests
└── integration/
    └── analyze_ai_test.go     # End-to-end analyze command with AI
```

**Structure Decision**: Single project structure (existing sdek-cli). This feature extends the existing `internal/` packages and adds new `internal/ai/` and `internal/policy/` packages for AI-specific logic. The `cmd/analyze.go` command is enhanced with AI flags and orchestration. Tests follow TDD principles with unit tests alongside implementation files and golden file fixtures in `testdata/ai/fixtures/`.

## Phase 0: Outline & Research ✅ COMPLETE

**Research Areas Investigated**:
1. **AI SDK Selection**: Evaluated Go SDKs for OpenAI and Anthropic
2. **Structured JSON Output**: Function calling vs tool use vs prompt engineering
3. **PII/Secret Redaction**: Regex patterns, hashing strategies, allowlists
4. **Caching Strategy**: Event-driven invalidation, content-based keys, filesystem storage
5. **Retry/Rate Limiting**: Exponential backoff, circuit breaker, rate limiter patterns
6. **Prompt Engineering**: System+user prompt structure, schema enforcement, truncation

**Key Decisions**:
- **OpenAI SDK**: `github.com/sashabaranov/go-openai` (community standard)
- **Anthropic SDK**: `github.com/anthropics/anthropic-sdk-go` (official)
- **Structured Output**: Function calling (OpenAI) + tool use (Anthropic) for reliable JSON
- **Redaction**: Regex-based detection with placeholders (preserves context)
- **Cache**: Event-driven invalidation with SHA256 content keys
- **Retry**: `github.com/cenkalti/backoff/v4` + circuit breaker + `golang.org/x/time/rate`
- **Prompts**: System prompt for role + JSON schema, user prompt for context

**Output**: ✅ `research.md` generated with 6 research decisions documented

## Phase 1: Design & Contracts ✅ COMPLETE

**Data Model Designed**:
- ✅ 6 core entities: AIConfig, AnalysisRequest, AnalysisResponse, CachedResult, Enhanced Evidence, PrivacyFilter
- ✅ Entity relationships diagram: AI Engine → Providers → Cache → Enhanced Evidence → Report
- ✅ State transitions: Cache lifecycle (create → read → invalidate → delete), AI analysis flow (check cache → call AI → fallback)
- ✅ Validation rules: Input validation (before AI), output validation (after AI), cache validation (on load)

**API Contracts Generated**:
- ✅ `contracts/engine-interface.md`: Core `ai.Engine` interface with Analyze(), Provider(), Health() methods
- ✅ Contract specifications: Input/output schemas, error types, retry policies, timeout behavior
- ✅ Test contract: 7 required tests for all Engine implementations (success, timeout, auth failure, invalid request, health check)
- ✅ Thread safety requirement: All implementations must be safe for concurrent use

**Test Scenarios Extracted**:
- ✅ `quickstart.md`: 8 integration test scenarios
  1. Enable AI with OpenAI (success path)
  2. AI fallback on provider error (resilience)
  3. Cache reuse on repeated analysis (performance)
  4. Cache invalidation on event change (correctness)
  5. Switch AI providers mid-stream (flexibility)
  6. Disable AI for CI/CD (offline mode)
  7. PII redaction before AI transmission (privacy)
  8. AI timeout and fallback (reliability)
- ✅ Integration test automation guidance: Test fixtures, manual validation checklist, rollback plan

**Agent Context Updated**:
- ✅ Ran `.specify/scripts/bash/update-agent-context.sh copilot`
- ✅ Updated `.github/copilot-instructions.md` with Go 1.23+ context
- ✅ Preserved existing manual additions between markers
- ✅ Added recent change for feature 002

**Output**: ✅ data-model.md, contracts/engine-interface.md, quickstart.md, .github/copilot-instructions.md updated

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:

1. **Foundation Tasks** (from data-model.md):
   - Create `internal/ai/types.go` with AIConfig, AnalysisRequest, AnalysisResponse, CachedResult structs [P]
   - Create `internal/ai/privacy.go` with PrivacyFilter and redaction logic [P]
   - Extend `internal/types/evidence.go` with AI metadata fields (AIAnalyzed, AIJustification, etc.) [P]
   - Create `internal/policy/loader.go` with policy excerpt loading for SOC2/ISO/PCI [P]

2. **Contract Test Tasks** (from contracts/engine-interface.md):
   - Create `internal/ai/engine_test.go` with 7 contract tests (TestEngine_AnalyzeSuccess, TestEngine_AnalyzeTimeout, etc.)
   - Create `internal/ai/providers/mock.go` with mock provider for testing [P]
   - Each test must fail initially (no implementation)

3. **Core Implementation Tasks** (from engine-interface.md):
   - Create `internal/ai/engine.go` with Engine interface definition [P]
   - Create `internal/ai/providers/openai.go` implementing Engine with go-openai SDK
   - Create `internal/ai/providers/anthropic.go` implementing Engine with anthropic-sdk-go
   - Create `internal/ai/cache.go` with filesystem caching and event-driven invalidation
   - Create `internal/ai/prompt.go` with system+user prompt template generation

4. **Integration Tasks** (from data-model.md relationships):
   - Extend `internal/analyze/mapper.go` to integrate AI analysis (batch controls, call Engine, fallback logic)
   - Extend `internal/analyze/evidence.go` to combine AI + heuristic confidence scores (70/30 weighted average)
   - Extend `internal/report/exporter.go` to include AI metadata fields in JSON output
   - Extend `internal/report/formatter.go` to format AI justifications and residual risk notes

5. **CLI Integration Tasks** (from Technical Context):
   - Extend `cmd/analyze.go` with `--ai-provider` flag and AI orchestration in PreRun/PostRun hooks
   - Extend `internal/config/config.go` to load ai.* Viper keys and validate AIConfig
   - Add AI configuration validation in analyze command PreRun (check API keys, test Health())

6. **Integration Test Tasks** (from quickstart.md scenarios):
   - Create `tests/integration/analyze_ai_test.go` with 8 scenario tests
   - Create `testdata/ai/fixtures/` with golden file responses for OpenAI and Anthropic
   - Create `testdata/ai/policies/` with SOC2/ISO/PCI policy excerpt fixtures

7. **Privacy Test Tasks** (from quickstart.md scenario 7):
   - Create `internal/ai/privacy_test.go` with redaction tests (email, phone, API key, credit card, SSN patterns)
   - Test preserving structure (keep domain for emails, last 4 for cards)
   - Test configurable allowlist (timestamps, status codes, log levels not redacted)

**Ordering Strategy**:

**Phase A: Foundation (parallel)**
- Task 1: Create ai/types.go [P]
- Task 2: Create ai/privacy.go [P]
- Task 3: Extend types/evidence.go [P]
- Task 4: Create policy/loader.go [P]
- Task 5: Create ai/engine.go interface [P]

**Phase B: Tests First (TDD)**
- Task 6: Create engine_test.go with contract tests (failing)
- Task 7: Create privacy_test.go with redaction tests (failing)
- Task 8: Create mock provider

**Phase C: Provider Implementations (parallel after Phase B)**
- Task 9: Implement openai.go (make tests pass) [P]
- Task 10: Implement anthropic.go (make tests pass) [P]

**Phase D: Supporting Services**
- Task 11: Implement cache.go (event-driven invalidation)
- Task 12: Implement prompt.go (template generation)
- Task 13: Implement privacy.go redaction logic (make tests pass)
- Task 14: Implement policy/loader.go (load excerpts)

**Phase E: Integration**
- Task 15: Extend analyze/mapper.go (AI integration)
- Task 16: Extend analyze/evidence.go (hybrid confidence scoring)
- Task 17: Extend report/exporter.go (AI metadata)
- Task 18: Extend report/formatter.go (AI formatting)
- Task 19: Extend cmd/analyze.go (AI flags, orchestration)
- Task 20: Extend config/config.go (AI config keys)

**Phase F: Integration Tests**
- Task 21: Create integration test file
- Task 22-29: Implement 8 quickstart scenarios
- Task 30: Create golden file fixtures

**Estimated Output**: 30-35 numbered, ordered tasks in tasks.md

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
- [x] Phase 0: Research complete (/plan command) ✅
- [x] Phase 1: Design complete (/plan command) ✅
- [x] Phase 2: Task planning complete (/plan command - describe approach only) ✅
- [ ] Phase 3: Tasks generated (/tasks command) - **READY TO EXECUTE**
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS ✅
- [x] Post-Design Constitution Check: PASS ✅ (no violations, clean integration)
- [x] All NEEDS CLARIFICATION resolved ✅ (5 clarifications completed in /clarify phase)
- [x] Complexity deviations documented ✅ (none identified)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
