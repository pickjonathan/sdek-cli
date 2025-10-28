package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// ConnectorAdapter adapts MCPManager to the ai.MCPConnector interface
// This allows the AI engine to use MCP servers for evidence collection
type ConnectorAdapter struct {
	manager *MCPManager
}

// NewConnectorAdapter creates a new adapter wrapping an MCP manager
func NewConnectorAdapter(manager *MCPManager) *ConnectorAdapter {
	return &ConnectorAdapter{
		manager: manager,
	}
}

// Manager returns the underlying MCP manager
func (a *ConnectorAdapter) Manager() *MCPManager {
	return a.manager
}

// Collect implements the ai.MCPConnector interface
// It routes evidence collection requests to the appropriate MCP server
func (a *ConnectorAdapter) Collect(ctx context.Context, source string, query string) ([]types.EvidenceEvent, error) {
	slog.Debug("MCP Connector: Collect called", "source", source, "query", query)

	// Map AI-generated source names to actual MCP servers
	// Ollama generates sources like "aws-cloudtrail", "okta", "github"
	// but we need to map these to actual MCP server names
	mappedSource := mapSourceToMCPServer(source)
	slog.Debug("MCP Connector: Source mapped", "original", source, "mapped", mappedSource)

	// Parse the source to determine server and tool
	// Source format can be:
	// - "server:tool" (e.g., "aws-api:call_aws")
	// - "server" (use default tool for that server)
	// - "tool" (search all servers for the tool)

	serverName, toolName := parseSource(mappedSource)
	slog.Debug("MCP Connector: Source parsed", "server", serverName, "tool", toolName)

	// If we have a specific server, use it
	if serverName != "" {
		return a.collectFromServer(ctx, serverName, toolName, query)
	}

	// Otherwise, search all servers for a tool with this name
	return a.collectFromAnyServer(ctx, toolName, query)
}

// collectFromServer collects evidence from a specific MCP server
func (a *ConnectorAdapter) collectFromServer(ctx context.Context, serverName, toolName, query string) ([]types.EvidenceEvent, error) {
	slog.Debug("MCP Connector: collectFromServer called", "server", serverName, "tool", toolName, "query", query)

	// Get the server
	server, exists := a.manager.GetServer(serverName)
	if !exists {
		slog.Error("MCP Connector: Server not found", "server", serverName)
		return nil, fmt.Errorf("MCP server not found: %s", serverName)
	}

	// Check server health
	if server.HealthStatus == StatusDown {
		slog.Error("MCP Connector: Server is down", "server", serverName, "error", server.Stats.LastError)
		return nil, fmt.Errorf("MCP server %s is down: %s", serverName, server.Stats.LastError)
	}
	slog.Debug("MCP Connector: Server health OK", "server", serverName, "status", server.HealthStatus)

	// If no tool name specified, try to infer it
	if toolName == "" {
		// Use the first available tool from the server
		if len(server.Tools) == 0 {
			slog.Error("MCP Connector: No tools available", "server", serverName)
			return nil, fmt.Errorf("no tools available from server %s", serverName)
		}
		toolName = server.Tools[0].Name
		slog.Debug("MCP Connector: Using first available tool", "tool", toolName)
	}

	// Prepare arguments for the tool
	// The query string becomes the tool arguments
	arguments := map[string]interface{}{
		"query":   query,
		"command": query, // Some tools use "command" instead of "query"
	}
	slog.Debug("MCP Connector: Executing tool", "server", serverName, "tool", toolName, "arguments", arguments)

	// Execute the tool
	result, err := a.manager.ExecuteTool(ctx, serverName, toolName, arguments)
	if err != nil {
		slog.Error("MCP Connector: Tool execution failed", "server", serverName, "tool", toolName, "error", err)
		return nil, fmt.Errorf("failed to execute tool %s on server %s: %w", toolName, serverName, err)
	}
	slog.Debug("MCP Connector: Tool executed successfully", "server", serverName, "tool", toolName)

	// Normalize the result to EvidenceEvent format
	events, err := NormalizeToEvidenceEvent(serverName, toolName, result)
	if err != nil {
		slog.Error("MCP Connector: Normalization failed", "error", err)
		return nil, fmt.Errorf("failed to normalize tool result: %w", err)
	}

	slog.Info("MCP Connector: Evidence collected", "server", serverName, "tool", toolName, "events", len(events))
	return events, nil
}

// collectFromAnyServer searches all servers for a tool and executes it
func (a *ConnectorAdapter) collectFromAnyServer(ctx context.Context, toolName, query string) ([]types.EvidenceEvent, error) {
	// Discover all tools
	allTools := a.manager.DiscoverTools()

	// Find the tool
	var foundServer string
	for _, tool := range allTools {
		if tool.Name == toolName {
			foundServer = tool.ServerName
			break
		}
	}

	if foundServer == "" {
		return nil, fmt.Errorf("tool %s not found on any MCP server", toolName)
	}

	// Execute via the found server
	return a.collectFromServer(ctx, foundServer, toolName, query)
}

// parseSource parses a source string into server and tool components
// Supports formats:
// - "server:tool" → ("server", "tool")
// - "server" → ("server", "")
// - "tool" → ("", "tool")
func parseSource(source string) (server, tool string) {
	// Check if it contains a colon
	for i, c := range source {
		if c == ':' {
			return source[:i], source[i+1:]
		}
	}

	// No colon found - could be either server or tool
	// Try to determine by checking if it looks like a server name
	// Server names typically contain hyphens (aws-api, github-mcp)
	// Tool names typically use underscores (call_aws, list_users)
	for _, c := range source {
		if c == '-' {
			// Likely a server name
			return source, ""
		}
	}

	// Assume it's a tool name
	return "", source
}

// mapSourceToMCPServer maps AI-generated source names to actual MCP server names
// Ollama and other AIs generate logical source names like "aws-cloudtrail", "okta", "github"
// but we need to map these to the actual configured MCP servers like "aws-api"
func mapSourceToMCPServer(source string) string {
	// Normalize source name for consistent matching:
	// 1. Convert to lowercase
	// 2. Replace spaces and underscores with hyphens
	// Examples:
	//   "AWS CloudTrail" → "aws-cloudtrail"
	//   "aws_cloudtrail" → "aws-cloudtrail"
	//   "AWS IAM" → "aws-iam"
	normalizedSource := ""
	for _, c := range source {
		if c == '_' || c == ' ' {
			normalizedSource += "-"
		} else if c >= 'A' && c <= 'Z' {
			// Convert uppercase to lowercase
			normalizedSource += string(c + 32)
		} else {
			normalizedSource += string(c)
		}
	}

	// Define mappings from AI-generated sources to MCP servers
	mappings := map[string]string{
		// AWS-related sources all map to aws-api server
		"aws-cloudtrail":    "aws-api:call_aws",
		"aws-iam":           "aws-api:call_aws",
		"aws":               "aws-api:call_aws",
		"cloudtrail":        "aws-api:call_aws",
		"iam":               "aws-api:call_aws",

		// Azure sources (if azure-mcp server is configured)
		"azure-ad":          "azure-mcp",
		"azure-ad-audit":    "azure-mcp",
		"azure":             "azure-mcp",

		// Okta sources
		"okta":              "okta-mcp",
		"okta-audit":        "okta-mcp",

		// GitHub sources
		"github":            "github-mcp",

		// Jira sources
		"jira":              "jira-mcp",

		// Other common sources
		"ldap-logs":         "ldap-mcp",
		"splunk":            "splunk-mcp",
		"terraform-state":   "terraform-mcp",
		"vault":             "vault-mcp",
		"vault-audit":       "vault-mcp",
		"service-now":       "servicenow-mcp",
		"security-hub":      "aws-api:call_aws", // AWS Security Hub
		"gcp-iam":           "gcp-mcp",
		"gcp-audit":         "gcp-mcp",
		"confluence":        "confluence-mcp",
		"hr-system":         "hr-mcp",
		"active-directory":  "ad-mcp",
		"slack":             "slack-mcp",
	}

	// Check if we have a mapping for normalized source
	if mapped, ok := mappings[normalizedSource]; ok {
		return mapped
	}

	// No mapping found, return original source
	// This allows for direct specification of "server:tool" format
	return source
}
