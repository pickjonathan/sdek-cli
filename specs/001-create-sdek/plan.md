
# Implementation Plan: Create sdek

**Branch**: `001-create-sdek` | **Date**: 2025-10-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/Users/pickjonathan/WorkSpacePrivate/sdek-cli/specs/001-create-sdek/spec.md`

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

Build sdek, a Go-based CLI and terminal UI tool that reduces audit preparation time by 30% through automated compliance evidence mapping. The tool ingests simulated data from five sources (Git, Jira, Slack, CI/CD, Docs), maps them to compliance frameworks (SOC2, ISO 27001, PCI DSS) using deterministic heuristics, and provides interactive TUI visualization with export capabilities. Phase 1 focuses on local-only operation with no external integrations.

**Technical Approach**: Use Cobra for command structure, Viper for configuration management, and Bubble Tea + Lip Gloss/Bubbles for rich terminal UI. Implement deterministic evidence mapping with severity-based risk scoring (3 high = 1 critical, 6 medium = 1 critical, 12 low = 1 critical). Store state in local JSON with auto-save. Provide keyboard-driven navigation with color-coded risk visualization.

## Technical Context

**Language/Version**: Go 1.23+ (latest stable)  
**Primary Dependencies**: 
- Cobra (CLI commands and flags)
- Viper (configuration management)
- Bubble Tea (terminal UI framework)
- Lip Gloss (TUI styling)
- Bubbles (TUI components: lists, spinners, text inputs)

**Storage**: Local JSON files (`$HOME/.sdek/state.json` for evidence cache, `$HOME/.sdek/config.yaml` for configuration)  
**Testing**: Go standard testing (`go test`), golden file tests for TUI rendering, integration tests for command flows  
**Target Platform**: Cross-platform (Linux, macOS, Windows with PowerShell/WSL)  
**Project Type**: Single project (CLI application)  
**Performance Goals**: 
- Cold start under 100ms
- TUI rendering at 60fps (16ms frame time)
- Handle up to 10,000 evidence items without degradation

**Constraints**: 
- No network calls or external API dependencies
- Deterministic data generation (no real AI inference)
- Minimum terminal size: 80 columns × 24 rows
- Local-only persistence (no database)
- Single-user operation (no concurrency control beyond file writes)

**Scale/Scope**: 
- 5 simulated data sources, each with 10-50 events
- 3 compliance frameworks with multiple controls each
- 3 simulated users (1 compliance manager, 2 engineers)
- Support for 10,000 evidence items total

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Based on SDEK CLI Constitution v1.0.0:

### I. Correctness and Safety
✅ **PASS** - All commands will validate flags/arguments in PreRun hooks. Using `fmt.Errorf` with context wrapping. No panics in error paths. File writes logged via structured logging.

### II. Configuration Management
✅ **PASS** - Viper as single source of truth. Precedence: CLI flags → env vars → config file → defaults. Auto-load from `$HOME/.sdek/config.yaml`. Support `--config` flag.

### III. Command Design (Cobra)
✅ **PASS** - All commands (`tui`, `ingest`, `analyze`, `report`, `seed`, `config`) will have Use/Short/Long/Example fields. PreRun for validation, PostRun for cleanup. Logical grouping under root command.

### IV. User Experience & Terminal UI (Bubble Tea)
✅ **PASS** - Clean responsive TUI with Lip Gloss styling. Keyboard shortcuts: Tab (sections), ↑/↓ (navigation), Enter (select), q (quit), r (refresh), e (export), / (search). Non-interactive mode via CLI flags. Minimum 80×24 terminal size detection.

### V. Test-Driven Development
✅ **PASS** - Unit tests for all RunE functions. Integration tests for ingest→analyze→report flow. Golden file tests for TUI rendering. All tests must pass before merge.

### VI. Performance & Efficiency
✅ **PASS** - Target <100ms cold start. No unnecessary goroutines. Cache expensive operations. Profile TUI for memory leaks during long sessions.

### VII. Cross-Platform Compatibility
✅ **PASS** - Support Linux/macOS/Windows. Use `filepath.Join` for all paths. Handle terminal sizing gracefully. Platform detection for fallback messages.

### VIII. Observability & Logging
✅ **PASS** - `--verbose` and `--log-level` flags. Structured logging (log/slog) to stderr. Levels: debug, info, warn, error. Telemetry events in JSON format.

### IX. Modularity & Code Organization
✅ **PASS** - Structure: `cmd/` (commands), `internal/ingest`, `internal/analyze`, `internal/report`, `internal/store`, `ui/` (Bubble Tea), `pkg/` (exportable). Thin command layer, business logic in internal packages.

### X. Extensibility & Versioning
✅ **PASS** - `--version` flag with build metadata. Semantic versioning. Backward-compatible flags. Extensible command structure for future phases.

### XI. Documentation & Clarity
✅ **PASS** - Auto-generate markdown help via Cobra. Examples in help output. Idiomatic Go style. Document non-obvious design decisions.

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
sdek-cli/
├── cmd/
│   ├── root.go              # Root command with global flags
│   ├── tui.go               # TUI launch command
│   ├── ingest.go            # Data ingestion command
│   ├── analyze.go           # Evidence mapping command
│   ├── report.go            # Export command
│   ├── seed.go              # Demo data generation command
│   └── config.go            # Configuration management command
├── internal/
│   ├── ingest/
│   │   ├── generator.go     # Simulated data generation (10-50 events per source)
│   │   ├── git.go           # Git commit simulator
│   │   ├── jira.go          # Jira ticket simulator
│   │   ├── slack.go         # Slack message simulator
│   │   ├── cicd.go          # CI/CD event simulator
│   │   └── docs.go          # Documentation source simulator
│   ├── analyze/
│   │   ├── mapper.go        # Evidence-to-control mapping logic
│   │   ├── confidence.go    # Confidence level calculation (Low/Medium/High)
│   │   ├── risk.go          # Severity-based risk scoring
│   │   └── frameworks.go    # SOC2, ISO 27001, PCI DSS definitions
│   ├── report/
│   │   ├── exporter.go      # JSON report generation
│   │   └── formatter.go     # Report formatting utilities
│   ├── store/
│   │   ├── state.go         # State persistence layer
│   │   ├── autosave.go      # Auto-save functionality
│   │   └── cache.go         # Evidence cache management
│   └── config/
│       ├── loader.go        # Viper configuration loader
│       └── validator.go     # Config validation
├── ui/
│   ├── app.go               # Main Bubble Tea application
│   ├── models/
│   │   ├── home.go          # Home screen model (Sources/Frameworks/Findings)
│   │   ├── sources.go       # Sources list model
│   │   ├── frameworks.go    # Frameworks list model
│   │   ├── controls.go      # Control detail model
│   │   ├── evidence.go      # Evidence list model
│   │   └── settings.go      # Settings menu model
│   ├── components/
│   │   ├── card.go          # Evidence card component
│   │   ├── riskbar.go       # Risk visualization bar component
│   │   ├── list.go          # Custom list component
│   │   └── statusbar.go     # Status bar component
│   └── styles/
│       └── theme.go         # Lip Gloss theme definitions
├── pkg/
│   └── types/
│       ├── source.go        # Source entity
│       ├── event.go         # Event entity
│       ├── framework.go     # Framework entity
│       ├── control.go       # Control entity
│       ├── evidence.go      # Evidence entity
│       ├── finding.go       # Finding entity
│       └── user.go          # User entity
├── tests/
│   ├── unit/
│   │   ├── ingest_test.go
│   │   ├── analyze_test.go
│   │   ├── report_test.go
│   │   └── store_test.go
│   ├── integration/
│   │   ├── flow_test.go     # ingest → analyze → report
│   │   └── config_test.go   # Configuration precedence tests
│   └── golden/
│       ├── tui_home_test.go
│       ├── tui_controls_test.go
│       └── fixtures/        # Golden file snapshots
├── main.go                  # Entry point
├── go.mod
├── go.sum
├── README.md
└── .github/
    └── copilot-instructions.md  # Agent-specific guidance
```

**Structure Decision**: Single project structure (Option 1). This is a CLI application with clear separation: `cmd/` for Cobra commands (thin layer), `internal/` for business logic (ingest/analyze/report/store/config), `ui/` for Bubble Tea components, `pkg/types/` for shared domain entities, and `tests/` organized by test type (unit/integration/golden).

## Phase 0: Outline & Research

✅ **COMPLETE** - All research complete, no NEEDS CLARIFICATION items remaining.

**Research Areas Completed**:
1. **Technology Stack**: Go 1.23+, Cobra, Viper, Bubble Tea, Lip Gloss, Bubbles, log/slog
2. **Data Model**: Local JSON persistence with in-memory caching, auto-save pattern
3. **Evidence Mapping**: Rule-based heuristics with keyword matching, confidence scoring
4. **Risk Scoring**: Severity-weighted algorithm (3H=1C, 6M=1C, 12L=1C)
5. **Testing Strategy**: Three-tier approach (unit, integration, golden file tests)
6. **Performance**: Lazy loading, pre-calculated scores, viewport rendering
7. **Cross-Platform**: Runtime terminal detection, graceful degradation
8. **Configuration**: YAML schema with Viper, hierarchical structure

**Key Decisions**:
- Deterministic seeding for reproducible demos
- Rule-based mapping (no AI inference needed)
- Write-through cache with auto-save
- Virtual scrolling for large datasets
- Table-driven tests for data generation

**Output**: ✅ [research.md](./research.md) - All technical unknowns resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

✅ **COMPLETE** - All design artifacts created, agent context updated.

1. ✅ **Extract entities from feature spec** → `data-model.md`:
   - 8 entities defined: Source, Event, Framework, Control, Evidence, Finding, User, Config
   - Validation rules extracted from requirements
   - Relationships documented with diagram
   - JSON state schema defined with indexes

2. ✅ **Generate CLI contracts** from functional requirements:
   - 7 command contracts: root, tui, ingest, analyze, report, seed, config
   - Each command documented with: usage, flags, behavior, output format, exit codes
   - Configuration precedence defined: flags → env → config → defaults
   - Error handling contracts defined

3. **Generate contract tests** from contracts:
   - (Phase 2 task: Create failing tests for each command)
   - Tests will assert: flag parsing, exit codes, output formats, error messages
   - Tests must fail (no implementation yet)

4. ✅ **Extract test scenarios** from user stories:
   - Created quickstart.md with 7 validation scenarios
   - Scenarios: setup, TUI navigation, CLI workflow, config management, role-based visibility, error handling, performance
   - Each scenario includes: steps, expected output, validation criteria

5. ✅ **Update agent file incrementally** (O(1) operation):
   - Ran `.specify/scripts/bash/update-agent-context.sh copilot`
   - Created `.github/copilot-instructions.md` with:
     * Active Technologies: Go 1.23+ (latest stable)
     * Project Structure: src/, tests/
     * Database: Local JSON files ($HOME/.sdek/state.json, $HOME/.sdek/config.yaml)
     * Recent Changes: 001-create-sdek context added

**Output**: 
- ✅ [data-model.md](./data-model.md) - 8 entities with validation rules
- ✅ [contracts/cli-commands.md](./contracts/cli-commands.md) - 7 command contracts
- ✅ [quickstart.md](./quickstart.md) - 7 end-to-end validation scenarios
- ✅ `.github/copilot-instructions.md` - Agent-specific guidance

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
1. **Load Template**: Start with `.specify/templates/tasks-template.md` structure
2. **Generate from Contracts**: For each CLI command in `contracts/cli-commands.md`:
   - Create contract test task (verify flags, exit codes, output format)
   - Create command implementation task (RunE function with PreRun validation)
   - Mark as [P] if commands are independent (e.g., config and seed)
3. **Generate from Data Model**: For each entity in `data-model.md`:
   - Create type definition task in `pkg/types/` [P]
   - Create validation function task [P]
   - Create serialization tests [P]
4. **Generate from Quickstart**: For each scenario in `quickstart.md`:
   - Create integration test task matching scenario steps
   - Link to contract test tasks as prerequisites
5. **Implementation Tasks**: Generate tasks to make tests pass:
   - Store layer (state.go, autosave.go, cache.go)
   - Business logic (ingest generators, analyze mapper, report exporter)
   - UI components (Bubble Tea models, styles, keyboard handling)
   - Configuration loader (Viper integration)

**Ordering Strategy**:
1. **Foundation First** (1-5): Project setup, go.mod, directory structure
2. **Types Layer** (6-13, all [P]): Entity definitions in `pkg/types/`
3. **Contract Tests** (14-20, mostly [P]): Failing command tests
4. **Store Layer** (21-23): state.go → cache.go → autosave.go
5. **Business Logic** (24-30): Ingest generators → analyze mapper → report exporter
6. **Command Implementation** (31-37): Make contract tests pass (seed → ingest → analyze → report → config → tui)
7. **UI Layer** (38-44): Bubble Tea models → components → keyboard handlers
8. **Integration Tests** (45-48): Quickstart scenarios as tests
9. **Polish** (49-50): Documentation, performance tuning

**TDD Principle**: Tests always before implementation. Use `// TODO: Implement` stubs to make tests compile but fail.

**Parallelization**: Mark [P] for tasks that:
- Operate on different files
- Have no shared dependencies
- Can be tested independently

**Estimated Output**: 48-50 numbered tasks with clear dependencies

**Task Template Format**:
```markdown
### Task [N]: [Title]
**Type**: [Test/Implementation/Documentation]
**Dependency**: Task [M] (if applicable)
**Parallel**: [P] (if independent)

Description of what to do.

**Acceptance Criteria**:
- [ ] Criterion 1
- [ ] Criterion 2

**Files**:
- Create: `path/to/new/file.go`
- Modify: `path/to/existing/file.go`

**Constitution Check**: [Relevant principle number]
```

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
- ✅ Phase 0: Research complete (/plan command) - research.md created
- ✅ Phase 1: Design complete (/plan command) - data-model.md, contracts/, quickstart.md, agent file updated
- ✅ Phase 2: Task planning complete (/plan command - approach described, ready for /tasks)
- [ ] Phase 3: Tasks generated (/tasks command) - Next step: Run /tasks to create tasks.md
- [ ] Phase 4: Implementation complete - Execute tasks in tasks.md
- [ ] Phase 5: Validation passed - Run quickstart.md scenarios

**Gate Status**:
- ✅ Initial Constitution Check: PASS (all 11 principles validated)
- ✅ Post-Design Constitution Check: PASS (no new violations introduced)
- ✅ All NEEDS CLARIFICATION resolved (5 questions answered in spec)
- ✅ Complexity deviations documented (none - single project structure justified)

**Artifacts Created**:
- ✅ `specs/001-create-sdek/research.md` (10 technical decisions)
- ✅ `specs/001-create-sdek/data-model.md` (8 entities with validation)
- ✅ `specs/001-create-sdek/contracts/cli-commands.md` (7 command contracts)
- ✅ `specs/001-create-sdek/quickstart.md` (7 validation scenarios)
- ✅ `.github/copilot-instructions.md` (agent context updated)

**Next Action**: Run `/tasks` command to generate tasks.md from contracts, data model, and quickstart scenarios

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
