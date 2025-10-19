package transport

import (
	"context"
	"time"
)

// Transport handles communication with an MCP tool server.
type Transport interface {
	// Invoke calls a method on the MCP tool server using JSON-RPC 2.0.
	Invoke(ctx context.Context, method string, params interface{}) (interface{}, error)

	// HealthCheck performs a lightweight health check (ping/pong).
	HealthCheck(ctx context.Context) error

	// Close gracefully shuts down the transport connection.
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

// TransportType represents the type of MCP transport.
type TransportType string

const (
	TransportTypeStdio TransportType = "stdio"
	TransportTypeHTTP  TransportType = "http"
)
