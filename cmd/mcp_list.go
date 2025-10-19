package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	mcpListFormat string
)

// mcpListCmd lists all configured MCP tools with their status
var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured MCP tools",
	Long: `List all configured MCP tools with their current status.

Displays:
  - Tool name
  - Status (ready, degraded, offline)
  - Last health check time
  - Average latency
  - Error count
  - Capabilities

The status reflects the health of the MCP tool connection:
  - ready: Tool is operational and responding
  - degraded: Tool has experienced recent failures but may still work
  - offline: Tool is not responding or disabled`,
	Example: `  # List all tools in table format
  sdek mcp list

  # List all tools in JSON format
  sdek mcp list --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize registry
		registry := mcp.NewRegistry()
		if _, err := registry.Init(ctx); err != nil {
			return fmt.Errorf("failed to initialize MCP registry: %w", err)
		}
		defer registry.Close(ctx)

		// Get all tools
		tools, err := registry.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tools: %w", err)
		}

		if len(tools) == 0 {
			fmt.Println("No MCP tools configured.")
			fmt.Println("\nTo add a tool, create a JSON config file in:")
			fmt.Println("  - ~/.sdek/mcp/")
			fmt.Println("  - ./.sdek/mcp/")
			fmt.Println("  - $SDEK_MCP_PATH")
			return nil
		}

		// Format output
		switch mcpListFormat {
		case "json":
			return outputJSON(tools)
		default:
			return outputTable(tools)
		}
	},
}

func outputTable(tools []types.MCPTool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "NAME\tSTATUS\tLATENCY\tERRORS\tCAPABILITIES\tLAST CHECK")

	// Rows
	for _, tool := range tools {
		status := string(tool.Status)
		latency := formatDuration(tool.Metrics.AverageLatency)
		errors := fmt.Sprintf("%d", tool.Metrics.ErrorCount)
		capabilities := fmt.Sprintf("%d", len(tool.Config.Capabilities))
		lastCheck := formatTime(tool.LastHealthCheck)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			tool.Name,
			status,
			latency,
			errors,
			capabilities,
			lastCheck,
		)
	}

	return nil
}

func outputJSON(tools []types.MCPTool) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tools)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dμs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	elapsed := time.Since(t)
	if elapsed < time.Minute {
		return fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
	}
	if elapsed < time.Hour {
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	}
	if elapsed < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	}
	return t.Format("2006-01-02")
}

func init() {
	mcpCmd.AddCommand(mcpListCmd)
	mcpListCmd.Flags().StringVar(&mcpListFormat, "format", "table", "Output format (table or json)")
}
