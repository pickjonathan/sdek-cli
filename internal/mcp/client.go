package mcp

import (
	"context"
	"fmt"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// MCP protocol version (using latest stable spec)
// The MCP spec uses date-based versioning: https://spec.modelcontextprotocol.io/
const MCPVersion = "2024-11-05"

// MCPClient represents an MCP client connection to a server
type MCPClient struct {
	config       types.MCPServerConfig
	transport    Transport
	capabilities ServerCapabilities
	tools        []types.Tool
}

// ServerCapabilities represents the capabilities returned by the MCP server
type ServerCapabilities struct {
	ProtocolVersion string   `json:"protocolVersion"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
	Capabilities struct {
		Tools     bool `json:"tools,omitempty"`
		Resources bool `json:"resources,omitempty"`
		Prompts   bool `json:"prompts,omitempty"`
	} `json:"capabilities"`
}

// InitializeParams are the parameters sent in the initialize request
type InitializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
	Capabilities struct {
		Roots bool `json:"roots,omitempty"`
	} `json:"capabilities"`
}

// InitializeResult is the result of the initialize request
type InitializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Tools     map[string]interface{} `json:"tools,omitempty"`
		Resources map[string]interface{} `json:"resources,omitempty"`
		Prompts   map[string]interface{} `json:"prompts,omitempty"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

// ListToolsResult is the result of the tools/list request
type ListToolsResult struct {
	Tools []struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		InputSchema map[string]interface{} `json:"inputSchema"`
	} `json:"tools"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(config types.MCPServerConfig) (*MCPClient, error) {
	transport, err := CreateTransport(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	return &MCPClient{
		config:    config,
		transport: transport,
	}, nil
}

// Initialize performs the MCP handshake
func (c *MCPClient) Initialize(ctx context.Context) error {
	// Initialize the transport (start subprocess or connect to HTTP endpoint)
	if err := c.transport.Initialize(ctx, c.config); err != nil {
		return fmt.Errorf("failed to initialize transport: %w", err)
	}

	// Send initialize request
	params := InitializeParams{
		ProtocolVersion: MCPVersion,
	}
	params.ClientInfo.Name = "sdek-cli"
	params.ClientInfo.Version = "1.0.0"

	request := NewRequest(1, "initialize", params)
	response, err := c.transport.Send(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	// Parse initialize result
	var result InitializeResult
	if err := response.UnmarshalResult(&result); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}

	// Validate protocol version (with flexible compatibility)
	// Log a warning for version mismatch but allow connection if basic compatibility exists
	if result.ProtocolVersion != MCPVersion {
		fmt.Printf("Warning: MCP protocol version mismatch (client=%s, server=%s) - attempting connection anyway\n",
			MCPVersion, result.ProtocolVersion)
		// Only fail for truly incompatible versions (old 1.0 vs new date-based)
		if result.ProtocolVersion == "1.0" && MCPVersion != "1.0" {
			return fmt.Errorf("incompatible protocol version: client=%s, server=%s", MCPVersion, result.ProtocolVersion)
		}
		if result.ProtocolVersion != "1.0" && MCPVersion == "1.0" {
			return fmt.Errorf("incompatible protocol version: client=%s, server=%s", MCPVersion, result.ProtocolVersion)
		}
	}

	// Send initialized notification
	initializedNotif := NewNotification("notifications/initialized", nil)
	if _, err := c.transport.Send(ctx, initializedNotif); err != nil {
		// Non-fatal: some servers may not require this
		fmt.Printf("Warning: failed to send initialized notification: %v\n", err)
	}

	// Discover tools
	if err := c.discoverTools(ctx); err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	return nil
}

// discoverTools requests the list of available tools from the server
func (c *MCPClient) discoverTools(ctx context.Context) error {
	request := NewRequest(2, "tools/list", nil)
	response, err := c.transport.Send(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to send tools/list request: %w", err)
	}

	var result ListToolsResult
	if err := response.UnmarshalResult(&result); err != nil {
		return fmt.Errorf("failed to unmarshal tools/list result: %w", err)
	}

	// Convert MCP tools to sdek-cli tool types
	c.tools = make([]types.Tool, 0, len(result.Tools))
	for _, mcpTool := range result.Tools {
		tool := types.Tool{
			Name:        mcpTool.Name,
			Description: mcpTool.Description,
			Parameters:  mcpTool.InputSchema,
			Source:      "mcp",
			ServerName:  c.config.Command, // Use command as server identifier for now
		}
		c.tools = append(c.tools, tool)
	}

	return nil
}

// ListTools returns the discovered tools
func (c *MCPClient) ListTools() []types.Tool {
	return c.tools
}

// CallTool executes a tool on the MCP server
func (c *MCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (interface{}, error) {
	// Find the tool
	var found bool
	for _, tool := range c.tools {
		if tool.Name == toolName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	// Send tools/call request
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	request := NewRequest(3, "tools/call", params)
	response, err := c.transport.Send(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	// Parse result
	var result map[string]interface{}
	if err := response.UnmarshalResult(&result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool result: %w", err)
	}

	return result, nil
}

// Close closes the MCP client connection
func (c *MCPClient) Close() error {
	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}
