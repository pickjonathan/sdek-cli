# Tasks: Create sdek

**Input**: Design documents from `/Users/pickjonathan/WorkSpacePrivate/sdek-cli/specs/001-create-sdek/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Execution Flow
```
1. Load plan.md → Extract tech stack (Go 1.23+, Cobra, Viper, Bubble Tea)
2. Load data-model.md → 8 entities (Source, Event, Framework, Control, Evidence, Finding, User, Config)
3. Load contracts/ → 7 CLI commands (root, tui, ingest, analyze, report, seed, config)
4. Load quickstart.md → 7 validation scenarios
5. Generate tasks by phase:
   → Setup: Project structure, Go modules, dependencies
   → Tests: Entity tests, command tests, integration tests
   → Core: Entity types, business logic, CLI commands
   → Integration: State persistence, configuration, logging
   → Polish: TUI implementation, performance, documentation
6. Apply TDD ordering: Tests before implementation
7. Mark [P] for independent tasks (different files)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- All paths relative to repository root

## Phase 3.1: Setup (Foundation)

- [X] **T001** Initialize Go module and project structure
  - Create `go.mod` with module name `github.com/yourorg/sdek-cli`
  - Create directory structure: `cmd/`, `internal/`, `ui/`, `pkg/types/`, `tests/`
  - Add `.gitignore` for Go projects
  - **Files**: `go.mod`, directory structure
  - **Constitution**: IX (Modularity)

- [X] **T002** Install core dependencies
  - Add Cobra: `github.com/spf13/cobra`
  - Add Viper: `github.com/spf13/viper`
  - Add Bubble Tea: `github.com/charmbracelet/bubbletea`
  - Add Lip Gloss: `github.com/charmbracelet/lipgloss`
  - Add Bubbles: `github.com/charmbracelet/bubbles`
  - Run `go mod tidy`
  - **Files**: `go.mod`, `go.sum`
  - **Constitution**: X (Dependencies)

- [X] **T003** [P] Configure development tooling
  - Create `Makefile` with build, test, run targets
  - Add `golangci-lint` configuration
  - Add VS Code debug configurations
  - **Files**: `Makefile`, `.golangci.yml`, `.vscode/launch.json`
  - **Constitution**: XI (Documentation)

## Phase 3.2: Entity Types (Models First)

- [X] **T004** [P] Create Source entity type
  - Define `Source` struct in `pkg/types/source.go`
  - Add validation function `ValidateSource()`
  - Add constants for source types (Git, Jira, Slack, CICD, Docs)
  - **Files**: `pkg/types/source.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [X] **T005** [P] Create Event entity type
  - Define `Event` struct in `pkg/types/event.go`
  - Add UUID generation helper
  - Add metadata helpers for each source type
  - Add validation function `ValidateEvent()`
  - **Files**: `pkg/types/event.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [X] **T006** [P] Create Framework entity type
  - Define `Framework` struct in `pkg/types/framework.go`
  - Add compliance percentage calculation method
  - Add constants for framework IDs (SOC2, ISO27001, PCIDSS)
  - **Files**: `pkg/types/framework.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [X] **T007** [P] Create Control entity type
  - Define `Control` struct in `pkg/types/control.go`
  - Add risk status enum (Green, Yellow, Red)
  - Add risk score calculation method
  - Add validation function `ValidateControl()`
  - **Files**: `pkg/types/control.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [X] **T008** [P] Create Evidence entity type
  - Define `Evidence` struct in `pkg/types/evidence.go`
  - Add confidence level enum (Low, Medium, High)
  - Add mapping relationship fields
  - **Files**: `pkg/types/evidence.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [X] **T009** [P] Create Finding entity type
  - Define `Finding` struct in `pkg/types/finding.go`
  - Add severity enum (Low, Medium, High, Critical)
  - Add status enum (Open, InProgress, Resolved)
  - **Files**: `pkg/types/finding.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [X] **T010** [P] Create User entity type
  - Define `User` struct in `pkg/types/user.go`
  - Add role enum (ComplianceManager, Engineer)
  - Add predefined users (Alice, Bob, Carol)
  - **Files**: `pkg/types/user.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [X] **T011** [P] Create Config entity type
  - Define `Config` struct in `pkg/types/config.go`
  - Add nested structs for export, frameworks, sources settings
  - Add default configuration constants
  - Add validation function `ValidateConfig()`
  - **Files**: `pkg/types/config.go`
  - **Constitution**: II (Configuration), IX (Modularity)

## Phase 3.3: Entity Tests (TDD)

- [X] **T012** [P] Write Source entity tests
  - Test validation rules (ID must be valid, EventCount 10-50)
  - Test state transitions
  - **Files**: `pkg/types/source_test.go`
  - **Constitution**: V (TDD)

- [X] **T013** [P] Write Event entity tests
  - Test UUID generation
  - Test timestamp validation
  - Test metadata helpers for each source type
  - **Files**: `pkg/types/event_test.go`
  - **Constitution**: V (TDD)

- [X] **T014** [P] Write Framework entity tests
  - Test compliance calculation
  - Test framework ID validation
  - **Files**: `pkg/types/framework_test.go`
  - **Constitution**: V (TDD)

- [X] **T015** [P] Write Control entity tests
  - Test risk score calculation
  - Test status determination logic
  - **Files**: `pkg/types/control_test.go`
  - **Constitution**: V (TDD)

- [X] **T016** [P] Write Evidence, Finding, User, Config entity tests
  - Test confidence level calculations
  - Test finding severity logic
  - Test user role permissions
  - Test config validation
  - **Files**: `pkg/types/evidence_test.go`, `pkg/types/finding_test.go`, `pkg/types/user_test.go`, `pkg/types/config_test.go`
  - **Constitution**: V (TDD)

## Phase 3.4: State & Storage Layer

- [ ] **T017** Create state persistence layer
  - Implement `State` struct in `internal/store/state.go`
  - Add JSON marshaling/unmarshaling
  - Add state file operations (Load, Save)
  - Add error handling with context wrapping
  - **Files**: `internal/store/state.go`
  - **Constitution**: I (Safety), IX (Modularity)

- [ ] **T018** Implement auto-save functionality
  - Create `AutoSave` in `internal/store/autosave.go`
  - Add debounce logic (save after 2s idle)
  - Add graceful shutdown handling
  - **Files**: `internal/store/autosave.go`
  - **Constitution**: I (Safety), VI (Performance)

- [ ] **T019** Implement state cache layer
  - Create `Cache` in `internal/store/cache.go`
  - Add in-memory indexes (by source, by framework, by control)
  - Add cache invalidation logic
  - **Files**: `internal/store/cache.go`
  - **Constitution**: VI (Performance), IX (Modularity)

- [ ] **T020** Write storage layer tests
  - Test state save/load operations
  - Test auto-save debounce timing
  - Test cache hit/miss scenarios
  - **Files**: `internal/store/state_test.go`, `internal/store/autosave_test.go`, `internal/store/cache_test.go`
  - **Constitution**: V (TDD)

## Phase 3.5: Configuration Management

- [ ] **T021** Implement Viper configuration loader
  - Create `ConfigLoader` in `internal/config/loader.go`
  - Implement precedence: flags → env → config file → defaults
  - Add environment variable binding (SDEK_*)
  - Add config file discovery ($HOME/.sdek/config.yaml)
  - **Files**: `internal/config/loader.go`
  - **Constitution**: II (Configuration), IX (Modularity)

- [ ] **T022** Implement configuration validation
  - Create `Validator` in `internal/config/validator.go`
  - Validate log levels, paths, enabled sources/frameworks
  - Return descriptive errors for invalid configs
  - **Files**: `internal/config/validator.go`
  - **Constitution**: I (Correctness), II (Configuration)

- [ ] **T023** [P] Write configuration tests
  - Test precedence order (flags override env override config)
  - Test environment variable binding
  - Test validation error messages
  - **Files**: `internal/config/loader_test.go`, `internal/config/validator_test.go`
  - **Constitution**: V (TDD)

## Phase 3.6: Data Generation (Ingest)

- [ ] **T024** Create base data generator
  - Implement `Generator` interface in `internal/ingest/generator.go`
  - Add deterministic seeding logic
  - Add timestamp generation (within 90 days)
  - Add common event generation helpers
  - **Files**: `internal/ingest/generator.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [ ] **T025** [P] Implement source-specific generators
  - Create `git.go` for Git commits (SHA, branch, files changed)
  - Create `jira.go` for Jira tickets (ticket ID, status, priority)
  - Create `slack.go` for Slack messages (channel, thread, reactions)
  - Create `cicd.go` for CI/CD pipelines (pipeline ID, status, duration)
  - Create `docs.go` for documentation (file path, change type, reviewer)
  - Each generates 10-50 events with realistic content
  - **Files**: `internal/ingest/git.go`, `internal/ingest/jira.go`, `internal/ingest/slack.go`, `internal/ingest/cicd.go`, `internal/ingest/docs.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [ ] **T026** [P] Write ingest tests
  - Test deterministic generation with same seed
  - Test event count boundaries (10-50)
  - Test timestamp ranges
  - Test source-specific metadata
  - **Files**: `internal/ingest/generator_test.go`, `internal/ingest/git_test.go`, etc.
  - **Constitution**: V (TDD)

## Phase 3.7: Evidence Mapping (Analyze)

- [ ] **T027** Implement framework definitions
  - Create `frameworks.go` in `internal/analyze/frameworks.go`
  - Define SOC2 controls (45 controls, categories)
  - Define ISO 27001 controls (60 controls)
  - Define PCI DSS controls (15 controls)
  - Add control keyword lists for mapping
  - **Files**: `internal/analyze/frameworks.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [ ] **T028** Implement evidence mapper
  - Create `Mapper` in `internal/analyze/mapper.go`
  - Implement keyword-based heuristic matching
  - Map events to controls based on keywords
  - Create evidence records with confidence scores
  - **Files**: `internal/analyze/mapper.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [ ] **T029** Implement confidence calculation
  - Create `ConfidenceCalculator` in `internal/analyze/confidence.go`
  - Calculate confidence based on keyword matches, recency, source type
  - Return Low (0-50%), Medium (51-75%), High (76-100%)
  - **Files**: `internal/analyze/confidence.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [ ] **T030** Implement risk scoring
  - Create `RiskScorer` in `internal/analyze/risk.go`
  - Implement severity-weighted algorithm (3H=1C, 6M=1C, 12L=1C)
  - Calculate control risk status (Green, Yellow, Red)
  - Generate findings for red/yellow controls
  - **Files**: `internal/analyze/risk.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [ ] **T031** [P] Write analyze tests
  - Test keyword matching accuracy
  - Test confidence calculation
  - Test risk scoring formula
  - Test finding generation rules
  - **Files**: `internal/analyze/mapper_test.go`, `internal/analyze/confidence_test.go`, `internal/analyze/risk_test.go`
  - **Constitution**: V (TDD)

## Phase 3.8: Report Export

- [ ] **T032** Implement report exporter
  - Create `Exporter` in `internal/report/exporter.go`
  - Generate JSON report with frameworks, controls, evidence, findings
  - Add metadata (generated timestamp, version)
  - Add summary statistics
  - Save to configured export path
  - **Files**: `internal/report/exporter.go`
  - **Constitution**: I (Correctness), IX (Modularity)

- [ ] **T033** Implement report formatter
  - Create `Formatter` in `internal/report/formatter.go`
  - Add JSON pretty-printing
  - Add role-based filtering (manager vs engineer views)
  - **Files**: `internal/report/formatter.go`
  - **Constitution**: IV (UX), IX (Modularity)

- [ ] **T034** [P] Write report tests
  - Test JSON structure matches schema
  - Test summary calculations
  - Test role-based filtering
  - Test file save operations
  - **Files**: `internal/report/exporter_test.go`, `internal/report/formatter_test.go`
  - **Constitution**: V (TDD)

## Phase 3.9: CLI Commands (Cobra)

- [ ] **T035** Implement root command
  - Create `root.go` in `cmd/root.go`
  - Add global flags (--config, --data-dir, --log-level, --verbose, --version)
  - Initialize Viper configuration
  - Set up structured logging (log/slog)
  - Add version command
  - **Files**: `cmd/root.go`
  - **Constitution**: II (Config), III (Cobra), VIII (Logging)

- [ ] **T036** Implement seed command
  - Create `seed.go` in `cmd/seed.go`
  - Add flags (--demo, --seed, --reset)
  - Generate demo data (5 sources, 130 events, 3 frameworks, 245 evidence, 18 findings)
  - Save to state file
  - Print summary output
  - **Files**: `cmd/seed.go`
  - **Constitution**: III (Cobra), I (Safety)

- [ ] **T037** Implement ingest command
  - Create `ingest.go` in `cmd/ingest.go`
  - Add flags (--source, --events, --seed)
  - Call data generators for specified sources
  - Update state with new events
  - Print ingest summary
  - **Files**: `cmd/ingest.go`
  - **Constitution**: III (Cobra), VIII (Logging)

- [ ] **T038** Implement analyze command
  - Create `analyze.go` in `cmd/analyze.go`
  - Load events from state
  - Run evidence mapper
  - Calculate risk scores
  - Generate findings
  - Update state with analysis results
  - Print analysis summary
  - **Files**: `cmd/analyze.go`
  - **Constitution**: III (Cobra), VIII (Logging)

- [ ] **T039** Implement report command
  - Create `report.go` in `cmd/report.go`
  - Add flags (--output)
  - Load analysis results from state
  - Generate JSON report
  - Save to specified path (default: $HOME/sdek/reports/)
  - Print export summary
  - **Files**: `cmd/report.go`
  - **Constitution**: III (Cobra), I (Safety)

- [ ] **T040** Implement config command
  - Create `config.go` in `cmd/config.go`
  - Add subcommands: init, get, set, list, validate
  - Implement config file operations
  - Print config values/errors
  - **Files**: `cmd/config.go`
  - **Constitution**: II (Config), III (Cobra)

- [ ] **T041** [P] Write command tests
  - Test flag parsing for each command
  - Test exit codes (0, 1, 2, 3, 4, 130)
  - Test error messages
  - Test output formats
  - **Files**: `cmd/root_test.go`, `cmd/seed_test.go`, `cmd/ingest_test.go`, `cmd/analyze_test.go`, `cmd/report_test.go`, `cmd/config_test.go`
  - **Constitution**: V (TDD), III (Cobra)

## Phase 3.10: Terminal UI (Bubble Tea)

- [ ] **T042** Create TUI application structure
  - Create `app.go` in `ui/app.go`
  - Define main Bubble Tea model
  - Implement Init, Update, View methods
  - Add terminal size detection (minimum 80×24)
  - Add screen navigation state machine
  - **Files**: `ui/app.go`
  - **Constitution**: IV (TUI), IX (Modularity)

- [ ] **T043** Implement home screen model
  - Create `home.go` in `ui/models/home.go`
  - Display three sections (Sources, Frameworks, Findings)
  - Implement Tab navigation between sections
  - Show summary data for each section
  - **Files**: `ui/models/home.go`
  - **Constitution**: IV (TUI), IX (Modularity)

- [ ] **T044** Implement list models
  - Create `sources.go` in `ui/models/sources.go` for sources list
  - Create `frameworks.go` in `ui/models/frameworks.go` for frameworks list
  - Create `controls.go` in `ui/models/controls.go` for controls detail view
  - Create `evidence.go` in `ui/models/evidence.go` for evidence list
  - Implement ↑/↓ navigation, Enter to select, ← to go back
  - **Files**: `ui/models/sources.go`, `ui/models/frameworks.go`, `ui/models/controls.go`, `ui/models/evidence.go`
  - **Constitution**: IV (TUI), IX (Modularity)

- [ ] **T045** Create reusable UI components
  - Create `card.go` in `ui/components/card.go` for evidence cards
  - Create `riskbar.go` in `ui/components/riskbar.go` for risk visualization
  - Create `list.go` in `ui/components/list.go` for custom list rendering
  - Create `statusbar.go` in `ui/components/statusbar.go` for keyboard shortcuts
  - **Files**: `ui/components/card.go`, `ui/components/riskbar.go`, `ui/components/list.go`, `ui/components/statusbar.go`
  - **Constitution**: IV (TUI), IX (Modularity)

- [ ] **T046** Implement TUI styling
  - Create `theme.go` in `ui/styles/theme.go`
  - Define Lip Gloss styles for colors, borders, spacing
  - Add risk status colors (green=#00ff00, yellow=#ffff00, red=#ff0000)
  - Add dark/light theme support
  - **Files**: `ui/styles/theme.go`
  - **Constitution**: IV (TUI), IX (Modularity)

- [ ] **T047** Implement keyboard handling
  - Add keyboard shortcuts to TUI models (q=quit, r=refresh, e=export, /=search)
  - Add SIGINT handling (Ctrl+C)
  - Add help screen toggle (?)
  - **Files**: `ui/app.go`, `ui/models/*.go`
  - **Constitution**: IV (TUI), I (Safety)

- [ ] **T048** Implement TUI command
  - Create `tui.go` in `cmd/tui.go`
  - Add flags (--role)
  - Load state from file
  - Initialize Bubble Tea program
  - Handle auto-save on exit
  - **Files**: `cmd/tui.go`
  - **Constitution**: III (Cobra), IV (TUI)

- [ ] **T049** [P] Write TUI golden file tests
  - Create golden file test fixtures in `tests/golden/fixtures/`
  - Test home screen rendering
  - Test control list rendering
  - Test evidence detail rendering
  - Compare output to golden files
  - **Files**: `tests/golden/tui_home_test.go`, `tests/golden/tui_controls_test.go`, `tests/golden/fixtures/*.golden`
  - **Constitution**: V (TDD), IV (TUI)

## Phase 3.11: Integration & E2E Tests

- [ ] **T050** Write CLI integration tests
  - Test workflow: seed → ingest → analyze → report
  - Test configuration precedence (flags → env → config → defaults)
  - Test state persistence across commands
  - Test error recovery scenarios
  - **Files**: `tests/integration/flow_test.go`, `tests/integration/config_test.go`
  - **Constitution**: V (TDD), I (Correctness)

- [ ] **T051** [P] Implement quickstart scenario tests
  - Scenario 1: First-time setup with demo data
  - Scenario 2: Interactive TUI navigation (simulate keypresses)
  - Scenario 3: CLI workflow (ingest → analyze → report)
  - Scenario 4: Configuration management
  - Scenario 5: Role-based visibility
  - Scenario 6: Error handling
  - Scenario 7: Performance validation
  - **Files**: `tests/integration/quickstart_test.go`
  - **Constitution**: V (TDD), IV (TUI)

## Phase 3.12: Main Entry Point & Build

- [ ] **T052** Create main.go entry point
  - Create `main.go` at repository root
  - Initialize root command
  - Set up panic recovery
  - Handle exit codes
  - **Files**: `main.go`
  - **Constitution**: I (Safety), VI (Performance)

- [ ] **T053** Add build configuration
  - Update `Makefile` with build targets (build, test, install, clean)
  - Add version injection via ldflags
  - Add cross-compilation targets (Linux, macOS, Windows)
  - **Files**: `Makefile`
  - **Constitution**: X (Versioning), VII (Cross-platform)

## Phase 3.13: Documentation & Polish

- [ ] **T054** [P] Write README.md
  - Add project overview
  - Add installation instructions
  - Add usage examples for all commands
  - Add TUI screenshots (ASCII art)
  - Add development setup guide
  - **Files**: `README.md`
  - **Constitution**: XI (Documentation)

- [ ] **T055** [P] Generate command documentation
  - Use Cobra's auto-generation: `cobra-cli gendocs`
  - Create `docs/` directory with command reference
  - Add examples for each command
  - **Files**: `docs/*.md`
  - **Constitution**: XI (Documentation)

- [ ] **T056** [P] Performance optimization
  - Profile cold start time (target <100ms)
  - Profile TUI rendering (target 60fps)
  - Add benchmarks for data generation, analysis, report export
  - Optimize any bottlenecks
  - **Files**: `*_bench_test.go` files
  - **Constitution**: VI (Performance)

- [ ] **T057** [P] Final validation
  - Run all tests: `go test ./...`
  - Run linter: `golangci-lint run`
  - Build for all platforms: `make build-all`
  - Execute quickstart scenarios manually
  - Verify <100ms startup time
  - Verify TUI runs smoothly at 60fps
  - **Files**: N/A (validation only)
  - **Constitution**: ALL

---

## Parallel Execution Examples

**Phase 3.2 (Entity Types)** - All can run in parallel:
```bash
# Tasks T004-T011 are independent (different files)
# Can be executed simultaneously
```

**Phase 3.3 (Entity Tests)** - All can run in parallel:
```bash
# Tasks T012-T016 are independent
```

**Phase 3.6 (Source Generators)** - Partial parallelization:
```bash
# T024 must complete first (base generator)
# Then T025 (all source-specific generators) can run in parallel
# Then T026 (tests) can run in parallel
```

**Phase 3.13 (Documentation)** - All can run in parallel:
```bash
# Tasks T054-T056 are independent
```

## Dependency Graph

```
T001 → T002 → T003
       ↓
T004-T011 [P] (Entity types)
       ↓
T012-T016 [P] (Entity tests)
       ↓
T017 → T018 → T019 → T020 (Storage layer)
       ↓
T021 → T022 → T023 [P] (Configuration)
       ↓
T024 → T025 [P] → T026 [P] (Ingest)
       ↓
T027 → T028 → T029 → T030 → T031 [P] (Analyze)
       ↓
T032 → T033 → T034 [P] (Report)
       ↓
T035 → T036-T040 → T041 [P] (CLI commands)
       ↓
T042 → T043 → T044 → T045 → T046 → T047 → T048 → T049 [P] (TUI)
       ↓
T050 → T051 [P] (Integration tests)
       ↓
T052 → T053 (Main & Build)
       ↓
T054-T057 [P] (Documentation & Polish)
```

## Task Execution Checklist

**Setup**: T001-T003 (3 tasks)
**Core Types**: T004-T016 (13 tasks)
**Storage**: T017-T020 (4 tasks)
**Configuration**: T021-T023 (3 tasks)
**Business Logic**: T024-T034 (11 tasks)
**CLI**: T035-T041 (7 tasks)
**TUI**: T042-T049 (8 tasks)
**Integration**: T050-T051 (2 tasks)
**Build**: T052-T053 (2 tasks)
**Polish**: T054-T057 (4 tasks)

**Total**: 57 tasks

---

*Generated from plan.md (Go 1.23+, Cobra, Viper, Bubble Tea, Lip Gloss)*
*Follows SDEK CLI Constitution v1.0.0 - TDD, <100ms startup, 60fps TUI*
