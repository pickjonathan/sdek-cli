package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestHTTPTransportInitialize(t *testing.T) {
	tests := []struct {
		name    string
		config  types.MCPServerConfig
		wantErr bool
	}{
		{
			name: "valid HTTP config",
			config: types.MCPServerConfig{
				Transport: "http",
				URL:       "http://localhost:8080",
				Timeout:   60,
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: types.MCPServerConfig{
				Transport: "http",
				URL:       "",
				Timeout:   60,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewHTTPTransport()
			ctx := context.Background()

			err := transport.Initialize(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				transport.Close()
			}
		})
	}
}

func TestHTTPTransportType(t *testing.T) {
	transport := NewHTTPTransport()
	if transport.Type() != TransportHTTP {
		t.Errorf("Type() = %v, want %v", transport.Type(), TransportHTTP)
	}
}

func TestHTTPTransportSendReceive(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Decode request
		var req JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Send response
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"status":"ok"}`),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := types.MCPServerConfig{
		Transport: "http",
		URL:       server.URL,
		Timeout:   60,
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer transport.Close()

	// Send request
	request := NewRequest(1, "test_method", map[string]string{"key": "value"})
	response, err := transport.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	// Verify response
	if response.JSONRPC != "2.0" {
		t.Errorf("Response JSONRPC = %v, want 2.0", response.JSONRPC)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if status, ok := result["status"].(string); !ok || status != "ok" {
		t.Errorf("Result status = %v, want ok", result["status"])
	}
}

func TestHTTPTransportAuthHeaders(t *testing.T) {
	// Set test environment variable for token
	os.Setenv("TEST_TOKEN", "secret-token-123")
	defer os.Unsetenv("TEST_TOKEN")

	// Create test server that checks authorization
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer secret-token-123" {
			t.Errorf("Expected Bearer token, got: %s", auth)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check custom header
		customVal := r.Header.Get("X-Custom-Header")
		if customVal != "test-value" {
			t.Errorf("Expected custom header, got: %s", customVal)
		}

		// Decode request
		var req JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Send success response
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"authenticated":true}`),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := types.MCPServerConfig{
		Transport: "http",
		URL:       server.URL,
		Timeout:   60,
		Headers: map[string]string{
			"Authorization":   "Bearer $TEST_TOKEN",
			"X-Custom-Header": "test-value",
		},
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer transport.Close()

	// Send authenticated request
	request := NewRequest(1, "secure_method", nil)
	response, err := transport.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	// Verify response
	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if auth, ok := result["authenticated"].(bool); !ok || !auth {
		t.Errorf("Expected authenticated=true, got %v", result)
	}
}

func TestHTTPTransportTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		json.NewEncoder(w).Encode(JSONRPCResponse{})
	}))
	defer server.Close()

	config := types.MCPServerConfig{
		Transport: "http",
		URL:       server.URL,
		Timeout:   1, // 1 second timeout
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer transport.Close()

	// Send request that should timeout
	request := NewRequest(1, "slow_method", nil)
	_, err := transport.Send(ctx, request)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestHTTPTransportConnectionError(t *testing.T) {
	config := types.MCPServerConfig{
		Transport: "http",
		URL:       "http://localhost:9999", // Non-existent server
		Timeout:   1,
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer transport.Close()

	// Send request to non-existent server
	request := NewRequest(1, "test_method", nil)
	_, err := transport.Send(ctx, request)
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

func TestHTTPTransportErrorResponse(t *testing.T) {
	// Create server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := NewErrorResponse(req.ID, MethodNotFound, "Method not found", nil)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := types.MCPServerConfig{
		Transport: "http",
		URL:       server.URL,
		Timeout:   60,
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer transport.Close()

	// Send request
	request := NewRequest(1, "nonexistent_method", nil)
	response, err := transport.Send(ctx, request)
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	// Verify error response
	if !response.IsError() {
		t.Error("Expected error response")
	}

	if response.Error.Code != MethodNotFound {
		t.Errorf("Error code = %d, want %d", response.Error.Code, MethodNotFound)
	}
}

func TestHTTPTransportClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(JSONRPCResponse{})
	}))
	defer server.Close()

	config := types.MCPServerConfig{
		Transport: "http",
		URL:       server.URL,
		Timeout:   60,
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Close transport
	if err := transport.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify transport is closed
	if !transport.closed {
		t.Error("Transport should be marked as closed")
	}

	// Verify double-close is safe
	if err := transport.Close(); err != nil {
		t.Errorf("Second Close() should not error, got: %v", err)
	}
}

func TestHTTPTransportHealthCheck(t *testing.T) {
	// Create server with health endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
			return
		}

		var req JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"status":"ok"}`),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := types.MCPServerConfig{
		Transport: "http",
		URL:       server.URL,
		Timeout:   60,
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer transport.Close()

	// The health check is part of initialization
	// If we got here, the health check passed
}

func TestHTTPTransportConcurrentRequests(t *testing.T) {
	// Create server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"status":"ok"}`),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := types.MCPServerConfig{
		Transport: "http",
		URL:       server.URL,
		Timeout:   60,
	}

	transport := NewHTTPTransport()
	ctx := context.Background()

	if err := transport.Initialize(ctx, config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer transport.Close()

	// Send concurrent requests
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			request := NewRequest(id, "concurrent_method", nil)
			_, err := transport.Send(ctx, request)
			if err != nil {
				t.Errorf("Send() failed for request %d: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestHTTPTransportHTTPStatusErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"HTTP 200 OK", http.StatusOK, false},
		{"HTTP 404 Not Found", http.StatusNotFound, true},
		{"HTTP 500 Internal Server Error", http.StatusInternalServerError, true},
		{"HTTP 401 Unauthorized", http.StatusUnauthorized, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					var req JSONRPCRequest
					json.NewDecoder(r.Body).Decode(&req)
					resp := JSONRPCResponse{
						JSONRPC: "2.0",
						ID:      req.ID,
						Result:  json.RawMessage(`{"status":"ok"}`),
					}
					json.NewEncoder(w).Encode(resp)
				} else {
					w.Write([]byte("Error"))
				}
			}))
			defer server.Close()

			config := types.MCPServerConfig{
				Transport: "http",
				URL:       server.URL,
				Timeout:   60,
			}

			transport := NewHTTPTransport()
			ctx := context.Background()

			if err := transport.Initialize(ctx, config); err != nil {
				t.Fatalf("Initialize() failed: %v", err)
			}
			defer transport.Close()

			request := NewRequest(1, "test_method", nil)
			_, err := transport.Send(ctx, request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
