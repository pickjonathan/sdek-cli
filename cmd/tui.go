package cmd

import (
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/ui"
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

	// Initialize the Bubble Tea application
	app := ui.NewModel(state, tuiRole)
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Run the TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	slog.Info("TUI command completed")
	return nil
}
