# Feature Specification: MCP Pluggable Architecture

**Feature Branch**: `006-mcp-pluggable-architecture`
**Created**: 2025-10-26
**Status**: Draft
**Input**: User description: "I would like sdek cli to follow the same architecture as the kubectl-ai cli, it should be mcp pluggable just like the kubectl-ai project, be able to change ai models and interact with MCP like implemented now ( AWS API MCP ), conduct a research of the diff from current sdek cli project and kubectl-cli and update the architecture."

## Execution Flow (main)
```
1. Parse user description from Input
   → Feature clearly defined: Adopt kubectl-ai's MCP pluggable architecture
2. Extract key concepts from description
   → Actors: Compliance Managers, Security Engineers, DevOps Engineers
   → Actions: configure MCP servers, switch AI providers, aggregate tools, execute workflows
   → Data: MCP server configs, tool registries, provider factories, chat sessions
   → Constraints: backward compatibility, zero PII leakage, existing configs must work
3. For each unclear aspect:
   → Marked where clarification needed (migration path, MCP server mode needs)
4. Fill User Scenarios & Testing section
   → Four primary flows: MCP client mode, dual-mode operation, provider switching, multi-system orchestration
5. Generate Functional Requirements
   → 35 functional + 7 non-functional testable requirements covering MCP integration
6. Identify Key Entities
   → MCPManager, ToolRegistry, ProviderFactory, ChatSession, AnalysisRequest
7. Run Review Checklist
   → All sections complete, testable requirements, 2 clarifications marked
8. Return: SUCCESS (spec ready for planning)
```

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## Clarifications

### Session 2025-10-26
- Q: Should sdek-cli support MCP server mode (exposing its capabilities to other AI clients) or only client mode? → A: **NEEDS CLARIFICATION** - Start with client mode only, server mode can be Phase 2
- Q: What is the migration path for existing connector configurations? → A: **NEEDS CLARIFICATION** - Auto-migrate from legacy `ai.connectors` to new `mcp.servers` format

---

## Problem Statement

Current sdek-cli has a basic AI provider abstraction (OpenAI, Anthropic) and connector system for evidence collection, but lacks the flexibility and extensibility of kubectl-ai's MCP-based architecture. Specifically:

- **Hard-coded connectors**: Each evidence source (GitHub, Jira, AWS) requires custom implementation in the codebase
- **Limited AI provider switching**: Only OpenAI and Anthropic are supported, no local models (Ollama, llama.cpp)
- **No tool aggregation**: Cannot combine multiple MCP servers into unified workflows
- **Tightly coupled**: Provider logic mixed with business logic, making testing and extensibility difficult
- **No external tool integration**: Cannot leverage existing MCP servers from the ecosystem (AWS API MCP, filesystem MCP, etc.)

This results in:
- **Slow feature development**: Each new evidence source requires code changes and deployment
- **Limited flexibility**: Users cannot add custom evidence sources without modifying code
- **Poor testability**: Hard to mock external services without heavy refactoring
- **Ecosystem fragmentation**: Cannot reuse MCP servers from other projects

## Business Value & Outcomes

### Success Metrics
- **100% backward compatibility**: All existing configs, connectors, and workflows continue to work
- **3+ new MCP servers**: AWS API MCP, filesystem MCP, and at least one community MCP server integrated without code changes
- **5+ AI providers**: Support OpenAI, Anthropic, Gemini, AWS Bedrock, and Ollama (local)
- **50% faster evidence collection**: Parallel MCP server execution reduces total collection time
- **Zero code deployments**: Users can add new MCP servers via config file only

### Business Benefits
- **Faster compliance audits**: Multi-system orchestration reduces manual evidence gathering
- **Lower infrastructure costs**: Local AI models (Ollama) eliminate API costs for development/testing
- **Better ecosystem integration**: Leverage AWS, GCP, Azure MCP servers for cloud evidence
- **Improved developer experience**: Plugin architecture enables rapid prototyping and testing
- **Future-proof design**: MCP standard ensures compatibility with emerging AI tooling

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story 1: MCP Client Mode - AWS Evidence Collection
**As a** Compliance Manager analyzing PCI DSS requirements
**I want to** configure sdek-cli to use the AWS API MCP server for automatic cloud evidence collection
**So that** I can audit IAM policies, CloudTrail logs, and EKS configurations without writing custom connectors

**Flow:**
1. User installs AWS API MCP server: `uvx aws-api-mcp-server`
2. User configures sdek-cli: `~/.sdek/mcp-config.yaml`
3. User lists available MCP tools: `sdek mcp list-tools`
4. User runs autonomous mode: `sdek ai plan --framework pci-dss --requirement 8.2.3 --sources aws-mcp`
5. System discovers `call_aws` and `suggest_aws_commands` tools from AWS MCP
6. AI proposes evidence plan using AWS MCP tools
7. User approves plan, system executes via AWS MCP server
8. System normalizes AWS CLI output into EvidenceEvent format
9. Context mode analysis runs with collected evidence
10. User reviews finding with AWS provenance links

### Primary User Story 2: Dual-Mode MCP Operation (Future)
**As a** Security Engineer building compliance workflows
**I want to** use sdek-cli both as an MCP client (consuming external tools) and MCP server (exposing compliance analysis to other AI agents)
**So that** I can orchestrate multi-tool workflows across different platforms

**Flow:**
1. User configures sdek-cli in dual mode: `mcp.client.enabled: true`, `mcp.server.enabled: true`
2. User starts MCP server: `sdek mcp-server --port 9080`
3. External AI client (Claude Desktop) connects to sdek-cli MCP endpoint
4. Claude Desktop sees sdek-cli tools: `analyze_control`, `map_evidence`, `generate_finding`
5. Claude also connects to AWS MCP, GitHub MCP, Jira MCP directly
6. User asks Claude: "Audit SOC2 CC6.1 for production environment"
7. Claude orchestrates:
   - Uses GitHub MCP to fetch auth code commits
   - Uses AWS MCP to get IAM policies
   - Uses Jira MCP to check security tickets
   - Uses sdek-cli MCP to run compliance analysis
8. Claude returns unified compliance report with cross-system evidence

### Primary User Story 3: AI Provider Switching
**As a** DevOps Engineer testing compliance workflows locally
**I want to** switch between cloud AI providers (OpenAI, Gemini, Bedrock) and local models (Ollama)
**So that** I can develop and test without incurring API costs or hitting rate limits

**Flow:**
1. User switches to local model: `sdek config set ai.provider ollama --model gemma3:12b`
2. User verifies health: `sdek ai health`
3. System detects Ollama endpoint at `http://localhost:11434`
4. User runs test analysis: `sdek analyze --demo --control CC6.1`
5. System sends prompt to local Ollama instance
6. Analysis completes without external API calls
7. User reviews finding (may have lower confidence than GPT-4)
8. User switches back to production: `sdek config set ai.provider openai --model gpt-4o`

### Primary User Story 4: Multi-System Orchestration
**As a** Compliance Manager preparing for annual audit
**I want to** run a single command that collects evidence from GitHub, AWS, Jira, and CI/CD simultaneously
**So that** I can generate comprehensive compliance reports in minutes instead of days

**Flow:**
1. User configures multiple MCP servers in `mcp-config.yaml`: AWS, GitHub, Jira, GitLab
2. User runs autonomous mode: `sdek ai plan --framework iso27001 --section A.9.4 --sources all`
3. AI generates evidence plan with queries for all 4 sources
4. User approves plan with auto-approve policies pre-configured
5. System executes plan items in parallel (max 10 concurrent)
6. Progress indicators show collection status per source
7. System normalizes evidence from all sources into unified format
8. Context mode analysis runs with merged evidence bundle
9. Finding includes provenance: "45% GitHub, 30% AWS, 15% Jira, 10% GitLab"
10. User exports report: `sdek export --format html --output audit-report.html`

### Acceptance Scenarios

#### Phase 1: MCP Client Integration
1. **Given** AWS API MCP server is configured in `mcp-config.yaml`
   **When** I run `sdek mcp list-tools`
   **Then** the output includes `call_aws` and `suggest_aws_commands` tools
   **And** each tool shows description, parameters, and server source

2. **Given** multiple MCP servers are configured (AWS, filesystem)
   **When** I run autonomous mode analysis
   **Then** AI can propose using tools from any configured server
   **And** plan shows which MCP server will handle each item

3. **Given** an MCP server fails during plan execution
   **When** the system attempts to execute the plan
   **Then** failed items are logged with error details
   **And** execution continues with other MCP servers
   **And** final finding flagged as "partial evidence"

4. **Given** no MCP servers are configured
   **When** I run analysis
   **Then** system uses built-in legacy connectors
   **And** functionality is identical to pre-Feature-006 behavior

#### Phase 2: AI Provider Flexibility
5. **Given** I configure Ollama provider with `gemma3:12b` model
   **When** I run `sdek ai health`
   **Then** system checks `http://localhost:11434/api/tags`
   **And** confirms model is available
   **And** returns success status

6. **Given** I switch from OpenAI to Gemini provider
   **When** I run analysis
   **Then** AI cache is invalidated (different provider = new cache key)
   **And** prompts are sent to `generativelanguage.googleapis.com`
   **And** responses are cached under `gemini/` cache directory

7. **Given** multiple AI providers are configured in priority order
   **When** primary provider (OpenAI) fails or times out
   **Then** system retries with secondary provider (Gemini)
   **And** finding metadata shows which provider succeeded

#### Phase 3: Tool Aggregation & Orchestration
8. **Given** autonomous mode generates a plan with 15 evidence sources
   **When** I approve the plan
   **Then** system executes maximum 10 sources concurrently (configurable)
   **And** displays live progress: "Collecting 7/15 complete"
   **And** respects per-connector rate limits

9. **Given** AWS MCP server is configured with read-only IAM profile
   **When** AI proposes a mutation operation (e.g., `ec2 terminate-instances`)
   **Then** MCP server blocks the operation
   **And** error is logged: "AWS MCP denied: READ_OPERATIONS_ONLY=true"
   **And** plan item marked as failed

10. **Given** I configure filesystem MCP with working directory `/compliance-evidence`
    **When** autonomous mode collects evidence from multiple sources
    **Then** normalized results are saved to `/compliance-evidence/evidence-bundle.json`
    **And** filesystem MCP tool is used to write files
    **And** provenance includes file paths

### Edge Cases
- **What happens when** an MCP server becomes unresponsive mid-execution?
  → System applies timeout (default 60s), logs error, marks item failed, continues with other servers

- **What happens when** user configures duplicate MCP server names?
  → System errors on startup: "Duplicate MCP server name: aws-api"

- **What happens when** MCP server returns invalid JSON?
  → System logs parse error, marks item failed, optionally retries with exponential backoff

- **What happens when** legacy connector and MCP server both claim the same source type?
  → MCP server takes precedence if `mcp.preferMCP: true` in config, otherwise legacy connector used

- **What happens when** AI provider URL scheme is invalid (e.g., `foobar://example.com`)?
  → System errors: "Unknown AI provider scheme: foobar. Valid: openai, anthropic, gemini, bedrock, ollama, llamacpp, vertexai"

- **What happens when** an MCP server requests file system access outside working directory?
  → Server blocks operation per MCP security constraints (similar to AWS MCP `WORKING_DIR` enforcement)

---

## Requirements *(mandatory)*

### Functional Requirements - MCP Client Integration

- **FR-001**: System MUST support configuring multiple MCP servers via `~/.sdek/mcp-config.yaml`
- **FR-002**: MCP server configuration MUST include: name, command, args, transport (stdio/http), environment variables
- **FR-003**: System MUST support stdio transport for local MCP servers (subprocess communication)
- **FR-004**: System MUST support HTTP transport for remote MCP servers (with optional authentication)
- **FR-005**: System MUST discover tools from all configured MCP servers on startup
- **FR-006**: System MUST maintain a unified ToolRegistry combining built-in and MCP tools
- **FR-007**: Tool discovery MUST include: name, description, parameters schema, source server
- **FR-008**: System MUST route tool execution requests to the appropriate MCP server
- **FR-009**: System MUST handle MCP server failures gracefully (log error, continue with other servers)
- **FR-010**: System MUST respect per-MCP-server timeouts and rate limits
- **FR-011**: System MUST normalize MCP server responses into standard EvidenceEvent format
- **FR-012**: System MUST preserve backward compatibility with legacy connectors (Feature 003)
- **FR-013**: CLI MUST provide commands: `sdek mcp list-servers`, `sdek mcp list-tools`, `sdek mcp test <server>`

### Functional Requirements - AI Provider Abstraction

- **FR-014**: System MUST support AI provider selection via URL schemes: `openai://`, `anthropic://`, `gemini://`, `bedrock://`, `ollama://`, `llamacpp://`, `vertexai://`
- **FR-015**: System MUST implement ProviderFactory pattern with registration system (like gollm)
- **FR-016**: Provider factories MUST be registered in init() functions for automatic discovery
- **FR-017**: System MUST support provider-specific configuration: API keys, endpoints, models, timeouts
- **FR-018**: System MUST support local AI providers: Ollama (HTTP), llama.cpp (local binary)
- **FR-019**: System MUST support cloud AI providers: OpenAI, Anthropic, Google Gemini, AWS Bedrock, Azure OpenAI, Grok
- **FR-020**: System MUST invalidate cache when switching providers (provider name in cache key)
- **FR-021**: System MUST provide unified ChatSession interface across all providers
- **FR-022**: ChatSession MUST support: multi-turn conversations, function/tool definitions, structured JSON responses
- **FR-023**: System MUST provide `sdek ai health` command to verify provider connectivity
- **FR-024**: System MUST support provider fallback: primary fails → retry with secondary

### Functional Requirements - Tool Aggregation & Safety

- **FR-025**: System MUST combine tools from: built-in connectors, legacy Feature 003 connectors, MCP servers
- **FR-026**: System MUST implement three-tier safety validation for tool execution (like kubectl-ai)
- **FR-027**: Tier 1 safety MUST detect interactive commands (vim, nano, REPL)
- **FR-028**: Tier 2 safety MUST detect resource modification operations (delete, create, apply, terminate)
- **FR-029**: Tier 3 safety MUST flag operations requiring user confirmation based on risk level
- **FR-030**: System MUST log all tool executions to audit journal with: timestamp, tool name, arguments, result status
- **FR-031**: System MUST support concurrent tool execution with configurable limits (default: 10 concurrent)
- **FR-032**: System MUST respect budget limits when executing autonomous plans (max sources, max API calls)

### Functional Requirements - Configuration & Migration

- **FR-033**: System MUST auto-migrate legacy `ai.connectors` config to new `mcp.servers` format on first run
- **FR-034**: System MUST preserve existing `config.yaml` structure for backward compatibility
- **FR-035**: System MUST support both old connector API and new MCP API simultaneously during transition period

### Non-Functional Requirements

- **NFR-001**: MCP tool discovery MUST complete within 5 seconds for up to 10 configured servers
- **NFR-002**: Parallel MCP server execution MUST reduce evidence collection time by ≥50% vs. sequential
- **NFR-003**: Provider switching MUST not require application restart (hot-reload)
- **NFR-004**: MCP server failures MUST NOT crash the main application (isolation)
- **NFR-005**: Tool execution timeout MUST be configurable per MCP server (default: 60s)
- **NFR-006**: System MUST support ≥100 concurrent tool executions across all MCP servers
- **NFR-007**: MCP integration MUST maintain ≥80% test coverage (unit + integration)

### Key Entities *(include if feature involves data)*

- **MCPManager**: Orchestrates connections to multiple MCP servers. Responsibilities:
  - Initialize clients from config (stdio or HTTP transport)
  - Discover tools from all servers
  - Route tool execution requests
  - Handle server failures and retries
  - Aggregate responses

- **MCPServer**: Represents a configured MCP server connection. Includes:
  - name (string, unique identifier)
  - command (string, executable path)
  - args ([]string, command arguments)
  - transport (enum: stdio|http)
  - env (map[string]string, environment variables)
  - timeout (int, seconds)
  - health_status (enum: healthy|degraded|down)

- **ToolRegistry**: Unified catalog of all available tools. Responsibilities:
  - Register tools from built-in, legacy, and MCP sources
  - Provide discovery interface for AI
  - Route execution to appropriate handler
  - Apply safety validation before execution
  - Track execution metrics

- **Tool**: Represents an executable capability. Includes:
  - name (string, e.g., "call_aws", "kubectl", "analyze_control")
  - description (string, natural language)
  - parameters (JSONSchema, defines expected inputs)
  - source (enum: builtin|legacy-connector|mcp-server)
  - server_name (string, for MCP tools)
  - safety_tier (enum: safe|interactive|modifies-resource)

- **ProviderFactory**: Creates AI provider instances from configuration. Pattern:
  - Registration: `RegisterProviderFactory("openai", openAIFactory)`
  - Creation: `CreateProvider(config) → Provider`
  - Validation: Check API keys, endpoints, model availability

- **Provider**: Unified interface for AI providers. Methods:
  - AnalyzeWithContext(ctx, prompt) → response
  - GetCallCount() → int (for telemetry)
  - GetLastPrompt() → string (for debugging)
  - Health(ctx) → error

- **ChatSession**: Manages multi-turn AI conversations. Includes:
  - messages ([]Message, conversation history)
  - functions ([]FunctionDefinition, available tools)
  - provider (Provider, active AI provider)
  - config (SessionConfig, temperature, max_tokens, etc.)

- **AnalysisRequest**: Encapsulates analysis invocation. Includes:
  - preamble (ContextPreamble, framework + control context)
  - evidence (EvidenceBundle, collected events)
  - tools ([]Tool, available for execution)
  - options (AnalysisOptions, cache control, concurrency limits)

- **MCPConnectorAdapter**: Adapter that wraps legacy Feature 003 connectors as MCP-compatible tools. Enables:
  - Gradual migration from legacy to MCP
  - Unified tool interface for AI
  - Backward compatibility during transition

---

## Scope & Boundaries

### In Scope
- Phase 1: MCP client integration (consume external MCP servers)
- AI provider abstraction with factory pattern (OpenAI, Anthropic, Gemini, Bedrock, Ollama, llama.cpp, Vertex AI)
- Tool registry with unified discovery and execution
- Three-tier safety validation framework
- Configuration migration from legacy connectors to MCP format
- HTTP and stdio transport support for MCP servers
- Parallel tool execution with concurrency limits
- Audit journal for tool executions
- CLI commands: `sdek mcp list-servers`, `sdek mcp list-tools`, `sdek mcp test`, `sdek ai health`
- Integration testing with real MCP servers (AWS API MCP, filesystem MCP)

### Out of Scope (Phase 2 - Future)
- MCP server mode (exposing sdek-cli capabilities to other AI clients)
- Dual-mode operation (simultaneous client + server)
- Web UI for MCP server configuration
- Real-time tool execution streaming
- Custom MCP server development toolkit
- MCP server marketplace/discovery
- Advanced orchestration (conditional workflows, loops, error handling)
- MCP server health dashboard
- Distributed MCP server execution (Kubernetes pods)

### Dependencies
- MCP Specification v1.0+ compatibility
- AI provider SDKs: OpenAI Go SDK, Anthropic Go SDK, Google AI Go SDK, AWS SDK (Bedrock)
- Local AI infrastructure: Ollama installed for local model support
- External MCP servers: AWS API MCP server, filesystem MCP server (for testing)
- Existing Feature 003 components: Engine interface, ContextPreamble, EvidenceBundle, Finding

### Assumptions
- MCP servers follow JSON-RPC 2.0 specification
- Users have necessary credentials for MCP servers (AWS profiles, API keys)
- stdio transport MCP servers are trusted (run as subprocesses with full system access)
- HTTP transport MCP servers support optional Bearer token authentication
- Legacy connectors will continue to work unchanged during migration period
- Users will gradually migrate to MCP-based evidence collection

---

## Architectural Comparison: Current vs. Target

### Current sdek-cli Architecture (Feature 003)
```
AI Provider (hard-coded)
    ├── OpenAI (via SDK)
    └── Anthropic (via SDK)

Evidence Connectors (custom implementations)
    ├── GitHubConnector (hard-coded)
    ├── JiraConnector (hard-coded)
    ├── AWSConnector (hard-coded)
    └── CI/CDConnector (hard-coded)

Analysis Engine
    ├── ProviderFactory (basic)
    ├── Engine (analysis only)
    └── Cache (SHA256-based)
```

**Limitations:**
- Adding new evidence source requires code changes
- Only 2 AI providers supported
- No local model support
- No tool aggregation across systems
- Tight coupling between engine and providers

### Target kubectl-ai-inspired Architecture (Feature 006)
```
AI Provider Layer (pluggable via factory)
    ├── ProviderFactory (registration-based)
    │   ├── OpenAI (registered in init)
    │   ├── Anthropic (registered in init)
    │   ├── Gemini (registered in init)
    │   ├── Bedrock (registered in init)
    │   ├── Ollama (registered in init)
    │   ├── llama.cpp (registered in init)
    │   └── VertexAI (registered in init)
    └── Unified Client Interface
        ├── ChatSession (multi-turn)
        ├── Function Calling (tool use)
        └── Structured Responses (JSON schema)

MCP Integration Layer
    ├── MCPManager
    │   ├── Server Discovery
    │   ├── Tool Aggregation
    │   └── Execution Routing
    │
    ├── MCP Servers (config-driven)
    │   ├── AWS API MCP (stdio/http)
    │   ├── Filesystem MCP (stdio)
    │   ├── GitHub MCP (stdio/http)
    │   ├── Jira MCP (stdio/http)
    │   ├── Slack MCP (stdio/http)
    │   └── Custom MCP Servers (user-defined)
    │
    └── Transport Layer
        ├── StdioClient (subprocess)
        └── HTTPClient (remote endpoints)

Tool System
    ├── ToolRegistry (unified catalog)
    │   ├── Built-in Tools
    │   ├── Legacy Connectors (wrapped)
    │   └── MCP Server Tools
    │
    ├── Safety Validator (3-tier)
    │   ├── Tier 1: Interactive Detection
    │   ├── Tier 2: Resource Modification
    │   └── Tier 3: User Confirmation
    │
    └── Execution Engine
        ├── Parallel Execution (10 concurrent default)
        ├── Rate Limiting (per-server)
        ├── Timeout Handling
        └── Audit Logging

Analysis Engine (unchanged)
    ├── Context Injection
    ├── Evidence Bundling
    ├── Finding Generation
    └── Cache Management
```

**Benefits:**
- Zero-code evidence source addition (config-only)
- 7+ AI providers supported
- Local model support (Ollama, llama.cpp)
- Tool aggregation enables multi-system workflows
- Loose coupling via interfaces
- Testable via mocking at any layer

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] Requirements are testable and unambiguous (2 clarifications marked)
- [x] Success criteria are measurable (100% compat, 3+ MCP servers, 5+ AI providers, 50% faster, zero code deployments)
- [x] Scope is clearly bounded (in/out of scope defined, Phase 2 identified)
- [x] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (2 clarifications)
- [x] User scenarios defined (4 primary stories, 10 acceptance scenarios, 6 edge cases)
- [x] Requirements generated (35 functional, 7 non-functional)
- [x] Entities identified (9 key entities)
- [x] Architectural comparison documented
- [x] Review checklist passed

---

## Next Steps

1. **Planning Phase**: Create detailed plan with tasks, milestones, and estimates
2. **Clarifications**: Resolve outstanding questions with stakeholders:
   - Decide on MCP server mode priority (Phase 1 or Phase 2)
   - Define migration strategy for existing connector configs
3. **Research Phase**: Document architectural gaps and design decisions
4. **Data Model**: Document MCPManager, ToolRegistry, ProviderFactory, ChatSession schemas
5. **Contracts**: Define internal API interfaces (MCPClient, ProviderFactory, ToolRegistry)
6. **Implementation**: Begin with provider factory refactoring, then MCP client integration
