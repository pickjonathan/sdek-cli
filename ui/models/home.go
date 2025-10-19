package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// HomeModel represents the home screen
type HomeModel struct {
	state          *store.State
	width          int
	height         int
	selectedScreen int // 0 = none, 1 = sources, 2 = frameworks, 3 = controls, 4 = evidence
}

// NewHomeModel creates a new home screen model
func NewHomeModel(state *store.State) HomeModel {
	return HomeModel{
		state:          state,
		selectedScreen: 0,
	}
}

// Init initializes the home model
func (m HomeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.selectedScreen = 1
		case "2":
			m.selectedScreen = 2
		case "3":
			m.selectedScreen = 3
		case "4":
			m.selectedScreen = 4
		case "right", "l":
			// Cycle forward through screens (1->2->3->4->1)
			if m.selectedScreen == 0 {
				m.selectedScreen = 1
			} else {
				m.selectedScreen = (m.selectedScreen % 4) + 1
			}
		case "left", "h":
			// Cycle backward through screens (4->3->2->1->4)
			if m.selectedScreen == 0 {
				m.selectedScreen = 4
			} else if m.selectedScreen == 1 {
				m.selectedScreen = 4
			} else {
				m.selectedScreen = m.selectedScreen - 1
			}
		case "enter":
			// Enter key doesn't change selectedScreen, just returns it
			// The app.go will handle the navigation
		}
	}
	return m, nil
}

// View renders the home screen
func (m HomeModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s string

	// Header
	title := styles.TitleStyle.Render("sdek - Compliance Evidence Mapping")
	subtitle := styles.SubtitleStyle.Render("Press 1-4 to navigate | ? for help | q to quit")

	s += title + "\n"
	s += subtitle + "\n\n"

	// Summary cards
	s += m.renderSummary() + "\n\n"

	// Compliance status
	s += m.renderCompliance() + "\n\n"

	// Quick actions
	s += m.renderQuickActions() + "\n"

	return s
}

// renderSummary renders the data summary section
func (m HomeModel) renderSummary() string {
	sourcesCard := m.makeCard("Sources", fmt.Sprintf("%d", len(m.state.Sources)), "1", "📊")
	frameworksCard := m.makeCard("Frameworks", fmt.Sprintf("%d", len(m.state.Frameworks)), "2", "📋")
	controlsCard := m.makeCard("Controls", fmt.Sprintf("%d", len(m.state.Controls)), "3", "🎯")
	evidenceCard := m.makeCard("Evidence", fmt.Sprintf("%d", len(m.state.Evidence)), "4", "📄")

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sourcesCard,
		"  ",
		frameworksCard,
		"  ",
		controlsCard,
		"  ",
		evidenceCard,
	)
}

// makeCard creates a summary card
func (m HomeModel) makeCard(title, value, key, icon string) string {
	cardWidth := (m.width - 8) / 4
	if cardWidth < 15 {
		cardWidth = 15
	}

	// Check if this card is selected
	keyNum := 0
	if key == "1" {
		keyNum = 1
	} else if key == "2" {
		keyNum = 2
	} else if key == "3" {
		keyNum = 3
	} else if key == "4" {
		keyNum = 4
	}

	cardStyle := styles.CardStyle.Width(cardWidth)
	if keyNum == m.selectedScreen {
		// Highlight selected card
		cardStyle = cardStyle.BorderForeground(lipgloss.Color(styles.AccentColor))
	}

	content := fmt.Sprintf("%s %s\n\n%s\n\nPress %s",
		icon,
		title,
		styles.HeaderStyle.Render(value),
		styles.KeyStyle.Render(key),
	)

	return cardStyle.Render(content)
}

// renderCompliance renders the compliance status section
func (m HomeModel) renderCompliance() string {
	var s string

	s += styles.HeaderStyle.Render("Compliance Status") + "\n\n"

	if len(m.state.Frameworks) == 0 {
		s += styles.SubtleStyle.Render("No frameworks configured. Run 'sdek seed --demo' to generate data.")
		return s
	}

	for _, framework := range m.state.Frameworks {
		// Calculate compliance percentage
		compliantControls := 0
		totalControls := 0

		for _, control := range m.state.Controls {
			if control.FrameworkID == framework.ID {
				totalControls++
				if control.RiskStatus == "green" || control.RiskStatus == "Green" {
					compliantControls++
				}
			}
		}

		percentage := 0.0
		if totalControls > 0 {
			percentage = float64(compliantControls) / float64(totalControls) * 100
		}

		// Render framework row
		nameCol := lipgloss.NewStyle().Width(20).Render(framework.Name)
		percentCol := m.formatPercentage(percentage)
		barCol := m.makeProgressBar(percentage, 30)
		statusCol := m.formatStatus(percentage)

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			nameCol,
			"  ",
			percentCol,
			"  ",
			barCol,
			"  ",
			statusCol,
		)

		s += row + "\n"
	}

	return s
}

// renderQuickActions renders quick action buttons
func (m HomeModel) renderQuickActions() string {
	s := styles.HeaderStyle.Render("Quick Actions") + "\n\n"

	actions := []struct {
		key  string
		desc string
	}{
		{"r", "Refresh data"},
		{"e", "Export report"},
		{"?", "Show help"},
		{"q", "Quit"},
	}

	for _, action := range actions {
		s += fmt.Sprintf("  %s  %s\n",
			styles.KeyStyle.Render(action.key),
			action.desc,
		)
	}

	return s
}

// formatPercentage formats a percentage with color
func (m HomeModel) formatPercentage(percent float64) string {
	var style lipgloss.Style
	if percent >= 80 {
		style = styles.StatusGreenStyle
	} else if percent >= 50 {
		style = styles.StatusYellowStyle
	} else {
		style = styles.StatusRedStyle
	}
	return style.Render(fmt.Sprintf("%5.1f%%", percent))
}

// formatStatus formats a status indicator
func (m HomeModel) formatStatus(percent float64) string {
	if percent >= 80 {
		return styles.StatusGreenStyle.Render("✓ Good")
	} else if percent >= 50 {
		return styles.StatusYellowStyle.Render("⚠ Review")
	} else {
		return styles.StatusRedStyle.Render("✗ Action Required")
	}
}

// makeProgressBar creates a progress bar
func (m HomeModel) makeProgressBar(percent float64, width int) string {
	filled := int(float64(width) * percent / 100)
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	var style lipgloss.Style
	if percent >= 80 {
		style = styles.StatusGreenStyle
	} else if percent >= 50 {
		style = styles.StatusYellowStyle
	} else {
		style = styles.StatusRedStyle
	}

	return style.Render(bar)
}

// SetSize updates the model dimensions
func (m *HomeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SelectedScreen returns the selected screen number
func (m HomeModel) SelectedScreen() int {
	return m.selectedScreen
}

// ResetSelection resets the selected screen
func (m *HomeModel) ResetSelection() {
	m.selectedScreen = 0
}
