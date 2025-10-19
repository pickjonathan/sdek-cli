# Feature Specification: MCP-Native Agent Orchestrator & Tooling Config

**Feature Branch**: `004-mcp-native-agent`  
**Created**: 2025-10-19  
**Status**: Draft  
**Input**: User description: "The agent orchestrator and agents its using should have connections to tools via MCP's, meaning the sdek agents should have MCP configurations like vs code and cursor in the same structure of the json configurations of MCP. the agents should use the MCP's to collect evidance for analysis."

## Execution Flow (main)
```
1. Parse user description from Input
   → Description clearly specifies MCP-based agent orchestration
2. Extract key concepts from description
   → Actors: agent orchestrator, agents, compliance engineers, security teams
   → Actions: configure MCP tools, collect evidence, validate configs, enforce RBAC
   → Data: MCP JSON configs, evidence, audit logs
   → Constraints: Config compatibility with VS Code/Cursor, RBAC enforcement
3. For each unclear aspect:
   → All aspects sufficiently specified in user's detailed description
4. Fill User Scenarios & Testing section
   → Multiple user flows identified for different personas
5. Generate Functional Requirements
   → All requirements testable and aligned with MCP spec compatibility
6. Identify Key Entities
   → MCP configs, tools, agents, evidence, audit logs
7. Run Review Checklist
   → No [NEEDS CLARIFICATION] markers
   → Implementation details avoided, focused on capabilities
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## Clarifications

### Session 2025-10-19

- Q: What is the maximum acceptable latency for a typical MCP tool invocation (excluding external API time)? → A: < 5 seconds (tolerable for long-running compliance scans)
- Q: What are the expected scalability targets? → A: Medium scale: 20-50 MCP tools, 100 invocations/second max
- Q: What are the data volume assumptions for audit logs? → A: Short-term: 7 days retention, expect ~10K invocations/day (default); retention should be configurable
- Q: What is the MCP protocol version compatibility strategy? → A: Best-effort - attempt compatibility, warn on mismatch
- Q: What is the maximum acceptable time for a degraded tool to recover to "ready" status? → A: Immediate: < 30 seconds (fast failover required)

---

## Problem Statement

Currently, sdek-cli's AI analysis uses direct connector calls embedded in the application. This approach creates several challenges:

1. **Lack of Interoperability**: Tool connections are proprietary and cannot be reused across environments
2. **Configuration Complexity**: Each tool requires custom integration code rather than standardized configuration
3. **Limited Portability**: Organizations cannot leverage existing MCP configurations from VS Code, Cursor, or other MCP-compatible tools
4. **Maintenance Burden**: Adding new tools requires code changes rather than configuration updates
5. **No Standard Protocol**: Evidence collection tools don't follow industry-standard Model Context Protocol (MCP)

The goal is to enable sdek-cli's agent orchestrator and agents to connect to tools using the Model Context Protocol (MCP) with JSON configurations that mirror the structure used by VS Code and Cursor. This creates a standardized, portable approach to tool integration that allows compliance teams to reuse configurations across their entire toolchain.

---

## User Scenarios & Testing *(mandatory)*

### Primary User Stories

**As a Platform/Compliance Engineer:**
- I want to drop VS Code/Cursor-style MCP JSON configurations into `~/.sdek/mcp/` so that the orchestrator automatically discovers and provisions tools for agents without code changes
- I want the same MCP configuration to work in both my IDE and sdek-cli so that I maintain consistency across my compliance tooling ecosystem
- I want to validate MCP configurations before deployment so that I can catch errors early

**As a Compliance Manager:**
- I want to run compliance analyses that automatically gather evidence via MCP tools so that I have comprehensive, auditable evidence trails
- I want to see which MCP tools were used to collect evidence so that I can verify the data sources in audit reports
- I want redaction policies to apply to MCP-collected evidence so that sensitive data remains protected

**As a Security Engineer:**
- I want to control which agents can invoke specific MCP tools via RBAC so that I can enforce least-privilege access
- I want to view audit logs of all MCP tool invocations so that I can investigate security incidents
- I want to set rate limits and concurrency controls on MCP tools so that I can prevent resource exhaustion

### Acceptance Scenarios

#### Scenario 1: Configuration Discovery and Loading
1. **Given** a valid VS Code/Cursor-style MCP config exists in `~/.sdek/mcp/github-tool.json`
   **When** the orchestrator starts or hot-reloads
   **Then** the tool appears in `sdek mcp list` with status "ready"
   **And** the tool is available for agent invocation

#### Scenario 2: Schema Validation
1. **Given** an MCP config file with missing required fields (e.g., missing "command" key)
   **When** I run `sdek mcp validate ./bad-config.json`
   **Then** I receive a validation error specifying the file path, line number, and missing property
   **And** the tool is not loaded into the registry

#### Scenario 3: Orchestrator Resilience
1. **Given** an MCP server is temporarily unavailable during startup
   **When** the orchestrator initializes
   **Then** it marks the tool as "degraded" and retries with exponential backoff
   **When** the server becomes available
   **Then** the tool transitions to "ready" without requiring an orchestrator restart

#### Scenario 4: RBAC Enforcement
1. **Given** agent "evidence-collector" is configured without "aws.iam.list" permission
   **When** the agent attempts to call the AWS IAM list method via MCP
   **Then** the call is denied immediately
   **And** an audit log entry is created with "permission_denied" status
   **And** the agent receives a clear error message indicating insufficient permissions

#### Scenario 5: Evidence Collection via MCP
1. **Given** MCP tools for GitHub, Jira, and AWS are configured and enabled
   **When** I run a SOC2 compliance analysis that requires evidence from these sources
   **Then** agents collect evidence via MCP tool calls
   **And** evidence is normalized into the evidence graph
   **And** redaction policies are applied to sensitive fields
   **And** audit logs include tool name, method, argument hashes, redaction status, and duration

#### Scenario 6: CLI and TUI Operations
1. **Given** multiple MCP tools are configured with varying statuses (ready, degraded, disabled)
   **When** I run `sdek mcp list`
   **Then** I see all tools with their current status, latency metrics, and error details
   **When** I run `sdek mcp test github-tool`
   **Then** I receive round-trip diagnostics including handshake status and response time
   **When** I open the TUI and navigate to the MCP Tools panel
   **Then** I can toggle tools on/off and see real-time status updates

### Edge Cases

- **What happens when an MCP tool crashes mid-execution?**
  - The orchestrator detects the failure, marks the tool as "degraded", logs the error with full context, and returns a structured error to the calling agent. The agent can decide whether to fail the analysis or continue with partial evidence.

- **What happens when the same MCP tool is defined in multiple config locations?**
  - The orchestrator follows precedence: project configs (`./.sdek/mcp/`) override global configs (`~/.sdek/mcp/`), which override `SDEK_MCP_PATH`. A warning is logged indicating which config was chosen.

- **What happens when an agent exceeds its rate limit for an MCP tool?**
  - The orchestrator throttles the request, logs a rate-limit event, and returns a structured error with retry-after timing. The agent can implement exponential backoff or fail the operation based on its policy.

- **What happens when an MCP config file changes while the orchestrator is running?**
  - The file watcher detects the change, validates the new config, and hot-reloads the tool connection. If validation fails, the old configuration remains active and an error is logged.

- **What happens when network connectivity to an MCP server is lost?**
  - The orchestrator transitions the tool to "degraded" status, begins retry attempts with backoff, and continues serving other tools. Agents receive clear errors when attempting to use the unavailable tool.

---

## Requirements *(mandatory)*

### Functional Requirements

#### Configuration Management
- **FR-001**: System MUST accept MCP JSON configurations with the same structure and key semantics as VS Code and Cursor MCP specifications
- **FR-002**: System MUST support MCP configuration discovery in three locations with defined precedence: project directory (`./.sdek/mcp/*.json`), global directory (`~/.sdek/mcp/*.json`), and environment variable `SDEK_MCP_PATH` (colon-separated list)
- **FR-003**: System MUST validate MCP configurations against a versioned JSON schema and report errors with file path, line number, and property path
- **FR-004**: System MUST support hot-reloading of MCP configurations when files change without requiring orchestrator restart
- **FR-005**: System MUST provide a CLI command `sdek mcp validate [path]` that validates one or more configuration files and reports detailed errors
- **FR-006**: System MUST attempt best-effort compatibility with MCP protocol versions, log warnings when version mismatches are detected, and proceed with connection attempts unless critical incompatibilities prevent operation

#### Orchestrator Runtime
- **FR-007**: System MUST discover, load, and initialize all MCP tool connections during startup
- **FR-008**: System MUST perform health checks on MCP tools and mark them as "ready", "degraded", or "offline"
- **FR-009**: System MUST retry failed MCP connections with exponential backoff strategy
- **FR-010**: System MUST continue operating when individual MCP tools are unavailable (graceful degradation)
- **FR-011**: System MUST expose a registry interface that provides authenticated MCP clients to authorized agents
- **FR-012**: System MUST watch MCP configuration directories for changes and hot-reload updated configurations
- **FR-013**: System MUST attempt to recover degraded MCP tools to "ready" status within 30 seconds through retry attempts and health checks

#### Agent Capability Mapping
- **FR-014**: System MUST map MCP tools to specific agent capabilities using a capability string format (e.g., "git.read", "jira.search", "aws.iam.list")
- **FR-015**: System MUST enforce role-based access control (RBAC) at MCP tool invocation time, denying calls when agents lack required permissions
- **FR-016**: System MUST support per-tool execution budgets including rate limits, concurrency limits, and timeout configurations
- **FR-017**: System MUST return structured errors to agents when RBAC denies access, including the denied capability and required permission

#### Evidence Collection
- **FR-018**: Agents MUST be able to invoke MCP tools to collect evidence from external systems (GitHub, Jira, Slack, AWS, CI/CD, documentation systems)
- **FR-019**: System MUST normalize evidence collected via MCP tools into the existing evidence graph data structure
- **FR-020**: System MUST preserve and apply data redaction policies to MCP-collected evidence
- **FR-021**: System MUST preserve and apply caching policies to MCP tool responses to minimize redundant API calls
- **FR-022**: System MUST emit detailed audit logs for each MCP tool invocation including tool name, method, argument hash, redaction status, duration, and result status
- **FR-023**: System MUST support configurable audit log retention periods with a default of 7 days
- **FR-024**: System MUST handle expected audit log volumes of approximately 10,000 invocations per day under normal operation

#### CLI Operations
- **FR-025**: System MUST provide `sdek mcp list` command that displays all discovered MCP tools with status, latency metrics, and error details
- **FR-026**: System MUST provide `sdek mcp test <tool>` command that executes health checks and reports diagnostics including round-trip time
- **FR-027**: System MUST provide `sdek mcp enable <tool>` and `sdek mcp disable <tool>` commands to toggle tool availability
- **FR-028**: System MUST persist enable/disable state across orchestrator restarts

#### TUI Operations
- **FR-029**: TUI MUST display an MCP Tools panel showing all tools with real-time status updates
- **FR-030**: TUI MUST display per-tool metrics including latency, error rates, and last invocation time
- **FR-031**: Users MUST be able to enable/disable tools via TUI with immediate visual feedback
- **FR-032**: TUI MUST provide a quick-test action that runs health checks and displays results inline

#### Observability and Diagnostics
- **FR-033**: System MUST log all MCP configuration loading attempts with success/failure status
- **FR-034**: System MUST log all MCP tool state transitions (ready → degraded → offline → ready)
- **FR-035**: System MUST emit metrics for MCP tool invocation counts, latencies, error rates, and concurrent requests
- **FR-036**: System MUST provide diagnostic output that includes MCP protocol version, transport type, and server capabilities

#### Performance Requirements
- **FR-037**: MCP tool invocations (orchestrator overhead only, excluding external API time) MUST complete within 5 seconds under normal operating conditions
- **FR-038**: System MUST log a warning when MCP tool invocations exceed the 5-second performance target
- **FR-039**: System MUST support 20-50 concurrent MCP tool connections under normal operation
- **FR-040**: System MUST handle up to 100 MCP tool invocations per second without degradation

### Key Entities *(include if feature involves data)*

- **MCP Configuration**: Represents a tool connection definition with attributes including name, command to execute, command-line arguments, environment variables, transport protocol (stdio/http), and declared capabilities. Configurations are validated against a versioned JSON schema and can be loaded from multiple sources with defined precedence.

- **MCP Tool**: Represents an active connection to an MCP server with runtime state including health status (ready/degraded/offline), latency metrics, error history, and capability registration. Tools are managed by the orchestrator registry and can be enabled/disabled by administrators.

- **Agent Capability**: Represents a specific operation an agent is permitted to perform via MCP tools (e.g., "git.read", "aws.iam.list"). Capabilities are mapped to MCP tool methods and enforced via RBAC policies. Each capability may have associated execution budgets (rate limits, timeouts).

- **MCP Invocation Audit Log**: Represents a record of an agent's use of an MCP tool including timestamp, agent identifier, tool name, method called, argument hash (not raw arguments for security), redaction applied flag, duration, result status, and any errors. Audit logs are persisted for compliance reporting and security investigations.

- **Evidence Item**: Represents data collected from external systems via MCP tools, normalized into the evidence graph with metadata including source tool, collection timestamp, redaction status, and relationships to compliance controls. Evidence items maintain provenance linking back to specific MCP tool invocations.

- **RBAC Policy**: Defines which agents are authorized to invoke which MCP tool capabilities. Policies include agent roles, capability grants, and optional contextual constraints (e.g., time-based access, conditional permissions). Policies are evaluated at invocation time before allowing MCP tool access.

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

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
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (none found)
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

---

## Dependencies and Assumptions

### Dependencies
- Existing AI analysis engine and evidence graph structure (from features 002 and 003)
- Existing redaction and caching policies (from feature 003)
- MCP specification version 1.0 or compatible (as defined by VS Code/Cursor implementations)
- File system watch capability for hot-reloading configurations

### Assumptions
- Organizations have existing or can obtain MCP-compatible tool servers (GitHub, Jira, AWS, etc.)
- VS Code and Cursor MCP configuration formats are stable and documented
- Network connectivity to MCP servers is generally reliable with transient failures handled via retry
- Configuration files are stored in accessible file system locations (not remote/cloud storage)
- JSON schema validation is sufficient for configuration safety (no runtime sandboxing required for MCP servers)
- System will operate at medium scale: 20-50 concurrent MCP tools, up to 100 invocations/second
- Audit log volume approximates 10,000 invocations per day under normal operation
- MCP protocol version mismatches will be handled with best-effort compatibility and warnings
- Tool recovery from degraded state should occur within 30 seconds to meet operational requirements

---

## Success Metrics

- **Configuration Portability**: 100% of valid VS Code/Cursor MCP configs work in sdek-cli without modification
- **Adoption Rate**: 80% of new tool integrations use MCP rather than custom connectors within 6 months
- **Operational Efficiency**: Mean time to add a new evidence source drops from 2 weeks (custom connector) to 1 hour (MCP config)
- **Reliability**: 99.5% of MCP tool invocations complete successfully or fail gracefully with clear errors
- **Audit Compliance**: 100% of evidence collection activities are traced via MCP invocation audit logs
- **Security Enforcement**: Zero unauthorized MCP tool accesses bypass RBAC controls

---

## Out of Scope

- Building new MCP tool servers (we leverage existing third-party MCP servers)
- Removing or replacing non-MCP bespoke connectors (they coexist with MCP tools)
- Modifying the existing AI scoring model or analysis algorithms
- Creating a visual configuration editor for MCP configs (users edit JSON directly or use IDE tooling)
- Implementing MCP server discovery protocols beyond file system scanning
- Supporting MCP transport protocols beyond stdio and HTTP
- Providing MCP server hosting or proxy services
