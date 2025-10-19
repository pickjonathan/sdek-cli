package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mcpCmd represents the parent mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) tool integrations",
	Long: `Manage MCP (Model Context Protocol) tool integrations.

MCP tools provide programmatic access to external services like GitHub, Jira,
AWS, and other data sources. This command allows you to list, validate, test,
enable, and disable MCP tools configured in your environment.

Configuration files are discovered from:
  - Project: ./.sdek/mcp/
  - Global: ~/.sdek/mcp/
  - Custom: $SDEK_MCP_PATH

Each MCP tool requires a JSON configuration file with:
  - name: Unique tool identifier
  - command: Executable to run (e.g., npx, uvx, python)
  - args: Command arguments
  - transport: Communication protocol (stdio or http)
  - capabilities: List of supported operations`,
	Example: `  # List all configured MCP tools
  sdek mcp list

  # Validate a configuration file
  sdek mcp validate ~/.sdek/mcp/aws.json

  # Test a specific tool
  sdek mcp test github

  # Enable a tool
  sdek mcp enable jira

  # Disable a tool
  sdek mcp disable aws`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if MCP feature is enabled in config
		if !viper.GetBool("features.mcp.enabled") {
			return fmt.Errorf("MCP feature is not enabled. Enable it in your config with: sdek config set features.mcp.enabled true")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
