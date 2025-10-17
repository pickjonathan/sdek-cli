# Tasks: AI Evidence Analysis

**Input**: Design documents from `/Users/pickjonathan/WorkSpacePrivate/sdek-cli/specs/002-ai-evidence-analysis/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/, quickstart.md

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Extract: Go 1.23+, AI SDKs (go-openai, anthropic-sdk-go), existing project structure
2. Load optional design documents:
   → data-model.md: 6 entities (AIConfig, AnalysisRequest, AnalysisResponse, CachedResult, Enhanced Evidence, PrivacyFilter)
   → contracts/engine-interface.md: ai.Engine interface with 3 methods, 7 contract tests
   → research.md: 4 decisions (SDK selection, structured JSON, PII redaction, caching strategy)
   → quickstart.md: 8 integration test scenarios
3. Generate tasks by category:
   → Setup: Dependencies, linting (2 tasks)
   → Tests: Contract tests, privacy tests (8 tasks - TDD)
   → Core: Interface, types, providers, cache, privacy, policy, prompt (15 tasks)
   → Integration: Extend analyze, report, config, cmd (7 tasks)
   → Polish: Integration tests, fixtures, docs (8 tasks)
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001-T040)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All contracts have tests? YES (7 tests in engine_test.go)
   → All entities have types? YES (types.go covers 6 entities)
   → All scenarios tested? YES (8 quickstart scenarios)
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- Single project structure at repository root
- `internal/` for implementation packages
- `cmd/` for CLI commands
- `tests/integration/` for integration tests
- `testdata/` for test fixtures

---

## Phase 3.1: Setup

- [x] T001 Install AI SDK dependencies (go-openai, anthropic-sdk-go, backoff, rate limiter)
- [x] T002 [P] Update go.mod and verify Go 1.23+ compatibility

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

- [x] T003 [P] Contract test Engine.Analyze() success in `internal/ai/engine_test.go`
- [x] T004 [P] Contract test Engine.Analyze() timeout in `internal/ai/engine_test.go`
- [x] T005 [P] Contract test Engine.Analyze() invalid request in `internal/ai/engine_test.go`
- [x] T006 [P] Contract test Engine.Analyze() auth failure in `internal/ai/engine_test.go`
- [x] T007 [P] Contract test Engine.Provider() in `internal/ai/engine_test.go`
- [x] T008 [P] Contract test Engine.Health() success in `internal/ai/engine_test.go`
- [x] T009 [P] Contract test Engine.Health() auth failure in `internal/ai/engine_test.go`
- [x] T010 [P] Privacy redaction tests (email, phone, API key, credit card, SSN) in `internal/ai/privacy_test.go`

## Phase 3.3: Core Implementation - Foundation (ONLY after tests are failing)

- [x] T011 [P] Create `internal/ai/types.go` with AIConfig, AnalysisRequest, AnalysisResponse, CachedResult structs
- [x] T012 [P] Create `internal/ai/engine.go` with Engine interface definition (Analyze, Provider, Health methods)
- [x] T013 [P] Create `internal/ai/errors.go` with error types (ErrInvalidRequest, ErrProviderTimeout, ErrProviderAuth, etc.)
- [x] T014 [P] Extend `pkg/types/evidence.go` with AI metadata fields (AIAnalyzed, AIJustification, AIConfidence, AIResidualRisk, HeuristicConfidence, CombinedConfidence, AnalysisMethod)

## Phase 3.4: Core Implementation - Privacy & Policy

- [x] T015 Create `internal/ai/privacy.go` with PrivacyFilter struct and redaction methods (implement to pass T010)
- [x] T016 [P] Create `internal/policy/loader.go` with policy excerpt loading for SOC2/ISO27001/PCI-DSS
- [x] T017 [P] Create `internal/policy/excerpts.go` with policy text data by control ID
- [x] T018 [P] Create `internal/policy/loader_test.go` with policy loading tests

## Phase 3.5: Core Implementation - Mock Provider

- [x] T019 Create `internal/ai/providers/mock.go` with mock Engine implementation (for test support)

## Phase 3.6: Core Implementation - Real Providers

- [x] T020 Create `internal/ai/providers/openai.go` implementing Engine with go-openai SDK (implement to pass T003-T009)
- [x] T021 Create `internal/ai/providers/anthropic.go` implementing Engine with anthropic-sdk-go SDK (implement to pass T003-T009)

## Phase 3.7: Core Implementation - Supporting Services

- [x] T022 Create `internal/ai/cache.go` with filesystem caching and event-driven invalidation
- [x] T023 [P] Create `internal/ai/cache_test.go` with cache tests (hit, miss, invalidation)
- [x] T024 Create `internal/ai/prompt.go` with system+user prompt template generation
- [x] T025 [P] Create `internal/ai/prompt_test.go` with prompt generation tests

## Phase 3.8: Integration - Analyze Command

- [x] T026 Extend `internal/analyze/mapper.go` to integrate AI analysis (batch controls, call Engine, fallback orchestration, weighted confidence scoring)
- [x] T027 Extend `internal/analyze/mapper_test.go` with AI-enhanced mapping tests
- [x] T028 Extend `internal/analyze/confidence.go` to compute weighted average (70% AI + 30% heuristic)
- [x] T029 Extend `internal/analyze/confidence_test.go` with hybrid confidence tests

## Phase 3.9: Integration - Report Command

- [x] T030 Extend `internal/report/exporter.go` to include AI metadata fields in JSON output
- [x] T031 Extend `internal/report/exporter_test.go` with AI metadata in report tests
- [x] T032 Extend `internal/report/formatter.go` to format AI justifications and residual risk notes
- [x] T033 Extend `internal/report/formatter_test.go` with AI formatting tests

## Phase 3.10: Integration - Configuration & CLI

- [ ] T034 Extend `internal/config/loader.go` to load ai.* Viper keys (ai.provider, ai.enabled, ai.model, ai.max_tokens, ai.temperature, ai.timeout, ai.rate_limit, ai.openai_key, ai.anthropic_key)
- [ ] T035 Extend `internal/config/validator.go` to validate AIConfig (provider values, token limits, temperature range, timeout range, API key presence)
- [ ] T036 Extend `cmd/analyze.go` with --ai-provider flag and AI orchestration in PreRun (Health check) and PostRun (cache statistics)

## Phase 3.11: Integration Tests (from quickstart.md scenarios)

- [ ] T037 [P] Create `tests/integration/analyze_ai_test.go` with scenario 1: Enable AI with OpenAI (success path)
- [ ] T038 [P] Create test scenario 2: AI fallback on provider error in `tests/integration/analyze_ai_test.go`
- [ ] T039 [P] Create test scenario 3: Cache reuse on repeated analysis in `tests/integration/analyze_ai_test.go`
- [ ] T040 [P] Create test scenario 4: Cache invalidation on event change in `tests/integration/analyze_ai_test.go`
- [ ] T041 [P] Create test scenario 5: Switch AI providers mid-stream in `tests/integration/analyze_ai_test.go`
- [ ] T042 [P] Create test scenario 6: Disable AI for CI/CD in `tests/integration/analyze_ai_test.go`
- [ ] T043 [P] Create test scenario 7: PII redaction before AI transmission in `tests/integration/analyze_ai_test.go`
- [ ] T044 [P] Create test scenario 8: AI timeout and fallback in `tests/integration/analyze_ai_test.go`

## Phase 3.12: Test Fixtures & Documentation

- [ ] T045 [P] Create `testdata/ai/fixtures/openai_response_soc2_cc1.1.json` (golden file for OpenAI)
- [ ] T046 [P] Create `testdata/ai/fixtures/anthropic_response_iso_a5.1.json` (golden file for Anthropic)
- [ ] T047 [P] Create `testdata/ai/policies/soc2_excerpts.json` (policy excerpt fixtures)
- [ ] T048 [P] Create `testdata/ai/policies/iso27001_excerpts.json` (policy excerpt fixtures)
- [ ] T049 [P] Create `testdata/ai/policies/pci_excerpts.json` (policy excerpt fixtures)
- [ ] T050 [P] Create `testdata/events_with_pii.json` (events for redaction testing)
- [ ] T051 Update `README.md` with AI configuration examples (enabling AI, switching providers, disabling for CI)
- [ ] T052 Update `docs/commands.md` with --ai-provider flag documentation

---

## Dependencies

**Sequential Dependencies**:
- Setup (T001-T002) before all other tasks
- Tests (T003-T010) before implementation (T011-T052)
- Foundation (T011-T014) before Privacy & Policy (T015-T018)
- Foundation (T011-T014) before Mock Provider (T019)
- Foundation (T011-T014) + Mock (T019) before Real Providers (T020-T021)
- Providers (T020-T021) before Supporting Services (T022-T025)
- Supporting Services (T022-T025) before Analyze Integration (T026-T029)
- Analyze Integration (T026-T029) before Report Integration (T030-T033)
- All core implementation (T011-T033) before Configuration & CLI (T034-T036)
- All implementation (T011-T036) before Integration Tests (T037-T044)
- Integration Tests (T037-T044) before Test Fixtures (T045-T050)
- All implementation before Documentation (T051-T052)

**File-Based Conflicts** (same file, must be sequential):
- T003-T009: All in `internal/ai/engine_test.go` (sequential within group, but [P] relative to other files)
- T026-T027: `internal/analyze/mapper.go` and its test (T026 before T027)
- T028-T029: `internal/analyze/confidence.go` and its test (T028 before T029)
- T030-T031: `internal/report/exporter.go` and its test (T030 before T031)
- T032-T033: `internal/report/formatter.go` and its test (T032 before T033)
- T034-T035: `internal/config/loader.go` and validator.go (can be parallel [P])
- T037-T044: All in `tests/integration/analyze_ai_test.go` (sequential within group)

**Parallel Execution Groups**:
- **Group A (Setup)**: T001, T002 [P]
- **Group B (Contract Tests)**: T003-T009 can all be written in parallel if using separate test functions [P]
- **Group C (Foundation)**: T011, T012, T013, T014 [P]
- **Group D (Privacy & Policy)**: T016, T017, T018 [P] (T015 depends on T010 passing)
- **Group E (Providers)**: T020, T021 [P] (after T019 completes)
- **Group F (Supporting Services Tests)**: T023, T025 [P] (after T022, T024 complete)
- **Group G (Integration Tests)**: T037-T044 [P] (different test scenarios in same file, but can be written in parallel)
- **Group H (Test Fixtures)**: T045-T050 [P]
- **Group I (Documentation)**: T051, T052 [P]

---

## Parallel Execution Examples

### Example 1: Foundation Tasks
```bash
# Launch T011-T014 together (different files, no dependencies):
Task: "Create internal/ai/types.go with AIConfig, AnalysisRequest, AnalysisResponse, CachedResult structs"
Task: "Create internal/ai/engine.go with Engine interface definition"
Task: "Create internal/ai/errors.go with error types"
Task: "Extend pkg/types/evidence.go with AI metadata fields"
```

### Example 2: Provider Implementations
```bash
# Launch T020-T021 together (different files, both implement same interface):
Task: "Create internal/ai/providers/openai.go implementing Engine with go-openai SDK"
Task: "Create internal/ai/providers/anthropic.go implementing Engine with anthropic-sdk-go SDK"
```

### Example 3: Integration Tests
```bash
# Launch T037-T044 together (different test scenarios):
Task: "Create tests/integration/analyze_ai_test.go with scenario 1: Enable AI with OpenAI"
Task: "Create test scenario 2: AI fallback on provider error"
Task: "Create test scenario 3: Cache reuse on repeated analysis"
Task: "Create test scenario 4: Cache invalidation on event change"
Task: "Create test scenario 5: Switch AI providers mid-stream"
Task: "Create test scenario 6: Disable AI for CI/CD"
Task: "Create test scenario 7: PII redaction before AI transmission"
Task: "Create test scenario 8: AI timeout and fallback"
```

### Example 4: Test Fixtures
```bash
# Launch T045-T050 together (independent JSON files):
Task: "Create testdata/ai/fixtures/openai_response_soc2_cc1.1.json"
Task: "Create testdata/ai/fixtures/anthropic_response_iso_a5.1.json"
Task: "Create testdata/ai/policies/soc2_excerpts.json"
Task: "Create testdata/ai/policies/iso27001_excerpts.json"
Task: "Create testdata/ai/policies/pci_excerpts.json"
Task: "Create testdata/events_with_pii.json"
```

---

## Notes

- **[P] tasks**: Different files, no dependencies - can run in parallel
- **Verify tests fail**: All T003-T010 must fail before implementing T011-T036
- **Commit after each task**: Use Git to track progress
- **Avoid**: Vague tasks, same file conflicts, skipping TDD

## Task Generation Rules
*Applied during main() execution*

1. **From Contracts** (contracts/engine-interface.md):
   - ai.Engine interface → 7 contract tests (T003-T009) [P]
   - Analyze(), Provider(), Health() → 3 implementation tasks (T020, T021, T012)
   
2. **From Data Model** (data-model.md):
   - 6 entities → T011 (types), T014 (evidence extension), T015 (privacy)
   - Relationships → T022 (cache), T024 (prompt), T026 (mapper integration)
   
3. **From User Stories** (quickstart.md):
   - 8 scenarios → 8 integration tests (T037-T044) [P]
   - Validation steps → fixture creation (T045-T050) [P]

4. **Ordering**:
   - Setup (T001-T002) → Tests (T003-T010) → Foundation (T011-T014) → Privacy/Policy (T015-T018) → Mock (T019) → Providers (T020-T021) → Services (T022-T025) → Integration (T026-T036) → Integration Tests (T037-T044) → Fixtures (T045-T050) → Docs (T051-T052)

## Validation Checklist
*GATE: Checked before marking tasks complete*

- [x] All contracts have corresponding tests (7 tests for Engine interface)
- [x] All entities have type definitions (6 entities in types.go + evidence.go)
- [x] All tests come before implementation (T003-T010 before T011+)
- [x] Parallel tasks truly independent (marked [P] only when different files)
- [x] Each task specifies exact file path (all tasks include file paths)
- [x] No task modifies same file as another [P] task (conflicts documented in Dependencies)
- [x] Integration test scenarios cover all quickstart examples (8 scenarios → T037-T044)
- [x] Privacy requirements testable (T010 for redaction, T043 for end-to-end)

---

## Implementation Guidelines

**TDD Approach**:
1. Write failing test (T003-T010)
2. Implement minimal code to pass test (T011-T036)
3. Refactor if needed
4. Repeat for each task

**AI Provider Implementation**:
- Use structured JSON output (function calling for OpenAI, tool use for Anthropic)
- Set temperature=0.3 for deterministic analysis
- Implement exponential backoff with `github.com/cenkalti/backoff/v4`
- Rate limit with `golang.org/x/time/rate`

**Privacy Requirements**:
- Redact email, phone, API key, credit card, SSN patterns
- Preserve text structure (e.g., "User <EMAIL_REDACTED> created ticket")
- Hash sensitive IDs for correlation
- Configurable allowlist for safe fields

**Caching Strategy**:
- Cache key: SHA256 hash of control ID + event IDs + policy excerpt
- Invalidation: When any referenced event is added/modified/deleted
- Storage: `~/.cache/sdek/ai-cache/{cache_key}.json`
- Format: JSON with CachedResult structure

**Confidence Scoring**:
- Heuristic-only: Use existing keyword-based score
- AI-only: Use AI confidence score
- Hybrid: 70% AI + 30% heuristic weighted average
- Combined confidence replaces individual scores in final evidence

**Error Handling**:
- Retryable errors: Timeout, rate limit, 5xx (max 3 retries with backoff)
- Non-retryable errors: Auth failure, invalid JSON, quota exceeded (fail fast)
- Fallback: Use heuristics if AI fails, log error, continue analysis
- No panics: All errors logged with context

**Configuration**:
- Keys: `ai.provider` (openai|anthropic|none), `ai.enabled` (bool), `ai.model` (string), `ai.max_tokens` (int), `ai.temperature` (float), `ai.timeout` (int), `ai.rate_limit` (int)
- Env vars: `SDEK_AI_OPENAI_KEY`, `SDEK_AI_ANTHROPIC_KEY`
- CLI flags: `--ai-provider`, `--no-ai`
- Validation: Provider values, token limits (0-32768), temperature (0.0-1.0), timeout (0-300s)

---

*Generated from specs/002-ai-evidence-analysis/ on 2025-10-16*
*Total tasks: 52 (Setup: 2, Tests: 8, Core: 15, Integration: 12, Polish: 15)*
*Estimated parallel groups: 9 (A-I)*
*Ready for execution via /implement command*
