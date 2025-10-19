package cmd

import (
	"context"
	"fmt"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/spf13/cobra"
)

// mcpDisableCmd disables an MCP tool
var mcpDisableCmd = &cobra.Command{
	Use:   "disable <tool-name>",
	Short: "Disable an MCP tool",
	Long: `Disable an MCP tool to prevent it from being invoked.

When a tool is disabled:
  - It will not be available for agent invocations
  - CLI test commands will fail
  - It will show as offline in the TUI
  - Configuration remains but tool is marked inactive

The tool's configuration file is not deleted and can be re-enabled later.
Use this to temporarily disable a tool without removing its configuration.`,
	Example: `  # Disable the AWS tool
  sdek mcp disable aws

  # Disable the GitHub tool  
  sdek mcp disable github`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		toolName := args[0]

		// Initialize registry
		registry := mcp.NewRegistry()
		if _, err := registry.Init(ctx); err != nil {
			return fmt.Errorf("failed to initialize MCP registry: %w", err)
		}
		defer registry.Close(ctx)

		// Disable the tool
		if err := registry.Disable(ctx, toolName); err != nil {
			return fmt.Errorf("failed to disable tool: %w", err)
		}

		fmt.Printf("✓ Disabled MCP tool: %s\n", toolName)
		fmt.Println("\nThe tool will no longer accept invocations.")
		fmt.Printf("Re-enable with: sdek mcp enable %s\n", toolName)

		return nil
	},
}

func init() {
	mcpCmd.AddCommand(mcpDisableCmd)
}
