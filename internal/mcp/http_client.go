package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// HTTPTransport implements MCP communication via HTTP
type HTTPTransport struct {
	config  types.MCPServerConfig
	client  *http.Client
	baseURL string
	headers map[string]string
	mu      sync.Mutex
	closed  bool
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport() *HTTPTransport {
	return &HTTPTransport{}
}

// Initialize establishes the HTTP connection
func (t *HTTPTransport) Initialize(ctx context.Context, config types.MCPServerConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.client != nil {
		return fmt.Errorf("transport already initialized")
	}

	// Validate config
	if err := ValidateTransportConfig(config); err != nil {
		return fmt.Errorf("invalid transport config: %w", err)
	}

	t.config = config
	t.baseURL = config.URL

	// Create HTTP client with timeout
	timeout := time.Duration(config.Timeout) * time.Second
	t.client = &http.Client{
		Timeout: timeout,
	}

	// Process headers with environment variable expansion
	t.headers = make(map[string]string)
	for key, value := range config.Headers {
		expandedValue := os.ExpandEnv(value)
		t.headers[key] = expandedValue
	}

	// Ensure Content-Type is set
	if _, exists := t.headers["Content-Type"]; !exists {
		t.headers["Content-Type"] = "application/json"
	}

	// Test connectivity if health_url is provided
	if config.HealthURL != "" {
		if err := t.checkHealth(ctx, config.HealthURL); err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}
	}

	return nil
}

// Send sends a JSON-RPC request via HTTP POST
func (t *HTTPTransport) Send(ctx context.Context, request *JSONRPCRequest) (*JSONRPCResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, ErrConnectionClosed
	}

	if t.client == nil {
		return nil, fmt.Errorf("transport not initialized")
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Marshal request to JSON
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	for key, value := range t.headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	httpResp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: HTTP request failed: %v", ErrTransportFailed, err)
	}
	defer httpResp.Body.Close()

	// Check HTTP status code
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("%w: HTTP %d: %s", ErrTransportFailed, httpResp.StatusCode, string(body))
	}

	// Read response body
	responseBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal response
	var response JSONRPCResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Validate response
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &response, nil
}

// Close closes the HTTP transport
func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	// HTTP client doesn't need explicit cleanup
	// Transport connections are managed by the http.Client
	return nil
}

// Type returns the transport type
func (t *HTTPTransport) Type() TransportType {
	return TransportHTTP
}

// checkHealth performs a health check on the server
func (t *HTTPTransport) checkHealth(ctx context.Context, healthURL string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Add authentication headers (if any)
	for key, value := range t.headers {
		if key == "Authorization" || key == "X-API-Key" {
			req.Header.Set(key, value)
		}
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
