package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// SourcesModel represents the sources list screen
type SourcesModel struct {
	state  *store.State
	width  int
	height int
	cursor int
}

// NewSourcesModel creates a new sources model
func NewSourcesModel(state *store.State) SourcesModel {
	return SourcesModel{
		state:  state,
		cursor: 0,
	}
}

// Init initializes the sources model
func (m SourcesModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SourcesModel) Update(msg tea.Msg) (SourcesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.state.Sources)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// View renders the sources screen
func (m SourcesModel) View() string {
	s := styles.TitleStyle.Render("Data Sources") + "\n\n"

	if len(m.state.Sources) == 0 {
		s += styles.SubtleStyle.Render("No sources found. Run 'sdek seed --demo' to generate data.")
		return s
	}

	for i, source := range m.state.Sources {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += styles.SelectedListItemStyle.Render(
				cursor+" "+source.Type+" ("+source.Name+")",
			) + "\n"
		} else {
			s += styles.ListItemStyle.Render(
				cursor+" "+source.Type+" ("+source.Name+")",
			) + "\n"
		}
	}

	s += "\n" + styles.SubtleStyle.Render("↑/↓: Navigate | Esc: Back | q: Quit")

	return s
}

// SetSize updates the model dimensions
func (m *SourcesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
