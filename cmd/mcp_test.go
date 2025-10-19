package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/spf13/cobra"
)

// mcpTestCmd tests an MCP tool connection
var mcpTestCmd = &cobra.Command{
	Use:   "test <tool-name>",
	Short: "Test an MCP tool connection",
	Long: `Test an MCP tool connection and display diagnostic information.

This command:
  - Performs a health check on the tool
  - Measures handshake latency
  - Verifies capabilities
  - Displays any errors

Use this to troubleshoot connectivity issues or verify a tool is working correctly.`,
	Example: `  # Test the AWS tool
  sdek mcp test aws

  # Test the GitHub tool
  sdek mcp test github`,
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

		// Test the tool
		fmt.Printf("Testing MCP tool: %s\n\n", toolName)
		
		report, err := registry.Test(ctx, toolName)
		if err != nil {
			return fmt.Errorf("test failed: %w", err)
		}

		// Display results
		fmt.Printf("Tool Name:    %s\n", report.ToolName)
		fmt.Printf("Status:       %s\n", report.Status)
		fmt.Printf("Latency:      %s\n", formatDuration(report.HandshakeLatency))
		fmt.Printf("Capabilities: %d\n", len(report.Capabilities))
		
		if len(report.Capabilities) > 0 {
			fmt.Println("\nCapabilities:")
			for _, cap := range report.Capabilities {
				fmt.Printf("  - %s\n", cap)
			}
		}

		if report.LastError != nil {
			fmt.Printf("\nLast Error:   %s\n", report.LastError.Error())
		}

		fmt.Printf("\nChecked:      %s\n", report.Timestamp.Format(time.RFC3339))

		if report.Status == "ready" {
			fmt.Println("\n✓ Tool is operational")
		} else {
			fmt.Println("\n⚠ Tool has issues")
		}

		return nil
	},
}

func init() {
	mcpCmd.AddCommand(mcpTestCmd)
}
