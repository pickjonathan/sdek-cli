package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestNewMCPManager(t *testing.T) {
	config := types.MCPConfig{
		Enabled: true,
		Servers: map[string]types.MCPServerConfig{
			"test-server": {
				Transport: "stdio",
				Command:   "echo",
				Args:      []string{"test"},
				Timeout:   60,
			},
		},
	}

	manager := NewMCPManager(config)
	if manager == nil {
		t.Fatal("NewMCPManager() returned nil")
	}

	if len(manager.servers) != 0 {
		t.Error("Manager should start with no initialized servers")
	}
}

func TestMCPManagerInitializeDisabled(t *testing.T) {
	config := types.MCPConfig{
		Enabled: false,
	}

	manager := NewMCPManager(config)
	ctx := context.Background()

	err := manager.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize() should not error when disabled, got: %v", err)
	}
}

func TestMCPManagerClose(t *testing.T) {
	config := types.MCPConfig{
		Enabled: true,
		Servers: map[string]types.MCPServerConfig{},
	}

	manager := NewMCPManager(config)
	ctx := context.Background()

	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Close should be safe even with no servers
	manager.Close()

	// Double-close should be safe
	manager.Close()
}

func TestMCPManagerListServers(t *testing.T) {
	config := types.MCPConfig{
		Enabled: true,
		Servers: map[string]types.MCPServerConfig{
			"server1": {
				Transport: "stdio",
				Command:   "cat",
				Timeout:   60,
			},
			"server2": {
				Transport: "http",
				URL:       "http://localhost:8080",
				Timeout:   60,
			},
		},
	}

	manager := NewMCPManager(config)
	servers := manager.ListServers()

	// Should list configured servers even before initialization
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}

	// Verify server names
	serverNames := make(map[string]bool)
	for _, server := range servers {
		serverNames[server.Name] = true
	}

	if !serverNames["server1"] || !serverNames["server2"] {
		t.Error("Expected servers 'server1' and 'server2'")
	}
}

func TestMCPManagerIsRetryable(t *testing.T) {
	config := types.MCPConfig{
		Enabled: true,
	}

	manager := NewMCPManager(config)

	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{
			name:      "timeout error",
			err:       context.DeadlineExceeded,
			wantRetry: true,
		},
		{
			name:      "canceled error",
			err:       context.Canceled,
			wantRetry: false,
		},
		{
			name:      "transport error",
			err:       ErrTransportFailed,
			wantRetry: true,
		},
		{
			name:      "connection closed",
			err:       ErrConnectionClosed,
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manager.isRetryable(tt.err)
			if got != tt.wantRetry {
				t.Errorf("isRetryable(%v) = %v, want %v", tt.err, got, tt.wantRetry)
			}
		})
	}
}

func TestMCPManagerExecuteToolInvalidServer(t *testing.T) {
	config := types.MCPConfig{
		Enabled: true,
		Servers: map[string]types.MCPServerConfig{},
	}

	manager := NewMCPManager(config)
	ctx := context.Background()

	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer manager.Close()

	// Try to execute tool on non-existent server
	_, err := manager.ExecuteTool(ctx, "nonexistent", "test_tool", nil)
	if err == nil {
		t.Error("Expected error executing tool on non-existent server")
	}
}

func TestMCPManagerGracefulFailure(t *testing.T) {
	// Test that one failing server doesn't crash the manager
	config := types.MCPConfig{
		Enabled: true,
		Servers: map[string]types.MCPServerConfig{
			"bad-server": {
				Transport: "stdio",
				Command:   "/nonexistent/command",
				Timeout:   1,
			},
		},
	}

	manager := NewMCPManager(config)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Initialize should complete even with failing server
	_ = manager.Initialize(ctx)

	// Manager should still be usable
	servers := manager.ListServers()
	if len(servers) == 0 {
		t.Error("Manager should still track configured servers")
	}

	manager.Close()
}

func TestMCPManagerDiscoverToolsEmpty(t *testing.T) {
	config := types.MCPConfig{
		Enabled: true,
		Servers: map[string]types.MCPServerConfig{},
	}

	manager := NewMCPManager(config)
	ctx := context.Background()

	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	defer manager.Close()

	tools := manager.DiscoverTools()
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools from empty manager, got %d", len(tools))
	}
}
