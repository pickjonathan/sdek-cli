
# Implementation Plan: MCP-Native Agent Orchestrator & Tooling Config

**Branch**: `004-mcp-native-agent` | **Date**: 2025-10-19 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/Users/pickjonathan/WorkSpacePrivate/sdek-cli/specs/004-mcp-native-agent/spec.md`

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

This feature enables sdek-cli's agent orchestrator and AI agents to connect to external tools using the Model Context Protocol (MCP) standard with JSON configurations that are compatible with VS Code and Cursor. Instead of custom-coded connectors, agents will discover and invoke MCP tools (GitHub, Jira, AWS, Slack, CI/CD, docs systems) for evidence collection. The implementation provides schema validation, hot-reload capabilities, RBAC enforcement, health monitoring, and comprehensive CLI/TUI operations for managing MCP tool connections. This creates a portable, standardized approach where organizations can reuse the same MCP configurations across their entire compliance toolchain.

## Technical Context

**Language/Version**: Go 1.23+ (per existing project standard)  
**Primary Dependencies**: Cobra (CLI), Viper (config), Bubble Tea + Lip Gloss (TUI), fsnotify (file watching), JSON Schema validator library  
**Storage**: File system (JSON configs in `~/.sdek/mcp/`, `./.sdek/mcp/`, and `$SDEK_MCP_PATH`); existing state management via internal/store  
**Testing**: Go standard testing, table-driven tests, golden file tests for TUI output, integration tests with mock MCP servers  
**Target Platform**: Linux, macOS, Windows (cross-platform via Go stdlib)  
**Project Type**: Single project (CLI tool with TUI)  
**Performance Goals**: <100ms cold start (existing target), MCP handshake <500ms, hot-reload <100ms, support 50+ concurrent MCP tools  
**Constraints**: Config compatibility with VS Code/Cursor MCP spec v1.0+, graceful degradation when tools unavailable, zero breaking changes to existing evidence collection API  
**Scale/Scope**: Support 50-100 MCP tool configs per deployment, 1000+ invocations per analysis run, multi-tenant RBAC with 10+ agent roles

**User-Provided Implementation Details**:
- 5 Milestones (M1-M5): Schema & Discovery, Orchestrator Runtime, RBAC & Budgets, Evidence Collection Integration, CLI/TUI Operations
- Feature flags: `mcp.enabled`, `mcp.hotReload`, `mcp.rbac.enforced`, `mcp.toolBudgets.enabled`
- Dependencies: VS Code/Cursor MCP schema parity, file watcher library, existing AI redaction/cache subsystems, RBAC module
- Testing: Unit (schema validation, loader precedence, state transitions, RBAC), Integration (mock MCP servers, retry/backoff), E2E (analysis via MCP evidence), Golden (sample configs, CLI/TUI snapshots)
- Telemetry: Metrics (tool readiness %, handshake latency, call success rate), Events (mcp_tool_loaded, mcp_tool_failed, mcp_invoked, permission_denied), Tracing (run_id → agent → tool → method)
- Risks: Config drift (versioned schema + compatibility shims), Tool instability (circuit breakers, backoff), Over-collection/cost (RBAC + budgets), Secret leakage (mandatory redaction)

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Correctness and Safety ✅
- **PASS**: All MCP configs will be validated before use (JSON schema validation with file/line/property errors)
- **PASS**: Typed errors with context (e.g., `fmt.Errorf("mcp registry: handshake failed for %s: %w", toolName, err)`)
- **PASS**: No panics; all tool failures handled gracefully with degraded state transitions
- **PASS**: Side effects (tool loading, enable/disable) will be logged and reflected in TUI

### Configuration Management ✅
- **PASS**: Viper will manage feature flags (`mcp.enabled`, `mcp.hotReload`, `mcp.rbac.enforced`, `mcp.toolBudgets.enabled`)
- **PASS**: MCP config discovery follows precedence: CLI flags (future) → env vars (`SDEK_MCP_PATH`) → project (`./.sdek/mcp/`) → global (`~/.sdek/mcp/`)
- **PASS**: No manual flag/env duplication; single source of truth via Viper

### Command Design (Cobra) ✅
- **PASS**: New `mcp` command group with subcommands: `list`, `validate`, `test`, `enable`, `disable`
- **PASS**: Each subcommand will have `Short`, `Long`, and `Example` help text
- **PASS**: PreRun hooks will validate flags and check feature enablement
- **PASS**: PostRun hooks will flush metrics and logs

### User Experience & Terminal UI (Bubble Tea) ✅
- **PASS**: New TUI panel "MCP Tools" with status, latency, enable/disable toggles
- **PASS**: Lip Gloss for consistent styling (status colors: green=ready, yellow=degraded, red=offline)
- **PASS**: Non-interactive modes: all CLI commands work without TUI (e.g., `sdek mcp list --format=json`)
- **PASS**: Intuitive keyboard shortcuts in TUI panel (standard arrow keys, Enter, toggle with space)

### Test-Driven Development ✅
- **PASS**: Unit tests for schema validation, config loading precedence, RBAC checks, state transitions
- **PASS**: Integration tests with mock MCP servers (handshake, health-check, retry/backoff)
- **PASS**: Golden file tests for CLI output and TUI snapshots
- **PASS**: E2E tests: run analysis with MCP-only evidence sources, assert audit trail
- **PASS**: All tests will pass before merge (`go test ./...`)

### Performance & Efficiency ✅
- **PASS**: MCP registry initialized once at startup; hot-reload on file changes (not full restart)
- **PASS**: Handshake/health-checks performed asynchronously; no blocking on startup
- **PASS**: Tool response caching preserved from existing cache subsystem (internal/store/cache.go)
- **PASS**: Circuit breaker pattern for failing tools (exponential backoff, degraded state)
- **PASS**: Profiling for memory leaks in long-lived TUI sessions (existing practice)

### Cross-Platform Compatibility ✅
- **PASS**: All path operations use `filepath.Join` for MCP config discovery
- **PASS**: File watcher (fsnotify) is cross-platform
- **PASS**: MCP tool invocation uses stdio and HTTP (both cross-platform transports)
- **PASS**: Terminal sizing and colors handled by Bubble Tea (already cross-platform)

### Observability & Logging ✅
- **PASS**: `--verbose` flag enables debug logs for MCP subsystem
- **PASS**: Structured logs (slog) to stderr: `mcp_tool_loaded`, `mcp_tool_failed`, `mcp_invoked`, `permission_denied`
- **PASS**: Metrics emitted: tool readiness %, handshake latency, call success rate, retries, RBAC denials
- **PASS**: Tracing: `run_id` → `agent_id` → `tool_name` → `method` (duration, redaction applied)

### Modularity & Code Organization ✅
- **PASS**: New packages align with existing structure:
  - `cmd/mcp.go` + `cmd/mcp_*.go` — Cobra commands (thin)
  - `internal/mcp/` — Registry, loader, validator, health-checker (business logic)
  - `internal/mcp/rbac/` — RBAC enforcement and capability mapping
  - `internal/mcp/transport/` — MCP protocol transport (stdio, HTTP)
  - `ui/models/mcp_tools.go` — Bubble Tea model for TUI panel
  - `pkg/types/mcp.go` — Public types (MCPConfig, MCPTool, MCPInvocationLog)
- **PASS**: No cyclic dependencies; MCP subsystem depends on existing AI/store/config modules but not vice versa
- **PASS**: Dependency injection for MCP registry (testable via mocks)

### Extensibility & Versioning ✅
- **PASS**: MCP config schema is versioned (v1.0.0 initial); future versions will use compatibility shims
- **PASS**: Feature flags allow gradual rollout and A/B testing (`mcp.enabled` default: true)
- **PASS**: Backward compatibility: existing connectors (internal/ai/connectors/) remain functional; MCP is additive
- **PASS**: Plugin-like architecture: new MCP tools require config files only (no code changes)

### Documentation & Clarity ✅
- **PASS**: Cobra auto-generates markdown help for `mcp` commands
- **PASS**: README will include MCP quickstart (copy VS Code config, run `sdek mcp list`)
- **PASS**: `docs/CONNECTORS.md` will be updated with MCP section
- **PASS**: Code comments for non-obvious decisions (e.g., precedence order, retry backoff algorithm)
- **PASS**: Idiomatic Go; no clever abstractions

**Constitution Compliance**: ✅ ALL CHECKS PASS — No violations detected

## Project Structure

### Documentation (this feature)
```
specs/004-mcp-native-agent/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
│   ├── mcp-config-schema.json      # JSON Schema for MCP configs
│   ├── mcp-registry-interface.md   # Go interface contracts
│   ├── mcp-transport-interface.md  # Transport abstraction
│   └── mcp-rbac-interface.md       # RBAC enforcement contracts
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
cmd/
├── mcp.go                    # Parent command for MCP subcommands
├── mcp_list.go               # List discovered tools with status
├── mcp_validate.go           # Validate config files
├── mcp_test.go               # Test tool health
├── mcp_enable.go             # Enable a tool
└── mcp_disable.go            # Disable a tool

internal/
├── mcp/
│   ├── registry.go           # MCPRegistry implementation (load, init, get, enable, disable)
│   ├── registry_test.go      # Unit tests for registry
│   ├── loader.go             # Config discovery and loading (precedence handling)
│   ├── loader_test.go        # Unit tests for loader (precedence, hot-reload)
│   ├── validator.go          # JSON schema validation
│   ├── validator_test.go     # Unit tests for validator
│   ├── health.go             # Health check and handshake logic
│   ├── health_test.go        # Unit tests for health checks
│   ├── invoker.go            # AgentInvoker implementation (call MCP tools)
│   ├── invoker_test.go       # Unit tests for invoker
│   ├── rbac/
│   │   ├── enforcer.go       # RBAC policy evaluation
│   │   ├── enforcer_test.go  # Unit tests for RBAC
│   │   ├── capability.go     # Capability mapping (tool → verbs)
│   │   └── budgets.go        # Rate limits, concurrency, timeouts
│   └── transport/
│       ├── stdio.go          # stdio MCP transport
│       ├── http.go           # HTTP MCP transport
│       ├── transport.go      # Transport interface
│       └── transport_test.go # Unit tests for transports
└── ai/
    └── mcp_evidence.go       # Integration: MCP invoker → evidence graph

pkg/types/
├── mcp.go                    # MCPConfig, MCPTool, MCPInvocationLog, MCPHealthReport
└── mcp_test.go               # Unit tests for types

ui/
├── models/
│   ├── mcp_tools.go          # Bubble Tea model for MCP Tools panel
│   └── mcp_tools_test.go     # Unit tests for TUI model
└── components/
    └── mcp_status.go         # Status badge component (ready/degraded/offline)

tests/
├── integration/
│   ├── mcp_handshake_test.go      # Integration test: handshake with mock MCP server
│   ├── mcp_hotreload_test.go      # Integration test: file watcher + hot-reload
│   ├── mcp_rbac_test.go           # Integration test: RBAC enforcement end-to-end
│   └── mcp_evidence_test.go       # Integration test: evidence collection via MCP
├── unit/
│   ├── mcp_loader_test.go         # Additional loader tests (edge cases)
│   ├── mcp_validator_test.go      # Additional validator tests (schema errors)
│   └── mcp_budgets_test.go        # Budget enforcement tests
└── golden/
    ├── fixtures/
    │   ├── mcp_list_output.txt         # Expected CLI output
    │   ├── mcp_tui_panel.txt           # Expected TUI rendering
    │   └── sample_mcp_configs/         # Example MCP configs
    │       ├── github.json
    │       ├── jira.json
    │       ├── aws.json
    │       └── invalid.json
    └── mcp_cli_test.go                 # Golden file tests for CLI

testdata/
└── mcp/
    ├── valid_config.json               # Valid MCP config for tests
    ├── invalid_config.json             # Invalid config (schema violation)
    ├── multiple_locations/             # Test loader precedence
    │   ├── global/github.json
    │   ├── project/github.json
    │   └── env/github.json
    └── mock_server/                    # Mock MCP server for integration tests
        └── main.go
```

**Structure Decision**: Single project (CLI tool). This feature extends the existing sdek-cli codebase with a new `mcp` command group and supporting packages. The structure follows constitutional requirements: commands in `cmd/`, business logic in `internal/mcp/`, public types in `pkg/types/`, TUI components in `ui/`, and comprehensive tests in `tests/`.

## Phase 0: Outline & Research

**Status**: ✅ Complete (no NEEDS CLARIFICATION in Technical Context; all details provided by user)

### Research Topics

1. **MCP Specification Alignment**
   - Task: Research VS Code and Cursor MCP configuration structure and semantics
   - Task: Identify MCP protocol version compatibility requirements (v1.0+)
   - Task: Document MCP transport protocols (stdio, HTTP)
   - Task: Understand MCP handshake and health-check patterns

2. **JSON Schema Validation in Go**
   - Task: Research Go libraries for JSON Schema validation (e.g., `github.com/xeipuuv/gojsonschema`, `github.com/santhosh-tekuri/jsonschema`)
   - Task: Best practices for reporting schema errors with file/line/property paths
   - Task: Schema versioning and compatibility strategies

3. **File Watching and Hot-Reload**
   - Task: Research `fsnotify` library for cross-platform file watching
   - Task: Patterns for debouncing file change events
   - Task: Safe hot-reload without race conditions (graceful connection swaps)

4. **MCP Transport Implementation**
   - Task: Research stdio-based IPC in Go (stdin/stdout communication)
   - Task: HTTP client patterns for MCP over HTTP
   - Task: Connection pooling and lifecycle management for long-lived MCP connections

5. **Circuit Breaker and Retry Patterns**
   - Task: Research exponential backoff algorithms in Go
   - Task: Circuit breaker patterns for failing external services
   - Task: Health check strategies and state transition logic (ready → degraded → offline)

6. **RBAC Integration**
   - Task: Review existing RBAC implementation in sdek-cli (if any)
   - Task: Capability-based access control patterns
   - Task: Per-resource rate limiting and concurrency budgets

7. **Evidence Graph Integration**
   - Task: Review existing evidence graph structure (`pkg/types/evidence.go`, `pkg/types/finding.go`)
   - Task: Patterns for normalizing diverse data sources into evidence entities
   - Task: Preserve provenance (source tool, method, timestamp) in evidence metadata

8. **Telemetry and Observability**
   - Task: Review existing metrics/tracing infrastructure
   - Task: Best practices for structured logging of external tool invocations
   - Task: Audit log design for compliance (what to log, retention policies)

**Output**: `research.md` documenting decisions, rationale, and alternatives for each topic

## Phase 1: Design & Contracts
*Prerequisites: research.md complete ✅*

### 1. Extract Entities → `data-model.md`
Entities from feature spec and research:
- **MCPConfig**: Config file structure (name, command, args, env, transport, capabilities, version)
- **MCPTool**: Runtime tool state (status, health, latency, error history, circuit breaker state)
- **MCPInvocationLog**: Audit record (timestamp, agent, tool, method, arg hash, redaction, duration, status)
- **AgentCapability**: RBAC mapping (agent role → capabilities)
- **ToolBudget**: Rate limits and concurrency controls
- **MCPHealthReport**: Health check results (handshake status, latency, error)
- **Evidence**: Integration point (source tool, method, timestamp, data, provenance)

### 2. Generate Interface Contracts → `/contracts/`
From functional requirements, define Go interfaces:
- **MCPRegistry**: Load, Init, Get, Enable, Disable, Validate, Test
- **MCPTransport**: Invoke, HealthCheck, Close (stdio and HTTP implementations)
- **RBACEnforcer**: CheckPermission, GetCapabilities, ApplyBudget
- **AgentInvoker**: InvokeTool (orchestrates transport + RBAC + audit + evidence normalization)

Output to:
- `contracts/mcp-config-schema.json` — JSON Schema for MCP configs
- `contracts/mcp-registry-interface.md` — Go interface for MCPRegistry
- `contracts/mcp-transport-interface.md` — Go interface for transports
- `contracts/mcp-rbac-interface.md` — Go interface for RBAC enforcement

### 3. Generate Contract Tests
From contracts, create failing tests:
- `internal/mcp/registry_test.go` — Test Load, Init, Get with mock configs
- `internal/mcp/transport/transport_test.go` — Test stdio/HTTP transports with mock servers
- `internal/mcp/rbac/enforcer_test.go` — Test permission checks and denials
- `tests/integration/mcp_evidence_test.go` — Test end-to-end evidence collection
- Tests will fail until implementation exists (TDD)

### 4. Extract Test Scenarios → `quickstart.md`
From user stories (spec.md), create quickstart validation:
- Scenario 1: Drop github.json into ~/.sdek/mcp/, run `sdek mcp list`, see "ready"
- Scenario 2: Validate invalid config, see detailed error
- Scenario 3: Simulate server outage, observe degraded state and recovery
- Scenario 4: Attempt unauthorized MCP call, see RBAC denial
- Scenario 5: Run analysis with MCP evidence, verify audit trail
- Scenario 6: Use TUI to toggle tool on/off, test health check

### 5. Update Agent Context File
Run script to update `.github/copilot-instructions.md`:
```bash
.specify/scripts/bash/update-agent-context.sh copilot
```
Add new technologies:
- fsnotify (file watching)
- github.com/santhosh-tekuri/jsonschema/v5 (schema validation)
- MCP protocol (stdio/HTTP transports)
- Circuit breaker pattern
- Capability-based RBAC

**Output**: ✅ data-model.md, contracts/*, failing tests, quickstart.md, .github/copilot-instructions.md updated

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:

The /tasks command will load `.specify/templates/tasks-template.md` and generate ordered, atomic tasks following TDD principles:

### Task Categories

1. **Foundation Tasks** (M1: Schema & Discovery)
   - Define Go types in `pkg/types/mcp.go` from data-model.md
   - Implement JSON Schema validator in `internal/mcp/validator.go`
   - Implement config loader with precedence in `internal/mcp/loader.go`
   - Write unit tests for validator and loader
   - **Est**: 8-10 tasks

2. **Transport Layer Tasks** (M1 + M2)
   - Define Transport interface in `internal/mcp/transport/transport.go`
   - Implement stdio transport with JSON-RPC 2.0
   - Implement HTTP transport with JSON-RPC 2.0
   - Write unit tests for both transports
   - Write integration tests with mock MCP servers
   - **Est**: 6-8 tasks

3. **Registry & Orchestrator Tasks** (M2)
   - Implement MCPRegistry in `internal/mcp/registry.go`
   - Implement health check and circuit breaker logic
   - Implement hot-reload via fsnotify file watcher
   - Write unit tests for registry lifecycle
   - Write integration tests for hot-reload
   - **Est**: 8-10 tasks

4. **RBAC Tasks** (M3)
   - Implement RBACEnforcer in `internal/mcp/rbac/enforcer.go`
   - Implement capability matching and budget enforcement
   - Implement audit logging in `internal/mcp/rbac/audit.go`
   - Write unit tests for RBAC checks
   - Write integration tests for permission denial
   - **Est**: 6-8 tasks

5. **Evidence Integration Tasks** (M4)
   - Implement AgentInvoker in `internal/mcp/invoker.go`
   - Integrate with existing redaction (`internal/ai/redactor.go`)
   - Integrate with existing caching (`internal/store/cache.go`)
   - Normalize MCP responses into Evidence entities
   - Write end-to-end integration tests
   - **Est**: 5-7 tasks

6. **CLI Commands Tasks** (M5)
   - Implement `cmd/mcp.go` (parent command)
   - Implement `cmd/mcp_list.go`, `cmd/mcp_validate.go`, `cmd/mcp_test.go`
   - Implement `cmd/mcp_enable.go`, `cmd/mcp_disable.go`
   - Write command unit tests
   - Write golden file tests for CLI output
   - **Est**: 7-9 tasks

7. **TUI Tasks** (M5)
   - Implement `ui/models/mcp_tools.go` (Bubble Tea model)
   - Implement `ui/components/mcp_status.go` (status badge)
   - Integrate with main TUI app
   - Write TUI unit tests
   - Write golden file tests for TUI rendering
   - **Est**: 5-7 tasks

8. **Documentation & Examples Tasks** (M5)
   - Update `docs/CONNECTORS.md` with MCP section
   - Create example MCP configs in `tests/golden/fixtures/sample_mcp_configs/`
   - Update README with MCP quickstart
   - Create demo walkthrough video/script
   - **Est**: 3-5 tasks

**Ordering Strategy**:
- **TDD Order**: Tests before implementation for each module
- **Dependency Order**: Types → Validator/Loader → Transport → Registry → RBAC → Evidence → CLI/TUI
- **Parallel Execution**: Tasks marked [P] can be executed independently (e.g., stdio transport [P] vs HTTP transport [P])
- **Milestone Alignment**: Tasks grouped by milestone (M1-M5) with explicit exit criteria

**Estimated Total**: **48-64 tasks** across 5 milestones

**Task Template Structure**:
Each task will include:
- Task number and title
- Milestone (M1-M5)
- Dependencies (if any)
- Acceptance criteria (from contracts and quickstart)
- Test requirements (unit, integration, golden)
- Parallel flag [P] where applicable

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

**Status**: ✅ No violations detected

All constitutional principles are satisfied by the design:
- Modular architecture with clear separation of concerns
- Standard Go idioms and best practices
- Proper error handling and observability
- Cross-platform compatibility
- TDD approach with comprehensive test coverage
- Thin CLI commands with business logic in internal packages

No complexity deviations require justification.


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command) ✅
- [x] Phase 1: Design complete (/plan command) ✅
- [x] Phase 2: Task planning complete (/plan command - describe approach only) ✅
- [x] Phase 3: Tasks generated (/tasks command) ✅
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS ✅
- [x] Post-Design Constitution Check: PASS ✅
- [x] All NEEDS CLARIFICATION resolved ✅
- [x] Complexity deviations documented: N/A (no deviations) ✅

**Artifacts Generated**:
- [x] `research.md` — Technology decisions and rationale
- [x] `data-model.md` — 7 core entities defined
- [x] `contracts/mcp-config-schema.json` — JSON Schema for configs
- [x] `contracts/mcp-registry-interface.md` — Registry contract
- [x] `contracts/mcp-transport-interface.md` — Transport contract
- [x] `contracts/mcp-rbac-interface.md` — RBAC contract
- [x] `quickstart.md` — 6 validation scenarios
- [x] `.github/copilot-instructions.md` — Updated with new technologies
- [x] `tasks.md` — 64 implementation tasks across 5 milestones

**Ready for Next Phase**: ✅ Yes — Execute tasks in tasks.md following TDD order

---
*Based on Constitution v1.0.0 - See `.specify/memory/constitution.md`*
