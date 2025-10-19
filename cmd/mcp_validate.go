package cmd

import (
	"context"
	"fmt"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/spf13/cobra"
)

// mcpValidateCmd validates MCP configuration files
var mcpValidateCmd = &cobra.Command{
	Use:   "validate [file...]",
	Short: "Validate MCP configuration files",
	Long: `Validate one or more MCP configuration files against the JSON schema.

This command checks that configuration files:
  - Are valid JSON
  - Match the required schema structure
  - Have all required fields
  - Use valid transport types (stdio or http)
  - Have properly formatted capabilities

Exit code 0 indicates all files are valid.
Exit code 1 indicates validation errors were found.`,
	Example: `  # Validate a single file
  sdek mcp validate ~/.sdek/mcp/aws.json

  # Validate multiple files
  sdek mcp validate ~/.sdek/mcp/*.json

  # Validate all configs in a directory
  sdek mcp validate ~/.sdek/mcp/`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize registry with validator
		registry := mcp.NewRegistry()

		// Validate each file
		hasErrors := false
		for _, path := range args {
			errors, err := registry.Validate(ctx, path)
			if err != nil {
				return fmt.Errorf("validation error for %s: %w", path, err)
			}
			if len(errors) > 0 {
				hasErrors = true
				fmt.Printf("❌ %s:\n", path)
				for _, schemaErr := range errors {
					fmt.Printf("  %s\n", schemaErr.Error())
				}
			} else {
				fmt.Printf("✓ %s\n", path)
			}
		}

		if hasErrors {
			return fmt.Errorf("validation failed for one or more files")
		}

		fmt.Println("\n✓ All configuration files are valid")
		return nil
	},
}

func init() {
	mcpCmd.AddCommand(mcpValidateCmd)
}
