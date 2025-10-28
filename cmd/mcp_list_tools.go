package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listToolsServer  string
	listToolsVerbose bool
)

var mcpListToolsCmd = &cobra.Command{
	Use:   "list-tools",
	Short: "List available tools from MCP servers",
	Long: `List all tools discovered from configured MCP servers.

Tools can be filtered by server using the --server flag. Use --verbose to see
detailed parameter schemas for each tool.`,
	RunE: runMCPListTools,
}

func init() {
	mcpCmd.AddCommand(mcpListToolsCmd)
	mcpListToolsCmd.Flags().StringVar(&listToolsServer, "server", "", "Filter tools by server name")
	mcpListToolsCmd.Flags().BoolVarP(&listToolsVerbose, "verbose", "v", false, "Show detailed parameter schemas")
}

func runMCPListTools(cmd *cobra.Command, args []string) error {
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

	// Initialize manager
	ctx := cmd.Context()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize MCP manager: %w", err)
	}
	defer manager.Close()

	// Get tools
	tools := manager.DiscoverTools()

	// Filter by server if specified
	if listToolsServer != "" {
		var filtered []struct {
			Name        string
			Description string
			ServerName  string
			Parameters  map[string]interface{}
		}
		for _, tool := range tools {
			if tool.ServerName == listToolsServer {
				filtered = append(filtered, struct {
					Name        string
					Description string
					ServerName  string
					Parameters  map[string]interface{}
				}{
					Name:        tool.Name,
					Description: tool.Description,
					ServerName:  tool.ServerName,
					Parameters:  tool.Parameters,
				})
			}
		}

		if len(filtered) == 0 {
			fmt.Printf("No tools found for server: %s\n", listToolsServer)
			return nil
		}

		// Display filtered tools
		displayTools(filtered, listToolsVerbose)
		return nil
	}

	if len(tools) == 0 {
		fmt.Println("No tools discovered from MCP servers.")
		return nil
	}

	// Convert to display format
	var displayList []struct {
		Name        string
		Description string
		ServerName  string
		Parameters  map[string]interface{}
	}
	for _, tool := range tools {
		displayList = append(displayList, struct {
			Name        string
			Description string
			ServerName  string
			Parameters  map[string]interface{}
		}{
			Name:        tool.Name,
			Description: tool.Description,
			ServerName:  tool.ServerName,
			Parameters:  tool.Parameters,
		})
	}

	displayTools(displayList, listToolsVerbose)
	return nil
}

func displayTools(tools []struct {
	Name        string
	Description string
	ServerName  string
	Parameters  map[string]interface{}
}, verbose bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !verbose {
		fmt.Fprintln(w, "TOOL NAME\tDESCRIPTION\tSERVER")
		fmt.Fprintln(w, "─────────\t───────────\t──────")

		for _, tool := range tools {
			// Truncate description if too long
			desc := tool.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\n",
				tool.Name,
				desc,
				tool.ServerName,
			)
		}

		w.Flush()
		fmt.Printf("\nTotal: %d tools\n", len(tools))
	} else {
		// Verbose mode: show detailed information
		for i, tool := range tools {
			if i > 0 {
				fmt.Println()
			}

			fmt.Printf("Tool: %s\n", tool.Name)
			fmt.Printf("Server: %s\n", tool.ServerName)
			fmt.Printf("Description: %s\n", tool.Description)

			if tool.Parameters != nil {
				fmt.Println("Parameters:")
				paramsJSON, err := json.MarshalIndent(tool.Parameters, "  ", "  ")
				if err != nil {
					fmt.Printf("  (error formatting parameters: %v)\n", err)
				} else {
					lines := strings.Split(string(paramsJSON), "\n")
					for _, line := range lines {
						fmt.Printf("  %s\n", line)
					}
				}
			} else {
				fmt.Println("Parameters: None")
			}

			fmt.Println(strings.Repeat("─", 80))
		}

		fmt.Printf("\nTotal: %d tools\n", len(tools))
	}
}
