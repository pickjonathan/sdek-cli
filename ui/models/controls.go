package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// ControlsModel represents the controls list screen
type ControlsModel struct {
	state       *store.State
	width       int
	height      int
	cursor      int
	frameworkID string
}

// NewControlsModel creates a new controls model
func NewControlsModel(state *store.State, frameworkID string) ControlsModel {
	return ControlsModel{
		state:       state,
		frameworkID: frameworkID,
		cursor:      0,
	}
}

// Init initializes the controls model
func (m ControlsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ControlsModel) Update(msg tea.Msg) (ControlsModel, tea.Cmd) {
	controls := m.getFilteredControls()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(controls)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// View renders the controls screen
func (m ControlsModel) View() string {
	s := styles.TitleStyle.Render("Controls") + "\n\n"

	controls := m.getFilteredControls()

	if len(controls) == 0 {
		s += styles.SubtleStyle.Render("No controls found. Run 'sdek seed --demo' to generate data.")
		return s
	}

	// Group by risk status
	greenCount := 0
	yellowCount := 0
	redCount := 0

	for _, ctrl := range controls {
		switch ctrl.RiskStatus {
		case "green", "Green":
			greenCount++
		case "yellow", "Yellow":
			yellowCount++
		case "red", "Red":
			redCount++
		}
	}

	s += fmt.Sprintf("Risk Summary: %s %d  %s %d  %s %d\n\n",
		styles.StatusGreenStyle.Render("●"),
		greenCount,
		styles.StatusYellowStyle.Render("●"),
		yellowCount,
		styles.StatusRedStyle.Render("●"),
		redCount,
	)

	// Show controls
	displayCount := 0
	maxDisplay := m.height - 10
	if maxDisplay < 5 {
		maxDisplay = 5
	}

	for i, ctrl := range controls {
		if displayCount >= maxDisplay {
			s += styles.SubtleStyle.Render(fmt.Sprintf("\n... and %d more controls", len(controls)-displayCount))
			break
		}

		cursor := " "
		statusIcon := m.getStatusIcon(ctrl.RiskStatus)
		line := fmt.Sprintf("%s %s %s - %s",
			cursor,
			statusIcon,
			ctrl.ID,
			ctrl.Title,
		)

		if m.cursor == i {
			cursor = ">"
			s += styles.SelectedListItemStyle.Render(line) + "\n"
		} else {
			s += styles.ListItemStyle.Render(line) + "\n"
		}
		displayCount++
	}

	s += "\n" + styles.SubtleStyle.Render("↑/↓: Navigate | Esc: Back | q: Quit")

	return s
}

// getFilteredControls returns controls filtered by framework if set
func (m ControlsModel) getFilteredControls() []types.Control {
	if m.frameworkID == "" {
		// Return all controls
		controls := make([]types.Control, len(m.state.Controls))
		copy(controls, m.state.Controls)
		return controls
	}

	// Filter by framework
	var filtered []types.Control
	for _, ctrl := range m.state.Controls {
		if ctrl.FrameworkID == m.frameworkID {
			filtered = append(filtered, ctrl)
		}
	}
	return filtered
}

// getStatusIcon returns an icon for the risk status
func (m ControlsModel) getStatusIcon(status string) string {
	switch status {
	case "green", "Green":
		return styles.StatusGreenStyle.Render("✓")
	case "yellow", "Yellow":
		return styles.StatusYellowStyle.Render("⚠")
	case "red", "Red":
		return styles.StatusRedStyle.Render("✗")
	default:
		return styles.SubtleStyle.Render("○")
	}
}

// SetSize updates the model dimensions
func (m *ControlsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
