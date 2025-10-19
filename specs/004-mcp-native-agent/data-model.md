# Data Model: MCP-Native Agent Orchestrator & Tooling Config

**Date**: 2025-10-19  
**Feature**: 004-mcp-native-agent  
**Status**: Complete

---

## Entity Definitions

### 1. MCPConfig

**Description**: Represents an MCP tool configuration loaded from a JSON file. Compatible with VS Code/Cursor MCP configuration structure.

**Fields**:
- `Name` (string, required): Unique identifier for the tool (e.g., "github", "jira", "aws")
- `Command` (string, required): Executable path or command name (e.g., "/usr/local/bin/mcp-github", "npx")
- `Args` ([]string, optional): Command-line arguments to pass to the executable
- `Env` (map[string]string, optional): Environment variables (e.g., `{"GITHUB_TOKEN": "${GITHUB_TOKEN}"}`)
- `Transport` (string, required): Communication protocol — "stdio" or "http"
- `Capabilities` ([]string, required): Declared capabilities (e.g., `["git.read", "git.commits.list"]`)
- `BaseURL` (string, optional): For HTTP transport, the base URL (e.g., "https://mcp-server.example.com")
- `Timeout` (duration, optional): Invocation timeout (default: 30s)
- `SchemaVersion` (string, required): Config schema version (e.g., "1.0.0")

**Validation Rules** (FR-003):
- `Name` must be non-empty, alphanumeric + hyphens only
- `Command` must be non-empty
- `Transport` must be "stdio" or "http"
- If `Transport == "http"`, `BaseURL` must be a valid URL
- `Capabilities` must be non-empty array
- `SchemaVersion` must match `^[0-9]+\.[0-9]+\.[0-9]+$`

**State Transitions**: None (config is immutable once loaded)

**Relationships**:
- One-to-one with `MCPTool` (after initialization)
- Loaded by `MCPLoader` from file system

**Example**:
```json
{
  "name": "github",
  "command": "/usr/local/bin/mcp-github",
  "args": ["--verbose"],
  "env": {
    "GITHUB_TOKEN": "${GITHUB_TOKEN}"
  },
  "transport": "stdio",
  "capabilities": ["git.read", "git.commits.list", "git.pr.list"],
  "timeout": "30s",
  "schemaVersion": "1.0.0"
}
```

---

### 2. MCPTool

**Description**: Represents the runtime state of an active MCP tool connection. Managed by the MCP registry.

**Fields**:
- `Name` (string, required): Tool identifier (from config)
- `Config` (*MCPConfig, required): Reference to the config
- `Status` (enum, required): Current health status — `Ready`, `Degraded`, `Offline`
- `Transport` (MCPTransport, required): Active transport instance (stdio or HTTP)
- `CircuitBreaker` (*CircuitBreaker, required): Circuit breaker state
- `Metrics` (ToolMetrics, required): Runtime metrics (see below)
- `LastHealthCheck` (time.Time, required): Timestamp of last health check
- `LastError` (error, optional): Most recent error (nil if healthy)
- `Enabled` (bool, required): Whether tool is administratively enabled (default: true)

**ToolMetrics Sub-Entity**:
- `HandshakeLatency` (time.Duration): Time to complete initial handshake
- `InvocationCount` (int64): Total invocations since startup
- `SuccessCount` (int64): Successful invocations
- `ErrorCount` (int64): Failed invocations
- `LastInvocationTime` (time.Time): Timestamp of most recent invocation
- `AverageLatency` (time.Duration): Moving average latency (last 100 calls)

**Validation Rules**:
- `Status` transitions: Ready ↔ Degraded ↔ Offline (managed by health checks)
- `Enabled` can be toggled by admin; if false, tool returns ErrToolDisabled

**State Transitions** (FR-007, FR-030):
```
Offline → [Handshake Success] → Ready
Ready → [N Consecutive Failures] → Degraded
Degraded → [Timeout Expired] → [Handshake Retry] → Ready or Offline
Degraded → [Circuit Breaker Opens] → Offline
Offline → [Circuit Breaker Half-Open] → [Test Success] → Ready
```

**Relationships**:
- One-to-one with `MCPConfig`
- Managed by `MCPRegistry`
- Invoked by `AgentInvoker`

---

### 3. MCPInvocationLog

**Description**: Audit record of an MCP tool invocation by an agent. Used for compliance reporting and security investigations.

**Fields**:
- `ID` (string, required): Unique log entry ID (UUID)
- `Timestamp` (time.Time, required): When invocation occurred
- `RunID` (string, required): Correlation ID for the analysis run
- `AgentID` (string, required): Identifier of the agent making the call (e.g., "evidence-collector")
- `AgentRole` (string, required): RBAC role of the agent
- `ToolName` (string, required): Tool invoked (e.g., "github")
- `Method` (string, required): Method called (e.g., "commits.list")
- `ArgsHash` (string, required): SHA256 hash of arguments (not raw args, for security)
- `RedactionApplied` (bool, required): Whether data was redacted
- `Duration` (time.Duration, required): Invocation duration
- `Status` (string, required): "success", "error", "permission_denied", "rate_limited"
- `ErrorMessage` (string, optional): Error details if status != "success"

**Validation Rules** (FR-020):
- All fields are required except `ErrorMessage`
- `ArgsHash` must be SHA256 (64 hex chars)
- `Status` must be one of: success, error, permission_denied, rate_limited

**State Transitions**: None (immutable once created)

**Relationships**:
- Many-to-one with `MCPTool` (many logs per tool)
- Many-to-one with analysis run (via `RunID`)
- Correlated with `Evidence` entities (via `RunID` + `Timestamp`)

**Storage**: Persisted to file system (e.g., `~/.sdek/logs/mcp-invocations.jsonl`) or external audit log system

---

### 4. AgentCapability

**Description**: Defines which RBAC capabilities an agent role possesses. Used for access control enforcement.

**Fields**:
- `Role` (string, required): Agent role name (e.g., "evidence-collector", "security-auditor", "read-only")
- `Capabilities` ([]string, required): List of allowed capabilities (e.g., `["github.read", "jira.search"]`)

**Capability Format**: `<tool>.<verb>` or `<tool>.*` (wildcard for all verbs on a tool)

**Validation Rules** (FR-012, FR-013):
- `Role` must be non-empty
- `Capabilities` must be non-empty array
- Each capability must match `^[a-z0-9-]+\.(([a-z0-9-]+\.)*[a-z0-9-]+|\*)$`

**State Transitions**: None (loaded from config)

**Relationships**:
- Many-to-many with `MCPTool` (a role can access multiple tools; a tool can be accessed by multiple roles)
- Evaluated by `RBACEnforcer` at invocation time

**Example**:
```yaml
roles:
  - role: evidence-collector
    capabilities:
      - github.read
      - github.commits.list
      - jira.search
      - aws.iam.list
  - role: security-auditor
    capabilities:
      - "*.*"  # Full access
  - role: read-only
    capabilities:
      - github.read
      - jira.read
```

---

### 5. ToolBudget

**Description**: Defines rate limits, concurrency limits, and timeout constraints for an MCP tool. Prevents abuse and cost overruns.

**Fields**:
- `ToolName` (string, required): Tool identifier
- `RateLimit` (RateLimit, required): Requests per second limit
- `ConcurrencyLimit` (int, required): Max parallel invocations
- `Timeout` (time.Duration, required): Per-invocation timeout
- `DailyQuota` (int, optional): Max invocations per day (0 = unlimited)

**RateLimit Sub-Entity**:
- `RequestsPerSecond` (float64): Max requests per second
- `BurstSize` (int): Max burst (e.g., 10 requests at once, then throttle)

**Validation Rules** (FR-014):
- `RateLimit.RequestsPerSecond` must be > 0
- `ConcurrencyLimit` must be >= 1
- `Timeout` must be > 0

**State Transitions**: None (loaded from config; enforced at runtime)

**Relationships**:
- One-to-one with `MCPTool`
- Enforced by `RBACEnforcer` (budget checks before allowing invocation)

**Example**:
```yaml
budgets:
  - tool: github
    rateLimit:
      requestsPerSecond: 10
      burstSize: 20
    concurrencyLimit: 5
    timeout: 30s
    dailyQuota: 5000
  - tool: aws
    rateLimit:
      requestsPerSecond: 5
      burstSize: 10
    concurrencyLimit: 3
    timeout: 60s
    dailyQuota: 1000
```

---

### 6. MCPHealthReport

**Description**: Result of a health check or handshake with an MCP tool. Used for diagnostics and status reporting.

**Fields**:
- `ToolName` (string, required): Tool identifier
- `Timestamp` (time.Time, required): When health check was performed
- `Status` (enum, required): `Healthy`, `Degraded`, `Unhealthy`
- `HandshakeSuccess` (bool, required): Whether handshake completed
- `HandshakeLatency` (time.Duration, required): Time to complete handshake
- `ErrorMessage` (string, optional): Error details if unhealthy
- `ServerCapabilities` ([]string, optional): Capabilities reported by server (for verification)
- `ProtocolVersion` (string, optional): MCP protocol version (e.g., "1.0.0")

**Validation Rules** (FR-007, FR-022):
- If `HandshakeSuccess == false`, `Status` must be `Unhealthy`
- `HandshakeLatency` must be >= 0

**State Transitions**: None (immutable snapshot)

**Relationships**:
- Many-to-one with `MCPTool` (multiple health checks per tool)
- Returned by `MCPRegistry.Test(toolName)`

**Usage**:
- CLI: `sdek mcp test <tool>` displays this report
- TUI: Health report shown in tool details view

---

### 7. Evidence (Integration Point)

**Description**: Represents evidence collected from external systems via MCP tools. This entity extends the existing `pkg/types/evidence.go` structure.

**New/Extended Fields for MCP**:
- `SourceType` (string): "mcp" (distinguishes MCP-collected evidence from other sources)
- `MCPToolName` (string): Tool that collected evidence (e.g., "github")
- `MCPMethod` (string): Method invoked (e.g., "commits.list")
- `MCPInvocationID` (string): Link to `MCPInvocationLog.ID` for audit trail

**Existing Fields** (preserved):
- `ID`, `Type`, `Description`, `Timestamp`, `Data`, `Metadata`, `Confidence`, `Tags`

**Validation Rules** (FR-017, FR-018, FR-019):
- If `SourceType == "mcp"`, `MCPToolName` and `MCPMethod` must be non-empty
- `Data` must be redacted if `Metadata["redacted"] == true`
- `Metadata["cache_key"]` used for caching (optional)

**State Transitions**: None (immutable once created)

**Relationships**:
- One-to-one with `MCPInvocationLog` (via `MCPInvocationID`)
- Many-to-one with analysis run (via `RunID`)
- Many-to-many with `Finding` (evidence supports findings)

**Normalization Pattern** (from research.md):
```go
func NormalizeMCPResponse(tool, method string, response any) (*Evidence, error) {
    // 1. Apply redaction
    redacted, didRedact := redactor.Redact(response)
    
    // 2. Create Evidence entity
    evidence := &Evidence{
        ID: uuid.New(),
        SourceType: "mcp",
        MCPToolName: tool,
        MCPMethod: method,
        Timestamp: time.Now(),
        Data: redacted,
        Metadata: map[string]any{
            "redacted": didRedact,
            "cache_key": fmt.Sprintf("%s:%s:%x", tool, method, argHash),
        },
    }
    
    return evidence, nil
}
```

---

## Entity Relationships Diagram

```
MCPConfig (file) ──1:1──> MCPTool (runtime)
                            │
                            ├──1:1──> CircuitBreaker
                            ├──1:1──> ToolMetrics
                            ├──1:1──> ToolBudget
                            │
                            └──1:N──> MCPInvocationLog (audit)
                                        │
                                        └──1:1──> Evidence (collected data)

AgentCapability (RBAC) ──M:N──> MCPTool (access control)

MCPHealthReport ──N:1──> MCPTool (diagnostics)
```

---

## Summary

This data model defines 7 core entities for the MCP-Native Agent Orchestrator feature:
1. **MCPConfig**: Configuration loaded from JSON
2. **MCPTool**: Runtime state and health
3. **MCPInvocationLog**: Audit trail for compliance
4. **AgentCapability**: RBAC capability mapping
5. **ToolBudget**: Rate limits and quotas
6. **MCPHealthReport**: Health check results
7. **Evidence**: Collected evidence (integration with existing types)

All entities align with functional requirements (FR-001 through FR-032) and constitutional principles (type safety, immutability where appropriate, clear relationships).

**Next Steps**: Generate Go type definitions and contracts in Phase 1.
