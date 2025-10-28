package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// StdioTransport implements MCP communication via stdio
type StdioTransport struct {
	config  types.MCPServerConfig
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	encoder *json.Encoder
	decoder *json.Decoder
	mu      sync.Mutex
	closed  bool
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{}
}

// Initialize establishes the subprocess connection
func (t *StdioTransport) Initialize(ctx context.Context, config types.MCPServerConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd != nil {
		return fmt.Errorf("transport already initialized")
	}

	// Validate config
	if err := ValidateTransportConfig(config); err != nil {
		return fmt.Errorf("invalid transport config: %w", err)
	}

	t.config = config

	// Create the command
	t.cmd = exec.CommandContext(ctx, config.Command, config.Args...)

	// Set environment variables
	t.cmd.Env = os.Environ()
	for key, value := range config.Env {
		// Expand environment variables in values
		expandedValue := os.ExpandEnv(value)
		t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("%s=%s", key, expandedValue))
	}

	// Create pipes for stdin, stdout, stderr
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the subprocess
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start subprocess: %w", err)
	}

	// Create JSON encoder and decoder
	t.encoder = json.NewEncoder(t.stdin)
	t.decoder = json.NewDecoder(t.stdout)

	// Start stderr logger in background
	go t.logStderr()

	return nil
}

// Send sends a JSON-RPC request and waits for response
func (t *StdioTransport) Send(ctx context.Context, request *JSONRPCRequest) (*JSONRPCResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, ErrConnectionClosed
	}

	if t.cmd == nil {
		return nil, fmt.Errorf("transport not initialized")
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Send request
	if err := t.encoder.Encode(request); err != nil {
		return nil, fmt.Errorf("%w: failed to send request: %v", ErrTransportFailed, err)
	}

	// Check if this is a notification (no ID means no response expected)
	if request.ID == nil {
		// Notifications don't get responses, return immediately
		return &JSONRPCResponse{
			JSONRPC: "2.0",
		}, nil
	}

	// Read response for requests (not notifications)
	var response JSONRPCResponse
	if err := t.decoder.Decode(&response); err != nil {
		if err == io.EOF {
			return nil, ErrConnectionClosed
		}
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrTransportFailed, err)
	}

	// Validate response
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &response, nil
}

// Close closes the transport and terminates the subprocess
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	var errs []error

	// Send shutdown notification if possible
	if t.encoder != nil && t.cmd != nil {
		shutdownReq := NewNotification("shutdown", nil)
		_ = t.encoder.Encode(shutdownReq)  // Best effort
	}

	// Close stdin
	if t.stdin != nil {
		if err := t.stdin.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close stdin: %w", err))
		}
	}

	// Close stdout
	if t.stdout != nil {
		if err := t.stdout.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close stdout: %w", err))
		}
	}

	// Close stderr
	if t.stderr != nil {
		if err := t.stderr.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close stderr: %w", err))
		}
	}

	// Kill the process
	if t.cmd != nil && t.cmd.Process != nil {
		if err := t.cmd.Process.Kill(); err != nil {
			errs = append(errs, fmt.Errorf("failed to kill process: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// Type returns the transport type
func (t *StdioTransport) Type() TransportType {
	return TransportStdio
}

// logStderr reads and logs stderr output from the subprocess
func (t *StdioTransport) logStderr() {
	if t.stderr == nil {
		return
	}

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: Use structured logging (slog) when integrated
		fmt.Fprintf(os.Stderr, "[MCP %s stderr] %s\n", t.config.Command, line)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "[MCP %s stderr] scanner error: %v\n", t.config.Command, err)
	}
}
