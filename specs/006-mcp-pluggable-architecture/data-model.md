# Data Model: MCP Pluggable Architecture

**Feature**: 006-mcp-pluggable-architecture
**Date**: 2025-10-26
**Status**: Complete

This document defines the core data structures and entities for the MCP pluggable architecture transformation of sdek-cli.

---

## Table of Contents
1. [MCP Configuration Entities](#mcp-configuration-entities)
2. [MCP Runtime Entities](#mcp-runtime-entities)
3. [AI Provider Entities](#ai-provider-entities)
4. [Tool System Entities](#tool-system-entities)
5. [Execution Entities](#execution-entities)
6. [Migration Entities](#migration-entities)

---

## MCP Configuration Entities

### MCPConfig
Top-level MCP configuration container.

**Fields:**
- `enabled` (boolean): Whether MCP integration is active (default: true)
- `prefer_mcp` (boolean): MCP tools take precedence over legacy connectors (default: true)
- `max_concurrent` (integer): Maximum concurrent MCP server connections (default: 10)
- `health_check_interval` (integer): Seconds between health checks (default: 300)
- `retry` (RetryConfig): Retry behavior configuration
- `servers` (map[string]MCPServerConfig): MCP server definitions by name

**Validation Rules:**
- `max_concurrent` must be > 0 and <= 100
- `health_check_interval` must be >= 60 (1 minute)
- Server names must be unique and match pattern `^[a-z0-9-]+$`

**Example:**
```yaml
mcp:
  enabled: true
  prefer_mcp: true
  max_concurrent: 10
  health_check_interval: 300
  retry:
    max_attempts: 3
    backoff: "exponential"
  servers:
    aws-api:
      command: "uvx"
      args: ["aws-api-mcp-server"]
      transport: "stdio"
    ...
```

---

### MCPServerConfig
Configuration for a single MCP server instance.

**Fields:**
- `command` (string): Executable path or command name (required for stdio)
- `args` ([]string): Command-line arguments (optional)
- `url` (string): HTTP endpoint URL (required for http transport)
- `transport` (enum): Communication mechanism ("stdio" | "http")
- `timeout` (integer): Request timeout in seconds (default: 60)
- `rate_limit` (integer): Requests per minute (0 = unlimited, default: 0)
- `env` (map[string]string): Environment variables for the process
- `headers` (map[string]string): HTTP headers (for http transport only)
- `health_url` (string): Health check endpoint (optional, for http transport)

**Validation Rules:**
- If `transport` = "stdio": `command` must be provided, `url` must be empty
- If `transport` = "http": `url` must be provided, `command` must be empty
- `timeout` must be > 0 and <= 600 (10 minutes)
- Environment variables support substitution: `${VAR_NAME}`
- Header values support substitution: `${VAR_NAME}`

**Example (stdio):**
```yaml
aws-api:
  command: "uvx"
  args: ["aws-api-mcp-server"]
  transport: "stdio"
  timeout: 60
  rate_limit: 30
  env:
    AWS_PROFILE: "readonly"
    READ_OPERATIONS_ONLY: "true"
    AWS_REGION: "${AWS_DEFAULT_REGION}"
```

**Example (http):**
```yaml
remote-mcp:
  url: "https://mcp.example.com/api"
  transport: "http"
  timeout: 30
  headers:
    Authorization: "Bearer ${MCP_API_TOKEN}"
    Content-Type: "application/json"
  health_url: "https://mcp.example.com/health"
```

---

### RetryConfig
Retry behavior for MCP server failures.

**Fields:**
- `max_attempts` (integer): Maximum retry attempts (default: 3)
- `backoff` (enum): Backoff strategy ("exponential" | "linear" | "constant")
- `initial_delay_ms` (integer): Initial retry delay in milliseconds (default: 1000)
- `max_delay_ms` (integer): Maximum retry delay in milliseconds (default: 30000)

**Backoff Formulas:**
- **Exponential**: `delay = initial_delay * (2 ^ attempt)` (capped at max_delay)
- **Linear**: `delay = initial_delay + (attempt * initial_delay)` (capped at max_delay)
- **Constant**: `delay = initial_delay` (always same)

**Example:**
```yaml
retry:
  max_attempts: 3
  backoff: "exponential"
  initial_delay_ms: 1000
  max_delay_ms: 30000
```

---

## MCP Runtime Entities

### MCPServer (Runtime State)
Runtime representation of an MCP server connection.

**Fields:**
- `name` (string): Server identifier from config
- `config` (MCPServerConfig): Configuration snapshot
- `transport` (Transport): Active transport instance
- `health_status` (enum): Current health ("healthy" | "degraded" | "down" | "unknown")
- `last_health_check` (timestamp): Last health check time
- `tools` ([]Tool): Discovered tools from this server
- `stats` (ServerStats): Runtime statistics

**State Transitions:**
```
unknown → healthy (successful initialize)
healthy → degraded (transient errors, but some calls succeed)
degraded → healthy (error rate drops below threshold)
degraded → down (error rate exceeds threshold or permanent error)
down → healthy (manual recovery or health check succeeds)
```

**Example:**
```json
{
  "name": "aws-api",
  "config": { /* MCPServerConfig */ },
  "transport": { /* StdioTransport or HTTPTransport */ },
  "health_status": "healthy",
  "last_health_check": "2025-10-26T10:30:00Z",
  "tools": [
    {
      "name": "call_aws",
      "description": "Execute AWS CLI commands",
      "parameters": { /* JSONSchema */ }
    },
    {
      "name": "suggest_aws_commands",
      "description": "Suggest AWS CLI commands",
      "parameters": { /* JSONSchema */ }
    }
  ],
  "stats": {
    "total_requests": 150,
    "successful_requests": 148,
    "failed_requests": 2,
    "error_rate": 0.013,
    "avg_latency_ms": 450
  }
}
```

---

### Transport (Interface)
Abstract transport mechanism for MCP communication.

**Methods:**
- `Initialize(ctx context.Context) error`: Establish connection
- `Send(ctx context.Context, request *JSONRPCRequest) (*JSONRPCResponse, error)`: Send message
- `Close() error`: Cleanup resources

**Implementations:**
1. **StdioTransport**: Subprocess communication via JSON-RPC over stdin/stdout
2. **HTTPTransport**: HTTP POST with JSON-RPC payload

---

### StdioTransport
Subprocess-based MCP communication.

**Fields:**
- `cmd` (*exec.Cmd): Running subprocess
- `stdin` (io.Writer): Subprocess stdin pipe
- `stdout` (io.Reader): Subprocess stdout pipe
- `stderr` (io.Reader): Subprocess stderr pipe (for logging)
- `encoder` (*json.Encoder): JSON encoder for requests
- `decoder` (*json.Decoder): JSON decoder for responses

**Lifecycle:**
```
1. Start subprocess: exec.Command(config.Command, config.Args...)
2. Set environment variables from config.Env
3. Connect stdin/stdout pipes
4. Send initialize handshake
5. Read tool list
6. Ready for CallTool requests
7. On Close(): Send shutdown, kill process, cleanup pipes
```

---

### HTTPTransport
HTTP-based MCP communication.

**Fields:**
- `base_url` (string): MCP server endpoint
- `headers` (map[string]string): HTTP headers
- `client` (*http.Client): Reusable HTTP client with timeout
- `sse_conn` (*SSEConnection): Optional Server-Sent Events connection (for server-to-client messages)

**Request Flow:**
```
1. POST base_url with JSON-RPC payload
2. Content-Type: application/json
3. Add custom headers from config
4. Parse response as JSON-RPC
5. Handle HTTP errors (timeout, 404, 500, etc.)
```

**SSE Support (Future):**
- MCP spec allows server → client messages via SSE
- Used for progress updates, notifications
- Optional feature, not in Phase 1

---

### ServerStats
Runtime statistics for an MCP server.

**Fields:**
- `total_requests` (integer): Total requests sent
- `successful_requests` (integer): Requests that returned success
- `failed_requests` (integer): Requests that returned error or timed out
- `error_rate` (float): `failed_requests / total_requests` (0.0 - 1.0)
- `avg_latency_ms` (integer): Average request latency in milliseconds
- `p95_latency_ms` (integer): 95th percentile latency (for alerting)
- `last_error` (string): Most recent error message
- `last_error_time` (timestamp): When last error occurred

**Thresholds:**
- **Degraded**: error_rate > 0.10 (10%) and < 0.50
- **Down**: error_rate >= 0.50 (50%) or consecutive failures > 5

---

## AI Provider Entities

### ProviderConfig
Configuration for an AI provider.

**Fields:**
- `url` (string): Provider URL with scheme (e.g., "openai://api.openai.com")
- `api_key` (string): Authentication key (supports `${VAR}` substitution)
- `model` (string): Model identifier (e.g., "gpt-4o", "gemma3:12b")
- `endpoint` (string): Optional custom endpoint override
- `timeout` (integer): Request timeout in seconds (default: 60)
- `max_retries` (integer): Maximum retry attempts (default: 3)
- `temperature` (float): Sampling temperature 0.0-2.0 (default: 0.0 for consistency)
- `max_tokens` (integer): Maximum response tokens (default: 4096)
- `extra` (map[string]string): Provider-specific settings

**URL Schemes:**
- `openai://` → OpenAI API
- `anthropic://` → Anthropic Claude API
- `gemini://` → Google Gemini API
- `bedrock://` → AWS Bedrock
- `vertexai://` → Google Vertex AI
- `ollama://` → Ollama local inference
- `llamacpp://` → llama.cpp local inference
- `azopenai://` → Azure OpenAI

**Example:**
```yaml
providers:
  openai:
    url: "openai://api.openai.com"
    api_key: "${SDEK_AI_OPENAI_KEY}"
    model: "gpt-4o"
    timeout: 60
    temperature: 0.0
    max_tokens: 4096

  ollama:
    url: "ollama://localhost:11434"
    model: "gemma3:12b"
    timeout: 120
    extra:
      num_ctx: "8192"
```

---

### Provider (Interface)
Unified interface for AI providers.

**Methods:**
- `AnalyzeWithContext(ctx context.Context, prompt string) (string, error)`: Send analysis request
- `Health(ctx context.Context) error`: Check provider availability
- `GetCallCount() int`: Get total calls made (for telemetry)
- `GetLastPrompt() string`: Get last prompt sent (for debugging)

**Responsibilities:**
- Send prompts to AI provider
- Parse responses
- Handle provider-specific authentication
- Implement retry logic
- Track usage statistics

---

### ChatSession
Multi-turn conversation with an AI provider.

**Fields:**
- `id` (string): Unique session identifier (UUID)
- `provider` (Provider): Active AI provider instance
- `messages` ([]Message): Conversation history
- `functions` ([]FunctionDefinition): Available tools/functions
- `config` (SessionConfig): Session-specific configuration
- `metadata` (map[string]string): Custom metadata (framework, control ID, etc.)

**Methods:**
- `AddMessage(role string, content string) error`: Append message to history
- `SetFunctions(functions []FunctionDefinition) error`: Register available tools
- `Send(ctx context.Context) (Response, error)`: Send conversation to provider
- `Reset()`: Clear message history (keep functions)

**Example:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "provider": { /* Provider instance */ },
  "messages": [
    {
      "role": "system",
      "content": "You are a compliance analysis expert..."
    },
    {
      "role": "user",
      "content": "Analyze this evidence for SOC2 CC6.1..."
    },
    {
      "role": "assistant",
      "content": "{\"finding_summary\": \"...\"}"
    }
  ],
  "functions": [
    {
      "name": "call_aws",
      "description": "Execute AWS CLI commands",
      "parameters": { /* JSONSchema */ }
    }
  ],
  "config": {
    "temperature": 0.0,
    "max_tokens": 4096
  },
  "metadata": {
    "framework": "soc2",
    "control_id": "CC6.1"
  }
}
```

---

### Message
Single message in a conversation.

**Fields:**
- `role` (enum): Message sender ("system" | "user" | "assistant" | "function")
- `content` (string): Message text
- `function_call` (FunctionCall): Optional function call (if role = "assistant")
- `function_result` (interface{}): Optional function result (if role = "function")
- `timestamp` (timestamp): When message was created

**Roles:**
- **system**: Instructions/context for the AI
- **user**: User input/questions
- **assistant**: AI responses
- **function**: Function execution results

---

### FunctionDefinition
Tool/function available to the AI.

**Fields:**
- `name` (string): Function identifier (e.g., "call_aws", "kubectl")
- `description` (string): Natural language description for AI
- `parameters` (JSONSchema): Expected input schema

**Example:**
```json
{
  "name": "call_aws",
  "description": "Execute validated AWS CLI commands. Use this to query AWS infrastructure for compliance evidence.",
  "parameters": {
    "type": "object",
    "properties": {
      "command": {
        "type": "string",
        "description": "AWS CLI command without 'aws' prefix (e.g., 'iam list-users')"
      }
    },
    "required": ["command"]
  }
}
```

---

## Tool System Entities

### ToolRegistry
Unified catalog of all available tools.

**Fields:**
- `builtin_tools` (map[string]Tool): Built-in tools (kubectl, bash)
- `mcp_tools` (map[string]Tool): Tools from MCP servers
- `legacy_tools` (map[string]Tool): Wrapped legacy connectors
- `safety_validator` (SafetyValidator): Safety analysis component

**Methods:**
- `Register(tool Tool) error`: Add a tool to registry
- `List() []Tool`: Get all registered tools
- `Get(name string) (Tool, error)`: Retrieve tool by name
- `Execute(ctx context.Context, call ToolCall) (interface{}, error)`: Execute tool
- `Analyze(call ToolCall) ToolCallAnalysis`: Analyze safety before execution

**Discovery Order:**
1. MCP tools (if `prefer_mcp: true`)
2. Built-in tools
3. Legacy tools (if not shadowed by MCP)

---

### Tool
Represents an executable capability.

**Fields:**
- `name` (string): Unique tool identifier
- `description` (string): Natural language description
- `parameters` (JSONSchema): Input parameter schema
- `source` (enum): Tool origin ("builtin" | "mcp" | "legacy")
- `server_name` (string): MCP server name (if source = "mcp")
- `safety_tier` (enum): Safety classification ("safe" | "interactive" | "modifies_resource")
- `handler` (ToolHandler): Execution implementation

**Example (MCP Tool):**
```json
{
  "name": "call_aws",
  "description": "Execute validated AWS CLI commands",
  "parameters": {
    "type": "object",
    "properties": {
      "command": {"type": "string"}
    },
    "required": ["command"]
  },
  "source": "mcp",
  "server_name": "aws-api",
  "safety_tier": "safe",
  "handler": { /* MCPToolHandler */ }
}
```

**Example (Built-in Tool):**
```json
{
  "name": "kubectl",
  "description": "Execute kubectl commands against Kubernetes cluster",
  "parameters": {
    "type": "object",
    "properties": {
      "command": {"type": "string"}
    },
    "required": ["command"]
  },
  "source": "builtin",
  "server_name": "",
  "safety_tier": "modifies_resource",
  "handler": { /* KubectlHandler */ }
}
```

---

### ToolCall
Request to execute a tool.

**Fields:**
- `tool_name` (string): Name of tool to execute
- `arguments` (map[string]interface{}): Input parameters as JSON object
- `context` (map[string]string): Additional context (user_id, session_id, etc.)

**Example:**
```json
{
  "tool_name": "call_aws",
  "arguments": {
    "command": "iam list-users --output json"
  },
  "context": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "analyst@example.com"
  }
}
```

---

### ToolCallAnalysis
Safety analysis result for a tool call.

**Fields:**
- `is_interactive` (boolean): Tool requires interactive terminal
- `modifies_resource` (boolean): Tool may mutate system state
- `requires_approval` (boolean): User confirmation required
- `risk_level` (enum): Risk classification ("low" | "medium" | "high")
- `rationale` (string): Explanation of safety assessment

**Decision Logic:**
```
if is_interactive:
    requires_approval = true
    risk_level = "high"
elif modifies_resource:
    requires_approval = true
    risk_level = "medium"
else:
    requires_approval = false
    risk_level = "low"
```

**Example:**
```json
{
  "is_interactive": false,
  "modifies_resource": true,
  "requires_approval": true,
  "risk_level": "medium",
  "rationale": "Command 'aws ec2 terminate-instances' detected - modifies AWS resources"
}
```

---

### ToolExecutionResult
Result of a tool execution.

**Fields:**
- `tool_name` (string): Tool that was executed
- `success` (boolean): Whether execution succeeded
- `output` (interface{}): Result data (format varies by tool)
- `error` (string): Error message if success = false
- `latency_ms` (integer): Execution time in milliseconds
- `timestamp` (timestamp): When execution completed

**Example (Success):**
```json
{
  "tool_name": "call_aws",
  "success": true,
  "output": {
    "Users": [
      {"UserName": "admin", "UserId": "AIDAI..."},
      {"UserName": "analyst", "UserId": "AIDAI..."}
    ]
  },
  "error": "",
  "latency_ms": 450,
  "timestamp": "2025-10-26T10:30:15Z"
}
```

**Example (Failure):**
```json
{
  "tool_name": "call_aws",
  "success": false,
  "output": null,
  "error": "AWS MCP server timeout after 60s",
  "latency_ms": 60000,
  "timestamp": "2025-10-26T10:31:00Z"
}
```

---

## Execution Entities

### AnalysisRequest
Request to perform compliance analysis.

**Fields:**
- `preamble` (types.ContextPreamble): Framework + control context (from Feature 003)
- `evidence` (types.EvidenceBundle): Collected evidence events (from Feature 003)
- `tools` ([]Tool): Available tools for AI to use
- `options` (AnalysisOptions): Request-specific options

**Example:**
```json
{
  "preamble": {
    "framework": "SOC2",
    "version": "2017",
    "section": "CC6.1",
    "excerpt": "The entity implements logical access security...",
    "control_ids": ["CC6.1"]
  },
  "evidence": {
    "events": [
      {
        "id": "evt-001",
        "source": "github",
        "type": "commit",
        "timestamp": "2025-10-20T10:00:00Z",
        "content": "feat: add MFA enforcement"
      }
    ],
    "metadata": {
      "total_events": 1,
      "redacted_fields": 0
    }
  },
  "tools": [
    {
      "name": "call_aws",
      "description": "Execute AWS CLI commands",
      "parameters": { /* JSONSchema */ }
    }
  ],
  "options": {
    "no_cache": false,
    "max_tool_calls": 10,
    "allow_tool_use": true
  }
}
```

---

### AnalysisOptions
Options for an analysis request.

**Fields:**
- `no_cache` (boolean): Bypass cache, always call provider (default: false)
- `max_tool_calls` (integer): Maximum tool executions per analysis (default: 10)
- `allow_tool_use` (boolean): Whether AI can use tools (default: true)
- `timeout_seconds` (integer): Maximum analysis duration (default: 120)

---

### ExecutionPlan
Orchestration plan for collecting evidence from multiple sources.

**Fields:**
- `id` (string): Unique plan identifier (UUID)
- `framework` (string): Target framework (e.g., "soc2")
- `control_ids` ([]string): Target controls
- `items` ([]PlanItem): Individual collection tasks
- `status` (enum): Plan state ("draft" | "approved" | "executing" | "completed" | "failed")
- `created_at` (timestamp): When plan was generated
- `executed_at` (timestamp): When execution started
- `completed_at` (timestamp): When execution finished

**Status Transitions:**
```
draft → approved (user approval)
approved → executing (ExecutePlan called)
executing → completed (all items finished)
executing → failed (critical error occurred)
```

---

### PlanItem
Single task in an execution plan.

**Fields:**
- `id` (string): Unique item identifier
- `source` (string): Evidence source (e.g., "aws-api", "github-mcp")
- `tool_name` (string): MCP tool to use
- `query` (string): Query/filter string
- `arguments` (map[string]interface{}): Tool-specific arguments
- `estimated_signal_strength` (float): Expected relevance (0.0 - 1.0)
- `approval_status` (enum): Approval state ("pending" | "approved" | "denied" | "auto-approved")
- `execution_status` (enum): Execution state ("pending" | "running" | "completed" | "failed")
- `result` (ToolExecutionResult): Execution result (if completed)

**Example:**
```json
{
  "id": "item-001",
  "source": "aws-api",
  "tool_name": "call_aws",
  "query": "iam list-users",
  "arguments": {
    "command": "iam list-users --output json"
  },
  "estimated_signal_strength": 0.85,
  "approval_status": "auto-approved",
  "execution_status": "completed",
  "result": {
    "tool_name": "call_aws",
    "success": true,
    "output": { /* AWS response */ },
    "latency_ms": 450
  }
}
```

---

## Migration Entities

### MigrationLog
Record of configuration migration.

**Fields:**
- `version` (string): Migration version (e.g., "003-to-006")
- `timestamp` (timestamp): When migration occurred
- `changes` ([]MigrationChange): List of changes applied
- `backup_path` (string): Path to backup of original config

**Example:**
```json
{
  "version": "003-to-006",
  "timestamp": "2025-10-26T10:00:00Z",
  "changes": [
    {
      "type": "connector_to_mcp",
      "source": "ai.connectors.github",
      "target": "mcp.servers.github-mcp",
      "details": "Migrated GitHub connector to MCP server"
    },
    {
      "type": "connector_to_mcp",
      "source": "ai.connectors.aws",
      "target": "mcp.servers.aws-api",
      "details": "Migrated AWS connector to MCP server"
    }
  ],
  "backup_path": "/Users/user/.sdek/config.yaml.backup.2025-10-26"
}
```

---

### MigrationChange
Single change during migration.

**Fields:**
- `type` (enum): Change type ("connector_to_mcp" | "provider_url" | "deprecation")
- `source` (string): Original config path
- `target` (string): New config path
- `details` (string): Human-readable description

---

## Validation Rules Summary

### Configuration Validation
1. **MCPConfig**:
   - `max_concurrent` in range [1, 100]
   - `health_check_interval` >= 60
   - Server names match `^[a-z0-9-]+$`
   - No duplicate server names

2. **MCPServerConfig**:
   - Exactly one of (`command`, `url`) must be set based on transport
   - `timeout` in range [1, 600]
   - Environment variables properly formatted: `${VAR_NAME}`

3. **ProviderConfig**:
   - URL scheme must be valid: openai|anthropic|gemini|bedrock|ollama|llamacpp|vertexai|azopenai
   - `temperature` in range [0.0, 2.0]
   - `max_tokens` > 0

### Runtime Validation
1. **ToolCall**:
   - `tool_name` must exist in ToolRegistry
   - `arguments` must match tool's parameter schema
   - Context must include `session_id` if auditing enabled

2. **ExecutionPlan**:
   - Must have at least one PlanItem
   - All items must reference valid MCP servers
   - Status transitions must follow state machine

---

## Relationships

```
MCPConfig (1) ──has many──> (N) MCPServerConfig
MCPServerConfig (1) ──creates──> (1) MCPServer (runtime)
MCPServer (1) ──discovers──> (N) Tool
ToolRegistry (1) ──aggregates──> (N) Tool
Provider (1) ──uses──> (1) ChatSession
ChatSession (1) ──contains──> (N) Message
AnalysisRequest (1) ──uses──> (1) ContextPreamble (Feature 003)
AnalysisRequest (1) ──uses──> (1) EvidenceBundle (Feature 003)
AnalysisRequest (1) ──references──> (N) Tool
ExecutionPlan (1) ──contains──> (N) PlanItem
PlanItem (1) ──executes via──> (1) Tool
Tool (1) ──routes to──> (1) MCPServer
```

---

## Entity Count Estimate

**Configuration Entities**: 3 (MCPConfig, MCPServerConfig, RetryConfig)
**Runtime Entities**: 4 (MCPServer, Transport, StdioTransport, HTTPTransport)
**AI Provider Entities**: 5 (ProviderConfig, Provider, ChatSession, Message, FunctionDefinition)
**Tool System Entities**: 5 (ToolRegistry, Tool, ToolCall, ToolCallAnalysis, ToolExecutionResult)
**Execution Entities**: 4 (AnalysisRequest, AnalysisOptions, ExecutionPlan, PlanItem)
**Migration Entities**: 2 (MigrationLog, MigrationChange)

**Total Core Entities**: 23

---

## Next Steps

With the data model complete, the next deliverables are:

1. **contracts/*.json**: JSON Schema definitions for API contracts
2. **quickstart.md**: User-friendly getting started guide
3. **tasks.md**: Implementation task breakdown (via `/speckit.tasks` command)

This data model serves as the foundation for implementation, ensuring all components share a common understanding of data structures and relationships.
