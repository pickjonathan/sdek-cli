# Tasks: MCP-Native Agent Orchestrator & Tooling Config

**Input**: Design documents from `/Users/pickjonathan/WorkSpacePrivate/sdek-cli/specs/004-mcp-native-agent/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Execution Flow (main)
```
1. Load plan.md from feature directory ✅
   → Tech stack: Go 1.23+, Cobra, Viper, Bubble Tea, Lip Gloss, fsnotify, JSON Schema validator
   → Structure: Single project (CLI tool), paths at repository root
2. Load design documents ✅
   → data-model.md: 7 entities (MCPConfig, MCPTool, MCPInvocationLog, AgentCapability, ToolBudget, MCPHealthReport, Evidence)
   → contracts/: 4 interface files (registry, transport, RBAC, schema)
   → quickstart.md: 6 validation scenarios (AC-01 to AC-06)
3. Generate tasks by category ✅
   → Setup: Dependencies, project structure, schema
   → Tests: Contract tests, unit tests, integration tests
   → Core: Types, validator, loader, transports, registry, RBAC, evidence
   → Integration: Redaction, caching, CLI, TUI
   → Polish: Documentation, examples, golden files
4. Apply task rules ✅
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001-T064) ✅
6. Task completeness validated ✅
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[M1-M5]**: Milestone markers for progress tracking
- Include exact file paths in descriptions

## Path Conventions
Single project structure (per plan.md):
- `pkg/types/` - Public types and interfaces
- `internal/mcp/` - MCP implementation
- `internal/mcp/transport/` - Transport implementations
- `internal/mcp/rbac/` - RBAC enforcement
- `cmd/` - CLI commands
- `ui/models/` and `ui/components/` - TUI components
- `tests/unit/`, `tests/integration/`, `tests/golden/` - Test files
- `testdata/mcp/` - Test fixtures

---

## Phase 3.1: Setup & Foundation (M1)

- [X] **T001** [P] Install JSON Schema validator library (github.com/santhosh-tekuri/jsonschema/v5) and fsnotify in go.mod
- [X] **T002** [P] Copy MCP config JSON Schema from contracts/mcp-config-schema.json to internal/mcp/schema/config-schema.json
- [X] **T003** [P] Create pkg/types/mcp.go with MCPConfig, MCPTool, MCPInvocationLog, AgentCapability, ToolBudget, MCPHealthReport structs
- [X] **T004** [P] Create internal/mcp/errors.go with ErrToolNotFound, ErrToolDisabled, ErrInvalidConfig, ErrHandshakeFailed, ErrPermissionDenied error types

---

## Phase 3.2: Tests First (TDD - M1) ⚠️ MUST COMPLETE BEFORE 3.3

**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Schema & Validation Tests
- [X] **T005** [P] Contract test for JSON Schema validation in tests/unit/mcp_validator_test.go
  - Test valid config passes validation
  - Test missing required fields rejected with file/line/property errors
  - Test invalid transport type rejected
  - Test invalid capability format rejected
  - Covers FR-003, FR-005

- [X] **T006** [P] Unit test for config loader precedence in tests/unit/mcp_loader_test.go
  - Test project config (./.sdek/mcp/) overrides global (~/.sdek/mcp/)
  - Test env var (SDEK_MCP_PATH) precedence
  - Test config with same name - precedence wins
  - Covers FR-002

### Transport Tests
- [X] **T007** [P] Contract test for stdio transport in tests/unit/mcp_transport_stdio_test.go
  - Test JSON-RPC 2.0 request/response over stdin/stdout
  - Test handshake sequence
  - Test error handling for crashed processes
  - Covers Transport interface contract

- [X] **T008** [P] Contract test for HTTP transport in tests/unit/mcp_transport_http_test.go
  - Test JSON-RPC 2.0 over HTTP POST
  - Test handshake with base URL
  - Test timeout handling
  - Covers Transport interface contract

### Registry Tests
- [X] **T009** [P] Contract test for MCPRegistry.Init in tests/unit/mcp_registry_test.go
  - Test discovery of configs from multiple paths
  - Test async handshake initialization
  - Test returns count of successful initializations
  - Covers Registry interface contract methods: Init

- [X] **T010** [P] Contract test for MCPRegistry lifecycle in tests/unit/mcp_registry_lifecycle_test.go
  - Test Close() waits for in-flight invocations
  - Test Reload() hot-reloads changed configs
  - Test List() returns all tools with status
  - Test Get() retrieves specific tool
  - Covers Registry interface contract methods: Close, Reload, List, Get

- [X] **T011** [P] Contract test for MCPRegistry admin operations in tests/unit/mcp_registry_admin_test.go
  - Test Enable() transitions tool to ready
  - Test Disable() transitions tool to offline
  - Test disabled tool rejects invocations
  - Covers Registry interface contract methods: Enable, Disable

- [X] **T012** [P] Contract test for MCPRegistry validation in tests/unit/mcp_registry_validate_test.go
  - Test Validate() checks schema errors
  - Test Test() performs health check
  - Covers Registry interface contract methods: Validate, Test

### RBAC Tests
- [X] **T013** [P] Contract test for RBACEnforcer.CheckPermission in tests/unit/mcp_rbac_test.go
  - Test capability matching (exact match)
  - Test wildcard capabilities (tool.*)
  - Test permission denied for missing capability
  - Covers RBAC interface contract method: CheckPermission, FR-013, FR-015

- [X] **T014** [P] Contract test for RBACEnforcer budget enforcement in tests/unit/mcp_budgets_test.go
  - Test rate limit enforcement (requests per second)
  - Test concurrency limit enforcement
  - Test timeout enforcement
  - Covers RBAC interface contract methods: ApplyBudget, FR-014

- [X] **T015** [P] Contract test for audit logging in tests/unit/mcp_audit_test.go
  - Test invocation log created with all required fields
  - Test args hashed (SHA256)
  - Test redaction flag set correctly
  - Covers FR-020, FR-021, FR-022

### Integration Tests
- [X] **T016** [P] Integration test for handshake with mock MCP server in tests/integration/mcp_handshake_test.go
  - Set up mock stdio MCP server (using testdata/mcp/mock_server/)
  - Test successful handshake
  - Test handshake failure handling
  - Covers FR-007, Scenario 1 (AC-01)

- [X] **T017** [P] Integration test for hot-reload in tests/integration/mcp_hotreload_test.go
  - Create config file, verify tool loaded
  - Modify config file, verify tool reloaded
  - Delete config file, verify tool removed
  - Covers FR-004, FR-011

- [X] **T018** [P] Integration test for RBAC enforcement end-to-end in tests/integration/mcp_rbac_test.go
  - Agent with permission invokes tool → success
  - Agent without permission invokes tool → permission denied
  - Verify audit log created
  - Covers Scenario 4 (AC-04), FR-013, FR-015

- [X] **T019** [P] Integration test for circuit breaker and retry in tests/integration/mcp_resilience_test.go
  - Tool fails N times → transitions to degraded
  - Circuit breaker opens → tool goes offline
  - Circuit breaker half-open → test success → tool returns to ready
  - Covers FR-008, FR-009, FR-010, Scenario 3 (AC-03)

- [X] **T020** [P] Integration test for evidence collection via MCP in tests/integration/mcp_evidence_test.go
  - Run analysis that calls MCP tool
  - Verify evidence normalized into evidence graph
  - Verify redaction applied
  - Verify audit log created
  - Covers Scenario 5 (AC-05), FR-016, FR-017, FR-018, FR-019

---

## Phase 3.3: Core Implementation (M1-M2) - ONLY after tests are failing

### Types & Schema (M1)
- [X] **T021** Implement MCPConfig struct validation methods in pkg/types/mcp.go
  - Implement Validate() method
  - Add field validation rules from data-model.md
  - Tests: T005 should pass
  - Covers FR-003

- [X] **T022** Implement MCPTool struct and ToolMetrics in pkg/types/mcp.go
  - Add Status enum (Ready, Degraded, Offline)
  - Add ToolMetrics fields
  - Add state transition methods
  - Tests: T009, T010 should start passing
  - Covers data-model.md entity definitions

- [X] **T023** [P] Implement MCPInvocationLog struct in pkg/types/mcp.go
  - Add all required fields
  - Add NewInvocationLog constructor with auto-generated ID/timestamp
  - Tests: T015 should pass
  - Covers FR-020

- [X] **T024** [P] Implement AgentCapability and ToolBudget structs in pkg/types/mcp.go
  - Add capability string validation
  - Add rate limit and budget fields
  - Tests: T013, T014 should start passing
  - Covers FR-012, FR-014

### Validator & Loader (M1)
- [X] **T025** Implement JSON Schema validator in internal/mcp/validator.go
  - Load schema from internal/mcp/schema/config-schema.json
  - Implement Validate(configPath) method
  - Return detailed SchemaError with file/line/property paths
  - Tests: T005 should fully pass
  - Covers FR-003, FR-005

- [X] **T026** Implement config loader with precedence in internal/mcp/loader.go
  - Implement LoadConfigs() with path precedence (./.sdek/mcp/, ~/.sdek/mcp/, SDEK_MCP_PATH)
  - Handle environment variable expansion (e.g., ${GITHUB_TOKEN})
  - Use filepath.Join for cross-platform paths
  - Tests: T006 should fully pass
  - Covers FR-001, FR-002

### Transport Layer (M1-M2)
- [X] **T027** Define Transport interface in internal/mcp/transport/transport.go
  - Define Invoke(method, args) method
  - Define HealthCheck() method
  - Define Close() method
  - Add transport type enum (Stdio, HTTP)

- [X] **T028** [P] Implement stdio transport in internal/mcp/transport/stdio.go
  - Implement JSON-RPC 2.0 over stdin/stdout
  - Start subprocess with command + args
  - Handle process crashes
  - Tests: T007 should fully pass
  - Covers data-model.md Transport field

- [X] **T029** [P] Implement HTTP transport in internal/mcp/transport/http.go
  - Implement JSON-RPC 2.0 over HTTP POST
  - Use BaseURL from config
  - Handle connection timeouts
  - Tests: T008 should fully pass
  - Covers data-model.md Transport field

### Registry & Orchestrator (M2)
- [X] **T030** Implement MCPRegistry struct in internal/mcp/registry.go
  - Add tools map (name → MCPTool)
  - Add mutex for concurrent access
  - Add loader and validator references

- [X] **T031** Implement MCPRegistry.Init method in internal/mcp/registry.go
  - Call loader.LoadConfigs()
  - Validate each config
  - Create transport for valid configs
  - Perform async handshake (use goroutines)
  - Start health monitor goroutine
  - Tests: T009 should fully pass
  - Covers FR-006, FR-007

- [X] **T032** Implement MCPRegistry.Close method in internal/mcp/registry.go
  - Wait for in-flight invocations (with timeout)
  - Close all transports
  - Stop health monitor
  - Tests: T010 should start passing
  - Covers Registry contract

- [X] **T033** Implement MCPRegistry query methods in internal/mcp/registry.go
  - Implement List() - return all tools
  - Implement Get(name) - return single tool or ErrToolNotFound
  - Tests: T010 should continue passing
  - Covers Registry contract

- [X] **T034** Implement MCPRegistry admin methods in internal/mcp/registry.go
  - Implement Enable(name) - set tool.Enabled = true, attempt ready transition
  - Implement Disable(name) - set tool.Enabled = false, transition to offline
  - Tests: T011 should fully pass
  - Covers FR-023, FR-024

- [X] **T035** Implement MCPRegistry.Validate and Test methods in internal/mcp/registry.go
  - Implement Validate(paths...) - call validator for each path
  - Implement Test(name) - perform health check, return diagnostics
  - Tests: T012 should fully pass
  - Covers FR-005, FR-026

- [X] **T036** Implement circuit breaker pattern in internal/mcp/circuitbreaker.go
  - Add CircuitBreaker struct (state: closed/open/half-open, failure count, last attempt)
  - Implement RecordSuccess(), RecordFailure()
  - Implement CanAttempt() with exponential backoff
  - Tests: T019 should start passing
  - Covers FR-009

- [X] **T037** Implement health monitor in internal/mcp/registry.go
  - Background goroutine polls tools every 30s
  - Call healthMonitor() and performHealthChecks()
  - Update tool.LastHealthCheck and Status
  - Integrated with circuit breaker on consecutive failures
  - Tests: T019 should continue passing
  - Covers FR-008, FR-030

- [X] **T038** Implement hot-reload with fsnotify in internal/mcp/watcher.go
  - Watch config directories for file changes
  - Debounce events (wait 500ms after last change)
  - Call Reloader.Reload() on change
  - Handle file create, modify, delete
  - Tests: T017 should fully pass
  - Covers FR-004, FR-011

### RBAC & Budgets (M3)
- [X] **T039** Implement RBACEnforcer interface in internal/mcp/rbac/enforcer.go
  - Define CheckPermission(agentRole, capability) method
  - Define GetCapabilities(agentRole) method
  - Define ApplyBudget(toolName) method
  - Define RecordInvocation(log) method

- [X] **T040** Implement RBACEnforcer.CheckPermission in internal/mcp/rbac/enforcer.go
  - Load role capabilities from config (e.g., Viper)
  - Match capability string (exact or wildcard)
  - Return true/false
  - Tests: T013 should fully pass
  - Covers FR-013, FR-014, FR-015

- [X] **T041** Implement budget enforcement in internal/mcp/rbac/budgets.go
  - Implement rate limiter (token bucket algorithm)
  - Implement concurrency limiter (semaphore)
  - Check limits before allowing invocation
  - Tests: T014 should fully pass
  - Covers FR-014

- [X] **T042** Implement audit logging in internal/mcp/rbac/audit.go
  - Create MCPInvocationLog struct
  - Hash arguments with SHA256
  - Write to log file (~/.sdek/logs/mcp-invocations.jsonl)
  - Implement log rotation (if > 7 days, truncate)
  - Tests: T015 should fully pass
  - Covers FR-020, FR-021, FR-022

### Evidence Integration (M4)
- [ ] **T043** Implement AgentInvoker interface in internal/mcp/invoker.go
  - Define InvokeTool(agentRole, toolName, method, args) method
  - Orchestrate: RBAC check → budget check → transport invoke → audit log → normalize evidence

- [ ] **T044** Implement AgentInvoker.InvokeTool in internal/mcp/invoker.go
  - Call rbac.CheckPermission(), return ErrPermissionDenied if false
  - Call rbac.ApplyBudget(), return ErrRateLimited if exceeded
  - Call transport.Invoke(method, args)
  - Create audit log
  - Call normalizeEvidence() helper
  - Tests: T018, T020 should start passing
  - Covers FR-016, FR-017

- [ ] **T045** Integrate with existing redaction in internal/mcp/invoker.go
  - Import internal/ai/redactor.go
  - Call redactor.Redact(response) before returning
  - Set invocationLog.RedactionApplied = true
  - Tests: T020 should continue passing
  - Covers FR-018

- [ ] **T046** Integrate with existing caching in internal/mcp/invoker.go
  - Import internal/store/cache.go
  - Check cache before invoking tool (cache key: toolName + method + argsHash)
  - Store response in cache after successful invocation
  - Respect cache TTL from existing policy
  - Tests: T020 should continue passing
  - Covers FR-019

- [ ] **T047** Implement normalizeEvidence helper in internal/mcp/evidence.go
  - Convert MCP response to pkg/types.Evidence struct
  - Set source metadata (tool name, method, timestamp)
  - Add provenance linking to invocation log
  - Return Evidence entity
  - Tests: T020 should fully pass
  - Covers FR-017

---

## Phase 3.4: CLI Commands (M5)

- [ ] **T048** Implement parent mcp command in cmd/mcp.go
  - Create cobra.Command with "mcp" name
  - Add Short, Long, Example help text
  - Add subcommands (list, validate, test, enable, disable)
  - Add PreRun hook to check mcp.enabled feature flag

- [ ] **T049** [P] Implement mcp list command in cmd/mcp_list.go
  - Call registry.List()
  - Format output as table (NAME, STATUS, LATENCY, CAPABILITIES, ERRORS)
  - Support --format=json flag
  - Tests: T021 (golden file test, next phase)
  - Covers FR-025

- [ ] **T050** [P] Implement mcp validate command in cmd/mcp_validate.go
  - Accept file paths as args
  - Call registry.Validate(paths...)
  - Print detailed schema errors
  - Exit code 1 if validation fails
  - Tests: T022 (golden file test, next phase)
  - Covers FR-005, Scenario 2 (AC-02)

- [ ] **T051** [P] Implement mcp test command in cmd/mcp_test.go
  - Accept tool name as arg
  - Call registry.Test(name)
  - Print health report (handshake status, latency, capabilities)
  - Tests: T023 (golden file test, next phase)
  - Covers FR-026, Scenario 6 (AC-06)

- [ ] **T052** [P] Implement mcp enable command in cmd/mcp_enable.go
  - Accept tool name as arg
  - Call registry.Enable(name)
  - Print success message
  - Tests: T024 (golden file test, next phase)
  - Covers FR-027

- [ ] **T053** [P] Implement mcp disable command in cmd/mcp_disable.go
  - Accept tool name as arg
  - Call registry.Disable(name)
  - Print success message
  - Tests: T025 (golden file test, next phase)
  - Covers FR-027

---

## Phase 3.5: TUI Components (M5)

- [ ] **T054** Implement MCP Tools panel model in ui/models/mcp_tools.go
  - Create Bubble Tea model struct
  - Implement Init() - fetch tools from registry
  - Implement Update() - handle key events (arrow keys, Enter, space for toggle)
  - Implement View() - render tool list with status badges
  - Tests: T026 (golden file test, next phase)
  - Covers FR-029, FR-031

- [ ] **T055** [P] Implement status badge component in ui/components/mcp_status.go
  - Use Lip Gloss for styling
  - Green badge for "ready"
  - Yellow badge for "degraded"
  - Red badge for "offline"
  - Tests: T027 (golden file test, next phase)
  - Covers FR-029

- [ ] **T056** Integrate MCP Tools panel into main TUI in ui/app.go
  - Add MCP Tools tab/panel
  - Wire up model to registry
  - Add real-time status updates (poll every 5s)
  - Tests: T028 (integration test, next phase)
  - Covers FR-029, FR-030

- [ ] **T057** Implement quick-test action in ui/models/mcp_tools.go
  - Add 't' key binding for test
  - Call registry.Test(selectedTool)
  - Display diagnostic results inline
  - Tests: T028 continues
  - Covers FR-032

---

## Phase 3.6: Integration & Polish (M5)

### Golden File Tests
- [ ] **T058** [P] Golden file test for mcp list output in tests/golden/mcp_cli_test.go
  - Create fixture with 3 tools (ready, degraded, offline)
  - Capture stdout from `sdek mcp list`
  - Compare with tests/golden/fixtures/mcp_list_output.txt
  - Covers Scenario 6 (AC-06)

- [ ] **T059** [P] Golden file test for mcp validate output in tests/golden/mcp_cli_test.go
  - Use tests/golden/fixtures/sample_mcp_configs/invalid.json
  - Capture stderr from `sdek mcp validate`
  - Compare with expected schema error output
  - Covers Scenario 2 (AC-02)

- [ ] **T060** [P] Golden file test for TUI rendering in tests/golden/mcp_tui_test.go
  - Render MCP Tools panel with mock data
  - Capture ANSI output
  - Compare with tests/golden/fixtures/mcp_tui_panel.txt
  - Covers TUI styling requirements

### Example Configs
- [ ] **T061** [P] Create example MCP configs in tests/golden/fixtures/sample_mcp_configs/
  - github.json (valid stdio config)
  - jira.json (valid HTTP config)
  - aws.json (valid stdio config with env vars)
  - invalid.json (missing required field)
  - Covers Scenario 1 (AC-01), Scenario 2 (AC-02)

### Documentation
- [ ] **T062** [P] Update docs/CONNECTORS.md with MCP section
  - Add "MCP Integration" section
  - Document config file structure
  - Document capability strings
  - Link to quickstart.md
  - Covers documentation requirements

- [ ] **T063** [P] Update README.md with MCP quickstart
  - Add "Using MCP Tools" section
  - Show example config creation
  - Show `sdek mcp list` example
  - Link to full docs
  - Covers user onboarding

### Final Validation
- [ ] **T064** Run all quickstart.md scenarios and verify success
  - Scenario 1: Config discovery and loading (AC-01)
  - Scenario 2: Schema validation (AC-02)
  - Scenario 3: Orchestrator resilience (AC-03)
  - Scenario 4: RBAC enforcement (AC-04)
  - Scenario 5: Evidence collection (AC-05)
  - Scenario 6: CLI/TUI operations (AC-06)
  - Covers all acceptance criteria

---

## Dependencies

### Critical Path (Must be sequential)
```
T001-T004 (Setup) 
  → T005-T020 (All tests - can run in parallel)
  → T021-T024 (Types)
  → T025-T026 (Validator/Loader)
  → T027-T029 (Transports)
  → T030-T038 (Registry/Orchestrator)
  → T039-T042 (RBAC)
  → T043-T047 (Evidence Integration)
  → T048-T053 (CLI) + T054-T057 (TUI) [parallel]
  → T058-T063 (Polish)
  → T064 (Final validation)
```

### Blocking Relationships
- **Tests (T005-T020) block implementation (T021-T047)** - TDD requirement
- **T021-T024 (Types)** blocks everything else - foundational
- **T025-T026 (Validator/Loader)** blocks T030 (Registry Init)
- **T027-T029 (Transports)** blocks T030 (Registry Init)
- **T030-T038 (Registry)** blocks T039 (RBAC - needs registry)
- **T039-T042 (RBAC)** blocks T043 (AgentInvoker - needs RBAC)
- **T043-T047 (Evidence)** blocks T020 (integration test passing)
- **T048 (Parent command)** blocks T049-T053 (subcommands)
- **T054 (TUI model)** blocks T056 (TUI integration)

### Parallel Opportunities
- **T001-T004**: All setup tasks parallel
- **T005-T020**: All test tasks parallel (different files)
- **T021, T023, T024**: Different struct definitions in same file - sequential preferred but could be parallel with care
- **T028, T029**: Different transport implementations - fully parallel
- **T049-T053**: Different command files - fully parallel
- **T055**: Independent component - parallel with T054
- **T058-T063**: Different golden files and docs - fully parallel

---

## Parallel Execution Examples

### Phase 3.1: Setup (all parallel)
```bash
# Launch T001-T004 together:
Task: "Install JSON Schema validator library and fsnotify in go.mod"
Task: "Copy MCP config JSON Schema to internal/mcp/schema/"
Task: "Create pkg/types/mcp.go with all MCP structs"
Task: "Create internal/mcp/errors.go with error types"
```

### Phase 3.2: Tests First (all parallel)
```bash
# Launch T005-T020 together (16 tasks):
Task: "Contract test for JSON Schema validation in tests/unit/mcp_validator_test.go"
Task: "Unit test for config loader precedence in tests/unit/mcp_loader_test.go"
Task: "Contract test for stdio transport in tests/unit/mcp_transport_stdio_test.go"
Task: "Contract test for HTTP transport in tests/unit/mcp_transport_http_test.go"
Task: "Contract test for MCPRegistry.Init in tests/unit/mcp_registry_test.go"
# ... (continue for all test tasks)
```

### Phase 3.4: CLI Commands (all parallel after T048)
```bash
# Launch T049-T053 together:
Task: "Implement mcp list command in cmd/mcp_list.go"
Task: "Implement mcp validate command in cmd/mcp_validate.go"
Task: "Implement mcp test command in cmd/mcp_test.go"
Task: "Implement mcp enable command in cmd/mcp_enable.go"
Task: "Implement mcp disable command in cmd/mcp_disable.go"
```

### Phase 3.6: Documentation (all parallel)
```bash
# Launch T061-T063 together:
Task: "Create example MCP configs in tests/golden/fixtures/sample_mcp_configs/"
Task: "Update docs/CONNECTORS.md with MCP section"
Task: "Update README.md with MCP quickstart"
```

---

## Notes

- **[P] tasks** = Different files, no dependencies - safe to run in parallel
- **TDD Critical**: All tests (T005-T020) MUST be written and failing before starting implementation (T021+)
- **Test Coverage Goal**: >80% for all internal/mcp/ packages
- **Commit Strategy**: Commit after each task or logical group (e.g., all transports)
- **Performance Validation**: After T064, run benchmarks to verify <5s latency, 100/s throughput
- **Constitution Compliance**: Review each implementation task against constitution principles before merging

---

## Task Generation Rules Applied

1. **From Contracts**:
   - mcp-registry-interface.md → T009-T012 (registry contract tests)
   - mcp-transport-interface.md → T007-T008 (transport contract tests)
   - mcp-rbac-interface.md → T013-T015 (RBAC contract tests)
   - mcp-config-schema.json → T005 (schema validation test)

2. **From Data Model**:
   - MCPConfig → T003, T021 (struct + validation)
   - MCPTool → T003, T022 (struct + state transitions)
   - MCPInvocationLog → T003, T023 (struct + audit logging)
   - AgentCapability → T003, T024 (struct + capability matching)
   - ToolBudget → T003, T024 (struct + budget enforcement)
   - MCPHealthReport → T003, T037 (struct + health monitoring)
   - Evidence → T047 (normalization)

3. **From Quickstart Scenarios**:
   - Scenario 1 (AC-01) → T016 (config discovery integration test), T061 (example configs)
   - Scenario 2 (AC-02) → T005 (validation test), T050 (validate command), T059 (golden test)
   - Scenario 3 (AC-03) → T019 (resilience integration test), T036-T037 (circuit breaker)
   - Scenario 4 (AC-04) → T018 (RBAC integration test), T040 (permission check)
   - Scenario 5 (AC-05) → T020 (evidence integration test), T044-T047 (evidence collection)
   - Scenario 6 (AC-06) → T049-T053 (CLI commands), T054-T057 (TUI), T058-T060 (golden tests)

4. **From Plan.md Milestones**:
   - M1 (Schema & Discovery) → T001-T026
   - M2 (Orchestrator Runtime) → T027-T038
   - M3 (RBAC & Budgets) → T039-T042
   - M4 (Evidence Integration) → T043-T047
   - M5 (CLI/TUI Operations) → T048-T064

---

## Validation Checklist

- [x] All contracts have corresponding tests (T005-T015)
- [x] All entities have model tasks (T021-T024)
- [x] All tests come before implementation (T005-T020 before T021+)
- [x] Parallel tasks are truly independent (different files, verified)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task (verified)
- [x] TDD workflow enforced (tests first, implementation second)
- [x] All 6 quickstart scenarios covered by tasks
- [x] All 40 functional requirements mapped to tasks
- [x] Constitutional principles followed (modular, typed errors, TDD)

---

**TOTAL TASKS**: 64 tasks across 5 milestones (M1-M5)
**ESTIMATED COMPLETION**: 3-4 weeks (assuming 3-5 tasks per day)
**READY FOR EXECUTION**: ✅ Yes - All prerequisites met, tasks ordered by dependencies

---

## Implementation Notes

### AWS MCP Configuration (2025-10-19)
✅ **Test Configuration Created**: `~/.sdek/mcp/aws.json`
- Using official `awslabs.aws-api-mcp-server@latest` package
- Transport: stdio (JSON-RPC over stdin/stdout)
- Command: `uvx awslabs.aws-api-mcp-server@latest`
- Security: `READ_OPERATIONS_ONLY=true` (read-only mode for safety)
- Region: `us-east-1`
- Capabilities: EC2, S3, Lambda, IAM, CloudFormation, RDS, DynamoDB, CloudWatch

**Prerequisites Verified**:
- ✅ uvx installed at `/opt/homebrew/bin/uvx`
- ✅ AWS credentials configured at `~/.aws/credentials`
- ✅ MCP server package downloads and starts successfully
- ✅ Configuration validates against MCPConfig schema

**Documentation**: See `AWS_MCP_TEST_RESULTS.md` in repository root

This configuration will be used to test the Registry implementation once T043-T047 are complete.
