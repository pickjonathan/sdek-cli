# MCP Registry Interface Contract

**Feature**: 004-mcp-native-agent  
**Date**: 2025-10-19  
**Status**: Contract Definition (Implementation Pending)

---

## Overview

The `MCPRegistry` interface provides the central orchestration point for managing MCP tool connections. It handles discovery, loading, initialization, health monitoring, and administrative operations for MCP tools.

---

## Interface Definition

```go
package mcp

import (
	"context"
	"time"
	
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// MCPRegistry manages the lifecycle of MCP tool connections.
type MCPRegistry interface {
	// Lifecycle Management
	
	// Init discovers and initializes all MCP tools from configured paths.
	// Returns count of successfully initialized tools and any critical errors.
	Init(ctx context.Context) (int, error)
	
	// Close gracefully shuts down all tool connections.
	// Waits up to timeout for in-flight invocations to complete.
	Close(ctx context.Context, timeout time.Duration) error
	
	// Reload re-scans config directories and hot-reloads changed tools.
	// Returns count of reloaded tools.
	Reload(ctx context.Context) (int, error)
	
	// Query Operations
	
	// List returns all discovered tools with their current status.
	List(ctx context.Context) ([]types.MCPTool, error)
	
	// Get retrieves a specific tool by name.
	// Returns ErrToolNotFound if tool doesn't exist.
	Get(ctx context.Context, name string) (types.MCPTool, error)
	
	// Administrative Operations
	
	// Enable marks a tool as administratively enabled.
	// Tool will transition from offline → ready if healthy.
	Enable(ctx context.Context, name string) error
	
	// Disable marks a tool as administratively disabled.
	// Tool will transition to offline and reject invocations.
	Disable(ctx context.Context, name string) error
	
	// Validation & Testing
	
	// Validate validates one or more config files against the schema.
	// Returns detailed schema errors with file/line/property paths.
	Validate(ctx context.Context, paths ...string) ([]types.SchemaError, error)
	
	// Test performs a health check and handshake on a tool.
	// Returns diagnostic information including latency and capabilities.
	Test(ctx context.Context, name string) (types.MCPHealthReport, error)
}
```

---

## Method Contracts

### Init(ctx context.Context) (int, error)

**Purpose**: Discover and initialize all MCP tools from configured paths.

**Behavior**:
1. Scan config directories in precedence order:
   - Project: `./.sdek/mcp/*.json`
   - Global: `~/.sdek/mcp/*.json`
   - Env: `$SDEK_MCP_PATH` (colon-separated)
2. For each config file:
   - Validate against schema
   - If valid: Load config → Create transport → Perform handshake (async)
   - If invalid: Log error, skip
3. Start health monitor background goroutine
4. Start file watcher for hot-reload (if enabled)

**Returns**:
- Count of successfully initialized tools
- Critical error (e.g., no write permissions, malformed SDEK_MCP_PATH) or nil

**Errors**:
- `ErrNoConfigDirs`: No config directories found or accessible
- `ErrSchemaLoadFailed`: Unable to load MCP config schema

**Acceptance Criteria** (AC-01):
- Given valid configs in `~/.sdek/mcp/`, when Init() called, all tools appear in List() with status ready or degraded

---

### Close(ctx context.Context, timeout time.Duration) error

**Purpose**: Gracefully shut down all tool connections.

**Behavior**:
1. Stop file watcher (if running)
2. Stop health monitor
3. For each tool:
   - Wait for in-flight invocations (up to timeout)
   - Close transport connection
   - Flush metrics
4. Persist tool enable/disable state to disk

**Returns**:
- nil on success
- Error if timeout exceeded or close fails

**Errors**:
- `ErrCloseTimeout`: Timeout exceeded while waiting for in-flight invocations

---

### Reload(ctx context.Context) (int, error)

**Purpose**: Hot-reload changed config files without restarting.

**Behavior**:
1. Re-scan config directories
2. For each config:
   - If new: Load and initialize
   - If changed: Validate → Load → Re-initialize → Swap old connection
   - If deleted: Gracefully close and remove from registry
3. Log all changes (mcp_tool_loaded, mcp_tool_removed)

**Returns**:
- Count of reloaded tools
- Error if reload fails critically

**Acceptance Criteria** (AC-03):
- Given a running registry, when a config file changes, tool is reloaded without restart

---

### List(ctx context.Context) ([]types.MCPTool, error)

**Purpose**: Return all discovered tools with current status.

**Behavior**:
- Query internal tool map
- Sort by name (alphabetical)
- Return snapshot (defensive copy)

**Returns**:
- Array of MCPTool entities
- Error if registry not initialized

**Acceptance Criteria** (AC-01, AC-06):
- CLI `sdek mcp list` displays all tools with status, latency, errors

---

### Get(ctx context.Context, name string) (types.MCPTool, error)

**Purpose**: Retrieve a specific tool by name.

**Behavior**:
- Lookup tool in registry by name
- Return copy (not pointer to internal state)

**Returns**:
- MCPTool entity
- `ErrToolNotFound` if tool doesn't exist

**Errors**:
- `ErrToolNotFound`: Tool with given name not in registry

---

### Enable(ctx context.Context, name string) error

**Purpose**: Administratively enable a tool.

**Behavior**:
1. Lookup tool
2. Set `Enabled = true`
3. If tool is offline due to admin disable, transition to ready (perform handshake)
4. Persist state to disk
5. Emit `mcp_tool_enabled` event

**Returns**:
- nil on success
- Error if tool not found or enable fails

**Errors**:
- `ErrToolNotFound`: Tool not in registry

**Acceptance Criteria** (AC-06):
- CLI `sdek mcp enable <tool>` enables tool and persists across restarts

---

### Disable(ctx context.Context, name string) error

**Purpose**: Administratively disable a tool.

**Behavior**:
1. Lookup tool
2. Set `Enabled = false`
3. Transition tool to offline
4. Reject future invocations with `ErrToolDisabled`
5. Persist state to disk
6. Emit `mcp_tool_disabled` event

**Returns**:
- nil on success
- Error if tool not found

**Errors**:
- `ErrToolNotFound`: Tool not in registry

**Acceptance Criteria** (AC-06):
- CLI `sdek mcp disable <tool>` disables tool and persists across restarts

---

### Validate(ctx context.Context, paths ...string) ([]types.SchemaError, error)

**Purpose**: Validate config files against MCP schema.

**Behavior**:
1. For each path:
   - Parse JSON (track line/column positions)
   - Validate against schema (`contracts/mcp-config-schema.json`)
   - Map schema errors to file positions
2. Return all errors (don't stop at first error)

**Returns**:
- Array of SchemaError (file, line, column, JSON path, message)
- Parsing error if JSON is malformed

**Errors**:
- `ErrFileNotFound`: Config file doesn't exist
- `ErrInvalidJSON`: JSON parsing failed

**Acceptance Criteria** (AC-02):
- Given invalid config, when validated, returns detailed error with file/line/property

---

### Test(ctx context.Context, name string) (types.MCPHealthReport, error)

**Purpose**: Perform health check and handshake on a tool.

**Behavior**:
1. Lookup tool
2. Execute handshake (send ping, await pong)
3. Measure latency
4. Verify capabilities match config
5. Return diagnostic report

**Returns**:
- MCPHealthReport with status, latency, capabilities
- Error if tool not found or test fails

**Errors**:
- `ErrToolNotFound`: Tool not in registry
- `ErrHandshakeFailed`: Handshake timed out or failed

**Acceptance Criteria** (AC-06):
- CLI `sdek mcp test <tool>` returns round-trip diagnostics

---

## Error Types

```go
var (
	ErrToolNotFound     = errors.New("mcp: tool not found in registry")
	ErrToolDisabled     = errors.New("mcp: tool is administratively disabled")
	ErrNoConfigDirs     = errors.New("mcp: no config directories found or accessible")
	ErrSchemaLoadFailed = errors.New("mcp: failed to load MCP config schema")
	ErrCloseTimeout     = errors.New("mcp: timeout exceeded during close")
	ErrFileNotFound     = errors.New("mcp: config file not found")
	ErrInvalidJSON      = errors.New("mcp: invalid JSON in config file")
	ErrHandshakeFailed  = errors.New("mcp: handshake with tool failed")
)
```

---

## Implementation Notes

### Concurrency
- Registry is thread-safe (use `sync.RWMutex` for tool map)
- Init() and Reload() acquire write lock
- List() and Get() acquire read lock
- Health checks run in background goroutines (don't block)

### Configuration Precedence
When the same tool is defined in multiple locations:
1. Project (`./.sdek/mcp/`) overrides
2. Global (`~/.sdek/mcp/`) overrides
3. Env (`$SDEK_MCP_PATH`)

Log warning when override occurs.

### Persistence
Tool enable/disable state persisted to `~/.sdek/state/mcp-tools.json`:
```json
{
  "github": {"enabled": true},
  "jira": {"enabled": false}
}
```

Loaded at Init(), updated on Enable/Disable.

---

## Testing Strategy

### Unit Tests
- `TestRegistryInit`: Valid/invalid configs, empty dirs, precedence
- `TestRegistryList`: Sorting, filtering by status
- `TestRegistryGet`: Found/not found
- `TestRegistryEnableDisable`: State transitions, persistence
- `TestRegistryValidate`: Schema errors with file/line

### Integration Tests
- `TestRegistryHotReload`: File watcher detects changes, tools reloaded
- `TestRegistryHandshake`: Mock MCP server, verify handshake protocol

---

## Dependencies

- `internal/config`: Viper for feature flags (`mcp.enabled`, `mcp.hotReload`)
- `internal/mcp/loader`: Config discovery and loading
- `internal/mcp/validator`: JSON Schema validation
- `internal/mcp/health`: Health check and handshake logic
- `internal/mcp/transport`: Transport interface (stdio, HTTP)
- `pkg/types`: MCPTool, MCPConfig, SchemaError, MCPHealthReport

---

**Contract Status**: ✅ Defined, awaiting implementation
