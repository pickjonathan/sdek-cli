# Research: MCP Pluggable Architecture

**Feature**: 006-mcp-pluggable-architecture
**Date**: 2025-10-26
**Status**: Complete

This document consolidates research findings, architectural gap analysis, and design decisions for transforming sdek-cli into an MCP-pluggable system following kubectl-ai's architecture patterns.

---

## Table of Contents
1. [Architectural Gap Analysis](#architectural-gap-analysis)
2. [Technology Choices](#technology-choices)
3. [Design Decisions](#design-decisions)
4. [Best Practices](#best-practices)
5. [Migration Strategy](#migration-strategy)
6. [Risk Assessment](#risk-assessment)

---

## Architectural Gap Analysis

### Current State (sdek-cli Feature 003)

#### AI Provider Layer
**Current Implementation:**
```go
// internal/ai/provider_factory.go
var providerFactories = make(map[string]ProviderFactory)

func RegisterProviderFactory(providerName string, factory ProviderFactory) {
    providerFactories[providerName] = factory
}
```

**Supported Providers:**
- OpenAI (via custom implementation)
- Anthropic (via custom implementation)
- Mock (for testing)

**Gaps:**
1. ❌ No URL scheme-based provider selection (uses string matching: "openai", "anthropic")
2. ❌ No unified ChatSession interface (direct API calls)
3. ❌ No local model support (Ollama, llama.cpp)
4. ❌ No cloud provider support beyond OpenAI/Anthropic (Gemini, Bedrock, Vertex AI)
5. ❌ No provider fallback mechanism
6. ❌ Provider configuration scattered across multiple files

#### Evidence Collection Layer
**Current Implementation:**
```go
// internal/ai/connectors/connector.go
type Connector interface {
    Collect(ctx context.Context, source string, query string) ([]types.EvidenceEvent, error)
}

// internal/ai/connectors/registry.go
var globalRegistry = &Registry{
    connectors: make(map[string]Connector),
}
```

**Supported Connectors:**
- GitHub (hard-coded)
- Jira (hard-coded)
- AWS (hard-coded)
- CI/CD (hard-coded)

**Gaps:**
1. ❌ No MCP server support (all connectors are Go code)
2. ❌ No external tool discovery (fixed set of connectors)
3. ❌ No stdio or HTTP transport abstraction
4. ❌ No config-driven connector registration
5. ❌ No tool aggregation across multiple systems
6. ❌ No safety validation framework

#### Configuration System
**Current Implementation:**
```go
// pkg/types/config.go
type AIConfig struct {
    Enabled      bool
    Provider     string  // "openai" | "anthropic"
    Model        string
    OpenAIKey    string
    AnthropicKey string
    Connectors   map[string]ConnectorConfig
}
```

**Gaps:**
1. ❌ No MCP server configuration section
2. ❌ No transport-specific settings (stdio vs HTTP)
3. ❌ No tool safety configuration
4. ❌ No multi-provider fallback configuration

### Target State (kubectl-ai Architecture)

#### AI Provider Layer (gollm)
**Target Implementation:**
```go
// URL scheme-based provider selection
client := gollm.NewClient("openai://api.openai.com")
client := gollm.NewClient("gemini://generativelanguage.googleapis.com")
client := gollm.NewClient("ollama://localhost:11434")

// Unified ChatSession interface
chat := client.NewChat()
chat.AddMessage("user", "Analyze this control")
chat.SetFunctions(toolDefinitions)
response, _ := chat.Send(ctx)
```

**Benefits:**
- URL schemes make provider selection declarative
- ChatSession abstraction enables multi-turn conversations
- Function calling support built-in
- Provider-specific logic isolated in factory implementations

#### MCP Integration Layer
**Target Implementation:**
```yaml
# ~/.sdek/mcp-config.yaml
servers:
  aws-api:
    command: "uvx"
    args: ["aws-api-mcp-server"]
    transport: "stdio"
    env:
      AWS_PROFILE: "readonly"
      READ_OPERATIONS_ONLY: "true"

  filesystem:
    command: "npx"
    args: ["@modelcontextprotocol/server-filesystem", "/evidence"]
    transport: "stdio"

  remote-api:
    url: "http://mcp-server:8080/mcp"
    transport: "http"
    headers:
      Authorization: "Bearer ${MCP_TOKEN}"
```

**Benefits:**
- Zero-code evidence source addition
- External MCP server ecosystem integration
- Transport abstraction (stdio for local, HTTP for remote)
- Environment variable injection for credentials

#### Tool System
**Target Implementation:**
```go
type ToolRegistry struct {
    builtinTools  map[string]Tool
    mcpTools      map[string]Tool  // From MCP servers
    legacyTools   map[string]Tool  // Wrapped connectors
}

type Tool struct {
    Name         string
    Description  string
    Parameters   JSONSchema
    Source       ToolSource  // builtin|mcp|legacy
    ServerName   string      // For MCP tools
    SafetyTier   SafetyTier  // safe|interactive|modifies
    Handler      ToolHandler
}

// Three-tier safety validation
func (r *ToolRegistry) Analyze(call ToolCall) ToolCallAnalysis {
    // Tier 1: Interactive command detection
    if isInteractive(call.Command) {
        return ToolCallAnalysis{IsInteractive: true}
    }

    // Tier 2: Resource modification detection
    if modifiesResources(call.Command) {
        return ToolCallAnalysis{ModifiesResource: true}
    }

    // Tier 3: User confirmation required
    return ToolCallAnalysis{RequiresApproval: needsApproval(call)}
}
```

**Benefits:**
- Unified tool catalog across all sources
- Safety validation before execution
- Audit trail for all tool calls
- Concurrent execution with limits

---

## Technology Choices

### 1. MCP Client Library

**Decision:** Implement custom MCP client following specification
**Rationale:**
- No mature Go MCP client library exists yet
- kubectl-ai implements custom client, proven in production
- Full control over transport layer (stdio + HTTP)
- Can optimize for sdek-cli's specific needs

**Alternatives Considered:**
- Wait for official Go MCP SDK → ❌ No ETA, blocks feature delivery
- Use kubectl-ai's gollm library directly → ❌ License compatibility, not published as module
- Python MCP SDK with CGo → ❌ Deployment complexity, performance overhead

**Implementation Approach:**
```go
// internal/mcp/client.go
type Client interface {
    Initialize(ctx context.Context, config ServerConfig) error
    ListTools(ctx context.Context) ([]Tool, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
    Close() error
}

// internal/mcp/stdio_client.go - subprocess communication
type StdioClient struct {
    cmd    *exec.Cmd
    stdin  io.Writer
    stdout io.Reader
    stderr io.Reader
}

// internal/mcp/http_client.go - remote HTTP endpoints
type HTTPClient struct {
    baseURL string
    headers map[string]string
    client  *http.Client
}
```

### 2. AI Provider Abstraction

**Decision:** Adapt gollm's factory pattern without direct dependency
**Rationale:**
- Factory pattern with registration proven effective
- URL scheme-based selection intuitive for users
- Avoids dependency on external module (license, versioning)
- Can extend with sdek-cli-specific features

**Alternatives Considered:**
- Use gollm as Go module → ❌ Not published, would require vendoring
- Abstract via LangChain Go → ❌ Heavy dependency, overkill for needs
- Keep current simple abstraction → ❌ Doesn't support local models or advanced features

**Implementation Approach:**
```go
// internal/ai/factory/registry.go
type ProviderFactory func(config ProviderConfig) (Provider, error)

var registry = make(map[string]ProviderFactory)

func Register(scheme string, factory ProviderFactory) {
    registry[scheme] = factory
}

func NewProvider(url string, config ProviderConfig) (Provider, error) {
    u, _ := url.Parse(url)
    factory, exists := registry[u.Scheme]
    if !exists {
        return nil, fmt.Errorf("unknown provider: %s", u.Scheme)
    }
    return factory(config)
}

// Provider implementations register in init()
func init() {
    factory.Register("openai", openAIFactory)
    factory.Register("anthropic", anthropicFactory)
    factory.Register("gemini", geminiFactory)
    factory.Register("bedrock", bedrockFactory)
    factory.Register("ollama", ollamaFactory)
    // ...
}
```

### 3. Configuration Management

**Decision:** Extend Viper-based config with MCP section, maintain backward compatibility
**Rationale:**
- Already using Viper for configuration
- YAML format familiar to users
- Environment variable substitution works
- Can auto-migrate legacy connector configs

**Alternatives Considered:**
- Separate MCP config file → ❌ Confusing, users have to manage multiple files
- JSON format → ❌ Less human-readable than YAML
- TOML format → ❌ Would require migration, no benefit

**Implementation Approach:**
```yaml
# ~/.sdek/config.yaml (extended)
ai:
  enabled: true
  provider_url: "openai://api.openai.com"  # NEW: URL scheme
  model: "gpt-4o"
  timeout: 60

  # Legacy fields (backward compat)
  provider: "openai"
  openai_key: "${SDEK_AI_OPENAI_KEY}"

  # NEW: Provider-specific configs
  providers:
    openai:
      api_key: "${SDEK_AI_OPENAI_KEY}"
      endpoint: "https://api.openai.com/v1"

    ollama:
      endpoint: "http://localhost:11434"
      model: "gemma3:12b"

    bedrock:
      region: "us-east-1"
      profile: "default"

  # NEW: Fallback chain
  fallback:
    enabled: true
    providers: ["openai", "gemini", "ollama"]

# NEW: MCP configuration section
mcp:
  enabled: true
  prefer_mcp: true  # MCP tools take precedence over legacy connectors

  servers:
    aws-api:
      command: "uvx"
      args: ["aws-api-mcp-server"]
      transport: "stdio"
      timeout: 60
      env:
        AWS_PROFILE: "readonly"
        READ_OPERATIONS_ONLY: "true"

    filesystem:
      command: "npx"
      args: ["@modelcontextprotocol/server-filesystem", "/evidence"]
      transport: "stdio"

    custom-api:
      url: "http://localhost:8080/mcp"
      transport: "http"
      headers:
        Authorization: "Bearer ${MCP_TOKEN}"
```

### 4. Transport Layer

**Decision:** Implement both stdio and HTTP transports
**Rationale:**
- stdio required for local MCP servers (AWS MCP, filesystem MCP)
- HTTP required for remote/containerized MCP servers
- MCP spec defines both transports
- kubectl-ai supports both

**Alternatives Considered:**
- stdio only → ❌ Can't integrate with remote services
- HTTP only → ❌ Can't use official MCP servers (most are stdio)
- gRPC → ❌ Not part of MCP spec, no ecosystem support

**Implementation Approach:**
```go
// internal/mcp/transport.go
type Transport interface {
    Send(ctx context.Context, request *jsonrpc.Request) (*jsonrpc.Response, error)
    Close() error
}

// stdio: subprocess communication via JSON-RPC over stdin/stdout
type StdioTransport struct {
    cmd    *exec.Cmd
    stdin  *json.Encoder
    stdout *json.Decoder
}

// HTTP: JSON-RPC over HTTP POST + SSE for server-to-client
type HTTPTransport struct {
    url     string
    headers map[string]string
    client  *http.Client
}
```

### 5. Backward Compatibility Strategy

**Decision:** Maintain legacy connector API, wrap as MCP-compatible tools
**Rationale:**
- Existing Feature 003 code must continue working
- Users don't want breaking changes
- Gradual migration path more realistic
- Can deprecate legacy API after Feature 006 stable

**Alternatives Considered:**
- Break API, force migration → ❌ User disruption, blocks adoption
- Duplicate code → ❌ Maintenance burden, bugs
- No backward compat → ❌ Not acceptable per spec requirements

**Implementation Approach:**
```go
// internal/mcp/adapter.go - Wraps legacy connectors as MCP tools
type LegacyConnectorAdapter struct {
    connector connectors.Connector
}

func (a *LegacyConnectorAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
    source := args["source"].(string)
    query := args["query"].(string)
    events, err := a.connector.Collect(ctx, source, query)
    // ... normalize to MCP response format
    return events, err
}

// Auto-wrap legacy connectors on startup
func WrapLegacyConnectors(registry *connectors.Registry) []Tool {
    var tools []Tool
    for name, connector := range registry.List() {
        tools = append(tools, Tool{
            Name:   fmt.Sprintf("legacy_%s", name),
            Source: ToolSourceLegacy,
            Handler: &LegacyConnectorAdapter{connector: connector},
        })
    }
    return tools
}
```

---

## Design Decisions

### Decision 1: Phased MCP Server Mode

**Question:** Should Feature 006 include MCP server mode (exposing sdek-cli to external AI clients)?

**Decision:** **No - Phase 1 is client mode only, server mode deferred to Phase 2**

**Rationale:**
1. **Complexity**: Server mode requires HTTP server, session management, authentication
2. **Use case priority**: 90% of value is in consuming external MCP servers (AWS, GitHub, Jira)
3. **Testing**: Client mode easier to test and validate
4. **Resource constraints**: Server mode would double implementation timeline

**Phase 1 Scope:**
- MCP client integration
- Tool aggregation from external servers
- AI provider abstraction
- Backward compatibility

**Phase 2 Scope (Future):**
- Expose sdek-cli as MCP server
- HTTP endpoint for external AI clients
- Authentication/authorization for MCP API
- Dual-mode operation (client + server simultaneously)

### Decision 2: Tool Safety Framework

**Question:** How aggressive should safety validation be for tool execution?

**Decision:** **Three-tier validation with user confirmation for Tier 2/3**

**Rationale:**
1. **Compliance context**: Accidental destructive operations could corrupt evidence
2. **kubectl-ai proven**: Three-tier approach battle-tested
3. **User trust**: Explicit confirmation builds confidence in autonomous mode
4. **Auditability**: All approvals logged for compliance audits

**Tier Definitions:**

**Tier 1 - Interactive Commands (Block by default):**
- Examples: `vim`, `nano`, `emacs`, `python` (REPL), `bash`
- Reason: Could hang process, require terminal interaction
- Action: Request user approval, show command preview

**Tier 2 - Resource Modification (Warn):**
- Examples: `aws ec2 terminate-instances`, `kubectl delete`, `git push --force`
- Reason: Could modify production systems
- Action: Require explicit user confirmation, log to audit trail

**Tier 3 - Safe Operations (Allow):**
- Examples: `aws iam list-users`, `kubectl get pods`, `git log`
- Reason: Read-only, no side effects
- Action: Execute immediately, log to audit trail

**Implementation:**
```go
type SafetyTier int

const (
    SafetyTierSafe SafetyTier = iota
    SafetyTierInteractive
    SafetyTierModifiesResource
)

var dangerousVerbs = []string{
    "delete", "terminate", "destroy", "drop", "remove",
    "create", "apply", "patch", "update", "scale",
    "exec", "ssh", "run",
}

var interactiveCommands = []string{
    "vim", "vi", "nano", "emacs", "pico",
    "python", "irb", "node", "bash", "sh",
}

func AnalyzeSafety(command string) SafetyTier {
    cmd := strings.Fields(command)[0]

    if slices.Contains(interactiveCommands, cmd) {
        return SafetyTierInteractive
    }

    for _, verb := range dangerousVerbs {
        if strings.Contains(command, verb) {
            return SafetyTierModifiesResource
        }
    }

    return SafetyTierSafe
}
```

### Decision 3: Configuration Migration Path

**Question:** How to migrate existing `ai.connectors` configs to new `mcp.servers` format?

**Decision:** **Auto-migrate on first run, preserve legacy configs, add deprecation warnings**

**Rationale:**
1. **User experience**: Zero manual work for migration
2. **Safety**: Don't delete old config (users can rollback)
3. **Transparency**: Log what was migrated
4. **Gradual**: Users can manually tune after auto-migration

**Migration Algorithm:**
```go
// internal/config/migration.go
func MigrateLegacyConnectors(cfg *types.Config) error {
    if cfg.MCP == nil {
        cfg.MCP = &types.MCPConfig{
            Enabled: true,
            Servers: make(map[string]types.MCPServerConfig),
        }
    }

    // Migrate each legacy connector
    for name, connCfg := range cfg.AI.Connectors {
        if !connCfg.Enabled {
            continue
        }

        // Map legacy connector to MCP server equivalent
        switch name {
        case "github":
            cfg.MCP.Servers["github-mcp"] = types.MCPServerConfig{
                Command:   "npx",
                Args:      []string{"@github/github-mcp-server"},
                Transport: "stdio",
                Env: map[string]string{
                    "GITHUB_TOKEN": connCfg.APIKey,
                },
            }
        case "aws":
            cfg.MCP.Servers["aws-api"] = types.MCPServerConfig{
                Command:   "uvx",
                Args:      []string{"aws-api-mcp-server"},
                Transport: "stdio",
                Env: map[string]string{
                    "AWS_PROFILE":          "default",
                    "READ_OPERATIONS_ONLY": "true",
                },
            }
        // ... more mappings
        }
    }

    // Log migration
    log.Info("Migrated legacy connectors to MCP format",
        "count", len(cfg.AI.Connectors),
        "servers", len(cfg.MCP.Servers))

    return nil
}
```

**User Communication:**
```
$ sdek analyze

INFO: Legacy connector config detected, auto-migrating to MCP format
INFO: Migrated 3 connectors: github → github-mcp, aws → aws-api, jira → jira-mcp
INFO: Review new config: ~/.sdek/config.yaml (section: mcp.servers)
INFO: Legacy connectors preserved for backward compatibility
WARN: Legacy connector API deprecated, will be removed in v2.0.0
```

### Decision 4: Concurrency Model

**Question:** How to handle concurrent tool execution during autonomous mode?

**Decision:** **Parallel execution with configurable limit (default: 10 concurrent)**

**Rationale:**
1. **Performance**: Evidence collection from 10 sources in parallel ~50% faster than sequential
2. **Resource control**: Unlimited concurrency could overwhelm MCP servers or hit rate limits
3. **Observability**: Progress tracking easier with bounded concurrency
4. **Failure isolation**: One slow/failed server doesn't block others

**Implementation:**
```go
// internal/mcp/executor.go
type Executor struct {
    maxConcurrency int
    semaphore      chan struct{}
    wg             sync.WaitGroup
}

func (e *Executor) ExecutePlan(ctx context.Context, plan *types.EvidencePlan) (*types.EvidenceBundle, error) {
    e.semaphore = make(chan struct{}, e.maxConcurrency)

    results := make(chan *PlanItemResult, len(plan.Items))
    errors := make(chan error, len(plan.Items))

    for _, item := range plan.Items {
        if !item.Approved {
            continue  // Skip non-approved items
        }

        e.wg.Add(1)
        go func(item types.PlanItem) {
            defer e.wg.Done()

            // Acquire semaphore slot
            e.semaphore <- struct{}{}
            defer func() { <-e.semaphore }()

            // Execute item via MCP server
            result, err := e.executeItem(ctx, item)
            if err != nil {
                errors <- err
                return
            }
            results <- result
        }(item)
    }

    // Wait for all items
    e.wg.Wait()
    close(results)
    close(errors)

    // Aggregate results
    bundle := aggregateResults(results)
    return bundle, firstError(errors)
}
```

### Decision 5: Error Handling Strategy

**Question:** How should the system handle MCP server failures?

**Decision:** **Graceful degradation - log errors, continue with available servers, flag partial results**

**Rationale:**
1. **Resilience**: One bad MCP server shouldn't crash entire analysis
2. **User value**: Partial evidence better than no evidence
3. **Transparency**: Users see which servers failed and why
4. **Debugging**: Detailed error logs help diagnose MCP server issues

**Error Categories:**

**Transient Errors (Retry):**
- Network timeouts
- Rate limiting (429)
- Temporary service unavailable (503)

**Permanent Errors (No Retry):**
- Authentication failure (401)
- Server not found (404)
- Invalid request (400)
- Unimplemented method

**Handling:**
```go
func (m *MCPManager) CallTool(ctx context.Context, serverName string, toolName string, args map[string]interface{}) (interface{}, error) {
    server := m.servers[serverName]

    var lastErr error
    for attempt := 0; attempt < maxRetries; attempt++ {
        result, err := server.CallTool(ctx, toolName, args)
        if err == nil {
            return result, nil
        }

        // Check if retryable
        if !isRetryable(err) {
            return nil, fmt.Errorf("permanent error from MCP server %s: %w", serverName, err)
        }

        lastErr = err
        backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        time.Sleep(backoff)
    }

    // Mark server as degraded
    m.markDegraded(serverName, lastErr)
    return nil, fmt.Errorf("MCP server %s failed after %d retries: %w", serverName, maxRetries, lastErr)
}
```

**User Feedback:**
```
$ sdek ai plan --framework soc2 --control CC6.1 --sources all

Collecting evidence from 4 MCP servers...
✓ github-mcp: 12 events collected
✗ aws-api: connection timeout (will retry)
✓ jira-mcp: 8 events collected
✓ filesystem: 5 events collected

WARNING: aws-api MCP server failed, evidence may be incomplete
Evidence collected: 25/~35 events (71% complete)

Proceeding with available evidence for analysis...
```

---

## Best Practices

### 1. MCP Server Configuration Best Practices

**Security:**
```yaml
mcp:
  servers:
    aws-api:
      # ✅ Use read-only IAM profiles for evidence collection
      env:
        AWS_PROFILE: "compliance-readonly"
        READ_OPERATIONS_ONLY: "true"
        AWS_API_MCP_WORKING_DIR: "/tmp/sdek-evidence"

      # ✅ Set aggressive timeouts to prevent hanging
      timeout: 30

      # ✅ Use environment variables for secrets (not inline)
      env:
        AWS_SECRET_ACCESS_KEY: "${AWS_SECRET}"

    # ❌ DON'T: Expose admin credentials
    aws-admin:
      env:
        AWS_PROFILE: "admin"  # ❌ Too permissive
```

**Performance:**
```yaml
mcp:
  # ✅ Set reasonable concurrency limits
  max_concurrent: 10

  servers:
    github-mcp:
      # ✅ Configure rate limits per server
      rate_limit: 30  # requests per minute
      timeout: 60     # seconds
```

**Reliability:**
```yaml
mcp:
  # ✅ Enable health checks
  health_check_interval: 300  # seconds

  # ✅ Configure retry behavior
  retry:
    max_attempts: 3
    backoff: "exponential"  # exponential|linear|constant

  servers:
    remote-api:
      # ✅ Use health endpoints
      health_url: "http://api.example.com/health"
```

### 2. AI Provider Selection Guidelines

**Development/Testing:**
```yaml
ai:
  provider_url: "ollama://localhost:11434"
  model: "gemma3:12b"
  # Local model, no API costs, fast iteration
```

**Staging/QA:**
```yaml
ai:
  provider_url: "gemini://generativelanguage.googleapis.com"
  model: "gemini-2.5-flash"
  # Fast, low-cost, good quality for validation
```

**Production:**
```yaml
ai:
  provider_url: "openai://api.openai.com"
  model: "gpt-4o"
  fallback:
    enabled: true
    providers: ["gemini", "anthropic"]
  # Best quality, fallback for reliability
```

**Compliance/Sensitive Data:**
```yaml
ai:
  provider_url: "bedrock://us-east-1"
  model: "us.anthropic.claude-sonnet-4"
  # AWS Bedrock keeps data in your VPC, no third-party API calls
```

### 3. Tool Safety Configuration

**Conservative (Compliance Audits):**
```yaml
ai:
  safety:
    require_approval_for:
      - "modifies_resource"
      - "interactive"

    deny_list:
      - "delete"
      - "terminate"
      - "destroy"
      - "exec"

    audit_all_executions: true
    audit_log: "/var/log/sdek/tool-audit.log"
```

**Moderate (Development):**
```yaml
ai:
  safety:
    require_approval_for:
      - "modifies_resource"

    # Allow read operations without approval
    allow_list:
      - "get"
      - "list"
      - "describe"
      - "show"
```

**Permissive (Automated Testing):**
```yaml
ai:
  safety:
    # Auto-approve all in CI/CD
    auto_approve_all: true  # ⚠️ Only for isolated test environments

    # Still audit everything
    audit_all_executions: true
```

### 4. Evidence Collection Optimization

**Parallel Collection:**
```yaml
ai:
  autonomous:
    # Optimize for speed
    max_concurrent_sources: 15
    timeout_per_source: 30

    # Fail fast if most sources succeed
    min_success_rate: 0.70  # 70% of sources must succeed
```

**Budget Control:**
```yaml
ai:
  budgets:
    # Prevent runaway costs
    max_api_calls: 500
    max_tokens: 250000
    max_sources: 50

    # Alert before hitting limits
    alert_threshold: 0.80  # 80% of budget
```

**Auto-Approval Policies:**
```yaml
ai:
  autonomous:
    auto_approve:
      # Auto-approve specific source+query patterns
      github: ["auth*", "login*", "security*"]
      aws: ["iam:List*", "iam:Get*", "cloudtrail:Describe*"]
      jira: ["project=SECURITY"]

      # Require manual approval for everything else
```

---

## Migration Strategy

### Phase 1: Preparation (Week 1)
1. **Feature flag**: Add `mcp.enabled` config flag (default: false)
2. **Backward compatibility**: Ensure all existing tests pass with MCP disabled
3. **Documentation**: Update CLAUDE.md with new architecture overview
4. **Dependencies**: Add required Go modules (no external MCP client dependencies yet)

### Phase 2: AI Provider Refactoring (Week 2-3)
1. **Factory pattern**: Implement URL scheme-based provider selection
2. **Provider registration**: Migrate OpenAI and Anthropic to new factory pattern
3. **New providers**: Add Gemini, Bedrock, Ollama, llama.cpp factories
4. **ChatSession**: Implement unified chat interface
5. **Testing**: Provider factory unit tests + integration tests with real APIs
6. **Migration**: Auto-detect legacy `ai.provider` string and convert to URL

### Phase 3: MCP Client Implementation (Week 4-5)
1. **Transport layer**: Implement stdio and HTTP transports
2. **JSON-RPC**: Implement MCP protocol messages (initialize, list_tools, call_tool)
3. **MCPManager**: Orchestrate multiple MCP server connections
4. **Tool discovery**: Aggregate tools from all servers into unified registry
5. **Testing**: Mock MCP servers for unit tests, real AWS MCP for integration

### Phase 4: Tool Registry & Safety (Week 6)
1. **ToolRegistry**: Unified catalog combining builtin, legacy, MCP tools
2. **Safety validator**: Implement three-tier validation
3. **Audit journal**: Log all tool executions
4. **Concurrent execution**: Parallel tool calls with semaphore
5. **Testing**: Safety validation test suite

### Phase 5: Integration & Migration (Week 7)
1. **Legacy adapter**: Wrap existing connectors as MCP-compatible tools
2. **Config migration**: Auto-migrate `ai.connectors` to `mcp.servers`
3. **CLI commands**: Add `sdek mcp list-servers`, `sdek mcp list-tools`, `sdek mcp test`
4. **End-to-end testing**: Full workflow with real MCP servers
5. **Performance testing**: Validate 50% speedup from parallel execution

### Phase 6: Documentation & Release (Week 8)
1. **User guide**: MCP configuration examples
2. **Migration guide**: Step-by-step for existing users
3. **Provider guide**: How to use different AI providers
4. **MCP server guide**: How to integrate new MCP servers
5. **Release notes**: Breaking changes, deprecations, new features

---

## Risk Assessment

### High Risk

**Risk 1: MCP Server Ecosystem Immaturity**
- **Issue**: Few production-ready MCP servers exist beyond AWS
- **Impact**: Limited immediate value for users
- **Mitigation**:
  - Maintain backward compatibility with legacy connectors
  - Provide adapter to wrap legacy connectors as MCP tools
  - Document how to build custom MCP servers
  - Partner with community to develop key MCP servers (Jira, GitHub, Slack)

**Risk 2: Breaking Changes in MCP Specification**
- **Issue**: MCP is relatively new (2024), spec may evolve
- **Impact**: Our implementation may become incompatible
- **Mitigation**:
  - Version MCP client code separately
  - Add version detection handshake
  - Support multiple MCP spec versions
  - Monitor MCP spec repo for changes

### Medium Risk

**Risk 3: Performance Regression with MCP Overhead**
- **Issue**: JSON-RPC + subprocess overhead may slow evidence collection
- **Impact**: Users perceive Feature 006 as slower than Feature 003
- **Mitigation**:
  - Benchmark legacy vs. MCP performance
  - Optimize JSON serialization (use streaming)
  - Pool subprocess connections
  - Cache tool discovery results

**Risk 4: Configuration Complexity**
- **Issue**: MCP config is more complex than legacy connector config
- **Impact**: Users struggle to configure MCP servers correctly
- **Mitigation**:
  - Provide config wizard: `sdek mcp init`
  - Include templates for popular MCP servers
  - Validate config on startup with helpful error messages
  - Create troubleshooting guide

### Low Risk

**Risk 5: AI Provider Rate Limiting**
- **Issue**: More providers may mean different rate limit behaviors
- **Impact**: Users hit rate limits, analysis fails
- **Mitigation**:
  - Implement per-provider rate limiting
  - Display remaining quota in `sdek ai status`
  - Queue requests when approaching limits
  - Fallback to alternative provider

**Risk 6: Security Vulnerabilities in MCP Servers**
- **Issue**: MCP servers run as subprocesses with system access
- **Impact**: Malicious MCP server could compromise host
- **Mitigation**:
  - Document trusted MCP server sources
  - Add config option: `mcp.trusted_commands` allowlist
  - Run MCP servers in restricted mode (no network, limited filesystem)
  - Audit log all MCP server executions

---

## Open Questions for Planning Phase

1. **MCP Server Priorities**: Which MCP servers should we develop/integrate first beyond AWS?
   - GitHub MCP (code evidence)
   - Jira MCP (ticket evidence)
   - Slack MCP (communication evidence)
   - CI/CD MCP (build/deploy evidence)

2. **Provider Testing**: How to test all AI providers without incurring high costs?
   - Use provider free tiers
   - Mock responses for expensive models
   - Limit integration tests to one provider (OpenAI)

3. **Deprecation Timeline**: When to remove legacy connector API?
   - Option A: v2.0.0 (12 months from Feature 006 release)
   - Option B: Keep indefinitely for stability
   - Option C: v3.0.0 (24 months, longer transition)

4. **MCP Server Distribution**: How should users install MCP servers?
   - Document manual installation (npm, pip, uvx)
   - Provide installer script: `sdek mcp install aws-api`
   - Bundle popular MCP servers with sdek-cli binary

5. **Performance Benchmarks**: What are acceptable performance targets?
   - Tool discovery: <5 seconds for 10 MCP servers
   - Evidence collection: 50% faster than Feature 003 for 10+ sources
   - Provider switching: <100ms overhead vs. direct API call

---

## Conclusion

This research establishes a clear path to transform sdek-cli into an MCP-pluggable system following kubectl-ai's proven architecture patterns. Key takeaways:

1. **Phased approach**: Start with MCP client mode, defer server mode to Phase 2
2. **Backward compatibility**: Critical for adoption, use adapter pattern
3. **Factory pattern**: URL-based provider selection simplifies multi-provider support
4. **Safety-first**: Three-tier validation prevents accidental destructive operations
5. **Performance**: Parallel execution reduces collection time by ~50%
6. **Ecosystem**: MCP standard enables zero-code evidence source addition

Next step: Generate `data-model.md` and `contracts/` to formalize interfaces.
