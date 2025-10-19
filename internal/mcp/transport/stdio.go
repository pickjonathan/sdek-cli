package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// StdioTransport implements Transport for stdio-based MCP servers.
type StdioTransport struct {
	config     *types.MCPConfig
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	scanner    *bufio.Scanner
	mu         sync.Mutex
	requestID  int
	latencies  []time.Duration
	maxLatency int
}

// NewStdioTransport creates a new stdio transport.
func NewStdioTransport(config *types.MCPConfig) (*StdioTransport, error) {
	cmd := exec.Command(config.Command, config.Args...)

	// Set environment variables
	if config.Env != nil {
		env := cmd.Env
		for key, value := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	return &StdioTransport{
		config:     config,
		cmd:        cmd,
		stdin:      stdin,
		stdout:     stdout,
		scanner:    bufio.NewScanner(stdout),
		latencies:  make([]time.Duration, 0, 100),
		maxLatency: 100,
	}, nil
}

// Invoke sends a JSON-RPC 2.0 request over stdin and reads the response from stdout.
func (t *StdioTransport) Invoke(ctx context.Context, method string, params interface{}) (interface{}, error) {
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

	// Send request
	reqData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	t.mu.Lock()
	_, err = t.stdin.Write(append(reqData, '\n'))
	t.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	if !t.scanner.Scan() {
		if err := t.scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		return nil, fmt.Errorf("process closed unexpectedly")
	}

	var response struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      int         `json:"id"`
		Result  interface{} `json:"result,omitempty"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(t.scanner.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// HealthCheck performs a simple health check by invoking a ping method.
func (t *StdioTransport) HealthCheck(ctx context.Context) error {
	_, err := t.Invoke(ctx, "ping", nil)
	return err
}

// Close terminates the process and closes pipes.
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stdin != nil {
		t.stdin.Close()
	}
	if t.stdout != nil {
		t.stdout.Close()
	}
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}
	return nil
}

// Metadata returns transport metadata.
func (t *StdioTransport) Metadata() TransportMetadata {
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

	isConnected := t.cmd != nil && t.cmd.Process != nil

	return TransportMetadata{
		Type:            "stdio",
		ProtocolVersion: "1.0.0",
		Latency:         avgLatency,
		IsConnected:     isConnected,
	}
}

// recordLatency records a latency measurement (keeps last 100).
func (t *StdioTransport) recordLatency(latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.latencies = append(t.latencies, latency)
	if len(t.latencies) > t.maxLatency {
		t.latencies = t.latencies[1:]
	}
}
