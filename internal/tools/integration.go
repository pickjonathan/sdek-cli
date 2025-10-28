package tools

import (
	"context"
	"fmt"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// MCPConnectorAdapter adapts the tool registry to work as an MCPConnector
// for backward compatibility with the existing AI Engine interface.
type MCPConnectorAdapter struct {
	registry *ToolRegistry
	manager  *mcp.MCPManager
}

// NewMCPConnectorAdapter creates a new adapter that bridges the tool registry
// to the legacy MCPConnector interface.
func NewMCPConnectorAdapter(registry *ToolRegistry, manager *mcp.MCPManager) *MCPConnectorAdapter {
	return &MCPConnectorAdapter{
		registry: registry,
		manager:  manager,
	}
}

// Collect implements the MCPConnector interface by routing to the tool registry.
// The source parameter is expected to be in "server:tool" format for MCP tools.
func (a *MCPConnectorAdapter) Collect(ctx context.Context, source string, query string) ([]types.EvidenceEvent, error) {
	// Parse source as "server:tool" format
	serverName, toolName := a.parseSource(source)

	// Create tool call
	call := &types.ToolCall{
		ToolName: toolName,
		Arguments: map[string]interface{}{
			"query":   query,
			"command": query, // Support both for compatibility
		},
		Context: map[string]string{
			"source": source,
		},
	}

	// Execute via manager directly (bypassing safety validation for MCP tools)
	result, err := a.manager.ExecuteTool(ctx, serverName, toolName, call.Arguments)
	if err != nil {
		return nil, fmt.Errorf("MCP tool execution failed: %w", err)
	}

	// Normalize result to evidence events
	events, err := mcp.NormalizeToEvidenceEvent(serverName, toolName, result)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize MCP result: %w", err)
	}

	return events, nil
}

// parseSource splits a source string into server and tool names.
// Format: "server:tool" or just "server" (uses default tool).
func (a *MCPConnectorAdapter) parseSource(source string) (string, string) {
	// Check if source contains ":"
	for i, c := range source {
		if c == ':' {
			return source[:i], source[i+1:]
		}
	}

	// Default tool name if no ":" found
	return source, "collect"
}

// EngineWithToolRegistry extends the AI Engine with tool registry access.
type EngineWithToolRegistry interface {
	ai.Engine

	// GetToolRegistry returns the underlying tool registry
	GetToolRegistry() *ToolRegistry

	// GetExecutor returns the parallel executor
	GetExecutor() *Executor
}

// engineWithToolsImpl wraps an existing engine and adds tool registry support.
type engineWithToolsImpl struct {
	ai.Engine
	registry *ToolRegistry
	executor *Executor
}

// WrapEngineWithTools wraps an existing engine with tool registry support.
func WrapEngineWithTools(engine ai.Engine, registry *ToolRegistry, executor *Executor) EngineWithToolRegistry {
	return &engineWithToolsImpl{
		Engine:   engine,
		registry: registry,
		executor: executor,
	}
}

// GetToolRegistry returns the underlying tool registry.
func (e *engineWithToolsImpl) GetToolRegistry() *ToolRegistry {
	return e.registry
}

// GetExecutor returns the parallel executor.
func (e *engineWithToolsImpl) GetExecutor() *Executor {
	return e.executor
}

// InitializeToolRegistryFromMCP initializes the tool registry with tools
// discovered from MCP servers.
func InitializeToolRegistryFromMCP(registry *ToolRegistry, manager *mcp.MCPManager) error {
	// Get list of servers from manager
	servers := manager.ListServers()

	// Discover tools from each server
	for _, serverInfo := range servers {
		tools := serverInfo.Tools

		for i := range tools {
			// MCP tools are already in types.Tool format
			// Just ensure they have the correct source and server name
			tools[i].Source = types.ToolSourceMCP
			tools[i].ServerName = serverInfo.Name
			if tools[i].SafetyTier == "" {
				tools[i].SafetyTier = types.SafetyTierSafe // Default
			}

			// Register tool
			if err := registry.Register(&tools[i]); err != nil {
				return fmt.Errorf("failed to register MCP tool %s: %w", tools[i].Name, err)
			}
		}
	}

	return nil
}
