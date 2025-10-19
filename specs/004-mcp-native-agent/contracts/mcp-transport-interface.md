# MCP Transport Interface Contract

**Feature**: 004-mcp-native-agent  
**Date**: 2025-10-19  
**Status**: Contract Definition (Implementation Pending)

---

## Overview

The `MCPTransport` interface abstracts the communication protocol between sdek-cli and MCP tool servers. It supports two transport types: **stdio** (local processes) and **HTTP** (remote servers).

---

## Interface Definition

```go
package transport

import (
	"context"
	"time"
)

// MCPTransport handles communication with an MCP tool server.
type MCPTransport interface {
	// Invoke calls a method on the MCP tool server.
	// Uses JSON-RPC 2.0 protocol (request/response).
	Invoke(ctx context.Context, method string, params any) (any, error)
	
	// HealthCheck performs a lightweight health check (ping/pong).
	// Returns nil if healthy, error otherwise.
	HealthCheck(ctx context.Context) error
	
	// Close gracefully shuts down the transport connection.
	// For stdio: terminates child process.
	// For HTTP: closes connection pool.
	Close() error
	
	// Metadata returns transport-specific information for diagnostics.
	Metadata() TransportMetadata
}

// TransportMetadata provides diagnostic information about the transport.
type TransportMetadata struct {
	Type            string        // "stdio" or "http"
	ProtocolVersion string        // e.g., "1.0.0"
	Latency         time.Duration // Average latency (last 100 calls)
	IsConnected     bool          // Connection status
}
```

---

## Implementations

### 1. Stdio Transport

**File**: `internal/mcp/transport/stdio.go`

**Behavior**:
- Spawns child process via `exec.Command(config.Command, config.Args...)`
- Communicates via stdin/stdout using JSON-RPC 2.0
- Long-lived process (pooled, not spawned per invocation)
- Monitors stderr for diagnostics

**JSON-RPC 2.0 Protocol**:
```json
// Request
{
  "jsonrpc": "2.0",
  "id": "req-uuid",
  "method": "commits.list",
  "params": {"repo": "owner/repo", "since": "2025-01-01"}
}

// Response
{
  "jsonrpc": "2.0",
  "id": "req-uuid",
  "result": [{"sha": "abc123", "message": "fix bug"}]
}

// Error Response
{
  "jsonrpc": "2.0",
  "id": "req-uuid",
  "error": {"code": -32600, "message": "Invalid params"}
}
```

**Implementation Notes**:
- Use `bufio.Scanner` for line-delimited JSON
- Handle process crashes (restart with backoff)
- Timeout on unresponsive processes (kill -9 after grace period)

**Acceptance Criteria**:
- Given valid stdio config, transport spawns process and completes handshake
- Given unresponsive process, transport returns timeout error and kills process

---

### 2. HTTP Transport

**File**: `internal/mcp/transport/http.go`

**Behavior**:
- Sends JSON-RPC 2.0 requests to `config.BaseURL + "/rpc"` via HTTP POST
- Uses `http.Client` with connection pooling
- Stateless (server maintains state, not client)

**HTTP Request**:
```http
POST /rpc HTTP/1.1
Host: mcp-server.example.com
Content-Type: application/json
Authorization: Bearer ${TOKEN}

{"jsonrpc": "2.0", "id": "req-uuid", "method": "search", "params": {...}}
```

**HTTP Response**:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{"jsonrpc": "2.0", "id": "req-uuid", "result": [...]}
```

**Implementation Notes**:
- Respect HTTP timeouts (`config.Timeout`)
- Handle retries on 5xx errors (exponential backoff)
- Support `Authorization` header for authenticated servers

**Acceptance Criteria**:
- Given valid HTTP config, transport sends request and receives response
- Given 503 error, transport retries with backoff

---

## Error Handling

```go
var (
	ErrTransportClosed       = errors.New("transport: connection closed")
	ErrInvokeTimeout         = errors.New("transport: invocation timeout")
	ErrProcessCrashed        = errors.New("transport: stdio process crashed")
	ErrInvalidJSONRPC        = errors.New("transport: invalid JSON-RPC response")
	ErrHTTPError             = errors.New("transport: HTTP request failed")
)
```

---

## Testing Strategy

### Unit Tests
- `TestStdioTransport`: Spawn mock process, send request, verify response
- `TestHTTPTransport`: Mock HTTP server, send request, verify response
- `TestStdioProcessCrash`: Verify error handling and restart logic
- `TestHTTPRetry`: Verify retry on 5xx errors

### Integration Tests
- `TestStdioHandshake`: Real MCP server (e.g., mock-mcp-server), verify handshake
- `TestHTTPHandshake`: Real HTTP MCP server, verify handshake

---

## Dependencies

- `os/exec`: Spawn child processes (stdio)
- `net/http`: HTTP client (HTTP transport)
- `encoding/json`: JSON-RPC 2.0 encoding/decoding
- `bufio`: Line-delimited JSON reading (stdio)

---

**Contract Status**: ✅ Defined, awaiting implementation
