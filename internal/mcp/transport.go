package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Common error types for transports
var (
	ErrTransportFailed   = errors.New("transport operation failed")
	ErrTimeout           = errors.New("transport operation timed out")
	ErrConnectionClosed  = errors.New("transport connection closed")
	ErrInvalidTransport  = errors.New("invalid transport type")
)

// TransportType represents the type of MCP transport
type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportHTTP  TransportType = "http"
)

// Transport defines the interface for MCP communication transports
type Transport interface {
	// Initialize establishes the connection and performs any necessary handshake
	Initialize(ctx context.Context, config types.MCPServerConfig) error

	// Send sends a JSON-RPC request and returns the response
	Send(ctx context.Context, request *JSONRPCRequest) (*JSONRPCResponse, error)

	// Close closes the transport and releases any resources
	Close() error

	// Type returns the transport type
	Type() TransportType
}

// CreateTransport creates a transport instance based on the config
func CreateTransport(config types.MCPServerConfig) (Transport, error) {
	transportType := TransportType(config.Transport)

	switch transportType {
	case TransportStdio:
		if config.Command == "" {
			return nil, fmt.Errorf("%w: command is required for stdio transport", ErrInvalidTransport)
		}
		return NewStdioTransport(), nil

	case TransportHTTP:
		if config.URL == "" {
			return nil, fmt.Errorf("%w: url is required for http transport", ErrInvalidTransport)
		}
		return NewHTTPTransport(), nil

	default:
		return nil, fmt.Errorf("%w: unsupported transport type: %s", ErrInvalidTransport, transportType)
	}
}

// ValidateTransportConfig validates the transport configuration
func ValidateTransportConfig(config types.MCPServerConfig) error {
	transportType := TransportType(config.Transport)

	switch transportType {
	case TransportStdio:
		if config.Command == "" {
			return fmt.Errorf("command is required for stdio transport")
		}
		if config.URL != "" {
			return fmt.Errorf("url should not be set for stdio transport")
		}

	case TransportHTTP:
		if config.URL == "" {
			return fmt.Errorf("url is required for http transport")
		}
		if config.Command != "" {
			return fmt.Errorf("command should not be set for http transport")
		}

	default:
		return fmt.Errorf("unsupported transport type: %s", transportType)
	}

	// Validate common settings
	if config.Timeout < 1 || config.Timeout > 600 {
		return fmt.Errorf("timeout must be between 1 and 600 seconds, got %d", config.Timeout)
	}

	return nil
}
