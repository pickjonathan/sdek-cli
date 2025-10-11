# Feature Specification: Create sdek

**Feature Branch**: `001-create-sdek`  
**Created**: 2025-10-11  
**Status**: Draft  
**Input**: User description: "Develop sdek, a CLI and terminal UI tool that reduces audit preparation time by 30%. It does this by ingesting signals from systems like Git, CI/CD, Jira, Slack, and documentation sources; then using AI to map those unstructured signals to compliance frameworks such as SOC2, ISO 27001, and PCI DSS. The goal is to make compliance evidence collection and risk visualization effortless."

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## Clarifications

### Session 2025-10-11
- Q: What are the data generation rules for simulated sources? → A: Each simulated data source (Git, Jira, Slack, CI/CD, Docs) generates 10-50 events with random distribution. Confidence levels (Low, Medium, High) are randomly distributed across evidence mappings. Each compliance framework (SOC2, ISO, PCI) must have at least one complete control (green), one partial control (yellow), and one missing control (red) to ensure all risk states are represented.
- Q: How should evidence completeness determine control risk status? → A: Severity-based risk scoring: If any critical issue exists, status cannot be green. Equivalencies: 3 high-severity issues = 1 critical, 6 medium-severity issues = 1 critical, 12 low-severity issues = 1 critical.
- Q: What are the minimum terminal dimensions for the TUI? → A: 80 columns × 24 rows (classic terminal size).
- Q: What should be the default export filename and location? → A: $HOME/sdek/reports/sdek-report-YYYY-MM-DD.json
- Q: How should "recent events" be defined for the simulated data display? → A: All events (10-50 items per source).

---

## User Scenarios & Testing

### Primary User Story

A compliance manager needs to prepare for an upcoming SOC2 audit. Instead of manually searching through Git commits, Jira tickets, and Slack conversations to gather evidence for each control, they launch sdek in their terminal. The tool shows them a visual dashboard of compliance frameworks, automatically maps available evidence to controls, and highlights gaps. They can export a comprehensive evidence report in minutes rather than days.

An engineering manager wants to understand which development activities contribute to compliance requirements. They use sdek's TUI to view recent code commits and see how they map to security controls. They can drill down into specific evidence items, assess confidence levels, and identify where documentation is missing.

### Acceptance Scenarios

1. **Given** the user has installed sdek and has no prior configuration, **When** they run `sdek tui` for the first time, **Then** the system displays a terminal UI home screen with three main sections (Sources, Frameworks, Findings) using simulated data, and the user can navigate between sections using Tab.

2. **Given** the user is viewing the Frameworks section, **When** they select "SOC2" and press Enter, **Then** the system displays a list of SOC2 controls, each showing mapped evidence count, confidence level (green/yellow/red), and remaining risk indicators.

3. **Given** the user is viewing a specific control, **When** they navigate through the evidence list, **Then** the system displays detailed information about each evidence item including source (Git/Jira/Slack), timestamp, and relevance score.

4. **Given** the user wants to generate a compliance report, **When** they run `sdek report --format json`, **Then** the system exports all evidence mappings and compliance status to a JSON file in the current directory.

5. **Given** the user wants to simulate new data ingestion, **When** they press 'r' in the TUI or run `sdek ingest`, **Then** the system updates the evidence cache and refreshes the display showing new evidence items and updated compliance scores.

6. **Given** the user is viewing as a compliance manager role, **When** they navigate the UI, **Then** they see summary views with high-level compliance status and risk indicators.

7. **Given** the user is viewing as an engineer role, **When** they navigate the UI, **Then** they see technical evidence details including commit SHAs, ticket IDs, and implementation specifics.

8. **Given** the user has launched the TUI, **When** they press 'q' at any screen, **Then** the application exits cleanly and returns to the shell prompt.

9. **Given** the user runs any CLI command, **When** they include the `--verbose` flag, **Then** the system outputs detailed debug information to stderr while maintaining clean output on stdout.

10. **Given** the user has set configuration values, **When** they run commands, **Then** the system applies configuration in precedence order: CLI flags, then environment variables, then config file, then defaults.

### Edge Cases

- What happens when the evidence cache file is corrupted or missing? System should regenerate from default simulated dataset and display a warning.
- What happens when the user's terminal window is too small to display the TUI properly? System should detect if dimensions are below 80 columns × 24 rows and display a helpful error message with required size.
- What happens when the user tries to export evidence to a directory without write permissions? System should display a clear error message and suggest alternative locations.
- What happens when the user rapidly presses keys in the TUI? System should queue inputs gracefully without crashing or displaying artifacts.
- What happens when the config file contains invalid YAML? System should display a clear error with line number and fall back to defaults.
- What happens when the user runs multiple sdek commands simultaneously? Each command should operate independently on the shared evidence cache with last-write-wins semantics.
- What happens when no simulated data is available for a selected framework? System should display "No evidence mapped yet" message and suggest running ingest command.

---

## Requirements

### Functional Requirements

#### Core Application
- **FR-001**: System MUST provide a command-line interface with four primary commands: `ingest`, `analyze`, `report`, and `tui`.
- **FR-002**: System MUST launch a terminal-based user interface when the `tui` command is executed.
- **FR-003**: System MUST display three main sections in the TUI home screen: Sources, Frameworks, and Findings.
- **FR-004**: System MUST support keyboard navigation using arrow keys for selection, Enter to open items, 'q' to quit, Tab to switch sections, 'r' to refresh, and 'e' to export.

#### Data Sources
- **FR-005**: System MUST simulate five data source types: Git commits, Jira tickets, Slack messages, CI/CD pipeline events, and documentation sources using preloaded datasets. Each source MUST generate between 10 and 50 events with random distribution.
- **FR-006**: System MUST display all simulated events from each source when a source is selected in the TUI (complete 10-50 item set per source).
- **FR-007**: System MUST show how each source event maps to evidence items when viewing source details.

#### Compliance Frameworks
- **FR-008**: System MUST support three compliance frameworks: SOC2, ISO 27001, and PCI DSS. Each framework MUST have at least one control with complete evidence (green), one control with partial evidence (yellow), and one control with missing evidence (red) to demonstrate all risk visualization states.
- **FR-009**: System MUST display a control list for each framework showing control ID, description, and compliance status.
- **FR-010**: System MUST show mapped evidence, confidence level, and remaining risk for each control when selected.
- **FR-011**: System MUST use color coding to indicate compliance status based on severity-weighted risk scoring: green for compliant (no critical-equivalent issues), yellow for partial compliance, red for missing/critical gaps. System MUST calculate critical-equivalent issues using: 1 critical issue, OR 3 high-severity issues, OR 6 medium-severity issues, OR 12 low-severity issues equals one critical-equivalent threshold. Any critical-equivalent issue prevents green status.

#### Evidence Mapping
- **FR-012**: System MUST map simulated source events to framework controls deterministically (no actual AI required in this phase).
- **FR-013**: System MUST display confidence levels for each evidence mapping indicating strength of the relationship. Confidence levels MUST be randomly distributed across three tiers: Low, Medium, and High.
- **FR-014**: System MUST aggregate evidence across multiple sources for each control.
- **FR-015**: System MUST identify and display compliance gaps where controls lack sufficient evidence.
- **FR-016**: System MUST track issue severity levels (critical, high, medium, low) for each compliance gap and calculate severity-weighted risk scores using the equivalency formula: 3 high = 1 critical, 6 medium = 1 critical, 12 low = 1 critical.

#### User Roles
- **FR-017**: System MUST predefine three simulated users: one compliance manager and two engineers.
- **FR-018**: System MUST adjust visibility of information based on selected user role: managers see summaries, engineers see technical details.
- **FR-019**: System MUST allow switching between user roles within the TUI settings menu.

#### CLI Commands
- **FR-020**: The `ingest` command MUST simulate loading new data from sources and update the evidence cache.
- **FR-021**: The `analyze` command MUST simulate running AI mapping to update evidence-to-control relationships.
- **FR-022**: The `report` command MUST export evidence data in JSON format to a specified file or stdout.
- **FR-023**: All CLI commands MUST accept `--verbose` flag to enable detailed logging.
- **FR-024**: All CLI commands MUST accept `--config` flag to specify a custom configuration file path.

#### Configuration Management
- **FR-025**: System MUST support configuration via CLI flags, environment variables, config file, and defaults in that precedence order.
- **FR-026**: System MUST store configuration file at `$HOME/.sdek/config.yaml` by default.
- **FR-027**: System MUST create the config directory and file if they don't exist on first run.
- **FR-028**: System MUST support configuration options for: color themes, verbosity levels, default export paths (default: `$HOME/sdek/reports/`), and user role selection.

#### Data Persistence
- **FR-029**: System MUST persist evidence cache as a local JSON file in `$HOME/.sdek/evidence.json`.
- **FR-030**: System MUST auto-save evidence cache after each command execution.
- **FR-031**: System MUST load evidence cache on startup and regenerate from defaults if missing or corrupted.

#### Terminal UI Experience
- **FR-032**: System MUST render the TUI with consistent styling using cards for evidence items and bars for risk visualization.
- **FR-033**: System MUST respond to terminal resize events and adjust layout accordingly.
- **FR-034**: System MUST work correctly on xterm, iTerm2, and Windows Terminal.
- **FR-035**: System MUST display clear status messages during loading, processing, and error states.
- **FR-036**: System MUST provide visual feedback for user actions (selections, exports, refreshes).

#### Export Functionality
- **FR-037**: System MUST export evidence reports in JSON format with structured data for all mappings. Default export location is `$HOME/sdek/reports/sdek-report-YYYY-MM-DD.json` where YYYY-MM-DD is the current date. System MUST create the reports directory if it doesn't exist.
- **FR-038**: System MUST include metadata in exports: export timestamp, sdek version, user role, and data sources included.
- **FR-039**: System MUST allow exporting from both CLI (`report` command) and TUI ('e' key).

#### Error Handling
- **FR-040**: System MUST display clear error messages for invalid commands or arguments.
- **FR-041**: System MUST validate configuration file syntax and display helpful error messages for invalid YAML.
- **FR-042**: System MUST handle missing or corrupted evidence cache gracefully by regenerating defaults.
- **FR-043**: System MUST detect insufficient terminal dimensions and display minimum size requirements. Minimum supported dimensions are 80 columns × 24 rows.

### Non-Functional Requirements

#### Performance
- **NFR-001**: System MUST start the CLI within 100ms on modern hardware (cold start target).
- **NFR-002**: System MUST render TUI updates within 16ms to maintain 60fps responsiveness.
- **NFR-003**: System MUST handle evidence caches with up to 10,000 evidence items without performance degradation.

#### Usability
- **NFR-004**: Terminal UI MUST be intuitive for users unfamiliar with the tool within 5 minutes of use.
- **NFR-005**: All keyboard shortcuts MUST be discoverable through help screens or status bar hints.
- **NFR-006**: Visual styling MUST be consistent across all screens with clear hierarchy and spacing.

#### Reliability
- **NFR-007**: System MUST not crash or corrupt data under normal operating conditions.
- **NFR-008**: System MUST exit cleanly on SIGTERM or SIGINT signals.
- **NFR-009**: System MUST handle concurrent command executions safely with last-write-wins for shared data.

#### Maintainability
- **NFR-010**: Code MUST follow the project constitution's modularity principles with clear separation of concerns.
- **NFR-011**: All commands MUST have unit test coverage for RunE logic.
- **NFR-012**: TUI rendering MUST have golden file tests to catch visual regressions.

### Key Entities

- **Source**: Represents a data integration point (Git, Jira, Slack, CI/CD, Docs). Contains: source type, display name, connection status, last sync timestamp, event count (10-50 per source).

- **Event**: A discrete signal from a source (commit, ticket, message). Contains: event ID, source reference, timestamp, event type, content summary, author/creator, extracted metadata.

- **Framework**: A compliance standard (SOC2, ISO 27001, PCI DSS). Contains: framework ID, full name, version, control count, overall compliance percentage.

- **Control**: A specific compliance requirement within a framework. Contains: control ID, description, category, severity level, mapped evidence list, confidence score, risk status.

- **Evidence**: A mapping between an event and a control. Contains: evidence ID, source event reference, target control reference, confidence level (0-100), relevance score, extraction timestamp, mapping rationale.

- **Finding**: An identified compliance gap or risk. Contains: finding ID, affected control, severity level (critical/high/medium/low), gap description, recommended actions, evidence deficit count, critical-equivalent score.

- **User**: Simulated user with role-based view preferences. Contains: user ID, display name, role (compliance_manager/engineer), visibility preferences, active status.

- **Config**: Application configuration. Contains: color theme, verbosity level, default export path, selected user role, framework preferences, cache paths.

---

## Review & Acceptance Checklist

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous  
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

---

## Notes

**Assumptions**:
- This is Phase 1 of the sdek project focused on establishing foundational architecture and UX patterns
- No real external system integrations required; all data is simulated
- No authentication or multi-tenancy required in this phase
- Simulated AI mapping uses deterministic rules rather than ML models
- Local-only deployment; no server or network components
- Single user at a time; no concurrent multi-user scenarios

**Out of Scope**:
- Real integrations with Git, Jira, Slack, or CI/CD systems
- Actual AI/ML inference or API calls to AI services
- User authentication or access control
- Multi-tenancy or team collaboration features
- Web-based UI or API endpoints
- Data synchronization across devices
- Audit trail or change history tracking
- Custom framework definitions or control customization

**Success Criteria**:
- Users can navigate the TUI intuitively using keyboard shortcuts
- Evidence mapping visualization clearly shows compliance status at a glance
- CLI commands execute reliably with clear output and error messages
- Configuration management works correctly with proper precedence
- Export functionality produces valid, structured JSON reports
- Performance meets <100ms startup and 60fps TUI responsiveness targets
- All functional requirements are testable and verifiable
- Architecture supports future phases with real integrations and AI

**Key Benefits**:
- Reduces audit preparation time by centralizing evidence visualization
- Makes compliance status transparent and actionable for multiple roles
- Provides scriptable CLI interface for automation workflows
- Establishes patterns for future real-world integration phases
- Validates UX approach before committing to external dependencies
