package cmd

import (
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP server connections",
	Long: `Manage Model Context Protocol (MCP) server connections.

MCP servers provide tools and resources that can be used for evidence collection
and compliance analysis. Use subcommands to list servers, list tools, and test
connections.

Examples:
  sdek mcp list-servers      # List all configured MCP servers
  sdek mcp list-tools        # List all available tools from MCP servers
  sdek mcp test aws-api      # Test connection to a specific server
`,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
