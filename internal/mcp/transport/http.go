package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// HTTPTransport implements Transport for HTTP-based MCP servers.
type HTTPTransport struct {
	config     *types.MCPConfig
	client     *http.Client
	mu         sync.Mutex
	requestID  int
	latencies  []time.Duration
	maxLatency int
}

// NewHTTPTransport creates a new HTTP transport.
func NewHTTPTransport(config *types.MCPConfig) (*HTTPTransport, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required for HTTP transport")
	}

	timeout := 30 * time.Second
	if config.Timeout != "" {
		if parsed, err := time.ParseDuration(config.Timeout); err == nil {
			timeout = parsed
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return &HTTPTransport{
		config:     config,
		client:     client,
		latencies:  make([]time.Duration, 0, 100),
		maxLatency: 100,
	}, nil
}

// Invoke sends a JSON-RPC 2.0 request via HTTP POST.
func (t *HTTPTransport) Invoke(ctx context.Context, method string, params interface{}) (interface{}, error) {
	t.mu.Lock()
	t.requestID++
	reqID := t.requestID
	t.mu.Unlock()

	start := time.Now()
	defer func() {
		latency := time.Since(start)
		t.recordLatency(latency)
	}()

	// Build JSON-RPC 2.0 request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      reqID,
		"method":  method,
	}
	if params != nil {
		request["params"] = params
	}

	reqData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP POST request
	req, err := http.NewRequestWithContext(ctx, "POST", t.config.BaseURL, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      int         `json:"id"`
		Result  interface{} `json:"result,omitempty"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// HealthCheck performs a health check via HTTP.
func (t *HTTPTransport) HealthCheck(ctx context.Context) error {
	_, err := t.Invoke(ctx, "ping", nil)
	return err
}

// Close closes the HTTP client.
func (t *HTTPTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}

// Metadata returns transport metadata.
func (t *HTTPTransport) Metadata() TransportMetadata {
	t.mu.Lock()
	defer t.mu.Unlock()

	avgLatency := time.Duration(0)
	if len(t.latencies) > 0 {
		total := time.Duration(0)
		for _, l := range t.latencies {
			total += l
		}
		avgLatency = total / time.Duration(len(t.latencies))
	}

	return TransportMetadata{
		Type:            "http",
		ProtocolVersion: "1.0.0",
		Latency:         avgLatency,
		IsConnected:     true, // HTTP is stateless
	}
}

// recordLatency records a latency measurement (keeps last 100).
func (t *HTTPTransport) recordLatency(latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.latencies = append(t.latencies, latency)
	if len(t.latencies) > t.maxLatency {
		t.latencies = t.latencies[1:]
	}
}
