package cmd

import (
	"fmt"
	"log/slog"

	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/spf13/cobra"
)

var (
	tuiRole string
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive terminal UI",
	Long: `Launch the interactive terminal user interface for exploring compliance data.

The TUI provides an interactive way to:
- Browse sources and events
- Explore compliance frameworks and controls
- View evidence mappings
- Investigate findings and risks
- Filter data by role (compliance manager or engineer)

Navigation:
- Tab:       Switch between sections
- ↑/↓:       Navigate lists
- Enter:     Select item for details
- ←:         Go back
- q:         Quit
- r:         Refresh data
- e:         Export report
- /:         Search
- ?:         Help

Minimum terminal size: 80 columns × 24 rows`,
	Example: `  # Launch TUI with default view
  sdek tui

  # Launch TUI with compliance manager view
  sdek tui --role manager

  # Launch TUI with engineer view
  sdek tui --role engineer`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate role if specified
		if tuiRole != "" {
			validRoles := []string{"manager", "engineer"}
			valid := false
			for _, r := range validRoles {
				if tuiRole == r {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid role '%s', must be one of: manager, engineer", tuiRole)
			}
		}
		return nil
	},
	RunE: runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	tuiCmd.Flags().StringVar(&tuiRole, "role", "", "Filter view by role (manager, engineer)")
}

func runTUI(cmd *cobra.Command, args []string) error {
	slog.Info("Starting TUI", "role", tuiRole)

	// Load existing state
	state, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Check if we have data to display
	if len(state.Events) == 0 {
		return fmt.Errorf("no data found to display, run 'sdek seed --demo' first")
	}

	// TODO: Implement Bubble Tea TUI
	// For now, provide a simple text-based interface

	fmt.Println("╔════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    sdek - Compliance Evidence Mapping                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("📊 Data Summary:")
	fmt.Println()
	fmt.Printf("  Sources:    %d\n", len(state.Sources))
	for _, source := range state.Sources {
		status := "✓"
		if !source.Enabled {
			status = "✗"
		}
		fmt.Printf("    %s %-15s %d events\n", status, source.Name, source.EventCount)
	}
	fmt.Println()

	fmt.Printf("  Events:     %d\n", len(state.Events))
	fmt.Printf("  Frameworks: %d\n", len(state.Frameworks))
	fmt.Printf("  Controls:   %d\n", len(state.Controls))
	fmt.Printf("  Evidence:   %d\n", len(state.Evidence))
	fmt.Printf("  Findings:   %d\n", len(state.Findings))
	fmt.Println()

	fmt.Println("🎯 Compliance Status:")
	fmt.Println()
	for _, fw := range state.Frameworks {
		status := "✗"
		if fw.CompliancePercentage >= 80 {
			status = "✓"
		} else if fw.CompliancePercentage >= 60 {
			status = "⚠"
		}

		bar := makeProgressBar(fw.CompliancePercentage, 30)
		fmt.Printf("  %s %-15s [%s] %.1f%%\n", status, fw.Name, bar, fw.CompliancePercentage)
	}
	fmt.Println()

	// Show risk distribution
	greenCount := 0
	yellowCount := 0
	redCount := 0
	for _, ctrl := range state.Controls {
		switch ctrl.RiskStatus {
		case "green":
			greenCount++
		case "yellow":
			yellowCount++
		case "red":
			redCount++
		}
	}

	fmt.Println("⚠️  Risk Distribution:")
	fmt.Println()
	fmt.Printf("  🟢 Green (Low Risk):     %3d controls\n", greenCount)
	fmt.Printf("  🟡 Yellow (Medium Risk): %3d controls\n", yellowCount)
	fmt.Printf("  🔴 Red (High Risk):      %3d controls\n", redCount)
	fmt.Println()

	// Show recent findings
	if len(state.Findings) > 0 {
		fmt.Println("🔍 Recent Findings:")
		fmt.Println()

		count := len(state.Findings)
		if count > 5 {
			count = 5
		}

		for i := 0; i < count; i++ {
			finding := state.Findings[i]
			severity := "  "
			switch finding.Severity {
			case "critical":
				severity = "🔴"
			case "high":
				severity = "🟠"
			case "medium":
				severity = "🟡"
			case "low":
				severity = "🟢"
			}

			fmt.Printf("  %s [%-8s] %s\n", severity, finding.Severity, finding.Title)
			fmt.Printf("     Control: %s - %s\n", finding.ControlID, finding.Description)
			fmt.Println()
		}

		if len(state.Findings) > 5 {
			fmt.Printf("  ... and %d more findings\n", len(state.Findings)-5)
			fmt.Println()
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("NOTE: Full interactive TUI (Bubble Tea) is under development.")
	fmt.Println("      Use CLI commands for now: analyze, report, ingest")
	fmt.Println()

	slog.Info("TUI command completed")
	return nil
}

// makeProgressBar creates a visual progress bar
func makeProgressBar(percentage float64, width int) string {
	filled := int(percentage / 100.0 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < width; i++ {
		bar += "░"
	}

	return bar
}
