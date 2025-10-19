package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/mcp/transport"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestHTTPTransportJSONRPC(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":{"status":"ok"}}`))
	}))
	defer server.Close()

	config := &types.MCPConfig{
		Name:      "test-http",
		Transport: "http",
		BaseURL:   server.URL,
	}

	trans, err := transport.NewHTTPTransport(config)
	if err != nil {
		t.Fatalf("failed to create HTTP transport: %v", err)
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

func TestHTTPTransportHandshakeWithBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"jsonrpc":"2.0","id":"handshake","result":{"capabilities":["test"]}}`))
	}))
	defer server.Close()

	config := &types.MCPConfig{
		Name:      "test-http",
		Transport: "http",
		BaseURL:   server.URL,
	}

	trans, err := transport.NewHTTPTransport(config)
	if err != nil {
		t.Fatalf("failed to create HTTP transport: %v", err)
	}
	defer trans.Close()

	ctx := context.Background()
	err = trans.HealthCheck(ctx)
	if err != nil {
		t.Errorf("expected successful health check, got: %v", err)
	}
}

func TestHTTPTransportTimeoutHandling(t *testing.T) {
	// Server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {} // block forever
	}))
	defer server.Close()

	config := &types.MCPConfig{
		Name:      "test-timeout",
		Transport: "http",
		BaseURL:   server.URL,
		Timeout:   "100ms",
	}

	trans, err := transport.NewHTTPTransport(config)
	if err != nil {
		t.Fatalf("failed to create HTTP transport: %v", err)
	}
	defer trans.Close()

	ctx := context.Background()
	_, err = trans.Invoke(ctx, "test.method", nil)
	if err == nil {
		t.Error("expected timeout error")
	}
}
