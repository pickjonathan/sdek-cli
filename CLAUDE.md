# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**sdek-cli** is a compliance evidence mapping tool that automates SOC2, ISO 27001, and PCI DSS compliance analysis. It ingests evidence from multiple sources (Git, Jira, Slack, CI/CD, Docs), maps them to compliance controls, and uses AI-powered context injection to generate policy-grounded compliance findings.

## Essential Commands

### Build and Test
```bash
# Build the binary
make build
# or
go build -o sdek main.go

# Run tests with race detection and coverage
make test
# or
go test -v -race -coverprofile=coverage.out ./...

# Run specific package tests
go test ./cmd/...
go test ./internal/ai/...

# Generate coverage report
make test-coverage

# Run linting
make lint
```

### Development Workflow
```bash
# Build and run
make run

# Run in development mode
make dev
# or
go run main.go

# Generate demo data for testing
./sdek seed --demo

# Launch TUI for manual testing
./sdek tui

# Test AI analysis (requires API key)
export SDEK_AI_OPENAI_KEY="sk-..."
./sdek config set ai.enabled true
./sdek config set ai.provider openai
./sdek analyze
```

### Testing AI Features
```bash
# Run integration tests (requires test data)
go test ./tests/integration/...

# Run specific integration test
go test ./tests/integration -run TestContextModeE2E

# Run unit tests only
go test ./internal/... ./pkg/...

# Run with verbose output
go test -v ./tests/integration/context_mode_test.go
```

## Architecture

### Core Data Flow

```
Sources → Events → Evidence → Findings → Reports
          ↓        ↓          ↓
       Heuristic  AI         Risk
       Mapping  Analysis   Scoring
```

**Key Concepts:**
- **Events**: Raw timestamped actions from sources (Git commits, Jira tickets, etc.)
- **Evidence**: Mappings between events and compliance controls with confidence scores
- **Findings**: Risk assessments for controls based on evidence quality and gaps
- **Context Injection**: AI prompts include exact policy language for grounded analysis

### State Management

All application state is persisted to `~/.sdek/state.json`:
- Sources, Events, Frameworks, Controls, Evidence, Findings
- State is loaded/saved via `internal/store/state.go`
- Use `store.Load()` and `state.Save()` for state operations
- Autosave wrapper available for CLI commands

### Configuration System

Configuration precedence (highest to lowest):
1. Command-line flags (Cobra)
2. Environment variables (`SDEK_*` prefix)
3. Config file (`~/.sdek/config.yaml`)
4. Default values in `pkg/types/config.go`

Viper handles all config binding automatically via `cmd/root.go:initConfig()`.

### AI Analysis Architecture

**Three modes** (Feature 003):
1. **Disabled**: Heuristic-only keyword matching
2. **Context Mode**: User triggers AI analysis with policy context injection
3. **Autonomous Mode**: AI proactively suggests evidence collection (experimental)

**Key Components:**
- `internal/ai/engine.go`: Core Engine interface with provider abstraction
- `internal/ai/providers/`: OpenAI and Anthropic implementations
- `internal/ai/prompt.go`: Context preamble injection
- `internal/ai/privacy.go`: PII/secret redaction (MANDATORY before AI calls)
- `internal/ai/cache.go`: SHA256-based prompt/response caching
- `internal/policy/excerpts.go`: Policy excerpt loading from JSON

**AI Analysis Flow:**
1. Load policy excerpts (framework + section)
2. Build `types.ContextPreamble` with exact policy text
3. Redact PII/secrets from evidence events
4. Check cache using prompt hash
5. Call AI provider with injected context
6. Parse structured JSON response into `types.Finding`
7. Save to cache + state

### MCP Integration Architecture (Feature 006)

**Overview:**
Feature 006 transforms sdek-cli into an MCP-pluggable system, enabling zero-code evidence source addition through external MCP servers.

**MCP Data Flow:**
```
AI Engine → ConnectorAdapter → MCPManager → MCPServer → Transport → External MCP Server
                                    ↓
                            Health Monitor + Retry Logic
```

**Key Components:**
- `internal/mcp/manager.go`: Multi-server orchestration with health monitoring
- `internal/mcp/client.go`: MCP client with handshake and tool discovery
- `internal/mcp/transport.go`: Transport abstraction (stdio, HTTP)
- `internal/mcp/stdio_client.go`: Subprocess-based MCP communication
- `internal/mcp/http_client.go`: HTTP-based MCP communication
- `internal/mcp/connector_adapter.go`: Bridges MCP to AI engine
- `internal/mcp/normalizer.go`: Converts MCP results to EvidenceEvent format
- `internal/ai/mcp_integration.go`: AI engine integration helpers

**MCP Configuration:**
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
      env:
        AWS_PROFILE: "readonly"
```

**Usage:**
```bash
# List MCP servers
sdek mcp list-servers

# List available tools
sdek mcp list-tools

# Test server connection
sdek mcp test aws-api

# Use in evidence collection (source format: "server:tool")
sdek ai plan --sources aws-api:call_aws
```

**Creating Engine with MCP:**
```go
// With MCP support
engine, err := ai.NewEngineWithMCP(ctx, config, provider)

// Without MCP (legacy)
engine := ai.NewEngine(config, provider)
```

**MCP Features:**
- **Graceful Degradation**: Individual server failures don't crash system
- **Health Monitoring**: Periodic background checks with status tracking
- **Retry Logic**: Configurable exponential/linear/constant backoff
- **Tool Discovery**: Automatic detection of available tools from servers
- **Result Normalization**: Converts diverse MCP formats to EvidenceEvent
- **Concurrent Execution**: Parallel tool calls with concurrency limits

### Tool Registry & Multi-System Orchestration (Feature 006 - Phase 5)

**Overview:**
Phase 5 adds a unified tool registry with three-tier safety validation, parallel execution, and comprehensive audit logging for multi-source evidence collection.

**Tool System Data Flow:**
```
User Request → AI Engine → Tool Registry → Safety Validator
                                ↓              ↓
                         Tool Lookup     Risk Assessment
                                ↓              ↓
                           Executor  ←  Approval Check
                                ↓
                         Parallel Execution (Semaphore)
                                ↓
                    ┌──────────┴──────────┐
                    ↓                     ↓
              MCP Manager          Builtin/Legacy
                    ↓                     ↓
              Tool Results          Tool Results
                    ↓                     ↓
                    └──────────┬──────────┘
                               ↓
                        Audit Logger
                               ↓
                      Aggregated Results
```

**Key Components:**
- `internal/tools/registry.go`: Unified tool catalog (builtin, MCP, legacy)
- `internal/tools/safety.go`: Three-tier safety validation framework
- `internal/tools/executor.go`: Parallel executor with semaphore-based concurrency
- `internal/tools/audit.go`: JSON-based audit trail with log rotation
- `internal/tools/integration.go`: Bridges tool system to AI engine

**Tool Registry Usage:**
```go
// Create registry with MCP manager
manager := mcp.NewMCPManager(config.MCP)
registry := tools.NewToolRegistry(true, manager) // preferMCP=true

// Initialize from MCP servers
if err := tools.InitializeToolRegistryFromMCP(registry, manager); err != nil {
    log.Fatalf("Failed to initialize tools: %v", err)
}

// Get tool
tool, err := registry.Get("call_aws")

// Analyze safety
call := &types.ToolCall{
    ToolName:  "delete_tool",
    Arguments: map[string]interface{}{"command": "delete users"},
    Context:   map[string]string{},
}
analysis := registry.Analyze(call)
if analysis.RequiresApproval {
    fmt.Printf("⚠️  Requires approval: %s\n", analysis.Rationale)
}

// Execute tool (with approval)
call.Context["approved"] = "true"
result, err := registry.Execute(ctx, call)
```

**Parallel Execution:**
```go
// Create executor
auditor, _ := tools.NewAuditLogger("/var/log/sdek/audit.log")
executor := tools.NewExecutor(registry, 10, 60*time.Second, auditor)

// Execute multiple tools in parallel
calls := []*types.ToolCall{
    {ToolName: "aws-api:call_aws", Arguments: map[string]interface{}{"command": "iam list-users"}},
    {ToolName: "github-mcp:search_code", Arguments: map[string]interface{}{"query": "auth"}},
    {ToolName: "jira-mcp:search_issues", Arguments: map[string]interface{}{"jql": "project=SEC"}},
}

results, err := executor.ExecuteParallel(ctx, calls)
```

**Safety Validation (Three-Tier):**
1. **Tier 1 - Interactive Commands**: Blocks vim, bash, python REPLs, etc.
2. **Tier 2 - Resource Modification**: Requires approval for delete, terminate, destroy, etc.
3. **Tier 3 - Safe Operations**: Auto-approves read-only operations (list, get, describe)

**Safety Configuration:**
```go
validator := tools.NewSafetyValidator()

// Customize rules
validator.AddDenyPattern("force-push")
validator.AddAllowPattern("list-*")
validator.SetDenyList([]string{"rm -rf", "drop table"})
```

**Audit Logging:**
- Every tool execution logged with timestamp, arguments, results, latency
- JSON-line format for easy parsing
- Concurrent-safe writes
- Automatic log rotation support

### Evidence Mapping Logic

Located in `internal/analyze/mapper.go`:

**Heuristic Mapping:**
- Keyword-based matching between event content and control keywords
- Recency scoring (newer events score higher)
- Source type weighting
- Result: 0.0-1.0 confidence score

**AI-Enhanced Mapping:**
- Hybrid approach: 70% AI confidence + 30% heuristic confidence
- Groups events by control for batch analysis
- Falls back to heuristic if AI fails
- Caches AI responses by control+events hash

**Finding Generation:**
- Risk scoring: Severity-weighted formula (3H=1C, 6M=1C, 12L=1C)
- Status determination: Green/Yellow/Red based on thresholds
- Review flagging: Confidence < 70% requires human review

## Code Patterns

### Adding a New Command

Commands use Cobra framework in `cmd/` directory:

```go
// cmd/mycommand.go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Long:  "Detailed description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Load state
        state, err := store.Load()
        if err != nil {
            return err
        }

        // Use autosave wrapper for automatic state persistence
        return store.WithAutosave(state, func() error {
            // Command logic here
            return nil
        })
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
    myCmd.Flags().StringP("flag", "f", "", "flag description")
}
```

### Testing Patterns

**Unit tests** use standard Go testing:
```go
func TestMyFunction(t *testing.T) {
    // Use testdata/ for fixtures
    got := MyFunction(input)
    if got != want {
        t.Errorf("MyFunction() = %v, want %v", got, want)
    }
}
```

**Integration tests** in `tests/integration/`:
- Use real state files and test data
- Test end-to-end workflows
- Mock AI providers via `internal/ai/providers/mock.go`
- Example: `context_mode_test.go` tests full AI analysis flow

### Working with AI Providers

**Always use the Engine interface** (`internal/ai/engine.go`):
```go
// Create engine with provider
cfg := types.DefaultConfig()
cfg.AI.Provider = "openai"
cfg.AI.OpenAIKey = os.Getenv("SDEK_AI_OPENAI_KEY")

engine, err := ai.NewEngine(cfg)
if err != nil {
    return err
}

// Analyze with context injection
preamble, _ := types.NewContextPreamble(framework, version, section, excerpt, controlIDs)
bundle := types.NewEvidenceBundle(events)

finding, err := engine.Analyze(ctx, preamble, *bundle)
if err != nil {
    // Handle error (falls back to heuristic in mapper)
}
```

**Never call providers directly** - use Engine for caching, redaction, and error handling.

### Privacy and Redaction

**PII redaction is MANDATORY** before sending to AI:
- Automatically applied by Engine
- Configured via `types.RedactionConfig`
- Default patterns: emails, phones, API keys, SSNs, credit cards
- Custom denylist via config: `ai.redaction.denylist`

Original events are **never modified** - redaction only applies to AI requests.

## Important Constraints

### AI Provider Requirements
- **API Keys**: Required via env vars or config file
  - OpenAI: `SDEK_AI_OPENAI_KEY`
  - Anthropic: `SDEK_AI_ANTHROPIC_KEY`
- **Rate Limits**: Default 10 req/min (configurable)
- **Timeouts**: Default 60s (configurable)
- **Caching**: SHA256-based, stored in `~/.cache/sdek/ai-cache/`

### State File Format
- JSON only (no other formats supported)
- Schema defined by `store.State` struct
- Version field for migration tracking
- Location: `~/.sdek/state.json` (configurable via `data_dir`)

### Testing Requirements
- Unit tests in same package as code (`*_test.go`)
- Integration tests in `tests/integration/`
- Use race detector: `go test -race`
- Coverage target: >80% (run `make test-coverage`)

### TUI Implementation
- Uses Bubble Tea framework (`ui/app.go`)
- Components in `ui/components/`
- Models in `ui/models/`
- Styles via Lip Gloss (`ui/styles/`)
- **Note**: TUI is partially implemented (structure exists, full interactivity planned)

## Common Gotchas

1. **State must be loaded before use**: Always call `store.Load()` first
2. **AI cache invalidation**: Provider switching invalidates cache (by design)
3. **Context preamble validation**: Use `types.NewContextPreamble()` for validation
4. **Config path resolution**: Use `viper` for env var substitution (`$HOME`)
5. **Log level**: Default is `info`, use `--log-level debug` or `-v` for verbose
6. **Concurrent access**: State operations are NOT thread-safe, use locks if needed
7. **AI errors**: Engine returns errors but mapper falls back to heuristic silently

## Development Principles

From `.github/copilot-instructions.md`:
- Go 1.23+ (latest stable)
- Follow standard Go conventions
- Use structured logging (slog)
- Prefer composition over inheritance
- Test all exported functions
- Document public APIs with godoc comments

## Key File Locations

- **Main entry**: `main.go` (minimal, delegates to cmd package)
- **CLI commands**: `cmd/*.go` (Cobra commands)
  - `cmd/mcp.go`, `cmd/mcp_*.go`: MCP management commands (Feature 006)
- **Core types**: `pkg/types/*.go` (public API)
  - `pkg/types/mcp.go`: MCP configuration types (Feature 006)
  - `pkg/types/provider.go`: AI provider types (Feature 006)
  - `pkg/types/tool.go`: Tool system types (Feature 006)
- **Analysis logic**: `internal/analyze/` (mapping, risk scoring)
- **AI integration**: `internal/ai/` (engine, providers, privacy)
  - `internal/ai/mcp_integration.go`: MCP engine integration (Feature 006)
- **MCP integration**: `internal/mcp/` (MCP client implementation - Feature 006)
  - `manager.go`: Multi-server orchestration
  - `client.go`: MCP protocol handshake
  - `transport.go`, `stdio_client.go`, `http_client.go`: Transport layer
  - `connector_adapter.go`: Bridge to AI engine
  - `normalizer.go`: Result normalization
- **Tool registry**: `internal/tools/` (Multi-system orchestration - Feature 006 Phase 5)
  - `registry.go`: Unified tool catalog (builtin, MCP, legacy)
  - `safety.go`: Three-tier safety validation framework
  - `executor.go`: Parallel executor with semaphore concurrency
  - `audit.go`: JSON-based audit trail with log rotation
  - `integration.go`: Bridges tool system to AI engine
- **State persistence**: `internal/store/state.go`
- **Config loading**: `internal/config/loader.go`
- **Test fixtures**: `testdata/` (JSON fixtures for tests)
- **Spec documentation**: `specs/*/` (feature specs with plans/tasks)

## Related Documentation

- [README.md](README.md): User-facing documentation and quickstart
- [config.example.yaml](config.example.yaml): Full configuration reference
- [specs/003-ai-context-injection/](specs/003-ai-context-injection/): Context injection feature spec
- [specs/003-ai-context-injection/quickstart.md](specs/003-ai-context-injection/quickstart.md): AI workflow examples
- [specs/006-mcp-pluggable-architecture/](specs/006-mcp-pluggable-architecture/): MCP integration feature spec
- [specs/006-mcp-pluggable-architecture/quickstart.md](specs/006-mcp-pluggable-architecture/quickstart.md): MCP usage guide
- [specs/006-mcp-pluggable-architecture/FINAL_IMPLEMENTATION_SUMMARY.md](specs/006-mcp-pluggable-architecture/FINAL_IMPLEMENTATION_SUMMARY.md): Implementation summary

## Recent Changes
- **2025-10-28**: Feature 006 Phase 5 Complete - Multi-System Orchestration implemented
  - Added unified tool registry combining builtin, MCP, and legacy tools
  - Implemented three-tier safety validation (interactive, modifies, safe)
  - Added parallel executor with configurable concurrency limits
  - Comprehensive audit logging with JSON-based trail
  - 10 unit tests covering all core functionality (100% pass rate)
  - Full backward compatibility maintained
- **2025-10-27**: Feature 006 Phase 4 Complete - MCP Client Mode implemented
  - Added MCP client infrastructure (JSON-RPC, transports, manager)
  - Integrated MCP with AI engine via ConnectorAdapter
  - Added CLI commands: `sdek mcp list-servers`, `list-tools`, `test`
  - Support for 7+ AI providers (OpenAI, Anthropic, Gemini, Ollama, etc.)
  - Full backward compatibility maintained

## Active Technologies
- Go 1.23+ (latest stable, matching current codebase) (006-mcp-pluggable-architecture)
- File-based (JSON state in `~/.sdek/state.json`, YAML config in `~/.sdek/config.yaml`, no database changes) (006-mcp-pluggable-architecture)
