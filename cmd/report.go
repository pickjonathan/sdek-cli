package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/report"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/spf13/cobra"
)

var (
	reportOutput string
	reportRole   string
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Export compliance report to JSON",
	Long: `Export a comprehensive compliance report in JSON format.

The report includes:
- Framework compliance percentages
- Control risk statuses and evidence counts
- Detailed evidence mappings
- Findings and recommendations
- Summary statistics

The report can be filtered by user role (compliance manager or engineer)
to show only relevant information.`,
	Example: `  # Export report to default location
  sdek report

  # Export report to specific file
  sdek report --output ~/reports/compliance-2024-10.json

  # Export report filtered for compliance manager view
  sdek report --role manager

  # Export report filtered for engineer view
  sdek report --role engineer`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate role if specified
		if reportRole != "" {
			validRoles := []string{"manager", "engineer"}
			valid := false
			for _, r := range validRoles {
				if reportRole == r {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid role '%s', must be one of: manager, engineer", reportRole)
			}
		}
		return nil
	},
	RunE: runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Get default report directory
	homeDir, _ := os.UserHomeDir()
	defaultOutput := filepath.Join(homeDir, ".sdek", "reports", fmt.Sprintf("compliance-%s.json", time.Now().Format("2006-01-02")))

	reportCmd.Flags().StringVarP(&reportOutput, "output", "o", defaultOutput, "Output file path for the report")
	reportCmd.Flags().StringVar(&reportRole, "role", "", "Filter report by role (manager, engineer)")
}

func runReport(cmd *cobra.Command, args []string) error {
	slog.Info("Starting report command", "output", reportOutput, "role", reportRole)

	// Load existing state
	state, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Check if we have data to report
	if len(state.Events) == 0 {
		return fmt.Errorf("no data found to report, run 'sdek seed --demo' first")
	}

	if len(state.Evidence) == 0 {
		slog.Warn("No evidence mappings found, run 'sdek analyze' to generate evidence")
	}

	// Create report directory if it doesn't exist
	reportDir := filepath.Dir(reportOutput)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Determine user role for filtering
	role := reportRole
	if role == "" {
		role = "all"
	}

	// Create exporter
	exporter := report.NewExporter(GetVersion())

	// Generate report
	slog.Info("Generating report", "role", role)
	reportData, err := exporter.GenerateReport(
		state.Sources,
		state.Events,
		state.Frameworks,
		state.Controls,
		state.Evidence,
		state.Findings,
		role,
	)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Format report
	formatter := report.NewFormatter()
	formattedData, err := formatter.FormatJSON(reportData, true) // pretty print
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	// Write report to file
	slog.Info("Writing report to file", "path", reportOutput)
	if err := os.WriteFile(reportOutput, formattedData, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	// Print summary
	fmt.Println("✓ Report generated successfully!")
	fmt.Println()
	fmt.Println("Report Details:")
	fmt.Printf("  Output file: %s\n", reportOutput)
	fmt.Printf("  File size:   %d bytes\n", len(formattedData))
	if reportRole != "" {
		fmt.Printf("  Role filter: %s\n", reportRole)
	}
	fmt.Println()
	fmt.Println("Report Contents:")
	fmt.Printf("  Frameworks:  %d\n", len(state.Frameworks))
	fmt.Printf("  Controls:    %d\n", len(state.Controls))
	fmt.Printf("  Evidence:    %d\n", len(state.Evidence))
	fmt.Printf("  Findings:    %d\n", len(state.Findings))
	fmt.Printf("  Events:      %d\n", len(state.Events))
	fmt.Println()
	
	// Show compliance summary
	fmt.Println("Compliance Summary:")
	for _, fw := range state.Frameworks {
		status := "✗"
		if fw.CompliancePercentage >= 80 {
			status = "✓"
		} else if fw.CompliancePercentage >= 60 {
			status = "⚠"
		}
		fmt.Printf("  %s %-15s %.1f%%\n", status, fw.Name, fw.CompliancePercentage)
	}
	fmt.Println()
	
	fmt.Printf("View the full report at: %s\n", reportOutput)

	slog.Info("Report command completed successfully")
	return nil
}
