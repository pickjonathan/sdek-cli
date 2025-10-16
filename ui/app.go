package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/ui/models"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// Screen represents the different screens in the TUI
type Screen int

const (
	ScreenHome Screen = iota
	ScreenSources
	ScreenFrameworks
	ScreenControls
	ScreenEvidence
	ScreenHelp
)

// Model is the main application model for the TUI
type Model struct {
	state         *store.State
	currentScreen Screen
	width         int
	height        int
	role          string

	// Screen models
	homeModel       models.HomeModel
	sourcesModel    models.SourcesModel
	frameworksModel models.FrameworksModel
	controlsModel   models.ControlsModel
	evidenceModel   models.EvidenceModel

	err error
}

// NewModel creates a new TUI model
func NewModel(state *store.State, role string) Model {
	return Model{
		state:           state,
		currentScreen:   ScreenHome,
		role:            role,
		homeModel:       models.NewHomeModel(state),
		sourcesModel:    models.NewSourcesModel(state),
		frameworksModel: models.NewFrameworksModel(state),
		controlsModel:   models.NewControlsModel(state, ""),
		evidenceModel:   models.NewEvidenceModel(state, ""),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Quit the application
			return m, tea.Quit

		case "?":
			// Toggle help screen
			if m.currentScreen == ScreenHelp {
				m.currentScreen = ScreenHome
			} else {
				m.currentScreen = ScreenHelp
			}
			return m, nil

		case "esc", "backspace":
			// Go back to home
			if m.currentScreen != ScreenHome {
				m.currentScreen = ScreenHome
			}
			return m, nil

		case "1":
			m.currentScreen = ScreenSources
			return m, nil
		case "2":
			m.currentScreen = ScreenFrameworks
			return m, nil
		case "3":
			m.currentScreen = ScreenControls
			return m, nil
		case "4":
			m.currentScreen = ScreenEvidence
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update all screen models with new size
		m.homeModel.SetSize(msg.Width, msg.Height)
		m.sourcesModel.SetSize(msg.Width, msg.Height)
		m.frameworksModel.SetSize(msg.Width, msg.Height)
		m.controlsModel.SetSize(msg.Width, msg.Height)
		m.evidenceModel.SetSize(msg.Width, msg.Height)
		return m, nil
	}

	// Delegate to current screen's model
	switch m.currentScreen {
	case ScreenHome:
		updatedModel, cmd := m.homeModel.Update(msg)
		m.homeModel = updatedModel

		// Check if home wants to navigate to another screen
		if m.homeModel.SelectedScreen() == 1 {
			m.currentScreen = ScreenSources
			m.homeModel.ResetSelection()
		} else if m.homeModel.SelectedScreen() == 2 {
			m.currentScreen = ScreenFrameworks
			m.homeModel.ResetSelection()
		} else if m.homeModel.SelectedScreen() == 3 {
			m.currentScreen = ScreenControls
			m.homeModel.ResetSelection()
		} else if m.homeModel.SelectedScreen() == 4 {
			m.currentScreen = ScreenEvidence
			m.homeModel.ResetSelection()
		}
		return m, cmd

	case ScreenSources:
		updatedModel, cmd := m.sourcesModel.Update(msg)
		m.sourcesModel = updatedModel
		return m, cmd

	case ScreenFrameworks:
		updatedModel, cmd := m.frameworksModel.Update(msg)
		m.frameworksModel = updatedModel
		return m, cmd

	case ScreenControls:
		updatedModel, cmd := m.controlsModel.Update(msg)
		m.controlsModel = updatedModel
		return m, cmd

	case ScreenEvidence:
		updatedModel, cmd := m.evidenceModel.Update(msg)
		m.evidenceModel = updatedModel
		return m, cmd
	}

	return m, nil
}

// View renders the model
func (m Model) View() string {
	if m.width < 80 || m.height < 24 {
		return styles.ErrorStyle.Render(
			"Terminal too small!\n\n" +
				fmt.Sprintf("Current size: %dx%d\n", m.width, m.height) +
				"Minimum required: 80x24\n\n" +
				"Please resize your terminal and press any key.",
		)
	}

	// Render help screen
	if m.currentScreen == ScreenHelp {
		return m.renderHelp()
	}

	// Render current screen
	switch m.currentScreen {
	case ScreenHome:
		return m.homeModel.View()
	case ScreenSources:
		return m.sourcesModel.View()
	case ScreenFrameworks:
		return m.frameworksModel.View()
	case ScreenControls:
		return m.controlsModel.View()
	case ScreenEvidence:
		return m.evidenceModel.View()
	default:
		return m.homeModel.View()
	}
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	help := styles.TitleStyle.Render("Keyboard Shortcuts") + "\n\n"

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"1-4", "Navigate to screen (Sources, Frameworks, Controls, Evidence)"},
		{"↑/↓", "Navigate lists"},
		{"←/→", "Navigate tabs"},
		{"Enter", "Select item"},
		{"Esc/Backspace", "Go back to home"},
		{"?", "Toggle help"},
		{"q/Ctrl+C", "Quit"},
	}

	for _, s := range shortcuts {
		help += fmt.Sprintf("  %s  %s\n",
			styles.KeyStyle.Render(s.key),
			s.desc,
		)
	}

	help += "\n" + styles.SubtleStyle.Render("Press ? to close help")

	return styles.ContainerStyle.Render(help)
}
