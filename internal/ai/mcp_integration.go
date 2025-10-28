package ai

import (
	"context"
	"fmt"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// NewEngineWithMCP creates a new Engine instance with MCP Manager support
// This initializes the MCP manager and connects to all configured MCP servers
func NewEngineWithMCP(ctx context.Context, cfg *types.Config, provider Provider) (Engine, error) {
	// Check if MCP is enabled
	if !cfg.MCP.Enabled {
		// Return engine without MCP connector
		return NewEngine(cfg, provider), nil
	}

	// Check if any servers are configured
	if len(cfg.MCP.Servers) == 0 {
		// Return engine without MCP connector
		return NewEngine(cfg, provider), nil
	}

	// Create MCP manager
	manager := mcp.NewMCPManager(cfg.MCP)

	// Initialize manager (connect to all servers)
	if err := manager.Initialize(ctx); err != nil {
		// Log warning but don't fail - some servers may be down
		fmt.Printf("Warning: MCP manager initialization had errors: %v\n", err)
		// Continue with whatever servers were successfully initialized
	}

	// Create connector adapter
	connector := mcp.NewConnectorAdapter(manager)

	// Create engine with MCP connector
	return NewEngineWithConnector(cfg, provider, connector), nil
}

// MCPManagerFromEngine extracts the MCP manager from an engine (if available)
// This is useful for CLI commands that need direct access to the manager
func MCPManagerFromEngine(engine Engine) (*mcp.MCPManager, bool) {
	// Type assert to engineImpl
	impl, ok := engine.(*engineImpl)
	if !ok {
		return nil, false
	}

	// Check if connector is an MCP connector adapter
	if impl.connector == nil {
		return nil, false
	}

	// Type assert connector to adapter
	adapter, ok := impl.connector.(*mcp.ConnectorAdapter)
	if !ok {
		return nil, false
	}

	// Extract manager from adapter
	return adapter.Manager(), true
}
