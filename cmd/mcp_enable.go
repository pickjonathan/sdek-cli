package cmd

import (
	"context"
	"fmt"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/spf13/cobra"
)

// mcpEnableCmd enables an MCP tool
var mcpEnableCmd = &cobra.Command{
	Use:   "enable <tool-name>",
	Short: "Enable an MCP tool",
	Long: `Enable an MCP tool that was previously disabled.

When a tool is enabled:
  - It will be initialized and health checked
  - It can be invoked by agents and CLI commands
  - It will appear in the TUI
  - Circuit breaker will attempt to transition it to ready state

If the tool was already enabled, this command is a no-op.`,
	Example: `  # Enable the AWS tool
  sdek mcp enable aws

  # Enable the Jira tool
  sdek mcp enable jira`,
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

		// Enable the tool
		if err := registry.Enable(ctx, toolName); err != nil {
			return fmt.Errorf("failed to enable tool: %w", err)
		}

		fmt.Printf("✓ Enabled MCP tool: %s\n", toolName)
		fmt.Println("\nThe tool will be initialized and health checked.")
		fmt.Printf("Check status with: sdek mcp test %s\n", toolName)

		return nil
	},
}

func init() {
	mcpCmd.AddCommand(mcpEnableCmd)
}
