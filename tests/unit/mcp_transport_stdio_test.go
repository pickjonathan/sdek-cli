package unit

import (
	"context"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/mcp/transport"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestStdioTransportJSONRPC(t *testing.T) {
	config := &types.MCPConfig{
		Name:    "test-stdio",
		Command: "/bin/echo",
		Args:    []string{`{"jsonrpc":"2.0","id":"1","result":{"status":"ok"}}`},
		Transport: "stdio",
	}

	trans, err := transport.NewStdioTransport(config)
	if err != nil {
		t.Fatalf("failed to create stdio transport: %v", err)
	}
	defer trans.Close()

	ctx := context.Background()
	result, err := trans.Invoke(ctx, "test.method", map[string]interface{}{"key": "value"})

	if err != nil {
		t.Errorf("expected successful invocation, got error: %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestStdioTransportHandshake(t *testing.T) {
	// This test would use a mock MCP server
	t.Skip("requires mock MCP server implementation")
}

func TestStdioTransportHandlesCrashedProcess(t *testing.T) {
	config := &types.MCPConfig{
		Name:      "test-crash",
		Command:   "/bin/false", // exits immediately
		Transport: "stdio",
	}

	trans, err := transport.NewStdioTransport(config)
	if err != nil {
		t.Fatalf("failed to create stdio transport: %v", err)
	}
	defer trans.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = trans.Invoke(ctx, "test.method", nil)
	if err == nil {
		t.Error("expected error when process crashes")
	}
}
