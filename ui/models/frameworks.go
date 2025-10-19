package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// FrameworksModel represents the frameworks list screen
type FrameworksModel struct {
	state  *store.State
	width  int
	height int
	cursor int
}

// NewFrameworksModel creates a new frameworks model
func NewFrameworksModel(state *store.State) FrameworksModel {
	return FrameworksModel{
		state:  state,
		cursor: 0,
	}
}

// Init initializes the frameworks model
func (m FrameworksModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m FrameworksModel) Update(msg tea.Msg) (FrameworksModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.state.Frameworks)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// View renders the frameworks screen
func (m FrameworksModel) View() string {
	s := styles.TitleStyle.Render("Compliance Frameworks") + "\n\n"

	if len(m.state.Frameworks) == 0 {
		s += styles.SubtleStyle.Render("No frameworks found. Run 'sdek seed --demo' to generate data.")
		return s
	}

	for i, framework := range m.state.Frameworks {
		// Calculate compliance
		compliant := 0
		total := 0
		for _, ctrl := range m.state.Controls {
			if ctrl.FrameworkID == framework.ID {
				total++
				if ctrl.RiskStatus == "green" || ctrl.RiskStatus == "Green" {
					compliant++
				}
			}
		}

		percent := 0.0
		if total > 0 {
			percent = float64(compliant) / float64(total) * 100
		}

		cursor := " "
		line := fmt.Sprintf("%s %s - %.1f%% compliant (%d/%d controls)",
			cursor,
			framework.Name,
			percent,
			compliant,
			total,
		)

		if m.cursor == i {
			cursor = ">"
			s += styles.SelectedListItemStyle.Render(line) + "\n"
		} else {
			s += styles.ListItemStyle.Render(line) + "\n"
		}
	}

	s += "\n" + styles.SubtleStyle.Render("↑/↓: Navigate | Esc: Back | q: Quit")

	return s
}

// SetSize updates the model dimensions
func (m *FrameworksModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
