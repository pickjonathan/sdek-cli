<!--
SYNC IMPACT REPORT
==================
Version Change: 0.0.0 → 1.0.0
Change Type: MAJOR (Initial constitution establishment)

Modified Principles:
- NEW: I. Correctness and Safety
- NEW: II. Configuration Management
- NEW: III. Command Design (Cobra)
- NEW: IV. User Experience & Terminal UI (Bubble Tea)
- NEW: V. Test-Driven Development
- NEW: VI. Performance & Efficiency
- NEW: VII. Cross-Platform Compatibility
- NEW: VIII. Observability & Logging
- NEW: IX. Modularity & Code Organization
- NEW: X. Extensibility & Versioning
- NEW: XI. Documentation & Clarity

Added Sections:
- Technology Stack & Constraints
- Quality Standards
- Governance

Templates Status:
✅ plan-template.md - Reviewed, compatible (Constitution Check section aligns)
✅ spec-template.md - Reviewed, compatible (Requirement completeness aligns with principles)
✅ tasks-template.md - Reviewed, compatible (TDD phase structure matches Principle V)
✅ agent-file-template.md - Reviewed, compatible (Will extract Go/Cobra/Viper/BubbleTea context)

Follow-up TODOs: None

Rationale:
This is the initial establishment of the constitution for the SDEK CLI project.
MAJOR version (1.0.0) is appropriate as this defines the foundational governance,
architectural principles, and quality standards that all future development must follow.
-->

# SDEK CLI Constitution

## Core Principles

### I. Correctness and Safety

All commands MUST validate flags and arguments before execution. The CLI MUST use clear, typed errors with context (`fmt.Errorf("context: %w", err)`) and MUST NOT use panics for error handling. Configuration and I/O errors MUST be handled gracefully and MUST NOT be silently ignored. Command side effects (file writes, network calls, etc.) MUST be logged or confirmed via UX feedback.

**Rationale**: Users rely on CLI tools for critical workflows. Silent failures or cryptic errors erode trust and create debugging nightmares. Explicit validation and contextual errors make the tool predictable and debuggable.

### II. Configuration Management

Viper MUST be the single source of truth for configuration—no manual flag/env duplication. Configuration precedence MUST follow (highest to lowest): CLI flags → environment variables → config file → defaults. The CLI MUST support a `--config` flag and MUST auto-load configuration from standard paths (`$HOME/.appname/config.yaml`). When writing configuration back, the tool SHOULD preserve comments and structure where possible.

**Rationale**: Configuration sprawl leads to inconsistencies and user confusion. A single, predictable precedence hierarchy ensures users can override settings at the appropriate level without surprises.

### III. Command Design (Cobra)

Each command MUST have a concise, descriptive `Use:` string, MUST provide `Short` and `Long` help messages, and SHOULD offer examples in the `Example:` field where applicable. Subcommands MUST group logically (e.g., `app user create`, `app user delete`). Commands MUST use PreRun hooks for setup and validation, and PostRun hooks for cleanup/logging.

**Rationale**: Discoverable, well-documented commands reduce support burden and improve user experience. Logical grouping and hook-based lifecycle management keep command logic clean and maintainable.

### IV. User Experience & Terminal UI (Bubble Tea)

The CLI MUST default to a clean, responsive TUI that works across major terminals (xterm, iTerm2, Windows Terminal). Lip Gloss MUST be used for consistent styling (colors, borders, padding). Layouts MUST be reactive and minimal—the tool MUST NOT block unnecessarily on animations or spinners. Every interactive command MUST offer a non-interactive mode (flags only) for CI/CD or scripting usage. Keyboard shortcuts MUST be intuitive (e.g., `q` to quit, `↑/↓` for selection, `Enter` to confirm). Errors and success states MUST be displayed clearly with visual distinction (e.g., red vs green).

**Rationale**: Terminal UIs significantly improve usability but can also create friction if poorly designed. Non-interactive fallbacks ensure the tool remains automatable. Consistent styling and intuitive controls make the tool feel polished and professional.

### V. Test-Driven Development

Every new command MUST have unit tests for command logic (`RunE` functions), integration tests for command invocation (`cobra.Command.Execute()`), and golden file tests for output rendering (TUI or text). Regression tests MUST be included for any bugs fixed. All code MUST pass `go test ./...` cleanly before merging.

**Rationale**: TDD catches bugs early, documents expected behavior, and enables confident refactoring. Golden file tests ensure UI changes are intentional and visible in reviews.

### VI. Performance & Efficiency

The CLI MUST start quickly (target: under 100ms cold start when possible). The tool MUST NOT spawn unnecessary background goroutines or long-running processes unless explicitly user-triggered. Long-lived TUI sessions MUST be profiled for memory leaks. Remote or expensive operations SHOULD be cached when appropriate (e.g., using `os.UserCacheDir()`).

**Rationale**: Slow startup kills CLI usability. Users expect instant responsiveness. Caching and efficient resource management ensure the tool scales to large projects without degrading performance.

### VII. Cross-Platform Compatibility

The CLI MUST support Linux, macOS, and Windows (PowerShell + WSL). Path operations MUST use only cross-platform-safe APIs (`filepath.Join`, not `/` literals). Terminal sizing and colors MUST be handled gracefully across OSes. For platform-specific system calls or file permissions, the tool MUST detect the platform and provide helpful fallback messages.

**Rationale**: Cross-platform support maximizes reach and reduces maintenance burden. Users expect tools to work consistently regardless of OS.

### VIII. Observability & Logging

The CLI MUST provide a `--verbose` flag to enable debug logs. Logs MUST be emitted to stderr (never stdout unless part of user output). The tool MUST use structured logs (`log/slog` or `zerolog`) with levels: `debug`, `info`, `warn`, `error`. All commands MUST emit telemetry events (if enabled) in JSON format.

**Rationale**: Observability is critical for troubleshooting production issues. Structured logs enable automated parsing and analysis. Separating logs from output ensures pipeable, scriptable commands.

### IX. Modularity & Code Organization

The project MUST follow this structure:
- `cmd/` — Cobra commands
- `internal/` — Business logic and reusable helpers
- `ui/` — Bubble Tea programs and components
- `pkg/` — Exported packages for reuse

Commands MUST be kept as thin as possible—business logic MUST be delegated to services or internal packages. Cyclic dependencies MUST be avoided; UI, config, and core logic MUST remain isolated. Dependency injection or functional options SHOULD be used for flexibility and testing.

**Rationale**: Clear separation of concerns makes the codebase navigable and testable. Thin command layers ensure business logic is reusable and independently testable.

### X. Extensibility & Versioning

Commands SHOULD be easily extensible (e.g., plugin architecture, dynamic subcommands). The tool MUST maintain backward compatibility for flags and command behavior. The CLI MUST expose `--version` and MAY optionally provide a `version` subcommand that includes build metadata. The project MUST use semantic versioning (MAJOR.MINOR.PATCH).

**Rationale**: Backward compatibility ensures user scripts don't break on upgrades. Semantic versioning communicates change impact clearly. Extensibility future-proofs the tool for evolving requirements.

### XI. Documentation & Clarity

The project MUST auto-generate markdown help via Cobra (`cmd/gen-docs` or `cobra doc`). Documentation MUST include examples in README and help output. Non-obvious design decisions MUST be documented in code comments or `docs/`. Code MUST favor readability and simplicity over cleverness—idiomatic Go is the standard.

**Rationale**: Documentation multiplies the tool's usability. Auto-generation ensures docs stay in sync with code. Idiomatic, readable code reduces onboarding friction for contributors.

## Technology Stack & Constraints

**Language**: Go (latest stable version)

**Core Dependencies**:
- **Cobra** — Command structure and help generation
- **Viper** — Configuration and environment variable management
- **Bubble Tea** — Rich terminal interfaces
- **Lip Gloss** — Styling and layout for TUI
- **Bubbles** — TUI components (lists, spinners, text inputs, etc.)

**Constraints**:
- Cold start time MUST be under 100ms (target)
- TUI MUST work on xterm, iTerm2, Windows Terminal
- All paths MUST use `filepath.Join` for cross-platform safety
- All tests MUST pass via `go test ./...`
- Linting MUST pass via `golangci-lint run`

## Quality Standards

**Code Quality**:
- All pull requests MUST pass linting (`golangci-lint run`) and tests
- Code MUST comply with this constitution; maintainers may reject non-compliant code
- Code MUST be idiomatic Go; readability over cleverness

**Testing Requirements**:
- Unit tests for all command logic (`RunE` functions)
- Integration tests for command invocation
- Golden file tests for TUI/text output
- Regression tests for bug fixes
- All tests MUST run cleanly before merging

**Performance Requirements**:
- Cold start under 100ms (target)
- No unnecessary background goroutines
- Profile long-lived TUI sessions for memory leaks
- Cache expensive operations where appropriate

## Governance

This constitution supersedes all other development practices. All code reviews MUST verify compliance with these principles. Amendments to this constitution require:
- Documentation of the change rationale
- At least one maintainer approval
- Migration plan if the change affects existing code

Complexity that violates simplicity principles MUST be justified in design documents. When in doubt, consult this constitution before implementation.

For runtime development guidance (agent-specific workflows, command examples, recent changes), reference the auto-generated agent guidance files (e.g., `CLAUDE.md`, `.github/copilot-instructions.md`, etc.).

**Version**: 1.0.0 | **Ratified**: 2025-10-11 | **Last Amended**: 2025-10-11