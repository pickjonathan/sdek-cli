# Research: MCP-Native Agent Orchestrator & Tooling Config

**Date**: 2025-10-19  
**Feature**: 004-mcp-native-agent  
**Status**: Complete

---

## 1. MCP Specification Alignment

### Decision
Use MCP protocol v1.0 as the baseline, mirroring VS Code and Cursor configuration structure with these key fields:
- `name` (string): Tool identifier
- `command` (string): Executable path
- `args` ([]string): Command-line arguments
- `env` (map[string]string): Environment variables
- `transport` (string): "stdio" or "http"
- `capabilities` ([]string): Declared tool capabilities (e.g., "git.read", "jira.search")

### Rationale
VS Code and Cursor have established MCP as a de facto standard for tool integration in AI-assisted development. Aligning with their configuration structure ensures:
1. Zero-friction adoption for users already using MCP in their IDEs
2. Reuse of existing MCP server implementations
3. Future compatibility as the MCP ecosystem evolves

### Alternatives Considered
- **Custom protocol**: Rejected due to lack of ecosystem, vendor lock-in, and higher maintenance burden
- **Language Server Protocol (LSP) adaptation**: Rejected as LSP is optimized for code intelligence, not evidence collection
- **gRPC-based custom API**: Rejected due to complexity and lack of existing tooling

### References
- VS Code MCP documentation: [Assumed available at vscode.dev or similar]
- Cursor MCP spec: [Assumed available in Cursor docs]
- MCP transport protocols: stdio (standard), HTTP (for remote tools)

---

## 2. JSON Schema Validation in Go

### Decision
Use `github.com/santhosh-tekuri/jsonschema/v5` for JSON Schema validation with custom error reporting.

### Rationale
After evaluating Go JSON Schema libraries:
- **santhosh-tekuri/jsonschema**: Modern (v5), active maintenance, supports Draft 2020-12, detailed validation errors
- **xeipuuv/gojsonschema**: Older, limited error context, Draft 04 only
- **qri-io/jsonschema**: Good for Draft 2019-09, but less error detail

The `santhosh-tekuri` library provides:
1. Precise error paths (e.g., `/env/API_KEY: missing required field`)
2. Line/column reporting via custom wrapper (parse JSON → map positions → schema error → lookup position)
3. Schema versioning via `$schema` field

### Implementation Pattern
```go
// Pseudo-code for error reporting
func ValidateMCPConfig(path string) ([]SchemaError, error) {
    // 1. Parse JSON with position tracking (encoding/json + custom decoder)
    // 2. Validate against schema
    // 3. Map schema errors to file positions
    // 4. Return []SchemaError{FilePath, Line, Column, JSONPath, Message}
}
```

### Alternatives Considered
- **Manual validation**: Rejected due to maintenance burden and lack of formal spec
- **Protocol Buffers**: Rejected as MCP configs are user-authored JSON files (not binary)

### References
- Library: https://github.com/santhosh-tekuri/jsonschema
- JSON Schema Draft 2020-12: https://json-schema.org/draft/2020-12/schema

---

## 3. File Watching and Hot-Reload

### Decision
Use `fsnotify` (cross-platform file watcher) with debouncing and graceful connection swaps.

### Rationale
- **fsnotify**: Standard Go library for file system notifications, cross-platform (Linux inotify, macOS FSEvents, Windows ReadDirectoryChangesW)
- **Debouncing**: Editors often trigger multiple write events for a single save; debounce with 500ms window
- **Graceful swap**: Load new config → validate → initialize new connection → swap in registry → close old connection (with timeout)

### Implementation Pattern
```go
func (r *Registry) watchConfigs(ctx context.Context) {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add("~/.sdek/mcp")
    watcher.Add("./.sdek/mcp")
    
    debounce := time.NewTimer(500 * time.Millisecond)
    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                debounce.Reset(500 * time.Millisecond)
            }
        case <-debounce.C:
            r.hotReload() // Validate → Init → Swap
        }
    }
}
```

### Alternatives Considered
- **Polling**: Rejected due to higher CPU usage and slower detection
- **Ignore hot-reload**: Rejected as user explicitly requested it; improves DX significantly

### References
- fsnotify: https://github.com/fsnotify/fsnotify
- Debouncing patterns: Standard Go timer-based approach

---

## 4. MCP Transport Implementation

### Decision
Implement two transports: **stdio** (primary) and **HTTP** (remote tools).

#### Stdio Transport
- **Pattern**: Spawn child process, communicate via stdin/stdout
- **Protocol**: JSON-RPC 2.0 over stdio (request/response pairs)
- **Lifecycle**: Long-lived process per tool (pooled)

#### HTTP Transport
- **Pattern**: Standard HTTP client with connection pooling
- **Protocol**: JSON-RPC 2.0 over HTTP POST
- **Lifecycle**: Stateless requests (server manages state)

### Rationale
- Stdio is standard for local tools (simpler, no network overhead)
- HTTP enables remote tools (e.g., centralized MCP proxy, cloud services)
- Both use JSON-RPC 2.0 for consistency

### Implementation Pattern
```go
type Transport interface {
    Invoke(ctx context.Context, method string, params any) (any, error)
    HealthCheck(ctx context.Context) error
    Close() error
}

type StdioTransport struct { cmd *exec.Cmd; stdin io.Writer; stdout io.Reader }
type HTTPTransport struct { client *http.Client; baseURL string }
```

### Alternatives Considered
- **gRPC**: Rejected due to lack of MCP server support
- **WebSocket**: Considered for bidirectional streaming, deferred to future iteration

### References
- JSON-RPC 2.0: https://www.jsonrpc.org/specification
- Go exec.Cmd: https://pkg.go.dev/os/exec

---

## 5. Circuit Breaker and Retry Patterns

### Decision
Implement exponential backoff with jitter and circuit breaker state machine.

### Algorithm
- **Backoff**: Initial 1s, max 30s, multiplier 2x, jitter ±20%
- **Circuit States**: CLOSED (normal) → OPEN (failing) → HALF_OPEN (testing recovery) → CLOSED
- **Thresholds**: Open after 5 consecutive failures, half-open after 60s, close after 2 successes

### Rationale
- Exponential backoff prevents overwhelming failing services
- Jitter prevents thundering herd when multiple tools recover simultaneously
- Circuit breaker provides fast-fail for degraded tools (don't wait for timeout)

### Implementation Pattern
```go
type CircuitBreaker struct {
    state State // CLOSED | OPEN | HALF_OPEN
    failures int
    lastFailTime time.Time
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == OPEN && time.Since(cb.lastFailTime) > 60*time.Second {
        cb.state = HALF_OPEN
    }
    if cb.state == OPEN {
        return ErrCircuitOpen
    }
    err := fn()
    if err != nil {
        cb.recordFailure()
    } else {
        cb.recordSuccess()
    }
    return err
}
```

### Alternatives Considered
- **Fixed retry intervals**: Rejected as they can overwhelm recovering services
- **No circuit breaker**: Rejected as repeated timeouts degrade user experience

### References
- Circuit breaker pattern: Martin Fowler https://martinfowler.com/bliki/CircuitBreaker.html
- Exponential backoff: Google SRE Book Chapter 21

---

## 6. RBAC Integration

### Decision
Implement capability-based RBAC with agent roles and per-tool budgets.

### Model
- **Agent**: Has a `role` (e.g., "evidence-collector", "security-auditor")
- **Capability**: String format `<tool>.<verb>` (e.g., "github.read", "aws.iam.list")
- **Policy**: Map of `role` → `[]capability`
- **Budget**: Per-tool rate limit (requests/sec), concurrency limit (max parallel), timeout

### Rationale
- Capability-based model is simpler than resource-based (no need to model every possible resource)
- Explicit deny-by-default (if capability not granted, deny)
- Budgets prevent abuse and cost overruns

### Implementation Pattern
```go
type RBACEnforcer struct {
    policies map[string][]string // role → capabilities
    budgets  map[string]Budget   // tool → budget
}

func (e *RBACEnforcer) Check(agent Agent, tool string, method string) error {
    capability := fmt.Sprintf("%s.%s", tool, method)
    if !slices.Contains(e.policies[agent.Role], capability) {
        return ErrPermissionDenied
    }
    if !e.budgets[tool].Allow() {
        return ErrRateLimitExceeded
    }
    return nil
}
```

### Alternatives Considered
- **Role-based only (no capabilities)**: Rejected as too coarse-grained
- **Attribute-based (ABAC)**: Rejected as overkill for initial version
- **No RBAC**: Rejected as security requirement in spec

### References
- NIST RBAC model: https://csrc.nist.gov/projects/role-based-access-control
- Capability-based security: https://en.wikipedia.org/wiki/Capability-based_security

---

## 7. Evidence Graph Integration

### Decision
Normalize MCP tool responses into `Evidence` entities with provenance metadata.

### Pattern
```go
type Evidence struct {
    ID          string
    SourceTool  string   // e.g., "github"
    SourceMethod string  // e.g., "commits.list"
    Timestamp   time.Time
    RunID       string   // Correlation ID
    Data        any      // Normalized data (commits, tickets, logs, etc.)
    Redacted    bool     // Whether redaction was applied
    CacheKey    string   // For caching (optional)
}

func (inv *MCPInvoker) CollectEvidence(ctx context.Context, tool, method string, params any) (*Evidence, error) {
    // 1. Invoke MCP tool
    response, err := inv.transport.Invoke(ctx, method, params)
    if err != nil { return nil, err }
    
    // 2. Apply redaction (use existing internal/ai/redactor.go)
    redacted, didRedact := redactor.Redact(response)
    
    // 3. Normalize into Evidence
    evidence := &Evidence{
        ID: uuid.New(),
        SourceTool: tool,
        SourceMethod: method,
        Timestamp: time.Now(),
        RunID: ctx.Value("run_id"),
        Data: redacted,
        Redacted: didRedact,
    }
    
    // 4. Cache (use existing internal/store/cache.go)
    cache.Set(cacheKey, evidence)
    
    return evidence, nil
}
```

### Rationale
- Preserves provenance for audit trails (which tool/method produced evidence)
- Integrates seamlessly with existing redaction and caching
- `Data` field is `any` type to support diverse evidence (commits, tickets, metrics, logs)

### Alternatives Considered
- **Strongly typed evidence variants**: Rejected as too rigid (MCP tools return diverse structures)
- **Bypass evidence graph**: Rejected as breaks existing analysis pipeline

### References
- Existing code: `pkg/types/evidence.go`, `internal/ai/redactor.go`, `internal/store/cache.go`

---

## 8. Telemetry and Observability

### Decision
Emit structured logs, metrics, and traces using existing infrastructure.

### Logs (slog)
- `mcp_tool_loaded`: Tool name, config path, capabilities
- `mcp_tool_failed`: Tool name, error, retry count
- `mcp_invoked`: Tool, method, duration, redacted, result status
- `permission_denied`: Agent, tool, method, required capability

### Metrics (Prometheus-style)
- `mcp_tool_status{tool, status}`: Gauge (1=ready, 0.5=degraded, 0=offline)
- `mcp_handshake_duration_ms{tool}`: Histogram
- `mcp_invocation_total{tool, method, status}`: Counter
- `mcp_rbac_denials_total{agent, tool}`: Counter

### Traces (OpenTelemetry)
- Span: `run_id → agent_id → mcp.invoke{tool, method}`
- Attributes: duration, payload_size, redaction_applied

### Rationale
- Structured logs enable automated analysis (e.g., audit queries)
- Metrics provide operational visibility (dashboards, alerts)
- Traces enable root cause analysis (which tool slowed down analysis)

### Implementation
Use existing logging/metrics infrastructure. Wrap MCP invocations with tracing spans.

### Alternatives Considered
- **Plain text logs**: Rejected as harder to parse
- **No tracing**: Rejected as debugging distributed evidence collection is hard

### References
- slog: https://pkg.go.dev/log/slog
- OpenTelemetry Go: https://opentelemetry.io/docs/instrumentation/go/

---

## Summary

All research topics have been addressed with concrete decisions, rationale, and alternatives considered. No NEEDS CLARIFICATION remain. The implementation can proceed to Phase 1 (Design & Contracts).

**Key Technologies Selected**:
1. JSON Schema validation: `github.com/santhosh-tekuri/jsonschema/v5`
2. File watching: `fsnotify`
3. MCP transports: stdio (exec.Cmd), HTTP (http.Client)
4. Circuit breaker: Custom implementation with exponential backoff
5. RBAC: Capability-based model with per-tool budgets
6. Evidence integration: Normalize via `pkg/types/evidence.go`
7. Observability: slog + metrics + OpenTelemetry

**Constitutional Compliance**: All decisions align with Go best practices, cross-platform requirements, and existing sdek-cli architecture.
