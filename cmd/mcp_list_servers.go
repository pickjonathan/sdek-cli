package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mcpListServersCmd = &cobra.Command{
	Use:   "list-servers",
	Short: "List configured MCP servers",
	Long: `List all configured MCP servers with their current status.

Shows server name, transport type, health status, number of tools, and last health check time.`,
	RunE: runMCPListServers,
}

func init() {
	mcpCmd.AddCommand(mcpListServersCmd)
}

func runMCPListServers(cmd *cobra.Command, args []string) error {
	// Load MCP config from Viper (reads from config.yaml)
	var mcpConfig types.MCPConfig
	if err := viper.UnmarshalKey("mcp", &mcpConfig); err != nil {
		return fmt.Errorf("failed to load MCP config: %w", err)
	}

	// Check if MCP is enabled
	if !mcpConfig.Enabled {
		fmt.Println("MCP integration is disabled. Enable it in config to use MCP servers.")
		return nil
	}

	// Check if any servers are configured
	if len(mcpConfig.Servers) == 0 {
		fmt.Println("No MCP servers configured.")
		fmt.Println("\nAdd servers to ~/.sdek/config.yaml under mcp.servers section.")
		return nil
	}

	// Create MCP manager
	manager := mcp.NewMCPManager(mcpConfig)

	// Initialize manager (this will attempt to connect to all servers)
	ctx := cmd.Context()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP manager: %w", err)
	}
	defer manager.Close()

	// Get all servers
	servers := manager.ListServers()

	if len(servers) == 0 {
		fmt.Println("No MCP servers available.")
		return nil
	}

	// Display servers in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVER NAME\tTRANSPORT\tSTATUS\tTOOLS\tLAST CHECK\tERROR RATE")
	fmt.Fprintln(w, "───────────\t─────────\t──────\t─────\t──────────\t──────────")

	for _, server := range servers {
		// Format status with emoji
		statusStr := formatStatus(server.HealthStatus)

		// Format last check time
		var lastCheck string
		if server.LastHealthCheck.IsZero() {
			lastCheck = "Never"
		} else {
			elapsed := time.Since(server.LastHealthCheck)
			if elapsed < time.Minute {
				lastCheck = fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
			} else if elapsed < time.Hour {
				lastCheck = fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
			} else {
				lastCheck = fmt.Sprintf("%dh ago", int(elapsed.Hours()))
			}
		}

		// Format error rate
		errorRate := fmt.Sprintf("%.1f%%", server.Stats.ErrorRate*100)
		if server.Stats.TotalRequests == 0 {
			errorRate = "N/A"
		}

		// Get tool count
		toolCount := fmt.Sprintf("%d", len(server.Tools))

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			server.Name,
			server.Config.Transport,
			statusStr,
			toolCount,
			lastCheck,
			errorRate,
		)
	}

	w.Flush()

	// Print summary
	fmt.Println()
	healthyCount := 0
	degradedCount := 0
	downCount := 0
	for _, server := range servers {
		switch server.HealthStatus {
		case mcp.StatusHealthy:
			healthyCount++
		case mcp.StatusDegraded:
			degradedCount++
		case mcp.StatusDown:
			downCount++
		}
	}

	fmt.Printf("Summary: %d healthy, %d degraded, %d down (total: %d)\n",
		healthyCount, degradedCount, downCount, len(servers))

	// Show errors if any servers are down
	if downCount > 0 {
		fmt.Println("\nServers with errors:")
		for _, server := range servers {
			if server.HealthStatus == mcp.StatusDown && server.Stats.LastError != "" {
				fmt.Printf("  • %s: %s\n", server.Name, server.Stats.LastError)
			}
		}
	}

	return nil
}

func formatStatus(status mcp.ServerStatus) string {
	switch status {
	case mcp.StatusHealthy:
		return "✓ Healthy"
	case mcp.StatusDegraded:
		return "⚠ Degraded"
	case mcp.StatusDown:
		return "✗ Down"
	default:
		return "? Unknown"
	}
}
